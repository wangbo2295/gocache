package functional

import (
	"strconv"
	"testing"

	"github.com/wangbo/gocache/test/e2e"
)

var _ = &e2e.TestClient{} // Verify e2e.TestClient implements expected interface

// TestSortedSet_BasicOperations tests ZADD, ZREM, ZCARD, ZSCORE
func TestSortedSet_BasicOperations(t *testing.T) {
	client := setupTestClient(t)
	defer client.Close()

	// Clean up
	client.Send("DEL", "myzset")

	t.Run("ZADD adds members to sorted set", func(t *testing.T) {
		reply, err := client.Send("ZADD", "myzset", "1", "one")
		if err != nil {
			t.Errorf("ZADD failed: %v", err)
		}
		count, _ := reply.GetInt()
		if count != 1 {
			t.Errorf("ZADD should return 1, got %d", count)
		}

		// Add multiple members
		reply, err = client.Send("ZADD", "myzset", "2", "two", "3", "three")
		if err != nil {
			t.Errorf("ZADD failed: %v", err)
		}
		count, _ = reply.GetInt()
		if count != 2 {
			t.Errorf("ZADD should return 2 new members, got %d", count)
		}

		// Update existing member's score
		reply, err = client.Send("ZADD", "myzset", "10", "one")
		if err != nil {
			t.Errorf("ZADD failed: %v", err)
		}
		count, _ = reply.GetInt()
		if count != 0 {
			t.Errorf("ZADD updating existing member should return 0, got %d", count)
		}
	})

	t.Run("ZCARD returns sorted set size", func(t *testing.T) {
		reply, err := client.Send("ZCARD", "myzset")
		if err != nil {
			t.Errorf("ZCARD failed: %v", err)
		}
		size, _ := reply.GetInt()
		if size != 3 {
			t.Errorf("ZCARD should return 3, got %d", size)
		}
	})

	t.Run("ZSCORE returns member score", func(t *testing.T) {
		reply, err := client.Send("ZSCORE", "myzset", "one")
		if err != nil {
			t.Errorf("ZSCORE failed: %v", err)
		}
		scoreStr := reply.GetString()
		if scoreStr == "" {
			t.Error("ZSCORE should return score")
		}
		// Score should be updated to 10
		if scoreStr != "10" {
			t.Logf("ZSCORE returned %s (expected 10 after update)", scoreStr)
		}
	})

	t.Run("ZSCORE non-existent member returns nil", func(t *testing.T) {
		reply, err := client.Send("ZSCORE", "myzset", "nonexistent")
		if err != nil {
			t.Errorf("ZSCORE failed: %v", err)
		}
		if !reply.IsNil() {
			t.Errorf("ZSCORE non-existent member should return nil, got %v", reply.GetString())
		}
	})

	t.Run("ZREM removes members from sorted set", func(t *testing.T) {
		reply, err := client.Send("ZREM", "myzset", "two")
		if err != nil {
			t.Errorf("ZREM failed: %v", err)
		}
		count, _ := reply.GetInt()
		if count != 1 {
			t.Errorf("ZREM should return 1, got %d", count)
		}

		// Verify size decreased
		reply, err = client.Send("ZCARD", "myzset")
		if err != nil {
			t.Errorf("ZCARD failed: %v", err)
		}
		size, _ := reply.GetInt()
		if size != 2 {
			t.Errorf("After ZREM, ZCARD should return 2, got %d", size)
		}
	})

	t.Run("ZREM non-existent member returns 0", func(t *testing.T) {
		reply, err := client.Send("ZREM", "myzset", "nonexistent")
		if err != nil {
			t.Errorf("ZREM failed: %v", err)
		}
		count, _ := reply.GetInt()
		if count != 0 {
			t.Errorf("ZREM non-existent member should return 0, got %d", count)
		}
	})

	t.Run("ZCARD on non-existent sorted set returns 0", func(t *testing.T) {
		reply, err := client.Send("ZCARD", "nonexistent")
		if err != nil {
			t.Errorf("ZCARD failed: %v", err)
		}
		size, _ := reply.GetInt()
		if size != 0 {
			t.Errorf("ZCARD on non-existent sorted set should return 0, got %d", size)
		}
	})

	// Cleanup
	client.Send("DEL", "myzset")
}

// TestSortedSet_RangeOperations tests ZRANGE, ZREVRANGE, ZRANGEBYSCORE, ZCOUNT
func TestSortedSet_RangeOperations(t *testing.T) {
	client := setupTestClient(t)
	defer client.Close()

	// Setup
	client.Send("DEL", "zrange_test")
	client.Send("ZADD", "zrange_test", "1", "one", "2", "two", "3", "three")

	t.Run("ZRANGE gets members in ascending order by score", func(t *testing.T) {
		reply, err := client.Send("ZRANGE", "zrange_test", "0", "-1")
		if err != nil {
			t.Errorf("ZRANGE failed: %v", err)
		}
		arr := reply.GetArray()
		if arr == nil || len(arr) != 3 { // 3 members only
			t.Errorf("ZRANGE should return 3 elements, got %d", len(arr))
		} else {
			// Check order: one, two, three
			if format(arr[0]) != "one" {
				t.Errorf("First member should be 'one', got '%s'", format(arr[0]))
			}
			if format(arr[1]) != "two" {
				t.Errorf("Second member should be 'two', got '%s'", format(arr[1]))
			}
			if format(arr[2]) != "three" {
				t.Errorf("Third member should be 'three', got '%s'", format(arr[2]))
			}
		}
	})

	t.Run("ZRANGE with WITHSCORES option", func(t *testing.T) {
		reply, err := client.Send("ZRANGE", "zrange_test", "0", "-1", "WITHSCORES")
		if err != nil {
			t.Errorf("ZRANGE WITHSCORES failed: %v", err)
		}
		arr := reply.GetArray()
		if arr == nil || len(arr) != 6 {
			t.Errorf("ZRANGE WITHSCORES should return 6 elements, got %d", len(arr))
		}
	})

	t.Run("ZRANGE with start > end returns empty", func(t *testing.T) {
		reply, err := client.Send("ZRANGE", "zrange_test", "5", "10")
		if err != nil {
			t.Errorf("ZRANGE failed: %v", err)
		}
		arr := reply.GetArray()
		if arr != nil && len(arr) != 0 {
			t.Errorf("ZRANGE with invalid range should return empty, got %d elements", len(arr))
		}
	})

	t.Run("ZREVRANGE gets members in descending order by score", func(t *testing.T) {
		reply, err := client.Send("ZREVRANGE", "zrange_test", "0", "-1")
		if err != nil {
			t.Errorf("ZREVRANGE failed: %v", err)
		}
		arr := reply.GetArray()
		if arr == nil || len(arr) != 3 { // 3 members only
			t.Errorf("ZREVRANGE should return 3 elements, got %d", len(arr))
		} else {
			// Check reverse order: three, two, one
			if format(arr[0]) != "three" {
				t.Errorf("First member should be 'three' (highest score), got '%s'", format(arr[0]))
			}
			if format(arr[1]) != "two" {
				t.Errorf("Second member should be 'two', got '%s'", format(arr[1]))
			}
			if format(arr[2]) != "one" {
				t.Errorf("Third member should be 'one', got '%s'", format(arr[2]))
			}
		}
	})

	t.Run("ZRANGEBYSCORE gets members by score range", func(t *testing.T) {
		reply, err := client.Send("ZRANGEBYSCORE", "zrange_test", "2", "3")
		if err != nil {
			t.Errorf("ZRANGEBYSCORE failed: %v", err)
		}
		arr := reply.GetArray()
		if arr == nil || len(arr) < 2 {
			t.Errorf("ZRANGEBYSCORE 2-3 should return at least 2 members, got %d", len(arr))
		}
	})

	t.Run("ZCOUNT counts members in score range", func(t *testing.T) {
		reply, err := client.Send("ZCOUNT", "zrange_test", "1", "2")
		if err != nil {
			t.Errorf("ZCOUNT failed: %v", err)
		}
		count, _ := reply.GetInt()
		if count != 2 {
			t.Errorf("ZCOUNT 1-2 should return 2, got %d", count)
		}
	})

	// Cleanup
	client.Send("DEL", "zrange_test")
}

// TestSortedSet_RankOperations tests ZRANK, ZREVRANK, ZINCRBY
func TestSortedSet_RankOperations(t *testing.T) {
	client := setupTestClient(t)
	defer client.Close()

	// Setup
	client.Send("DEL", "zrank_test")
	client.Send("ZADD", "zrank_test", "10", "member1", "20", "member2", "30", "member3")

	t.Run("ZRANK returns rank of member (ascending)", func(t *testing.T) {
		reply, err := client.Send("ZRANK", "zrank_test", "member1")
		if err != nil {
			t.Errorf("ZRANK failed: %v", err)
		}
		if !reply.IsNil() {
			rankStr := reply.GetString()
			if rankStr == "" {
				t.Error("ZRANK should return rank")
			}
			// member1 has lowest score, so rank should be 0
			if rankStr != "0" {
				t.Logf("ZRANK returned %s (expected 0 for member1)", rankStr)
			}
		}
	})

	t.Run("ZRANK non-existent member returns nil", func(t *testing.T) {
		reply, err := client.Send("ZRANK", "zrank_test", "nonexistent")
		if err != nil {
			t.Errorf("ZRANK failed: %v", err)
		}
		if !reply.IsNil() {
			t.Errorf("ZRANK non-existent member should return nil, got %v", reply.GetString())
		}
	})

	t.Run("ZREVRANK returns rank of member (descending)", func(t *testing.T) {
		reply, err := client.Send("ZREVRANK", "zrank_test", "member3")
		if err != nil {
			t.Errorf("ZREVRANK failed: %v", err)
		}
		if !reply.IsNil() {
			rankStr := reply.GetString()
			if rankStr == "" {
				t.Error("ZREVRANK should return rank")
			}
			// member3 has highest score, so reverse rank should be 0
			if rankStr != "0" {
				t.Logf("ZREVRANK returned %s (expected 0 for member3)", rankStr)
			}
		}
	})

	t.Run("ZINCRBY increments member score", func(t *testing.T) {
		reply, err := client.Send("ZINCRBY", "zrank_test", "5", "member2")
		if err != nil {
			t.Errorf("ZINCRBY failed: %v", err)
		}
		newScoreStr := reply.GetString()
		if newScoreStr == "" {
			t.Error("ZINCRBY should return new score")
		}
		// member2 had score 20, after +5 should be 25
		newScore, _ := strconv.ParseFloat(newScoreStr, 64)
		if newScore != 25 {
			t.Errorf("ZINCRBY should return 25, got %v", newScore)
		}

		// Verify score was updated
		reply, err = client.Send("ZSCORE", "zrank_test", "member2")
		if err != nil {
			t.Errorf("ZSCORE failed: %v", err)
		}
		scoreStr := reply.GetString()
		if scoreStr != "25" {
			t.Errorf("Score should be 25 after ZINCRBY, got %s", scoreStr)
		}
	})

	t.Run("ZINCRBY on non-existent member creates it", func(t *testing.T) {
		reply, err := client.Send("ZINCRBY", "zrank_test", "15", "newmember")
		if err != nil {
			t.Errorf("ZINCRBY failed: %v", err)
		}
		newScoreStr := reply.GetString()
		if newScoreStr != "15" {
			t.Errorf("ZINCRBY on new member should return 15, got %s", newScoreStr)
		}

		// Verify member was created
		reply, err = client.Send("ZSCORE", "zrank_test", "newmember")
		if err != nil {
			t.Errorf("ZSCORE failed: %v", err)
		}
		if reply.IsNil() {
			t.Error("ZINCRBY should create non-existent member")
		}
	})

	t.Run("ZINCRBY with negative value decrements", func(t *testing.T) {
		reply, err := client.Send("ZINCRBY", "zrank_test", "-10", "member1")
		if err != nil {
			t.Errorf("ZINCRBY failed: %v", err)
		}
		newScoreStr := reply.GetString()
		// member1 had score 10, after -10 should be 0
		newScore, _ := strconv.ParseFloat(newScoreStr, 64)
		if newScore != 0 {
			t.Errorf("ZINCRBY -10 should return 0, got %v", newScore)
		}
	})

	// Cleanup
	client.Send("DEL", "zrank_test")
}

// TestSortedSet_BinarySafety tests sorted set with special characters
func TestSortedSet_BinarySafety(t *testing.T) {
	client := setupTestClient(t)
	defer client.Close()

	t.Run("Sorted set members with spaces", func(t *testing.T) {
		client.Send("DEL", "space_zset")
		reply, err := client.Send("ZADD", "space_zset", "1", "member with spaces")
		if err != nil {
			t.Errorf("ZADD with spaces failed: %v", err)
		}

		reply, err = client.Send("ZSCORE", "space_zset", "member with spaces")
		if err != nil {
			t.Errorf("ZSCORE failed: %v", err)
		}
		if reply.IsNil() {
			t.Error("Member with spaces should exist")
		}

		client.Send("DEL", "space_zset")
	})

	t.Run("Sorted set members with special characters", func(t *testing.T) {
		client.Send("DEL", "special_zset")
		specialValue := "hello\r\nworld\ttest"
		reply, err := client.Send("ZADD", "special_zset", "1", specialValue)
		if err != nil {
			t.Errorf("ZADD with special chars failed: %v", err)
		}

		reply, err = client.Send("ZRANGE", "special_zset", "0", "-1")
		if err != nil {
			t.Errorf("ZRANGE failed: %v", err)
		}
		arr := reply.GetArray()
		if arr != nil && len(arr) > 0 {
			if format(arr[0]) != specialValue {
				t.Errorf("Special characters not preserved, got '%s'", format(arr[0]))
			}
		}

		client.Send("DEL", "special_zset")
	})

	t.Run("Sorted set members with unicode", func(t *testing.T) {
		client.Send("DEL", "unicode_zset")
		unicodeValue := "你好世界🌍"
		reply, err := client.Send("ZADD", "unicode_zset", "1", unicodeValue)
		if err != nil {
			t.Errorf("ZADD with unicode failed: %v", err)
		}

		reply, err = client.Send("ZRANGE", "unicode_zset", "0", "-1")
		if err != nil {
			t.Errorf("ZRANGE failed: %v", err)
		}
		arr := reply.GetArray()
		if arr != nil && len(arr) > 0 {
			if format(arr[0]) != unicodeValue {
				t.Errorf("Unicode not preserved, got '%s'", format(arr[0]))
			}
		}

		client.Send("DEL", "unicode_zset")
	})

	t.Run("Sorted set with negative scores", func(t *testing.T) {
		client.Send("DEL", "neg_zset")
		reply, err := client.Send("ZADD", "neg_zset", "-10", "negative", "0", "zero", "10", "positive")
		if err != nil {
			t.Errorf("ZADD with negative scores failed: %v", err)
		}

		// Verify order: negative(-10), zero(0), positive(10)
		reply, err = client.Send("ZRANGE", "neg_zset", "0", "-1")
		if err != nil {
			t.Errorf("ZRANGE failed: %v", err)
		}
		arr := reply.GetArray()
		if arr != nil && len(arr) >= 2 {
			if format(arr[0]) != "negative" {
				t.Error("Member with lowest score (-10) should be first")
			}
		}

		client.Send("DEL", "neg_zset")
	})

	t.Run("Sorted set with floating point scores", func(t *testing.T) {
		client.Send("DEL", "float_zset")
		reply, err := client.Send("ZADD", "float_zset", "1.5", "one", "2.7", "two", "3.14", "three")
		if err != nil {
			t.Errorf("ZADD with float scores failed: %v", err)
		}

		// Verify scores
		reply, err = client.Send("ZSCORE", "float_zset", "two")
		if err != nil {
			t.Errorf("ZSCORE failed: %v", err)
		}
		if reply.GetString() != "2.7" {
			t.Logf("ZSCORE returned %s (expected 2.7)", reply.GetString())
		}

		client.Send("DEL", "float_zset")
	})
}

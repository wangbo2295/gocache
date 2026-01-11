package functional

import (
	"testing"

	"github.com/wangbo/gocache/test/e2e"
)

var _ = &e2e.TestClient{} // Verify e2e.TestClient implements expected interface

// TestSet_BasicOperations tests SADD, SREM, SMEMBERS, SCARD
func TestSet_BasicOperations(t *testing.T) {
	client := setupTestClient(t)
	defer client.Close()

	// Clean up
	client.Send("DEL", "myset")

	t.Run("SADD adds members to set", func(t *testing.T) {
		reply, err := client.Send("SADD", "myset", "a")
		if err != nil {
			t.Errorf("SADD failed: %v", err)
		}
		count, _ := reply.GetInt()
		if count != 1 {
			t.Errorf("SADD should return 1, got %d", count)
		}

		// Add more members
		reply, err = client.Send("SADD", "myset", "b", "c")
		if err != nil {
			t.Errorf("SADD failed: %v", err)
		}
		count, _ = reply.GetInt()
		if count != 2 {
			t.Errorf("SADD should return 2 new members, got %d", count)
		}

		// Add duplicate member
		reply, err = client.Send("SADD", "myset", "a")
		if err != nil {
			t.Errorf("SADD failed: %v", err)
		}
		count, _ = reply.GetInt()
		if count != 0 {
			t.Errorf("SADD duplicate should return 0, got %d", count)
		}
	})

	t.Run("SMEMBERS gets all members", func(t *testing.T) {
		reply, err := client.Send("SMEMBERS", "myset")
		if err != nil {
			t.Errorf("SMEMBERS failed: %v", err)
		}
		arr := reply.GetArray()
		if arr == nil || len(arr) != 3 {
			t.Errorf("SMEMBERS should return 3 members, got %d", len(arr))
		}
	})

	t.Run("SCARD returns set size", func(t *testing.T) {
		reply, err := client.Send("SCARD", "myset")
		if err != nil {
			t.Errorf("SCARD failed: %v", err)
		}
		size, _ := reply.GetInt()
		if size != 3 {
			t.Errorf("SCARD should return 3, got %d", size)
		}
	})

	t.Run("SREM removes members from set", func(t *testing.T) {
		reply, err := client.Send("SREM", "myset", "b")
		if err != nil {
			t.Errorf("SREM failed: %v", err)
		}
		count, _ := reply.GetInt()
		if count != 1 {
			t.Errorf("SREM should return 1, got %d", count)
		}

		// Verify size decreased
		reply, err = client.Send("SCARD", "myset")
		if err != nil {
			t.Errorf("SCARD failed: %v", err)
		}
		size, _ := reply.GetInt()
		if size != 2 {
			t.Errorf("After SREM, SCARD should return 2, got %d", size)
		}
	})

	t.Run("SREM non-existent member returns 0", func(t *testing.T) {
		reply, err := client.Send("SREM", "myset", "nonexistent")
		if err != nil {
			t.Errorf("SREM failed: %v", err)
		}
		count, _ := reply.GetInt()
		if count != 0 {
			t.Errorf("SREM non-existent member should return 0, got %d", count)
		}
	})

	t.Run("SMEMBERS on non-existent set returns empty", func(t *testing.T) {
		reply, err := client.Send("SMEMBERS", "nonexistent")
		if err != nil {
			t.Errorf("SMEMBERS failed: %v", err)
		}
		arr := reply.GetArray()
		if arr != nil && len(arr) != 0 {
			t.Errorf("SMEMBERS on non-existent set should return empty, got %d elements", len(arr))
		}
	})

	t.Run("SCARD on non-existent set returns 0", func(t *testing.T) {
		reply, err := client.Send("SCARD", "nonexistent")
		if err != nil {
			t.Errorf("SCARD failed: %v", err)
		}
		size, _ := reply.GetInt()
		if size != 0 {
			t.Errorf("SCARD on non-existent set should return 0, got %d", size)
		}
	})

	// Cleanup
	client.Send("DEL", "myset")
}

// TestSet_MemberOperations tests SISMEMBER, SPOP, SRANDMEMBER
func TestSet_MemberOperations(t *testing.T) {
	client := setupTestClient(t)
	defer client.Close()

	t.Run("SISMEMBER checks if member exists", func(t *testing.T) {
		client.Send("DEL", "testset")
		client.Send("SADD", "testset", "a", "b", "c")

		// Existing member
		reply, err := client.Send("SISMEMBER", "testset", "b")
		if err != nil {
			t.Errorf("SISMEMBER failed: %v", err)
		}
		count, _ := reply.GetInt()
		if count != 1 {
			t.Errorf("SISMEMBER should return 1 for existing member, got %d", count)
		}

		// Non-existent member
		reply, err = client.Send("SISMEMBER", "testset", "x")
		if err != nil {
			t.Errorf("SISMEMBER failed: %v", err)
		}
		count, _ = reply.GetInt()
		if count != 0 {
			t.Errorf("SISMEMBER should return 0 for non-existent member, got %d", count)
		}

		client.Send("DEL", "testset")
	})

	t.Run("SPOP randomly removes and returns member", func(t *testing.T) {
		client.Send("DEL", "popset")
		client.Send("SADD", "popset", "x", "y", "z")

		// Pop one element
		reply, err := client.Send("SPOP", "popset")
		if err != nil {
			t.Errorf("SPOP failed: %v", err)
		}
		popped := reply.GetString()
		if popped != "x" && popped != "y" && popped != "z" {
			t.Errorf("SPOP returned unexpected value: %s", popped)
		}

		// Verify size decreased
		reply, err = client.Send("SCARD", "popset")
		if err != nil {
			t.Errorf("SCARD failed: %v", err)
		}
		size, _ := reply.GetInt()
		if size != 2 {
			t.Errorf("After SPOP, set size should be 2, got %d", size)
		}

		client.Send("DEL", "popset")
	})

	t.Run("SPOP from empty set returns nil", func(t *testing.T) {
		client.Send("DEL", "emptyset")
		reply, err := client.Send("SPOP", "emptyset")
		if err != nil {
			t.Errorf("SPOP failed: %v", err)
		}
		if !reply.IsNil() {
			t.Errorf("SPOP from empty set should return nil, got %v", reply.GetString())
		}
	})

	t.Run("SRANDMEMBER gets random member without removing", func(t *testing.T) {
		client.Send("DEL", "randset")
		client.Send("SADD", "randset", "a", "b", "c")

		reply, err := client.Send("SRANDMEMBER", "randset")
		if err != nil {
			t.Errorf("SRANDMEMBER failed: %v", err)
		}
		member := reply.GetString()
		if member != "a" && member != "b" && member != "c" {
			t.Errorf("SRANDMEMBER returned unexpected value: %s", member)
		}

		// Verify size unchanged
		reply, err = client.Send("SCARD", "randset")
		if err != nil {
			t.Errorf("SCARD failed: %v", err)
		}
		size, _ := reply.GetInt()
		if size != 3 {
			t.Errorf("SRANDMEMBER should not remove member, size should be 3, got %d", size)
		}

		client.Send("DEL", "randset")
	})

	t.Run("SRANDMEMBER on empty set returns nil", func(t *testing.T) {
		client.Send("DEL", "emptyset2")
		reply, err := client.Send("SRANDMEMBER", "emptyset2")
		if err != nil {
			t.Errorf("SRANDMEMBER failed: %v", err)
		}
		if !reply.IsNil() {
			t.Errorf("SRANDMEMBER on empty set should return nil, got %v", reply.GetString())
		}
	})
}

// TestSet_MoveOperations tests SMOVE
func TestSet_MoveOperations(t *testing.T) {
	client := setupTestClient(t)
	defer client.Close()

	t.Run("SMOVE moves member from one set to another", func(t *testing.T) {
		client.Send("DEL", "set1", "set2")
		client.Send("SADD", "set1", "a", "b")
		client.Send("SADD", "set2", "c")

		reply, err := client.Send("SMOVE", "set1", "set2", "a")
		if err != nil {
			t.Errorf("SMOVE failed: %v", err)
		}
		count, _ := reply.GetInt()
		if count != 1 {
			t.Errorf("SMOVE should return 1, got %d", count)
		}

		// Verify "a" is in set2
		reply, err = client.Send("SISMEMBER", "set2", "a")
		if err != nil {
			t.Errorf("SISMEMBER failed: %v", err)
		}
		count, _ = reply.GetInt()
		if count != 1 {
			t.Error("SMOVE should move element to destination")
		}

		// Verify "a" is not in set1
		reply, err = client.Send("SISMEMBER", "set1", "a")
		if err != nil {
			t.Errorf("SISMEMBER failed: %v", err)
		}
		count, _ = reply.GetInt()
		if count != 0 {
			t.Error("SMOVE should remove element from source")
		}

		client.Send("DEL", "set1", "set2")
	})

	t.Run("SMOVE non-existent member returns 0", func(t *testing.T) {
		client.Send("DEL", "set3", "set4")
		client.Send("SADD", "set3", "a")

		reply, err := client.Send("SMOVE", "set3", "set4", "x")
		if err != nil {
			t.Errorf("SMOVE failed: %v", err)
		}
		count, _ := reply.GetInt()
		if count != 0 {
			t.Errorf("SMOVE non-existent member should return 0, got %d", count)
		}

		client.Send("DEL", "set3", "set4")
	})
}

// TestSet_SetOperations tests SDIFF, SINTER, SUNION and their STORE variants
func TestSet_SetOperations(t *testing.T) {
	client := setupTestClient(t)
	defer client.Close()

	// Setup sets
	client.Send("DEL", "setA", "setB", "setC")
	client.Send("SADD", "setA", "a", "b", "c")
	client.Send("SADD", "setB", "c", "d", "e")

	t.Run("SDIFF returns difference", func(t *testing.T) {
		// setA - setB = {a, b}
		reply, err := client.Send("SDIFF", "setA", "setB")
		if err != nil {
			t.Errorf("SDIFF failed: %v", err)
		}
		arr := reply.GetArray()
		if arr == nil || len(arr) != 2 {
			t.Errorf("SDIFF should return 2 elements, got %d", len(arr))
		}
	})

	t.Run("SINTER returns intersection", func(t *testing.T) {
		reply, err := client.Send("SINTER", "setA", "setB")
		if err != nil {
			t.Errorf("SINTER failed: %v", err)
		}
		arr := reply.GetArray()
		if arr == nil || len(arr) != 1 {
			t.Errorf("SINTER should return 1 element (c), got %d", len(arr))
		}
	})

	t.Run("SUNION returns union", func(t *testing.T) {
		reply, err := client.Send("SUNION", "setA", "setB")
		if err != nil {
			t.Errorf("SUNION failed: %v", err)
		}
		arr := reply.GetArray()
		if arr == nil || len(arr) != 5 {
			t.Errorf("SUNION should return 5 elements, got %d", len(arr))
		}
	})

	t.Run("SDIFFSTORE stores difference", func(t *testing.T) {
		reply, err := client.Send("SDIFFSTORE", "setC", "setA", "setB")
		if err != nil {
			t.Errorf("SDIFFSTORE failed: %v", err)
		}
		count, _ := reply.GetInt()
		if count != 2 {
			t.Errorf("SDIFFSTORE should return 2, got %d", count)
		}

		// Verify setC has 2 elements
		reply, err = client.Send("SCARD", "setC")
		if err != nil {
			t.Errorf("SCARD failed: %v", err)
		}
		size, _ := reply.GetInt()
		if size != 2 {
			t.Errorf("setC should have 2 elements, got %d", size)
		}
	})

	t.Run("SINTERSTORE stores intersection", func(t *testing.T) {
		client.Send("DEL", "setD")
		reply, err := client.Send("SINTERSTORE", "setD", "setA", "setB")
		if err != nil {
			t.Errorf("SINTERSTORE failed: %v", err)
		}
		count, _ := reply.GetInt()
		if count != 1 {
			t.Logf("SINTERSTORE returned %d (expected 1)", count)
		}

		// Verify setD has 1 element
		reply, err = client.Send("SCARD", "setD")
		if err != nil {
			t.Errorf("SCARD failed: %v", err)
		}
		size, _ := reply.GetInt()
		if size != 1 {
			t.Errorf("setD should have 1 element, got %d", size)
		}

		client.Send("DEL", "setD")
	})

	t.Run("SUNIONSTORE stores union", func(t *testing.T) {
		client.Send("DEL", "setE")
		reply, err := client.Send("SUNIONSTORE", "setE", "setA", "setB")
		if err != nil {
			t.Errorf("SUNIONSTORE failed: %v", err)
		}
		count, _ := reply.GetInt()
		if count != 5 {
			t.Logf("SUNIONSTORE returned %d (expected 5)", count)
		}

		// Verify setE has 5 elements
		reply, err = client.Send("SCARD", "setE")
		if err != nil {
			t.Errorf("SCARD failed: %v", err)
		}
		size, _ := reply.GetInt()
		if size != 5 {
			t.Errorf("setE should have 5 elements, got %d", size)
		}

		client.Send("DEL", "setE")
	})

	// Cleanup
	client.Send("DEL", "setA", "setB", "setC")
}

// TestSet_BinarySafety tests set with special characters
func TestSet_BinarySafety(t *testing.T) {
	client := setupTestClient(t)
	defer client.Close()

	t.Run("Set members with spaces", func(t *testing.T) {
		client.Send("DEL", "spaceset")
		reply, err := client.Send("SADD", "spaceset", "member with spaces")
		if err != nil {
			t.Errorf("SADD with spaces failed: %v", err)
		}

		reply, err = client.Send("SISMEMBER", "spaceset", "member with spaces")
		if err != nil {
			t.Errorf("SISMEMBER failed: %v", err)
		}
		count, _ := reply.GetInt()
		if count != 1 {
			t.Error("Member with spaces should exist")
		}

		client.Send("DEL", "spaceset")
	})

	t.Run("Set members with special characters", func(t *testing.T) {
		client.Send("DEL", "specialset")
		specialValue := "hello\r\nworld\ttest"
		reply, err := client.Send("SADD", "specialset", specialValue)
		if err != nil {
			t.Errorf("SADD with special chars failed: %v", err)
		}

		reply, err = client.Send("SPOP", "specialset")
		if err != nil {
			t.Errorf("SPOP failed: %v", err)
		}
		if reply.GetString() != specialValue {
			t.Errorf("Special characters not preserved, got '%s'", reply.GetString())
		}

		client.Send("DEL", "specialset")
	})

	t.Run("Set members with unicode", func(t *testing.T) {
		client.Send("DEL", "unicodeset")
		unicodeValue := "你好世界🌍"
		reply, err := client.Send("SADD", "unicodeset", unicodeValue)
		if err != nil {
			t.Errorf("SADD with unicode failed: %v", err)
		}

		reply, err = client.Send("SMEMBERS", "unicodeset")
		if err != nil {
			t.Errorf("SMEMBERS failed: %v", err)
		}
		arr := reply.GetArray()
		if arr != nil && len(arr) > 0 {
			if format(arr[0]) != unicodeValue {
				t.Errorf("Unicode not preserved, got '%s'", format(arr[0]))
			}
		}

		client.Send("DEL", "unicodeset")
	})
}

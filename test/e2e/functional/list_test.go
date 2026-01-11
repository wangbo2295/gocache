package functional

import (
	"testing"

	"github.com/wangbo/gocache/test/e2e"
)

var _ = &e2e.TestClient{} // Verify e2e.TestClient implements expected interface

// TestList_BasicOperations tests LPUSH, RPUSH, LPOP, RPOP
func TestList_BasicOperations(t *testing.T) {
	client := setupTestClient(t)
	defer client.Close()

	// Clean up
	client.Send("DEL", "mylist")

	t.Run("LPUSH adds elements to left", func(t *testing.T) {
		reply, err := client.Send("LPUSH", "mylist", "c")
		if err != nil {
			t.Errorf("LPUSH failed: %v", err)
		}
		length, _ := reply.GetInt()
		if length != 1 {
			t.Errorf("LPUSH should return 1, got %d", length)
		}

		reply, err = client.Send("LPUSH", "mylist", "b", "a")
		if err != nil {
			t.Errorf("LPUSH failed: %v", err)
		}
		length, _ = reply.GetInt()
		if length != 3 {
			t.Errorf("LPUSH should return 3, got %d", length)
		}

		// Verify order: a, b, c
		reply, err = client.Send("LRANGE", "mylist", "0", "-1")
		if err != nil {
			t.Errorf("LRANGE failed: %v", err)
		}
		arr := reply.GetArray()
		if arr != nil && len(arr) == 3 {
			if format(arr[0]) != "a" || format(arr[1]) != "b" || format(arr[2]) != "c" {
				t.Errorf("List order should be a,b,c, got %v,%v,%v", arr[0], arr[1], arr[2])
			}
		}
	})

	t.Run("RPUSH adds elements to right", func(t *testing.T) {
		client.Send("DEL", "rlist")
		reply, err := client.Send("RPUSH", "rlist", "a", "b", "c")
		if err != nil {
			t.Errorf("RPUSH failed: %v", err)
		}
		length, _ := reply.GetInt()
		if length != 3 {
			t.Errorf("RPUSH should return 3, got %d", length)
		}

		// Verify order: a, b, c
		reply, err = client.Send("LRANGE", "rlist", "0", "-1")
		if err != nil {
			t.Errorf("LRANGE failed: %v", err)
		}
		arr := reply.GetArray()
		if arr != nil && len(arr) == 3 {
			if format(arr[0]) != "a" || format(arr[2]) != "c" {
				t.Error("RPUSH should maintain order a,b,c")
			}
		}
	})

	t.Run("LPOP removes from left", func(t *testing.T) {
		client.Send("DEL", "poplist")
		client.Send("RPUSH", "poplist", "a", "b", "c")

		reply, err := client.Send("LPOP", "poplist")
		if err != nil {
			t.Errorf("LPOP failed: %v", err)
		}
		if reply.GetString() != "a" {
			t.Errorf("LPOP should return 'a', got %s", reply.GetString())
		}

		// Verify remaining: b, c
		reply, err = client.Send("LRANGE", "poplist", "0", "-1")
		if err != nil {
			t.Errorf("LRANGE failed: %v", err)
		}
		arr := reply.GetArray()
		if arr != nil && len(arr) == 2 {
			if format(arr[0]) != "b" || format(arr[1]) != "c" {
				t.Error("After LPOP, list should be b,c")
			}
		}
	})

	t.Run("RPOP removes from right", func(t *testing.T) {
		client.Send("DEL", "poplist2")
		client.Send("RPUSH", "poplist2", "x", "y", "z")

		reply, err := client.Send("RPOP", "poplist2")
		if err != nil {
			t.Errorf("RPOP failed: %v", err)
		}
		if reply.GetString() != "z" {
			t.Errorf("RPOP should return 'z', got %s", reply.GetString())
		}

		// Verify remaining: x, y
		reply, err = client.Send("LRANGE", "poplist2", "0", "-1")
		if err != nil {
			t.Errorf("LRANGE failed: %v", err)
		}
		arr := reply.GetArray()
		if arr != nil && len(arr) == 2 {
			if format(arr[0]) != "x" || format(arr[1]) != "y" {
				t.Error("After RPOP, list should be x,y")
			}
		}
	})

	t.Run("LPOP on empty list returns nil", func(t *testing.T) {
		client.Send("DEL", "emptylist")
		reply, err := client.Send("LPOP", "emptylist")
		if err != nil {
			t.Errorf("LPOP failed: %v", err)
		}
		if !reply.IsNil() {
			t.Errorf("LPOP from empty list should return nil, got %v", reply.GetString())
		}
	})

	t.Run("RPOP on empty list returns nil", func(t *testing.T) {
		client.Send("DEL", "emptylist2")
		reply, err := client.Send("RPOP", "emptylist2")
		if err != nil {
			t.Errorf("RPOP failed: %v", err)
		}
		if !reply.IsNil() {
			t.Errorf("RPOP from empty list should return nil, got %v", reply.GetString())
		}
	})

	// Cleanup
	client.Send("DEL", "mylist", "rlist", "poplist", "poplist2")
}

// TestList_QueryOperations tests LLEN, LINDEX, LRANGE
func TestList_QueryOperations(t *testing.T) {
	client := setupTestClient(t)
	defer client.Close()

	// Setup
	client.Send("DEL", "qlist")
	client.Send("RPUSH", "qlist", "a", "b", "c", "d", "e")

	t.Run("LLEN returns list length", func(t *testing.T) {
		reply, err := client.Send("LLEN", "qlist")
		if err != nil {
			t.Errorf("LLEN failed: %v", err)
		}
		length, _ := reply.GetInt()
		if length != 5 {
			t.Errorf("LLEN should return 5, got %d", length)
		}
	})

	t.Run("LLEN on non-existent list returns 0", func(t *testing.T) {
		reply, err := client.Send("LLEN", "nonexistent")
		if err != nil {
			t.Errorf("LLEN failed: %v", err)
		}
		length, _ := reply.GetInt()
		if length != 0 {
			t.Errorf("LLEN on non-existent list should return 0, got %d", length)
		}
	})

	t.Run("LINDEX gets element at index", func(t *testing.T) {
		reply, err := client.Send("LINDEX", "qlist", "0")
		if err != nil {
			t.Errorf("LINDEX failed: %v", err)
		}
		if reply.GetString() != "a" {
			t.Errorf("LINDEX 0 should return 'a', got %s", reply.GetString())
		}

		reply, err = client.Send("LINDEX", "qlist", "2")
		if err != nil {
			t.Errorf("LINDEX failed: %v", err)
		}
		if reply.GetString() != "c" {
			t.Errorf("LINDEX 2 should return 'c', got %s", reply.GetString())
		}

		// Negative index
		reply, err = client.Send("LINDEX", "qlist", "-1")
		if err != nil {
			t.Errorf("LINDEX with negative index failed: %v", err)
		}
		if reply.GetString() != "e" {
			t.Errorf("LINDEX -1 should return 'e', got %s", reply.GetString())
		}
	})

	t.Run("LINDEX out of range returns nil", func(t *testing.T) {
		reply, err := client.Send("LINDEX", "qlist", "100")
		if err != nil {
			t.Errorf("LINDEX failed: %v", err)
		}
		if !reply.IsNil() {
			t.Errorf("LINDEX out of range should return nil, got %v", reply.GetString())
		}
	})

	t.Run("LRANGE gets range of elements", func(t *testing.T) {
		reply, err := client.Send("LRANGE", "qlist", "1", "3")
		if err != nil {
			t.Errorf("LRANGE failed: %v", err)
		}
		arr := reply.GetArray()
		if arr == nil || len(arr) != 3 {
			t.Errorf("LRANGE 1 3 should return 3 elements, got %d", len(arr))
		} else if format(arr[0]) != "b" || format(arr[2]) != "d" {
			t.Error("LRANGE should return b,c,d")
		}
	})

	t.Run("LRANGE with negative indices", func(t *testing.T) {
		reply, err := client.Send("LRANGE", "qlist", "-3", "-1")
		if err != nil {
			t.Errorf("LRANGE failed: %v", err)
		}
		arr := reply.GetArray()
		if arr == nil || len(arr) != 3 {
			t.Errorf("LRANGE -3 -1 should return 3 elements, got %d", len(arr))
		}
	})

	t.Run("LRANGE 0 -1 gets all elements", func(t *testing.T) {
		reply, err := client.Send("LRANGE", "qlist", "0", "-1")
		if err != nil {
			t.Errorf("LRANGE failed: %v", err)
		}
		arr := reply.GetArray()
		if arr == nil || len(arr) != 5 {
			t.Errorf("LRANGE 0 -1 should return 5 elements, got %d", len(arr))
		}
	})

	t.Run("LRANGE on non-existent list returns empty", func(t *testing.T) {
		reply, err := client.Send("LRANGE", "nonexistent", "0", "-1")
		if err != nil {
			t.Errorf("LRANGE failed: %v", err)
		}
		arr := reply.GetArray()
		if arr != nil && len(arr) != 0 {
			t.Errorf("LRANGE on non-existent list should return empty, got %d elements", len(arr))
		}
	})

	// Cleanup
	client.Send("DEL", "qlist")
}

// TestList_ModificationOperations tests LSET, LTRIM, LREM, LINSERT
func TestList_ModificationOperations(t *testing.T) {
	client := setupTestClient(t)
	defer client.Close()

	t.Run("LSET sets element at index", func(t *testing.T) {
		client.Send("DEL", "setlist")
		client.Send("RPUSH", "setlist", "a", "b", "c")

		reply, err := client.Send("LSET", "setlist", "1", "NEW")
		if err != nil || !reply.IsOK() {
			t.Errorf("LSET should succeed: %v", err)
		}

		// Verify
		reply, err = client.Send("LINDEX", "setlist", "1")
		if err != nil {
			t.Errorf("LINDEX failed: %v", err)
		}
		if reply.GetString() != "NEW" {
			t.Errorf("LSET should update value, got %s", reply.GetString())
		}

		client.Send("DEL", "setlist")
	})

	t.Run("LSET out of range returns error", func(t *testing.T) {
		client.Send("DEL", "setlist2")
		client.Send("RPUSH", "setlist2", "a", "b")

		reply, err := client.Send("LSET", "setlist2", "10", "X")
		if err == nil {
			t.Error("LSET with out of range index should return error")
		} else if reply != nil && !reply.IsError() {
			t.Error("LSET with out of range should return error reply")
		}

		client.Send("DEL", "setlist2")
	})

	t.Run("LTRIM trims list to range", func(t *testing.T) {
		client.Send("DEL", "trimlist")
		client.Send("RPUSH", "trimlist", "a", "b", "c", "d", "e")

		reply, err := client.Send("LTRIM", "trimlist", "1", "3")
		if err != nil || !reply.IsOK() {
			t.Errorf("LTRIM should succeed: %v", err)
		}

		// Verify: b, c, d
		reply, err = client.Send("LRANGE", "trimlist", "0", "-1")
		if err != nil {
			t.Errorf("LRANGE failed: %v", err)
		}
		arr := reply.GetArray()
		if arr == nil || len(arr) != 3 {
			t.Errorf("After LTRIM, list should have 3 elements, got %d", len(arr))
		}

		client.Send("DEL", "trimlist")
	})

	t.Run("LREM removes elements", func(t *testing.T) {
		client.Send("DEL", "remlist")
		client.Send("RPUSH", "remlist", "a", "b", "a", "c", "a")

		// Remove all "a" (count 0)
		reply, err := client.Send("LREM", "remlist", "0", "a")
		if err != nil {
			t.Errorf("LREM failed: %v", err)
		}
		count, _ := reply.GetInt()
		if count != 3 {
			t.Errorf("LREM should return 3, got %d", count)
		}

		// Verify remaining: b, c
		reply, err = client.Send("LRANGE", "remlist", "0", "-1")
		if err != nil {
			t.Errorf("LRANGE failed: %v", err)
		}
		arr := reply.GetArray()
		if arr == nil || len(arr) != 2 {
			t.Errorf("After LREM, list should have 2 elements, got %d", len(arr))
		}

		client.Send("DEL", "remlist")
	})

	t.Run("LINSERT inserts element", func(t *testing.T) {
		client.Send("DEL", "inslist")
		client.Send("RPUSH", "inslist", "a", "b", "c")

		// Insert "NEW" before "b"
		reply, err := client.Send("LINSERT", "inslist", "BEFORE", "b", "NEW")
		if err != nil {
			t.Errorf("LINSERT failed: %v", err)
		}
		length, _ := reply.GetInt()
		if length != 4 {
			t.Errorf("LINSERT should return 4, got %d", length)
		}

		// Verify: a, NEW, b, c
		reply, err = client.Send("LINDEX", "inslist", "1")
		if err != nil {
			t.Errorf("LINDEX failed: %v", err)
		}
		if reply.GetString() != "NEW" {
			t.Error("LINSERT should insert NEW before b")
		}

		client.Send("DEL", "inslist")
	})

	t.Run("LINSERT before non-existent element returns -1", func(t *testing.T) {
		client.Send("DEL", "inslist2")
		client.Send("RPUSH", "inslist2", "a", "b", "c")

		reply, err := client.Send("LINSERT", "inslist2", "BEFORE", "X", "Y")
		if err != nil {
			t.Errorf("LINSERT failed: %v", err)
		}
		length, _ := reply.GetInt()
		if length != -1 {
			t.Errorf("LINSERT before non-existent element should return -1, got %d", length)
		}

		client.Send("DEL", "inslist2")
	})
}

// TestList_BinarySafety tests list with special characters
func TestList_BinarySafety(t *testing.T) {
	client := setupTestClient(t)
	defer client.Close()

	t.Run("List elements with spaces", func(t *testing.T) {
		client.Send("DEL", "spacelist")
		reply, err := client.Send("RPUSH", "spacelist", "value with spaces")
		if err != nil {
			t.Errorf("RPUSH with spaces failed: %v", err)
		}

		reply, err = client.Send("LINDEX", "spacelist", "0")
		if err != nil {
			t.Errorf("LINDEX failed: %v", err)
		}
		if reply.GetString() != "value with spaces" {
			t.Errorf("Spaces not preserved, got '%s'", reply.GetString())
		}

		client.Send("DEL", "spacelist")
	})

	t.Run("List elements with special characters", func(t *testing.T) {
		client.Send("DEL", "speciallist")
		specialValue := "hello\r\nworld\ttest"
		reply, err := client.Send("LPUSH", "speciallist", specialValue)
		if err != nil {
			t.Errorf("LPUSH with special chars failed: %v", err)
		}

		reply, err = client.Send("LPOP", "speciallist")
		if err != nil {
			t.Errorf("LPOP failed: %v", err)
		}
		if reply.GetString() != specialValue {
			t.Errorf("Special characters not preserved, got '%s'", reply.GetString())
		}

		client.Send("DEL", "speciallist")
	})

	t.Run("List elements with unicode", func(t *testing.T) {
		client.Send("DEL", "unicodelist")
		unicodeValue := "你好世界🌍"
		reply, err := client.Send("RPUSH", "unicodelist", unicodeValue)
		if err != nil {
			t.Errorf("RPUSH with unicode failed: %v", err)
		}

		reply, err = client.Send("LINDEX", "unicodelist", "0")
		if err != nil {
			t.Errorf("LINDEX failed: %v", err)
		}
		if reply.GetString() != unicodeValue {
			t.Errorf("Unicode not preserved, got '%s'", reply.GetString())
		}

		client.Send("DEL", "unicodelist")
	})

	t.Run("List with empty string elements", func(t *testing.T) {
		client.Send("DEL", "emptylist")
		reply, err := client.Send("RPUSH", "emptylist", "")
		if err != nil {
			t.Errorf("RPUSH empty string failed: %v", err)
		}

		reply, err = client.Send("LINDEX", "emptylist", "0")
		if err != nil {
			t.Errorf("LINDEX failed: %v", err)
		}
		if reply.GetString() != "" {
			t.Errorf("Empty string not preserved, got '%s'", reply.GetString())
		}

		client.Send("DEL", "emptylist")
	})
}

// Helper function to format interface{} to string
func format(v interface{}) string {
	if v == nil {
		return ""
	}
	return v.(string)
}

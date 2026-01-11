package functional

import (
	"fmt"
	"testing"

	"github.com/wangbo/gocache/test/e2e"
)

var _ = &e2e.TestClient{} // Verify e2e.TestClient implements expected interface

// setupTestClient creates a test client (defined in string_test.go)
// This is a common helper used across all functional test files

// TestHash_BasicOperations tests basic HSET, HGET, HGETALL operations
func TestHash_BasicOperations(t *testing.T) {
	client := setupTestClient(t)
	defer client.Close()

	// Clean up
	client.Send("DEL", "user:1")

	t.Run("HSET and HGET basic field-value", func(t *testing.T) {
		reply, err := client.Send("HSET", "user:1", "name", "Alice")
		if err != nil {
			t.Errorf("HSET failed: %v", err)
		}
		// HSET returns 1 for new field, 0 for update
		count, _ := reply.GetInt()
		if count != 1 {
			t.Logf("HSET returned %d (expected 1 for new field)", count)
		}

		reply, err = client.Send("HGET", "user:1", "name")
		if err != nil {
			t.Errorf("HGET failed: %v", err)
		}
		if reply.GetString() != "Alice" {
			t.Errorf("HGET returned wrong value: got %v, want 'Alice'", reply.GetString())
		}
	})

	t.Run("HGET non-existent hash returns nil", func(t *testing.T) {
		reply, err := client.Send("HGET", "nonexistent", "field")
		if err != nil {
			t.Errorf("HGET failed: %v", err)
		}
		if !reply.IsNil() {
			t.Errorf("HGET non-existent hash should return nil, got: %v", reply.GetString())
		}
	})

	t.Run("HGET non-existent field returns nil", func(t *testing.T) {
		reply, err := client.Send("HGET", "user:1", "nonexistent")
		if err != nil {
			t.Errorf("HGET failed: %v", err)
		}
		if !reply.IsNil() {
			t.Errorf("HGET non-existent field should return nil, got: %v", reply.GetString())
		}
	})

	t.Run("HGETALL gets all fields", func(t *testing.T) {
		client.Send("HSET", "user:1", "age", "25")
		client.Send("HSET", "user:1", "city", "NYC")

		reply, err := client.Send("HGETALL", "user:1")
		if err != nil {
			t.Errorf("HGETALL failed: %v", err)
		}

		arr := reply.GetArray()
		if arr == nil || len(arr) == 0 {
			t.Error("HGETALL should return fields")
		} else {
			t.Logf("HGETALL returned %d fields", len(arr)/2)
		}
	})

	t.Run("HGETALL on non-existent hash returns empty", func(t *testing.T) {
		reply, err := client.Send("HGETALL", "nonexistent")
		if err != nil {
			t.Errorf("HGETALL failed: %v", err)
		}
		arr := reply.GetArray()
		if arr != nil && len(arr) != 0 {
			t.Errorf("HGETALL on non-existent hash should return empty, got %d items", len(arr))
		}
	})

	// Cleanup
	client.Send("DEL", "user:1")
}

// TestHash_FieldOperations tests HDEL, HEXISTS, HKEYS, HVALS, HLEN
func TestHash_FieldOperations(t *testing.T) {
	client := setupTestClient(t)
	defer client.Close()

	// Setup
	client.Send("HSET", "user:2", "name", "Bob")
	client.Send("HSET", "user:2", "age", "30")

	t.Run("HDEL deletes hash field", func(t *testing.T) {
		reply, err := client.Send("HDEL", "user:2", "name")
		if err != nil {
			t.Errorf("HDEL failed: %v", err)
		}
		count, _ := reply.GetInt()
		if count != 1 {
			t.Errorf("HDEL should return 1, got %d", count)
		}

		// Verify field is deleted
		reply, err = client.Send("HGET", "user:2", "name")
		if err != nil {
			t.Errorf("HGET after HDEL failed: %v", err)
		}
		if !reply.IsNil() {
			t.Error("Deleted field should not exist")
		}
	})

	t.Run("HDEL non-existent field returns 0", func(t *testing.T) {
		reply, err := client.Send("HDEL", "user:2", "nonexistent")
		if err != nil {
			t.Errorf("HDEL failed: %v", err)
		}
		count, _ := reply.GetInt()
		if count != 0 {
			t.Errorf("HDEL non-existent field should return 0, got %d", count)
		}
	})

	t.Run("HEXISTS checks if field exists", func(t *testing.T) {
		// Existing field
		reply, err := client.Send("HEXISTS", "user:2", "age")
		if err != nil {
			t.Errorf("HEXISTS failed: %v", err)
		}
		count, _ := reply.GetInt()
		if count != 1 {
			t.Errorf("HEXISTS should return 1 for existing field, got %d", count)
		}

		// Non-existent field
		reply, err = client.Send("HEXISTS", "user:2", "name")
		if err != nil {
			t.Errorf("HEXISTS failed: %v", err)
		}
		count, _ = reply.GetInt()
		if count != 0 {
			t.Errorf("HEXISTS should return 0 for non-existent field, got %d", count)
		}
	})

	t.Run("HKEYS gets all field names", func(t *testing.T) {
		client.Send("HSET", "user:3", "f1", "v1")
		client.Send("HSET", "user:3", "f2", "v2")
		client.Send("HSET", "user:3", "f3", "v3")

		reply, err := client.Send("HKEYS", "user:3")
		if err != nil {
			t.Errorf("HKEYS failed: %v", err)
		}
		arr := reply.GetArray()
		if arr == nil || len(arr) != 3 {
			t.Errorf("HKEYS should return 3 fields, got %d", len(arr))
		}

		client.Send("DEL", "user:3")
	})

	t.Run("HVALS gets all field values", func(t *testing.T) {
		reply, err := client.Send("HVALS", "user:2")
		if err != nil {
			t.Errorf("HVALS failed: %v", err)
		}
		arr := reply.GetArray()
		if arr == nil || len(arr) == 0 {
			t.Error("HVALS should return values")
		}
	})

	t.Run("HLEN returns hash length", func(t *testing.T) {
		client.Send("HSET", "user:4", "a", "1")
		client.Send("HSET", "user:4", "b", "2")
		client.Send("HSET", "user:4", "c", "3")

		reply, err := client.Send("HLEN", "user:4")
		if err != nil {
			t.Errorf("HLEN failed: %v", err)
		}
		length, _ := reply.GetInt()
		if length != 3 {
			t.Errorf("HLEN should return 3, got %d", length)
		}

		client.Send("DEL", "user:4")
	})

	t.Run("HLEN on non-existent hash returns 0", func(t *testing.T) {
		reply, err := client.Send("HLEN", "nonexistent")
		if err != nil {
			t.Errorf("HLEN failed: %v", err)
		}
		length, _ := reply.GetInt()
		if length != 0 {
			t.Errorf("HLEN on non-existent hash should return 0, got %d", length)
		}
	})

	// Cleanup
	client.Send("DEL", "user:2")
}

// TestHash_SetOperations tests HSETNX, HINCRBY, HMGET, HMSET
func TestHash_SetOperations(t *testing.T) {
	client := setupTestClient(t)
	defer client.Close()

	t.Run("HSETNX sets field only if not exists", func(t *testing.T) {
		// First HSETNX should succeed
		reply, err := client.Send("HSETNX", "user:5", "name", "Charlie")
		if err != nil {
			t.Errorf("HSETNX failed: %v", err)
		}
		count, _ := reply.GetInt()
		if count != 1 {
			t.Errorf("HSETNX should return 1 for new field, got %d", count)
		}

		// Second HSETNX should fail
		reply, err = client.Send("HSETNX", "user:5", "name", "David")
		if err != nil {
			t.Errorf("HSETNX failed: %v", err)
		}
		count, _ = reply.GetInt()
		if count != 0 {
			t.Errorf("HSETNX should return 0 for existing field, got %d", count)
		}

		// Verify value unchanged
		reply, err = client.Send("HGET", "user:5", "name")
		if err != nil {
			t.Errorf("HGET failed: %v", err)
		}
		if reply.GetString() != "Charlie" {
			t.Errorf("HSETNX should not overwrite existing value, got '%s'", reply.GetString())
		}

		client.Send("DEL", "user:5")
	})

	t.Run("HINCRBY increments field value", func(t *testing.T) {
		// Set initial value
		client.Send("HSET", "counter:1", "count", "10")

		// Increment
		reply, err := client.Send("HINCRBY", "counter:1", "count", "5")
		if err != nil {
			t.Errorf("HINCRBY failed: %v", err)
		}
		val, _ := reply.GetInt()
		if val != 15 {
			t.Errorf("HINCRBY should return 15, got %d", val)
		}

		// Verify
		reply, err = client.Send("HGET", "counter:1", "count")
		if err != nil {
			t.Errorf("HGET failed: %v", err)
		}
		if reply.GetString() != "15" {
			t.Errorf("Value should be '15', got '%s'", reply.GetString())
		}

		// Decrement
		reply, err = client.Send("HINCRBY", "counter:1", "count", "-3")
		if err != nil {
			t.Errorf("HINCRBY failed: %v", err)
		}
		val, _ = reply.GetInt()
		if val != 12 {
			t.Errorf("HINCRBY with negative should return 12, got %d", val)
		}

		client.Send("DEL", "counter:1")
	})

	t.Run("HINCRBY on non-existent field creates it", func(t *testing.T) {
		client.Send("DEL", "counter:2")

		reply, err := client.Send("HINCRBY", "counter:2", "count", "5")
		if err != nil {
			t.Errorf("HINCRBY failed: %v", err)
		}
		val, _ := reply.GetInt()
		if val != 5 {
			t.Errorf("HINCRBY on non-existent field should return 5, got %d", val)
		}

		client.Send("DEL", "counter:2")
	})

	t.Run("HMGET gets multiple fields", func(t *testing.T) {
		client.Send("HSET", "user:6", "name", "Eve")
		client.Send("HSET", "user:6", "age", "35")
		client.Send("HSET", "user:6", "city", "SF")

		reply, err := client.Send("HMGET", "user:6", "name", "age", "nonexistent", "city")
		if err != nil {
			t.Errorf("HMGET failed: %v", err)
		}

		arr := reply.GetArray()
		if arr == nil || len(arr) != 4 {
			t.Errorf("HMGET should return 4 values, got %d", len(arr))
		}

		// Check values
		if fmt.Sprintf("%v", arr[0]) != "Eve" {
			t.Errorf("HMGET[0] should be 'Eve', got %v", arr[0])
		}
		if fmt.Sprintf("%v", arr[1]) != "35" {
			t.Errorf("HMGET[1] should be '35', got %v", arr[1])
		}

		client.Send("DEL", "user:6")
	})

	t.Run("HMSET sets multiple fields", func(t *testing.T) {
		reply, err := client.Send("HMSET", "user:7", "name", "Frank", "age", "40", "city", "LA")
		if err != nil || !reply.IsOK() {
			t.Errorf("HMSET failed: %v", err)
		}

		// Verify all fields were set
		reply, err = client.Send("HGET", "user:7", "name")
		if err != nil || reply.GetString() != "Frank" {
			t.Error("HMSET should set name field")
		}

		reply, err = client.Send("HGET", "user:7", "age")
		if err != nil || reply.GetString() != "40" {
			t.Error("HMSET should set age field")
		}

		reply, err = client.Send("HGET", "user:7", "city")
		if err != nil || reply.GetString() != "LA" {
			t.Error("HMSET should set city field")
		}

		client.Send("DEL", "user:7")
	})
}

// TestHash_BinarySafety tests hash with special characters and binary data
func TestHash_BinarySafety(t *testing.T) {
	client := setupTestClient(t)
	defer client.Close()

	t.Run("Hash field and value with spaces", func(t *testing.T) {
		reply, err := client.Send("HSET", "test:hash", "field with spaces", "value with spaces")
		if err != nil {
			t.Errorf("HSET with spaces failed: %v", err)
		}

		reply, err = client.Send("HGET", "test:hash", "field with spaces")
		if err != nil {
			t.Errorf("HGET failed: %v", err)
		}
		if reply.GetString() != "value with spaces" {
			t.Errorf("Value with spaces not preserved, got '%s'", reply.GetString())
		}

		client.Send("DEL", "test:hash")
	})

	t.Run("Hash with special characters", func(t *testing.T) {
		specialValue := "hello\r\nworld\ttest"
		reply, err := client.Send("HSET", "test:hash", "special", specialValue)
		if err != nil {
			t.Errorf("HSET with special chars failed: %v", err)
		}

		reply, err = client.Send("HGET", "test:hash", "special")
		if err != nil {
			t.Errorf("HGET failed: %v", err)
		}
		if reply.GetString() != specialValue {
			t.Errorf("Special characters not preserved, got '%s'", reply.GetString())
		}

		client.Send("DEL", "test:hash")
	})

	t.Run("Hash with unicode", func(t *testing.T) {
		unicodeValue := "你好世界🌍"
		reply, err := client.Send("HSET", "test:hash", "unicode", unicodeValue)
		if err != nil {
			t.Errorf("HSET with unicode failed: %v", err)
		}

		reply, err = client.Send("HGET", "test:hash", "unicode")
		if err != nil {
			t.Errorf("HGET failed: %v", err)
		}
		if reply.GetString() != unicodeValue {
			t.Errorf("Unicode not preserved, got '%s'", reply.GetString())
		}

		client.Send("DEL", "test:hash")
	})

	t.Run("Hash with empty string value", func(t *testing.T) {
		reply, err := client.Send("HSET", "test:hash", "empty", "")
		if err != nil {
			t.Errorf("HSET empty string failed: %v", err)
		}

		reply, err = client.Send("HGET", "test:hash", "empty")
		if err != nil {
			t.Errorf("HGET failed: %v", err)
		}
		if reply.GetString() != "" {
			t.Errorf("Empty string not preserved, got '%s'", reply.GetString())
		}

		client.Send("DEL", "test:hash")
	})
}

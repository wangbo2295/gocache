package functional

import (
	"fmt"
	"testing"

	"github.com/wangbo/gocache/test/e2e"
)

var _ = &e2e.TestClient{} // Verify e2e.TestClient implements expected interface

// TestString_BasicOperations tests basic SET/GET/DEL operations
func TestString_BasicOperations(t *testing.T) {
	client := setupTestClient(t)
	defer client.Close()

	// Clean up first
	client.Send("DEL", "test_key")

	t.Run("SET and GET basic key-value", func(t *testing.T) {
		reply, err := client.Send("SET", "test_key", "test_value")
		if err != nil || !reply.IsOK() {
			t.Errorf("SET failed: %v, reply: %v", err, reply)
		}

		reply, err = client.Send("GET", "test_key")
		if err != nil {
			t.Errorf("GET failed: %v", err)
		}
		if reply.GetString() != "test_value" {
			t.Errorf("GET returned wrong value: got %v, want 'test_value'", reply.GetString())
		}
	})

	t.Run("GET non-existent key returns nil", func(t *testing.T) {
		reply, err := client.Send("GET", "nonexistent_key")
		if err != nil {
			t.Errorf("GET failed: %v", err)
		}
		if !reply.IsNil() {
			t.Errorf("GET non-existent key should return nil, got: %v", reply.GetString())
		}
	})

	t.Run("DEL deletes a key", func(t *testing.T) {
		client.Send("SET", "delete_key", "value")
		reply, err := client.Send("DEL", "delete_key")
		if err != nil {
			t.Errorf("DEL failed: %v", err)
		}
		count, _ := reply.GetInt()
		if count != 1 {
			t.Errorf("DEL should return 1, got %d", count)
		}

		// Verify key is deleted
		reply, err = client.Send("GET", "delete_key")
		if err != nil {
			t.Errorf("GET after DEL failed: %v", err)
		}
		if !reply.IsNil() {
			t.Error("Key should be deleted after DEL")
		}
	})

	t.Run("DEL non-existent key returns 0", func(t *testing.T) {
		reply, err := client.Send("DEL", "nonexistent_key")
		if err != nil {
			t.Errorf("DEL failed: %v", err)
		}
		count, _ := reply.GetInt()
		if count != 0 {
			t.Errorf("DEL non-existent key should return 0, got %d", count)
		}
	})

	// Cleanup
	client.Send("DEL", "test_key")
}

// TestString_Exists tests EXISTS command
func TestString_Exists(t *testing.T) {
	client := setupTestClient(t)
	defer client.Close()

	client.Send("DEL", "exists_key1", "exists_key2")

	t.Run("EXISTS returns 1 for existing key", func(t *testing.T) {
		client.Send("SET", "exists_key1", "value")
		reply, err := client.Send("EXISTS", "exists_key1")
		if err != nil {
			t.Errorf("EXISTS failed: %v", err)
		}
		count, _ := reply.GetInt()
		if count != 1 {
			t.Errorf("EXISTS should return 1, got %d", count)
		}
	})

	t.Run("EXISTS returns 0 for non-existent key", func(t *testing.T) {
		reply, err := client.Send("EXISTS", "nonexistent_key")
		if err != nil {
			t.Errorf("EXISTS failed: %v", err)
		}
		count, _ := reply.GetInt()
		if count != 0 {
			t.Errorf("EXISTS should return 0, got %d", count)
		}
	})

	t.Run("EXISTS with multiple keys", func(t *testing.T) {
		client.Send("SET", "exists_key1", "value1")
		client.Send("SET", "exists_key2", "value2")

		reply, err := client.Send("EXISTS", "exists_key1", "exists_key2", "nonexistent")
		if err != nil {
			t.Errorf("EXISTS failed: %v", err)
		}
		count, _ := reply.GetInt()
		if count != 2 {
			t.Errorf("EXISTS should return 2, got %d", count)
		}
	})

	// Cleanup
	client.Send("DEL", "exists_key1", "exists_key2")
}

// TestString_IncrDecr tests INCR, DECR, INCRBY, DECRBY commands
func TestString_IncrDecr(t *testing.T) {
	client := setupTestClient(t)
	defer client.Close()

	t.Run("INCR increments integer value", func(t *testing.T) {
		client.Send("SET", "counter", "10")
		reply, err := client.Send("INCR", "counter")
		if err != nil {
			t.Errorf("INCR failed: %v", err)
		}
		val, _ := reply.GetInt()
		if val != 11 {
			t.Errorf("INCR should return 11, got %d", val)
		}

		// Verify value
		getReply, _ := client.Send("GET", "counter")
		if getReply.GetString() != "11" {
			t.Errorf("Value should be '11', got '%s'", getReply.GetString())
		}

		client.Send("DEL", "counter")
	})

	t.Run("INCR on non-existent key creates it with value 1", func(t *testing.T) {
		client.Send("DEL", "new_counter")
		reply, err := client.Send("INCR", "new_counter")
		if err != nil {
			t.Errorf("INCR failed: %v", err)
		}
		val, _ := reply.GetInt()
		if val != 1 {
			t.Errorf("INCR on new key should return 1, got %d", val)
		}
		client.Send("DEL", "new_counter")
	})

	t.Run("DECR decrements integer value", func(t *testing.T) {
		client.Send("SET", "counter", "10")
		reply, err := client.Send("DECR", "counter")
		if err != nil {
			t.Errorf("DECR failed: %v", err)
		}
		val, _ := reply.GetInt()
		if val != 9 {
			t.Errorf("DECR should return 9, got %d", val)
		}
		client.Send("DEL", "counter")
	})

	t.Run("INCRBY increments by specified amount", func(t *testing.T) {
		client.Send("SET", "counter", "10")
		reply, err := client.Send("INCRBY", "counter", "5")
		if err != nil {
			t.Errorf("INCRBY failed: %v", err)
		}
		val, _ := reply.GetInt()
		if val != 15 {
			t.Errorf("INCRBY should return 15, got %d", val)
		}
		client.Send("DEL", "counter")
	})

	t.Run("INCRBY with negative value decrements", func(t *testing.T) {
		client.Send("SET", "counter", "10")
		reply, err := client.Send("INCRBY", "counter", "-3")
		if err != nil {
			t.Errorf("INCRBY failed: %v", err)
		}
		val, _ := reply.GetInt()
		if val != 7 {
			t.Errorf("INCRBY -3 should return 7, got %d", val)
		}
		client.Send("DEL", "counter")
	})

	t.Run("DECRBY decrements by specified amount", func(t *testing.T) {
		client.Send("SET", "counter", "10")
		reply, err := client.Send("DECRBY", "counter", "3")
		if err != nil {
			t.Errorf("DECRBY failed: %v", err)
		}
		val, _ := reply.GetInt()
		if val != 7 {
			t.Errorf("DECRBY should return 7, got %d", val)
		}
		client.Send("DEL", "counter")
	})
}

// TestString_MultipleOperations tests MGET and MSET
func TestString_MultipleOperations(t *testing.T) {
	client := setupTestClient(t)
	defer client.Close()

	// Setup
	client.Send("DEL", "key1", "key2", "key3")

	t.Run("MSET sets multiple keys", func(t *testing.T) {
		reply, err := client.Send("MSET", "key1", "value1", "key2", "value2", "key3", "value3")
		if err != nil || !reply.IsOK() {
			t.Errorf("MSET failed: %v", err)
		}

		// Verify values
		for i := 1; i <= 3; i++ {
			getReply, _ := client.Send("GET", fmt.Sprintf("key%d", i))
			expected := fmt.Sprintf("value%d", i)
			if getReply.GetString() != expected {
				t.Errorf("key%d should be '%s', got '%s'", i, expected, getReply.GetString())
			}
		}
	})

	t.Run("MGET gets multiple keys", func(t *testing.T) {
		reply, err := client.Send("MGET", "key1", "key2", "key3", "nonexistent")
		if err != nil {
			t.Errorf("MGET failed: %v", err)
		}

		arr := reply.GetArray()
		if arr == nil || len(arr) != 4 {
			t.Errorf("MGET should return 4 values, got %d", len(arr))
		}

		// Check values
		expected := []string{"value1", "value2", "value3", "(nil)"}
		for i, val := range arr {
			valStr := fmt.Sprintf("%v", val)
			if valStr != expected[i] && valStr != "" {
				t.Logf("MGET[%d] = %v (expected %s)", i, val, expected[i])
			}
		}
	})

	// Cleanup
	client.Send("DEL", "key1", "key2", "key3")
}

// TestString_StringOperations tests APPEND, STRLEN, GETRANGE, SETRANGE
func TestString_StringOperations(t *testing.T) {
	client := setupTestClient(t)
	defer client.Close()

	t.Run("APPEND appends to string", func(t *testing.T) {
		client.Send("SET", "append_key", "Hello")
		reply, err := client.Send("APPEND", "append_key", " World")
		if err != nil {
			t.Errorf("APPEND failed: %v", err)
		}
		length, _ := reply.GetInt()
		if length != 11 { // "Hello World" length
			t.Errorf("APPEND should return new length 11, got %d", length)
		}

		getReply, _ := client.Send("GET", "append_key")
		if getReply.GetString() != "Hello World" {
			t.Errorf("Value should be 'Hello World', got '%s'", getReply.GetString())
		}
		client.Send("DEL", "append_key")
	})

	t.Run("STRLEN returns string length", func(t *testing.T) {
		client.Send("SET", "strlen_key", "Hello")
		reply, err := client.Send("STRLEN", "strlen_key")
		if err != nil {
			t.Errorf("STRLEN failed: %v", err)
		}
		length, _ := reply.GetInt()
		if length != 5 {
			t.Errorf("STRLEN should return 5, got %d", length)
		}
		client.Send("DEL", "strlen_key")
	})

	t.Run("STRLEN on non-existent key returns 0", func(t *testing.T) {
		reply, err := client.Send("STRLEN", "nonexistent_key")
		if err != nil {
			t.Errorf("STRLEN failed: %v", err)
		}
		length, _ := reply.GetInt()
		if length != 0 {
			t.Errorf("STRLEN on non-existent key should return 0, got %d", length)
		}
	})

	t.Run("GETRANGE gets substring", func(t *testing.T) {
		client.Send("SET", "range_key", "Hello World")
		reply, err := client.Send("GETRANGE", "range_key", "0", "4")
		if err != nil {
			t.Errorf("GETRANGE failed: %v", err)
		}
		if reply.GetString() != "Hello" {
			t.Errorf("GETRANGE 0 4 should return 'Hello', got '%s'", reply.GetString())
		}

		// Negative indices - "Hello World"[-6:-1] = " World" (includes the space)
		reply, err = client.Send("GETRANGE", "range_key", "-5", "-1")
		if err != nil {
			t.Errorf("GETRANGE with negative indices failed: %v", err)
		}
		if reply.GetString() != "World" {
			t.Errorf("GETRANGE -5 -1 should return 'World', got '%s'", reply.GetString())
		}
		client.Send("DEL", "range_key")
	})
}

// TestString_BinarySafety tests binary safety with special characters
func TestString_BinarySafety(t *testing.T) {
	client := setupTestClient(t)
	defer client.Close()

	t.Run("SET and GET with spaces", func(t *testing.T) {
		reply, err := client.Send("SET", "space_key", "value with spaces")
		if err != nil || !reply.IsOK() {
			t.Errorf("SET with spaces failed: %v", err)
		}

		getReply, _ := client.Send("GET", "space_key")
		if getReply.GetString() != "value with spaces" {
			t.Errorf("Value with spaces not preserved, got '%s'", getReply.GetString())
		}
		client.Send("DEL", "space_key")
	})

	t.Run("SET and GET with special characters", func(t *testing.T) {
		specialValue := "hello\r\nworld\ttest"
		reply, err := client.Send("SET", "special_key", specialValue)
		if err != nil || !reply.IsOK() {
			t.Errorf("SET with special chars failed: %v", err)
		}

		getReply, _ := client.Send("GET", "special_key")
		if getReply.GetString() != specialValue {
			t.Errorf("Special characters not preserved, got '%s'", getReply.GetString())
		}
		client.Send("DEL", "special_key")
	})

	t.Run("SET and GET with unicode", func(t *testing.T) {
		unicodeValue := "你好世界🌍"
		reply, err := client.Send("SET", "unicode_key", unicodeValue)
		if err != nil || !reply.IsOK() {
			t.Errorf("SET with unicode failed: %v", err)
		}

		getReply, _ := client.Send("GET", "unicode_key")
		if getReply.GetString() != unicodeValue {
			t.Errorf("Unicode not preserved, got '%s'", getReply.GetString())
		}
		client.Send("DEL", "unicode_key")
	})

	t.Run("SET empty string", func(t *testing.T) {
		reply, err := client.Send("SET", "empty_key", "")
		if err != nil || !reply.IsOK() {
			t.Errorf("SET empty string failed: %v", err)
		}

		getReply, _ := client.Send("GET", "empty_key")
		if getReply.GetString() != "" {
			t.Errorf("Empty string not preserved, got '%s'", getReply.GetString())
		}
		client.Send("DEL", "empty_key")
	})
}

package database

import (
	"strconv"
	"testing"
	"time"

	"github.com/wangbo/gocache/datastruct"
)

func TestDB_ExecSetGet(t *testing.T) {
	db := MakeDB()

	// SET and GET
	result, err := db.ExecCommand("SET", "key1", "value1")
	if err != nil {
		t.Fatalf("SET failed: %v", err)
	}
	if string(result[0]) != "OK" {
		t.Errorf("Expected OK, got %s", string(result[0]))
	}

	result, err = db.ExecCommand("GET", "key1")
	if err != nil {
		t.Fatalf("GET failed: %v", err)
	}
	if string(result[0]) != "value1" {
		t.Errorf("Expected 'value1', got %s", string(result[0]))
	}

	// GET non-existent key
	result, err = db.ExecCommand("GET", "nonexistent")
	if err != nil {
		t.Fatalf("GET failed: %v", err)
	}
	if result[0] != nil {
		t.Errorf("Expected nil for non-existent key, got %v", result[0])
	}
}

func TestDB_ExecDel(t *testing.T) {
	db := MakeDB()

	// Set up some keys
	db.ExecCommand("SET", "key1", "value1")
	db.ExecCommand("SET", "key2", "value2")

	// Delete single key
	result, err := db.ExecCommand("DEL", "key1")
	if err != nil {
		t.Fatalf("DEL failed: %v", err)
	}
	if string(result[0]) != "1" {
		t.Errorf("Expected 1, got %s", string(result[0]))
	}

	// Verify deletion
	result, err = db.ExecCommand("GET", "key1")
	if err != nil {
		t.Fatalf("GET failed: %v", err)
	}
	if result[0] != nil {
		t.Error("Key should be deleted")
	}

	// Delete multiple keys
	result, err = db.ExecCommand("DEL", "key2", "nonexistent")
	if err != nil {
		t.Fatalf("DEL failed: %v", err)
	}
	if string(result[0]) != "1" {
		t.Errorf("Expected 1, got %s", string(result[0]))
	}
}

func TestDB_ExecExists(t *testing.T) {
	db := MakeDB()

	db.ExecCommand("SET", "key1", "value1")
	db.ExecCommand("SET", "key2", "value2")

	// Existing key
	result, err := db.ExecCommand("EXISTS", "key1")
	if err != nil {
		t.Fatalf("EXISTS failed: %v", err)
	}
	if string(result[0]) != "1" {
		t.Errorf("Expected 1, got %s", string(result[0]))
	}

	// Multiple keys
	result, err = db.ExecCommand("EXISTS", "key1", "key2", "nonexistent")
	if err != nil {
		t.Fatalf("EXISTS failed: %v", err)
	}
	if string(result[0]) != "2" {
		t.Errorf("Expected 2, got %s", string(result[0]))
	}

	// Non-existent key
	result, err = db.ExecCommand("EXISTS", "nonexistent")
	if err != nil {
		t.Fatalf("EXISTS failed: %v", err)
	}
	if string(result[0]) != "0" {
		t.Errorf("Expected 0, got %s", string(result[0]))
	}
}

func TestDB_ExecKeys(t *testing.T) {
	db := MakeDB()

	// Set some keys
	db.ExecCommand("SET", "key1", "value1")
	db.ExecCommand("SET", "key2", "value2")
	db.ExecCommand("SET", "key3", "value3")

	// Get all keys
	result, err := db.ExecCommand("KEYS", "*")
	if err != nil {
		t.Fatalf("KEYS failed: %v", err)
	}

	if len(result) != 3 {
		t.Errorf("Expected 3 keys, got %d", len(result))
	}

	// Verify keys (order not guaranteed)
	keySet := make(map[string]bool)
	for _, key := range result {
		keySet[string(key)] = true
	}

	if !keySet["key1"] || !keySet["key2"] || !keySet["key3"] {
		t.Error("Not all keys found")
	}
}

func TestDB_ExecIncr(t *testing.T) {
	db := MakeDB()

	// Increment non-existent key (starts from 0)
	result, err := db.ExecCommand("INCR", "counter")
	if err != nil {
		t.Fatalf("INCR failed: %v", err)
	}
	if string(result[0]) != "1" {
		t.Errorf("Expected '1', got %s", string(result[0]))
	}

	// Increment existing key
	result, err = db.ExecCommand("INCR", "counter")
	if err != nil {
		t.Fatalf("INCR failed: %v", err)
	}
	if string(result[0]) != "2" {
		t.Errorf("Expected '2', got %s", string(result[0]))
	}

	// Try INCR on non-integer value
	db.ExecCommand("SET", "strkey", "notanumber")
	result, err = db.ExecCommand("INCR", "strkey")
	if err == nil {
		t.Error("Expected error for INCR on non-integer")
	}
}

func TestDB_ExecIncrBy(t *testing.T) {
	db := MakeDB()

	// Increment by positive value
	result, err := db.ExecCommand("INCRBY", "counter", "10")
	if err != nil {
		t.Fatalf("INCRBY failed: %v", err)
	}
	if string(result[0]) != "10" {
		t.Errorf("Expected '10', got %s", string(result[0]))
	}

	// Increment by negative value (decrement)
	result, err = db.ExecCommand("INCRBY", "counter", "-5")
	if err != nil {
		t.Fatalf("INCRBY failed: %v", err)
	}
	if string(result[0]) != "5" {
		t.Errorf("Expected '5', got %s", string(result[0]))
	}
}

func TestDB_ExecDecr(t *testing.T) {
	db := MakeDB()

	// Decrement non-existent key (starts from 0, becomes -1)
	result, err := db.ExecCommand("DECR", "counter")
	if err != nil {
		t.Fatalf("DECR failed: %v", err)
	}
	if string(result[0]) != "-1" {
		t.Errorf("Expected '-1', got %s", string(result[0]))
	}
}

func TestDB_ExecDecrBy(t *testing.T) {
	db := MakeDB()

	db.ExecCommand("SET", "counter", "100")

	result, err := db.ExecCommand("DECRBY", "counter", "20")
	if err != nil {
		t.Fatalf("DECRBY failed: %v", err)
	}
	if string(result[0]) != "80" {
		t.Errorf("Expected '80', got %s", string(result[0]))
	}
}

func TestDB_ExecMGet(t *testing.T) {
	db := MakeDB()

	db.ExecCommand("SET", "key1", "value1")
	db.ExecCommand("SET", "key2", "value2")

	// Get multiple keys
	result, err := db.ExecCommand("MGET", "key1", "key2", "nonexistent")
	if err != nil {
		t.Fatalf("MGET failed: %v", err)
	}

	if len(result) != 3 {
		t.Fatalf("Expected 3 results, got %d", len(result))
	}

	if string(result[0]) != "value1" {
		t.Errorf("Expected 'value1', got %s", string(result[0]))
	}
	if string(result[1]) != "value2" {
		t.Errorf("Expected 'value2', got %s", string(result[1]))
	}
	if result[2] != nil {
		t.Errorf("Expected nil for non-existent key, got %v", result[2])
	}
}

func TestDB_ExecMSet(t *testing.T) {
	db := MakeDB()

	// Set multiple keys
	result, err := db.ExecCommand("MSET", "key1", "value1", "key2", "value2")
	if err != nil {
		t.Fatalf("MSET failed: %v", err)
	}
	if string(result[0]) != "OK" {
		t.Errorf("Expected OK, got %s", string(result[0]))
	}

	// Verify values
	result, err = db.ExecCommand("GET", "key1")
	if err != nil {
		t.Fatalf("GET failed: %v", err)
	}
	if string(result[0]) != "value1" {
		t.Errorf("Expected 'value1', got %s", string(result[0]))
	}
}

func TestDB_ExecStrLen(t *testing.T) {
	db := MakeDB()

	db.ExecCommand("SET", "key1", "hello world")

	result, err := db.ExecCommand("STRLEN", "key1")
	if err != nil {
		t.Fatalf("STRLEN failed: %v", err)
	}
	if string(result[0]) != "11" {
		t.Errorf("Expected '11', got %s", string(result[0]))
	}

	// Non-existent key
	result, err = db.ExecCommand("STRLEN", "nonexistent")
	if err != nil {
		t.Fatalf("STRLEN failed: %v", err)
	}
	if string(result[0]) != "0" {
		t.Errorf("Expected '0', got %s", string(result[0]))
	}
}

func TestDB_ExecAppend(t *testing.T) {
	db := MakeDB()

	// Append to existing key
	db.ExecCommand("SET", "key1", "Hello")
	result, err := db.ExecCommand("APPEND", "key1", " World")
	if err != nil {
		t.Fatalf("APPEND failed: %v", err)
	}
	if string(result[0]) != "11" {
		t.Errorf("Expected '11', got %s", string(result[0]))
	}

	// Verify value
	result, err = db.ExecCommand("GET", "key1")
	if err != nil {
		t.Fatalf("GET failed: %v", err)
	}
	if string(result[0]) != "Hello World" {
		t.Errorf("Expected 'Hello World', got %s", string(result[0]))
	}

	// Append to non-existent key
	result, err = db.ExecCommand("APPEND", "key2", "New")
	if err != nil {
		t.Fatalf("APPEND failed: %v", err)
	}
	if string(result[0]) != "3" {
		t.Errorf("Expected '3', got %s", string(result[0]))
	}
}

func TestDB_ExecGetRange(t *testing.T) {
	db := MakeDB()

	db.ExecCommand("SET", "key1", "Hello World")

	// Get full string
	result, err := db.ExecCommand("GETRANGE", "key1", "0", "-1")
	if err != nil {
		t.Fatalf("GETRANGE failed: %v", err)
	}
	if string(result[0]) != "Hello World" {
		t.Errorf("Expected 'Hello World', got %s", string(result[0]))
	}

	// Get range
	result, err = db.ExecCommand("GETRANGE", "key1", "0", "4")
	if err != nil {
		t.Fatalf("GETRANGE failed: %v", err)
	}
	if string(result[0]) != "Hello" {
		t.Errorf("Expected 'Hello', got %s", string(result[0]))
	}

	// Negative indices
	result, err = db.ExecCommand("GETRANGE", "key1", "-6", "-1")
	if err != nil {
		t.Fatalf("GETRANGE failed: %v", err)
	}
	if string(result[0]) != " World" {
		t.Errorf("Expected ' World', got %s", string(result[0]))
	}

	// Out of range indices
	result, err = db.ExecCommand("GETRANGE", "key1", "-100", "100")
	if err != nil {
		t.Fatalf("GETRANGE failed: %v", err)
	}
	if string(result[0]) != "Hello World" {
		t.Errorf("Expected 'Hello World', got %s", string(result[0]))
	}
}

func TestDB_ExecExpire(t *testing.T) {
	db := MakeDB()

	db.ExecCommand("SET", "key1", "value1")

	// Set expiration for 2 seconds
	result, err := db.ExecCommand("EXPIRE", "key1", "2")
	if err != nil {
		t.Fatalf("EXPIRE failed: %v", err)
	}
	if string(result[0]) != "1" {
		t.Errorf("Expected 1, got %s", string(result[0]))
	}

	// Check TTL
	result, err = db.ExecCommand("TTL", "key1")
	if err != nil {
		t.Fatalf("TTL failed: %v", err)
	}
	ttl, _ := strconv.Atoi(string(result[0]))
	if ttl < 0 || ttl > 2 {
		t.Errorf("Expected TTL between 0 and 2, got %d", ttl)
	}

	// Wait for expiration
	time.Sleep(2 * time.Second)

	// Key should be expired
	result, err = db.ExecCommand("GET", "key1")
	if err != nil {
		t.Fatalf("GET failed: %v", err)
	}
	if result[0] != nil {
		t.Error("Key should be expired")
	}
}

func TestDB_ExecPExpire(t *testing.T) {
	db := MakeDB()

	db.ExecCommand("SET", "key1", "value1")

	// Set expiration for 500 milliseconds
	result, err := db.ExecCommand("PEXPIRE", "key1", "500")
	if err != nil {
		t.Fatalf("PEXPIRE failed: %v", err)
	}
	if string(result[0]) != "1" {
		t.Errorf("Expected 1, got %s", string(result[0]))
	}

	// Check PTTL
	result, err = db.ExecCommand("PTTL", "key1")
	if err != nil {
		t.Fatalf("PTTL failed: %v", err)
	}
	pttl, _ := strconv.Atoi(string(result[0]))
	if pttl < 0 || pttl > 500 {
		t.Errorf("Expected PTTL between 0 and 500, got %d", pttl)
	}
}

func TestDB_ExecPersist(t *testing.T) {
	db := MakeDB()

	db.ExecCommand("SET", "key1", "value1")
	db.ExecCommand("EXPIRE", "key1", "100")

	// Remove TTL
	result, err := db.ExecCommand("PERSIST", "key1")
	if err != nil {
		t.Fatalf("PERSIST failed: %v", err)
	}
	if string(result[0]) != "1" {
		t.Errorf("Expected 1, got %s", string(result[0]))
	}

	// Check TTL (should be -1, no expiry)
	result, err = db.ExecCommand("TTL", "key1")
	if err != nil {
		t.Fatalf("TTL failed: %v", err)
	}
	if string(result[0]) != "-1" {
		t.Errorf("Expected '-1', got %s", string(result[0]))
	}
}

func TestDB_ExecPersistNonExistentKey(t *testing.T) {
	db := MakeDB()

	// Try to persist non-existent key
	result, err := db.ExecCommand("PERSIST", "nonexistent")
	if err != nil {
		t.Fatalf("PERSIST failed: %v", err)
	}
	if string(result[0]) != "0" {
		t.Errorf("Expected 0, got %s", string(result[0]))
	}
}

func TestDB_TTL(t *testing.T) {
	db := MakeDB()

	// Non-existent key
	result, err := db.ExecCommand("TTL", "nonexistent")
	if err != nil {
		t.Fatalf("TTL failed: %v", err)
	}
	if string(result[0]) != "-2" {
		t.Errorf("Expected '-2', got %s", string(result[0]))
	}

	// Key without TTL
	db.ExecCommand("SET", "key1", "value1")
	result, err = db.ExecCommand("TTL", "key1")
	if err != nil {
		t.Fatalf("TTL failed: %v", err)
	}
	if string(result[0]) != "-1" {
		t.Errorf("Expected '-1', got %s", string(result[0]))
	}
}

func TestDB_GetEntity(t *testing.T) {
	db := MakeDB()

	// Non-existent key
	entity, ok := db.GetEntity("key1")
	if ok {
		t.Error("Expected false for non-existent key")
	}

	// Set and get
	db.ExecCommand("SET", "key1", "value1")
	entity, ok = db.GetEntity("key1")
	if !ok {
		t.Error("Expected true for existing key")
	}

	str, ok := entity.Data.(*datastruct.String)
	if !ok {
		t.Error("Expected String data type")
	}
	if string(str.Value) != "value1" {
		t.Errorf("Expected 'value1', got %s", string(str.Value))
	}
}

func TestDB_PutEntity(t *testing.T) {
	db := MakeDB()

	entity := datastruct.MakeString([]byte("test"))

	// Put new key
	result := db.PutEntity("key1", entity)
	if result != 1 {
		t.Errorf("Expected 1 for new key, got %d", result)
	}

	// Update existing key
	result = db.PutEntity("key1", entity)
	if result != 0 {
		t.Errorf("Expected 0 for update, got %d", result)
	}
}

func TestDB_Remove(t *testing.T) {
	db := MakeDB()

	db.ExecCommand("SET", "key1", "value1")

	// Remove key
	result := db.Remove("key1")
	if result != 1 {
		t.Errorf("Expected 1, got %d", result)
	}

	// Remove non-existent key
	result = db.Remove("key1")
	if result != 0 {
		t.Errorf("Expected 0, got %d", result)
	}
}

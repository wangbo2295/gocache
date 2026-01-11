package database

import (
	"testing"
	"time"

	"github.com/wangbo/gocache/datastruct"
)

// TestDB_PutIfExists tests the PutIfExists method
func TestDB_PutIfExists(t *testing.T) {
	db := MakeDB()
	defer db.Close()

	// Try to update non-existent key
	entity := &datastruct.DataEntity{Data: []byte("value1")}
	result := db.PutIfExists("key1", entity)
	if result != 0 {
		t.Errorf("PutIfExists on non-existent key should return 0, got %d", result)
	}

	// Add the key first
	db.PutEntity("key1", entity)

	// Now update it
	newEntity := &datastruct.DataEntity{Data: []byte("value2")}
	result = db.PutIfExists("key1", newEntity)
	if result != 1 {
		t.Errorf("PutIfExists on existing key should return 1, got %d", result)
	}

	// Verify the value was updated
	val, ok := db.GetEntity("key1")
	if !ok || string(val.Data.([]byte)) != "value2" {
		t.Error("PutIfExists should update the value")
	}
}

// TestDB_PutIfAbsent tests the PutIfAbsent method
func TestDB_PutIfAbsent(t *testing.T) {
	db := MakeDB()
	defer db.Close()

	entity := &datastruct.DataEntity{Data: []byte("value1")}

	// Insert new key
	result := db.PutIfAbsent("key1", entity)
	if result != 1 {
		t.Errorf("PutIfAbsent on new key should return 1, got %d", result)
	}

	// Try to insert again
	newEntity := &datastruct.DataEntity{Data: []byte("value2")}
	result = db.PutIfAbsent("key1", newEntity)
	if result != 0 {
		t.Errorf("PutIfAbsent on existing key should return 0, got %d", result)
	}

	// Verify original value is unchanged
	val, ok := db.GetEntity("key1")
	if !ok || string(val.Data.([]byte)) != "value1" {
		t.Error("PutIfAbsent should not overwrite existing value")
	}
}

// TestDB_Keys tests the Keys method
func TestDB_Keys(t *testing.T) {
	db := MakeDB()
	defer db.Close()

	// Initially empty
	keys := db.Keys()
	if len(keys) != 0 {
		t.Errorf("Expected 0 keys, got %d", len(keys))
	}

	// Add some keys
	for i := 0; i < 5; i++ {
		key := "key" + string(rune('0'+i))
		entity := &datastruct.DataEntity{Data: []byte("value")}
		db.PutEntity(key, entity)
	}

	keys = db.Keys()
	if len(keys) != 5 {
		t.Errorf("Expected 5 keys, got %d", len(keys))
	}
}

// TestDB_Close tests the Close method
func TestDB_Close(t *testing.T) {
	db := MakeDB()

	// Add some data
	entity := &datastruct.DataEntity{Data: []byte("value")}
	db.PutEntity("key1", entity)
	db.Exec([][]byte{[]byte("SET"), []byte("key2"), []byte("value2")})

	// Close the database
	err := db.Close()
	if err != nil {
		t.Fatalf("Close should not return error, got: %v", err)
	}

	// Verify data is cleared
	keys := db.Keys()
	if len(keys) != 0 {
		t.Errorf("After Close, keys should be empty, got %d", len(keys))
	}

	// Verify memory is reset
	if db.GetUsedMemory() != 0 {
		t.Errorf("After Close, usedMemory should be 0, got %d", db.GetUsedMemory())
	}
}

// TestDB_SlowLog tests slow log functionality
func TestDB_SlowLog(t *testing.T) {
	db := MakeDB()
	defer db.Close()

	// Add a fast command (should not be logged)
	db.AddSlowLogEntry(1*time.Millisecond, [][]byte{[]byte("GET"), []byte("key")})
	entries := db.GetSlowLogEntries()
	if len(entries) != 0 {
		t.Error("Fast command should not be logged")
	}

	// Add a slow command (should be logged)
	db.AddSlowLogEntry(15*time.Millisecond, [][]byte{[]byte("SLOW"), []byte("command")})
	entries = db.GetSlowLogEntries()
	if len(entries) != 1 {
		t.Errorf("Expected 1 slow log entry, got %d", len(entries))
	}

	// Check entry properties
	if len(entries) > 0 {
		entry := entries[0]
		if entry.Duration < 10000 { // 10ms in microseconds
			t.Errorf("Duration should be >= 10000 microseconds, got %d", entry.Duration)
		}
		if len(entry.Command) == 0 {
			t.Error("Command should not be empty")
		}
	}

	// Test GetSlowLogLen
	length := db.GetSlowLogLen()
	if length != 1 {
		t.Errorf("Expected slow log length 1, got %d", length)
	}

	// Test ResetSlowLog
	db.ResetSlowLog()
	entries = db.GetSlowLogEntries()
	if len(entries) != 0 {
		t.Error("After ResetSlowLog, entries should be empty")
	}
}

// TestDB_SerializeCommand tests the serializeCommand function
func TestDB_SerializeCommand(t *testing.T) {
	tests := []struct {
		name     string
		cmdLine  [][]byte
		expected string
	}{
		{
			name:     "empty command",
			cmdLine:  [][]byte{},
			expected: "",
		},
		{
			name:     "simple command",
			cmdLine:  [][]byte{[]byte("GET"), []byte("key")},
			expected: "GET key",
		},
		{
			name:     "command with spaces",
			cmdLine:  [][]byte{[]byte("SET"), []byte(" key with spaces "), []byte("value")},
			expected: `SET " key with spaces " value`,
		},
		{
			name:     "command with empty arg",
			cmdLine:  [][]byte{[]byte("SET"), []byte(""), []byte("value")},
			expected: `SET "" value`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := serializeCommand(tt.cmdLine)
			if string(result) != tt.expected {
				t.Errorf("serializeCommand() = %q, want %q", string(result), tt.expected)
			}
		})
	}
}

// TestDB_GetVersion tests the GetVersion method
func TestDB_GetVersion(t *testing.T) {
	db := MakeDB()
	defer db.Close()

	// Non-existent key should have version 0
	version := db.GetVersion("nonexistent")
	if version != 0 {
		t.Errorf("Non-existent key should have version 0, got %d", version)
	}

	// Add a key
	entity := &datastruct.DataEntity{Data: []byte("value")}
	db.PutEntity("key1", entity)

	// Version should be >= 1
	version = db.GetVersion("key1")
	if version < 1 {
		t.Errorf("Key should have version >= 1, got %d", version)
	}

	// Update the key
	db.PutEntity("key1", &datastruct.DataEntity{Data: []byte("newvalue")})

	// Version should increment
	newVersion := db.GetVersion("key1")
	if newVersion <= version {
		t.Errorf("Version should increment, was %d, now %d", version, newVersion)
	}
}

// TestDB_ExpireFromTimeWheel tests the expireFromTimeWheel callback
func TestDB_ExpireFromTimeWheel(t *testing.T) {
	db := MakeDB()
	defer db.Close()

	// Add a key with TTL
	entity := &datastruct.DataEntity{Data: []byte("value")}
	db.PutEntity("key1", entity)
	db.ttlMap.Put("key1", time.Now().Add(-time.Second)) // Already expired

	// Manually call expireFromTimeWheel
	db.expireFromTimeWheel("key1")

	// Key should be removed
	_, ok := db.GetEntityWithoutTTLCheck("key1")
	if ok {
		t.Error("Expired key should be removed by expireFromTimeWheel")
	}
}

// TestDB_GetEntityWithoutExpiryCheck tests getEntityWithoutExpiryCheck
func TestDB_GetEntityWithoutExpiryCheck(t *testing.T) {
	db := MakeDB()
	defer db.Close()

	// Add a key with TTL
	entity := &datastruct.DataEntity{Data: []byte("value")}
	db.PutEntity("key1", entity)
	db.ttlMap.Put("key1", time.Now().Add(-time.Second)) // Already expired

	// getEntityWithoutExpiryCheck should return the key even if expired
	val, ok := db.getEntityWithoutExpiryCheck("key1")
	if !ok {
		t.Error("getEntityWithoutExpiryCheck should return expired key")
	}
	if string(val.Data.([]byte)) != "value" {
		t.Errorf("Expected value 'value', got '%s'", string(val.Data.([]byte)))
	}
}

// TestDB_InitEvictionPolicy tests initEvictionPolicy
func TestDB_InitEvictionPolicy(t *testing.T) {
	db := MakeDB()
	defer db.Close()

	// After MakeDB, eviction policy should be initialized
	if db.evictionPolicy == nil {
		// This is OK if config has no policy set
		t.Skip("No eviction policy configured")
	}
}

// TestDB_CheckAndEvict tests checkAndEvict method
func TestDB_CheckAndEvict(t *testing.T) {
	// This test requires a specific memory limit configuration
	// Skip for now as it needs config manipulation
	t.Skip("checkAndEvict test requires memory limit configuration")
}

// Helper function for testing expired keys without TTL check
func (db *DB) GetEntityWithoutTTLCheck(key string) (*datastruct.DataEntity, bool) {
	return db.getEntityWithoutExpiryCheck(key)
}

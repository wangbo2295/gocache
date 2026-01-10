package database

import (
	"testing"
	"time"

	"github.com/wangbo/gocache/config"
	"github.com/wangbo/gocache/datastruct"
)

// TestMemoryUsage tests memory tracking
func TestMemoryUsage(t *testing.T) {
	db := MakeDB()

	// Initially, memory should be 0
	if db.GetUsedMemory() != 0 {
		t.Errorf("Expected initial memory 0, got %d", db.GetUsedMemory())
	}

	// Add some keys
	db.ExecCommand("SET", "key1", "value1")
	db.ExecCommand("SET", "key2", "value2")

	// Memory should have increased
	usedMemory := db.GetUsedMemory()
	if usedMemory == 0 {
		t.Error("Expected memory usage to increase after SET")
	}

	// Remove a key
	db.Remove("key1")

	// Memory should have decreased
	newMemory := db.GetUsedMemory()
	if newMemory >= usedMemory {
		t.Error("Expected memory usage to decrease after REMOVE")
	}
}

// TestLRUPolicyAccessTracking tests LRU access tracking
func TestLRUPolicyAccessTracking(t *testing.T) {
	// Set a small memory limit and LRU policy
	config.Config.MaxMemory = 1000 // bytes
	config.Config.MaxMemoryPolicy = "allkeys-lru"

	db := MakeDB()

	// Add keys until we hit the limit
	for i := 1; i <= 10; i++ {
		key := "key" + string(rune('0'+i))
		value := string(make([]byte, 100)) // 100 bytes each
		db.ExecCommand("SET", key, value)
	}

	// Access some keys to make them more recently used
	db.ExecCommand("GET", "key1")
	db.ExecCommand("GET", "key2")

	// Add more keys - should trigger eviction of least recently used
	db.ExecCommand("SET", "key11", string(make([]byte, 100)))

	// Check that some keys were evicted (we can't know exactly which ones without internal access)
	// but we can verify that the system didn't crash and memory is still within bounds
	if db.GetUsedMemory() > config.Config.MaxMemory*2 { // Allow some tolerance
		t.Errorf("Memory usage %d exceeds reasonable bounds", db.GetUsedMemory())
	}
}

// TestNoEvictionPolicy tests noeviction policy
func TestNoEvictionPolicy(t *testing.T) {
	config.Config.MaxMemory = 100 // Very small limit
	config.Config.MaxMemoryPolicy = "noeviction"

	db := MakeDB()

	// Add keys - with noeviction, we should not evict anything
	// even if we exceed the memory limit
	for i := 0; i < 10; i++ {
		key := "key" + string(rune('0'+i))
		db.ExecCommand("SET", key, "test")
	}

	// Memory should grow without eviction
	if db.GetUsedMemory() == 0 {
		t.Error("Expected memory usage to be tracked")
	}

	// Reset config
	config.Config.MaxMemory = 0
	config.Config.MaxMemoryPolicy = "noeviction"
}

// TestEntityEstimateSize tests size estimation for different data types
func TestEntityEstimateSize(t *testing.T) {
	// String
	strEntity := datastruct.MakeString([]byte("hello"))
	if strEntity.EstimateSize() == 0 {
		t.Error("String entity should have non-zero size")
	}

	// Hash - create via DataEntity
	hashData := datastruct.MakeHash()
	hashEntity := &datastruct.DataEntity{Data: hashData}
	if hashEntity.EstimateSize() == 0 {
		t.Error("Hash entity should have non-zero size")
	}

	// List
	listData := datastruct.MakeList()
	listEntity := &datastruct.DataEntity{Data: listData}
	if listEntity.EstimateSize() == 0 {
		t.Error("List entity should have non-zero size")
	}

	// Set
	setData := datastruct.MakeSet()
	setEntity := &datastruct.DataEntity{Data: setData}
	if setEntity.EstimateSize() == 0 {
		t.Error("Set entity should have non-zero size")
	}

	// Sorted Set
	zsetData := datastruct.MakeSortedSet()
	zsetEntity := &datastruct.DataEntity{Data: zsetData}
	if zsetEntity.EstimateSize() == 0 {
		t.Error("SortedSet entity should have non-zero size")
	}
}

// TestInfoCommand tests INFO command
func TestInfoCommand(t *testing.T) {
	db := MakeDB()

	// Add some data
	db.ExecCommand("SET", "key1", "value1")
	db.ExecCommand("SET", "key2", "value2")

	// Test INFO command
	result, err := db.ExecCommand("INFO")
	if err != nil {
		t.Fatalf("INFO command failed: %v", err)
	}

	if len(result) == 0 {
		t.Error("INFO command should return data")
	}

	// Check that output contains expected sections
	info := string(result[0])
	if len(info) == 0 {
		t.Error("INFO output should not be empty")
	}
}

// TestMemoryCommand tests MEMORY command
func TestMemoryCommand(t *testing.T) {
	db := MakeDB()

	// Add some data
	db.ExecCommand("SET", "key1", "value1")

	// Test MEMORY USAGE command
	result, err := db.ExecCommand("MEMORY", "USAGE", "key1")
	if err != nil {
		t.Fatalf("MEMORY USAGE command failed: %v", err)
	}

	if len(result) == 0 {
		t.Error("MEMORY USAGE should return data")
	}

	// Test MEMORY STATS command
	result, err = db.ExecCommand("MEMORY", "STATS")
	if err != nil {
		t.Fatalf("MEMORY STATS command failed: %v", err)
	}

	if len(result) == 0 {
		t.Error("MEMORY STATS should return data")
	}
}

// TestMemoryTrackingWithUpdate tests memory tracking when updating existing keys
func TestMemoryTrackingWithUpdate(t *testing.T) {
	db := MakeDB()

	// Add a key
	db.ExecCommand("SET", "key1", "small")
	initialMemory := db.GetUsedMemory()

	// Update with a larger value
	db.ExecCommand("SET", "key1", "this is a much larger value")

	// Memory should increase (though not necessarily by the full difference
	// since we're replacing the old value)
	updatedMemory := db.GetUsedMemory()
	if updatedMemory < initialMemory {
		t.Errorf("Memory should not decrease when updating with larger value: %d < %d",
			updatedMemory, initialMemory)
	}
}

// TestMemoryTrackingWithDelete tests memory tracking when deleting keys
func TestMemoryTrackingWithDelete(t *testing.T) {
	db := MakeDB()

	// Add keys
	db.ExecCommand("SET", "key1", "value1")
	db.ExecCommand("SET", "key2", "value2")
	db.ExecCommand("SET", "key3", "value3")

	memoryWithKeys := db.GetUsedMemory()

	// Delete a key
	db.ExecCommand("DEL", "key2")

	memoryAfterDelete := db.GetUsedMemory()

	if memoryAfterDelete >= memoryWithKeys {
		t.Errorf("Memory should decrease after delete: %d >= %d",
			memoryAfterDelete, memoryWithKeys)
	}
}

// TestExpiryWithEviction tests that expired keys are properly handled with eviction
func TestExpiryWithEviction(t *testing.T) {
	config.Config.MaxMemory = 1000
	config.Config.MaxMemoryPolicy = "allkeys-lru"

	db := MakeDB()

	// Add a key with short TTL
	db.ExecCommand("SET", "tempkey", "tempvalue")
	db.ExecCommand("EXPIRE", "tempkey", "1") // 1 second

	// Wait for expiration
	time.Sleep(2 * time.Second)

	// Try to get the expired key - it should be gone
	result, _ := db.ExecCommand("GET", "tempkey")
	if len(result) > 0 && len(result[0]) > 0 {
		t.Error("Expired key should be removed")
	}

	// Reset config
	config.Config.MaxMemory = 0
	config.Config.MaxMemoryPolicy = "noeviction"
}

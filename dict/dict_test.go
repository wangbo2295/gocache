package dict

import (
	"fmt"
	"strconv"
	"sync"
	"testing"
)

func TestConcurrentDict_Get(t *testing.T) {
	dict := MakeConcurrentDict(4)

	// Get non-existent key
	_, ok := dict.Get("key1")
	if ok {
		t.Error("Expected false for non-existent key")
	}

	// Put and Get
	dict.Put("key1", "value1")
	val, ok := dict.Get("key1")
	if !ok {
		t.Error("Expected true for existing key")
	}
	if val != "value1" {
		t.Errorf("Expected 'value1', got %v", val)
	}
}

func TestConcurrentDict_Put(t *testing.T) {
	dict := MakeConcurrentDict(4)

	// Put new key
	result := dict.Put("key1", "value1")
	if result != 1 {
		t.Errorf("Expected 1 for new key, got %d", result)
	}

	// Update existing key
	result = dict.Put("key1", "value2")
	if result != 0 {
		t.Errorf("Expected 0 for update, got %d", result)
	}

	val, _ := dict.Get("key1")
	if val != "value2" {
		t.Errorf("Expected 'value2', got %v", val)
	}
}

func TestConcurrentDict_PutIfExists(t *testing.T) {
	dict := MakeConcurrentDict(4)

	// Key doesn't exist
	result := dict.PutIfExists("key1", "value1")
	if result != 0 {
		t.Errorf("Expected 0 for non-existent key, got %d", result)
	}

	// Key exists
	dict.Put("key1", "value1")
	result = dict.PutIfExists("key1", "value2")
	if result != 1 {
		t.Errorf("Expected 1 for existing key, got %d", result)
	}

	val, _ := dict.Get("key1")
	if val != "value2" {
		t.Errorf("Expected 'value2', got %v", val)
	}
}

func TestConcurrentDict_PutIfAbsent(t *testing.T) {
	dict := MakeConcurrentDict(4)

	// Key doesn't exist
	result := dict.PutIfAbsent("key1", "value1")
	if result != 1 {
		t.Errorf("Expected 1 for new key, got %d", result)
	}

	// Key already exists
	result = dict.PutIfAbsent("key1", "value2")
	if result != 0 {
		t.Errorf("Expected 0 for existing key, got %d", result)
	}

	val, _ := dict.Get("key1")
	if val != "value1" {
		t.Errorf("Expected 'value1', got %v", val)
	}
}

func TestConcurrentDict_Remove(t *testing.T) {
	dict := MakeConcurrentDict(4)

	// Remove non-existent key
	result := dict.Remove("key1")
	if result != 0 {
		t.Errorf("Expected 0 for non-existent key, got %d", result)
	}

	// Remove existing key
	dict.Put("key1", "value1")
	result = dict.Remove("key1")
	if result != 1 {
		t.Errorf("Expected 1 for existing key, got %d", result)
	}

	// Verify key is removed
	_, ok := dict.Get("key1")
	if ok {
		t.Error("Key should be removed")
	}
}

func TestConcurrentDict_Len(t *testing.T) {
	dict := MakeConcurrentDict(4)

	if dict.Len() != 0 {
		t.Errorf("Expected 0, got %d", dict.Len())
	}

	dict.Put("key1", "value1")
	if dict.Len() != 1 {
		t.Errorf("Expected 1, got %d", dict.Len())
	}

	dict.Put("key2", "value2")
	dict.Put("key3", "value3")
	if dict.Len() != 3 {
		t.Errorf("Expected 3, got %d", dict.Len())
	}

	dict.Remove("key1")
	if dict.Len() != 2 {
		t.Errorf("Expected 2, got %d", dict.Len())
	}
}

func TestConcurrentDict_ForEach(t *testing.T) {
	dict := MakeConcurrentDict(4)

	dict.Put("key1", "value1")
	dict.Put("key2", "value2")
	dict.Put("key3", "value3")

	count := 0
	dict.ForEach(func(key string, val interface{}) bool {
		count++
		return true
	})

	if count != 3 {
		t.Errorf("Expected 3 iterations, got %d", count)
	}

	// Test early termination
	count = 0
	dict.ForEach(func(key string, val interface{}) bool {
		count++
		return count < 2 // Stop after 2 iterations
	})

	if count != 2 {
		t.Errorf("Expected 2 iterations, got %d", count)
	}
}

func TestConcurrentDict_Keys(t *testing.T) {
	dict := MakeConcurrentDict(4)

	dict.Put("key1", "value1")
	dict.Put("key2", "value2")
	dict.Put("key3", "value3")

	keys := dict.Keys()
	if len(keys) != 3 {
		t.Errorf("Expected 3 keys, got %d", len(keys))
	}

	// Verify all keys are present
	keySet := make(map[string]bool)
	for _, key := range keys {
		keySet[key] = true
	}

	if !keySet["key1"] || !keySet["key2"] || !keySet["key3"] {
		t.Error("Not all keys found")
	}
}

func TestConcurrentDict_RandomKeys(t *testing.T) {
	dict := MakeConcurrentDict(4)

	// Empty dict
	keys := dict.RandomKeys(5)
	if len(keys) != 0 {
		t.Errorf("Expected 0 keys for empty dict, got %d", len(keys))
	}

	// Add some keys
	for i := 0; i < 10; i++ {
		dict.Put("key"+strconv.Itoa(i), "value"+strconv.Itoa(i))
	}

	// Request fewer keys than available
	keys = dict.RandomKeys(5)
	if len(keys) != 5 {
		t.Errorf("Expected 5 keys, got %d", len(keys))
	}

	// Request more keys than available
	keys = dict.RandomKeys(20)
	if len(keys) != 10 {
		t.Errorf("Expected 10 keys, got %d", len(keys))
	}

	// Request 0 keys
	keys = dict.RandomKeys(0)
	if len(keys) != 0 {
		t.Errorf("Expected 0 keys, got %d", len(keys))
	}
}

func TestConcurrentDict_Clear(t *testing.T) {
	dict := MakeConcurrentDict(4)

	dict.Put("key1", "value1")
	dict.Put("key2", "value2")
	dict.Put("key3", "value3")

	if dict.Len() != 3 {
		t.Errorf("Expected 3 before clear, got %d", dict.Len())
	}

	dict.Clear()

	if dict.Len() != 0 {
		t.Errorf("Expected 0 after clear, got %d", dict.Len())
	}

	_, ok := dict.Get("key1")
	if ok {
		t.Error("Key should not exist after clear")
	}
}

func TestConcurrentDict_ShardDistribution(t *testing.T) {
	dict := MakeConcurrentDict(16)

	// Add 100 keys and verify they're distributed across shards
	for i := 0; i < 100; i++ {
		dict.Put("key"+strconv.Itoa(i), "value"+strconv.Itoa(i))
	}

	if dict.Len() != 100 {
		t.Errorf("Expected 100 keys, got %d", dict.Len())
	}

	// Verify all keys can be retrieved
	for i := 0; i < 100; i++ {
		key := "key" + strconv.Itoa(i)
		val, ok := dict.Get(key)
		if !ok {
			t.Errorf("Key %s not found", key)
		}
		expected := "value" + strconv.Itoa(i)
		if val != expected {
			t.Errorf("Expected %s, got %v", expected, val)
		}
	}
}

func TestConcurrentDict_ShardCountAdjustment(t *testing.T) {
	// Test that shard count is adjusted to power of 2
	tests := []struct {
		input    int
		expected int
	}{
		{1, 1},
		{2, 2},
		{3, 4},
		{4, 4},
		{5, 8},
		{10, 16},
		{16, 16},
		{17, 32},
		{100, 128},
	}

	for _, tt := range tests {
		t.Run(strconv.Itoa(tt.input), func(t *testing.T) {
			dict := MakeConcurrentDict(tt.input)
			if dict.shardCount != tt.expected {
				t.Errorf("Expected shard count %d, got %d", tt.expected, dict.shardCount)
			}
		})
	}
}

func TestConcurrentDict_ConcurrentWrites(t *testing.T) {
	dict := MakeConcurrentDict(16)
	const numGoroutines = 100
	const keysPerGoroutine = 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < keysPerGoroutine; j++ {
				key := "goroutine" + strconv.Itoa(id) + "_key" + strconv.Itoa(j)
				value := id*1000 + j
				dict.Put(key, value)
			}
		}(i)
	}

	wg.Wait()

	expectedLen := numGoroutines * keysPerGoroutine
	if dict.Len() != expectedLen {
		t.Errorf("Expected %d keys, got %d", expectedLen, dict.Len())
	}

	// Verify all keys can be retrieved
	for i := 0; i < numGoroutines; i++ {
		for j := 0; j < keysPerGoroutine; j++ {
			key := "goroutine" + strconv.Itoa(i) + "_key" + strconv.Itoa(j)
			val, ok := dict.Get(key)
			if !ok {
				t.Errorf("Key %s not found", key)
			}
			expected := i*1000 + j
			if val != expected {
				t.Errorf("Key %s: expected %d, got %v", key, expected, val)
			}
		}
	}
}

func TestConcurrentDict_ConcurrentReads(t *testing.T) {
	dict := MakeConcurrentDict(16)

	// Populate dict
	for i := 0; i < 1000; i++ {
		dict.Put("key"+strconv.Itoa(i), i)
	}

	const numGoroutines = 100
	const readsPerGoroutine = 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	errors := make(chan error, numGoroutines*readsPerGoroutine)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < readsPerGoroutine; j++ {
				key := "key" + strconv.Itoa(j%1000)
				val, ok := dict.Get(key)
				if !ok {
					errors <- fmt.Errorf("Key %s not found", key)
					return
				}
				expected := j % 1000
				if val != expected {
					errors <- fmt.Errorf("Key %s: expected %d, got %v", key, expected, val)
					return
				}
			}
		}()
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		if err != nil {
			t.Error(err)
		}
	}
}

func TestConcurrentDict_ConcurrentMixed(t *testing.T) {
	dict := MakeConcurrentDict(16)

	const numGoroutines = 50
	const operationsPerGoroutine = 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines * 3) // writers, readers, removers

	// Writers
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < operationsPerGoroutine; j++ {
				key := "key" + strconv.Itoa(j%100)
				dict.Put(key, id*1000+j)
			}
		}(i)
	}

	// Readers
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < operationsPerGoroutine; j++ {
				key := "key" + strconv.Itoa(j%100)
				dict.Get(key)
			}
		}()
	}

	// Removers
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < operationsPerGoroutine; j++ {
				key := "key" + strconv.Itoa(j%100)
				if id%2 == 0 { // Only half of them remove
					dict.Remove(key)
				}
			}
		}(i)
	}

	wg.Wait()

	// Just verify no deadlocks or panics occurred
	// The final state is non-deterministic due to concurrent operations
	t.Logf("Final dict size: %d", dict.Len())
}

func TestConcurrentDict_Spread(t *testing.T) {
	dict := MakeConcurrentDict(16)

	// Test that different keys are distributed across shards
	shardCounts := make(map[uint32]int)
	for i := 0; i < 1000; i++ {
		key := "key" + strconv.Itoa(i)
		shard := dict.spread(key)
		shardCounts[shard]++
	}

	// Check that keys are distributed (not all in one shard)
	if len(shardCounts) < 8 {
		t.Errorf("Keys not well distributed, only %d shards used", len(shardCounts))
	}

	// Check each shard has some keys (reasonable distribution)
	for shard, count := range shardCounts {
		if count < 10 {
			t.Errorf("Shard %d has only %d keys, distribution may be poor", shard, count)
		}
	}
}

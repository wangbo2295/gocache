package eviction

import (
	"testing"
)

func TestLRU_RecordAccess(t *testing.T) {
	lru := NewLRU(3)

	// Record accesses
	lru.RecordAccess("a")
	lru.RecordAccess("b")
	lru.RecordAccess("c")

	if lru.Len() != 3 {
		t.Errorf("Expected length 3, got %d", lru.Len())
	}

	// Access 'a' again (should move it to the back)
	lru.RecordAccess("a")

	// Get LRU list
	lruList := lru.GetLRUList()
	if len(lruList) != 3 {
		t.Errorf("Expected LRU list length 3, got %d", len(lruList))
	}

	// 'b' should be least recently used (at front)
	if lruList[0] != "b" {
		t.Errorf("Expected 'b' to be least recently used, got '%s'", lruList[0])
	}

	// 'a' should be most recently used (at back)
	if lruList[2] != "a" {
		t.Errorf("Expected 'a' to be most recently used, got '%s'", lruList[2])
	}
}

func TestLRU_Capacity(t *testing.T) {
	lru := NewLRU(3)

	// Add items up to capacity
	lru.RecordAccess("a")
	lru.RecordAccess("b")
	lru.RecordAccess("c")

	if lru.Len() != 3 {
		t.Errorf("Expected length 3, got %d", lru.Len())
	}

	// Add one more (should evict 'a')
	lru.RecordAccess("d")

	if lru.Len() != 3 {
		t.Errorf("Expected length 3 after adding beyond capacity, got %d", lru.Len())
	}

	lruList := lru.GetLRUList()
	if contains(lruList, "a") {
		t.Error("Expected 'a' to be evicted")
	}
}

func TestLRU_RecordDelete(t *testing.T) {
	lru := NewLRU(10)

	lru.RecordAccess("a")
	lru.RecordAccess("b")
	lru.RecordAccess("c")

	lru.RecordDelete("b")

	if lru.Len() != 2 {
		t.Errorf("Expected length 2, got %d", lru.Len())
	}

	lruList := lru.GetLRUList()
	if contains(lruList, "b") {
		t.Error("Expected 'b' to be deleted from LRU list")
	}
}

func TestLRU_Evict(t *testing.T) {
	lru := NewLRU(10)

	lru.RecordAccess("a")
	lru.RecordAccess("b")
	lru.RecordAccess("c")
	lru.RecordAccess("d")
	lru.RecordAccess("e")

	// Evict 2 items (should evict 'a' and 'b')
	keys := lru.Evict(2)

	if len(keys) != 2 {
		t.Errorf("Expected 2 evicted keys, got %d", len(keys))
	}

	if keys[0] != "a" || keys[1] != "b" {
		t.Errorf("Expected to evict 'a' and 'b', got %v", keys)
	}

	if lru.Len() != 3 {
		t.Errorf("Expected length 3 after eviction, got %d", lru.Len())
	}
}

func TestLRU_Reset(t *testing.T) {
	lru := NewLRU(10)

	lru.RecordAccess("a")
	lru.RecordAccess("b")
	lru.RecordAccess("c")

	lru.Reset()

	if lru.Len() != 0 {
		t.Errorf("Expected length 0 after reset, got %d", lru.Len())
	}

	lruList := lru.GetLRUList()
	if len(lruList) != 0 {
		t.Errorf("Expected empty LRU list after reset, got %v", lruList)
	}
}

func TestLRU_UpdateIsAccess(t *testing.T) {
	lru := NewLRU(3)

	lru.RecordAccess("a")
	lru.RecordUpdate("b")

	lru.RecordAccess("a") // Move 'a' to back

	lruList := lru.GetLRUList()
	if lruList[0] != "b" {
		t.Errorf("Expected 'b' to be least recently used, got '%s'", lruList[0])
	}
}

func TestLRU_DeleteNonExistent(t *testing.T) {
	lru := NewLRU(10)

	lru.RecordAccess("a")
	lru.RecordDelete("x") // Delete non-existent key

	if lru.Len() != 1 {
		t.Errorf("Expected length 1, got %d", lru.Len())
	}
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

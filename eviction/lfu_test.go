package eviction

import (
	"testing"
)

func TestLFU_RecordAccess(t *testing.T) {
	lfu := NewLFU(10)

	// Record accesses
	lfu.RecordAccess("a")
	lfu.RecordAccess("a")
	lfu.RecordAccess("b")

	freqs := lfu.GetFrequencies()

	if freqs["a"] != 2 {
		t.Errorf("Expected frequency 2 for 'a', got %d", freqs["a"])
	}

	if freqs["b"] != 1 {
		t.Errorf("Expected frequency 1 for 'b', got %d", freqs["b"])
	}
}

func TestLFU_Capacity(t *testing.T) {
	lfu := NewLFU(3)

	// Add items up to capacity
	lfu.RecordAccess("a")
	lfu.RecordAccess("b")
	lfu.RecordAccess("c")

	if lfu.Len() != 3 {
		t.Errorf("Expected length 3, got %d", lfu.Len())
	}

	// Add one more (should evict the one with lowest frequency)
	lfu.RecordAccess("d")

	if lfu.Len() != 3 {
		t.Errorf("Expected length 3 after adding beyond capacity, got %d", lfu.Len())
	}

	// All items should have frequency 1, so oldest should be evicted
	freqs := lfu.GetFrequencies()
	if _, exists := freqs["a"]; exists {
		t.Error("Expected 'a' to be evicted (oldest with same frequency)")
	}
}

func TestLFU_EvictionOrder(t *testing.T) {
	lfu := NewLFU(10)

	// Create different access frequencies
	lfu.RecordAccess("a")
	lfu.RecordAccess("a")
	lfu.RecordAccess("a") // freq = 3

	lfu.RecordAccess("b")
	lfu.RecordAccess("b") // freq = 2

	lfu.RecordAccess("c") // freq = 1

	// Evict 2 items (should evict 'c' then 'b')
	keys := lfu.Evict(2)

	if len(keys) != 2 {
		t.Errorf("Expected 2 evicted keys, got %d", len(keys))
	}

	if keys[0] != "c" || keys[1] != "b" {
		t.Errorf("Expected to evict 'c' and 'b', got %v", keys)
	}
}

func TestLFU_RecordDelete(t *testing.T) {
	lfu := NewLFU(10)

	lfu.RecordAccess("a")
	lfu.RecordAccess("b")
	lfu.RecordAccess("c")

	lfu.RecordDelete("b")

	if lfu.Len() != 2 {
		t.Errorf("Expected length 2, got %d", lfu.Len())
	}

	freqs := lfu.GetFrequencies()
	if _, exists := freqs["b"]; exists {
		t.Error("Expected 'b' to be deleted from LFU")
	}
}

func TestLFU_Evict(t *testing.T) {
	lfu := NewLFU(10)

	lfu.RecordAccess("a")
	lfu.RecordAccess("b")
	lfu.RecordAccess("c")
	lfu.RecordAccess("d")
	lfu.RecordAccess("e")

	// Evict 2 items
	keys := lfu.Evict(2)

	if len(keys) != 2 {
		t.Errorf("Expected 2 evicted keys, got %d", len(keys))
	}

	// All have frequency 1, so oldest should be evicted first
	if keys[0] != "a" || keys[1] != "b" {
		t.Errorf("Expected to evict 'a' and 'b', got %v", keys)
	}

	if lfu.Len() != 3 {
		t.Errorf("Expected length 3 after eviction, got %d", lfu.Len())
	}
}

func TestLFU_Reset(t *testing.T) {
	lfu := NewLFU(10)

	lfu.RecordAccess("a")
	lfu.RecordAccess("b")
	lfu.RecordAccess("c")

	lfu.Reset()

	if lfu.Len() != 0 {
		t.Errorf("Expected length 0 after reset, got %d", lfu.Len())
	}

	freqs := lfu.GetFrequencies()
	if len(freqs) != 0 {
		t.Errorf("Expected empty frequencies after reset, got %v", freqs)
	}
}

func TestLFU_UpdateIsAccess(t *testing.T) {
	lfu := NewLFU(10)

	lfu.RecordUpdate("a")
	lfu.RecordUpdate("a")
	lfu.RecordUpdate("b")

	freq := lfu.GetFrequency("a")
	if freq != 2 {
		t.Errorf("Expected frequency 2 for 'a', got %d", freq)
	}
}

func TestLFU_DeleteNonExistent(t *testing.T) {
	lfu := NewLFU(10)

	lfu.RecordAccess("a")
	lfu.RecordDelete("x") // Delete non-existent key

	if lfu.Len() != 1 {
		t.Errorf("Expected length 1, got %d", lfu.Len())
	}
}

func TestLFU_EvictBeyondCapacity(t *testing.T) {
	lfu := NewLFU(10)

	lfu.RecordAccess("a")
	lfu.RecordAccess("b")

	// Evict more than available
	keys := lfu.Evict(5)

	if len(keys) != 2 {
		t.Errorf("Expected 2 evicted keys, got %d", len(keys))
	}

	if lfu.Len() != 0 {
		t.Errorf("Expected length 0, got %d", lfu.Len())
	}
}

func TestLFU_GetFrequency(t *testing.T) {
	lfu := NewLFU(10)

	lfu.RecordAccess("a")
	lfu.RecordAccess("a")
	lfu.RecordAccess("a")

	freq := lfu.GetFrequency("a")
	if freq != 3 {
		t.Errorf("Expected frequency 3, got %d", freq)
	}

	// Non-existent key
	freq = lfu.GetFrequency("x")
	if freq != 0 {
		t.Errorf("Expected frequency 0 for non-existent key, got %d", freq)
	}
}

func TestLFU_FrequencyBasedEviction(t *testing.T) {
	lfu := NewLFU(5)

	// Create items with different frequencies
	lfu.RecordAccess("a") // freq 1
	lfu.RecordAccess("b") // freq 1
	lfu.RecordAccess("c") // freq 1

	lfu.RecordAccess("a") // freq 2
	lfu.RecordAccess("b") // freq 2

	lfu.RecordAccess("a") // freq 3

	freqs := lfu.GetFrequencies()
	if freqs["a"] != 3 || freqs["b"] != 2 || freqs["c"] != 1 {
		t.Error("Frequencies not as expected")
	}

	// Evict should remove 'c' (lowest frequency, oldest among ties)
	keys := lfu.Evict(1)
	if len(keys) != 1 || keys[0] != "c" {
		t.Errorf("Expected to evict 'c', got %v", keys)
	}
}

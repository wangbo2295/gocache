package eviction

import (
	"testing"
	"time"
)

func TestNewRandom(t *testing.T) {
	rand := NewRandom()
	if rand == nil {
		t.Fatal("NewRandom returned nil")
	}
}

func TestRandom_RecordAccess(t *testing.T) {
	rand := NewRandom()
	rand.RecordAccess("key1")
	// Should not panic
}

func TestRandom_RecordUpdate(t *testing.T) {
	rand := NewRandom()
	rand.RecordUpdate("key1")
	// Should not panic
}

func TestRandom_RecordDelete(t *testing.T) {
	rand := NewRandom()
	rand.RecordDelete("key1")
	// Should not panic
}

func TestRandom_Evict(t *testing.T) {
	rand := NewRandom()
	
	// Add some keys
	for i := 0; i < 10; i++ {
		rand.RecordUpdate(string(rune('a'+i)))
	}
	
	keys := rand.Evict(3)
	if len(keys) != 3 {
		t.Errorf("Expected 3 keys, got %d", len(keys))
	}
}

func TestRandom_Reset(t *testing.T) {
	rand := NewRandom()
	rand.RecordUpdate("key1")
	rand.Reset()
	// Should not panic
}

func TestRandom_Len(t *testing.T) {
	rand := NewRandom()
	
	// Add keys
	for i := 0; i < 5; i++ {
		rand.RecordUpdate(string(rune('a'+i)))
	}
	
	if rand.Len() != 5 {
		t.Errorf("Expected length 5, got %d", rand.Len())
	}
}

func TestNewTTL(t *testing.T) {
	ttl := NewTTL()
	if ttl == nil {
		t.Fatal("NewTTL returned nil")
	}
}

func TestTTL_RecordAccess(t *testing.T) {
	ttl := NewTTL()
	ttl.RecordAccess("key1")
	// Should not panic
}

func TestTTL_RecordUpdate(t *testing.T) {
	ttl := NewTTL()
	ttl.RecordUpdate("key1")
	// Should not panic
}

func TestTTL_RecordDelete(t *testing.T) {
	ttl := NewTTL()
	ttl.RecordDelete("key1")
	// Should not panic
}

func TestTTL_SetExpire(t *testing.T) {
	ttl := NewTTL()
	expireTime := time.Now().Add(1 * time.Second)
	ttl.SetExpire("key1", expireTime)
	// Should not panic
}

func TestTTL_Evict(t *testing.T) {
	ttl := NewTTL()
	baseTime := time.Now()
	
	// Add keys with expirations
	for i := 0; i < 5; i++ {
		expireTime := baseTime.Add(time.Duration(1000+i*100) * time.Millisecond)
		ttl.SetExpire(string(rune('a'+i)), expireTime)
	}
	
	keys := ttl.Evict(2)
	if len(keys) != 2 {
		t.Errorf("Expected 2 keys, got %d", len(keys))
	}
}

func TestTTL_Reset(t *testing.T) {
	ttl := NewTTL()
	expireTime := time.Now().Add(1 * time.Second)
	ttl.SetExpire("key1", expireTime)
	ttl.Reset()
	// Should not panic
}

func TestTTL_Len(t *testing.T) {
	ttl := NewTTL()
	baseTime := time.Now()
	
	for i := 0; i < 3; i++ {
		expireTime := baseTime.Add(time.Duration(1000+i*100) * time.Millisecond)
		ttl.SetExpire(string(rune('a'+i)), expireTime)
	}
	
	if ttl.Len() != 3 {
		t.Errorf("Expected length 3, got %d", ttl.Len())
	}
}

func TestTTLHeapOperations(t *testing.T) {
	ttl := NewTTL()
	baseTime := time.Now()
	
	// Test heap operations
	ttl.SetExpire("key1", baseTime.Add(3*time.Second))
	ttl.SetExpire("key2", baseTime.Add(1*time.Second))
	ttl.SetExpire("key3", baseTime.Add(2*time.Second))
	
	// Evict should return shortest TTL first
	keys := ttl.Evict(1)
	if len(keys) != 1 || keys[0] != "key2" {
		t.Errorf("Expected key2 to be evicted first, got %v", keys)
	}
}

package datastruct

import (
	"sync"
	"testing"
	"time"
)

func TestNewTimeWheel(t *testing.T) {
	expiredKeys := make([]string, 0)
	var mu sync.Mutex

	onExpire := func(key string) {
		mu.Lock()
		defer mu.Unlock()
		expiredKeys = append(expiredKeys, key)
	}

	tw := NewTimeWheel(10*time.Millisecond, 100, onExpire)

	if tw == nil {
		t.Fatal("NewTimeWheel returned nil")
	}

	if tw.wheelSize != 100 {
		t.Errorf("Expected wheelSize 100, got %d", tw.wheelSize)
	}

	if tw.interval != 10*time.Millisecond {
		t.Errorf("Expected interval 10ms, got %v", tw.interval)
	}
}

func TestTimeWheel_AddAndExpire(t *testing.T) {
	expiredKeys := make([]string, 0)
	var mu sync.Mutex

	onExpire := func(key string) {
		mu.Lock()
		defer mu.Unlock()
		expiredKeys = append(expiredKeys, key)
	}

	tw := NewTimeWheel(10*time.Millisecond, 100, onExpire)
	tw.Start()
	defer tw.Stop()

	// Add a key to expire in 50ms
	tw.Add("key1", 50*time.Millisecond)

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	mu.Lock()
	keys := expiredKeys
	mu.Unlock()

	if len(keys) == 0 {
		t.Error("Expected key to expire, but no keys expired")
	}

	found := false
	for _, k := range keys {
		if k == "key1" {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("Expected key1 to expire, got %v", keys)
	}
}

func TestTimeWheel_AddMultipleKeys(t *testing.T) {
	expiredKeys := make([]string, 0)
	var mu sync.Mutex

	onExpire := func(key string) {
		mu.Lock()
		defer mu.Unlock()
		expiredKeys = append(expiredKeys, key)
	}

	tw := NewTimeWheel(10*time.Millisecond, 100, onExpire)
	tw.Start()
	defer tw.Stop()

	// Add multiple keys with different expiration times
	tw.Add("key1", 30*time.Millisecond)  // Will expire at ~30ms
	tw.Add("key2", 50*time.Millisecond)  // Will expire at ~50ms
	tw.Add("key3", 70*time.Millisecond)  // Will expire at ~70ms

	// Wait for all to expire
	time.Sleep(150 * time.Millisecond)

	mu.Lock()
	keys := expiredKeys
	mu.Unlock()

	if len(keys) < 3 {
		t.Errorf("Expected at least 3 keys to expire, got %d", len(keys))
	}

	// Check that all keys are present
	keyMap := make(map[string]bool)
	for _, k := range keys {
		keyMap[k] = true
	}

	if !keyMap["key1"] || !keyMap["key2"] || !keyMap["key3"] {
		t.Errorf("Missing expected keys, got %v", keys)
	}
}

func TestTimeWheel_Remove(t *testing.T) {
	expiredKeys := make([]string, 0)
	var mu sync.Mutex

	onExpire := func(key string) {
		mu.Lock()
		defer mu.Unlock()
		expiredKeys = append(expiredKeys, key)
	}

	tw := NewTimeWheel(10*time.Millisecond, 100, onExpire)
	tw.Start()
	defer tw.Stop()

	// Add a key
	tw.Add("key1", 50*time.Millisecond)

	// Remove it immediately
	tw.Remove("key1")

	// Wait for expiration time
	time.Sleep(150 * time.Millisecond)

	mu.Lock()
	keys := expiredKeys
	mu.Unlock()

	// Key should not have expired since we removed it
	for _, k := range keys {
		if k == "key1" {
			t.Error("key1 should not have expired after being removed")
		}
	}
}

func TestTimeWheel_Size(t *testing.T) {
	onExpire := func(key string) {}

	tw := NewTimeWheel(10*time.Millisecond, 100, onExpire)
	tw.Start()
	defer tw.Stop()

	// Initially empty
	if size := tw.Size(); size != 0 {
		t.Errorf("Expected size 0, got %d", size)
	}

	// Add some keys
	tw.Add("key1", 100*time.Millisecond)
	tw.Add("key2", 100*time.Millisecond)
	tw.Add("key3", 100*time.Millisecond)

	if size := tw.Size(); size != 3 {
		t.Errorf("Expected size 3, got %d", size)
	}

	// Remove a key
	tw.Remove("key1")

	if size := tw.Size(); size != 2 {
		t.Errorf("Expected size 2 after removal, got %d", size)
	}
}

func TestTimeWheel_StopAndStart(t *testing.T) {
	expiredKeys := make([]string, 0)
	var mu sync.Mutex

	onExpire := func(key string) {
		mu.Lock()
		defer mu.Unlock()
		expiredKeys = append(expiredKeys, key)
	}

	tw := NewTimeWheel(10*time.Millisecond, 100, onExpire)
	tw.Start()

	// Add a key
	tw.Add("key1", 50*time.Millisecond)

	// Stop before it expires
	tw.Stop()

	// Wait
	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	keys := expiredKeys
	mu.Unlock()

	// Key should not have expired since we stopped
	if len(keys) > 0 {
		t.Error("Keys should not expire after time wheel is stopped")
	}

	// Restart
	tw.Start()
	defer tw.Stop()

	// Add another key
	tw.Add("key2", 50*time.Millisecond)

	// Wait for expiration
	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	keys = expiredKeys
	mu.Unlock()

	// Now key2 should expire
	found := false
	for _, k := range keys {
		if k == "key2" {
			found = true
			break
		}
	}

	if !found {
		t.Error("key2 should have expired after restart")
	}
}

func TestTimeWheel_ConcurrentAccess(t *testing.T) {
	expiredKeys := make([]string, 0)
	var mu sync.Mutex

	onExpire := func(key string) {
		mu.Lock()
		defer mu.Unlock()
		expiredKeys = append(expiredKeys, key)
	}

	tw := NewTimeWheel(1*time.Millisecond, 100, onExpire)
	tw.Start()
	defer tw.Stop()

	// Concurrent adds
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			tw.Add(string(rune('a'+idx)), 10*time.Millisecond)
		}(i)
	}

	wg.Wait()

	// Wait for expirations
	time.Sleep(50 * time.Millisecond)

	mu.Lock()
	keys := expiredKeys
	mu.Unlock()

	// At least some keys should have expired
	if len(keys) == 0 {
		t.Error("Expected some keys to expire")
	}
}

func TestTimeWheel_ZeroDuration(t *testing.T) {
	onExpire := func(key string) {
		t.Error("Should not call expire for zero duration")
	}

	tw := NewTimeWheel(10*time.Millisecond, 100, onExpire)
	tw.Start()
	defer tw.Stop()

	// Add with zero duration - should be ignored
	tw.Add("key1", 0*time.Millisecond)

	// Wait
	time.Sleep(50 * time.Millisecond)

	// No expiration should have occurred (callback was not called)
}

func TestTimeWheel_VeryShortDuration(t *testing.T) {
	expiredKeys := make([]string, 0)
	var mu sync.Mutex

	onExpire := func(key string) {
		mu.Lock()
		defer mu.Unlock()
		expiredKeys = append(expiredKeys, key)
	}

	tw := NewTimeWheel(10*time.Millisecond, 100, onExpire)
	tw.Start()
	defer tw.Stop()

	// Add with duration shorter than interval - should still work
	tw.Add("key1", 1*time.Millisecond)

	// Wait for expiration
	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	keys := expiredKeys
	mu.Unlock()

	if len(keys) == 0 {
		t.Error("Expected key to expire even with very short duration")
	}
}

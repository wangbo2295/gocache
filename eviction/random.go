package eviction

import (
	"math/rand"
	"sync"
	"time"

	"github.com/wangbo/gocache/evictionpkg"
)

// Random implements random eviction policy
type Random struct {
	mu    sync.Mutex
	keys  map[string]bool
	rand  *rand.Rand
}

// NewRandom creates a new Random eviction policy
func NewRandom() *Random {
	return &Random{
		keys: make(map[string]bool),
		rand: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// RecordAccess records that a key was accessed
func (r *Random) RecordAccess(key string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.keys[key] = true
}

// RecordUpdate records that a key was updated
func (r *Random) RecordUpdate(key string) {
	r.RecordAccess(key)
}

// RecordDelete records that a key was deleted
func (r *Random) RecordDelete(key string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.keys, key)
}

// Evict returns random keys to evict
func (r *Random) Evict(count int) []string {
	r.mu.Lock()
	defer r.mu.Unlock()

	if len(r.keys) == 0 {
		return nil
	}

	keys := make([]string, 0, count)
	// Collect all keys
	allKeys := make([]string, 0, len(r.keys))
	for key := range r.keys {
		allKeys = append(allKeys, key)
	}

	// Randomly select keys
	for i := 0; i < count && len(allKeys) > 0; i++ {
		idx := r.rand.Intn(len(allKeys))
		key := allKeys[idx]
		keys = append(keys, key)

		// Remove selected key
		delete(r.keys, key)
		allKeys = append(allKeys[:idx], allKeys[idx+1:]...)
	}

	return keys
}

// Reset clears all tracking data
func (r *Random) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.keys = make(map[string]bool)
}

// Len returns the number of keys being tracked
func (r *Random) Len() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.keys)
}

// Ensure Random implements the EvictionPolicy interface
var _ evictionpkg.EvictionPolicy = (*Random)(nil)

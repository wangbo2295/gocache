package eviction

import (
	"container/list"
	"sync"

	"github.com/wangbo/gocache/evictionpkg"
)

// LRU implements Least Recently Used eviction policy
type LRU struct {
	mu       sync.Mutex
	lruList  *list.List              // Doubly linked list of keys (least recent at front)
	lruMap   map[string]*list.Element // Map from key to its element in the list
	capacity int                      // Maximum number of keys to track
}

// NewLRU creates a new LRU eviction policy
func NewLRU(capacity int) *LRU {
	return &LRU{
		lruList:  list.New(),
		lruMap:   make(map[string]*list.Element),
		capacity: capacity,
	}
}

// RecordAccess records that a key was accessed (moves it to the back)
func (l *LRU) RecordAccess(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	// If key exists, move it to the back (most recently used)
	if elem, exists := l.lruMap[key]; exists {
		l.lruList.MoveToBack(elem)
		return
	}

	// Add new key at the back
	elem := l.lruList.PushBack(key)
	l.lruMap[key] = elem

	// Evict if over capacity
	if l.lruList.Len() > l.capacity && l.capacity > 0 {
		l.evictOldest()
	}
}

// RecordUpdate records that a key was updated (same as access)
func (l *LRU) RecordUpdate(key string) {
	l.RecordAccess(key)
}

// RecordDelete records that a key was deleted
func (l *LRU) RecordDelete(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if elem, exists := l.lruMap[key]; exists {
		l.lruList.Remove(elem)
		delete(l.lruMap, key)
	}
}

// Evict returns keys to evict (least recently used first)
func (l *LRU) Evict(count int) []string {
	l.mu.Lock()
	defer l.mu.Unlock()

	keys := make([]string, 0, count)
	for i := 0; i < count && l.lruList.Len() > 0; i++ {
		keys = append(keys, l.evictOldest())
	}
	return keys
}

// evictOldest removes and returns the least recently used key
// Must be called with mu locked
func (l *LRU) evictOldest() string {
	elem := l.lruList.Front()
	if elem == nil {
		return ""
	}

	key := elem.Value.(string)
	l.lruList.Remove(elem)
	delete(l.lruMap, key)

	return key
}

// Reset clears all tracking data
func (l *LRU) Reset() {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.lruList.Init()
	l.lruMap = make(map[string]*list.Element)
}

// Len returns the number of keys being tracked
func (l *LRU) Len() int {
	l.mu.Lock()
	defer l.mu.Unlock()

	return l.lruList.Len()
}

// GetLRUList returns a copy of the LRU list (for debugging/testing)
func (l *LRU) GetLRUList() []string {
	l.mu.Lock()
	defer l.mu.Unlock()

	keys := make([]string, 0, l.lruList.Len())
	for elem := l.lruList.Front(); elem != nil; elem = elem.Next() {
		keys = append(keys, elem.Value.(string))
	}
	return keys
}

// Ensure LRU implements the EvictionPolicy interface
var _ evictionpkg.EvictionPolicy = (*LRU)(nil)

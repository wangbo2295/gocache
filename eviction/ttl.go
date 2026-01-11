package eviction

import (
	"container/heap"
	"sync"
	"time"

	"github.com/wangbo/gocache/evictionpkg"
)

// ttlItem represents an item with TTL in the heap
type ttlItem struct {
	key          string
	expireTime   time.Time
	index        int
}

// ttlHeap implements a min-heap based on expiration time
type ttlHeap []*ttlItem

func (h ttlHeap) Len() int { return len(h) }

func (h ttlHeap) Less(i, j int) bool {
	return h[i].expireTime.Before(h[j].expireTime)
}

func (h ttlHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
	h[i].index = i
	h[j].index = j
}

func (h *ttlHeap) Push(x interface{}) {
	n := len(*h)
	item := x.(*ttlItem)
	item.index = n
	*h = append(*h, item)
}

func (h *ttlHeap) Pop() interface{} {
	old := *h
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.index = -1
	*h = old[0 : n-1]
	return item
}

// TTL implements TTL-based eviction policy (evict keys with earliest expiration)
type TTL struct {
	mu     sync.Mutex
	heap   *ttlHeap
	items  map[string]*ttlItem
}

// NewTTL creates a new TTL eviction policy
func NewTTL() *TTL {
	h := make(ttlHeap, 0)
	heap.Init(&h)

	return &TTL{
		heap:  &h,
		items: make(map[string]*ttlItem),
	}
}

// RecordAccess is a no-op for TTL policy
func (t *TTL) RecordAccess(key string) {
	// TTL policy doesn't track access
}

// RecordUpdate is a no-op for TTL policy
func (t *TTL) RecordUpdate(key string) {
	// TTL policy doesn't track updates
}

// RecordDelete records that a key was deleted
func (t *TTL) RecordDelete(key string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if item, exists := t.items[key]; exists {
		heap.Remove(t.heap, item.index)
		delete(t.items, key)
	}
}

// SetExpire sets the expiration time for a key
func (t *TTL) SetExpire(key string, expireTime time.Time) {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Remove existing item if present
	if item, exists := t.items[key]; exists {
		heap.Remove(t.heap, item.index)
	}

	// Add new item
	item := &ttlItem{
		key:        key,
		expireTime: expireTime,
	}
	heap.Push(t.heap, item)
	t.items[key] = item
}

// Evict returns keys with earliest expiration times
func (t *TTL) Evict(count int) []string {
	t.mu.Lock()
	defer t.mu.Unlock()

	keys := make([]string, 0, count)
	for i := 0; i < count && t.heap.Len() > 0; i++ {
		item := heap.Pop(t.heap).(*ttlItem)
		keys = append(keys, item.key)
		delete(t.items, item.key)
	}
	return keys
}

// Reset clears all tracking data
func (t *TTL) Reset() {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.heap = &ttlHeap{}
	heap.Init(t.heap)
	t.items = make(map[string]*ttlItem)
}

// Len returns the number of keys being tracked
func (t *TTL) Len() int {
	t.mu.Lock()
	defer t.mu.Unlock()

	return t.heap.Len()
}

// Ensure TTL implements the EvictionPolicy interface
var _ evictionpkg.EvictionPolicy = (*TTL)(nil)

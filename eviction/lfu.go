package eviction

import (
	"container/heap"
	"sync"
	"time"

	"github.com/wangbo/gocache/evictionpkg"
)

// lfuItem represents an item in the LFU heap
type lfuItem struct {
	key       string
	frequency int
	lastAccess time.Time
	index     int // Index in the heap
}

// lfuHeap implements a min-heap based on frequency (and last access as tiebreaker)
type lfuHeap []*lfuItem

func (h lfuHeap) Len() int { return len(h) }

func (h lfuHeap) Less(i, j int) bool {
	// Lower frequency comes first
	if h[i].frequency != h[j].frequency {
		return h[i].frequency < h[j].frequency
	}
	// If frequencies are equal, older access comes first
	return h[i].lastAccess.Before(h[j].lastAccess)
}

func (h lfuHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
	h[i].index = i
	h[j].index = j
}

func (h *lfuHeap) Push(x interface{}) {
	n := len(*h)
	item := x.(*lfuItem)
	item.index = n
	*h = append(*h, item)
}

func (h *lfuHeap) Pop() interface{} {
	old := *h
	n := len(old)
	item := old[n-1]
	old[n-1] = nil  // avoid memory leak
	item.index = -1 // for safety
	*h = old[0 : n-1]
	return item
}

// LFU implements Least Frequently Used eviction policy
type LFU struct {
	mu       sync.Mutex
	heap     *lfuHeap                // Min-heap of items by frequency
	items    map[string]*lfuItem     // Map from key to its item in the heap
	capacity int                      // Maximum number of keys to track
}

// NewLFU creates a new LFU eviction policy
func NewLFU(capacity int) *LFU {
	h := make(lfuHeap, 0)
	heap.Init(&h)

	return &LFU{
		heap:     &h,
		items:    make(map[string]*lfuItem),
		capacity: capacity,
	}
}

// RecordAccess records that a key was accessed (increments frequency)
func (l *LFU) RecordAccess(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if item, exists := l.items[key]; exists {
		// Increment frequency and update last access time
		item.frequency++
		item.lastAccess = time.Now()
		heap.Fix(l.heap, item.index)
		return
	}

	// Add new item
	item := &lfuItem{
		key:        key,
		frequency:  1,
		lastAccess: time.Now(),
	}
	heap.Push(l.heap, item)
	l.items[key] = item

	// Evict if over capacity
	if l.heap.Len() > l.capacity && l.capacity > 0 {
		l.evictLeastFrequent()
	}
}

// RecordUpdate records that a key was updated (same as access)
func (l *LFU) RecordUpdate(key string) {
	l.RecordAccess(key)
}

// RecordDelete records that a key was deleted
func (l *LFU) RecordDelete(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if item, exists := l.items[key]; exists {
		heap.Remove(l.heap, item.index)
		delete(l.items, key)
	}
}

// Evict returns keys to evict (least frequently used first)
func (l *LFU) Evict(count int) []string {
	l.mu.Lock()
	defer l.mu.Unlock()

	keys := make([]string, 0, count)
	for i := 0; i < count && l.heap.Len() > 0; i++ {
		keys = append(keys, l.evictLeastFrequent())
	}
	return keys
}

// evictLeastFrequent removes and returns the least frequently used key
// Must be called with mu locked
func (l *LFU) evictLeastFrequent() string {
	if l.heap.Len() == 0 {
		return ""
	}

	item := heap.Pop(l.heap).(*lfuItem)
	delete(l.items, item.key)
	return item.key
}

// Reset clears all tracking data
func (l *LFU) Reset() {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.heap = &lfuHeap{}
	heap.Init(l.heap)
	l.items = make(map[string]*lfuItem)
}

// Len returns the number of keys being tracked
func (l *LFU) Len() int {
	l.mu.Lock()
	defer l.mu.Unlock()

	return l.heap.Len()
}

// GetFrequency returns the access frequency of a key (for debugging/testing)
func (l *LFU) GetFrequency(key string) int {
	l.mu.Lock()
	defer l.mu.Unlock()

	if item, exists := l.items[key]; exists {
		return item.frequency
	}
	return 0
}

// GetFrequencies returns all items with their frequencies (for debugging/testing)
func (l *LFU) GetFrequencies() map[string]int {
	l.mu.Lock()
	defer l.mu.Unlock()

	result := make(map[string]int, len(l.items))
	for key, item := range l.items {
		result[key] = item.frequency
	}
	return result
}

// Ensure LFU implements the EvictionPolicy interface
var _ evictionpkg.EvictionPolicy = (*LFU)(nil)

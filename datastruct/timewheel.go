package datastruct

import (
	"sync"
	"sync/atomic"
	"time"
)

// TimeWheel implements a hierarchical time wheel for efficient TTL management
// Based on the paper "Hashed and Hierarchical Timing Wheels"
type TimeWheel struct {
	sync.Mutex
	interval    time.Duration        // Tick interval (e.g., 1ms)
	ticker      *time.Ticker         // Time ticker
	currentTime int64                // Current time in ticks
	buckets     []*bucket            // Timing buckets
	wheelSize   int                  // Number of buckets per wheel
	stopChan    chan struct{}        // Channel to stop the time wheel
	onExpire    func(key string)     // Callback when a key expires
	running     atomic.Int32         // 1 if running, 0 if stopped
	wg          sync.WaitGroup       // Wait for goroutine to stop
}

// bucket represents a single bucket in the time wheel
type bucket struct {
	sync.Mutex
	entries map[string]struct{} // Set of keys in this bucket
}

// newBucket creates a new bucket
func newBucket() *bucket {
	return &bucket{
		entries: make(map[string]struct{}),
	}
}

// add adds a key to the bucket
func (b *bucket) add(key string) {
	b.Lock()
	defer b.Unlock()
	b.entries[key] = struct{}{}
}

// remove removes a key from the bucket
func (b *bucket) remove(key string) {
	b.Lock()
	defer b.Unlock()
	delete(b.entries, key)
}

// getAndClear returns all keys in the bucket and clears it
// Must be called with bucket lock held
func (b *bucket) getAndClear() []string {
	b.Lock()
	defer b.Unlock()

	keys := make([]string, 0, len(b.entries))
	for key := range b.entries {
		keys = append(keys, key)
	}
	b.entries = make(map[string]struct{})
	return keys
}

// NewTimeWheel creates a new time wheel
// interval: tick interval (e.g., 1ms)
// wheelSize: number of buckets in the wheel (e.g., 1024)
// onExpire: callback function when a key expires
func NewTimeWheel(interval time.Duration, wheelSize int, onExpire func(string)) *TimeWheel {
	if interval <= 0 {
		interval = time.Millisecond
	}
	if wheelSize <= 0 {
		wheelSize = 1024
	}

	tw := &TimeWheel{
		interval:  interval,
		wheelSize: wheelSize,
		buckets:   make([]*bucket, wheelSize),
		stopChan:  make(chan struct{}),
		onExpire:  onExpire,
	}

	// Initialize buckets
	for i := 0; i < wheelSize; i++ {
		tw.buckets[i] = newBucket()
	}

	return tw
}

// Start starts the time wheel
func (tw *TimeWheel) Start() {
	tw.Lock()
	defer tw.Unlock()

	if tw.running.Load() == 1 {
		return // Already started
	}

	tw.ticker = time.NewTicker(tw.interval)
	tw.stopChan = make(chan struct{})
	tw.running.Store(1)

	tw.wg.Add(1)
	go func() {
		defer tw.wg.Done()
		for tw.running.Load() == 1 {
			select {
			case <-tw.ticker.C:
				tw.tick()
			case <-tw.stopChan:
				return
			}
		}
	}()
}

// Stop stops the time wheel
func (tw *TimeWheel) Stop() {
	// First, signal stop
	tw.Lock()
	if tw.running.Load() == 0 {
		tw.Unlock()
		return // Already stopped
	}
	tw.running.Store(0)

	if tw.ticker != nil {
		tw.ticker.Stop()
	}
	close(tw.stopChan)
	tw.ticker = nil
	tw.Unlock()

	// Then wait for goroutine to finish (without holding lock)
	tw.wg.Wait()
}

// tick advances the time wheel by one tick
func (tw *TimeWheel) tick() {
	tw.Lock()
	defer tw.Unlock()

	// Move to next bucket
	index := tw.currentTime % int64(tw.wheelSize)
	bucket := tw.buckets[index]

	// Get all keys in this bucket and expire them
	keys := bucket.getAndClear()
	for _, key := range keys {
		if tw.onExpire != nil {
			tw.onExpire(key)
		}
	}

	// Advance current time
	tw.currentTime++
}

// Add adds a key to expire after the given duration
func (tw *TimeWheel) Add(key string, duration time.Duration) {
	if duration <= 0 {
		return
	}

	tw.Lock()
	defer tw.Unlock()

	// Calculate which bucket this key belongs to
	ticks := duration.Milliseconds() / tw.interval.Milliseconds()
	if ticks <= 0 {
		ticks = 1
	}

	bucketIndex := (tw.currentTime + ticks) % int64(tw.wheelSize)
	bucket := tw.buckets[bucketIndex]

	// Add key to bucket
	bucket.add(key)
}

// Remove removes a key from the time wheel
// Note: This is a simplified version that scans all buckets.
// In production, you'd want a more efficient lookup mechanism.
func (tw *TimeWheel) Remove(key string) {
	tw.Lock()
	defer tw.Unlock()

	// Remove from all buckets
	for _, bucket := range tw.buckets {
		bucket.remove(key)
	}
}

// Size returns the total number of keys in the time wheel
func (tw *TimeWheel) Size() int {
	tw.Lock()
	defer tw.Unlock()

	total := 0
	for _, bucket := range tw.buckets {
		bucket.Lock()
		total += len(bucket.entries)
		bucket.Unlock()
	}
	return total
}

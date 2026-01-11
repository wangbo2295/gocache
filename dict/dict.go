package dict

import (
	"sync"
	"sync/atomic"
)

// ConcurrentDict is a thread-safe dictionary with sharded locks
type ConcurrentDict struct {
	table      []*shard
	count      int32
	shardCount int
}

// shard represents a single shard with its own lock
type shard struct {
	m     map[string]interface{}
	mutex sync.RWMutex
}

const (
	defaultShardCount = 16
)

// MakeConcurrentDict creates a new concurrent dictionary
func MakeConcurrentDict(shardCount int) *ConcurrentDict {
	// Adjust shard count to power of 2
	if shardCount < 1 {
		shardCount = defaultShardCount
	}
	// Find next power of 2
	shardCount = 1 << (32 - leadingZeros(uint32(shardCount-1)))

	dict := &ConcurrentDict{
		table:      make([]*shard, shardCount),
		shardCount: shardCount,
	}
	for i := 0; i < shardCount; i++ {
		dict.table[i] = &shard{
			m: make(map[string]interface{}),
		}
	}
	return dict
}

// leadingZeros counts the number of leading zeros in a 32-bit unsigned integer
func leadingZeros(x uint32) uint32 {
	if x == 0 {
		return 32
	}
	var n uint32 = 0
	if x>>16 == 0 {
		n += 16
		x <<= 16
	}
	if x>>24 == 0 {
		n += 8
		x <<= 8
	}
	if x>>28 == 0 {
		n += 4
		x <<= 4
	}
	if x>>30 == 0 {
		n += 2
		x <<= 2
	}
	if x>>31 == 0 {
		n++
	}
	return n
}

// spread calculates the shard index for a given key using optimized FNV-1a hash
// This inline implementation avoids memory allocation from hash/fnv package
func (d *ConcurrentDict) spread(key string) uint32 {
	// FNV-1a 32-bit hash algorithm
	// This is a zero-allocation implementation using bitwise operations
	const (
		offset32 = 2166136261
		prime32  = 16777619
	)

	hash := uint32(offset32)
	for i := 0; i < len(key); i++ {
		hash ^= uint32(key[i])
		hash *= prime32
	}

	// Use higher bits for better distribution and modulo for shard selection
	// Since shardCount is always power of 2, we can use bitwise AND for faster computation
	return (hash >> 16) & (uint32(d.shardCount) - 1)
}

// Get retrieves the value for a given key
func (d *ConcurrentDict) Get(key string) (interface{}, bool) {
	index := d.spread(key)
	shard := d.table[index]
	shard.mutex.RLock()
	defer shard.mutex.RUnlock()
	val, ok := shard.m[key]
	return val, ok
}

// Put stores a key-value pair, returns 1 if key is new, 0 if updating existing key
func (d *ConcurrentDict) Put(key string, val interface{}) (result int) {
	index := d.spread(key)
	shard := d.table[index]
	shard.mutex.Lock()
	defer shard.mutex.Unlock()

	_, existed := shard.m[key]
	shard.m[key] = val

	if !existed {
		atomic.AddInt32(&d.count, 1)
		return 1
	}
	return 0
}

// PutIfExists puts value only if key exists, returns 1 if updated, 0 otherwise
func (d *ConcurrentDict) PutIfExists(key string, val interface{}) (result int) {
	index := d.spread(key)
	shard := d.table[index]
	shard.mutex.Lock()
	defer shard.mutex.Unlock()

	if _, existed := shard.m[key]; existed {
		shard.m[key] = val
		return 1
	}
	return 0
}

// PutIfAbsent puts value only if key does not exist, returns 1 if inserted, 0 otherwise
func (d *ConcurrentDict) PutIfAbsent(key string, val interface{}) (result int) {
	index := d.spread(key)
	shard := d.table[index]
	shard.mutex.Lock()
	defer shard.mutex.Unlock()

	if _, existed := shard.m[key]; !existed {
		shard.m[key] = val
		atomic.AddInt32(&d.count, 1)
		return 1
	}
	return 0
}

// Remove deletes a key, returns 1 if key existed, 0 otherwise
func (d *ConcurrentDict) Remove(key string) (result int) {
	index := d.spread(key)
	shard := d.table[index]
	shard.mutex.Lock()
	defer shard.mutex.Unlock()

	if _, existed := shard.m[key]; existed {
		delete(shard.m, key)
		atomic.AddInt32(&d.count, -1)
		return 1
	}
	return 0
}

// Len returns the number of keys in the dictionary
func (d *ConcurrentDict) Len() int {
	return int(atomic.LoadInt32(&d.count))
}

// ForEach iterates over all key-value pairs in the dictionary
// The iteration is not atomic - keys may be added or removed during iteration
func (d *ConcurrentDict) ForEach(consumer func(key string, val interface{}) bool) {
	for _, shard := range d.table {
		shard.mutex.RLock()
		for key, val := range shard.m {
			// Return false to stop iteration
			if !consumer(key, val) {
				shard.mutex.RUnlock()
				return
			}
		}
		shard.mutex.RUnlock()
	}
}

// Keys returns all keys in the dictionary
// Warning: Not atomic, keys may be added or removed during iteration
func (d *ConcurrentDict) Keys() []string {
	keys := make([]string, 0, d.Len())
	d.ForEach(func(key string, val interface{}) bool {
		keys = append(keys, key)
		return true
	})
	return keys
}

// RandomKeys returns n random keys from the dictionary
// May return fewer keys if dictionary has less than n keys
func (d *ConcurrentDict) RandomKeys(n int) []string {
	if n <= 0 {
		return []string{}
	}

	// Get total number of keys
	size := d.Len()
	if size == 0 {
		return []string{}
	}

	if n > size {
		n = size
	}

	result := make([]string, 0, n)
	// Simple approach: iterate through shards and collect keys
	// TODO: Use reservoir sampling for better randomness
	for _, shard := range d.table {
		shard.mutex.RLock()
		for key := range shard.m {
			result = append(result, key)
			if len(result) >= n {
				shard.mutex.RUnlock()
				return result
			}
		}
		shard.mutex.RUnlock()
	}

	return result
}

// RandomDistinctKeys returns n distinct random keys
func (d *ConcurrentDict) RandomDistinctKeys(n int) []string {
	// Use RandomKeys which already returns distinct keys
	return d.RandomKeys(n)
}

// Clear removes all keys from the dictionary
func (d *ConcurrentDict) Clear() {
	for _, shard := range d.table {
		shard.mutex.Lock()
		shard.m = make(map[string]interface{})
		shard.mutex.Unlock()
	}
	atomic.StoreInt32(&d.count, 0)
}

// AtomicUpdate performs a read-modify-write operation atomically on a key
// The updater function receives the current value (or nil if key doesn't exist)
// and returns the new value. The shard lock is held during the entire operation.
// Returns the previous value (or nil if key didn't exist) and true if key existed.
func (d *ConcurrentDict) AtomicUpdate(key string, updater func(interface{}) interface{}) (interface{}, bool) {
	index := d.spread(key)
	shard := d.table[index]
	shard.mutex.Lock()
	defer shard.mutex.Unlock()

	val, existed := shard.m[key]
	// Call updater with current value to get new value
	newVal := updater(val)
	shard.m[key] = newVal

	if !existed {
		atomic.AddInt32(&d.count, 1)
	}
	return val, existed
}

// AtomicGetAndUpdate atomically gets the current value and updates it with a new value
// Returns the previous value (or nil if key didn't exist) and whether it existed.
func (d *ConcurrentDict) AtomicGetAndUpdate(key string, newVal interface{}) (interface{}, bool) {
	index := d.spread(key)
	shard := d.table[index]
	shard.mutex.Lock()
	defer shard.mutex.Unlock()

	val, existed := shard.m[key]
	shard.m[key] = newVal

	if !existed {
		atomic.AddInt32(&d.count, 1)
	}
	return val, existed
}

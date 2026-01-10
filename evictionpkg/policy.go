package evictionpkg

// EvictionPolicy defines the interface for data eviction policies
type EvictionPolicy interface {
	// RecordAccess records that a key was accessed
	RecordAccess(key string)

	// RecordUpdate records that a key was updated
	RecordUpdate(key string)

	// RecordDelete records that a key was deleted
	RecordDelete(key string)

	// Evict selects and returns keys to evict
	Evict(count int) []string

	// Reset clears all tracking data
	Reset()
}

// EvictionPolicyType represents the type of eviction policy
type EvictionPolicyType string

const (
	// NoEviction - never evict (return errors when memory limit reached)
	NoEviction EvictionPolicyType = "noeviction"
	// AllKeysLRU - evict least recently used keys from all keys
	AllKeysLRU EvictionPolicyType = "allkeys-lru"
	// AllKeysLFU - evict least frequently used keys from all keys
	AllKeysLFU EvictionPolicyType = "allkeys-lfu"
	// VolatileLRU - evict least recently used keys with expiry set
	VolatileLRU EvictionPolicyType = "volatile-lru"
	// VolatileLFU - evict least frequently used keys with expiry set
	VolatileLFU EvictionPolicyType = "volatile-lfu"
	// AllKeysRandom - evict random keys
	AllKeysRandom EvictionPolicyType = "allkeys-random"
	// VolatileRandom - evict random keys with expiry set
	VolatileRandom EvictionPolicyType = "volatile-random"
	// VolatileTTL - evict keys with shortest TTL first
	VolatileTTL EvictionPolicyType = "volatile-ttl"
)

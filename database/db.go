package database

import (
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/wangbo/gocache/config"
	"github.com/wangbo/gocache/datastruct"
	"github.com/wangbo/gocache/dict"
	"github.com/wangbo/gocache/eviction"
	"github.com/wangbo/gocache/evictionpkg"
)

// DB represents a single database instance
type DB struct {
	index      int
	data       *dict.ConcurrentDict
	ttlMap     *dict.ConcurrentDict
	versionMap *dict.ConcurrentDict
	mu         sync.RWMutex

	// Eviction support
	evictionPolicy evictionpkg.EvictionPolicy
	usedMemory     int64 // Current memory usage in bytes

	// Time wheel for TTL management
	timeWheel *datastruct.TimeWheel

	// Transaction support
	multiState *MultiState

	// RDB save state
	lastSaveTime       time.Time
	bgSaveInProgress   bool
	bgSaveStartTime    time.Time
	bgSaveMu           sync.Mutex // Protects bgSave fields

	// Slow log
	slowLog        []*SlowLogEntry
	slowLogMu      sync.Mutex
	slowLogMaxLen  int // Maximum number of slow log entries (default 128)
}

// toLowerBytes converts a byte slice to lowercase in-place without allocation
// This is an optimized version for ASCII commands, avoiding string conversion
func toLowerBytes(b []byte) string {
	// Convert to lowercase in-place (ASCII only)
	for i := 0; i < len(b); i++ {
		if b[i] >= 'A' && b[i] <= 'Z' {
			b[i] += 32
		}
	}
	// Use BytesToString for zero-allocation conversion
	return BytesToString(b)
}

// SlowLogEntry represents a slow log entry
type SlowLogEntry struct {
	ID        int64
	Timestamp time.Time
	Duration  int64  // Execution time in microseconds
	Command   []byte // The command that was executed
}

// MakeDB creates a new database instance
func MakeDB() *DB {
	db := &DB{
		index:         0,
		data:          dict.MakeConcurrentDict(16),
		ttlMap:        dict.MakeConcurrentDict(16),
		versionMap:    dict.MakeConcurrentDict(16),
		usedMemory:    0,
		slowLogMaxLen: 128, // Default max 128 slow log entries
	}

	// Initialize eviction policy based on config
	db.initEvictionPolicy()

	// Initialize time wheel for TTL management (10ms interval, 1024 buckets)
	db.timeWheel = datastruct.NewTimeWheel(
		10*time.Millisecond,    // 10ms tick interval
		1024,                   // 1024 buckets (covers ~10 seconds)
		db.expireFromTimeWheel, // Callback when key expires
	)
	db.timeWheel.Start()

	// Initialize transaction state
	db.multiState = NewMultiState(db)

	return db
}

// initEvictionPolicy initializes the eviction policy based on config
func (db *DB) initEvictionPolicy() {
	policy := config.Config.MaxMemoryPolicy

	switch policy {
	case "allkeys-lru", "volatile-lru":
		// Use LRU with a large capacity (will be limited by memory)
		db.evictionPolicy = eviction.NewLRU(1000000)
	case "allkeys-lfu", "volatile-lfu":
		// Use LFU with a large capacity
		db.evictionPolicy = eviction.NewLFU(1000000)
	case "allkeys-random", "volatile-random":
		// Use Random eviction
		db.evictionPolicy = eviction.NewRandom()
	case "volatile-ttl":
		// Use TTL-based eviction (only for keys with TTL)
		db.evictionPolicy = eviction.NewTTL()
	default:
		// No eviction or other policies not yet implemented
		db.evictionPolicy = nil
	}
}

// GetUsedMemory returns the current memory usage in bytes
func (db *DB) GetUsedMemory() int64 {
	return atomic.LoadInt64(&db.usedMemory)
}

// addMemoryUsage adds to the memory usage counter
func (db *DB) addMemoryUsage(delta int64) {
	atomic.AddInt64(&db.usedMemory, delta)
}

// checkAndEvict checks if memory limit is exceeded and evicts if necessary
func (db *DB) checkAndEvict() {
	if config.Config.MaxMemory <= 0 {
		return // No memory limit set
	}

	if db.evictionPolicy == nil {
		return // No eviction policy
	}

	usedMemory := db.GetUsedMemory()
	maxMemory := config.Config.MaxMemory

	// If over limit, evict keys
	for usedMemory > maxMemory {
		// Evict up to 10 keys at a time to reduce lock contention
		keys := db.evictionPolicy.Evict(10)
		if len(keys) == 0 {
			break
		}

		for _, key := range keys {
			// Remove from database (will subtract memory usage and record deletion)
			db.Remove(key)
		}

		usedMemory = db.GetUsedMemory()
	}
}

// Exec executes a command and returns a reply
func (db *DB) Exec(cmdLine [][]byte) (result [][]byte, err error) {
	if len(cmdLine) == 0 {
		return nil, errors.New("empty command")
	}

	// Make a copy of cmdLine[0] to avoid modifying the original
	cmdBytes := make([]byte, len(cmdLine[0]))
	copy(cmdBytes, cmdLine[0])
	cmd := toLowerBytes(cmdBytes)
	args := cmdLine[1:]

	// Parse command type using registry
	cmdType, ok := ParseCommandType(cmd)
	if !ok {
		return nil, errors.New("unknown command: " + cmd)
	}

	// Get command executor from registry
	executor, ok := GetCommandExecutor(cmdType)
	if !ok {
		return nil, errors.New("command not implemented: " + cmd)
	}

	// Transaction commands (MULTI, EXEC, DISCARD, WATCH, UNWATCH) are always executed immediately
	// They control transaction state and should not be queued
	switch cmdType {
	case CmdMulti, CmdExec, CmdDiscard, CmdWatch, CmdUnwatch:
		return executor.Execute(db, args)
	}

	// If in MULTI mode, queue non-transaction commands instead of executing
	if db.multiState.IsInMulti() {
		// Convert cmdLine to []string for queuing (using SafeBytesToString for safety)
		cmdStr := make([]string, len(cmdLine))
		for i, b := range cmdLine {
			cmdStr[i] = SafeBytesToString(b)
		}

		if err := db.multiState.Enqueue(cmdStr); err != nil {
			return nil, err
		}

		return [][]byte{[]byte("QUEUED")}, nil
	}

	// Execute command using command executor - no more switch-case!
	return executor.Execute(db, args)
}

// GetEntity retrieves the data entity for a given key
// It checks TTL and removes expired keys automatically
func (db *DB) GetEntity(key string) (*datastruct.DataEntity, bool) {
	// Check if key is expired
	db.expireIfNeeded(key)

	val, ok := db.data.Get(key)
	if !ok {
		return nil, false
	}
	entity, ok := val.(*datastruct.DataEntity)
	if !ok {
		return nil, false
	}

	// Record access in eviction policy
	if db.evictionPolicy != nil {
		db.evictionPolicy.RecordAccess(key)
	}

	return entity, true
}

// getEntityWithoutExpiryCheck retrieves the data entity without checking TTL
// This is used internally to avoid circular calls
func (db *DB) getEntityWithoutExpiryCheck(key string) (*datastruct.DataEntity, bool) {
	val, ok := db.data.Get(key)
	if !ok {
		return nil, false
	}
	entity, ok := val.(*datastruct.DataEntity)
	if !ok {
		return nil, false
	}
	return entity, true
}

// PutEntity stores a data entity
func (db *DB) PutEntity(key string, entity *datastruct.DataEntity) int {
	// Check if key already exists
	_, exists := db.data.Get(key)

	// Put the entity
	result := db.data.Put(key, entity)

	// Increment version for WATCH
	db.incrementVersion(key)

	// Track memory and eviction based on whether it was new or existing
	if !exists {
		// New key - add to memory usage
		size := entity.EstimateSize()
		db.addMemoryUsage(size)

		// Record in eviction policy
		if db.evictionPolicy != nil {
			db.evictionPolicy.RecordAccess(key)
		}

		// Check if we need to evict
		db.checkAndEvict()
	} else {
		// Existing key - record update in eviction policy
		if db.evictionPolicy != nil {
			db.evictionPolicy.RecordUpdate(key)
		}
	}

	return result
}

// PutIfExists updates entity only if key exists
func (db *DB) PutIfExists(key string, entity *datastruct.DataEntity) int {
	result := db.data.PutIfExists(key, entity)

	if result == 1 && db.evictionPolicy != nil {
		db.evictionPolicy.RecordUpdate(key)
	}

	return result
}

// PutIfAbsent inserts entity only if key does not exist
func (db *DB) PutIfAbsent(key string, entity *datastruct.DataEntity) int {
	result := db.data.PutIfAbsent(key, entity)

	if result == 1 {
		// New key - add to memory usage
		size := entity.EstimateSize()
		db.addMemoryUsage(size)

		// Record in eviction policy
		if db.evictionPolicy != nil {
			db.evictionPolicy.RecordAccess(key)
		}

		// Check if we need to evict
		db.checkAndEvict()
	}

	return result
}

// Remove removes a key from the database
func (db *DB) Remove(key string) int {
	// Calculate size before removing (use internal method to avoid circular call)
	entity, ok := db.getEntityWithoutExpiryCheck(key)
	var size int64
	if ok && entity != nil {
		size = entity.EstimateSize()
	}

	result := db.data.Remove(key)

	// Increment version for WATCH (even on delete) - do this BEFORE removing from versionMap
	db.incrementVersion(key)

	db.ttlMap.Remove(key)
	db.versionMap.Remove(key)

	// Remove from time wheel
	db.timeWheel.Remove(key)

	// Subtract from memory usage
	if size > 0 {
		db.addMemoryUsage(-size)
	}

	// Record deletion in eviction policy
	if result > 0 && db.evictionPolicy != nil {
		db.evictionPolicy.RecordDelete(key)
	}

	return result
}

// Exists checks if a key exists
func (db *DB) Exists(key string) bool {
	db.expireIfNeeded(key)
	_, ok := db.data.Get(key)
	return ok
}

// Expire sets a TTL for a key (in seconds)
func (db *DB) Expire(key string, ttl time.Duration) int {
	if _, ok := db.data.Get(key); !ok {
		return 0
	}

	// Remove from time wheel if it was there
	_, hasExistingTTL := db.ttlMap.Get(key)
	if hasExistingTTL {
		db.timeWheel.Remove(key)
	}

	// Store exact expiration time in ttlMap for precise TTL queries
	db.ttlMap.Put(key, time.Now().Add(ttl))

	// Add to time wheel for active expiration
	db.timeWheel.Add(key, ttl)

	return 1
}

// Persist removes the TTL from a key
func (db *DB) Persist(key string) int {
	if _, ok := db.ttlMap.Get(key); !ok {
		return 0
	}
	db.ttlMap.Remove(key)

	// Remove from time wheel
	db.timeWheel.Remove(key)

	return 1
}

// expireFromTimeWheel is called by the time wheel when a key expires
// Note: This is called from within the time wheel's tick loop, so we must
// avoid calling timeWheel.Remove() to prevent deadlock
func (db *DB) expireFromTimeWheel(key string) {
	// Check if key still exists and is expired
	val, ok := db.ttlMap.Get(key)
	if !ok {
		return // Key already removed or doesn't have TTL
	}

	expireTime, ok := val.(time.Time)
	if !ok {
		return
	}

	// Double-check that it's actually expired
	if time.Now().Before(expireTime) {
		return // Not expired yet, might have been updated
	}

	// Remove the key from data structures (but don't call timeWheel.Remove
	// since we're already in the time wheel's callback)
	entity, ok := db.getEntityWithoutExpiryCheck(key)
	var size int64
	if ok && entity != nil {
		size = entity.EstimateSize()
	}

	db.data.Remove(key)
	db.ttlMap.Remove(key)
	db.versionMap.Remove(key)

	// Subtract from memory usage
	if size > 0 {
		db.addMemoryUsage(-size)
	}

	// Record deletion in eviction policy
	if db.evictionPolicy != nil {
		db.evictionPolicy.RecordDelete(key)
	}
}

// TTL returns the remaining TTL in seconds
// Returns -2 if key does not exist, -1 if key exists but has no expiry
func (db *DB) TTL(key string) time.Duration {
	db.expireIfNeeded(key)

	if _, ok := db.data.Get(key); !ok {
		return -2
	}

	val, ok := db.ttlMap.Get(key)
	if !ok {
		return -1
	}

	expireTime := val.(time.Time)
	remaining := time.Until(expireTime)
	if remaining < 0 {
		// Already expired
		db.Remove(key)
		return -2
	}

	return remaining
}

// expireIfNeeded checks and removes expired key
func (db *DB) expireIfNeeded(key string) {
	val, ok := db.ttlMap.Get(key)
	if !ok {
		return
	}

	expireTime := val.(time.Time)
	if time.Now().After(expireTime) {
		db.Remove(key)
	}
}

// ExecCommand is a convenience method to execute command from strings
func (db *DB) ExecCommand(cmd string, args ...string) ([][]byte, error) {
	cmdLine := make([][]byte, 0, len(args)+1)
	cmdLine = append(cmdLine, []byte(cmd))
	for _, arg := range args {
		cmdLine = append(cmdLine, []byte(arg))
	}
	return db.Exec(cmdLine)
}

// atomicIncr performs an atomic increment operation on a key
// This uses ConcurrentDict's AtomicUpdate to ensure the read-modify-write is atomic
func (db *DB) atomicIncr(key string, delta int64) (int64, error) {
	// Check if key is expired first
	db.expireIfNeeded(key)

	var result int64
	var err error

	// Use AtomicUpdate to perform the increment atomically
	db.data.AtomicUpdate(key, func(val interface{}) interface{} {
		var str *datastruct.String

		if val != nil {
			entity, ok := val.(*datastruct.DataEntity)
			if !ok {
				err = errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
				return nil
			}
			str, ok = entity.Data.(*datastruct.String)
			if !ok {
				err = errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
				return nil
			}
		} else {
			// Key doesn't exist, create new String with value "0"
			str = &datastruct.String{Value: []byte("0")}
		}

		// Perform the increment
		var newVal int64
		newVal, err = str.Increment(delta)
		result = newVal

		// Return updated entity
		return &datastruct.DataEntity{Data: str}
	})

	if err != nil {
		return 0, err
	}

	// Increment version for WATCH
	db.incrementVersion(key)

	// Record access in eviction policy
	if db.evictionPolicy != nil {
		db.evictionPolicy.RecordAccess(key)
	}

	return result, nil
}

// Close stops the time wheel and cleans up resources gracefully
func (db *DB) Close() error {
	// 1. Stop time wheel (stop accepting new TTL callbacks)
	if db.timeWheel != nil {
		db.timeWheel.Stop()
	}

	// 2. Clear all data structures
	if db.data != nil {
		db.data.Clear()
	}
	if db.ttlMap != nil {
		db.ttlMap.Clear()
	}
	if db.versionMap != nil {
		db.versionMap.Clear()
	}

	// 3. Reset counters
	atomic.StoreInt64(&db.usedMemory, 0)

	// 4. Clear slow log
	db.slowLogMu.Lock()
	db.slowLog = nil
	db.slowLogMu.Unlock()

	// 5. Reset transaction state
	if db.multiState != nil {
		db.multiState.Discard()
	}

	return nil
}

// Keys returns all keys in the database
func (db *DB) Keys() []string {
	return db.data.Keys()
}

// GetVersion returns the version of a key (for WATCH)
func (db *DB) GetVersion(key string) uint64 {
	val, ok := db.versionMap.Get(key)
	if !ok {
		return 0
	}

	version, ok := val.(uint64)
	if !ok {
		return 0
	}

	return version
}

// incrementVersion increments the version of a key
func (db *DB) incrementVersion(key string) {
	// Use atomic add for version increment
	val, ok := db.versionMap.Get(key)
	var version uint64
	if ok {
		if v, ok := val.(uint64); ok {
			version = v + 1
		} else {
			version = 1
		}
	} else {
		version = 1
	}

	db.versionMap.Put(key, version)
}

// SlowLog methods

// AddSlowLogEntry adds an entry to the slow log if the duration exceeds the threshold
func (db *DB) AddSlowLogEntry(duration time.Duration, cmdLine [][]byte) {
	// Only log if execution time exceeds threshold (default 10ms)
	const slowLogThreshold = 10 * time.Millisecond
	if duration < slowLogThreshold {
		return
	}

	entry := &SlowLogEntry{
		ID:        time.Now().UnixNano(), // Simple ID generation
		Timestamp: time.Now(),
		Duration:  duration.Microseconds(),
		Command:   serializeCommand(cmdLine),
	}

	db.slowLogMu.Lock()
	defer db.slowLogMu.Unlock()

	// Add to beginning of log (most recent first)
	db.slowLog = append([]*SlowLogEntry{entry}, db.slowLog...)

	// Trim if exceeds max length
	if len(db.slowLog) > db.slowLogMaxLen {
		db.slowLog = db.slowLog[:db.slowLogMaxLen]
	}
}

// GetSlowLogEntries returns all slow log entries
func (db *DB) GetSlowLogEntries() []*SlowLogEntry {
	db.slowLogMu.Lock()
	defer db.slowLogMu.Unlock()

	// Return a copy to avoid race conditions
	result := make([]*SlowLogEntry, len(db.slowLog))
	copy(result, db.slowLog)
	return result
}

// GetSlowLogLen returns the number of slow log entries
func (db *DB) GetSlowLogLen() int {
	db.slowLogMu.Lock()
	defer db.slowLogMu.Unlock()
	return len(db.slowLog)
}

// ResetSlowLog clears all slow log entries
func (db *DB) ResetSlowLog() {
	db.slowLogMu.Lock()
	defer db.slowLogMu.Unlock()
	db.slowLog = nil
}

// serializeCommand converts command line to string for logging
func serializeCommand(cmdLine [][]byte) []byte {
	if len(cmdLine) == 0 {
		return []byte{}
	}

	var result []byte
	for i, arg := range cmdLine {
		if i > 0 {
			result = append(result, ' ')
		}
		// Escape arguments with spaces
		if len(arg) == 0 || (len(arg) > 0 && (arg[0] == ' ' || arg[len(arg)-1] == ' ')) {
			result = append(result, '"')
			result = append(result, arg...)
			result = append(result, '"')
		} else {
			result = append(result, arg...)
		}
	}
	return result
}

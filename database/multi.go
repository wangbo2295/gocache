package database

import (
	"errors"
	"sync"
)

// MultiState manages the state of a transaction (MULTI/EXEC)
type MultiState struct {
	mu           sync.Mutex
	inMulti      bool                // Whether in MULTI mode
	commands     [][]string          // Queued commands
	aborted      bool                // Whether transaction is aborted
	watchedKeys  map[string]uint64   // WATCHed keys and their versions
	dirtyKeys    map[string]struct{} // Keys modified during transaction
	db           *DB                 // Reference to the database
}

// NewMultiState creates a new transaction state
func NewMultiState(db *DB) *MultiState {
	return &MultiState{
		inMulti:     false,
		commands:    make([][]string, 0, 16),
		aborted:     false,
		watchedKeys: make(map[string]uint64),
		dirtyKeys:   make(map[string]struct{}),
		db:          db,
	}
}

// Begin starts a MULTI transaction
func (ms *MultiState) Begin() error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	if ms.inMulti {
		return errors.New("ERR MULTI calls can not be nested")
	}

	ms.inMulti = true
	ms.commands = ms.commands[:0] // Clear command queue
	ms.aborted = false
	ms.dirtyKeys = make(map[string]struct{})

	return nil
}

// Discard discards the transaction
func (ms *MultiState) Discard() error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	if !ms.inMulti {
		return errors.New("ERR DISCARD without MULTI")
	}

	ms.inMulti = false
	ms.commands = ms.commands[:0]
	ms.aborted = false
	ms.dirtyKeys = make(map[string]struct{})

	return nil
}

// IsInMulti returns whether currently in MULTI mode
func (ms *MultiState) IsInMulti() bool {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	return ms.inMulti
}

// Enqueue queues a command for execution in transaction
func (ms *MultiState) Enqueue(cmd []string) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	if !ms.inMulti {
		return errors.New("ERR ENQUEUE without MULTI")
	}

	// Clone the command to avoid modification
	cmdCopy := make([]string, len(cmd))
	copy(cmdCopy, cmd)

	ms.commands = append(ms.commands, cmdCopy)
	return nil
}

// GetCommands returns the queued commands
func (ms *MultiState) GetCommands() [][]string {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	// Return a copy to avoid modification
	cmds := make([][]string, len(ms.commands))
	copy(cmds, ms.commands)
	return cmds
}

// Abort marks the transaction as aborted
func (ms *MultiState) Abort() {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.aborted = true
}

// IsAborted returns whether the transaction is aborted
func (ms *MultiState) IsAborted() bool {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	return ms.aborted
}

// Watch marks a key to be watched for modifications
func (ms *MultiState) Watch(keys ...string) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	if ms.inMulti {
		return errors.New("ERR WATCH inside MULTI is not allowed")
	}

	for _, key := range keys {
		// Get current version of the key
		version := ms.db.GetVersion(key)
		ms.watchedKeys[key] = version
	}

	return nil
}

// Unwatch clears all watched keys
func (ms *MultiState) Unwatch() {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	ms.watchedKeys = make(map[string]uint64)
}

// CheckWatchedKeys checks if any watched keys have been modified
// Returns true if conflict detected
func (ms *MultiState) CheckWatchedKeys() bool {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	for key, oldVersion := range ms.watchedKeys {
		currentVersion := ms.db.GetVersion(key)
		if currentVersion != oldVersion {
			// Key has been modified
			return true
		}
	}

	return false
}

// MarkDirty marks a key as modified during transaction
func (ms *MultiState) MarkDirty(key string) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	ms.dirtyKeys[key] = struct{}{}
}

// GetDirtyKeys returns all keys modified during transaction
func (ms *MultiState) GetDirtyKeys() []string {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	keys := make([]string, 0, len(ms.dirtyKeys))
	for key := range ms.dirtyKeys {
		keys = append(keys, key)
	}
	return keys
}

// Clear resets the transaction state
func (ms *MultiState) Clear() {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	ms.inMulti = false
	ms.commands = ms.commands[:0]
	ms.aborted = false
	ms.dirtyKeys = make(map[string]struct{})
}

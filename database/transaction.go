package database

import (
	"errors"
	"fmt"
)

// execMulti executes the MULTI command
func execMulti(db *DB, args [][]byte) ([][]byte, error) {
	if len(args) != 0 {
		return nil, errors.New("ERR wrong number of arguments for MULTI")
	}

	if err := db.multiState.Begin(); err != nil {
		return nil, err
	}

	return [][]byte{[]byte("OK")}, nil
}

// execDiscard executes the DISCARD command
func execDiscard(db *DB, args [][]byte) ([][]byte, error) {
	if len(args) != 0 {
		return nil, errors.New("ERR wrong number of arguments for DISCARD")
	}

	if err := db.multiState.Discard(); err != nil {
		return nil, err
	}

	return [][]byte{[]byte("OK")}, nil
}

// execExec executes the EXEC command
func execExec(db *DB, args [][]byte) ([][]byte, error) {
	if len(args) != 0 {
		return nil, errors.New("ERR wrong number of arguments for EXEC")
	}

	// Check if we're in MULTI mode
	if !db.multiState.IsInMulti() {
		return nil, errors.New("ERR EXEC without MULTI")
	}

	// Check for WATCH conflicts
	if db.multiState.CheckWatchedKeys() {
		db.multiState.Clear()
		return nil, errors.New("WATCH key has been modified")
	}

	// Check if transaction was aborted
	if db.multiState.IsAborted() {
		db.multiState.Clear()
		return nil, errors.New("Transaction aborted due to errors")
	}

	// Get queued commands and clear MULTI state before executing
	commands := db.multiState.GetCommands()
	db.multiState.Clear()

	// Execute all commands atomically
	results := make([][]byte, 0, len(commands))

	for _, cmdArgs := range commands {
		if len(cmdArgs) == 0 {
			continue
		}

		// Convert []string to [][]byte
		cmdBytes := make([][]byte, len(cmdArgs))
		for i, arg := range cmdArgs {
			cmdBytes[i] = []byte(arg)
		}

		// Execute command directly (now that we're not in MULTI mode)
		result, err := db.Exec(cmdBytes)
		if err != nil {
			// Continue execution even on error - append error as result
			// This matches Redis behavior where all commands are executed
			results = append(results, []byte(fmt.Sprintf("ERR %v", err)))
		} else {
			// Append results
			results = append(results, result...)
		}
	}

	// Clear watched keys after EXEC
	db.multiState.Unwatch()

	return results, nil
}

// execWatch executes the WATCH command
func execWatch(db *DB, args [][]byte) ([][]byte, error) {
	if len(args) == 0 {
		return nil, errors.New("ERR wrong number of arguments for WATCH")
	}

	// Convert [][]byte to []string
	keys := make([]string, len(args))
	for i, arg := range args {
		keys[i] = string(arg)
	}

	if err := db.multiState.Watch(keys...); err != nil {
		return nil, err
	}

	return [][]byte{[]byte("OK")}, nil
}

// execUnwatch executes the UNWATCH command
func execUnwatch(db *DB, args [][]byte) ([][]byte, error) {
	if len(args) != 0 {
		return nil, errors.New("ERR wrong number of arguments for UNWATCH")
	}

	db.multiState.Unwatch()

	return [][]byte{[]byte("OK")}, nil
}

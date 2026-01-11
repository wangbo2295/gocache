package database

import (
	"testing"
)

// TestMultiExecBasic tests basic MULTI/EXEC functionality
func TestMultiExecBasic(t *testing.T) {
	db := MakeDB()
	defer db.Close()

	// Start transaction
	result, err := db.ExecCommand("MULTI")
	if err != nil {
		t.Fatalf("MULTI failed: %v", err)
	}
	if len(result) == 0 || string(result[0]) != "OK" {
		t.Errorf("Expected OK, got %v", result)
	}

	// Queue commands
	result, err = db.ExecCommand("SET", "key1", "value1")
	if err != nil {
		t.Fatalf("SET in MULTI failed: %v", err)
	}
	if len(result) == 0 || string(result[0]) != "QUEUED" {
		t.Errorf("Expected QUEUED, got %v", result)
	}

	result, err = db.ExecCommand("SET", "key2", "value2")
	if err != nil {
		t.Fatalf("SET in MULTI failed: %v", err)
	}
	if len(result) == 0 || string(result[0]) != "QUEUED" {
		t.Errorf("Expected QUEUED, got %v", result)
	}

	// Execute transaction
	result, err = db.ExecCommand("EXEC")
	if err != nil {
		t.Fatalf("EXEC failed: %v", err)
	}

	// Should have 2 results (each SET returns OK)
	if len(result) != 2 {
		t.Errorf("Expected 2 results, got %d", len(result))
	}

	// EXEC returns array of command results
	// For SET commands, each returns "OK"
	for i, r := range result {
		if string(r) != "OK" {
			t.Errorf("Expected OK at index %d, got %v", i, string(r))
		}
	}

	// Verify keys were set
	val, err := db.ExecCommand("GET", "key1")
	if err != nil {
		t.Fatalf("GET failed: %v", err)
	}
	if len(val) == 0 || string(val[0]) != "value1" {
		t.Errorf("Expected value1, got %v", val)
	}

	val, err = db.ExecCommand("GET", "key2")
	if err != nil {
		t.Fatalf("GET failed: %v", err)
	}
	if len(val) == 0 || string(val[0]) != "value2" {
		t.Errorf("Expected value2, got %v", val)
	}
}

// TestDiscard tests DISCARD command
func TestDiscard(t *testing.T) {
	db := MakeDB()
	defer db.Close()

	// Start transaction
	result, err := db.ExecCommand("MULTI")
	if err != nil {
		t.Fatalf("MULTI failed: %v", err)
	}

	// Queue a command
	result, err = db.ExecCommand("SET", "key1", "value1")
	if err != nil {
		t.Fatalf("SET in MULTI failed: %v", err)
	}

	// Discard transaction
	result, err = db.ExecCommand("DISCARD")
	if err != nil {
		t.Fatalf("DISCARD failed: %v", err)
	}
	if len(result) == 0 || string(result[0]) != "OK" {
		t.Errorf("Expected OK, got %v", result)
	}

	// Key should not be set
	result, err = db.ExecCommand("GET", "key1")
	if err != nil {
		t.Fatalf("GET failed: %v", err)
	}
	if len(result) > 0 && len(result[0]) > 0 {
		t.Errorf("Expected no value, got %v", result)
	}
}

// TestExecWithoutMulti tests EXEC without MULTI
func TestExecWithoutMulti(t *testing.T) {
	db := MakeDB()
	defer db.Close()

	// EXEC without MULTI should fail
	_, err := db.ExecCommand("EXEC")
	if err == nil {
		t.Error("Expected error for EXEC without MULTI")
	}
}

// TestDiscardWithoutMulti tests DISCARD without MULTI
func TestDiscardWithoutMulti(t *testing.T) {
	db := MakeDB()
	defer db.Close()

	// DISCARD without MULTI should fail
	_, err := db.ExecCommand("DISCARD")
	if err == nil {
		t.Error("Expected error for DISCARD without MULTI")
	}
}

// TestNestedMulti tests that MULTI cannot be nested
func TestNestedMulti(t *testing.T) {
	db := MakeDB()
	defer db.Close()

	// First MULTI
	_, err := db.ExecCommand("MULTI")
	if err != nil {
		t.Fatalf("MULTI failed: %v", err)
	}

	// Second MULTI should fail
	_, err = db.ExecCommand("MULTI")
	if err == nil {
		t.Error("Expected error for nested MULTI")
	}
}

// TestTransactionIsolation tests that transactions are isolated
func TestTransactionIsolation(t *testing.T) {
	db := MakeDB()
	defer db.Close()

	// Set initial value
	db.ExecCommand("SET", "key1", "initial")

	// Get initial value to verify
	result, _ := db.ExecCommand("GET", "key1")
	if len(result) == 0 || string(result[0]) != "initial" {
		t.Fatalf("Expected initial value, got %v", result)
	}

	// Start transaction
	db.ExecCommand("MULTI")
	db.ExecCommand("SET", "key1", "transaction")

	// While in transaction, GET should be queued, not executed
	result, _ = db.ExecCommand("GET", "key1")
	if len(result) == 0 || string(result[0]) != "QUEUED" {
		t.Errorf("Expected QUEUED for GET during transaction, got %v", result)
	}

	// Execute transaction (will execute SET and GET)
	result, err := db.ExecCommand("EXEC")
	if err != nil {
		t.Fatalf("EXEC failed: %v", err)
	}

	// Should have 2 results: SET OK, GET value
	if len(result) != 2 {
		t.Errorf("Expected 2 results, got %d", len(result))
	}

	// Second result should be the value from GET
	if len(result[1]) == 0 || string(result[1]) != "transaction" {
		t.Errorf("Expected 'transaction' from GET in EXEC, got %v", result[1])
	}

	// Now value should be updated
	result, _ = db.ExecCommand("GET", "key1")
	if len(result) == 0 || string(result[0]) != "transaction" {
		t.Errorf("Expected transaction value after EXEC, got %v", result)
	}
}

// TestWatchBasic tests basic WATCH functionality
func TestWatchBasic(t *testing.T) {
	db := MakeDB()
	defer db.Close()

	// Set initial value
	db.ExecCommand("SET", "key1", "value1")

	// Watch key
	result, err := db.ExecCommand("WATCH", "key1")
	if err != nil {
		t.Fatalf("WATCH failed: %v", err)
	}
	if len(result) == 0 || string(result[0]) != "OK" {
		t.Errorf("Expected OK, got %v", result)
	}

	// Modify the watched key
	db.ExecCommand("SET", "key1", "value2")

	// Start transaction and try to EXEC
	db.ExecCommand("MULTI")
	db.ExecCommand("SET", "key1", "value3")
	_, err = db.ExecCommand("EXEC")

	// EXEC should fail because watched key was modified
	if err == nil {
		t.Error("Expected error for EXEC after WATCHed key was modified")
	}
}

// TestUnwatch tests UNWATCH command
func TestUnwatch(t *testing.T) {
	db := MakeDB()
	defer db.Close()

	// Set initial value
	db.ExecCommand("SET", "key1", "value1")

	// Watch key
	db.ExecCommand("WATCH", "key1")

	// Unwatch
	result, err := db.ExecCommand("UNWATCH")
	if err != nil {
		t.Fatalf("UNWATCH failed: %v", err)
	}
	if len(result) == 0 || string(result[0]) != "OK" {
		t.Errorf("Expected OK, got %v", result)
	}

	// Modify the key
	db.ExecCommand("SET", "key1", "value2")

	// Start transaction and EXEC - should succeed because we unwatched
	db.ExecCommand("MULTI")
	db.ExecCommand("SET", "key1", "value3")
	result, err = db.ExecCommand("EXEC")
	if err != nil {
		t.Fatalf("EXEC failed after UNWATCH: %v", err)
	}

	// Verify value was set
	result, _ = db.ExecCommand("GET", "key1")
	if len(result) == 0 || string(result[0]) != "value3" {
		t.Errorf("Expected value3, got %v", result)
	}
}

// TestWatchDetectsDelete tests that WATCH detects DEL operations
func TestWatchDetectsDelete(t *testing.T) {
	db := MakeDB()
	defer db.Close()

	// Set initial value
	db.ExecCommand("SET", "key1", "value1")

	// Watch key
	db.ExecCommand("WATCH", "key1")

	// Delete the key (this should increment version and trigger conflict on EXEC)
	db.ExecCommand("DEL", "key1")

	// Start transaction and try to EXEC
	db.ExecCommand("MULTI")
	db.ExecCommand("SET", "key1", "value2")
	_, err := db.ExecCommand("EXEC")

	// EXEC should fail because watched key was deleted (version changed)
	if err == nil {
		t.Error("Expected error for EXEC after WATCHed key was deleted")
	}
}

// TestWatchMultipleKeys tests watching multiple keys
func TestWatchMultipleKeys(t *testing.T) {
	db := MakeDB()
	defer db.Close()

	// Set initial values
	db.ExecCommand("SET", "key1", "value1")
	db.ExecCommand("SET", "key2", "value2")
	db.ExecCommand("SET", "key3", "value3")

	// Watch multiple keys
	_, err := db.ExecCommand("WATCH", "key1", "key2", "key3")
	if err != nil {
		t.Fatalf("WATCH failed: %v", err)
	}

	// Modify one of the watched keys
	db.ExecCommand("SET", "key2", "modified")

	// Start transaction and try to EXEC
	db.ExecCommand("MULTI")
	db.ExecCommand("SET", "key1", "newvalue")
	_, err = db.ExecCommand("EXEC")

	// EXEC should fail because one watched key was modified
	if err == nil {
		t.Error("Expected error for EXEC after one WATCHed key was modified")
	}
}

// TestTransactionWithDifferentTypes tests transaction with different data types
func TestTransactionWithDifferentTypes(t *testing.T) {
	db := MakeDB()
	defer db.Close()

	// Start transaction
	db.ExecCommand("MULTI")

	// Queue different types of commands
	db.ExecCommand("SET", "stringkey", "stringvalue")
	db.ExecCommand("HSET", "hashkey", "field1", "value1")
	db.ExecCommand("LPUSH", "listkey", "item1")
	db.ExecCommand("SADD", "setkey", "member1")

	// Execute
	result, err := db.ExecCommand("EXEC")
	if err != nil {
		t.Fatalf("EXEC failed: %v", err)
	}

	// Should have 4 results
	if len(result) != 4 {
		t.Errorf("Expected 4 results, got %d", len(result))
	}

	// Verify all keys were created
	// Check string
	val, _ := db.ExecCommand("GET", "stringkey")
	if len(val) == 0 || string(val[0]) != "stringvalue" {
		t.Error("String key not set correctly")
	}

	// Check hash
	val, _ = db.ExecCommand("HGET", "hashkey", "field1")
	if len(val) == 0 || string(val[0]) != "value1" {
		t.Error("Hash field not set correctly")
	}

	// Check list
	val, _ = db.ExecCommand("LINDEX", "listkey", "0")
	if len(val) == 0 || string(val[0]) != "item1" {
		t.Error("List item not added correctly")
	}

	// Check set
	val, _ = db.ExecCommand("SISMEMBER", "setkey", "member1")
	if len(val) == 0 || string(val[0]) != "1" {
		t.Error("Set member not added correctly")
	}
}

// TestTransactionAtomicity tests that transaction is atomic
func TestTransactionAtomicity(t *testing.T) {
	db := MakeDB()
	defer db.Close()

	// Start transaction
	db.ExecCommand("MULTI")

	// Queue commands that modify multiple keys
	db.ExecCommand("SET", "key1", "value1")
	db.ExecCommand("SET", "key2", "value2")
	db.ExecCommand("SET", "key3", "value3")

	// Before EXEC, keys should not exist (they're only queued, not executed)
	// Note: EXISTS in MULTI mode will be queued, not executed
	// We need to check with a new DB connection or use a different approach

	// Execute
	result, err := db.ExecCommand("EXEC")
	if err != nil {
		t.Fatalf("EXEC failed: %v", err)
	}

	// Should have 3 results (all SETs return OK)
	if len(result) != 3 {
		t.Errorf("Expected 3 results, got %d", len(result))
	}

	// After EXEC, all keys should exist atomically
	for _, key := range []string{"key1", "key2", "key3"} {
		result, _ := db.ExecCommand("EXISTS", key)
		if len(result) == 0 || string(result[0]) != "1" {
			t.Errorf("Key %s should exist after EXEC", key)
		}
	}
}

// TestMultiStateIsInMulti tests IsInMulti method
func TestMultiStateIsInMulti(t *testing.T) {
	db := MakeDB()
	defer db.Close()

	// Initially not in MULTI
	if db.multiState.IsInMulti() {
		t.Error("Should not be in MULTI initially")
	}

	// After MULTI, should be in MULTI
	db.ExecCommand("MULTI")
	if !db.multiState.IsInMulti() {
		t.Error("Should be in MULTI after MULTI command")
	}

	// After EXEC, should not be in MULTI
	db.ExecCommand("SET", "key", "value")
	db.ExecCommand("EXEC")
	if db.multiState.IsInMulti() {
		t.Error("Should not be in MULTI after EXEC")
	}
}

// TestVersionIncrement tests version increment on modifications
func TestVersionIncrement(t *testing.T) {
	db := MakeDB()
	defer db.Close()

	// Initially version is 0
	version := db.GetVersion("key1")
	if version != 0 {
		t.Errorf("Expected initial version 0, got %d", version)
	}

	// After SET, version should be incremented
	db.ExecCommand("SET", "key1", "value1")
	version = db.GetVersion("key1")
	if version != 1 {
		t.Errorf("Expected version 1 after SET, got %d", version)
	}

	// After another SET, version should be incremented again
	db.ExecCommand("SET", "key1", "value2")
	version = db.GetVersion("key1")
	if version != 2 {
		t.Errorf("Expected version 2 after second SET, got %d", version)
	}

	// After DEL, version is incremented but key is removed from versionMap
	// (version increment happens before removal for WATCH detection)
	// After DEL, the key no longer exists, so GetVersion returns 0
	db.ExecCommand("DEL", "key1")
	version = db.GetVersion("key1")
	if version != 0 {
		t.Errorf("Expected version 0 after DEL (key removed), got %d", version)
	}

	// Re-SET the same key, version should start fresh at 1
	db.ExecCommand("SET", "key1", "value3")
	version = db.GetVersion("key1")
	if version != 1 {
		t.Errorf("Expected version 1 after re-SET, got %d", version)
	}
}

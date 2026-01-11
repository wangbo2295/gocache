package rdb

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/wangbo/gocache/database"
)

// TestRDBSaveLoad tests RDB save and load functionality
func TestRDBSaveLoad(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()
	rdbFile := filepath.Join(tmpDir, "test.rdb")

	// Create database and add data
	db := database.MakeDB()
	defer db.Close()

	// Add various types of data
	db.ExecCommand("SET", "stringkey", "stringvalue")
	db.ExecCommand("HSET", "hashkey", "field1", "value1")
	db.ExecCommand("HSET", "hashkey", "field2", "value2")
	db.ExecCommand("LPUSH", "listkey", "item1", "item2")
	db.ExecCommand("SADD", "setkey", "member1", "member2")
	db.ExecCommand("ZADD", "zsetkey", "1.0", "one", "2.0", "two")

	// Save to RDB file
	if err := SaveToFile(db, rdbFile); err != nil {
		t.Fatalf("Failed to save RDB: %v", err)
	}

	// Verify file was created
	info, err := os.Stat(rdbFile)
	if err != nil {
		t.Fatalf("RDB file not created: %v", err)
	}

	if info.Size() == 0 {
		t.Error("RDB file is empty")
	}

	// Load into new database
	db2 := database.MakeDB()
	defer db2.Close()

	if err := LoadFromFile(db2, rdbFile); err != nil {
		t.Fatalf("Failed to load RDB: %v", err)
	}

	// Verify data was restored
	val, _ := db2.ExecCommand("GET", "stringkey")
	if len(val) == 0 || string(val[0]) != "stringvalue" {
		t.Error("String key not restored correctly")
	}

	val, _ = db2.ExecCommand("HGET", "hashkey", "field1")
	if len(val) == 0 || string(val[0]) != "value1" {
		t.Error("Hash field not restored correctly")
	}

	val, _ = db2.ExecCommand("LINDEX", "listkey", "0")
	if len(val) == 0 || string(val[0]) != "item2" {
		t.Error("List not restored correctly")
	}

	val, _ = db2.ExecCommand("SISMEMBER", "setkey", "member1")
	if len(val) == 0 || string(val[0]) != "1" {
		t.Error("Set member not restored correctly")
	}

	val, _ = db2.ExecCommand("ZSCORE", "zsetkey", "one")
	if len(val) == 0 || string(val[0]) != "1" {
		t.Error("Sorted set not restored correctly")
	}
}

// TestRDBEmpty tests saving and loading an empty database
func TestRDBEmpty(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()
	rdbFile := filepath.Join(tmpDir, "empty.rdb")

	// Create empty database
	db := database.MakeDB()
	defer db.Close()

	// Save to RDB file
	if err := SaveToFile(db, rdbFile); err != nil {
		t.Fatalf("Failed to save empty RDB: %v", err)
	}

	// Load into new database
	db2 := database.MakeDB()
	defer db2.Close()

	if err := LoadFromFile(db2, rdbFile); err != nil {
		t.Fatalf("Failed to load empty RDB: %v", err)
	}

	// Verify no keys exist
	keys := db2.Keys()
	if len(keys) != 0 {
		t.Errorf("Expected 0 keys, got %d", len(keys))
	}
}

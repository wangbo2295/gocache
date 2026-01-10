package aof

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/wangbo/gocache/database"
)

func TestMakeAOFHandler(t *testing.T) {
	// Create temp file
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "test.aof")

	db := database.MakeDB()
	handler, err := MakeAOFHandler(filename, db)
	if err != nil {
		t.Fatalf("MakeAOFHandler failed: %v", err)
	}
	defer handler.Close()

	if handler == nil {
		t.Fatal("Handler is nil")
	}
	if handler.db != db {
		t.Error("DB not set correctly")
	}
}

func TestAOFHandler_AddCommand(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "test.aof")

	db := database.MakeDB()
	handler, err := MakeAOFHandler(filename, db)
	if err != nil {
		t.Fatalf("MakeAOFHandler failed: %v", err)
	}
	defer handler.Close()

	// Add SET command
	cmd := [][]byte{[]byte("SET"), []byte("key1"), []byte("value1")}
	if err := handler.AddCommand(cmd); err != nil {
		t.Fatalf("AddCommand failed: %v", err)
	}

	// Verify file was written
	info, err := os.Stat(filename)
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}
	if info.Size() == 0 {
		t.Error("File is empty")
	}
}

func TestAOFHandler_Load(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "test.aof")

	// Create AOF handler and add commands
	db := database.MakeDB()
	handler, err := MakeAOFHandler(filename, db)
	if err != nil {
		t.Fatalf("MakeAOFHandler failed: %v", err)
	}

	// Add some commands
	handler.AddCommand([][]byte{[]byte("SET"), []byte("key1"), []byte("value1")})
	handler.AddCommand([][]byte{[]byte("SET"), []byte("key2"), []byte("value2")})
	handler.AddCommand([][]byte{[]byte("INCR"), []byte("counter")})

	// Close handler
	handler.Close()

	// Create new handler and load data
	db2 := database.MakeDB()
	handler2, err := MakeAOFHandler(filename, db2)
	if err != nil {
		t.Fatalf("MakeAOFHandler failed: %v", err)
	}
	defer handler2.Close()

	// Verify data was loaded
	result, err := db2.ExecCommand("GET", "key1")
	if err != nil {
		t.Fatalf("GET failed: %v", err)
	}
	if len(result) == 0 || string(result[0]) != "value1" {
		t.Errorf("Expected 'value1', got %v", result)
	}

	result, err = db2.ExecCommand("GET", "key2")
	if err != nil {
		t.Fatalf("GET failed: %v", err)
	}
	if len(result) == 0 || string(result[0]) != "value2" {
		t.Errorf("Expected 'value2', got %v", result)
	}

	result, err = db2.ExecCommand("GET", "counter")
	if err != nil {
		t.Fatalf("GET failed: %v", err)
	}
	if len(result) == 0 || string(result[0]) != "1" {
		t.Errorf("Expected '1', got %v", result)
	}
}

func TestAOFHandler_Close(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "test.aof")

	db := database.MakeDB()
	handler, err := MakeAOFHandler(filename, db)
	if err != nil {
		t.Fatalf("MakeAOFHandler failed: %v", err)
	}

	// Add command
	if err := handler.AddCommand([][]byte{[]byte("SET"), []byte("key"), []byte("value")}); err != nil {
		t.Fatalf("AddCommand failed: %v", err)
	}

	// Close handler
	if err := handler.Close(); err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	// Try to add command after close (should fail)
	if err := handler.AddCommand([][]byte{[]byte("SET"), []byte("key2"), []byte("value2")}); err == nil {
		t.Error("Expected error when adding command after close")
	}
}

func TestAOFHandler_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "test.aof")

	// Create empty file
	file, err := os.Create(filename)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	file.Close()

	// Load empty file should work
	db := database.MakeDB()
	handler, err := MakeAOFHandler(filename, db)
	if err != nil {
		t.Fatalf("MakeAOFHandler failed: %v", err)
	}
	defer handler.Close()

	// Verify DB is empty
	result, err := db.ExecCommand("GET", "nonexistent")
	if err != nil {
		t.Fatalf("GET failed: %v", err)
	}
	if len(result) != 1 || result[0] != nil {
		t.Errorf("Expected nil result, got %v", result)
	}
}

func TestAOFHandler_ConcurrentWrites(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "test.aof")

	db := database.MakeDB()
	handler, err := MakeAOFHandler(filename, db)
	if err != nil {
		t.Fatalf("MakeAOFHandler failed: %v", err)
	}
	defer handler.Close()

	// Write concurrently
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(id int) {
			cmd := [][]byte{[]byte("SET"), []byte("key" + string(rune('0'+id))), []byte("value" + string(rune('0'+id)))}
			if err := handler.AddCommand(cmd); err != nil {
				t.Errorf("AddCommand failed: %v", err)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestAOFHandler_LoadWithSpecialCharacters(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "test.aof")

	db := database.MakeDB()
	handler, err := MakeAOFHandler(filename, db)
	if err != nil {
		t.Fatalf("MakeAOFHandler failed: %v", err)
	}

	// Add command with special characters
	handler.AddCommand([][]byte{[]byte("SET"), []byte("key with spaces"), []byte("value with \r\n characters")})

	// Close and reload
	handler.Close()

	db2 := database.MakeDB()
	handler2, err := MakeAOFHandler(filename, db2)
	if err != nil {
		t.Fatalf("MakeAOFHandler failed: %v", err)
	}
	defer handler2.Close()

	// Verify data
	result, err := db2.ExecCommand("GET", "key with spaces")
	if err != nil {
		t.Fatalf("GET failed: %v", err)
	}
	if len(result) == 0 || string(result[0]) != "value with \r\n characters" {
		t.Errorf("Expected 'value with \\r\\n characters', got %v", result)
	}
}

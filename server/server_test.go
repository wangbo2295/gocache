package server

import (
	"testing"
	
	"github.com/wangbo/gocache/database"
)

func TestMakeHandler(t *testing.T) {
	db := database.MakeDB()
	defer db.Close()
	
	handler := MakeHandler(db)
	if handler == nil {
		t.Fatal("Handler should not be nil")
	}
}

func TestMakeHandlerWithAOF(t *testing.T) {
	db := database.MakeDB()
	defer db.Close()
	
	handler := MakeHandlerWithAOF(db, nil)
	if handler == nil {
		t.Fatal("Handler should not be nil")
	}
}

func TestHandlerExecCommand(t *testing.T) {
	db := database.MakeDB()
	defer db.Close()
	
	handler := MakeHandler(db)
	
	// Test PING command
	_, err := handler.ExecCommand([][]byte{[]byte("PING")})
	if err != nil {
		t.Fatalf("ExecCommand PING failed: %v", err)
	}
}

func TestExecCommandSetGet(t *testing.T) {
	db := database.MakeDB()
	defer db.Close()
	
	handler := MakeHandler(db)
	
	// SET command
	_, err := handler.ExecCommand([][]byte{
		[]byte("SET"), []byte("key"), []byte("value"),
	})
	if err != nil {
		t.Fatalf("SET failed: %v", err)
	}
	
	// GET command
 getResult, err := handler.ExecCommand([][]byte{
		[]byte("GET"), []byte("key"),
	})
	if err != nil {
		t.Fatalf("GET failed: %v", err)
	}
	
	if getResult == nil {
		t.Fatal("Expected GET response")
	}
}

func TestExecCommandWithMulti(t *testing.T) {
	db := database.MakeDB()
	defer db.Close()
	
	handler := MakeHandler(db)
	
	// MULTI
	_, err := handler.ExecCommand([][]byte{[]byte("MULTI")})
	if err != nil {
		t.Fatalf("MULTI failed: %v", err)
	}
	
	// SET in transaction
	_, err = handler.ExecCommand([][]byte{
		[]byte("SET"), []byte("txkey"), []byte("txvalue"),
	})
	if err != nil {
		t.Fatalf("SET in MULTI failed: %v", err)
	}
	
	// EXEC
	_, err = handler.ExecCommand([][]byte{[]byte("EXEC")})
	if err != nil {
		t.Fatalf("EXEC failed: %v", err)
	}
	
	// Verify value
	getResult, err := handler.ExecCommand([][]byte{
		[]byte("GET"), []byte("txkey"),
	})
	if err != nil {
		t.Fatalf("GET failed: %v", err)
	}
	
	if getResult == nil {
		t.Fatal("Expected GET response")
	}
}

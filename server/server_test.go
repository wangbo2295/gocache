package server

import (
	"bufio"
	"net"
	"testing"
	"time"

	"github.com/wangbo/gocache/config"
	"github.com/wangbo/gocache/database"
)

func TestMakeServer(t *testing.T) {
	cfg := &config.Properties{
		Bind: "127.0.0.1",
		Port: 16379, // Use different port for testing
	}
	db := database.MakeDB()
	handler := MakeHandler(db)

	server := MakeServer(cfg, handler)
	if server == nil {
		t.Fatal("MakeServer returned nil")
	}
	if server.config != cfg {
		t.Error("Server config not set correctly")
	}
	if server.handler != handler {
		t.Error("Server handler not set correctly")
	}
}

func TestMakeHandler(t *testing.T) {
	db := database.MakeDB()
	handler := MakeHandler(db)

	if handler == nil {
		t.Fatal("MakeHandler returned nil")
	}
	if handler.db != db {
		t.Error("Handler db not set correctly")
	}
}

func TestHandler_ExecCommand_PING(t *testing.T) {
	db := database.MakeDB()
	handler := MakeHandler(db)

	// PING without arguments
	result, err := handler.ExecCommand([][]byte{[]byte("PING")})
	if err != nil {
		t.Fatalf("ExecCommand failed: %v", err)
	}
	if result.ToBytes() == nil {
		t.Error("Expected non-nil result")
	}

	// PING with argument
	result, err = handler.ExecCommand([][]byte{[]byte("PING"), []byte("hello")})
	if err != nil {
		t.Fatalf("ExecCommand failed: %v", err)
	}
	if result.ToBytes() == nil {
		t.Error("Expected non-nil result")
	}
}

func TestHandler_ExecCommand_SET_GET(t *testing.T) {
	db := database.MakeDB()
	handler := MakeHandler(db)

	// SET command
	result, err := handler.ExecCommand([][]byte{[]byte("SET"), []byte("key1"), []byte("value1")})
	if err != nil {
		t.Fatalf("SET failed: %v", err)
	}
	replyBytes := result.ToBytes()
	if string(replyBytes) != "+OK\r\n" {
		t.Errorf("Expected '+OK\\r\\n', got %q", string(replyBytes))
	}

	// GET command
	result, err = handler.ExecCommand([][]byte{[]byte("GET"), []byte("key1")})
	if err != nil {
		t.Fatalf("GET failed: %v", err)
	}
	replyBytes = result.ToBytes()
	if string(replyBytes) != "$6\r\nvalue1\r\n" {
		t.Errorf("Expected bulk reply with 'value1', got %q", string(replyBytes))
	}

	// GET non-existent key
	result, err = handler.ExecCommand([][]byte{[]byte("GET"), []byte("nonexistent")})
	if err != nil {
		t.Fatalf("GET failed: %v", err)
	}
	replyBytes = result.ToBytes()
	if string(replyBytes) != "$-1\r\n" {
		t.Errorf("Expected null bulk reply, got %q", string(replyBytes))
	}
}

func TestHandler_ExecCommand_DEL(t *testing.T) {
	db := database.MakeDB()
	handler := MakeHandler(db)

	// Set up a key
	handler.ExecCommand([][]byte{[]byte("SET"), []byte("key1"), []byte("value1")})

	// Delete the key
	result, err := handler.ExecCommand([][]byte{[]byte("DEL"), []byte("key1")})
	if err != nil {
		t.Fatalf("DEL failed: %v", err)
	}
	replyBytes := result.ToBytes()
	if string(replyBytes) != ":1\r\n" {
		t.Errorf("Expected ':1\\r\\n', got %q", string(replyBytes))
	}

	// Verify deletion
	result, err = handler.ExecCommand([][]byte{[]byte("GET"), []byte("key1")})
	if err != nil {
		t.Fatalf("GET failed: %v", err)
	}
	replyBytes = result.ToBytes()
	if string(replyBytes) != "$-1\r\n" {
		t.Errorf("Expected null bulk reply, got %q", string(replyBytes))
	}
}

func TestHandler_ExecCommand_INCR(t *testing.T) {
	db := database.MakeDB()
	handler := MakeHandler(db)

	// INCR non-existent key (starts from 0)
	result, err := handler.ExecCommand([][]byte{[]byte("INCR"), []byte("counter")})
	if err != nil {
		t.Fatalf("INCR failed: %v", err)
	}
	replyBytes := result.ToBytes()
	if string(replyBytes) != ":1\r\n" {
		t.Errorf("Expected ':1\\r\\n', got %q", string(replyBytes))
	}

	// INCR again
	result, err = handler.ExecCommand([][]byte{[]byte("INCR"), []byte("counter")})
	if err != nil {
		t.Fatalf("INCR failed: %v", err)
	}
	replyBytes = result.ToBytes()
	if string(replyBytes) != ":2\r\n" {
		t.Errorf("Expected ':2\\r\\n', got %q", string(replyBytes))
	}
}

func TestHandler_ExecCommand_Keys(t *testing.T) {
	db := database.MakeDB()
	handler := MakeHandler(db)

	// Set multiple keys
	handler.ExecCommand([][]byte{[]byte("SET"), []byte("key1"), []byte("value1")})
	handler.ExecCommand([][]byte{[]byte("SET"), []byte("key2"), []byte("value2")})

	// KEYS command
	result, err := handler.ExecCommand([][]byte{[]byte("KEYS"), []byte("*")})
	if err != nil {
		t.Fatalf("KEYS failed: %v", err)
	}
	replyBytes := result.ToBytes()
	// Should be array reply: *2\r\n$4\r\nkey1\r\n$4\r\nkey2\r\n
	if len(replyBytes) < 3 || replyBytes[0] != '*' {
		t.Errorf("Expected array reply, got %q", string(replyBytes))
	}
}

func TestIntegration_Server(t *testing.T) {
	cfg := &config.Properties{
		Bind: "127.0.0.1",
		Port: 16379,
	}
	db := database.MakeDB()
	handler := MakeHandler(db)
	server := MakeServer(cfg, handler)

	// Start server in background
	go func() {
		if err := server.Start(); err != nil {
			t.Logf("Server error: %v", err)
		}
	}()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)

	// Connect to server
	conn, err := net.Dial("tcp", "127.0.0.1:16379")
	if err != nil {
		t.Fatalf("Failed to connect to server: %v", err)
	}
	defer conn.Close()

	// Send PING command using inline command format
	conn.Write([]byte("PING\r\n"))

	// Read response
	reader := bufio.NewReader(conn)
	response, err := reader.ReadString('\n')
	if err != nil {
		t.Fatalf("Failed to read response: %v", err)
	}
	if response != "+PONG\r\n" {
		t.Errorf("Expected '+PONG\\r\\n', got %q", response)
	}

	// Stop server
	conn.Close()
	server.Stop()

	// Wait for server to stop
	time.Sleep(100 * time.Millisecond)
}

func TestIntegration_ConcurrentClients(t *testing.T) {
	cfg := &config.Properties{
		Bind: "127.0.0.1",
		Port: 16380,
	}
	db := database.MakeDB()
	handler := MakeHandler(db)
	server := MakeServer(cfg, handler)

	// Start server in background
	go func() {
		if err := server.Start(); err != nil {
			t.Logf("Server error: %v", err)
		}
	}()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)

	const numClients = 10
	const operationsPerClient = 100

	// Create multiple clients
	for i := 0; i < numClients; i++ {
		go func(clientID int) {
			conn, err := net.Dial("tcp", "127.0.0.1:16380")
			if err != nil {
				t.Errorf("Client %d: failed to connect: %v", clientID, err)
				return
			}
			defer conn.Close()

			for j := 0; j < operationsPerClient; j++ {
				// Send PING command
				conn.Write([]byte("PING\r\n"))

				// Read response
				reader := bufio.NewReader(conn)
				response, _ := reader.ReadString('\n')
				if response != "+PONG\r\n" {
					t.Errorf("Client %d: expected '+PONG\\r\\n', got %q", clientID, response)
					return
				}
			}
		}(i)
	}

	// Wait for all clients to finish
	time.Sleep(1 * time.Second)

	// Stop server
	server.Stop()
}

func TestHandler_ExecCommand_Empty(t *testing.T) {
	db := database.MakeDB()
	handler := MakeHandler(db)

	// Empty command
	_, err := handler.ExecCommand([][]byte{})
	if err == nil {
		t.Error("Expected error for empty command")
	}
}

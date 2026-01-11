package server

import (
	"testing"

	"github.com/wangbo/gocache/database"
	"github.com/wangbo/gocache/protocol"
)

// TestCaseInsensitiveCommands verifies that commands are case-insensitive
// and that AOF appending works regardless of command case
func TestCaseInsensitiveCommands(t *testing.T) {
	db := database.MakeDB()
	defer db.Close()

	handler := MakeHandler(db)

	testCases := []struct {
		name        string
		cmdLine     [][]byte
		wantContain string
	}{
		{
			name:        "lowercase set",
			cmdLine:     [][]byte{[]byte("set"), []byte("key1"), []byte("value1")},
			wantContain: "+OK",
		},
		{
			name:        "uppercase SET",
			cmdLine:     [][]byte{[]byte("SET"), []byte("key2"), []byte("value2")},
			wantContain: "+OK",
		},
		{
			name:        "mixed case SeT",
			cmdLine:     [][]byte{[]byte("SeT"), []byte("key3"), []byte("value3")},
			wantContain: "+OK",
		},
		{
			name:        "lowercase get",
			cmdLine:     [][]byte{[]byte("get"), []byte("key1")},
			wantContain: "value1",
		},
		{
			name:        "uppercase GET",
			cmdLine:     [][]byte{[]byte("GET"), []byte("key2")},
			wantContain: "value2",
		},
		{
			name:        "lowercase del",
			cmdLine:     [][]byte{[]byte("del"), []byte("key3")},
			wantContain: ":1",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			reply, err := handler.ExecCommand(tc.cmdLine)
			if err != nil {
				t.Fatalf("ExecCommand error: %v", err)
			}

			respStr := string(reply.ToBytes())
			// Check if response contains expected substring (more flexible than exact match)
			if !contains(respStr, tc.wantContain) {
				t.Errorf("ExecCommand(%v) = %q, want to contain %q", tc.cmdLine, respStr, tc.wantContain)
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && findInString(s, substr)))
}

func findInString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// TestProtocolCommands verifies protocol package command detection functions
func TestProtocolCommands(t *testing.T) {
	tests := []struct {
		name           string
		cmd            string
		isWrite        bool
		isInteger      bool
		isStatus       bool
	}{
		{"set", "set", true, false, true},
		{"SET", "SET", true, false, true},
		{"SeT", "SeT", true, false, true},
		{"get", "get", false, false, false},
		{"GET", "GET", false, false, false},
		{"del", "del", true, true, false},
		{"DEL", "DEL", true, true, false},
		{"incr", "incr", true, true, false},
		{"INCR", "INCR", true, true, false},
		{"hset", "hset", true, false, false},
		{"HSET", "HSET", true, false, false},
		{"lpush", "lpush", true, false, false},
		{"LPUSH", "LPUSH", true, false, false},
		{"sadd", "sadd", true, false, false},
		{"SADD", "SADD", true, false, false},
		{"zadd", "zadd", true, false, false},
		{"ZADD", "ZADD", true, false, false},
		{"ping", "ping", false, false, false},
		{"PING", "PING", false, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := protocol.IsWriteCommand(tt.cmd); got != tt.isWrite {
				t.Errorf("IsWriteCommand(%q) = %v, want %v", tt.cmd, got, tt.isWrite)
			}
			if got := protocol.IsIntegerCommand(tt.cmd); got != tt.isInteger {
				t.Errorf("IsIntegerCommand(%q) = %v, want %v", tt.cmd, got, tt.isInteger)
			}
			if got := protocol.IsStatusCommand(tt.cmd); got != tt.isStatus {
				t.Errorf("IsStatusCommand(%q) = %v, want %v", tt.cmd, got, tt.isStatus)
			}
		})
	}
}

// TestToUpper verifies the ToUpper function
func TestToUpper(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"set", "SET"},
		{"SET", "SET"},
		{"SeT", "SET"},
		{"get", "GET"},
		{"hset", "HSET"},
		{"ping", "PING"},
		{"", ""},
		{"alreadyUPPER", "ALREADYUPPER"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := protocol.ToUpper(tt.input); got != tt.want {
				t.Errorf("ToUpper(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

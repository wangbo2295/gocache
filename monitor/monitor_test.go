package monitor

import (
	"bytes"
	"net"
	"sync"
	"testing"
	"time"
)

// MockConn implements net.Conn for testing
type MockConn struct {
	net.Conn
	readBuffer  bytes.Buffer
	writeBuffer bytes.Buffer
	closed      bool
	mu          sync.Mutex
}

func (m *MockConn) Read(b []byte) (n int, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.readBuffer.Read(b)
}

func (m *MockConn) Write(b []byte) (n int, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed {
		return 0, net.ErrClosed
	}
	return m.writeBuffer.Write(b)
}

func (m *MockConn) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closed = true
	return nil
}

func (m *MockConn) LocalAddr() net.Addr {
	return &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 6379}
}

func (m *MockConn) RemoteAddr() net.Addr {
	return &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 12345}
}

func (m *MockConn) SetDeadline(t time.Time) error {
	return nil
}

func (m *MockConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (m *MockConn) SetWriteDeadline(t time.Time) error {
	return nil
}

func (m *MockConn) GetWrittenData() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.writeBuffer.String()
}

func (m *MockConn) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.writeBuffer.Reset()
}

func TestGetMonitor(t *testing.T) {
	monitor := GetMonitor()

	if monitor == nil {
		t.Fatal("GetMonitor returned nil")
	}

	// Should return the same instance
	monitor2 := GetMonitor()
	if monitor != monitor2 {
		t.Error("GetMonitor should return the same global instance")
	}
}

func TestAddClient(t *testing.T) {
	monitor := &Monitor{
		clients:   make([]net.Conn, 0),
		enabled:   false,
		monitorCh: make(chan *MonitoredCommand, 1000),
	}

	client1 := &MockConn{}
	monitor.AddClient(client1)

	time.Sleep(100 * time.Millisecond) // Give time for goroutine to start

	if !monitor.enabled {
		t.Error("Monitor should be enabled after adding first client")
	}

	monitor.clientsMu.RLock()
	if len(monitor.clients) != 1 {
		t.Errorf("Expected 1 client, got %d", len(monitor.clients))
	}
	monitor.clientsMu.RUnlock()

	// Add second client
	client2 := &MockConn{}
	monitor.AddClient(client2)

	monitor.clientsMu.RLock()
	if len(monitor.clients) != 2 {
		t.Errorf("Expected 2 clients, got %d", len(monitor.clients))
	}
	monitor.clientsMu.RUnlock()
}

func TestRemoveClient(t *testing.T) {
	monitor := &Monitor{
		clients:   make([]net.Conn, 0),
		enabled:   false,
		monitorCh: make(chan *MonitoredCommand, 1000),
	}

	client1 := &MockConn{}
	client2 := &MockConn{}

	monitor.AddClient(client1)
	monitor.AddClient(client2)

	time.Sleep(100 * time.Millisecond)

	monitor.RemoveClient(client1)

	monitor.clientsMu.RLock()
	if len(monitor.clients) != 1 {
		t.Errorf("Expected 1 client after removal, got %d", len(monitor.clients))
	}
	monitor.clientsMu.RUnlock()

	// Remove second client - should disable monitoring
	monitor.RemoveClient(client2)

	time.Sleep(100 * time.Millisecond)

	if monitor.enabled {
		t.Error("Monitor should be disabled when no clients")
	}

	monitor.clientsMu.RLock()
	if len(monitor.clients) != 0 {
		t.Errorf("Expected 0 clients, got %d", len(monitor.clients))
	}
	monitor.clientsMu.RUnlock()
}

func TestLogCommand(t *testing.T) {
	monitor := &Monitor{
		clients:   make([]net.Conn, 0),
		enabled:   false,
		monitorCh: make(chan *MonitoredCommand, 1000),
	}

	// Log when not enabled - should not panic
	cmdLine := [][]byte{[]byte("SET"), []byte("key"), []byte("value")}
	monitor.LogCommand(cmdLine, "127.0.0.1:12345")

	// Channel should be empty
	select {
	case <-monitor.monitorCh:
		t.Error("Should not receive command when monitor is disabled")
	default:
		// Expected
	}

	// Enable and log
	client := &MockConn{}
	monitor.AddClient(client)

	time.Sleep(100 * time.Millisecond)

	monitor.LogCommand(cmdLine, "127.0.0.1:12345")

	// Give time for broadcast
	time.Sleep(200 * time.Millisecond)

	data := client.GetWrittenData()
	if len(data) == 0 {
		t.Error("Client should have received command")
	}
}

func TestSerializeCommand(t *testing.T) {
	tests := []struct {
		name     string
		cmdLine  [][]byte
		expected string
	}{
		{
			name:     "empty command",
			cmdLine:  [][]byte{},
			expected: "",
		},
		{
			name:     "simple SET command",
			cmdLine:  [][]byte{[]byte("SET"), []byte("key"), []byte("value")},
			expected: "SET key value",
		},
		{
			name:     "GET command",
			cmdLine:  [][]byte{[]byte("GET"), []byte("mykey")},
			expected: "GET mykey",
		},
		{
			name:     "command with spaces",
			cmdLine:  [][]byte{[]byte("SET"), []byte(" key with spaces "), []byte("value")},
			expected: `SET " key with spaces " value`,
		},
		{
			name:     "command with empty argument",
			cmdLine:  [][]byte{[]byte("SET"), []byte(""), []byte("value")},
			expected: `SET "" value`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := serializeCommand(tt.cmdLine)
			if result != tt.expected {
				t.Errorf("serializeCommand() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestBroadcastLoop(t *testing.T) {
	monitor := &Monitor{
		clients:   make([]net.Conn, 0),
		enabled:   false,
		monitorCh: make(chan *MonitoredCommand, 1000),
	}

	client1 := &MockConn{}
	client2 := &MockConn{}

	monitor.AddClient(client1)
	monitor.AddClient(client2)

	time.Sleep(100 * time.Millisecond)

	// Log a command
	cmdLine := [][]byte{[]byte("SET"), []byte("key"), []byte("value")}
	monitor.LogCommand(cmdLine, "127.0.0.1:12345")

	// Wait for broadcast
	time.Sleep(200 * time.Millisecond)

	// Check both clients received the command
	data1 := client1.GetWrittenData()
	data2 := client2.GetWrittenData()

	if len(data1) == 0 {
		t.Error("Client1 should have received command")
	}

	if len(data2) == 0 {
		t.Error("Client2 should have received command")
	}

	if data1 != data2 {
		t.Error("Both clients should receive the same command")
	}

	// Verify format - should start with timestamp
	if len(data1) < 10 {
		t.Errorf("Unexpected command format, too short: %s", data1)
	}
	// Just verify it contains expected elements
	if len(data1) == 0 {
		t.Error("Should have received data")
	}
}

func TestMultipleCommands(t *testing.T) {
	monitor := &Monitor{
		clients:   make([]net.Conn, 0),
		enabled:   false,
		monitorCh: make(chan *MonitoredCommand, 1000),
	}

	client := &MockConn{}
	monitor.AddClient(client)

	time.Sleep(100 * time.Millisecond)

	commands := [][][]byte{
		{[]byte("SET"), []byte("key1"), []byte("value1")},
		{[]byte("GET"), []byte("key1")},
		{[]byte("SET"), []byte("key2"), []byte("value2")},
		{[]byte("DEL"), []byte("key1")},
	}

	for _, cmd := range commands {
		monitor.LogCommand(cmd, "127.0.0.1:12345")
	}

	// Wait for all broadcasts
	time.Sleep(500 * time.Millisecond)

	data := client.GetWrittenData()
	if len(data) == 0 {
		t.Fatal("Client should have received commands")
	}

	// Should have all 4 commands
	lines := 0
	for i := 0; i < len(data); i++ {
		if data[i] == '\n' {
			lines++
		}
	}

	if lines < 4 {
		t.Errorf("Expected at least 4 command lines, got %d", lines)
	}
}

func TestClientWithWriteError(t *testing.T) {
	monitor := &Monitor{
		clients:   make([]net.Conn, 0),
		enabled:   false,
		monitorCh: make(chan *MonitoredCommand, 1000),
	}

	// Create a client that will be closed
	errorClient := &MockConn{}
	goodClient := &MockConn{}

	monitor.AddClient(errorClient)
	monitor.AddClient(goodClient)

	time.Sleep(100 * time.Millisecond)

	// Close the error client
	errorClient.Close()

	// Log a command - should remove error client
	cmdLine := [][]byte{[]byte("SET"), []byte("key"), []byte("value")}
	monitor.LogCommand(cmdLine, "127.0.0.1:12345")

	time.Sleep(200 * time.Millisecond)

	// Check that error client was removed
	monitor.clientsMu.RLock()
	clientCount := len(monitor.clients)
	monitor.clientsMu.RUnlock()

	if clientCount != 1 {
		t.Errorf("Expected 1 client after error, got %d", clientCount)
	}

	// Good client should still receive data
	data := goodClient.GetWrittenData()
	if len(data) == 0 {
		t.Error("Good client should have received command")
	}
}

func TestConcurrentClientAccess(t *testing.T) {
	monitor := &Monitor{
		clients:   make([]net.Conn, 0),
		enabled:   false,
		monitorCh: make(chan *MonitoredCommand, 1000),
	}

	var wg sync.WaitGroup

	// Concurrently add clients
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			client := &MockConn{}
			monitor.AddClient(client)
		}()
	}

	// Concurrently log commands
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			cmdLine := [][]byte{[]byte("SET"), []byte("key"), []byte("value")}
			monitor.LogCommand(cmdLine, "127.0.0.1:12345")
		}(i)
	}

	wg.Wait()

	// Should have some clients
	monitor.clientsMu.RLock()
	clientCount := len(monitor.clients)
	monitor.clientsMu.RUnlock()

	if clientCount == 0 {
		t.Error("Should have added some clients")
	}

	if clientCount > 10 {
		t.Errorf("Should have at most 10 clients, got %d", clientCount)
	}
}

func TestChannelFullBehavior(t *testing.T) {
	monitor := &Monitor{
		clients:   make([]net.Conn, 0),
		enabled:   false,
		monitorCh: make(chan *MonitoredCommand, 2), // Small buffer
	}

	client := &MockConn{}
	monitor.AddClient(client)

	time.Sleep(100 * time.Millisecond)

	// Fill the channel
	for i := 0; i < 10; i++ {
		cmdLine := [][]byte{[]byte("SET"), []byte("key"), []byte("value")}
		monitor.LogCommand(cmdLine, "127.0.0.1:12345")
	}

	// Should not block or panic
	time.Sleep(200 * time.Millisecond)

	// At least some commands should be processed
	data := client.GetWrittenData()
	if len(data) == 0 {
		t.Error("Should have processed some commands")
	}
}

func TestMonitoredCommand(t *testing.T) {
	cmdLine := [][]byte{[]byte("SET"), []byte("key"), []byte("value")}
	cmd := &MonitoredCommand{
		Timestamp: time.Now(),
		Command:   serializeCommand(cmdLine),
		Client:    "127.0.0.1:12345",
	}

	if cmd.Command != "SET key value" {
		t.Errorf("Unexpected command: %s", cmd.Command)
	}

	if cmd.Client != "127.0.0.1:12345" {
		t.Errorf("Unexpected client: %s", cmd.Client)
	}

	if time.Since(cmd.Timestamp) > time.Second {
		t.Error("Timestamp should be recent")
	}
}

func TestRemoveNonExistentClient(t *testing.T) {
	monitor := &Monitor{
		clients:   make([]net.Conn, 0),
		enabled:   false,
		monitorCh: make(chan *MonitoredCommand, 1000),
	}

	client1 := &MockConn{}
	client2 := &MockConn{}
	client3 := &MockConn{}

	monitor.AddClient(client1)
	monitor.AddClient(client2)

	time.Sleep(100 * time.Millisecond)

	// Try to remove client3 which was never added
	monitor.RemoveClient(client3)

	monitor.clientsMu.RLock()
	clientCount := len(monitor.clients)
	monitor.clientsMu.RUnlock()

	if clientCount != 2 {
		t.Errorf("Should still have 2 clients, got %d", clientCount)
	}
}

func TestGlobalMonitor(t *testing.T) {
	// Test that global monitor can be used safely
	globalMonitor := GetMonitor()

	client := &MockConn{}
	globalMonitor.AddClient(client)

	time.Sleep(100 * time.Millisecond)

	cmdLine := [][]byte{[]byte("PING")}
	globalMonitor.LogCommand(cmdLine, "127.0.0.1:9999")

	time.Sleep(200 * time.Millisecond)

	data := client.GetWrittenData()
	if len(data) == 0 {
		t.Error("Global monitor should work")
	}

	// Cleanup
	globalMonitor.RemoveClient(client)
}

package replication

import (
	"bufio"
	"bytes"
	"io"
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
	readDelay   time.Duration
	writeDelay  time.Duration
}

func (m *MockConn) Read(b []byte) (n int, err error) {
	if m.readDelay > 0 {
		time.Sleep(m.readDelay)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed {
		return 0, io.EOF
	}
	return m.readBuffer.Read(b)
}

func (m *MockConn) Write(b []byte) (n int, err error) {
	if m.writeDelay > 0 {
		time.Sleep(m.writeDelay)
	}
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
	return &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 6379}
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

func (m *MockConn) WriteString(s string) (int, error) {
	return m.Write([]byte(s))
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
	m.readBuffer.Reset()
}

func TestReplicationRole_String(t *testing.T) {
	tests := []struct {
		name     string
		role     ReplicationRole
		expected string
	}{
		{"master", RoleMaster, "master"},
		{"slave", RoleSlave, "slave"},
		{"unknown", ReplicationRole(99), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.role.String(); got != tt.expected {
				t.Errorf("ReplicationRole.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestReplicationState_IsMaster(t *testing.T) {
	rs := &ReplicationState{
		role: RoleMaster,
	}

	if !rs.IsMaster() {
		t.Error("Should be master")
	}

	rs.role = RoleSlave
	if rs.IsMaster() {
		t.Error("Should not be master")
	}
}

func TestReplicationState_IsSlave(t *testing.T) {
	rs := &ReplicationState{
		role: RoleSlave,
	}

	if !rs.IsSlave() {
		t.Error("Should be slave")
	}

	rs.role = RoleMaster
	if rs.IsSlave() {
		t.Error("Should not be slave")
	}
}

func TestReplicationState_GetRole(t *testing.T) {
	rs := &ReplicationState{
		role: RoleMaster,
	}

	if rs.GetRole() != RoleMaster {
		t.Error("Should return RoleMaster")
	}

	rs.role = RoleSlave
	if rs.GetRole() != RoleSlave {
		t.Error("Should return RoleSlave")
	}
}

func TestReplicationState_GetMasterInfo(t *testing.T) {
	rs := &ReplicationState{
		masterHost: "localhost",
		masterPort: 6379,
	}

	host, port := rs.GetMasterInfo()
	if host != "localhost" {
		t.Errorf("Expected host 'localhost', got '%s'", host)
	}

	if port != 6379 {
		t.Errorf("Expected port 6379, got %d", port)
	}
}

func TestReplicationState_GetReplicationID(t *testing.T) {
	rs := &ReplicationState{
		replID: 12345,
	}

	if rs.GetReplicationID() != 12345 {
		t.Errorf("Expected replID 12345, got %d", rs.GetReplicationID())
	}
}

func TestReplicationState_GetReplicationOffset(t *testing.T) {
	rs := &ReplicationState{
		replOffset: 9999,
	}

	if rs.GetReplicationOffset() != 9999 {
		t.Errorf("Expected offset 9999, got %d", rs.GetReplicationOffset())
	}
}

func TestReplicationState_IncrementReplicationOffset(t *testing.T) {
	rs := &ReplicationState{
		replOffset: 100,
	}

	rs.IncrementReplicationOffset(50)
	if rs.GetReplicationOffset() != 150 {
		t.Errorf("Expected offset 150, got %d", rs.GetReplicationOffset())
	}

	rs.IncrementReplicationOffset(25)
	if rs.GetReplicationOffset() != 175 {
		t.Errorf("Expected offset 175, got %d", rs.GetReplicationOffset())
	}
}

func TestReplicationState_SetAsSlave(t *testing.T) {
	rs := &ReplicationState{
		role:   RoleMaster,
		replID: 1,
	}

	err := rs.SetAsSlave("localhost", 6380)
	if err != nil {
		t.Fatalf("SetAsSlave failed: %v", err)
	}

	if rs.role != RoleSlave {
		t.Error("Should be slave after SetAsSlave")
	}

	if rs.masterHost != "localhost" {
		t.Errorf("Expected masterHost 'localhost', got '%s'", rs.masterHost)
	}

	if rs.masterPort != 6380 {
		t.Errorf("Expected masterPort 6380, got %d", rs.masterPort)
	}

	if rs.replID != 0 {
		t.Errorf("Slave should have replID 0, got %d", rs.replID)
	}
}

func TestReplicationState_SetAsMaster(t *testing.T) {
	rs := &ReplicationState{
		role:       RoleSlave,
		masterHost: "localhost",
		masterPort: 6380,
		replID:     0,
	}

	rs.SetAsMaster()

	if rs.role != RoleMaster {
		t.Error("Should be master after SetAsMaster")
	}

	if rs.masterHost != "" {
		t.Errorf("Master should not have masterHost, got '%s'", rs.masterHost)
	}

	if rs.masterPort != 0 {
		t.Errorf("Master should not have masterPort, got %d", rs.masterPort)
	}

	if rs.replID != 1 {
		t.Errorf("Master should have replID 1, got %d", rs.replID)
	}
}

func TestReplicationState_ConnectToMaster_NotSlave(t *testing.T) {
	rs := &ReplicationState{
		role: RoleMaster,
	}

	err := rs.ConnectToMaster()
	if err == nil {
		t.Error("Should return error when not slave")
	}

	if err.Error() != "not configured as slave" {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestReplicationState_ConnectToMaster_NoMasterConfigured(t *testing.T) {
	rs := &ReplicationState{
		role:       RoleSlave,
		masterHost: "",
		masterPort: 0,
	}

	err := rs.ConnectToMaster()
	if err == nil {
		t.Error("Should return error when no master configured")
	}

	if err.Error() != "no master configured" {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestReplicationState_DisconnectFromMaster(t *testing.T) {
	rs := &ReplicationState{
		role:       RoleSlave,
		masterConn: &MockConn{},
	}

	err := rs.DisconnectFromMaster()
	if err != nil {
		t.Fatalf("DisconnectFromMaster failed: %v", err)
	}

	if rs.masterConn != nil {
		t.Error("masterConn should be nil after disconnect")
	}
}

func TestReplicationState_DisconnectFromMaster_NoConnection(t *testing.T) {
	rs := &ReplicationState{
		role:       RoleSlave,
		masterConn: nil,
	}

	err := rs.DisconnectFromMaster()
	if err != nil {
		t.Fatalf("DisconnectFromMaster should not error when no connection: %v", err)
	}
}

func TestReplicationState_SendPSync_NotConnected(t *testing.T) {
	rs := &ReplicationState{
		role:       RoleSlave,
		masterConn: nil,
	}

	err := rs.SendPSync(1, 100)
	if err == nil {
		t.Error("Should return error when not connected")
	}

	if err.Error() != "not connected to master" {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestReplicationState_SendSync_NotConnected(t *testing.T) {
	rs := &ReplicationState{
		role:       RoleSlave,
		masterConn: nil,
	}

	err := rs.SendSync()
	if err == nil {
		t.Error("Should return error when not connected")
	}

	if err.Error() != "not connected to master" {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestReplicationState_SendPSync_Connected(t *testing.T) {
	conn := &MockConn{}
	rs := &ReplicationState{
		role:       RoleSlave,
		masterConn: conn,
	}

	err := rs.SendPSync(123, 456)
	if err != nil {
		t.Fatalf("SendPSync failed: %v", err)
	}

	data := conn.GetWrittenData()
	expected := "PSYNC 123 456\r\n"
	if data != expected {
		t.Errorf("Expected %q, got %q", expected, data)
	}
}

func TestReplicationState_SendSync_Connected(t *testing.T) {
	conn := &MockConn{}
	rs := &ReplicationState{
		role:       RoleSlave,
		masterConn: conn,
	}

	err := rs.SendSync()
	if err != nil {
		t.Fatalf("SendSync failed: %v", err)
	}

	data := conn.GetWrittenData()
	expected := "SYNC\r\n"
	if data != expected {
		t.Errorf("Expected %q, got %q", expected, data)
	}
}

func TestReplicationState_ReceiveSyncResponse_NotConnected(t *testing.T) {
	rs := &ReplicationState{
		masterConn: nil,
	}

	_, err := rs.ReceiveSyncResponse()
	if err == nil {
		t.Error("Should return error when not connected")
	}

	if err.Error() != "not connected to master" {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestReplicationState_ReceiveSyncResponse_InvalidResponse(t *testing.T) {
	conn := &MockConn{}
	conn.mu.Lock()
	conn.readBuffer.WriteString("-ERROR invalid response\r\n")
	conn.mu.Unlock()

	rs := &ReplicationState{
		masterConn: conn,
	}

	_, err := rs.ReceiveSyncResponse()
	if err == nil {
		t.Error("Should return error for invalid response")
	}
}

func TestReplicationState_ReceiveSyncResponse_Valid(t *testing.T) {
	conn := &MockConn{}
	// Write a valid SYNC response to read buffer
	response := "+FULLRESYNC 123 456\r\n$10\r\n0123456789\r\n"
	conn.mu.Lock()
	conn.readBuffer.WriteString(response)
	conn.mu.Unlock()

	rs := &ReplicationState{
		masterConn: conn,
	}

	data, err := rs.ReceiveSyncResponse()
	if err != nil {
		t.Fatalf("ReceiveSyncResponse failed: %v", err)
	}

	if len(data) != 10 {
		t.Errorf("Expected 10 bytes, got %d", len(data))
	}

	if string(data) != "0123456789" {
		t.Errorf("Unexpected data: %s", string(data))
	}

	if rs.replID != 123 {
		t.Errorf("Expected replID 123, got %d", rs.replID)
	}

	if rs.replOffset != 456 {
		t.Errorf("Expected offset 456, got %d", rs.replOffset)
	}
}

func TestReplicationState_RegisterSlave(t *testing.T) {
	rs := &ReplicationState{
		slaveConns: make([]net.Conn, 0),
	}

	conn1 := &MockConn{}
	conn2 := &MockConn{}

	rs.RegisterSlave(conn1)
	rs.RegisterSlave(conn2)

	if rs.GetSlaveCount() != 2 {
		t.Errorf("Expected 2 slaves, got %d", rs.GetSlaveCount())
	}
}

func TestReplicationState_UnregisterSlave(t *testing.T) {
	rs := &ReplicationState{
		slaveConns: make([]net.Conn, 0),
	}

	conn1 := &MockConn{}
	conn2 := &MockConn{}

	rs.RegisterSlave(conn1)
	rs.RegisterSlave(conn2)

	rs.UnregisterSlave(conn1)

	if rs.GetSlaveCount() != 1 {
		t.Errorf("Expected 1 slave after unregister, got %d", rs.GetSlaveCount())
	}
}

func TestReplicationState_PropagateCommand_NotMaster(t *testing.T) {
	rs := &ReplicationState{
		role:       RoleSlave,
		slaveConns: make([]net.Conn, 0),
	}

	cmd := [][]byte{[]byte("SET"), []byte("key"), []byte("value")}
	err := rs.PropagateCommand(cmd)
	if err != nil {
		t.Fatalf("PropagateCommand should not error when not master: %v", err)
	}
}

func TestReplicationState_PropagateCommand_MasterNoSlaves(t *testing.T) {
	rs := &ReplicationState{
		role:       RoleMaster,
		slaveConns: make([]net.Conn, 0),
	}

	cmd := [][]byte{[]byte("SET"), []byte("key"), []byte("value")}
	err := rs.PropagateCommand(cmd)
	if err != nil {
		t.Fatalf("PropagateCommand failed: %v", err)
	}
}

func TestReplicationState_AddToBacklog(t *testing.T) {
	rs := &ReplicationState{
		replicationBacklog: make([]byte, 0, 100),
		backlogSize:       100,
		replOffset:        0,
	}

	// Add data
	rs.addToBacklog([]byte("command1"))
	rs.addToBacklog([]byte("command2"))

	if len(rs.replicationBacklog) != 16 { // "command1" (8) + "command2" (8)
		t.Errorf("Expected backlog length 16, got %d", len(rs.replicationBacklog))
	}
}

func TestReplicationState_AddToBacklog_Trim(t *testing.T) {
	rs := &ReplicationState{
		replicationBacklog: make([]byte, 0),
		backlogSize:       10,
		replOffset:        0,
	}

	// Add data that exceeds backlog size
	rs.addToBacklog([]byte("12345678901")) // 11 bytes

	// Should be trimmed to 10 bytes
	if len(rs.replicationBacklog) > 10 {
		t.Errorf("Backlog should be trimmed to 10 bytes, got %d", len(rs.replicationBacklog))
	}
}

func TestReplicationState_GetBacklogData(t *testing.T) {
	rs := &ReplicationState{
		replicationBacklog: []byte("abcdefghij"),
		backlogSize:       100,
		replOffset:        100,
	}

	// Request data from offset 95 (should return last 5 bytes)
	data, err := rs.GetBacklogData(95)
	if err != nil {
		t.Fatalf("GetBacklogData failed: %v", err)
	}

	if string(data) != "fghij" {
		t.Errorf("Expected 'fghij', got '%s'", string(data))
	}
}

func TestReplicationState_GetBacklogData_OffsetInFuture(t *testing.T) {
	rs := &ReplicationState{
		replicationBacklog: []byte("abc"),
		backlogSize:       100,
		replOffset:        100,
	}

	_, err := rs.GetBacklogData(200)
	if err == nil {
		t.Error("Should return error when offset is in future")
	}
}

func TestReplicationState_GetBacklogData_OffsetTooOld(t *testing.T) {
	rs := &ReplicationState{
		replicationBacklog: []byte("abc"),
		backlogSize:       100,
		replOffset:        100,
	}

	_, err := rs.GetBacklogData(50)
	if err == nil {
		t.Error("Should return error when offset is too old")
	}
}

func TestReplicationState_SetBacklogSize(t *testing.T) {
	rs := &ReplicationState{
		replicationBacklog: []byte("abcdefghij"),
		backlogSize:       100,
	}

	rs.SetBacklogSize(5)

	if rs.GetBacklogSize() != 5 {
		t.Errorf("Expected backlog size 5, got %d", rs.GetBacklogSize())
	}

	// Backlog should be trimmed
	if len(rs.replicationBacklog) > 5 {
		t.Errorf("Backlog should be trimmed to 5 bytes, got %d", len(rs.replicationBacklog))
	}
}

func TestSerializeCommand(t *testing.T) {
	tests := []struct {
		name     string
		cmdLine  [][]byte
		expected string
	}{
		{
			name:     "SET command",
			cmdLine:  [][]byte{[]byte("SET"), []byte("key"), []byte("value")},
			expected: "*3\r\n$3\r\nSET\r\n$3\r\nkey\r\n$5\r\nvalue\r\n",
		},
		{
			name:     "GET command",
			cmdLine:  [][]byte{[]byte("GET"), []byte("mykey")},
			expected: "*2\r\n$3\r\nGET\r\n$5\r\nmykey\r\n",
		},
		{
			name:     "empty command",
			cmdLine:  [][]byte{},
			expected: "*0\r\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := serializeCommand(tt.cmdLine)
			if string(result) != tt.expected {
				t.Errorf("serializeCommand() = %q, want %q", string(result), tt.expected)
			}
		})
	}
}

func TestNewDBCommandAdapter(t *testing.T) {
	adapter := NewDBCommandAdapter(nil)
	if adapter == nil {
		t.Fatal("NewDBCommandAdapter returned nil")
	}

	if adapter.db != nil {
		t.Error("db should be nil")
	}
}

func TestDBCommandAdapter_ExecCommand_NoExecMethod(t *testing.T) {
	adapter := NewDBCommandAdapter("not a database")

	cmd := [][]byte{[]byte("PING")}
	_, err := adapter.ExecCommand(cmd)
	if err == nil {
		t.Error("Should return error when db doesn't implement Exec")
	}

	if err.Error() != "database does not implement Exec method" {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestReplicationState_StartReplicationLoop_NotSlave(t *testing.T) {
	rs := &ReplicationState{
		role: RoleMaster,
	}

	err := rs.StartReplicationLoop(nil)
	if err == nil {
		t.Error("Should return error when not slave")
	}

	if err.Error() != "not configured as slave" {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestReplicationState_StartReplicationLoop_NotConnected(t *testing.T) {
	rs := &ReplicationState{
		role:       RoleSlave,
		masterConn: nil,
	}

	err := rs.StartReplicationLoop(nil)
	if err == nil {
		t.Error("Should return error when not connected")
	}

	if err.Error() != "not connected to master" {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestReplicationState_readCommand_InvalidType(t *testing.T) {
	rs := &ReplicationState{}
	buf := bytes.NewBufferString("+PING\r\n")

	reader := bufio.NewReader(buf)
	_, err := rs.readCommand(reader)
	if err == nil {
		t.Error("Should return error for unsupported RESP type")
	}
}

func TestReplicationState_readCommand_EOF(t *testing.T) {
	rs := &ReplicationState{}
	buf := bytes.NewBufferString("")

	reader := bufio.NewReader(buf)
	_, err := rs.readCommand(reader)
	if err != io.EOF {
		t.Errorf("Expected EOF, got %v", err)
	}
}

func TestReplicationState_readCommand_ValidArray(t *testing.T) {
	rs := &ReplicationState{}
	buf := bytes.NewBufferString("*2\r\n$3\r\nGET\r\n$3\r\nkey\r\n")

	reader := bufio.NewReader(buf)
	cmdLine, err := rs.readCommand(reader)
	if err != nil {
		t.Fatalf("readCommand failed: %v", err)
	}

	if len(cmdLine) != 2 {
		t.Errorf("Expected 2 args, got %d", len(cmdLine))
	}

	if string(cmdLine[0]) != "GET" {
		t.Errorf("Expected 'GET', got '%s'", string(cmdLine[0]))
	}

	if string(cmdLine[1]) != "key" {
		t.Errorf("Expected 'key', got '%s'", string(cmdLine[1]))
	}
}

func TestReplicationState_readCommand_NullArray(t *testing.T) {
	rs := &ReplicationState{}
	buf := bytes.NewBufferString("*-1\r\n")

	reader := bufio.NewReader(buf)
	cmdLine, err := rs.readCommand(reader)
	if err != nil {
		t.Fatalf("readCommand failed: %v", err)
	}

	if cmdLine != nil {
		t.Errorf("Expected nil for null array, got %v", cmdLine)
	}
}

func TestRegisterRDBLoader(t *testing.T) {
	loader := &mockRDBLoader{}
	RegisterRDBLoader(loader)

	// Can't easily test the registered loader without accessing private var
	// Just verify it doesn't panic
	RegisterRDBLoader(nil)
}

func TestLoadRDBData_NoLoader(t *testing.T) {
	// Clear the loader first
	RegisterRDBLoader(nil)

	err := LoadRDBData(nil, []byte("data"))
	if err == nil {
		t.Error("Should return error when no loader registered")
	}

	if err.Error() != "no RDB loader registered" {
		t.Errorf("Unexpected error: %v", err)
	}
}

// mockRDBLoader implements RDBLoader for testing
type mockRDBLoader struct{}

func (m *mockRDBLoader) LoadRDBFromBytes(db interface{}, data []byte) error {
	return nil
}

func TestGlobalState(t *testing.T) {
	// Test that global State is accessible
	if State == nil {
		t.Error("Global State should not be nil")
	}

	if !State.IsMaster() {
		t.Error("Global state should default to master")
	}

	if State.GetReplicationID() != 1 {
		t.Errorf("Expected default replID 1, got %d", State.GetReplicationID())
	}
}

func TestConcurrentAccess(t *testing.T) {
	rs := &ReplicationState{
		role:              RoleMaster,
		replicationBacklog: make([]byte, 0),
		backlogSize:       1000,
		slaveConns:        make([]net.Conn, 0),
	}

	var wg sync.WaitGroup

	// Concurrently increment offset
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			rs.IncrementReplicationOffset(1)
		}()
	}

	// Concurrently add to backlog
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			rs.addToBacklog([]byte("data"))
		}()
	}

	wg.Wait()

	// Verify final state
	if rs.GetReplicationOffset() != 100 {
		t.Errorf("Expected offset 100, got %d", rs.GetReplicationOffset())
	}
}

func TestBacklogEdgeCases(t *testing.T) {
	rs := &ReplicationState{
		replicationBacklog: make([]byte, 0),
		backlogSize:       10,
		replOffset:        100,
	}

	// Add exactly 10 bytes
	rs.addToBacklog([]byte("0123456789"))

	// Should be exactly at limit
	if len(rs.replicationBacklog) != 10 {
		t.Errorf("Expected backlog size 10, got %d", len(rs.replicationBacklog))
	}

	// Add one more byte
	rs.addToBacklog([]byte("x"))

	// Should still be at limit
	if len(rs.replicationBacklog) > 10 {
		t.Errorf("Backlog should not exceed limit")
	}
}

func TestPropagateCommandWithSlaves(t *testing.T) {
	rs := &ReplicationState{
		role:              RoleMaster,
		slaveConns:        make([]net.Conn, 0),
		replicationBacklog: make([]byte, 0),
		backlogSize:       1000,
		replOffset:        0,
	}

	// Add mock slaves
	slave1 := &MockConn{}
	slave2 := &MockConn{}

	rs.RegisterSlave(slave1)
	rs.RegisterSlave(slave2)

	cmd := [][]byte{[]byte("SET"), []byte("key"), []byte("value")}
	err := rs.PropagateCommand(cmd)
	if err != nil {
		t.Fatalf("PropagateCommand failed: %v", err)
	}

	// Give goroutines time to write
	time.Sleep(100 * time.Millisecond)

	// Check that both slaves received data
	data1 := slave1.GetWrittenData()
	data2 := slave2.GetWrittenData()

	if len(data1) == 0 {
		t.Error("Slave1 should have received command")
	}

	if len(data2) == 0 {
		t.Error("Slave2 should have received command")
	}

	// Both should receive the same data
	if data1 != data2 {
		t.Error("Both slaves should receive the same command")
	}
}

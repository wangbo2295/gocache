package replication

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

// ReplicationRole defines the role of the instance
type ReplicationRole int

const (
	RoleMaster ReplicationRole = iota
	RoleSlave
)

func (r ReplicationRole) String() string {
	switch r {
	case RoleMaster:
		return "master"
	case RoleSlave:
		return "slave"
	default:
		return "unknown"
	}
}

// ReplicationState holds the replication state
type ReplicationState struct {
	role          ReplicationRole
	masterHost    string
	masterPort    int
	masterConn    net.Conn
	replID        uint64
	replOffset    uint64
	mu            sync.RWMutex

	// Master-side: slave connections
	slaveConns    []net.Conn
	slaveConnsMu  sync.Mutex

	// Replication backlog for PSYNC
	replicationBacklog []byte
	backlogSize       int // Maximum size of backlog (default 1MB)
	backlogMu         sync.Mutex
}

// Global replication state
var State = &ReplicationState{
	role:       RoleMaster,
	replID:     1,                // Default replication ID
	replOffset: 0,
	backlogSize: 1 << 20,        // 1MB default backlog
}

// IsMaster returns true if this instance is a master
func (rs *ReplicationState) IsMaster() bool {
	rs.mu.RLock()
	defer rs.mu.RUnlock()
	return rs.role == RoleMaster
}

// IsSlave returns true if this instance is a slave
func (rs *ReplicationState) IsSlave() bool {
	rs.mu.RLock()
	defer rs.mu.RUnlock()
	return rs.role == RoleSlave
}

// GetRole returns the current role
func (rs *ReplicationState) GetRole() ReplicationRole {
	rs.mu.RLock()
	defer rs.mu.RUnlock()
	return rs.role
}

// GetMasterInfo returns master host and port
func (rs *ReplicationState) GetMasterInfo() (string, int) {
	rs.mu.RLock()
	defer rs.mu.RUnlock()
	return rs.masterHost, rs.masterPort
}

// GetReplicationID returns the replication ID
func (rs *ReplicationState) GetReplicationID() uint64 {
	rs.mu.RLock()
	defer rs.mu.RUnlock()
	return rs.replID
}

// GetReplicationOffset returns the replication offset
func (rs *ReplicationState) GetReplicationOffset() uint64 {
	rs.mu.RLock()
	defer rs.mu.RUnlock()
	return rs.replOffset
}

// IncrementReplicationOffset increments the replication offset
func (rs *ReplicationState) IncrementReplicationOffset(delta uint64) {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	rs.replOffset += delta
}

// SetAsSlave sets this instance as a slave of the given master
func (rs *ReplicationState) SetAsSlave(host string, port int) error {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	// Disconnect from existing master if any
	if rs.masterConn != nil {
		rs.masterConn.Close()
		rs.masterConn = nil
	}

	rs.role = RoleSlave
	rs.masterHost = host
	rs.masterPort = port
	rs.replID = 0 // Slaves don't have a replication ID

	return nil
}

// SetAsMaster sets this instance as a master
func (rs *ReplicationState) SetAsMaster() {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	// Disconnect from master if connected as slave
	if rs.masterConn != nil {
		rs.masterConn.Close()
		rs.masterConn = nil
	}

	rs.role = RoleMaster
	rs.masterHost = ""
	rs.masterPort = 0
	rs.replID = 1 // Master has replication ID 1
}

// ConnectToMaster connects to the master server
func (rs *ReplicationState) ConnectToMaster() error {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	if rs.role != RoleSlave {
		return fmt.Errorf("not configured as slave")
	}

	if rs.masterHost == "" {
		return fmt.Errorf("no master configured")
	}

	// Connect to master
	addr := fmt.Sprintf("%s:%d", rs.masterHost, rs.masterPort)
	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		return fmt.Errorf("failed to connect to master: %w", err)
	}

	rs.masterConn = conn
	return nil
}

// DisconnectFromMaster disconnects from the master
func (rs *ReplicationState) DisconnectFromMaster() error {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	if rs.masterConn != nil {
		err := rs.masterConn.Close()
		rs.masterConn = nil
		return err
	}

	return nil
}

// SendPSync sends a PSYNC command to the master
func (rs *ReplicationState) SendPSync(replID uint64, offset uint64) error {
	rs.mu.RLock()
	conn := rs.masterConn
	rs.mu.RUnlock()

	if conn == nil {
		return fmt.Errorf("not connected to master")
	}

	// Send PSYNC command
	cmd := fmt.Sprintf("PSYNC %d %d\r\n", replID, offset)
	_, err := conn.Write([]byte(cmd))
	if err != nil {
		return fmt.Errorf("failed to send PSYNC: %w", err)
	}

	return nil
}

// SendSync sends a SYNC command to the master
func (rs *ReplicationState) SendSync() error {
	rs.mu.RLock()
	conn := rs.masterConn
	rs.mu.RUnlock()

	if conn == nil {
		return fmt.Errorf("not connected to master")
	}

	// Send SYNC command
	cmd := "SYNC\r\n"
	_, err := conn.Write([]byte(cmd))
	if err != nil {
		return fmt.Errorf("failed to send SYNC: %w", err)
	}

	return nil
}

// ReceiveSyncResponse receives and processes the SYNC response from master
// Returns the RDB data received from the master
func (rs *ReplicationState) ReceiveSyncResponse() ([]byte, error) {
	rs.mu.RLock()
	conn := rs.masterConn
	rs.mu.RUnlock()

	if conn == nil {
		return nil, fmt.Errorf("not connected to master")
	}

	// Set read timeout
	conn.SetReadDeadline(time.Now().Add(30 * time.Second))
	defer conn.SetReadDeadline(time.Time{})

	reader := bufio.NewReader(conn)

	// Read response line: +FULLRESYNC <replid> <offset>\r\n
	line, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read SYNC response: %w", err)
	}

	// Parse response
	if len(line) < 11 || line[0] != '+' {
		return nil, fmt.Errorf("invalid SYNC response: %s", line)
	}

	// Parse: FULLRESYNC <replid> <offset>
	parts := bytes.Fields([]byte(line[1 : len(line)-2]))
	if len(parts) != 3 || string(parts[0]) != "FULLRESYNC" {
		return nil, fmt.Errorf("invalid SYNC response format: %s", line)
	}

	// Parse replID and offset
	var replID uint64
	var replOffset uint64
	if _, err := fmt.Sscanf(string(parts[1]), "%d", &replID); err != nil {
		return nil, fmt.Errorf("invalid replID: %w", err)
	}
	if _, err := fmt.Sscanf(string(parts[2]), "%d", &replOffset); err != nil {
		return nil, fmt.Errorf("invalid offset: %w", err)
	}

	// Update replication state
	rs.mu.Lock()
	rs.replID = replID
	rs.replOffset = replOffset
	rs.mu.Unlock()

	fmt.Printf("Received SYNC response: replID=%d, offset=%d\n", replID, replOffset)

	// Read RDB file length: $<length>\r\n
	lengthLine, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read RDB length: %w", err)
	}

	if len(lengthLine) < 3 || lengthLine[0] != '$' {
		return nil, fmt.Errorf("invalid RDB length format: %s", lengthLine)
	}

	// Parse length
	var rdbLength int64
	if _, err := fmt.Sscanf(lengthLine[1:], "%d", &rdbLength); err != nil {
		return nil, fmt.Errorf("invalid RDB length: %w", err)
	}

	fmt.Printf("Receiving RDB file: %d bytes\n", rdbLength)

	// Read RDB data
	rdbData := make([]byte, rdbLength)
	bytesRead, err := io.ReadFull(reader, rdbData)
	if err != nil {
		return nil, fmt.Errorf("failed to read RDB data: %w", err)
	}

	if int64(bytesRead) != rdbLength {
		return nil, fmt.Errorf("incomplete RDB data: expected %d, got %d", rdbLength, bytesRead)
	}

	// Read trailing \r\n
	trailing := make([]byte, 2)
	if _, err := io.ReadFull(reader, trailing); err != nil {
		return nil, fmt.Errorf("failed to read trailing CRLF: %w", err)
	}

	if trailing[0] != '\r' || trailing[1] != '\n' {
		return nil, fmt.Errorf("invalid trailing bytes after RDB")
	}

	fmt.Printf("Successfully received RDB file (%d bytes)\n", len(rdbData))

	return rdbData, nil
}

// PerformFullSync performs a full synchronization with the master
// This is the main entry point for slave to sync with master
func (rs *ReplicationState) PerformFullSync() ([]byte, error) {
	// Connect to master if not already connected
	if rs.masterConn == nil {
		if err := rs.ConnectToMaster(); err != nil {
			return nil, fmt.Errorf("failed to connect to master: %w", err)
		}
	}

	// Send SYNC command
	if err := rs.SendSync(); err != nil {
		return nil, fmt.Errorf("failed to send SYNC: %w", err)
	}

	// Receive and return RDB data
	return rs.ReceiveSyncResponse()
}

// PerformPartialSync performs a partial synchronization with the master
func (rs *ReplicationState) PerformPartialSync(replID uint64, offset uint64) ([]byte, error) {
	// Connect to master if not already connected
	if rs.masterConn == nil {
		if err := rs.ConnectToMaster(); err != nil {
			return nil, fmt.Errorf("failed to connect to master: %w", err)
		}
	}

	// Send PSYNC command
	if err := rs.SendPSync(replID, offset); err != nil {
		return nil, fmt.Errorf("failed to send PSYNC: %w", err)
	}

	// For now, PSYNC will fallback to full sync if master doesn't support incremental
	// Receive response (may be FULLRESYNC or CONTINUE)
	return rs.ReceiveSyncResponse()
}

// RDBLoader defines the interface for loading RDB data
// Using interface{} to avoid circular imports
type RDBLoader interface {
	LoadRDBFromBytes(db interface{}, data []byte) error
}

// rdbLoader holds the registered RDB loader
var rdbLoader RDBLoader

// RegisterRDBLoader registers an RDB loader implementation
func RegisterRDBLoader(loader RDBLoader) {
	rdbLoader = loader
}

// LoadRDBData loads RDB data using the registered loader
func LoadRDBData(db interface{}, data []byte) error {
	if rdbLoader == nil {
		return fmt.Errorf("no RDB loader registered")
	}
	return rdbLoader.LoadRDBFromBytes(db, data)
}

// RegisterSlave registers a slave connection on the master
func (rs *ReplicationState) RegisterSlave(conn net.Conn) {
	rs.slaveConnsMu.Lock()
	defer rs.slaveConnsMu.Unlock()

	rs.slaveConns = append(rs.slaveConns, conn)
	fmt.Printf("Registered slave: %s (total slaves: %d)\n", conn.RemoteAddr(), len(rs.slaveConns))
}

// UnregisterSlave removes a slave connection
func (rs *ReplicationState) UnregisterSlave(conn net.Conn) {
	rs.slaveConnsMu.Lock()
	defer rs.slaveConnsMu.Unlock()

	for i, c := range rs.slaveConns {
		if c == conn {
			rs.slaveConns = append(rs.slaveConns[:i], rs.slaveConns[i+1:]...)
			fmt.Printf("Unregistered slave: %s (remaining slaves: %d)\n", conn.RemoteAddr(), len(rs.slaveConns))
			return
		}
	}
}

// GetSlaveCount returns the number of connected slaves
func (rs *ReplicationState) GetSlaveCount() int {
	rs.slaveConnsMu.Lock()
	defer rs.slaveConnsMu.Unlock()
	return len(rs.slaveConns)
}

// PropagateCommand sends a write command to all connected slaves
// This is called by the master after executing a write command
func (rs *ReplicationState) PropagateCommand(cmdLine [][]byte) error {
	// Only propagate if we're a master
	if !rs.IsMaster() {
		return nil
	}

	rs.slaveConnsMu.Lock()
	slaves := make([]net.Conn, len(rs.slaveConns))
	copy(slaves, rs.slaveConns)
	rs.slaveConnsMu.Unlock()

	if len(slaves) == 0 {
		return nil
	}

	// Convert command to RESP format
	cmdData := serializeCommand(cmdLine)

	// Add to replication backlog for PSYNC
	rs.addToBacklog(cmdData)

	// Send to all slaves (non-blocking)
	var wg sync.WaitGroup
	for _, slave := range slaves {
		wg.Add(1)
		go func(conn net.Conn) {
			defer wg.Done()
			if _, err := conn.Write(cmdData); err != nil {
				fmt.Printf("Failed to send command to slave %s: %v\n", conn.RemoteAddr(), err)
				// Don't unregister here, let the connection handler do it
			}
		}(slave)
	}
	wg.Wait()

	// Increment replication offset
	rs.IncrementReplicationOffset(uint64(len(cmdData)))

	return nil
}

// addToBacklog adds command data to the replication backlog
func (rs *ReplicationState) addToBacklog(data []byte) {
	rs.backlogMu.Lock()
	defer rs.backlogMu.Unlock()

	// Append new data to backlog
	rs.replicationBacklog = append(rs.replicationBacklog, data...)

	// Trim backlog if it exceeds max size
	if len(rs.replicationBacklog) > rs.backlogSize {
		// Keep only the most recent data
		excess := len(rs.replicationBacklog) - rs.backlogSize
		rs.replicationBacklog = rs.replicationBacklog[excess:]
	}
}

// GetBacklogData returns backlog data starting from the specified offset
// Returns nil if the offset is too old (no longer in backlog)
func (rs *ReplicationState) GetBacklogData(offset uint64) ([]byte, error) {
	rs.backlogMu.Lock()
	defer rs.backlogMu.Unlock()

	// Calculate starting position in backlog
	// The backlog contains bytes from (currentOffset - backlogLen) to currentOffset
	backlogLen := uint64(len(rs.replicationBacklog))
	currentOffset := rs.GetReplicationOffset()

	if offset > currentOffset {
		return nil, fmt.Errorf("offset %d is in the future (current: %d)", offset, currentOffset)
	}

	// Check if offset is within backlog range
	if currentOffset-backlogLen > offset {
		return nil, fmt.Errorf("offset %d is too old (not in backlog)", offset)
	}

	// Calculate position within backlog
	position := offset - (currentOffset - backlogLen)

	if position >= backlogLen {
		return nil, nil // No new data
	}

	return rs.replicationBacklog[position:], nil
}

// SetBacklogSize sets the maximum size of the replication backlog
func (rs *ReplicationState) SetBacklogSize(size int) {
	rs.backlogMu.Lock()
	defer rs.backlogMu.Unlock()
	rs.backlogSize = size

	// Trim existing backlog if needed
	if len(rs.replicationBacklog) > size {
		excess := len(rs.replicationBacklog) - size
		rs.replicationBacklog = rs.replicationBacklog[excess:]
	}
}

// GetBacklogSize returns the current backlog size limit
func (rs *ReplicationState) GetBacklogSize() int {
	rs.backlogMu.Lock()
	defer rs.backlogMu.Unlock()
	return rs.backlogSize
}

// serializeCommand converts a command to RESP format
func serializeCommand(cmdLine [][]byte) []byte {
	var buf bytes.Buffer

	// Write array header
	buf.WriteString(fmt.Sprintf("*%d\r\n", len(cmdLine)))

	// Write each argument as bulk string
	for _, arg := range cmdLine {
		buf.WriteString(fmt.Sprintf("$%d\r\n", len(arg)))
		buf.Write(arg)
		buf.WriteString("\r\n")
	}

	return buf.Bytes()
}

// CommandHandler defines the interface for handling propagated commands
type CommandHandler interface {
	ExecCommand(cmdLine [][]byte) ([][]byte, error)
}

// DBCommandAdapter wraps a database.DB to implement CommandHandler
type DBCommandAdapter struct {
	db interface{}
}

// NewDBCommandAdapter creates a new adapter
func NewDBCommandAdapter(db interface{}) *DBCommandAdapter {
	return &DBCommandAdapter{db: db}
}

// ExecCommand executes a command using the database.Exec method
func (a *DBCommandAdapter) ExecCommand(cmdLine [][]byte) ([][]byte, error) {
	// Use reflection or type assertion to call the right method
	// For now, we'll use a simple approach
	type executor interface {
		Exec(cmdLine [][]byte) ([][]byte, error)
	}

	if db, ok := a.db.(executor); ok {
		return db.Exec(cmdLine)
	}

	return nil, fmt.Errorf("database does not implement Exec method")
}

// StartReplicationLoop starts the replication loop for a slave
// This continuously receives and executes commands from the master
func (rs *ReplicationState) StartReplicationLoop(handler CommandHandler) error {
	if !rs.IsSlave() {
		return fmt.Errorf("not configured as slave")
	}

	rs.mu.RLock()
	conn := rs.masterConn
	rs.mu.RUnlock()

	if conn == nil {
		return fmt.Errorf("not connected to master")
	}

	// Start replication loop in background
	go func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("Replication loop panic: %v\n", r)
			}
		}()

		reader := bufio.NewReader(conn)

		for {
			// Set read deadline to detect stale connections
			conn.SetReadDeadline(time.Now().Add(60 * time.Second))

			// Read command from master
			cmdLine, err := rs.readCommand(reader)
			if err != nil {
				if err == io.EOF {
					fmt.Printf("Master closed connection\n")
				} else {
					fmt.Printf("Replication read error: %v\n", err)
				}
				return
			}

			if len(cmdLine) == 0 {
				continue
			}

			// Execute command locally
			if _, err := handler.ExecCommand(cmdLine); err != nil {
				fmt.Printf("Replication command execution error: %v\n", err)
			}

			// Update replication offset
			rs.IncrementReplicationOffset(1)
		}
	}()

	return nil
}

// readCommand reads a RESP command from the reader
func (rs *ReplicationState) readCommand(reader *bufio.Reader) ([][]byte, error) {
	// Read first character to determine type
	leadByte, err := reader.ReadByte()
	if err != nil {
		return nil, err
	}

	switch leadByte {
	case '*': // Array
		// Read array length
		line, err := reader.ReadString('\n')
		if err != nil {
			return nil, err
		}

		var arrayLen int
		if _, err := fmt.Sscanf(line, "%d\r\n", &arrayLen); err != nil {
			return nil, fmt.Errorf("invalid array length: %w", err)
		}

		if arrayLen < 0 {
			return nil, nil // Null array
		}

		// Read each bulk string
		cmdLine := make([][]byte, arrayLen)
		for i := 0; i < arrayLen; i++ {
			// Read bulk string marker
			marker, err := reader.ReadByte()
			if err != nil {
				return nil, err
			}

			if marker != '$' {
				return nil, fmt.Errorf("expected bulk string, got %c", marker)
			}

			// Read length
			lengthLine, err := reader.ReadString('\n')
			if err != nil {
				return nil, err
			}

			var length int
			if _, err := fmt.Sscanf(lengthLine, "%d\r\n", &length); err != nil {
				return nil, fmt.Errorf("invalid bulk string length: %w", err)
			}

			if length < 0 {
				cmdLine[i] = nil
				continue
			}

			// Read data
			data := make([]byte, length+2) // +2 for \r\n
			if _, err := io.ReadFull(reader, data); err != nil {
				return nil, err
			}

			cmdLine[i] = data[:length]
		}

		return cmdLine, nil

	default:
		return nil, fmt.Errorf("unsupported RESP type: %c", leadByte)
	}
}

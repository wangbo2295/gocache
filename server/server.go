package server

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/wangbo/gocache/auth"
	"github.com/wangbo/gocache/config"
	"github.com/wangbo/gocache/database"
	"github.com/wangbo/gocache/monitor"
	"github.com/wangbo/gocache/persistence"
	"github.com/wangbo/gocache/persistence/aof"
	"github.com/wangbo/gocache/protocol"
	"github.com/wangbo/gocache/protocol/resp"
	"github.com/wangbo/gocache/replication"
)

// Handler represents a command handler
type Handler struct {
	db            *database.DB
	aof           *aof.AOFHandler
	authenticator *auth.Authenticator
}

// MakeHandler creates a new handler
func MakeHandler(db *database.DB) *Handler {
	return &Handler{db: db}
}

// MakeHandlerWithAOF creates a new handler with AOF persistence
func MakeHandlerWithAOF(db *database.DB, aofHandler *aof.AOFHandler) *Handler {
	return &Handler{db: db, aof: aofHandler}
}

// MakeHandlerWithAuth creates a new handler with authenticator
func MakeHandlerWithAuth(db *database.DB, aofHandler *aof.AOFHandler, authenticator *auth.Authenticator) *Handler {
	return &Handler{db: db, aof: aofHandler, authenticator: authenticator}
}

// ExecCommand executes a command and returns a reply
func (h *Handler) ExecCommand(cmdLine [][]byte) (resp.Reply, error) {
	if len(cmdLine) == 0 {
		return nil, errors.New("empty command")
	}

	cmd := string(cmdLine[0])
	cmdUpper := protocol.ToUpper(cmd)

	// Handle PING command specially
	if cmdUpper == protocol.CmdPing {
		if len(cmdLine) == 1 {
			return resp.MakePongReply(), nil
		}
		return resp.MakeStatusReply(string(cmdLine[1])), nil
	}

	// Track execution time for slow log
	startTime := time.Now()

	// Execute command in database
	result, err := h.db.Exec(cmdLine)
	if err != nil {
		return resp.MakeErrorReply(err.Error()), nil
	}

	// Calculate execution time and log to slow log if needed
	duration := time.Since(startTime)
	h.db.AddSlowLogEntry(duration, cmdLine)

	// Log command to monitor if enabled (skip MONITOR command itself)
	if cmdUpper != protocol.CmdMonitor {
		monitor.GetMonitor().LogCommand(cmdLine, "")
	}

	// Write to AOF if enabled and command is write operation
	if h.aof != nil && protocol.IsWriteCommand(cmdUpper) {
		if err := h.aof.AddCommand(cmdLine); err != nil {
			// Log error but don't fail the command
			fmt.Printf("AOF write error: %v\n", err)
		}
	}

	// Propagate write commands to slaves
	if protocol.IsWriteCommand(cmdUpper) {
		if err := replication.State.PropagateCommand(cmdLine); err != nil {
			// Log error but don't fail the command
			fmt.Printf("Replication propagation error: %v\n", err)
		}
	}

	// Convert result to appropriate reply type
	if len(result) == 0 {
		return resp.MakeNullBulkReply(), nil
	}

	// For SET/MSET commands, return OK
	if protocol.IsStatusCommand(cmdUpper) {
		return resp.MakeStatusReply("OK"), nil
	}

	// For commands that return integers (DEL, EXISTS, INCR, DECR, etc.)
	if protocol.IsIntegerCommand(cmdUpper) {
		if len(result) == 1 && result[0] != nil {
			// Parse integer from result
			val := string(result[0])
			var num int64
			if _, err := fmt.Sscanf(val, "%d", &num); err == nil {
				return resp.MakeIntReply(num), nil
			}
		}
	}

	// For commands that return arrays (HGETALL, LRANGE, etc.)
	// These should always return arrays even if there's only 1 element
	if protocol.IsArrayCommand(cmdUpper) {
		return resp.MakeMultiBulkReply(result), nil
	}

	// For single result commands (GET, STRLEN, etc.)
	if len(result) == 1 {
		if result[0] == nil {
			return resp.MakeNullBulkReply(), nil
		}
		return resp.MakeBulkReply(result[0]), nil
	}

	// For multiple results (MGET, KEYS), return as array
	return resp.MakeMultiBulkReply(result), nil
}

// Client represents a connected client
type Client struct {
	conn          net.Conn
	server        *Server
	authenticated bool
	clientID      string
}

// Server represents the Redis server
type Server struct {
	config    *config.Properties
	handler   *Handler
	listener  net.Listener
	closing   bool
	wg        sync.WaitGroup
}

// MakeServer creates a new server
func MakeServer(cfg *config.Properties, handler *Handler) *Server {
	return &Server{
		config:  cfg,
		handler: handler,
		closing: false,
	}
}

// Start starts the server
func (s *Server) Start() error {
	addr := fmt.Sprintf("%s:%d", s.config.Bind, s.config.Port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", addr, err)
	}
	s.listener = listener

	fmt.Printf("Server is listening on %s\n", addr)

	// Accept connections in a loop
	for !s.closing {
		conn, err := listener.Accept()
		if err != nil {
			if s.closing {
				return nil
			}
			return fmt.Errorf("accept error: %w", err)
		}

		// Handle each connection in a separate goroutine
		client := &Client{
			conn:          conn,
			server:        s,
			authenticated: false,
			clientID:      conn.RemoteAddr().String(),
		}
		s.wg.Add(1)
		go client.handleConnection()
	}

	return nil
}

// Stop stops the server
func (s *Server) Stop() {
	s.closing = true
	if s.listener != nil {
		s.listener.Close()
	}
	s.wg.Wait()
}

// handleConnection handles a client connection
func (c *Client) handleConnection() {
	defer c.conn.Close()
	defer c.server.wg.Done()

	remoteAddr := c.conn.RemoteAddr().String()
	fmt.Printf("Client connected: %s\n", remoteAddr)

	// Parse and execute commands
	parser := resp.MakeParser()

	for {
		// Read and parse command
		cmdLine, err := parser.ParseStream(c.conn)
		if err != nil {
			if err == io.EOF {
				fmt.Printf("Client disconnected: %s\n", remoteAddr)
				return
			}
			// Send error reply
			errReply := resp.MakeErrorReply(err.Error())
			c.conn.Write(errReply.ToBytes())
			continue
		}

		if len(cmdLine) == 0 {
			continue
		}

		// Check if this is a SYNC or PSYNC command (replication commands)
		cmdUpper := protocol.ToUpper(string(cmdLine[0]))
		if cmdUpper == protocol.CmdSync || cmdUpper == protocol.CmdPSync {
			// Handle replication commands specially
			if err := c.handleReplicationCommand(cmdLine); err != nil {
				fmt.Printf("Replication command error: %v\n", err)
				errReply := resp.MakeErrorReply(err.Error())
				c.conn.Write(errReply.ToBytes())
			}
			return
		}

		// Check if this is a MONITOR command
		if cmdUpper == protocol.CmdMonitor {
			// Handle MONITOR command specially
			if err := c.handleMonitor(); err != nil {
				fmt.Printf("Monitor command error: %v\n", err)
				errReply := resp.MakeErrorReply(err.Error())
				c.conn.Write(errReply.ToBytes())
			}
			return
		}

		// Check if this is an AUTH command
		if cmdUpper == protocol.CmdAuth {
			// Handle AUTH command specially
			if err := c.handleAuth(cmdLine); err != nil {
				errReply := resp.MakeErrorReply(err.Error())
				c.conn.Write(errReply.ToBytes())
			}
			continue
		}

		// Check authentication if required
		if c.server.handler.authenticator != nil &&
			c.server.handler.authenticator.IsEnabled() &&
			!c.authenticated {
			errReply := resp.MakeErrorReply("NOAUTH Authentication required.")
			c.conn.Write(errReply.ToBytes())
			continue
		}

		// Execute command
		result, _ := c.server.handler.ExecCommand(cmdLine)

		// Send reply
		c.conn.Write(result.ToBytes())
	}
}

// handleReplicationCommand handles SYNC and PSYNC commands
// These commands require special handling because they send large RDB files
func (c *Client) handleReplicationCommand(cmdLine [][]byte) error {
	cmdUpper := protocol.ToUpper(string(cmdLine[0]))

	if cmdUpper == protocol.CmdSync {
		return c.handleSync()
	}

	if cmdUpper == protocol.CmdPSync {
		return c.handlePSync(cmdLine)
	}

	return fmt.Errorf("unknown replication command: %s", cmdUpper)
}

// handleSync handles a full synchronization request from a slave
func (c *Client) handleSync() error {
	// Verify this instance is a master
	if !replication.State.IsMaster() {
		return fmt.Errorf("SYNC is only valid on master")
	}

	// Generate RDB file to a buffer
	var rdbBuffer bytes.Buffer
	if err := persistence.SaveDatabaseToWriter(c.server.handler.db, &rdbBuffer); err != nil {
		return fmt.Errorf("failed to generate RDB: %w", err)
	}

	rdbData := rdbBuffer.Bytes()

	// Send SYNC response: +FULLRESYNC <replid> <offset>\r\n
	replID := replication.State.GetReplicationID()
	replOffset := replication.State.GetReplicationOffset()
	syncResponse := fmt.Sprintf("+FULLRESYNC %d %d\r\n", replID, replOffset)

	if _, err := c.conn.Write([]byte(syncResponse)); err != nil {
		return fmt.Errorf("failed to send SYNC response: %w", err)
	}

	// Send RDB file length: $<length>\r\n
	lengthLine := fmt.Sprintf("$%d\r\n", len(rdbData))
	if _, err := c.conn.Write([]byte(lengthLine)); err != nil {
		return fmt.Errorf("failed to send RDB length: %w", err)
	}

	// Send RDB file content
	if _, err := c.conn.Write(rdbData); err != nil {
		return fmt.Errorf("failed to send RDB data: %w", err)
	}

	// Send trailing \r\n
	if _, err := c.conn.Write([]byte("\r\n")); err != nil {
		return fmt.Errorf("failed to send trailing CRLF: %w", err)
	}

	fmt.Printf("Sent RDB file (%d bytes) to slave %s\n", len(rdbData), c.conn.RemoteAddr())

	// Register this slave connection for command propagation
	replication.State.RegisterSlave(c.conn)

	// Start a goroutine to handle command propagation to this slave
	go c.propagateCommandsToSlave()

	return nil
}

// propagateCommandsToSlave continuously sends propagated commands to slave
// This goroutine runs after SYNC completes to forward subsequent write commands
func (c *Client) propagateCommandsToSlave() {
	defer func() {
		c.conn.Close()
		replication.State.UnregisterSlave(c.conn)
	}()

	// For now, we just keep the connection alive
	// In the future, we would have a channel that receives commands to propagate
	// For the current implementation, commands are propagated immediately when executed
	// This goroutine mainly serves to keep the connection open and handle cleanup

	// Keep reading from slave (PING, etc.)
	parser := resp.MakeParser()
	for {
		cmdLine, err := parser.ParseStream(c.conn)
		if err != nil {
			if err != io.EOF {
				fmt.Printf("Slave connection error: %v\n", err)
			}
			return
		}

		if len(cmdLine) == 0 {
			continue
		}

		// Handle slave commands (currently just PING)
		cmd := string(cmdLine[0])
		cmdUpper := protocol.ToUpper(cmd)

		if cmdUpper == protocol.CmdPing {
			c.conn.Write(resp.MakePongReply().ToBytes())
		}
		// Other slave commands can be added here
	}
}

// handlePSync handles a partial synchronization request from a slave
func (c *Client) handlePSync(cmdLine [][]byte) error {
	// Verify this instance is a master
	if !replication.State.IsMaster() {
		return fmt.Errorf("PSYNC is only valid on master")
	}

	// Parse arguments: PSYNC <replid> <offset>
	if len(cmdLine) != 3 {
		return fmt.Errorf("wrong number of arguments for PSYNC")
	}

	replIDStr := string(cmdLine[1])
	offsetStr := string(cmdLine[2])

	// Parse offset
	var offset uint64
	if _, err := fmt.Sscanf(offsetStr, "%d", &offset); err != nil {
		return fmt.Errorf("invalid offset: %w", err)
	}

	// Check if we can do incremental sync
	// For now, we don't match replID (simplified)
	// In production, you would check if replID matches
	_ = replIDStr // Will be used for replID matching in future

	// Try to get incremental data from backlog
	backlogData, err := replication.State.GetBacklogData(offset)
	if err != nil || backlogData == nil {
		// Fallback to full sync
		fmt.Printf("PSYNC: backlog not available, doing full sync (offset=%d)\n", offset)
		return c.handleSync()
	}

	// Send CONTINUE response with incremental data
	replOffset := replication.State.GetReplicationOffset()
	continueResponse := fmt.Sprintf("+CONTINUE %d\r\n", replOffset)

	if _, err := c.conn.Write([]byte(continueResponse)); err != nil {
		return fmt.Errorf("failed to send CONTINUE response: %w", err)
	}

	// Send backlog data
	if _, err := c.conn.Write(backlogData); err != nil {
		return fmt.Errorf("failed to send backlog data: %w", err)
	}

	fmt.Printf("Sent incremental sync (%d bytes) to slave %s\n", len(backlogData), c.conn.RemoteAddr())

	// Register this slave connection for command propagation
	replication.State.RegisterSlave(c.conn)

	// Start a goroutine to handle command propagation to this slave
	go c.propagateCommandsToSlave()

	return nil
}

// handleMonitor handles the MONITOR command
func (c *Client) handleMonitor() error {
	// Send OK response to indicate monitoring has started
	okReply := resp.MakeStatusReply("OK")
	if _, err := c.conn.Write(okReply.ToBytes()); err != nil {
		return fmt.Errorf("failed to send OK response: %w", err)
	}

	// Add this client to the monitor
	monitor.GetMonitor().AddClient(c.conn)
	defer monitor.GetMonitor().RemoveClient(c.conn)

	// Send a welcome message
	welcomeMsg := fmt.Sprintf("+OK %d\r\n", time.Now().Unix())
	c.conn.Write([]byte(welcomeMsg))

	// Keep connection open and continue streaming commands
	// The monitor broadcast loop will send commands to this client
	// We just need to keep the connection alive
	parser := resp.MakeParser()
	for {
		cmdLine, err := parser.ParseStream(c.conn)
		if err != nil {
			if err == io.EOF {
				fmt.Printf("Monitor client disconnected: %s\n", c.conn.RemoteAddr())
				return nil
			}
			fmt.Printf("Monitor client error: %v\n", err)
			return err
		}

		if len(cmdLine) == 0 {
			continue
		}

		// In monitoring mode, we don't execute commands from this client
		// Just send an error reply
		errReply := resp.MakeErrorReply("MONITOR mode - cannot execute commands")
		c.conn.Write(errReply.ToBytes())
	}
}

// handleAuth handles the AUTH command
func (c *Client) handleAuth(cmdLine [][]byte) error {
	if len(cmdLine) != 2 {
		return fmt.Errorf("wrong number of arguments for AUTH")
	}

	password := string(cmdLine[1])

	// If no authenticator configured, accept any password
	if c.server.handler.authenticator == nil {
		c.authenticated = true
		okReply := resp.MakeStatusReply("OK")
		c.conn.Write(okReply.ToBytes())
		return nil
	}

	// Check if authentication is enabled
	if !c.server.handler.authenticator.IsEnabled() {
		c.authenticated = true
		okReply := resp.MakeStatusReply("OK")
		c.conn.Write(okReply.ToBytes())
		return nil
	}

	// Authenticate using the authenticator
	if c.server.handler.authenticator.Authenticate(password) {
		c.authenticated = true
		okReply := resp.MakeStatusReply("OK")
		c.conn.Write(okReply.ToBytes())
		return nil
	}

	// Authentication failed
	return fmt.Errorf("invalid password")
}

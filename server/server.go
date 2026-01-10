package server

import (
	"errors"
	"fmt"
	"io"
	"net"
	"sync"

	"github.com/wangbo/gocache/config"
	"github.com/wangbo/gocache/database"
	"github.com/wangbo/gocache/protocol/resp"
)

// Handler represents a command handler
type Handler struct {
	db *database.DB
}

// MakeHandler creates a new handler
func MakeHandler(db *database.DB) *Handler {
	return &Handler{db: db}
}

// ExecCommand executes a command and returns a reply
func (h *Handler) ExecCommand(cmdLine [][]byte) (resp.Reply, error) {
	if len(cmdLine) == 0 {
		return nil, errors.New("empty command")
	}

	cmd := string(cmdLine[0])

	// Handle PING command specially
	if cmd == "PING" {
		if len(cmdLine) == 1 {
			return resp.MakePongReply(), nil
		}
		return resp.MakeStatusReply(string(cmdLine[1])), nil
	}

	// Execute command in database
	result, err := h.db.Exec(cmdLine)
	if err != nil {
		return resp.MakeErrorReply(err.Error()), nil
	}

	// Convert result to appropriate reply type
	if len(result) == 0 {
		return resp.MakeNullBulkReply(), nil
	}

	// For SET/MSET commands, return OK
	if cmd == "SET" || cmd == "MSET" {
		return resp.MakeStatusReply("OK"), nil
	}

	// For commands that return integers (DEL, EXISTS, INCR, DECR, etc.)
	if isIntegerCommand(cmd) {
		if len(result) == 1 && result[0] != nil {
			// Parse integer from result
			val := string(result[0])
			var num int64
			if _, err := fmt.Sscanf(val, "%d", &num); err == nil {
				return resp.MakeIntReply(num), nil
			}
		}
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

// isIntegerCommand checks if a command returns an integer result
func isIntegerCommand(cmd string) bool {
	intCmds := map[string]bool{
		"DEL":     true,
		"EXISTS":  true,
		"INCR":    true,
		"INCRBY":  true,
		"DECR":    true,
		"DECRBY":  true,
		"STRLEN":  true,
		"APPEND":  true,
		"EXPIRE":  true,
		"PEXPIRE": true,
		"PERSIST": true,
		"TTL":     true,
		"PTTL":    true,
	}
	return intCmds[cmd]
}

// Client represents a connected client
type Client struct {
	conn net.Conn
	server *Server
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
			conn:   conn,
			server: s,
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

		// Execute command
		result, _ := c.server.handler.ExecCommand(cmdLine)

		// Send reply
		c.conn.Write(result.ToBytes())
	}
}

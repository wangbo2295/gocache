package monitor

import (
	"fmt"
	"net"
	"sync"
	"time"
)

// Monitor manages command monitoring
type Monitor struct {
	clients    []net.Conn
	clientsMu  sync.RWMutex
	enabled    bool
	monitorCh  chan *MonitoredCommand
}

// MonitoredCommand represents a command being monitored
type MonitoredCommand struct {
	Timestamp time.Time
	Command   string
	Client    string // Client address
}

var (
	// Global monitor instance
	globalMonitor = &Monitor{
		clients:   make([]net.Conn, 0),
		enabled:   false,
		monitorCh: make(chan *MonitoredCommand, 1000),
	}
)

// GetMonitor returns the global monitor instance
func GetMonitor() *Monitor {
	return globalMonitor
}

// AddClient adds a monitoring client
func (m *Monitor) AddClient(conn net.Conn) {
	m.clientsMu.Lock()
	defer m.clientsMu.Unlock()

	m.clients = append(m.clients, conn)

	// Start monitoring if this is the first client
	if len(m.clients) == 1 {
		m.enabled = true
		go m.broadcastLoop()
	}

	fmt.Printf("Monitor: client added (total: %d)\n", len(m.clients))
}

// RemoveClient removes a monitoring client
func (m *Monitor) RemoveClient(conn net.Conn) {
	m.clientsMu.Lock()
	defer m.clientsMu.Unlock()

	for i, c := range m.clients {
		if c == conn {
			m.clients = append(m.clients[:i], m.clients[i+1:]...)
			break
		}
	}

	// Stop monitoring if no more clients
	if len(m.clients) == 0 {
		m.enabled = false
	}

	fmt.Printf("Monitor: client removed (remaining: %d)\n", len(m.clients))
}

// LogCommand logs a command for monitoring
func (m *Monitor) LogCommand(cmdLine [][]byte, clientAddr string) {
	if !m.enabled {
		return
	}

	// Serialize command
	cmd := serializeCommand(cmdLine)

	cmdMon := &MonitoredCommand{
		Timestamp: time.Now(),
		Command:   cmd,
		Client:    clientAddr,
	}

	// Send to monitor channel (non-blocking)
	select {
	case m.monitorCh <- cmdMon:
	default:
		// Channel full, drop the command
	}
}

// broadcastLoop broadcasts commands to all monitoring clients
func (m *Monitor) broadcastLoop() {
	for cmdMon := range m.monitorCh {
		m.clientsMu.RLock()
		clients := make([]net.Conn, len(m.clients))
		copy(clients, m.clients)
		m.clientsMu.RUnlock()

		if len(clients) == 0 {
			// No more clients, stop monitoring
			m.enabled = false
			return
		}

		// Format: timestamp in microseconds + command
		timestampMicros := cmdMon.Timestamp.UnixNano() / 1000
		message := fmt.Sprintf("%d [db 0] \"%s\"\r\n", timestampMicros, cmdMon.Command)

		// Send to all clients
		for _, client := range clients {
			if _, err := client.Write([]byte(message)); err != nil {
				// Remove client on error
				m.RemoveClient(client)
			}
		}
	}
}

// serializeCommand serializes a command line to string
func serializeCommand(cmdLine [][]byte) string {
	if len(cmdLine) == 0 {
		return ""
	}

	result := ""
	for i, arg := range cmdLine {
		if i > 0 {
			result += " "
		}
		// Escape arguments with spaces or quotes
		argStr := string(arg)
		if len(argStr) == 0 || (len(argStr) > 0 && (argStr[0] == ' ' || argStr[len(argStr)-1] == ' ')) {
			result += `"` + argStr + `"`
		} else {
			result += argStr
		}
	}
	return result
}

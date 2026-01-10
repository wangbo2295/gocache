package aof

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/wangbo/gocache/database"
	"github.com/wangbo/gocache/protocol/resp"
)

// AOFHandler represents an AOF persistence handler
type AOFHandler struct {
	file    *os.File
	writer  *bufio.Writer
	db      *database.DB
	mu      sync.Mutex
	closing bool
}

// MakeAOFHandler creates a new AOF handler
func MakeAOFHandler(filename string, db *database.DB) (*AOFHandler, error) {
	// Open file in append mode, create if not exists
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open AOF file: %w", err)
	}

	handler := &AOFHandler{
		file:   file,
		writer: bufio.NewWriter(file),
		db:     db,
	}

	// Load existing data from AOF file
	if err := handler.Load(); err != nil {
		file.Close()
		return nil, fmt.Errorf("failed to load AOF file: %w", err)
	}

	return handler, nil
}

// Load loads and replays commands from AOF file
func (h *AOFHandler) Load() error {
	// Seek to beginning of file
	if _, err := h.file.Seek(0, 0); err != nil {
		return err
	}

	// Create reader
	reader := bufio.NewReader(h.file)
	parser := resp.MakeParser()

	// Read and execute commands line by line
	for {
		// Read command
		cmdLine, err := parser.ParseStream(reader)
		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("parse error: %w", err)
		}

		if len(cmdLine) == 0 {
			continue
		}

		// Execute command in database (don't write to AOF during load)
		// We use a flag to prevent recursive AOF writes
		_, err = h.db.Exec(cmdLine)
		if err != nil {
			// Log error but continue processing
			fmt.Printf("Error executing command from AOF: %v\n", err)
		}
	}

	// Seek back to end for appending
	if _, err := h.file.Seek(0, 2); err != nil {
		return err
	}

	return nil
}

// AddCommand writes a command to AOF file
func (h *AOFHandler) AddCommand(cmdLine [][]byte) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.closing {
		return fmt.Errorf("AOF handler is closing")
	}

	// Write command in RESP array format
	// Format: *<count>\r\n$<len1>\r\n<arg1>\r\n$<len2>\r\n<arg2>\r\n...

	// Write array header
	if _, err := h.writer.WriteString(fmt.Sprintf("*%d\r\n", len(cmdLine))); err != nil {
		return err
	}

	// Write each argument as bulk string
	for _, arg := range cmdLine {
		if arg == nil {
			// Write null bulk string
			if _, err := h.writer.WriteString("$-1\r\n"); err != nil {
				return err
			}
		} else {
			// Write bulk string
			if _, err := h.writer.WriteString(fmt.Sprintf("$%d\r\n%s\r\n", len(arg), string(arg))); err != nil {
				return err
			}
		}
	}

	// Flush to disk
	return h.writer.Flush()
}

// Close closes the AOF handler
func (h *AOFHandler) Close() error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.closing {
		return nil
	}

	h.closing = true

	// Flush buffer
	if err := h.writer.Flush(); err != nil {
		return err
	}

	// Sync to disk
	if err := h.file.Sync(); err != nil {
		return err
	}

	// Close file
	return h.file.Close()
}

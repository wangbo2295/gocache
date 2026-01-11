package aof

import (
	"bufio"
	"fmt"
	"os"
	"sync"

	"github.com/wangbo/gocache/database"
	"github.com/wangbo/gocache/datastruct"
)

// Rewriter handles AOF file rewriting
type Rewriter struct {
	aof       *AOFHandler
	db        *database.DB
	rewriting bool
	mu        sync.Mutex
}

// MakeRewriter creates a new AOF rewriter
func MakeRewriter(aof *AOFHandler, db *database.DB) *Rewriter {
	return &Rewriter{
		aof: aof,
		db:  db,
	}
}

// Rewrite performs AOF rewrite
// It creates a new compacted AOF file and atomically replaces the old one
func (r *Rewriter) Rewrite() error {
	r.mu.Lock()
	if r.rewriting {
		r.mu.Unlock()
		return fmt.Errorf("AOF rewrite already in progress")
	}
	r.rewriting = true
	r.mu.Unlock()

	defer func() {
		r.mu.Lock()
		r.rewriting = false
		r.mu.Unlock()
	}()

	// Get AOF file path
	aofPath := r.aof.file.Name()
	tmpPath := aofPath + ".tmp"
	rewritePath := aofPath + ".rewrite"

	// Create temporary rewrite file
	tmpFile, err := os.Create(tmpPath)
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}

	// Create handler for rewrite with buffered writer
	rewriteHandler := &AOFHandler{
		file:   tmpFile,
		writer: bufio.NewWriter(tmpFile),
		db:     r.db,
		closing: false,
	}

	// Write all current data to rewrite file
	if err := r.writeAllData(rewriteHandler); err != nil {
		tmpFile.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("failed to write data: %w", err)
	}

	// Sync and close temp file
	if err := rewriteHandler.writer.Flush(); err != nil {
		tmpFile.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("failed to flush: %w", err)
	}

	if err := tmpFile.Sync(); err != nil {
		tmpFile.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("failed to sync: %w", err)
	}

	if err := tmpFile.Close(); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("failed to close: %w", err)
	}

	// Rename temp to rewrite file
	if err := os.Rename(tmpPath, rewritePath); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("failed to rename: %w", err)
	}

	// Atomically replace old AOF file
	if err := os.Rename(rewritePath, aofPath); err != nil {
		return fmt.Errorf("failed to replace AOF file: %w", err)
	}

	// Reopen AOF file for appending
	r.aof.mu.Lock()
	defer r.aof.mu.Unlock()

	// Close old file
	r.aof.file.Close()

	// Open new file
	newFile, err := os.OpenFile(aofPath, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("failed to reopen AOF file: %w", err)
	}

	r.aof.file = newFile
	r.aof.writer.Reset(newFile)
	r.aof.closing = false

	return nil
}

// writeAllData writes all current database data to AOF handler
func (r *Rewriter) writeAllData(handler *AOFHandler) error {
	// Get all keys from database
	keys := r.db.Keys()

	// For each key, write the minimal command to recreate it
	for _, key := range keys {
		entity, ok := r.db.GetEntity(key)
		if !ok || entity == nil {
			continue // Skip keys that can't be accessed
		}

		// Write data based on type using type assertion
		switch data := entity.Data.(type) {
		case *datastruct.String:
			if err := r.writeString(key, data, handler); err != nil {
				return err
			}
		case *datastruct.Hash:
			if err := r.writeHash(key, data, handler); err != nil {
				return err
			}
		case *datastruct.List:
			if err := r.writeList(key, data, handler); err != nil {
				return err
			}
		case *datastruct.Set:
			if err := r.writeSet(key, data, handler); err != nil {
				return err
			}
		case *datastruct.SortedSet:
			if err := r.writeSortedSet(key, data, handler); err != nil {
				return err
			}
		}

		// Write TTL if exists
		if ttl := r.db.TTL(key); ttl > 0 {
			// Use PEXPIRE for millisecond precision
			cmd := [][]byte{[]byte("PEXPIRE"), []byte(key), []byte(fmt.Sprintf("%d", ttl))}
			if err := handler.AddCommand(cmd); err != nil {
				return err
			}
		}
	}

	return nil
}

// writeString writes a string key to AOF
func (r *Rewriter) writeString(key string, data *datastruct.String, handler *AOFHandler) error {
	cmd := [][]byte{[]byte("SET"), []byte(key), data.Value}
	return handler.AddCommand(cmd)
}

// writeHash writes a hash key to AOF
func (r *Rewriter) writeHash(key string, data *datastruct.Hash, handler *AOFHandler) error {
	// Use HMSET for efficiency (set all fields at once)
	args := [][]byte{[]byte("HMSET"), []byte(key)}
	allData := data.GetAll()
	for field, value := range allData {
		args = append(args, []byte(field), value)
	}

	return handler.AddCommand(args)
}

// writeList writes a list key to AOF
func (r *Rewriter) writeList(key string, data *datastruct.List, handler *AOFHandler) error {
	// Use RPUSH to add all elements at once
	args := [][]byte{[]byte("RPUSH"), []byte(key)}
	elements := data.GetAll()
	for _, elem := range elements {
		args = append(args, elem)
	}

	return handler.AddCommand(args)
}

// writeSet writes a set key to AOF
func (r *Rewriter) writeSet(key string, data *datastruct.Set, handler *AOFHandler) error {
	// Use SADD to add all members
	args := [][]byte{[]byte("SADD"), []byte(key)}
	members := data.Members()
	for _, member := range members {
		args = append(args, member)
	}

	return handler.AddCommand(args)
}

// writeSortedSet writes a sorted set key to AOF
func (r *Rewriter) writeSortedSet(key string, data *datastruct.SortedSet, handler *AOFHandler) error {
	// Use ZADD to add all members
	args := [][]byte{[]byte("ZADD"), []byte(key)}

	// Get all members with their scores
	// We need to iterate by rank to get score-member pairs
	for i := 0; i < data.Len(); i++ {
		member := data.GetMemberByRank(i)
		score := data.GetScoreByRank(i)
		args = append(args, []byte(fmt.Sprintf("%f", score)), member)
	}

	return handler.AddCommand(args)
}

// IsRewriting returns true if a rewrite is in progress
func (r *Rewriter) IsRewriting() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.rewriting
}

// RewriteInBackground performs AOF rewrite in a goroutine
func (r *Rewriter) RewriteInBackground() error {
	if r.IsRewriting() {
		return fmt.Errorf("AOF rewrite already in progress")
	}

	go func() {
		if err := r.Rewrite(); err != nil {
			fmt.Printf("AOF rewrite failed: %v\n", err)
		}
	}()

	return nil
}

package resp

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"
)

const (
	// RESP type markers
	SimpleString byte = '+'
	Error        byte = '-'
	Integer      byte = ':'
	BulkString   byte = '$'
	Array        byte = '*'
)

var (
	ErrInvalidSyntax = errors.New("resp: invalid syntax")
	ErrInvalidFormat = errors.New("resp: invalid format")
)

// ParseStream reads and parses one RESP command from reader
func ParseStream(reader io.Reader) ([][]byte, error) {
	// For now, we'll implement a simpler version that reads line by line
	// A full implementation would handle bulk strings and arrays properly
	bufReader := bufio.NewReader(reader)

	// Read first character to determine type
	line, err := bufReader.ReadString('\n')
	if err != nil {
		return nil, err
	}

	if len(line) < 3 {
		return nil, ErrInvalidSyntax
	}

	// Remove \r\n
	line = line[:len(line)-2]

	switch line[0] {
	case Array:
		// Array: *2\r\n$3\r\nGET\r\n$3\r\nkey\r\n
		count, err := strconv.Atoi(line[1:])
		if err != nil {
			return nil, ErrInvalidFormat
		}
		return parseArray(bufReader, count)
	case BulkString:
		// Bulk string: $6\r\nfoobar\r\n
		size, err := strconv.Atoi(line[1:])
		if err != nil {
			return nil, ErrInvalidFormat
		}
		data, err := parseBulkString(bufReader, size)
		if err != nil {
			return nil, err
		}
		return [][]byte{data}, nil
	case SimpleString, Error, Integer:
		// Simple types: +OK\r\n, -Error\r\n, :123\r\n
		return [][]byte{[]byte(line[1:])}, nil
	default:
		// Treat as inline command (simple string without prefix)
		return [][]byte{[]byte(line)}, nil
	}
}

// parseArray parses RESP array
func parseArray(reader *bufio.Reader, count int) ([][]byte, error) {
	if count < 0 {
		return nil, ErrInvalidFormat
	}

	args := make([][]byte, 0, count)

	for i := 0; i < count; i++ {
		// Read the bulk string header ($size\r\n)
		line, err := reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		if len(line) < 3 {
			return nil, ErrInvalidSyntax
		}
		line = line[:len(line)-2] // Remove \r\n

		if line[0] != BulkString {
			return nil, fmt.Errorf("resp: expected bulk string, got %c", line[0])
		}

		size, err := strconv.Atoi(line[1:])
		if err != nil {
			return nil, ErrInvalidFormat
		}

		// Read the bulk string data
		data, err := parseBulkString(reader, size)
		if err != nil {
			return nil, err
		}

		args = append(args, data)
	}

	return args, nil
}

// parseBulkString parses RESP bulk string
func parseBulkString(reader *bufio.Reader, size int) ([]byte, error) {
	if size < 0 {
		// Null bulk string ($-1\r\n)
		// We already read the size line in the caller, so just return nil
		return nil, nil
	}

	// Read the data
	data := make([]byte, size+2) // +2 for \r\n
	n, err := io.ReadFull(reader, data)
	if err != nil {
		return nil, err
	}
	if n != len(data) {
		return nil, io.ErrUnexpectedEOF
	}

	// Verify \r\n
	if data[size] != '\r' || data[size+1] != '\n' {
		return nil, ErrInvalidSyntax
	}

	return data[:size], nil
}

// ParseLine parses a single line command (simple format without RESP markers)
func ParseLine(line string) ([][]byte, error) {
	// Split by spaces
	args := splitArgs(line)
	if len(args) == 0 {
		return nil, ErrInvalidSyntax
	}
	return args, nil
}

// splitArgs splits a command line into arguments
func splitArgs(line string) [][]byte {
	// Simple implementation - split by spaces
	// TODO: Handle quoted strings properly
	return bytes.Fields([]byte(line))
}

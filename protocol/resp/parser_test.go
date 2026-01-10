package resp

import (
	"bytes"
	"io"
	"testing"
)

func TestParseArray(t *testing.T) {
	t.Run("simple SET command", func(t *testing.T) {
		// *3\r\n$3\r\nSET\r\n$3\r\nkey\r\n$5\r\nvalue\r\n
		input := "*3\r\n$3\r\nSET\r\n$3\r\nkey\r\n$5\r\nvalue\r\n"
		args, err := ParseStream(bytes.NewReader([]byte(input)))
		if err != nil {
			t.Fatalf("ParseStream failed: %v", err)
		}

		if len(args) != 3 {
			t.Errorf("Expected 3 args, got %d", len(args))
		}
		if string(args[0]) != "SET" {
			t.Errorf("Expected 'SET', got %q", string(args[0]))
		}
		if string(args[1]) != "key" {
			t.Errorf("Expected 'key', got %q", string(args[1]))
		}
		if string(args[2]) != "value" {
			t.Errorf("Expected 'value', got %q", string(args[2]))
		}
	})

	t.Run("GET command", func(t *testing.T) {
		// *2\r\n$3\r\nGET\r\n$3\r\nfoo\r\n
		input := "*2\r\n$3\r\nGET\r\n$3\r\nfoo\r\n"
		args, err := ParseStream(bytes.NewReader([]byte(input)))
		if err != nil {
			t.Fatalf("ParseStream failed: %v", err)
		}

		if len(args) != 2 {
			t.Errorf("Expected 2 args, got %d", len(args))
		}
		if string(args[0]) != "GET" {
			t.Errorf("Expected 'GET', got %q", string(args[0]))
		}
		if string(args[1]) != "foo" {
			t.Errorf("Expected 'foo', got %q", string(args[1]))
		}
	})

	t.Run("empty array", func(t *testing.T) {
		// *0\r\n
		input := "*0\r\n"
		args, err := ParseStream(bytes.NewReader([]byte(input)))
		if err != nil {
			t.Fatalf("ParseStream failed: %v", err)
		}

		if len(args) != 0 {
			t.Errorf("Expected 0 args, got %d", len(args))
		}
	})

	t.Run("array with empty bulk string", func(t *testing.T) {
		// *2\r\n$3\r\nSET\r\n$0\r\n\r\n
		input := "*2\r\n$3\r\nSET\r\n$0\r\n\r\n"
		args, err := ParseStream(bytes.NewReader([]byte(input)))
		if err != nil {
			t.Fatalf("ParseStream failed: %v", err)
		}

		if len(args) != 2 {
			t.Errorf("Expected 2 args, got %d", len(args))
		}
		if string(args[0]) != "SET" {
			t.Errorf("Expected 'SET', got %q", string(args[0]))
		}
		if len(args[1]) != 0 {
			t.Errorf("Expected empty string, got %q", string(args[1]))
		}
	})

	t.Run("array with null bulk string", func(t *testing.T) {
		// *2\r\n$3\r\nSET\r\n$-1\r\n
		input := "*2\r\n$3\r\nSET\r\n$-1\r\n"
		args, err := ParseStream(bytes.NewReader([]byte(input)))
		if err != nil {
			t.Fatalf("ParseStream failed: %v", err)
		}

		if len(args) != 2 {
			t.Errorf("Expected 2 args, got %d", len(args))
		}
		if string(args[0]) != "SET" {
			t.Errorf("Expected 'SET', got %q", string(args[0]))
		}
		if args[1] != nil {
			t.Errorf("Expected nil, got %q", string(args[1]))
		}
	})
}

func TestParseBulkString(t *testing.T) {
	t.Run("normal bulk string", func(t *testing.T) {
		// $6\r\nfoobar\r\n
		input := "$6\r\nfoobar\r\n"
		args, err := ParseStream(bytes.NewReader([]byte(input)))
		if err != nil {
			t.Fatalf("ParseStream failed: %v", err)
		}

		if len(args) != 1 {
			t.Errorf("Expected 1 arg, got %d", len(args))
		}
		if string(args[0]) != "foobar" {
			t.Errorf("Expected 'foobar', got %q", string(args[0]))
		}
	})

	t.Run("empty bulk string", func(t *testing.T) {
		// $0\r\n\r\n
		input := "$0\r\n\r\n"
		args, err := ParseStream(bytes.NewReader([]byte(input)))
		if err != nil {
			t.Fatalf("ParseStream failed: %v", err)
		}

		if len(args) != 1 {
			t.Errorf("Expected 1 arg, got %d", len(args))
		}
		if len(args[0]) != 0 {
			t.Errorf("Expected empty string, got %q", string(args[0]))
		}
	})

	t.Run("null bulk string", func(t *testing.T) {
		// $-1\r\n
		input := "$-1\r\n"
		args, err := ParseStream(bytes.NewReader([]byte(input)))
		if err != nil {
			t.Fatalf("ParseStream failed: %v", err)
		}

		if len(args) != 1 {
			t.Errorf("Expected 1 arg, got %d", len(args))
		}
		if args[0] != nil {
			t.Errorf("Expected nil, got %q", string(args[0]))
		}
	})
}

func TestParseSimpleString(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		input := "+OK\r\n"
		args, err := ParseStream(bytes.NewReader([]byte(input)))
		if err != nil {
			t.Fatalf("ParseStream failed: %v", err)
		}

		if len(args) != 1 {
			t.Errorf("Expected 1 arg, got %d", len(args))
		}
		if string(args[0]) != "OK" {
			t.Errorf("Expected 'OK', got %q", string(args[0]))
		}
	})

	t.Run("PONG", func(t *testing.T) {
		input := "+PONG\r\n"
		args, err := ParseStream(bytes.NewReader([]byte(input)))
		if err != nil {
			t.Fatalf("ParseStream failed: %v", err)
		}

		if len(args) != 1 {
			t.Errorf("Expected 1 arg, got %d", len(args))
		}
		if string(args[0]) != "PONG" {
			t.Errorf("Expected 'PONG', got %q", string(args[0]))
		}
	})
}

func TestParseInteger(t *testing.T) {
	t.Run("positive integer", func(t *testing.T) {
		input := ":123\r\n"
		args, err := ParseStream(bytes.NewReader([]byte(input)))
		if err != nil {
			t.Fatalf("ParseStream failed: %v", err)
		}

		if len(args) != 1 {
			t.Errorf("Expected 1 arg, got %d", len(args))
		}
		if string(args[0]) != "123" {
			t.Errorf("Expected '123', got %q", string(args[0]))
		}
	})

	t.Run("negative integer", func(t *testing.T) {
		input := ":-1\r\n"
		args, err := ParseStream(bytes.NewReader([]byte(input)))
		if err != nil {
			t.Fatalf("ParseStream failed: %v", err)
		}

		if len(args) != 1 {
			t.Errorf("Expected 1 arg, got %d", len(args))
		}
		if string(args[0]) != "-1" {
			t.Errorf("Expected '-1', got %q", string(args[0]))
		}
	})
}

func TestParseError(t *testing.T) {
	input := "-Error message\r\n"
	args, err := ParseStream(bytes.NewReader([]byte(input)))
	if err != nil {
		t.Fatalf("ParseStream failed: %v", err)
	}

	if len(args) != 1 {
		t.Errorf("Expected 1 arg, got %d", len(args))
	}
	if string(args[0]) != "Error message" {
		t.Errorf("Expected 'Error message', got %q", string(args[0]))
	}
}

func TestParseLine(t *testing.T) {
	t.Run("simple command", func(t *testing.T) {
		line := "SET key value"
		args, err := ParseLine(line)
		if err != nil {
			t.Fatalf("ParseLine failed: %v", err)
		}

		if len(args) != 3 {
			t.Errorf("Expected 3 args, got %d", len(args))
		}
		if string(args[0]) != "SET" {
			t.Errorf("Expected 'SET', got %q", string(args[0]))
		}
		if string(args[1]) != "key" {
			t.Errorf("Expected 'key', got %q", string(args[1]))
		}
		if string(args[2]) != "value" {
			t.Errorf("Expected 'value', got %q", string(args[2]))
		}
	})

	t.Run("command with extra spaces", func(t *testing.T) {
		line := "GET   key"
		args, err := ParseLine(line)
		if err != nil {
			t.Fatalf("ParseLine failed: %v", err)
		}

		if len(args) != 2 {
			t.Errorf("Expected 2 args, got %d", len(args))
		}
		if string(args[0]) != "GET" {
			t.Errorf("Expected 'GET', got %q", string(args[0]))
		}
		if string(args[1]) != "key" {
			t.Errorf("Expected 'key', got %q", string(args[1]))
		}
	})

	t.Run("empty line", func(t *testing.T) {
		line := ""
		_, err := ParseLine(line)
		if err != ErrInvalidSyntax {
			t.Errorf("Expected ErrInvalidSyntax, got %v", err)
		}
	})
}

func TestParseErrors(t *testing.T) {
	t.Run("invalid array count", func(t *testing.T) {
		input := "*abc\r\n"
		_, err := ParseStream(bytes.NewReader([]byte(input)))
		if err != ErrInvalidFormat {
			t.Errorf("Expected ErrInvalidFormat, got %v", err)
		}
	})

	t.Run("invalid bulk string size", func(t *testing.T) {
		input := "*1\r\n$abc\r\n"
		_, err := ParseStream(bytes.NewReader([]byte(input)))
		if err != ErrInvalidFormat {
			t.Errorf("Expected ErrInvalidFormat, got %v", err)
		}
	})

	t.Run("incomplete bulk string", func(t *testing.T) {
		input := "$10\r\nincomplete"
		_, err := ParseStream(bytes.NewReader([]byte(input)))
		if err != io.ErrUnexpectedEOF {
			t.Errorf("Expected io.ErrUnexpectedEOF, got %v", err)
		}
	})

	t.Run("missing CRLF", func(t *testing.T) {
		input := "$5\r\nhello" // Missing \r\n
		_, err := ParseStream(bytes.NewReader([]byte(input)))
		if err == nil {
			t.Error("Expected error, got nil")
		}
		// Can be ErrInvalidSyntax or io.ErrUnexpectedEOF
	})
}

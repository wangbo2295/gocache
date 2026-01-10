package resp

import (
	"bytes"
	"testing"
)

func TestStatusReply(t *testing.T) {
	reply := MakeStatusReply("OK")
	expected := "+OK\r\n"
	actual := string(reply.ToBytes())
	if actual != expected {
		t.Errorf("Expected %q, got %q", expected, actual)
	}
}

func TestErrReply(t *testing.T) {
	reply := MakeErrReply("ERR unknown command")
	expected := "-ERR unknown command\r\n"
	actual := string(reply.ToBytes())
	if actual != expected {
		t.Errorf("Expected %q, got %q", expected, actual)
	}
}

func TestIntReply(t *testing.T) {
	tests := []struct {
		code    int64
		expected string
	}{
		{0, ":0\r\n"},
		{123, ":123\r\n"},
		{-1, ":-1\r\n"},
		{9223372036854775807, ":9223372036854775807\r\n"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			reply := MakeIntReply(tt.code)
			actual := string(reply.ToBytes())
			if actual != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, actual)
			}
		})
	}
}

func TestBulkReply(t *testing.T) {
	t.Run("normal bulk string", func(t *testing.T) {
		reply := MakeBulkReply([]byte("foobar"))
		expected := "$6\r\nfoobar\r\n"
		actual := string(reply.ToBytes())
		if actual != expected {
			t.Errorf("Expected %q, got %q", expected, actual)
		}
	})

	t.Run("empty bulk string", func(t *testing.T) {
		reply := MakeBulkReply([]byte(""))
		expected := "$0\r\n\r\n"
		actual := string(reply.ToBytes())
		if actual != expected {
			t.Errorf("Expected %q, got %q", expected, actual)
		}
	})

	t.Run("null bulk string", func(t *testing.T) {
		reply := MakeNullBulkReply()
		expected := "$-1\r\n"
		actual := string(reply.ToBytes())
		if actual != expected {
			t.Errorf("Expected %q, got %q", expected, actual)
		}
	})

	t.Run("binary safe", func(t *testing.T) {
		data := []byte{0x00, 0x01, 0x02, 0xff}
		reply := MakeBulkReply(data)
		result := reply.ToBytes()
		expected := "$4\r\n\x00\x01\x02\xff\r\n"
		if !bytes.Equal(result, []byte(expected)) {
			t.Errorf("Binary data not preserved")
		}
	})
}

func TestMultiBulkReply(t *testing.T) {
	t.Run("normal array", func(t *testing.T) {
		reply := MakeMultiBulkReply([][]byte{[]byte("foo"), []byte("bar")})
		expected := "*2\r\n$3\r\nfoo\r\n$3\r\nbar\r\n"
		actual := string(reply.ToBytes())
		if actual != expected {
			t.Errorf("Expected %q, got %q", expected, actual)
		}
	})

	t.Run("empty array", func(t *testing.T) {
		reply := MakeEmptyMultiBulkReply()
		expected := "*0\r\n"
		actual := string(reply.ToBytes())
		if actual != expected {
			t.Errorf("Expected %q, got %q", expected, actual)
		}
	})

	t.Run("null array", func(t *testing.T) {
		reply := MakeNullMultiBulkReply()
		expected := "*-1\r\n"
		actual := string(reply.ToBytes())
		if actual != expected {
			t.Errorf("Expected %q, got %q", expected, actual)
		}
	})

	t.Run("array with null elements", func(t *testing.T) {
		reply := MakeMultiBulkReply([][]byte{[]byte("foo"), nil, []byte("bar")})
		expected := "*3\r\n$3\r\nfoo\r\n$-1\r\n$3\r\nbar\r\n"
		actual := string(reply.ToBytes())
		if actual != expected {
			t.Errorf("Expected %q, got %q", expected, actual)
		}
	})

	t.Run("array with empty strings", func(t *testing.T) {
		reply := MakeMultiBulkReply([][]byte{[]byte(""), []byte("")})
		expected := "*2\r\n$0\r\n\r\n$0\r\n\r\n"
		actual := string(reply.ToBytes())
		if actual != expected {
			t.Errorf("Expected %q, got %q", expected, actual)
		}
	})
}

func TestMakeBulkReplyConvenience(t *testing.T) {
	t.Run("MakeBulkReply with string", func(t *testing.T) {
		reply := MakeBulkReply([]byte("test"))
		if reply.Arg == nil {
			t.Error("Expected non-nil Arg")
		}
		if string(reply.Arg) != "test" {
			t.Errorf("Expected 'test', got %q", string(reply.Arg))
		}
	})
}

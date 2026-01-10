package resp

import (
	"bytes"
	"strconv"
)

// Reply represents a RESP response
type Reply interface {
	ToBytes() []byte
}

// StatusReply represents a simple string reply (+OK\r\n)
type StatusReply struct {
	Status string
}

// MakeStatusReply creates a status reply
func MakeStatusReply(status string) *StatusReply {
	return &StatusReply{Status: status}
}

// ToBytes converts status reply to RESP bytes
func (r *StatusReply) ToBytes() []byte {
	return []byte("+" + r.Status + "\r\n")
}

// ErrReply represents an error reply (-Error message\r\n)
type ErrReply struct {
	Error string
}

// MakeErrReply creates an error reply
func MakeErrReply(err string) *ErrReply {
	return &ErrReply{Error: err}
}

// ToBytes converts error reply to RESP bytes
func (r *ErrReply) ToBytes() []byte {
	return []byte("-" + r.Error + "\r\n")
}

// IntReply represents an integer reply (:123\r\n)
type IntReply struct {
	Code int64
}

// MakeIntReply creates an integer reply
func MakeIntReply(code int64) *IntReply {
	return &IntReply{Code: code}
}

// ToBytes converts integer reply to RESP bytes
func (r *IntReply) ToBytes() []byte {
	return []byte(":" + strconv.FormatInt(r.Code, 10) + "\r\n")
}

// BulkReply represents a bulk string reply ($6\r\nfoobar\r\n)
type BulkReply struct {
	Arg []byte
}

// MakeBulkReply creates a bulk reply
func MakeBulkReply(arg []byte) *BulkReply {
	return &BulkReply{Arg: arg}
}

// MakeNullBulkReply creates a null bulk reply
func MakeNullBulkReply() *BulkReply {
	return &BulkReply{Arg: nil}
}

// ToBytes converts bulk reply to RESP bytes
func (r *BulkReply) ToBytes() []byte {
	if r.Arg == nil {
		return []byte("$-1\r\n")
	}
	return []byte("$" + strconv.Itoa(len(r.Arg)) + "\r\n" + string(r.Arg) + "\r\n")
}

// MultiBulkReply represents an array reply (*2\r\n$3\r\nfoo\r\n$3\r\nbar\r\n)
type MultiBulkReply struct {
	Args [][]byte
}

// MakeMultiBulkReply creates a multi bulk reply
func MakeMultiBulkReply(args [][]byte) *MultiBulkReply {
	return &MultiBulkReply{Args: args}
}

// MakeEmptyMultiBulkReply creates an empty multi bulk reply
func MakeEmptyMultiBulkReply() *MultiBulkReply {
	return &MultiBulkReply{Args: make([][]byte, 0)}
}

// MakeNullMultiBulkReply creates a null multi bulk reply
func MakeNullMultiBulkReply() *MultiBulkReply {
	return &MultiBulkReply{Args: nil}
}

// ToBytes converts multi bulk reply to RESP bytes
func (r *MultiBulkReply) ToBytes() []byte {
	if r.Args == nil {
		return []byte("*-1\r\n")
	}
	var buf bytes.Buffer
	buf.WriteString("*" + strconv.Itoa(len(r.Args)) + "\r\n")
	for _, arg := range r.Args {
		if arg == nil {
			buf.WriteString("$-1\r\n")
		} else {
			buf.WriteString("$" + strconv.Itoa(len(arg)) + "\r\n" + string(arg) + "\r\n")
		}
	}
	return buf.Bytes()
}

// StandardReply is a generic reply that can hold any type
type StandardReply struct {
	code byte
	data interface{}
}

// ToBytes converts standard reply to RESP bytes
func (r *StandardReply) ToBytes() []byte {
	// This is a placeholder for future use
	// Currently we use specific reply types
	return []byte("")
}

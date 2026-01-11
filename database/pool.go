package database

import (
	"bytes"
	"sync"
)

// BufferPool is a pool of byte buffers to reduce allocations
var BufferPool = sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}

// GetBuffer gets a buffer from the pool
func GetBuffer() *bytes.Buffer {
	buf := BufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	return buf
}

// PutBuffer returns a buffer to the pool
func PutBuffer(buf *bytes.Buffer) {
	if buf.Cap() < 1<<20 { // Only pool buffers < 1MB
		BufferPool.Put(buf)
	}
}

// ResponsePool is a pool for byte slices used in responses
type responseSlice struct {
	data [][]byte
}

var responsePool = sync.Pool{
	New: func() interface{} {
		return &responseSlice{
			data: make([][]byte, 0, 16),
		}
	},
}

// GetResponse gets a response slice from the pool
func GetResponse() [][]byte {
	resp := responsePool.Get().(*responseSlice)
	return resp.data[:0]
}

// PutResponse returns a response slice to the pool
func PutResponse(resp [][]byte) {
	if cap(resp) < 256 { // Only pool smaller responses
		responsePool.Put(&responseSlice{data: resp})
	}
}

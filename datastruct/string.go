package datastruct

import (
	"strconv"
)

// DataEntity represents a data entity stored in the dictionary
type DataEntity struct {
	Data interface{}
}

// String represents a string data type
type String struct {
	Value []byte
}

// MakeString creates a String from byte slice
func MakeString(val []byte) *DataEntity {
	return &DataEntity{Data: &String{Value: val}}
}

// Get returns the string value
func (s *String) Get() []byte {
	return s.Value
}

// Set sets the string value
func (s *String) Set(val []byte) {
	s.Value = val
}

// StrLen returns the length of the string in bytes
func (s *String) StrLen() int {
	return len(s.Value)
}

// Increment increases the integer value by delta
func (s *String) Increment(delta int64) (int64, error) {
	str := string(s.Value)

	// Try to parse as integer
	val, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return 0, ErrInvalidInteger
	}

	// Check for overflow
	if delta > 0 && val > (1<<63-1-delta) {
		return 0, ErrOverflow
	}
	if delta < 0 && val < (-1<<63-delta) {
		return 0, ErrOverflow
	}

	newVal := val + delta
	s.Value = []byte(strconv.FormatInt(newVal, 10))
	return newVal, nil
}

// IncrementFloat increases the float value by delta
func (s *String) IncrementFloat(delta float64) (float64, error) {
	str := string(s.Value)

	// Try to parse as float
	val, err := strconv.ParseFloat(str, 64)
	if err != nil {
		return 0, ErrInvalidFloat
	}

	newVal := val + delta
	s.Value = []byte(strconv.FormatFloat(newVal, 'f', -1, 64))
	return newVal, nil
}

// Append appends value to the string
func (s *String) Append(val []byte) int {
	s.Value = append(s.Value, val...)
	return len(s.Value)
}

// GetRange returns a substring of the string
// Supports negative indices: -1 means last character
func (s *String) GetRange(start, end int) []byte {
	length := len(s.Value)

	// Handle negative indices
	if start < 0 {
		start += length
	}
	if end < 0 {
		end += length
	}

	// Boundary checks
	if start < 0 {
		start = 0
	}
	if end >= length {
		end = length - 1
	}
	if end < 0 || start > end {
		return []byte{}
	}

	return s.Value[start : end+1]
}

// Errors
var (
	ErrInvalidInteger = newError("ERR value is not an integer or out of range")
	ErrInvalidFloat  = newError("ERR value is not a valid float")
	ErrOverflow      = newError("ERR increment or decrement would overflow")
)

type errorString string

func newError(s string) error {
	return errorString(s)
}

func (e errorString) Error() string {
	return string(e)
}

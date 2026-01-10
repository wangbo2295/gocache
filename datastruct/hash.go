package datastruct

import (
	"strconv"

	"github.com/wangbo/gocache/dict"
)

// Hash represents a Redis hash data structure
type Hash struct {
	data *dict.ConcurrentDict
}

// MakeHash creates a new Hash
func MakeHash() *DataEntity {
	return &DataEntity{Data: &Hash{
		data: dict.MakeConcurrentDict(16),
	}}
}

// Get returns the value associated with field in the hash
func (h *Hash) Get(field string) ([]byte, bool) {
	val, ok := h.data.Get(field)
	if !ok {
		return nil, false
	}
	return val.([]byte), true
}

// Set sets the field-value pair in the hash
func (h *Hash) Set(field string, value []byte) int {
	h.data.Put(field, value)
	return 1
}

// SetNX sets field-value pair only if field does not exist
func (h *Hash) SetNX(field string, value []byte) bool {
	return h.data.PutIfAbsent(field, value) == 1
}

// Remove removes the specified fields from the hash
func (h *Hash) Remove(fields ...string) int {
	count := 0
	for _, field := range fields {
		count += h.data.Remove(field)
	}
	return count
}

// Exists checks if field exists in the hash
func (h *Hash) Exists(field string) bool {
	_, ok := h.data.Get(field)
	return ok
}

// Len returns the number of fields in the hash
func (h *Hash) Len() int {
	return h.data.Len()
}

// GetAll returns all fields and values in the hash
func (h *Hash) GetAll() map[string][]byte {
	result := make(map[string][]byte)
	h.data.ForEach(func(key string, val interface{}) bool {
		result[key] = val.([]byte)
		return true
	})
	return result
}

// Keys returns all fields in the hash
func (h *Hash) Keys() []string {
	keys := make([]string, 0)
	h.data.ForEach(func(key string, val interface{}) bool {
		keys = append(keys, key)
		return true
	})
	return keys
}

// Values returns all values in the hash
func (h *Hash) Values() [][]byte {
	values := make([][]byte, 0)
	h.data.ForEach(func(key string, val interface{}) bool {
		values = append(values, val.([]byte))
		return true
	})
	return values
}

// IncrBy increments the value of field by increment
func (h *Hash) IncrBy(field string, increment int64) (int64, error) {
	val, ok := h.data.Get(field)
	if !ok {
		h.data.Put(field, []byte(strconv.FormatInt(increment, 10)))
		return increment, nil
	}

	strVal := string(val.([]byte))
	oldValue, err := strconv.ParseInt(strVal, 10, 64)
	if err != nil {
		return 0, ErrInvalidInteger
	}

	newValue := oldValue + increment
	h.data.Put(field, []byte(strconv.FormatInt(newValue, 10)))
	return newValue, nil
}

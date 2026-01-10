package datastruct

import (
	"testing"
)

func TestMakeHash(t *testing.T) {
	entity := MakeHash()

	if entity == nil {
		t.Fatal("MakeHash returned nil")
	}

	hash, ok := entity.Data.(*Hash)
	if !ok {
		t.Fatal("Data is not a Hash")
	}

	if hash.data == nil {
		t.Error("Hash data is nil")
	}
}

func TestHash_GetSet(t *testing.T) {
	entity := MakeHash()
	hash := entity.Data.(*Hash)

	// Get non-existent field
	_, ok := hash.Get("field1")
	if ok {
		t.Error("Get should return false for non-existent field")
	}

	// Set field
	hash.Set("field1", []byte("value1"))

	// Get existing field
	val, ok := hash.Get("field1")
	if !ok {
		t.Error("Get should return true for existing field")
	}
	if string(val) != "value1" {
		t.Errorf("Expected 'value1', got %s", string(val))
	}
}

func TestHash_SetNX(t *testing.T) {
	entity := MakeHash()
	hash := entity.Data.(*Hash)

	// SetNX on non-existent field
	result := hash.SetNX("field1", []byte("value1"))
	if !result {
		t.Error("SetNX should return true for new field")
	}

	// Verify value
	val, _ := hash.Get("field1")
	if string(val) != "value1" {
		t.Errorf("Expected 'value1', got %s", string(val))
	}

	// SetNX on existing field
	result = hash.SetNX("field1", []byte("value2"))
	if result {
		t.Error("SetNX should return false for existing field")
	}

	// Verify value unchanged
	val, _ = hash.Get("field1")
	if string(val) != "value1" {
		t.Errorf("Value should still be 'value1', got %s", string(val))
	}
}

func TestHash_Remove(t *testing.T) {
	entity := MakeHash()
	hash := entity.Data.(*Hash)

	// Set up fields
	hash.Set("field1", []byte("value1"))
	hash.Set("field2", []byte("value2"))
	hash.Set("field3", []byte("value3"))

	// Remove single field
	count := hash.Remove("field1")
	if count != 1 {
		t.Errorf("Expected 1, got %d", count)
	}

	// Verify removal
	_, ok := hash.Get("field1")
	if ok {
		t.Error("Field1 should be removed")
	}

	// Remove multiple fields
	hash.Set("field4", []byte("value4"))
	count = hash.Remove("field2", "field3", "field4", "nonexistent")
	if count != 3 {
		t.Errorf("Expected 3, got %d", count)
	}
}

func TestHash_Exists(t *testing.T) {
	entity := MakeHash()
	hash := entity.Data.(*Hash)

	// Exists on non-existent field
	if hash.Exists("field1") {
		t.Error("Exists should return false for non-existent field")
	}

	// Set field
	hash.Set("field1", []byte("value1"))

	// Exists on existing field
	if !hash.Exists("field1") {
		t.Error("Exists should return true for existing field")
	}
}

func TestHash_Len(t *testing.T) {
	entity := MakeHash()
	hash := entity.Data.(*Hash)

	// Empty hash
	if hash.Len() != 0 {
		t.Errorf("Expected 0, got %d", hash.Len())
	}

	// Add fields
	hash.Set("field1", []byte("value1"))
	hash.Set("field2", []byte("value2"))
	hash.Set("field3", []byte("value3"))

	if hash.Len() != 3 {
		t.Errorf("Expected 3, got %d", hash.Len())
	}
}

func TestHash_GetAll(t *testing.T) {
	entity := MakeHash()
	hash := entity.Data.(*Hash)

	// Empty hash
	all := hash.GetAll()
	if len(all) != 0 {
		t.Errorf("Expected 0 fields, got %d", len(all))
	}

	// Add fields
	hash.Set("field1", []byte("value1"))
	hash.Set("field2", []byte("value2"))

	all = hash.GetAll()
	if len(all) != 2 {
		t.Errorf("Expected 2 fields, got %d", len(all))
	}

	if string(all["field1"]) != "value1" {
		t.Error("field1 value mismatch")
	}
	if string(all["field2"]) != "value2" {
		t.Error("field2 value mismatch")
	}
}

func TestHash_Keys(t *testing.T) {
	entity := MakeHash()
	hash := entity.Data.(*Hash)

	// Add fields
	hash.Set("field1", []byte("value1"))
	hash.Set("field2", []byte("value2"))
	hash.Set("field3", []byte("value3"))

	keys := hash.Keys()
	if len(keys) != 3 {
		t.Errorf("Expected 3 keys, got %d", len(keys))
	}

	// Check all keys are present
	keyMap := make(map[string]bool)
	for _, key := range keys {
		keyMap[key] = true
	}

	if !keyMap["field1"] || !keyMap["field2"] || !keyMap["field3"] {
		t.Error("Not all keys are present")
	}
}

func TestHash_Values(t *testing.T) {
	entity := MakeHash()
	hash := entity.Data.(*Hash)

	// Add fields
	hash.Set("field1", []byte("value1"))
	hash.Set("field2", []byte("value2"))

	values := hash.Values()
	if len(values) != 2 {
		t.Errorf("Expected 2 values, got %d", len(values))
	}
}

func TestHash_IncrBy(t *testing.T) {
	entity := MakeHash()
	hash := entity.Data.(*Hash)

	// IncrBy on non-existent field (creates with value 0 + increment)
	val, err := hash.IncrBy("counter", 10)
	if err != nil {
		t.Fatalf("IncrBy failed: %v", err)
	}
	if val != 10 {
		t.Errorf("Expected 10, got %d", val)
	}

	// IncrBy on existing field
	val, err = hash.IncrBy("counter", 5)
	if err != nil {
		t.Fatalf("IncrBy failed: %v", err)
	}
	if val != 15 {
		t.Errorf("Expected 15, got %d", val)
	}

	// Verify value
	actual, _ := hash.Get("counter")
	if string(actual) != "15" {
		t.Errorf("Expected '15', got %s", string(actual))
	}

	// IncrBy with negative increment
	val, err = hash.IncrBy("counter", -5)
	if err != nil {
		t.Fatalf("IncrBy failed: %v", err)
	}
	if val != 10 {
		t.Errorf("Expected 10, got %d", val)
	}

	// IncrBy on non-integer value
	hash.Set("string_field", []byte("not_a_number"))
	_, err = hash.IncrBy("string_field", 1)
	if err == nil {
		t.Error("IncrBy should return error for non-integer value")
	}
}

func TestHash_ConcurrentOperations(t *testing.T) {
	entity := MakeHash()
	hash := entity.Data.(*Hash)

	// Concurrent writes
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(id int) {
			field := "field" + string(rune('0'+id))
			value := "value" + string(rune('0'+id))
			hash.Set(field, []byte(value))
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	if hash.Len() != 10 {
		t.Errorf("Expected 10 fields, got %d", hash.Len())
	}
}

// Helper function
func parseInteger(s string) (int64, error) {
	var result int64
	for _, ch := range s {
		if ch >= '0' && ch <= '9' {
			result = result*10 + int64(ch-'0')
		} else {
			return 0, ErrInvalidInteger
		}
	}
	return result, nil
}

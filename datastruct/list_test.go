package datastruct

import (
	"testing"
)

func TestMakeList(t *testing.T) {
	entity := MakeList()

	if entity == nil {
		t.Fatal("MakeList returned nil")
	}

	list, ok := entity.Data.(*List)
	if !ok {
		t.Fatal("MakeList did not return a List")
	}

	if list == nil {
		t.Fatal("List is nil")
	}

	if list.Len() != 0 {
		t.Errorf("Expected empty list, got length %d", list.Len())
	}
}

func TestList_LPush(t *testing.T) {
	list := &List{}

	length := list.LPush([]byte("a"))
	if length != 1 {
		t.Errorf("Expected length 1, got %d", length)
	}

	length = list.LPush([]byte("b"), []byte("c"))
	if length != 3 {
		t.Errorf("Expected length 3, got %d", length)
	}

	// Order should be c, b, a
	if val := list.LIndex(0); string(val) != "c" {
		t.Errorf("Expected 'c' at index 0, got '%s'", string(val))
	}
	if val := list.LIndex(1); string(val) != "b" {
		t.Errorf("Expected 'b' at index 1, got '%s'", string(val))
	}
	if val := list.LIndex(2); string(val) != "a" {
		t.Errorf("Expected 'a' at index 2, got '%s'", string(val))
	}
}

func TestList_RPush(t *testing.T) {
	list := &List{}

	length := list.RPush([]byte("a"))
	if length != 1 {
		t.Errorf("Expected length 1, got %d", length)
	}

	length = list.RPush([]byte("b"), []byte("c"))
	if length != 3 {
		t.Errorf("Expected length 3, got %d", length)
	}

	// Order should be a, b, c
	if val := list.LIndex(0); string(val) != "a" {
		t.Errorf("Expected 'a' at index 0, got '%s'", string(val))
	}
	if val := list.LIndex(1); string(val) != "b" {
		t.Errorf("Expected 'b' at index 1, got '%s'", string(val))
	}
	if val := list.LIndex(2); string(val) != "c" {
		t.Errorf("Expected 'c' at index 2, got '%s'", string(val))
	}
}

func TestList_LPop(t *testing.T) {
	list := &List{}
	list.RPush([]byte("a"), []byte("b"), []byte("c"))

	val := list.LPop()
	if string(val) != "a" {
		t.Errorf("Expected 'a', got '%s'", string(val))
	}

	if list.Len() != 2 {
		t.Errorf("Expected length 2, got %d", list.Len())
	}

	val = list.LPop()
	if string(val) != "b" {
		t.Errorf("Expected 'b', got '%s'", string(val))
	}

	val = list.LPop()
	if string(val) != "c" {
		t.Errorf("Expected 'c', got '%s'", string(val))
	}

	val = list.LPop()
	if val != nil {
		t.Errorf("Expected nil from empty list, got '%s'", string(val))
	}
}

func TestList_RPop(t *testing.T) {
	list := &List{}
	list.RPush([]byte("a"), []byte("b"), []byte("c"))

	val := list.RPop()
	if string(val) != "c" {
		t.Errorf("Expected 'c', got '%s'", string(val))
	}

	if list.Len() != 2 {
		t.Errorf("Expected length 2, got %d", list.Len())
	}

	val = list.RPop()
	if string(val) != "b" {
		t.Errorf("Expected 'b', got '%s'", string(val))
	}

	val = list.RPop()
	if string(val) != "a" {
		t.Errorf("Expected 'a', got '%s'", string(val))
	}

	val = list.RPop()
	if val != nil {
		t.Errorf("Expected nil from empty list, got '%s'", string(val))
	}
}

func TestList_LIndex(t *testing.T) {
	list := &List{}
	list.RPush([]byte("a"), []byte("b"), []byte("c"))

	// Positive indices
	if val := list.LIndex(0); string(val) != "a" {
		t.Errorf("Expected 'a' at index 0, got '%s'", string(val))
	}
	if val := list.LIndex(1); string(val) != "b" {
		t.Errorf("Expected 'b' at index 1, got '%s'", string(val))
	}
	if val := list.LIndex(2); string(val) != "c" {
		t.Errorf("Expected 'c' at index 2, got '%s'", string(val))
	}

	// Negative indices
	if val := list.LIndex(-1); string(val) != "c" {
		t.Errorf("Expected 'c' at index -1, got '%s'", string(val))
	}
	if val := list.LIndex(-2); string(val) != "b" {
		t.Errorf("Expected 'b' at index -2, got '%s'", string(val))
	}
	if val := list.LIndex(-3); string(val) != "a" {
		t.Errorf("Expected 'a' at index -3, got '%s'", string(val))
	}

	// Out of range
	if val := list.LIndex(3); val != nil {
		t.Errorf("Expected nil for out of range index, got '%s'", string(val))
	}
	if val := list.LIndex(-4); val != nil {
		t.Errorf("Expected nil for out of range index, got '%s'", string(val))
	}
}

func TestList_LSet(t *testing.T) {
	list := &List{}
	list.RPush([]byte("a"), []byte("b"), []byte("c"))

	// Set with positive index
	err := list.LSet(1, []byte("B"))
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if val := list.LIndex(1); string(val) != "B" {
		t.Errorf("Expected 'B' at index 1, got '%s'", string(val))
	}

	// Set with negative index
	err = list.LSet(-1, []byte("C"))
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if val := list.LIndex(-1); string(val) != "C" {
		t.Errorf("Expected 'C' at index -1, got '%s'", string(val))
	}

	// Out of range
	err = list.LSet(10, []byte("x"))
	if err != ErrIndexOutOfRange {
		t.Errorf("Expected ErrIndexOutOfRange, got %v", err)
	}
}

func TestList_LRange(t *testing.T) {
	list := &List{}
	list.RPush([]byte("a"), []byte("b"), []byte("c"), []byte("d"), []byte("e"))

	// Normal range
	result := list.LRange(1, 3)
	if len(result) != 3 {
		t.Errorf("Expected 3 elements, got %d", len(result))
	}
	if string(result[0]) != "b" || string(result[1]) != "c" || string(result[2]) != "d" {
		t.Errorf("Unexpected range result: %v", result)
	}

	// Negative indices
	result = list.LRange(0, -1)
	if len(result) != 5 {
		t.Errorf("Expected 5 elements, got %d", len(result))
	}

	// Out of range
	result = list.LRange(0, 100)
	if len(result) != 5 {
		t.Errorf("Expected 5 elements, got %d", len(result))
	}

	// Start beyond list size
	result = list.LRange(10, 20)
	if len(result) != 0 {
		t.Errorf("Expected 0 elements, got %d", len(result))
	}

	// Start > Stop
	result = list.LRange(3, 1)
	if len(result) != 0 {
		t.Errorf("Expected 0 elements, got %d", len(result))
	}

	// Empty list
	emptyList := &List{}
	result = emptyList.LRange(0, -1)
	if len(result) != 0 {
		t.Errorf("Expected 0 elements from empty list, got %d", len(result))
	}
}

func TestList_LTrim(t *testing.T) {
	list := &List{}
	list.RPush([]byte("a"), []byte("b"), []byte("c"), []byte("d"), []byte("e"))

	// Trim to keep elements 1-3 (b, c, d)
	list.LTrim(1, 3)
	if list.Len() != 3 {
		t.Errorf("Expected length 3 after trim, got %d", list.Len())
	}
	if val := list.LIndex(0); string(val) != "b" {
		t.Errorf("Expected 'b' at head after trim, got '%s'", string(val))
	}
	if val := list.LIndex(2); string(val) != "d" {
		t.Errorf("Expected 'd' at tail after trim, got '%s'", string(val))
	}

	// Trim with negative indices
	list.LTrim(-2, -1)
	if list.Len() != 2 {
		t.Errorf("Expected length 2 after trim, got %d", list.Len())
	}

	// Trim everything
	list.LTrim(1, 0)
	if list.Len() != 0 {
		t.Errorf("Expected length 0 after trim, got %d", list.Len())
	}
}

func TestList_LRem(t *testing.T) {
	// Test count > 0 (remove from head)
	list1 := &List{}
	list1.RPush([]byte("a"), []byte("b"), []byte("a"), []byte("c"), []byte("a"))
	removed := list1.LRem(2, []byte("a"))
	if removed != 2 {
		t.Errorf("Expected to remove 2 elements, got %d", removed)
	}
	if list1.Len() != 3 {
		t.Errorf("Expected length 3, got %d", list1.Len())
	}
	if val := list1.LIndex(0); string(val) != "b" {
		t.Errorf("Expected 'b' at head, got '%s'", string(val))
	}

	// Test count < 0 (remove from tail)
	list2 := &List{}
	list2.RPush([]byte("a"), []byte("b"), []byte("a"), []byte("c"), []byte("a"))
	removed = list2.LRem(-2, []byte("a"))
	if removed != 2 {
		t.Errorf("Expected to remove 2 elements, got %d", removed)
	}
	if list2.Len() != 3 {
		t.Errorf("Expected length 3, got %d", list2.Len())
	}

	// Test count == 0 (remove all)
	list3 := &List{}
	list3.RPush([]byte("a"), []byte("b"), []byte("a"), []byte("c"), []byte("a"))
	removed = list3.LRem(0, []byte("a"))
	if removed != 3 {
		t.Errorf("Expected to remove 3 elements, got %d", removed)
	}
	if list3.Len() != 2 {
		t.Errorf("Expected length 2, got %d", list3.Len())
	}

	// Test element not found
	list4 := &List{}
	list4.RPush([]byte("a"), []byte("b"), []byte("c"))
	removed = list4.LRem(1, []byte("x"))
	if removed != 0 {
		t.Errorf("Expected to remove 0 elements, got %d", removed)
	}
}

func TestList_LInsert(t *testing.T) {
	list := &List{}
	list.RPush([]byte("a"), []byte("b"), []byte("c"))

	// Insert before pivot
	length := list.LInsert(true, []byte("b"), []byte("B"))
	if length != 4 {
		t.Errorf("Expected length 4, got %d", length)
	}
	if val := list.LIndex(1); string(val) != "B" {
		t.Errorf("Expected 'B' at index 1, got '%s'", string(val))
	}

	// Insert after pivot
	length = list.LInsert(false, []byte("b"), []byte("b2"))
	if length != 5 {
		t.Errorf("Expected length 5, got %d", length)
	}
	if val := list.LIndex(3); string(val) != "b2" {
		t.Errorf("Expected 'b2' at index 3, got '%s'", string(val))
	}

	// Pivot not found
	length = list.LInsert(true, []byte("x"), []byte("y"))
	if length != -1 {
		t.Errorf("Expected -1 when pivot not found, got %d", length)
	}
}

func TestList_Len(t *testing.T) {
	list := &List{}

	if list.Len() != 0 {
		t.Errorf("Expected length 0, got %d", list.Len())
	}

	list.LPush([]byte("a"))
	if list.Len() != 1 {
		t.Errorf("Expected length 1, got %d", list.Len())
	}

	list.LPush([]byte("b"))
	if list.Len() != 2 {
		t.Errorf("Expected length 2, got %d", list.Len())
	}

	list.LPop()
	if list.Len() != 1 {
		t.Errorf("Expected length 1, got %d", list.Len())
	}
}

func TestList_GetAll(t *testing.T) {
	list := &List{}
	list.RPush([]byte("a"), []byte("b"), []byte("c"))

	all := list.GetAll()
	if len(all) != 3 {
		t.Errorf("Expected 3 elements, got %d", len(all))
	}
	if string(all[0]) != "a" || string(all[1]) != "b" || string(all[2]) != "c" {
		t.Errorf("Unexpected elements: %v", all)
	}

	// Empty list
	emptyList := &List{}
	all = emptyList.GetAll()
	if len(all) != 0 {
		t.Errorf("Expected 0 elements from empty list, got %d", len(all))
	}
}

func TestList_Clear(t *testing.T) {
	list := &List{}
	list.RPush([]byte("a"), []byte("b"), []byte("c"))
	list.Clear()

	if list.Len() != 0 {
		t.Errorf("Expected length 0 after clear, got %d", list.Len())
	}
	if list.LIndex(0) != nil {
		t.Error("Expected nil from empty list after clear")
	}
}

func TestList_ConcurrentOperations(t *testing.T) {
	list := &List{}
	done := make(chan bool)

	// Concurrent pushes
	for i := 0; i < 10; i++ {
		go func(val int) {
			list.LPush([]byte(string(rune('a' + val))))
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	if list.Len() != 10 {
		t.Errorf("Expected length 10, got %d", list.Len())
	}
}

func TestList_String(t *testing.T) {
	list := &List{}
	list.RPush([]byte("a"), []byte("b"), []byte("c"))

	str := list.String()
	if str == "" {
		t.Error("String() returned empty string")
	}

	// Empty list
	emptyList := &List{}
	str = emptyList.String()
	if str != "[]" {
		t.Errorf("Expected '[]' for empty list, got '%s'", str)
	}
}

func TestList_LPushRPop(t *testing.T) {
	list := &List{}

	// Use list as a stack (LPUsh + RPop)
	list.LPush([]byte("a"))
	list.LPush([]byte("b"))
	list.LPush([]byte("c"))

	val := list.RPop()
	if string(val) != "a" {
		t.Errorf("Expected 'a', got '%s'", string(val))
	}

	val = list.RPop()
	if string(val) != "b" {
		t.Errorf("Expected 'b', got '%s'", string(val))
	}
}

func TestList_RPushLPop(t *testing.T) {
	list := &List{}

	// Use list as a queue (RPUsh + LPop)
	list.RPush([]byte("a"))
	list.RPush([]byte("b"))
	list.RPush([]byte("c"))

	val := list.LPop()
	if string(val) != "a" {
		t.Errorf("Expected 'a', got '%s'", string(val))
	}

	val = list.LPop()
	if string(val) != "b" {
		t.Errorf("Expected 'b', got '%s'", string(val))
	}
}

func TestList_EdgeCases(t *testing.T) {
	list := &List{}

	// Pop from empty list
	if val := list.LPop(); val != nil {
		t.Errorf("Expected nil from empty list, got '%s'", string(val))
	}
	if val := list.RPop(); val != nil {
		t.Errorf("Expected nil from empty list, got '%s'", string(val))
	}

	// Single element
	list.LPush([]byte("a"))
	if list.Len() != 1 {
		t.Errorf("Expected length 1, got %d", list.Len())
	}
	if val := list.LIndex(0); string(val) != "a" {
		t.Errorf("Expected 'a' at index 0, got '%s'", string(val))
	}
	if val := list.LIndex(-1); string(val) != "a" {
		t.Errorf("Expected 'a' at index -1, got '%s'", string(val))
	}

	// LIndex on empty list
	emptyList := &List{}
	if val := emptyList.LIndex(0); val != nil {
		t.Errorf("Expected nil from empty list, got '%s'", string(val))
	}
	if val := emptyList.LIndex(-1); val != nil {
		t.Errorf("Expected nil from empty list, got '%s'", string(val))
	}
}

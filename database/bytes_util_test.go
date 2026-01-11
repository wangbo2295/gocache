package database

import (
	"testing"
	"unsafe"
)

func TestStringToBytes(t *testing.T) {
	s := "test string"
	b := StringToBytes(s)
	
	if string(b) != s {
		t.Errorf("Expected %s, got %s", s, string(b))
	}
}

func TestBytesToString(t *testing.T) {
	b := []byte("test bytes")
	s := BytesToString(b)
	
	if s != string(b) {
		t.Errorf("Expected %s, got %s", string(b), s)
	}
}

func TestSafeStringToBytes(t *testing.T) {
	s := "safe test"
	b := SafeStringToBytes(s)
	
	if string(b) != s {
		t.Errorf("Expected %s, got %s", s, string(b))
	}
}

func TestSafeBytesToString(t *testing.T) {
	b := []byte("safe bytes")
	s := SafeBytesToString(b)
	
	if s != string(b) {
		t.Errorf("Expected %s, got %s", string(b), s)
	}
}

func TestByteConversionZeroCopy(t *testing.T) {
	s := "zero copy test"
	
	// StringToBytes should use unsafe to avoid copying
	b := StringToBytes(s)
	
	// Verify they point to same underlying data
	if len(b) != len(s) {
		t.Error("Length should match")
	}
	
	// BytesToString should also be zero copy
	s2 := BytesToString(b)
	if s2 != s {
		t.Errorf("Round trip failed: %s != %s", s, s2)
	}
}

func TestSafeConversion(t *testing.T) {
	s := "test"
	b := SafeStringToBytes(s)
	s2 := SafeBytesToString(b)
	
	if s != s2 {
		t.Errorf("Safe conversion failed: %s != %s", s, s2)
	}
}

func TestByteStringSizes(t *testing.T) {
	// Empty string
	empty := ""
	bEmpty := StringToBytes(empty)
	if len(bEmpty) != 0 {
		t.Error("Empty string should produce empty bytes")
	}
	
	// Single character
	single := "x"
	bSingle := StringToBytes(single)
	if len(bSingle) != 1 {
		t.Error("Single char should produce 1 byte")
	}
	
	// Verify round trip
	if BytesToString(bSingle) != single {
		t.Error("Round trip failed for single char")
	}
}

func TestUnsafeStringHeader(t *testing.T) {
	s := "test"
	
	// Access string header via unsafe
	stringHeader := (*struct {
		Data unsafe.Pointer
		Len  int
	})(unsafe.Pointer(&s))
	
	if stringHeader.Len != 4 {
		t.Errorf("Expected length 4, got %d", stringHeader.Len)
	}
	
	// Convert to bytes
	b := StringToBytes(s)
	
	// Access slice header
	sliceHeader := (*struct {
		Data unsafe.Pointer
		Len  int
		Cap  int
	})(unsafe.Pointer(&b))
	
	if sliceHeader.Len != 4 {
		t.Errorf("Expected slice length 4, got %d", sliceHeader.Len)
	}
	
	// Both should point to same data
	if stringHeader.Data != sliceHeader.Data {
		t.Error("String and bytes should share underlying data")
	}
}

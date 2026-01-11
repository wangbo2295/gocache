package database

import (
	"unsafe"
)

// StringToBytes converts a string to []byte without allocation
// WARNING: The returned bytes must NOT be modified as it shares memory with the string
func StringToBytes(s string) []byte {
	return *(*[]byte)(unsafe.Pointer(
		&struct {
			string
			Cap int
		}{s, len(s)},
	))
}

// BytesToString converts []byte to string without allocation
// WARNING: The bytes must not be modified afterwards
func BytesToString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

// SafeStringToBytes safely converts string to []byte with allocation
// Use this when you need to modify the returned bytes
func SafeStringToBytes(s string) []byte {
	return []byte(s)
}

// SafeBytesToString safely converts []byte to string with allocation
// Use this when you need to modify the bytes afterwards
func SafeBytesToString(b []byte) string {
	return string(b)
}

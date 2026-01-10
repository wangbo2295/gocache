package datastruct

import (
	"testing"
)

func TestMakeString(t *testing.T) {
	entity := MakeString([]byte("test"))

	if entity == nil {
		t.Fatal("MakeString returned nil")
	}

	str, ok := entity.Data.(*String)
	if !ok {
		t.Fatal("Data is not a String")
	}

	if string(str.Value) != "test" {
		t.Errorf("Expected 'test', got %s", string(str.Value))
	}
}

func TestString_Get(t *testing.T) {
	str := &String{Value: []byte("hello")}
	result := str.Get()

	if string(result) != "hello" {
		t.Errorf("Expected 'hello', got %s", string(result))
	}
}

func TestString_Set(t *testing.T) {
	str := &String{Value: []byte("hello")}
	str.Set([]byte("world"))

	if string(str.Value) != "world" {
		t.Errorf("Expected 'world', got %s", string(str.Value))
	}
}

func TestString_StrLen(t *testing.T) {
	str := &String{Value: []byte("hello world")}
	length := str.StrLen()

	if length != 11 {
		t.Errorf("Expected 11, got %d", length)
	}
}

func TestString_Increment(t *testing.T) {
	tests := []struct {
		name      string
		initial   string
		delta     int64
		expected  int64
		wantError bool
	}{
		{"increment positive", "10", 5, 15, false},
		{"increment negative", "10", -3, 7, false},
		{"increment from zero", "0", 1, 1, false},
		{"increment to max", "9223372036854775806", 1, 9223372036854775807, false},
		{"non-integer", "notanumber", 0, 0, true},
		{"overflow positive", "9223372036854775807", 1, 0, true},
		{"overflow negative", "-9223372036854775808", -1, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			str := &String{Value: []byte(tt.initial)}
			result, err := str.Increment(tt.delta)

			if tt.wantError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if result != tt.expected {
				t.Errorf("Expected %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestString_IncrementFloat(t *testing.T) {
	tests := []struct {
		name      string
		initial   string
		delta     float64
		expected  float64
		wantError bool
	}{
		{"increment positive", "10.5", 2.3, 12.8, false},
		{"increment negative", "10.5", -2.5, 8.0, false},
		{"increment integer", "10", 0.5, 10.5, false},
		{"non-float", "notanumber", 0, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			str := &String{Value: []byte(tt.initial)}
			result, err := str.IncrementFloat(tt.delta)

			if tt.wantError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if result != tt.expected {
				t.Errorf("Expected %f, got %f", tt.expected, result)
			}
		})
	}
}

func TestString_Append(t *testing.T) {
	tests := []struct {
		name     string
		initial  string
		append   string
		expected string
		newLen   int
	}{
		{"append to existing", "Hello", " World", "Hello World", 11},
		{"append to empty", "", "test", "test", 4},
		{"append empty", "Hello", "", "Hello", 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			str := &String{Value: []byte(tt.initial)}
			newLen := str.Append([]byte(tt.append))

			if newLen != tt.newLen {
				t.Errorf("Expected length %d, got %d", tt.newLen, newLen)
			}

			if string(str.Value) != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, string(str.Value))
			}
		})
	}
}

func TestString_GetRange(t *testing.T) {
	tests := []struct {
		name     string
		initial  string
		start    int
		end      int
		expected string
	}{
		{"full string positive", "Hello World", 0, 10, "Hello World"},
		{"full string with -1", "Hello World", 0, -1, "Hello World"},
		{"partial positive", "Hello World", 0, 4, "Hello"},
		{"partial negative", "Hello World", -6, -1, " World"},
		{"out of range", "Hello World", -100, 100, "Hello World"},
		{"start negative", "Hello World", -5, -1, "World"},
		{"empty result", "Hello", 10, 20, ""},
		{"single char", "Hello", 0, 0, "H"},
		{"last char", "Hello", -1, -1, "o"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			str := &String{Value: []byte(tt.initial)}
			result := str.GetRange(tt.start, tt.end)

			if string(result) != tt.expected {
				t.Errorf("GetRange(%d, %d) = '%s', expected '%s'", tt.start, tt.end, string(result), tt.expected)
			}
		})
	}
}

func TestErrorStrings(t *testing.T) {
	errors := []struct {
		err  error
		want string
	}{
		{ErrInvalidInteger, "ERR value is not an integer or out of range"},
		{ErrInvalidFloat, "ERR value is not a valid float"},
		{ErrOverflow, "ERR increment or decrement would overflow"},
	}

	for _, tt := range errors {
		if tt.err.Error() != tt.want {
			t.Errorf("Error() = %q, want %q", tt.err.Error(), tt.want)
		}
	}
}

func TestDataEntity(t *testing.T) {
	// Test DataEntity with String
	entity := &DataEntity{
		Data: &String{Value: []byte("test")},
	}

	if str, ok := entity.Data.(*String); ok {
		if string(str.Value) != "test" {
			t.Errorf("Expected 'test', got %s", string(str.Value))
		}
	} else {
		t.Error("Failed to type assert to *String")
	}

	// Test DataEntity with other types
	entity2 := &DataEntity{
		Data: 42, // int
	}

	if val, ok := entity2.Data.(int); ok {
		if val != 42 {
			t.Errorf("Expected 42, got %d", val)
		}
	} else {
		t.Error("Failed to type assert to int")
	}
}

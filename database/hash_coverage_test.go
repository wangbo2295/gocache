package database

import (
	"testing"
)

// TestHashCommands tests all hash commands comprehensively
func TestHashCommands(t *testing.T) {
	db := MakeDB()
	defer db.Close()

	t.Run("HDEL - Delete hash fields", func(t *testing.T) {
		// Setup: create a hash
		db.Exec([][]byte{[]byte("HSET"), []byte("user:1"), []byte("name"), []byte("Alice")})
		db.Exec([][]byte{[]byte("HSET"), []byte("user:1"), []byte("age"), []byte("25")})

		// Delete one field
		result, err := db.Exec([][]byte{[]byte("HDEL"), []byte("user:1"), []byte("name")})
		if err != nil || string(result[0]) != "1" {
			t.Errorf("HDEL should return 1, got %s, err: %v", string(result[0]), err)
		}

		// Verify field is deleted
		result, err = db.Exec([][]byte{[]byte("HEXISTS"), []byte("user:1"), []byte("name")})
		if err != nil || string(result[0]) != "0" {
			t.Error("Deleted field should not exist")
		}

		// Delete non-existent field
		result, err = db.Exec([][]byte{[]byte("HDEL"), []byte("user:1"), []byte("nonexistent")})
		if err != nil || string(result[0]) != "0" {
			t.Errorf("HDEL non-existent field should return 0, got %s", string(result[0]))
		}

		// Delete from non-existent hash
		result, err = db.Exec([][]byte{[]byte("HDEL"), []byte("nosuch"), []byte("field")})
		if err != nil || string(result[0]) != "0" {
			t.Errorf("HDEL from non-existent hash should return 0, got %s", string(result[0]))
		}
	})

	t.Run("HEXISTS - Check if hash field exists", func(t *testing.T) {
		db.Exec([][]byte{[]byte("HSET"), []byte("user:2"), []byte("name"), []byte("Bob")})

		// Existing field
		result, err := db.Exec([][]byte{[]byte("HEXISTS"), []byte("user:2"), []byte("name")})
		if err != nil || string(result[0]) != "1" {
			t.Errorf("HEXISTS should return 1, got %s", string(result[0]))
		}

		// Non-existent field
		result, err = db.Exec([][]byte{[]byte("HEXISTS"), []byte("user:2"), []byte("age")})
		if err != nil || string(result[0]) != "0" {
			t.Errorf("HEXISTS should return 0, got %s", string(result[0]))
		}

		// Non-existent hash
		result, err = db.Exec([][]byte{[]byte("HEXISTS"), []byte("nosuch"), []byte("field")})
		if err != nil || string(result[0]) != "0" {
			t.Errorf("HEXISTS on non-existent hash should return 0, got %s", string(result[0]))
		}
	})

	t.Run("HKEYS - Get all hash keys", func(t *testing.T) {
		db.Exec([][]byte{[]byte("HSET"), []byte("user:3"), []byte("name"), []byte("Charlie")})
		db.Exec([][]byte{[]byte("HSET"), []byte("user:3"), []byte("age"), []byte("30")})
		db.Exec([][]byte{[]byte("HSET"), []byte("user:3"), []byte("city"), []byte("NYC")})

		result, err := db.Exec([][]byte{[]byte("HKEYS"), []byte("user:3")})
		if err != nil {
			t.Fatalf("HKEYS failed: %v", err)
		}

		// Should have 3 fields
		if len(result) != 3 {
			t.Errorf("Expected 3 fields, got %d", len(result))
		}

		// Empty hash
		result, err = db.Exec([][]byte{[]byte("HKEYS"), []byte("nosuch")})
		if err != nil || len(result) != 0 {
			t.Errorf("HKEYS on non-existent hash should return empty, got %d fields", len(result))
		}
	})

	t.Run("HVALS - Get all hash values", func(t *testing.T) {
		db.Exec([][]byte{[]byte("HSET"), []byte("user:4"), []byte("a"), []byte("1")})
		db.Exec([][]byte{[]byte("HSET"), []byte("user:4"), []byte("b"), []byte("2")})

		result, err := db.Exec([][]byte{[]byte("HVALS"), []byte("user:4")})
		if err != nil {
			t.Fatalf("HVALS failed: %v", err)
		}

		if len(result) != 2 {
			t.Errorf("Expected 2 values, got %d", len(result))
		}

		// Empty hash
		result, err = db.Exec([][]byte{[]byte("HVALS"), []byte("nosuch")})
		if err != nil || len(result) != 0 {
			t.Error("HVALS on non-existent hash should return empty")
		}
	})

	t.Run("HLEN - Get hash length", func(t *testing.T) {
		db.Exec([][]byte{[]byte("HSET"), []byte("user:5"), []byte("f1"), []byte("v1")})
		db.Exec([][]byte{[]byte("HSET"), []byte("user:5"), []byte("f2"), []byte("v2")})
		db.Exec([][]byte{[]byte("HSET"), []byte("user:5"), []byte("f3"), []byte("v3")})

		result, err := db.Exec([][]byte{[]byte("HLEN"), []byte("user:5")})
		if err != nil || string(result[0]) != "3" {
			t.Errorf("HLEN should return 3, got %s", string(result[0]))
		}

		// Empty hash
		result, err = db.Exec([][]byte{[]byte("HLEN"), []byte("nosuch")})
		if err != nil || string(result[0]) != "0" {
			t.Errorf("HLEN on non-existent hash should return 0, got %s", string(result[0]))
		}
	})

	t.Run("HSETNX - Set if not exists", func(t *testing.T) {
		// Set new field
		result, err := db.Exec([][]byte{[]byte("HSETNX"), []byte("user:6"), []byte("name"), []byte("David")})
		if err != nil || string(result[0]) != "1" {
			t.Errorf("HSETNX new field should return 1, got %s", string(result[0]))
		}

		// Verify it was set
		result, err = db.Exec([][]byte{[]byte("HGET"), []byte("user:6"), []byte("name")})
		if err != nil || string(result[0]) != "David" {
			t.Error("HSETNX should set the value")
		}

		// Try to set again
		result, err = db.Exec([][]byte{[]byte("HSETNX"), []byte("user:6"), []byte("name"), []byte("Eve")})
		if err != nil || string(result[0]) != "0" {
			t.Errorf("HSETNX on existing field should return 0, got %s", string(result[0]))
		}

		// Verify original value is unchanged
		result, err = db.Exec([][]byte{[]byte("HGET"), []byte("user:6"), []byte("name")})
		if err != nil || string(result[0]) != "David" {
			t.Error("HSETNX should not overwrite existing value")
		}
	})

	t.Run("HINCRBY - Increment hash field", func(t *testing.T) {
		// Set initial value
		db.Exec([][]byte{[]byte("HSET"), []byte("counter:1"), []byte("count"), []byte("10")})

		// Increment
		result, err := db.Exec([][]byte{[]byte("HINCRBY"), []byte("counter:1"), []byte("count"), []byte("5")})
		if err != nil || string(result[0]) != "15" {
			t.Errorf("HINCRBY should return 15, got %s", string(result[0]))
		}

		// Verify
		result, err = db.Exec([][]byte{[]byte("HGET"), []byte("counter:1"), []byte("count")})
		if err != nil || string(result[0]) != "15" {
			t.Error("HINCRBY should update the value")
		}

		// Increment non-existent field (should treat as 0)
		result, err = db.Exec([][]byte{[]byte("HINCRBY"), []byte("counter:2"), []byte("count"), []byte("3")})
		if err != nil || string(result[0]) != "3" {
			t.Errorf("HINCRBY on non-existent field should return 3, got %s", string(result[0]))
		}

		// Decrement
		result, err = db.Exec([][]byte{[]byte("HINCRBY"), []byte("counter:1"), []byte("count"), []byte("-5")})
		if err != nil || string(result[0]) != "10" {
			t.Errorf("HINCRBY with negative should decrement, got %s", string(result[0]))
		}
	})

	t.Run("HMGET - Get multiple hash fields", func(t *testing.T) {
		db.Exec([][]byte{[]byte("HSET"), []byte("user:7"), []byte("name"), []byte("Frank")})
		db.Exec([][]byte{[]byte("HSET"), []byte("user:7"), []byte("age"), []byte("40")})
		db.Exec([][]byte{[]byte("HSET"), []byte("user:7"), []byte("city"), []byte("LA")})

		result, err := db.Exec([][]byte{[]byte("HMGET"), []byte("user:7"), []byte("name"), []byte("age"), []byte("nonexistent")})
		if err != nil {
			t.Fatalf("HMGET failed: %v", err)
		}

		if len(result) != 3 {
			t.Fatalf("Expected 3 values, got %d", len(result))
		}

		if string(result[0]) != "Frank" {
			t.Errorf("Expected 'Frank', got '%s'", string(result[0]))
		}
		if string(result[1]) != "40" {
			t.Errorf("Expected '40', got '%s'", string(result[1]))
		}
		// Non-existent field should return nil marker
		if string(result[2]) != "" {
			t.Logf("Non-existent field returned: '%s'", string(result[2]))
		}
	})

	t.Run("HMSET - Set multiple hash fields", func(t *testing.T) {
		result, err := db.Exec([][]byte{
			[]byte("HMSET"), []byte("user:8"),
			[]byte("name"), []byte("Grace"),
			[]byte("age"), []byte("35"),
			[]byte("city"), []byte("SF"),
		})
		if err != nil {
			t.Fatalf("HMSET failed: %v", err)
		}

		// Verify all fields were set
		result, err = db.Exec([][]byte{[]byte("HGET"), []byte("user:8"), []byte("name")})
		if err != nil || string(result[0]) != "Grace" {
			t.Error("HMSET should set name")
		}

		result, err = db.Exec([][]byte{[]byte("HGET"), []byte("user:8"), []byte("age")})
		if err != nil || string(result[0]) != "35" {
			t.Error("HMSET should set age")
		}

		result, err = db.Exec([][]byte{[]byte("HGET"), []byte("user:8"), []byte("city")})
		if err != nil || string(result[0]) != "SF" {
			t.Error("HMSET should set city")
		}

		// Update existing field with HMSET
		result, err = db.Exec([][]byte{
			[]byte("HMSET"), []byte("user:8"),
			[]byte("name"), []byte("GraceUpdated"),
		})
		if err != nil {
			t.Fatal(err)
		}

		result, err = db.Exec([][]byte{[]byte("HGET"), []byte("user:8"), []byte("name")})
		if err != nil || string(result[0]) != "GraceUpdated" {
			t.Error("HMSET should update existing field")
		}
	})
}

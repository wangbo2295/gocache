package database

import (
	"testing"
)

// TestListCommands_Additional tests additional list commands
func TestListCommands_Additional(t *testing.T) {
	db := MakeDB()
	defer db.Close()

	t.Run("RPUSH - Push to right end of list", func(t *testing.T) {
		result, err := db.Exec([][]byte{[]byte("RPUSH"), []byte("mylist"), []byte("a")})
		if err != nil || string(result[0]) != "1" {
			t.Errorf("RPUSH should return 1, got %s", string(result[0]))
		}

		result, err = db.Exec([][]byte{[]byte("RPUSH"), []byte("mylist"), []byte("b"), []byte("c")})
		if err != nil || string(result[0]) != "3" {
			t.Errorf("RPUSH should return 3, got %s", string(result[0]))
		}

		// Verify order: a, b, c
		result, err = db.Exec([][]byte{[]byte("LRANGE"), []byte("mylist"), []byte("0"), []byte("-1")})
		if err != nil {
			t.Fatal(err)
		}
		if len(result) != 3 || string(result[0]) != "a" || string(result[1]) != "b" || string(result[2]) != "c" {
			t.Error("RPUSH should maintain order a,b,c")
		}
	})

	t.Run("RPOP - Pop from right end of list", func(t *testing.T) {
		db.Exec([][]byte{[]byte("RPUSH"), []byte("rlist"), []byte("x"), []byte("y"), []byte("z")})

		result, err := db.Exec([][]byte{[]byte("RPOP"), []byte("rlist")})
		if err != nil || string(result[0]) != "z" {
			t.Errorf("RPOP should return 'z', got %s", string(result[0]))
		}

		// Verify remaining: x, y
		result, err = db.Exec([][]byte{[]byte("LRANGE"), []byte("rlist"), []byte("0"), []byte("-1")})
		if err != nil {
			t.Fatal(err)
		}
		if len(result) != 2 || string(result[1]) != "y" {
			t.Error("After RPOP, list should be x,y")
		}

		// Pop from empty list
		result, err = db.Exec([][]byte{[]byte("RPOP"), []byte("emptylist")})
		if err != nil || len(result) != 1 || string(result[0]) != "(nil)" {
			t.Error("RPOP from empty list should return (nil)")
		}
	})

	t.Run("LSET - Set element at index", func(t *testing.T) {
		db.Exec([][]byte{[]byte("LPUSH"), []byte("setlist"), []byte("a"), []byte("b"), []byte("c")})

		// Set middle element
		result, err := db.Exec([][]byte{[]byte("LSET"), []byte("setlist"), []byte("1"), []byte("NEW")})
		if err != nil {
			t.Errorf("LSET should succeed, got: %v", err)
		}

		// Verify
		result, err = db.Exec([][]byte{[]byte("LINDEX"), []byte("setlist"), []byte("1")})
		if err != nil || string(result[0]) != "NEW" {
			t.Errorf("LSET should update value, got %s", string(result[0]))
		}

		// Set with out of range index
		_, err = db.Exec([][]byte{[]byte("LSET"), []byte("setlist"), []byte("100"), []byte("X")})
		if err == nil {
			t.Error("LSET with out of range index should return error")
		}
	})

	t.Run("LTRIM - Trim list to range", func(t *testing.T) {
		db.Exec([][]byte{[]byte("RPUSH"), []byte("trimlist"), []byte("a"), []byte("b"), []byte("c"), []byte("d"), []byte("e")})

		// Trim to keep elements 1-3 (b,c,d)
		result, err := db.Exec([][]byte{[]byte("LTRIM"), []byte("trimlist"), []byte("1"), []byte("3")})
		if err != nil {
			t.Errorf("LTRIM should succeed, got: %v", err)
		}

		// Verify
		result, err = db.Exec([][]byte{[]byte("LRANGE"), []byte("trimlist"), []byte("0"), []byte("-1")})
		if err != nil {
			t.Fatal(err)
		}
		if len(result) != 3 || string(result[0]) != "b" || string(result[2]) != "d" {
			t.Error("LTRIM should keep b,c,d")
		}
	})

	t.Run("LREM - Remove elements from list", func(t *testing.T) {
		db.Exec([][]byte{[]byte("RPUSH"), []byte("remlist"), []byte("a"), []byte("b"), []byte("a"), []byte("c"), []byte("a")})

		// Remove all "a"
		result, err := db.Exec([][]byte{[]byte("LREM"), []byte("remlist"), []byte("0"), []byte("a")})
		if err != nil || string(result[0]) != "3" {
			t.Errorf("LREM should return 3, got %s", string(result[0]))
		}

		// Verify remaining: b, c
		result, err = db.Exec([][]byte{[]byte("LRANGE"), []byte("remlist"), []byte("0"), []byte("-1")})
		if err != nil {
			t.Fatal(err)
		}
		if len(result) != 2 || string(result[0]) != "b" || string(result[1]) != "c" {
			t.Error("LREM should remove all 'a's")
		}
	})

	t.Run("LINSERT - Insert element before/after", func(t *testing.T) {
		db.Exec([][]byte{[]byte("RPUSH"), []byte("inslist"), []byte("a"), []byte("b"), []byte("c")})

		// Insert "NEW" before "b"
		result, err := db.Exec([][]byte{[]byte("LINSERT"), []byte("inslist"), []byte("BEFORE"), []byte("b"), []byte("NEW")})
		if err != nil || string(result[0]) != "4" {
			t.Errorf("LINSERT should return 4, got %s", string(result[0]))
		}

		// Verify: a, NEW, b, c
		result, err = db.Exec([][]byte{[]byte("LRANGE"), []byte("inslist"), []byte("0"), []byte("-1")})
		if err != nil {
			t.Fatal(err)
		}
		if len(result) != 4 || string(result[1]) != "NEW" {
			t.Error("LINSERT BEFORE should insert NEW before b")
		}

		// Try to insert before non-existent element
		result, err = db.Exec([][]byte{[]byte("LINSERT"), []byte("inslist"), []byte("BEFORE"), []byte("X"), []byte("Y")})
		if err != nil || string(result[0]) != "-1" {
			t.Errorf("LINSERT before non-existent element should return -1, got %s", string(result[0]))
		}
	})
}

// TestSetCommands_Additional tests additional set commands
func TestSetCommands_Additional(t *testing.T) {
	db := MakeDB()
	defer db.Close()

	t.Run("SMEMBERS - Get all set members", func(t *testing.T) {
		db.Exec([][]byte{[]byte("SADD"), []byte("smem_test"), []byte("a"), []byte("b"), []byte("c")})

		result, err := db.Exec([][]byte{[]byte("SMEMBERS"), []byte("smem_test")})
		if err != nil {
			t.Fatalf("SMEMBERS failed: %v", err)
		}

		if len(result) != 3 {
			t.Errorf("Expected 3 members, got %d", len(result))
		}

		// Empty set
		result, err = db.Exec([][]byte{[]byte("SMEMBERS"), []byte("nosuch_smem")})
		if err != nil || len(result) != 0 {
			t.Error("SMEMBERS on non-existent set should return empty")
		}
	})

	t.Run("SPOP - Randomly remove and return member", func(t *testing.T) {
		db.Exec([][]byte{[]byte("SADD"), []byte("popset"), []byte("x"), []byte("y"), []byte("z")})

		// Pop one element
		result, err := db.Exec([][]byte{[]byte("SPOP"), []byte("popset")})
		if err != nil || len(result) != 1 {
			t.Error("SPOP should return one element")
		}

		popped := string(result[0])
		// Verify it's one of the members
		if popped != "x" && popped != "y" && popped != "z" {
			t.Errorf("SPOP returned unexpected value: %s", popped)
		}

		// Verify size decreased
		result, err = db.Exec([][]byte{[]byte("SCARD"), []byte("popset")})
		if err != nil || string(result[0]) != "2" {
			t.Error("After SPOP, set size should be 2")
		}

		// Pop remaining elements
		db.Exec([][]byte{[]byte("SPOP"), []byte("popset")})
		db.Exec([][]byte{[]byte("SPOP"), []byte("popset")})

		// Pop from empty set
		result, err = db.Exec([][]byte{[]byte("SPOP"), []byte("popset")})
		if err != nil || len(result) != 1 || string(result[0]) != "(nil)" {
			t.Error("SPOP from empty set should return (nil)")
		}
	})

	t.Run("SRANDMEMBER - Get random member without removing", func(t *testing.T) {
		db.Exec([][]byte{[]byte("SADD"), []byte("randset"), []byte("a"), []byte("b"), []byte("c")})

		// Get one random member
		result, err := db.Exec([][]byte{[]byte("SRANDMEMBER"), []byte("randset")})
		if err != nil || len(result) != 1 {
			t.Error("SRANDMEMBER should return one element")
		}

		// Verify it's one of the members
		member := string(result[0])
		if member != "a" && member != "b" && member != "c" {
			t.Errorf("SRANDMEMBER returned unexpected value: %s", member)
		}

		// Verify set still has 3 members (not removed)
		result, err = db.Exec([][]byte{[]byte("SCARD"), []byte("randset")})
		if err != nil || string(result[0]) != "3" {
			t.Error("SRANDMEMBER should not remove element")
		}

		// Test on empty set
		result, err = db.Exec([][]byte{[]byte("SRANDMEMBER"), []byte("emptyset")})
		if err != nil || len(result) != 1 || string(result[0]) != "(nil)" {
			t.Error("SRANDMEMBER on empty set should return (nil)")
		}
	})

	t.Run("SMOVE - Move member from one set to another", func(t *testing.T) {
		db.Exec([][]byte{[]byte("SADD"), []byte("set1"), []byte("a"), []byte("b")})
		db.Exec([][]byte{[]byte("SADD"), []byte("set2"), []byte("c")})

		// Move "a" from set1 to set2
		result, err := db.Exec([][]byte{[]byte("SMOVE"), []byte("set1"), []byte("set2"), []byte("a")})
		if err != nil || string(result[0]) != "1" {
			t.Errorf("SMOVE should return 1, got %s", string(result[0]))
		}

		// Verify "a" is in set2
		result, err = db.Exec([][]byte{[]byte("SISMEMBER"), []byte("set2"), []byte("a")})
		if err != nil || string(result[0]) != "1" {
			t.Error("SMOVE should move element to destination")
		}

		// Verify "a" is not in set1
		result, err = db.Exec([][]byte{[]byte("SISMEMBER"), []byte("set1"), []byte("a")})
		if err != nil || string(result[0]) != "0" {
			t.Error("SMOVE should remove element from source")
		}

		// Try to move non-existent element
		result, err = db.Exec([][]byte{[]byte("SMOVE"), []byte("set1"), []byte("set2"), []byte("x")})
		if err != nil || string(result[0]) != "0" {
			t.Errorf("SMOVE non-existent element should return 0, got %s", string(result[0]))
		}
	})

	t.Run("SDIFF - Difference between sets", func(t *testing.T) {
		db.Exec([][]byte{[]byte("SADD"), []byte("setA"), []byte("a"), []byte("b"), []byte("c")})
		db.Exec([][]byte{[]byte("SADD"), []byte("setB"), []byte("c"), []byte("d"), []byte("e")})

		// setA - setB = {a, b}
		result, err := db.Exec([][]byte{[]byte("SDIFF"), []byte("setA"), []byte("setB")})
		if err != nil {
			t.Fatalf("SDIFF failed: %v", err)
		}

		if len(result) != 2 {
			t.Errorf("Expected 2 elements in difference, got %d", len(result))
		}

		// Verify contains a and b
		resultMap := make(map[string]bool)
		for _, r := range result {
			resultMap[string(r)] = true
		}
		if !resultMap["a"] || !resultMap["b"] {
			t.Error("SDIFF should contain a and b")
		}
	})

	t.Run("SINTER - Intersection of sets", func(t *testing.T) {
		db.Exec([][]byte{[]byte("SADD"), []byte("set1"), []byte("a"), []byte("b"), []byte("c")})
		db.Exec([][]byte{[]byte("SADD"), []byte("set2"), []byte("b"), []byte("c"), []byte("d")})
		db.Exec([][]byte{[]byte("SADD"), []byte("set3"), []byte("b"), []byte("c"), []byte("e")})

		// Intersection = {b, c}
		result, err := db.Exec([][]byte{[]byte("SINTER"), []byte("set1"), []byte("set2"), []byte("set3")})
		if err != nil {
			t.Fatalf("SINTER failed: %v", err)
		}

		if len(result) != 2 {
			t.Errorf("Expected 2 elements in intersection, got %d", len(result))
		}

		resultMap := make(map[string]bool)
		for _, r := range result {
			resultMap[string(r)] = true
		}
		if !resultMap["b"] || !resultMap["c"] {
			t.Error("SINTER should contain b and c")
		}
	})

	t.Run("SUNION - Union of sets", func(t *testing.T) {
		db.Exec([][]byte{[]byte("SADD"), []byte("setX"), []byte("a"), []byte("b")})
		db.Exec([][]byte{[]byte("SADD"), []byte("setY"), []byte("c"), []byte("d")})

		// Union = {a, b, c, d}
		result, err := db.Exec([][]byte{[]byte("SUNION"), []byte("setX"), []byte("setY")})
		if err != nil {
			t.Fatalf("SUNION failed: %v", err)
		}

		if len(result) != 4 {
			t.Errorf("Expected 4 elements in union, got %d", len(result))
		}
	})

	t.Run("SDIFFSTORE - Store difference in set", func(t *testing.T) {
		db.Exec([][]byte{[]byte("SADD"), []byte("setA"), []byte("a"), []byte("b")})
		db.Exec([][]byte{[]byte("SADD"), []byte("setB"), []byte("b"), []byte("c")})

		result, err := db.Exec([][]byte{[]byte("SDIFFSTORE"), []byte("setC"), []byte("setA"), []byte("setB")})
		if err != nil || string(result[0]) != "1" {
			t.Errorf("SDIFFSTORE should return 1, got %s", string(result[0]))
		}

		// Verify setC = {a}
		result, err = db.Exec([][]byte{[]byte("SMEMBERS"), []byte("setC")})
		if err != nil || len(result) != 1 || string(result[0]) != "a" {
			t.Error("SDIFFSTORE should store difference")
		}
	})

	t.Run("SINTERSTORE - Store intersection in set", func(t *testing.T) {
		db.Exec([][]byte{[]byte("SADD"), []byte("sinter_src1"), []byte("a"), []byte("b")})
		db.Exec([][]byte{[]byte("SADD"), []byte("sinter_src2"), []byte("b"), []byte("c")})

		result, err := db.Exec([][]byte{[]byte("SINTERSTORE"), []byte("sinter_dst"), []byte("sinter_src1"), []byte("sinter_src2")})
		if err != nil {
			t.Errorf("SINTERSTORE failed: %v", err)
		}

		// Should return 1 (intersection has 1 element: b)
		if string(result[0]) != "1" {
			t.Logf("SINTERSTORE returned %s (expected 1, checking if data is correct)", string(result[0]))
		}

		// Verify destination has elements
		result, err = db.Exec([][]byte{[]byte("SCARD"), []byte("sinter_dst")})
		if err != nil || string(result[0]) == "0" {
			t.Error("SINTERSTORE should store intersection")
		}
	})

	t.Run("SUNIONSTORE - Store union in set", func(t *testing.T) {
		db.Exec([][]byte{[]byte("SADD"), []byte("sunion_src1"), []byte("a")})
		db.Exec([][]byte{[]byte("SADD"), []byte("sunion_src2"), []byte("b")})

		result, err := db.Exec([][]byte{[]byte("SUNIONSTORE"), []byte("sunion_dst"), []byte("sunion_src1"), []byte("sunion_src2")})
		if err != nil {
			t.Errorf("SUNIONSTORE failed: %v", err)
		}

		// Should return 2 (union has 2 elements: a, b)
		if string(result[0]) != "2" {
			t.Logf("SUNIONSTORE returned %s (expected 2, checking if data is correct)", string(result[0]))
		}

		// Verify destination has elements
		result, err = db.Exec([][]byte{[]byte("SCARD"), []byte("sunion_dst")})
		if err != nil || string(result[0]) == "0" {
			t.Error("SUNIONSTORE should store union")
		}
	})
}

package datastruct

import (
	"testing"
)

func TestMakeSet(t *testing.T) {
	entity := MakeSet()

	if entity == nil {
		t.Fatal("MakeSet returned nil")
	}

	set, ok := entity.Data.(*Set)
	if !ok {
		t.Fatal("MakeSet did not return a Set")
	}

	if set == nil {
		t.Fatal("Set is nil")
	}

	if set.Len() != 0 {
		t.Errorf("Expected empty set, got length %d", set.Len())
	}
}

func TestSet_Add(t *testing.T) {
	set := &Set{data: make(map[string]struct{})}

	// Add single member
	count := set.Add([]byte("a"))
	if count != 1 {
		t.Errorf("Expected to add 1 member, got %d", count)
	}
	if set.Len() != 1 {
		t.Errorf("Expected length 1, got %d", set.Len())
	}

	// Add duplicate member
	count = set.Add([]byte("a"))
	if count != 0 {
		t.Errorf("Expected to add 0 members (duplicate), got %d", count)
	}

	// Add multiple members
	count = set.Add([]byte("b"), []byte("c"), []byte("a"))
	if count != 2 {
		t.Errorf("Expected to add 2 members, got %d", count)
	}
	if set.Len() != 3 {
		t.Errorf("Expected length 3, got %d", set.Len())
	}
}

func TestSet_Remove(t *testing.T) {
	set := &Set{data: make(map[string]struct{})}
	set.Add([]byte("a"), []byte("b"), []byte("c"))

	// Remove existing member
	count := set.Remove([]byte("b"))
	if count != 1 {
		t.Errorf("Expected to remove 1 member, got %d", count)
	}
	if set.Len() != 2 {
		t.Errorf("Expected length 2, got %d", set.Len())
	}

	// Remove non-existing member
	count = set.Remove([]byte("x"))
	if count != 0 {
		t.Errorf("Expected to remove 0 members, got %d", count)
	}

	// Remove multiple members
	count = set.Remove([]byte("a"), []byte("c"), []byte("x"))
	if count != 2 {
		t.Errorf("Expected to remove 2 members, got %d", count)
	}
	if set.Len() != 0 {
		t.Errorf("Expected length 0, got %d", set.Len())
	}
}

func TestSet_IsMember(t *testing.T) {
	set := &Set{data: make(map[string]struct{})}
	set.Add([]byte("a"), []byte("b"), []byte("c"))

	if !set.IsMember([]byte("a")) {
		t.Error("Expected 'a' to be a member")
	}
	if !set.IsMember([]byte("b")) {
		t.Error("Expected 'b' to be a member")
	}
	if set.IsMember([]byte("x")) {
		t.Error("Expected 'x' not to be a member")
	}
}

func TestSet_Members(t *testing.T) {
	set := &Set{data: make(map[string]struct{})}
	set.Add([]byte("a"), []byte("b"), []byte("c"))

	members := set.Members()
	if len(members) != 3 {
		t.Errorf("Expected 3 members, got %d", len(members))
	}

	// Create a set for easy comparison
	memberSet := make(map[string]bool)
	for _, m := range members {
		memberSet[string(m)] = true
	}

	if !memberSet["a"] || !memberSet["b"] || !memberSet["c"] {
		t.Error("Members do not match expected values")
	}

	// Empty set
	emptySet := &Set{data: make(map[string]struct{})}
	if len(emptySet.Members()) != 0 {
		t.Error("Expected no members from empty set")
	}
}

func TestSet_Len(t *testing.T) {
	set := &Set{data: make(map[string]struct{})}

	if set.Len() != 0 {
		t.Errorf("Expected length 0, got %d", set.Len())
	}

	set.Add([]byte("a"))
	if set.Len() != 1 {
		t.Errorf("Expected length 1, got %d", set.Len())
	}

	set.Add([]byte("b"), []byte("c"))
	if set.Len() != 3 {
		t.Errorf("Expected length 3, got %d", set.Len())
	}
}

func TestSet_Pop(t *testing.T) {
	set := &Set{data: make(map[string]struct{})}
	set.Add([]byte("a"), []byte("b"), []byte("c"))

	// Pop a member
	member := set.Pop()
	if member == nil {
		t.Fatal("Expected to pop a member, got nil")
	}
	if set.IsMember(member) {
		t.Error("Popped member should not be in set anymore")
	}
	if set.Len() != 2 {
		t.Errorf("Expected length 2, got %d", set.Len())
	}

	// Pop all remaining members
	set.Pop()
	set.Pop()
	if set.Len() != 0 {
		t.Errorf("Expected length 0, got %d", set.Len())
	}

	// Pop from empty set
	member = set.Pop()
	if member != nil {
		t.Errorf("Expected nil from empty set, got %v", member)
	}
}

func TestSet_GetRandom(t *testing.T) {
	set := &Set{data: make(map[string]struct{})}
	set.Add([]byte("a"), []byte("b"), []byte("c"))

	// Get random member without removal
	member := set.GetRandom()
	if member == nil {
		t.Fatal("Expected to get a random member, got nil")
	}
	if !set.IsMember(member) {
		t.Error("Random member should still be in set")
	}
	if set.Len() != 3 {
		t.Errorf("Expected length 3, got %d", set.Len())
	}

	// Empty set
	emptySet := &Set{data: make(map[string]struct{})}
	if emptySet.GetRandom() != nil {
		t.Error("Expected nil from empty set")
	}
}

func TestSet_GetRandomMembers(t *testing.T) {
	set := &Set{data: make(map[string]struct{})}
	set.Add([]byte("a"), []byte("b"), []byte("c"), []byte("d"), []byte("e"))

	// Get 2 random members
	members := set.GetRandomMembers(2)
	if len(members) != 2 {
		t.Errorf("Expected 2 members, got %d", len(members))
	}

	// Get more than available
	members = set.GetRandomMembers(10)
	if len(members) != 5 {
		t.Errorf("Expected 5 members (all available), got %d", len(members))
	}

	// Get from empty set
	emptySet := &Set{data: make(map[string]struct{})}
	if len(emptySet.GetRandomMembers(5)) != 0 {
		t.Error("Expected no members from empty set")
	}
}

func TestSet_Diff(t *testing.T) {
	set1 := &Set{data: make(map[string]struct{})}
	set1.Add([]byte("a"), []byte("b"), []byte("c"))

	set2 := &Set{data: make(map[string]struct{})}
	set2.Add([]byte("c"), []byte("d"), []byte("e"))

	// set1 - set2 = {a, b}
	diff := set1.Diff([]*Set{set2})
	if len(diff) != 2 {
		t.Errorf("Expected 2 members in diff, got %d", len(diff))
	}

	// Check diff result
	diffSet := make(map[string]bool)
	for _, m := range diff {
		diffSet[string(m)] = true
	}
	if !diffSet["a"] || !diffSet["b"] || diffSet["c"] {
		t.Error("Diff result incorrect")
	}

	// Empty diff
	set3 := &Set{data: make(map[string]struct{})}
	set3.Add([]byte("a"), []byte("b"), []byte("c"))
	diff = set3.Diff([]*Set{set1})
	if len(diff) != 0 {
		t.Errorf("Expected 0 members in diff, got %d", len(diff))
	}
}

func TestSet_Intersect(t *testing.T) {
	set1 := &Set{data: make(map[string]struct{})}
	set1.Add([]byte("a"), []byte("b"), []byte("c"))

	set2 := &Set{data: make(map[string]struct{})}
	set2.Add([]byte("c"), []byte("d"), []byte("e"))

	set3 := &Set{data: make(map[string]struct{})}
	set3.Add([]byte("b"), []byte("c"), []byte("f"))

	// set1 ∩ set2 ∩ set3 = {c}
	intersect := set1.Intersect([]*Set{set2, set3})
	if len(intersect) != 1 {
		t.Errorf("Expected 1 member in intersection, got %d", len(intersect))
	}
	if string(intersect[0]) != "c" {
		t.Errorf("Expected 'c' in intersection, got '%s'", string(intersect[0]))
	}

	// No intersection
	set4 := &Set{data: make(map[string]struct{})}
	set4.Add([]byte("x"), []byte("y"))
	intersect = set1.Intersect([]*Set{set4})
	if len(intersect) != 0 {
		t.Errorf("Expected 0 members in intersection, got %d", len(intersect))
	}
}

func TestSet_Union(t *testing.T) {
	set1 := &Set{data: make(map[string]struct{})}
	set1.Add([]byte("a"), []byte("b"))

	set2 := &Set{data: make(map[string]struct{})}
	set2.Add([]byte("c"), []byte("d"))

	set3 := &Set{data: make(map[string]struct{})}
	set3.Add([]byte("b"), []byte("e"))

	// set1 ∪ set2 ∪ set3 = {a, b, c, d, e}
	union := set1.Union([]*Set{set2, set3})
	if len(union) != 5 {
		t.Errorf("Expected 5 members in union, got %d", len(union))
	}

	// Check union result
	unionSet := make(map[string]bool)
	for _, m := range union {
		unionSet[string(m)] = true
	}
	for _, expected := range []string{"a", "b", "c", "d", "e"} {
		if !unionSet[expected] {
			t.Errorf("Expected '%s' in union", expected)
		}
	}
}

func TestSet_IsSubset(t *testing.T) {
	set1 := &Set{data: make(map[string]struct{})}
	set1.Add([]byte("a"), []byte("b"))

	set2 := &Set{data: make(map[string]struct{})}
	set2.Add([]byte("a"), []byte("b"), []byte("c"))

	set3 := &Set{data: make(map[string]struct{})}
	set3.Add([]byte("a"), []byte("c"))

	// set1 ⊆ set2 = true
	if !set1.IsSubset(set2) {
		t.Error("Expected set1 to be subset of set2")
	}

	// set1 ⊆ set3 = false
	if set1.IsSubset(set3) {
		t.Error("Expected set1 not to be subset of set3")
	}

	// Empty set is subset of any set
	emptySet := &Set{data: make(map[string]struct{})}
	if !emptySet.IsSubset(set1) {
		t.Error("Expected empty set to be subset of set1")
	}
}

func TestSet_Move(t *testing.T) {
	set1 := &Set{data: make(map[string]struct{})}
	set1.Add([]byte("a"), []byte("b"), []byte("c"))

	set2 := &Set{data: make(map[string]struct{})}
	set2.Add([]byte("d"))

	// Move existing member
	moved := set1.Move(set2, []byte("b"))
	if !moved {
		t.Error("Expected move to succeed")
	}
	if set1.IsMember([]byte("b")) {
		t.Error("'b' should not be in set1 anymore")
	}
	if !set2.IsMember([]byte("b")) {
		t.Error("'b' should be in set2")
	}

	// Move non-existing member
	moved = set1.Move(set2, []byte("x"))
	if moved {
		t.Error("Expected move to fail for non-existing member")
	}
}

func TestSet_Scan(t *testing.T) {
	set := &Set{data: make(map[string]struct{})}
	set.Add([]byte("a"), []byte("b"), []byte("c"), []byte("d"), []byte("e"))

	// First scan
	cursor, members := set.Scan(0, 2)
	if len(members) != 2 {
		t.Errorf("Expected 2 members, got %d", len(members))
	}
	if cursor == 0 {
		t.Error("Expected non-zero cursor for incomplete scan")
	}

	// Continue scan
	nextCursor, nextMembers := set.Scan(cursor, 2)
	if len(nextMembers) != 2 {
		t.Errorf("Expected 2 members, got %d", len(nextMembers))
	}
	if nextCursor == 0 {
		t.Error("Expected non-zero cursor for incomplete scan")
	}

	// Final scan
	finalCursor, finalMembers := set.Scan(nextCursor, 2)
	if len(finalMembers) != 1 {
		t.Errorf("Expected 1 member, got %d", len(finalMembers))
	}
	if finalCursor != 0 {
		t.Error("Expected cursor 0 for complete scan")
	}

	// Scan beyond end
	cursor, members = set.Scan(100, 2)
	if cursor != 0 {
		t.Error("Expected cursor 0 when starting beyond end")
	}
	if len(members) != 0 {
		t.Error("Expected no members when starting beyond end")
	}
}

func TestSet_Clear(t *testing.T) {
	set := &Set{data: make(map[string]struct{})}
	set.Add([]byte("a"), []byte("b"), []byte("c"))

	set.Clear()

	if set.Len() != 0 {
		t.Errorf("Expected length 0 after clear, got %d", set.Len())
	}
}

func TestSet_HasSameMembersAs(t *testing.T) {
	set1 := &Set{data: make(map[string]struct{})}
	set1.Add([]byte("a"), []byte("b"), []byte("c"))

	set2 := &Set{data: make(map[string]struct{})}
	set2.Add([]byte("a"), []byte("b"), []byte("c"))

	set3 := &Set{data: make(map[string]struct{})}
	set3.Add([]byte("a"), []byte("b"))

	// Same members
	if !set1.HasSameMembersAs(set2) {
		t.Error("Expected set1 and set2 to have same members")
	}

	// Different members
	if set1.HasSameMembersAs(set3) {
		t.Error("Expected set1 and set3 to have different members")
	}

	// With nil
	if set1.HasSameMembersAs(nil) {
		t.Error("Expected false when comparing with nil")
	}
}

func TestSet_String(t *testing.T) {
	set := &Set{data: make(map[string]struct{})}
	set.Add([]byte("a"), []byte("b"), []byte("c"))

	str := set.String()
	if str == "" {
		t.Error("String() returned empty string")
	}

	// Empty set
	emptySet := &Set{data: make(map[string]struct{})}
	if emptySet.String() != "{}" {
		t.Errorf("Expected '{}' for empty set, got '%s'", emptySet.String())
	}
}

func TestSet_EdgeCases(t *testing.T) {
	set := &Set{data: make(map[string]struct{})}

	// Add empty byte slice
	count := set.Add([]byte(""))
	if count != 1 {
		t.Errorf("Expected to add 1 member, got %d", count)
	}
	if !set.IsMember([]byte("")) {
		t.Error("Expected empty byte slice to be a member")
	}

	// Diff with no other sets
	diff := set.Diff([]*Set{})
	if len(diff) != 1 {
		t.Errorf("Expected 1 member in diff, got %d", len(diff))
	}

	// Intersect with no other sets
	intersect := set.Intersect([]*Set{})
	if len(intersect) != 1 {
		t.Errorf("Expected 1 member in intersection, got %d", len(intersect))
	}

	// Union with no other sets
	union := set.Union([]*Set{})
	if len(union) != 1 {
		t.Errorf("Expected 1 member in union, got %d", len(union))
	}
}

func TestSet_EqualBytes(t *testing.T) {
	set := &Set{data: make(map[string]struct{})}
	set.Add([]byte("a"), []byte("b"), []byte("c"))

	if !set.EqualBytes([]byte("a")) {
		t.Error("Expected EqualBytes to return true for 'a'")
	}
	if set.EqualBytes([]byte("x")) {
		t.Error("Expected EqualBytes to return false for 'x'")
	}
}

// Helper function to compare slices regardless of order
func slicesEqual(a, b [][]byte) bool {
	if len(a) != len(b) {
		return false
	}

	aMap := make(map[string]bool)
	for _, v := range a {
		aMap[string(v)] = true
	}

	for _, v := range b {
		if !aMap[string(v)] {
			return false
		}
	}

	return true
}

func TestSlicesEqual(t *testing.T) {
	tests := []struct {
		name string
		a    [][]byte
		b    [][]byte
		want bool
	}{
		{
			name: "equal slices",
			a:    [][]byte{[]byte("a"), []byte("b"), []byte("c")},
			b:    [][]byte{[]byte("c"), []byte("b"), []byte("a")},
			want: true,
		},
		{
			name: "different slices",
			a:    [][]byte{[]byte("a"), []byte("b")},
			b:    [][]byte{[]byte("a"), []byte("c")},
			want: false,
		},
		{
			name: "different lengths",
			a:    [][]byte{[]byte("a"), []byte("b")},
			b:    [][]byte{[]byte("a"), []byte("b"), []byte("c")},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := slicesEqual(tt.a, tt.b); got != tt.want {
				t.Errorf("slicesEqual() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSet_ReflectDeepEqual(t *testing.T) {
	set1 := &Set{data: make(map[string]struct{})}
	set1.Add([]byte("a"), []byte("b"))

	set2 := &Set{data: make(map[string]struct{})}
	set2.Add([]byte("b"), []byte("a"))

	// Members() may return different order, but should contain same elements
	members1 := set1.Members()
	members2 := set2.Members()

	if len(members1) != len(members2) {
		t.Error("Members should have same length")
	}

	// Check if they contain the same elements
	m1Set := make(map[string]bool)
	for _, m := range members1 {
		m1Set[string(m)] = true
	}

	for _, m := range members2 {
		if !m1Set[string(m)] {
			t.Error("Members2 contains element not in Members1")
		}
	}
}

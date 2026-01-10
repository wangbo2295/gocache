package datastruct

import (
	"math"
	"testing"
)

func TestMakeSortedSet(t *testing.T) {
	entity := MakeSortedSet()

	if entity == nil {
		t.Fatal("MakeSortedSet returned nil")
	}

	zset, ok := entity.Data.(*SortedSet)
	if !ok {
		t.Fatal("MakeSortedSet did not return a SortedSet")
	}

	if zset == nil {
		t.Fatal("SortedSet is nil")
	}

	if zset.Len() != 0 {
		t.Errorf("Expected empty sorted set, got length %d", zset.Len())
	}
}

func TestSortedSet_Add(t *testing.T) {
	zset := &SortedSet{
		members:  make(map[string]*sortedSetMember),
		elements: make([]*sortedSetMember, 0),
	}

	// Add new member
	count := zset.Add(1.0, []byte("a"))
	if count != 1 {
		t.Errorf("Expected to add 1 member, got %d", count)
	}
	if zset.Len() != 1 {
		t.Errorf("Expected length 1, got %d", zset.Len())
	}

	// Add another member
	count = zset.Add(2.0, []byte("b"))
	if count != 1 {
		t.Errorf("Expected to add 1 member, got %d", count)
	}
	if zset.Len() != 2 {
		t.Errorf("Expected length 2, got %d", zset.Len())
	}

	// Update existing member
	count = zset.Add(3.0, []byte("a"))
	if count != 0 {
		t.Errorf("Expected to add 0 members (update), got %d", count)
	}
	if zset.Len() != 2 {
		t.Errorf("Expected length 2, got %d", zset.Len())
	}

	// Check order
	if zset.elements[0].score != 2.0 {
		t.Errorf("Expected first score to be 2.0, got %f", zset.elements[0].score)
	}
	if zset.elements[1].score != 3.0 {
		t.Errorf("Expected second score to be 3.0, got %f", zset.elements[1].score)
	}
}

func TestSortedSet_Remove(t *testing.T) {
	zset := &SortedSet{
		members:  make(map[string]*sortedSetMember),
		elements: make([]*sortedSetMember, 0),
	}

	zset.Add(1.0, []byte("a"))
	zset.Add(2.0, []byte("b"))
	zset.Add(3.0, []byte("c"))

	// Remove existing member
	count := zset.Remove([]byte("b"))
	if count != 1 {
		t.Errorf("Expected to remove 1 member, got %d", count)
	}
	if zset.Len() != 2 {
		t.Errorf("Expected length 2, got %d", zset.Len())
	}

	// Remove non-existing member
	count = zset.Remove([]byte("x"))
	if count != 0 {
		t.Errorf("Expected to remove 0 members, got %d", count)
	}

	// Remove multiple members
	count = zset.Remove([]byte("a"), []byte("c"))
	if count != 2 {
		t.Errorf("Expected to remove 2 members, got %d", count)
	}
	if zset.Len() != 0 {
		t.Errorf("Expected length 0, got %d", zset.Len())
	}
}

func TestSortedSet_Score(t *testing.T) {
	zset := &SortedSet{
		members:  make(map[string]*sortedSetMember),
		elements: make([]*sortedSetMember, 0),
	}

	zset.Add(1.5, []byte("a"))
	zset.Add(2.5, []byte("b"))

	// Get existing member score
	score := zset.Score([]byte("a"))
	if score != 1.5 {
		t.Errorf("Expected score 1.5, got %f", score)
	}

	// Get non-existing member score
	score = zset.Score([]byte("x"))
	if !math.IsNaN(score) {
		t.Errorf("Expected NaN for non-existing member, got %f", score)
	}
}

func TestSortedSet_Rank(t *testing.T) {
	zset := &SortedSet{
		members:  make(map[string]*sortedSetMember),
		elements: make([]*sortedSetMember, 0),
	}

	zset.Add(1.0, []byte("a"))
	zset.Add(2.0, []byte("b"))
	zset.Add(3.0, []byte("c"))

	// Get existing member rank
	rank := zset.Rank([]byte("a"))
	if rank != 0 {
		t.Errorf("Expected rank 0, got %d", rank)
	}

	rank = zset.Rank([]byte("c"))
	if rank != 2 {
		t.Errorf("Expected rank 2, got %d", rank)
	}

	// Get non-existing member rank
	rank = zset.Rank([]byte("x"))
	if rank != -1 {
		t.Errorf("Expected rank -1 for non-existing member, got %d", rank)
	}
}

func TestSortedSet_RevRank(t *testing.T) {
	zset := &SortedSet{
		members:  make(map[string]*sortedSetMember),
		elements: make([]*sortedSetMember, 0),
	}

	zset.Add(1.0, []byte("a"))
	zset.Add(2.0, []byte("b"))
	zset.Add(3.0, []byte("c"))

	// Get existing member reverse rank
	rank := zset.RevRank([]byte("c"))
	if rank != 0 {
		t.Errorf("Expected reverse rank 0, got %d", rank)
	}

	rank = zset.RevRank([]byte("a"))
	if rank != 2 {
		t.Errorf("Expected reverse rank 2, got %d", rank)
	}

	// Get non-existing member reverse rank
	rank = zset.RevRank([]byte("x"))
	if rank != -1 {
		t.Errorf("Expected reverse rank -1 for non-existing member, got %d", rank)
	}
}

func TestSortedSet_Count(t *testing.T) {
	zset := &SortedSet{
		members:  make(map[string]*sortedSetMember),
		elements: make([]*sortedSetMember, 0),
	}

	zset.Add(1.0, []byte("a"))
	zset.Add(2.0, []byte("b"))
	zset.Add(3.0, []byte("c"))
	zset.Add(4.0, []byte("d"))
	zset.Add(5.0, []byte("e"))

	// Count in range
	count := zset.Count(2.0, 4.0)
	if count != 3 {
		t.Errorf("Expected count 3, got %d", count)
	}

	// Count with no matches
	count = zset.Count(10.0, 20.0)
	if count != 0 {
		t.Errorf("Expected count 0, got %d", count)
	}

	// Count all
	count = zset.Count(0.0, 100.0)
	if count != 5 {
		t.Errorf("Expected count 5, got %d", count)
	}
}

func TestSortedSet_Range(t *testing.T) {
	zset := &SortedSet{
		members:  make(map[string]*sortedSetMember),
		elements: make([]*sortedSetMember, 0),
	}

	zset.Add(1.0, []byte("a"))
	zset.Add(2.0, []byte("b"))
	zset.Add(3.0, []byte("c"))
	zset.Add(4.0, []byte("d"))
	zset.Add(5.0, []byte("e"))

	// Range without scores
	result := zset.Range(0, 2, false)
	if len(result) != 3 {
		t.Errorf("Expected 3 members, got %d", len(result))
	}
	if string(result[0]) != "a" || string(result[1]) != "b" || string(result[2]) != "c" {
		t.Error("Range result incorrect")
	}

	// Range with scores
	result = zset.Range(1, 3, true)
	if len(result) != 6 { // 3 members * 2 (member + score)
		t.Errorf("Expected 6 elements, got %d", len(result))
	}

	// Negative indices
	result = zset.Range(-2, -1, false)
	if len(result) != 2 {
		t.Errorf("Expected 2 members, got %d", len(result))
	}
	if string(result[0]) != "d" || string(result[1]) != "e" {
		t.Error("Negative range result incorrect")
	}

	// Out of range
	result = zset.Range(10, 20, false)
	if len(result) != 0 {
		t.Errorf("Expected 0 members, got %d", len(result))
	}
}

func TestSortedSet_RevRange(t *testing.T) {
	zset := &SortedSet{
		members:  make(map[string]*sortedSetMember),
		elements: make([]*sortedSetMember, 0),
	}

	zset.Add(1.0, []byte("a"))
	zset.Add(2.0, []byte("b"))
	zset.Add(3.0, []byte("c"))
	zset.Add(4.0, []byte("d"))
	zset.Add(5.0, []byte("e"))

	// Reverse range without scores
	result := zset.RevRange(0, 2, false)
	if len(result) != 3 {
		t.Errorf("Expected 3 members, got %d", len(result))
	}
	if string(result[0]) != "e" || string(result[1]) != "d" || string(result[2]) != "c" {
		t.Error("Reverse range result incorrect")
	}

	// Reverse range with scores
	result = zset.RevRange(0, 1, true)
	if len(result) != 4 { // 2 members * 2
		t.Errorf("Expected 4 elements, got %d", len(result))
	}
}

func TestSortedSet_RangeByScore(t *testing.T) {
	zset := &SortedSet{
		members:  make(map[string]*sortedSetMember),
		elements: make([]*sortedSetMember, 0),
	}

	zset.Add(1.0, []byte("a"))
	zset.Add(2.0, []byte("b"))
	zset.Add(3.0, []byte("c"))
	zset.Add(4.0, []byte("d"))
	zset.Add(5.0, []byte("e"))

	// Range by score without scores
	result := zset.RangeByScore(2.0, 4.0, false)
	if len(result) != 3 {
		t.Errorf("Expected 3 members, got %d", len(result))
	}

	// Range by score with scores
	result = zset.RangeByScore(2.0, 3.0, true)
	if len(result) != 4 { // 2 members * 2
		t.Errorf("Expected 4 elements, got %d", len(result))
	}

	// No matches
	result = zset.RangeByScore(10.0, 20.0, false)
	if len(result) != 0 {
		t.Errorf("Expected 0 members, got %d", len(result))
	}
}

func TestSortedSet_Len(t *testing.T) {
	zset := &SortedSet{
		members:  make(map[string]*sortedSetMember),
		elements: make([]*sortedSetMember, 0),
	}

	if zset.Len() != 0 {
		t.Errorf("Expected length 0, got %d", zset.Len())
	}

	zset.Add(1.0, []byte("a"))
	if zset.Len() != 1 {
		t.Errorf("Expected length 1, got %d", zset.Len())
	}

	zset.Add(2.0, []byte("b"))
	if zset.Len() != 2 {
		t.Errorf("Expected length 2, got %d", zset.Len())
	}
}

func TestSortedSet_IncrBy(t *testing.T) {
	zset := &SortedSet{
		members:  make(map[string]*sortedSetMember),
		elements: make([]*sortedSetMember, 0),
	}

	// Increment existing member
	zset.Add(1.0, []byte("a"))
	newScore := zset.IncrBy(2.5, []byte("a"))
	if newScore != 3.5 {
		t.Errorf("Expected new score 3.5, got %f", newScore)
	}
	if zset.Score([]byte("a")) != 3.5 {
		t.Errorf("Expected score 3.5, got %f", zset.Score([]byte("a")))
	}

	// Increment non-existing member (should add with increment as score)
	newScore = zset.IncrBy(5.0, []byte("b"))
	if newScore != 5.0 {
		t.Errorf("Expected new score 5.0, got %f", newScore)
	}
	if !zset.IsMember([]byte("b")) {
		t.Error("Expected member 'b' to exist")
	}
}

func TestSortedSet_Clear(t *testing.T) {
	zset := &SortedSet{
		members:  make(map[string]*sortedSetMember),
		elements: make([]*sortedSetMember, 0),
	}

	zset.Add(1.0, []byte("a"))
	zset.Add(2.0, []byte("b"))
	zset.Add(3.0, []byte("c"))

	zset.Clear()

	if zset.Len() != 0 {
		t.Errorf("Expected length 0 after clear, got %d", zset.Len())
	}
}

func TestSortedSet_Members(t *testing.T) {
	zset := &SortedSet{
		members:  make(map[string]*sortedSetMember),
		elements: make([]*sortedSetMember, 0),
	}

	zset.Add(1.0, []byte("a"))
	zset.Add(2.0, []byte("b"))
	zset.Add(3.0, []byte("c"))

	members := zset.Members()
	if len(members) != 3 {
		t.Errorf("Expected 3 members, got %d", len(members))
	}
	if string(members[0]) != "a" || string(members[1]) != "b" || string(members[2]) != "c" {
		t.Error("Members not in correct order")
	}

	// Empty sorted set
	emptyZset := &SortedSet{
		members:  make(map[string]*sortedSetMember),
		elements: make([]*sortedSetMember, 0),
	}
	if len(emptyZset.Members()) != 0 {
		t.Error("Expected no members from empty sorted set")
	}
}

func TestSortedSet_String(t *testing.T) {
	zset := &SortedSet{
		members:  make(map[string]*sortedSetMember),
		elements: make([]*sortedSetMember, 0),
	}

	zset.Add(1.0, []byte("a"))
	zset.Add(2.0, []byte("b"))

	str := zset.String()
	if str == "" {
		t.Error("String() returned empty string")
	}

	// Empty sorted set
	emptyZset := &SortedSet{
		members:  make(map[string]*sortedSetMember),
		elements: make([]*sortedSetMember, 0),
	}
	if emptyZset.String() != "[]" {
		t.Errorf("Expected '[]' for empty sorted set, got '%s'", emptyZset.String())
	}
}

func TestSortedSet_GetMemberByRank(t *testing.T) {
	zset := &SortedSet{
		members:  make(map[string]*sortedSetMember),
		elements: make([]*sortedSetMember, 0),
	}

	zset.Add(1.0, []byte("a"))
	zset.Add(2.0, []byte("b"))
	zset.Add(3.0, []byte("c"))

	// Get member by rank
	member := zset.GetMemberByRank(0)
	if string(member) != "a" {
		t.Errorf("Expected 'a', got '%s'", string(member))
	}

	member = zset.GetMemberByRank(2)
	if string(member) != "c" {
		t.Errorf("Expected 'c', got '%s'", string(member))
	}

	// Out of range
	member = zset.GetMemberByRank(10)
	if member != nil {
		t.Errorf("Expected nil for out of range rank, got '%s'", string(member))
	}
}

func TestSortedSet_GetMemberByRevRank(t *testing.T) {
	zset := &SortedSet{
		members:  make(map[string]*sortedSetMember),
		elements: make([]*sortedSetMember, 0),
	}

	zset.Add(1.0, []byte("a"))
	zset.Add(2.0, []byte("b"))
	zset.Add(3.0, []byte("c"))

	// Get member by reverse rank
	member := zset.GetMemberByRevRank(0)
	if string(member) != "c" {
		t.Errorf("Expected 'c', got '%s'", string(member))
	}

	member = zset.GetMemberByRevRank(2)
	if string(member) != "a" {
		t.Errorf("Expected 'a', got '%s'", string(member))
	}

	// Out of range
	member = zset.GetMemberByRevRank(10)
	if member != nil {
		t.Errorf("Expected nil for out of range rank, got '%s'", string(member))
	}
}

func TestSortedSet_GetScoreByRank(t *testing.T) {
	zset := &SortedSet{
		members:  make(map[string]*sortedSetMember),
		elements: make([]*sortedSetMember, 0),
	}

	zset.Add(1.5, []byte("a"))
	zset.Add(2.5, []byte("b"))
	zset.Add(3.5, []byte("c"))

	// Get score by rank
	score := zset.GetScoreByRank(1)
	if score != 2.5 {
		t.Errorf("Expected 2.5, got %f", score)
	}

	// Out of range
	score = zset.GetScoreByRank(10)
	if !math.IsNaN(score) {
		t.Errorf("Expected NaN for out of range rank, got %f", score)
	}
}

func TestSortedSet_EdgeCases(t *testing.T) {
	zset := &SortedSet{
		members:  make(map[string]*sortedSetMember),
		elements: make([]*sortedSetMember, 0),
	}

	// Add members with same score (should maintain insertion order after sorting)
	zset.Add(1.0, []byte("a"))
	zset.Add(1.0, []byte("b"))
	zset.Add(1.0, []byte("c"))

	if zset.Len() != 3 {
		t.Errorf("Expected length 3, got %d", zset.Len())
	}

	// Range with same scores
	result := zset.Range(0, -1, false)
	if len(result) != 3 {
		t.Errorf("Expected 3 members, got %d", len(result))
	}
}

// Helper method to check if member exists
func (z *SortedSet) IsMember(member []byte) bool {
	_, exists := z.members[string(member)]
	return exists
}

func TestSortedSet_DuplicateScores(t *testing.T) {
	zset := &SortedSet{
		members:  make(map[string]*sortedSetMember),
		elements: make([]*sortedSetMember, 0),
	}

	// Add members with duplicate scores
	zset.Add(1.0, []byte("first"))
	zset.Add(1.0, []byte("second"))
	zset.Add(1.0, []byte("third"))

	// All should be stored
	if zset.Len() != 3 {
		t.Errorf("Expected 3 members, got %d", zset.Len())
	}

	// Count should return all
	count := zset.Count(1.0, 1.0)
	if count != 3 {
		t.Errorf("Expected count 3, got %d", count)
	}

	// Range by score should return all
	result := zset.RangeByScore(1.0, 1.0, false)
	if len(result) != 3 {
		t.Errorf("Expected 3 members, got %d", len(result))
	}
}

func TestSortedSet_FractionalScores(t *testing.T) {
	zset := &SortedSet{
		members:  make(map[string]*sortedSetMember),
		elements: make([]*sortedSetMember, 0),
	}

	// Add members with fractional scores
	zset.Add(1.5, []byte("a"))
	zset.Add(2.7, []byte("b"))
	zset.Add(3.14, []byte("c"))
	zset.Add(0.1, []byte("d"))

	if zset.Len() != 4 {
		t.Errorf("Expected 4 members, got %d", zset.Len())
	}

	// Check order
	if zset.GetScoreByRank(0) != 0.1 {
		t.Errorf("Expected first score to be 0.1, got %f", zset.GetScoreByRank(0))
	}
	if zset.GetScoreByRank(3) != 3.14 {
		t.Errorf("Expected last score to be 3.14, got %f", zset.GetScoreByRank(3))
	}
}

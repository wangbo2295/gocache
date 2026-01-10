package datastruct

import (
	"bytes"
	"math"
	"strconv"
)

// SortedSet represents a Redis sorted set data structure
// Uses a simple slice-based implementation for simplicity
type SortedSet struct {
	// map maintains O(1) lookups by member
	members map[string]*sortedSetMember
	// slice maintains sorted order by score
	elements []*sortedSetMember
}

// sortedSetMember represents a member in the sorted set
type sortedSetMember struct {
	member []byte
	score  float64
}

// MakeSortedSet creates a new SortedSet wrapped in DataEntity
func MakeSortedSet() *DataEntity {
	return &DataEntity{Data: &SortedSet{
		members:  make(map[string]*sortedSetMember),
		elements: make([]*sortedSetMember, 0),
	}}
}

// Add adds or updates a member with a score
// Returns the number of new members added (0 if member already existed)
func (z *SortedSet) Add(score float64, member []byte) int {
	key := string(member)

	// Check if member already exists
	if existing, ok := z.members[key]; ok {
		// Update score if changed
		if existing.score != score {
			existing.score = score
			// Re-sort the elements
			z.resort()
		}
		return 0
	}

	// Add new member
	newMember := &sortedSetMember{
		member: member,
		score:  score,
	}
	z.members[key] = newMember
	z.elements = append(z.elements, newMember)

	// Sort elements by score
	z.resort()

	return 1
}

// Remove removes one or more members from the sorted set
// Returns the number of members removed
func (z *SortedSet) Remove(members ...[]byte) int {
	count := 0
	for _, member := range members {
		key := string(member)
		if _, exists := z.members[key]; exists {
			delete(z.members, key)
			count++
		}
	}

	// Rebuild elements slice
	if count > 0 {
		z.rebuildElements()
	}

	return count
}

// Score returns the score of a member
// Returns NaN if member doesn't exist
func (z *SortedSet) Score(member []byte) float64 {
	key := string(member)
	if m, ok := z.members[key]; ok {
		return m.score
	}
	return math.NaN()
}

// Rank returns the rank of a member (0-based, ordered by score ascending)
// Returns -1 if member doesn't exist
func (z *SortedSet) Rank(member []byte) int {
	key := string(member)
	if _, exists := z.members[key]; !exists {
		return -1
	}

	for i, elem := range z.elements {
		if bytes.Equal(elem.member, member) {
			return i
		}
	}

	return -1
}

// RevRank returns the rank of a member (0-based, ordered by score descending)
// Returns -1 if member doesn't exist
func (z *SortedSet) RevRank(member []byte) int {
	rank := z.Rank(member)
	if rank == -1 {
		return -1
	}
	return len(z.elements) - 1 - rank
}

// Count returns the number of members with scores between min and max (inclusive)
func (z *SortedSet) Count(min, max float64) int {
	count := 0
	for _, elem := range z.elements {
		if elem.score >= min && elem.score <= max {
			count++
		}
	}
	return count
}

// RangeByScore returns members with scores between min and max (inclusive)
// With scores determines if scores are included in the result
func (z *SortedSet) RangeByScore(min, max float64, withScores bool) [][]byte {
	result := make([][]byte, 0)
	for _, elem := range z.elements {
		if elem.score >= min && elem.score <= max {
			result = append(result, elem.member)
			if withScores {
				result = append(result, []byte(strconv.FormatFloat(elem.score, 'f', -1, 64)))
			}
		}
	}
	return result
}

// Range returns members in the given range [start, stop] by rank (ascending)
// With scores determines if scores are included in the result
func (z *SortedSet) Range(start, stop int, withScores bool) [][]byte {
	return z.rangeByIndex(start, stop, withScores, false)
}

// RevRange returns members in the given range [start, stop] by rank (descending)
// With scores determines if scores are included in the result
func (z *SortedSet) RevRange(start, stop int, withScores bool) [][]byte {
	return z.rangeByIndex(start, stop, withScores, true)
}

// rangeByIndex is the internal implementation for Range and RevRange
func (z *SortedSet) rangeByIndex(start, stop int, withScores, reverse bool) [][]byte {
	length := len(z.elements)

	// Normalize negative indices
	if start < 0 {
		start = length + start
		if start < 0 {
			start = 0
		}
	}
	if stop < 0 {
		stop = length + stop
		if stop < 0 {
			return [][]byte{}
		}
	}

	// Boundary checks
	if start >= length || stop < 0 || start > stop {
		return [][]byte{}
	}

	if stop >= length {
		stop = length - 1
	}

	result := make([][]byte, 0, (stop-start+1)*(1+boolToInt(withScores)))

	if !reverse {
		// Ascending order
		for i := start; i <= stop; i++ {
			result = append(result, z.elements[i].member)
			if withScores {
				result = append(result, []byte(strconv.FormatFloat(z.elements[i].score, 'f', -1, 64)))
			}
		}
	} else {
		// Descending order
		for i := length - 1 - start; i >= length - 1 - stop; i-- {
			result = append(result, z.elements[i].member)
			if withScores {
				result = append(result, []byte(strconv.FormatFloat(z.elements[i].score, 'f', -1, 64)))
			}
		}
	}

	return result
}

// RangeByScoreWithLimit returns members with scores between min and max
// Limited to offset and count, ordered by score ascending
func (z *SortedSet) RangeByScoreWithLimit(min, max float64, offset, count int, withScores bool, reverse bool) [][]byte {
	result := make([][]byte, 0)
	skipped := 0

	var elements []*sortedSetMember
	if !reverse {
		elements = z.elements
	} else {
		// Reverse the elements for descending order
		elements = make([]*sortedSetMember, len(z.elements))
		for i := 0; i < len(z.elements); i++ {
			elements[len(z.elements)-1-i] = z.elements[i]
		}
	}

	for _, elem := range elements {
		if elem.score < min {
			continue
		}
		if elem.score > max {
			break
		}

		if skipped < offset {
			skipped++
			continue
		}

		if count > 0 && len(result)/ (1 + boolToInt(withScores)) >= count {
			break
		}

		result = append(result, elem.member)
		if withScores {
			result = append(result, []byte(strconv.FormatFloat(elem.score, 'f', -1, 64)))
		}
	}

	return result
}

// Len returns the number of members in the sorted set
func (z *SortedSet) Len() int {
	return len(z.elements)
}

// GetScore returns the score of a member (alias for Score)
func (z *SortedSet) GetScore(member []byte) float64 {
	return z.Score(member)
}

// IncrBy increments the score of a member by increment
// Returns the new score
func (z *SortedSet) IncrBy(increment float64, member []byte) float64 {
	key := string(member)
	if existing, ok := z.members[key]; ok {
		existing.score += increment
		z.resort()
		return existing.score
	}

	// Member doesn't exist, add it with the increment as initial score
	newMember := &sortedSetMember{
		member: member,
		score:  increment,
	}
	z.members[key] = newMember
	z.elements = append(z.elements, newMember)
	z.resort()

	return increment
}

// resort sorts the elements slice by score
func (z *SortedSet) resort() {
	// Simple insertion sort for simplicity (can be optimized with quicksort/mergesort)
	for i := 1; i < len(z.elements); i++ {
		key := z.elements[i]
		j := i - 1
		for j >= 0 && z.elements[j].score > key.score {
			z.elements[j+1] = z.elements[j]
			j--
		}
		z.elements[j+1] = key
	}
}

// rebuildElements rebuilds the elements slice from the members map
func (z *SortedSet) rebuildElements() {
	z.elements = make([]*sortedSetMember, 0, len(z.members))
	for _, member := range z.members {
		z.elements = append(z.elements, member)
	}
	z.resort()
}

// Clear removes all members from the sorted set
func (z *SortedSet) Clear() {
	z.members = make(map[string]*sortedSetMember)
	z.elements = make([]*sortedSetMember, 0)
}

// Members returns all members (ordered by score)
func (z *SortedSet) Members() [][]byte {
	result := make([][]byte, len(z.elements))
	for i, elem := range z.elements {
		result[i] = elem.member
	}
	return result
}

// String returns a string representation of the sorted set
func (z *SortedSet) String() string {
	if len(z.elements) == 0 {
		return "[]"
	}

	result := "["
	for i, elem := range z.elements {
		if i > 0 {
			result += ", "
		}
		result += strconv.Quote(string(elem.member))
		result += ":"
		result += strconv.FormatFloat(elem.score, 'f', -1, 64)
	}
	result += "]"
	return result
}

// boolToInt converts a boolean to int (0 or 1)
func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// GetMemberByRank returns the member at the given rank (0-based, ascending)
func (z *SortedSet) GetMemberByRank(rank int) []byte {
	if rank < 0 || rank >= len(z.elements) {
		return nil
	}
	return z.elements[rank].member
}

// GetMemberByRevRank returns the member at the given rank (0-based, descending)
func (z *SortedSet) GetMemberByRevRank(rank int) []byte {
	if rank < 0 || rank >= len(z.elements) {
		return nil
	}
	return z.elements[len(z.elements)-1-rank].member
}

// GetScoreByRank returns the score at the given rank (0-based, ascending)
func (z *SortedSet) GetScoreByRank(rank int) float64 {
	if rank < 0 || rank >= len(z.elements) {
		return math.NaN()
	}
	return z.elements[rank].score
}

// GetScoreByRevRank returns the score at the given rank (0-based, descending)
func (z *SortedSet) GetScoreByRevRank(rank int) float64 {
	if rank < 0 || rank >= len(z.elements) {
		return math.NaN()
	}
	return z.elements[len(z.elements)-1-rank].score
}

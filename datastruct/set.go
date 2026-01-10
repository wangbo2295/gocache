package datastruct

import (
	"bytes"
)

// Set represents a Redis set data structure (unordered collection of unique strings)
type Set struct {
	data map[string]struct{}
}

// MakeSet creates a new Set wrapped in DataEntity
func MakeSet() *DataEntity {
	return &DataEntity{Data: &Set{
		data: make(map[string]struct{}),
	}}
}

// Add adds one or more members to the set
// Returns the number of members that were added (excluding those already present)
func (s *Set) Add(members ...[]byte) int {
	count := 0
	for _, member := range members {
		key := string(member)
		if _, exists := s.data[key]; !exists {
			s.data[key] = struct{}{}
			count++
		}
	}
	return count
}

// Remove removes one or more members from the set
// Returns the number of members that were removed
func (s *Set) Remove(members ...[]byte) int {
	count := 0
	for _, member := range members {
		key := string(member)
		if _, exists := s.data[key]; exists {
			delete(s.data, key)
			count++
		}
	}
	return count
}

// IsMember checks if member is in the set
func (s *Set) IsMember(member []byte) bool {
	_, exists := s.data[string(member)]
	return exists
}

// Members returns all members of the set
func (s *Set) Members() [][]byte {
	result := make([][]byte, 0, len(s.data))
	for member := range s.data {
		result = append(result, []byte(member))
	}
	return result
}

// Len returns the number of members in the set
func (s *Set) Len() int {
	return len(s.data)
}

// Pop removes and returns a random member from the set
// Returns nil if set is empty
func (s *Set) Pop() []byte {
	for member := range s.data {
		delete(s.data, member)
		return []byte(member)
	}
	return nil
}

// GetRandom returns a random member from the set without removing it
// Returns nil if set is empty
func (s *Set) GetRandom() []byte {
	for member := range s.data {
		return []byte(member)
	}
	return nil
}

// GetRandomMembers returns n random members from the set without removing them
// Returns at most n members (fewer if set has less than n members)
func (s *Set) GetRandomMembers(n int) [][]byte {
	result := make([][]byte, 0, n)
	for member := range s.data {
		if len(result) >= n {
			break
		}
		result = append(result, []byte(member))
	}
	return result
}

// Diff returns the difference between this set and other sets (members in this set but not in others)
func (s *Set) Diff(others []*Set) [][]byte {
	if len(s.data) == 0 {
		return [][]byte{}
	}

	// Build a set of members to exclude
	exclude := make(map[string]struct{})
	for _, other := range others {
		if other == nil {
			continue
		}
		for member := range other.data {
			exclude[member] = struct{}{}
		}
	}

	// Collect members not in exclude
	result := make([][]byte, 0)
	for member := range s.data {
		if _, excluded := exclude[member]; !excluded {
			result = append(result, []byte(member))
		}
	}
	return result
}

// Intersect returns the intersection of this set with other sets
func (s *Set) Intersect(others []*Set) [][]byte {
	if len(s.data) == 0 {
		return [][]byte{}
	}

	// Find members that exist in all sets
	result := make([][]byte, 0)
	for member := range s.data {
		inAll := true
		for _, other := range others {
			if other == nil {
				inAll = false
				break
			}
			if _, exists := other.data[member]; !exists {
				inAll = false
				break
			}
		}
		if inAll {
			result = append(result, []byte(member))
		}
	}
	return result
}

// Union returns the union of this set with other sets
func (s *Set) Union(others []*Set) [][]byte {
	// Use a map to deduplicate
	seen := make(map[string]struct{})

	// Add members from this set
	for member := range s.data {
		seen[member] = struct{}{}
	}

	// Add members from other sets
	for _, other := range others {
		if other == nil {
			continue
		}
		for member := range other.data {
			seen[member] = struct{}{}
		}
	}

	// Convert to result
	result := make([][]byte, 0, len(seen))
	for member := range seen {
		result = append(result, []byte(member))
	}
	return result
}

// IsSubset checks if this set is a subset of another set
func (s *Set) IsSubset(other *Set) bool {
	if other == nil {
		return false
	}
	for member := range s.data {
		if _, exists := other.data[member]; !exists {
			return false
		}
	}
	return true
}

// Move moves a member from this set to another set
// Returns true if member was moved, false if member was not in this set
func (s *Set) Move(other *Set, member []byte) bool {
	key := string(member)
	if _, exists := s.data[key]; !exists {
		return false
	}

	delete(s.data, key)
	if other.data == nil {
		other.data = make(map[string]struct{})
	}
	other.data[key] = struct{}{}
	return true
}

// Scan iterates over members with a cursor
// Returns the next cursor and members in this batch
// Cursor 0 indicates start, cursor 0 indicates end
func (s *Set) Scan(cursor int64, count int64) (int64, [][]byte) {
	members := s.Members()
	total := int64(len(members))

	if cursor >= total {
		return 0, [][]byte{}
	}

	end := cursor + count
	if end > total {
		end = total
	}

	batch := members[cursor:end]

	if end >= total {
		return 0, batch
	}

	return end, batch
}

// Clear removes all members from the set
func (s *Set) Clear() {
	s.data = make(map[string]struct{})
}

// HasSameMembersAs checks if two sets have exactly the same members
func (s *Set) HasSameMembersAs(other *Set) bool {
	if other == nil {
		return false
	}

	if len(s.data) != len(other.data) {
		return false
	}

	for member := range s.data {
		if _, exists := other.data[member]; !exists {
			return false
		}
	}

	return true
}

// String returns a string representation of the set
func (s *Set) String() string {
	if len(s.data) == 0 {
		return "{}"
	}

	result := "{"
	members := s.Members()
	for i, member := range members {
		if i > 0 {
			result += ", "
		}
		result += string(member)
	}
	result += "}"
	return result
}

// EqualBytes compares if a byte slice equals a member in the set
func (s *Set) EqualBytes(member []byte) bool {
	for m := range s.data {
		if bytes.Equal([]byte(m), member) {
			return true
		}
	}
	return false
}

package datastruct

import (
	"strconv"
)

// List represents a Redis list data structure (implemented as doubly linked list)
type List struct {
	head *listNode
	tail *listNode
	size int
}

// listNode represents a node in the doubly linked list
type listNode struct {
	prev  *listNode
	next  *listNode
	value []byte
}

// MakeList creates a new List wrapped in DataEntity
func MakeList() *DataEntity {
	return &DataEntity{Data: &List{}}
}

// Len returns the number of elements in the list
func (l *List) Len() int {
	return l.size
}

// LPush inserts one or more values at the head of the list
// Returns the new length of the list
func (l *List) LPush(values ...[]byte) int {
	for _, value := range values {
		node := &listNode{
			value: value,
		}

		if l.head == nil {
			l.head = node
			l.tail = node
		} else {
			node.next = l.head
			l.head.prev = node
			l.head = node
		}
		l.size++
	}
	return l.size
}

// RPush inserts one or more values at the tail of the list
// Returns the new length of the list
func (l *List) RPush(values ...[]byte) int {
	for _, value := range values {
		node := &listNode{
			value: value,
		}

		if l.tail == nil {
			l.head = node
			l.tail = node
		} else {
			node.prev = l.tail
			l.tail.next = node
			l.tail = node
		}
		l.size++
	}
	return l.size
}

// LPop removes and returns the first element of the list
// Returns nil if list is empty
func (l *List) LPop() []byte {
	if l.head == nil {
		return nil
	}

	value := l.head.value
	l.head = l.head.next

	if l.head != nil {
		l.head.prev = nil
	} else {
		l.tail = nil
	}

	l.size--
	return value
}

// RPop removes and returns the last element of the list
// Returns nil if list is empty
func (l *List) RPop() []byte {
	if l.tail == nil {
		return nil
	}

	value := l.tail.value
	l.tail = l.tail.prev

	if l.tail != nil {
		l.tail.next = nil
	} else {
		l.head = nil
	}

	l.size--
	return value
}

// LIndex returns the element at index in the list
// Index 0 is the head, index -1 is the tail
// Returns nil if index is out of range
func (l *List) LIndex(index int) []byte {
	if index < 0 {
		index = l.size + index
	}

	if index < 0 || index >= l.size {
		return nil
	}

	node := l.head
	for i := 0; i < index; i++ {
		node = node.next
	}

	return node.value
}

// LSet sets the element at index to value
// Index 0 is the head, index -1 is the tail
// Returns error if index is out of range
func (l *List) LSet(index int, value []byte) error {
	if index < 0 {
		index = l.size + index
	}

	if index < 0 || index >= l.size {
		return ErrIndexOutOfRange
	}

	node := l.head
	for i := 0; i < index; i++ {
		node = node.next
	}

	node.value = value
	return nil
}

// LRange returns a slice of elements from start to stop (inclusive)
// Supports negative indices (index -1 is the tail)
// Returns empty slice if range is invalid
func (l *List) LRange(start, stop int) [][]byte {
	if l.size == 0 {
		return [][]byte{}
	}

	// Normalize negative indices
	if start < 0 {
		start = l.size + start
		if start < 0 {
			start = 0
		}
	}
	if stop < 0 {
		stop = l.size + stop
		if stop < 0 {
			return [][]byte{}
		}
	}

	// Clamp to valid range
	if start >= l.size {
		return [][]byte{}
	}
	if stop >= l.size {
		stop = l.size - 1
	}

	if start > stop {
		return [][]byte{}
	}

	// Collect elements
	result := make([][]byte, 0, stop-start+1)
	node := l.head
	for i := 0; i <= stop && node != nil; i++ {
		if i >= start {
			result = append(result, node.value)
		}
		node = node.next
	}

	return result
}

// LTrim trims the list to only contain elements from start to stop (inclusive)
// Supports negative indices
func (l *List) LTrim(start, stop int) {
	if l.size == 0 {
		return
	}

	// Normalize negative indices
	if start < 0 {
		start = l.size + start
		if start < 0 {
			start = 0
		}
	}
	if stop < 0 {
		stop = l.size + stop
		if stop < 0 {
			// Trim everything
			l.head = nil
			l.tail = nil
			l.size = 0
			return
		}
	}

	// Clamp to valid range
	if start >= l.size {
		// Trim everything
		l.head = nil
		l.tail = nil
		l.size = 0
		return
	}
	if stop >= l.size {
		stop = l.size - 1
	}

	if start > stop {
		// Trim everything
		l.head = nil
		l.tail = nil
		l.size = 0
		return
	}

	// Find new head and tail nodes
	newHead := l.head
	for i := 0; i < start; i++ {
		newHead = newHead.next
	}

	newTail := newHead
	for i := start; i < stop; i++ {
		newTail = newTail.next
	}

	// Update pointers and count
	newHead.prev = nil
	newTail.next = nil
	l.head = newHead
	l.tail = newTail
	l.size = stop - start + 1
}

// LRem removes the first count occurrences of elements equal to value
// If count > 0: remove first count elements
// If count < 0: remove last count elements
// If count == 0: remove all elements
// Returns the number of removed elements
func (l *List) LRem(count int, value []byte) int {
	if l.size == 0 {
		return 0
	}

	removed := 0

	if count >= 0 {
		// Remove from head
		node := l.head
		for node != nil && (count == 0 || removed < count) {
			if string(node.value) == string(value) {
				// Remove this node
				if node.prev != nil {
					node.prev.next = node.next
				} else {
					l.head = node.next
				}

				if node.next != nil {
					node.next.prev = node.prev
				} else {
					l.tail = node.prev
				}

				l.size--
				removed++
				node = node.next
			} else {
				node = node.next
			}
		}
	} else {
		// Remove from tail
		node := l.tail
		countToRemove := -count
		for node != nil && removed < countToRemove {
			if string(node.value) == string(value) {
				// Remove this node
				if node.prev != nil {
					node.prev.next = node.next
				} else {
					l.head = node.next
				}

				if node.next != nil {
					node.next.prev = node.prev
				} else {
					l.tail = node.prev
				}

				l.size--
				removed++
				node = node.prev
			} else {
				node = node.prev
			}
		}
	}

	return removed
}

// LInsert inserts value before or after pivot value
// Returns the new length of the list, or -1 if pivot not found, or 0 if error
func (l *List) LInsert(before bool, pivot, value []byte) int {
	// Find pivot node
	node := l.head
	for node != nil {
		if string(node.value) == string(pivot) {
			break
		}
		node = node.next
	}

	if node == nil {
		return -1 // pivot not found
	}

	newNode := &listNode{
		value: value,
	}

	if before {
		// Insert before pivot
		newNode.next = node
		newNode.prev = node.prev

		if node.prev != nil {
			node.prev.next = newNode
		} else {
			l.head = newNode
		}
		node.prev = newNode
	} else {
		// Insert after pivot
		newNode.prev = node
		newNode.next = node.next

		if node.next != nil {
			node.next.prev = newNode
		} else {
			l.tail = newNode
		}
		node.next = newNode
	}

	l.size++
	return l.size
}

// LLen returns the length of the list (alias for Len)
func (l *List) LLen() int {
	return l.Len()
}

// GetAll returns all elements in the list as a slice
func (l *List) GetAll() [][]byte {
	if l.size == 0 {
		return [][]byte{}
	}

	result := make([][]byte, l.size)
	node := l.head
	for i := 0; i < l.size; i++ {
		result[i] = node.value
		node = node.next
	}
	return result
}

// Clear removes all elements from the list
func (l *List) Clear() {
	l.head = nil
	l.tail = nil
	l.size = 0
}

// String returns a string representation of the list
func (l *List) String() string {
	if l.size == 0 {
		return "[]"
	}

	result := "["
	node := l.head
	first := true
	for node != nil {
		if !first {
			result += ", "
		}
		result += strconv.Quote(string(node.value))
		first = false
		node = node.next
	}
	result += "]"
	return result
}

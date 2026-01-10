package datastruct

import "unsafe"

// SizeEstimator provides size estimation for data structures
type SizeEstimator interface {
	// EstimateSize returns the estimated memory size in bytes
	EstimateSize() int64
}

// EstimateSize returns the estimated memory size of a DataEntity
func (e *DataEntity) EstimateSize() int64 {
	if estimator, ok := e.Data.(SizeEstimator); ok {
		return estimator.EstimateSize()
	}
	return estimateBasicSize(e.Data)
}

// estimateBasicSize provides a basic size estimation for any data structure
func estimateBasicSize(data interface{}) int64 {
	if data == nil {
		return 0
	}

	switch v := data.(type) {
	case *String:
		return int64(unsafe.Sizeof(String{})) + int64(len(v.Value))
	case *Hash:
		size := int64(unsafe.Sizeof(Hash{}))
		if v.data != nil {
			// Rough estimation: overhead + entries
			size += int64(v.data.Len()) * 100 // Approximate 100 bytes per entry
		}
		return size
	case *List:
		size := int64(unsafe.Sizeof(List{}))
		if v.Len() > 0 {
			// Rough estimation: overhead + elements
			size += int64(v.Len()) * 50 // Approximate 50 bytes per element
		}
		return size
	case *Set:
		size := int64(unsafe.Sizeof(Set{}))
		if v.data != nil {
			// Rough estimation: overhead + members
			size += int64(len(v.data)) * 80 // Approximate 80 bytes per member
		}
		return size
	case *SortedSet:
		size := int64(unsafe.Sizeof(SortedSet{}))
		if v.members != nil {
			// Rough estimation: overhead + members
			size += int64(len(v.members)) * 120 // Approximate 120 bytes per member (including score)
		}
		return size
	default:
		return 100 // Default minimal size
	}
}

// GetEstimatedSize returns the estimated size for a specific data type
func (s *String) GetEstimatedSize() int64 {
	return int64(unsafe.Sizeof(String{})) + int64(len(s.Value))
}

func (h *Hash) GetEstimatedSize() int64 {
	size := int64(unsafe.Sizeof(Hash{}))
	if h.data != nil {
		size += int64(h.data.Len()) * 100
	}
	return size
}

func (l *List) GetEstimatedSize() int64 {
	size := int64(unsafe.Sizeof(List{}))
	size += int64(l.Len()) * 50
	return size
}

func (s *Set) GetEstimatedSize() int64 {
	size := int64(unsafe.Sizeof(Set{}))
	if s.data != nil {
		size += int64(len(s.data)) * 80
	}
	return size
}

func (z *SortedSet) GetEstimatedSize() int64 {
	size := int64(unsafe.Sizeof(SortedSet{}))
	if z.members != nil {
		size += int64(len(z.members)) * 120
	}
	return size
}

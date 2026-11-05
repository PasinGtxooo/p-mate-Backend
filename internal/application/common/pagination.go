package common

const (
	DefaultLimit  = 10
	DefaultOffset = 0
)

// ValidatePagination returns the limit and offset values with default fallback
func ValidatePagination(limitPtr, offsetPtr *int) (int, int) {
	limit := DefaultLimit
	if limitPtr != nil {
		limit = *limitPtr
	}

	offset := DefaultOffset
	if offsetPtr != nil {
		offset = *offsetPtr
	}

	return limit, offset
}

// CalculateHasMore returns true if there are more items to fetch
func CalculateHasMore(offset, limit int, total int) bool {
	return offset+limit < total
}

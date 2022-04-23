package util

// CopyRef returns a pointer to a copy of value
func CopyRef[T any](value T) *T {
	return &value
}

// NilOr returns true if value is nil or expected
func NilOr[T comparable](value *T, expected T) bool {
	return value == nil || *value == expected
}

// NotNilAnd returns true if value is not nil and expected
func NotNilAnd[T comparable](value *T, expected T) bool {
	return value != nil && *value == expected
}

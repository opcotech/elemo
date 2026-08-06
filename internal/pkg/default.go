package pkg

// Default returns the value if it is not the zero value of the type, otherwise
// it returns the fallback.
func Default[T comparable](value, fallback T) T {
	var zero T
	if value == zero {
		return fallback
	}

	return value
}

// DefaultPtr returns the value if it is not nil, otherwise it returns the
// fallback.
func DefaultPtr[T any](value *T, fallback T) T {
	if value == nil {
		return fallback
	}

	return *value
}

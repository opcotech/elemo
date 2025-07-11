package pkg

// GetDefault returns the value if it is not the zero value of the type, otherwise
// it returns the fallback.
func RenderDefault[T comparable](value, fallback T) T {
	var zero T
	if value == zero {
		return fallback
	}

	return value
}

// GetDefaultPtr returns the value if it is not nil, otherwise it returns the
// fallback.
func RenderDefaultPtr[T any](value *T, fallback T) T {
	if value == nil {
		return fallback
	}

	return *value
}

package optional

import (
	"github.com/goccy/go-json"
)

// NullProtectedValueTag is the JSON key used when marshaling an explicit null
// for a pointer-typed optional value.
const NullProtectedValueTag = "null_protected_value"

type protected[T any] struct {
	Value *T `json:"null_protected_value"`
}

// Optional represents a value that may be undefined, set, or explicitly null.
//
// Defined == false           -> field was not provided (do not update)
// Defined == true, Value nil -> field was explicitly set to null/clear
// Defined == true, Value set -> field was set to *Value
type Optional[T any] struct {
	Defined bool `json:"defined"`
	Value   *T   `json:"value"`
}

// Some returns an Optional with a defined non-nil value.
func Some[T any](v T) Optional[T] {
	return Optional[T]{Defined: true, Value: &v}
}

// None returns an undefined Optional.
func None[T any]() Optional[T] {
	return Optional[T]{}
}

// Null returns an Optional that is explicitly set to nil.
func Null[T any]() Optional[T] {
	return Optional[T]{Defined: true, Value: nil}
}

// MarshalJSON is implemented by deferring to the wrapped type (T).
func (o Optional[T]) MarshalJSON() ([]byte, error) {
	if !o.Defined {
		return []byte("null"), nil
	}

	if o.Value == nil {
		return json.Marshal(protected[T]{
			Value: o.Value,
		})
	}

	return json.Marshal(o.Value)
}

// UnmarshalJSON is implemented by deferring to the wrapped type (T).
// It will be called only if the value is defined in the JSON payload.
func (o *Optional[T]) UnmarshalJSON(data []byte) error {
	o.Defined = true
	return json.Unmarshal(data, &o.Value)
}

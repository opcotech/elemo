package api

import (
	"reflect"

	"github.com/goccy/go-json"

	"github.com/opcotech/elemo/internal/pkg/convert"
)

const nullProtectedValueTag = "null_protected_value"

type protected[T any] struct {
	Value *T `json:"null_protected_value"`
}

// Optional is a wrapper type for optional fields in the JSON API.
type Optional[T any] struct {
	Defined bool `json:"defined"`
	Value   *T   `json:"value"`
}

// MarshalJSON is implemented by deferring to the wrapped type (T).
func (o Optional[T]) MarshalJSON() ([]byte, error) {
	if !o.Defined {
		return []byte("null"), nil
	}

	if reflect.ValueOf(o.Value).Kind() == reflect.Ptr && o.Value == nil {
		return json.Marshal(protected[T]{
			Value: o.Value,
		})
	}

	if o.Value == nil {
		var zero T
		o.Value = &zero
	}

	return json.Marshal(o.Value)
}

// UnmarshalJSON is implemented by deferring to the wrapped type (T).
// It will be called only if the value is defined in the JSON payload.
func (o *Optional[T]) UnmarshalJSON(data []byte) error {
	o.Defined = true
	return json.Unmarshal(data, &o.Value)
}

// ConvertRequestToMap converts a request to a map[string]any and removes all
// nil values. This is useful for the JSON API, where the client can send
// optional fields as null. Use this function to convert the request to a map
// only if the optional fields are Optional[T] types.
func ConvertRequestToMap(input any) (map[string]any, error) {
	res := make(map[string]any)
	if err := convert.AnyToAny(input, &res); err != nil {
		return nil, err
	}

	for k, v := range res {
		if v == nil {
			delete(res, k)
		}

		// Refill the value if it is a pointer to a zero value
		if reflect.ValueOf(v).Kind() == reflect.Map {
			if nullProtected, ok := v.(map[string]any)[nullProtectedValueTag]; ok {
				res[k] = nullProtected
			}
		}
	}

	return res, nil
}

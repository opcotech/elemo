package api

import (
	"github.com/goccy/go-json"

	"github.com/opcotech/elemo/internal/pkg/convert"
)

type Optional[T any] struct {
	Defined bool `json:"defined"`
	Value   *T   `json:"value,omitempty"`
}

// MarshalJSON is implemented by deferring to the wrapped type (T).
func (o Optional[T]) MarshalJSON() ([]byte, error) {
	if o.Defined && o.Value == nil {
		var zero T
		o.Value = &zero
	}
	if !o.Defined {
		return []byte("null"), nil
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
	}

	return res, nil
}

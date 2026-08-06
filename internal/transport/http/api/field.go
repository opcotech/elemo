package api

import (
	"reflect"

	"github.com/opcotech/elemo/internal/pkg/convert"
	"github.com/opcotech/elemo/internal/pkg/optional"
)

// Optional is a type alias for optional.Optional so OpenAPI-generated types
// continue to compile against this package.
type Optional[T any] = optional.Optional[T]

// ConvertRequestToMap converts a request to a map[string]any and removes all
// nil values. This is useful for the JSON API, where the client can send
// optional fields as null. Use this function to convert the request to a map
// only if the optional fields are Optional[T] types.
func ConvertRequestToMap(input any) (map[string]any, error) {
	res := make(map[string]any)
	if err := convert.AnyToAny(input, &res); err != nil {
		return nil, err
	}

	nullProtectedValueTag := optional.NullProtectedValueTag

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

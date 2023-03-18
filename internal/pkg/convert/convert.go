package convert

import (
	"encoding/json"
	"errors"
)

var (
	ErrMarshal   = errors.New("could not marshal value")   // value could not be marshaled
	ErrUnmarshal = errors.New("could not unmarshal value") // value could not be unmarshaled
)

// ToPointer converts any value to a pointer, but nil.
//
// If the value is nil, it panics.
func ToPointer[T any](src T) *T {
	return &src
}

// AnyToAny converts any value to any type that can be converted using
// json.Marshal and json.Unmarshal. If the output is set to nil, the
// function converts to nil.
func AnyToAny(input any, output any) error {
	if output == nil {
		return nil
	}

	var err error

	b, err := json.Marshal(&input)
	if err != nil {
		return errors.Join(ErrMarshal, err)
	}

	if err = json.Unmarshal(b, output); err != nil {
		return errors.Join(ErrUnmarshal, err)
	}

	return nil
}

// MustAnyToAny converts any value to any type that can be converted
// using json.Marshal and json.Unmarshal. If the output is set to nil, the
// function converts to nil. If the conversion fails, it panics.
func MustAnyToAny(input any, output any) {
	err := AnyToAny(input, output)
	if err != nil {
		panic(err)
	}
}

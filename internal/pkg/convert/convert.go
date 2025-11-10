package convert

import (
	"errors"
	"strings"
	"unicode"

	"github.com/goccy/go-json"
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

// EmailToNameParts extracts first and last name parts from an email address.
//
// The function extracts the local part (before the @ symbol) and attempts to
// split it into first and last name parts. It handles common separators like
// dots (.), hyphens (-), and underscores (_). If multiple separators are found,
// the first part is used as the first name, and the remaining parts are joined
// as the last name. The first letter of each name part is capitalized.
func EmailToNameParts(email string) (firstName, lastName string) {
	if email == "" {
		return "", ""
	}

	// Extract local part (before @)
	localPart, _, _ := strings.Cut(email, "@")
	localPart = strings.TrimSpace(localPart)
	if localPart == "" {
		return "", ""
	}

	// Split by first common separator found
	var parts []string
	for _, sep := range []string{".", "-", "_"} {
		if strings.Contains(localPart, sep) {
			parts = strings.Split(localPart, sep)
			break
		}
	}
	if len(parts) == 0 {
		parts = []string{localPart}
	}

	// Filter empty parts, trim whitespace, and capitalize in one pass
	filteredParts := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			filteredParts = append(filteredParts, capitalize(trimmed))
		}
	}

	if len(filteredParts) == 0 {
		return "", ""
	}

	firstName = filteredParts[0]
	if len(filteredParts) > 1 {
		lastName = strings.Join(filteredParts[1:], " ")
	}

	return firstName, lastName
}

// capitalize capitalizes the first letter of a string and lowercases the rest.
func capitalize(s string) string {
	if s == "" {
		return ""
	}

	runes := []rune(s)
	if len(runes) == 0 {
		return s
	}

	runes[0] = unicode.ToUpper(runes[0])
	for i := 1; i < len(runes); i++ {
		runes[i] = unicode.ToLower(runes[i])
	}

	return string(runes)
}

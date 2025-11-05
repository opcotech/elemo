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
// as the last name.
//
// Examples:
//   - "john.doe@example.com" -> ("John", "Doe")
//   - "john@example.com" -> ("John", "")
//   - "john.doe.smith@example.com" -> ("John", "Doe Smith")
//   - "john-doe@example.com" -> ("John", "Doe")
//   - "jane_doe@example.com" -> ("Jane", "Doe")
//
// The function capitalizes the first letter of each name part.
func EmailToNameParts(email string) (firstName, lastName string) {
	if email == "" {
		return "", ""
	}

	// Extract local part (before @)
	localPart := email
	if idx := strings.Index(email, "@"); idx != -1 {
		localPart = email[:idx]
	}

	if localPart == "" {
		return "", ""
	}

	// Split by common separators
	parts := []string{}
	for _, sep := range []string{".", "-", "_"} {
		if strings.Contains(localPart, sep) {
			parts = strings.Split(localPart, sep)
			break
		}
	}

	// If no separator found, use the whole local part as first name
	if len(parts) == 0 {
		parts = []string{localPart}
	}

	// Filter out empty parts
	filteredParts := []string{}
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			filteredParts = append(filteredParts, trimmed)
		}
	}

	if len(filteredParts) == 0 {
		return "", ""
	}

	// First part is first name
	firstName = capitalize(filteredParts[0])

	// Remaining parts joined as last name
	if len(filteredParts) > 1 {
		lastNameParts := []string{}
		for _, part := range filteredParts[1:] {
			lastNameParts = append(lastNameParts, capitalize(part))
		}
		lastName = strings.Join(lastNameParts, " ")
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

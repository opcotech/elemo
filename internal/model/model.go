package model

import (
	"database/sql/driver"
	"errors"
	"strings"

	"github.com/rs/xid"
)

// ID represents a unique identifier for a resource, combining a resource Type
// and a unique identifier.
type ID struct {
	Inner xid.ID
	Type  ResourceType
}

func (id ID) Validate() error {
	if !id.Type.IsAResourceType() {
		return ErrInvalidID
	}
	return nil
}

func (id ID) Value() (driver.Value, error) {
	return id.Composite(), nil
}

func (id *ID) Scan(value any) error {
	raw, err := scanString(value)
	if err != nil {
		return err
	}
	parsed, err := ParseCompositeID(raw)
	if err != nil {
		return err
	}
	*id = parsed
	return nil
}

func scanString(value any) (string, error) {
	switch v := value.(type) {
	case string:
		return v, nil
	case []byte:
		return string(v), nil
	default:
		return "", ErrInvalidID
	}
}

// String returns the string representation of the ID. The type is not part of
// the string representation. This is to allow for the ID to be used as a
// label or flag in a database or log aggregation system.
func (id ID) String() string {
	return id.Inner.String()
}

// Composite returns the typed identifier in ResourceType:xid form.
func (id ID) Composite() string {
	return id.Type.String() + ":" + id.Inner.String()
}

// SearchKey returns a Meilisearch-safe primary key (ResourceType_xid).
func (id ID) SearchKey() string {
	return id.Type.String() + "_" + id.Inner.String()
}

// ParseCompositeID parses a ResourceType:xid identifier.
func ParseCompositeID(value string) (ID, error) {
	typ, xid, ok := strings.Cut(value, ":")
	if !ok || typ == "" || xid == "" {
		return ID{}, ErrInvalidID
	}
	return NewIDFromString(xid, typ)
}

// ParseSearchKey parses a Meilisearch primary key in ResourceType_xid form.
func ParseSearchKey(value string) (ID, error) {
	typ, xid, ok := strings.Cut(value, "_")
	if !ok || typ == "" || xid == "" {
		return ID{}, ErrInvalidID
	}
	return NewIDFromString(xid, typ)
}

// Label returns the Type of the ID.
func (id ID) Label() string {
	return id.Type.String()
}

// IsNil returns true if the ID is nil.
func (id ID) IsNil() bool {
	return id.Inner == xid.NilID()
}

// NewID creates a new ID.
func NewID(typ ResourceType) (ID, error) {
	id := ID{Inner: xid.New(), Type: typ}

	if err := id.Validate(); err != nil {
		return ID{}, err
	}

	return id, nil
}

// MustNewID creates a new ID. It panics if the type is invalid.
func MustNewID(typ ResourceType) ID {
	id, err := NewID(typ)
	if err != nil {
		panic(err)
	}

	return id
}

// NewNilID creates a new ID with a nil xid.ID.
func NewNilID(typ ResourceType) (ID, error) {
	id := ID{Inner: xid.NilID(), Type: typ}

	if err := id.Validate(); err != nil {
		return ID{}, err
	}

	return id, nil
}

// MustNewNilID creates a new ID with a nil xid.ID. It panics if the type is
// invalid.
func MustNewNilID(typ ResourceType) ID {
	id, err := NewNilID(typ)
	if err != nil {
		panic(err)
	}

	return id
}

// NewIDFromString creates a new ID from a string. The string must be a valid
// xid string.
func NewIDFromString(id, typ string) (ID, error) {
	var rt ResourceType
	if err := rt.UnmarshalText([]byte(typ)); err != nil {
		return ID{}, errors.Join(ErrInvalidID, err)
	}

	newID, err := NewNilID(rt)
	if err != nil {
		return ID{}, err
	}

	parsed, err := xid.FromString(id)
	if err != nil {
		return ID{}, errors.Join(ErrInvalidID, err)
	}

	newID.Inner = parsed
	return newID, nil
}

// NewRawID creates a new xid.ID.
func NewRawID() string {
	return xid.New().String()
}

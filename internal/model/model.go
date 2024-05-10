package model

import (
	"errors"

	"github.com/rs/xid"
)

// ID represents a unique identifier for a resource, combining a resource Type
// and a unique identifier.
type ID struct {
	Inner xid.ID
	Type  ResourceType
}

func (id ID) Validate() error {
	if id.Type < 1 || id.Type > 16 {
		return ErrInvalidID
	}
	return nil
}

// String returns the string representation of the ID. The type is not part of
// the string representation. This is to allow for the ID to be used as a
// label or flag in a database or log aggregation system.
func (id ID) String() string {
	return id.Inner.String()
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

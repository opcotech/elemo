package validate

import (
	"errors"
	"regexp"
	"strings"

	"github.com/rs/xid"
)

const (
	slugMinLength = 3
	slugMaxLength = 50
	projectKeyMin = 2
	projectKeyMax = 6
)

// Canonical slug: lowercase ASCII kebab-case.
var slugPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

var (
	// ErrInvalidSlug is returned when a slug is not canonical kebab-case.
	ErrInvalidSlug = errors.New("invalid slug")
	// ErrReservedSlug is returned when a slug is reserved for routing.
	ErrReservedSlug = errors.New("reserved slug")
	// ErrXIDShapedSlug is returned when a slug would parse as an xid.
	ErrXIDShapedSlug = errors.New("xid-shaped slug")
	// ErrReservedProjectKey is returned when a project key is reserved.
	ErrReservedProjectKey = errors.New("reserved project key")
	// ErrInvalidProjectKey is returned when a project key is not 2-6 letters.
	ErrInvalidProjectKey = errors.New("invalid project key")
	// ErrInvalidRef is returned when a path ref is neither a valid xid nor slug.
	ErrInvalidRef = errors.New("invalid resource ref")
)

// Reserved organization slugs that collide with static routes.
var reservedOrganizationSlugs = map[string]struct{}{
	"join": {},
	"new":  {},
}

// Reserved namespace slugs that collide with static routes.
var reservedNamespaceSlugs = map[string]struct{}{
	"new": {},
}

const reservedProjectKey = "NEW"

// IsXIDShaped reports whether value parses as an xid.
func IsXIDShaped(value string) bool {
	_, err := xid.FromString(value)
	return err == nil
}

// Slug checks canonical kebab-case form and rejects xid-shaped values.
// Reserved names are not checked here; use OrganizationSlug or NamespaceSlug.
func Slug(value string) error {
	if value != strings.TrimSpace(value) {
		return ErrInvalidSlug
	}
	if len(value) < slugMinLength || len(value) > slugMaxLength {
		return ErrInvalidSlug
	}
	if !slugPattern.MatchString(value) {
		return ErrInvalidSlug
	}
	if IsXIDShaped(value) {
		return ErrXIDShapedSlug
	}
	return nil
}

// OrganizationSlug validates a globally unique organization slug.
func OrganizationSlug(value string) error {
	if err := Slug(value); err != nil {
		return err
	}
	if _, reserved := reservedOrganizationSlugs[value]; reserved {
		return ErrReservedSlug
	}
	return nil
}

// NamespaceSlug validates an organization-scoped namespace slug.
func NamespaceSlug(value string) error {
	if err := Slug(value); err != nil {
		return err
	}
	if _, reserved := reservedNamespaceSlugs[value]; reserved {
		return ErrReservedSlug
	}
	return nil
}

// ProjectKey validates a namespace-scoped project key. The value must already
// be uppercased.
func ProjectKey(value string) error {
	if value != strings.ToUpper(value) {
		return ErrInvalidProjectKey
	}
	if len(value) < projectKeyMin || len(value) > projectKeyMax {
		return ErrInvalidProjectKey
	}
	for _, r := range value {
		if r < 'A' || r > 'Z' {
			return ErrInvalidProjectKey
		}
	}
	if value == reservedProjectKey {
		return ErrReservedProjectKey
	}
	return nil
}

// ParseRef treats a valid xid as an identifier; otherwise it requires a
// canonical slug. Reserved values are accepted as slugs so lookup returns
// not-found instead of a client error.
func ParseRef(value string) (id string, slug string, err error) {
	if value == "" {
		return "", "", ErrInvalidRef
	}
	if IsXIDShaped(value) {
		return value, "", nil
	}
	if err := Slug(value); err != nil {
		return "", "", errors.Join(ErrInvalidRef, err)
	}
	return "", value, nil
}

package validate

import (
	"testing"

	"github.com/rs/xid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSlug(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		wantErr error
	}{
		{name: "canonical three chars", input: "abc"},
		{name: "canonical with hyphen", input: "acme-inc"},
		{name: "canonical with digits", input: "org-2"},
		{name: "max length", input: hyphenated(50)},
		{name: "empty", input: "", wantErr: ErrInvalidSlug},
		{name: "too short", input: "ab", wantErr: ErrInvalidSlug},
		{name: "too long", input: hyphenated(51), wantErr: ErrInvalidSlug},
		{name: "uppercase", input: "Acme", wantErr: ErrInvalidSlug},
		{name: "whitespace", input: " acme", wantErr: ErrInvalidSlug},
		{name: "trailing whitespace", input: "acme ", wantErr: ErrInvalidSlug},
		{name: "underscore", input: "acme_inc", wantErr: ErrInvalidSlug},
		{name: "leading hyphen", input: "-acme", wantErr: ErrInvalidSlug},
		{name: "trailing hyphen", input: "acme-", wantErr: ErrInvalidSlug},
		{name: "repeated hyphen", input: "acme--inc", wantErr: ErrInvalidSlug},
		{name: "percent encoded", input: "acme%2Finc", wantErr: ErrInvalidSlug},
		{name: "unicode", input: "ácme", wantErr: ErrInvalidSlug},
		{name: "confusable cyrillic a", input: "аcme", wantErr: ErrInvalidSlug},
		{name: "slash", input: "acme/inc", wantErr: ErrInvalidSlug},
		{name: "dot", input: "acme.inc", wantErr: ErrInvalidSlug},
		{name: "xid shaped", input: xid.New().String(), wantErr: ErrXIDShapedSlug},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := Slug(tt.input)
			if tt.wantErr == nil {
				assert.NoError(t, err)
				return
			}
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestOrganizationSlug(t *testing.T) {
	t.Parallel()

	assert.NoError(t, OrganizationSlug("acme"))
	assert.ErrorIs(t, OrganizationSlug("new"), ErrReservedSlug)
	assert.ErrorIs(t, OrganizationSlug("join"), ErrReservedSlug)
	assert.ErrorIs(t, OrganizationSlug("NEW"), ErrInvalidSlug)
}

func TestNamespaceSlug(t *testing.T) {
	t.Parallel()

	assert.NoError(t, NamespaceSlug("platform"))
	assert.ErrorIs(t, NamespaceSlug("new"), ErrReservedSlug)
	assert.NoError(t, NamespaceSlug("join"))
}

func TestProjectKey(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		wantErr error
	}{
		{name: "two letters", input: "AB"},
		{name: "six letters", input: "ABCDEF"},
		{name: "lowercase", input: "ab", wantErr: ErrInvalidProjectKey},
		{name: "too short", input: "A", wantErr: ErrInvalidProjectKey},
		{name: "too long", input: "ABCDEFG", wantErr: ErrInvalidProjectKey},
		{name: "digits", input: "AB1", wantErr: ErrInvalidProjectKey},
		{name: "reserved", input: "NEW", wantErr: ErrReservedProjectKey},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := ProjectKey(tt.input)
			if tt.wantErr == nil {
				assert.NoError(t, err)
				return
			}
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestParseRef(t *testing.T) {
	t.Parallel()

	id := xid.New().String()

	gotID, gotSlug, err := ParseRef(id)
	require.NoError(t, err)
	assert.Equal(t, id, gotID)
	assert.Empty(t, gotSlug)

	gotID, gotSlug, err = ParseRef("acme")
	require.NoError(t, err)
	assert.Empty(t, gotID)
	assert.Equal(t, "acme", gotSlug)

	_, _, err = ParseRef("")
	assert.ErrorIs(t, err, ErrInvalidRef)

	_, _, err = ParseRef("Acme")
	assert.ErrorIs(t, err, ErrInvalidRef)

	_, _, err = ParseRef("new")
	require.NoError(t, err)
}

func hyphenated(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = 'a'
	}
	return string(b)
}

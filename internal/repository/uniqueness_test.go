package repository

import (
	"errors"
	"testing"

	"github.com/neo4j/neo4j-go-driver/v6/neo4j"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMapUniquenessError(t *testing.T) {
	t.Parallel()

	t.Run("nil", func(t *testing.T) {
		t.Parallel()
		require.NoError(t, mapUniquenessError(nil))
	})

	t.Run("non neo4j", func(t *testing.T) {
		t.Parallel()
		err := errors.New("boom")
		assert.Equal(t, err, mapUniquenessError(err))
	})

	tests := []struct {
		name string
		msg  string
		want error
	}{
		{name: "organization slug constraint", msg: "ConstraintValidationFailed: " + constraintOrganizationSlug, want: ErrSlugConflict},
		{name: "namespace slug constraint", msg: "ConstraintValidationFailed: " + constraintNamespaceOrgSlug, want: ErrSlugConflict},
		{name: "project key constraint", msg: "ConstraintValidationFailed: " + constraintProjectNamespaceKey, want: ErrProjectKeyConflict},
		{name: "slug property fallback", msg: "already exists with label `Organization` and property `slug` = 'acme'", want: ErrSlugConflict},
		{name: "key property fallback", msg: "already exists with label `Project` and property `key` = 'PLAT'", want: ErrProjectKeyConflict},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := mapUniquenessError(&neo4j.Neo4jError{
				Code: neo4jConstraintValidationFailed,
				Msg:  tt.msg,
			})
			assert.ErrorIs(t, err, tt.want)
		})
	}

	t.Run("unrelated uniqueness", func(t *testing.T) {
		t.Parallel()
		original := &neo4j.Neo4jError{
			Code: neo4jConstraintValidationFailed,
			Msg:  "ConstraintValidationFailed: organization_email_unique",
		}
		err := mapUniquenessError(original)
		assert.Equal(t, original, err)
		assert.NotErrorIs(t, err, ErrSlugConflict)
	})
}

package repository

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/opcotech/elemo/internal/model"
)

func TestListSearchableRecords_Validation(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("nil database", func(t *testing.T) {
		t.Parallel()
		_, err := ListSearchableRecords(ctx, nil, model.ResourceTypeIssue, "", 10)
		assert.ErrorIs(t, err, ErrNoDriver)
	})

	t.Run("invalid resource type", func(t *testing.T) {
		t.Parallel()
		_, err := ListSearchableRecords(ctx, &Neo4jDatabase{}, model.ResourceType(0), "", 10)
		assert.ErrorIs(t, err, model.ErrInvalidResourceType)
	})

	t.Run("invalid limit", func(t *testing.T) {
		t.Parallel()
		_, err := ListSearchableRecords(ctx, &Neo4jDatabase{}, model.ResourceTypeIssue, "", 0)
		assert.ErrorIs(t, err, ErrInvalidPageSize)
	})
}

func TestListSearchableRecordsByIDs_Validation(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("nil database", func(t *testing.T) {
		t.Parallel()
		_, err := ListSearchableRecordsByIDs(ctx, nil, model.ResourceTypeIssue, nil)
		assert.ErrorIs(t, err, ErrNoDriver)
	})

	t.Run("empty ids", func(t *testing.T) {
		t.Parallel()
		got, err := ListSearchableRecordsByIDs(ctx, &Neo4jDatabase{}, model.ResourceTypeIssue, nil)
		require.NoError(t, err)
		assert.Empty(t, got)
	})
}

func TestListSearchableIDs_Validation(t *testing.T) {
	t.Parallel()

	_, err := ListSearchableIDs(context.Background(), nil, model.ResourceTypeIssue)
	assert.ErrorIs(t, err, ErrNoDriver)
}

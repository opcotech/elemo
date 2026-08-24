package repository

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/opcotech/elemo/internal/model"
)

func TestCompileCursorBounds(t *testing.T) {
	t.Parallel()

	t.Run("defaults order and sets limit", func(t *testing.T) {
		t.Parallel()

		params := map[string]any{}
		bounds, err := compileCursorBounds("u", CursorPage{Size: 10}, SortDirectionUnknown, params)
		require.NoError(t, err)
		assert.Equal(t, SortDirectionDesc, bounds.Order)
		assert.Equal(t, 11, params["limit"])
		assert.Empty(t, bounds.Where)
		assert.NotContains(t, params, "cursor_id")
	})

	t.Run("empty token does not emit cursor params", func(t *testing.T) {
		t.Parallel()

		empty := ""
		params := map[string]any{}
		bounds, err := compileCursorBounds("u", CursorPage{Size: 10, Token: &empty}, SortDirectionAsc, params)
		require.NoError(t, err)
		assert.Nil(t, bounds.Page.Token)
		assert.Empty(t, bounds.Where)
		assert.NotContains(t, params, "cursor_id")
	})

	t.Run("token emits cursor where and params", func(t *testing.T) {
		t.Parallel()

		id := model.MustNewID(model.ResourceTypeUser)
		token, err := EncodeCursor(id)
		require.NoError(t, err)

		params := map[string]any{}
		bounds, err := compileCursorBounds("u", CursorPage{Size: 10, Token: &token}, SortDirectionDesc, params)
		require.NoError(t, err)
		assert.Equal(t, "u.id < $cursor_id", bounds.Where)
		assert.Equal(t, id.String(), params["cursor_id"])
		assert.NotContains(t, params, "cursor_created_at")
	})

	t.Run("rejects invalid page size", func(t *testing.T) {
		t.Parallel()

		_, err := compileCursorBounds("u", CursorPage{Size: MaxPageSize + 1}, SortDirectionUnknown, map[string]any{})
		require.ErrorIs(t, err, ErrInvalidPageSize)
	})

	t.Run("rejects unsupported order", func(t *testing.T) {
		t.Parallel()

		_, err := compileCursorBounds("u", CursorPage{Size: 10}, SortDirection(100), map[string]any{})
		require.ErrorIs(t, err, ErrUnsupportedOrder)
	})
}

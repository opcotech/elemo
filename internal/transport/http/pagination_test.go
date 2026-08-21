package http

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg/convert"
	"github.com/opcotech/elemo/internal/repository"
)

func TestCursorPageFromParams(t *testing.T) {
	t.Parallel()

	t.Run("defaults size", func(t *testing.T) {
		t.Parallel()

		got, err := cursorPageFromParams(nil, nil)
		require.NoError(t, err)
		assert.Equal(t, DefaultPageSize, got.Size)
		assert.Nil(t, got.Token)
	})

	t.Run("coerces empty token to nil", func(t *testing.T) {
		t.Parallel()

		got, err := cursorPageFromParams(convert.ToPointer(10), convert.ToPointer(""))
		require.NoError(t, err)
		assert.Equal(t, 10, got.Size)
		assert.Nil(t, got.Token)
	})

	t.Run("rejects invalid page size", func(t *testing.T) {
		t.Parallel()

		_, err := cursorPageFromParams(convert.ToPointer(repository.MaxPageSize+1), nil)
		require.Error(t, err)
		assert.ErrorIs(t, err, repository.ErrInvalidPageSize)
	})

	t.Run("passes through opaque cursor", func(t *testing.T) {
		t.Parallel()

		token := "not-a-token"
		got, err := cursorPageFromParams(convert.ToPointer(10), &token)
		require.NoError(t, err)
		require.NotNil(t, got.Token)
		assert.Equal(t, token, *got.Token)
	})

	t.Run("keeps valid cursor", func(t *testing.T) {
		t.Parallel()

		id := model.MustNewID(model.ResourceTypeIssue)
		token, err := repository.EncodeCursor(id)
		require.NoError(t, err)

		got, err := cursorPageFromParams(convert.ToPointer(25), &token)
		require.NoError(t, err)
		assert.Equal(t, 25, got.Size)
		require.NotNil(t, got.Token)
		assert.Equal(t, token, *got.Token)
	})
}

package repository

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/opcotech/elemo/internal/model"
)

func TestCompositeIDArg(t *testing.T) {
	t.Parallel()

	t.Run("nil", func(t *testing.T) {
		t.Parallel()
		assert.Nil(t, compositeIDArg(nil))
	})

	t.Run("value", func(t *testing.T) {
		t.Parallel()
		id := model.MustNewID(model.ResourceTypeUser)
		assert.Equal(t, id.Composite(), compositeIDArg(&id))
	})
}

func TestParseOptionalCompositeID(t *testing.T) {
	t.Parallel()

	t.Run("nil", func(t *testing.T) {
		t.Parallel()
		got, err := parseOptionalCompositeID(nil)
		require.NoError(t, err)
		assert.Nil(t, got)
	})

	t.Run("empty", func(t *testing.T) {
		t.Parallel()
		raw := ""
		got, err := parseOptionalCompositeID(&raw)
		require.NoError(t, err)
		assert.Nil(t, got)
	})

	t.Run("valid", func(t *testing.T) {
		t.Parallel()
		id := model.MustNewID(model.ResourceTypeUser)
		raw := id.Composite()
		got, err := parseOptionalCompositeID(&raw)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, id, *got)
	})

	t.Run("invalid", func(t *testing.T) {
		t.Parallel()
		raw := "not-a-composite"
		_, err := parseOptionalCompositeID(&raw)
		assert.ErrorIs(t, err, model.ErrInvalidID)
	})
}

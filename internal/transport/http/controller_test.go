package http

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/opcotech/elemo/internal/service"
)

func TestWithProjectService(t *testing.T) {
	t.Parallel()

	t.Run("set project service", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ps := service.NewMockProjectService(ctrl)
		var c baseController
		err := WithProjectService(ps)(&c)
		require.NoError(t, err)
		assert.Equal(t, ps, c.projectService)
	})

	t.Run("nil project service", func(t *testing.T) {
		t.Parallel()
		var c baseController
		err := WithProjectService(nil)(&c)
		assert.ErrorIs(t, err, ErrNoProjectService)
	})
}

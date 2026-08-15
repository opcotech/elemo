package http

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/opcotech/elemo/internal/service"
	"github.com/opcotech/elemo/internal/testutil/mock"
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

func TestWithIssueService(t *testing.T) {
	t.Parallel()

	t.Run("set issue service", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		is := service.NewMockIssueService(ctrl)
		var c baseController
		err := WithIssueService(is)(&c)
		require.NoError(t, err)
		assert.Equal(t, is, c.issueService)
	})

	t.Run("nil issue service", func(t *testing.T) {
		t.Parallel()
		var c baseController
		err := WithIssueService(nil)(&c)
		assert.ErrorIs(t, err, ErrNoIssueService)
	})
}

func TestWithLabelService(t *testing.T) {
	t.Parallel()

	t.Run("set label service", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ls := service.NewMockLabelService(ctrl)
		var c baseController
		err := WithLabelService(ls)(&c)
		require.NoError(t, err)
		assert.Equal(t, ls, c.labelService)
	})

	t.Run("nil label service", func(t *testing.T) {
		t.Parallel()
		var c baseController
		err := WithLabelService(nil)(&c)
		assert.ErrorIs(t, err, ErrNoLabelService)
	})
}

func TestWithEmailService(t *testing.T) {
	t.Parallel()

	t.Run("set email service", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		es := mock.NewEmailService(ctrl)
		var c baseController
		err := WithEmailService(es)(&c)
		require.NoError(t, err)
		assert.Equal(t, es, c.emailService)
	})

	t.Run("nil email service", func(t *testing.T) {
		t.Parallel()
		var c baseController
		err := WithEmailService(nil)(&c)
		assert.ErrorIs(t, err, ErrNoEmailService)
	})
}

func TestWithNotificationService(t *testing.T) {
	t.Parallel()

	t.Run("nil notification service", func(t *testing.T) {
		t.Parallel()
		var c baseController
		err := WithNotificationService(nil)(&c)
		assert.ErrorIs(t, err, ErrNoNotificationService)
	})
}

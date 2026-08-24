package service_test

import (
	"context"
	"testing"

	"github.com/opcotech/elemo/internal/service"
	mocksvc "github.com/opcotech/elemo/internal/service/mock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg"
)

func TestCtxUserID(t *testing.T) {
	t.Parallel()

	t.Run("returns authenticated user", func(t *testing.T) {
		t.Parallel()
		userID := model.MustNewID(model.ResourceTypeUser)
		got, err := service.CtxUserID(context.WithValue(context.Background(), pkg.CtxKeyUserID, userID))
		require.NoError(t, err)
		assert.Equal(t, userID, got)
	})

	t.Run("missing user", func(t *testing.T) {
		t.Parallel()
		_, err := service.CtxUserID(context.Background())
		require.ErrorIs(t, err, service.ErrNoUser)
	})
}

func TestRequireAction(t *testing.T) {
	t.Parallel()

	resource := model.MustNewID(model.ResourceTypeProject)
	ctx := context.Background()

	t.Run("allows", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		perm := mocksvc.NewMockPermissionService(ctrl)
		perm.EXPECT().CtxUserHas(ctx, resource, model.ActionProjectRead).Return(true, nil)
		require.NoError(t, service.RequireAction(ctx, perm, resource, model.ActionProjectRead))
	})

	t.Run("denies", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		perm := mocksvc.NewMockPermissionService(ctrl)
		perm.EXPECT().CtxUserHas(ctx, resource, model.ActionProjectRead).Return(false, nil)
		err := service.RequireAction(ctx, perm, resource, model.ActionProjectRead)
		require.ErrorIs(t, err, service.ErrNoPermission)
	})

	t.Run("propagates infrastructure error", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		perm := mocksvc.NewMockPermissionService(ctrl)
		perm.EXPECT().CtxUserHas(ctx, resource, model.ActionProjectRead).Return(false, assert.AnError)
		err := service.RequireAction(ctx, perm, resource, model.ActionProjectRead)
		require.ErrorIs(t, err, assert.AnError)
	})
}

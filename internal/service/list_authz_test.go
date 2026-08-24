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
)

func TestResolvedListScopeIDs(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	root := model.MustNewID(model.ResourceTypeNamespace)
	org := model.MustNewID(model.ResourceTypeOrganization)
	other := model.MustNewID(model.ResourceTypeProject)

	t.Run("no grants", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		perm := mocksvc.NewMockPermissionService(ctrl)
		perm.EXPECT().CtxUserListGrantScopes(ctx, model.ActionProjectRead).Return(nil, nil)

		scopeIDs, allowed, err := service.ResolvedListScopeIDs(ctx, perm, root, model.ActionProjectRead)
		require.NoError(t, err)
		assert.False(t, allowed)
		assert.Nil(t, scopeIDs)
	})

	t.Run("covering grant skips EXISTS", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		perm := mocksvc.NewMockPermissionService(ctrl)
		perm.EXPECT().CtxUserListGrantScopes(ctx, model.ActionProjectRead).Return([]model.ID{org}, nil)
		perm.EXPECT().ListScopeAncestry(ctx, root).Return([]model.ID{root, org}, nil)

		scopeIDs, allowed, err := service.ResolvedListScopeIDs(ctx, perm, root, model.ActionProjectRead)
		require.NoError(t, err)
		assert.True(t, allowed)
		assert.Nil(t, scopeIDs)
	})

	t.Run("narrow grant keeps scope ids", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		perm := mocksvc.NewMockPermissionService(ctrl)
		perm.EXPECT().CtxUserListGrantScopes(ctx, model.ActionProjectRead).Return([]model.ID{other}, nil)
		perm.EXPECT().ListScopeAncestry(ctx, root).Return([]model.ID{root, org}, nil)

		scopeIDs, allowed, err := service.ResolvedListScopeIDs(ctx, perm, root, model.ActionProjectRead)
		require.NoError(t, err)
		assert.True(t, allowed)
		assert.Equal(t, []model.ID{other}, scopeIDs)
	})
}

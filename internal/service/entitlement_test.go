package service_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/opcotech/elemo/internal/deployment"
	"github.com/opcotech/elemo/internal/entitlement"
	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/service"
	mocksvc "github.com/opcotech/elemo/internal/service/mock"
)

type reporterFunc func(context.Context) (entitlement.Status, error)

func (f reporterFunc) Status(ctx context.Context) (entitlement.Status, error) {
	return f(ctx)
}

func TestEntitlementServiceAuthorization(t *testing.T) {
	t.Parallel()

	t.Run("installation administrator can read status", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		permissions := mocksvc.NewMockPermissionService(ctrl)
		permissions.EXPECT().CtxUserHas(gomock.Any(), model.InstallationID(), model.ActionOrganizationCreate).Return(true, nil)
		reporter := reporterFunc(func(context.Context) (entitlement.Status, error) {
			return entitlement.Status{DeploymentMode: deployment.ModeSelfHosted}, nil
		})
		svc, err := service.NewEntitlementService(reporter, permissions)
		require.NoError(t, err)

		status, err := svc.Get(context.Background())
		require.NoError(t, err)
		require.Equal(t, deployment.ModeSelfHosted, status.DeploymentMode)
	})

	t.Run("ordinary user is denied", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		permissions := mocksvc.NewMockPermissionService(ctrl)
		permissions.EXPECT().CtxUserHas(gomock.Any(), model.InstallationID(), model.ActionOrganizationCreate).Return(false, nil)
		reporter := reporterFunc(func(context.Context) (entitlement.Status, error) {
			t.Fatal("status reporter must not run for unauthorized callers")
			return entitlement.Status{}, nil
		})
		svc, err := service.NewEntitlementService(reporter, permissions)
		require.NoError(t, err)

		_, err = svc.Get(context.Background())
		require.ErrorIs(t, err, service.ErrNoPermission)
	})
}

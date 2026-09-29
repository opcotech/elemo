package http

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/opcotech/elemo/internal/deployment"
	"github.com/opcotech/elemo/internal/entitlement"
	"github.com/opcotech/elemo/internal/entitlement/license"
	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/service"
	mocksvc "github.com/opcotech/elemo/internal/service/mock"
	"github.com/opcotech/elemo/internal/transport/http/api"
)

func TestSystemController_V1SystemEntitlements(t *testing.T) {
	t.Parallel()

	t.Run("self hosted", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		ent := mocksvc.NewMockEntitlementService(ctrl)
		ent.EXPECT().Get(gomock.Any()).Return(entitlement.Status{DeploymentMode: deployment.ModeSelfHosted}, nil)

		c, err := NewSystemController(mocksvc.NewMockSystemService(ctrl), ent)
		require.NoError(t, err)
		resp, err := c.V1SystemEntitlements(context.Background(), api.V1SystemEntitlementsRequestObject{})
		require.NoError(t, err)
		got, ok := resp.(api.V1SystemEntitlements200JSONResponse)
		require.True(t, ok)
		assert.Equal(t, api.SystemEntitlementsDeploymentModeSelfHosted, got.DeploymentMode)
		assert.Nil(t, got.Airgap)
	})

	t.Run("airgap grace", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		ent := mocksvc.NewMockEntitlementService(ctrl)
		ent.EXPECT().Get(gomock.Any()).Return(entitlement.Status{
			DeploymentMode: deployment.ModeAirGap,
			AirGap: &entitlement.AirGapStatus{
				State:          license.StateGrace,
				Reason:         "in grace period",
				InstallationID: "11111111-1111-1111-1111-111111111111",
				SeatsLicensed:  5,
				SeatsActive:    3,
			},
		}, nil)

		c, err := NewSystemController(mocksvc.NewMockSystemService(ctrl), ent)
		require.NoError(t, err)
		resp, err := c.V1SystemEntitlements(context.Background(), api.V1SystemEntitlementsRequestObject{})
		require.NoError(t, err)
		got, ok := resp.(api.V1SystemEntitlements200JSONResponse)
		require.True(t, ok)
		require.NotNil(t, got.Airgap)
		assert.Equal(t, api.SystemEntitlementsDeploymentModeAirgap, got.DeploymentMode)
		assert.Equal(t, api.SystemAirGapEntitlementStateGrace, got.Airgap.State)
		assert.Equal(t, 3, got.Airgap.SeatsActive)
	})

	t.Run("forbidden", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		ent := mocksvc.NewMockEntitlementService(ctrl)
		ent.EXPECT().Get(gomock.Any()).Return(entitlement.Status{}, service.ErrNoPermission)

		c, err := NewSystemController(mocksvc.NewMockSystemService(ctrl), ent)
		require.NoError(t, err)
		resp, err := c.V1SystemEntitlements(context.Background(), api.V1SystemEntitlementsRequestObject{})
		require.NoError(t, err)
		_, ok := resp.(api.V1SystemEntitlements403JSONResponse)
		assert.True(t, ok)
	})

	t.Run("internal errors are not exposed", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		ent := mocksvc.NewMockEntitlementService(ctrl)
		ent.EXPECT().Get(gomock.Any()).Return(entitlement.Status{}, errors.New("neo4j connection secret"))

		c, err := NewSystemController(mocksvc.NewMockSystemService(ctrl), ent)
		require.NoError(t, err)
		resp, err := c.V1SystemEntitlements(context.Background(), api.V1SystemEntitlementsRequestObject{})
		require.NoError(t, err)
		got, ok := resp.(api.V1SystemEntitlements500JSONResponse)
		require.True(t, ok)
		assert.Equal(t, "Internal Server Error", got.Message)
	})

	t.Run("invalid installation identity fails the response", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		ent := mocksvc.NewMockEntitlementService(ctrl)
		ent.EXPECT().Get(gomock.Any()).Return(entitlement.Status{
			DeploymentMode: deployment.ModeAirGap,
			AirGap: &entitlement.AirGapStatus{
				State:          license.StateInvalid,
				InstallationID: "invalid",
			},
		}, nil)

		c, err := NewSystemController(mocksvc.NewMockSystemService(ctrl), ent)
		require.NoError(t, err)
		resp, err := c.V1SystemEntitlements(context.Background(), api.V1SystemEntitlementsRequestObject{})
		require.NoError(t, err)
		got, ok := resp.(api.V1SystemEntitlements500JSONResponse)
		require.True(t, ok)
		assert.Equal(t, "Internal Server Error", got.Message)
	})
}

func TestUserController_V1UsersCreate_EntitlementConflict(t *testing.T) {
	t.Parallel()

	body := &api.V1UsersCreateJSONRequestBody{
		Email:     "user@example.com",
		FirstName: "Test",
		LastName:  "User",
		Password:  "super-secret",
		Username:  "test-user",
	}

	t.Run("seat limit reached", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		users := mocksvc.NewMockUserService(ctrl)
		users.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil, entitlement.ErrSeatLimitReached)
		c, err := NewUserController(users, mocksvc.NewMockEmailService(ctrl))
		require.NoError(t, err)

		resp, err := c.V1UsersCreate(context.Background(), api.V1UsersCreateRequestObject{Body: body})
		require.NoError(t, err)
		got, ok := resp.(api.V1UsersCreate409JSONResponse)
		require.True(t, ok)
		require.NotNil(t, got.Code)
		assert.Equal(t, api.HTTPErrorCodeSeatLimitReached, *got.Code)
	})

	t.Run("activation denied", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		users := mocksvc.NewMockUserService(ctrl)
		users.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil, entitlement.ErrActivationDenied)
		c, err := NewUserController(users, mocksvc.NewMockEmailService(ctrl))
		require.NoError(t, err)

		resp, err := c.V1UsersCreate(context.Background(), api.V1UsersCreateRequestObject{Body: body})
		require.NoError(t, err)
		got, ok := resp.(api.V1UsersCreate409JSONResponse)
		require.True(t, ok)
		require.NotNil(t, got.Code)
		assert.Equal(t, api.HTTPErrorCodeActivationDenied, *got.Code)
	})

	t.Run("mutation denied", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		users := mocksvc.NewMockUserService(ctrl)
		users.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil, entitlement.ErrMutationDenied)
		c, err := NewUserController(users, mocksvc.NewMockEmailService(ctrl))
		require.NoError(t, err)

		resp, err := c.V1UsersCreate(context.Background(), api.V1UsersCreateRequestObject{Body: body})
		require.NoError(t, err)
		got, ok := resp.(api.V1UsersCreate409JSONResponse)
		require.True(t, ok)
		require.NotNil(t, got.Code)
		assert.Equal(t, api.HTTPErrorCodeEntitlementReadOnly, *got.Code)
	})
}

func TestUserController_V1UserUpdate_EntitlementConflict(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	users := mocksvc.NewMockUserService(ctrl)
	users.EXPECT().Update(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, entitlement.ErrSeatLimitReached)
	c, err := NewUserController(users, mocksvc.NewMockEmailService(ctrl))
	require.NoError(t, err)

	status := api.UserStatusActive
	resp, err := c.V1UserUpdate(context.Background(), api.V1UserUpdateRequestObject{
		Id: "9bsv0s46s6s002p9ltq0",
		Body: &api.V1UserUpdateJSONRequestBody{
			Status: &status,
		},
	})
	require.NoError(t, err)
	got, ok := resp.(api.V1UserUpdate409JSONResponse)
	require.True(t, ok)
	require.NotNil(t, got.Code)
	assert.Equal(t, api.HTTPErrorCodeSeatLimitReached, *got.Code)
}

func TestOrganizationController_V1OrganizationMembersAccept_EntitlementConflict(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	orgID := "9bsv0s46s6s002p9ltq0"
	orgModelID, err := model.NewIDFromString(orgID, model.ResourceTypeOrganization.String())
	require.NoError(t, err)
	organizations := mocksvc.NewMockOrganizationService(ctrl)
	organizations.EXPECT().Resolve(gomock.Any(), gomock.Any(), "").Return(&service.Organization{
		ID: orgModelID,
	}, nil)
	organizations.EXPECT().AcceptInvitation(gomock.Any(), gomock.Any(), gomock.Any()).Return(entitlement.ErrActivationDenied)
	c, err := NewOrganizationController(
		organizations,
		mocksvc.NewMockRoleService(ctrl),
		mocksvc.NewMockTeamService(ctrl),
		mocksvc.NewMockUserService(ctrl),
	)
	require.NoError(t, err)

	resp, err := c.V1OrganizationMembersAccept(context.Background(), api.V1OrganizationMembersAcceptRequestObject{
		OrganizationRef: orgID,
		Body: &api.V1OrganizationMembersAcceptJSONRequestBody{
			Token: "invitation-token",
		},
	})
	require.NoError(t, err)
	got, ok := resp.(api.V1OrganizationMembersAccept409JSONResponse)
	require.True(t, ok)
	require.NotNil(t, got.Code)
	assert.Equal(t, api.HTTPErrorCodeActivationDenied, *got.Code)
}

func TestOrganizationController_V1OrganizationMembersAccept_PasswordRequired(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	orgID := "9bsv0s46s6s002p9ltq0"
	orgModelID, err := model.NewIDFromString(orgID, model.ResourceTypeOrganization.String())
	require.NoError(t, err)
	organizations := mocksvc.NewMockOrganizationService(ctrl)
	organizations.EXPECT().Resolve(gomock.Any(), gomock.Any(), "").Return(&service.Organization{
		ID: orgModelID,
	}, nil)
	organizations.EXPECT().AcceptInvitation(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(service.ErrOrganizationInvitePassword)
	c, err := NewOrganizationController(
		organizations,
		mocksvc.NewMockRoleService(ctrl),
		mocksvc.NewMockTeamService(ctrl),
		mocksvc.NewMockUserService(ctrl),
	)
	require.NoError(t, err)

	resp, err := c.V1OrganizationMembersAccept(context.Background(), api.V1OrganizationMembersAcceptRequestObject{
		OrganizationRef: orgID,
		Body: &api.V1OrganizationMembersAcceptJSONRequestBody{
			Token: "invitation-token",
		},
	})
	require.NoError(t, err)
	_, ok := resp.(api.V1OrganizationMembersAccept400JSONResponse)
	require.True(t, ok)
}

func TestIssueController_V1ProjectsIssuesCreate_ReadOnly(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	issues := mocksvc.NewMockIssueService(ctrl)
	issues.EXPECT().Create(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, entitlement.ErrMutationDenied)
	c, _, _ := newTestIssueController(t, ctrl, issues)

	resp, err := c.V1ProjectsIssuesCreate(context.Background(), api.V1ProjectsIssuesCreateRequestObject{
		ProjectId: model.MustNewID(model.ResourceTypeProject).String(),
		Body: &api.V1ProjectsIssuesCreateJSONRequestBody{
			Kind:  api.IssueKindStory,
			Title: "Blocked write",
		},
	})
	require.NoError(t, err)
	got, ok := resp.(api.V1ProjectsIssuesCreate409JSONResponse)
	require.True(t, ok)
	require.NotNil(t, got.Code)
	assert.Equal(t, api.HTTPErrorCodeEntitlementReadOnly, *got.Code)
}

func TestUserController_V1UserResetPassword_ExpiredToken(t *testing.T) {
	t.Parallel()

	userID := model.MustNewID(model.ResourceTypeUser)
	user := &service.User{
		ID:        userID,
		Email:     "user@example.com",
		FirstName: "Ada",
		LastName:  "Lovelace",
	}
	body := &api.V1UserResetPasswordJSONRequestBody{
		Token:    "expired-token",
		Password: "new-password",
	}

	t.Run("expired token with claims sends a new reset email", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		users := mocksvc.NewMockUserService(ctrl)
		users.EXPECT().VerifyToken(gomock.Any(), body.Token).Return(map[string]any{"user_id": userID.String()}, service.ErrExpiredToken)
		users.EXPECT().Get(gomock.Any(), userID).Return(user, nil)
		users.EXPECT().DeleteToken(gomock.Any(), userID, model.UserTokenContextResetPassword).Return(nil)
		users.EXPECT().CreateToken(gomock.Any(), userID, user.Email, model.UserTokenContextResetPassword, nil).Return("new-token", nil)
		emails := mocksvc.NewMockEmailService(ctrl)
		emails.EXPECT().SendAuthPasswordResetEmail(gomock.Any(), gomock.Any(), "new-token").Return(nil)

		c, err := NewUserController(users, emails)
		require.NoError(t, err)
		resp, err := c.V1UserResetPassword(context.Background(), api.V1UserResetPasswordRequestObject{Body: body})
		require.NoError(t, err)
		_, ok := resp.(api.V1UserResetPassword204Response)
		assert.True(t, ok)
	})

	t.Run("expired token without claims does not panic", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		users := mocksvc.NewMockUserService(ctrl)
		users.EXPECT().VerifyToken(gomock.Any(), body.Token).Return(nil, service.ErrExpiredToken)
		c, err := NewUserController(users, mocksvc.NewMockEmailService(ctrl))
		require.NoError(t, err)
		resp, err := c.V1UserResetPassword(context.Background(), api.V1UserResetPasswordRequestObject{Body: body})
		require.NoError(t, err)
		_, ok := resp.(api.V1UserResetPassword400JSONResponse)
		assert.True(t, ok)
	})

	t.Run("password update conflict is 409", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		users := mocksvc.NewMockUserService(ctrl)
		users.EXPECT().VerifyToken(gomock.Any(), body.Token).Return(map[string]any{"user_id": userID.String()}, nil)
		users.EXPECT().Get(gomock.Any(), userID).Return(user, nil)
		users.EXPECT().Update(gomock.Any(), userID, gomock.Any()).Return(nil, entitlement.ErrMutationDenied)
		c, err := NewUserController(users, mocksvc.NewMockEmailService(ctrl))
		require.NoError(t, err)
		resp, err := c.V1UserResetPassword(context.Background(), api.V1UserResetPasswordRequestObject{Body: body})
		require.NoError(t, err)
		got, ok := resp.(api.V1UserResetPassword409JSONResponse)
		require.True(t, ok)
		require.NotNil(t, got.Code)
		assert.Equal(t, api.HTTPErrorCodeEntitlementReadOnly, *got.Code)
	})
}

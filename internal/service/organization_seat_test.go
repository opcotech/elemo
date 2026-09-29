package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/opcotech/elemo/internal/entitlement"
	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg/auth"
	mocklog "github.com/opcotech/elemo/internal/pkg/log/mock"
	mocktrace "github.com/opcotech/elemo/internal/pkg/tracing/mock"
	"github.com/opcotech/elemo/internal/repository"
	mockrepo "github.com/opcotech/elemo/internal/repository/mock"
	"github.com/opcotech/elemo/internal/service"
	mocksvc "github.com/opcotech/elemo/internal/service/mock"
	testModel "github.com/opcotech/elemo/internal/testutil/model"
)

func TestOrganizationService_AcceptInvitation_SeatLimit(t *testing.T) {
	t.Parallel()

	userID := model.MustNewID(model.ResourceTypeUser)
	orgID := model.MustNewID(model.ResourceTypeOrganization)
	ctx := context.Background()
	tokenData := map[string]any{
		"organization_id": orgID.String(),
		"user_id":         userID.String(),
	}
	publicToken, tokenHash, err := auth.GenerateToken(model.UserTokenContextInvite.String(), tokenData)
	require.NoError(t, err)

	user := testModel.NewUser()
	user.ID = userID
	user.Status = model.UserStatusPending
	now := time.Now()
	userToken := &repository.UserToken{
		ID:        model.MustNewID(model.ResourceTypeUserToken),
		UserID:    userID,
		SentTo:    user.Email,
		Token:     tokenHash,
		Context:   model.UserTokenContextInvite,
		CreatedAt: &now,
	}

	newOrgService := func(
		ctrl *gomock.Controller,
		userRepo repository.UserRepository,
		seats entitlement.SeatPolicy,
	) service.OrganizationService {
		t.Helper()
		span := mocktrace.NewMockSpan(ctrl)
		span.EXPECT().End(gomock.Len(0))
		tracer := mocktrace.NewMockTracer(ctrl)
		tracer.EXPECT().Start(gomock.Any(), "service.organizationService/AcceptInvitation", gomock.Len(0)).Return(ctx, span)

		orgRepo := mockrepo.NewMockOrganizationRepository(ctrl)

		userTokenRepo := mockrepo.NewMockUserTokenRepository(ctrl)
		userTokenRepo.EXPECT().Get(gomock.Any(), userID, model.UserTokenContextInvite).Return(userToken, nil)
		userTokenRepo.EXPECT().Delete(gomock.Any(), userID, model.UserTokenContextInvite).Return(nil).MaxTimes(1)

		logger := mocklog.NewMockLogger(ctrl)
		logger.EXPECT().Warn(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

		svc, err := service.NewOrganizationService(
			orgRepo,
			userRepo,
			userTokenRepo,
			mockrepo.NewMockRoleRepository(ctrl),
			mocksvc.NewMockPermissionService(ctrl),
			seats,
			mocksvc.NewMockEmailService(ctrl),
			mocksvc.NewMockNotificationService(ctrl),
			mocksvc.NewMockSearchService(ctrl),
			service.WithLogger(logger),
			service.WithTracer(tracer),
		)
		require.NoError(t, err)
		return svc
	}

	t.Run("activate pending user when seats remain", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		userRepo := mockrepo.NewMockUserRepository(ctrl)
		userRepo.EXPECT().Get(gomock.Any(), userID, repository.UserDetailProjection()).Return(user, nil)
		userRepo.EXPECT().AcceptInvitation(gomock.Any(), userID, orgID, gomock.Any(), gomock.Cond(func(got *repository.ActivationAuthorization) bool {
			return got != nil && *got == repository.FiniteActivation(2)
		}), nil).Return(&repository.User{
			ID:     userID,
			Status: model.UserStatusActive,
		}, nil)

		svc := newOrgService(ctrl, userRepo, finiteSeats{max: 2})
		err := svc.AcceptInvitation(ctx, orgID, service.AcceptOrganizationInvitationOpts{
			Token:    publicToken,
			Password: "password123",
		})
		require.NoError(t, err)
	})

	t.Run("reject invitation when the seat limit is reached", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		userRepo := mockrepo.NewMockUserRepository(ctrl)
		userRepo.EXPECT().Get(gomock.Any(), userID, repository.UserDetailProjection()).Return(user, nil)
		userRepo.EXPECT().AcceptInvitation(gomock.Any(), userID, orgID, gomock.Any(), gomock.Any(), nil).Return(nil, repository.ErrSeatLimitReached)

		svc := newOrgService(ctrl, userRepo, finiteSeats{max: 1})
		err := svc.AcceptInvitation(ctx, orgID, service.AcceptOrganizationInvitationOpts{
			Token:    publicToken,
			Password: "password123",
		})
		require.ErrorIs(t, err, entitlement.ErrSeatLimitReached)
	})

	t.Run("reject invitation when activation is denied", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		userRepo := mockrepo.NewMockUserRepository(ctrl)
		userRepo.EXPECT().Get(gomock.Any(), userID, repository.UserDetailProjection()).Return(user, nil)

		svc := newOrgService(ctrl, userRepo, finiteSeats{err: entitlement.ErrActivationDenied})
		err := svc.AcceptInvitation(ctx, orgID, service.AcceptOrganizationInvitationOpts{
			Token:    publicToken,
			Password: "password123",
		})
		require.ErrorIs(t, err, entitlement.ErrActivationDenied)
	})

	t.Run("reject activation when an active user becomes pending before commit", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		activeUser := *user
		activeUser.Status = model.UserStatusActive
		userRepo := mockrepo.NewMockUserRepository(ctrl)
		userRepo.EXPECT().Get(gomock.Any(), userID, repository.UserDetailProjection()).Return(&activeUser, nil)
		userRepo.EXPECT().AcceptInvitation(gomock.Any(), userID, orgID, "", nil, nil).
			Return(nil, repository.ErrUserActivationStatus)

		svc := newOrgService(ctrl, userRepo, finiteSeats{err: entitlement.ErrActivationDenied})
		err := svc.AcceptInvitation(ctx, orgID, service.AcceptOrganizationInvitationOpts{
			Token: publicToken,
		})
		require.ErrorIs(t, err, repository.ErrUserActivationStatus)
	})
}

func TestOrganizationService_AcceptInvitation_AssignsRoleAtomically(t *testing.T) {
	t.Parallel()

	userID := model.MustNewID(model.ResourceTypeUser)
	orgID := model.MustNewID(model.ResourceTypeOrganization)
	roleID := model.MustNewID(model.ResourceTypeRole)
	ctx := context.Background()
	tokenData := map[string]any{
		"organization_id": orgID.String(),
		"user_id":         userID.String(),
		"role_id":         roleID.String(),
	}
	publicToken, tokenHash, err := auth.GenerateToken(model.UserTokenContextInvite.String(), tokenData)
	require.NoError(t, err)

	user := testModel.NewUser()
	user.ID = userID
	user.Status = model.UserStatusPending
	now := time.Now()
	userToken := &repository.UserToken{
		ID:        model.MustNewID(model.ResourceTypeUserToken),
		UserID:    userID,
		SentTo:    user.Email,
		Token:     tokenHash,
		Context:   model.UserTokenContextInvite,
		CreatedAt: &now,
	}

	ctrl := gomock.NewController(t)
	span := mocktrace.NewMockSpan(ctrl)
	span.EXPECT().End(gomock.Len(0))
	tracer := mocktrace.NewMockTracer(ctrl)
	tracer.EXPECT().Start(gomock.Any(), "service.organizationService/AcceptInvitation", gomock.Len(0)).Return(ctx, span)

	userRepo := mockrepo.NewMockUserRepository(ctrl)
	userRepo.EXPECT().Get(gomock.Any(), userID, repository.UserDetailProjection()).Return(user, nil)
	userRepo.EXPECT().AcceptInvitation(gomock.Any(), userID, orgID, gomock.Any(), gomock.Cond(func(got *repository.ActivationAuthorization) bool {
		return got != nil && *got == repository.UnrestrictedActivation()
	}), gomock.Cond(func(got *model.ID) bool {
		return got != nil && *got == roleID
	})).Return(&repository.User{ID: userID, Status: model.UserStatusActive}, nil)

	userTokenRepo := mockrepo.NewMockUserTokenRepository(ctrl)
	userTokenRepo.EXPECT().Get(gomock.Any(), userID, model.UserTokenContextInvite).Return(userToken, nil)
	userTokenRepo.EXPECT().Delete(gomock.Any(), userID, model.UserTokenContextInvite).Return(nil)

	svc, err := service.NewOrganizationService(
		mockrepo.NewMockOrganizationRepository(ctrl),
		userRepo,
		userTokenRepo,
		mockrepo.NewMockRoleRepository(ctrl),
		mocksvc.NewMockPermissionService(ctrl),
		entitlement.Unrestricted(),
		mocksvc.NewMockEmailService(ctrl),
		mocksvc.NewMockNotificationService(ctrl),
		mocksvc.NewMockSearchService(ctrl),
		service.WithLogger(mocklog.NewMockLogger(ctrl)),
		service.WithTracer(tracer),
	)
	require.NoError(t, err)
	require.NoError(t, svc.AcceptInvitation(ctx, orgID, service.AcceptOrganizationInvitationOpts{
		Token:    publicToken,
		Password: "password123",
	}))
}

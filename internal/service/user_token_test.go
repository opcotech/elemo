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
	"github.com/opcotech/elemo/internal/repository"
	mockrepo "github.com/opcotech/elemo/internal/repository/mock"
	"github.com/opcotech/elemo/internal/service"
)

func TestUserService_VerifyToken_ExpiredReturnsClaims(t *testing.T) {
	t.Parallel()

	userID := model.MustNewID(model.ResourceTypeUser)
	public, hash, err := auth.GenerateToken(model.UserTokenContextResetPassword.String(), map[string]any{
		"user_id": userID.String(),
	})
	require.NoError(t, err)

	createdAt := time.Now().Add(-service.UserPasswordResetDeadline - time.Minute)
	ctrl := gomock.NewController(t)
	tokens := mockrepo.NewMockUserTokenRepository(ctrl)
	tokens.EXPECT().Get(gomock.Any(), userID, model.UserTokenContextResetPassword).Return(&repository.UserToken{
		ID:        model.MustNewID(model.ResourceTypeUserToken),
		UserID:    userID,
		Token:     hash,
		Context:   model.UserTokenContextResetPassword,
		CreatedAt: &createdAt,
	}, nil)

	svc, err := service.NewUserService(
		mockrepo.NewMockUserRepository(ctrl),
		tokens,
		entitlement.Unrestricted(),
	)
	require.NoError(t, err)

	claims, err := svc.VerifyToken(context.Background(), public)
	require.ErrorIs(t, err, service.ErrExpiredToken)
	require.Equal(t, userID.String(), claims["user_id"])
}

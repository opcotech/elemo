package service_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/opcotech/elemo/internal/entitlement"
	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg"
	mocklog "github.com/opcotech/elemo/internal/pkg/log/mock"
	"github.com/opcotech/elemo/internal/pkg/optional"
	mocktrace "github.com/opcotech/elemo/internal/pkg/tracing/mock"
	"github.com/opcotech/elemo/internal/repository"
	mockrepo "github.com/opcotech/elemo/internal/repository/mock"
	"github.com/opcotech/elemo/internal/service"
	testModel "github.com/opcotech/elemo/internal/testutil/model"
)

type finiteSeats struct {
	max uint32
	err error
}

func (f finiteSeats) ActivationLimit(_ context.Context) (entitlement.Limit, error) {
	if f.err != nil {
		return entitlement.Limit{}, f.err
	}
	return entitlement.Limit{Max: f.max}, nil
}

func TestUserService_Create_SeatLimit(t *testing.T) {
	t.Parallel()

	userID := model.MustNewID(model.ResourceTypeUser)
	ctx := context.WithValue(context.Background(), pkg.CtxKeyUserID, userID)
	opts := createUserOptsFromRepo(testModel.NewCreateUserOpts())
	opts.Status = model.UserStatusActive

	t.Run("create active user when seats remain", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		span := mocktrace.NewMockSpan(ctrl)
		span.EXPECT().End(gomock.Len(0))
		tracer := mocktrace.NewMockTracer(ctrl)
		tracer.EXPECT().Start(gomock.Any(), "service.userService/Create", gomock.Len(0)).Return(ctx, span)

		userRepo := mockrepo.NewMockUserRepository(ctrl)
		userRepo.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, got repository.CreateUserOpts) (*repository.User, error) {
			require.NotNil(t, got.Activation)
			require.Equal(t, repository.FiniteActivation(2), *got.Activation)
			return &repository.User{ID: userID, Status: model.UserStatusActive}, nil
		})

		svc, err := service.NewUserService(
			userRepo,
			mockrepo.NewMockUserTokenRepository(ctrl),
			finiteSeats{max: 2},
			service.WithLogger(mocklog.NewMockLogger(ctrl)),
			service.WithTracer(tracer),
		)
		require.NoError(t, err)
		_, err = svc.Create(ctx, opts)
		require.NoError(t, err)
	})

	t.Run("create pending user does not consume a seat", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		span := mocktrace.NewMockSpan(ctrl)
		span.EXPECT().End(gomock.Len(0))
		tracer := mocktrace.NewMockTracer(ctrl)
		tracer.EXPECT().Start(gomock.Any(), "service.userService/Create", gomock.Len(0)).Return(ctx, span)

		pending := opts
		pending.Status = model.UserStatusPending
		userRepo := mockrepo.NewMockUserRepository(ctrl)
		userRepo.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, got repository.CreateUserOpts) (*repository.User, error) {
			require.Nil(t, got.Activation)
			return &repository.User{Status: model.UserStatusPending}, nil
		})

		svc, err := service.NewUserService(
			userRepo,
			mockrepo.NewMockUserTokenRepository(ctrl),
			finiteSeats{max: 1},
			service.WithLogger(mocklog.NewMockLogger(ctrl)),
			service.WithTracer(tracer),
		)
		require.NoError(t, err)
		_, err = svc.Create(ctx, pending)
		require.NoError(t, err)
	})

	t.Run("create active user when the seat limit is reached", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		span := mocktrace.NewMockSpan(ctrl)
		span.EXPECT().End(gomock.Len(0))
		tracer := mocktrace.NewMockTracer(ctrl)
		tracer.EXPECT().Start(gomock.Any(), "service.userService/Create", gomock.Len(0)).Return(ctx, span)

		userRepo := mockrepo.NewMockUserRepository(ctrl)
		userRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil, repository.ErrSeatLimitReached)

		svc, err := service.NewUserService(
			userRepo,
			mockrepo.NewMockUserTokenRepository(ctrl),
			finiteSeats{max: 1},
			service.WithLogger(mocklog.NewMockLogger(ctrl)),
			service.WithTracer(tracer),
		)
		require.NoError(t, err)
		_, err = svc.Create(ctx, opts)
		require.ErrorIs(t, err, entitlement.ErrSeatLimitReached)
	})

	t.Run("create active user when activation is denied", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		span := mocktrace.NewMockSpan(ctrl)
		span.EXPECT().End(gomock.Len(0))
		tracer := mocktrace.NewMockTracer(ctrl)
		tracer.EXPECT().Start(gomock.Any(), "service.userService/Create", gomock.Len(0)).Return(ctx, span)

		svc, err := service.NewUserService(
			mockrepo.NewMockUserRepository(ctrl),
			mockrepo.NewMockUserTokenRepository(ctrl),
			finiteSeats{err: entitlement.ErrActivationDenied},
			service.WithLogger(mocklog.NewMockLogger(ctrl)),
			service.WithTracer(tracer),
		)
		require.NoError(t, err)
		_, err = svc.Create(ctx, opts)
		require.ErrorIs(t, err, entitlement.ErrActivationDenied)
	})
}

func TestUserService_Update_SeatActivation(t *testing.T) {
	t.Parallel()

	userID := model.MustNewID(model.ResourceTypeUser)
	ctx := context.WithValue(context.Background(), pkg.CtxKeyUserID, userID)
	opts := service.UpdateUserOpts{Status: optional.Some(model.UserStatusActive)}

	t.Run("activate pending user when seats remain", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		span := mocktrace.NewMockSpan(ctrl)
		span.EXPECT().End(gomock.Len(0))
		tracer := mocktrace.NewMockTracer(ctrl)
		tracer.EXPECT().Start(gomock.Any(), "service.userService/Update", gomock.Len(0)).Return(ctx, span)

		userRepo := mockrepo.NewMockUserRepository(ctrl)
		userRepo.EXPECT().Activate(gomock.Any(), userID, gomock.Any(), repository.FiniteActivation(2)).Return(&repository.User{ID: userID, Status: model.UserStatusActive}, nil)

		svc, err := service.NewUserService(
			userRepo,
			mockrepo.NewMockUserTokenRepository(ctrl),
			finiteSeats{max: 2},
			service.WithLogger(mocklog.NewMockLogger(ctrl)),
			service.WithTracer(tracer),
		)
		require.NoError(t, err)
		_, err = svc.Update(ctx, userID, opts)
		require.NoError(t, err)
	})

	t.Run("active to active does not consume a seat", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		span := mocktrace.NewMockSpan(ctrl)
		span.EXPECT().End(gomock.Len(0))
		tracer := mocktrace.NewMockTracer(ctrl)
		tracer.EXPECT().Start(gomock.Any(), "service.userService/Update", gomock.Len(0)).Return(ctx, span)

		userRepo := mockrepo.NewMockUserRepository(ctrl)
		userRepo.EXPECT().Activate(gomock.Any(), userID, gomock.Any(), repository.FiniteActivation(1)).Return(&repository.User{ID: userID, Status: model.UserStatusActive}, nil)

		svc, err := service.NewUserService(
			userRepo,
			mockrepo.NewMockUserTokenRepository(ctrl),
			finiteSeats{max: 1},
			service.WithLogger(mocklog.NewMockLogger(ctrl)),
			service.WithTracer(tracer),
		)
		require.NoError(t, err)
		_, err = svc.Update(ctx, userID, opts)
		require.NoError(t, err)
	})

	t.Run("invalid status is not an entitlement denial", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		span := mocktrace.NewMockSpan(ctrl)
		span.EXPECT().End(gomock.Len(0))
		tracer := mocktrace.NewMockTracer(ctrl)
		tracer.EXPECT().Start(gomock.Any(), "service.userService/Update", gomock.Len(0)).Return(ctx, span)

		userRepo := mockrepo.NewMockUserRepository(ctrl)
		userRepo.EXPECT().Activate(gomock.Any(), userID, gomock.Any(), repository.FiniteActivation(1)).Return(nil, repository.ErrUserActivationStatus)

		svc, err := service.NewUserService(
			userRepo,
			mockrepo.NewMockUserTokenRepository(ctrl),
			finiteSeats{max: 1},
			service.WithLogger(mocklog.NewMockLogger(ctrl)),
			service.WithTracer(tracer),
		)
		require.NoError(t, err)
		_, err = svc.Update(ctx, userID, opts)
		require.ErrorIs(t, err, repository.ErrUserActivationStatus)
		require.NotErrorIs(t, err, entitlement.ErrActivationDenied)
	})
}

package service

import (
	"context"
	"errors"

	"github.com/opcotech/elemo/internal/entitlement"
	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/repository"
)

func activationFromPolicy(ctx context.Context, seats entitlement.SeatPolicy) (*repository.ActivationAuthorization, error) {
	limit, err := seats.ActivationLimit(ctx)
	if err != nil {
		return nil, err
	}
	if limit.Unlimited {
		auth := repository.UnrestrictedActivation()
		return &auth, nil
	}
	auth := repository.FiniteActivation(limit.Max)
	return &auth, nil
}

func createUserWithSeats(
	ctx context.Context,
	userRepo repository.UserRepository,
	seats entitlement.SeatPolicy,
	opts repository.CreateUserOpts,
) (*repository.User, error) {
	if opts.Status == 0 || opts.Status == model.UserStatusActive {
		opts.Status = model.UserStatusActive
		auth, err := activationFromPolicy(ctx, seats)
		if err != nil {
			return nil, err
		}
		opts.Activation = auth
	}

	user, err := userRepo.Create(ctx, opts)
	if err != nil {
		if errors.Is(err, repository.ErrSeatLimitReached) {
			return nil, entitlement.ErrSeatLimitReached
		}
		return nil, err
	}
	return user, nil
}

func activateUserWithSeats(
	ctx context.Context,
	userRepo repository.UserRepository,
	seats entitlement.SeatPolicy,
	id model.ID,
	opts repository.UpdateUserOpts,
) (*repository.User, error) {
	activating := opts.Status.Defined && opts.Status.Value != nil && *opts.Status.Value == model.UserStatusActive
	if !activating {
		return userRepo.Update(ctx, id, opts)
	}

	auth, err := activationFromPolicy(ctx, seats)
	if err != nil {
		return nil, err
	}

	user, err := userRepo.Activate(ctx, id, opts, *auth)
	if err != nil {
		if errors.Is(err, repository.ErrSeatLimitReached) {
			return nil, entitlement.ErrSeatLimitReached
		}
		return nil, err
	}
	return user, nil
}

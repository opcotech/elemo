package service

import (
	"context"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg"
)

func ctxUserID(ctx context.Context) (model.ID, error) {
	userID, ok := pkg.CtxUserIDValue(ctx)
	if !ok {
		return model.ID{}, ErrNoUser
	}
	return userID, nil
}

func requireAction(ctx context.Context, perm PermissionService, resource model.ID, action model.Action) error {
	allowed, err := perm.CtxUserHas(ctx, resource, action)
	if err != nil {
		return err
	}
	if !allowed {
		return ErrNoPermission
	}
	return nil
}

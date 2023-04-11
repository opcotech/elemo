package service

import (
	"context"
	"errors"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg"
	"github.com/opcotech/elemo/internal/repository"
)

// PermissionRepository defines the interface for the permission repository.
type PermissionRepository interface {
	Create(ctx context.Context, perm *model.Permission) error
	Get(ctx context.Context, id model.ID) (*model.Permission, error)
	GetBySubject(ctx context.Context, id model.ID) ([]*model.Permission, error)
	GetByTarget(ctx context.Context, id model.ID) ([]*model.Permission, error)
	Update(ctx context.Context, id model.ID, kind model.PermissionKind) (*model.Permission, error)
	Delete(ctx context.Context, id model.ID) error
	HasPermission(ctx context.Context, subject, target model.ID, kind model.PermissionKind) (bool, error)
	HasAnyPermission(ctx context.Context, subject, target model.ID, kinds ...model.PermissionKind) (bool, error)
}

func ctxUserPermitted(ctx context.Context, repo PermissionRepository, target model.ID, permissions ...model.PermissionKind) bool {
	userID, ok := ctx.Value(pkg.CtxKeyUserID).(model.ID)
	if !ok {
		return false
	}

	hasPerm, err := repo.HasAnyPermission(ctx, userID, target, append(permissions, model.PermissionKindAll)...)
	if err != nil && !errors.Is(err, repository.ErrPermissionRead) {
		return false
	}

	return hasPerm
}

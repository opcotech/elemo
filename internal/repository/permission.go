package repository

import (
	"context"

	"github.com/opcotech/elemo/internal/model"
)

// PermissionRepository defines the interface for the permission repository.
type PermissionRepository interface {
	Create(ctx context.Context, perm *model.Permission) error
	Get(ctx context.Context, id model.ID) (*model.Permission, error)
	GetBySubject(ctx context.Context, id model.ID) ([]*model.Permission, error)
	GetByTarget(ctx context.Context, id model.ID) ([]*model.Permission, error)
	Update(ctx context.Context, id model.ID, kind model.PermissionKind) (*model.Permission, error)
	Delete(ctx context.Context, id model.ID) error
	HasPermission(ctx context.Context, subject, target model.ID, kinds ...model.PermissionKind) (bool, error)
	HasAnyRelation(ctx context.Context, source, target model.ID) (bool, error)
	HasSystemRole(ctx context.Context, source model.ID, targets ...model.SystemRole) (bool, error)
}

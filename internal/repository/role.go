package repository

import (
	"context"

	"github.com/opcotech/elemo/internal/model"
)

// RoleRepository is a repository for managing roles.
type RoleRepository interface {
	Create(ctx context.Context, createdBy, belongsTo model.ID, role *model.Role) error
	Get(ctx context.Context, id, belongsTo model.ID) (*model.Role, error)
	GetAllBelongsTo(ctx context.Context, belongsTo model.ID, offset, limit int) ([]*model.Role, error)
	Update(ctx context.Context, id, belongsTo model.ID, patch map[string]any) (*model.Role, error)
	AddMember(ctx context.Context, roleID, memberID, belongsToID model.ID) error
	RemoveMember(ctx context.Context, roleID, memberID, belongsToID model.ID) error
	Delete(ctx context.Context, id, belongsTo model.ID) error
}

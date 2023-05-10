package repository

import (
	"context"

	"github.com/opcotech/elemo/internal/model"
)

// OrganizationRepository is a repository for managing organizations.
type OrganizationRepository interface {
	Create(ctx context.Context, owner model.ID, organization *model.Organization) error
	Get(ctx context.Context, id model.ID) (*model.Organization, error)
	GetAll(ctx context.Context, offset, limit int) ([]*model.Organization, error)
	Update(ctx context.Context, id model.ID, patch map[string]any) (*model.Organization, error)
	AddMember(ctx context.Context, orgID, memberID model.ID) error
	RemoveMember(ctx context.Context, orgID, memberID model.ID) error
	Delete(ctx context.Context, id model.ID) error
}

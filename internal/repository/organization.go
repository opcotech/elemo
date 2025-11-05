package repository

import (
	"context"

	"github.com/opcotech/elemo/internal/model"
)

// OrganizationRepository is a repository for managing organizations.
//
//go:generate mockgen -source=organization.go -destination=../testutil/mock/organization_repo_gen.go -package=mock -mock_names "OrganizationRepository=OrganizationRepository"
type OrganizationRepository interface {
	Create(ctx context.Context, owner model.ID, organization *model.Organization) error
	Get(ctx context.Context, id model.ID) (*model.Organization, error)
	GetAll(ctx context.Context, userID model.ID, offset, limit int) ([]*model.Organization, error)
	Update(ctx context.Context, id model.ID, patch map[string]any) (*model.Organization, error)
	GetMembers(ctx context.Context, orgID model.ID) ([]*model.OrganizationMember, error)
	AddMember(ctx context.Context, orgID, memberID model.ID) error
	RemoveMember(ctx context.Context, orgID, memberID model.ID) error
	AddInvitation(ctx context.Context, orgID, userID model.ID) error
	RemoveInvitation(ctx context.Context, orgID, userID model.ID) error
	GetInvitations(ctx context.Context, orgID model.ID) ([]*model.OrganizationMember, error)
	Delete(ctx context.Context, id model.ID) error
}

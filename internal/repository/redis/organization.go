package redis

import (
	"context"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/repository"
)

func clearOrganizationsPattern(ctx context.Context, r *baseRepository, pattern ...string) error {
	return r.DeletePattern(ctx, composeCacheKey(model.ResourceTypeOrganization.String(), pattern))
}

func clearOrganizationsKey(ctx context.Context, r *baseRepository, id model.ID) error {
	return r.Delete(ctx, composeCacheKey(model.ResourceTypeOrganization.String(), id.String()))
}

func clearOrganizationAllGetAll(ctx context.Context, r *baseRepository) error {
	return clearOrganizationsPattern(ctx, r, "GetAll", "*")
}

// CachedOrganizationRepository implements caching on the
// repository.OrganizationRepository.
type CachedOrganizationRepository struct {
	cacheRepo        *baseRepository
	organizationRepo repository.OrganizationRepository
}

func (r *CachedOrganizationRepository) Create(ctx context.Context, owner model.ID, organization *model.Organization) error {
	if err := clearOrganizationAllGetAll(ctx, r.cacheRepo); err != nil {
		return err
	}

	return r.organizationRepo.Create(ctx, owner, organization)
}

func (r *CachedOrganizationRepository) Get(ctx context.Context, id model.ID) (*model.Organization, error) {
	var organization *model.Organization
	var err error

	key := composeCacheKey(model.ResourceTypeOrganization.String(), id.String())
	if err = r.cacheRepo.Get(ctx, key, &organization); err != nil {
		return nil, err
	}

	if organization != nil {
		return organization, nil
	}

	if organization, err = r.organizationRepo.Get(ctx, id); err != nil {
		return nil, err
	}

	if err = r.cacheRepo.Set(ctx, key, organization); err != nil {
		return nil, err
	}

	return organization, nil
}

func (r *CachedOrganizationRepository) GetAll(ctx context.Context, offset, limit int) ([]*model.Organization, error) {
	var organizations []*model.Organization
	var err error

	key := composeCacheKey(model.ResourceTypeAssignment.String(), "GetAll", offset, limit)
	if err = r.cacheRepo.Get(ctx, key, &organizations); err != nil {
		return nil, err
	}

	if organizations != nil {
		return organizations, nil
	}

	if organizations, err = r.organizationRepo.GetAll(ctx, offset, limit); err != nil {
		return nil, err
	}

	if err = r.cacheRepo.Set(ctx, key, organizations); err != nil {
		return nil, err
	}

	return organizations, nil
}

func (r *CachedOrganizationRepository) Update(ctx context.Context, id model.ID, patch map[string]any) (*model.Organization, error) {
	var organization *model.Organization
	var err error

	organization, err = r.organizationRepo.Update(ctx, id, patch)
	if err != nil {
		return nil, err
	}

	key := composeCacheKey(model.ResourceTypeOrganization.String(), id.String())
	if err = r.cacheRepo.Set(ctx, key, organization); err != nil {
		return nil, err
	}

	if err := clearOrganizationAllGetAll(ctx, r.cacheRepo); err != nil {
		return nil, err
	}

	return organization, nil
}

func (r *CachedOrganizationRepository) AddMember(ctx context.Context, orgID, memberID model.ID) error {
	if err := clearOrganizationsKey(ctx, r.cacheRepo, orgID); err != nil {
		return err
	}

	if err := clearOrganizationAllGetAll(ctx, r.cacheRepo); err != nil {
		return err
	}

	return r.organizationRepo.AddMember(ctx, orgID, memberID)
}

func (r *CachedOrganizationRepository) RemoveMember(ctx context.Context, orgID, memberID model.ID) error {
	if err := clearOrganizationsKey(ctx, r.cacheRepo, orgID); err != nil {
		return err
	}

	if err := clearOrganizationAllGetAll(ctx, r.cacheRepo); err != nil {
		return err
	}

	return r.organizationRepo.RemoveMember(ctx, orgID, memberID)
}

func (r *CachedOrganizationRepository) Delete(ctx context.Context, id model.ID) error {
	if err := clearOrganizationsKey(ctx, r.cacheRepo, id); err != nil {
		return err
	}

	if err := clearOrganizationAllGetAll(ctx, r.cacheRepo); err != nil {
		return err
	}

	return r.organizationRepo.Delete(ctx, id)
}

// NewCachedOrganizationRepository returns a new CachedOrganizationRepository.
func NewCachedOrganizationRepository(repo repository.OrganizationRepository, opts ...RepositoryOption) (*CachedOrganizationRepository, error) {
	r, err := newBaseRepository(opts...)
	if err != nil {
		return nil, err
	}

	return &CachedOrganizationRepository{
		cacheRepo:        r,
		organizationRepo: repo,
	}, nil
}

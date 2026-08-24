package service

import (
	"context"
	"errors"
	"time"

	"github.com/opcotech/elemo/internal/license"
	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg/log"
	"github.com/opcotech/elemo/internal/pkg/optional"
	"github.com/opcotech/elemo/internal/pkg/validate"
	"github.com/opcotech/elemo/internal/repository"
)

// PartialNamespace is a lean namespace used on issue reads.
type PartialNamespace struct {
	ID   model.ID
	Name string
}

// PartialOrganization is a lean organization used on accessible namespace lists.
type PartialOrganization struct {
	ID   model.ID
	Name string
}

// Namespace represents a namespace returned by the service.
type Namespace struct {
	ID            model.ID
	Name          string
	Description   string
	ProjectCount  *int64
	DocumentCount *int64
	CreatedAt     *time.Time
	UpdatedAt     *time.Time
}

// AccessibleNamespace is a reachable namespace with its owning organization.
type AccessibleNamespace struct {
	Namespace
	Organization PartialOrganization
}

// CreateNamespaceOpts holds the data required to create a namespace.
type CreateNamespaceOpts struct {
	Name        string `json:"name" validate:"required,min=3,max=120"`
	Description string `json:"description" validate:"omitempty,min=5,max=500"`
}

// Validate validates the create options.
func (o *CreateNamespaceOpts) Validate() error {
	if err := validate.Struct(o); err != nil {
		return errors.Join(model.ErrInvalidNamespaceDetails, err)
	}
	return nil
}

// UpdateNamespaceOpts holds the fields that can be updated on a namespace.
// Undefined fields (Defined == false) are left unchanged.
type UpdateNamespaceOpts struct {
	Name        optional.Optional[string]
	Description optional.Optional[string]
}

// NamespaceService serves the business logic of interacting with namespaces.
//
//go:generate go tool mockgen -destination=mock/mock_namespace_gen.go -package=mocksvc . NamespaceService
type NamespaceService interface {
	// Create creates a new namespace in an organization. If the organization
	// does not exist, an error is returned.
	Create(ctx context.Context, orgID model.ID, opts CreateNamespaceOpts) (*Namespace, error)
	// Get returns a namespace by its ID. If the namespace does not exist, an
	// error is returned.
	Get(ctx context.Context, id model.ID) (*Namespace, error)
	// List returns a cursor-paginated page of namespaces for an organization.
	List(ctx context.Context, orgID model.ID, page CursorPage) (Page[*Namespace], error)
	// ListAccessible returns namespaces the actor can reach, each with an
	// organization stub. Reachability does not require organization.read.
	ListAccessible(ctx context.Context, page CursorPage) (Page[*AccessibleNamespace], error)
	// Update updates a namespace. If the namespace does not exist, an error
	// is returned.
	Update(ctx context.Context, id model.ID, opts UpdateNamespaceOpts) (*Namespace, error)
	// Delete deletes a namespace. If the namespace does not exist, an error
	// is returned.
	Delete(ctx context.Context, id model.ID) error
}

// namespaceService is the concrete implementation of NamespaceService.
type namespaceService struct {
	runtime
	namespaceRepo     repository.NamespaceRepository
	permissionService PermissionService
	licenseService    LicenseService
	searchService     SearchService
}

func partialNamespaceFromRepository(n *repository.PartialNamespace) *PartialNamespace {
	if n == nil {
		return nil
	}
	return &PartialNamespace{
		ID:   n.ID,
		Name: n.Name,
	}
}

func namespaceFromRepository(n *repository.Namespace) *Namespace {
	if n == nil {
		return nil
	}

	return &Namespace{
		ID:            n.ID,
		Name:          n.Name,
		Description:   n.Description,
		ProjectCount:  n.ProjectCount,
		DocumentCount: n.DocumentCount,
		CreatedAt:     n.CreatedAt,
		UpdatedAt:     n.UpdatedAt,
	}
}

func namespacesFromRepository(namespaces []*repository.Namespace) []*Namespace {
	out := make([]*Namespace, len(namespaces))
	for i, n := range namespaces {
		out[i] = namespaceFromRepository(n)
	}
	return out
}

func accessibleNamespaceFromRepository(n *repository.AccessibleNamespace) *AccessibleNamespace {
	if n == nil {
		return nil
	}

	ns := namespaceFromRepository(&n.Namespace)
	if ns == nil {
		return nil
	}

	return &AccessibleNamespace{
		Namespace: *ns,
		Organization: PartialOrganization{
			ID:   n.Organization.ID,
			Name: n.Organization.Name,
		},
	}
}

func (s *namespaceService) Create(ctx context.Context, orgID model.ID, opts CreateNamespaceOpts) (*Namespace, error) {
	ctx, span := s.tracer.Start(ctx, "service.namespaceService/Create")
	defer span.End()

	if expired, err := s.licenseService.Expired(ctx); expired || err != nil {
		return nil, errors.Join(ErrNamespaceCreate, license.ErrLicenseExpired)
	}

	if err := orgID.Validate(); err != nil {
		return nil, errors.Join(ErrNamespaceCreate, err)
	}

	if err := opts.Validate(); err != nil {
		return nil, errors.Join(ErrNamespaceCreate, err)
	}

	if err := requireAction(ctx, s.permissionService, orgID, model.ActionNamespaceCreate); err != nil {
		return nil, errors.Join(ErrNamespaceCreate, err)
	}

	userID, err := ctxUserID(ctx)
	if err != nil {
		return nil, errors.Join(ErrNamespaceCreate, err)
	}

	namespace, err := s.namespaceRepo.Create(ctx, repository.CreateNamespaceOpts{
		Name:        opts.Name,
		Description: opts.Description,
		CreatorID:   userID,
		OrgID:       orgID,
	})
	if err != nil {
		return nil, errors.Join(ErrNamespaceCreate, err)
	}

	actions, err := roleTemplateActions(model.RoleKeyNamespaceAdmin)
	if err != nil {
		return nil, errors.Join(ErrNamespaceCreate, err)
	}
	if err := s.permissionService.BootstrapCreator(ctx, userID, namespace.ID, actions); err != nil {
		return nil, errors.Join(ErrNamespaceCreate, err)
	}

	out := namespaceFromRepository(namespace)
	enqueueSearchIndex(ctx, s.logger, s.searchService, out.ID)
	return out, nil
}

func (s *namespaceService) Get(ctx context.Context, id model.ID) (*Namespace, error) {
	ctx, span := s.tracer.Start(ctx, "service.namespaceService/Get")
	defer span.End()

	if err := id.Validate(); err != nil {
		return nil, errors.Join(ErrNamespaceGet, err)
	}

	if err := requireAction(ctx, s.permissionService, id, model.ActionNamespaceRead); err != nil {
		return nil, errors.Join(ErrNamespaceGet, err)
	}

	namespace, err := s.namespaceRepo.Get(ctx, id, repository.NamespaceDetailProjection())
	if err != nil {
		return nil, errors.Join(ErrNamespaceGet, err)
	}

	return namespaceFromRepository(namespace), nil
}

func (s *namespaceService) List(ctx context.Context, orgID model.ID, page CursorPage) (Page[*Namespace], error) {
	ctx, span := s.tracer.Start(ctx, "service.namespaceService/List")
	defer span.End()

	if err := orgID.Validate(); err != nil {
		return Page[*Namespace]{}, errors.Join(ErrNamespaceList, err)
	}

	normalized, err := page.Normalize()
	if err != nil {
		return Page[*Namespace]{}, errors.Join(ErrNamespaceList, err)
	}

	userID, err := ctxUserID(ctx)
	if err != nil {
		return Page[*Namespace]{}, errors.Join(ErrNamespaceList, err)
	}

	namespaces, err := s.namespaceRepo.ListForOrganization(ctx, repository.NamespaceListQuery{
		OrgID:      orgID,
		ActorID:    userID,
		Page:       normalized,
		Order:      repository.SortDirectionDesc,
		Projection: repository.NamespaceListProjection(),
	})
	if err != nil {
		return Page[*Namespace]{}, errors.Join(ErrNamespaceList, err)
	}

	return mapPage(namespaces, namespaceFromRepository), nil
}

func (s *namespaceService) ListAccessible(ctx context.Context, page CursorPage) (Page[*AccessibleNamespace], error) {
	ctx, span := s.tracer.Start(ctx, "service.namespaceService/ListAccessible")
	defer span.End()

	normalized, err := page.Normalize()
	if err != nil {
		return Page[*AccessibleNamespace]{}, errors.Join(ErrNamespaceList, err)
	}

	userID, err := ctxUserID(ctx)
	if err != nil {
		return Page[*AccessibleNamespace]{}, errors.Join(ErrNamespaceList, err)
	}

	namespaces, err := s.namespaceRepo.ListAccessible(ctx, repository.NamespaceListAccessibleQuery{
		ActorID:    userID,
		Page:       normalized,
		Order:      repository.SortDirectionDesc,
		Projection: repository.NamespaceListProjection(),
	})
	if err != nil {
		return Page[*AccessibleNamespace]{}, errors.Join(ErrNamespaceList, err)
	}

	return mapPage(namespaces, accessibleNamespaceFromRepository), nil
}

func (s *namespaceService) Update(ctx context.Context, id model.ID, opts UpdateNamespaceOpts) (*Namespace, error) {
	ctx, span := s.tracer.Start(ctx, "service.namespaceService/Update")
	defer span.End()

	if expired, err := s.licenseService.Expired(ctx); expired || err != nil {
		return nil, errors.Join(ErrNamespaceUpdate, license.ErrLicenseExpired)
	}

	if err := id.Validate(); err != nil {
		return nil, errors.Join(ErrNamespaceUpdate, err)
	}

	if err := requireAction(ctx, s.permissionService, id, model.ActionNamespaceUpdate); err != nil {
		return nil, errors.Join(ErrNamespaceUpdate, err)
	}

	namespace, err := s.namespaceRepo.Update(ctx, id, repository.UpdateNamespaceOpts{
		Name:        opts.Name,
		Description: opts.Description,
	})
	if err != nil {
		return nil, errors.Join(ErrNamespaceUpdate, err)
	}

	out := namespaceFromRepository(namespace)
	enqueueSearchIndex(ctx, s.logger, s.searchService, out.ID)
	return out, nil
}

func (s *namespaceService) Delete(ctx context.Context, id model.ID) error {
	ctx, span := s.tracer.Start(ctx, "service.namespaceService/Delete")
	defer span.End()

	if expired, err := s.licenseService.Expired(ctx); expired || err != nil {
		return errors.Join(ErrNamespaceDelete, license.ErrLicenseExpired)
	}

	if err := id.Validate(); err != nil {
		return errors.Join(ErrNamespaceDelete, err)
	}

	if err := requireAction(ctx, s.permissionService, id, model.ActionNamespaceDelete); err != nil {
		return errors.Join(ErrNamespaceDelete, err)
	}

	if err := s.namespaceRepo.Delete(ctx, id); err != nil {
		return errors.Join(ErrNamespaceDelete, err)
	}

	if err := s.searchService.DeleteByScope(ctx, id); err != nil {
		s.logger.Warn(ctx, "failed to delete search documents by scope",
			log.WithError(err),
			log.WithValue(id.Composite()),
		)
	}
	return nil
}

// NewNamespaceService returns a new instance of the NamespaceService interface.
func NewNamespaceService(
	namespaceRepo repository.NamespaceRepository,
	permissionService PermissionService,
	licenseService LicenseService,
	searchService SearchService,
	opts ...Option,
) (NamespaceService, error) {
	rt, err := newRuntime(opts...)
	if err != nil {
		return nil, err
	}

	svc := &namespaceService{
		runtime:           rt,
		namespaceRepo:     namespaceRepo,
		permissionService: permissionService,
		licenseService:    licenseService,
		searchService:     searchService,
	}

	if svc.namespaceRepo == nil {
		return nil, ErrNoNamespaceRepository
	}

	if svc.permissionService == nil {
		return nil, ErrNoPermissionService
	}

	if svc.licenseService == nil {
		return nil, ErrNoLicenseService
	}

	if svc.searchService == nil {
		return nil, ErrNoSearchService
	}

	return svc, nil
}

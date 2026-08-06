package service

import (
	"context"
	"errors"
	"time"

	"github.com/opcotech/elemo/internal/license"
	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg"
	"github.com/opcotech/elemo/internal/pkg/optional"
	"github.com/opcotech/elemo/internal/pkg/validate"
	"github.com/opcotech/elemo/internal/repository"
)

// NamespaceProject represents a simplified project within a namespace.
type NamespaceProject struct {
	ID          model.ID
	Key         string
	Name        string
	Description string
	Logo        string
	Status      model.ProjectStatus
}

// NamespaceDocument represents a simplified document within a namespace.
type NamespaceDocument struct {
	ID        model.ID
	Name      string
	Excerpt   string
	CreatedBy model.ID
	CreatedAt *time.Time
}

// Namespace represents a namespace returned by the service.
type Namespace struct {
	ID          model.ID
	Name        string
	Description string
	Projects    []*NamespaceProject
	Documents   []*NamespaceDocument
	CreatedAt   *time.Time
	UpdatedAt   *time.Time
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
type NamespaceService interface {
	// Create creates a new namespace in an organization. If the organization
	// does not exist, an error is returned.
	Create(ctx context.Context, orgID model.ID, opts CreateNamespaceOpts) (*Namespace, error)
	// Get returns a namespace by its ID. If the namespace does not exist, an
	// error is returned.
	Get(ctx context.Context, id model.ID) (*Namespace, error)
	// GetAll returns all namespaces for an organization. The offset and limit
	// parameters are used to paginate the results. If the offset is greater
	// than the number of namespaces in the organization, an empty slice is
	// returned.
	GetAll(ctx context.Context, orgID model.ID, offset, limit int) ([]*Namespace, error)
	// Update updates a namespace. If the namespace does not exist, an error
	// is returned.
	Update(ctx context.Context, id model.ID, opts UpdateNamespaceOpts) (*Namespace, error)
	// Delete deletes a namespace. If the namespace does not exist, an error
	// is returned.
	Delete(ctx context.Context, id model.ID) error
}

// namespaceService is the concrete implementation of NamespaceService.
type namespaceService struct {
	*baseService
}

func namespaceProjectFromRepository(p *repository.NamespaceProject) *NamespaceProject {
	if p == nil {
		return nil
	}
	return &NamespaceProject{
		ID:          p.ID,
		Key:         p.Key,
		Name:        p.Name,
		Description: p.Description,
		Logo:        p.Logo,
		Status:      p.Status,
	}
}

func namespaceDocumentFromRepository(d *repository.NamespaceDocument) *NamespaceDocument {
	if d == nil {
		return nil
	}
	return &NamespaceDocument{
		ID:        d.ID,
		Name:      d.Name,
		Excerpt:   d.Excerpt,
		CreatedBy: d.CreatedBy,
		CreatedAt: d.CreatedAt,
	}
}

func namespaceFromRepository(n *repository.Namespace) *Namespace {
	if n == nil {
		return nil
	}

	projects := make([]*NamespaceProject, len(n.Projects))
	for i, p := range n.Projects {
		projects[i] = namespaceProjectFromRepository(p)
	}

	documents := make([]*NamespaceDocument, len(n.Documents))
	for i, d := range n.Documents {
		documents[i] = namespaceDocumentFromRepository(d)
	}

	return &Namespace{
		ID:          n.ID,
		Name:        n.Name,
		Description: n.Description,
		Projects:    projects,
		Documents:   documents,
		CreatedAt:   n.CreatedAt,
		UpdatedAt:   n.UpdatedAt,
	}
}

func namespacesFromRepository(namespaces []*repository.Namespace) []*Namespace {
	out := make([]*Namespace, len(namespaces))
	for i, n := range namespaces {
		out[i] = namespaceFromRepository(n)
	}
	return out
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

	if !s.permissionService.CtxUserHasPermission(ctx, orgID, model.PermissionKindWrite) {
		return nil, errors.Join(ErrNamespaceCreate, ErrNoPermission)
	}

	userID, ok := ctx.Value(pkg.CtxKeyUserID).(model.ID)
	if !ok {
		return nil, errors.Join(ErrNamespaceCreate, model.ErrInvalidID)
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

	return namespaceFromRepository(namespace), nil
}

func (s *namespaceService) Get(ctx context.Context, id model.ID) (*Namespace, error) {
	ctx, span := s.tracer.Start(ctx, "service.namespaceService/Get")
	defer span.End()

	if err := id.Validate(); err != nil {
		return nil, errors.Join(ErrNamespaceGet, err)
	}

	if !s.permissionService.CtxUserHasPermission(ctx, id, model.PermissionKindRead) {
		return nil, errors.Join(ErrNamespaceGet, ErrNoPermission)
	}

	namespace, err := s.namespaceRepo.Get(ctx, id)
	if err != nil {
		return nil, errors.Join(ErrNamespaceGet, err)
	}

	return namespaceFromRepository(namespace), nil
}

func (s *namespaceService) GetAll(ctx context.Context, orgID model.ID, offset, limit int) ([]*Namespace, error) {
	ctx, span := s.tracer.Start(ctx, "service.namespaceService/GetAll")
	defer span.End()

	if err := orgID.Validate(); err != nil {
		return nil, errors.Join(ErrNamespaceGetAll, err)
	}

	if offset < 0 || limit <= 0 {
		return nil, errors.Join(ErrNamespaceGetAll, ErrInvalidPaginationParams)
	}

	if !s.permissionService.CtxUserHasPermission(ctx, orgID, model.PermissionKindRead) {
		return nil, errors.Join(ErrNamespaceGetAll, ErrNoPermission)
	}

	namespaces, err := s.namespaceRepo.GetAll(ctx, orgID, offset, limit)
	if err != nil {
		return nil, errors.Join(ErrNamespaceGetAll, err)
	}

	return namespacesFromRepository(namespaces), nil
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

	if !s.permissionService.CtxUserHasPermission(ctx, id, model.PermissionKindWrite) {
		return nil, errors.Join(ErrNamespaceUpdate, ErrNoPermission)
	}

	namespace, err := s.namespaceRepo.Update(ctx, id, repository.UpdateNamespaceOpts{
		Name:        opts.Name,
		Description: opts.Description,
	})
	if err != nil {
		return nil, errors.Join(ErrNamespaceUpdate, err)
	}

	return namespaceFromRepository(namespace), nil
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

	if !s.permissionService.CtxUserHasPermission(ctx, id, model.PermissionKindDelete) {
		return errors.Join(ErrNamespaceDelete, ErrNoPermission)
	}

	if err := s.namespaceRepo.Delete(ctx, id); err != nil {
		return errors.Join(ErrNamespaceDelete, err)
	}

	return nil
}

// NewNamespaceService returns a new instance of the NamespaceService interface.
func NewNamespaceService(opts ...Option) (NamespaceService, error) {
	s, err := newService(opts...)
	if err != nil {
		return nil, err
	}

	svc := &namespaceService{
		baseService: s,
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

	return svc, nil
}

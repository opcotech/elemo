package service

import (
	"context"
	"errors"
	"time"

	"github.com/opcotech/elemo/internal/license"
	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg/optional"
	"github.com/opcotech/elemo/internal/pkg/validate"
	"github.com/opcotech/elemo/internal/repository"
)

// PartialProject represents a simplified project within a namespace.
type PartialProject struct {
	ID          model.ID
	Key         string
	Name        string
	Description string
	Logo        string
	Status      model.ProjectStatus
}

// Project represents a project returned by the service.
type Project struct {
	ID          model.ID
	Key         string
	Name        string
	Description string
	Logo        string
	Status      model.ProjectStatus
	Teams       []model.ID
	Documents   []*PartialDocument
	Issues      []model.ID
	CreatedAt   *time.Time
	UpdatedAt   *time.Time
}

// CreateProjectOpts holds the data required to create a project.
type CreateProjectOpts struct {
	Key         string              `json:"key" validate:"required,alpha,min=3,max=6"`
	Name        string              `json:"name" validate:"required,min=3,max=120"`
	Description string              `json:"description" validate:"omitempty,min=10,max=500"`
	Logo        string              `json:"logo" validate:"omitempty,url"`
	Status      model.ProjectStatus `json:"status" validate:"omitempty,min=1,max=2"`
}

// Validate validates the create options.
func (o *CreateProjectOpts) Validate() error {
	if err := validate.Struct(o); err != nil {
		return errors.Join(model.ErrInvalidProjectDetails, err)
	}
	return nil
}

// UpdateProjectOpts holds the fields that can be updated on a project.
// Undefined fields (Defined == false) are left unchanged.
type UpdateProjectOpts struct {
	Key         optional.Optional[string]
	Name        optional.Optional[string]
	Description optional.Optional[string]
	Logo        optional.Optional[string]
	Status      optional.Optional[model.ProjectStatus]
}

// ProjectService serves the business logic of interacting with projects.
type ProjectService interface {
	// Create creates a new project in a namespace. If the namespace does not
	// exist, an error is returned.
	Create(ctx context.Context, namespaceID model.ID, opts CreateProjectOpts) (*Project, error)
	// Get returns a project by its ID. If the project does not exist, an error is
	// returned.
	Get(ctx context.Context, id model.ID) (*Project, error)
	// GetByKey returns a project by its key. If the project does not exist, an
	// error is returned.
	GetByKey(ctx context.Context, key string) (*Project, error)
	// GetAll returns all projects for a namespace. The offset and limit
	// parameters are used to paginate the results. If the offset is greater
	// than the number of projects in the namespace, an empty slice is returned.
	GetAll(ctx context.Context, namespaceID model.ID, offset, limit int) ([]*Project, error)
	// Update updates a project. If the project does not exist, an error is
	// returned.
	Update(ctx context.Context, id model.ID, opts UpdateProjectOpts) (*Project, error)
	// Delete deletes a project. If the project does not exist, an error is
	// returned.
	Delete(ctx context.Context, id model.ID) error
}

// projectService is the concrete implementation of ProjectService.
type projectService struct {
	*baseService
}

func partialProjectFromRepository(p *repository.PartialProject) *PartialProject {
	if p == nil {
		return nil
	}

	return &PartialProject{
		ID:          p.ID,
		Key:         p.Key,
		Name:        p.Name,
		Description: p.Description,
		Logo:        p.Logo,
		Status:      p.Status,
	}
}

func projectFromRepository(p *repository.Project) *Project {
	if p == nil {
		return nil
	}

	documents := make([]*PartialDocument, len(p.Documents))
	for i, d := range p.Documents {
		documents[i] = partialDocumentFromRepository(d)
	}

	return &Project{
		ID:          p.ID,
		Key:         p.Key,
		Name:        p.Name,
		Description: p.Description,
		Logo:        p.Logo,
		Status:      p.Status,
		Teams:       p.Teams,
		Documents:   documents,
		Issues:      p.Issues,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}

func projectsFromRepository(projects []*repository.Project) []*Project {
	out := make([]*Project, len(projects))
	for i, p := range projects {
		out[i] = projectFromRepository(p)
	}
	return out
}

func (s *projectService) Create(ctx context.Context, namespaceID model.ID, opts CreateProjectOpts) (*Project, error) {
	ctx, span := s.tracer.Start(ctx, "service.projectService/Create")
	defer span.End()

	if expired, err := s.licenseService.Expired(ctx); expired || err != nil {
		return nil, errors.Join(ErrProjectCreate, license.ErrLicenseExpired)
	}

	if err := namespaceID.Validate(); err != nil {
		return nil, errors.Join(ErrProjectCreate, err)
	}

	if err := opts.Validate(); err != nil {
		return nil, errors.Join(ErrProjectCreate, err)
	}

	if !s.permissionService.CtxUserHasPermission(ctx, namespaceID, model.PermissionKindWrite) {
		return nil, errors.Join(ErrProjectCreate, ErrNoPermission)
	}

	status := opts.Status
	if status == 0 {
		status = model.ProjectStatusActive
	}

	project, err := s.projectRepo.Create(ctx, repository.CreateProjectOpts{
		NamespaceID: namespaceID,
		Key:         opts.Key,
		Name:        opts.Name,
		Description: opts.Description,
		Logo:        opts.Logo,
		Status:      status,
	})
	if err != nil {
		return nil, errors.Join(ErrProjectCreate, err)
	}

	return projectFromRepository(project), nil
}

func (s *projectService) Get(ctx context.Context, id model.ID) (*Project, error) {
	ctx, span := s.tracer.Start(ctx, "service.projectService/Get")
	defer span.End()

	if err := id.Validate(); err != nil {
		return nil, errors.Join(ErrProjectGet, err)
	}

	if !s.permissionService.CtxUserHasPermission(ctx, id, model.PermissionKindRead) {
		return nil, errors.Join(ErrProjectGet, ErrNoPermission)
	}

	project, err := s.projectRepo.Get(ctx, id)
	if err != nil {
		return nil, errors.Join(ErrProjectGet, err)
	}

	return projectFromRepository(project), nil
}

func (s *projectService) GetByKey(ctx context.Context, key string) (*Project, error) {
	ctx, span := s.tracer.Start(ctx, "service.projectService/GetByKey")
	defer span.End()

	if key == "" {
		return nil, errors.Join(ErrProjectGet, model.ErrInvalidProjectDetails)
	}

	project, err := s.projectRepo.GetByKey(ctx, key)
	if err != nil {
		return nil, errors.Join(ErrProjectGet, err)
	}

	if !s.permissionService.CtxUserHasPermission(ctx, project.ID, model.PermissionKindRead) {
		return nil, errors.Join(ErrProjectGet, ErrNoPermission)
	}

	return projectFromRepository(project), nil
}

func (s *projectService) GetAll(ctx context.Context, namespaceID model.ID, offset, limit int) ([]*Project, error) {
	ctx, span := s.tracer.Start(ctx, "service.projectService/GetAll")
	defer span.End()

	if err := namespaceID.Validate(); err != nil {
		return nil, errors.Join(ErrProjectGetAll, err)
	}

	if offset < 0 || limit <= 0 {
		return nil, errors.Join(ErrProjectGetAll, ErrInvalidPaginationParams)
	}

	if !s.permissionService.CtxUserHasPermission(ctx, namespaceID, model.PermissionKindRead) {
		return nil, errors.Join(ErrProjectGetAll, ErrNoPermission)
	}

	projects, err := s.projectRepo.GetAll(ctx, namespaceID, offset, limit)
	if err != nil {
		return nil, errors.Join(ErrProjectGetAll, err)
	}

	return projectsFromRepository(projects), nil
}

func (s *projectService) Update(ctx context.Context, id model.ID, opts UpdateProjectOpts) (*Project, error) {
	ctx, span := s.tracer.Start(ctx, "service.projectService/Update")
	defer span.End()

	if expired, err := s.licenseService.Expired(ctx); expired || err != nil {
		return nil, errors.Join(ErrProjectUpdate, license.ErrLicenseExpired)
	}

	if err := id.Validate(); err != nil {
		return nil, errors.Join(ErrProjectUpdate, err)
	}

	if !s.permissionService.CtxUserHasPermission(ctx, id, model.PermissionKindWrite) {
		return nil, errors.Join(ErrProjectUpdate, ErrNoPermission)
	}

	project, err := s.projectRepo.Update(ctx, id, repository.UpdateProjectOpts{
		Key:         opts.Key,
		Name:        opts.Name,
		Description: opts.Description,
		Logo:        opts.Logo,
		Status:      opts.Status,
	})
	if err != nil {
		return nil, errors.Join(ErrProjectUpdate, err)
	}

	return projectFromRepository(project), nil
}

func (s *projectService) Delete(ctx context.Context, id model.ID) error {
	ctx, span := s.tracer.Start(ctx, "service.projectService/Delete")
	defer span.End()

	if expired, err := s.licenseService.Expired(ctx); expired || err != nil {
		return errors.Join(ErrProjectDelete, license.ErrLicenseExpired)
	}

	if err := id.Validate(); err != nil {
		return errors.Join(ErrProjectDelete, err)
	}

	if !s.permissionService.CtxUserHasPermission(ctx, id, model.PermissionKindDelete) {
		return errors.Join(ErrProjectDelete, ErrNoPermission)
	}

	if err := s.projectRepo.Delete(ctx, id); err != nil {
		return errors.Join(ErrProjectDelete, err)
	}

	return nil
}

// NewProjectService returns a new instance of the ProjectService interface.
func NewProjectService(opts ...Option) (ProjectService, error) {
	s, err := newService(opts...)
	if err != nil {
		return nil, err
	}

	svc := &projectService{
		baseService: s,
	}

	if svc.projectRepo == nil {
		return nil, ErrNoProjectRepository
	}

	if svc.permissionService == nil {
		return nil, ErrNoPermissionService
	}

	if svc.licenseService == nil {
		return nil, ErrNoLicenseService
	}

	return svc, nil
}

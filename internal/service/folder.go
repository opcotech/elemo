package service

import (
	"context"
	"errors"
	"time"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg"
	"github.com/opcotech/elemo/internal/pkg/optional"
	"github.com/opcotech/elemo/internal/pkg/validate"
	"github.com/opcotech/elemo/internal/repository"
)

// DocumentLibrary is the organization or namespace a document or folder is scoped to.
type DocumentLibrary struct {
	ID   model.ID
	Type model.ResourceType
	Name string
}

// DocumentFolder is a lean folder location on a document or nested folder.
type DocumentFolder struct {
	ID       model.ID
	Name     string
	ParentID *model.ID
}

// Folder represents a folder returned by the service.
type Folder struct {
	ID        model.ID
	Name      string
	Library   DocumentLibrary
	Parent    *DocumentFolder
	CreatedBy PartialUser
	CreatedAt *time.Time
	UpdatedAt *time.Time
}

// CreateFolderOpts holds the data required to create a folder.
type CreateFolderOpts struct {
	Name     string `json:"name" validate:"required,min=1,max=120"`
	ParentID *model.ID
}

// Validate validates the create options.
func (o *CreateFolderOpts) Validate() error {
	if err := validate.Struct(o); err != nil {
		return errors.Join(model.ErrInvalidFolderDetails, err)
	}
	if o.ParentID != nil {
		if err := o.ParentID.Validate(); err != nil {
			return errors.Join(model.ErrInvalidFolderDetails, err)
		}
		if o.ParentID.Type != model.ResourceTypeFolder {
			return errors.Join(model.ErrInvalidFolderDetails, model.ErrInvalidID)
		}
	}
	return nil
}

// UpdateFolderOpts holds the fields that can be updated on a folder.
type UpdateFolderOpts struct {
	Name     optional.Optional[string]
	ParentID optional.Optional[model.ID]
}

// FolderService serves the business logic of interacting with folders.
//
//go:generate go tool mockgen -destination=folder_mock_gen.go -package=service -mock_names FolderService=MockFolderService . FolderService
type FolderService interface {
	Create(ctx context.Context, libraryID model.ID, opts CreateFolderOpts) (*Folder, error)
	Get(ctx context.Context, id model.ID) (*Folder, error)
	List(ctx context.Context, libraryID model.ID, parentID *model.ID, page CursorPage) (Page[*Folder], error)
	Update(ctx context.Context, id model.ID, opts UpdateFolderOpts) (*Folder, error)
	Delete(ctx context.Context, id model.ID) error
}

type folderService struct {
	*baseService
}

func documentLibraryFromRepository(lib repository.DocumentLibrary) DocumentLibrary {
	return DocumentLibrary{
		ID:   lib.ID,
		Type: lib.Type,
		Name: lib.Name,
	}
}

func documentFolderFromRepository(folder *repository.DocumentFolder) *DocumentFolder {
	if folder == nil {
		return nil
	}
	return &DocumentFolder{
		ID:       folder.ID,
		Name:     folder.Name,
		ParentID: folder.ParentID,
	}
}

func folderFromRepository(f *repository.Folder) *Folder {
	if f == nil {
		return nil
	}
	return &Folder{
		ID:        f.ID,
		Name:      f.Name,
		Library:   documentLibraryFromRepository(f.Library),
		Parent:    documentFolderFromRepository(f.Parent),
		CreatedBy: partialUserValueFromRepository(f.CreatedBy),
		CreatedAt: f.CreatedAt,
		UpdatedAt: f.UpdatedAt,
	}
}

func (s *folderService) Create(ctx context.Context, libraryID model.ID, opts CreateFolderOpts) (*Folder, error) {
	ctx, span := s.tracer.Start(ctx, "service.folderService/Create")
	defer span.End()

	if err := libraryID.Validate(); err != nil {
		return nil, errors.Join(ErrFolderCreate, err)
	}
	if libraryID.Type != model.ResourceTypeOrganization && libraryID.Type != model.ResourceTypeNamespace {
		return nil, errors.Join(ErrFolderCreate, model.ErrInvalidID)
	}
	if err := opts.Validate(); err != nil {
		return nil, errors.Join(ErrFolderCreate, err)
	}

	if !s.permissionService.CtxUserHasPermission(ctx, libraryID, model.PermissionKindWrite) {
		return nil, errors.Join(ErrFolderCreate, ErrNoPermission)
	}

	userID, ok := ctx.Value(pkg.CtxKeyUserID).(model.ID)
	if !ok {
		return nil, errors.Join(ErrFolderCreate, model.ErrInvalidID)
	}

	folder, err := s.folderRepo.Create(ctx, repository.CreateFolderOpts{
		Library:   libraryID,
		ParentID:  opts.ParentID,
		Name:      opts.Name,
		CreatedBy: userID,
	})
	if err != nil {
		return nil, errors.Join(ErrFolderCreate, err)
	}

	return folderFromRepository(folder), nil
}

func (s *folderService) Get(ctx context.Context, id model.ID) (*Folder, error) {
	ctx, span := s.tracer.Start(ctx, "service.folderService/Get")
	defer span.End()

	if err := id.Validate(); err != nil {
		return nil, errors.Join(ErrFolderGet, err)
	}

	folder, err := s.folderRepo.Get(ctx, id)
	if err != nil {
		return nil, errors.Join(ErrFolderGet, err)
	}

	if !s.permissionService.CtxUserHasPermission(ctx, folder.Library.ID, model.PermissionKindRead) {
		return nil, errors.Join(ErrFolderGet, ErrNoPermission)
	}

	return folderFromRepository(folder), nil
}

func (s *folderService) List(ctx context.Context, libraryID model.ID, parentID *model.ID, page CursorPage) (Page[*Folder], error) {
	ctx, span := s.tracer.Start(ctx, "service.folderService/List")
	defer span.End()

	if err := libraryID.Validate(); err != nil {
		return Page[*Folder]{}, errors.Join(ErrFolderGetAll, err)
	}
	if libraryID.Type != model.ResourceTypeOrganization && libraryID.Type != model.ResourceTypeNamespace {
		return Page[*Folder]{}, errors.Join(ErrFolderGetAll, model.ErrInvalidID)
	}

	normalized, err := page.Normalize()
	if err != nil {
		return Page[*Folder]{}, errors.Join(ErrFolderGetAll, err)
	}

	if !s.permissionService.CtxUserHasPermission(ctx, libraryID, model.PermissionKindRead) {
		return Page[*Folder]{}, errors.Join(ErrFolderGetAll, ErrNoPermission)
	}

	folders, err := s.folderRepo.List(ctx, libraryID, parentID, normalized)
	if err != nil {
		return Page[*Folder]{}, errors.Join(ErrFolderGetAll, err)
	}

	return mapPage(folders, folderFromRepository), nil
}

func (s *folderService) Update(ctx context.Context, id model.ID, opts UpdateFolderOpts) (*Folder, error) {
	ctx, span := s.tracer.Start(ctx, "service.folderService/Update")
	defer span.End()

	if err := id.Validate(); err != nil {
		return nil, errors.Join(ErrFolderUpdate, err)
	}

	current, err := s.folderRepo.Get(ctx, id)
	if err != nil {
		return nil, errors.Join(ErrFolderUpdate, err)
	}

	if !s.permissionService.CtxUserHasPermission(ctx, current.Library.ID, model.PermissionKindWrite) {
		return nil, errors.Join(ErrFolderUpdate, ErrNoPermission)
	}

	folder, err := s.folderRepo.Update(ctx, id, repository.UpdateFolderOpts{
		Name:     opts.Name,
		ParentID: opts.ParentID,
	})
	if err != nil {
		return nil, errors.Join(ErrFolderUpdate, err)
	}

	return folderFromRepository(folder), nil
}

func (s *folderService) Delete(ctx context.Context, id model.ID) error {
	ctx, span := s.tracer.Start(ctx, "service.folderService/Delete")
	defer span.End()

	if err := id.Validate(); err != nil {
		return errors.Join(ErrFolderDelete, err)
	}

	current, err := s.folderRepo.Get(ctx, id)
	if err != nil {
		return errors.Join(ErrFolderDelete, err)
	}

	if !s.permissionService.CtxUserHasPermission(ctx, current.Library.ID, model.PermissionKindDelete) {
		return errors.Join(ErrFolderDelete, ErrNoPermission)
	}

	if err := s.folderRepo.Delete(ctx, id); err != nil {
		return errors.Join(ErrFolderDelete, err)
	}

	return nil
}

// NewFolderService returns a new instance of the FolderService interface.
func NewFolderService(opts ...Option) (FolderService, error) {
	s, err := newService(opts...)
	if err != nil {
		return nil, err
	}

	svc := &folderService{
		baseService: s,
	}

	if svc.folderRepo == nil {
		return nil, ErrNoFolderRepository
	}

	if svc.permissionService == nil {
		return nil, ErrNoPermissionService
	}

	return svc, nil
}

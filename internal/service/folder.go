package service

import (
	"context"
	"errors"
	"time"

	"github.com/opcotech/elemo/internal/model"
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
//go:generate go tool mockgen -destination=mock/mock_folder_gen.go -package=mocksvc . FolderService
type FolderService interface {
	Create(ctx context.Context, libraryID model.ID, opts CreateFolderOpts) (*Folder, error)
	Get(ctx context.Context, id model.ID) (*Folder, error)
	List(ctx context.Context, libraryID model.ID, parentID *model.ID, page CursorPage) (Page[*Folder], error)
	Update(ctx context.Context, id model.ID, opts UpdateFolderOpts) (*Folder, error)
	Delete(ctx context.Context, id model.ID) error
}

type folderService struct {
	runtime
	folderRepo        repository.FolderRepository
	permissionService PermissionService
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

	if err := requireAction(ctx, s.permissionService, libraryID, model.ActionFolderCreate); err != nil {
		return nil, errors.Join(ErrFolderCreate, err)
	}

	userID, err := ctxUserID(ctx)
	if err != nil {
		return nil, errors.Join(ErrFolderCreate, err)
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

	actions, err := roleTemplateActions(model.RoleKeyDocumentMaintainer)
	if err != nil {
		return nil, errors.Join(ErrFolderCreate, err)
	}
	if err := s.permissionService.BootstrapCreator(ctx, userID, folder.ID, actions); err != nil {
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

	if err := requireAction(ctx, s.permissionService, folder.ID, model.ActionDocumentRead); err != nil {
		return nil, errors.Join(ErrFolderGet, err)
	}

	return folderFromRepository(folder), nil
}

func (s *folderService) List(ctx context.Context, libraryID model.ID, parentID *model.ID, page CursorPage) (Page[*Folder], error) {
	ctx, span := s.tracer.Start(ctx, "service.folderService/List")
	defer span.End()

	if err := libraryID.Validate(); err != nil {
		return Page[*Folder]{}, errors.Join(ErrFolderList, err)
	}
	if libraryID.Type != model.ResourceTypeOrganization && libraryID.Type != model.ResourceTypeNamespace {
		return Page[*Folder]{}, errors.Join(ErrFolderList, model.ErrInvalidID)
	}

	normalized, err := page.Normalize()
	if err != nil {
		return Page[*Folder]{}, errors.Join(ErrFolderList, err)
	}

	userID, err := ctxUserID(ctx)
	if err != nil {
		return Page[*Folder]{}, errors.Join(ErrFolderList, err)
	}

	scopeIDs, allowed, err := resolvedListScopeIDs(ctx, s.permissionService, libraryID, model.ActionDocumentRead)
	if err != nil {
		return Page[*Folder]{}, errors.Join(ErrFolderList, err)
	}
	if !allowed {
		return repository.EmptyPage[*Folder](), nil
	}

	folders, err := s.folderRepo.ListForLibrary(ctx, repository.FolderListQuery{
		LibraryID: libraryID,
		ActorID:   userID,
		ScopeIDs:  scopeIDs,
		ParentID:  parentID,
		Page:      normalized,
		Order:     repository.SortDirectionDesc,
	})
	if err != nil {
		return Page[*Folder]{}, errors.Join(ErrFolderList, err)
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

	if err := requireAction(ctx, s.permissionService, current.ID, model.ActionDocumentUpdate); err != nil {
		return nil, errors.Join(ErrFolderUpdate, err)
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

	if err := requireAction(ctx, s.permissionService, current.ID, model.ActionDocumentDelete); err != nil {
		return errors.Join(ErrFolderDelete, err)
	}

	if err := s.folderRepo.Delete(ctx, id); err != nil {
		return errors.Join(ErrFolderDelete, err)
	}

	return nil
}

// NewFolderService returns a new instance of the FolderService interface.
func NewFolderService(folderRepo repository.FolderRepository, permissionService PermissionService, opts ...Option) (FolderService, error) {
	rt, err := newRuntime(opts...)
	if err != nil {
		return nil, err
	}

	svc := &folderService{
		runtime:           rt,
		folderRepo:        folderRepo,
		permissionService: permissionService,
	}

	if svc.folderRepo == nil {
		return nil, ErrNoFolderRepository
	}

	if svc.permissionService == nil {
		return nil, ErrNoPermissionService
	}

	return svc, nil
}

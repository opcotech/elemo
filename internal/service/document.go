package service

import (
	"context"
	"errors"
	"time"

	"github.com/opcotech/elemo/internal/license"
	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg"
	"github.com/opcotech/elemo/internal/pkg/log"
	"github.com/opcotech/elemo/internal/pkg/optional"
	"github.com/opcotech/elemo/internal/pkg/validate"
	"github.com/opcotech/elemo/internal/repository"
)

const documentFilePrefix = "documents/"

// PartialDocument represents a simplified document within a library or related view.
type PartialDocument struct {
	ID        model.ID
	Title     string
	Excerpt   string
	CreatedBy PartialUser
	CreatedAt *time.Time
	UpdatedAt *time.Time
}

// DocumentRelation is a project or issue a document is related to.
type DocumentRelation struct {
	ID   model.ID
	Type model.ResourceType
	Name string
}

// Document represents a document returned by the service.
type Document struct {
	ID              model.ID
	Title           string
	Excerpt         string
	FileID          string
	CreatedBy       PartialUser
	Library         DocumentLibrary
	Folder          *DocumentFolder
	Relations       []DocumentRelation
	Labels          []PartialLabel
	CommentCount    *int64
	AttachmentCount *int64
	CreatedAt       *time.Time
	UpdatedAt       *time.Time
	Content         []byte
}

// CreateDocumentOpts holds the data required to create a document.
type CreateDocumentOpts struct {
	Title   string `json:"title" validate:"required,min=3,max=120"`
	Excerpt string `json:"excerpt" validate:"omitempty,min=10,max=500"`
	Content []byte `json:"content" validate:"omitempty"`
}

// Validate validates the create options.
func (o *CreateDocumentOpts) Validate() error {
	if err := validate.Struct(o); err != nil {
		return errors.Join(model.ErrInvalidDocumentDetails, err)
	}
	return nil
}

// UpdateDocumentOpts holds the fields that can be updated on a document.
type UpdateDocumentOpts struct {
	Title     optional.Optional[string]
	Excerpt   optional.Optional[string]
	Content   optional.Optional[[]byte]
	LibraryID optional.Optional[model.ID]
	FolderID  optional.Optional[model.ID]
}

// LibraryListFilter selects which documents to return from a library.
type LibraryListFilter struct {
	FolderID *model.ID
	All      bool
}

// DocumentService serves the business logic of interacting with documents.
//
//go:generate go tool mockgen -destination=document_mock_gen.go -package=service -mock_names DocumentService=MockDocumentService . DocumentService
type DocumentService interface {
	// Create creates a new document in the library inferred from contextID.
	Create(ctx context.Context, contextID model.ID, opts CreateDocumentOpts) (*Document, error)
	// Get returns a document by its ID, including the stored body.
	Get(ctx context.Context, id model.ID) (*Document, error)
	// ListLibrary returns documents scoped to an organization or namespace.
	ListLibrary(ctx context.Context, libraryID model.ID, filter LibraryListFilter, page CursorPage) (Page[*PartialDocument], error)
	// ListRelated returns documents related to a project or issue.
	ListRelated(ctx context.Context, relatedTo model.ID, page CursorPage) (Page[*PartialDocument], error)
	// Update updates a document. Optional library and folder fields move it.
	Update(ctx context.Context, id model.ID, opts UpdateDocumentOpts) (*Document, error)
	// MoveLibrary replaces SCOPED_TO and clears LOCATED_IN.
	MoveLibrary(ctx context.Context, id, libraryID model.ID) (*Document, error)
	// MoveToFolder sets or clears LOCATED_IN. nil folderID is the library root.
	MoveToFolder(ctx context.Context, id model.ID, folderID *model.ID) (*Document, error)
	// Relate creates RELATED_TO a project or issue.
	Relate(ctx context.Context, id, targetID model.ID) error
	// Unrelate removes RELATED_TO a project or issue.
	Unrelate(ctx context.Context, id, targetID model.ID) error
	// Delete deletes a document.
	Delete(ctx context.Context, id model.ID) error
}

type documentService struct {
	*baseService
}

func partialDocumentFromDocument(d *repository.Document) *PartialDocument {
	if d == nil {
		return nil
	}
	return &PartialDocument{
		ID:        d.ID,
		Title:     d.Title,
		Excerpt:   d.Excerpt,
		CreatedBy: partialUserValueFromRepository(d.CreatedBy),
		CreatedAt: d.CreatedAt,
		UpdatedAt: d.UpdatedAt,
	}
}

func documentRelationsFromRepository(relations []repository.DocumentRelation) []DocumentRelation {
	out := make([]DocumentRelation, len(relations))
	for i, relation := range relations {
		out[i] = DocumentRelation{
			ID:   relation.ID,
			Type: relation.Type,
			Name: relation.Name,
		}
	}
	return out
}

func documentFromRepository(d *repository.Document, content []byte) *Document {
	if d == nil {
		return nil
	}

	return &Document{
		ID:              d.ID,
		Title:           d.Title,
		Excerpt:         d.Excerpt,
		FileID:          d.FileID,
		CreatedBy:       partialUserValueFromRepository(d.CreatedBy),
		Library:         documentLibraryFromRepository(d.Library),
		Folder:          documentFolderFromRepository(d.Folder),
		Relations:       documentRelationsFromRepository(d.Relations),
		Labels:          partialLabelsFromRepository(d.Labels),
		CommentCount:    d.CommentCount,
		AttachmentCount: d.AttachmentCount,
		CreatedAt:       d.CreatedAt,
		UpdatedAt:       d.UpdatedAt,
		Content:         content,
	}
}

// documentContent returns stored file bytes. A missing blob is treated as
// empty, so metadata and location updates can succeed when object storage has no object.
func (s *documentService) documentContent(ctx context.Context, fileID string) ([]byte, error) {
	content, err := s.staticFileService.Get(ctx, fileID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return []byte{}, nil
		}
		return nil, err
	}
	return content, nil
}

func (s *documentService) hasDocumentPermission(ctx context.Context, docID model.ID, action model.Action) bool {
	return s.permissionService.CtxUserHas(ctx, docID, action)
}

func relatedResourceReadAction(id model.ID) (model.Action, bool) {
	return model.ReadActionFor(id.Type)
}

func isLibraryID(id model.ID) bool {
	return id.Type == model.ResourceTypeOrganization || id.Type == model.ResourceTypeNamespace
}

func isRelatedID(id model.ID) bool {
	return id.Type == model.ResourceTypeProject || id.Type == model.ResourceTypeIssue
}

func (s *documentService) Create(ctx context.Context, contextID model.ID, opts CreateDocumentOpts) (*Document, error) {
	ctx, span := s.tracer.Start(ctx, "service.documentService/Create")
	defer span.End()

	if expired, err := s.licenseService.Expired(ctx); expired || err != nil {
		return nil, errors.Join(ErrDocumentCreate, license.ErrLicenseExpired)
	}

	if err := contextID.Validate(); err != nil {
		return nil, errors.Join(ErrDocumentCreate, err)
	}

	if err := opts.Validate(); err != nil {
		return nil, errors.Join(ErrDocumentCreate, err)
	}

	libraryID := contextID
	var relatedTo *model.ID
	switch contextID.Type {
	case model.ResourceTypeOrganization, model.ResourceTypeNamespace:
	case model.ResourceTypeProject, model.ResourceTypeIssue:
		resolved, err := s.documentRepo.ResolveLibrary(ctx, contextID)
		if err != nil {
			return nil, errors.Join(ErrDocumentCreate, err)
		}
		libraryID = resolved
		relatedTo = &contextID
	default:
		return nil, errors.Join(ErrDocumentCreate, model.ErrInvalidID)
	}

	if !s.permissionService.CtxUserHas(ctx, libraryID, model.ActionDocumentCreate) {
		return nil, errors.Join(ErrDocumentCreate, ErrNoPermission)
	}

	if ok, err := s.licenseService.WithinThreshold(ctx, license.QuotaDocuments); !ok || err != nil {
		return nil, errors.Join(ErrDocumentCreate, ErrQuotaExceeded)
	}

	userID, ok := ctx.Value(pkg.CtxKeyUserID).(model.ID)
	if !ok {
		return nil, errors.Join(ErrDocumentCreate, model.ErrInvalidID)
	}

	fileID := documentFilePrefix + model.NewRawID()
	if err := s.staticFileService.Create(ctx, fileID, opts.Content); err != nil {
		return nil, errors.Join(ErrDocumentCreate, err)
	}

	doc, err := s.documentRepo.Create(ctx, repository.CreateDocumentOpts{
		Library:   libraryID,
		RelatedTo: relatedTo,
		Title:     opts.Title,
		Excerpt:   opts.Excerpt,
		FileID:    fileID,
		CreatedBy: userID,
	})
	if err != nil {
		_ = s.staticFileService.Delete(ctx, fileID)
		return nil, errors.Join(ErrDocumentCreate, err)
	}

	actions, err := roleTemplateActions(model.RoleKeyDocumentMaintainer)
	if err != nil {
		return nil, errors.Join(ErrDocumentCreate, err)
	}
	if err := s.permissionService.BootstrapCreator(ctx, userID, doc.ID, actions); err != nil {
		return nil, errors.Join(ErrDocumentCreate, err)
	}

	out := documentFromRepository(doc, opts.Content)
	if err := s.searchService.Index(ctx, IndexInput{
		ID:        out.ID,
		Title:     out.Title,
		Content:   out.Excerpt,
		CreatedAt: out.CreatedAt,
		UpdatedAt: out.UpdatedAt,
	}); err != nil {
		s.logger.Warn(ctx, "failed to index search document",
			log.WithError(err),
			log.WithValue(out.ID.Composite()),
		)
	}
	return out, nil
}

func (s *documentService) Get(ctx context.Context, id model.ID) (*Document, error) {
	ctx, span := s.tracer.Start(ctx, "service.documentService/Get")
	defer span.End()

	if err := id.Validate(); err != nil {
		return nil, errors.Join(ErrDocumentGet, err)
	}

	doc, err := s.documentRepo.Get(ctx, id, repository.DocumentDetailProjection())
	if err != nil {
		return nil, errors.Join(ErrDocumentGet, err)
	}

	if !s.hasDocumentPermission(ctx, doc.ID, model.ActionDocumentRead) {
		return nil, errors.Join(ErrDocumentGet, ErrNoPermission)
	}

	content, err := s.staticFileService.Get(ctx, doc.FileID)
	if err != nil {
		return nil, errors.Join(ErrDocumentGet, err)
	}

	return documentFromRepository(doc, content), nil
}

func (s *documentService) ListLibrary(ctx context.Context, libraryID model.ID, filter LibraryListFilter, page CursorPage) (Page[*PartialDocument], error) {
	ctx, span := s.tracer.Start(ctx, "service.documentService/ListLibrary")
	defer span.End()

	if err := libraryID.Validate(); err != nil {
		return Page[*PartialDocument]{}, errors.Join(ErrDocumentGetAll, err)
	}
	if !isLibraryID(libraryID) {
		return Page[*PartialDocument]{}, errors.Join(ErrDocumentGetAll, model.ErrInvalidID)
	}

	normalized, err := page.Normalize()
	if err != nil {
		return Page[*PartialDocument]{}, errors.Join(ErrDocumentGetAll, err)
	}

	userID, ok := ctx.Value(pkg.CtxKeyUserID).(model.ID)
	if !ok {
		return Page[*PartialDocument]{}, errors.Join(ErrDocumentGetAll, ErrNoUser)
	}

	documents, err := s.documentRepo.ListLibrary(ctx, libraryID, userID, repository.LibraryListFilter{
		FolderID: filter.FolderID,
		All:      filter.All,
	}, normalized, repository.DocumentSummaryProjection())
	if err != nil {
		return Page[*PartialDocument]{}, errors.Join(ErrDocumentGetAll, err)
	}

	return mapPage(documents, partialDocumentFromDocument), nil
}

func (s *documentService) ListRelated(ctx context.Context, relatedTo model.ID, page CursorPage) (Page[*PartialDocument], error) {
	ctx, span := s.tracer.Start(ctx, "service.documentService/ListRelated")
	defer span.End()

	if err := relatedTo.Validate(); err != nil {
		return Page[*PartialDocument]{}, errors.Join(ErrDocumentGetAll, err)
	}
	if !isRelatedID(relatedTo) {
		return Page[*PartialDocument]{}, errors.Join(ErrDocumentGetAll, model.ErrInvalidID)
	}

	normalized, err := page.Normalize()
	if err != nil {
		return Page[*PartialDocument]{}, errors.Join(ErrDocumentGetAll, err)
	}

	action, ok := relatedResourceReadAction(relatedTo)
	if !ok || !s.permissionService.CtxUserHas(ctx, relatedTo, action) {
		return Page[*PartialDocument]{}, errors.Join(ErrDocumentGetAll, ErrNoPermission)
	}

	userID, ok := ctx.Value(pkg.CtxKeyUserID).(model.ID)
	if !ok {
		return Page[*PartialDocument]{}, errors.Join(ErrDocumentGetAll, ErrNoUser)
	}

	documents, err := s.documentRepo.ListRelated(ctx, relatedTo, userID, normalized, repository.DocumentSummaryProjection())
	if err != nil {
		return Page[*PartialDocument]{}, errors.Join(ErrDocumentGetAll, err)
	}

	return mapPage(documents, partialDocumentFromDocument), nil
}

func (s *documentService) Update(ctx context.Context, id model.ID, opts UpdateDocumentOpts) (*Document, error) {
	ctx, span := s.tracer.Start(ctx, "service.documentService/Update")
	defer span.End()

	if expired, err := s.licenseService.Expired(ctx); expired || err != nil {
		return nil, errors.Join(ErrDocumentUpdate, license.ErrLicenseExpired)
	}

	if err := id.Validate(); err != nil {
		return nil, errors.Join(ErrDocumentUpdate, err)
	}

	current, err := s.documentRepo.Get(ctx, id, repository.DocumentDetailProjection())
	if err != nil {
		return nil, errors.Join(ErrDocumentUpdate, err)
	}

	if !s.hasDocumentPermission(ctx, current.ID, model.ActionDocumentUpdate) {
		return nil, errors.Join(ErrDocumentUpdate, ErrNoPermission)
	}

	if opts.Content.Defined && opts.Content.Value != nil {
		if err := s.staticFileService.Update(ctx, current.FileID, *opts.Content.Value); err != nil {
			return nil, errors.Join(ErrDocumentUpdate, err)
		}
	}

	if opts.Title.Defined || opts.Excerpt.Defined {
		current, err = s.documentRepo.Update(ctx, id, repository.UpdateDocumentOpts{
			Title:   opts.Title,
			Excerpt: opts.Excerpt,
		})
		if err != nil {
			return nil, errors.Join(ErrDocumentUpdate, err)
		}
	}

	if opts.LibraryID.Defined && opts.LibraryID.Value != nil {
		moved, err := s.MoveLibrary(ctx, id, *opts.LibraryID.Value)
		if err != nil {
			return nil, err
		}
		if !opts.FolderID.Defined {
			return moved, nil
		}
	}

	if opts.FolderID.Defined {
		moved, err := s.MoveToFolder(ctx, id, opts.FolderID.Value)
		if err != nil {
			return nil, err
		}
		if err := s.searchService.Index(ctx, IndexInput{
			ID:        moved.ID,
			Title:     moved.Title,
			Content:   moved.Excerpt,
			CreatedAt: moved.CreatedAt,
			UpdatedAt: moved.UpdatedAt,
		}); err != nil {
			s.logger.Warn(ctx, "failed to index search document",
				log.WithError(err),
				log.WithValue(moved.ID.Composite()),
			)
		}
		return moved, nil
	}

	content, err := s.documentContent(ctx, current.FileID)
	if err != nil {
		return nil, errors.Join(ErrDocumentUpdate, err)
	}

	out := documentFromRepository(current, content)
	if err := s.searchService.Index(ctx, IndexInput{
		ID:        out.ID,
		Title:     out.Title,
		Content:   out.Excerpt,
		CreatedAt: out.CreatedAt,
		UpdatedAt: out.UpdatedAt,
	}); err != nil {
		s.logger.Warn(ctx, "failed to index search document",
			log.WithError(err),
			log.WithValue(out.ID.Composite()),
		)
	}
	return out, nil
}

func (s *documentService) MoveLibrary(ctx context.Context, id, libraryID model.ID) (*Document, error) {
	ctx, span := s.tracer.Start(ctx, "service.documentService/MoveLibrary")
	defer span.End()

	if err := id.Validate(); err != nil {
		return nil, errors.Join(ErrDocumentMove, err)
	}
	if err := libraryID.Validate(); err != nil {
		return nil, errors.Join(ErrDocumentMove, err)
	}
	if !isLibraryID(libraryID) {
		return nil, errors.Join(ErrDocumentMove, model.ErrInvalidID)
	}

	resolved, err := s.documentRepo.ResolveLibrary(ctx, libraryID)
	if err != nil {
		return nil, errors.Join(ErrDocumentMove, err)
	}
	libraryID = resolved

	current, err := s.documentRepo.Get(ctx, id, repository.DocumentDetailProjection())
	if err != nil {
		return nil, errors.Join(ErrDocumentMove, err)
	}

	if !s.hasDocumentPermission(ctx, current.ID, model.ActionDocumentUpdate) {
		return nil, errors.Join(ErrDocumentMove, ErrNoPermission)
	}
	if !s.permissionService.CtxUserHas(ctx, libraryID, model.ActionDocumentUpdate) {
		return nil, errors.Join(ErrDocumentMove, ErrNoPermission)
	}

	content, err := s.documentContent(ctx, current.FileID)
	if err != nil {
		return nil, errors.Join(ErrDocumentMove, err)
	}

	doc, err := s.documentRepo.MoveLibrary(ctx, id, libraryID)
	if err != nil {
		return nil, errors.Join(ErrDocumentMove, err)
	}

	out := documentFromRepository(doc, content)
	if err := s.searchService.Index(ctx, IndexInput{
		ID:        out.ID,
		Title:     out.Title,
		Content:   out.Excerpt,
		CreatedAt: out.CreatedAt,
		UpdatedAt: out.UpdatedAt,
	}); err != nil {
		s.logger.Warn(ctx, "failed to index search document",
			log.WithError(err),
			log.WithValue(out.ID.Composite()),
		)
	}
	return out, nil
}

func (s *documentService) MoveToFolder(ctx context.Context, id model.ID, folderID *model.ID) (*Document, error) {
	ctx, span := s.tracer.Start(ctx, "service.documentService/MoveToFolder")
	defer span.End()

	if err := id.Validate(); err != nil {
		return nil, errors.Join(ErrDocumentMove, err)
	}
	if folderID != nil {
		if err := folderID.Validate(); err != nil {
			return nil, errors.Join(ErrDocumentMove, err)
		}
		if folderID.Type != model.ResourceTypeFolder {
			return nil, errors.Join(ErrDocumentMove, model.ErrInvalidID)
		}
	}

	current, err := s.documentRepo.Get(ctx, id, repository.DocumentDetailProjection())
	if err != nil {
		return nil, errors.Join(ErrDocumentMove, err)
	}

	if !s.hasDocumentPermission(ctx, current.ID, model.ActionDocumentUpdate) {
		return nil, errors.Join(ErrDocumentMove, ErrNoPermission)
	}

	content, err := s.documentContent(ctx, current.FileID)
	if err != nil {
		return nil, errors.Join(ErrDocumentMove, err)
	}

	doc, err := s.documentRepo.MoveToFolder(ctx, id, folderID)
	if err != nil {
		return nil, errors.Join(ErrDocumentMove, err)
	}

	return documentFromRepository(doc, content), nil
}

func (s *documentService) Relate(ctx context.Context, id, targetID model.ID) error {
	ctx, span := s.tracer.Start(ctx, "service.documentService/Relate")
	defer span.End()

	if err := id.Validate(); err != nil {
		return errors.Join(ErrDocumentRelate, err)
	}
	if err := targetID.Validate(); err != nil {
		return errors.Join(ErrDocumentRelate, err)
	}
	if !isRelatedID(targetID) {
		return errors.Join(ErrDocumentRelate, model.ErrInvalidID)
	}

	current, err := s.documentRepo.Get(ctx, id, repository.DocumentDetailProjection())
	if err != nil {
		return errors.Join(ErrDocumentRelate, err)
	}

	if !s.hasDocumentPermission(ctx, current.ID, model.ActionDocumentUpdate) {
		return errors.Join(ErrDocumentRelate, ErrNoPermission)
	}
	action, ok := relatedResourceReadAction(targetID)
	if !ok || !s.permissionService.CtxUserHas(ctx, targetID, action) {
		return errors.Join(ErrDocumentRelate, ErrNoPermission)
	}

	if err := s.documentRepo.Relate(ctx, id, targetID); err != nil {
		return errors.Join(ErrDocumentRelate, err)
	}
	return nil
}

func (s *documentService) Unrelate(ctx context.Context, id, targetID model.ID) error {
	ctx, span := s.tracer.Start(ctx, "service.documentService/Unrelate")
	defer span.End()

	if err := id.Validate(); err != nil {
		return errors.Join(ErrDocumentUnrelate, err)
	}
	if err := targetID.Validate(); err != nil {
		return errors.Join(ErrDocumentUnrelate, err)
	}
	if !isRelatedID(targetID) {
		return errors.Join(ErrDocumentUnrelate, model.ErrInvalidID)
	}

	current, err := s.documentRepo.Get(ctx, id, repository.DocumentDetailProjection())
	if err != nil {
		return errors.Join(ErrDocumentUnrelate, err)
	}

	if !s.hasDocumentPermission(ctx, current.ID, model.ActionDocumentUpdate) {
		return errors.Join(ErrDocumentUnrelate, ErrNoPermission)
	}
	action, ok := relatedResourceReadAction(targetID)
	if !ok || !s.permissionService.CtxUserHas(ctx, targetID, action) {
		return errors.Join(ErrDocumentUnrelate, ErrNoPermission)
	}

	if err := s.documentRepo.Unrelate(ctx, id, targetID); err != nil {
		return errors.Join(ErrDocumentUnrelate, err)
	}
	return nil
}

func (s *documentService) Delete(ctx context.Context, id model.ID) error {
	ctx, span := s.tracer.Start(ctx, "service.documentService/Delete")
	defer span.End()

	if expired, err := s.licenseService.Expired(ctx); expired || err != nil {
		return errors.Join(ErrDocumentDelete, license.ErrLicenseExpired)
	}

	if err := id.Validate(); err != nil {
		return errors.Join(ErrDocumentDelete, err)
	}

	current, err := s.documentRepo.Get(ctx, id, repository.DocumentDetailProjection())
	if err != nil {
		return errors.Join(ErrDocumentDelete, err)
	}

	if !s.hasDocumentPermission(ctx, current.ID, model.ActionDocumentDelete) {
		return errors.Join(ErrDocumentDelete, ErrNoPermission)
	}

	if err := s.documentRepo.Delete(ctx, id); err != nil {
		return errors.Join(ErrDocumentDelete, err)
	}

	if err := s.staticFileService.Delete(ctx, current.FileID); err != nil {
		return errors.Join(ErrDocumentDelete, err)
	}

	if err := s.searchService.Delete(ctx, id); err != nil {
		s.logger.Warn(ctx, "failed to delete search document",
			log.WithError(err),
			log.WithValue(id.Composite()),
		)
	}
	return nil
}

// NewDocumentService returns a new instance of the DocumentService interface.
func NewDocumentService(opts ...Option) (DocumentService, error) {
	s, err := newService(opts...)
	if err != nil {
		return nil, err
	}

	svc := &documentService{
		baseService: s,
	}

	if svc.documentRepo == nil {
		return nil, ErrNoDocumentRepository
	}

	if svc.licenseService == nil {
		return nil, ErrNoLicenseService
	}

	if svc.permissionService == nil {
		return nil, ErrNoPermissionService
	}

	if svc.staticFileService == nil {
		return nil, ErrNoStaticFileService
	}

	if svc.searchService == nil {
		return nil, ErrNoSearchService
	}

	return svc, nil
}

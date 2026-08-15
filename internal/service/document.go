package service

import (
	"context"
	"errors"
	"time"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/repository"
)

// PartialDocument represents a simplified document within a namespace or project.
type PartialDocument struct {
	ID        model.ID
	Name      string
	Excerpt   string
	CreatedBy PartialUser
	CreatedAt *time.Time
}

// DocumentService serves the business logic of reading documents.
//
//go:generate go tool mockgen -destination=document_mock_gen.go -package=service -mock_names DocumentService=MockDocumentService . DocumentService
type DocumentService interface {
	// ListBelongsTo returns a cursor-paginated page of documents that belong
	// to a namespace or project.
	ListBelongsTo(ctx context.Context, belongsTo model.ID, page CursorPage) (Page[*PartialDocument], error)
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
		Name:      d.Name,
		Excerpt:   d.Excerpt,
		CreatedBy: partialUserValueFromRepository(d.CreatedBy),
		CreatedAt: d.CreatedAt,
	}
}

func (s *documentService) ListBelongsTo(ctx context.Context, belongsTo model.ID, page CursorPage) (Page[*PartialDocument], error) {
	ctx, span := s.tracer.Start(ctx, "service.documentService/ListBelongsTo")
	defer span.End()

	if err := belongsTo.Validate(); err != nil {
		return Page[*PartialDocument]{}, errors.Join(ErrDocumentGetAll, err)
	}

	normalized, err := page.Normalize()
	if err != nil {
		return Page[*PartialDocument]{}, errors.Join(ErrDocumentGetAll, err)
	}

	if !s.permissionService.CtxUserHasPermission(ctx, belongsTo, model.PermissionKindRead) {
		return Page[*PartialDocument]{}, errors.Join(ErrDocumentGetAll, ErrNoPermission)
	}

	documents, err := s.documentRepo.ListBelongsTo(ctx, belongsTo, normalized, repository.DocumentSummaryProjection())
	if err != nil {
		return Page[*PartialDocument]{}, errors.Join(ErrDocumentGetAll, err)
	}

	return mapPage(documents, partialDocumentFromDocument), nil
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

	if svc.permissionService == nil {
		return nil, ErrNoPermissionService
	}

	return svc, nil
}

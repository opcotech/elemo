package service

import (
	"context"
	"errors"
	"time"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/repository"
)

// PartialLabel is a lean label used on issue and document reads.
type PartialLabel struct {
	ID   model.ID
	Name string
}

// Label represents a label returned by the service.
type Label struct {
	ID          model.ID
	Name        string
	Description string
	CreatedAt   *time.Time
	UpdatedAt   *time.Time
}

// LabelService serves the business logic of interacting with labels.
//
//go:generate go tool mockgen -destination=mock/mock_label_gen.go -package=mocksvc . LabelService
type LabelService interface {
	// List returns a cursor-paginated page of labels.
	List(ctx context.Context, page CursorPage) (Page[*Label], error)
}

// labelService is the concrete implementation of LabelService.
type labelService struct {
	runtime
	labelRepo repository.LabelRepository
}

func labelFromRepository(l *repository.Label) *Label {
	if l == nil {
		return nil
	}
	return &Label{
		ID:          l.ID,
		Name:        l.Name,
		Description: l.Description,
		CreatedAt:   l.CreatedAt,
		UpdatedAt:   l.UpdatedAt,
	}
}

func (s *labelService) List(ctx context.Context, page CursorPage) (Page[*Label], error) {
	ctx, span := s.tracer.Start(ctx, "service.labelService/List")
	defer span.End()

	normalized, err := page.Normalize()
	if err != nil {
		return Page[*Label]{}, errors.Join(ErrLabelList, err)
	}

	if _, err := ctxUserID(ctx); err != nil {
		return Page[*Label]{}, errors.Join(ErrLabelList, err)
	}

	labels, err := s.labelRepo.List(ctx, normalized, repository.LabelListProjection())
	if err != nil {
		return Page[*Label]{}, errors.Join(ErrLabelList, err)
	}

	return mapPage(labels, labelFromRepository), nil
}

// NewLabelService returns a new instance of the LabelService interface.
func NewLabelService(labelRepo repository.LabelRepository, opts ...Option) (LabelService, error) {
	rt, err := newRuntime(opts...)
	if err != nil {
		return nil, err
	}

	svc := &labelService{
		runtime:   rt,
		labelRepo: labelRepo,
	}

	if svc.labelRepo == nil {
		return nil, ErrNoLabelRepository
	}

	return svc, nil
}

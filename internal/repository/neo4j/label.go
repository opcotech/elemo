package neo4j

import (
	"context"

	"github.com/opcotech/elemo/internal/model"
)

// LabelRepository is a repository for managing labels.
type LabelRepository struct {
	*repository
}

func (r *LabelRepository) Create(ctx context.Context, label *model.Label) error {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.LabelRepository/Create")
	defer span.End()

	panic("not implemented")
}

func (r *LabelRepository) Get(ctx context.Context, id model.ID) (*model.Label, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.LabelRepository/Get")
	defer span.End()

	panic("not implemented")
}

func (r *LabelRepository) Update(ctx context.Context, label *model.Label) error {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.LabelRepository/Update")
	defer span.End()

	panic("not implemented")
}

func (r *LabelRepository) Delete(ctx context.Context, id model.ID) error {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.LabelRepository/Delete")
	defer span.End()

	panic("not implemented")
}

// NewLabelRepository creates a new label repository.
func NewLabelRepository(opts ...RepositoryOption) (*LabelRepository, error) {
	baseRepo, err := newRepository(opts...)
	if err != nil {
		return nil, err
	}

	return &LabelRepository{
		repository: baseRepo,
	}, nil
}

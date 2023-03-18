package neo4j

import (
	"context"

	"github.com/opcotech/elemo/internal/model"
)

// DocumentRepository is a repository for managing documents.
type DocumentRepository struct {
	*repository
}

func (r *DocumentRepository) Create(ctx context.Context, document *model.Document) error {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.DocumentRepository/Create")
	defer span.End()

	panic("not implemented")
}

func (r *DocumentRepository) Get(ctx context.Context, id model.ID) (*model.Document, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.DocumentRepository/Get")
	defer span.End()

	panic("not implemented")
}

func (r *DocumentRepository) GetByOwner(ctx context.Context, id model.ID) (*[]model.Document, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.DocumentRepository/GetByOwner")
	defer span.End()

	panic("not implemented")
}

func (r *DocumentRepository) Update(ctx context.Context, document *model.Document) error {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.DocumentRepository/Update")
	defer span.End()

	panic("not implemented")
}

func (r *DocumentRepository) Delete(ctx context.Context, id model.ID) error {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.DocumentRepository/Delete")
	defer span.End()

	panic("not implemented")
}

// NewDocumentRepository creates a new document repository.
func NewDocumentRepository(opts ...RepositoryOption) (*DocumentRepository, error) {
	baseRepo, err := newRepository(opts...)
	if err != nil {
		return nil, err
	}

	return &DocumentRepository{
		repository: baseRepo,
	}, nil
}

package neo4j

import (
	"context"

	"github.com/opcotech/elemo/internal/model"
)

// NamespaceRepository is a repository for managing namespaces.
type NamespaceRepository struct {
	*repository
}

func (r *NamespaceRepository) Create(ctx context.Context, namespace *model.Namespace) error {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.NamespaceRepository/Create")
	defer span.End()

	panic("not implemented")
}

func (r *NamespaceRepository) Get(ctx context.Context, id model.ID) (*model.Namespace, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.NamespaceRepository/Get")
	defer span.End()

	panic("not implemented")
}

func (r *NamespaceRepository) Update(ctx context.Context, namespace *model.Namespace) error {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.NamespaceRepository/Update")
	defer span.End()

	panic("not implemented")
}

func (r *NamespaceRepository) Delete(ctx context.Context, id model.ID) error {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.NamespaceRepository/Delete")
	defer span.End()

	panic("not implemented")
}

// NewNamespaceRepository creates a new namespace repository.
func NewNamespaceRepository(opts ...RepositoryOption) (*NamespaceRepository, error) {
	baseRepo, err := newRepository(opts...)
	if err != nil {
		return nil, err
	}

	return &NamespaceRepository{
		repository: baseRepo,
	}, nil
}

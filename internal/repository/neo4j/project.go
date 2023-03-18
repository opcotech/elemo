package neo4j

import (
	"context"

	"github.com/opcotech/elemo/internal/model"
)

// ProjectRepository is a repository for managing projects.
type ProjectRepository struct {
	*repository
}

func (r *ProjectRepository) Create(ctx context.Context, owner model.ID, project *model.Project) error {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.ProjectRepository/Create")
	defer span.End()

	panic("not implemented")
}

func (r *ProjectRepository) Get(ctx context.Context, id model.ID) (*model.Project, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.ProjectRepository/Get")
	defer span.End()

	panic("not implemented")
}

func (r *ProjectRepository) Update(ctx context.Context, project *model.Project) error {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.ProjectRepository/Update")
	defer span.End()

	panic("not implemented")
}

func (r *ProjectRepository) Delete(ctx context.Context, id model.ID) error {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.ProjectRepository/Delete")
	defer span.End()

	panic("not implemented")
}

// NewProjectRepository creates a new project repository.
func NewProjectRepository(opts ...RepositoryOption) (*ProjectRepository, error) {
	baseRepo, err := newRepository(opts...)
	if err != nil {
		return nil, err
	}

	return &ProjectRepository{
		repository: baseRepo,
	}, nil
}

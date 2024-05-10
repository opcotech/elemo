package neo4j

import (
	"context"
	"errors"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg/convert"
	"github.com/opcotech/elemo/internal/repository"
)

// ProjectRepository is a repository for managing projects.
type ProjectRepository struct {
	*baseRepository
}

func (r *ProjectRepository) scan(pp, dp, tp, ip string) func(rec *neo4j.Record) (*model.Project, error) {
	return func(rec *neo4j.Record) (*model.Project, error) {
		p := new(model.Project)

		val, _, err := neo4j.GetRecordValue[neo4j.Node](rec, pp)
		if err != nil {
			return nil, err
		}

		if err := ScanIntoStruct(&val, &p, []string{"id"}); err != nil {
			return nil, err
		}

		p.ID, _ = model.NewIDFromString(val.GetProperties()["id"].(string), model.ResourceTypeProject.String())

		if p.Documents, err = ParseIDsFromRecord(rec, dp, model.ResourceTypeDocument.String()); err != nil {
			return nil, err
		}

		if p.Teams, err = ParseIDsFromRecord(rec, tp, model.ResourceTypeRole.String()); err != nil {
			return nil, err
		}

		if p.Issues, err = ParseIDsFromRecord(rec, ip, model.ResourceTypeIssue.String()); err != nil {
			return nil, err
		}

		if err := p.Validate(); err != nil {
			return nil, err
		}

		return p, nil
	}
}

func (r *ProjectRepository) Create(ctx context.Context, namespaceID model.ID, project *model.Project) error {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.ProjectRepository/Create")
	defer span.End()

	if err := namespaceID.Validate(); err != nil {
		return errors.Join(repository.ErrProjectCreate, err)
	}

	if err := project.Validate(); err != nil {
		return errors.Join(repository.ErrProjectCreate, err)
	}

	createdAt := time.Now().UTC()

	project.ID = model.MustNewID(model.ResourceTypeProject)
	project.CreatedAt = convert.ToPointer(createdAt)
	project.UpdatedAt = nil

	cypher := `
	MATCH (n:` + namespaceID.Label() + ` {id: $namespace_id})
	CREATE
		(p:` + project.ID.Label() + ` {
			id: $id, key: $key, name: $name, description: $description, logo: $logo, status: $status,
			created_at: datetime($created_at)
		}),
		(n)-[:` + EdgeKindHasProject.String() + `]->(p)`

	params := map[string]any{
		"id":           project.ID.String(),
		"key":          project.Key,
		"name":         project.Name,
		"description":  project.Description,
		"logo":         project.Logo,
		"status":       project.Status.String(),
		"created_at":   createdAt.Format(time.RFC3339Nano),
		"namespace_id": namespaceID.String(),
	}

	if err := ExecuteWriteAndConsume(ctx, r.db, cypher, params); err != nil {
		return errors.Join(repository.ErrProjectCreate, err)
	}

	return nil
}

func (r *ProjectRepository) Get(ctx context.Context, id model.ID) (*model.Project, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.ProjectRepository/Get")
	defer span.End()

	cypher := `
	MATCH (p:` + id.Label() + ` {id: $id})
	OPTIONAL MATCH (d:` + model.ResourceTypeDocument.String() + `)-[:` + EdgeKindBelongsTo.String() + `]->(p)
	OPTIONAL MATCH (p)-[:` + EdgeKindHasTeam.String() + `]->(t:` + model.ResourceTypeRole.String() + `)
	OPTIONAL MATCH (p)<-[:` + EdgeKindBelongsTo.String() + `]-(i:` + model.ResourceTypeIssue.String() + `)
	RETURN p, d, t, i`

	params := map[string]any{
		"id": id.String(),
	}

	project, err := ExecuteReadAndReadSingle(ctx, r.db, cypher, params, r.scan("p", "d", "t", "i"))
	if err != nil {
		return nil, errors.Join(repository.ErrProjectRead, err)
	}

	return project, nil
}

func (r *ProjectRepository) GetByKey(ctx context.Context, key string) (*model.Project, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.ProjectRepository/GetByKey")
	defer span.End()

	cypher := `
	MATCH (p:` + model.ResourceTypeProject.String() + ` {key: $key})
	OPTIONAL MATCH (d:` + model.ResourceTypeDocument.String() + `)-[:` + EdgeKindBelongsTo.String() + `]->(p)
	OPTIONAL MATCH (p)-[:` + EdgeKindHasTeam.String() + `]->(t:` + model.ResourceTypeRole.String() + `)
	OPTIONAL MATCH (p)<-[:` + EdgeKindBelongsTo.String() + `]-(i:` + model.ResourceTypeIssue.String() + `)
	RETURN p, d, t, i`

	params := map[string]any{
		"key": key,
	}

	project, err := ExecuteReadAndReadSingle(ctx, r.db, cypher, params, r.scan("p", "d", "t", "i"))
	if err != nil {
		return nil, errors.Join(repository.ErrProjectRead, err)
	}

	return project, nil
}

func (r *ProjectRepository) GetAll(ctx context.Context, namespaceID model.ID, offset, limit int) ([]*model.Project, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.ProjectRepository/GetAll")
	defer span.End()

	cypher := `
	MATCH (:` + namespaceID.Label() + ` {id: $namespace_id})-[:` + EdgeKindHasProject.String() + `]->(p)
	OPTIONAL MATCH (d:` + model.ResourceTypeDocument.String() + `)-[:` + EdgeKindBelongsTo.String() + `]->(p)
	OPTIONAL MATCH (p)-[:` + EdgeKindHasTeam.String() + `]->(t:` + model.ResourceTypeRole.String() + `)
	OPTIONAL MATCH (p)<-[:` + EdgeKindBelongsTo.String() + `]-(i:` + model.ResourceTypeIssue.String() + `)
	RETURN p, d, t, i
	ORDER BY p.created_at DESC
	SKIP $offset LIMIT $limit`

	params := map[string]any{
		"namespace_id": namespaceID.String(),
		"offset":       offset,
		"limit":        limit,
	}

	projects, err := ExecuteReadAndReadAll(ctx, r.db, cypher, params, r.scan("p", "d", "t", "i"))
	if err != nil {
		return nil, errors.Join(repository.ErrProjectRead, err)
	}

	return projects, nil
}

func (r *ProjectRepository) Update(ctx context.Context, id model.ID, patch map[string]any) (*model.Project, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.ProjectRepository/Update")
	defer span.End()

	cypher := `
	MATCH (p:` + id.Label() + ` {id: $id})
	SET p += $patch, p.updated_at = datetime()
	WITH p
	OPTIONAL MATCH (d:` + model.ResourceTypeDocument.String() + `)-[:` + EdgeKindBelongsTo.String() + `]->(p)
	OPTIONAL MATCH (p)-[:` + EdgeKindHasTeam.String() + `]->(t:` + model.ResourceTypeRole.String() + `)
	OPTIONAL MATCH (p)<-[:` + EdgeKindBelongsTo.String() + `]-(i:` + model.ResourceTypeIssue.String() + `)
	RETURN p, d, t, i`

	params := map[string]any{
		"id":    id.String(),
		"patch": patch,
	}

	project, err := ExecuteWriteAndReadSingle(ctx, r.db, cypher, params, r.scan("p", "d", "t", "i"))
	if err != nil {
		return nil, errors.Join(repository.ErrProjectUpdate, err)
	}

	return project, nil
}

func (r *ProjectRepository) Delete(ctx context.Context, id model.ID) error {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.ProjectRepository/Delete")
	defer span.End()

	cypher := `MATCH (p:` + id.Label() + ` {id: $id}) DETACH DELETE p`
	params := map[string]any{
		"id": id.String(),
	}

	if err := ExecuteWriteAndConsume(ctx, r.db, cypher, params); err != nil {
		return errors.Join(repository.ErrProjectDelete, err)
	}

	return nil
}

// NewProjectRepository creates a new project baseRepository.
func NewProjectRepository(opts ...RepositoryOption) (*ProjectRepository, error) {
	baseRepo, err := newRepository(opts...)
	if err != nil {
		return nil, err
	}

	return &ProjectRepository{
		baseRepository: baseRepo,
	}, nil
}

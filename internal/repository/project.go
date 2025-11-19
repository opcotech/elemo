package repository

import (
	"context"
	"errors"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg/convert"
)

var (
	ErrProjectCreate = errors.New("failed to create project") // project cannot be created
	ErrProjectDelete = errors.New("failed to delete project") // project cannot be deleted
	ErrProjectRead   = errors.New("failed to read project")   // project cannot be read
	ErrProjectUpdate = errors.New("failed to update project") // project cannot be updated
)

//go:generate mockgen -source=project.go -destination=../testutil/mock/project_repo_gen.go -package=mock -mock_names "ProjectRepository=ProjectRepository"
type ProjectRepository interface {
	Create(ctx context.Context, namespaceID model.ID, project *model.Project) error
	Get(ctx context.Context, id model.ID) (*model.Project, error)
	GetByKey(ctx context.Context, key string) (*model.Project, error)
	GetAll(ctx context.Context, namespaceID model.ID, offset, limit int) ([]*model.Project, error)
	Update(ctx context.Context, id model.ID, patch map[string]any) (*model.Project, error)
	Delete(ctx context.Context, id model.ID) error
}

// ProjectRepository is a repository for managing projects.
type Neo4jProjectRepository struct {
	*neo4jBaseRepository
}

func (r *Neo4jProjectRepository) scan(pp, dp, tp, ip string) func(rec *neo4j.Record) (*model.Project, error) {
	return func(rec *neo4j.Record) (*model.Project, error) {
		p := new(model.Project)

		val, _, err := neo4j.GetRecordValue[neo4j.Node](rec, pp)
		if err != nil {
			return nil, err
		}

		if err := Neo4jScanIntoStruct(&val, &p, []string{"id"}); err != nil {
			return nil, err
		}

		p.ID, _ = model.NewIDFromString(val.GetProperties()["id"].(string), model.ResourceTypeProject.String())

		if p.Documents, err = Neo4jParseIDsFromRecord(rec, dp, model.ResourceTypeDocument.String()); err != nil {
			return nil, err
		}

		if p.Teams, err = Neo4jParseIDsFromRecord(rec, tp, model.ResourceTypeRole.String()); err != nil {
			return nil, err
		}

		if p.Issues, err = Neo4jParseIDsFromRecord(rec, ip, model.ResourceTypeIssue.String()); err != nil {
			return nil, err
		}

		if err := p.Validate(); err != nil {
			return nil, err
		}

		return p, nil
	}
}

func (r *Neo4jProjectRepository) Create(ctx context.Context, namespaceID model.ID, project *model.Project) error {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.ProjectRepository/Create")
	defer span.End()

	if err := namespaceID.Validate(); err != nil {
		return errors.Join(ErrProjectCreate, err)
	}

	if err := project.Validate(); err != nil {
		return errors.Join(ErrProjectCreate, err)
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

	if err := Neo4jExecuteWriteAndConsume(ctx, r.db, cypher, params); err != nil {
		return errors.Join(ErrProjectCreate, err)
	}

	return nil
}

func (r *Neo4jProjectRepository) Get(ctx context.Context, id model.ID) (*model.Project, error) {
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

	project, err := Neo4jExecuteReadAndReadSingle(ctx, r.db, cypher, params, r.scan("p", "d", "t", "i"))
	if err != nil {
		return nil, errors.Join(ErrProjectRead, err)
	}

	return project, nil
}

func (r *Neo4jProjectRepository) GetByKey(ctx context.Context, key string) (*model.Project, error) {
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

	project, err := Neo4jExecuteReadAndReadSingle(ctx, r.db, cypher, params, r.scan("p", "d", "t", "i"))
	if err != nil {
		return nil, errors.Join(ErrProjectRead, err)
	}

	return project, nil
}

func (r *Neo4jProjectRepository) GetAll(ctx context.Context, namespaceID model.ID, offset, limit int) ([]*model.Project, error) {
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

	projects, err := Neo4jExecuteReadAndReadAll(ctx, r.db, cypher, params, r.scan("p", "d", "t", "i"))
	if err != nil {
		return nil, errors.Join(ErrProjectRead, err)
	}

	return projects, nil
}

func (r *Neo4jProjectRepository) Update(ctx context.Context, id model.ID, patch map[string]any) (*model.Project, error) {
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

	project, err := Neo4jExecuteWriteAndReadSingle(ctx, r.db, cypher, params, r.scan("p", "d", "t", "i"))
	if err != nil {
		return nil, errors.Join(ErrProjectUpdate, err)
	}

	return project, nil
}

func (r *Neo4jProjectRepository) Delete(ctx context.Context, id model.ID) error {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.ProjectRepository/Delete")
	defer span.End()

	cypher := `MATCH (p:` + id.Label() + ` {id: $id}) DETACH DELETE p`
	params := map[string]any{
		"id": id.String(),
	}

	if err := Neo4jExecuteWriteAndConsume(ctx, r.db, cypher, params); err != nil {
		return errors.Join(ErrProjectDelete, err)
	}

	return nil
}

// NewNeo4jProjectRepository creates a new project neo4jBaseRepository.
func NewNeo4jProjectRepository(opts ...Neo4jRepositoryOption) (*Neo4jProjectRepository, error) {
	baseRepo, err := newNeo4jRepository(opts...)
	if err != nil {
		return nil, err
	}

	return &Neo4jProjectRepository{
		neo4jBaseRepository: baseRepo,
	}, nil
}

func clearProjectsPattern(ctx context.Context, r *redisBaseRepository, pattern ...string) error {
	return r.DeletePattern(ctx, composeCacheKey(model.ResourceTypeProject.String(), pattern))
}

func clearProjectsKey(ctx context.Context, r *redisBaseRepository, id model.ID) error {
	return r.Delete(ctx, composeCacheKey(model.ResourceTypeProject.String(), id.String()))
}

func clearProjectsByKey(ctx context.Context, r *redisBaseRepository, id model.ID) error {
	return clearProjectsPattern(ctx, r, "GetByKey", id.String(), "*")
}

func clearProjectsAllGetAll(ctx context.Context, r *redisBaseRepository) error {
	return clearProjectsPattern(ctx, r, "GetAll", "*")
}

func clearProjectsAllCrossCache(ctx context.Context, r *redisBaseRepository) error {
	deleteFns := []func(context.Context, *redisBaseRepository, ...string) error{
		clearNamespacesPattern,
	}

	for _, fn := range deleteFns {
		if err := fn(ctx, r, "*"); err != nil {
			return err
		}
	}

	return nil
}

// CachedProjectRepository implements caching on the
// repository.ProjectRepository.
type RedisCachedProjectRepository struct {
	cacheRepo   *redisBaseRepository
	projectRepo ProjectRepository
}

func (r *RedisCachedProjectRepository) Create(ctx context.Context, namespaceID model.ID, project *model.Project) error {
	if err := clearProjectsAllGetAll(ctx, r.cacheRepo); err != nil {
		return err
	}
	if err := clearProjectsAllCrossCache(ctx, r.cacheRepo); err != nil {
		return err
	}

	return r.projectRepo.Create(ctx, namespaceID, project)
}

func (r *RedisCachedProjectRepository) Get(ctx context.Context, id model.ID) (*model.Project, error) {
	var project *model.Project
	var err error

	key := composeCacheKey(model.ResourceTypeProject.String(), id.String())
	if err = r.cacheRepo.Get(ctx, key, &project); err != nil {
		return nil, err
	}

	if project != nil {
		return project, nil
	}

	if project, err = r.projectRepo.Get(ctx, id); err != nil {
		return nil, err
	}

	if err = r.cacheRepo.Set(ctx, key, project); err != nil {
		return nil, err
	}

	return project, nil
}

func (r *RedisCachedProjectRepository) GetByKey(ctx context.Context, key string) (*model.Project, error) {
	var project *model.Project
	var err error

	cacheKey := composeCacheKey(model.ResourceTypeProject.String(), "GetByKey", key)
	if err = r.cacheRepo.Get(ctx, cacheKey, &project); err != nil {
		return nil, err
	}

	if project != nil {
		return project, nil
	}

	if project, err = r.projectRepo.GetByKey(ctx, key); err != nil {
		return nil, err
	}

	if err = r.cacheRepo.Set(ctx, cacheKey, project); err != nil {
		return nil, err
	}

	return project, nil
}

func (r *RedisCachedProjectRepository) GetAll(ctx context.Context, namespaceID model.ID, offset, limit int) ([]*model.Project, error) {
	var projects []*model.Project
	var err error

	key := composeCacheKey(model.ResourceTypeProject.String(), "GetAll", namespaceID.String(), offset, limit)
	if err = r.cacheRepo.Get(ctx, key, &projects); err != nil {
		return nil, err
	}

	if projects != nil {
		return projects, nil
	}

	if projects, err = r.projectRepo.GetAll(ctx, namespaceID, offset, limit); err != nil {
		return nil, err
	}

	if err = r.cacheRepo.Set(ctx, key, projects); err != nil {
		return nil, err
	}

	return projects, nil
}

func (r *RedisCachedProjectRepository) Update(ctx context.Context, id model.ID, patch map[string]any) (*model.Project, error) {
	project, err := r.projectRepo.Update(ctx, id, patch)
	if err != nil {
		return nil, err
	}

	key := composeCacheKey(model.ResourceTypeProject.String(), id.String())
	if err := r.cacheRepo.Set(ctx, key, project); err != nil {
		return nil, err
	}

	if err := clearProjectsByKey(ctx, r.cacheRepo, id); err != nil {
		return nil, err
	}

	if err := clearProjectsAllGetAll(ctx, r.cacheRepo); err != nil {
		return nil, err
	}

	return project, nil
}

func (r *RedisCachedProjectRepository) Delete(ctx context.Context, id model.ID) error {
	if err := clearProjectsKey(ctx, r.cacheRepo, id); err != nil {
		return err
	}

	if err := clearProjectsByKey(ctx, r.cacheRepo, id); err != nil {
		return err
	}

	if err := clearProjectsAllGetAll(ctx, r.cacheRepo); err != nil {
		return err
	}

	if err := clearProjectsAllCrossCache(ctx, r.cacheRepo); err != nil {
		return err
	}

	return r.projectRepo.Delete(ctx, id)
}

// NewCachedProjectRepository returns a new CachedProjectRepository.
func NewCachedProjectRepository(repo ProjectRepository, opts ...RedisRepositoryOption) (*RedisCachedProjectRepository, error) {
	r, err := newRedisBaseRepository(opts...)
	if err != nil {
		return nil, err
	}

	return &RedisCachedProjectRepository{
		cacheRepo:   r,
		projectRepo: repo,
	}, nil
}

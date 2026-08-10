package repository

import (
	"context"
	"errors"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg/convert"
	"github.com/opcotech/elemo/internal/pkg/optional"
)

var (
	ErrProjectCreate = errors.New("failed to create project") // project cannot be created
	ErrProjectDelete = errors.New("failed to delete project") // project cannot be deleted
	ErrProjectRead   = errors.New("failed to read project")   // project cannot be read
	ErrProjectUpdate = errors.New("failed to update project") // project cannot be updated
)

// PartialProject represents a simplified project that can be used in lists.
type PartialProject struct {
	ID          model.ID            `json:"id"`
	Key         string              `json:"key"`
	Name        string              `json:"name"`
	Description string              `json:"description"`
	Logo        string              `json:"logo"`
	Status      model.ProjectStatus `json:"status"`
}

// Project represents a project persisted by the repository.
type Project struct {
	ID          model.ID            `json:"id"`
	Key         string              `json:"key"`
	Name        string              `json:"name"`
	Description string              `json:"description"`
	Logo        string              `json:"logo"`
	Status      model.ProjectStatus `json:"status"`
	Teams       []model.ID          `json:"teams"`
	Documents   []*PartialDocument  `json:"documents"`
	Issues      []model.ID          `json:"issues"`
	CreatedAt   *time.Time          `json:"created_at"`
	UpdatedAt   *time.Time          `json:"updated_at"`
}

// CreateProjectOpts holds the data required to create a project.
type CreateProjectOpts struct {
	NamespaceID model.ID
	CreatorID   model.ID
	Key         string
	Name        string
	Description string
	Logo        string
	Status      model.ProjectStatus
}

// UpdateProjectOpts holds the fields that can be updated on a project.
// Undefined fields (Defined == false) are left unchanged.
type UpdateProjectOpts struct {
	Key         optional.Optional[string]
	Name        optional.Optional[string]
	Description optional.Optional[string]
	Logo        optional.Optional[string]
	Status      optional.Optional[model.ProjectStatus]
}

// patch builds a Neo4j property map from defined optional fields.
func (o UpdateProjectOpts) patch() map[string]any {
	p := make(map[string]any)

	if o.Key.Defined {
		p["key"] = *o.Key.Value
	}
	if o.Name.Defined {
		p["name"] = *o.Name.Value
	}
	if o.Description.Defined {
		p["description"] = *o.Description.Value
	}
	if o.Logo.Defined {
		p["logo"] = *o.Logo.Value
	}
	if o.Status.Defined {
		p["status"] = o.Status.Value.String()
	}

	return p
}

// scanPartialProjects scans the record into partial projects.
func scanPartialProjects(record *neo4j.Record, key string) ([]*PartialProject, error) {
	projectsVal, err := Neo4jParseValueFromRecord[[]any](record, key)
	if err != nil {
		projectsVal = []any{}
	}

	projects := make([]*PartialProject, 0, len(projectsVal))
	for _, pVal := range projectsVal {
		if pVal == nil {
			return nil, err
		}
		pNode, ok := pVal.(neo4j.Node)
		if !ok {
			return nil, err
		}

		projectID, err := model.NewIDFromString(pNode.GetProperties()["id"].(string), model.ResourceTypeProject.String())
		if err != nil {
			return nil, err
		}

		var tempProject struct {
			Key         string `json:"key"`
			Name        string `json:"name"`
			Description string `json:"description"`
			Logo        string `json:"logo"`
			Status      string `json:"status"`
		}
		if err := Neo4jScanIntoStruct(&pNode, &tempProject, []string{"id"}); err != nil {
			return nil, err
		}

		var status model.ProjectStatus
		if err := status.UnmarshalText([]byte(tempProject.Status)); err != nil {
			return nil, err
		}

		projects = append(projects, &PartialProject{
			ID:          projectID,
			Key:         tempProject.Key,
			Name:        tempProject.Name,
			Description: tempProject.Description,
			Logo:        tempProject.Logo,
			Status:      status,
		})
	}

	return projects, nil
}

//go:generate go tool mockgen -source=project.go -destination=project_mock_gen.go -package=repository -mock_names "ProjectRepository=MockProjectRepository"
type ProjectRepository interface {
	Create(ctx context.Context, opts CreateProjectOpts) (*Project, error)
	Get(ctx context.Context, id model.ID) (*Project, error)
	GetByKey(ctx context.Context, key string) (*Project, error)
	GetAll(ctx context.Context, namespaceID model.ID, offset, limit int) ([]*Project, error)
	Update(ctx context.Context, id model.ID, opts UpdateProjectOpts) (*Project, error)
	Delete(ctx context.Context, id model.ID) error
}

// Neo4jProjectRepository is a repository for managing projects.
type Neo4jProjectRepository struct {
	*neo4jBaseRepository
}

func (r *Neo4jProjectRepository) scan(pp, dp, tp, ip string) func(rec *neo4j.Record) (*Project, error) {
	return func(rec *neo4j.Record) (*Project, error) {
		p := new(Project)

		val, _, err := neo4j.GetRecordValue[neo4j.Node](rec, pp)
		if err != nil {
			return nil, err
		}

		if err := Neo4jScanIntoStruct(&val, &p, []string{"id"}); err != nil {
			return nil, err
		}

		p.ID, _ = model.NewIDFromString(val.GetProperties()["id"].(string), model.ResourceTypeProject.String())

		if p.Documents, err = scanPartialDocuments(rec, dp); err != nil {
			return nil, err
		}

		if p.Teams, err = Neo4jParseIDsFromRecord(rec, tp, model.ResourceTypeRole.String()); err != nil {
			return nil, err
		}

		if p.Issues, err = Neo4jParseIDsFromRecord(rec, ip, model.ResourceTypeIssue.String()); err != nil {
			return nil, err
		}

		return p, nil
	}
}

func (r *Neo4jProjectRepository) Create(ctx context.Context, opts CreateProjectOpts) (*Project, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.ProjectRepository/Create")
	defer span.End()

	createdAt := convert.ToPointer(time.Now().UTC())

	project := &Project{
		ID:          model.MustNewID(model.ResourceTypeProject),
		Key:         opts.Key,
		Name:        opts.Name,
		Description: opts.Description,
		Logo:        opts.Logo,
		Status:      opts.Status,
		Teams:       make([]model.ID, 0),
		Documents:   make([]*PartialDocument, 0),
		Issues:      make([]model.ID, 0),
		CreatedAt:   createdAt,
		UpdatedAt:   nil,
	}

	cypher := `
	MATCH (u:` + opts.CreatorID.Label() + ` {id: $creator_id})
	MATCH (n:` + opts.NamespaceID.Label() + ` {id: $namespace_id})
	CREATE
		(p:` + project.ID.Label() + ` {
			id: $id, key: $key, name: $name, description: $description, logo: $logo, status: $status,
			created_at: datetime($created_at)
		}),
		(n)-[:` + EdgeKindHasProject.String() + `]->(p),
		(u)-[:` + EdgeKindHasPermission.String() + ` {id: $perm_id, kind: $perm_kind, created_at: datetime($created_at)}]->(p)`

	params := map[string]any{
		"id":           project.ID.String(),
		"key":          project.Key,
		"name":         project.Name,
		"description":  project.Description,
		"logo":         project.Logo,
		"status":       project.Status.String(),
		"created_at":   createdAt.Format(time.RFC3339Nano),
		"creator_id":   opts.CreatorID.String(),
		"namespace_id": opts.NamespaceID.String(),
		"perm_id":      model.NewRawID(),
		"perm_kind":    model.PermissionKindAll.String(),
	}

	if err := Neo4jExecuteWriteAndConsume(ctx, r.db, cypher, params); err != nil {
		return nil, errors.Join(ErrProjectCreate, err)
	}

	return project, nil
}

func (r *Neo4jProjectRepository) Get(ctx context.Context, id model.ID) (*Project, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.ProjectRepository/Get")
	defer span.End()

	cypher := `
	MATCH (p:` + id.Label() + ` {id: $id})
	OPTIONAL MATCH (d:` + model.ResourceTypeDocument.String() + `)-[:` + EdgeKindBelongsTo.String() + `]->(p)
	OPTIONAL MATCH (p)-[:` + EdgeKindHasTeam.String() + `]->(t:` + model.ResourceTypeRole.String() + `)
	OPTIONAL MATCH (p)<-[:` + EdgeKindBelongsTo.String() + `]-(i:` + model.ResourceTypeIssue.String() + `)
	RETURN p, collect(DISTINCT d) AS d, collect(DISTINCT t.id) AS t, collect(DISTINCT i.id) AS i`

	params := map[string]any{
		"id": id.String(),
	}

	project, err := Neo4jExecuteReadAndReadSingle(ctx, r.db, cypher, params, r.scan("p", "d", "t", "i"))
	if err != nil {
		return nil, errors.Join(ErrProjectRead, err)
	}

	return project, nil
}

func (r *Neo4jProjectRepository) GetByKey(ctx context.Context, key string) (*Project, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.ProjectRepository/GetByKey")
	defer span.End()

	cypher := `
	MATCH (p:` + model.ResourceTypeProject.String() + ` {key: $key})
	OPTIONAL MATCH (d:` + model.ResourceTypeDocument.String() + `)-[:` + EdgeKindBelongsTo.String() + `]->(p)
	OPTIONAL MATCH (p)-[:` + EdgeKindHasTeam.String() + `]->(t:` + model.ResourceTypeRole.String() + `)
	OPTIONAL MATCH (p)<-[:` + EdgeKindBelongsTo.String() + `]-(i:` + model.ResourceTypeIssue.String() + `)
	RETURN p, collect(DISTINCT d) AS d, collect(DISTINCT t.id) AS t, collect(DISTINCT i.id) AS i`

	params := map[string]any{
		"key": key,
	}

	project, err := Neo4jExecuteReadAndReadSingle(ctx, r.db, cypher, params, r.scan("p", "d", "t", "i"))
	if err != nil {
		return nil, errors.Join(ErrProjectRead, err)
	}

	return project, nil
}

func (r *Neo4jProjectRepository) GetAll(ctx context.Context, namespaceID model.ID, offset, limit int) ([]*Project, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.ProjectRepository/GetAll")
	defer span.End()

	cypher := `
	MATCH (:` + namespaceID.Label() + ` {id: $namespace_id})-[:` + EdgeKindHasProject.String() + `]->(p)
	OPTIONAL MATCH (d:` + model.ResourceTypeDocument.String() + `)-[:` + EdgeKindBelongsTo.String() + `]->(p)
	OPTIONAL MATCH (p)-[:` + EdgeKindHasTeam.String() + `]->(t:` + model.ResourceTypeRole.String() + `)
	OPTIONAL MATCH (p)<-[:` + EdgeKindBelongsTo.String() + `]-(i:` + model.ResourceTypeIssue.String() + `)
	RETURN p, collect(DISTINCT d) AS d, collect(DISTINCT t.id) AS t, collect(DISTINCT i.id) AS i
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

func (r *Neo4jProjectRepository) Update(ctx context.Context, id model.ID, opts UpdateProjectOpts) (*Project, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.ProjectRepository/Update")
	defer span.End()

	cypher := `
	MATCH (p:` + id.Label() + ` {id: $id})
	SET p += $patch, p.updated_at = datetime()
	WITH p
	OPTIONAL MATCH (d:` + model.ResourceTypeDocument.String() + `)-[:` + EdgeKindBelongsTo.String() + `]->(p)
	OPTIONAL MATCH (p)-[:` + EdgeKindHasTeam.String() + `]->(t:` + model.ResourceTypeRole.String() + `)
	OPTIONAL MATCH (p)<-[:` + EdgeKindBelongsTo.String() + `]-(i:` + model.ResourceTypeIssue.String() + `)
	RETURN p, collect(DISTINCT d) AS d, collect(DISTINCT t.id) AS t, collect(DISTINCT i.id) AS i`

	params := map[string]any{
		"id":    id.String(),
		"patch": opts.patch(),
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

// RedisCachedProjectRepository implements caching on the ProjectRepository.
type RedisCachedProjectRepository struct {
	cacheRepo   *redisBaseRepository
	projectRepo ProjectRepository
}

func (r *RedisCachedProjectRepository) Create(ctx context.Context, opts CreateProjectOpts) (*Project, error) {
	if err := clearProjectsAllGetAll(ctx, r.cacheRepo); err != nil {
		return nil, err
	}
	if err := clearProjectsAllCrossCache(ctx, r.cacheRepo); err != nil {
		return nil, err
	}

	return r.projectRepo.Create(ctx, opts)
}

func (r *RedisCachedProjectRepository) Get(ctx context.Context, id model.ID) (*Project, error) {
	var project *Project
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

func (r *RedisCachedProjectRepository) GetByKey(ctx context.Context, key string) (*Project, error) {
	var project *Project
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

func (r *RedisCachedProjectRepository) GetAll(ctx context.Context, namespaceID model.ID, offset, limit int) ([]*Project, error) {
	var projects []*Project
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

func (r *RedisCachedProjectRepository) Update(ctx context.Context, id model.ID, opts UpdateProjectOpts) (*Project, error) {
	project, err := r.projectRepo.Update(ctx, id, opts)
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

	// Namespace Get embeds PartialProject rows; clear so list status/name stay fresh.
	if err := clearProjectsAllCrossCache(ctx, r.cacheRepo); err != nil {
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

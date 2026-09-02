package repository

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/neo4j/neo4j-go-driver/v6/neo4j"

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
	ID            model.ID            `json:"id"`
	Key           string              `json:"key"`
	Name          string              `json:"name"`
	Description   string              `json:"description"`
	Logo          string              `json:"logo"`
	Status        model.ProjectStatus `json:"status"`
	Teams         []model.ID          `json:"teams"`
	DocumentCount *int64              `json:"document_count"`
	IssueCount    *int64              `json:"issue_count"`
	CreatedAt     *time.Time          `json:"created_at"`
	UpdatedAt     *time.Time          `json:"updated_at"`
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
	Name        optional.Optional[string]
	Description optional.Optional[string]
	Logo        optional.Optional[string]
	Status      optional.Optional[model.ProjectStatus]
}

// patch builds a Neo4j property map from defined optional fields.
func (o UpdateProjectOpts) patch() map[string]any {
	p := make(map[string]any)

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

//go:generate go tool mockgen -source=project.go -destination=mock/mock_project_gen.go -package=mockrepo
type ProjectRepository interface {
	Create(ctx context.Context, opts CreateProjectOpts) (*Project, error)
	Get(ctx context.Context, id model.ID, proj ProjectProjection) (*Project, error)
	GetByKey(ctx context.Context, namespaceID model.ID, key string, proj ProjectProjection) (*Project, error)
	ListForNamespace(ctx context.Context, query ProjectListQuery) (Page[*Project], error)
	Update(ctx context.Context, id model.ID, opts UpdateProjectOpts, proj ProjectProjection) (*Project, error)
	Delete(ctx context.Context, id model.ID) error
}

// Neo4jProjectRepository is a repository for managing projects.
type Neo4jProjectRepository struct {
	*neo4jBaseRepository
}

func decodeProjectRecord(record *neo4j.Record, proj ProjectProjection) (*Project, error) {
	node, err := Neo4jRecordNode(record, "p")
	if err != nil {
		return nil, err
	}

	project := new(Project)
	if err := Neo4jScanNodeScalars(node, project, []string{"id", "issue_count", "document_count"}); err != nil {
		return nil, err
	}

	project.ID, err = Neo4jDecodeID(node, model.ResourceTypeProject)
	if err != nil {
		return nil, err
	}
	if proj.Teams {
		project.Teams = make([]model.ID, 0)
	}
	if proj.IssueCount {
		issueCount, err := Neo4jParseValueFromRecord[int64](record, "issue_count")
		if err != nil {
			return nil, err
		}
		project.IssueCount = convert.ToPointer(issueCount)
	}
	if proj.DocumentCount {
		documentCount, err := Neo4jParseValueFromRecord[int64](record, "document_count")
		if err != nil {
			return nil, err
		}
		project.DocumentCount = convert.ToPointer(documentCount)
	}

	return project, nil
}

func (r *Neo4jProjectRepository) applyProjectLoaders(
	ctx context.Context,
	tx neo4j.ManagedTransaction,
	plan QueryPlan,
	projects []*Project,
) error {
	if len(projects) == 0 || len(plan.Loaders) == 0 {
		return nil
	}

	ids := make([]string, 0, len(projects))
	projectsByID := make(map[string]*Project, len(projects))
	for _, project := range projects {
		if project == nil {
			continue
		}
		id := project.ID.String()
		ids = append(ids, id)
		projectsByID[id] = project
	}
	if len(ids) == 0 {
		return nil
	}

	for _, loader := range plan.Loaders {
		query := loader
		query.Params = cloneParams(loader.Params)
		query.Params["ids"] = ids

		switch {
		case strings.HasSuffix(loader.Name, ".teams"):
			rows, _, err := Neo4jRunQuery(ctx, tx, query, func(record *neo4j.Record) (struct {
				projectID string
				teamIDs   []model.ID
			}, error,
			) {
				projectID, err := Neo4jParseValueFromRecord[string](record, "project_id")
				if err != nil {
					return struct {
						projectID string
						teamIDs   []model.ID
					}{}, err
				}
				teamIDs, err := Neo4jRecordIDs(record, "team_ids", model.ResourceTypeRole)
				if err != nil {
					return struct {
						projectID string
						teamIDs   []model.ID
					}{}, err
				}
				return struct {
					projectID string
					teamIDs   []model.ID
				}{
					projectID: projectID,
					teamIDs:   teamIDs,
				}, nil
			})
			if err != nil {
				return err
			}
			for _, row := range rows {
				project, ok := projectsByID[row.projectID]
				if !ok {
					continue
				}
				project.Teams = row.teamIDs
			}
		default:
			return ErrQueryCompile
		}
	}

	return nil
}

func (r *Neo4jProjectRepository) readProjectPlan(
	ctx context.Context,
	plan QueryPlan,
	proj ProjectProjection,
) ([]*Project, error) {
	projects := make([]*Project, 0)

	if err := Neo4jExecuteReadPlan(ctx, r.db, plan, func(tx neo4j.ManagedTransaction) error {
		rootProjects, _, err := Neo4jRunQuery(ctx, tx, plan.Root, func(record *neo4j.Record) (*Project, error) {
			return decodeProjectRecord(record, proj)
		})
		if err != nil {
			return err
		}

		projects = rootProjects
		return r.applyProjectLoaders(ctx, tx, plan, projects)
	}); err != nil {
		return nil, err
	}

	return projects, nil
}

func (r *Neo4jProjectRepository) Create(ctx context.Context, opts CreateProjectOpts) (*Project, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.ProjectRepository/Create")
	defer span.End()

	createdAt := time.Now().UTC()
	id := model.MustNewID(model.ResourceTypeProject)

	cypher := `
	MATCH (u:` + opts.CreatorID.Label() + ` {id: $creator_id})
	MATCH (n:` + opts.NamespaceID.Label() + ` {id: $namespace_id})
	CREATE
		(p:` + id.Label() + ` {
			id: $id, key: $key, namespace_id: $namespace_id, name: $name, description: $description, logo: $logo, status: $status,
			next_issue_id: $next_issue_id, created_at: datetime($created_at)
		}),
		(n)-[:` + EdgeKindHasProject.String() + `]->(p),
		(p)-[:` + EdgeKindInScopeOf.String() + ` {id: $scope_id, created_at: datetime($created_at)}]->(n)`

	params := map[string]any{
		"id":            id.String(),
		"key":           opts.Key,
		"name":          opts.Name,
		"description":   opts.Description,
		"logo":          opts.Logo,
		"status":        opts.Status.String(),
		"next_issue_id": 0,
		"created_at":    createdAt.Format(time.RFC3339Nano),
		"creator_id":    opts.CreatorID.String(),
		"namespace_id":  opts.NamespaceID.String(),
		"scope_id":      model.NewRawID(),
	}

	if err := Neo4jExecuteWriteAndConsume(ctx, r.db, cypher, params); err != nil {
		return nil, errors.Join(ErrProjectCreate, mapUniquenessError(err))
	}

	project, err := r.Get(ctx, id, ProjectDetailProjection())
	if err != nil {
		return nil, errors.Join(ErrProjectCreate, err)
	}

	return project, nil
}

func (r *Neo4jProjectRepository) Get(ctx context.Context, id model.ID, proj ProjectProjection) (*Project, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.ProjectRepository/Get")
	defer span.End()

	plan, err := CompileQuery(ProjectGetQuery{
		ID:         id,
		Projection: proj,
	})
	if err != nil {
		return nil, errors.Join(ErrProjectRead, err)
	}

	projects, err := r.readProjectPlan(ctx, plan, proj)
	if err != nil {
		return nil, errors.Join(ErrProjectRead, err)
	}
	if len(projects) == 0 {
		return nil, errors.Join(ErrProjectRead, ErrNotFound)
	}
	if len(projects) > 1 {
		return nil, errors.Join(ErrProjectRead, ErrMalformedResult)
	}

	return projects[0], nil
}

func (r *Neo4jProjectRepository) GetByKey(ctx context.Context, namespaceID model.ID, key string, proj ProjectProjection) (*Project, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.ProjectRepository/GetByKey")
	defer span.End()

	plan, err := CompileQuery(ProjectGetByKeyQuery{
		NamespaceID: namespaceID,
		Key:         key,
		Projection:  proj,
	})
	if err != nil {
		return nil, errors.Join(ErrProjectRead, err)
	}

	projects, err := r.readProjectPlan(ctx, plan, proj)
	if err != nil {
		return nil, errors.Join(ErrProjectRead, err)
	}
	if len(projects) == 0 {
		return nil, errors.Join(ErrProjectRead, ErrNotFound)
	}
	if len(projects) > 1 {
		return nil, errors.Join(ErrProjectRead, ErrMalformedResult)
	}

	return projects[0], nil
}

func (r *Neo4jProjectRepository) ListForNamespace(ctx context.Context, query ProjectListQuery) (Page[*Project], error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.ProjectRepository/ListForNamespace")
	defer span.End()

	normalizedPage, err := query.Page.Normalize()
	if err != nil {
		return Page[*Project]{}, errors.Join(ErrProjectRead, err)
	}

	query.Page = normalizedPage
	plan, err := CompileQuery(query)
	if err != nil {
		return Page[*Project]{}, errors.Join(ErrProjectRead, err)
	}

	projects, err := r.readProjectPlan(ctx, plan, query.Projection)
	if err != nil {
		return Page[*Project]{}, errors.Join(ErrProjectRead, err)
	}

	return PaginateSlice(projects, normalizedPage.Size, func(project *Project) model.ID {
		return project.ID
	})
}

func (r *Neo4jProjectRepository) Update(ctx context.Context, id model.ID, opts UpdateProjectOpts, _ ProjectProjection) (*Project, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.ProjectRepository/Update")
	defer span.End()

	cypher := `
	MATCH (p:` + id.Label() + ` {id: $id})
	SET p += $patch
	SET p.updated_at = datetime.statement()
	RETURN p.id AS id`

	params := map[string]any{
		"id":    id.String(),
		"patch": opts.patch(),
	}

	if _, err := Neo4jExecuteWriteAndReadSingle(ctx, r.db, cypher, params, func(_ *neo4j.Record) (*struct{}, error) {
		return &struct{}{}, nil
	}); err != nil {
		return nil, errors.Join(ErrProjectUpdate, err)
	}

	project, err := r.Get(ctx, id, ProjectDetailProjection())
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

func clearProjectsGet(ctx context.Context, r *redisBaseRepository, id model.ID) error {
	return clearProjectsPattern(ctx, r, "*", "Get", id.String(), "*")
}

func clearProjectsAllByKey(ctx context.Context, r *redisBaseRepository) error {
	return clearProjectsPattern(ctx, r, "*", "GetByKey", "*")
}

func clearProjectsAllList(ctx context.Context, r *redisBaseRepository) error {
	return clearProjectsPattern(ctx, r, "*", "ListForNamespace", "*")
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
	if err := clearProjectsAllList(ctx, r.cacheRepo); err != nil {
		return nil, err
	}
	if err := clearProjectsAllCrossCache(ctx, r.cacheRepo); err != nil {
		return nil, err
	}
	project, err := r.projectRepo.Create(ctx, opts)
	if err != nil {
		return nil, err
	}
	if err := bumpIssueListNamespaceGeneration(ctx, r.cacheRepo, opts.NamespaceID); err != nil {
		return nil, err
	}
	if err := bumpIssueListProjectGeneration(ctx, r.cacheRepo, project.ID); err != nil {
		return nil, err
	}
	if err := bumpIssueListProjectionEpoch(ctx, r.cacheRepo); err != nil {
		return nil, err
	}
	return project, nil
}

func (r *RedisCachedProjectRepository) Get(ctx context.Context, id model.ID, proj ProjectProjection) (*Project, error) {
	var project *Project
	var err error

	plan, err := CompileQuery(ProjectGetQuery{ID: id, Projection: proj})
	if err != nil {
		return nil, err
	}
	key := plan.CacheKey(model.ResourceTypeProject.String(), "Get", id.String())
	if err = r.cacheRepo.Get(ctx, key, &project); err != nil {
		return nil, err
	}

	if project != nil {
		return project, nil
	}

	if project, err = r.projectRepo.Get(ctx, id, proj); err != nil {
		return nil, err
	}

	if err = r.cacheRepo.Set(ctx, key, project); err != nil {
		return nil, err
	}

	return project, nil
}

func (r *RedisCachedProjectRepository) GetByKey(ctx context.Context, namespaceID model.ID, key string, proj ProjectProjection) (*Project, error) {
	var project *Project
	var err error

	plan, err := CompileQuery(ProjectGetByKeyQuery{NamespaceID: namespaceID, Key: key, Projection: proj})
	if err != nil {
		return nil, err
	}
	cacheKey := plan.CacheKey(model.ResourceTypeProject.String(), "GetByKey", namespaceID.String(), key)
	if err = r.cacheRepo.Get(ctx, cacheKey, &project); err != nil {
		return nil, err
	}

	if project != nil {
		return project, nil
	}

	if project, err = r.projectRepo.GetByKey(ctx, namespaceID, key, proj); err != nil {
		return nil, err
	}

	if err = r.cacheRepo.Set(ctx, cacheKey, project); err != nil {
		return nil, err
	}

	return project, nil
}

func (r *RedisCachedProjectRepository) ListForNamespace(ctx context.Context, query ProjectListQuery) (Page[*Project], error) {
	var projects Page[*Project]
	var err error

	normalized, err := normalizedPage(query.Page)
	if err != nil {
		return Page[*Project]{}, err
	}
	query.Page = normalized

	plan, err := CompileQuery(query)
	if err != nil {
		return Page[*Project]{}, err
	}
	key := plan.CacheKey(model.ResourceTypeProject.String(), "ListForNamespace", query.NamespaceID.String())
	if err = r.cacheRepo.Get(ctx, key, &projects); err != nil {
		return Page[*Project]{}, err
	}

	if projects.Items != nil {
		return projects, nil
	}

	if projects, err = r.projectRepo.ListForNamespace(ctx, query); err != nil {
		return Page[*Project]{}, err
	}

	if err = r.cacheRepo.Set(ctx, key, projects); err != nil {
		return Page[*Project]{}, err
	}

	return projects, nil
}

func (r *RedisCachedProjectRepository) Update(ctx context.Context, id model.ID, opts UpdateProjectOpts, proj ProjectProjection) (*Project, error) {
	project, err := r.projectRepo.Update(ctx, id, opts, proj)
	if err != nil {
		return nil, err
	}

	plan, err := CompileQuery(ProjectGetQuery{ID: id, Projection: ProjectDetailProjection()})
	if err != nil {
		return nil, err
	}
	key := plan.CacheKey(model.ResourceTypeProject.String(), "Get", id.String())
	if err := r.cacheRepo.Set(ctx, key, project); err != nil {
		return nil, err
	}

	if err := clearProjectsAllByKey(ctx, r.cacheRepo); err != nil {
		return nil, err
	}

	if err := clearProjectsAllList(ctx, r.cacheRepo); err != nil {
		return nil, err
	}

	// Namespace Get embeds PartialProject rows; clear so list status/name stay fresh.
	if err := clearProjectsAllCrossCache(ctx, r.cacheRepo); err != nil {
		return nil, err
	}
	if err := bumpIssueListProjectGeneration(ctx, r.cacheRepo, id); err != nil {
		return nil, err
	}
	if err := bumpIssueListProjectionEpoch(ctx, r.cacheRepo); err != nil {
		return nil, err
	}

	return project, nil
}

func (r *RedisCachedProjectRepository) Delete(ctx context.Context, id model.ID) error {
	if err := clearProjectsGet(ctx, r.cacheRepo, id); err != nil {
		return err
	}

	if err := clearProjectsAllByKey(ctx, r.cacheRepo); err != nil {
		return err
	}

	if err := clearProjectsAllList(ctx, r.cacheRepo); err != nil {
		return err
	}

	if err := clearProjectsAllCrossCache(ctx, r.cacheRepo); err != nil {
		return err
	}
	if err := bumpIssueListProjectGeneration(ctx, r.cacheRepo, id); err != nil {
		return err
	}
	if err := bumpIssueListProjectionEpoch(ctx, r.cacheRepo); err != nil {
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

package repository

import (
	"context"
	"errors"
	"time"

	"github.com/neo4j/neo4j-go-driver/v6/neo4j"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg/convert"
	"github.com/opcotech/elemo/internal/pkg/optional"
)

var (
	ErrNamespaceCreate = errors.New("failed to create namespace") // the namespace could not be created
	ErrNamespaceDelete = errors.New("failed to delete namespace") // the namespace could not be deleted
	ErrNamespaceRead   = errors.New("failed to read namespace")   // the namespace could not be retrieved
	ErrNamespaceUpdate = errors.New("failed to update namespace") // the namespace could not be updated
)

// PartialNamespace is a lean namespace used on issue reads.
type PartialNamespace struct {
	ID   model.ID `json:"id"`
	Name string   `json:"name"`
}

// Namespace represents a namespace persisted by the repository.
type Namespace struct {
	ID            model.ID   `json:"id"`
	Name          string     `json:"name"`
	Description   string     `json:"description"`
	ProjectCount  *int64     `json:"project_count"`
	DocumentCount *int64     `json:"document_count"`
	CreatedAt     *time.Time `json:"created_at"`
	UpdatedAt     *time.Time `json:"updated_at"`
}

// CreateNamespaceOpts holds the data required to create a namespace.
type CreateNamespaceOpts struct {
	Name        string
	Description string
	CreatorID   model.ID
	OrgID       model.ID
}

// UpdateNamespaceOpts holds the fields that can be updated on a namespace.
// Undefined fields (Defined == false) are left unchanged.
type UpdateNamespaceOpts struct {
	Name        optional.Optional[string]
	Description optional.Optional[string]
}

// patch builds a Neo4j property map from defined optional fields.
func (o UpdateNamespaceOpts) patch() map[string]any {
	p := make(map[string]any)

	if o.Name.Defined {
		p["name"] = *o.Name.Value
	}
	if o.Description.Defined {
		if o.Description.Value == nil {
			p["description"] = nil
		} else {
			p["description"] = *o.Description.Value
		}
	}

	return p
}

//go:generate go tool mockgen -source=namespace.go -destination=namespace_mock_gen.go -package=repository -mock_names "NamespaceRepository=MockNamespaceRepository"
type NamespaceRepository interface {
	Create(ctx context.Context, opts CreateNamespaceOpts) (*Namespace, error)
	Get(ctx context.Context, id model.ID, proj NamespaceProjection) (*Namespace, error)
	List(ctx context.Context, orgID model.ID, page CursorPage, proj NamespaceProjection) (Page[*Namespace], error)
	Update(ctx context.Context, id model.ID, opts UpdateNamespaceOpts) (*Namespace, error)
	Delete(ctx context.Context, id model.ID) error
}

// Neo4jNamespaceRepository is a repository for managing namespaces.
type Neo4jNamespaceRepository struct {
	*neo4jBaseRepository
}

func (r *Neo4jNamespaceRepository) scan(proj NamespaceProjection) func(rec *neo4j.Record) (*Namespace, error) {
	return func(rec *neo4j.Record) (*Namespace, error) {
		node, err := Neo4jRecordNode(rec, "ns")
		if err != nil {
			return nil, err
		}

		ns := new(Namespace)
		if err := Neo4jScanNodeScalars(node, ns, []string{"id", "project_count", "document_count"}); err != nil {
			return nil, err
		}

		ns.ID, err = Neo4jDecodeID(node, model.ResourceTypeNamespace)
		if err != nil {
			return nil, err
		}
		if proj.ProjectCount {
			projectCount, err := Neo4jParseValueFromRecord[int64](rec, "project_count")
			if err != nil {
				return nil, err
			}
			ns.ProjectCount = convert.ToPointer(projectCount)
		}
		if proj.DocumentCount {
			documentCount, err := Neo4jParseValueFromRecord[int64](rec, "document_count")
			if err != nil {
				return nil, err
			}
			ns.DocumentCount = convert.ToPointer(documentCount)
		}

		return ns, nil
	}
}

func (r *Neo4jNamespaceRepository) Create(ctx context.Context, opts CreateNamespaceOpts) (*Namespace, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.NamespaceRepository/Create")
	defer span.End()

	createdAt := time.Now().UTC()
	id := model.MustNewID(model.ResourceTypeNamespace)

	cypher := `
	MATCH (u:` + opts.CreatorID.Label() + ` {id: $creator_id})
	MATCH (org:` + opts.OrgID.Label() + ` {id: $org_id})
	CREATE (ns:` + id.Label() + ` {id: $id, name: $name, description: $description, created_at: datetime($created_at)}),
		(org)-[:` + EdgeKindHasNamespace.String() + ` {id: $has_ns_id, created_at: datetime($created_at)}]->(ns),
		(u)-[:` + EdgeKindHasPermission.String() + ` {id: $perm_id, kind: $perm_kind, created_at: datetime($created_at)}]->(ns)`

	params := map[string]any{
		"id":          id.String(),
		"name":        opts.Name,
		"description": opts.Description,
		"created_at":  createdAt.Format(time.RFC3339Nano),
		"creator_id":  opts.CreatorID.String(),
		"org_id":      opts.OrgID.String(),
		"has_ns_id":   model.NewRawID(),
		"perm_id":     model.NewRawID(),
		"perm_kind":   model.PermissionKindAll.String(),
	}

	if err := Neo4jExecuteWriteAndConsume(ctx, r.db, cypher, params); err != nil {
		return nil, errors.Join(ErrNamespaceCreate, err)
	}

	return r.Get(ctx, id, NamespaceDetailProjection())
}

func (r *Neo4jNamespaceRepository) Get(ctx context.Context, id model.ID, proj NamespaceProjection) (*Namespace, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.NamespaceRepository/Get")
	defer span.End()

	plan, err := CompileQuery(NamespaceGetQuery{
		ID:         id,
		Projection: proj,
	})
	if err != nil {
		return nil, errors.Join(ErrNamespaceRead, err)
	}

	var namespace *Namespace
	err = Neo4jExecuteReadPlan(ctx, r.db, plan, func(tx neo4j.ManagedTransaction) error {
		var runErr error
		namespace, _, runErr = Neo4jRunQuerySingle(ctx, tx, plan.Root, r.scan(proj))
		if runErr != nil {
			return runErr
		}
		return nil
	})
	if err != nil {
		return nil, errors.Join(ErrNamespaceRead, err)
	}

	return namespace, nil
}

func (r *Neo4jNamespaceRepository) List(ctx context.Context, orgID model.ID, page CursorPage, proj NamespaceProjection) (Page[*Namespace], error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.NamespaceRepository/List")
	defer span.End()

	normalized, err := page.Normalize()
	if err != nil {
		return Page[*Namespace]{}, errors.Join(ErrNamespaceRead, err)
	}
	plan, err := CompileQuery(NamespaceListQuery{
		OrgID:      orgID,
		Page:       normalized,
		Order:      SortDirectionDesc,
		Projection: proj,
	})
	if err != nil {
		return Page[*Namespace]{}, errors.Join(ErrNamespaceRead, err)
	}

	namespaces := make([]*Namespace, 0)
	err = Neo4jExecuteReadPlan(ctx, r.db, plan, func(tx neo4j.ManagedTransaction) error {
		var runErr error
		namespaces, _, runErr = Neo4jRunQuery(ctx, tx, plan.Root, r.scan(proj))
		if runErr != nil {
			return runErr
		}
		return nil
	})
	if err != nil {
		return Page[*Namespace]{}, errors.Join(ErrNamespaceRead, err)
	}

	return PaginateSlice(namespaces, normalized.Size, func(namespace *Namespace) model.ID {
		return namespace.ID
	})
}

func (r *Neo4jNamespaceRepository) Update(ctx context.Context, id model.ID, opts UpdateNamespaceOpts) (*Namespace, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.NamespaceRepository/Update")
	defer span.End()

	cypher := `
	MATCH (ns:` + id.Label() + ` {id: $id})
	SET ns += $patch, ns.updated_at = $updated_at
	RETURN ns.id AS id`

	params := map[string]any{
		"id":         id.String(),
		"patch":      opts.patch(),
		"updated_at": time.Now().UTC().Format(time.RFC3339Nano),
	}

	_, err := Neo4jExecuteWriteAndReadSingle(ctx, r.db, cypher, params, func(_ *neo4j.Record) (*struct{}, error) {
		return &struct{}{}, nil
	})
	if err != nil {
		return nil, errors.Join(ErrNamespaceUpdate, err)
	}

	return r.Get(ctx, id, NamespaceDetailProjection())
}

func (r *Neo4jNamespaceRepository) Delete(ctx context.Context, id model.ID) error {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.NamespaceRepository/Delete")
	defer span.End()

	cypher := `
	MATCH (ns:` + id.Label() + ` {id: $id}) DETACH DELETE ns`

	params := map[string]any{
		"id": id.String(),
	}

	if err := Neo4jExecuteWriteAndConsume(ctx, r.db, cypher, params); err != nil {
		return errors.Join(ErrNamespaceDelete, err)
	}

	return nil
}

// NewNeo4jNamespaceRepository creates a new namespace neo4jBaseRepository.
func NewNeo4jNamespaceRepository(opts ...Neo4jRepositoryOption) (*Neo4jNamespaceRepository, error) {
	baseRepo, err := newNeo4jRepository(opts...)
	if err != nil {
		return nil, err
	}

	return &Neo4jNamespaceRepository{
		neo4jBaseRepository: baseRepo,
	}, nil
}

func clearNamespacesPattern(ctx context.Context, r *redisBaseRepository, pattern ...string) error {
	return r.DeletePattern(ctx, composeCacheKey(model.ResourceTypeNamespace.String(), pattern))
}

func clearNamespacesKey(ctx context.Context, r *redisBaseRepository, id model.ID) error {
	return clearNamespacesPattern(ctx, r, "Get", id.String(), "*")
}

func clearNamespacesAllLists(ctx context.Context, r *redisBaseRepository) error {
	return clearNamespacesPattern(ctx, r, "List", "*", "*", "*", "*")
}

func clearNamespaceAllCrossCache(ctx context.Context, r *redisBaseRepository) error {
	deleteFns := []func(context.Context, *redisBaseRepository, ...string) error{
		clearOrganizationsPattern,
	}

	for _, fn := range deleteFns {
		if err := fn(ctx, r, "*"); err != nil {
			return err
		}
	}

	return nil
}

// RedisCachedNamespaceRepository implements caching on the NamespaceRepository.
type RedisCachedNamespaceRepository struct {
	cacheRepo     *redisBaseRepository
	namespaceRepo NamespaceRepository
}

func (r *RedisCachedNamespaceRepository) Create(ctx context.Context, opts CreateNamespaceOpts) (*Namespace, error) {
	if err := clearNamespacesAllLists(ctx, r.cacheRepo); err != nil {
		return nil, err
	}
	if err := clearNamespaceAllCrossCache(ctx, r.cacheRepo); err != nil {
		return nil, err
	}

	return r.namespaceRepo.Create(ctx, opts)
}

func (r *RedisCachedNamespaceRepository) Get(ctx context.Context, id model.ID, proj NamespaceProjection) (*Namespace, error) {
	var namespace *Namespace
	var err error

	key := composeCacheKey(model.ResourceTypeNamespace.String(), "Get", id.String(), projectionCacheValue(proj))
	if err = r.cacheRepo.Get(ctx, key, &namespace); err != nil {
		return nil, err
	}

	if namespace != nil {
		return namespace, nil
	}

	if namespace, err = r.namespaceRepo.Get(ctx, id, proj); err != nil {
		return nil, err
	}

	if err = r.cacheRepo.Set(ctx, key, namespace); err != nil {
		return nil, err
	}

	return namespace, nil
}

func (r *RedisCachedNamespaceRepository) List(ctx context.Context, orgID model.ID, page CursorPage, proj NamespaceProjection) (Page[*Namespace], error) {
	var namespaces Page[*Namespace]
	var err error

	normalized, err := normalizedPage(page)
	if err != nil {
		return Page[*Namespace]{}, err
	}

	key := composeCacheKey(model.ResourceTypeNamespace.String(), "List", orgID.String(), projectionCacheValue(proj), pageTokenValue(normalized.Token), normalized.Size)
	if err = r.cacheRepo.Get(ctx, key, &namespaces); err != nil {
		return Page[*Namespace]{}, err
	}

	if namespaces.Items != nil {
		return namespaces, nil
	}

	namespaces, err = r.namespaceRepo.List(ctx, orgID, normalized, proj)
	if err != nil {
		return Page[*Namespace]{}, err
	}

	if err = r.cacheRepo.Set(ctx, key, namespaces); err != nil {
		return Page[*Namespace]{}, err
	}

	return namespaces, nil
}

func (r *RedisCachedNamespaceRepository) Update(ctx context.Context, id model.ID, opts UpdateNamespaceOpts) (*Namespace, error) {
	namespace, err := r.namespaceRepo.Update(ctx, id, opts)
	if err != nil {
		return nil, err
	}

	key := composeCacheKey(model.ResourceTypeNamespace.String(), "Get", id.String(), projectionCacheValue(NamespaceDetailProjection()))
	if err = r.cacheRepo.Set(ctx, key, namespace); err != nil {
		return nil, err
	}

	if err := clearNamespacesAllLists(ctx, r.cacheRepo); err != nil {
		return nil, err
	}

	return namespace, nil
}

func (r *RedisCachedNamespaceRepository) Delete(ctx context.Context, id model.ID) error {
	if err := clearNamespacesKey(ctx, r.cacheRepo, id); err != nil {
		return err
	}

	if err := clearNamespacesAllLists(ctx, r.cacheRepo); err != nil {
		return err
	}

	if err := clearNamespaceAllCrossCache(ctx, r.cacheRepo); err != nil {
		return err
	}

	return r.namespaceRepo.Delete(ctx, id)
}

// NewCachedNamespaceRepository returns a new CachedNamespaceRepository.
func NewCachedNamespaceRepository(repo NamespaceRepository, opts ...RedisRepositoryOption) (*RedisCachedNamespaceRepository, error) {
	r, err := newRedisBaseRepository(opts...)
	if err != nil {
		return nil, err
	}

	return &RedisCachedNamespaceRepository{
		cacheRepo:     r,
		namespaceRepo: repo,
	}, nil
}

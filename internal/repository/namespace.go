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
	ErrNamespaceCreate = errors.New("failed to create namespace") // the namespace could not be created
	ErrNamespaceDelete = errors.New("failed to delete namespace") // the namespace could not be deleted
	ErrNamespaceRead   = errors.New("failed to read namespace")   // the namespace could not be retrieved
	ErrNamespaceUpdate = errors.New("failed to update namespace") // the namespace could not be updated
)

// NamespaceProject represents a simplified project within a namespace.
type NamespaceProject struct {
	ID          model.ID            `json:"id"`
	Key         string              `json:"key"`
	Name        string              `json:"name"`
	Description string              `json:"description"`
	Logo        string              `json:"logo"`
	Status      model.ProjectStatus `json:"status"`
}

// NamespaceDocument represents a simplified document within a namespace.
type NamespaceDocument struct {
	ID        model.ID   `json:"id"`
	Name      string     `json:"name"`
	Excerpt   string     `json:"excerpt"`
	CreatedBy model.ID   `json:"created_by"`
	CreatedAt *time.Time `json:"created_at"`
}

// Namespace represents a namespace persisted by the repository.
type Namespace struct {
	ID          model.ID             `json:"id"`
	Name        string               `json:"name"`
	Description string               `json:"description"`
	Projects    []*NamespaceProject  `json:"projects"`
	Documents   []*NamespaceDocument `json:"documents"`
	CreatedAt   *time.Time           `json:"created_at"`
	UpdatedAt   *time.Time           `json:"updated_at"`
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
	Get(ctx context.Context, id model.ID) (*Namespace, error)
	GetAll(ctx context.Context, orgID model.ID, offset, limit int) ([]*Namespace, error)
	Update(ctx context.Context, id model.ID, opts UpdateNamespaceOpts) (*Namespace, error)
	Delete(ctx context.Context, id model.ID) error
}

// Neo4jNamespaceRepository is a repository for managing namespaces.
type Neo4jNamespaceRepository struct {
	*neo4jBaseRepository
}

func (r *Neo4jNamespaceRepository) scan(nsp, pp, dp string) func(rec *neo4j.Record) (*Namespace, error) {
	return func(rec *neo4j.Record) (*Namespace, error) {
		ns := new(Namespace)

		val, _, err := neo4j.GetRecordValue[neo4j.Node](rec, nsp)
		if err != nil {
			return nil, err
		}

		if err := Neo4jScanIntoStruct(&val, &ns, []string{"id"}); err != nil {
			return nil, err
		}

		ns.ID, _ = model.NewIDFromString(val.GetProperties()["id"].(string), model.ResourceTypeNamespace.String())

		projectsVal, err := Neo4jParseValueFromRecord[[]any](rec, pp)
		if err != nil {
			projectsVal = []any{}
		}

		projects := make([]*NamespaceProject, 0, len(projectsVal))
		for _, pVal := range projectsVal {
			if pVal == nil {
				continue
			}
			pNode, ok := pVal.(neo4j.Node)
			if !ok {
				continue
			}

			projectID, err := model.NewIDFromString(pNode.GetProperties()["id"].(string), model.ResourceTypeProject.String())
			if err != nil {
				continue
			}

			var tempProject struct {
				Key         string `json:"key"`
				Name        string `json:"name"`
				Description string `json:"description"`
				Logo        string `json:"logo"`
				Status      string `json:"status"`
			}
			if err := Neo4jScanIntoStruct(&pNode, &tempProject, []string{"id"}); err != nil {
				continue
			}

			var status model.ProjectStatus
			if err := status.UnmarshalText([]byte(tempProject.Status)); err != nil {
				continue
			}

			projects = append(projects, &NamespaceProject{
				ID:          projectID,
				Key:         tempProject.Key,
				Name:        tempProject.Name,
				Description: tempProject.Description,
				Logo:        tempProject.Logo,
				Status:      status,
			})
		}
		ns.Projects = projects

		documentsVal, err := Neo4jParseValueFromRecord[[]any](rec, dp)
		if err != nil {
			documentsVal = []any{}
		}

		documents := make([]*NamespaceDocument, 0, len(documentsVal))
		for _, dVal := range documentsVal {
			if dVal == nil {
				continue
			}
			dNode, ok := dVal.(neo4j.Node)
			if !ok {
				continue
			}

			documentID, err := model.NewIDFromString(dNode.GetProperties()["id"].(string), model.ResourceTypeDocument.String())
			if err != nil {
				continue
			}

			var tempDocument struct {
				Name      string     `json:"name"`
				Excerpt   string     `json:"excerpt"`
				CreatedBy string     `json:"created_by"`
				CreatedAt *time.Time `json:"created_at"`
			}
			if err := Neo4jScanIntoStruct(&dNode, &tempDocument, []string{"id"}); err != nil {
				continue
			}

			createdBy, err := model.NewIDFromString(tempDocument.CreatedBy, model.ResourceTypeUser.String())
			if err != nil {
				continue
			}

			documents = append(documents, &NamespaceDocument{
				ID:        documentID,
				Name:      tempDocument.Name,
				Excerpt:   tempDocument.Excerpt,
				CreatedBy: createdBy,
				CreatedAt: tempDocument.CreatedAt,
			})
		}
		ns.Documents = documents

		return ns, nil
	}
}

func (r *Neo4jNamespaceRepository) Create(ctx context.Context, opts CreateNamespaceOpts) (*Namespace, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.NamespaceRepository/Create")
	defer span.End()

	createdAt := convert.ToPointer(time.Now().UTC())

	namespace := &Namespace{
		ID:          model.MustNewID(model.ResourceTypeNamespace),
		Name:        opts.Name,
		Description: opts.Description,
		Projects:    make([]*NamespaceProject, 0),
		Documents:   make([]*NamespaceDocument, 0),
		CreatedAt:   createdAt,
		UpdatedAt:   nil,
	}

	cypher := `
	MATCH (u:` + opts.CreatorID.Label() + ` {id: $creator_id})
	MATCH (org:` + opts.OrgID.Label() + ` {id: $org_id})
	CREATE (ns:` + namespace.ID.Label() + ` {id: $id, name: $name, description: $description, created_at: datetime($created_at)}),
		(org)-[:` + EdgeKindHasNamespace.String() + ` {id: $has_ns_id, created_at: datetime($created_at)}]->(ns),
		(u)-[:` + EdgeKindHasPermission.String() + ` {id: $perm_id, kind: $perm_kind, created_at: datetime($created_at)}]->(ns)`

	params := map[string]any{
		"id":          namespace.ID.String(),
		"name":        namespace.Name,
		"description": namespace.Description,
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

	return namespace, nil
}

func (r *Neo4jNamespaceRepository) Get(ctx context.Context, id model.ID) (*Namespace, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.NamespaceRepository/Get")
	defer span.End()

	cypher := `
	MATCH (ns:` + id.Label() + ` {id: $id})
	OPTIONAL MATCH (p:` + model.ResourceTypeProject.String() + `)<-[:` + EdgeKindHasProject.String() + `]-(ns)
	OPTIONAL MATCH (d:` + model.ResourceTypeDocument.String() + `)-[:` + EdgeKindBelongsTo.String() + `]->(ns)
	RETURN ns, collect(DISTINCT p) as p, collect(DISTINCT d) as d`

	params := map[string]any{
		"id": id.String(),
	}

	ns, err := Neo4jExecuteReadAndReadSingle(ctx, r.db, cypher, params, r.scan("ns", "p", "d"))
	if err != nil {
		return nil, errors.Join(ErrNamespaceRead, err)
	}

	return ns, nil
}

func (r *Neo4jNamespaceRepository) GetAll(ctx context.Context, orgID model.ID, offset, limit int) ([]*Namespace, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.NamespaceRepository/GetAll")
	defer span.End()

	cypher := `
	MATCH (org:` + orgID.Label() + ` {id: $org_id})-[:` + EdgeKindHasNamespace.String() + `]->(ns:` + model.ResourceTypeNamespace.String() + `)
	OPTIONAL MATCH (p:` + model.ResourceTypeProject.String() + `)<-[:` + EdgeKindHasProject.String() + `]-(ns)
	OPTIONAL MATCH (d:` + model.ResourceTypeDocument.String() + `)-[:` + EdgeKindBelongsTo.String() + `]->(ns)
	RETURN ns, collect(DISTINCT p) as p, collect(DISTINCT d) as d
	ORDER BY ns.created_at DESC
	SKIP $offset LIMIT $limit`

	params := map[string]any{
		"org_id": orgID.String(),
		"offset": offset,
		"limit":  limit,
	}

	nss, err := Neo4jExecuteReadAndReadAll(ctx, r.db, cypher, params, r.scan("ns", "p", "d"))
	if err != nil {
		return nil, errors.Join(ErrNamespaceRead, err)
	}

	return nss, nil
}

func (r *Neo4jNamespaceRepository) Update(ctx context.Context, id model.ID, opts UpdateNamespaceOpts) (*Namespace, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.NamespaceRepository/Update")
	defer span.End()

	cypher := `
	MATCH (ns:` + id.Label() + ` {id: $id}) SET ns += $patch, ns.updated_at = $updated_at
	WITH ns
	OPTIONAL MATCH (p:` + model.ResourceTypeProject.String() + `)<-[:` + EdgeKindHasProject.String() + `]->(ns)
	OPTIONAL MATCH (d:` + model.ResourceTypeDocument.String() + `)-[:` + EdgeKindBelongsTo.String() + `]->(ns)
	RETURN ns, collect(DISTINCT p) as p, collect(DISTINCT d) as d`

	params := map[string]any{
		"id":         id.String(),
		"patch":      opts.patch(),
		"updated_at": time.Now().UTC().Format(time.RFC3339Nano),
	}

	ns, err := Neo4jExecuteWriteAndReadSingle(ctx, r.db, cypher, params, r.scan("ns", "p", "d"))
	if err != nil {
		return nil, errors.Join(ErrNamespaceUpdate, err)
	}

	return ns, nil
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
	return r.Delete(ctx, composeCacheKey(model.ResourceTypeNamespace.String(), id.String()))
}

func clearNamespacesAllGetAll(ctx context.Context, r *redisBaseRepository) error {
	return clearNamespacesPattern(ctx, r, "GetAll", "*")
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
	if err := clearNamespacesAllGetAll(ctx, r.cacheRepo); err != nil {
		return nil, err
	}
	if err := clearNamespaceAllCrossCache(ctx, r.cacheRepo); err != nil {
		return nil, err
	}

	return r.namespaceRepo.Create(ctx, opts)
}

func (r *RedisCachedNamespaceRepository) Get(ctx context.Context, id model.ID) (*Namespace, error) {
	var namespace *Namespace
	var err error

	key := composeCacheKey(model.ResourceTypeNamespace.String(), id.String())
	if err = r.cacheRepo.Get(ctx, key, &namespace); err != nil {
		return nil, err
	}

	if namespace != nil {
		return namespace, nil
	}

	if namespace, err = r.namespaceRepo.Get(ctx, id); err != nil {
		return nil, err
	}

	if err = r.cacheRepo.Set(ctx, key, namespace); err != nil {
		return nil, err
	}

	return namespace, nil
}

func (r *RedisCachedNamespaceRepository) GetAll(ctx context.Context, orgID model.ID, offset, limit int) ([]*Namespace, error) {
	var namespaces []*Namespace
	var err error

	key := composeCacheKey(model.ResourceTypeNamespace.String(), "GetAll", orgID.String(), offset, limit)
	if err = r.cacheRepo.Get(ctx, key, &namespaces); err != nil {
		return nil, err
	}

	if namespaces != nil {
		return namespaces, nil
	}

	namespaces, err = r.namespaceRepo.GetAll(ctx, orgID, offset, limit)
	if err != nil {
		return nil, err
	}

	if err = r.cacheRepo.Set(ctx, key, namespaces); err != nil {
		return nil, err
	}

	return namespaces, nil
}

func (r *RedisCachedNamespaceRepository) Update(ctx context.Context, id model.ID, opts UpdateNamespaceOpts) (*Namespace, error) {
	namespace, err := r.namespaceRepo.Update(ctx, id, opts)
	if err != nil {
		return nil, err
	}

	key := composeCacheKey(model.ResourceTypeNamespace.String(), id.String())
	if err = r.cacheRepo.Set(ctx, key, namespace); err != nil {
		return nil, err
	}

	if err := clearNamespacesAllGetAll(ctx, r.cacheRepo); err != nil {
		return nil, err
	}

	return namespace, nil
}

func (r *RedisCachedNamespaceRepository) Delete(ctx context.Context, id model.ID) error {
	if err := clearNamespacesKey(ctx, r.cacheRepo, id); err != nil {
		return err
	}

	if err := clearNamespacesAllGetAll(ctx, r.cacheRepo); err != nil {
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

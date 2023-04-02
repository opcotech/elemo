package neo4j

import (
	"context"
	"errors"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"

	"github.com/opcotech/elemo/internal/model"
)

var (
	ErrNamespaceCreate = errors.New("failed to create namespace") // the namespace could not be created
	ErrNamespaceRead   = errors.New("failed to read namespace")   // the namespace could not be retrieved
	ErrNamespaceUpdate = errors.New("failed to update namespace") // the namespace could not be updated
	ErrNamespaceDelete = errors.New("failed to delete namespace") // the namespace could not be deleted
)

// NamespaceRepository is a repository for managing namespaces.
type NamespaceRepository struct {
	*repository
}

func (r *NamespaceRepository) scan(nsp, pp, dp string) func(rec *neo4j.Record) (*model.Namespace, error) {
	return func(rec *neo4j.Record) (*model.Namespace, error) {
		ns := new(model.Namespace)

		val, _, err := neo4j.GetRecordValue[neo4j.Node](rec, nsp)
		if err != nil {
			return nil, err
		}

		if err := ScanIntoStruct(&val, &ns, []string{"id"}); err != nil {
			return nil, err
		}

		ns.ID, _ = model.NewIDFromString(val.GetProperties()["id"].(string), model.NamespaceIDType)

		if ns.Projects, err = ParseIDsFromRecord(rec, pp, model.ProjectIDType); err != nil {
			return nil, err
		}

		if ns.Documents, err = ParseIDsFromRecord(rec, dp, model.NamespaceIDType); err != nil {
			return nil, err
		}

		if err := ns.Validate(); err != nil {
			return nil, err
		}

		return ns, nil
	}
}

func (r *NamespaceRepository) Create(ctx context.Context, orgID model.ID, namespace *model.Namespace) error {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.NamespaceRepository/Create")
	defer span.End()

	if err := orgID.Validate(); err != nil {
		return errors.Join(ErrAttachmentCreate, err)
	}

	if err := namespace.Validate(); err != nil {
		return errors.Join(ErrNamespaceCreate, err)
	}

	createdAt := time.Now()

	hasNsID := model.MustNewID(EdgeKindHasNamespace.String())

	namespace.ID = model.MustNewID(model.NamespaceIDType)
	namespace.CreatedAt = &createdAt
	namespace.UpdatedAt = nil

	cypher := `
	MATCH (org:` + orgID.Label() + ` {id: $org_id})
	CREATE (ns:` + namespace.ID.Label() + ` {id: $id, name: $name, description: $description, created_at: datetime($created_at)}),
		(org)-[:` + hasNsID.Label() + ` {id: $has_ns_id, created_at: datetime($created_at)}]->(ns)`

	params := map[string]any{
		"id":          namespace.ID.String(),
		"name":        namespace.Name,
		"description": namespace.Description,
		"created_at":  createdAt.Format(time.RFC3339Nano),
		"org_id":      orgID.String(),
		"has_ns_id":   hasNsID.String(),
	}

	if err := ExecuteWriteAndConsume(ctx, r.db, cypher, params); err != nil {
		return errors.Join(ErrNamespaceCreate, err)
	}

	return nil
}

func (r *NamespaceRepository) Get(ctx context.Context, id model.ID) (*model.Namespace, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.NamespaceRepository/Get")
	defer span.End()

	cypher := `
	MATCH (ns:` + id.Label() + ` {id: $id})
	OPTIONAL MATCH (p:` + model.ProjectIDType + `)<-[:` + EdgeKindHasProject.String() + `]-(ns)
	OPTIONAL MATCH (d:` + model.DocumentIDType + `)-[:` + EdgeKindBelongsTo.String() + `]->(ns)
	RETURN ns, collect(DISTINCT p.id) as p, collect(DISTINCT d.id) as d`

	params := map[string]any{
		"id": id.String(),
	}

	ns, err := ExecuteReadAndReadSingle(ctx, r.db, cypher, params, r.scan("ns", "p", "d"))
	if err != nil {
		return nil, errors.Join(ErrNamespaceRead, err)
	}

	return ns, nil
}

func (r *NamespaceRepository) GetAll(ctx context.Context, orgID model.ID, offset, limit int) ([]*model.Namespace, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.NamespaceRepository/GetAll")
	defer span.End()

	cypher := `
	MATCH (org:` + orgID.Label() + ` {id: $org_id})-[:` + EdgeKindHasNamespace.String() + `]->(ns:` + model.NamespaceIDType + `)
	OPTIONAL MATCH (p:` + model.ProjectIDType + `)<-[:` + EdgeKindHasProject.String() + `]-(ns)
	OPTIONAL MATCH (d:` + model.DocumentIDType + `)-[:` + EdgeKindBelongsTo.String() + `]->(ns)
	RETURN ns, collect(DISTINCT p.id) as p, collect(DISTINCT d.id) as d
	ORDER BY ns.created_at DESC
	SKIP $offset LIMIT $limit`

	params := map[string]any{
		"org_id": orgID.String(),
		"offset": offset,
		"limit":  limit,
	}

	nss, err := ExecuteReadAndReadAll(ctx, r.db, cypher, params, r.scan("ns", "p", "d"))
	if err != nil {
		return nil, errors.Join(ErrNamespaceRead, err)
	}

	return nss, nil
}

func (r *NamespaceRepository) Update(ctx context.Context, id model.ID, patch map[string]any) (*model.Namespace, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.NamespaceRepository/Update")
	defer span.End()

	cypher := `
	MATCH (ns:` + id.Label() + ` {id: $id}) SET ns += $patch, ns.updated_at = $updated_at
	WITH ns
	OPTIONAL MATCH (p:` + model.ProjectIDType + `)<-[:` + EdgeKindHasProject.String() + `]->(ns)
	OPTIONAL MATCH (d:` + model.DocumentIDType + `)-[:` + EdgeKindBelongsTo.String() + `]->(ns)
	RETURN ns, collect(DISTINCT p.id) as p, collect(DISTINCT d.id) as d`

	params := map[string]any{
		"id":         id.String(),
		"patch":      patch,
		"updated_at": time.Now().Format(time.RFC3339Nano),
	}

	ns, err := ExecuteReadAndReadSingle(ctx, r.db, cypher, params, r.scan("ns", "p", "d"))
	if err != nil {
		return nil, errors.Join(ErrNamespaceUpdate, err)
	}

	return ns, nil
}

func (r *NamespaceRepository) Delete(ctx context.Context, id model.ID) error {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.NamespaceRepository/Delete")
	defer span.End()

	cypher := `
	MATCH (ns:` + id.Label() + ` {id: $id}) DETACH DELETE ns`

	params := map[string]any{
		"id": id.String(),
	}

	if err := ExecuteWriteAndConsume(ctx, r.db, cypher, params); err != nil {
		return errors.Join(ErrNamespaceDelete, err)
	}

	return nil
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

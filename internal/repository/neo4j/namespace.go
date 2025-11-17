package neo4j

import (
	"context"
	"errors"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/repository"
)

// NamespaceRepository is a repository for managing namespaces.
type NamespaceRepository struct {
	*baseRepository
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

		ns.ID, _ = model.NewIDFromString(val.GetProperties()["id"].(string), model.ResourceTypeNamespace.String())

		// Parse projects from collected nodes
		projectsVal, err := ParseValueFromRecord[[]any](rec, pp)
		if err != nil {
			projectsVal = []any{}
		}

		projects := make([]*model.NamespaceProject, 0, len(projectsVal))
		for _, pVal := range projectsVal {
			if pVal == nil {
				continue
			}
			if pNode, ok := pVal.(neo4j.Node); ok {
				projectID, err := model.NewIDFromString(pNode.GetProperties()["id"].(string), model.ResourceTypeProject.String())
				if err != nil {
					continue
				}

				// Use ScanIntoStruct to parse project fields
				var tempProject struct {
					Key         string `json:"key"`
					Name        string `json:"name"`
					Description string `json:"description"`
					Logo        string `json:"logo"`
					Status      string `json:"status"`
				}
				if err := ScanIntoStruct(&pNode, &tempProject, []string{"id"}); err != nil {
					continue
				}

				var status model.ProjectStatus
				if err := status.UnmarshalText([]byte(tempProject.Status)); err != nil {
					continue
				}

				project, err := model.NewNamespaceProject(projectID, tempProject.Key, tempProject.Name, tempProject.Description, tempProject.Logo, status)
				if err != nil {
					continue
				}

				projects = append(projects, project)
			}
		}
		ns.Projects = projects

		// Parse documents from collected nodes
		documentsVal, err := ParseValueFromRecord[[]any](rec, dp)
		if err != nil {
			documentsVal = []any{}
		}

		documents := make([]*model.NamespaceDocument, 0, len(documentsVal))
		for _, dVal := range documentsVal {
			if dVal == nil {
				continue
			}
			if dNode, ok := dVal.(neo4j.Node); ok {
				documentID, err := model.NewIDFromString(dNode.GetProperties()["id"].(string), model.ResourceTypeDocument.String())
				if err != nil {
					continue
				}

				// Use ScanIntoStruct to parse document fields
				var tempDocument struct {
					Name      string     `json:"name"`
					Excerpt   string     `json:"excerpt"`
					CreatedBy string     `json:"created_by"`
					CreatedAt *time.Time `json:"created_at"`
				}
				if err := ScanIntoStruct(&dNode, &tempDocument, []string{"id"}); err != nil {
					continue
				}

				createdBy, err := model.NewIDFromString(tempDocument.CreatedBy, model.ResourceTypeUser.String())
				if err != nil {
					continue
				}

				document, err := model.NewNamespaceDocument(documentID, tempDocument.Name, tempDocument.Excerpt, createdBy, tempDocument.CreatedAt)
				if err != nil {
					continue
				}

				documents = append(documents, document)
			}
		}
		ns.Documents = documents

		if err := ns.Validate(); err != nil {
			return nil, err
		}

		return ns, nil
	}
}

func (r *NamespaceRepository) Create(ctx context.Context, creatorID, orgID model.ID, namespace *model.Namespace) error {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.NamespaceRepository/Create")
	defer span.End()

	if err := creatorID.Validate(); err != nil {
		return errors.Join(repository.ErrNamespaceCreate, err)
	}

	if err := orgID.Validate(); err != nil {
		return errors.Join(repository.ErrNamespaceCreate, err)
	}

	if err := namespace.Validate(); err != nil {
		return errors.Join(repository.ErrNamespaceCreate, err)
	}

	createdAt := time.Now().UTC()

	namespace.ID = model.MustNewID(model.ResourceTypeNamespace)
	namespace.CreatedAt = &createdAt
	namespace.UpdatedAt = nil

	cypher := `
	MATCH (u:` + creatorID.Label() + ` {id: $creator_id})
	MATCH (org:` + orgID.Label() + ` {id: $org_id})
	CREATE (ns:` + namespace.ID.Label() + ` {id: $id, name: $name, description: $description, created_at: datetime($created_at)}),
		(org)-[:` + EdgeKindHasNamespace.String() + ` {id: $has_ns_id, created_at: datetime($created_at)}]->(ns),
		(u)-[:` + EdgeKindHasPermission.String() + ` {id: $perm_id, kind: $perm_kind, created_at: datetime($created_at)}]->(ns)`

	params := map[string]any{
		"id":          namespace.ID.String(),
		"name":        namespace.Name,
		"description": namespace.Description,
		"created_at":  createdAt.Format(time.RFC3339Nano),
		"creator_id":  creatorID.String(),
		"org_id":      orgID.String(),
		"has_ns_id":   model.NewRawID(),
		"perm_id":     model.NewRawID(),
		"perm_kind":   model.PermissionKindAll.String(),
	}

	if err := ExecuteWriteAndConsume(ctx, r.db, cypher, params); err != nil {
		return errors.Join(repository.ErrNamespaceCreate, err)
	}

	return nil
}

func (r *NamespaceRepository) Get(ctx context.Context, id model.ID) (*model.Namespace, error) {
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

	ns, err := ExecuteReadAndReadSingle(ctx, r.db, cypher, params, r.scan("ns", "p", "d"))
	if err != nil {
		return nil, errors.Join(repository.ErrNamespaceRead, err)
	}

	return ns, nil
}

func (r *NamespaceRepository) GetAll(ctx context.Context, orgID model.ID, offset, limit int) ([]*model.Namespace, error) {
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

	nss, err := ExecuteReadAndReadAll(ctx, r.db, cypher, params, r.scan("ns", "p", "d"))
	if err != nil {
		return nil, errors.Join(repository.ErrNamespaceRead, err)
	}

	return nss, nil
}

func (r *NamespaceRepository) Update(ctx context.Context, id model.ID, patch map[string]any) (*model.Namespace, error) {
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
		"patch":      patch,
		"updated_at": time.Now().UTC().Format(time.RFC3339Nano),
	}

	ns, err := ExecuteWriteAndReadSingle(ctx, r.db, cypher, params, r.scan("ns", "p", "d"))
	if err != nil {
		return nil, errors.Join(repository.ErrNamespaceUpdate, err)
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
		return errors.Join(repository.ErrNamespaceDelete, err)
	}

	return nil
}

// NewNamespaceRepository creates a new namespace baseRepository.
func NewNamespaceRepository(opts ...RepositoryOption) (*NamespaceRepository, error) {
	baseRepo, err := newRepository(opts...)
	if err != nil {
		return nil, err
	}

	return &NamespaceRepository{
		baseRepository: baseRepo,
	}, nil
}

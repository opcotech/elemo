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

// TodoRepository is a repository for managing todos.
type TodoRepository struct {
	*baseRepository
}

func (r *TodoRepository) scan(tp, op, cp string) func(rec *neo4j.Record) (*model.Todo, error) {
	return func(rec *neo4j.Record) (*model.Todo, error) {
		todo := new(model.Todo)

		val, _, err := neo4j.GetRecordValue[neo4j.Node](rec, tp)
		if err != nil {
			return nil, err
		}

		ownerID, err := ParseValueFromRecord[string](rec, op)
		if err != nil {
			return nil, err
		}

		creatorID, err := ParseValueFromRecord[string](rec, cp)
		if err != nil {
			return nil, err
		}

		if err := ScanIntoStruct(&val, &todo, []string{"id"}); err != nil {
			return nil, err
		}

		todo.ID, _ = model.NewIDFromString(val.GetProperties()["id"].(string), model.TodoIDType)
		todo.OwnedBy, _ = model.NewIDFromString(ownerID, model.UserIDType)
		todo.CreatedBy, _ = model.NewIDFromString(creatorID, model.UserIDType)

		if err := todo.Validate(); err != nil {
			return nil, err
		}

		return todo, nil
	}
}

func (r *TodoRepository) Create(ctx context.Context, todo *model.Todo) error {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.TodoRepository/Create")
	defer span.End()

	if err := todo.Validate(); err != nil {
		return errors.Join(repository.ErrTodoCreate, err)
	}

	createdAt := convert.ToPointer(time.Now())

	ownerRelID := model.MustNewID(EdgeKindBelongsTo.String())
	ownerPermID := model.MustNewID(EdgeKindHasPermission.String())
	creatorRelID := model.MustNewID(EdgeKindCreated.String())
	creatorPermID := model.MustNewID(EdgeKindHasPermission.String())

	todo.ID = model.MustNewID(model.TodoIDType)
	todo.CreatedAt = createdAt
	todo.UpdatedAt = nil

	cypher := `
	MATCH (o:` + todo.OwnedBy.Label() + ` {id: $owner_id}), (c:` + todo.CreatedBy.Label() + ` {id: $creator_id})
	CREATE
		(t:` + todo.ID.Label() + ` {
			id: $id, title: $title, description: $description, priority: $priority, completed: $completed,
			due_date: datetime($due_date), created_at: datetime($created_at)
		}),
		(t)-[:` + ownerRelID.Label() + ` {id: $owned_rel_id, created_at: datetime($created_at)}]->(o),
		(t)<-[:` + creatorRelID.Label() + ` {id: $created_rel_id, created_at: datetime($created_at)}]-(c),
		(o)-[:` + ownerPermID.Label() + ` {id: $owner_perm_id, kind: $owner_perm_kind, created_at: datetime($created_at)}]->(t)
	MERGE (c)-[rel:` + creatorPermID.Label() + `]->(t)
	ON CREATE SET rel += {id: $creator_perm_id, kind: $creator_perm_kind, created_at: datetime($created_at)}`

	params := map[string]any{
		"id":                todo.ID.String(),
		"title":             todo.Title,
		"description":       todo.Description,
		"priority":          todo.Priority.String(),
		"completed":         todo.Completed,
		"due_date":          nil,
		"created_at":        todo.CreatedAt.Format(time.RFC3339Nano),
		"owner_id":          todo.OwnedBy.String(),
		"owned_rel_id":      ownerRelID.String(),
		"owner_perm_id":     ownerPermID.String(),
		"owner_perm_kind":   model.PermissionKindAll.String(),
		"creator_id":        todo.CreatedBy.String(),
		"created_rel_id":    creatorRelID.String(),
		"creator_perm_id":   creatorPermID.String(),
		"creator_perm_kind": model.PermissionKindAll.String(),
	}

	if todo.DueDate != nil {
		params["due_date"] = todo.DueDate.Format(time.RFC3339Nano)
	}

	if err := ExecuteWriteAndConsume(ctx, r.db, cypher, params); err != nil {
		return errors.Join(err, repository.ErrTodoCreate)
	}

	return nil
}

func (r *TodoRepository) Get(ctx context.Context, id model.ID) (*model.Todo, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.TodoRepository/Get")
	defer span.End()

	cypher := `
	MATCH (t:` + id.Label() + ` {id: $id})
	OPTIONAL MATCH (t)-[:` + EdgeKindBelongsTo.String() + `]->(o)
	OPTIONAL MATCH (t)<-[:` + EdgeKindCreated.String() + `]-(c)
	RETURN t, o.id as o, c.id as c`

	params := map[string]any{
		"id": id.String(),
	}

	todo, err := ExecuteReadAndReadSingle(ctx, r.db, cypher, params, r.scan("t", "o", "c"))
	if err != nil {
		return nil, errors.Join(err, repository.ErrTodoRead)
	}

	return todo, nil
}

func (r *TodoRepository) GetByOwner(ctx context.Context, ownerID model.ID, completed *bool) ([]*model.Todo, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.TodoRepository/GetByCreator")
	defer span.End()

	cypher := `
	MATCH (t:` + model.TodoIDType + `)-[:` + EdgeKindBelongsTo.String() + `]->(o:` + ownerID.Label() + ` {id: $owner_id})
	WHERE $completed IS NULL OR t.completed = $completed
	OPTIONAL MATCH (t)<-[:` + EdgeKindCreated.String() + `]-(c)
	RETURN t, o.id as o, c.id as c
	ORDER BY t.created_at DESC`

	params := map[string]any{
		"owner_id":  ownerID.String(),
		"completed": completed,
	}

	todos, err := ExecuteReadAndReadAll(ctx, r.db, cypher, params, r.scan("t", "o", "c"))
	if err != nil {
		return nil, errors.Join(err, repository.ErrTodoRead)
	}

	return todos, nil
}

func (r *TodoRepository) Update(ctx context.Context, id model.ID, patch map[string]any) (*model.Todo, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.TodoRepository/Update")
	defer span.End()

	cypher := `
	MATCH (t:` + id.Label() + ` {id: $id})
	SET t += $patch, t.updated_at = datetime($updated_at)
	WITH t
	OPTIONAL MATCH (t)-[:` + EdgeKindBelongsTo.String() + `]->(o)
	OPTIONAL MATCH (t)<-[:` + EdgeKindCreated.String() + `]-(c)
	RETURN t, o.id as o, c.id as c`

	params := map[string]any{
		"id":         id.String(),
		"patch":      patch,
		"updated_at": time.Now().Format(time.RFC3339Nano),
	}

	todo, err := ExecuteWriteAndReadSingle(ctx, r.db, cypher, params, r.scan("t", "o", "c"))
	if err != nil {
		return nil, errors.Join(repository.ErrTodoUpdate, err)
	}

	return todo, nil
}

func (r *TodoRepository) Delete(ctx context.Context, id model.ID) error {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.TodoRepository/Delete")
	defer span.End()

	cypher := `MATCH (t:` + id.Label() + ` {id: $id}) DETACH DELETE t`
	params := map[string]any{
		"id": id.String(),
	}

	if err := ExecuteWriteAndConsume(ctx, r.db, cypher, params); err != nil {
		return errors.Join(repository.ErrTodoDelete, err)
	}

	return nil
}

// NewTodoRepository creates a new todo baseRepository.
func NewTodoRepository(opts ...RepositoryOption) (*TodoRepository, error) {
	baseRepo, err := newRepository(opts...)
	if err != nil {
		return nil, err
	}

	return &TodoRepository{
		baseRepository: baseRepo,
	}, nil
}

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
	ErrTodoCreate = errors.New("failed to create todo") // todo cannot be created
	ErrTodoDelete = errors.New("failed to delete todo") // todo cannot be deleted
	ErrTodoRead   = errors.New("failed to read todo")   // todo cannot be read
	ErrTodoUpdate = errors.New("failed to update todo") // todo cannot be updated
)

// Todo represents a todo persisted by the repository.
type Todo struct {
	ID          model.ID           `json:"id"`
	Title       string             `json:"title"`
	Description string             `json:"description"`
	Priority    model.TodoPriority `json:"priority"`
	Completed   bool               `json:"completed"`
	OwnedBy     model.ID           `json:"owned_by"`
	CreatedBy   model.ID           `json:"created_by"`
	DueDate     *time.Time         `json:"due_date"`
	CreatedAt   *time.Time         `json:"created_at"`
	UpdatedAt   *time.Time         `json:"updated_at"`
}

// CreateTodoOpts holds the data required to create a todo.
type CreateTodoOpts struct {
	Title       string
	Description string
	Priority    model.TodoPriority
	Completed   bool
	OwnedBy     model.ID
	CreatedBy   model.ID
	DueDate     *time.Time
}

// UpdateTodoOpts holds the fields that can be updated on a todo.
// Undefined fields (Defined == false) are left unchanged.
type UpdateTodoOpts struct {
	Title       optional.Optional[string]
	Description optional.Optional[string]
	Priority    optional.Optional[model.TodoPriority]
	Completed   optional.Optional[bool]
	DueDate     optional.Optional[time.Time]
}

// patch builds a Neo4j property map from defined optional fields.
func (o UpdateTodoOpts) patch() map[string]any {
	p := make(map[string]any)

	if o.Title.Defined {
		p["title"] = *o.Title.Value
	}
	if o.Description.Defined {
		p["description"] = *o.Description.Value
	}
	if o.Priority.Defined {
		p["priority"] = o.Priority.Value.String()
	}
	if o.Completed.Defined {
		p["completed"] = *o.Completed.Value
	}
	if o.DueDate.Defined {
		if o.DueDate.Value == nil {
			p["due_date"] = nil
		} else {
			p["due_date"] = o.DueDate.Value.Format(time.RFC3339Nano)
		}
	}

	return p
}

//go:generate go tool mockgen -source=todo.go -destination=todo_mock_gen.go -package=repository -mock_names "TodoRepository=MockTodoRepository"
type TodoRepository interface {
	Create(ctx context.Context, opts CreateTodoOpts) (*Todo, error)
	Get(ctx context.Context, id model.ID) (*Todo, error)
	GetByOwner(ctx context.Context, ownerID model.ID, offset, limit int, completed *bool) ([]*Todo, error)
	Update(ctx context.Context, id model.ID, opts UpdateTodoOpts) (*Todo, error)
	Delete(ctx context.Context, id model.ID) error
}

// Neo4jTodoRepository is a repository for managing todos.
type Neo4jTodoRepository struct {
	*neo4jBaseRepository
}

func (r *Neo4jTodoRepository) scan(tp, op, cp string) func(rec *neo4j.Record) (*Todo, error) {
	return func(rec *neo4j.Record) (*Todo, error) {
		todo := new(Todo)

		val, _, err := neo4j.GetRecordValue[neo4j.Node](rec, tp)
		if err != nil {
			return nil, err
		}

		ownerID, err := Neo4jParseValueFromRecord[string](rec, op)
		if err != nil {
			return nil, err
		}

		creatorID, err := Neo4jParseValueFromRecord[string](rec, cp)
		if err != nil {
			return nil, err
		}

		if err := Neo4jScanIntoStruct(&val, &todo, []string{"id"}); err != nil {
			return nil, err
		}

		todo.ID, _ = model.NewIDFromString(val.GetProperties()["id"].(string), model.ResourceTypeTodo.String())
		todo.OwnedBy, _ = model.NewIDFromString(ownerID, model.ResourceTypeUser.String())
		todo.CreatedBy, _ = model.NewIDFromString(creatorID, model.ResourceTypeUser.String())

		return todo, nil
	}
}

func (r *Neo4jTodoRepository) Create(ctx context.Context, opts CreateTodoOpts) (*Todo, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.TodoRepository/Create")
	defer span.End()

	createdAt := convert.ToPointer(time.Now().UTC())

	todo := &Todo{
		ID:          model.MustNewID(model.ResourceTypeTodo),
		Title:       opts.Title,
		Description: opts.Description,
		Priority:    opts.Priority,
		Completed:   opts.Completed,
		OwnedBy:     opts.OwnedBy,
		CreatedBy:   opts.CreatedBy,
		DueDate:     opts.DueDate,
		CreatedAt:   createdAt,
		UpdatedAt:   nil,
	}

	cypher := `
	MATCH (o:` + todo.OwnedBy.Label() + ` {id: $owner_id})
	MATCH (c:` + todo.CreatedBy.Label() + ` {id: $creator_id})
	CREATE
		(t:` + todo.ID.Label() + ` {
			id: $id, title: $title, description: $description, priority: $priority, completed: $completed,
			due_date: datetime($due_date), created_at: datetime($created_at)
		}),
		(t)-[:` + EdgeKindBelongsTo.String() + ` {id: $owned_rel_id, created_at: datetime($created_at)}]->(o),
		(t)<-[:` + EdgeKindCreated.String() + ` {id: $created_rel_id, created_at: datetime($created_at)}]-(c),
		(o)-[:` + EdgeKindHasPermission.String() + ` {id: $owner_perm_id, kind: $owner_perm_kind, created_at: datetime($created_at)}]->(t)
	MERGE (c)-[rel:` + EdgeKindHasPermission.String() + `]->(t)
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
		"owned_rel_id":      model.NewRawID(),
		"owner_perm_id":     model.NewRawID(),
		"owner_perm_kind":   model.PermissionKindAll.String(),
		"creator_id":        todo.CreatedBy.String(),
		"created_rel_id":    model.NewRawID(),
		"creator_perm_id":   model.NewRawID(),
		"creator_perm_kind": model.PermissionKindAll.String(),
	}

	if todo.DueDate != nil {
		params["due_date"] = todo.DueDate.Format(time.RFC3339Nano)
	}

	if err := Neo4jExecuteWriteAndConsume(ctx, r.db, cypher, params); err != nil {
		return nil, errors.Join(err, ErrTodoCreate)
	}

	return todo, nil
}

func (r *Neo4jTodoRepository) Get(ctx context.Context, id model.ID) (*Todo, error) {
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

	todo, err := Neo4jExecuteReadAndReadSingle(ctx, r.db, cypher, params, r.scan("t", "o", "c"))
	if err != nil {
		return nil, errors.Join(err, ErrTodoRead)
	}

	return todo, nil
}

func (r *Neo4jTodoRepository) GetByOwner(ctx context.Context, ownerID model.ID, offset, limit int, completed *bool) ([]*Todo, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.TodoRepository/GetByOwner")
	defer span.End()

	cypher := `
	MATCH (t:` + model.ResourceTypeTodo.String() + `)-[:` + EdgeKindBelongsTo.String() + `]->(o:` + ownerID.Label() + ` {id: $owner_id})
	WHERE $completed IS NULL OR t.completed = $completed
	OPTIONAL MATCH (t)<-[:` + EdgeKindCreated.String() + `]-(c)
	RETURN t, o.id as o, c.id as c
	ORDER BY t.created_at DESC
	SKIP $offset LIMIT $limit`

	params := map[string]any{
		"owner_id":  ownerID.String(),
		"offset":    offset,
		"limit":     limit,
		"completed": completed,
	}

	todos, err := Neo4jExecuteReadAndReadAll(ctx, r.db, cypher, params, r.scan("t", "o", "c"))
	if err != nil {
		return nil, errors.Join(err, ErrTodoRead)
	}

	return todos, nil
}

func (r *Neo4jTodoRepository) Update(ctx context.Context, id model.ID, opts UpdateTodoOpts) (*Todo, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.TodoRepository/Update")
	defer span.End()

	cypher := `
	MATCH (t:` + id.Label() + ` {id: $id})
	SET t += $patch, t.updated_at = datetime()
	WITH t
	OPTIONAL MATCH (t)-[:` + EdgeKindBelongsTo.String() + `]->(o)
	OPTIONAL MATCH (t)<-[:` + EdgeKindCreated.String() + `]-(c)
	RETURN t, o.id as o, c.id as c`

	params := map[string]any{
		"id":    id.String(),
		"patch": opts.patch(),
	}

	todo, err := Neo4jExecuteWriteAndReadSingle(ctx, r.db, cypher, params, r.scan("t", "o", "c"))
	if err != nil {
		return nil, errors.Join(ErrTodoUpdate, err)
	}

	return todo, nil
}

func (r *Neo4jTodoRepository) Delete(ctx context.Context, id model.ID) error {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.TodoRepository/Delete")
	defer span.End()

	cypher := `MATCH (t:` + id.Label() + ` {id: $id}) DETACH DELETE t`
	params := map[string]any{
		"id": id.String(),
	}

	if err := Neo4jExecuteWriteAndConsume(ctx, r.db, cypher, params); err != nil {
		return errors.Join(ErrTodoDelete, err)
	}

	return nil
}

// NewNeo4jTodoRepository creates a new todo neo4jBaseRepository.
func NewNeo4jTodoRepository(opts ...Neo4jRepositoryOption) (*Neo4jTodoRepository, error) {
	baseRepo, err := newNeo4jRepository(opts...)
	if err != nil {
		return nil, err
	}

	return &Neo4jTodoRepository{
		neo4jBaseRepository: baseRepo,
	}, nil
}

// RedisCachedTodoRepository implements caching on the TodoRepository.
type RedisCachedTodoRepository struct {
	cacheRepo *redisBaseRepository
	todoRepo  TodoRepository
}

func (r *RedisCachedTodoRepository) Create(ctx context.Context, opts CreateTodoOpts) (*Todo, error) {
	pattern := composeCacheKey(model.ResourceTypeTodo.String(), "GetByOwner", opts.OwnedBy.String(), "*")
	if err := r.cacheRepo.DeletePattern(ctx, pattern); err != nil {
		return nil, err
	}

	return r.todoRepo.Create(ctx, opts)
}

func (r *RedisCachedTodoRepository) Get(ctx context.Context, id model.ID) (*Todo, error) {
	var todo *Todo
	var err error

	key := composeCacheKey(model.ResourceTypeTodo.String(), id.String())
	if err = r.cacheRepo.Get(ctx, key, &todo); err != nil {
		return nil, err
	}

	if todo != nil {
		return todo, nil
	}

	if todo, err = r.todoRepo.Get(ctx, id); err != nil {
		return nil, err
	}

	if err = r.cacheRepo.Set(ctx, key, todo); err != nil {
		return nil, err
	}

	return todo, nil
}

func (r *RedisCachedTodoRepository) GetByOwner(ctx context.Context, ownerID model.ID, offset, limit int, completed *bool) ([]*Todo, error) {
	var todos []*Todo
	var err error

	key := composeCacheKey(model.ResourceTypeTodo.String(), "GetByOwner", ownerID.String(), offset, limit, completed)
	if err = r.cacheRepo.Get(ctx, key, &todos); err != nil {
		return nil, err
	}

	if todos != nil {
		return todos, nil
	}

	todos, err = r.todoRepo.GetByOwner(ctx, ownerID, offset, limit, completed)
	if err != nil {
		return nil, err
	}

	if err = r.cacheRepo.Set(ctx, key, todos); err != nil {
		return nil, err
	}

	return todos, nil
}

func (r *RedisCachedTodoRepository) Update(ctx context.Context, id model.ID, opts UpdateTodoOpts) (*Todo, error) {
	todo, err := r.todoRepo.Update(ctx, id, opts)
	if err != nil {
		return nil, err
	}

	key := composeCacheKey(model.ResourceTypeTodo.String(), id.String())
	if err = r.cacheRepo.Set(ctx, key, todo); err != nil {
		return nil, err
	}

	pattern := composeCacheKey(model.ResourceTypeTodo.String(), "GetByOwner", todo.OwnedBy.String(), "*")
	if err := r.cacheRepo.DeletePattern(ctx, pattern); err != nil {
		return nil, err
	}

	return todo, nil
}

func (r *RedisCachedTodoRepository) Delete(ctx context.Context, id model.ID) error {
	key := composeCacheKey(model.ResourceTypeTodo.String(), id.String())
	if err := r.cacheRepo.Delete(ctx, key); err != nil {
		return err
	}

	pattern := composeCacheKey(model.ResourceTypeTodo.String(), "GetByOwner", "*")
	if err := r.cacheRepo.DeletePattern(ctx, pattern); err != nil {
		return err
	}

	return r.todoRepo.Delete(ctx, id)
}

// NewCachedTodoRepository returns a new CachedTodoRepository.
func NewCachedTodoRepository(repo TodoRepository, opts ...RedisRepositoryOption) (*RedisCachedTodoRepository, error) {
	r, err := newRedisBaseRepository(opts...)
	if err != nil {
		return nil, err
	}

	return &RedisCachedTodoRepository{
		cacheRepo: r,
		todoRepo:  repo,
	}, nil
}

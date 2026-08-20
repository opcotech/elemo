package repository

import (
	"context"
	"errors"
	"time"

	"github.com/neo4j/neo4j-go-driver/v6/neo4j"

	"github.com/opcotech/elemo/internal/model"
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
	ListByOwner(ctx context.Context, ownerID model.ID, page CursorPage, completed *bool) (Page[*Todo], error)
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

	createdAt := time.Now().UTC()
	id := model.MustNewID(model.ResourceTypeTodo)

	cypher := `
	MATCH (o:` + opts.OwnedBy.Label() + ` {id: $owner_id})
	MATCH (c:` + opts.CreatedBy.Label() + ` {id: $creator_id})
	CREATE
		(t:` + id.Label() + ` {
			id: $id, title: $title, description: $description, priority: $priority, completed: $completed,
			due_date: datetime($due_date), created_at: datetime($created_at)
		}),
		(t)-[:` + EdgeKindBelongsTo.String() + ` {id: $owned_rel_id, created_at: datetime($created_at)}]->(o),
		(t)-[:` + EdgeKindInScopeOf.String() + ` {id: $scope_id, created_at: datetime($created_at)}]->(o),
		(t)<-[:` + EdgeKindCreated.String() + ` {id: $created_rel_id, created_at: datetime($created_at)}]-(c)`

	params := map[string]any{
		"id":             id.String(),
		"title":          opts.Title,
		"description":    opts.Description,
		"priority":       opts.Priority.String(),
		"completed":      opts.Completed,
		"due_date":       nil,
		"created_at":     createdAt.Format(time.RFC3339Nano),
		"owner_id":       opts.OwnedBy.String(),
		"owned_rel_id":   model.NewRawID(),
		"scope_id":       model.NewRawID(),
		"creator_id":     opts.CreatedBy.String(),
		"created_rel_id": model.NewRawID(),
	}

	if opts.DueDate != nil {
		params["due_date"] = opts.DueDate.Format(time.RFC3339Nano)
	}

	if err := Neo4jExecuteWriteAndConsume(ctx, r.db, cypher, params); err != nil {
		return nil, errors.Join(err, ErrTodoCreate)
	}

	return r.Get(ctx, id)
}

func (r *Neo4jTodoRepository) Get(ctx context.Context, id model.ID) (*Todo, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.TodoRepository/Get")
	defer span.End()

	plan, err := CompileQuery(TodoGetQuery{
		ID: id,
	})
	if err != nil {
		return nil, errors.Join(err, ErrTodoRead)
	}

	var todo *Todo
	err = Neo4jExecuteReadPlan(ctx, r.db, plan, func(tx neo4j.ManagedTransaction) error {
		var runErr error
		todo, _, runErr = Neo4jRunQuerySingle(ctx, tx, plan.Root, r.scan("t", "o", "c"))
		return runErr
	})
	if err != nil {
		return nil, errors.Join(err, ErrTodoRead)
	}

	return todo, nil
}

func (r *Neo4jTodoRepository) ListByOwner(ctx context.Context, ownerID model.ID, page CursorPage, completed *bool) (Page[*Todo], error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.TodoRepository/ListByOwner")
	defer span.End()

	normalized, err := page.Normalize()
	if err != nil {
		return Page[*Todo]{}, errors.Join(err, ErrTodoRead)
	}
	plan, err := CompileQuery(TodoListByOwnerQuery{
		OwnerID:   ownerID,
		Page:      normalized,
		Order:     SortDirectionDesc,
		Completed: completed,
	})
	if err != nil {
		return Page[*Todo]{}, errors.Join(err, ErrTodoRead)
	}

	todos := make([]*Todo, 0)
	err = Neo4jExecuteReadPlan(ctx, r.db, plan, func(tx neo4j.ManagedTransaction) error {
		var runErr error
		todos, _, runErr = Neo4jRunQuery(ctx, tx, plan.Root, r.scan("t", "o", "c"))
		return runErr
	})
	if err != nil {
		return Page[*Todo]{}, errors.Join(err, ErrTodoRead)
	}

	return PaginateSlice(todos, normalized.Size, func(todo *Todo) model.ID {
		return todo.ID
	})
}

func (r *Neo4jTodoRepository) Update(ctx context.Context, id model.ID, opts UpdateTodoOpts) (*Todo, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.TodoRepository/Update")
	defer span.End()

	cypher := `
	MATCH (t:` + id.Label() + ` {id: $id})
	SET t += $patch, t.updated_at = datetime()
	RETURN t.id AS id`

	params := map[string]any{
		"id":    id.String(),
		"patch": opts.patch(),
	}

	if _, err := Neo4jExecuteWriteAndReadSingle(ctx, r.db, cypher, params, func(_ *neo4j.Record) (*struct{}, error) {
		return &struct{}{}, nil
	}); err != nil {
		return nil, errors.Join(ErrTodoUpdate, err)
	}

	return r.Get(ctx, id)
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
	pattern := composeCacheKey(model.ResourceTypeTodo.String(), "ListByOwner", opts.OwnedBy.String(), "*")
	if err := r.cacheRepo.DeletePattern(ctx, pattern); err != nil {
		return nil, err
	}

	return r.todoRepo.Create(ctx, opts)
}

func (r *RedisCachedTodoRepository) Get(ctx context.Context, id model.ID) (*Todo, error) {
	var todo *Todo
	var err error

	key := composeCacheKey(model.ResourceTypeTodo.String(), "Get", id.String())
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

func (r *RedisCachedTodoRepository) ListByOwner(ctx context.Context, ownerID model.ID, page CursorPage, completed *bool) (Page[*Todo], error) {
	var todos Page[*Todo]
	var err error

	normalized, err := normalizedPage(page)
	if err != nil {
		return Page[*Todo]{}, err
	}

	key := composeCacheKey(model.ResourceTypeTodo.String(), "ListByOwner", ownerID.String(), pageTokenValue(normalized.Token), normalized.Size, completed)
	if err = r.cacheRepo.Get(ctx, key, &todos); err != nil {
		return Page[*Todo]{}, err
	}

	if todos.Items != nil {
		return todos, nil
	}

	todos, err = r.todoRepo.ListByOwner(ctx, ownerID, normalized, completed)
	if err != nil {
		return Page[*Todo]{}, err
	}

	if err = r.cacheRepo.Set(ctx, key, todos); err != nil {
		return Page[*Todo]{}, err
	}

	return todos, nil
}

func (r *RedisCachedTodoRepository) Update(ctx context.Context, id model.ID, opts UpdateTodoOpts) (*Todo, error) {
	todo, err := r.todoRepo.Update(ctx, id, opts)
	if err != nil {
		return nil, err
	}

	key := composeCacheKey(model.ResourceTypeTodo.String(), "Get", id.String())
	if err = r.cacheRepo.Set(ctx, key, todo); err != nil {
		return nil, err
	}

	pattern := composeCacheKey(model.ResourceTypeTodo.String(), "ListByOwner", todo.OwnedBy.String(), "*")
	if err := r.cacheRepo.DeletePattern(ctx, pattern); err != nil {
		return nil, err
	}

	return todo, nil
}

func (r *RedisCachedTodoRepository) Delete(ctx context.Context, id model.ID) error {
	key := composeCacheKey(model.ResourceTypeTodo.String(), "Get", id.String()+"*")
	if err := r.cacheRepo.DeletePattern(ctx, key); err != nil {
		return err
	}

	pattern := composeCacheKey(model.ResourceTypeTodo.String(), "ListByOwner", "*")
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

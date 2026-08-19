package repository

import (
	"context"
	"errors"
	"time"

	"github.com/neo4j/neo4j-go-driver/v6/neo4j"

	"github.com/opcotech/elemo/internal/model"
)

var (
	ErrAssignmentCreate = errors.New("failed to create assignment") // the assignment could not be created
	ErrAssignmentDelete = errors.New("failed to delete assignment") // the assignment could not be deleted
	ErrAssignmentRead   = errors.New("failed to read assignment")   // the assignment could not be retrieved
)

// Assignment represents an assignment persisted by the repository.
type Assignment struct {
	ID        model.ID             `json:"id"`
	Kind      model.AssignmentKind `json:"kind"`
	User      model.ID             `json:"user_id"`
	Resource  model.ID             `json:"resource_id"`
	CreatedAt *time.Time           `json:"created_at"`
}

// CreateAssignmentOpts holds the data required to create an assignment.
type CreateAssignmentOpts struct {
	Kind     model.AssignmentKind
	User     model.ID
	Resource model.ID
}

//go:generate go tool mockgen -source=assignment.go -destination=assignment_mock_gen.go -package=repository -mock_names "AssignmentRepository=MockAssignmentRepository"
type AssignmentRepository interface {
	Create(ctx context.Context, opts CreateAssignmentOpts) (*Assignment, error)
	Get(ctx context.Context, id model.ID, proj AssignmentProjection) (*Assignment, error)
	ListByUser(ctx context.Context, userID model.ID, page CursorPage, proj AssignmentProjection) (Page[*Assignment], error)
	ListByResource(ctx context.Context, resourceID model.ID, page CursorPage, proj AssignmentProjection) (Page[*Assignment], error)
	Delete(ctx context.Context, id model.ID) error
}

// Neo4jAssignmentRepository is a repository for managing user assignments.
type Neo4jAssignmentRepository struct {
	*neo4jBaseRepository
}

func (r *Neo4jAssignmentRepository) scan(up, ap, rp string) func(rec *neo4j.Record) (*Assignment, error) {
	return func(rec *neo4j.Record) (*Assignment, error) {
		a := new(Assignment)

		val, _, err := neo4j.GetRecordValue[neo4j.Relationship](rec, ap)
		if err != nil {
			return nil, err
		}

		user, _, err := neo4j.GetRecordValue[neo4j.Node](rec, up)
		if err != nil {
			return nil, err
		}

		resource, _, err := neo4j.GetRecordValue[neo4j.Node](rec, rp)
		if err != nil {
			return nil, err
		}

		if err := Neo4jScanIntoStruct(&val, &a, []string{"id"}); err != nil {
			return nil, err
		}

		a.ID, err = model.NewIDFromString(val.GetProperties()["id"].(string), model.ResourceTypeAssignment.String())
		if err != nil {
			return nil, err
		}
		a.User, err = model.NewIDFromString(user.GetProperties()["id"].(string), domainLabel(user.Labels))
		if err != nil {
			return nil, err
		}
		a.Resource, err = model.NewIDFromString(resource.GetProperties()["id"].(string), domainLabel(resource.Labels))
		if err != nil {
			return nil, err
		}

		return a, nil
	}
}

func (r *Neo4jAssignmentRepository) Create(ctx context.Context, opts CreateAssignmentOpts) (*Assignment, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.AssignmentRepository/Create")
	defer span.End()

	createdAt := time.Now().UTC()
	id := model.MustNewID(model.ResourceTypeAssignment)

	cypher := `
	MATCH (u:` + opts.User.Label() + ` {id: $user_id})
	MATCH (r:` + opts.Resource.Label() + ` {id: $resource_id})
	MERGE (u)-[a:` + EdgeKindAssignedTo.String() + ` {kind: $kind}]->(r)
	ON CREATE SET a.id = $id, a.created_at = datetime($created_at)
	RETURN a.id AS id`

	params := map[string]any{
		"id":          id.String(),
		"user_id":     opts.User.String(),
		"resource_id": opts.Resource.String(),
		"kind":        opts.Kind.String(),
		"created_at":  createdAt.Format(time.RFC3339Nano),
	}

	storedID, err := Neo4jExecuteWriteAndReadSingle(ctx, r.db, cypher, params, func(rec *neo4j.Record) (*model.ID, error) {
		raw, parseErr := Neo4jParseValueFromRecord[string](rec, "id")
		if parseErr != nil {
			return nil, parseErr
		}
		parsed, parseErr := model.NewIDFromString(raw, model.ResourceTypeAssignment.String())
		if parseErr != nil {
			return nil, parseErr
		}
		return &parsed, nil
	})
	if err != nil {
		return nil, errors.Join(ErrAssignmentCreate, err)
	}

	return r.Get(ctx, *storedID, AssignmentDetailProjection())
}

func (r *Neo4jAssignmentRepository) Get(ctx context.Context, id model.ID, proj AssignmentProjection) (*Assignment, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.AssignmentRepository/Get")
	defer span.End()

	plan, err := CompileQuery(AssignmentGetQuery{ID: id, Projection: proj})
	if err != nil {
		return nil, errors.Join(ErrAssignmentRead, err)
	}

	var assignment *Assignment
	err = Neo4jExecuteReadPlan(ctx, r.db, plan, func(tx neo4j.ManagedTransaction) error {
		var runErr error
		assignment, _, runErr = Neo4jRunQuerySingle(ctx, tx, plan.Root, r.scan("u", "a", "r"))
		return runErr
	})
	if err != nil {
		return nil, errors.Join(ErrAssignmentRead, err)
	}

	return assignment, nil
}

func (r *Neo4jAssignmentRepository) ListByUser(ctx context.Context, userID model.ID, page CursorPage, proj AssignmentProjection) (Page[*Assignment], error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.AssignmentRepository/ListByUser")
	defer span.End()

	normalized, err := page.Normalize()
	if err != nil {
		return Page[*Assignment]{}, errors.Join(ErrAssignmentRead, err)
	}
	plan, err := CompileQuery(AssignmentListByUserQuery{
		UserID:     userID,
		Page:       normalized,
		Order:      SortDirectionDesc,
		Projection: proj,
	})
	if err != nil {
		return Page[*Assignment]{}, errors.Join(ErrAssignmentRead, err)
	}

	assignments := make([]*Assignment, 0)
	err = Neo4jExecuteReadPlan(ctx, r.db, plan, func(tx neo4j.ManagedTransaction) error {
		var runErr error
		assignments, _, runErr = Neo4jRunQuery(ctx, tx, plan.Root, r.scan("u", "a", "r"))
		return runErr
	})
	if err != nil {
		return Page[*Assignment]{}, errors.Join(ErrAssignmentRead, err)
	}

	return PaginateSlice(assignments, normalized.Size, func(assignment *Assignment) model.ID {
		return assignment.ID
	})
}

func (r *Neo4jAssignmentRepository) ListByResource(ctx context.Context, resourceID model.ID, page CursorPage, proj AssignmentProjection) (Page[*Assignment], error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.AssignmentRepository/ListByResource")
	defer span.End()

	normalized, err := page.Normalize()
	if err != nil {
		return Page[*Assignment]{}, errors.Join(ErrAssignmentRead, err)
	}
	plan, err := CompileQuery(AssignmentListByResourceQuery{
		ResourceID: resourceID,
		Page:       normalized,
		Order:      SortDirectionDesc,
		Projection: proj,
	})
	if err != nil {
		return Page[*Assignment]{}, errors.Join(ErrAssignmentRead, err)
	}

	assignments := make([]*Assignment, 0)
	err = Neo4jExecuteReadPlan(ctx, r.db, plan, func(tx neo4j.ManagedTransaction) error {
		var runErr error
		assignments, _, runErr = Neo4jRunQuery(ctx, tx, plan.Root, r.scan("u", "a", "r"))
		return runErr
	})
	if err != nil {
		return Page[*Assignment]{}, errors.Join(ErrAssignmentRead, err)
	}

	return PaginateSlice(assignments, normalized.Size, func(assignment *Assignment) model.ID {
		return assignment.ID
	})
}

func (r *Neo4jAssignmentRepository) Delete(ctx context.Context, id model.ID) error {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.AssignmentRepository/Delete")
	defer span.End()

	cypher := `MATCH (u)-[a:` + EdgeKindAssignedTo.String() + ` {id: $id}]->(r) DELETE a`
	params := map[string]any{
		"id": id.String(),
	}

	if err := Neo4jExecuteWriteAndConsume(ctx, r.db, cypher, params); err != nil {
		return errors.Join(ErrAssignmentDelete, err)
	}

	return nil
}

// NewNeo4jAssignmentRepository creates a new assignment neo4jBaseRepository.
func NewNeo4jAssignmentRepository(opts ...Neo4jRepositoryOption) (*Neo4jAssignmentRepository, error) {
	baseRepo, err := newNeo4jRepository(opts...)
	if err != nil {
		return nil, err
	}

	return &Neo4jAssignmentRepository{
		neo4jBaseRepository: baseRepo,
	}, nil
}

func clearAssignmentsKey(ctx context.Context, r *redisBaseRepository, id model.ID) error {
	return clearAssignmentsPattern(ctx, r, "Get", id.String(), "*")
}

func clearAssignmentsPattern(ctx context.Context, r *redisBaseRepository, pattern ...string) error {
	return r.DeletePattern(ctx, composeCacheKey(model.ResourceTypeAssignment.String(), pattern))
}

func clearAssignmentByResource(ctx context.Context, r *redisBaseRepository, resourceID model.ID) error {
	return clearAssignmentsPattern(ctx, r, "ListByResource", resourceID.String(), "*", "*")
}

func clearAssignmentAllByResource(ctx context.Context, r *redisBaseRepository) error {
	return clearAssignmentsPattern(ctx, r, "ListByResource", "*", "*", "*")
}

func clearAssignmentByUser(ctx context.Context, r *redisBaseRepository, userID model.ID) error {
	return clearAssignmentsPattern(ctx, r, "ListByUser", userID.String(), "*", "*")
}

func clearAssignmentAllByUser(ctx context.Context, r *redisBaseRepository) error {
	return clearAssignmentsPattern(ctx, r, "ListByUser", "*", "*", "*")
}

func clearAssignmentAllCrossCache(ctx context.Context, r *redisBaseRepository, assignment *Assignment) error {
	var deleteFn func(ctx context.Context, r *redisBaseRepository, pattern ...string) error

	if assignment == nil {
		deleteFn = clearIssuesPattern
	} else {
		switch assignment.Resource.Type {
		case model.ResourceTypeIssue:
			deleteFn = clearIssuesPattern
		default:
			return ErrUnexpectedCachedResource
		}
	}

	return deleteFn(ctx, r, "*")
}

// RedisCachedAssignmentRepository implements caching on the AssignmentRepository.
type RedisCachedAssignmentRepository struct {
	cacheRepo      *redisBaseRepository
	assignmentRepo AssignmentRepository
}

func (r *RedisCachedAssignmentRepository) Create(ctx context.Context, opts CreateAssignmentOpts) (*Assignment, error) {
	if err := clearAssignmentByResource(ctx, r.cacheRepo, opts.Resource); err != nil {
		return nil, err
	}

	if err := clearAssignmentByUser(ctx, r.cacheRepo, opts.User); err != nil {
		return nil, err
	}

	assignment := &Assignment{User: opts.User, Resource: opts.Resource}
	if err := clearAssignmentAllCrossCache(ctx, r.cacheRepo, assignment); err != nil {
		return nil, err
	}

	return r.assignmentRepo.Create(ctx, opts)
}

func (r *RedisCachedAssignmentRepository) Get(ctx context.Context, id model.ID, proj AssignmentProjection) (*Assignment, error) {
	var assignment *Assignment
	var err error

	key := composeCacheKey(model.ResourceTypeAssignment.String(), "Get", id.String(), projectionCacheValue(proj))
	if err = r.cacheRepo.Get(ctx, key, &assignment); err != nil {
		return nil, err
	}

	if assignment != nil {
		return assignment, nil
	}

	if assignment, err = r.assignmentRepo.Get(ctx, id, proj); err != nil {
		return nil, err
	}

	if err = r.cacheRepo.Set(ctx, key, assignment); err != nil {
		return nil, err
	}

	return assignment, nil
}

func (r *RedisCachedAssignmentRepository) ListByUser(ctx context.Context, userID model.ID, page CursorPage, proj AssignmentProjection) (Page[*Assignment], error) {
	var assignments Page[*Assignment]
	var err error

	normalized, err := normalizedPage(page)
	if err != nil {
		return Page[*Assignment]{}, err
	}

	key := composeCacheKey(
		model.ResourceTypeAssignment.String(),
		"ListByUser",
		userID.String(),
		projectionCacheValue(proj),
		pageTokenValue(normalized.Token),
		normalized.Size,
	)
	if err = r.cacheRepo.Get(ctx, key, &assignments); err != nil {
		return Page[*Assignment]{}, err
	}

	if assignments.Items != nil {
		return assignments, nil
	}

	if assignments, err = r.assignmentRepo.ListByUser(ctx, userID, normalized, proj); err != nil {
		return Page[*Assignment]{}, err
	}

	if err = r.cacheRepo.Set(ctx, key, assignments); err != nil {
		return Page[*Assignment]{}, err
	}

	return assignments, nil
}

func (r *RedisCachedAssignmentRepository) ListByResource(ctx context.Context, resourceID model.ID, page CursorPage, proj AssignmentProjection) (Page[*Assignment], error) {
	var assignments Page[*Assignment]
	var err error

	normalized, err := normalizedPage(page)
	if err != nil {
		return Page[*Assignment]{}, err
	}

	key := composeCacheKey(
		model.ResourceTypeAssignment.String(),
		"ListByResource",
		resourceID.String(),
		projectionCacheValue(proj),
		pageTokenValue(normalized.Token),
		normalized.Size,
	)
	if err = r.cacheRepo.Get(ctx, key, &assignments); err != nil {
		return Page[*Assignment]{}, err
	}

	if assignments.Items != nil {
		return assignments, nil
	}

	if assignments, err = r.assignmentRepo.ListByResource(ctx, resourceID, normalized, proj); err != nil {
		return Page[*Assignment]{}, err
	}

	if err = r.cacheRepo.Set(ctx, key, assignments); err != nil {
		return Page[*Assignment]{}, err
	}

	return assignments, nil
}

func (r *RedisCachedAssignmentRepository) Delete(ctx context.Context, id model.ID) error {
	if err := clearAssignmentsKey(ctx, r.cacheRepo, id); err != nil {
		return err
	}

	if err := clearAssignmentAllByResource(ctx, r.cacheRepo); err != nil {
		return err
	}

	if err := clearAssignmentAllByUser(ctx, r.cacheRepo); err != nil {
		return err
	}

	if err := clearAssignmentAllCrossCache(ctx, r.cacheRepo, nil); err != nil {
		return err
	}

	return r.assignmentRepo.Delete(ctx, id)
}

// NewCachedAssignmentRepository returns a new CachedAssignmentRepository.
func NewCachedAssignmentRepository(repo AssignmentRepository, opts ...RedisRepositoryOption) (*RedisCachedAssignmentRepository, error) {
	r, err := newRedisBaseRepository(opts...)
	if err != nil {
		return nil, err
	}

	return &RedisCachedAssignmentRepository{
		cacheRepo:      r,
		assignmentRepo: repo,
	}, nil
}

package repository

import (
	"context"
	"errors"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg/convert"
)

var (
	ErrAssignmentCreate = errors.New("failed to create assignment") // the assignment could not be created
	ErrAssignmentDelete = errors.New("failed to delete assignment") // the assignment could not be deleted
	ErrAssignmentRead   = errors.New("failed to read assignment")   // the assignment could not be retrieved
)

// AssignmentRepository is a repository for managing resource assignments.
//
//go:generate mockgen -source=assignment.go -destination=../testutil/mock/assignment_repo_gen.go -package=mock -mock_names "AssignmentRepository=AssignmentRepository"
type AssignmentRepository interface {
	Create(ctx context.Context, assignment *model.Assignment) error
	Get(ctx context.Context, id model.ID) (*model.Assignment, error)
	GetByUser(ctx context.Context, userID model.ID, offset, limit int) ([]*model.Assignment, error)
	GetByResource(ctx context.Context, resourceID model.ID, offset, limit int) ([]*model.Assignment, error)
	Delete(ctx context.Context, id model.ID) error
}

// Neo4jAssignmentRepository is a repository for managing user assignments.
type Neo4jAssignmentRepository struct {
	*neo4jBaseRepository
}

func (r *Neo4jAssignmentRepository) scan(up, ap, rp string) func(rec *neo4j.Record) (*model.Assignment, error) {
	return func(rec *neo4j.Record) (*model.Assignment, error) {
		a := new(model.Assignment)

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

		a.ID, _ = model.NewIDFromString(val.GetProperties()["id"].(string), model.ResourceTypeAssignment.String())
		a.User, _ = model.NewIDFromString(user.GetProperties()["id"].(string), user.Labels[0])
		a.Resource, _ = model.NewIDFromString(resource.GetProperties()["id"].(string), resource.Labels[0])

		if err := a.Validate(); err != nil {
			return nil, err
		}

		return a, nil
	}
}

func (r *Neo4jAssignmentRepository) Create(ctx context.Context, assignment *model.Assignment) error {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.AssignmentRepository/Create")
	defer span.End()

	if err := assignment.Validate(); err != nil {
		return errors.Join(ErrAssignmentCreate, err)
	}

	createdAt := time.Now().UTC()

	assignment.ID = model.MustNewID(model.ResourceTypeAssignment)
	assignment.CreatedAt = convert.ToPointer(createdAt)

	cypher := `
	MATCH (u:` + assignment.User.Label() + ` {id: $user_id})
	MATCH (r:` + assignment.Resource.Label() + ` {id: $resource_id})
	MERGE (u)-[a:` + EdgeKindAssignedTo.String() + ` {kind: $kind}]->(r)
	ON CREATE SET a.id = $id, a.created_at = datetime($created_at)`

	params := map[string]any{
		"id":          assignment.ID.String(),
		"user_id":     assignment.User.String(),
		"resource_id": assignment.Resource.String(),
		"kind":        assignment.Kind.String(),
		"created_at":  assignment.CreatedAt.Format(time.RFC3339Nano),
	}

	if err := Neo4jExecuteWriteAndConsume(ctx, r.db, cypher, params); err != nil {
		return errors.Join(ErrAssignmentCreate, err)
	}

	return nil
}

func (r *Neo4jAssignmentRepository) Get(ctx context.Context, id model.ID) (*model.Assignment, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.AssignmentRepository/Get")
	defer span.End()

	cypher := `
	MATCH (u)-[a:` + EdgeKindAssignedTo.String() + ` {id: $id}]->(r)
	RETURN u, a, r`

	params := map[string]any{
		"id": id.String(),
	}

	assignment, err := Neo4jExecuteReadAndReadSingle(ctx, r.db, cypher, params, r.scan("u", "a", "r"))
	if err != nil {
		return nil, errors.Join(ErrAssignmentRead, err)
	}

	return assignment, nil
}

func (r *Neo4jAssignmentRepository) GetByUser(ctx context.Context, userID model.ID, offset, limit int) ([]*model.Assignment, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.AssignmentRepository/GetByUser")
	defer span.End()

	cypher := `
	MATCH (u:` + userID.Label() + ` {id: $user_id})-[a:` + EdgeKindAssignedTo.String() + `]->(r)
	RETURN u, a, r
	ORDER BY a.created_at DESC
	SKIP $offset LIMIT $limit`

	params := map[string]any{
		"user_id": userID.String(),
		"offset":  offset,
		"limit":   limit,
	}

	assignments, err := Neo4jExecuteReadAndReadAll(ctx, r.db, cypher, params, r.scan("u", "a", "r"))
	if err != nil {
		return nil, errors.Join(ErrAssignmentRead, err)
	}

	return assignments, nil
}

func (r *Neo4jAssignmentRepository) GetByResource(ctx context.Context, resourceID model.ID, offset, limit int) ([]*model.Assignment, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.AssignmentRepository/GetByResource")
	defer span.End()

	cypher := `
	MATCH (u)-[a:` + EdgeKindAssignedTo.String() + `]->(r:` + resourceID.Label() + ` {id: $resource_id})
	RETURN u, a, r
	ORDER BY a.created_at DESC
	SKIP $offset LIMIT $limit`

	params := map[string]any{
		"resource_id": resourceID.String(),
		"offset":      offset,
		"limit":       limit,
	}

	assignments, err := Neo4jExecuteReadAndReadAll(ctx, r.db, cypher, params, r.scan("u", "a", "r"))
	if err != nil {
		return nil, errors.Join(ErrAssignmentRead, err)
	}

	return assignments, nil
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
	return r.Delete(ctx, composeCacheKey(model.ResourceTypeAssignment.String(), id.String()))
}

func clearAssignmentsPattern(ctx context.Context, r *redisBaseRepository, pattern ...string) error {
	return r.DeletePattern(ctx, composeCacheKey(model.ResourceTypeAssignment.String(), pattern))
}

func clearAssignmentByResource(ctx context.Context, r *redisBaseRepository, resourceID model.ID) error {
	return clearAssignmentsPattern(ctx, r, "GetByResource", resourceID.String(), "*")
}

func clearAssignmentAllByResource(ctx context.Context, r *redisBaseRepository) error {
	return clearAssignmentsPattern(ctx, r, "GetByResource", "*")
}

func clearAssignmentByUser(ctx context.Context, r *redisBaseRepository, userID model.ID) error {
	return clearAssignmentsPattern(ctx, r, "GetByUser", userID.String(), "*")
}

func clearAssignmentAllByUser(ctx context.Context, r *redisBaseRepository) error {
	return clearAssignmentsPattern(ctx, r, "GetByUser", "*")
}

func clearAssignmentAllCrossCache(ctx context.Context, r *redisBaseRepository, assignment *model.Assignment) error {
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

func (r *RedisCachedAssignmentRepository) Create(ctx context.Context, assignment *model.Assignment) error {
	if err := clearAssignmentByResource(ctx, r.cacheRepo, assignment.Resource); err != nil {
		return err
	}

	if err := clearAssignmentByUser(ctx, r.cacheRepo, assignment.User); err != nil {
		return err
	}

	if err := clearAssignmentAllCrossCache(ctx, r.cacheRepo, assignment); err != nil {
		return err
	}

	return r.assignmentRepo.Create(ctx, assignment)
}

func (r *RedisCachedAssignmentRepository) Get(ctx context.Context, id model.ID) (*model.Assignment, error) {
	var assignment *model.Assignment
	var err error

	key := composeCacheKey(model.ResourceTypeAssignment.String(), id.String())
	if err = r.cacheRepo.Get(ctx, key, &assignment); err != nil {
		return nil, err
	}

	if assignment != nil {
		return assignment, nil
	}

	if assignment, err = r.assignmentRepo.Get(ctx, id); err != nil {
		return nil, err
	}

	if err = r.cacheRepo.Set(ctx, key, assignment); err != nil {
		return nil, err
	}

	return assignment, nil
}

func (r *RedisCachedAssignmentRepository) GetByUser(ctx context.Context, userID model.ID, offset, limit int) ([]*model.Assignment, error) {
	var assignments []*model.Assignment
	var err error

	key := composeCacheKey(model.ResourceTypeAssignment.String(), "GetByUser", userID.String(), offset, limit)
	if err = r.cacheRepo.Get(ctx, key, &assignments); err != nil {
		return nil, err
	}

	if assignments != nil {
		return assignments, nil
	}

	if assignments, err = r.assignmentRepo.GetByUser(ctx, userID, offset, limit); err != nil {
		return nil, err
	}

	if err = r.cacheRepo.Set(ctx, key, assignments); err != nil {
		return nil, err
	}

	return assignments, nil
}

func (r *RedisCachedAssignmentRepository) GetByResource(ctx context.Context, resourceID model.ID, offset, limit int) ([]*model.Assignment, error) {
	var assignments []*model.Assignment
	var err error

	key := composeCacheKey(model.ResourceTypeAssignment.String(), "GetByResource", resourceID.String(), offset, limit)
	if err = r.cacheRepo.Get(ctx, key, &assignments); err != nil {
		return nil, err
	}

	if assignments != nil {
		return assignments, nil
	}

	if assignments, err = r.assignmentRepo.GetByResource(ctx, resourceID, offset, limit); err != nil {
		return nil, err
	}

	if err = r.cacheRepo.Set(ctx, key, assignments); err != nil {
		return nil, err
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

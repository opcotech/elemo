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

// AssignmentRepository is a repository for managing user assignments.
type AssignmentRepository struct {
	*baseRepository
}

func (r *AssignmentRepository) scan(up, ap, rp string) func(rec *neo4j.Record) (*model.Assignment, error) {
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

		if err := ScanIntoStruct(&val, &a, []string{"id"}); err != nil {
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

func (r *AssignmentRepository) Create(ctx context.Context, assignment *model.Assignment) error {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.AssignmentRepository/Create")
	defer span.End()

	if err := assignment.Validate(); err != nil {
		return errors.Join(repository.ErrAssignmentCreate, err)
	}

	createdAt := time.Now()

	assignment.ID = model.MustNewID(model.ResourceTypeAssignment)
	assignment.CreatedAt = convert.ToPointer(createdAt)

	cypher := `
	MATCH (u:` + assignment.User.Label() + ` {id: $user_id}), (r:` + assignment.Resource.Label() + ` {id: $resource_id})
	MERGE (u)-[a:` + EdgeKindAssignedTo.String() + ` {kind: $kind}]->(r)
	ON CREATE SET a.id = $id, a.created_at = datetime($created_at)`

	params := map[string]any{
		"id":          assignment.ID.String(),
		"user_id":     assignment.User.String(),
		"resource_id": assignment.Resource.String(),
		"kind":        assignment.Kind.String(),
		"created_at":  assignment.CreatedAt.Format(time.RFC3339Nano),
	}

	if err := ExecuteWriteAndConsume(ctx, r.db, cypher, params); err != nil {
		return errors.Join(repository.ErrAssignmentCreate, err)
	}

	return nil
}

func (r *AssignmentRepository) Get(ctx context.Context, id model.ID) (*model.Assignment, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.AssignmentRepository/Get")
	defer span.End()

	cypher := `
	MATCH (u)-[a:` + EdgeKindAssignedTo.String() + ` {id: $id}]->(r)
	RETURN u, a, r`

	params := map[string]any{
		"id": id.String(),
	}

	assignment, err := ExecuteReadAndReadSingle(ctx, r.db, cypher, params, r.scan("u", "a", "r"))
	if err != nil {
		return nil, errors.Join(repository.ErrAssignmentRead, err)
	}

	return assignment, nil
}

func (r *AssignmentRepository) GetByUser(ctx context.Context, userID model.ID, offset, limit int) ([]*model.Assignment, error) {
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

	assignments, err := ExecuteReadAndReadAll(ctx, r.db, cypher, params, r.scan("u", "a", "r"))
	if err != nil {
		return nil, errors.Join(repository.ErrAssignmentRead, err)
	}

	return assignments, nil
}

func (r *AssignmentRepository) GetByResource(ctx context.Context, resourceID model.ID, offset, limit int) ([]*model.Assignment, error) {
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

	assignments, err := ExecuteReadAndReadAll(ctx, r.db, cypher, params, r.scan("u", "a", "r"))
	if err != nil {
		return nil, errors.Join(repository.ErrAssignmentRead, err)
	}

	return assignments, nil
}

func (r *AssignmentRepository) Delete(ctx context.Context, id model.ID) error {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.AssignmentRepository/Delete")
	defer span.End()

	cypher := `MATCH (u)-[a:` + EdgeKindAssignedTo.String() + ` {id: $id}]->(r) DELETE a`
	params := map[string]any{
		"id": id.String(),
	}

	if err := ExecuteWriteAndConsume(ctx, r.db, cypher, params); err != nil {
		return errors.Join(repository.ErrAssignmentDelete, err)
	}

	return nil
}

// NewAssignmentRepository creates a new assignment baseRepository.
func NewAssignmentRepository(opts ...RepositoryOption) (*AssignmentRepository, error) {
	baseRepo, err := newRepository(opts...)
	if err != nil {
		return nil, err
	}

	return &AssignmentRepository{
		baseRepository: baseRepo,
	}, nil
}

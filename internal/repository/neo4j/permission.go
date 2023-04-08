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

// PermissionRepository is a baseRepository for managing permissions.
type PermissionRepository struct {
	*baseRepository
}

// scan is a helper function for scanning a permission from a neo4j.Record.
func (r *PermissionRepository) scan(permParam, subjectParam, targetParam string) func(rec *neo4j.Record) (*model.Permission, error) {
	return func(rec *neo4j.Record) (*model.Permission, error) {
		parsed := new(model.Permission)

		val, _, err := neo4j.GetRecordValue[neo4j.Relationship](rec, permParam)
		if err != nil {
			return nil, err
		}

		subject, _, err := neo4j.GetRecordValue[neo4j.Node](rec, subjectParam)
		if err != nil {
			return nil, err
		}

		target, _, err := neo4j.GetRecordValue[neo4j.Node](rec, targetParam)
		if err != nil {
			return nil, err
		}

		if err := ScanIntoStruct(&val, &parsed, []string{"id"}); err != nil {
			return nil, err
		}

		parsed.ID, _ = model.NewIDFromString(val.GetProperties()["id"].(string), EdgeKindHasPermission.String())
		parsed.Subject, _ = model.NewIDFromString(subject.GetProperties()["id"].(string), subject.Labels[0])
		parsed.Target, _ = model.NewIDFromString(target.GetProperties()["id"].(string), target.Labels[0])

		if err := parsed.Validate(); err != nil {
			return nil, err
		}

		return parsed, nil
	}
}

// Create creates a new permission if it does not already exist between the
// subject and target. If the permission already exists, no action is taken.
func (r *PermissionRepository) Create(ctx context.Context, perm *model.Permission) error {
	ctx, span := r.tracer.Start(ctx, "baseRepository.neo4j.PermissionRepository/Create")
	defer span.End()

	if err := perm.Validate(); err != nil {
		return errors.Join(repository.ErrPermissionCreate, err)
	}

	perm.ID = model.MustNewID(EdgeKindHasPermission.String())
	perm.CreatedAt = convert.ToPointer(time.Now())
	perm.UpdatedAt = nil

	cypher := `
	MATCH (subject:` + perm.Subject.Label() + ` {id: $subject}), (target:` + perm.Target.Label() + ` {id: $target})
	MERGE (subject)-[p:` + perm.ID.Label() + ` {id: $id, kind: $kind}]->(target) ON CREATE SET p.created_at = datetime($created_at)
	`

	params := map[string]any{
		"id":         perm.ID.String(),
		"subject":    perm.Subject.String(),
		"target":     perm.Target.String(),
		"kind":       perm.Kind.String(),
		"created_at": perm.CreatedAt.Format(time.RFC3339Nano),
	}

	if err := ExecuteWriteAndConsume(ctx, r.db, cypher, params); err != nil {
		return errors.Join(err, repository.ErrPermissionCreate)
	}

	return nil
}

// Get returns an existing permission, its subject and target. If the
// permission does not exist, an error is returned.
func (r *PermissionRepository) Get(ctx context.Context, id model.ID) (*model.Permission, error) {
	ctx, span := r.tracer.Start(ctx, "baseRepository.neo4j.PermissionRepository/Get")
	defer span.End()

	cypher := `
	MATCH (s)-[p:` + id.Label() + ` {id: $id}]->(t)
	RETURN s, p, t
	`

	params := map[string]any{
		"id": id.String(),
	}

	perm, err := ExecuteWriteAndReadSingle(ctx, r.db, cypher, params, r.scan("p", "s", "t"))
	if err != nil {
		return nil, errors.Join(err, repository.ErrPermissionRead)
	}

	return perm, nil
}

// GetBySubject returns all permissions for a given subject. If no permissions
// exist, an empty slice is returned.
func (r *PermissionRepository) GetBySubject(ctx context.Context, id model.ID) ([]*model.Permission, error) {
	ctx, span := r.tracer.Start(ctx, "baseRepository.neo4j.PermissionRepository/GetBySubject")
	defer span.End()

	cypher := `
	MATCH (s:` + id.Label() + ` {id: $id})-[p:` + EdgeKindHasPermission.String() + `]->(t)
	RETURN s, p, t
	ORDER BY p.created_at DESC`

	params := map[string]any{
		"id": id.String(),
	}

	perms, err := ExecuteReadAndReadAll(ctx, r.db, cypher, params, r.scan("p", "s", "t"))
	if err != nil {
		return nil, errors.Join(err, repository.ErrPermissionRead)
	}

	return perms, nil
}

// GetByTarget returns all permissions for a given target. If no permissions
// exist, an empty slice is returned.
func (r *PermissionRepository) GetByTarget(ctx context.Context, id model.ID) ([]*model.Permission, error) {
	ctx, span := r.tracer.Start(ctx, "baseRepository.neo4j.PermissionRepository/GetByTarget")
	defer span.End()

	cypher := `
	MATCH (s)-[p:` + EdgeKindHasPermission.String() + `]->(t:` + id.Label() + ` {id: $id})
	RETURN s, p, t
	ORDER BY p.created_at DESC`

	params := map[string]any{
		"id": id.String(),
	}

	perms, err := ExecuteReadAndReadAll(ctx, r.db, cypher, params, r.scan("p", "s", "t"))
	if err != nil {
		return nil, errors.Join(err, repository.ErrPermissionRead)
	}

	return perms, nil
}

func (r *PermissionRepository) HasPermission(ctx context.Context, subject, target model.ID, kind model.PermissionKind) (bool, error) {
	ctx, span := r.tracer.Start(ctx, "baseRepository.neo4j.PermissionRepository/HasPermission")
	defer span.End()

	return r.HasAnyPermission(ctx, subject, target, kind, model.PermissionKindAll)
}

func (r *PermissionRepository) HasAnyPermission(ctx context.Context, subject, target model.ID, kinds ...model.PermissionKind) (bool, error) {
	ctx, span := r.tracer.Start(ctx, "baseRepository.neo4j.PermissionRepository/HasAnyPermission")
	defer span.End()

	permissions := make([]string, len(kinds))
	for i, kind := range kinds {
		permissions[i] = kind.String()
	}

	cypher := `
	OPTIONAL MATCH (s:` + subject.Label() + ` {id: $subject})-[p:` + EdgeKindHasPermission.String() + `]->(t:` + target.Label() + ` {id: $target})
	WHERE p.kind IN $permissions
	RETURN count(p) >= 1 as has_permission`

	params := map[string]any{
		"subject":     subject.String(),
		"target":      target.String(),
		"permissions": permissions,
	}

	hasPermission, err := ExecuteReadAndReadSingle(ctx, r.db, cypher, params, func(rec *neo4j.Record) (*bool, error) {
		val, _, err := neo4j.GetRecordValue[bool](rec, "has_permission")
		if err != nil {
			return nil, err
		}
		return &val, nil
	})
	if err != nil {
		return false, errors.Join(repository.ErrPermissionRead, err)
	}

	return *hasPermission, nil
}

// Update updates an existing permission's kind. If the permission does not
// exist, an error is returned. If the permission's kind is already the same
// as the one provided, the kind is overwritten and the updated_at timestamp
// is updated.
func (r *PermissionRepository) Update(ctx context.Context, id model.ID, kind model.PermissionKind) (*model.Permission, error) {
	ctx, span := r.tracer.Start(ctx, "baseRepository.neo4j.PermissionRepository/Update")
	defer span.End()

	cypher := `
	MATCH (s)-[p:` + id.Label() + ` {id: $id}]->(t)
	SET p.kind = $kind, p.updated_at = datetime($updated_at)
	RETURN s, p, t
	`

	params := map[string]any{
		"id":         id.String(),
		"kind":       kind.String(),
		"updated_at": time.Now().Format(time.RFC3339Nano),
	}

	perm, err := ExecuteWriteAndReadSingle(ctx, r.db, cypher, params, r.scan("p", "s", "t"))
	if err != nil {
		return nil, errors.Join(err, repository.ErrPermissionUpdate)
	}

	return perm, nil
}

// Delete deletes an existing permission. If the permission does not exist, no
// errors are returned.
func (r *PermissionRepository) Delete(ctx context.Context, id model.ID) error {
	ctx, span := r.tracer.Start(ctx, "baseRepository.neo4j.PermissionRepository/Delete")
	defer span.End()

	cypher := `MATCH (s)-[p:` + id.Label() + ` {id: $id}]->(t) DELETE p`

	params := map[string]any{
		"id": id.String(),
	}

	if err := ExecuteWriteAndConsume(ctx, r.db, cypher, params); err != nil {
		return errors.Join(err, repository.ErrPermissionDelete)
	}

	return nil
}

// NewPermissionRepository creates a new permission baseRepository.
func NewPermissionRepository(opts ...RepositoryOption) (*PermissionRepository, error) {
	baseRepo, err := newRepository(opts...)
	if err != nil {
		return nil, err
	}

	return &PermissionRepository{
		baseRepository: baseRepo,
	}, nil
}

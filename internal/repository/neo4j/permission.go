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

// PermissionRepository is a repository for managing permissions.
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

		parsed.ID, _ = model.NewIDFromString(val.GetProperties()["id"].(string), model.ResourceTypePermission.String())
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
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.PermissionRepository/Create")
	defer span.End()

	if err := perm.Validate(); err != nil {
		return errors.Join(repository.ErrPermissionCreate, err)
	}

	perm.ID = model.MustNewID(model.ResourceTypePermission)
	perm.CreatedAt = convert.ToPointer(time.Now().UTC())
	perm.UpdatedAt = nil

	cypher := `
	MATCH (subject:` + perm.Subject.Label() + ` {id: $subject}), (target:` + perm.Target.Label() + ` {id: $target})
	MERGE (subject)-[p:` + EdgeKindHasPermission.String() + ` {id: $id, kind: $kind}]->(target) ON CREATE SET p.created_at = datetime($created_at)
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
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.PermissionRepository/Get")
	defer span.End()

	cypher := `
	MATCH (s)-[p:` + EdgeKindHasPermission.String() + ` {id: $id}]->(t)
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
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.PermissionRepository/GetBySubject")
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
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.PermissionRepository/GetByTarget")
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

// GetBySubjectAndTarget returns all permissions for a given target that the
// source has. If no permissions exist, an empty slice is returned.
func (r *PermissionRepository) GetBySubjectAndTarget(ctx context.Context, source, target model.ID) ([]*model.Permission, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.PermissionRepository/GetBySubjectAndTarget")
	defer span.End()

	cypher := `
	MATCH (s:` + source.Label() + ` {id: $source})-[p:` + EdgeKindHasPermission.String() + `]->(t:` + target.Label() + ` {id: $target})
	RETURN s, p, t
	ORDER BY p.created_at DESC`

	params := map[string]any{
		"source": source.String(),
		"target": target.String(),
	}

	perms, err := ExecuteReadAndReadAll(ctx, r.db, cypher, params, r.scan("p", "s", "t"))
	if err != nil {
		return nil, errors.Join(err, repository.ErrPermissionRead)
	}

	return perms, nil
}

// HasPermission returns true if the subject has the given permission on the
// target. If the permission does not exist, false is returned.
// TODO: Refactor this code. This is a mess.
func (r *PermissionRepository) HasPermission(ctx context.Context, subject, target model.ID, kinds ...model.PermissionKind) (bool, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.PermissionRepository/HasPermission")
	defer span.End()

	hasCreatePermission := false
	permissions := make([]string, len(kinds))
	for i, kind := range kinds {
		if kind == model.PermissionKindCreate {
			hasCreatePermission = true
		}
		permissions[i] = kind.String()
	}

	var cypher string
	if hasCreatePermission {
		cypher = `
		MATCH (s:` + subject.Label() + ` {id: $subject_id})
		MATCH (rt:` + model.ResourceTypeResourceType.String() + ` {id: $target_label})
		OPTIONAL MATCH (s)-[perm:` + EdgeKindHasPermission.String() + `]->(t) WHERE perm.kind IN $permissions
		WITH s, rt, perm

		OPTIONAL MATCH st=(s)-[:` + EdgeKindHasPermission.String() + `|` + EdgeKindMemberOf.String() + `*..2]->(t)
		OPTIONAL MATCH srt=(s)-[:` + EdgeKindHasPermission.String() + `|` + EdgeKindMemberOf.String() + `*..2]->(rt)
		WITH perm, st, srt
		WHERE any(r IN relationships(srt) WHERE type(r) = "` + EdgeKindHasPermission.String() + `" AND r.kind IN $permissions)

		RETURN perm IS NOT NULL OR srt IS NOT NULL AS has_permission
		LIMIT 1`
	} else {
		cypher = `
		MATCH (s:` + subject.Label() + ` {id: $subject_id})
		MATCH (t:` + target.Label() + ` {id: $target_id})
		MATCH (rt:` + model.ResourceTypeResourceType.String() + ` {id: $target_label})
		OPTIONAL MATCH (s)-[perm:` + EdgeKindHasPermission.String() + `]->(t) WHERE perm.kind IN $permissions
		WITH s, t, rt, perm

		OPTIONAL MATCH st=(s)-[:` + EdgeKindHasPermission.String() + `|` + EdgeKindMemberOf.String() + `*..2]->(t)
		OPTIONAL MATCH srt=(s)-[:` + EdgeKindHasPermission.String() + `|` + EdgeKindMemberOf.String() + `*..2]->(rt)
		WITH perm, st, srt
		WHERE (
			any(r IN relationships(st) WHERE type(r) = "` + EdgeKindHasPermission.String() + `" AND r.kind IN $permissions) OR
			any(r IN relationships(srt) WHERE type(r) = "` + EdgeKindHasPermission.String() + `" AND r.kind IN $permissions)
		)

		RETURN perm IS NOT NULL OR st IS NOT NULL OR srt IS NOT NULL AS has_permission
		LIMIT 1`
	}

	params := map[string]any{
		"subject_id":   subject.String(),
		"target_label": target.Label(),
		"permissions":  permissions,
	}

	if !hasCreatePermission {
		params["target_id"] = target.String()
	}

	hasPermission, err := ExecuteReadAndReadSingle(ctx, r.db, cypher, params, func(rec *neo4j.Record) (*bool, error) {
		val, _, err := neo4j.GetRecordValue[bool](rec, "has_permission")
		if err != nil {
			return nil, err
		}
		return &val, nil
	})
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return false, nil
		}
		return false, errors.Join(repository.ErrPermissionRead, err)
	}

	return *hasPermission, nil
}

// HasAnyRelation returns true if there is a relation between the source and
// target. If there is no relation, false is returned.
func (r *PermissionRepository) HasAnyRelation(ctx context.Context, source, target model.ID) (bool, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.RelationRepository/HasAnyRelation")
	defer span.End()

	if err := source.Validate(); err != nil {
		return false, errors.Join(repository.ErrRelationRead, err)
	}

	if err := target.Validate(); err != nil {
		return false, errors.Join(repository.ErrRelationRead, err)
	}

	cypher := `
	MATCH (s:` + source.Label() + ` {id: $source_id})
	MATCH (t:` + target.Label() + ` {id: $target_id})
	MATCH path = shortestPath((s)-[*]-(t))
	WITH path
	WHERE length(path) > 0
	RETURN count(path) > 0 AS has_relation`

	params := map[string]any{
		"source_id": source.String(),
		"target_id": target.String(),
	}

	hasRelation, err := ExecuteReadAndReadSingle(ctx, r.db, cypher, params, func(rec *neo4j.Record) (*bool, error) {
		val, _, err := neo4j.GetRecordValue[bool](rec, "has_relation")
		if err != nil {
			return nil, err
		}
		return &val, nil
	})
	if err != nil {
		return false, errors.Join(repository.ErrRelationRead, err)
	}

	return *hasRelation, nil
}

// HasSystemRole returns true if there is a relation between the source and
// target that is a system role. If there is no relation, false is returned.
func (r *PermissionRepository) HasSystemRole(ctx context.Context, source model.ID, roles ...model.SystemRole) (bool, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.RelationRepository/HasAnyRelation")
	defer span.End()

	if err := source.Validate(); err != nil {
		return false, errors.Join(repository.ErrRelationRead, err)
	}

	if len(roles) == 0 {
		return false, errors.Join(repository.ErrRelationRead, model.ErrInvalidID)
	}

	roleIDs := make([]string, len(roles))
	for i, role := range roles {
		roleIDs[i] = role.String()
	}

	cypher := `
	MATCH path = (s:` + source.Label() + ` {id: $source_id})-[:` + EdgeKindMemberOf.String() + `]->(r:` + model.ResourceTypeRole.String() + ` {system: true})
	WHERE r.id IN $target_ids
	RETURN count(path) > 0 AS has_system_role
	LIMIT 1`

	params := map[string]any{
		"source_id":  source.String(),
		"target_ids": roleIDs,
	}

	hasSystemRole, err := ExecuteReadAndReadSingle(ctx, r.db, cypher, params, func(rec *neo4j.Record) (*bool, error) {
		val, _, err := neo4j.GetRecordValue[bool](rec, "has_system_role")
		if err != nil {
			return nil, err
		}
		return &val, nil
	})
	if err != nil {
		return false, errors.Join(repository.ErrSystemRoleRead, err)
	}

	return *hasSystemRole, nil
}

// Update updates an existing permission's kind. If the permission does not
// exist, an error is returned. If the permission's kind is already the same
// as the one provided, the kind is overwritten and the updated_at timestamp
// is updated.
func (r *PermissionRepository) Update(ctx context.Context, id model.ID, kind model.PermissionKind) (*model.Permission, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.PermissionRepository/Update")
	defer span.End()

	cypher := `
	MATCH (s)-[p:` + EdgeKindHasPermission.String() + ` {id: $id}]->(t)
	SET p.kind = $kind, p.updated_at = datetime()
	RETURN s, p, t
	`

	params := map[string]any{
		"id":         id.String(),
		"kind":       kind.String(),
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
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.PermissionRepository/Delete")
	defer span.End()

	cypher := `MATCH (s)-[p:` + EdgeKindHasPermission.String() + ` {id: $id}]->(t) DELETE p`

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

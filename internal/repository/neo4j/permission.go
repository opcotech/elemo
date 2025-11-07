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

		// Exclude "id" and "kind" from ScanIntoStruct to handle them manually
		// "kind" must be manually unmarshaled because JSON unmarshaling doesn't properly call UnmarshalText for uint8 types
		if err := ScanIntoStruct(&val, &parsed, []string{"id", "kind"}); err != nil {
			return nil, err
		}

		parsed.ID, _ = model.NewIDFromString(val.GetProperties()["id"].(string), model.ResourceTypePermission.String())
		parsed.Subject, _ = model.NewIDFromString(subject.GetProperties()["id"].(string), subject.Labels[0])
		parsed.Target, _ = model.NewIDFromString(target.GetProperties()["id"].(string), target.Labels[0])

		// Manually extract and unmarshal kind to ensure proper conversion from string to PermissionKind
		// This is necessary because JSON unmarshaling might not properly call UnmarshalText for uint8 types
		kindStr := val.GetProperties()["kind"].(string)
		if err := parsed.Kind.UnmarshalText([]byte(kindStr)); err != nil {
			return nil, err
		}

		if err := parsed.Validate(); err != nil {
			return nil, err
		}

		return parsed, nil
	}
}

// scanSystemLevelPermission is a helper function for scanning a permission from a ResourceType node.
// The target is preserved as a nil ID (system-level permission) rather than parsing from the node.
func (r *PermissionRepository) scanSystemLevelPermission(target model.ID) func(rec *neo4j.Record) (*model.Permission, error) {
	return func(rec *neo4j.Record) (*model.Permission, error) {
		val, _, err := neo4j.GetRecordValue[neo4j.Relationship](rec, "p")
		if err != nil {
			return nil, err
		}

		subject, _, err := neo4j.GetRecordValue[neo4j.Node](rec, "s")
		if err != nil {
			return nil, err
		}

		perm := &model.Permission{}
		if err := ScanIntoStruct(&val, &perm, []string{"id"}); err != nil {
			return nil, err
		}

		perm.ID, _ = model.NewIDFromString(val.GetProperties()["id"].(string), model.ResourceTypePermission.String())
		perm.Subject, _ = model.NewIDFromString(subject.GetProperties()["id"].(string), subject.Labels[0])
		perm.Target = target // Preserve nil ID for system-level permissions

		kindStr := val.GetProperties()["kind"].(string)
		if err := perm.Kind.UnmarshalText([]byte(kindStr)); err != nil {
			return nil, err
		}

		if err := perm.Validate(); err != nil {
			return nil, err
		}

		return perm, nil
	}
}

// scanSystemRolePermission is a helper function for scanning permissions from system roles.
// The target is preserved as a nil ID (system-level permission) rather than parsing from the node.
func (r *PermissionRepository) scanSystemRolePermission(target model.ID) func(rec *neo4j.Record) (*model.Permission, error) {
	return func(rec *neo4j.Record) (*model.Permission, error) {
		val, _, err := neo4j.GetRecordValue[neo4j.Relationship](rec, "p")
		if err != nil {
			return nil, err
		}

		subject, _, err := neo4j.GetRecordValue[neo4j.Node](rec, "s")
		if err != nil {
			return nil, err
		}

		perm := &model.Permission{}
		if err := ScanIntoStruct(&val, &perm, []string{"id"}); err != nil {
			return nil, err
		}

		// Generate a new ID for virtual permission (system role permission)
		perm.ID = model.MustNewID(model.ResourceTypePermission)
		perm.Subject, _ = model.NewIDFromString(subject.GetProperties()["id"].(string), subject.Labels[0])
		perm.Target = target // Preserve nil ID for system-level permissions

		kindStr := val.GetProperties()["kind"].(string)
		if err := perm.Kind.UnmarshalText([]byte(kindStr)); err != nil {
			return nil, err
		}

		// Set default CreatedAt if not set by ScanIntoStruct
		if perm.CreatedAt == nil {
			now := time.Now().UTC()
			perm.CreatedAt = &now
		}

		if err := perm.Validate(); err != nil {
			return nil, err
		}

		return perm, nil
	}
}

// getDirectResourceTypePermissions returns direct permissions on a ResourceType node.
func (r *PermissionRepository) getDirectResourceTypePermissions(ctx context.Context, source, target model.ID) ([]*model.Permission, error) {
	cypher := `
	MATCH (s:` + source.Label() + ` {id: $source})-[p:` + EdgeKindHasPermission.String() + `]->(rt:` + model.ResourceTypeResourceType.String() + ` {id: $target_label})
	RETURN s, p, rt
	ORDER BY p.created_at DESC`

	params := map[string]any{
		"source":       source.String(),
		"target_label": target.Label(),
	}

	perms, err := ExecuteReadAndReadAll(ctx, r.db, cypher, params, r.scanSystemLevelPermission(target))
	if err != nil {
		return nil, errors.Join(err, repository.ErrPermissionRead)
	}

	return perms, nil
}

// getSystemRolePermissions returns permissions derived from system roles (Owner, Admin, Support).
func (r *PermissionRepository) getSystemRolePermissions(ctx context.Context, source, target model.ID) ([]*model.Permission, error) {
	cypher := `
	MATCH (s:` + source.Label() + ` {id: $source})-[m:` + EdgeKindMemberOf.String() + `]->(r:` + model.ResourceTypeRole.String() + ` {system: true})
	MATCH (r)-[p:` + EdgeKindHasPermission.String() + `]->(rt:` + model.ResourceTypeResourceType.String() + ` {id: $target_label})
	RETURN DISTINCT s, p, rt
	ORDER BY p.created_at DESC`

	params := map[string]any{
		"source":       source.String(),
		"target_label": target.Label(),
	}

	perms, err := ExecuteReadAndReadAll(ctx, r.db, cypher, params, r.scanSystemRolePermission(target))
	if err != nil {
		return nil, errors.Join(err, repository.ErrPermissionRead)
	}

	return perms, nil
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
	MATCH (subject:` + perm.Subject.Label() + ` {id: $subject})
	MATCH (target:` + perm.Target.Label() + ` {id: $target})
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
// source has. If no permissions exist, an empty slice is returned. For system
// level permissions (nil IDs), it checks permissions on ResourceType nodes
// and system roles.
func (r *PermissionRepository) GetBySubjectAndTarget(ctx context.Context, source, target model.ID) ([]*model.Permission, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.PermissionRepository/GetBySubjectAndTarget")
	defer span.End()

	if target.IsNil() {
		directPerms, err := r.getDirectResourceTypePermissions(ctx, source, target)
		if err != nil {
			return nil, err
		}

		systemRolePerms, err := r.getSystemRolePermissions(ctx, source, target)
		if err != nil {
			return nil, err
		}

		return deduplicatePermissions(directPerms, systemRolePerms), nil
	}

	cypher := `
	MATCH (s:` + source.Label() + ` {id: $source})
	MATCH (t:` + target.Label() + ` {id: $target})
	MATCH path=(s)-[:` + EdgeKindHasPermission.String() + `|` + EdgeKindMemberOf.String() + `*..2]->(t)
	UNWIND relationships(path) AS rel
	WITH s, t, rel, startNode(rel) AS relStart, endNode(rel) AS relEnd
	WHERE type(rel) = "` + EdgeKindHasPermission.String() + `" AND relEnd = t
	RETURN s, rel AS p, t
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

	// If source and target are the same, they always have a relation (self-relation)
	if source.String() == target.String() {
		return true, nil
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
		"id":   id.String(),
		"kind": kind.String(),
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

// deduplicatePermissions merges permissions from different sources, giving precedence to
// PermissionKindAll ("*") when present. If "*" permission exists, all other permissions are
// removed as "*" grants all permissions.
func deduplicatePermissions(directPerms, systemRolePerms []*model.Permission) []*model.Permission {
	permissionMap := make(map[model.PermissionKind]*model.Permission)

	// Add direct permissions first
	for _, perm := range directPerms {
		permissionMap[perm.Kind] = perm
	}

	// Add system role permissions with deduplication logic
	for _, perm := range systemRolePerms {
		// If we already have "*" permission and this is not "*", skip it
		if _, hasAll := permissionMap[model.PermissionKindAll]; hasAll && perm.Kind != model.PermissionKindAll {
			continue
		}

		// If we're adding "*" permission, it overrides all others
		if perm.Kind == model.PermissionKindAll {
			permissionMap = map[model.PermissionKind]*model.Permission{
				model.PermissionKindAll: perm,
			}
			break
		}

		// Add permission if not already present
		if _, exists := permissionMap[perm.Kind]; !exists {
			permissionMap[perm.Kind] = perm
		}
	}

	// Convert map to slice
	result := make([]*model.Permission, 0, len(permissionMap))
	for _, perm := range permissionMap {
		result = append(result, perm)
	}

	return result
}

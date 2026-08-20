package repository

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/neo4j/neo4j-go-driver/v6/neo4j"

	"github.com/opcotech/elemo/internal/model"
)

var (
	ErrPermissionCreate = errors.New("failed to create permission")        // grant cannot be created
	ErrPermissionDelete = errors.New("failed to delete permission")        // grant cannot be deleted
	ErrPermissionRead   = errors.New("failed to read permission")          // grant cannot be read
	ErrPermissionUpdate = errors.New("failed to update permission")        // grant cannot be updated
	ErrInScopeOfLink    = errors.New("failed to link authorization scope") // IN_SCOPE_OF edge cannot be created
)

// Grant is a scoped authorization relationship from a principal to a resource.
type Grant struct {
	ID        model.ID       `json:"id"`
	Principal model.ID       `json:"principal"`
	Scope     model.ID       `json:"scope"`
	RoleID    *model.ID      `json:"role_id,omitempty"`
	Actions   []model.Action `json:"actions"`
	CreatedAt *time.Time     `json:"created_at"`
	UpdatedAt *time.Time     `json:"updated_at"`
}

// CreateGrantOpts holds the data required to create a grant.
type CreateGrantOpts struct {
	Principal model.ID
	Scope     model.ID
	RoleID    *model.ID
	Actions   []model.Action
}

// Validate reports whether the grant has a principal, a scope, and either a
// role or at least one action.
func (o CreateGrantOpts) Validate() error {
	if err := o.Principal.Validate(); err != nil {
		return errors.Join(model.ErrInvalidGrant, err)
	}
	if err := o.Scope.Validate(); err != nil {
		return errors.Join(model.ErrInvalidGrant, err)
	}
	if !model.IsPrincipalType(o.Principal.Type) {
		return errors.Join(model.ErrInvalidGrant, model.ErrNotAPrincipal)
	}
	if o.RoleID == nil && len(o.Actions) == 0 {
		return errors.Join(model.ErrInvalidGrant, model.ErrInvalidAction)
	}
	if o.RoleID != nil {
		if err := o.RoleID.Validate(); err != nil {
			return errors.Join(model.ErrInvalidGrant, err)
		}
		if o.RoleID.Type != model.ResourceTypeRole {
			return errors.Join(model.ErrInvalidGrant, model.ErrInvalidID)
		}
	}
	for _, action := range o.Actions {
		if err := action.Validate(); err != nil {
			return errors.Join(model.ErrInvalidGrant, err)
		}
	}
	return nil
}

// Decision explains why an authorization check was allowed or denied.
type Decision struct {
	Allowed   bool
	Action    model.Action
	Actor     model.ID
	Resource  model.ID
	Principal *model.ID
	Scope     *model.ID
	GrantID   *model.ID
	RoleID    *model.ID
}

//go:generate go tool mockgen -source=permission.go -destination=permission_mock_gen.go -package=repository -mock_names "PermissionRepository=MockPermissionRepository"
type PermissionRepository interface {
	Create(ctx context.Context, opts CreateGrantOpts) (*Grant, error)
	Get(ctx context.Context, id model.ID) (*Grant, error)
	ListByPrincipal(ctx context.Context, principal model.ID) ([]*Grant, error)
	ListByScope(ctx context.Context, scope model.ID) ([]*Grant, error)
	Delete(ctx context.Context, id model.ID) error
	Has(ctx context.Context, actor, resource model.ID, action model.Action) (bool, error)
	EffectiveActions(ctx context.Context, actor, resource model.ID) ([]model.Action, error)
	Explain(ctx context.Context, actor, resource model.ID, action model.Action) (*Decision, error)
	ListVisible(ctx context.Context, actor model.ID, action model.Action, parent model.ID, resourceType model.ResourceType) ([]model.ID, error)
	ListGrantScopes(ctx context.Context, actor model.ID, action model.Action) ([]model.ID, error)
	ListScopeAncestry(ctx context.Context, resource model.ID) ([]model.ID, error)
	LinkInScopeOf(ctx context.Context, child, parent model.ID) error
	BumpGeneration(ctx context.Context, principal model.ID) error
}

// Neo4jPermissionRepository is a repository for managing grants and evaluating authorization.
type Neo4jPermissionRepository struct {
	*neo4jBaseRepository
}

func (r *Neo4jPermissionRepository) scanGrant() func(rec *neo4j.Record) (*Grant, error) {
	return func(rec *neo4j.Record) (*Grant, error) {
		rel, _, err := neo4j.GetRecordValue[neo4j.Relationship](rec, "g")
		if err != nil {
			return nil, err
		}
		principal, _, err := neo4j.GetRecordValue[neo4j.Node](rec, "principal")
		if err != nil {
			return nil, err
		}
		scope, _, err := neo4j.GetRecordValue[neo4j.Node](rec, "scope")
		if err != nil {
			return nil, err
		}

		grant := &Grant{
			Actions: []model.Action{},
		}
		grant.ID, err = model.NewIDFromString(rel.GetProperties()["id"].(string), model.ResourceTypePermission.String())
		if err != nil {
			return nil, err
		}
		grant.Principal, err = model.NewIDFromString(principal.GetProperties()["id"].(string), domainLabel(principal.Labels))
		if err != nil {
			return nil, err
		}
		grant.Scope, err = model.NewIDFromString(scope.GetProperties()["id"].(string), domainLabel(scope.Labels))
		if err != nil {
			return nil, err
		}
		if roleID, ok := rel.GetProperties()["role_id"].(string); ok && roleID != "" {
			id, err := model.NewIDFromString(roleID, model.ResourceTypeRole.String())
			if err != nil {
				return nil, err
			}
			grant.RoleID = &id
		}
		if raw, ok := rel.GetProperties()["actions"].([]any); ok {
			values := make([]string, 0, len(raw))
			for _, item := range raw {
				if s, ok := item.(string); ok {
					values = append(values, s)
				}
			}
			actions, err := model.ParseActions(values)
			if err != nil {
				return nil, err
			}
			grant.Actions = actions
		}
		if createdAt, err := Neo4jDecodeTime(rel.GetProperties()["created_at"]); err == nil {
			grant.CreatedAt = createdAt
		}
		if raw, ok := rel.GetProperties()["updated_at"]; ok {
			if updatedAt, err := Neo4jDecodeTime(raw); err == nil {
				grant.UpdatedAt = updatedAt
			}
		}
		return grant, nil
	}
}

// Create records a GRANTED relationship from principal to scope.
func (r *Neo4jPermissionRepository) Create(ctx context.Context, opts CreateGrantOpts) (*Grant, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.PermissionRepository/Create")
	defer span.End()

	if err := opts.Validate(); err != nil {
		return nil, errors.Join(ErrPermissionCreate, err)
	}

	createdAt := time.Now().UTC()
	id := model.MustNewID(model.ResourceTypePermission)
	roleID := ""
	if opts.RoleID != nil {
		roleID = opts.RoleID.String()
	}

	cypher := `
	MATCH (principal:` + opts.Principal.Label() + ` {id: $principal_id})
	MATCH (scope:` + opts.Scope.Label() + ` {id: $scope_id})
	SET principal:` + model.LabelPrincipal + `
	CREATE (principal)-[g:` + EdgeKindGranted.String() + ` {
		id: $id,
		role_id: $role_id,
		actions: $actions,
		created_at: datetime($created_at)
	}]->(scope)
	RETURN principal, g, scope`

	params := map[string]any{
		"principal_id": opts.Principal.String(),
		"scope_id":     opts.Scope.String(),
		"id":           id.String(),
		"role_id":      roleID,
		"actions":      model.ActionStrings(opts.Actions),
		"created_at":   createdAt.Format(time.RFC3339Nano),
	}

	grant, err := Neo4jExecuteWriteAndReadSingle(ctx, r.db, cypher, params, r.scanGrant())
	if err != nil {
		return nil, errors.Join(ErrPermissionCreate, err)
	}
	return grant, nil
}

// Get returns a grant by ID.
func (r *Neo4jPermissionRepository) Get(ctx context.Context, id model.ID) (*Grant, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.PermissionRepository/Get")
	defer span.End()

	if err := id.Validate(); err != nil {
		return nil, errors.Join(ErrPermissionRead, err)
	}

	cypher := `
	MATCH (principal)-[g:` + EdgeKindGranted.String() + ` {id: $id}]->(scope)
	RETURN principal, g, scope`

	grant, err := Neo4jExecuteReadAndReadSingle(ctx, r.db, cypher, map[string]any{"id": id.String()}, r.scanGrant())
	if err != nil {
		return nil, errors.Join(ErrPermissionRead, err)
	}
	return grant, nil
}

// ListByPrincipal returns grants issued to principal.
func (r *Neo4jPermissionRepository) ListByPrincipal(ctx context.Context, principal model.ID) ([]*Grant, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.PermissionRepository/ListByPrincipal")
	defer span.End()

	if err := principal.Validate(); err != nil {
		return nil, errors.Join(ErrPermissionRead, err)
	}

	cypher := `
	MATCH (principal:` + principal.Label() + ` {id: $principal_id})-[g:` + EdgeKindGranted.String() + `]->(scope)
	RETURN principal, g, scope`

	grants, err := Neo4jExecuteReadAndReadAll(ctx, r.db, cypher, map[string]any{
		"principal_id": principal.String(),
	}, r.scanGrant())
	if err != nil {
		return nil, errors.Join(ErrPermissionRead, err)
	}
	if grants == nil {
		grants = []*Grant{}
	}
	return grants, nil
}

// ListByScope returns grants whose scope is the given resource.
func (r *Neo4jPermissionRepository) ListByScope(ctx context.Context, scope model.ID) ([]*Grant, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.PermissionRepository/ListByScope")
	defer span.End()

	if err := scope.Validate(); err != nil {
		return nil, errors.Join(ErrPermissionRead, err)
	}

	cypher := `
	MATCH (principal)-[g:` + EdgeKindGranted.String() + `]->(scope:` + scope.Label() + ` {id: $scope_id})
	RETURN principal, g, scope`

	grants, err := Neo4jExecuteReadAndReadAll(ctx, r.db, cypher, map[string]any{
		"scope_id": scope.String(),
	}, r.scanGrant())
	if err != nil {
		return nil, errors.Join(ErrPermissionRead, err)
	}
	if grants == nil {
		grants = []*Grant{}
	}
	return grants, nil
}

// Delete removes a grant by ID.
func (r *Neo4jPermissionRepository) Delete(ctx context.Context, id model.ID) error {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.PermissionRepository/Delete")
	defer span.End()

	if err := id.Validate(); err != nil {
		return errors.Join(ErrPermissionDelete, err)
	}

	cypher := `MATCH ()-[g:` + EdgeKindGranted.String() + ` {id: $id}]->() DELETE g`
	if err := Neo4jExecuteWriteAndConsume(ctx, r.db, cypher, map[string]any{"id": id.String()}); err != nil {
		return errors.Join(ErrPermissionDelete, err)
	}
	return nil
}

// Has reports whether actor may perform action on resource via a direct or
// inherited grant. MEMBER_OF is followed at most one hop; inactive users never
// match.
func (r *Neo4jPermissionRepository) Has(ctx context.Context, actor, resource model.ID, action model.Action) (bool, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.PermissionRepository/Has")
	defer span.End()

	if err := actor.Validate(); err != nil {
		return false, errors.Join(ErrPermissionRead, err)
	}
	if err := resource.Validate(); err != nil {
		return false, errors.Join(ErrPermissionRead, err)
	}
	if err := action.Validate(); err != nil {
		return false, errors.Join(ErrPermissionRead, err)
	}

	cypher := `
	MATCH (actor:` + actor.Label() + ` {id: $actor_id})
	WHERE actor.status IS NULL OR actor.status = $active_status
	MATCH (resource:` + resource.Label() + ` {id: $resource_id})
	MATCH (actor)-[:` + EdgeKindMemberOf.String() + `*0..1]->(principal)
	WHERE principal:User OR principal:Team OR principal:Organization
	AND (principal.status IS NULL OR principal.status = $active_status)
	MATCH path = (resource)-[:` + EdgeKindInScopeOf.String() + `*0..]->(scope)
	WHERE ` + authzAcyclicPathPredicate("path") + `
	MATCH (principal)-[g:` + EdgeKindGranted.String() + `]->(scope)
	WHERE ($action IN coalesce(g.actions, [])) OR (
		g.role_id IS NOT NULL AND g.role_id <> "" AND EXISTS {
			MATCH (role:` + model.ResourceTypeRole.String() + ` {id: g.role_id})
			WHERE $action IN coalesce(role.actions, [])
		}
	)
	RETURN true AS allowed
	LIMIT 1`

	params := map[string]any{
		"actor_id":      actor.String(),
		"resource_id":   resource.String(),
		"action":        action.String(),
		"active_status": model.UserStatusActive.String(),
	}

	allowed, err := Neo4jExecuteReadAndReadSingle(ctx, r.db, cypher, params, func(rec *neo4j.Record) (*bool, error) {
		val, _, err := neo4j.GetRecordValue[bool](rec, "allowed")
		if err != nil {
			return nil, err
		}
		return &val, nil
	})
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return false, nil
		}
		return false, errors.Join(ErrPermissionRead, err)
	}
	return *allowed, nil
}

// EffectiveActions returns the distinct union of grant actions and referenced
// role actions the actor holds on resource, including inherited scopes.
func (r *Neo4jPermissionRepository) EffectiveActions(ctx context.Context, actor, resource model.ID) ([]model.Action, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.PermissionRepository/EffectiveActions")
	defer span.End()

	if err := actor.Validate(); err != nil {
		return nil, errors.Join(ErrPermissionRead, err)
	}
	if err := resource.Validate(); err != nil {
		return nil, errors.Join(ErrPermissionRead, err)
	}

	cypher := `
	MATCH (actor:` + actor.Label() + ` {id: $actor_id})
	WHERE actor.status IS NULL OR actor.status = $active_status
	MATCH (resource:` + resource.Label() + ` {id: $resource_id})
	MATCH (actor)-[:` + EdgeKindMemberOf.String() + `*0..1]->(principal)
	WHERE principal:User OR principal:Team OR principal:Organization
	AND (principal.status IS NULL OR principal.status = $active_status)
	MATCH path = (resource)-[:` + EdgeKindInScopeOf.String() + `*0..]->(scope)
	WHERE ` + authzAcyclicPathPredicate("path") + `
	MATCH (principal)-[g:` + EdgeKindGranted.String() + `]->(scope)
	OPTIONAL MATCH (role:` + model.ResourceTypeRole.String() + ` {id: g.role_id})
	WITH coalesce(g.actions, []) + coalesce(role.actions, []) AS actions
	UNWIND actions AS action
	RETURN DISTINCT action`

	params := map[string]any{
		"actor_id":      actor.String(),
		"resource_id":   resource.String(),
		"active_status": model.UserStatusActive.String(),
	}

	values, err := Neo4jExecuteReadAndReadAll(ctx, r.db, cypher, params, func(rec *neo4j.Record) (*string, error) {
		val, _, err := neo4j.GetRecordValue[string](rec, "action")
		if err != nil {
			return nil, err
		}
		return &val, nil
	})
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return []model.Action{}, nil
		}
		return nil, errors.Join(ErrPermissionRead, err)
	}

	raw := make([]string, 0, len(values))
	for _, value := range values {
		if value != nil && *value != "" {
			raw = append(raw, *value)
		}
	}
	actions, err := model.ParseActions(raw)
	if err != nil {
		return nil, errors.Join(ErrPermissionRead, err)
	}
	return actions, nil
}

// Explain returns why Has allowed or denied the check. On deny, Principal,
// Scope, GrantID, and RoleID are left nil.
func (r *Neo4jPermissionRepository) Explain(ctx context.Context, actor, resource model.ID, action model.Action) (*Decision, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.PermissionRepository/Explain")
	defer span.End()

	decision := &Decision{
		Action:   action,
		Actor:    actor,
		Resource: resource,
	}

	allowed, err := r.Has(ctx, actor, resource, action)
	if err != nil {
		return nil, err
	}
	decision.Allowed = allowed
	if !allowed {
		return decision, nil
	}

	cypher := `
	MATCH (actor:` + actor.Label() + ` {id: $actor_id})
	WHERE actor.status IS NULL OR actor.status = $active_status
	MATCH (resource:` + resource.Label() + ` {id: $resource_id})
	MATCH (actor)-[:` + EdgeKindMemberOf.String() + `*0..1]->(principal)
	WHERE principal:User OR principal:Team OR principal:Organization
	AND (principal.status IS NULL OR principal.status = $active_status)
	MATCH path = (resource)-[:` + EdgeKindInScopeOf.String() + `*0..]->(scope)
	WHERE ` + authzAcyclicPathPredicate("path") + `
	MATCH (principal)-[g:` + EdgeKindGranted.String() + `]->(scope)
	WHERE ($action IN coalesce(g.actions, [])) OR (
		g.role_id IS NOT NULL AND g.role_id <> "" AND EXISTS {
			MATCH (role:` + model.ResourceTypeRole.String() + ` {id: g.role_id})
			WHERE $action IN coalesce(role.actions, [])
		}
	)
	RETURN principal, g, scope
	LIMIT 1`

	params := map[string]any{
		"actor_id":      actor.String(),
		"resource_id":   resource.String(),
		"action":        action.String(),
		"active_status": model.UserStatusActive.String(),
	}

	grant, err := Neo4jExecuteReadAndReadSingle(ctx, r.db, cypher, params, r.scanGrant())
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return decision, nil
		}
		return nil, errors.Join(ErrPermissionRead, err)
	}
	decision.Principal = &grant.Principal
	decision.Scope = &grant.Scope
	decision.GrantID = &grant.ID
	decision.RoleID = grant.RoleID
	return decision, nil
}

// ListVisible returns IDs of resourceType that actor may perform action on.
// When parent is the installation, every matching node is considered;
// otherwise only direct IN_SCOPE_OF children of parent are listed.
func (r *Neo4jPermissionRepository) ListVisible(ctx context.Context, actor model.ID, action model.Action, parent model.ID, resourceType model.ResourceType) ([]model.ID, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.PermissionRepository/ListVisible")
	defer span.End()

	if err := actor.Validate(); err != nil {
		return nil, errors.Join(ErrPermissionRead, err)
	}
	if err := action.Validate(); err != nil {
		return nil, errors.Join(ErrPermissionRead, err)
	}
	if err := parent.Validate(); err != nil {
		return nil, errors.Join(ErrPermissionRead, err)
	}
	if !resourceType.IsAResourceType() {
		return nil, errors.Join(ErrPermissionRead, model.ErrInvalidResourceType)
	}

	params := map[string]any{
		"user_id":       actor.String(),
		"action":        action.String(),
		"active_status": model.UserStatusActive.String(),
	}

	var cypher string
	if parent.Type == model.ResourceTypeInstallation {
		cypher = `
		MATCH (n:` + resourceType.String() + `)
		WHERE ` + AuthzVisibleExistsClause("n", "$user_id", "$action") + `
		RETURN n.id AS id`
	} else {
		cypher = `
		MATCH (child:` + resourceType.String() + `)-[:` + EdgeKindInScopeOf.String() + `]->(parent:` + parent.Label() + ` {id: $parent_id})
		WHERE ` + AuthzVisibleExistsClause("child", "$user_id", "$action") + `
		RETURN child.id AS id`
		params["parent_id"] = parent.String()
	}

	ids, err := Neo4jExecuteReadAndReadAll(ctx, r.db, cypher, params, func(rec *neo4j.Record) (model.ID, error) {
		raw, err := Neo4jParseValueFromRecord[string](rec, "id")
		if err != nil {
			return model.ID{}, err
		}
		return model.NewIDFromString(raw, resourceType.String())
	})
	if err != nil {
		return nil, errors.Join(ErrPermissionRead, err)
	}
	if ids == nil {
		ids = []model.ID{}
	}
	return ids, nil
}

// ListGrantScopes returns distinct scopes the actor holds action on via a
// principal (the actor, a team, or an organization) GRANTED edge. It does not
// expand descendants; callers intersect these IDs with indexed scope ancestry.
func (r *Neo4jPermissionRepository) ListGrantScopes(ctx context.Context, actor model.ID, action model.Action) ([]model.ID, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.PermissionRepository/ListGrantScopes")
	defer span.End()

	if err := actor.Validate(); err != nil {
		return nil, errors.Join(ErrPermissionRead, err)
	}
	if err := action.Validate(); err != nil {
		return nil, errors.Join(ErrPermissionRead, err)
	}

	cypher := `
	MATCH (actor:` + actor.Label() + ` {id: $actor_id})
	WHERE actor.status IS NULL OR actor.status = $active_status
	MATCH (actor)-[:` + EdgeKindMemberOf.String() + `*0..1]->(principal)
	WHERE (principal:User OR principal:Team OR principal:Organization)
	AND (principal.status IS NULL OR principal.status = $active_status)
	MATCH (principal)-[g:` + EdgeKindGranted.String() + `]->(scope)
	WHERE ($action IN coalesce(g.actions, [])) OR (
		g.role_id IS NOT NULL AND g.role_id <> "" AND EXISTS {
			MATCH (role:` + model.ResourceTypeRole.String() + ` {id: g.role_id})
			WHERE $action IN coalesce(role.actions, [])
		}
	)
	RETURN DISTINCT scope`

	params := map[string]any{
		"actor_id":      actor.String(),
		"action":        action.String(),
		"active_status": model.UserStatusActive.String(),
	}

	ids, err := Neo4jExecuteReadAndReadAll(ctx, r.db, cypher, params, func(rec *neo4j.Record) (model.ID, error) {
		node, _, err := neo4j.GetRecordValue[neo4j.Node](rec, "scope")
		if err != nil {
			return model.ID{}, err
		}
		return Neo4jDecodeIDFromLabel(node)
	})
	if err != nil {
		return nil, errors.Join(ErrPermissionRead, err)
	}
	if ids == nil {
		ids = []model.ID{}
	}
	return ids, nil
}

// ListScopeAncestry returns resource and every IN_SCOPE_OF ancestor, nearest
// first. The walk is acyclic.
func (r *Neo4jPermissionRepository) ListScopeAncestry(ctx context.Context, resource model.ID) ([]model.ID, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.PermissionRepository/ListScopeAncestry")
	defer span.End()

	if err := resource.Validate(); err != nil {
		return nil, errors.Join(ErrPermissionRead, err)
	}

	cypher := `
	MATCH (n:` + resource.Label() + ` {id: $resource_id})
	MATCH path = (n)-[:` + EdgeKindInScopeOf.String() + `*0..]->(scope)
	WHERE ` + authzAcyclicPathPredicate("path") + `
	WITH DISTINCT scope, min(length(path)) AS depth
	RETURN scope
	ORDER BY depth`

	params := map[string]any{
		"resource_id": resource.String(),
	}

	ids, err := Neo4jExecuteReadAndReadAll(ctx, r.db, cypher, params, func(rec *neo4j.Record) (model.ID, error) {
		node, _, err := neo4j.GetRecordValue[neo4j.Node](rec, "scope")
		if err != nil {
			return model.ID{}, err
		}
		return Neo4jDecodeIDFromLabel(node)
	})
	if err != nil {
		return nil, errors.Join(ErrPermissionRead, err)
	}
	if ids == nil {
		ids = []model.ID{}
	}
	return ids, nil
}

// LinkInScopeOf creates (child)-[:IN_SCOPE_OF]->(parent). Self-links and links
// that would close a cycle return model.ErrGrantCycle.
func (r *Neo4jPermissionRepository) LinkInScopeOf(ctx context.Context, child, parent model.ID) error {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.PermissionRepository/LinkInScopeOf")
	defer span.End()

	if err := child.Validate(); err != nil {
		return errors.Join(ErrInScopeOfLink, err)
	}
	if err := parent.Validate(); err != nil {
		return errors.Join(ErrInScopeOfLink, err)
	}
	if child.String() == parent.String() && child.Type == parent.Type {
		return errors.Join(ErrInScopeOfLink, model.ErrGrantCycle)
	}

	cycleCypher := `
	MATCH (child:` + child.Label() + ` {id: $child_id})
	MATCH (parent:` + parent.Label() + ` {id: $parent_id})
	OPTIONAL MATCH path = (parent)-[:` + EdgeKindInScopeOf.String() + `*0..]->(child)
	WHERE path IS NOT NULL
	RETURN count(path) > 0 AS would_cycle`

	wouldCycle, err := Neo4jExecuteReadAndReadSingle(ctx, r.db, cycleCypher, map[string]any{
		"child_id":  child.String(),
		"parent_id": parent.String(),
	}, func(rec *neo4j.Record) (*bool, error) {
		val, _, recErr := neo4j.GetRecordValue[bool](rec, "would_cycle")
		if recErr != nil {
			return nil, recErr
		}
		return &val, nil
	})
	if err != nil {
		return errors.Join(ErrInScopeOfLink, err)
	}
	if wouldCycle != nil && *wouldCycle {
		return errors.Join(ErrInScopeOfLink, model.ErrGrantCycle)
	}

	cypher := `
	MATCH (child:` + child.Label() + ` {id: $child_id})
	MATCH (parent:` + parent.Label() + ` {id: $parent_id})
	MERGE (child)-[rel:` + EdgeKindInScopeOf.String() + `]->(parent)
	ON CREATE SET rel.id = $rel_id, rel.created_at = datetime($created_at)`

	params := map[string]any{
		"child_id":   child.String(),
		"parent_id":  parent.String(),
		"rel_id":     model.NewRawID(),
		"created_at": time.Now().UTC().Format(time.RFC3339Nano),
	}
	if err := Neo4jExecuteWriteAndConsume(ctx, r.db, cypher, params); err != nil {
		return errors.Join(ErrInScopeOfLink, err)
	}
	return nil
}

// BumpGeneration is a no-op on Neo4j; generation is stored in Redis.
func (r *Neo4jPermissionRepository) BumpGeneration(ctx context.Context, principal model.ID) error {
	_, span := r.tracer.Start(ctx, "repository.neo4j.PermissionRepository/BumpGeneration")
	defer span.End()
	_ = principal
	return nil
}

// NewNeo4jPermissionRepository creates a new permission neo4jBaseRepository.
func NewNeo4jPermissionRepository(opts ...Neo4jRepositoryOption) (*Neo4jPermissionRepository, error) {
	baseRepo, err := newNeo4jRepository(opts...)
	if err != nil {
		return nil, err
	}

	return &Neo4jPermissionRepository{
		neo4jBaseRepository: baseRepo,
	}, nil
}

// authzAcyclicPathPredicate rejects cyclic IN_SCOPE_OF walks. Names are
// prefixed so the fragment can be embedded in EXISTS subqueries that inherit
// outer aliases such as n (namespace).
func authzAcyclicPathPredicate(pathVar string) string {
	return "ALL(authz_node IN nodes(" + pathVar + ") WHERE size([authz_other IN nodes(" + pathVar + ") WHERE authz_other = authz_node]) = 1)"
}

// AuthzVisibleExistsClause returns a Cypher EXISTS fragment that is true when
// the node bound to resourceAlias is visible to the actor for actionParam.
// Aliases inside the fragment are prefixed so it can be embedded in queries
// that already bind n or other names.
func AuthzVisibleExistsClause(resourceAlias, actorParam, actionParam string) string {
	return `
	EXISTS {
		MATCH (actor:User {id: ` + actorParam + `})
		WHERE actor.status IS NULL OR actor.status = $active_status
		MATCH (actor)-[:` + EdgeKindMemberOf.String() + `*0..1]->(principal)
		WHERE principal:User OR principal:Team OR principal:Organization
		AND (principal.status IS NULL OR principal.status = $active_status)
		MATCH path = (` + resourceAlias + `)-[:` + EdgeKindInScopeOf.String() + `*0..]->(scope)
		WHERE ` + authzAcyclicPathPredicate("path") + `
		MATCH (principal)-[g:` + EdgeKindGranted.String() + `]->(scope)
		WHERE (` + actionParam + ` IN coalesce(g.actions, [])) OR (
			g.role_id IS NOT NULL AND g.role_id <> "" AND EXISTS {
				MATCH (role:` + model.ResourceTypeRole.String() + ` {id: g.role_id})
				WHERE ` + actionParam + ` IN coalesce(role.actions, [])
			}
		)
	}`
}

func applyAuthzVisible(actor model.ID, action model.Action, alias, actorParam string, params map[string]any) string {
	if err := actor.Validate(); err != nil {
		return ""
	}
	if actorParam == "" {
		actorParam = "$user_id"
	}
	params[strings.TrimPrefix(actorParam, "$")] = actor.String()
	params["action"] = action.String()
	params["active_status"] = model.UserStatusActive.String()
	return AuthzVisibleExistsClause(alias, actorParam, "$action")
}

// applyAuthzReachableNamespace is true when the actor can namespace.read the
// bound namespace, or can read at least one descendant project, issue,
// document, or folder.
func applyAuthzReachableNamespace(actor model.ID, alias, actorParam string, params map[string]any) string {
	if err := actor.Validate(); err != nil {
		return ""
	}
	if actorParam == "" {
		actorParam = "$user_id"
	}
	params[strings.TrimPrefix(actorParam, "$")] = actor.String()
	params["active_status"] = model.UserStatusActive.String()
	params["namespace_read"] = model.ActionNamespaceRead.String()
	params["project_read"] = model.ActionProjectRead.String()
	params["issue_read"] = model.ActionIssueRead.String()
	params["document_read"] = model.ActionDocumentRead.String()

	inScope := `[:` + EdgeKindInScopeOf.String() + `*1..]`
	projectDesc := `
	EXISTS {
		MATCH desc_path = (authz_project:` + model.ResourceTypeProject.String() + `)-` + inScope + `->(` + alias + `)
		WHERE ` + authzAcyclicPathPredicate("desc_path") + `
		AND ` + AuthzVisibleExistsClause("authz_project", actorParam, "$project_read") + `
	}`
	issueDesc := `
	EXISTS {
		MATCH desc_path = (authz_issue:` + model.ResourceTypeIssue.String() + `)-` + inScope + `->(` + alias + `)
		WHERE ` + authzAcyclicPathPredicate("desc_path") + `
		AND ` + AuthzVisibleExistsClause("authz_issue", actorParam, "$issue_read") + `
	}`
	documentDesc := `
	EXISTS {
		MATCH desc_path = (authz_doc)-` + inScope + `->(` + alias + `)
		WHERE (authz_doc:` + model.ResourceTypeDocument.String() + ` OR authz_doc:` + model.ResourceTypeFolder.String() + `)
		AND ` + authzAcyclicPathPredicate("desc_path") + `
		AND ` + AuthzVisibleExistsClause("authz_doc", actorParam, "$document_read") + `
	}`

	return "(" + AuthzVisibleExistsClause(alias, actorParam, "$namespace_read") +
		" OR " + projectDesc +
		" OR " + issueDesc +
		" OR " + documentDesc + ")"
}

func authzGenKey(principal model.ID) string {
	return composeCacheKey("authz", "gen", principal.String())
}

// RedisCachedPermissionRepository wraps PermissionRepository and invalidates
// authz-filtered list caches on grant and ancestry writes. Evaluator methods
// are not cached.
type RedisCachedPermissionRepository struct {
	cacheRepo      *redisBaseRepository
	permissionRepo PermissionRepository
}

func (c *RedisCachedPermissionRepository) Create(ctx context.Context, opts CreateGrantOpts) (*Grant, error) {
	grant, err := c.permissionRepo.Create(ctx, opts)
	if err != nil {
		return nil, err
	}
	_ = c.BumpGeneration(ctx, opts.Principal)
	_ = clearPermissionAllCrossCache(ctx, c.cacheRepo)
	return grant, nil
}

func (c *RedisCachedPermissionRepository) Get(ctx context.Context, id model.ID) (*Grant, error) {
	return c.permissionRepo.Get(ctx, id)
}

func (c *RedisCachedPermissionRepository) ListByPrincipal(ctx context.Context, principal model.ID) ([]*Grant, error) {
	return c.permissionRepo.ListByPrincipal(ctx, principal)
}

func (c *RedisCachedPermissionRepository) ListByScope(ctx context.Context, scope model.ID) ([]*Grant, error) {
	return c.permissionRepo.ListByScope(ctx, scope)
}

func (c *RedisCachedPermissionRepository) Delete(ctx context.Context, id model.ID) error {
	grant, getErr := c.permissionRepo.Get(ctx, id)
	if err := c.permissionRepo.Delete(ctx, id); err != nil {
		return err
	}
	if getErr == nil && grant != nil {
		_ = c.BumpGeneration(ctx, grant.Principal)
	}
	_ = clearPermissionAllCrossCache(ctx, c.cacheRepo)
	return nil
}

func (c *RedisCachedPermissionRepository) Has(ctx context.Context, actor, resource model.ID, action model.Action) (bool, error) {
	return c.permissionRepo.Has(ctx, actor, resource, action)
}

func (c *RedisCachedPermissionRepository) EffectiveActions(ctx context.Context, actor, resource model.ID) ([]model.Action, error) {
	return c.permissionRepo.EffectiveActions(ctx, actor, resource)
}

func (c *RedisCachedPermissionRepository) Explain(ctx context.Context, actor, resource model.ID, action model.Action) (*Decision, error) {
	return c.permissionRepo.Explain(ctx, actor, resource, action)
}

func (c *RedisCachedPermissionRepository) ListVisible(ctx context.Context, actor model.ID, action model.Action, parent model.ID, resourceType model.ResourceType) ([]model.ID, error) {
	return c.permissionRepo.ListVisible(ctx, actor, action, parent, resourceType)
}

func (c *RedisCachedPermissionRepository) ListGrantScopes(ctx context.Context, actor model.ID, action model.Action) ([]model.ID, error) {
	return c.permissionRepo.ListGrantScopes(ctx, actor, action)
}

func (c *RedisCachedPermissionRepository) ListScopeAncestry(ctx context.Context, resource model.ID) ([]model.ID, error) {
	return c.permissionRepo.ListScopeAncestry(ctx, resource)
}

func (c *RedisCachedPermissionRepository) LinkInScopeOf(ctx context.Context, child, parent model.ID) error {
	if err := c.permissionRepo.LinkInScopeOf(ctx, child, parent); err != nil {
		return err
	}
	_ = clearPermissionAllCrossCache(ctx, c.cacheRepo)
	return nil
}

func (c *RedisCachedPermissionRepository) BumpGeneration(ctx context.Context, principal model.ID) error {
	key := authzGenKey(principal)
	var gen int64
	_ = c.cacheRepo.Get(ctx, key, &gen)
	return c.cacheRepo.Set(ctx, key, gen+1)
}

// NewCachedPermissionRepository returns a new CachedPermissionRepository.
func NewCachedPermissionRepository(repo PermissionRepository, opts ...RedisRepositoryOption) (*RedisCachedPermissionRepository, error) {
	r, err := newRedisBaseRepository(opts...)
	if err != nil {
		return nil, err
	}

	return &RedisCachedPermissionRepository{
		cacheRepo:      r,
		permissionRepo: repo,
	}, nil
}

func clearPermissionAllCrossCache(ctx context.Context, r *redisBaseRepository) error {
	if err := clearRolesPattern(ctx, r, "*"); err != nil {
		return err
	}
	if err := clearUsersPattern(ctx, r, "*"); err != nil {
		return err
	}
	if err := clearOrganizationAllLists(ctx, r); err != nil {
		return err
	}
	if err := clearNamespacesAllLists(ctx, r); err != nil {
		return err
	}
	if err := clearProjectsAllList(ctx, r); err != nil {
		return err
	}
	if err := clearIssueAllForProject(ctx, r); err != nil {
		return err
	}
	if err := clearDocumentAllLibrary(ctx, r); err != nil {
		return err
	}
	if err := clearDocumentAllRelated(ctx, r); err != nil {
		return err
	}
	if err := clearDocumentAllByCreator(ctx, r); err != nil {
		return err
	}
	if err := clearFolderAllLists(ctx, r); err != nil {
		return err
	}
	return nil
}

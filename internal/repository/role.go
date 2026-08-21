package repository

import (
	"context"
	"errors"
	"time"

	"github.com/neo4j/neo4j-go-driver/v6/neo4j"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg/convert"
	"github.com/opcotech/elemo/internal/pkg/optional"
)

var (
	ErrRoleAddMember    = errors.New("failed to add member to role")      // member cannot be added to role
	ErrRoleCreate       = errors.New("failed to create role")             // role cannot be created
	ErrRoleDelete       = errors.New("failed to delete role")             // role cannot be deleted
	ErrRoleRead         = errors.New("failed to read role")               // role cannot be read
	ErrRoleRemoveMember = errors.New("failed to remove member from role") // member cannot be removed from role
	ErrRoleUpdate       = errors.New("failed to update role")             // role cannot be updated
)

// Role represents a role persisted by the repository.
type Role struct {
	ID          model.ID   `json:"id"`
	Key         string     `json:"key"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Actions     []string   `json:"actions"`
	MemberCount *int64     `json:"member_count"`
	Permissions []model.ID `json:"permissions"`
	CreatedAt   *time.Time `json:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at"`
}

// CreateRoleOpts holds the data required to create a role.
type CreateRoleOpts struct {
	Name        string
	Description string
	Key         string
	Actions     []string
	CreatedBy   model.ID
	BelongsTo   model.ID
}

// UpdateRoleOpts holds the fields that can be updated on a role.
// Undefined fields (Defined == false) are left unchanged.
type UpdateRoleOpts struct {
	Name        optional.Optional[string]
	Description optional.Optional[string]
	Actions     optional.Optional[[]string]
}

// patch builds a Neo4j property map from defined optional fields.
func (o UpdateRoleOpts) patch() map[string]any {
	p := make(map[string]any)

	if o.Name.Defined {
		p["name"] = *o.Name.Value
	}
	if o.Description.Defined {
		if o.Description.Value == nil {
			p["description"] = nil
		} else {
			p["description"] = *o.Description.Value
		}
	}
	if o.Actions.Defined {
		if o.Actions.Value == nil {
			p["actions"] = []string{}
		} else {
			p["actions"] = *o.Actions.Value
		}
	}

	return p
}

//go:generate go tool mockgen -source=role.go -destination=role_mock_gen.go -package=repository -mock_names "RoleRepository=MockRoleRepository"
type RoleRepository interface {
	Create(ctx context.Context, opts CreateRoleOpts) (*Role, error)
	Get(ctx context.Context, id, belongsTo model.ID, proj RoleProjection) (*Role, error)
	GetByID(ctx context.Context, id model.ID) (*Role, error)
	GetByKey(ctx context.Context, belongsTo model.ID, key string) (*Role, error)
	ListBelongsTo(ctx context.Context, belongsTo model.ID, page CursorPage, proj RoleProjection) (Page[*Role], error)
	ListMembers(ctx context.Context, roleID, belongsTo model.ID, page CursorPage) (Page[*User], error)
	Update(ctx context.Context, id, belongsTo model.ID, opts UpdateRoleOpts) (*Role, error)
	AddMember(ctx context.Context, roleID, memberID, belongsToID model.ID) error
	RemoveMember(ctx context.Context, roleID, memberID, belongsToID model.ID) error
	Delete(ctx context.Context, id, belongsTo model.ID) error
}

// Neo4jRoleRepository is a repository for managing roles.
type Neo4jRoleRepository struct {
	*neo4jBaseRepository
}

func (r *Neo4jRoleRepository) scan(proj RoleProjection) func(rec *neo4j.Record) (*Role, error) {
	return func(rec *neo4j.Record) (*Role, error) {
		node, err := Neo4jRecordNode(rec, "r")
		if err != nil {
			return nil, err
		}

		role := new(Role)
		if err := Neo4jScanIntoStruct(&node, &role, []string{"id"}); err != nil {
			return nil, err
		}

		role.ID, err = Neo4jDecodeID(node, model.ResourceTypeRole)
		if err != nil {
			return nil, err
		}
		if role.Actions == nil {
			role.Actions = []string{}
		}
		if proj.MemberCount {
			memberCount, err := Neo4jParseValueFromRecord[int64](rec, "member_count")
			if err != nil {
				return nil, err
			}
			role.MemberCount = convert.ToPointer(memberCount)
		}
		if proj.Permissions {
			role.Permissions = make([]model.ID, 0)
		}

		return role, nil
	}
}

func (r *Neo4jRoleRepository) applyRoleLoaders(ctx context.Context, tx neo4j.ManagedTransaction, plan QueryPlan, roles []*Role) error {
	if len(plan.Loaders) == 0 || len(roles) == 0 {
		return nil
	}

	roleByID := make(map[string]*Role, len(roles))
	ids := make([]string, 0, len(roles))
	for _, role := range roles {
		if role == nil {
			continue
		}
		id := role.ID.String()
		roleByID[id] = role
		ids = append(ids, id)
	}

	for _, loader := range plan.Loaders {
		query := loader
		query.Params = cloneParams(loader.Params)
		query.Params["ids"] = ids
		if loader.Name == "role.load_permissions" {
			rows, _, err := Neo4jRunQuery(ctx, tx, query, func(rec *neo4j.Record) (struct {
				RoleID        string
				PermissionIDs []model.ID
			}, error) {
				roleID, err := Neo4jParseValueFromRecord[string](rec, "role_id")
				if err != nil {
					return struct {
						RoleID        string
						PermissionIDs []model.ID
					}{}, err
				}
				permissionIDs, err := Neo4jRecordIDs(rec, "permission_ids", model.ResourceTypePermission)
				if err != nil {
					return struct {
						RoleID        string
						PermissionIDs []model.ID
					}{}, err
				}
				return struct {
					RoleID        string
					PermissionIDs []model.ID
				}{RoleID: roleID, PermissionIDs: permissionIDs}, nil
			})
			if err != nil {
				return err
			}
			for _, row := range rows {
				if role := roleByID[row.RoleID]; role != nil {
					role.Permissions = row.PermissionIDs
				}
			}
		}
	}

	return nil
}

// Create creates a new role owned by BelongsTo.
func (r *Neo4jRoleRepository) Create(ctx context.Context, opts CreateRoleOpts) (*Role, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.RoleRepository/Create")
	defer span.End()

	createdAt := time.Now().UTC()
	id := model.MustNewID(model.ResourceTypeRole)

	cypher := `
	MATCH (b:` + opts.BelongsTo.Label() + ` {id: $belongs_to_id})
	MERGE (r:` + id.Label() + ` {id: $role_id})
	ON CREATE SET r += {
		key: $key,
		name: $name,
		description: $description,
		actions: $actions,
		created_at: datetime($created_at)
	}
	MERGE (b)-[:` + EdgeKindDefinesRole.String() + ` {id: $defines_id, created_at: datetime($created_at)}]->(r)
	MERGE (r)-[:` + EdgeKindInScopeOf.String() + ` {id: $scope_id, created_at: datetime($created_at)}]->(b)
	`

	params := map[string]any{
		"belongs_to_id": opts.BelongsTo.String(),
		"role_id":       id.String(),
		"defines_id":    model.NewRawID(),
		"scope_id":      model.NewRawID(),
		"key":           opts.Key,
		"name":          opts.Name,
		"description":   opts.Description,
		"actions":       opts.Actions,
		"created_at":    createdAt.Format(time.RFC3339Nano),
	}

	if err := Neo4jExecuteWriteAndConsume(ctx, r.db, cypher, params); err != nil {
		return nil, errors.Join(err, ErrRoleCreate)
	}

	return r.Get(ctx, id, opts.BelongsTo, RoleDetailProjection())
}

// Get returns a role by ID that belongs to belongsTo.
func (r *Neo4jRoleRepository) Get(ctx context.Context, id, belongsTo model.ID, proj RoleProjection) (*Role, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.RoleRepository/Get")
	defer span.End()

	plan, err := CompileQuery(RoleGetQuery{ID: id, BelongsTo: belongsTo, Projection: proj})
	if err != nil {
		return nil, errors.Join(err, ErrRoleRead)
	}

	var role *Role
	err = Neo4jExecuteReadPlan(ctx, r.db, plan, func(tx neo4j.ManagedTransaction) error {
		var readErr error
		role, _, readErr = Neo4jRunQuerySingle(ctx, tx, plan.Root, r.scan(proj))
		if readErr != nil {
			return readErr
		}
		return r.applyRoleLoaders(ctx, tx, plan, []*Role{role})
	})
	if err != nil {
		return nil, errors.Join(err, ErrRoleRead)
	}

	return role, nil
}

// GetByID returns a role by ID without requiring the owner resource.
func (r *Neo4jRoleRepository) GetByID(ctx context.Context, id model.ID) (*Role, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.RoleRepository/GetByID")
	defer span.End()

	if err := id.Validate(); err != nil {
		return nil, errors.Join(err, ErrRoleRead)
	}

	cypher := `
	MATCH (r:` + model.ResourceTypeRole.String() + ` {id: $id})
	RETURN r, COUNT { (:` + model.ResourceTypeUser.String() + `)-[:` + EdgeKindMemberOf.String() + `]->(r) } AS member_count`

	role, err := Neo4jExecuteReadAndReadSingle(ctx, r.db, cypher, map[string]any{"id": id.String()}, r.scan(RoleDetailProjection()))
	if err != nil {
		return nil, errors.Join(err, ErrRoleRead)
	}
	return role, nil
}

// GetByKey returns the role with the given key that belongs to belongsTo.
func (r *Neo4jRoleRepository) GetByKey(ctx context.Context, belongsTo model.ID, key string) (*Role, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.RoleRepository/GetByKey")
	defer span.End()

	if err := belongsTo.Validate(); err != nil {
		return nil, errors.Join(err, ErrRoleRead)
	}

	cypher := `
	MATCH (:` + belongsTo.Label() + ` {id: $belongs_to_id})-[:` + EdgeKindDefinesRole.String() + `]->(r:` + model.ResourceTypeRole.String() + ` {key: $key})
	RETURN r, COUNT { (:` + model.ResourceTypeUser.String() + `)-[:` + EdgeKindMemberOf.String() + `]->(r) } AS member_count`

	role, err := Neo4jExecuteReadAndReadSingle(ctx, r.db, cypher, map[string]any{
		"belongs_to_id": belongsTo.String(),
		"key":           key,
	}, r.scan(RoleDetailProjection()))
	if err != nil {
		return nil, errors.Join(err, ErrRoleRead)
	}
	return role, nil
}

// ListBelongsTo returns a cursor-paginated page of roles for belongsTo.
func (r *Neo4jRoleRepository) ListBelongsTo(ctx context.Context, belongsTo model.ID, page CursorPage, proj RoleProjection) (Page[*Role], error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.RoleRepository/ListBelongsTo")
	defer span.End()

	normalized, err := page.Normalize()
	if err != nil {
		return Page[*Role]{}, errors.Join(ErrRoleRead, err)
	}
	plan, err := CompileQuery(RoleListBelongsToQuery{
		BelongsTo:  belongsTo,
		Page:       normalized,
		Order:      SortDirectionDesc,
		Projection: proj,
	})
	if err != nil {
		return Page[*Role]{}, errors.Join(ErrRoleRead, err)
	}

	roles := make([]*Role, 0)
	err = Neo4jExecuteReadPlan(ctx, r.db, plan, func(tx neo4j.ManagedTransaction) error {
		var readErr error
		roles, _, readErr = Neo4jRunQuery(ctx, tx, plan.Root, r.scan(proj))
		if readErr != nil {
			return readErr
		}
		return r.applyRoleLoaders(ctx, tx, plan, roles)
	})
	if err != nil {
		return Page[*Role]{}, errors.Join(ErrRoleRead, err)
	}

	return PaginateSlice(roles, normalized.Size, func(role *Role) model.ID {
		return role.ID
	})
}

// ListMembers returns a cursor-paginated page of users who are members of the role.
func (r *Neo4jRoleRepository) ListMembers(ctx context.Context, roleID, belongsTo model.ID, page CursorPage) (Page[*User], error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.RoleRepository/ListMembers")
	defer span.End()

	normalized, err := page.Normalize()
	if err != nil {
		return Page[*User]{}, errors.Join(ErrRoleRead, err)
	}
	plan, err := CompileQuery(RoleMemberListQuery{
		RoleID:    roleID,
		BelongsTo: belongsTo,
		Page:      normalized,
		Order:     SortDirectionDesc,
	})
	if err != nil {
		return Page[*User]{}, errors.Join(ErrRoleRead, err)
	}

	users := make([]*User, 0)
	err = Neo4jExecuteReadPlan(ctx, r.db, plan, func(tx neo4j.ManagedTransaction) error {
		scanned, _, readErr := Neo4jRunQuery(ctx, tx, plan.Root, scanRoleMemberUser)
		if readErr != nil {
			return readErr
		}
		users = scanned
		return nil
	})
	if err != nil {
		return Page[*User]{}, errors.Join(ErrRoleRead, err)
	}

	return PaginateSlice(users, normalized.Size, func(user *User) model.ID {
		return user.ID
	})
}

func scanRoleMemberUser(rec *neo4j.Record) (*User, error) {
	user := new(User)
	user.Links = make([]string, 0)
	user.Permissions = make([]model.ID, 0)

	val, _, err := neo4j.GetRecordValue[neo4j.Node](rec, "u")
	if err != nil {
		return nil, err
	}
	if err := Neo4jScanIntoStruct(&val, &user, []string{"id"}); err != nil {
		return nil, err
	}
	user.ID, err = model.NewIDFromString(val.GetProperties()["id"].(string), model.ResourceTypeUser.String())
	if err != nil {
		return nil, err
	}
	return user, nil
}

// Update updates a role that belongs to belongsTo with any given opts.
func (r *Neo4jRoleRepository) Update(ctx context.Context, id, belongsTo model.ID, opts UpdateRoleOpts) (*Role, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.RoleRepository/Update")
	defer span.End()

	cypher := `
	MATCH (:` + belongsTo.Label() + ` {id: $belongs_to_id})-[:` + EdgeKindDefinesRole.String() + `]->(r:` + id.Label() + ` {id: $id})
	SET r += $patch, r.updated_at = datetime()
	RETURN r.id AS id`

	params := map[string]any{
		"id":            id.String(),
		"belongs_to_id": belongsTo.String(),
		"patch":         opts.patch(),
	}

	_, err := Neo4jExecuteWriteAndReadSingle(ctx, r.db, cypher, params, func(_ *neo4j.Record) (*struct{}, error) {
		return &struct{}{}, nil
	})
	if err != nil {
		return nil, errors.Join(err, ErrRoleUpdate)
	}

	return r.Get(ctx, id, belongsTo, RoleDetailProjection())
}

// AddMember adds a user as a member of the role.
func (r *Neo4jRoleRepository) AddMember(ctx context.Context, roleID, memberID, belongsToID model.ID) error {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.RoleRepository/AddMember")
	defer span.End()

	cypher := `
	MATCH (r:` + roleID.Label() + ` {id: $role_id})
	MATCH (u:` + memberID.Label() + ` {id: $member_id})
	MATCH (b:` + belongsToID.Label() + ` {id: $belongs_to_id})
	MERGE (u)-[m:` + EdgeKindMemberOf.String() + `]->(r)
	ON CREATE SET m.created_at = datetime($now), m.id = $membership_id
	ON MATCH SET m.updated_at = datetime($now)`

	params := map[string]any{
		"role_id":       roleID.String(),
		"member_id":     memberID.String(),
		"belongs_to_id": belongsToID.String(),
		"membership_id": model.NewRawID(),
		"now":           time.Now().UTC().Format(time.RFC3339Nano),
	}

	if err := Neo4jExecuteWriteAndConsume(ctx, r.db, cypher, params); err != nil {
		return errors.Join(ErrRoleAddMember, err)
	}

	return nil
}

// RemoveMember removes a user from the role.
func (r *Neo4jRoleRepository) RemoveMember(ctx context.Context, roleID, memberID, belongsToID model.ID) error {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.RoleRepository/RemoveMember")
	defer span.End()

	cypher := `
	MATCH (:` + roleID.Label() + ` {id: $role_id})<-[r:` + EdgeKindMemberOf.String() + `]-(:` + memberID.Label() + ` {id: $member_id})
	MATCH (b:` + belongsToID.Label() + ` {id: $belongs_to_id})
	DELETE r`

	params := map[string]any{
		"role_id":       roleID.String(),
		"member_id":     memberID.String(),
		"belongs_to_id": belongsToID.String(),
	}

	if err := Neo4jExecuteWriteAndConsume(ctx, r.db, cypher, params); err != nil {
		return errors.Join(ErrRoleRemoveMember, err)
	}

	return nil
}

// Delete permanently deletes a role that belongs to belongsTo.
func (r *Neo4jRoleRepository) Delete(ctx context.Context, id, belongsTo model.ID) error {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.RoleRepository/Delete")
	defer span.End()

	cypher := `
	MATCH (:` + belongsTo.Label() + ` {id: $belongs_to_id})-[:` + EdgeKindDefinesRole.String() + `]->(r:` + id.Label() + ` {id: $id})
	DETACH DELETE r`
	params := map[string]any{
		"id":            id.String(),
		"belongs_to_id": belongsTo.String(),
	}

	if err := Neo4jExecuteWriteAndConsume(ctx, r.db, cypher, params); err != nil {
		return errors.Join(ErrRoleDelete, err)
	}

	return nil
}

// NewNeo4jRoleRepository creates a new role neo4jBaseRepository.
func NewNeo4jRoleRepository(opts ...Neo4jRepositoryOption) (*Neo4jRoleRepository, error) {
	baseRepo, err := newNeo4jRepository(opts...)
	if err != nil {
		return nil, err
	}

	return &Neo4jRoleRepository{
		neo4jBaseRepository: baseRepo,
	}, nil
}

func clearRolesPattern(ctx context.Context, r *redisBaseRepository, pattern ...string) error {
	return r.DeletePattern(ctx, composeCacheKey(model.ResourceTypeRole.String(), pattern))
}

func clearRolesKey(ctx context.Context, r *redisBaseRepository, id model.ID) error {
	if err := clearRolesPattern(ctx, r, "Get", id.String(), "*"); err != nil {
		return err
	}
	return clearRolesPattern(ctx, r, "GetByID", id.String())
}

func clearRolesGetByKey(ctx context.Context, r *redisBaseRepository, belongsTo model.ID) error {
	return clearRolesPattern(ctx, r, "GetByKey", belongsTo.String(), "*")
}

func clearRolesBelongsTo(ctx context.Context, r *redisBaseRepository, id model.ID) error {
	return clearRolesPattern(ctx, r, "ListBelongsTo", id.String(), "*", "*", "*")
}

func clearRolesAllBelongsTo(ctx context.Context, r *redisBaseRepository) error {
	return clearRolesPattern(ctx, r, "ListBelongsTo", "*")
}

func clearRoleAllCrossCache(ctx context.Context, r *redisBaseRepository) error {
	deleteFns := []func(context.Context, *redisBaseRepository, ...string) error{
		clearOrganizationsPattern,
		clearProjectsPattern,
	}

	for _, fn := range deleteFns {
		if err := fn(ctx, r, "*"); err != nil {
			return err
		}
	}

	return nil
}

// RedisCachedRoleRepository implements caching on the RoleRepository.
type RedisCachedRoleRepository struct {
	cacheRepo *redisBaseRepository
	roleRepo  RoleRepository
}

func (r *RedisCachedRoleRepository) Create(ctx context.Context, opts CreateRoleOpts) (*Role, error) {
	if err := clearRolesBelongsTo(ctx, r.cacheRepo, opts.BelongsTo); err != nil {
		return nil, err
	}
	if err := clearRolesGetByKey(ctx, r.cacheRepo, opts.BelongsTo); err != nil {
		return nil, err
	}
	if err := clearRoleAllCrossCache(ctx, r.cacheRepo); err != nil {
		return nil, err
	}

	role, err := r.roleRepo.Create(ctx, opts)
	if err != nil {
		return nil, err
	}
	if err := bumpIssueListAuthzEpoch(ctx, r.cacheRepo); err != nil {
		return nil, err
	}
	return role, nil
}

func (r *RedisCachedRoleRepository) Get(ctx context.Context, id, belongsTo model.ID, proj RoleProjection) (*Role, error) {
	var role *Role
	var err error

	key := composeCacheKey(model.ResourceTypeRole.String(), "Get", id.String(), projectionCacheValue(proj))
	if err = r.cacheRepo.Get(ctx, key, &role); err != nil {
		return nil, err
	}

	if role != nil {
		return role, nil
	}

	if role, err = r.roleRepo.Get(ctx, id, belongsTo, proj); err != nil {
		return nil, err
	}

	if err = r.cacheRepo.Set(ctx, key, role); err != nil {
		return nil, err
	}

	return role, nil
}

func (r *RedisCachedRoleRepository) GetByID(ctx context.Context, id model.ID) (*Role, error) {
	var role *Role
	var err error

	key := composeCacheKey(model.ResourceTypeRole.String(), "GetByID", id.String())
	if err = r.cacheRepo.Get(ctx, key, &role); err != nil {
		return nil, err
	}

	if role != nil {
		return role, nil
	}

	if role, err = r.roleRepo.GetByID(ctx, id); err != nil {
		return nil, err
	}

	if err = r.cacheRepo.Set(ctx, key, role); err != nil {
		return nil, err
	}

	return role, nil
}

func (r *RedisCachedRoleRepository) GetByKey(ctx context.Context, belongsTo model.ID, key string) (*Role, error) {
	var role *Role
	var err error

	cacheKey := composeCacheKey(model.ResourceTypeRole.String(), "GetByKey", belongsTo.String(), key)
	if err = r.cacheRepo.Get(ctx, cacheKey, &role); err != nil {
		return nil, err
	}

	if role != nil {
		return role, nil
	}

	if role, err = r.roleRepo.GetByKey(ctx, belongsTo, key); err != nil {
		return nil, err
	}

	if err = r.cacheRepo.Set(ctx, cacheKey, role); err != nil {
		return nil, err
	}

	return role, nil
}

func (r *RedisCachedRoleRepository) ListBelongsTo(ctx context.Context, belongsTo model.ID, page CursorPage, proj RoleProjection) (Page[*Role], error) {
	var roles Page[*Role]
	var err error

	normalized, err := normalizedPage(page)
	if err != nil {
		return Page[*Role]{}, err
	}

	key := composeCacheKey(model.ResourceTypeRole.String(), "ListBelongsTo", belongsTo.String(), projectionCacheValue(proj), pageTokenValue(normalized.Token), normalized.Size)
	if err = r.cacheRepo.Get(ctx, key, &roles); err != nil {
		return Page[*Role]{}, err
	}

	if roles.Items != nil {
		return roles, nil
	}

	if roles, err = r.roleRepo.ListBelongsTo(ctx, belongsTo, normalized, proj); err != nil {
		return Page[*Role]{}, err
	}

	if err = r.cacheRepo.Set(ctx, key, roles); err != nil {
		return Page[*Role]{}, err
	}

	return roles, nil
}

func (r *RedisCachedRoleRepository) ListMembers(ctx context.Context, roleID, belongsTo model.ID, page CursorPage) (Page[*User], error) {
	return r.roleRepo.ListMembers(ctx, roleID, belongsTo, page)
}

func (r *RedisCachedRoleRepository) Update(ctx context.Context, id, belongsTo model.ID, opts UpdateRoleOpts) (*Role, error) {
	role, err := r.roleRepo.Update(ctx, id, belongsTo, opts)
	if err != nil {
		return nil, err
	}

	key := composeCacheKey(model.ResourceTypeRole.String(), "Get", id.String(), projectionCacheValue(RoleDetailProjection()))
	if err = r.cacheRepo.Set(ctx, key, role); err != nil {
		return nil, err
	}

	if err := clearRolesPattern(ctx, r.cacheRepo, "GetByID", id.String()); err != nil {
		return nil, err
	}
	if err := clearRolesGetByKey(ctx, r.cacheRepo, belongsTo); err != nil {
		return nil, err
	}
	if err := clearRolesAllBelongsTo(ctx, r.cacheRepo); err != nil {
		return nil, err
	}
	if opts.Actions.Defined {
		if err := clearPermissionAllCrossCache(ctx, r.cacheRepo); err != nil {
			return nil, err
		}
	}
	if err := bumpIssueListAuthzEpoch(ctx, r.cacheRepo); err != nil {
		return nil, err
	}

	return role, nil
}

func (r *RedisCachedRoleRepository) AddMember(ctx context.Context, roleID, memberID, belongsToID model.ID) error {
	if err := clearRolesKey(ctx, r.cacheRepo, roleID); err != nil {
		return err
	}
	if err := clearRolesAllBelongsTo(ctx, r.cacheRepo); err != nil {
		return err
	}
	// Clear organization cache since GetMembers includes role information
	if err := clearOrganizationsKey(ctx, r.cacheRepo, belongsToID); err != nil {
		return err
	}

	if err := r.roleRepo.AddMember(ctx, roleID, memberID, belongsToID); err != nil {
		return err
	}
	return bumpIssueListAuthzEpoch(ctx, r.cacheRepo)
}

func (r *RedisCachedRoleRepository) RemoveMember(ctx context.Context, roleID, memberID, belongsToID model.ID) error {
	if err := clearRolesKey(ctx, r.cacheRepo, roleID); err != nil {
		return err
	}
	if err := clearRolesAllBelongsTo(ctx, r.cacheRepo); err != nil {
		return err
	}
	// Clear organization cache since GetMembers includes role information
	if err := clearOrganizationsKey(ctx, r.cacheRepo, belongsToID); err != nil {
		return err
	}

	if err := r.roleRepo.RemoveMember(ctx, roleID, memberID, belongsToID); err != nil {
		return err
	}
	return bumpIssueListAuthzEpoch(ctx, r.cacheRepo)
}

func (r *RedisCachedRoleRepository) Delete(ctx context.Context, id, belongsTo model.ID) error {
	if err := clearRolesKey(ctx, r.cacheRepo, id); err != nil {
		return err
	}
	if err := clearRolesGetByKey(ctx, r.cacheRepo, belongsTo); err != nil {
		return err
	}
	if err := clearRolesAllBelongsTo(ctx, r.cacheRepo); err != nil {
		return err
	}
	if err := clearRoleAllCrossCache(ctx, r.cacheRepo); err != nil {
		return err
	}

	if err := r.roleRepo.Delete(ctx, id, belongsTo); err != nil {
		return err
	}
	return bumpIssueListAuthzEpoch(ctx, r.cacheRepo)
}

// NewCachedRoleRepository returns a new CachedRoleRepository.
func NewCachedRoleRepository(repo RoleRepository, opts ...RedisRepositoryOption) (*RedisCachedRoleRepository, error) {
	r, err := newRedisBaseRepository(opts...)
	if err != nil {
		return nil, err
	}

	return &RedisCachedRoleRepository{
		cacheRepo: r,
		roleRepo:  repo,
	}, nil
}

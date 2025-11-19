package repository

import (
	"context"
	"errors"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/opcotech/elemo/internal/model"
)

var (
	ErrRoleAddMember    = errors.New("failed to add member to role")      // member cannot be added to role
	ErrRoleCreate       = errors.New("failed to create role")             // role cannot be created
	ErrRoleDelete       = errors.New("failed to delete role")             // role cannot be deleted
	ErrRoleRead         = errors.New("failed to read role")               // role cannot be read
	ErrRoleRemoveMember = errors.New("failed to remove member from role") // member cannot be removed from role
	ErrRoleUpdate       = errors.New("failed to update role")             // role cannot be updated
)

//go:generate mockgen -source=role.go -destination=../testutil/mock/role_repo_gen.go -package=mock -mock_names "RoleRepository=RoleRepository"
type RoleRepository interface {
	Create(ctx context.Context, createdBy, belongsTo model.ID, role *model.Role) error
	Get(ctx context.Context, id, belongsTo model.ID) (*model.Role, error)
	GetAllBelongsTo(ctx context.Context, belongsTo model.ID, offset, limit int) ([]*model.Role, error)
	Update(ctx context.Context, id, belongsTo model.ID, patch map[string]any) (*model.Role, error)
	AddMember(ctx context.Context, roleID, memberID, belongsToID model.ID) error
	RemoveMember(ctx context.Context, roleID, memberID, belongsToID model.ID) error
	Delete(ctx context.Context, id, belongsTo model.ID) error
}

// RoleRepository is a repository for managing roles.
type Neo4jRoleRepository struct {
	*neo4jBaseRepository
}

func (r *Neo4jRoleRepository) scan(rp, mp, pp string) func(rec *neo4j.Record) (*model.Role, error) {
	return func(rec *neo4j.Record) (*model.Role, error) {
		role := new(model.Role)

		val, _, err := neo4j.GetRecordValue[neo4j.Node](rec, rp)
		if err != nil {
			return nil, err
		}

		if err := Neo4jScanIntoStruct(&val, &role, []string{"id"}); err != nil {
			return nil, err
		}

		role.ID, _ = model.NewIDFromString(val.GetProperties()["id"].(string), model.ResourceTypeRole.String())

		if role.Members, err = Neo4jParseIDsFromRecord(rec, mp, model.ResourceTypeUser.String()); err != nil {
			return nil, err
		}

		if role.Permissions, err = Neo4jParseIDsFromRecord(rec, pp, model.ResourceTypePermission.String()); err != nil {
			return nil, err
		}

		if err := role.Validate(); err != nil {
			return nil, err
		}

		return role, nil
	}
}

func (r *Neo4jRoleRepository) Create(ctx context.Context, createdBy, belongsTo model.ID, role *model.Role) error {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.RoleRepository/Create")
	defer span.End()

	if err := belongsTo.Validate(); err != nil {
		return errors.Join(ErrRoleCreate, err)
	}

	if err := role.Validate(); err != nil {
		return errors.Join(ErrRoleCreate, err)
	}

	createdAt := time.Now().UTC()

	role.ID = model.MustNewID(model.ResourceTypeRole)
	role.CreatedAt = &createdAt
	role.UpdatedAt = nil

	cypher := `
	MATCH (u:` + createdBy.Label() + ` {id: $owner_id})
	MATCH (b:` + belongsTo.Label() + ` {id: $belongs_to_id})
	MERGE (r:` + role.ID.Label() + ` {id: $role_id})
	ON CREATE SET r += { name: $name, description: $description, created_at: datetime($created_at) }
	CREATE (r)<-[:` + EdgeKindHasTeam.String() + ` { id: $has_team_id, created_at: datetime($created_at) }]-(b)
	CREATE (u)-[:` + EdgeKindMemberOf.String() + ` { id: $membership_id, created_at: datetime($created_at) }]->(r)
	MERGE (u)-[p:` + EdgeKindHasPermission.String() + ` {id: $perm_id, kind: $perm_kind}]->(r) ON CREATE SET p.created_at = datetime($created_at)
	`

	params := map[string]any{
		"owner_id":      createdBy.String(),
		"belongs_to_id": belongsTo.String(),
		"role_id":       role.ID.String(),
		"membership_id": model.NewRawID(),
		"has_team_id":   model.NewRawID(),
		"perm_id":       model.NewRawID(),
		"perm_kind":     model.PermissionKindAll.String(),
		"name":          role.Name,
		"description":   role.Description,
		"created_at":    createdAt.Format(time.RFC3339Nano),
	}

	if err := Neo4jExecuteWriteAndConsume(ctx, r.db, cypher, params); err != nil {
		return errors.Join(err, ErrRoleCreate)
	}

	return nil
}

func (r *Neo4jRoleRepository) Get(ctx context.Context, id, belongsTo model.ID) (*model.Role, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.RoleRepository/Get")
	defer span.End()

	cypher := `
	MATCH (r:` + id.Label() + ` {id: $id})
	MATCH (b:` + belongsTo.Label() + ` {id: $belongs_to_id})
	OPTIONAL MATCH (r)<-[:` + EdgeKindMemberOf.String() + `]-(u:` + model.ResourceTypeUser.String() + `)
	OPTIONAL MATCH (r)-[p:` + EdgeKindHasPermission.String() + `]->()
	RETURN r, collect(DISTINCT u.id) AS m, collect(DISTINCT p.id) AS p
	`

	params := map[string]any{
		"id":            id.String(),
		"belongs_to_id": belongsTo.String(),
	}

	role, err := Neo4jExecuteReadAndReadSingle(ctx, r.db, cypher, params, r.scan("r", "m", "p"))
	if err != nil {
		return nil, errors.Join(err, ErrRoleRead)
	}

	return role, nil
}

func (r *Neo4jRoleRepository) GetAllBelongsTo(ctx context.Context, belongsTo model.ID, offset, limit int) ([]*model.Role, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.RoleRepository/GetAllBelongsTo")
	defer span.End()

	if err := belongsTo.Validate(); err != nil {
		return nil, errors.Join(ErrRoleRead, err)
	}

	cypher := `
	MATCH (r:` + model.ResourceTypeRole.String() + `)<-[:` + EdgeKindHasTeam.String() + `]-(:` + belongsTo.Label() + ` {id: $id})
	OPTIONAL MATCH (r)<-[:` + EdgeKindMemberOf.String() + `]-(u:` + model.ResourceTypeUser.String() + `)
	OPTIONAL MATCH (r)-[p:` + EdgeKindHasPermission.String() + `]->()
	RETURN r, collect(DISTINCT u.id) AS m, collect(DISTINCT p.id) AS p
	ORDER BY r.created_at DESC
	SKIP $offset LIMIT $limit`

	params := map[string]any{
		"id":     belongsTo.String(),
		"offset": offset,
		"limit":  limit,
	}

	roles, err := Neo4jExecuteReadAndReadAll(ctx, r.db, cypher, params, r.scan("r", "m", "p"))
	if err != nil {
		return nil, errors.Join(ErrRoleRead, err)
	}

	return roles, nil
}

func (r *Neo4jRoleRepository) Update(ctx context.Context, id, belongsTo model.ID, patch map[string]any) (*model.Role, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.RoleRepository/Update")
	defer span.End()

	cypher := `
	MATCH (r:` + id.Label() + ` {id: $id})
	MATCH (b:` + belongsTo.Label() + ` {id: $belongs_to_id})
	SET r += $patch, r.updated_at = datetime()
	WITH r
	OPTIONAL MATCH (r)<-[:` + EdgeKindMemberOf.String() + `]-(u:` + model.ResourceTypeUser.String() + `)
	OPTIONAL MATCH (r)-[p:` + EdgeKindHasPermission.String() + `]->()
	RETURN r, collect(DISTINCT u.id) AS m, collect(DISTINCT p.id) AS p`

	params := map[string]any{
		"id":            id.String(),
		"belongs_to_id": belongsTo.String(),
		"patch":         patch,
	}

	role, err := Neo4jExecuteWriteAndReadSingle(ctx, r.db, cypher, params, r.scan("r", "m", "p"))
	if err != nil {
		return nil, errors.Join(err, ErrRoleUpdate)
	}

	return role, nil
}

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

func (r *Neo4jRoleRepository) Delete(ctx context.Context, id, belongsTo model.ID) error {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.RoleRepository/Delete")
	defer span.End()

	cypher := `MATCH (r:` + id.Label() + ` {id: $id})
	MATCH (b:` + belongsTo.Label() + ` {id: $belongs_to_id})
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
	return r.Delete(ctx, composeCacheKey(model.ResourceTypeRole.String(), id.String()))
}

func clearRolesBelongsTo(ctx context.Context, r *redisBaseRepository, id model.ID) error {
	return clearRolesPattern(ctx, r, "GetAllBelongsTo", id.String(), "*")
}

func clearRolesAllBelongsTo(ctx context.Context, r *redisBaseRepository) error {
	return clearRolesPattern(ctx, r, "GetAllBelongsTo", "*")
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

// CachedRoleRepository implements caching on the
// repository.RoleRepository.
type RedisCachedRoleRepository struct {
	cacheRepo *redisBaseRepository
	roleRepo  RoleRepository
}

func (r *RedisCachedRoleRepository) Create(ctx context.Context, createdBy, belongsTo model.ID, role *model.Role) error {
	if err := clearRolesBelongsTo(ctx, r.cacheRepo, belongsTo); err != nil {
		return err
	}
	if err := clearRoleAllCrossCache(ctx, r.cacheRepo); err != nil {
		return err
	}

	return r.roleRepo.Create(ctx, createdBy, belongsTo, role)
}

func (r *RedisCachedRoleRepository) Get(ctx context.Context, id, belongsTo model.ID) (*model.Role, error) {
	var role *model.Role
	var err error

	key := composeCacheKey(model.ResourceTypeRole.String(), id.String())
	if err = r.cacheRepo.Get(ctx, key, &role); err != nil {
		return nil, err
	}

	if role != nil {
		return role, nil
	}

	if role, err = r.roleRepo.Get(ctx, id, belongsTo); err != nil {
		return nil, err
	}

	if err = r.cacheRepo.Set(ctx, key, role); err != nil {
		return nil, err
	}

	return role, nil
}

func (r *RedisCachedRoleRepository) GetAllBelongsTo(ctx context.Context, belongsTo model.ID, offset, limit int) ([]*model.Role, error) {
	var roles []*model.Role
	var err error

	key := composeCacheKey(model.ResourceTypeRole.String(), "GetAllBelongsTo", belongsTo.String(), offset, limit)
	if err = r.cacheRepo.Get(ctx, key, &roles); err != nil {
		return nil, err
	}

	if roles != nil {
		return roles, nil
	}

	if roles, err = r.roleRepo.GetAllBelongsTo(ctx, belongsTo, offset, limit); err != nil {
		return nil, err
	}

	if err = r.cacheRepo.Set(ctx, key, roles); err != nil {
		return nil, err
	}

	return roles, nil
}

func (r *RedisCachedRoleRepository) Update(ctx context.Context, id, belongsTo model.ID, patch map[string]any) (*model.Role, error) {
	var role *model.Role
	var err error

	role, err = r.roleRepo.Update(ctx, id, belongsTo, patch)
	if err != nil {
		return nil, err
	}

	key := composeCacheKey(model.ResourceTypeRole.String(), id.String())
	if err = r.cacheRepo.Set(ctx, key, role); err != nil {
		return nil, err
	}

	if err := clearRolesAllBelongsTo(ctx, r.cacheRepo); err != nil {
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

	return r.roleRepo.AddMember(ctx, roleID, memberID, belongsToID)
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

	return r.roleRepo.RemoveMember(ctx, roleID, memberID, belongsToID)
}

func (r *RedisCachedRoleRepository) Delete(ctx context.Context, id, belongsTo model.ID) error {
	if err := clearRolesKey(ctx, r.cacheRepo, id); err != nil {
		return err
	}

	if err := clearRolesAllBelongsTo(ctx, r.cacheRepo); err != nil {
		return err
	}

	if err := clearRoleAllCrossCache(ctx, r.cacheRepo); err != nil {
		return err
	}

	return r.roleRepo.Delete(ctx, id, belongsTo)
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

package neo4j

import (
	"context"
	"errors"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/repository"
)

// RoleRepository is a repository for managing roles.
type RoleRepository struct {
	*baseRepository
}

func (r *RoleRepository) scan(rp, mp, pp string) func(rec *neo4j.Record) (*model.Role, error) {
	return func(rec *neo4j.Record) (*model.Role, error) {
		role := new(model.Role)

		val, _, err := neo4j.GetRecordValue[neo4j.Node](rec, rp)
		if err != nil {
			return nil, err
		}

		if err := ScanIntoStruct(&val, &role, []string{"id"}); err != nil {
			return nil, err
		}

		role.ID, _ = model.NewIDFromString(val.GetProperties()["id"].(string), model.ResourceTypeRole.String())

		if role.Members, err = ParseIDsFromRecord(rec, mp, model.ResourceTypeUser.String()); err != nil {
			return nil, err
		}

		if role.Permissions, err = ParseIDsFromRecord(rec, pp, model.ResourceTypePermission.String()); err != nil {
			return nil, err
		}

		if err := role.Validate(); err != nil {
			return nil, err
		}

		return role, nil
	}
}

func (r *RoleRepository) Create(ctx context.Context, createdBy, belongsTo model.ID, role *model.Role) error {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.RoleRepository/Create")
	defer span.End()

	if err := belongsTo.Validate(); err != nil {
		return errors.Join(repository.ErrRoleCreate, err)
	}

	if err := role.Validate(); err != nil {
		return errors.Join(repository.ErrRoleCreate, err)
	}

	createdAt := time.Now()

	role.ID = model.MustNewID(model.ResourceTypeRole)
	role.CreatedAt = &createdAt
	role.UpdatedAt = nil

	cypher := `
	MATCH (u:` + createdBy.Label() + ` {id: $owner_id}), (b:` + belongsTo.Label() + ` {id: $belongs_to_id})
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

	if err := ExecuteWriteAndConsume(ctx, r.db, cypher, params); err != nil {
		return errors.Join(err, repository.ErrRoleCreate)
	}

	return nil
}

func (r *RoleRepository) Get(ctx context.Context, id model.ID) (*model.Role, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.RoleRepository/Get")
	defer span.End()

	cypher := `
	MATCH (r:` + id.Label() + ` {id: $id})
	OPTIONAL MATCH (r)<-[:` + EdgeKindMemberOf.String() + `]-(u:` + model.ResourceTypeUser.String() + `)
	OPTIONAL MATCH (r)-[p:` + EdgeKindHasPermission.String() + `]->()
	RETURN r, collect(DISTINCT u.id) AS m, collect(DISTINCT p.id) AS p
	`

	params := map[string]any{
		"id": id.String(),
	}

	role, err := ExecuteReadAndReadSingle(ctx, r.db, cypher, params, r.scan("r", "m", "p"))
	if err != nil {
		return nil, errors.Join(err, repository.ErrRoleRead)
	}

	return role, nil
}

func (r *RoleRepository) GetAllBelongsTo(ctx context.Context, belongsTo model.ID, offset, limit int) ([]*model.Role, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.RoleRepository/GetAllBelongsTo")
	defer span.End()

	if err := belongsTo.Validate(); err != nil {
		return nil, errors.Join(repository.ErrRoleRead, err)
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

	roles, err := ExecuteReadAndReadAll(ctx, r.db, cypher, params, r.scan("r", "m", "p"))
	if err != nil {
		return nil, errors.Join(repository.ErrRoleRead, err)
	}

	return roles, nil
}

func (r *RoleRepository) Update(ctx context.Context, id model.ID, patch map[string]any) (*model.Role, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.RoleRepository/Update")
	defer span.End()

	cypher := `
	MATCH (r:` + id.Label() + ` {id: $id}) SET r += $patch, r.updated_at = datetime($updated_at)
	WITH r
	OPTIONAL MATCH (r)<-[:` + EdgeKindMemberOf.String() + `]-(u:` + model.ResourceTypeUser.String() + `)
	OPTIONAL MATCH (r)-[p:` + EdgeKindHasPermission.String() + `]->()
	RETURN r, collect(DISTINCT u.id) AS m, collect(DISTINCT p.id) AS p`

	params := map[string]any{
		"id":         id.String(),
		"patch":      patch,
		"updated_at": time.Now().Format(time.RFC3339Nano),
	}

	role, err := ExecuteWriteAndReadSingle(ctx, r.db, cypher, params, r.scan("r", "m", "p"))
	if err != nil {
		return nil, errors.Join(err, repository.ErrRoleUpdate)
	}

	return role, nil
}

func (r *RoleRepository) AddMember(ctx context.Context, roleID, memberID model.ID) error {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.RoleRepository/AddMember")
	defer span.End()

	cypher := `
	MATCH (r:` + roleID.Label() + ` {id: $role_id}), (u:` + memberID.Label() + ` {id: $member_id})
	MERGE (u)-[m:` + EdgeKindMemberOf.String() + `]->(r)
	ON CREATE SET m.created_at = datetime($now), m.id = $membership_id
	ON MATCH SET m.updated_at = datetime($now)`

	params := map[string]any{
		"role_id":       roleID.String(),
		"member_id":     memberID.String(),
		"membership_id": model.NewRawID(),
		"now":           time.Now().Format(time.RFC3339Nano),
	}

	if err := ExecuteWriteAndConsume(ctx, r.db, cypher, params); err != nil {
		return errors.Join(repository.ErrRoleAddMember, err)
	}

	return nil
}

func (r *RoleRepository) RemoveMember(ctx context.Context, roleID, memberID model.ID) error {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.RoleRepository/RemoveMember")
	defer span.End()

	cypher := `
	MATCH (:` + roleID.Label() + ` {id: $role_id})<-[r:` + EdgeKindMemberOf.String() + `]-(:` + memberID.Label() + ` {id: $member_id})
	DELETE r`

	params := map[string]any{
		"role_id":   roleID.String(),
		"member_id": memberID.String(),
	}

	if err := ExecuteWriteAndConsume(ctx, r.db, cypher, params); err != nil {
		return errors.Join(repository.ErrRoleRemoveMember, err)
	}

	return nil
}

func (r *RoleRepository) Delete(ctx context.Context, id model.ID) error {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.RoleRepository/Delete")
	defer span.End()

	cypher := `MATCH (r:` + id.Label() + ` {id: $id}) DETACH DELETE r`
	params := map[string]any{
		"id": id.String(),
	}

	if err := ExecuteWriteAndConsume(ctx, r.db, cypher, params); err != nil {
		return errors.Join(repository.ErrRoleDelete, err)
	}

	return nil
}

// NewRoleRepository creates a new role baseRepository.
func NewRoleRepository(opts ...RepositoryOption) (*RoleRepository, error) {
	baseRepo, err := newRepository(opts...)
	if err != nil {
		return nil, err
	}

	return &RoleRepository{
		baseRepository: baseRepo,
	}, nil
}

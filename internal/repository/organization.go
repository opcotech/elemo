package repository

import (
	"context"
	"errors"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/opcotech/elemo/internal/model"
)

var (
	ErrOrganizationAddMember    = errors.New("failed to add member to organization")      // member cannot be added to organization
	ErrOrganizationCreate       = errors.New("failed to create organization")             // organization cannot be created
	ErrOrganizationDelete       = errors.New("failed to delete organization")             // organization cannot be deleted
	ErrOrganizationRead         = errors.New("failed to read organization")               // organization cannot be read
	ErrOrganizationRemoveMember = errors.New("failed to remove member from organization") // member cannot be removed from organization
	ErrOrganizationUpdate       = errors.New("failed to update organization")             // organization cannot be updated
)

//go:generate mockgen -source=organization.go -destination=../testutil/mock/organization_repo_gen.go -package=mock -mock_names "OrganizationRepository=OrganizationRepository"
type OrganizationRepository interface {
	Create(ctx context.Context, owner model.ID, organization *model.Organization) error
	Get(ctx context.Context, id model.ID) (*model.Organization, error)
	GetAll(ctx context.Context, userID model.ID, offset, limit int) ([]*model.Organization, error)
	Update(ctx context.Context, id model.ID, patch map[string]any) (*model.Organization, error)
	GetMembers(ctx context.Context, orgID model.ID) ([]*model.OrganizationMember, error)
	AddMember(ctx context.Context, orgID, memberID model.ID) error
	RemoveMember(ctx context.Context, orgID, memberID model.ID) error
	AddInvitation(ctx context.Context, orgID, userID model.ID) error
	RemoveInvitation(ctx context.Context, orgID, userID model.ID) error
	GetInvitations(ctx context.Context, orgID model.ID) ([]*model.OrganizationMember, error)
	Delete(ctx context.Context, id model.ID) error
}

// OrganizationRepository is a repository for managing organizations.
type Neo4jOrganizationRepository struct {
	*neo4jBaseRepository
}

func (r *Neo4jOrganizationRepository) scan(op, np, tp, mp string) func(rec *neo4j.Record) (*model.Organization, error) {
	return func(rec *neo4j.Record) (*model.Organization, error) {
		org := new(model.Organization)

		val, _, err := neo4j.GetRecordValue[neo4j.Node](rec, op)
		if err != nil {
			return nil, err
		}

		if err := Neo4jScanIntoStruct(&val, &org, []string{"id"}); err != nil {
			return nil, err
		}

		org.ID, _ = model.NewIDFromString(val.GetProperties()["id"].(string), model.ResourceTypeOrganization.String())

		if org.Namespaces, err = Neo4jParseIDsFromRecord(rec, np, model.ResourceTypeNamespace.String()); err != nil {
			return nil, err
		}

		if org.Teams, err = Neo4jParseIDsFromRecord(rec, tp, model.ResourceTypeRole.String()); err != nil {
			return nil, err
		}

		if org.Members, err = Neo4jParseIDsFromRecord(rec, mp, model.ResourceTypeUser.String()); err != nil {
			return nil, err
		}

		if err := org.Validate(); err != nil {
			return nil, err
		}

		return org, nil
	}
}

func (r *Neo4jOrganizationRepository) scanOrganizationMember(up string) func(rec *neo4j.Record) (model.OrganizationMember, error) {
	return func(rec *neo4j.Record) (model.OrganizationMember, error) {
		val, _, err := neo4j.GetRecordValue[neo4j.Node](rec, up)
		if err != nil {
			return model.OrganizationMember{}, err
		}

		userID, err := model.NewIDFromString(val.GetProperties()["id"].(string), model.ResourceTypeUser.String())
		if err != nil {
			return model.OrganizationMember{}, err
		}

		firstName := ""
		if v, ok := val.GetProperties()["first_name"]; ok {
			firstName = v.(string)
		}

		lastName := ""
		if v, ok := val.GetProperties()["last_name"]; ok {
			lastName = v.(string)
		}

		email := ""
		if v, ok := val.GetProperties()["email"]; ok {
			email = v.(string)
		}

		var picture *string
		if v, ok := val.GetProperties()["picture"]; ok && v != nil {
			pic := v.(string)
			if pic != "" {
				picture = &pic
			}
		}

		statusStr := ""
		if v, ok := val.GetProperties()["status"]; ok {
			statusStr = v.(string)
		}
		var status model.UserStatus
		if err := status.UnmarshalText([]byte(statusStr)); err != nil {
			return model.OrganizationMember{}, err
		}

		roleNamesVal, err := Neo4jParseValueFromRecord[[]any](rec, "roles")
		if err != nil {
			roleNamesVal = []any{}
		}

		roleNames := make([]string, 0, len(roleNamesVal))
		for _, rn := range roleNamesVal {
			if rn != nil {
				roleNames = append(roleNames, rn.(string))
			}
		}

		member, err := model.NewOrganizationMember(userID, firstName, lastName, email, picture, status, roleNames)
		if err != nil {
			return model.OrganizationMember{}, err
		}

		return *member, nil
	}
}

func (r *Neo4jOrganizationRepository) Create(ctx context.Context, owner model.ID, organization *model.Organization) error {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.OrganizationRepository/Create")
	defer span.End()

	if err := owner.Validate(); err != nil {
		return errors.Join(ErrOrganizationCreate, err)
	}

	if err := organization.Validate(); err != nil {
		return errors.Join(ErrOrganizationCreate, err)
	}

	createdAt := time.Now().UTC()

	organization.ID = model.MustNewID(model.ResourceTypeOrganization)
	organization.CreatedAt = &createdAt
	organization.UpdatedAt = nil

	cypher := `
	MATCH (u:` + owner.Label() + ` {id: $owner_id})
	CREATE (o:` + organization.ID.Label() + ` { id: $id, name: $name, email: $email, logo: $logo, website: $website,
		status: $status, created_at: datetime($created_at)
	}),
	(u)-[:` + EdgeKindMemberOf.String() + ` {id: $membership_id, created_at: datetime($created_at)}]->(o),
	(u)-[:` + EdgeKindHasPermission.String() + `{id: $permission_id, created_at: datetime($created_at), kind: $permission_kind}]->(o)`

	params := map[string]any{
		"id":              organization.ID.String(),
		"name":            organization.Name,
		"email":           organization.Email,
		"logo":            organization.Logo,
		"website":         organization.Website,
		"status":          organization.Status.String(),
		"created_at":      createdAt.Format(time.RFC3339Nano),
		"owner_id":        owner.String(),
		"membership_id":   model.NewRawID(),
		"permission_id":   model.NewRawID(),
		"permission_kind": model.PermissionKindAll.String(),
	}

	if err := Neo4jExecuteWriteAndConsume(ctx, r.db, cypher, params); err != nil {
		return errors.Join(ErrOrganizationCreate, err)
	}

	return nil
}

func (r *Neo4jOrganizationRepository) Get(ctx context.Context, id model.ID) (*model.Organization, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.OrganizationRepository/Get")
	defer span.End()

	cypher := `
	MATCH (o:` + id.Label() + ` {id: $id})
	OPTIONAL MATCH (u:` + model.ResourceTypeUser.String() + `)-[:` + EdgeKindMemberOf.String() + `]->(o)
	OPTIONAL MATCH (o)-[:` + EdgeKindHasNamespace.String() + `]->(n:` + model.ResourceTypeNamespace.String() + `)
	OPTIONAL MATCH (o)-[:` + EdgeKindHasTeam.String() + `]->(t:` + model.ResourceTypeRole.String() + `)
	RETURN o, collect(DISTINCT u.id) AS m, collect(DISTINCT n.id) AS n, collect(DISTINCT t.id) AS t
	`

	params := map[string]any{
		"id": id.String(),
	}

	org, err := Neo4jExecuteReadAndReadSingle(ctx, r.db, cypher, params, r.scan("o", "n", "t", "m"))
	if err != nil {
		return nil, errors.Join(ErrOrganizationRead, err)
	}

	return org, nil
}

func (r *Neo4jOrganizationRepository) GetAll(ctx context.Context, userID model.ID, offset, limit int) ([]*model.Organization, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.OrganizationRepository/GetAllBelongsTo")
	defer span.End()

	if err := userID.Validate(); err != nil {
		return nil, errors.Join(ErrOrganizationRead, err)
	}

	cypher := `
	MATCH (u:` + userID.Label() + ` {id: $user_id})-[m:` + EdgeKindMemberOf.String() + `]->(o:` + model.ResourceTypeOrganization.String() + `)
	OPTIONAL MATCH (u2:` + model.ResourceTypeUser.String() + `)-[:` + EdgeKindMemberOf.String() + `]->(o)
	OPTIONAL MATCH (o)-[:` + EdgeKindHasNamespace.String() + `]->(n:` + model.ResourceTypeNamespace.String() + `)
	OPTIONAL MATCH (o)-[:` + EdgeKindHasTeam.String() + `]->(t:` + model.ResourceTypeRole.String() + `)
	RETURN o, collect(DISTINCT u2.id) AS m, collect(DISTINCT n.id) AS n, collect(DISTINCT t.id) AS t
	ORDER BY o.created_at DESC
	SKIP $offset LIMIT $limit`

	params := map[string]any{
		"user_id": userID.String(),
		"offset":  offset,
		"limit":   limit,
	}

	orgs, err := Neo4jExecuteReadAndReadAll(ctx, r.db, cypher, params, r.scan("o", "n", "t", "m"))
	if err != nil {
		return nil, errors.Join(ErrOrganizationRead, err)
	}

	return orgs, nil
}

func (r *Neo4jOrganizationRepository) Update(ctx context.Context, id model.ID, patch map[string]any) (*model.Organization, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.OrganizationRepository/Update")
	defer span.End()

	cypher := `
	MATCH (o:` + id.Label() + ` {id: $id}) SET o += $patch, o.updated_at = datetime()
	WITH o
	OPTIONAL MATCH (u:` + model.ResourceTypeUser.String() + `)-[:` + EdgeKindMemberOf.String() + `]->(o)
	OPTIONAL MATCH (o)-[:` + EdgeKindHasNamespace.String() + `]->(n:` + model.ResourceTypeNamespace.String() + `)
	OPTIONAL MATCH (o)-[:` + EdgeKindHasTeam.String() + `]->(t:` + model.ResourceTypeRole.String() + `)
	RETURN o, collect(DISTINCT u.id) AS m, collect(DISTINCT n.id) AS n, collect(DISTINCT t.id) AS t`

	params := map[string]any{
		"id":    id.String(),
		"patch": patch,
	}

	org, err := Neo4jExecuteWriteAndReadSingle(ctx, r.db, cypher, params, r.scan("o", "n", "t", "m"))
	if err != nil {
		return nil, errors.Join(ErrOrganizationUpdate, err)
	}

	return org, nil
}

func (r *Neo4jOrganizationRepository) GetMembers(ctx context.Context, orgID model.ID) ([]*model.OrganizationMember, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.OrganizationRepository/GetMembers")
	defer span.End()

	if err := orgID.Validate(); err != nil {
		return nil, errors.Join(ErrOrganizationRead, err)
	}

	cypher := `
	MATCH (o:` + orgID.Label() + ` {id: $org_id})
	MATCH (u:` + model.ResourceTypeUser.String() + `)-[rel:` + EdgeKindMemberOf.String() + `|` + EdgeKindInvitedTo.String() + `]->(o)
	WITH DISTINCT u, o, collect(DISTINCT type(rel)) AS relTypes
	WITH u, o, relTypes, CASE WHEN '` + EdgeKindMemberOf.String() + `' IN relTypes THEN true ELSE false END AS isMember
	WHERE isMember = true OR NOT EXISTS((u)-[:` + EdgeKindMemberOf.String() + `]->(o))
	OPTIONAL MATCH (u)-[:` + EdgeKindMemberOf.String() + `]->(r:` + model.ResourceTypeRole.String() + `)<-[:` + EdgeKindHasTeam.String() + `]-(o)
	WITH u, isMember, collect(DISTINCT r) AS roleNodes
	WITH u, isMember,
	CASE WHEN isMember THEN [role IN roleNodes WHERE role IS NOT NULL | role.name] ELSE [] END AS roles
	RETURN u AS u, roles AS roles, isMember AS isMember
	ORDER BY isMember DESC, u.created_at ASC`

	params := map[string]any{
		"org_id": orgID.String(),
	}

	members, err := Neo4jExecuteReadAndReadAll(ctx, r.db, cypher, params, func(rec *neo4j.Record) (model.OrganizationMember, error) {
		member, err := r.scanOrganizationMember("u")(rec)
		if err != nil {
			return model.OrganizationMember{}, err
		}

		isMemberVal, err := Neo4jParseValueFromRecord[bool](rec, "isMember")
		if err != nil {
			isMemberVal = false
		}

		// If user is not a member (has INVITED_TO but not MEMBER_OF), set status to pending
		if !isMemberVal {
			member.Status = model.UserStatusPending
		}

		return member, nil
	})
	if err != nil {
		return nil, errors.Join(ErrOrganizationRead, err)
	}

	membersPtr := make([]*model.OrganizationMember, len(members))
	for i := range members {
		membersPtr[i] = &members[i]
	}

	return membersPtr, nil
}

func (r *Neo4jOrganizationRepository) AddMember(ctx context.Context, orgID, memberID model.ID) error {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.OrganizationRepository/AddMember")
	defer span.End()

	if err := orgID.Validate(); err != nil {
		return errors.Join(ErrOrganizationAddMember, err)
	}

	if err := memberID.Validate(); err != nil {
		return errors.Join(ErrOrganizationAddMember, err)
	}

	cypher := `
	MATCH (o:` + orgID.Label() + ` {id: $org_id})
	MATCH (u:` + memberID.Label() + ` {id: $member_id})
	MERGE (u)-[m:` + EdgeKindMemberOf.String() + `]->(o)
	ON CREATE SET m.created_at = datetime($now), m.id = $membership_id
	ON MATCH SET m.updated_at = datetime($now)`

	params := map[string]any{
		"org_id":        orgID.String(),
		"member_id":     memberID.String(),
		"membership_id": model.NewRawID(),
		"now":           time.Now().UTC().Format(time.RFC3339Nano),
	}

	if err := Neo4jExecuteWriteAndConsume(ctx, r.db, cypher, params); err != nil {
		return errors.Join(ErrOrganizationAddMember, err)
	}

	return nil
}

func (r *Neo4jOrganizationRepository) RemoveMember(ctx context.Context, orgID, memberID model.ID) error {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.OrganizationRepository/RemoveMember")
	defer span.End()

	if err := orgID.Validate(); err != nil {
		return errors.Join(ErrOrganizationRemoveMember, err)
	}

	if err := memberID.Validate(); err != nil {
		return errors.Join(ErrOrganizationRemoveMember, err)
	}

	cypher := `
	MATCH (:` + memberID.Label() + ` {id: $member_id})-[r:` + EdgeKindMemberOf.String() + `]->(:` + orgID.Label() + ` {id: $org_id})
	DELETE r`

	params := map[string]any{
		"org_id":    orgID.String(),
		"member_id": memberID.String(),
	}

	if err := Neo4jExecuteWriteAndConsume(ctx, r.db, cypher, params); err != nil {
		return errors.Join(ErrOrganizationRemoveMember, err)
	}

	return nil
}

func (r *Neo4jOrganizationRepository) AddInvitation(ctx context.Context, orgID, userID model.ID) error {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.OrganizationRepository/AddInvitation")
	defer span.End()

	if err := orgID.Validate(); err != nil {
		return errors.Join(ErrOrganizationAddMember, err)
	}

	if err := userID.Validate(); err != nil {
		return errors.Join(ErrOrganizationAddMember, err)
	}

	// Check if IDs are nil (invalid)
	if orgID.IsNil() || userID.IsNil() {
		return errors.Join(ErrOrganizationAddMember, model.ErrInvalidID)
	}

	cypher := `
	MATCH (o:` + orgID.Label() + ` {id: $org_id})
	MATCH (u:` + userID.Label() + ` {id: $user_id})
	MERGE (u)-[i:` + EdgeKindInvitedTo.String() + `]->(o)
	ON CREATE SET i.created_at = datetime($now), i.id = $invitation_id
	ON MATCH SET i.updated_at = datetime($now)
	RETURN o.id AS org_id`

	params := map[string]any{
		"org_id":        orgID.String(),
		"user_id":       userID.String(),
		"invitation_id": model.NewRawID(),
		"now":           time.Now().UTC().Format(time.RFC3339Nano),
	}

	_, err := Neo4jExecuteWriteAndReadSingle(ctx, r.db, cypher, params, func(_ *neo4j.Record) (*struct{}, error) {
		return &struct{}{}, nil
	})
	if err != nil {
		return errors.Join(ErrOrganizationAddMember, err)
	}

	return nil
}

func (r *Neo4jOrganizationRepository) RemoveInvitation(ctx context.Context, orgID, userID model.ID) error {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.OrganizationRepository/RemoveInvitation")
	defer span.End()

	if err := orgID.Validate(); err != nil {
		return errors.Join(ErrOrganizationRemoveMember, err)
	}

	if err := userID.Validate(); err != nil {
		return errors.Join(ErrOrganizationRemoveMember, err)
	}

	// Check if IDs are nil (invalid)
	if orgID.IsNil() || userID.IsNil() {
		return errors.Join(ErrOrganizationRemoveMember, model.ErrInvalidID)
	}

	cypher := `
	MATCH (:` + userID.Label() + ` {id: $user_id})-[r:` + EdgeKindInvitedTo.String() + `]->(:` + orgID.Label() + ` {id: $org_id})
	DELETE r`

	params := map[string]any{
		"org_id":  orgID.String(),
		"user_id": userID.String(),
	}

	if err := Neo4jExecuteWriteAndConsume(ctx, r.db, cypher, params); err != nil {
		return errors.Join(ErrOrganizationRemoveMember, err)
	}

	return nil
}

func (r *Neo4jOrganizationRepository) GetInvitations(ctx context.Context, orgID model.ID) ([]*model.OrganizationMember, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.OrganizationRepository/GetInvitations")
	defer span.End()

	if err := orgID.Validate(); err != nil {
		return nil, errors.Join(ErrOrganizationRead, err)
	}

	// Check if ID is nil (invalid)
	if orgID.IsNil() {
		return nil, errors.Join(ErrOrganizationRead, model.ErrInvalidID)
	}

	cypher := `
	MATCH (u:` + model.ResourceTypeUser.String() + `)-[:` + EdgeKindInvitedTo.String() + `]->(o:` + orgID.Label() + ` {id: $org_id})
	RETURN u, [] AS roles
	ORDER BY u.created_at ASC`

	params := map[string]any{
		"org_id": orgID.String(),
	}

	members, err := Neo4jExecuteReadAndReadAll(ctx, r.db, cypher, params, r.scanOrganizationMember("u"))
	if err != nil {
		return nil, errors.Join(ErrOrganizationRead, err)
	}

	membersPtr := make([]*model.OrganizationMember, len(members))
	for i := range members {
		membersPtr[i] = &members[i]
	}

	return membersPtr, nil
}

func (r *Neo4jOrganizationRepository) Delete(ctx context.Context, id model.ID) error {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.OrganizationRepository/Delete")
	defer span.End()

	cypher := `MATCH (o:` + id.Label() + ` {id: $id}), (o)-[r]-() DETACH DELETE o, r`
	params := map[string]any{
		"id": id.String(),
	}

	if err := Neo4jExecuteWriteAndConsume(ctx, r.db, cypher, params); err != nil {
		return errors.Join(ErrOrganizationDelete, err)
	}

	return nil
}

// NewNeo4jOrganizationRepository creates a new organization neo4jBaseRepository.
func NewNeo4jOrganizationRepository(opts ...Neo4jRepositoryOption) (*Neo4jOrganizationRepository, error) {
	baseRepo, err := newNeo4jRepository(opts...)
	if err != nil {
		return nil, err
	}

	return &Neo4jOrganizationRepository{
		neo4jBaseRepository: baseRepo,
	}, nil
}

func clearOrganizationsPattern(ctx context.Context, r *redisBaseRepository, pattern ...string) error {
	return r.DeletePattern(ctx, composeCacheKey(model.ResourceTypeOrganization.String(), pattern))
}

func clearOrganizationsKey(ctx context.Context, r *redisBaseRepository, id model.ID) error {
	return r.Delete(ctx, composeCacheKey(model.ResourceTypeOrganization.String(), id.String()))
}

func clearOrganizationAllGetAll(ctx context.Context, r *redisBaseRepository) error {
	return clearOrganizationsPattern(ctx, r, "GetAll", "*", "*")
}

// CachedOrganizationRepository implements caching on the
// repository.OrganizationRepository.
type RedisCachedOrganizationRepository struct {
	cacheRepo        *redisBaseRepository
	organizationRepo OrganizationRepository
}

func (r *RedisCachedOrganizationRepository) Create(ctx context.Context, owner model.ID, organization *model.Organization) error {
	if err := clearOrganizationAllGetAll(ctx, r.cacheRepo); err != nil {
		return err
	}

	return r.organizationRepo.Create(ctx, owner, organization)
}

func (r *RedisCachedOrganizationRepository) Get(ctx context.Context, id model.ID) (*model.Organization, error) {
	var organization *model.Organization
	var err error

	key := composeCacheKey(model.ResourceTypeOrganization.String(), id.String())
	if err = r.cacheRepo.Get(ctx, key, &organization); err != nil {
		return nil, err
	}

	if organization != nil {
		return organization, nil
	}

	if organization, err = r.organizationRepo.Get(ctx, id); err != nil {
		return nil, err
	}

	if err = r.cacheRepo.Set(ctx, key, organization); err != nil {
		return nil, err
	}

	return organization, nil
}

func (r *RedisCachedOrganizationRepository) GetAll(ctx context.Context, userID model.ID, offset, limit int) ([]*model.Organization, error) {
	var organizations []*model.Organization
	var err error

	key := composeCacheKey(model.ResourceTypeOrganization.String(), "GetAll", userID.String(), offset, limit)
	if err = r.cacheRepo.Get(ctx, key, &organizations); err != nil {
		return nil, err
	}

	if organizations != nil {
		return organizations, nil
	}

	if organizations, err = r.organizationRepo.GetAll(ctx, userID, offset, limit); err != nil {
		return nil, err
	}

	if err = r.cacheRepo.Set(ctx, key, organizations); err != nil {
		return nil, err
	}

	return organizations, nil
}

func (r *RedisCachedOrganizationRepository) Update(ctx context.Context, id model.ID, patch map[string]any) (*model.Organization, error) {
	var organization *model.Organization
	var err error

	organization, err = r.organizationRepo.Update(ctx, id, patch)
	if err != nil {
		return nil, err
	}

	key := composeCacheKey(model.ResourceTypeOrganization.String(), id.String())
	if err = r.cacheRepo.Set(ctx, key, organization); err != nil {
		return nil, err
	}

	if err := clearOrganizationAllGetAll(ctx, r.cacheRepo); err != nil {
		return nil, err
	}

	return organization, nil
}

func (r *RedisCachedOrganizationRepository) AddMember(ctx context.Context, orgID, memberID model.ID) error {
	if err := clearOrganizationsKey(ctx, r.cacheRepo, orgID); err != nil {
		return err
	}

	if err := clearOrganizationAllGetAll(ctx, r.cacheRepo); err != nil {
		return err
	}

	return r.organizationRepo.AddMember(ctx, orgID, memberID)
}

func (r *RedisCachedOrganizationRepository) RemoveMember(ctx context.Context, orgID, memberID model.ID) error {
	if err := clearOrganizationsKey(ctx, r.cacheRepo, orgID); err != nil {
		return err
	}

	if err := clearOrganizationAllGetAll(ctx, r.cacheRepo); err != nil {
		return err
	}

	return r.organizationRepo.RemoveMember(ctx, orgID, memberID)
}

func (r *RedisCachedOrganizationRepository) Delete(ctx context.Context, id model.ID) error {
	if err := clearOrganizationsKey(ctx, r.cacheRepo, id); err != nil {
		return err
	}

	if err := clearOrganizationAllGetAll(ctx, r.cacheRepo); err != nil {
		return err
	}

	return r.organizationRepo.Delete(ctx, id)
}

func (r *RedisCachedOrganizationRepository) GetMembers(ctx context.Context, orgID model.ID) ([]*model.OrganizationMember, error) {
	return r.organizationRepo.GetMembers(ctx, orgID)
}

func (r *RedisCachedOrganizationRepository) AddInvitation(ctx context.Context, orgID, userID model.ID) error {
	if err := clearOrganizationsKey(ctx, r.cacheRepo, orgID); err != nil {
		return err
	}

	if err := clearOrganizationAllGetAll(ctx, r.cacheRepo); err != nil {
		return err
	}

	return r.organizationRepo.AddInvitation(ctx, orgID, userID)
}

func (r *RedisCachedOrganizationRepository) RemoveInvitation(ctx context.Context, orgID, userID model.ID) error {
	if err := clearOrganizationsKey(ctx, r.cacheRepo, orgID); err != nil {
		return err
	}

	if err := clearOrganizationAllGetAll(ctx, r.cacheRepo); err != nil {
		return err
	}

	return r.organizationRepo.RemoveInvitation(ctx, orgID, userID)
}

func (r *RedisCachedOrganizationRepository) GetInvitations(ctx context.Context, orgID model.ID) ([]*model.OrganizationMember, error) {
	return r.organizationRepo.GetInvitations(ctx, orgID)
}

// NewCachedOrganizationRepository returns a new CachedOrganizationRepository.
func NewCachedOrganizationRepository(repo OrganizationRepository, opts ...RedisRepositoryOption) (*RedisCachedOrganizationRepository, error) {
	r, err := newRedisBaseRepository(opts...)
	if err != nil {
		return nil, err
	}

	return &RedisCachedOrganizationRepository{
		cacheRepo:        r,
		organizationRepo: repo,
	}, nil
}

package neo4j

import (
	"context"
	"errors"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/repository"
)

// OrganizationRepository is a repository for managing organizations.
type OrganizationRepository struct {
	*baseRepository
}

func (r *OrganizationRepository) scan(op, np, tp, mp string) func(rec *neo4j.Record) (*model.Organization, error) {
	return func(rec *neo4j.Record) (*model.Organization, error) {
		org := new(model.Organization)

		val, _, err := neo4j.GetRecordValue[neo4j.Node](rec, op)
		if err != nil {
			return nil, err
		}

		if err := ScanIntoStruct(&val, &org, []string{"id"}); err != nil {
			return nil, err
		}

		org.ID, _ = model.NewIDFromString(val.GetProperties()["id"].(string), model.ResourceTypeOrganization.String())

		if org.Namespaces, err = ParseIDsFromRecord(rec, np, model.ResourceTypeNamespace.String()); err != nil {
			return nil, err
		}

		if org.Teams, err = ParseIDsFromRecord(rec, tp, model.ResourceTypeRole.String()); err != nil {
			return nil, err
		}

		if org.Members, err = ParseIDsFromRecord(rec, mp, model.ResourceTypeUser.String()); err != nil {
			return nil, err
		}

		if err := org.Validate(); err != nil {
			return nil, err
		}

		return org, nil
	}
}

func (r *OrganizationRepository) scanOrganizationMember(up string) func(rec *neo4j.Record) (model.OrganizationMember, error) {
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

		roleNamesVal, err := ParseValueFromRecord[[]any](rec, "roles")
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

func (r *OrganizationRepository) Create(ctx context.Context, owner model.ID, organization *model.Organization) error {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.OrganizationRepository/Create")
	defer span.End()

	if err := owner.Validate(); err != nil {
		return errors.Join(repository.ErrOrganizationCreate, err)
	}

	if err := organization.Validate(); err != nil {
		return errors.Join(repository.ErrOrganizationCreate, err)
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

	if err := ExecuteWriteAndConsume(ctx, r.db, cypher, params); err != nil {
		return errors.Join(repository.ErrOrganizationCreate, err)
	}

	return nil
}

func (r *OrganizationRepository) Get(ctx context.Context, id model.ID) (*model.Organization, error) {
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

	org, err := ExecuteReadAndReadSingle(ctx, r.db, cypher, params, r.scan("o", "n", "t", "m"))
	if err != nil {
		return nil, errors.Join(repository.ErrOrganizationRead, err)
	}

	return org, nil
}

func (r *OrganizationRepository) GetAll(ctx context.Context, userID model.ID, offset, limit int) ([]*model.Organization, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.OrganizationRepository/GetAllBelongsTo")
	defer span.End()

	if err := userID.Validate(); err != nil {
		return nil, errors.Join(repository.ErrOrganizationRead, err)
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

	orgs, err := ExecuteReadAndReadAll(ctx, r.db, cypher, params, r.scan("o", "n", "t", "m"))
	if err != nil {
		return nil, errors.Join(repository.ErrOrganizationRead, err)
	}

	return orgs, nil
}

func (r *OrganizationRepository) Update(ctx context.Context, id model.ID, patch map[string]any) (*model.Organization, error) {
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

	org, err := ExecuteWriteAndReadSingle(ctx, r.db, cypher, params, r.scan("o", "n", "t", "m"))
	if err != nil {
		return nil, errors.Join(repository.ErrOrganizationUpdate, err)
	}

	return org, nil
}

func (r *OrganizationRepository) GetMembers(ctx context.Context, orgID model.ID) ([]*model.OrganizationMember, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.OrganizationRepository/GetMembers")
	defer span.End()

	if err := orgID.Validate(); err != nil {
		return nil, errors.Join(repository.ErrOrganizationRead, err)
	}

	// Query both MEMBER_OF (active members) and INVITED_TO (pending invitations)
	// Single query using UNION to combine active members with pending invitations
	cypher := `
	MATCH (o:` + orgID.Label() + ` {id: $org_id})
	MATCH (u:` + model.ResourceTypeUser.String() + `)-[:` + EdgeKindMemberOf.String() + `]->(o)
	OPTIONAL MATCH (u)-[:` + EdgeKindMemberOf.String() + `]->(r:` + model.ResourceTypeRole.String() + `)<-[:` + EdgeKindHasTeam.String() + `]-(o)
	WITH u, [r IN collect(DISTINCT r) WHERE r IS NOT NULL | r.name] AS roles, true AS isMember
	RETURN u AS u, roles AS roles, isMember AS isMember
	UNION
	MATCH (o:` + orgID.Label() + ` {id: $org_id})
	MATCH (u:` + model.ResourceTypeUser.String() + `)-[:` + EdgeKindInvitedTo.String() + `]->(o)
	WHERE NOT EXISTS((u)-[:` + EdgeKindMemberOf.String() + `]->(o))
	RETURN u AS u, [] AS roles, false AS isMember
	ORDER BY isMember DESC, u.created_at ASC`

	params := map[string]any{
		"org_id": orgID.String(),
	}

	members, err := ExecuteReadAndReadAll(ctx, r.db, cypher, params, func(rec *neo4j.Record) (model.OrganizationMember, error) {
		member, err := r.scanOrganizationMember("u")(rec)
		if err != nil {
			return model.OrganizationMember{}, err
		}

		isMemberVal, err := ParseValueFromRecord[bool](rec, "isMember")
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
		return nil, errors.Join(repository.ErrOrganizationRead, err)
	}

	membersPtr := make([]*model.OrganizationMember, len(members))
	for i := range members {
		membersPtr[i] = &members[i]
	}

	return membersPtr, nil
}

func (r *OrganizationRepository) AddMember(ctx context.Context, orgID, memberID model.ID) error {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.OrganizationRepository/AddMember")
	defer span.End()

	if err := orgID.Validate(); err != nil {
		return errors.Join(repository.ErrOrganizationAddMember, err)
	}

	if err := memberID.Validate(); err != nil {
		return errors.Join(repository.ErrOrganizationAddMember, err)
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

	if err := ExecuteWriteAndConsume(ctx, r.db, cypher, params); err != nil {
		return errors.Join(repository.ErrOrganizationAddMember, err)
	}

	return nil
}

func (r *OrganizationRepository) RemoveMember(ctx context.Context, orgID, memberID model.ID) error {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.OrganizationRepository/RemoveMember")
	defer span.End()

	if err := orgID.Validate(); err != nil {
		return errors.Join(repository.ErrOrganizationRemoveMember, err)
	}

	if err := memberID.Validate(); err != nil {
		return errors.Join(repository.ErrOrganizationRemoveMember, err)
	}

	cypher := `
	MATCH (:` + memberID.Label() + ` {id: $member_id})-[r:` + EdgeKindMemberOf.String() + `]->(:` + orgID.Label() + ` {id: $org_id})
	DELETE r`

	params := map[string]any{
		"org_id":    orgID.String(),
		"member_id": memberID.String(),
	}

	if err := ExecuteWriteAndConsume(ctx, r.db, cypher, params); err != nil {
		return errors.Join(repository.ErrOrganizationRemoveMember, err)
	}

	return nil
}

func (r *OrganizationRepository) AddInvitation(ctx context.Context, orgID, userID model.ID) error {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.OrganizationRepository/AddInvitation")
	defer span.End()

	if err := orgID.Validate(); err != nil {
		return errors.Join(repository.ErrOrganizationAddMember, err)
	}

	if err := userID.Validate(); err != nil {
		return errors.Join(repository.ErrOrganizationAddMember, err)
	}

	// Check if IDs are nil (invalid)
	if orgID.IsNil() || userID.IsNil() {
		return errors.Join(repository.ErrOrganizationAddMember, model.ErrInvalidID)
	}

	cypher := `
	MATCH (o:` + orgID.Label() + ` {id: $org_id})
	MATCH (u:` + userID.Label() + ` {id: $user_id})
	WITH o, u
	WHERE o IS NOT NULL AND u IS NOT NULL
	MERGE (u)-[i:` + EdgeKindInvitedTo.String() + `]->(o)
	ON CREATE SET i.created_at = datetime($now), i.id = $invitation_id
	ON MATCH SET i.updated_at = datetime($now)
	RETURN i`

	params := map[string]any{
		"org_id":        orgID.String(),
		"user_id":       userID.String(),
		"invitation_id": model.NewRawID(),
		"now":           time.Now().UTC().Format(time.RFC3339Nano),
	}

	result, err := r.db.GetWriteSession(ctx).Run(ctx, cypher, params)
	if err != nil {
		return errors.Join(repository.ErrOrganizationAddMember, err)
	}

	// Check if any records were returned (nodes were found)
	hasRecord := false
	for result.Next(ctx) {
		hasRecord = true
		break
	}

	if err := result.Err(); err != nil {
		return errors.Join(repository.ErrOrganizationAddMember, err)
	}

	if !hasRecord {
		return errors.Join(repository.ErrOrganizationAddMember, repository.ErrNotFound)
	}

	return nil
}

func (r *OrganizationRepository) RemoveInvitation(ctx context.Context, orgID, userID model.ID) error {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.OrganizationRepository/RemoveInvitation")
	defer span.End()

	if err := orgID.Validate(); err != nil {
		return errors.Join(repository.ErrOrganizationRemoveMember, err)
	}

	if err := userID.Validate(); err != nil {
		return errors.Join(repository.ErrOrganizationRemoveMember, err)
	}

	// Check if IDs are nil (invalid)
	if orgID.IsNil() || userID.IsNil() {
		return errors.Join(repository.ErrOrganizationRemoveMember, model.ErrInvalidID)
	}

	cypher := `
	MATCH (:` + userID.Label() + ` {id: $user_id})-[r:` + EdgeKindInvitedTo.String() + `]->(:` + orgID.Label() + ` {id: $org_id})
	DELETE r`

	params := map[string]any{
		"org_id":  orgID.String(),
		"user_id": userID.String(),
	}

	if err := ExecuteWriteAndConsume(ctx, r.db, cypher, params); err != nil {
		return errors.Join(repository.ErrOrganizationRemoveMember, err)
	}

	return nil
}

func (r *OrganizationRepository) GetInvitations(ctx context.Context, orgID model.ID) ([]*model.OrganizationMember, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.OrganizationRepository/GetInvitations")
	defer span.End()

	if err := orgID.Validate(); err != nil {
		return nil, errors.Join(repository.ErrOrganizationRead, err)
	}

	// Check if ID is nil (invalid)
	if orgID.IsNil() {
		return nil, errors.Join(repository.ErrOrganizationRead, model.ErrInvalidID)
	}

	cypher := `
	MATCH (o:` + orgID.Label() + ` {id: $org_id})
	MATCH (u:` + model.ResourceTypeUser.String() + `)-[:` + EdgeKindInvitedTo.String() + `]->(o)
	RETURN u, [] AS roles
	ORDER BY u.created_at ASC`

	params := map[string]any{
		"org_id": orgID.String(),
	}

	members, err := ExecuteReadAndReadAll(ctx, r.db, cypher, params, r.scanOrganizationMember("u"))
	if err != nil {
		return nil, errors.Join(repository.ErrOrganizationRead, err)
	}

	membersPtr := make([]*model.OrganizationMember, len(members))
	for i := range members {
		membersPtr[i] = &members[i]
	}

	return membersPtr, nil
}

func (r *OrganizationRepository) Delete(ctx context.Context, id model.ID) error {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.OrganizationRepository/Delete")
	defer span.End()

	cypher := `MATCH (o:` + id.Label() + ` {id: $id}), (o)-[r]-() DETACH DELETE o, r`
	params := map[string]any{
		"id": id.String(),
	}

	if err := ExecuteWriteAndConsume(ctx, r.db, cypher, params); err != nil {
		return errors.Join(repository.ErrOrganizationDelete, err)
	}

	return nil
}

// NewOrganizationRepository creates a new organization baseRepository.
func NewOrganizationRepository(opts ...RepositoryOption) (*OrganizationRepository, error) {
	baseRepo, err := newRepository(opts...)
	if err != nil {
		return nil, err
	}

	return &OrganizationRepository{
		baseRepository: baseRepo,
	}, nil
}

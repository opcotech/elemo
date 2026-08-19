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
	ErrOrganizationAddMember    = errors.New("failed to add member to organization")      // member cannot be added to organization
	ErrOrganizationCreate       = errors.New("failed to create organization")             // organization cannot be created
	ErrOrganizationDelete       = errors.New("failed to delete organization")             // organization cannot be deleted
	ErrOrganizationRead         = errors.New("failed to read organization")               // organization cannot be read
	ErrOrganizationRemoveMember = errors.New("failed to remove member from organization") // member cannot be removed from organization
	ErrOrganizationUpdate       = errors.New("failed to update organization")             // organization cannot be updated
)

// Organization represents an organization persisted by the repository.
type Organization struct {
	ID             model.ID                 `json:"id"`
	Name           string                   `json:"name"`
	Email          string                   `json:"email"`
	Logo           string                   `json:"logo"`
	Website        string                   `json:"website"`
	Status         model.OrganizationStatus `json:"status"`
	NamespaceCount *int64                   `json:"namespace_count"`
	TeamCount      *int64                   `json:"team_count"`
	MemberCount    *int64                   `json:"member_count"`
	DocumentCount  *int64                   `json:"document_count"`
	CreatedAt      *time.Time               `json:"created_at"`
	UpdatedAt      *time.Time               `json:"updated_at"`
}

// OrganizationMember represents a member of an organization.
type OrganizationMember struct {
	ID        model.ID         `json:"id"`
	FirstName string           `json:"first_name"`
	LastName  string           `json:"last_name"`
	Email     string           `json:"email"`
	Picture   *string          `json:"picture"`
	Status    model.UserStatus `json:"status"`
	Roles     []string         `json:"roles"`
	CreatedAt *time.Time       `json:"created_at"`
}

// CreateOrganizationOpts holds the data required to create an organization.
type CreateOrganizationOpts struct {
	Owner   model.ID
	Name    string
	Email   string
	Logo    string
	Website string
	Status  model.OrganizationStatus
}

// UpdateOrganizationOpts holds the fields that can be updated on an organization.
// Undefined fields (Defined == false) are left unchanged.
type UpdateOrganizationOpts struct {
	Name    optional.Optional[string]
	Email   optional.Optional[string]
	Logo    optional.Optional[string]
	Website optional.Optional[string]
	Status  optional.Optional[model.OrganizationStatus]
}

// patch builds a Neo4j property map from defined optional fields.
func (o UpdateOrganizationOpts) patch() map[string]any {
	p := make(map[string]any)

	if o.Name.Defined {
		p["name"] = *o.Name.Value
	}
	if o.Email.Defined {
		p["email"] = *o.Email.Value
	}
	if o.Logo.Defined {
		if o.Logo.Value == nil {
			p["logo"] = nil
		} else {
			p["logo"] = *o.Logo.Value
		}
	}
	if o.Website.Defined {
		if o.Website.Value == nil {
			p["website"] = nil
		} else {
			p["website"] = *o.Website.Value
		}
	}
	if o.Status.Defined {
		p["status"] = o.Status.Value.String()
	}

	return p
}

//go:generate go tool mockgen -source=organization.go -destination=organization_mock_gen.go -package=repository -mock_names "OrganizationRepository=MockOrganizationRepository"
type OrganizationRepository interface {
	Create(ctx context.Context, opts CreateOrganizationOpts) (*Organization, error)
	Get(ctx context.Context, id model.ID, proj OrganizationProjection) (*Organization, error)
	List(ctx context.Context, userID model.ID, page CursorPage, proj OrganizationProjection) (Page[*Organization], error)
	Update(ctx context.Context, id model.ID, opts UpdateOrganizationOpts) (*Organization, error)
	ListMembers(ctx context.Context, orgID model.ID, page CursorPage) (Page[*OrganizationMember], error)
	AddMember(ctx context.Context, orgID, memberID model.ID) error
	RemoveMember(ctx context.Context, orgID, memberID model.ID) error
	AddInvitation(ctx context.Context, orgID, userID model.ID) error
	RemoveInvitation(ctx context.Context, orgID, userID model.ID) error
	GetInvitations(ctx context.Context, orgID model.ID) ([]*OrganizationMember, error)
	Delete(ctx context.Context, id model.ID) error
}

// Neo4jOrganizationRepository is a repository for managing organizations.
type Neo4jOrganizationRepository struct {
	*neo4jBaseRepository
}

func (r *Neo4jOrganizationRepository) scan(proj OrganizationProjection) func(rec *neo4j.Record) (*Organization, error) {
	return func(rec *neo4j.Record) (*Organization, error) {
		node, err := Neo4jRecordNode(rec, "o")
		if err != nil {
			return nil, err
		}

		org := new(Organization)
		if err := Neo4jScanIntoStruct(&node, &org, []string{"id", "namespace_count", "team_count", "member_count", "document_count"}); err != nil {
			return nil, err
		}

		org.ID, err = Neo4jDecodeID(node, model.ResourceTypeOrganization)
		if err != nil {
			return nil, err
		}
		if proj.NamespaceCount {
			namespaceCount, err := Neo4jParseValueFromRecord[int64](rec, "namespace_count")
			if err != nil {
				return nil, err
			}
			org.NamespaceCount = convert.ToPointer(namespaceCount)
		}
		if proj.TeamCount {
			teamCount, err := Neo4jParseValueFromRecord[int64](rec, "team_count")
			if err != nil {
				return nil, err
			}
			org.TeamCount = convert.ToPointer(teamCount)
		}
		if proj.MemberCount {
			memberCount, err := Neo4jParseValueFromRecord[int64](rec, "member_count")
			if err != nil {
				return nil, err
			}
			org.MemberCount = convert.ToPointer(memberCount)
		}
		if proj.DocumentCount {
			documentCount, err := Neo4jParseValueFromRecord[int64](rec, "document_count")
			if err != nil {
				return nil, err
			}
			org.DocumentCount = convert.ToPointer(documentCount)
		}

		return org, nil
	}
}

func (r *Neo4jOrganizationRepository) scanOrganizationMember(up string) func(rec *neo4j.Record) (OrganizationMember, error) {
	return func(rec *neo4j.Record) (OrganizationMember, error) {
		val, _, err := neo4j.GetRecordValue[neo4j.Node](rec, up)
		if err != nil {
			return OrganizationMember{}, err
		}

		userID, err := model.NewIDFromString(val.GetProperties()["id"].(string), model.ResourceTypeUser.String())
		if err != nil {
			return OrganizationMember{}, err
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
			return OrganizationMember{}, err
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

		createdAt, err := Neo4jNodeTime(val, "created_at")
		if err != nil {
			return OrganizationMember{}, err
		}

		return OrganizationMember{
			ID:        userID,
			FirstName: firstName,
			LastName:  lastName,
			Email:     email,
			Picture:   picture,
			Status:    status,
			Roles:     roleNames,
			CreatedAt: createdAt,
		}, nil
	}
}

func (r *Neo4jOrganizationRepository) Create(ctx context.Context, opts CreateOrganizationOpts) (*Organization, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.OrganizationRepository/Create")
	defer span.End()

	createdAt := time.Now().UTC()
	id := model.MustNewID(model.ResourceTypeOrganization)

	status := opts.Status
	if status == 0 {
		status = model.OrganizationStatusActive
	}

	cypher := `
	MATCH (u:` + opts.Owner.Label() + ` {id: $owner_id})
	SET u:` + model.LabelPrincipal + `
	CREATE (o:` + id.Label() + `:` + model.LabelPrincipal + ` { id: $id, name: $name, email: $email, logo: $logo, website: $website,
		status: $status, created_at: datetime($created_at)
	}),
	(u)-[:` + EdgeKindMemberOf.String() + ` {id: $membership_id, created_at: datetime($created_at)}]->(o)`

	params := map[string]any{
		"id":            id.String(),
		"name":          opts.Name,
		"email":         opts.Email,
		"logo":          opts.Logo,
		"website":       opts.Website,
		"status":        status.String(),
		"created_at":    createdAt.Format(time.RFC3339Nano),
		"owner_id":      opts.Owner.String(),
		"membership_id": model.NewRawID(),
	}

	if err := Neo4jExecuteWriteAndConsume(ctx, r.db, cypher, params); err != nil {
		return nil, errors.Join(ErrOrganizationCreate, err)
	}

	return r.Get(ctx, id, OrganizationDetailProjection())
}

func (r *Neo4jOrganizationRepository) Get(ctx context.Context, id model.ID, proj OrganizationProjection) (*Organization, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.OrganizationRepository/Get")
	defer span.End()

	plan, err := CompileQuery(OrganizationGetQuery{ID: id, Projection: proj})
	if err != nil {
		return nil, errors.Join(ErrOrganizationRead, err)
	}

	var organization *Organization
	err = Neo4jExecuteReadPlan(ctx, r.db, plan, func(tx neo4j.ManagedTransaction) error {
		var readErr error
		organization, _, readErr = Neo4jRunQuerySingle(ctx, tx, plan.Root, r.scan(proj))
		if readErr != nil {
			return readErr
		}
		return nil
	})
	if err != nil {
		return nil, errors.Join(ErrOrganizationRead, err)
	}

	return organization, nil
}

func (r *Neo4jOrganizationRepository) List(ctx context.Context, userID model.ID, page CursorPage, proj OrganizationProjection) (Page[*Organization], error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.OrganizationRepository/List")
	defer span.End()

	normalized, err := page.Normalize()
	if err != nil {
		return Page[*Organization]{}, errors.Join(ErrOrganizationRead, err)
	}
	plan, err := CompileQuery(OrganizationListQuery{
		UserID:     userID,
		Action:     model.ActionOrganizationRead,
		Page:       normalized,
		Order:      SortDirectionDesc,
		Projection: proj,
	})
	if err != nil {
		return Page[*Organization]{}, errors.Join(ErrOrganizationRead, err)
	}

	organizations := make([]*Organization, 0)
	err = Neo4jExecuteReadPlan(ctx, r.db, plan, func(tx neo4j.ManagedTransaction) error {
		var readErr error
		organizations, _, readErr = Neo4jRunQuery(ctx, tx, plan.Root, r.scan(proj))
		if readErr != nil {
			return readErr
		}
		return nil
	})
	if err != nil {
		return Page[*Organization]{}, errors.Join(ErrOrganizationRead, err)
	}

	return PaginateSlice(organizations, normalized.Size, func(organization *Organization) model.ID {
		return organization.ID
	})
}

func (r *Neo4jOrganizationRepository) Update(ctx context.Context, id model.ID, opts UpdateOrganizationOpts) (*Organization, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.OrganizationRepository/Update")
	defer span.End()

	cypher := `
	MATCH (o:` + id.Label() + ` {id: $id})
	SET o += $patch, o.updated_at = datetime()
	RETURN o.id AS id`

	params := map[string]any{
		"id":    id.String(),
		"patch": opts.patch(),
	}

	_, err := Neo4jExecuteWriteAndReadSingle(ctx, r.db, cypher, params, func(_ *neo4j.Record) (*struct{}, error) {
		return &struct{}{}, nil
	})
	if err != nil {
		return nil, errors.Join(ErrOrganizationUpdate, err)
	}

	return r.Get(ctx, id, OrganizationDetailProjection())
}

func (r *Neo4jOrganizationRepository) ListMembers(ctx context.Context, orgID model.ID, page CursorPage) (Page[*OrganizationMember], error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.OrganizationRepository/ListMembers")
	defer span.End()

	normalized, err := page.Normalize()
	if err != nil {
		return Page[*OrganizationMember]{}, errors.Join(ErrOrganizationRead, err)
	}
	plan, err := CompileQuery(OrganizationMemberListQuery{
		OrgID: orgID,
		Page:  normalized,
		Order: SortDirectionDesc,
	})
	if err != nil {
		return Page[*OrganizationMember]{}, errors.Join(ErrOrganizationRead, err)
	}

	members := make([]*OrganizationMember, 0)
	err = Neo4jExecuteReadPlan(ctx, r.db, plan, func(tx neo4j.ManagedTransaction) error {
		rootMembers, _, readErr := Neo4jRunQuery(ctx, tx, plan.Root, func(rec *neo4j.Record) (*OrganizationMember, error) {
			member, err := r.scanOrganizationMember("u")(rec)
			if err != nil {
				return nil, err
			}
			isMemberVal, err := Neo4jParseValueFromRecord[bool](rec, "isMember")
			if err != nil {
				isMemberVal = false
			}
			if !isMemberVal {
				member.Status = model.UserStatusPending
			}
			return &member, nil
		})
		if readErr != nil {
			return readErr
		}
		members = rootMembers
		return r.applyMemberRoleLoader(ctx, tx, plan, members)
	})
	if err != nil {
		return Page[*OrganizationMember]{}, errors.Join(ErrOrganizationRead, err)
	}

	return PaginateSlice(members, normalized.Size, func(member *OrganizationMember) model.ID {
		return member.ID
	})
}

func (r *Neo4jOrganizationRepository) applyMemberRoleLoader(
	ctx context.Context,
	tx neo4j.ManagedTransaction,
	plan QueryPlan,
	members []*OrganizationMember,
) error {
	if len(plan.Loaders) == 0 || len(members) == 0 {
		return nil
	}

	ids := make([]string, 0, len(members))
	membersByID := make(map[string]*OrganizationMember, len(members))
	for _, member := range members {
		if member == nil {
			continue
		}
		id := member.ID.String()
		ids = append(ids, id)
		membersByID[id] = member
		if member.Roles == nil {
			member.Roles = make([]string, 0)
		}
	}

	for _, loader := range plan.Loaders {
		query := loader
		query.Params = cloneParams(loader.Params)
		query.Params["ids"] = ids

		rows, _, err := Neo4jRunQuery(ctx, tx, query, func(rec *neo4j.Record) (struct {
			UserID string
			Roles  []string
		}, error) {
			userID, err := Neo4jParseValueFromRecord[string](rec, "user_id")
			if err != nil {
				return struct {
					UserID string
					Roles  []string
				}{}, err
			}
			roleNamesVal, err := Neo4jParseValueFromRecord[[]any](rec, "roles")
			if err != nil {
				roleNamesVal = []any{}
			}
			roles := make([]string, 0, len(roleNamesVal))
			for _, roleName := range roleNamesVal {
				if roleName == nil {
					continue
				}
				name, ok := roleName.(string)
				if !ok || name == "" {
					continue
				}
				roles = append(roles, name)
			}
			return struct {
				UserID string
				Roles  []string
			}{UserID: userID, Roles: roles}, nil
		})
		if err != nil {
			return err
		}
		for _, row := range rows {
			member, ok := membersByID[row.UserID]
			if !ok {
				continue
			}
			member.Roles = row.Roles
		}
	}

	return nil
}

func (r *Neo4jOrganizationRepository) AddMember(ctx context.Context, orgID, memberID model.ID) error {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.OrganizationRepository/AddMember")
	defer span.End()

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

func (r *Neo4jOrganizationRepository) GetInvitations(ctx context.Context, orgID model.ID) ([]*OrganizationMember, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.OrganizationRepository/GetInvitations")
	defer span.End()

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

	membersPtr := make([]*OrganizationMember, len(members))
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
	return clearOrganizationsPattern(ctx, r, "Get", id.String(), "*")
}

func clearOrganizationAllLists(ctx context.Context, r *redisBaseRepository) error {
	return clearOrganizationsPattern(ctx, r, "List", "*", "*", "*", "*")
}

// RedisCachedOrganizationRepository implements caching on the
// repository.OrganizationRepository.
type RedisCachedOrganizationRepository struct {
	cacheRepo        *redisBaseRepository
	organizationRepo OrganizationRepository
}

func (r *RedisCachedOrganizationRepository) Create(ctx context.Context, opts CreateOrganizationOpts) (*Organization, error) {
	if err := clearOrganizationAllLists(ctx, r.cacheRepo); err != nil {
		return nil, err
	}

	return r.organizationRepo.Create(ctx, opts)
}

func (r *RedisCachedOrganizationRepository) Get(ctx context.Context, id model.ID, proj OrganizationProjection) (*Organization, error) {
	var organization *Organization
	var err error

	key := composeCacheKey(model.ResourceTypeOrganization.String(), "Get", id.String(), projectionCacheValue(proj))
	if err = r.cacheRepo.Get(ctx, key, &organization); err != nil {
		return nil, err
	}

	if organization != nil {
		return organization, nil
	}

	if organization, err = r.organizationRepo.Get(ctx, id, proj); err != nil {
		return nil, err
	}

	if err = r.cacheRepo.Set(ctx, key, organization); err != nil {
		return nil, err
	}

	return organization, nil
}

func (r *RedisCachedOrganizationRepository) List(ctx context.Context, userID model.ID, page CursorPage, proj OrganizationProjection) (Page[*Organization], error) {
	var organizations Page[*Organization]
	var err error

	normalized, err := normalizedPage(page)
	if err != nil {
		return Page[*Organization]{}, err
	}

	key := composeCacheKey(model.ResourceTypeOrganization.String(), "List", userID.String(), projectionCacheValue(proj), pageTokenValue(normalized.Token), normalized.Size)
	if err = r.cacheRepo.Get(ctx, key, &organizations); err != nil {
		return Page[*Organization]{}, err
	}

	if organizations.Items != nil {
		return organizations, nil
	}

	if organizations, err = r.organizationRepo.List(ctx, userID, normalized, proj); err != nil {
		return Page[*Organization]{}, err
	}

	if err = r.cacheRepo.Set(ctx, key, organizations); err != nil {
		return Page[*Organization]{}, err
	}

	return organizations, nil
}

func (r *RedisCachedOrganizationRepository) Update(ctx context.Context, id model.ID, opts UpdateOrganizationOpts) (*Organization, error) {
	organization, err := r.organizationRepo.Update(ctx, id, opts)
	if err != nil {
		return nil, err
	}

	key := composeCacheKey(model.ResourceTypeOrganization.String(), "Get", id.String(), projectionCacheValue(OrganizationDetailProjection()))
	if err = r.cacheRepo.Set(ctx, key, organization); err != nil {
		return nil, err
	}

	if err := clearOrganizationAllLists(ctx, r.cacheRepo); err != nil {
		return nil, err
	}

	return organization, nil
}

func (r *RedisCachedOrganizationRepository) AddMember(ctx context.Context, orgID, memberID model.ID) error {
	if err := clearOrganizationsKey(ctx, r.cacheRepo, orgID); err != nil {
		return err
	}

	if err := clearOrganizationAllLists(ctx, r.cacheRepo); err != nil {
		return err
	}

	return r.organizationRepo.AddMember(ctx, orgID, memberID)
}

func (r *RedisCachedOrganizationRepository) RemoveMember(ctx context.Context, orgID, memberID model.ID) error {
	if err := clearOrganizationsKey(ctx, r.cacheRepo, orgID); err != nil {
		return err
	}

	if err := clearOrganizationAllLists(ctx, r.cacheRepo); err != nil {
		return err
	}

	return r.organizationRepo.RemoveMember(ctx, orgID, memberID)
}

func (r *RedisCachedOrganizationRepository) Delete(ctx context.Context, id model.ID) error {
	if err := clearOrganizationsKey(ctx, r.cacheRepo, id); err != nil {
		return err
	}

	if err := clearOrganizationAllLists(ctx, r.cacheRepo); err != nil {
		return err
	}

	return r.organizationRepo.Delete(ctx, id)
}

func (r *RedisCachedOrganizationRepository) ListMembers(ctx context.Context, orgID model.ID, page CursorPage) (Page[*OrganizationMember], error) {
	return r.organizationRepo.ListMembers(ctx, orgID, page)
}

func (r *RedisCachedOrganizationRepository) AddInvitation(ctx context.Context, orgID, userID model.ID) error {
	if err := clearOrganizationsKey(ctx, r.cacheRepo, orgID); err != nil {
		return err
	}

	if err := clearOrganizationAllLists(ctx, r.cacheRepo); err != nil {
		return err
	}

	return r.organizationRepo.AddInvitation(ctx, orgID, userID)
}

func (r *RedisCachedOrganizationRepository) RemoveInvitation(ctx context.Context, orgID, userID model.ID) error {
	if err := clearOrganizationsKey(ctx, r.cacheRepo, orgID); err != nil {
		return err
	}

	if err := clearOrganizationAllLists(ctx, r.cacheRepo); err != nil {
		return err
	}

	return r.organizationRepo.RemoveInvitation(ctx, orgID, userID)
}

func (r *RedisCachedOrganizationRepository) GetInvitations(ctx context.Context, orgID model.ID) ([]*OrganizationMember, error) {
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

package neo4j

import (
	"context"
	"errors"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"

	"github.com/opcotech/elemo/internal/model"
)

var (
	ErrOrganizationCreate       = errors.New("failed to create organization")             // organization cannot be created
	ErrOrganizationRead         = errors.New("failed to read organization")               // organization cannot be read
	ErrOrganizationUpdate       = errors.New("failed to update organization")             // organization cannot be updated
	ErrOrganizationAddMember    = errors.New("failed to add member to organization")      // member cannot be added to organization
	ErrOrganizationRemoveMember = errors.New("failed to remove member from organization") // member cannot be removed from organization
	ErrOrganizationDelete       = errors.New("failed to delete organization")             // organization cannot be deleted
)

// OrganizationRepository is a repository for managing organizations.
type OrganizationRepository struct {
	*repository
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

		org.ID, _ = model.NewIDFromString(val.GetProperties()["id"].(string), model.OrganizationIDType)

		if org.Namespaces, err = ParseIDsFromRecord(rec, np, model.NamespaceIDType); err != nil {
			return nil, err
		}

		if org.Teams, err = ParseIDsFromRecord(rec, tp, model.RoleIDType); err != nil {
			return nil, err
		}

		if org.Members, err = ParseIDsFromRecord(rec, mp, model.UserIDType); err != nil {
			return nil, err
		}

		if err := org.Validate(); err != nil {
			return nil, err
		}

		return org, nil
	}
}

func (r *OrganizationRepository) Create(ctx context.Context, owner model.ID, organization *model.Organization) error {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.OrganizationRepository/Create")
	defer span.End()

	if err := organization.Validate(); err != nil {
		return errors.Join(ErrOrganizationCreate, err)
	}

	createdAt := time.Now()

	membershipID := model.MustNewID(EdgeKindMemberOf.String())

	organization.ID = model.MustNewID(model.OrganizationIDType)
	organization.CreatedAt = &createdAt
	organization.UpdatedAt = nil

	cypher := `
	MATCH (u:` + owner.Label() + ` {id: $owner_id})
	CREATE (o:` + organization.ID.Label() + ` { id: $id, name: $name, email: $email, logo: $logo, website: $website,
		status: $status, created_at: datetime($created_at)
	}),
	(u)-[:` + membershipID.Label() + ` {id: $membership_id, created_at: datetime($created_at)}]->(o)`

	params := map[string]interface{}{
		"id":            organization.ID.String(),
		"name":          organization.Name,
		"email":         organization.Email,
		"logo":          organization.Logo,
		"website":       organization.Website,
		"status":        organization.Status.String(),
		"created_at":    createdAt.Format(time.RFC3339Nano),
		"owner_id":      owner.String(),
		"membership_id": membershipID.String(),
	}

	if err := ExecuteWriteAndConsume(ctx, r.db, cypher, params); err != nil {
		return errors.Join(ErrOrganizationCreate, err)
	}

	return nil
}

func (r *OrganizationRepository) Get(ctx context.Context, id model.ID) (*model.Organization, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.OrganizationRepository/Get")
	defer span.End()

	cypher := `
	MATCH (o:` + id.Label() + ` {id: $id})
	OPTIONAL MATCH (u:` + model.UserIDType + `)-[:` + EdgeKindMemberOf.String() + `]->(o)
	OPTIONAL MATCH (o)-[:` + EdgeKindHasNamespace.String() + `]->(n:` + model.NamespaceIDType + `)
	OPTIONAL MATCH (o)-[:` + EdgeKindHasTeam.String() + `]->(t:` + model.RoleIDType + `)
	RETURN o, collect(u.id) AS m, collect(n.id) AS n, collect(t.id) AS t
	`

	params := map[string]any{
		"id": id.String(),
	}

	org, err := ExecuteReadAndReadSingle(ctx, r.db, cypher, params, r.scan("o", "n", "t", "m"))
	if err != nil {
		return nil, errors.Join(ErrOrganizationRead, err)
	}

	return org, nil
}

func (r *OrganizationRepository) GetAll(ctx context.Context, offset, limit int) ([]*model.Organization, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.OrganizationRepository/GetAllBelongsTo")
	defer span.End()

	cypher := `
	MATCH (o:` + model.OrganizationIDType + `)
	OPTIONAL MATCH (u:` + model.UserIDType + `)-[:` + EdgeKindMemberOf.String() + `]->(o)
	OPTIONAL MATCH (o)-[:` + EdgeKindHasNamespace.String() + `]->(n:` + model.NamespaceIDType + `)
	OPTIONAL MATCH (o)-[:` + EdgeKindHasTeam.String() + `]->(t:` + model.RoleIDType + `)
	RETURN o, collect(u.id) AS m, collect(n.id) AS n, collect(t.id) AS t
	ORDER BY o.created_at DESC
	SKIP $offset LIMIT $limit`

	params := map[string]any{
		"offset": offset,
		"limit":  limit,
	}

	orgs, err := ExecuteReadAndReadAll(ctx, r.db, cypher, params, r.scan("o", "n", "t", "m"))
	if err != nil {
		return nil, errors.Join(ErrOrganizationRead, err)
	}

	return orgs, nil
}

func (r *OrganizationRepository) Update(ctx context.Context, id model.ID, patch map[string]any) (*model.Organization, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.OrganizationRepository/Update")
	defer span.End()

	cypher := `
	MATCH (o:` + id.Label() + ` {id: $id}) SET o += $patch SET o.updated_at = datetime($updated_at)
	WITH o
	OPTIONAL MATCH (u:` + model.UserIDType + `)-[:` + EdgeKindMemberOf.String() + `]->(o)
	OPTIONAL MATCH (o)-[:` + EdgeKindHasNamespace.String() + `]->(n:` + model.NamespaceIDType + `)
	OPTIONAL MATCH (o)-[:` + EdgeKindHasTeam.String() + `]->(t:` + model.RoleIDType + `)
	RETURN o, collect(u.id) AS m, collect(n.id) AS n, collect(t.id) AS t`

	params := map[string]any{
		"id":         id.String(),
		"patch":      patch,
		"updated_at": time.Now().Format(time.RFC3339Nano),
	}

	org, err := ExecuteReadAndReadSingle(ctx, r.db, cypher, params, r.scan("o", "n", "t", "m"))
	if err != nil {
		return nil, errors.Join(ErrOrganizationUpdate, err)
	}

	return org, nil
}

func (r *OrganizationRepository) AddMember(ctx context.Context, orgID, memberID model.ID) error {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.OrganizationRepository/AddMember")
	defer span.End()

	cypher := `
	MATCH (o:` + orgID.Label() + ` {id: $org_id}), (u:` + memberID.Label() + ` {id: $member_id})
	MERGE (u)-[m:` + EdgeKindMemberOf.String() + `]->(o)
	ON CREATE SET m.created_at = datetime($now), m.id = $membership_id
	ON MATCH SET m.updated_at = datetime($now)`

	params := map[string]any{
		"org_id":        orgID.String(),
		"member_id":     memberID.String(),
		"membership_id": model.MustNewID(EdgeKindMemberOf.String()).String(),
		"now":           time.Now().Format(time.RFC3339Nano),
	}

	if err := ExecuteWriteAndConsume(ctx, r.db, cypher, params); err != nil {
		return errors.Join(ErrOrganizationAddMember, err)
	}

	return nil
}

func (r *OrganizationRepository) RemoveMember(ctx context.Context, orgID, memberID model.ID) error {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.OrganizationRepository/RemoveMember")
	defer span.End()

	cypher := `
	MATCH (:` + memberID.Label() + ` {id: $member_id})-[r:` + EdgeKindMemberOf.String() + `]->(:` + orgID.Label() + ` {id: $org_id})
	DELETE r`

	params := map[string]any{
		"org_id":    orgID.String(),
		"member_id": memberID.String(),
	}

	if err := ExecuteWriteAndConsume(ctx, r.db, cypher, params); err != nil {
		return errors.Join(ErrOrganizationRemoveMember, err)
	}

	return nil
}

func (r *OrganizationRepository) Delete(ctx context.Context, id model.ID) error {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.OrganizationRepository/Delete")
	defer span.End()

	cypher := `MATCH (o:` + id.Label() + ` {id: $id}), (o)-[r]-() DETACH DELETE o, r`
	params := map[string]any{
		"id": id.String(),
	}

	if err := ExecuteWriteAndConsume(ctx, r.db, cypher, params); err != nil {
		return errors.Join(ErrOrganizationDelete, err)
	}

	return nil
}

// NewOrganizationRepository creates a new organization repository.
func NewOrganizationRepository(opts ...RepositoryOption) (*OrganizationRepository, error) {
	baseRepo, err := newRepository(opts...)
	if err != nil {
		return nil, err
	}

	return &OrganizationRepository{
		repository: baseRepo,
	}, nil
}

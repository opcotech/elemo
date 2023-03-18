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

func (r *OrganizationRepository) Create(ctx context.Context, owner model.ID, organization *model.Organization) error {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.OrganizationRepository/Create")
	defer span.End()

	if err := owner.Validate(); err != nil {
		return errors.Join(repository.ErrOrganizationCreate, err)
	}

	if err := organization.Validate(); err != nil {
		return errors.Join(repository.ErrOrganizationCreate, err)
	}

	createdAt := time.Now()

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

	params := map[string]interface{}{
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

func (r *OrganizationRepository) GetAll(ctx context.Context, offset, limit int) ([]*model.Organization, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.OrganizationRepository/GetAllBelongsTo")
	defer span.End()

	cypher := `
	MATCH (o:` + model.ResourceTypeOrganization.String() + `)
	OPTIONAL MATCH (u:` + model.ResourceTypeUser.String() + `)-[:` + EdgeKindMemberOf.String() + `]->(o)
	OPTIONAL MATCH (o)-[:` + EdgeKindHasNamespace.String() + `]->(n:` + model.ResourceTypeNamespace.String() + `)
	OPTIONAL MATCH (o)-[:` + EdgeKindHasTeam.String() + `]->(t:` + model.ResourceTypeRole.String() + `)
	RETURN o, collect(DISTINCT u.id) AS m, collect(DISTINCT n.id) AS n, collect(DISTINCT t.id) AS t
	ORDER BY o.created_at DESC
	SKIP $offset LIMIT $limit`

	params := map[string]any{
		"offset": offset,
		"limit":  limit,
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
	MATCH (o:` + id.Label() + ` {id: $id}) SET o += $patch, o.updated_at = datetime($updated_at)
	WITH o
	OPTIONAL MATCH (u:` + model.ResourceTypeUser.String() + `)-[:` + EdgeKindMemberOf.String() + `]->(o)
	OPTIONAL MATCH (o)-[:` + EdgeKindHasNamespace.String() + `]->(n:` + model.ResourceTypeNamespace.String() + `)
	OPTIONAL MATCH (o)-[:` + EdgeKindHasTeam.String() + `]->(t:` + model.ResourceTypeRole.String() + `)
	RETURN o, collect(DISTINCT u.id) AS m, collect(DISTINCT n.id) AS n, collect(DISTINCT t.id) AS t`

	params := map[string]any{
		"id":         id.String(),
		"patch":      patch,
		"updated_at": time.Now().Format(time.RFC3339Nano),
	}

	org, err := ExecuteWriteAndReadSingle(ctx, r.db, cypher, params, r.scan("o", "n", "t", "m"))
	if err != nil {
		return nil, errors.Join(repository.ErrOrganizationUpdate, err)
	}

	return org, nil
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
	MATCH (o:` + orgID.Label() + ` {id: $org_id}), (u:` + memberID.Label() + ` {id: $member_id})
	MERGE (u)-[m:` + EdgeKindMemberOf.String() + `]->(o)
	ON CREATE SET m.created_at = datetime($now), m.id = $membership_id
	ON MATCH SET m.updated_at = datetime($now)`

	params := map[string]any{
		"org_id":        orgID.String(),
		"member_id":     memberID.String(),
		"membership_id": model.NewRawID(),
		"now":           time.Now().Format(time.RFC3339Nano),
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

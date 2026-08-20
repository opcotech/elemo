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
	ErrTeamAddMember    = errors.New("failed to add member to team")      // member cannot be added to team
	ErrTeamCreate       = errors.New("failed to create team")             // team cannot be created
	ErrTeamDelete       = errors.New("failed to delete team")             // team cannot be deleted
	ErrTeamRead         = errors.New("failed to read team")               // team cannot be read
	ErrTeamRemoveMember = errors.New("failed to remove member from team") // member cannot be removed from team
	ErrTeamUpdate       = errors.New("failed to update team")             // team cannot be updated
)

// Team represents a team persisted by the repository.
type Team struct {
	ID          model.ID   `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	MemberCount *int64     `json:"member_count"`
	CreatedAt   *time.Time `json:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at"`
}

// CreateTeamOpts holds the data required to create a team.
type CreateTeamOpts struct {
	Name        string
	Description string
	CreatedBy   model.ID
	BelongsTo   model.ID
}

// UpdateTeamOpts holds the fields that can be updated on a team.
// Undefined fields (Defined == false) are left unchanged.
type UpdateTeamOpts struct {
	Name        optional.Optional[string]
	Description optional.Optional[string]
}

// patch builds a Neo4j property map from defined optional fields.
func (o UpdateTeamOpts) patch() map[string]any {
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

	return p
}

//go:generate go tool mockgen -source=team.go -destination=team_mock_gen.go -package=repository -mock_names "TeamRepository=MockTeamRepository"
type TeamRepository interface {
	Create(ctx context.Context, opts CreateTeamOpts) (*Team, error)
	Get(ctx context.Context, id, belongsTo model.ID, proj TeamProjection) (*Team, error)
	ListBelongsTo(ctx context.Context, belongsTo model.ID, page CursorPage, proj TeamProjection) (Page[*Team], error)
	ListMembers(ctx context.Context, teamID, belongsTo model.ID, page CursorPage) (Page[*User], error)
	Update(ctx context.Context, id, belongsTo model.ID, opts UpdateTeamOpts) (*Team, error)
	AddMember(ctx context.Context, teamID, memberID, belongsToID model.ID) error
	RemoveMember(ctx context.Context, teamID, memberID, belongsToID model.ID) error
	Delete(ctx context.Context, id, belongsTo model.ID) error
}

// Neo4jTeamRepository is a repository for managing teams.
type Neo4jTeamRepository struct {
	*neo4jBaseRepository
}

func (r *Neo4jTeamRepository) scan(proj TeamProjection) func(rec *neo4j.Record) (*Team, error) {
	return func(rec *neo4j.Record) (*Team, error) {
		node, err := Neo4jRecordNode(rec, "t")
		if err != nil {
			return nil, err
		}

		team := new(Team)
		if err := Neo4jScanIntoStruct(&node, &team, []string{"id"}); err != nil {
			return nil, err
		}

		team.ID, err = Neo4jDecodeID(node, model.ResourceTypeTeam)
		if err != nil {
			return nil, err
		}
		if proj.MemberCount {
			memberCount, err := Neo4jParseValueFromRecord[int64](rec, "member_count")
			if err != nil {
				return nil, err
			}
			team.MemberCount = convert.ToPointer(memberCount)
		}

		return team, nil
	}
}

// Create creates a new team owned by BelongsTo and labels it as a principal.
func (r *Neo4jTeamRepository) Create(ctx context.Context, opts CreateTeamOpts) (*Team, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.TeamRepository/Create")
	defer span.End()

	createdAt := time.Now().UTC()
	id := model.MustNewID(model.ResourceTypeTeam)

	cypher := `
	MATCH (b:` + opts.BelongsTo.Label() + ` {id: $belongs_to_id})
	MATCH (u:` + opts.CreatedBy.Label() + ` {id: $created_by})
	CREATE (t:` + id.Label() + `:` + model.LabelPrincipal + ` {id: $id, name: $name, description: $description, created_at: datetime($created_at)})
	CREATE (b)-[:` + EdgeKindHasTeam.String() + ` {id: $has_team_id, created_at: datetime($created_at)}]->(t)
	CREATE (t)-[:` + EdgeKindInScopeOf.String() + ` {id: $scope_id, created_at: datetime($created_at)}]->(b)
	`

	params := map[string]any{
		"belongs_to_id": opts.BelongsTo.String(),
		"created_by":    opts.CreatedBy.String(),
		"id":            id.String(),
		"has_team_id":   model.NewRawID(),
		"scope_id":      model.NewRawID(),
		"name":          opts.Name,
		"description":   opts.Description,
		"created_at":    createdAt.Format(time.RFC3339Nano),
	}

	if err := Neo4jExecuteWriteAndConsume(ctx, r.db, cypher, params); err != nil {
		return nil, errors.Join(err, ErrTeamCreate)
	}

	return r.Get(ctx, id, opts.BelongsTo, TeamDetailProjection())
}

// Get returns a team by ID that belongs to belongsTo.
func (r *Neo4jTeamRepository) Get(ctx context.Context, id, belongsTo model.ID, proj TeamProjection) (*Team, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.TeamRepository/Get")
	defer span.End()

	plan, err := CompileQuery(TeamGetQuery{ID: id, BelongsTo: belongsTo, Projection: proj})
	if err != nil {
		return nil, errors.Join(err, ErrTeamRead)
	}

	var team *Team
	err = Neo4jExecuteReadPlan(ctx, r.db, plan, func(tx neo4j.ManagedTransaction) error {
		var readErr error
		team, _, readErr = Neo4jRunQuerySingle(ctx, tx, plan.Root, r.scan(proj))
		return readErr
	})
	if err != nil {
		return nil, errors.Join(err, ErrTeamRead)
	}

	return team, nil
}

// ListBelongsTo returns a cursor-paginated page of teams for belongsTo.
func (r *Neo4jTeamRepository) ListBelongsTo(ctx context.Context, belongsTo model.ID, page CursorPage, proj TeamProjection) (Page[*Team], error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.TeamRepository/ListBelongsTo")
	defer span.End()

	normalized, err := page.Normalize()
	if err != nil {
		return Page[*Team]{}, errors.Join(ErrTeamRead, err)
	}
	plan, err := CompileQuery(TeamListBelongsToQuery{
		BelongsTo:  belongsTo,
		Page:       normalized,
		Order:      SortDirectionDesc,
		Projection: proj,
	})
	if err != nil {
		return Page[*Team]{}, errors.Join(ErrTeamRead, err)
	}

	teams := make([]*Team, 0)
	err = Neo4jExecuteReadPlan(ctx, r.db, plan, func(tx neo4j.ManagedTransaction) error {
		var readErr error
		teams, _, readErr = Neo4jRunQuery(ctx, tx, plan.Root, r.scan(proj))
		return readErr
	})
	if err != nil {
		return Page[*Team]{}, errors.Join(ErrTeamRead, err)
	}

	return PaginateSlice(teams, normalized.Size, func(team *Team) model.ID {
		return team.ID
	})
}

// ListMembers returns a cursor-paginated page of users who are members of the team.
func (r *Neo4jTeamRepository) ListMembers(ctx context.Context, teamID, belongsTo model.ID, page CursorPage) (Page[*User], error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.TeamRepository/ListMembers")
	defer span.End()

	normalized, err := page.Normalize()
	if err != nil {
		return Page[*User]{}, errors.Join(ErrTeamRead, err)
	}
	plan, err := CompileQuery(TeamMemberListQuery{
		TeamID:    teamID,
		BelongsTo: belongsTo,
		Page:      normalized,
		Order:     SortDirectionDesc,
	})
	if err != nil {
		return Page[*User]{}, errors.Join(ErrTeamRead, err)
	}

	users := make([]*User, 0)
	err = Neo4jExecuteReadPlan(ctx, r.db, plan, func(tx neo4j.ManagedTransaction) error {
		scanned, _, readErr := Neo4jRunQuery(ctx, tx, plan.Root, scanTeamMemberUser)
		if readErr != nil {
			return readErr
		}
		users = scanned
		return nil
	})
	if err != nil {
		return Page[*User]{}, errors.Join(ErrTeamRead, err)
	}

	return PaginateSlice(users, normalized.Size, func(user *User) model.ID {
		return user.ID
	})
}

func scanTeamMemberUser(rec *neo4j.Record) (*User, error) {
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

// Update updates a team that belongs to belongsTo with any given opts.
func (r *Neo4jTeamRepository) Update(ctx context.Context, id, belongsTo model.ID, opts UpdateTeamOpts) (*Team, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.TeamRepository/Update")
	defer span.End()

	cypher := `
	MATCH (:` + belongsTo.Label() + ` {id: $belongs_to_id})-[:` + EdgeKindHasTeam.String() + `]->(t:` + id.Label() + ` {id: $id})
	SET t += $patch, t.updated_at = datetime()
	RETURN t.id AS id`

	params := map[string]any{
		"id":            id.String(),
		"belongs_to_id": belongsTo.String(),
		"patch":         opts.patch(),
	}

	_, err := Neo4jExecuteWriteAndReadSingle(ctx, r.db, cypher, params, func(_ *neo4j.Record) (*struct{}, error) {
		return &struct{}{}, nil
	})
	if err != nil {
		return nil, errors.Join(err, ErrTeamUpdate)
	}

	return r.Get(ctx, id, belongsTo, TeamDetailProjection())
}

// AddMember adds a user as a member of the team.
func (r *Neo4jTeamRepository) AddMember(ctx context.Context, teamID, memberID, belongsToID model.ID) error {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.TeamRepository/AddMember")
	defer span.End()

	cypher := `
	MATCH (b:` + belongsToID.Label() + ` {id: $belongs_to_id})-[:` + EdgeKindHasTeam.String() + `]->(t:` + teamID.Label() + ` {id: $team_id})
	MATCH (u:` + memberID.Label() + ` {id: $member_id})
	MERGE (u)-[m:` + EdgeKindMemberOf.String() + `]->(t)
	ON CREATE SET m.created_at = datetime($now), m.id = $membership_id
	ON MATCH SET m.updated_at = datetime($now)`

	params := map[string]any{
		"team_id":       teamID.String(),
		"member_id":     memberID.String(),
		"belongs_to_id": belongsToID.String(),
		"membership_id": model.NewRawID(),
		"now":           time.Now().UTC().Format(time.RFC3339Nano),
	}

	if err := Neo4jExecuteWriteAndConsume(ctx, r.db, cypher, params); err != nil {
		return errors.Join(ErrTeamAddMember, err)
	}

	return nil
}

// RemoveMember removes a user from the team.
func (r *Neo4jTeamRepository) RemoveMember(ctx context.Context, teamID, memberID, belongsToID model.ID) error {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.TeamRepository/RemoveMember")
	defer span.End()

	cypher := `
	MATCH (b:` + belongsToID.Label() + ` {id: $belongs_to_id})-[:` + EdgeKindHasTeam.String() + `]->(t:` + teamID.Label() + ` {id: $team_id})
	MATCH (t)<-[r:` + EdgeKindMemberOf.String() + `]-(:` + memberID.Label() + ` {id: $member_id})
	DELETE r`

	params := map[string]any{
		"team_id":       teamID.String(),
		"member_id":     memberID.String(),
		"belongs_to_id": belongsToID.String(),
	}

	if err := Neo4jExecuteWriteAndConsume(ctx, r.db, cypher, params); err != nil {
		return errors.Join(ErrTeamRemoveMember, err)
	}

	return nil
}

// Delete permanently deletes a team that belongs to belongsTo.
func (r *Neo4jTeamRepository) Delete(ctx context.Context, id, belongsTo model.ID) error {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.TeamRepository/Delete")
	defer span.End()

	cypher := `
	MATCH (b:` + belongsTo.Label() + ` {id: $belongs_to_id})-[:` + EdgeKindHasTeam.String() + `]->(t:` + id.Label() + ` {id: $id})
	DETACH DELETE t`
	params := map[string]any{
		"id":            id.String(),
		"belongs_to_id": belongsTo.String(),
	}

	if err := Neo4jExecuteWriteAndConsume(ctx, r.db, cypher, params); err != nil {
		return errors.Join(ErrTeamDelete, err)
	}

	return nil
}

// NewNeo4jTeamRepository creates a new team neo4jBaseRepository.
func NewNeo4jTeamRepository(opts ...Neo4jRepositoryOption) (*Neo4jTeamRepository, error) {
	baseRepo, err := newNeo4jRepository(opts...)
	if err != nil {
		return nil, err
	}

	return &Neo4jTeamRepository{
		neo4jBaseRepository: baseRepo,
	}, nil
}

func clearTeamsPattern(ctx context.Context, r *redisBaseRepository, pattern ...string) error {
	return r.DeletePattern(ctx, composeCacheKey(model.ResourceTypeTeam.String(), pattern))
}

func clearTeamsKey(ctx context.Context, r *redisBaseRepository, id model.ID) error {
	return clearTeamsPattern(ctx, r, "Get", id.String(), "*")
}

func clearTeamsBelongsTo(ctx context.Context, r *redisBaseRepository, id model.ID) error {
	return clearTeamsPattern(ctx, r, "ListBelongsTo", id.String(), "*", "*", "*")
}

func clearTeamsAllBelongsTo(ctx context.Context, r *redisBaseRepository) error {
	return clearTeamsPattern(ctx, r, "ListBelongsTo", "*")
}

func clearTeamAllCrossCache(ctx context.Context, r *redisBaseRepository) error {
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

// RedisCachedTeamRepository implements caching on the TeamRepository.
type RedisCachedTeamRepository struct {
	cacheRepo *redisBaseRepository
	teamRepo  TeamRepository
}

func (r *RedisCachedTeamRepository) Create(ctx context.Context, opts CreateTeamOpts) (*Team, error) {
	if err := clearTeamsBelongsTo(ctx, r.cacheRepo, opts.BelongsTo); err != nil {
		return nil, err
	}
	if err := clearTeamAllCrossCache(ctx, r.cacheRepo); err != nil {
		return nil, err
	}

	return r.teamRepo.Create(ctx, opts)
}

func (r *RedisCachedTeamRepository) Get(ctx context.Context, id, belongsTo model.ID, proj TeamProjection) (*Team, error) {
	var team *Team
	var err error

	key := composeCacheKey(model.ResourceTypeTeam.String(), "Get", id.String(), projectionCacheValue(proj))
	if err = r.cacheRepo.Get(ctx, key, &team); err != nil {
		return nil, err
	}

	if team != nil {
		return team, nil
	}

	if team, err = r.teamRepo.Get(ctx, id, belongsTo, proj); err != nil {
		return nil, err
	}

	if err = r.cacheRepo.Set(ctx, key, team); err != nil {
		return nil, err
	}

	return team, nil
}

func (r *RedisCachedTeamRepository) ListBelongsTo(ctx context.Context, belongsTo model.ID, page CursorPage, proj TeamProjection) (Page[*Team], error) {
	var teams Page[*Team]
	var err error

	normalized, err := normalizedPage(page)
	if err != nil {
		return Page[*Team]{}, err
	}

	key := composeCacheKey(model.ResourceTypeTeam.String(), "ListBelongsTo", belongsTo.String(), projectionCacheValue(proj), pageTokenValue(normalized.Token), normalized.Size)
	if err = r.cacheRepo.Get(ctx, key, &teams); err != nil {
		return Page[*Team]{}, err
	}

	if teams.Items != nil {
		return teams, nil
	}

	if teams, err = r.teamRepo.ListBelongsTo(ctx, belongsTo, normalized, proj); err != nil {
		return Page[*Team]{}, err
	}

	if err = r.cacheRepo.Set(ctx, key, teams); err != nil {
		return Page[*Team]{}, err
	}

	return teams, nil
}

func (r *RedisCachedTeamRepository) ListMembers(ctx context.Context, teamID, belongsTo model.ID, page CursorPage) (Page[*User], error) {
	return r.teamRepo.ListMembers(ctx, teamID, belongsTo, page)
}

func (r *RedisCachedTeamRepository) Update(ctx context.Context, id, belongsTo model.ID, opts UpdateTeamOpts) (*Team, error) {
	team, err := r.teamRepo.Update(ctx, id, belongsTo, opts)
	if err != nil {
		return nil, err
	}

	key := composeCacheKey(model.ResourceTypeTeam.String(), "Get", id.String(), projectionCacheValue(TeamDetailProjection()))
	if err = r.cacheRepo.Set(ctx, key, team); err != nil {
		return nil, err
	}

	if err := clearTeamsAllBelongsTo(ctx, r.cacheRepo); err != nil {
		return nil, err
	}

	return team, nil
}

func (r *RedisCachedTeamRepository) AddMember(ctx context.Context, teamID, memberID, belongsToID model.ID) error {
	if err := clearTeamsKey(ctx, r.cacheRepo, teamID); err != nil {
		return err
	}
	if err := clearTeamsAllBelongsTo(ctx, r.cacheRepo); err != nil {
		return err
	}
	if err := clearOrganizationsKey(ctx, r.cacheRepo, belongsToID); err != nil {
		return err
	}
	if err := clearProjectsGet(ctx, r.cacheRepo, belongsToID); err != nil {
		return err
	}

	return r.teamRepo.AddMember(ctx, teamID, memberID, belongsToID)
}

func (r *RedisCachedTeamRepository) RemoveMember(ctx context.Context, teamID, memberID, belongsToID model.ID) error {
	if err := clearTeamsKey(ctx, r.cacheRepo, teamID); err != nil {
		return err
	}
	if err := clearTeamsAllBelongsTo(ctx, r.cacheRepo); err != nil {
		return err
	}
	if err := clearOrganizationsKey(ctx, r.cacheRepo, belongsToID); err != nil {
		return err
	}
	if err := clearProjectsGet(ctx, r.cacheRepo, belongsToID); err != nil {
		return err
	}

	return r.teamRepo.RemoveMember(ctx, teamID, memberID, belongsToID)
}

func (r *RedisCachedTeamRepository) Delete(ctx context.Context, id, belongsTo model.ID) error {
	if err := clearTeamsKey(ctx, r.cacheRepo, id); err != nil {
		return err
	}

	if err := clearTeamsAllBelongsTo(ctx, r.cacheRepo); err != nil {
		return err
	}

	if err := clearTeamAllCrossCache(ctx, r.cacheRepo); err != nil {
		return err
	}

	return r.teamRepo.Delete(ctx, id, belongsTo)
}

// NewCachedTeamRepository returns a new CachedTeamRepository.
func NewCachedTeamRepository(repo TeamRepository, opts ...RedisRepositoryOption) (*RedisCachedTeamRepository, error) {
	r, err := newRedisBaseRepository(opts...)
	if err != nil {
		return nil, err
	}

	return &RedisCachedTeamRepository{
		cacheRepo: r,
		teamRepo:  repo,
	}, nil
}

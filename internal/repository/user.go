package repository

import (
	"context"
	"errors"
	"time"

	"github.com/neo4j/neo4j-go-driver/v6/neo4j"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg/convert"
	"github.com/opcotech/elemo/internal/pkg/log"
	"github.com/opcotech/elemo/internal/pkg/optional"
)

var (
	ErrUserCreate = errors.New("failed to create user") // user cannot be created
	ErrUserDelete = errors.New("failed to delete user") // user cannot be deleted
	ErrUserRead   = errors.New("failed to read user")   // user cannot be read
	ErrUserUpdate = errors.New("failed to update user") // user cannot be updated

	ErrUserActivationPolicy = errors.New("active user mutation requires an explicit seat policy")
	ErrUserActivationStatus = errors.New("user status does not allow activation")
	ErrUserAcceptInvitation = errors.New("failed to activate user and accept organization invitation")
)

// PartialUser is a lean user used on issue and document reads.
type PartialUser struct {
	ID        model.ID `json:"id"`
	FirstName string   `json:"first_name"`
	LastName  string   `json:"last_name"`
	Picture   string   `json:"picture"`
}

// User represents a user persisted by the repository.
type User struct {
	ID            model.ID         `json:"id"`
	Username      string           `json:"username"`
	Email         string           `json:"email"`
	Password      string           `json:"password"`
	Status        model.UserStatus `json:"status"`
	FirstName     string           `json:"first_name"`
	LastName      string           `json:"last_name"`
	Picture       string           `json:"picture"`
	Title         string           `json:"title"`
	Bio           string           `json:"bio"`
	Phone         string           `json:"phone"`
	Address       string           `json:"address"`
	Links         []string         `json:"links"`
	Languages     []model.Language `json:"languages"`
	DocumentCount *int64           `json:"document_count"`
	Permissions   []model.ID       `json:"permissions"`
	CreatedAt     *time.Time       `json:"created_at"`
	UpdatedAt     *time.Time       `json:"updated_at"`
}

// ActivationAuthorization is the explicit seat decision for creating or
// activating a human user. The zero value is invalid. Nil on CreateUserOpts
// means no activation was requested, never unrestricted.
type ActivationAuthorization struct {
	unlimited bool
	limit     uint32
}

// UnrestrictedActivation is self-hosted or otherwise unlimited activation.
func UnrestrictedActivation() ActivationAuthorization {
	return ActivationAuthorization{unlimited: true}
}

// FiniteActivation authorizes activation up to limit active humans.
func FiniteActivation(limit uint32) ActivationAuthorization {
	return ActivationAuthorization{limit: limit}
}

func (a ActivationAuthorization) params() (unlimited bool, limit int64) {
	if a.unlimited {
		return true, 0
	}
	return false, int64(a.limit)
}

// CreateUserOpts holds the data required to create a user.
type CreateUserOpts struct {
	Username  string
	Email     string
	Password  string
	Status    model.UserStatus
	FirstName string
	LastName  string
	Picture   string
	Title     string
	Bio       string
	Phone     string
	Address   string
	Links     []string
	Languages []model.Language
	// Activation is required when Status is active (including zero Status).
	// nil means no activation was requested and is rejected for active users.
	Activation *ActivationAuthorization
}

// UpdateUserOpts holds the fields that can be updated on a user.
// Undefined fields (Defined == false) are left unchanged.
type UpdateUserOpts struct {
	Username  optional.Optional[string]
	Email     optional.Optional[string]
	Password  optional.Optional[string]
	Status    optional.Optional[model.UserStatus]
	FirstName optional.Optional[string]
	LastName  optional.Optional[string]
	Picture   optional.Optional[string]
	Title     optional.Optional[string]
	Bio       optional.Optional[string]
	Phone     optional.Optional[string]
	Address   optional.Optional[string]
	Links     optional.Optional[[]string]
	Languages optional.Optional[[]model.Language]
}

// patch builds a Neo4j property map from defined optional fields.
func (o UpdateUserOpts) patch() map[string]any {
	p := make(map[string]any)

	if o.Username.Defined {
		p["username"] = *o.Username.Value
	}
	if o.Email.Defined {
		p["email"] = *o.Email.Value
	}
	if o.Password.Defined {
		p["password"] = *o.Password.Value
	}
	if o.Status.Defined {
		p["status"] = o.Status.Value.String()
	}
	if o.FirstName.Defined {
		p["first_name"] = *o.FirstName.Value
	}
	if o.LastName.Defined {
		p["last_name"] = *o.LastName.Value
	}
	if o.Picture.Defined {
		if o.Picture.Value == nil {
			p["picture"] = nil
		} else {
			p["picture"] = *o.Picture.Value
		}
	}
	if o.Title.Defined {
		if o.Title.Value == nil {
			p["title"] = nil
		} else {
			p["title"] = *o.Title.Value
		}
	}
	if o.Bio.Defined {
		if o.Bio.Value == nil {
			p["bio"] = nil
		} else {
			p["bio"] = *o.Bio.Value
		}
	}
	if o.Phone.Defined {
		if o.Phone.Value == nil {
			p["phone"] = nil
		} else {
			p["phone"] = *o.Phone.Value
		}
	}
	if o.Address.Defined {
		if o.Address.Value == nil {
			p["address"] = nil
		} else {
			p["address"] = *o.Address.Value
		}
	}
	if o.Links.Defined {
		if o.Links.Value == nil {
			p["links"] = nil
		} else {
			p["links"] = *o.Links.Value
		}
	}
	if o.Languages.Defined {
		if o.Languages.Value == nil {
			p["languages"] = nil
		} else {
			languages := make([]string, len(*o.Languages.Value))
			for i, l := range *o.Languages.Value {
				languages[i] = l.String()
			}
			p["languages"] = languages
		}
	}

	return p
}

//go:generate go tool mockgen -source=user.go -destination=mock/mock_user_gen.go -package=mockrepo
type UserRepository interface {
	// Create persists a user. Active users (including zero Status) require
	// Activation; nil means no activation was requested and is rejected.
	// Pending and inactive users skip the seat lock.
	Create(ctx context.Context, opts CreateUserOpts) (*User, error)
	Get(ctx context.Context, id model.ID, proj UserProjection) (*User, error)
	GetByEmail(ctx context.Context, email string, proj UserProjection) (*User, error)
	List(ctx context.Context, page CursorPage, proj UserProjection) (Page[*User], error)
	Update(ctx context.Context, id model.ID, opts UpdateUserOpts) (*User, error)
	Delete(ctx context.Context, id model.ID) error
	Activate(ctx context.Context, id model.ID, opts UpdateUserOpts, auth ActivationAuthorization) (*User, error)
	// AcceptInvitation is the persistence used by
	// OrganizationService.AcceptInvitation. It atomically activates the
	// user when entitled, deletes the invitation edge, creates membership,
	// and optionally grants a role. OrganizationRepository still owns
	// invite/revoke edges.
	AcceptInvitation(ctx context.Context, userID, orgID model.ID, password string, activation *ActivationAuthorization, roleID *model.ID) (*User, error)
	ActiveHumanCount(ctx context.Context) (int, error)
}

// Neo4jUserRepository is a repository for managing users.
type Neo4jUserRepository struct {
	*neo4jBaseRepository
}

// scan is a helper function for scanning a user from a Neo4j Record.
func (r *Neo4jUserRepository) scan(up string, proj UserProjection) func(rec *neo4j.Record) (*User, error) {
	return func(rec *neo4j.Record) (*User, error) {
		user := new(User)
		user.Links = make([]string, 0)
		user.Permissions = make([]model.ID, 0)

		val, _, err := neo4j.GetRecordValue[neo4j.Node](rec, up)
		if err != nil {
			return nil, err
		}

		if err := Neo4jScanIntoStruct(&val, &user, []string{"id", "document_count"}); err != nil {
			return nil, err
		}

		user.ID, _ = model.NewIDFromString(val.GetProperties()["id"].(string), model.ResourceTypeUser.String())

		if proj.DocumentCount {
			documentCount, err := Neo4jParseValueFromRecord[int64](rec, "document_count")
			if err != nil {
				return nil, err
			}
			user.DocumentCount = convert.ToPointer(documentCount)
		}

		return user, nil
	}
}

func (r *Neo4jUserRepository) Create(ctx context.Context, opts CreateUserOpts) (*User, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.UserRepository/Create")
	defer span.End()

	if opts.Status == 0 {
		opts.Status = model.UserStatusActive
	}
	if opts.Status == model.UserStatusActive {
		if opts.Activation == nil {
			return nil, errors.Join(ErrUserCreate, ErrUserActivationPolicy)
		}
		return r.createActive(ctx, opts)
	}
	return r.createNonActive(ctx, opts)
}

func (r *Neo4jUserRepository) createNonActive(ctx context.Context, opts CreateUserOpts) (*User, error) {
	id, params := userCreateParams(opts)

	cypher := `
	MERGE (u:` + id.Label() + ` {id: $id})
	ON CREATE SET u += {
		username: $username, email: $email, password: $password, status: $status, first_name: $first_name,
		last_name: $last_name, picture: $picture, title: $title, bio: $bio, phone: $phone, address: $address,
		links: $links, languages: $languages, created_at: datetime($created_at)
	}
	SET u:` + model.LabelPrincipal

	if err := Neo4jExecuteWriteAndConsume(ctx, r.db, cypher, params); err != nil {
		return nil, errors.Join(ErrUserCreate, err)
	}

	return r.Get(ctx, id, UserDetailProjection())
}

func (r *Neo4jUserRepository) applyUserLoaders(ctx context.Context, tx neo4j.ManagedTransaction, plan QueryPlan, users []*User) error {
	if len(plan.Loaders) == 0 || len(users) == 0 {
		return nil
	}

	userByID := make(map[string]*User, len(users))
	ids := make([]string, 0, len(users))
	for _, user := range users {
		if user == nil {
			continue
		}
		userByID[user.ID.String()] = user
		ids = append(ids, user.ID.String())
	}

	for _, loader := range plan.Loaders {
		query := loader
		query.Params = cloneParams(loader.Params)
		query.Params["ids"] = ids
		switch loader.Name {
		case "user.load_permissions":
			type permissionRow struct {
				UserID        string
				PermissionIDs []model.ID
			}
			rows, _, err := Neo4jRunQuery(ctx, tx, query, func(rec *neo4j.Record) (permissionRow, error) {
				userID, err := Neo4jParseValueFromRecord[string](rec, "user_id")
				if err != nil {
					return permissionRow{}, err
				}
				ids, err := Neo4jParseIDsFromRecord(rec, "permission_ids", model.ResourceTypePermission.String())
				if err != nil {
					return permissionRow{}, err
				}
				return permissionRow{UserID: userID, PermissionIDs: ids}, nil
			})
			if err != nil {
				return err
			}
			for _, row := range rows {
				if user := userByID[row.UserID]; user != nil {
					user.Permissions = row.PermissionIDs
				}
			}
		default:
			return ErrQueryCompile
		}
	}

	return nil
}

// Get returns a user by its ID.
func (r *Neo4jUserRepository) Get(ctx context.Context, id model.ID, proj UserProjection) (*User, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.UserRepository/Get")
	defer span.End()

	plan, err := CompileQuery(UserGetQuery{
		ID:         id,
		Projection: proj,
	})
	if err != nil {
		return nil, errors.Join(ErrUserRead, err)
	}

	var user *User
	err = Neo4jExecuteReadPlan(ctx, r.db, plan, func(tx neo4j.ManagedTransaction) error {
		var readErr error
		user, _, readErr = Neo4jRunQuerySingle(ctx, tx, plan.Root, r.scan("u", proj))
		if readErr != nil {
			return readErr
		}
		return r.applyUserLoaders(ctx, tx, plan, []*User{user})
	})
	if err != nil {
		if errors.As(err, &ErrNoMoreRecords) {
			return nil, errors.Join(ErrUserRead, ErrNotFound)
		}
		return nil, errors.Join(ErrUserRead, err)
	}

	return user, nil
}

// GetByEmail returns a user by its email.
func (r *Neo4jUserRepository) GetByEmail(ctx context.Context, email string, proj UserProjection) (*User, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.UserRepository/GetByEmail")
	defer span.End()

	plan, err := CompileQuery(UserGetByEmailQuery{
		Email:      email,
		Projection: proj,
	})
	if err != nil {
		return nil, errors.Join(ErrUserRead, err)
	}

	var user *User
	err = Neo4jExecuteReadPlan(ctx, r.db, plan, func(tx neo4j.ManagedTransaction) error {
		var readErr error
		user, _, readErr = Neo4jRunQuerySingle(ctx, tx, plan.Root, r.scan("u", proj))
		if readErr != nil {
			return readErr
		}
		return r.applyUserLoaders(ctx, tx, plan, []*User{user})
	})
	if err != nil {
		if errors.As(err, &ErrNoMoreRecords) {
			return nil, errors.Join(ErrUserRead, ErrNotFound)
		}
		return nil, errors.Join(ErrUserRead, err)
	}

	return user, nil
}

// List returns users with cursor pagination.
func (r *Neo4jUserRepository) List(ctx context.Context, page CursorPage, proj UserProjection) (Page[*User], error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.UserRepository/List")
	defer span.End()

	normalized, err := page.Normalize()
	if err != nil {
		return Page[*User]{}, errors.Join(ErrUserRead, err)
	}
	plan, err := CompileQuery(UserListQuery{
		Page:       normalized,
		Order:      SortDirectionDesc,
		Projection: proj,
	})
	if err != nil {
		return Page[*User]{}, errors.Join(ErrUserRead, err)
	}

	users := make([]*User, 0)
	err = Neo4jExecuteReadPlan(ctx, r.db, plan, func(tx neo4j.ManagedTransaction) error {
		var readErr error
		users, _, readErr = Neo4jRunQuery(ctx, tx, plan.Root, r.scan("u", proj))
		if readErr != nil {
			return readErr
		}
		return r.applyUserLoaders(ctx, tx, plan, users)
	})
	if err != nil {
		if errors.As(err, &ErrNoMoreRecords) {
			return Page[*User]{}, errors.Join(ErrUserRead, ErrNotFound)
		}
		return Page[*User]{}, errors.Join(ErrUserRead, err)
	}

	return PaginateSlice(users, normalized.Size, func(user *User) model.ID {
		return user.ID
	})
}

// Update updates a user by its ID with any given opts.
func (r *Neo4jUserRepository) Update(ctx context.Context, id model.ID, opts UpdateUserOpts) (*User, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.UserRepository/Update")
	defer span.End()

	if opts.Status.Defined && opts.Status.Value != nil && *opts.Status.Value == model.UserStatusActive {
		return nil, errors.Join(ErrUserUpdate, ErrUserActivationPolicy)
	}

	cypher := `
	MATCH (u:` + id.Label() + ` {id: $id})
	SET u += $patch
	SET u.updated_at = datetime.statement()
	RETURN u.id AS id
	`
	params := map[string]any{
		"id":    id.String(),
		"patch": opts.patch(),
	}

	_, err := Neo4jExecuteWriteAndReadSingle(ctx, r.db, cypher, params, func(_ *neo4j.Record) (*struct{}, error) {
		return &struct{}{}, nil
	})
	if err != nil {
		if errors.As(err, &ErrNoMoreRecords) {
			return nil, errors.Join(ErrUserRead, ErrNotFound)
		}
		return nil, errors.Join(ErrUserUpdate, err)
	}

	return r.Get(ctx, id, UserDetailProjection())
}

// Delete deletes a user by its ID.
func (r *Neo4jUserRepository) Delete(ctx context.Context, id model.ID) error {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.UserRepository/Delete")
	defer span.End()

	cypher := `MATCH (u:` + id.Label() + ` {id: $id}) DETACH DELETE u`
	params := map[string]any{
		"id": id.String(),
	}

	if err := Neo4jExecuteWriteAndConsume(ctx, r.db, cypher, params); err != nil {
		return errors.Join(ErrUserDelete, err)
	}

	return nil
}

func userCreateParams(opts CreateUserOpts) (model.ID, map[string]any) {
	createdAt := time.Now().UTC()
	id := model.MustNewID(model.ResourceTypeUser)

	status := opts.Status
	if status == 0 {
		status = model.UserStatusActive
	}

	links := opts.Links
	if links == nil {
		links = make([]string, 0)
	}

	languages := opts.Languages
	if languages == nil {
		languages = make([]model.Language, 0)
	}
	languageValues := make([]string, len(languages))
	for i, language := range languages {
		languageValues[i] = language.String()
	}

	return id, map[string]any{
		"id":         id.String(),
		"username":   opts.Username,
		"email":      opts.Email,
		"password":   opts.Password,
		"status":     status.String(),
		"first_name": opts.FirstName,
		"last_name":  opts.LastName,
		"picture":    opts.Picture,
		"title":      opts.Title,
		"bio":        opts.Bio,
		"phone":      opts.Phone,
		"address":    opts.Address,
		"links":      links,
		"languages":  languageValues,
		"created_at": createdAt.Format(time.RFC3339Nano),
	}
}

func (r *Neo4jUserRepository) createActive(ctx context.Context, opts CreateUserOpts) (*User, error) {
	id, params := userCreateParams(opts)
	bindSeatParams(params, *opts.Activation)

	cypher := seatLockCypher() + `
	WITH i
	OPTIONAL MATCH (n:` + model.ResourceTypeUser.String() + ` {status: $active_status})
	WITH i, count(n) AS active_count
	WITH i, i IS NOT NULL AS installation_found,
		CASE WHEN $unlimited OR (i IS NOT NULL AND active_count < $limit) THEN true ELSE false END AS allowed
	FOREACH (_ IN CASE WHEN allowed THEN [1] ELSE [] END |
		CREATE (u:` + id.Label() + ` {id: $id})
		SET u += {
			username: $username, email: $email, password: $password, status: $status, first_name: $first_name,
			last_name: $last_name, picture: $picture, title: $title, bio: $bio, phone: $phone, address: $address,
			links: $links, languages: $languages, created_at: datetime($created_at)
		}
		SET u:` + model.LabelPrincipal + `
	)
	RETURN installation_found, allowed
	`

	type createResult struct {
		InstallationFound bool
		Allowed           bool
	}
	result, err := Neo4jExecuteWriteAndReadSingle(ctx, r.db, cypher, params, func(rec *neo4j.Record) (*createResult, error) {
		installationFound, _, err := neo4j.GetRecordValue[bool](rec, "installation_found")
		if err != nil {
			return nil, err
		}
		allowed, _, err := neo4j.GetRecordValue[bool](rec, "allowed")
		if err != nil {
			return nil, err
		}
		return &createResult{InstallationFound: installationFound, Allowed: allowed}, nil
	})
	if err != nil {
		return nil, errors.Join(ErrUserCreate, mapUniquenessError(err))
	}
	if !opts.Activation.unlimited && !result.InstallationFound {
		return nil, errors.Join(ErrUserCreate, ErrInstallationRead, ErrNotFound)
	}
	if !result.Allowed {
		return nil, errors.Join(ErrUserCreate, ErrSeatLimitReached)
	}

	return r.Get(ctx, id, UserDetailProjection())
}

func (r *Neo4jUserRepository) Activate(ctx context.Context, id model.ID, opts UpdateUserOpts, auth ActivationAuthorization) (*User, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.UserRepository/Activate")
	defer span.End()

	if !opts.Status.Defined || opts.Status.Value == nil || *opts.Status.Value != model.UserStatusActive {
		return nil, errors.Join(ErrUserUpdate, ErrUserActivationStatus)
	}

	params := map[string]any{
		"id":              id.String(),
		"patch":           opts.patch(),
		"pending_status":  model.UserStatusPending.String(),
		"inactive_status": model.UserStatusInactive.String(),
	}
	bindSeatParams(params, auth)

	cypher := seatLockCypher() + `
	WITH i
	OPTIONAL MATCH (u:` + id.Label() + ` {id: $id})
	WITH i, u, u.status AS prev
	OPTIONAL MATCH (n:` + model.ResourceTypeUser.String() + ` {status: $active_status})
	WITH i, u, prev, count(n) AS active_count
	WITH i, u, prev, i IS NOT NULL AS installation_found, u IS NOT NULL AS user_found,
		CASE WHEN prev IN [$active_status, $pending_status, $inactive_status] THEN true ELSE false END AS status_allowed,
		CASE
			WHEN u IS NULL THEN false
			WHEN NOT (prev IN [$active_status, $pending_status, $inactive_status]) THEN false
			WHEN NOT $unlimited AND i IS NULL THEN false
			WHEN prev = $active_status THEN true
			WHEN $unlimited THEN true
			ELSE active_count < $limit
		END AS allowed
	FOREACH (_ IN CASE WHEN allowed THEN [1] ELSE [] END |
		SET u += $patch
		SET u.updated_at = datetime.statement()
	)
	RETURN installation_found, user_found, status_allowed, allowed
	`

	type activateResult struct {
		InstallationFound bool
		UserFound         bool
		StatusAllowed     bool
		Allowed           bool
	}

	result, err := Neo4jExecuteWriteAndReadSingle(ctx, r.db, cypher, params, func(rec *neo4j.Record) (*activateResult, error) {
		installationFound, _, err := neo4j.GetRecordValue[bool](rec, "installation_found")
		if err != nil {
			return nil, err
		}
		userFound, _, err := neo4j.GetRecordValue[bool](rec, "user_found")
		if err != nil {
			return nil, err
		}
		statusAllowed, _, err := neo4j.GetRecordValue[bool](rec, "status_allowed")
		if err != nil {
			return nil, err
		}
		allowed, _, err := neo4j.GetRecordValue[bool](rec, "allowed")
		if err != nil {
			return nil, err
		}
		return &activateResult{
			InstallationFound: installationFound,
			UserFound:         userFound,
			StatusAllowed:     statusAllowed,
			Allowed:           allowed,
		}, nil
	})
	if err != nil {
		return nil, errors.Join(ErrUserUpdate, err)
	}
	if !auth.unlimited && !result.InstallationFound {
		return nil, errors.Join(ErrUserUpdate, ErrInstallationRead, ErrNotFound)
	}
	if !result.UserFound {
		return nil, errors.Join(ErrUserUpdate, ErrNotFound)
	}
	if !result.StatusAllowed {
		return nil, errors.Join(ErrUserUpdate, ErrUserActivationStatus)
	}
	if !result.Allowed {
		return nil, errors.Join(ErrUserUpdate, ErrSeatLimitReached)
	}

	return r.Get(ctx, id, UserDetailProjection())
}

func bindSeatParams(params map[string]any, auth ActivationAuthorization) {
	params["installation_id"] = model.InstallationID().String()
	params["unlimited"], params["limit"] = auth.params()
	params["active_status"] = model.UserStatusActive.String()
}

func seatLockCypher() string {
	return `
	OPTIONAL MATCH (i:` + model.ResourceTypeInstallation.String() + ` {id: $installation_id})
	FOREACH (_ IN CASE WHEN i IS NULL THEN [] ELSE [1] END |
		SET i.seat_lock = timestamp()
	)`
}

// AcceptInvitation atomically activates a user, removes the invitation,
// creates organization membership, and optionally grants a role while
// holding the installation seat lock.
func (r *Neo4jUserRepository) AcceptInvitation(
	ctx context.Context,
	userID, orgID model.ID,
	hashedPassword string,
	activation *ActivationAuthorization,
	roleID *model.ID,
) (*User, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.UserRepository/AcceptInvitation")
	defer span.End()

	allowActivation := activation != nil
	auth := UnrestrictedActivation()
	if activation != nil {
		auth = *activation
	}

	roleIDValue := ""
	grantID := ""
	if roleID != nil && !roleID.IsNil() {
		roleIDValue = roleID.String()
		grantID = model.MustNewID(model.ResourceTypePermission).String()
	}

	params := map[string]any{
		"user_id":          userID.String(),
		"org_id":           orgID.String(),
		"password":         hashedPassword,
		"pending_status":   model.UserStatusPending.String(),
		"membership_id":    model.NewRawID(),
		"now":              time.Now().UTC().Format(time.RFC3339Nano),
		"allow_activation": allowActivation,
		"role_id":          roleIDValue,
		"grant_id":         grantID,
	}
	bindSeatParams(params, auth)

	cypher := seatLockCypher() + `
	WITH i
	OPTIONAL MATCH (u:` + model.ResourceTypeUser.String() + ` {id: $user_id})
	OPTIONAL MATCH (o:` + model.ResourceTypeOrganization.String() + ` {id: $org_id})
	OPTIONAL MATCH (u)-[invitation:` + EdgeKindInvitedTo.String() + `]->(o)
	OPTIONAL MATCH (role:` + model.ResourceTypeRole.String() + ` {id: $role_id})
	WITH i, u, o, invitation, role, u.status AS prev
	OPTIONAL MATCH (n:` + model.ResourceTypeUser.String() + ` {status: $active_status})
	WITH i, u, o, invitation, role, prev, count(n) AS active_count
	WITH i, u, o, invitation, role, prev,
		i IS NOT NULL AS installation_found,
		u IS NOT NULL AS user_found,
		o IS NOT NULL AS organization_found,
		invitation IS NOT NULL AS invitation_found,
		($role_id = "" OR role IS NOT NULL) AS role_found,
		CASE WHEN prev IN [$active_status, $pending_status] THEN true ELSE false END AS status_allowed,
		CASE
			WHEN u IS NULL OR o IS NULL OR invitation IS NULL THEN false
			WHEN $role_id <> "" AND role IS NULL THEN false
			WHEN NOT (prev IN [$active_status, $pending_status]) THEN false
			WHEN NOT $unlimited AND i IS NULL THEN false
			WHEN prev = $active_status THEN true
			WHEN NOT $allow_activation THEN false
			WHEN $unlimited THEN true
			ELSE active_count < $limit
		END AS allowed
	FOREACH (_ IN CASE WHEN allowed THEN [1] ELSE [] END |
		SET u.status = $active_status
		SET u.password = CASE WHEN prev = $pending_status THEN $password ELSE u.password END
		SET u.updated_at = datetime($now)
		DELETE invitation
		MERGE (u)-[membership:` + EdgeKindMemberOf.String() + `]->(o)
		ON CREATE SET membership.id = $membership_id, membership.created_at = datetime($now)
		ON MATCH SET membership.updated_at = datetime($now)
	)
	FOREACH (_ IN CASE WHEN allowed AND $role_id <> "" THEN [1] ELSE [] END |
		CREATE (u)-[g:` + EdgeKindGranted.String() + ` {
			id: $grant_id,
			role_id: $role_id,
			actions: [],
			created_at: datetime($now)
		}]->(o)
	)
	RETURN installation_found, user_found, organization_found, invitation_found, role_found, status_allowed, allowed, u,
		COUNT { (u)-[:` + EdgeKindCreated.String() + `]->(:` + model.ResourceTypeDocument.String() + `) } AS document_count
	`

	type acceptResult struct {
		InstallationFound bool
		UserFound         bool
		OrganizationFound bool
		InvitationFound   bool
		RoleFound         bool
		StatusAllowed     bool
		Allowed           bool
		User              *User
	}
	result, err := Neo4jExecuteWriteAndReadSingle(ctx, r.db, cypher, params, func(rec *neo4j.Record) (*acceptResult, error) {
		values := make([]bool, 0, 7)
		for _, field := range []string{
			"installation_found", "user_found", "organization_found",
			"invitation_found", "role_found", "status_allowed", "allowed",
		} {
			value, _, err := neo4j.GetRecordValue[bool](rec, field)
			if err != nil {
				return nil, err
			}
			values = append(values, value)
		}
		result := &acceptResult{
			InstallationFound: values[0],
			UserFound:         values[1],
			OrganizationFound: values[2],
			InvitationFound:   values[3],
			RoleFound:         values[4],
			StatusAllowed:     values[5],
			Allowed:           values[6],
		}
		if result.Allowed {
			user, scanErr := r.scan("u", UserDetailProjection())(rec)
			if scanErr != nil {
				return nil, scanErr
			}
			result.User = user
		}
		return result, nil
	})
	if err != nil {
		return nil, errors.Join(ErrUserAcceptInvitation, err)
	}
	if allowActivation && !auth.unlimited && !result.InstallationFound {
		return nil, errors.Join(ErrUserAcceptInvitation, ErrInstallationRead, ErrNotFound)
	}
	if !result.UserFound || !result.OrganizationFound || !result.InvitationFound {
		return nil, errors.Join(ErrUserAcceptInvitation, ErrNotFound)
	}
	if !result.RoleFound {
		return nil, errors.Join(ErrUserAcceptInvitation, ErrNotFound)
	}
	if !result.StatusAllowed {
		return nil, errors.Join(ErrUserAcceptInvitation, ErrUserActivationStatus)
	}
	if !result.Allowed {
		if !allowActivation {
			return nil, errors.Join(ErrUserAcceptInvitation, ErrUserActivationStatus)
		}
		return nil, errors.Join(ErrUserAcceptInvitation, ErrSeatLimitReached)
	}

	return result.User, nil
}

func (r *Neo4jUserRepository) ActiveHumanCount(ctx context.Context) (int, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.UserRepository/ActiveHumanCount")
	defer span.End()

	cypher := `MATCH (n:` + model.ResourceTypeUser.String() + ` {status: $status}) RETURN count(n) AS c`
	params := map[string]any{
		"status": model.UserStatusActive.String(),
	}

	count, err := Neo4jExecuteReadAndReadSingle(ctx, r.db, cypher, params, func(rec *neo4j.Record) (*int, error) {
		val, _, err := neo4j.GetRecordValue[int64](rec, "c")
		if err != nil {
			return nil, err
		}
		n := int(val)
		return &n, nil
	})
	if err != nil {
		return 0, errors.Join(ErrReadResourceCount, err)
	}

	return *count, nil
}

// NewNeo4jUserRepository creates a new user neo4jBaseRepository.
func NewNeo4jUserRepository(opts ...Neo4jRepositoryOption) (*Neo4jUserRepository, error) {
	baseRepo, err := newNeo4jRepository(opts...)
	if err != nil {
		return nil, err
	}

	return &Neo4jUserRepository{
		neo4jBaseRepository: baseRepo,
	}, nil
}

func clearUsersPattern(ctx context.Context, r *redisBaseRepository, pattern ...string) error {
	return r.DeletePattern(ctx, composeCacheKey(model.ResourceTypeUser.String(), pattern))
}

func clearUsersKey(ctx context.Context, r *redisBaseRepository, id model.ID) error {
	return clearUsersPattern(ctx, r, "Get", id.String(), "*")
}

func clearUsersByEmail(ctx context.Context, r *redisBaseRepository, email string) error {
	return clearUsersPattern(ctx, r, "GetByEmail", email, "*")
}

func clearUsersAllByEmail(ctx context.Context, r *redisBaseRepository) error {
	return clearUsersPattern(ctx, r, "GetByEmail", "*")
}

func clearUserAll(ctx context.Context, r *redisBaseRepository) error {
	return clearUsersPattern(ctx, r, "List", "*", "*", "*")
}

func clearUserAllCrossCache(ctx context.Context, r *redisBaseRepository) error {
	deleteFns := []func(context.Context, *redisBaseRepository, ...string) error{
		clearOrganizationsPattern,
		clearRolesPattern,
	}

	for _, fn := range deleteFns {
		if err := fn(ctx, r, "*"); err != nil {
			return err
		}
	}

	return nil
}

// RedisCachedUserRepository implements caching on the UserRepository.
type RedisCachedUserRepository struct {
	cacheRepo *redisBaseRepository
	userRepo  UserRepository
}

func (r *RedisCachedUserRepository) Create(ctx context.Context, opts CreateUserOpts) (*User, error) {
	if err := clearUserAll(ctx, r.cacheRepo); err != nil {
		return nil, err
	}
	if err := clearUserAllCrossCache(ctx, r.cacheRepo); err != nil {
		return nil, err
	}

	return r.userRepo.Create(ctx, opts)
}

func (r *RedisCachedUserRepository) Activate(ctx context.Context, id model.ID, opts UpdateUserOpts, auth ActivationAuthorization) (*User, error) {
	user, err := r.userRepo.Activate(ctx, id, opts, auth)
	if err != nil {
		return nil, err
	}

	r.invalidateAfterWrite(ctx, "failed to invalidate cache after user activation",
		func() error {
			key := composeCacheKey(model.ResourceTypeUser.String(), "Get", id.String(), projectionCacheValue(UserDetailProjection()))
			return r.cacheRepo.Set(ctx, key, user)
		},
		func() error { return clearUsersByEmail(ctx, r.cacheRepo, user.Email) },
		func() error { return clearUserAll(ctx, r.cacheRepo) },
		func() error { return bumpIssueListUserGeneration(ctx, r.cacheRepo, id) },
		func() error { return bumpIssueListProjectionEpoch(ctx, r.cacheRepo) },
	)

	return user, nil
}

func (r *RedisCachedUserRepository) AcceptInvitation(
	ctx context.Context,
	userID, orgID model.ID,
	hashedPassword string,
	activation *ActivationAuthorization,
	roleID *model.ID,
) (*User, error) {
	user, err := r.userRepo.AcceptInvitation(ctx, userID, orgID, hashedPassword, activation, roleID)
	if err != nil {
		return nil, err
	}

	r.invalidateAfterWrite(ctx, "failed to invalidate cache after accepting invitation",
		func() error { return clearUsersKey(ctx, r.cacheRepo, userID) },
		func() error { return clearUsersAllByEmail(ctx, r.cacheRepo) },
		func() error { return clearUserAll(ctx, r.cacheRepo) },
		func() error { return clearOrganizationsPattern(ctx, r.cacheRepo, "*") },
		func() error { return bumpIssueListUserGeneration(ctx, r.cacheRepo, userID) },
		func() error { return bumpIssueListProjectionEpoch(ctx, r.cacheRepo) },
		func() error { return clearPermissionAllCrossCache(ctx, r.cacheRepo) },
		func() error { return r.bumpAuthzGeneration(ctx, userID) },
	)

	return user, nil
}

func (r *RedisCachedUserRepository) bumpAuthzGeneration(ctx context.Context, principal model.ID) error {
	key := authzGenKey(principal)
	var gen int64
	_ = r.cacheRepo.Get(ctx, key, &gen)
	return r.cacheRepo.Set(ctx, key, gen+1)
}

func (r *RedisCachedUserRepository) invalidateAfterWrite(ctx context.Context, message string, invalidations ...func() error) {
	for _, invalidate := range invalidations {
		if err := invalidate(); err != nil {
			r.cacheRepo.logger.Warn(ctx, message, log.WithError(err))
		}
	}
}

func (r *RedisCachedUserRepository) ActiveHumanCount(ctx context.Context) (int, error) {
	return r.userRepo.ActiveHumanCount(ctx)
}

func (r *RedisCachedUserRepository) Get(ctx context.Context, id model.ID, proj UserProjection) (*User, error) {
	var user *User
	var err error

	key := composeCacheKey(model.ResourceTypeUser.String(), "Get", id.String(), projectionCacheValue(proj))
	if err = r.cacheRepo.Get(ctx, key, &user); err != nil {
		return nil, err
	}

	if user != nil {
		return user, nil
	}

	if user, err = r.userRepo.Get(ctx, id, proj); err != nil {
		return nil, err
	}

	if err = r.cacheRepo.Set(ctx, key, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (r *RedisCachedUserRepository) GetByEmail(ctx context.Context, email string, proj UserProjection) (*User, error) {
	var user *User
	var err error

	key := composeCacheKey(model.ResourceTypeUser.String(), "GetByEmail", email, projectionCacheValue(proj))
	if err = r.cacheRepo.Get(ctx, key, &user); err != nil {
		return nil, err
	}

	if user != nil {
		return user, nil
	}

	if user, err = r.userRepo.GetByEmail(ctx, email, proj); err != nil {
		return nil, err
	}

	if err = r.cacheRepo.Set(ctx, key, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (r *RedisCachedUserRepository) List(ctx context.Context, page CursorPage, proj UserProjection) (Page[*User], error) {
	var users Page[*User]
	var err error

	normalized, err := normalizedPage(page)
	if err != nil {
		return Page[*User]{}, err
	}

	key := composeCacheKey(model.ResourceTypeUser.String(), "List", projectionCacheValue(proj), pageTokenValue(normalized.Token), normalized.Size)
	if err = r.cacheRepo.Get(ctx, key, &users); err != nil {
		return Page[*User]{}, err
	}

	if users.Items != nil {
		return users, nil
	}

	if users, err = r.userRepo.List(ctx, normalized, proj); err != nil {
		return Page[*User]{}, err
	}

	if err = r.cacheRepo.Set(ctx, key, users); err != nil {
		return Page[*User]{}, err
	}

	return users, nil
}

func (r *RedisCachedUserRepository) Update(ctx context.Context, id model.ID, opts UpdateUserOpts) (*User, error) {
	user, err := r.userRepo.Update(ctx, id, opts)
	if err != nil {
		return nil, err
	}

	key := composeCacheKey(model.ResourceTypeUser.String(), "Get", id.String(), projectionCacheValue(UserDetailProjection()))
	if err = r.cacheRepo.Set(ctx, key, user); err != nil {
		return nil, err
	}

	if err = clearUsersByEmail(ctx, r.cacheRepo, user.Email); err != nil {
		return nil, err
	}

	if err = clearUserAll(ctx, r.cacheRepo); err != nil {
		return nil, err
	}
	if err = bumpIssueListUserGeneration(ctx, r.cacheRepo, id); err != nil {
		return nil, err
	}
	if err = bumpIssueListProjectionEpoch(ctx, r.cacheRepo); err != nil {
		return nil, err
	}

	return user, nil
}

func (r *RedisCachedUserRepository) Delete(ctx context.Context, id model.ID) error {
	if err := clearUsersKey(ctx, r.cacheRepo, id); err != nil {
		return err
	}

	if err := clearUsersAllByEmail(ctx, r.cacheRepo); err != nil {
		return err
	}

	if err := clearUserAll(ctx, r.cacheRepo); err != nil {
		return err
	}

	if err := clearUserAllCrossCache(ctx, r.cacheRepo); err != nil {
		return err
	}
	if err := bumpIssueListUserGeneration(ctx, r.cacheRepo, id); err != nil {
		return err
	}
	if err := bumpIssueListProjectionEpoch(ctx, r.cacheRepo); err != nil {
		return err
	}

	return r.userRepo.Delete(ctx, id)
}

// NewCachedUserRepository returns a new CachedUserRepository.
func NewCachedUserRepository(repo UserRepository, opts ...RedisRepositoryOption) (*RedisCachedUserRepository, error) {
	r, err := newRedisBaseRepository(opts...)
	if err != nil {
		return nil, err
	}

	return &RedisCachedUserRepository{
		cacheRepo: r,
		userRepo:  repo,
	}, nil
}

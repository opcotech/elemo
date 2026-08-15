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
	ErrUserCreate = errors.New("failed to create user") // user cannot be created
	ErrUserDelete = errors.New("failed to delete user") // user cannot be deleted
	ErrUserRead   = errors.New("failed to read user")   // user cannot be read
	ErrUserUpdate = errors.New("failed to update user") // user cannot be updated
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

//go:generate go tool mockgen -source=user.go -destination=user_mock_gen.go -package=repository -mock_names "UserRepository=MockUserRepository"
type UserRepository interface {
	Create(ctx context.Context, opts CreateUserOpts) (*User, error)
	Get(ctx context.Context, id model.ID, proj UserProjection) (*User, error)
	GetByEmail(ctx context.Context, email string, proj UserProjection) (*User, error)
	List(ctx context.Context, page CursorPage, proj UserProjection) (Page[*User], error)
	Update(ctx context.Context, id model.ID, opts UpdateUserOpts) (*User, error)
	Delete(ctx context.Context, id model.ID) error
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

// Create creates a new user if it does not already exist.
func (r *Neo4jUserRepository) Create(ctx context.Context, opts CreateUserOpts) (*User, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.UserRepository/Create")
	defer span.End()

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

	cypher := `
	MERGE (u:` + id.Label() + ` {id: $id})
	ON CREATE SET u += {
		username: $username, email: $email, password: $password, status: $status, first_name: $first_name,
		last_name: $last_name, picture: $picture, title: $title, bio: $bio, phone: $phone, address: $address,
		links: $links, languages: $languages, created_at: datetime($created_at)
	}`

	params := map[string]any{
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

	if err := Neo4jExecuteWriteAndConsume(ctx, r.db, cypher, params); err != nil {
		return nil, errors.Join(err, ErrUserCreate)
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

	cypher := `
	MATCH (u:` + id.Label() + ` {id: $id})
	SET u += $patch, u.updated_at = datetime()
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
		return errors.Join(err, ErrUserDelete)
	}

	return nil
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

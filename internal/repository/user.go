package repository

import (
	"context"
	"errors"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"

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

// User represents a user persisted by the repository.
type User struct {
	ID          model.ID         `json:"id"`
	Username    string           `json:"username"`
	Email       string           `json:"email"`
	Password    string           `json:"password"`
	Status      model.UserStatus `json:"status"`
	FirstName   string           `json:"first_name"`
	LastName    string           `json:"last_name"`
	Picture     string           `json:"picture"`
	Title       string           `json:"title"`
	Bio         string           `json:"bio"`
	Phone       string           `json:"phone"`
	Address     string           `json:"address"`
	Links       []string         `json:"links"`
	Languages   []model.Language `json:"languages"`
	Documents   []model.ID       `json:"documents"`
	Permissions []model.ID       `json:"permissions"`
	CreatedAt   *time.Time       `json:"created_at"`
	UpdatedAt   *time.Time       `json:"updated_at"`
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

//go:generate mockgen -source=user.go -destination=user_mock_gen.go -package=repository -mock_names "UserRepository=MockUserRepository"
type UserRepository interface {
	Create(ctx context.Context, opts CreateUserOpts) (*User, error)
	Get(ctx context.Context, id model.ID) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetAll(ctx context.Context, offset, limit int) ([]*User, error)
	Update(ctx context.Context, id model.ID, opts UpdateUserOpts) (*User, error)
	Delete(ctx context.Context, id model.ID) error
}

const (
	languageIDType = "Language" // label for language nodes
)

// Neo4jUserRepository is a repository for managing users.
type Neo4jUserRepository struct {
	*neo4jBaseRepository
}

// scan is a helper function for scanning a user from a Neo4j Record.
func (r *Neo4jUserRepository) scan(up, pp, dp string) func(rec *neo4j.Record) (*User, error) {
	return func(rec *neo4j.Record) (*User, error) {
		user := new(User)
		user.Links = make([]string, 0)

		val, _, err := neo4j.GetRecordValue[neo4j.Node](rec, up)
		if err != nil {
			return nil, err
		}

		if err := Neo4jScanIntoStruct(&val, &user, []string{"id"}); err != nil {
			return nil, err
		}

		user.ID, _ = model.NewIDFromString(val.GetProperties()["id"].(string), model.ResourceTypeUser.String())

		if user.Permissions, err = Neo4jParseIDsFromRecord(rec, pp, model.ResourceTypePermission.String()); err != nil {
			return nil, err
		}

		if user.Documents, err = Neo4jParseIDsFromRecord(rec, dp, model.ResourceTypeDocument.String()); err != nil {
			return nil, err
		}

		return user, nil
	}
}

// Create creates a new user if it does not already exist.
func (r *Neo4jUserRepository) Create(ctx context.Context, opts CreateUserOpts) (*User, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.UserRepository/Create")
	defer span.End()

	createdAt := time.Now().UTC()

	user := &User{
		ID:          model.MustNewID(model.ResourceTypeUser),
		Username:    opts.Username,
		Email:       opts.Email,
		Password:    opts.Password,
		Status:      opts.Status,
		FirstName:   opts.FirstName,
		LastName:    opts.LastName,
		Picture:     opts.Picture,
		Title:       opts.Title,
		Bio:         opts.Bio,
		Phone:       opts.Phone,
		Address:     opts.Address,
		Links:       opts.Links,
		Languages:   opts.Languages,
		Documents:   make([]model.ID, 0),
		Permissions: make([]model.ID, 0),
		CreatedAt:   convert.ToPointer(createdAt),
		UpdatedAt:   nil,
	}

	if user.Links == nil {
		user.Links = make([]string, 0)
	}
	if user.Languages == nil {
		user.Languages = make([]model.Language, 0)
	}
	if user.Status == 0 {
		user.Status = model.UserStatusActive
	}

	languages := make([]string, len(user.Languages))
	for i, l := range user.Languages {
		languages[i] = l.String()
	}

	cypher := `
	MERGE (u:` + user.ID.Label() + ` {id: $id})
	ON CREATE SET u += {
		username: $username, email: $email, password: $password, status: $status, first_name: $first_name,
		last_name: $last_name, picture: $picture, title: $title, bio: $bio, phone: $phone, address: $address,
		links: $links, languages: $languages, created_at: datetime($created_at)
	}`

	params := map[string]any{
		"id":         user.ID.String(),
		"username":   user.Username,
		"email":      user.Email,
		"password":   user.Password,
		"status":     user.Status.String(),
		"first_name": user.FirstName,
		"last_name":  user.LastName,
		"picture":    user.Picture,
		"title":      user.Title,
		"bio":        user.Bio,
		"phone":      user.Phone,
		"address":    user.Address,
		"links":      user.Links,
		"languages":  languages,
		"created_at": createdAt.Format(time.RFC3339Nano),
	}

	if err := Neo4jExecuteWriteAndConsume(ctx, r.db, cypher, params); err != nil {
		return nil, errors.Join(err, ErrUserCreate)
	}

	return user, nil
}

// Get returns a user by its ID.
func (r *Neo4jUserRepository) Get(ctx context.Context, id model.ID) (*User, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.UserRepository/Get")
	defer span.End()

	cypher := `MATCH (u:` + model.ResourceTypeUser.String() + ` {id: $id})
	OPTIONAL MATCH (u)-[p:` + EdgeKindHasPermission.String() + `]->()
	OPTIONAL MATCH (u)<-[r:` + EdgeKindBelongsTo.String() + `]-(d:` + model.ResourceTypeDocument.String() + `)
	RETURN u, collect(DISTINCT p.id) AS p, collect(DISTINCT d.id) AS d`

	params := map[string]any{
		"id": id.String(),
	}

	user, err := Neo4jExecuteReadAndReadSingle(ctx, r.db, cypher, params, r.scan("u", "p", "d"))
	if err != nil {
		if errors.As(err, &ErrNoMoreRecords) {
			return nil, errors.Join(ErrUserRead, ErrNotFound)
		}
		return nil, errors.Join(ErrUserRead, err)
	}

	return user, nil
}

// GetByEmail returns a user by its email.
func (r *Neo4jUserRepository) GetByEmail(ctx context.Context, email string) (*User, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.UserRepository/GetByEmail")
	defer span.End()

	cypher := `MATCH (u:` + model.ResourceTypeUser.String() + ` {email: $email})
	OPTIONAL MATCH (u)-[p:` + EdgeKindHasPermission.String() + `]->()
	OPTIONAL MATCH (u)<-[r:` + EdgeKindBelongsTo.String() + `]-(d:` + model.ResourceTypeDocument.String() + `)
	RETURN u, collect(DISTINCT p.id) AS p, collect(DISTINCT d.id) AS d`

	params := map[string]any{
		"email": email,
	}

	user, err := Neo4jExecuteReadAndReadSingle(ctx, r.db, cypher, params, r.scan("u", "p", "d"))
	if err != nil {
		if errors.As(err, &ErrNoMoreRecords) {
			return nil, errors.Join(ErrUserRead, ErrNotFound)
		}
		return nil, errors.Join(ErrUserRead, err)
	}

	return user, nil
}

// GetAll returns all users respecting the given offset and limit.
func (r *Neo4jUserRepository) GetAll(ctx context.Context, offset, limit int) ([]*User, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.UserRepository/GetAllBelongsTo")
	defer span.End()

	cypher := `
	MATCH (u:` + model.ResourceTypeUser.String() + `)
	OPTIONAL MATCH (u)-[p:` + EdgeKindHasPermission.String() + `]->()
	OPTIONAL MATCH (u)<-[r:` + EdgeKindBelongsTo.String() + `]-(d:` + model.ResourceTypeDocument.String() + `)
	RETURN u, collect(DISTINCT p.id) AS p, collect(DISTINCT d.id) AS d
	ORDER BY u.created_at DESC
	SKIP $offset LIMIT $limit`

	params := map[string]any{
		"offset": offset,
		"limit":  limit,
	}

	users, err := Neo4jExecuteWriteAndReadAll(ctx, r.db, cypher, params, r.scan("u", "p", "d"))
	if err != nil {
		if errors.As(err, &ErrNoMoreRecords) {
			return nil, errors.Join(ErrUserRead, ErrNotFound)
		}
		return nil, errors.Join(ErrUserRead, err)
	}

	return users, nil
}

// Update updates a user by its ID with any given opts.
func (r *Neo4jUserRepository) Update(ctx context.Context, id model.ID, opts UpdateUserOpts) (*User, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.UserRepository/Update")
	defer span.End()

	cypher := `
	MATCH (u:` + id.Label() + ` {id: $id})
	SET u += $patch, u.updated_at = datetime()
	WITH u
	OPTIONAL MATCH (u)-[p:` + EdgeKindHasPermission.String() + `]->()
	OPTIONAL MATCH (u)<-[r:` + EdgeKindBelongsTo.String() + `]-(d:` + model.ResourceTypeDocument.String() + `)
	RETURN u, collect(DISTINCT p.id) AS p, collect(DISTINCT d.id) AS d
	`
	params := map[string]any{
		"id":    id.String(),
		"patch": opts.patch(),
	}

	updated, err := Neo4jExecuteWriteAndReadSingle(ctx, r.db, cypher, params, r.scan("u", "p", "d"))
	if err != nil {
		if errors.As(err, &ErrNoMoreRecords) {
			return nil, errors.Join(ErrUserRead, ErrNotFound)
		}
		return nil, errors.Join(ErrUserUpdate, err)
	}

	return updated, nil
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
	return r.Delete(ctx, composeCacheKey(model.ResourceTypeUser.String(), id.String()))
}

func clearUsersByEmail(ctx context.Context, r *redisBaseRepository, email string) error {
	return r.Delete(ctx, composeCacheKey(model.ResourceTypeUser.String(), "GetByEmail", email))
}

func clearUsersAllByEmail(ctx context.Context, r *redisBaseRepository) error {
	return clearUsersPattern(ctx, r, "GetByEmail", "*")
}

func clearUserAll(ctx context.Context, r *redisBaseRepository) error {
	return clearUsersPattern(ctx, r, "GetAll", "*")
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

func (r *RedisCachedUserRepository) Get(ctx context.Context, id model.ID) (*User, error) {
	var user *User
	var err error

	key := composeCacheKey(model.ResourceTypeUser.String(), id.String())
	if err = r.cacheRepo.Get(ctx, key, &user); err != nil {
		return nil, err
	}

	if user != nil {
		return user, nil
	}

	if user, err = r.userRepo.Get(ctx, id); err != nil {
		return nil, err
	}

	if err = r.cacheRepo.Set(ctx, key, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (r *RedisCachedUserRepository) GetByEmail(ctx context.Context, email string) (*User, error) {
	var user *User
	var err error

	key := composeCacheKey(model.ResourceTypeUser.String(), "GetByEmail", email)
	if err = r.cacheRepo.Get(ctx, key, &user); err != nil {
		return nil, err
	}

	if user != nil {
		return user, nil
	}

	if user, err = r.userRepo.GetByEmail(ctx, email); err != nil {
		return nil, err
	}

	if err = r.cacheRepo.Set(ctx, key, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (r *RedisCachedUserRepository) GetAll(ctx context.Context, offset, limit int) ([]*User, error) {
	var users []*User
	var err error

	key := composeCacheKey(model.ResourceTypeUser.String(), "GetAll", offset, limit)
	if err = r.cacheRepo.Get(ctx, key, &users); err != nil {
		return nil, err
	}

	if users != nil {
		return users, nil
	}

	if users, err = r.userRepo.GetAll(ctx, offset, limit); err != nil {
		return nil, err
	}

	if err = r.cacheRepo.Set(ctx, key, users); err != nil {
		return nil, err
	}

	return users, nil
}

func (r *RedisCachedUserRepository) Update(ctx context.Context, id model.ID, opts UpdateUserOpts) (*User, error) {
	user, err := r.userRepo.Update(ctx, id, opts)
	if err != nil {
		return nil, err
	}

	key := composeCacheKey(model.ResourceTypeUser.String(), id.String())
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

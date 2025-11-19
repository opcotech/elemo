package repository

import (
	"context"
	"errors"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg/convert"
)

var (
	ErrUserCreate = errors.New("failed to create user") // user cannot be created
	ErrUserDelete = errors.New("failed to delete user") // user cannot be deleted
	ErrUserRead   = errors.New("failed to read user")   // user cannot be read
	ErrUserUpdate = errors.New("failed to update user") // user cannot be updated
)

//go:generate mockgen -source=user.go -destination=../testutil/mock/user_repo_gen.go -package=mock -mock_names "UserRepository=UserRepository"
type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	Get(ctx context.Context, id model.ID) (*model.User, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	GetAll(ctx context.Context, offset, limit int) ([]*model.User, error)
	Update(ctx context.Context, id model.ID, patch map[string]any) (*model.User, error)
	Delete(ctx context.Context, id model.ID) error
}

const (
	languageIDType = "Language" // label for language nodes
)

// UserRepository is a repository for managing users.
type Neo4jUserRepository struct {
	*neo4jBaseRepository
}

// scan is a helper function for scanning a user from a Neo4j Record.
func (r *Neo4jUserRepository) scan(up, pp, dp string) func(rec *neo4j.Record) (*model.User, error) {
	return func(rec *neo4j.Record) (*model.User, error) {
		user := new(model.User)
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

		if err := user.Validate(); err != nil {
			return nil, err
		}

		return user, nil
	}
}

// Create creates a new user if it does not already exist. Also, create all
// missing languages and user-language relationships.
func (r *Neo4jUserRepository) Create(ctx context.Context, user *model.User) error {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.UserRepository/Create")
	defer span.End()

	if err := user.Validate(); err != nil {
		return errors.Join(ErrUserCreate, err)
	}

	createdAt := time.Now().UTC()

	user.ID = model.MustNewID(model.ResourceTypeUser)
	user.CreatedAt = convert.ToPointer(createdAt)
	user.UpdatedAt = nil

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
		return errors.Join(err, ErrUserCreate)
	}

	return nil
}

// Get returns a user by its ID.
func (r *Neo4jUserRepository) Get(ctx context.Context, id model.ID) (*model.User, error) {
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
func (r *Neo4jUserRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
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
func (r *Neo4jUserRepository) GetAll(ctx context.Context, offset, limit int) ([]*model.User, error) {
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

// Update updates a user by its ID with any given patch.
func (r *Neo4jUserRepository) Update(ctx context.Context, id model.ID, patch map[string]any) (*model.User, error) {
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
		"patch": patch,
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

// CachedUserRepository implements caching on the
// repository.UserRepository.
type RedisCachedUserRepository struct {
	cacheRepo *redisBaseRepository
	userRepo  UserRepository
}

func (r *RedisCachedUserRepository) Create(ctx context.Context, user *model.User) error {
	if err := clearUserAll(ctx, r.cacheRepo); err != nil {
		return err
	}
	if err := clearUserAllCrossCache(ctx, r.cacheRepo); err != nil {
		return err
	}

	return r.userRepo.Create(ctx, user)
}

func (r *RedisCachedUserRepository) Get(ctx context.Context, id model.ID) (*model.User, error) {
	var user *model.User
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

func (r *RedisCachedUserRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	var user *model.User
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

func (r *RedisCachedUserRepository) GetAll(ctx context.Context, offset, limit int) ([]*model.User, error) {
	var users []*model.User
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

func (r *RedisCachedUserRepository) Update(ctx context.Context, id model.ID, patch map[string]any) (*model.User, error) {
	var user *model.User
	var err error

	user, err = r.userRepo.Update(ctx, id, patch)
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

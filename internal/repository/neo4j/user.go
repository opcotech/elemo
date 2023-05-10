package neo4j

import (
	"context"
	"errors"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg/convert"
	"github.com/opcotech/elemo/internal/repository"
)

const (
	languageIDType = "Language" // label for language nodes
)

// UserRepository is a repository for managing users.
type UserRepository struct {
	*baseRepository
}

// scan is a helper function for scanning a user from a Neo4j Record.
func (r *UserRepository) scan(up, pp, dp string) func(rec *neo4j.Record) (*model.User, error) {
	return func(rec *neo4j.Record) (*model.User, error) {
		user := new(model.User)
		user.Links = make([]string, 0)

		val, _, err := neo4j.GetRecordValue[neo4j.Node](rec, up)
		if err != nil {
			return nil, err
		}

		if err := ScanIntoStruct(&val, &user, []string{"id"}); err != nil {
			return nil, err
		}

		user.ID, _ = model.NewIDFromString(val.GetProperties()["id"].(string), model.ResourceTypeUser.String())

		if user.Permissions, err = ParseIDsFromRecord(rec, pp, model.ResourceTypePermission.String()); err != nil {
			return nil, err
		}

		if user.Documents, err = ParseIDsFromRecord(rec, dp, model.ResourceTypeDocument.String()); err != nil {
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
func (r *UserRepository) Create(ctx context.Context, user *model.User) error {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.UserRepository/Create")
	defer span.End()

	if err := user.Validate(); err != nil {
		return errors.Join(repository.ErrUserCreate, err)
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

	if err := ExecuteWriteAndConsume(ctx, r.db, cypher, params); err != nil {
		return errors.Join(err, repository.ErrUserCreate)
	}

	return nil
}

// Get returns a user by its ID.
func (r *UserRepository) Get(ctx context.Context, id model.ID) (*model.User, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.UserRepository/Get")
	defer span.End()

	cypher := `MATCH (u:` + model.ResourceTypeUser.String() + ` {id: $id})
	OPTIONAL MATCH (u)-[p:` + EdgeKindHasPermission.String() + `]->()
	OPTIONAL MATCH (u)<-[r:` + EdgeKindBelongsTo.String() + `]-(d:` + model.ResourceTypeDocument.String() + `)
	RETURN u, collect(DISTINCT p.id) AS p, collect(DISTINCT d.id) AS d`

	params := map[string]any{
		"id": id.String(),
	}

	user, err := ExecuteReadAndReadSingle(ctx, r.db, cypher, params, r.scan("u", "p", "d"))
	if err != nil {
		if errors.As(err, &ErrNoMoreRecords) {
			return nil, errors.Join(repository.ErrUserRead, repository.ErrNotFound)
		}
		return nil, errors.Join(repository.ErrUserRead, err)
	}

	return user, nil
}

// GetByEmail returns a user by its email.
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.UserRepository/GetByEmail")
	defer span.End()

	cypher := `MATCH (u:` + model.ResourceTypeUser.String() + ` {email: $email})
	OPTIONAL MATCH (u)-[p:` + EdgeKindHasPermission.String() + `]->()
	OPTIONAL MATCH (u)<-[r:` + EdgeKindBelongsTo.String() + `]-(d:` + model.ResourceTypeDocument.String() + `)
	RETURN u, collect(DISTINCT p.id) AS p, collect(DISTINCT d.id) AS d`

	params := map[string]any{
		"email": email,
	}

	user, err := ExecuteReadAndReadSingle(ctx, r.db, cypher, params, r.scan("u", "p", "d"))
	if err != nil {
		if errors.As(err, &ErrNoMoreRecords) {
			return nil, errors.Join(repository.ErrUserRead, repository.ErrNotFound)
		}
		return nil, errors.Join(repository.ErrUserRead, err)
	}

	return user, nil
}

// GetAll returns all users respecting the given offset and limit.
func (r *UserRepository) GetAll(ctx context.Context, offset, limit int) ([]*model.User, error) {
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

	users, err := ExecuteWriteAndReadAll(ctx, r.db, cypher, params, r.scan("u", "p", "d"))
	if err != nil {
		if errors.As(err, &ErrNoMoreRecords) {
			return nil, errors.Join(repository.ErrUserRead, repository.ErrNotFound)
		}
		return nil, errors.Join(repository.ErrUserRead, err)
	}

	return users, nil
}

// Update updates a user by its ID with any given patch.
func (r *UserRepository) Update(ctx context.Context, id model.ID, patch map[string]any) (*model.User, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.UserRepository/Update")
	defer span.End()

	cypher := `
	MATCH (u:` + id.Label() + ` {id: $id})
	SET u += $patch, u.updated_at = datetime($updated_at)
	WITH u
	OPTIONAL MATCH (u)-[p:` + EdgeKindHasPermission.String() + `]->()
	OPTIONAL MATCH (u)<-[r:` + EdgeKindBelongsTo.String() + `]-(d:` + model.ResourceTypeDocument.String() + `)
	RETURN u, collect(DISTINCT p.id) AS p, collect(DISTINCT d.id) AS d
	`
	params := map[string]any{
		"id":         id.String(),
		"patch":      patch,
		"updated_at": time.Now().UTC().Format(time.RFC3339Nano),
	}

	updated, err := ExecuteWriteAndReadSingle(ctx, r.db, cypher, params, r.scan("u", "p", "d"))
	if err != nil {
		if errors.As(err, &ErrNoMoreRecords) {
			return nil, errors.Join(repository.ErrUserRead, repository.ErrNotFound)
		}
		return nil, errors.Join(repository.ErrUserUpdate, err)
	}

	return updated, nil
}

// Delete deletes a user by its ID.
func (r *UserRepository) Delete(ctx context.Context, id model.ID) error {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.UserRepository/Delete")
	defer span.End()

	cypher := `MATCH (u:` + id.Label() + ` {id: $id}) DETACH DELETE u`
	params := map[string]any{
		"id": id.String(),
	}

	if err := ExecuteWriteAndConsume(ctx, r.db, cypher, params); err != nil {
		return errors.Join(err, repository.ErrUserDelete)
	}

	return nil
}

// NewUserRepository creates a new user baseRepository.
func NewUserRepository(opts ...RepositoryOption) (*UserRepository, error) {
	baseRepo, err := newRepository(opts...)
	if err != nil {
		return nil, err
	}

	return &UserRepository{
		baseRepository: baseRepo,
	}, nil
}

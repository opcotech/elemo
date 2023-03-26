package neo4j

import (
	"context"
	"errors"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg/convert"
)

// User language is a special case _scenario where we need to create a
// relationship between a user and a language, but the language has no ID.
// Therefore, we must keep the label name and relationship type in a constant
// so that we can use it in the cypher query.
// TODO: This should be changed and the Language should have an ID too, so we
//
//	can use the same pattern as the other nodes, and eliminate this shit.
const (
	languageIDType = "Language" // label for language nodes
)

var (
	ErrUserCreate = errors.New("failed to create user") // user cannot be created
	ErrUserRead   = errors.New("failed to read user")   // user cannot be read
	ErrUserUpdate = errors.New("failed to update user") // user cannot be updated
	ErrUserDelete = errors.New("failed to delete user") // user cannot be deleted
)

// UserRepository is a repository for managing users.
type UserRepository struct {
	*repository
}

// scan is a helper function for scanning a user from a Neo4j Record.
func (r *UserRepository) scan(up, lp, pp, dp string) func(rec *neo4j.Record) (*model.User, error) {
	return func(rec *neo4j.Record) (*model.User, error) {
		user := new(model.User)

		val, _, err := neo4j.GetRecordValue[neo4j.Node](rec, up)
		if err != nil {
			return nil, err
		}

		if err := ScanIntoStruct(&val, &user, []string{"id"}); err != nil {
			return nil, err
		}

		user.ID, _ = model.NewIDFromString(val.GetProperties()["id"].(string), model.UserIDType)

		ln, err := ParseValueFromRecord[[]any](rec, lp)
		if err != nil {
			return nil, err
		}

		user.Languages = make([]model.Language, len(ln))
		for i, n := range ln {
			lc := model.Language(0)
			if err := lc.UnmarshalText([]byte(n.(string))); err != nil {
				return nil, err
			}

			user.Languages[i] = lc
		}

		if user.Permissions, err = ParseIDsFromRecord(rec, pp, EdgeKindHasPermission.String()); err != nil {
			return nil, err
		}

		if user.Documents, err = ParseIDsFromRecord(rec, dp, model.DocumentIDType); err != nil {
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
		return errors.Join(ErrUserCreate, err)
	}

	createdAt := time.Now()

	user.ID = model.MustNewID(model.UserIDType)
	user.CreatedAt = convert.ToPointer(createdAt)
	user.UpdatedAt = nil

	languages := make([]string, len(user.Languages))
	for i, l := range user.Languages {
		languages[i] = l.String()
	}

	cypher := `
	MERGE (u:` + user.ID.Label() + ` {id: $id})
	ON CREATE SET u += { username: $username, email: $email, password: $password, status: $status,
		first_name: $first_name, last_name: $last_name, picture: $picture, title: $title, bio: $bio, phone: $phone,
		address: $address, links: $links, created_at: datetime($created_at) }
	WITH u, $languages AS languages
	UNWIND $languages AS language
	MERGE (l:` + languageIDType + ` { code: language }) CREATE (u)-[:` + EdgeKindSpeaks.String() + ` {created_at: datetime($created_at)}]->(l)
	`

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
		return errors.Join(err, ErrUserCreate)
	}

	return nil
}

// Get returns a user by its ID.
func (r *UserRepository) Get(ctx context.Context, id model.ID) (*model.User, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.UserRepository/Get")
	defer span.End()

	cypher := `MATCH (u:` + model.UserIDType + ` {id: toString($id)})
	OPTIONAL MATCH (u)-[:` + EdgeKindSpeaks.String() + `]->(l:` + languageIDType + `)
	OPTIONAL MATCH (u)-[p:` + EdgeKindHasPermission.String() + `]->()
	OPTIONAL MATCH (u)-[r:` + EdgeKindCreated.String() + `]->(d:` + model.DocumentIDType + `)
	RETURN u, collect(DISTINCT l.code) AS l, collect(DISTINCT p.id) AS p, collect(DISTINCT d.id) AS d`

	params := map[string]any{
		"id": id.String(),
	}

	user, err := ExecuteReadAndReadSingle(ctx, r.db, cypher, params, r.scan("u", "l", "p", "d"))
	if err != nil {
		return nil, errors.Join(ErrUserRead, err)
	}

	return user, nil
}

// GetAll returns all users respecting the given offset and limit.
func (r *UserRepository) GetAll(ctx context.Context, offset, limit int) ([]*model.User, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.UserRepository/GetAllBelongsTo")
	defer span.End()

	cypher := `
	MATCH (u:` + model.UserIDType + `)
	OPTIONAL MATCH (u)-[:` + EdgeKindSpeaks.String() + `]->(l:` + languageIDType + `)
	OPTIONAL MATCH (u)-[p:` + EdgeKindHasPermission.String() + `]->()
	OPTIONAL MATCH (u)-[r:` + EdgeKindCreated.String() + `]->(d:` + model.DocumentIDType + `)
	RETURN u, collect(DISTINCT l.code) AS l, collect(DISTINCT p.id) AS p, collect(DISTINCT d.id) AS d
	ORDER BY u.created_at DESC
	SKIP $offset LIMIT $limit`

	params := map[string]any{
		"offset": offset,
		"limit":  limit,
	}

	users, err := ExecuteWriteAndReadAll(ctx, r.db, cypher, params, r.scan("u", "l", "p", "d"))
	if err != nil {
		return nil, errors.Join(ErrUserRead, err)
	}

	return users, nil
}

// Update updates a user by its ID with any given patch. The patch data is not
// typed, so it is up to the caller to ensure that the patch data is valid.
// This behaviour is intentional, as it allows the caller to extend the user
// with additional data if needed, however, that data will not be returned if
// the user model has not been extended.
//
// The patch data cannot be used to update relationships. To update a relation
// use the dedicated repository.
func (r *UserRepository) Update(ctx context.Context, id model.ID, patch map[string]any) (*model.User, error) {
	ctx, span := r.tracer.Start(ctx, "repository.neo4j.UserRepository/Update")
	defer span.End()

	cypher := `MATCH (u:` + id.Label() + ` {id: $id}) SET u += $patch SET u.updated_at = datetime($updated_at)`
	params := map[string]any{
		"id":         id.String(),
		"patch":      patch,
		"updated_at": time.Now().Format(time.RFC3339Nano),
	}

	if languages, ok := patch["languages"]; ok {
		cypher += `
		WITH u, $languages AS languages
		UNWIND languages AS language
		MATCH (u:` + id.Label() + ` {id: $id})
		MERGE (l:` + languageIDType + ` {code: language})
		MERGE (u)-[nl:` + EdgeKindSpeaks.String() + `]->(l) ON CREATE SET nl.created_at = datetime($created_at)
		WITH u, languages
		MATCH (u)-[r:` + EdgeKindSpeaks.String() + `]->(l:` + languageIDType + `) WHERE NOT l.code IN languages
		DELETE r`

		delete(patch, "languages")
		params["languages"] = languages
		params["created_at"] = time.Now().Format(time.RFC3339Nano)
	}

	params["patch"] = patch

	// Update the user
	if err := ExecuteWriteAndConsume(ctx, r.db, cypher, params); err != nil {
		return nil, errors.Join(ErrUserUpdate, err)
	}

	// Delete dangling languages
	deleteDanglingCypher := `
	MATCH (l:` + languageIDType + `) WHERE NOT (l)<-[:` + EdgeKindSpeaks.String() + `]-()
	DETACH DELETE l`

	if err := ExecuteWriteAndConsume(ctx, r.db, deleteDanglingCypher, nil); err != nil {
		return nil, errors.Join(ErrUserUpdate, err)
	}

	// Return the updated user and its languages
	getUpdatedCypher := `
	MATCH (u:` + id.Label() + ` {id: $id})-[:` + EdgeKindSpeaks.String() + `]->(l:` + languageIDType + `)
	OPTIONAL MATCH (u)-[p:` + EdgeKindHasPermission.String() + `]->()
	OPTIONAL MATCH (u)-[r:` + EdgeKindCreated.String() + `]->(d:` + model.DocumentIDType + `)
	RETURN u, collect(DISTINCT l.code) AS l, collect(DISTINCT p.id) AS p, collect(DISTINCT d.id) AS d`

	getUpdatedParams := map[string]any{"id": id.String()}
	updated, err := ExecuteWriteAndReadSingle(ctx, r.db, getUpdatedCypher, getUpdatedParams, r.scan("u", "l", "p", "d"))
	if err != nil {
		return nil, errors.Join(ErrUserUpdate, err)
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
		return errors.Join(err, ErrUserDelete)
	}

	return nil
}

// NewUserRepository creates a new user repository.
func NewUserRepository(opts ...RepositoryOption) (*UserRepository, error) {
	baseRepo, err := newRepository(opts...)
	if err != nil {
		return nil, err
	}

	return &UserRepository{
		repository: baseRepo,
	}, nil
}

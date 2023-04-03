//go:build integration

package repository

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/repository/neo4j"
	"github.com/opcotech/elemo/internal/testutil"
)

func prepareUser(t *testing.T) *model.User {
	username := strings.ToLower(testutil.GenerateRandomString(10))
	email := testutil.GenerateEmail(10)

	user, err := model.NewUser(username, email, "AppleTree")
	require.NoError(t, err)

	user.Languages = []model.Language{model.LanguageHU, model.LanguageEN, model.LanguageAR}
	user.Address = "1234 Main St, Anytown, USA"
	user.Bio = "I am a test user"
	user.FirstName = "Test"
	user.LastName = "User"
	user.Links = []string{"https://example.com/"}
	user.Picture = "https://www.gravatar.com/avatar"
	user.Phone = "+1234567890"
	user.Title = "Senior Test User"

	return user
}

func TestUserRepository_Create(t *testing.T) {
	ctx := context.Background()

	db, closer := testutil.NewNeo4jDatabase(t, neo4jDBConf)
	defer func(ctx context.Context, closer func(ctx context.Context) error) {
		require.NoError(t, closer(ctx))
	}(ctx, closer)

	defer testutil.CleanupNeo4jStore(t, ctx, db)

	repo, err := neo4j.NewUserRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

	err = repo.Create(ctx, prepareUser(t))
	require.NoError(t, err)
}

func TestUserRepository_Get(t *testing.T) {
	ctx := context.Background()

	db, closer := testutil.NewNeo4jDatabase(t, neo4jDBConf)
	defer func(ctx context.Context, closer func(ctx context.Context) error) {
		require.NoError(t, closer(ctx))
	}(ctx, closer)

	defer testutil.CleanupNeo4jStore(t, ctx, db)

	repo, err := neo4j.NewUserRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

	organizationID := model.MustNewID(model.OrganizationIDType)
	_, err = db.GetWriteSession(ctx).Run(ctx,
		"CREATE (:"+organizationID.Label()+" {id: $organization_id})",
		map[string]any{
			"organization_id": organizationID.String(),
		},
	)
	require.NoError(t, err)

	user := prepareUser(t)
	err = repo.Create(ctx, user)
	require.NoError(t, err)

	permRepo, err := neo4j.NewPermissionRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

	readPerm, err := model.NewPermission(user.ID, organizationID, model.PermissionKindRead)
	require.NoError(t, err)

	writePerm, err := model.NewPermission(user.ID, organizationID, model.PermissionKindWrite)
	require.NoError(t, err)

	err = permRepo.Create(ctx, readPerm)
	require.NoError(t, err)

	err = permRepo.Create(ctx, writePerm)
	require.NoError(t, err)

	got, err := repo.Get(ctx, user.ID)
	require.NoError(t, err)

	assert.Equal(t, user.ID, got.ID)
	assert.Equal(t, user.Username, got.Username)
	assert.Equal(t, user.Email, got.Email)
	assert.Equal(t, user.Password, got.Password)
	assert.Equal(t, user.Status, got.Status)
	assert.Equal(t, user.FirstName, got.FirstName)
	assert.Equal(t, user.LastName, got.LastName)
	assert.Equal(t, user.Picture, got.Picture)
	assert.Equal(t, user.Title, got.Title)
	assert.Equal(t, user.Bio, got.Bio)
	assert.Equal(t, user.Phone, got.Phone)
	assert.Equal(t, user.Address, got.Address)
	assert.Equal(t, user.Links, got.Links)
	assert.ElementsMatch(t, user.Languages, got.Languages)
	assert.WithinDuration(t, *user.CreatedAt, *got.CreatedAt, 0)
	assert.Nil(t, got.UpdatedAt)

	permIDs := []model.ID{readPerm.ID, writePerm.ID}
	for _, perm := range got.Permissions {
		assert.Contains(t, permIDs, perm)
	}

	assert.Equal(t, 0, len(got.Documents))
}

func TestUserRepository_GetAll(t *testing.T) {
	ctx := context.Background()

	db, closer := testutil.NewNeo4jDatabase(t, neo4jDBConf)
	defer func(ctx context.Context, closer func(ctx context.Context) error) {
		require.NoError(t, closer(ctx))
	}(ctx, closer)

	defer testutil.CleanupNeo4jStore(t, ctx, db)
	testutil.CleanupNeo4jStore(t, ctx, db)

	repo, err := neo4j.NewUserRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

	user := prepareUser(t)
	err = repo.Create(ctx, user)
	require.NoError(t, err)

	user2 := prepareUser(t)
	err = repo.Create(ctx, user2)
	require.NoError(t, err)

	user3 := prepareUser(t)
	err = repo.Create(ctx, user3)
	require.NoError(t, err)

	got, err := repo.GetAll(ctx, 0, 3)
	require.NoError(t, err)
	assert.Len(t, got, 3)

	got, err = repo.GetAll(ctx, 0, 2)
	require.NoError(t, err)
	assert.Len(t, got, 2)

	got, err = repo.GetAll(ctx, 0, 1)
	require.NoError(t, err)
	assert.Len(t, got, 1)

	got, err = repo.GetAll(ctx, 0, 0)
	require.NoError(t, err)
	assert.Len(t, got, 0)

	got, err = repo.GetAll(ctx, 1, 3)
	require.NoError(t, err)
	assert.Len(t, got, 2)

	got, err = repo.GetAll(ctx, 2, 3)
	require.NoError(t, err)
	assert.Len(t, got, 1)

	got, err = repo.GetAll(ctx, 3, 3)
	require.NoError(t, err)
	assert.Len(t, got, 0)
}

func TestUserRepository_Update(t *testing.T) {
	ctx := context.Background()

	db, closer := testutil.NewNeo4jDatabase(t, neo4jDBConf)
	defer func(ctx context.Context, closer func(ctx context.Context) error) {
		require.NoError(t, closer(ctx))
	}(ctx, closer)

	defer testutil.CleanupNeo4jStore(t, ctx, db)

	repo, err := neo4j.NewUserRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

	user := prepareUser(t)
	err = repo.Create(ctx, user)
	require.NoError(t, err)

	other := prepareUser(t)
	err = repo.Create(ctx, other)
	require.NoError(t, err)

	patch := map[string]any{
		"email":    "info@example.com",
		"username": "new username",
		"languages": []string{
			model.LanguageHU.String(),
			model.LanguageEN.String(),
			model.LanguageAR.String(),
		},
	}

	updated, err := repo.Update(ctx, user.ID, patch)
	require.NoError(t, err)

	assert.Equal(t, user.ID, updated.ID)
	assert.Equal(t, patch["username"], updated.Username)
	assert.Equal(t, patch["email"], updated.Email)
	assert.Equal(t, user.Password, updated.Password)
	assert.Equal(t, user.Status, updated.Status)
	assert.Equal(t, user.FirstName, updated.FirstName)
	assert.Equal(t, user.LastName, updated.LastName)
	assert.Equal(t, user.Picture, updated.Picture)
	assert.Equal(t, user.Title, updated.Title)
	assert.Equal(t, user.Bio, updated.Bio)
	assert.Equal(t, user.Phone, updated.Phone)
	assert.Equal(t, user.Address, updated.Address)
	assert.WithinDuration(t, *user.CreatedAt, *updated.CreatedAt, 0)
	assert.NotEqual(t, user.UpdatedAt, updated.UpdatedAt)
	assert.ElementsMatch(t, user.Links, updated.Links)

	// Check if the languages are updated
	assert.ElementsMatch(t, []model.Language{
		model.LanguageHU,
		model.LanguageEN,
		model.LanguageAR,
	}, updated.Languages)

	// Check if the other user is not affected by the update
	otherGot, err := repo.Get(ctx, other.ID)
	require.NoError(t, err)
	assert.Equal(t, other.ID, otherGot.ID)
}

func TestUserRepository_Delete(t *testing.T) {
	ctx := context.Background()

	db, closer := testutil.NewNeo4jDatabase(t, neo4jDBConf)
	defer func(ctx context.Context, closer func(ctx context.Context) error) {
		require.NoError(t, closer(ctx))
	}(ctx, closer)

	defer testutil.CleanupNeo4jStore(t, ctx, db)

	repo, err := neo4j.NewUserRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

	user := prepareUser(t)
	err = repo.Create(ctx, user)
	require.NoError(t, err)

	got, err := repo.Get(ctx, user.ID)
	require.NoError(t, err)
	assert.Equal(t, user.ID, got.ID)

	err = repo.Delete(ctx, user.ID)
	require.NoError(t, err)

	_, err = repo.Get(ctx, user.ID)
	require.Error(t, err)
}

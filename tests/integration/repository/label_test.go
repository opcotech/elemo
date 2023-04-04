//go:build integration

package repository

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/repository/neo4j"
	"github.com/opcotech/elemo/internal/testutil"
	testRepo "github.com/opcotech/elemo/internal/testutil/repository"
)

func TestLabelRepository_Create(t *testing.T) {
	ctx := context.Background()

	db, closer := testRepo.NewNeo4jDatabase(t, neo4jDBConf)
	defer func(ctx context.Context, closer func(ctx context.Context) error) {
		require.NoError(t, closer(ctx))
	}(ctx, closer)

	defer testRepo.CleanupNeo4jStore(t, ctx, db)

	labelRepo, err := neo4j.NewLabelRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

	label, err := model.NewLabel(testutil.GenerateRandomString(10))
	require.NoError(t, err)

	err = labelRepo.Create(ctx, label)
	require.NoError(t, err)

	assert.NotEqual(t, model.ID{}, label.ID)
	assert.NotNil(t, label.CreatedAt)
	assert.Nil(t, label.UpdatedAt)
}

func TestLabelRepository_Get(t *testing.T) {
	ctx := context.Background()

	db, closer := testRepo.NewNeo4jDatabase(t, neo4jDBConf)
	defer func(ctx context.Context, closer func(ctx context.Context) error) {
		require.NoError(t, closer(ctx))
	}(ctx, closer)

	defer testRepo.CleanupNeo4jStore(t, ctx, db)

	labelRepo, err := neo4j.NewLabelRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

	label, err := model.NewLabel(testutil.GenerateRandomString(10))
	require.NoError(t, err)

	err = labelRepo.Create(ctx, label)
	require.NoError(t, err)

	got, err := labelRepo.Get(ctx, label.ID)
	require.NoError(t, err)

	assert.Equal(t, label.ID, got.ID)
	assert.Equal(t, label.Name, got.Name)
	assert.Equal(t, label.Description, got.Description)
	assert.WithinDuration(t, *label.CreatedAt, *got.CreatedAt, 1*time.Second)
	assert.Nil(t, got.UpdatedAt)
}

func TestLabelRepository_Update(t *testing.T) {
	ctx := context.Background()

	db, closer := testRepo.NewNeo4jDatabase(t, neo4jDBConf)
	defer func(ctx context.Context, closer func(ctx context.Context) error) {
		require.NoError(t, closer(ctx))
	}(ctx, closer)

	defer testRepo.CleanupNeo4jStore(t, ctx, db)

	labelRepo, err := neo4j.NewLabelRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

	label, err := model.NewLabel(testutil.GenerateRandomString(10))
	require.NoError(t, err)

	err = labelRepo.Create(ctx, label)
	require.NoError(t, err)

	patch := map[string]any{
		"name":        testutil.GenerateRandomString(10),
		"description": testutil.GenerateRandomString(10),
	}

	got, err := labelRepo.Update(ctx, label.ID, patch)
	require.NoError(t, err)

	assert.Equal(t, label.ID, got.ID)
	assert.Equal(t, patch["name"], got.Name)
	assert.Equal(t, patch["description"], got.Description)
	assert.WithinDuration(t, *label.CreatedAt, *got.CreatedAt, 1*time.Second)
	assert.NotNil(t, got.UpdatedAt)
	assert.WithinDuration(t, time.Now(), *got.UpdatedAt, 1*time.Second)
}

func TestLabelRepository_AttachTo(t *testing.T) {
	ctx := context.Background()

	db, closer := testRepo.NewNeo4jDatabase(t, neo4jDBConf)
	defer func(ctx context.Context, closer func(ctx context.Context) error) {
		require.NoError(t, closer(ctx))
	}(ctx, closer)

	defer testRepo.CleanupNeo4jStore(t, ctx, db)

	labelRepo, err := neo4j.NewLabelRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

	userRepo, err := neo4j.NewUserRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

	orgRepo, err := neo4j.NewOrganizationRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

	documentRepo, err := neo4j.NewDocumentRepository(
		neo4j.WithDatabase(db),
	)

	user := prepareUser(t)
	err = userRepo.Create(ctx, user)
	require.NoError(t, err)

	organization := prepareOrganization(t)
	err = orgRepo.Create(ctx, user.ID, organization)
	require.NoError(t, err)

	document := prepareDocument(t, user.ID)
	err = documentRepo.Create(ctx, organization.ID, document)
	require.NoError(t, err)

	label, err := model.NewLabel(testutil.GenerateRandomString(10))
	require.NoError(t, err)

	err = labelRepo.Create(ctx, label)
	require.NoError(t, err)

	err = labelRepo.AttachTo(ctx, label.ID, document.ID)
	require.NoError(t, err)
}

func TestLabelRepository_DetachFrom(t *testing.T) {
	ctx := context.Background()

	db, closer := testRepo.NewNeo4jDatabase(t, neo4jDBConf)
	defer func(ctx context.Context, closer func(ctx context.Context) error) {
		require.NoError(t, closer(ctx))
	}(ctx, closer)

	defer testRepo.CleanupNeo4jStore(t, ctx, db)

	labelRepo, err := neo4j.NewLabelRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

	userRepo, err := neo4j.NewUserRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

	orgRepo, err := neo4j.NewOrganizationRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

	documentRepo, err := neo4j.NewDocumentRepository(
		neo4j.WithDatabase(db),
	)

	user := prepareUser(t)
	err = userRepo.Create(ctx, user)
	require.NoError(t, err)

	organization := prepareOrganization(t)
	err = orgRepo.Create(ctx, user.ID, organization)
	require.NoError(t, err)

	document := prepareDocument(t, user.ID)
	err = documentRepo.Create(ctx, organization.ID, document)
	require.NoError(t, err)

	label, err := model.NewLabel(testutil.GenerateRandomString(10))
	require.NoError(t, err)

	err = labelRepo.Create(ctx, label)
	require.NoError(t, err)

	err = labelRepo.AttachTo(ctx, label.ID, document.ID)
	require.NoError(t, err)

	err = labelRepo.DetachFrom(ctx, label.ID, document.ID)
	require.NoError(t, err)
}

func TestLabelRepository_Delete(t *testing.T) {
	ctx := context.Background()

	db, closer := testRepo.NewNeo4jDatabase(t, neo4jDBConf)
	defer func(ctx context.Context, closer func(ctx context.Context) error) {
		require.NoError(t, closer(ctx))
	}(ctx, closer)

	defer testRepo.CleanupNeo4jStore(t, ctx, db)

	labelRepo, err := neo4j.NewLabelRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

	label, err := model.NewLabel(testutil.GenerateRandomString(10))
	require.NoError(t, err)

	err = labelRepo.Create(ctx, label)
	require.NoError(t, err)

	err = labelRepo.Delete(ctx, label.ID)
	require.NoError(t, err)

	_, err = labelRepo.Get(ctx, label.ID)
	assert.Error(t, err)
}

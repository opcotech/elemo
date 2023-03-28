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
)

func TestCommentRepository_Create(t *testing.T) {
	ctx := context.Background()

	db, closer := newNeo4jDatabase(t)
	defer func(ctx context.Context, closer func(ctx context.Context) error) {
		require.NoError(t, closer(ctx))
	}(ctx, closer)

	defer cleanupNeo4jStore(t, ctx, db)

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

	commentRepo, err := neo4j.NewCommentRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

	user := prepareUser(t)
	err = userRepo.Create(ctx, user)
	require.NoError(t, err)

	organization := prepareOrganization(t)
	err = orgRepo.Create(ctx, user.ID, organization)
	require.NoError(t, err)

	document := prepareDocument(t, user.ID)
	err = documentRepo.Create(ctx, organization.ID, document)
	require.NoError(t, err)

	comment, err := model.NewComment("this is a test comment from a user", user.ID)
	require.NoError(t, err)

	err = commentRepo.Create(ctx, document.ID, comment)
	require.NoError(t, err)

	assert.NotEqual(t, model.ID{}, comment.ID)
	assert.NotNil(t, comment.CreatedAt)
	assert.Nil(t, comment.UpdatedAt)
}

func TestCommentRepository_Get(t *testing.T) {
	ctx := context.Background()

	db, closer := newNeo4jDatabase(t)
	defer func(ctx context.Context, closer func(ctx context.Context) error) {
		require.NoError(t, closer(ctx))
	}(ctx, closer)

	defer cleanupNeo4jStore(t, ctx, db)

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

	commentRepo, err := neo4j.NewCommentRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

	user := prepareUser(t)
	err = userRepo.Create(ctx, user)
	require.NoError(t, err)

	organization := prepareOrganization(t)
	err = orgRepo.Create(ctx, user.ID, organization)
	require.NoError(t, err)

	document := prepareDocument(t, user.ID)
	err = documentRepo.Create(ctx, organization.ID, document)
	require.NoError(t, err)

	comment, err := model.NewComment("this is a test comment from a user", user.ID)
	require.NoError(t, err)

	err = commentRepo.Create(ctx, document.ID, comment)
	require.NoError(t, err)

	got, err := commentRepo.Get(ctx, comment.ID)
	require.NoError(t, err)

	assert.Equal(t, comment.ID, got.ID)
	assert.Equal(t, comment.Content, got.Content)
	assert.WithinDuration(t, *comment.CreatedAt, *got.CreatedAt, 1*time.Second)
	assert.Nil(t, got.UpdatedAt)
}

func TestCommentRepository_GetAllBelongsTo(t *testing.T) {
	ctx := context.Background()

	db, closer := newNeo4jDatabase(t)
	defer func(ctx context.Context, closer func(ctx context.Context) error) {
		require.NoError(t, closer(ctx))
	}(ctx, closer)

	defer cleanupNeo4jStore(t, ctx, db)

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

	commentRepo, err := neo4j.NewCommentRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

	user := prepareUser(t)
	err = userRepo.Create(ctx, user)
	require.NoError(t, err)

	organization := prepareOrganization(t)
	err = orgRepo.Create(ctx, user.ID, organization)
	require.NoError(t, err)

	document := prepareDocument(t, user.ID)
	err = documentRepo.Create(ctx, organization.ID, document)
	require.NoError(t, err)

	comment, err := model.NewComment("this is a test comment from a user", user.ID)
	require.NoError(t, err)

	require.NoError(t, commentRepo.Create(ctx, document.ID, comment))
	require.NoError(t, commentRepo.Create(ctx, document.ID, comment))
	require.NoError(t, commentRepo.Create(ctx, document.ID, comment))

	got, err := commentRepo.GetAllBelongsTo(ctx, document.ID, 0, 10)
	require.NoError(t, err)
	assert.Len(t, got, 3)

	got, err = commentRepo.GetAllBelongsTo(ctx, document.ID, 0, 2)
	require.NoError(t, err)
	assert.Len(t, got, 2)

	got, err = commentRepo.GetAllBelongsTo(ctx, document.ID, 1, 1)
	require.NoError(t, err)
	assert.Len(t, got, 1)

	got, err = commentRepo.GetAllBelongsTo(ctx, document.ID, 2, 2)
	require.NoError(t, err)
	assert.Len(t, got, 1)

	got, err = commentRepo.GetAllBelongsTo(ctx, document.ID, 3, 1)
	require.NoError(t, err)
	assert.Len(t, got, 0)
}

func TestCommentRepository_Update(t *testing.T) {
	ctx := context.Background()

	db, closer := newNeo4jDatabase(t)
	defer func(ctx context.Context, closer func(ctx context.Context) error) {
		require.NoError(t, closer(ctx))
	}(ctx, closer)

	defer cleanupNeo4jStore(t, ctx, db)

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

	commentRepo, err := neo4j.NewCommentRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

	user := prepareUser(t)
	err = userRepo.Create(ctx, user)
	require.NoError(t, err)

	organization := prepareOrganization(t)
	err = orgRepo.Create(ctx, user.ID, organization)
	require.NoError(t, err)

	document := prepareDocument(t, user.ID)
	err = documentRepo.Create(ctx, organization.ID, document)
	require.NoError(t, err)

	comment, err := model.NewComment("this is a test comment from a user", user.ID)
	require.NoError(t, err)

	err = commentRepo.Create(ctx, document.ID, comment)
	require.NoError(t, err)

	patch := map[string]any{
		"content": "this is an updated comment",
	}

	updated, err := commentRepo.Update(ctx, comment.ID, patch)
	require.NoError(t, err)

	assert.Equal(t, comment.ID, updated.ID)
	assert.Equal(t, patch["content"], updated.Content)
	assert.WithinDuration(t, *comment.CreatedAt, *updated.CreatedAt, 1*time.Second)
	assert.NotNil(t, updated.UpdatedAt)
}

func TestCommentRepository_Delete(t *testing.T) {
	ctx := context.Background()

	db, closer := newNeo4jDatabase(t)
	defer func(ctx context.Context, closer func(ctx context.Context) error) {
		require.NoError(t, closer(ctx))
	}(ctx, closer)

	defer cleanupNeo4jStore(t, ctx, db)

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

	commentRepo, err := neo4j.NewCommentRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

	user := prepareUser(t)
	err = userRepo.Create(ctx, user)
	require.NoError(t, err)

	organization := prepareOrganization(t)
	err = orgRepo.Create(ctx, user.ID, organization)
	require.NoError(t, err)

	document := prepareDocument(t, user.ID)
	err = documentRepo.Create(ctx, organization.ID, document)
	require.NoError(t, err)

	comment, err := model.NewComment("this is a test comment from a user", user.ID)
	require.NoError(t, err)

	err = commentRepo.Create(ctx, document.ID, comment)
	require.NoError(t, err)

	err = commentRepo.Delete(ctx, comment.ID)
	require.NoError(t, err)

	_, err = commentRepo.Get(ctx, comment.ID)
	require.Error(t, err)
}

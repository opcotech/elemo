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

func TestAttachmentRepository_Create(t *testing.T) {
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

	attachmentRepo, err := neo4j.NewAttachmentRepository(
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

	attachment, err := model.NewAttachment("test attachment", "file_id", user.ID)
	require.NoError(t, err)

	err = attachmentRepo.Create(ctx, document.ID, attachment)
	require.NoError(t, err)

	assert.NotEqual(t, model.ID{}, attachment.ID)
	assert.NotNil(t, attachment.CreatedAt)
	assert.Nil(t, attachment.UpdatedAt)
}

func TestAttachmentRepository_Get(t *testing.T) {
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

	attachmentRepo, err := neo4j.NewAttachmentRepository(
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

	attachment, err := model.NewAttachment("test attachment", "file_id", user.ID)
	require.NoError(t, err)

	err = attachmentRepo.Create(ctx, document.ID, attachment)
	require.NoError(t, err)

	got, err := attachmentRepo.Get(ctx, attachment.ID)
	require.NoError(t, err)

	assert.Equal(t, attachment.ID, got.ID)
	assert.Equal(t, attachment.Name, got.Name)
	assert.Equal(t, attachment.FileID, got.FileID)
	assert.WithinDuration(t, *attachment.CreatedAt, *got.CreatedAt, 1*time.Second)
	assert.Nil(t, got.UpdatedAt)
}

func TestAttachmentRepository_GetAllBelongsTo(t *testing.T) {
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

	attachmentRepo, err := neo4j.NewAttachmentRepository(
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

	attachment, err := model.NewAttachment("test attachment", "file_id", user.ID)
	require.NoError(t, err)

	require.NoError(t, attachmentRepo.Create(ctx, document.ID, attachment))
	require.NoError(t, attachmentRepo.Create(ctx, document.ID, attachment))
	require.NoError(t, attachmentRepo.Create(ctx, document.ID, attachment))

	got, err := attachmentRepo.GetAllBelongsTo(ctx, document.ID, 0, 10)
	require.NoError(t, err)
	assert.Len(t, got, 3)

	got, err = attachmentRepo.GetAllBelongsTo(ctx, document.ID, 0, 2)
	require.NoError(t, err)
	assert.Len(t, got, 2)

	got, err = attachmentRepo.GetAllBelongsTo(ctx, document.ID, 1, 1)
	require.NoError(t, err)
	assert.Len(t, got, 1)

	got, err = attachmentRepo.GetAllBelongsTo(ctx, document.ID, 2, 2)
	require.NoError(t, err)
	assert.Len(t, got, 1)

	got, err = attachmentRepo.GetAllBelongsTo(ctx, document.ID, 3, 1)
	require.NoError(t, err)
	assert.Len(t, got, 0)
}

func TestAttachmentRepository_Update(t *testing.T) {
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

	attachmentRepo, err := neo4j.NewAttachmentRepository(
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

	attachment, err := model.NewAttachment("test attachment", "file_id", user.ID)
	require.NoError(t, err)

	err = attachmentRepo.Create(ctx, document.ID, attachment)
	require.NoError(t, err)

	newName := "test name"
	updated, err := attachmentRepo.Update(ctx, attachment.ID, newName)
	require.NoError(t, err)

	assert.Equal(t, attachment.ID, updated.ID)
	assert.Equal(t, newName, updated.Name)
	assert.WithinDuration(t, *attachment.CreatedAt, *updated.CreatedAt, 1*time.Second)
	assert.NotNil(t, updated.UpdatedAt)
}

func TestAttachmentRepository_Delete(t *testing.T) {
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

	attachmentRepo, err := neo4j.NewAttachmentRepository(
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

	attachment, err := model.NewAttachment("test attachment", "file_id", user.ID)
	require.NoError(t, err)

	err = attachmentRepo.Create(ctx, document.ID, attachment)
	require.NoError(t, err)

	err = attachmentRepo.Delete(ctx, attachment.ID)
	require.NoError(t, err)

	_, err = attachmentRepo.Get(ctx, attachment.ID)
	require.Error(t, err)
}

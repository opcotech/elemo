//go:build integration

package repository

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/repository"
	"github.com/opcotech/elemo/internal/repository/neo4j"
	"github.com/opcotech/elemo/internal/testutil"
	testRepo "github.com/opcotech/elemo/internal/testutil/repository"
)

func prepareDocument(t *testing.T, createdBy model.ID) *model.Document {
	document, err := model.NewDocument(testutil.GenerateRandomString(10), "file_id", createdBy)
	require.NoError(t, err)

	document.Excerpt = testutil.GenerateRandomString(10)

	return document
}

func TestDocumentRepository_Create(t *testing.T) {
	ctx := context.Background()

	db, closer := testRepo.NewNeo4jDatabase(t, neo4jDBConf)
	defer func(ctx context.Context, closer func(ctx context.Context) error) {
		require.NoError(t, closer(ctx))
	}(ctx, closer)

	defer testRepo.CleanupNeo4jStore(t, ctx, db)

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
}

func TestDocumentRepository_Get(t *testing.T) {
	ctx := context.Background()

	db, closer := testRepo.NewNeo4jDatabase(t, neo4jDBConf)
	defer func(ctx context.Context, closer func(ctx context.Context) error) {
		require.NoError(t, closer(ctx))
	}(ctx, closer)

	defer testRepo.CleanupNeo4jStore(t, ctx, db)

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

	labelRepo, err := neo4j.NewLabelRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

	commentRepo, err := neo4j.NewCommentRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

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

	label, err := model.NewLabel("label")
	require.NoError(t, err)

	err = labelRepo.Create(ctx, label)
	require.NoError(t, err)

	err = labelRepo.AttachTo(ctx, label.ID, document.ID)
	require.NoError(t, err)

	comment, err := model.NewComment("comment", user.ID)
	require.NoError(t, err)

	err = commentRepo.Create(ctx, document.ID, comment)
	require.NoError(t, err)

	attachment, err := model.NewAttachment("attachment", "file_id", user.ID)
	require.NoError(t, err)

	err = attachmentRepo.Create(ctx, document.ID, attachment)
	require.NoError(t, err)

	got, err := documentRepo.Get(ctx, document.ID)
	require.NoError(t, err)

	assert.Equal(t, document.ID, got.ID)
	assert.Equal(t, document.Name, got.Name)
	assert.Equal(t, document.Excerpt, got.Excerpt)
	assert.Equal(t, document.FileID, got.FileID)
	assert.Equal(t, document.CreatedBy, got.CreatedBy)
	assert.ElementsMatch(t, []model.ID{label.ID}, got.Labels)
	assert.ElementsMatch(t, []model.ID{comment.ID}, got.Comments)
	assert.ElementsMatch(t, []model.ID{attachment.ID}, got.Attachments)
	assert.WithinDuration(t, *document.CreatedAt, *got.CreatedAt, 1*time.Second)
	assert.Nil(t, got.UpdatedAt)
}

func TestDocumentRepository_GetByCreator(t *testing.T) {
	ctx := context.Background()

	db, closer := testRepo.NewNeo4jDatabase(t, neo4jDBConf)
	defer func(ctx context.Context, closer func(ctx context.Context) error) {
		require.NoError(t, closer(ctx))
	}(ctx, closer)

	defer testRepo.CleanupNeo4jStore(t, ctx, db)
	testRepo.CleanupNeo4jStore(t, ctx, db)

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

	user2 := prepareUser(t)
	err = userRepo.Create(ctx, user2)
	require.NoError(t, err)

	organization := prepareOrganization(t)
	err = orgRepo.Create(ctx, user.ID, organization)
	require.NoError(t, err)

	err = documentRepo.Create(ctx, organization.ID, prepareDocument(t, user.ID))
	require.NoError(t, err)

	err = documentRepo.Create(ctx, organization.ID, prepareDocument(t, user.ID))
	require.NoError(t, err)

	err = documentRepo.Create(ctx, organization.ID, prepareDocument(t, user.ID))
	require.NoError(t, err)

	err = documentRepo.Create(ctx, organization.ID, prepareDocument(t, user2.ID))
	require.NoError(t, err)

	got, err := documentRepo.GetByCreator(ctx, user2.ID, 0, 10)
	require.NoError(t, err)
	assert.Len(t, got, 1)

	got, err = documentRepo.GetByCreator(ctx, user.ID, 0, 10)
	require.NoError(t, err)
	assert.Len(t, got, 3)

	got, err = documentRepo.GetByCreator(ctx, user.ID, 1, 10)
	require.NoError(t, err)
	assert.Len(t, got, 2)

	got, err = documentRepo.GetByCreator(ctx, user.ID, 2, 10)
	require.NoError(t, err)
	assert.Len(t, got, 1)

	got, err = documentRepo.GetByCreator(ctx, user.ID, 3, 10)
	require.NoError(t, err)
	assert.Len(t, got, 0)

	got, err = documentRepo.GetByCreator(ctx, user.ID, 0, 2)
	require.NoError(t, err)
	assert.Len(t, got, 2)
}

func TestDocumentRepository_GetAllBelongsTo(t *testing.T) {
	ctx := context.Background()

	db, closer := testRepo.NewNeo4jDatabase(t, neo4jDBConf)
	defer func(ctx context.Context, closer func(ctx context.Context) error) {
		require.NoError(t, closer(ctx))
	}(ctx, closer)

	defer testRepo.CleanupNeo4jStore(t, ctx, db)
	testRepo.CleanupNeo4jStore(t, ctx, db)

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

	organization2 := prepareOrganization(t)
	err = orgRepo.Create(ctx, user.ID, organization2)
	require.NoError(t, err)

	err = documentRepo.Create(ctx, organization.ID, prepareDocument(t, user.ID))
	require.NoError(t, err)

	err = documentRepo.Create(ctx, organization.ID, prepareDocument(t, user.ID))
	require.NoError(t, err)

	err = documentRepo.Create(ctx, organization.ID, prepareDocument(t, user.ID))
	require.NoError(t, err)

	err = documentRepo.Create(ctx, organization2.ID, prepareDocument(t, user.ID))
	require.NoError(t, err)

	documents, err := documentRepo.GetAllBelongsTo(ctx, organization2.ID, 0, 10)
	require.NoError(t, err)
	assert.Len(t, documents, 1)

	documents, err = documentRepo.GetAllBelongsTo(ctx, organization.ID, 0, 10)
	require.NoError(t, err)
	assert.Len(t, documents, 3)

	documents, err = documentRepo.GetAllBelongsTo(ctx, organization.ID, 1, 3)
	require.NoError(t, err)
	assert.Len(t, documents, 2)

	documents, err = documentRepo.GetAllBelongsTo(ctx, organization.ID, 2, 3)
	require.NoError(t, err)
	assert.Len(t, documents, 1)

	documents, err = documentRepo.GetAllBelongsTo(ctx, organization.ID, 3, 3)
	require.NoError(t, err)
	assert.Len(t, documents, 0)
}

func TestDocumentRepository_Update(t *testing.T) {
	ctx := context.Background()

	db, closer := testRepo.NewNeo4jDatabase(t, neo4jDBConf)
	defer func(ctx context.Context, closer func(ctx context.Context) error) {
		require.NoError(t, closer(ctx))
	}(ctx, closer)

	defer testRepo.CleanupNeo4jStore(t, ctx, db)

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

	patch := map[string]any{
		"name":    "new name",
		"excerpt": "new excerpt",
	}

	got, err := documentRepo.Update(ctx, document.ID, patch)
	require.NoError(t, err)

	assert.Equal(t, document.ID, got.ID)
	assert.Equal(t, patch["name"], got.Name)
	assert.Equal(t, patch["excerpt"], got.Excerpt)
	assert.WithinDuration(t, *document.CreatedAt, *got.CreatedAt, 1*time.Second)
	assert.NotNil(t, got.UpdatedAt)
}

func TestDocumentRepository_Delete(t *testing.T) {
	ctx := context.Background()

	db, closer := testRepo.NewNeo4jDatabase(t, neo4jDBConf)
	defer func(ctx context.Context, closer func(ctx context.Context) error) {
		require.NoError(t, closer(ctx))
	}(ctx, closer)

	defer testRepo.CleanupNeo4jStore(t, ctx, db)

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

	_, err = documentRepo.Get(ctx, document.ID)
	require.NoError(t, err)

	err = documentRepo.Delete(ctx, document.ID)
	require.NoError(t, err)

	_, err = documentRepo.Get(ctx, document.ID)
	assert.ErrorIs(t, err, repository.ErrNotFound)
}

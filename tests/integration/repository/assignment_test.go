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
	testRepo "github.com/opcotech/elemo/internal/testutil/repository"
)

func TestAssignmentRepository_Create(t *testing.T) {
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

	assignmentRepo, err := neo4j.NewAssignmentRepository(
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

	assignment, err := model.NewAssignment(user.ID, document.ID, model.AssignmentKindReviewer)
	require.NoError(t, err)

	err = assignmentRepo.Create(ctx, assignment)
	require.NoError(t, err)
}

func TestAssignmentRepository_Get(t *testing.T) {
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

	assignmentRepo, err := neo4j.NewAssignmentRepository(
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

	assignment, err := model.NewAssignment(user.ID, document.ID, model.AssignmentKindReviewer)
	require.NoError(t, err)

	err = assignmentRepo.Create(ctx, assignment)
	require.NoError(t, err)

	got, err := assignmentRepo.Get(ctx, assignment.ID)
	require.NoError(t, err)

	assert.Equal(t, assignment.ID, got.ID)
	assert.Equal(t, assignment.User, got.User)
	assert.Equal(t, assignment.Resource, got.Resource)
	assert.Equal(t, assignment.Kind, got.Kind)
	assert.WithinDuration(t, *assignment.CreatedAt, *got.CreatedAt, 1*time.Second)
}

func TestAssignmentRepository_GetByUser(t *testing.T) {
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

	assignmentRepo, err := neo4j.NewAssignmentRepository(
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

	document2 := prepareDocument(t, user.ID)
	err = documentRepo.Create(ctx, organization.ID, document2)
	require.NoError(t, err)

	assignment, err := model.NewAssignment(user.ID, document.ID, model.AssignmentKindReviewer)
	require.NoError(t, err)

	assignment2, err := model.NewAssignment(user.ID, document2.ID, model.AssignmentKindReviewer)
	require.NoError(t, err)

	err = assignmentRepo.Create(ctx, assignment)
	require.NoError(t, err)

	err = assignmentRepo.Create(ctx, assignment2)
	require.NoError(t, err)

	got, err := assignmentRepo.GetByUser(ctx, user.ID, 0, 10)
	require.NoError(t, err)
	assert.Len(t, got, 2)

	got, err = assignmentRepo.GetByUser(ctx, user.ID, 1, 10)
	require.NoError(t, err)
	assert.Len(t, got, 1)

	got, err = assignmentRepo.GetByUser(ctx, user.ID, 2, 10)
	require.NoError(t, err)
	assert.Len(t, got, 0)

	got, err = assignmentRepo.GetByUser(ctx, user.ID, 0, 1)
	require.NoError(t, err)
	assert.Len(t, got, 1)

	got, err = assignmentRepo.GetByUser(ctx, user.ID, 0, 2)
	require.NoError(t, err)
	assert.Len(t, got, 2)
}

func TestAssignmentRepository_GetByResource(t *testing.T) {
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

	assignmentRepo, err := neo4j.NewAssignmentRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

	user := prepareUser(t)
	err = userRepo.Create(ctx, user)
	require.NoError(t, err)

	user2 := prepareUser(t)
	err = userRepo.Create(ctx, user2)
	require.NoError(t, err)

	organization := prepareOrganization(t)
	err = orgRepo.Create(ctx, user.ID, organization)
	require.NoError(t, err)

	document := prepareDocument(t, user.ID)
	err = documentRepo.Create(ctx, organization.ID, document)
	require.NoError(t, err)

	assignment, err := model.NewAssignment(user.ID, document.ID, model.AssignmentKindReviewer)
	require.NoError(t, err)

	assignment2, err := model.NewAssignment(user2.ID, document.ID, model.AssignmentKindReviewer)
	require.NoError(t, err)

	err = assignmentRepo.Create(ctx, assignment)
	require.NoError(t, err)

	err = assignmentRepo.Create(ctx, assignment2)
	require.NoError(t, err)

	got, err := assignmentRepo.GetByResource(ctx, document.ID, 0, 10)
	require.NoError(t, err)
	assert.Len(t, got, 2)

	got, err = assignmentRepo.GetByResource(ctx, document.ID, 1, 10)
	require.NoError(t, err)
	assert.Len(t, got, 1)

	got, err = assignmentRepo.GetByResource(ctx, document.ID, 2, 10)
	require.NoError(t, err)
	assert.Len(t, got, 0)

	got, err = assignmentRepo.GetByResource(ctx, document.ID, 0, 1)
	require.NoError(t, err)
	assert.Len(t, got, 1)

	got, err = assignmentRepo.GetByResource(ctx, document.ID, 0, 2)
	require.NoError(t, err)
	assert.Len(t, got, 2)
}

func TestAssignmentRepository_Delete(t *testing.T) {
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

	assignmentRepo, err := neo4j.NewAssignmentRepository(
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

	assignment, err := model.NewAssignment(user.ID, document.ID, model.AssignmentKindReviewer)
	require.NoError(t, err)

	err = assignmentRepo.Create(ctx, assignment)
	require.NoError(t, err)

	err = assignmentRepo.Delete(ctx, assignment.ID)
	require.NoError(t, err)

	_, err = assignmentRepo.Get(ctx, assignment.ID)
	require.Error(t, err)
}

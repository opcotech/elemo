//go:build integration

package repository

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg/convert"
	"github.com/opcotech/elemo/internal/repository/neo4j"
	"github.com/opcotech/elemo/internal/testutil"
)

func TestIssueRepository_Create(t *testing.T) {
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

	namespaceRepo, err := neo4j.NewNamespaceRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

	projectRepo, err := neo4j.NewProjectRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

	issueRepo, err := neo4j.NewIssueRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

	user := prepareUser(t)
	err = userRepo.Create(ctx, user)
	require.NoError(t, err)

	organization := prepareOrganization(t)
	err = orgRepo.Create(ctx, user.ID, organization)
	require.NoError(t, err)

	namespace, err := model.NewNamespace(testutil.GenerateRandomString(10))
	require.NoError(t, err)

	err = namespaceRepo.Create(ctx, organization.ID, namespace)
	require.NoError(t, err)

	project := prepareProject(t)

	err = projectRepo.Create(ctx, namespace.ID, project)
	require.NoError(t, err)

	issue, err := model.NewIssue(1, "My test epic", model.IssueKindEpic, user.ID)
	require.NoError(t, err)

	err = issueRepo.Create(ctx, project.ID, issue)
	require.NoError(t, err)
}

func TestIssueRepository_Get(t *testing.T) {
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

	namespaceRepo, err := neo4j.NewNamespaceRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

	projectRepo, err := neo4j.NewProjectRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

	assignmentRepo, err := neo4j.NewAssignmentRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

	commentRepo, err := neo4j.NewCommentRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

	labelRepo, err := neo4j.NewLabelRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

	attachmentRepo, err := neo4j.NewAttachmentRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

	issueRepo, err := neo4j.NewIssueRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

	user := prepareUser(t)
	err = userRepo.Create(ctx, user)
	require.NoError(t, err)

	organization := prepareOrganization(t)
	err = orgRepo.Create(ctx, user.ID, organization)
	require.NoError(t, err)

	namespace, err := model.NewNamespace(testutil.GenerateRandomString(10))
	require.NoError(t, err)

	err = namespaceRepo.Create(ctx, organization.ID, namespace)
	require.NoError(t, err)

	project := prepareProject(t)

	err = projectRepo.Create(ctx, namespace.ID, project)
	require.NoError(t, err)

	epic, err := model.NewIssue(1, "My test epic", model.IssueKindEpic, user.ID)
	require.NoError(t, err)

	err = issueRepo.Create(ctx, project.ID, epic)
	require.NoError(t, err)

	issue, err := model.NewIssue(2, "My test issue", model.IssueKindStory, user.ID)
	require.NoError(t, err)

	issue.Parent = convert.ToPointer(epic.ID)
	issue.DueDate = convert.ToPointer(time.Now().AddDate(0, 0, 1))
	issue.Links = []string{
		"https://www.google.com",
	}

	err = issueRepo.Create(ctx, project.ID, issue)
	require.NoError(t, err)

	assignment, err := model.NewAssignment(user.ID, issue.ID, model.AssignmentKindAssignee)
	require.NoError(t, err)

	err = assignmentRepo.Create(ctx, assignment)
	require.NoError(t, err)

	attachment, err := model.NewAttachment("test attachment", "file_id", user.ID)
	require.NoError(t, err)

	err = attachmentRepo.Create(ctx, issue.ID, attachment)
	require.NoError(t, err)

	comment1, err := model.NewComment("this is a test comment from a user", user.ID)
	require.NoError(t, err)

	err = commentRepo.Create(ctx, issue.ID, comment1)
	require.NoError(t, err)

	comment2, err := model.NewComment("and another test comment from a user", user.ID)
	require.NoError(t, err)

	err = commentRepo.Create(ctx, issue.ID, comment2)
	require.NoError(t, err)

	label, err := model.NewLabel(testutil.GenerateRandomString(10))
	require.NoError(t, err)

	err = labelRepo.Create(ctx, label)
	require.NoError(t, err)

	err = labelRepo.AttachTo(ctx, label.ID, issue.ID)
	require.NoError(t, err)

	got, err := issueRepo.Get(ctx, issue.ID)
	require.NoError(t, err)

	assert.Equal(t, issue.ID, got.ID)
	assert.Equal(t, issue.Parent, got.Parent)
	assert.Equal(t, issue.Kind, got.Kind)
	assert.Equal(t, issue.Title, got.Title)
	assert.Equal(t, issue.Description, got.Description)
	assert.Equal(t, issue.Status, got.Status)
	assert.Equal(t, issue.Priority, got.Priority)
	assert.Equal(t, issue.Resolution, got.Resolution)
	assert.Equal(t, issue.ReportedBy, got.ReportedBy)
	assert.ElementsMatch(t, []model.ID{user.ID}, got.Assignees)
	assert.ElementsMatch(t, []model.ID{label.ID}, got.Labels)
	assert.ElementsMatch(t, []model.ID{comment1.ID, comment2.ID}, got.Comments)
	assert.ElementsMatch(t, []model.ID{attachment.ID}, got.Attachments)
	assert.ElementsMatch(t, []model.ID{user.ID}, got.Watchers)
	assert.ElementsMatch(t, []model.ID{epic.ID}, got.Relations)
	assert.ElementsMatch(t, issue.Links, got.Links)
	assert.WithinDuration(t, *issue.DueDate, *got.DueDate, 1*time.Second)
	assert.WithinDuration(t, *issue.CreatedAt, *got.CreatedAt, 1*time.Second)
	assert.Nil(t, got.UpdatedAt)
}

func TestIssueRepository_AddWatcher(t *testing.T) {
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

	namespaceRepo, err := neo4j.NewNamespaceRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

	projectRepo, err := neo4j.NewProjectRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

	issueRepo, err := neo4j.NewIssueRepository(
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

	namespace, err := model.NewNamespace(testutil.GenerateRandomString(10))
	require.NoError(t, err)

	err = namespaceRepo.Create(ctx, organization.ID, namespace)
	require.NoError(t, err)

	project := prepareProject(t)

	err = projectRepo.Create(ctx, namespace.ID, project)
	require.NoError(t, err)

	issue, err := model.NewIssue(1, "My test epic", model.IssueKindEpic, user.ID)
	require.NoError(t, err)

	err = issueRepo.Create(ctx, project.ID, issue)
	require.NoError(t, err)

	err = issueRepo.AddWatcher(ctx, issue.ID, user2.ID)
	require.NoError(t, err)
}

func TestIssueRepository_GetWatchers(t *testing.T) {
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

	namespaceRepo, err := neo4j.NewNamespaceRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

	projectRepo, err := neo4j.NewProjectRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

	issueRepo, err := neo4j.NewIssueRepository(
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

	namespace, err := model.NewNamespace(testutil.GenerateRandomString(10))
	require.NoError(t, err)

	err = namespaceRepo.Create(ctx, organization.ID, namespace)
	require.NoError(t, err)

	project := prepareProject(t)

	err = projectRepo.Create(ctx, namespace.ID, project)
	require.NoError(t, err)

	issue, err := model.NewIssue(1, "My test epic", model.IssueKindEpic, user.ID)
	require.NoError(t, err)

	err = issueRepo.Create(ctx, project.ID, issue)
	require.NoError(t, err)

	err = issueRepo.AddWatcher(ctx, issue.ID, user2.ID)
	require.NoError(t, err)

	watchers, err := issueRepo.GetWatchers(ctx, issue.ID)
	require.NoError(t, err)

	assert.Len(t, watchers, 2)
	assert.ElementsMatch(t, []model.ID{user.ID, user2.ID}, []model.ID{watchers[0].ID, watchers[1].ID})
}

func TestIssueRepository_RemoveWatcher(t *testing.T) {
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

	namespaceRepo, err := neo4j.NewNamespaceRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

	projectRepo, err := neo4j.NewProjectRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

	issueRepo, err := neo4j.NewIssueRepository(
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

	namespace, err := model.NewNamespace(testutil.GenerateRandomString(10))
	require.NoError(t, err)

	err = namespaceRepo.Create(ctx, organization.ID, namespace)
	require.NoError(t, err)

	project := prepareProject(t)

	err = projectRepo.Create(ctx, namespace.ID, project)
	require.NoError(t, err)

	issue, err := model.NewIssue(1, "My test epic", model.IssueKindEpic, user.ID)
	require.NoError(t, err)

	err = issueRepo.Create(ctx, project.ID, issue)
	require.NoError(t, err)

	err = issueRepo.AddWatcher(ctx, issue.ID, user2.ID)
	require.NoError(t, err)

	watchers, err := issueRepo.GetWatchers(ctx, issue.ID)
	require.NoError(t, err)

	assert.Len(t, watchers, 2)
	assert.ElementsMatch(t, []model.ID{user.ID, user2.ID}, []model.ID{watchers[0].ID, watchers[1].ID})

	err = issueRepo.RemoveWatcher(ctx, issue.ID, user2.ID)
	require.NoError(t, err)

	watchers, err = issueRepo.GetWatchers(ctx, issue.ID)
	require.NoError(t, err)

	assert.Len(t, watchers, 1)
	assert.Equal(t, user.ID, watchers[0].ID)
}

func TestIssueRepository_AddRelation(t *testing.T) {
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

	namespaceRepo, err := neo4j.NewNamespaceRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

	projectRepo, err := neo4j.NewProjectRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

	issueRepo, err := neo4j.NewIssueRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

	user := prepareUser(t)
	err = userRepo.Create(ctx, user)
	require.NoError(t, err)

	organization := prepareOrganization(t)
	err = orgRepo.Create(ctx, user.ID, organization)
	require.NoError(t, err)

	namespace, err := model.NewNamespace(testutil.GenerateRandomString(10))
	require.NoError(t, err)

	err = namespaceRepo.Create(ctx, organization.ID, namespace)
	require.NoError(t, err)

	project := prepareProject(t)

	err = projectRepo.Create(ctx, namespace.ID, project)
	require.NoError(t, err)

	issue1, err := model.NewIssue(1, "My test epic", model.IssueKindEpic, user.ID)
	require.NoError(t, err)

	err = issueRepo.Create(ctx, project.ID, issue1)
	require.NoError(t, err)

	issue2, err := model.NewIssue(2, "My test issue", model.IssueKindTask, user.ID)
	require.NoError(t, err)

	err = issueRepo.Create(ctx, project.ID, issue2)
	require.NoError(t, err)

	relation, err := model.NewIssueRelation(issue2.ID, issue1.ID, model.IssueRelationKindSubtaskOf)
	require.NoError(t, err)

	err = issueRepo.AddRelation(ctx, relation)
	require.NoError(t, err)
}

func TestIssueRepository_GetRelations(t *testing.T) {
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

	namespaceRepo, err := neo4j.NewNamespaceRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

	projectRepo, err := neo4j.NewProjectRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

	issueRepo, err := neo4j.NewIssueRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

	user := prepareUser(t)
	err = userRepo.Create(ctx, user)
	require.NoError(t, err)

	organization := prepareOrganization(t)
	err = orgRepo.Create(ctx, user.ID, organization)
	require.NoError(t, err)

	namespace, err := model.NewNamespace(testutil.GenerateRandomString(10))
	require.NoError(t, err)

	err = namespaceRepo.Create(ctx, organization.ID, namespace)
	require.NoError(t, err)

	project := prepareProject(t)

	err = projectRepo.Create(ctx, namespace.ID, project)
	require.NoError(t, err)

	issue1, err := model.NewIssue(1, "My test epic", model.IssueKindEpic, user.ID)
	require.NoError(t, err)

	err = issueRepo.Create(ctx, project.ID, issue1)
	require.NoError(t, err)

	issue2, err := model.NewIssue(2, "My test issue", model.IssueKindTask, user.ID)
	require.NoError(t, err)

	err = issueRepo.Create(ctx, project.ID, issue2)
	require.NoError(t, err)

	relation1, err := model.NewIssueRelation(issue2.ID, issue1.ID, model.IssueRelationKindSubtaskOf)
	require.NoError(t, err)

	err = issueRepo.AddRelation(ctx, relation1)
	require.NoError(t, err)

	relation2, err := model.NewIssueRelation(issue2.ID, issue1.ID, model.IssueRelationKindBlocks)
	require.NoError(t, err)

	err = issueRepo.AddRelation(ctx, relation2)
	require.NoError(t, err)

	relations, err := issueRepo.GetRelations(ctx, issue1.ID)
	require.NoError(t, err)

	require.Len(t, relations, 2)
	assert.Equal(t, []model.ID{relation1.ID, relation2.ID}, []model.ID{relations[0].ID, relations[1].ID})

	relations, err = issueRepo.GetRelations(ctx, issue2.ID)
	require.NoError(t, err)

	require.Len(t, relations, 2)
	assert.Equal(t, []model.ID{relation1.ID, relation2.ID}, []model.ID{relations[0].ID, relations[1].ID})
}

func TestIssueRepository_RemoveRelation(t *testing.T) {
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

	namespaceRepo, err := neo4j.NewNamespaceRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

	projectRepo, err := neo4j.NewProjectRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

	issueRepo, err := neo4j.NewIssueRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

	user := prepareUser(t)
	err = userRepo.Create(ctx, user)
	require.NoError(t, err)

	organization := prepareOrganization(t)
	err = orgRepo.Create(ctx, user.ID, organization)
	require.NoError(t, err)

	namespace, err := model.NewNamespace(testutil.GenerateRandomString(10))
	require.NoError(t, err)

	err = namespaceRepo.Create(ctx, organization.ID, namespace)
	require.NoError(t, err)

	project := prepareProject(t)

	err = projectRepo.Create(ctx, namespace.ID, project)
	require.NoError(t, err)

	issue1, err := model.NewIssue(1, "My test epic", model.IssueKindEpic, user.ID)
	require.NoError(t, err)

	err = issueRepo.Create(ctx, project.ID, issue1)
	require.NoError(t, err)

	issue2, err := model.NewIssue(2, "My test issue", model.IssueKindTask, user.ID)
	require.NoError(t, err)

	err = issueRepo.Create(ctx, project.ID, issue2)
	require.NoError(t, err)

	relation1, err := model.NewIssueRelation(issue2.ID, issue1.ID, model.IssueRelationKindSubtaskOf)
	require.NoError(t, err)

	err = issueRepo.AddRelation(ctx, relation1)
	require.NoError(t, err)

	relation2, err := model.NewIssueRelation(issue2.ID, issue1.ID, model.IssueRelationKindBlocks)
	require.NoError(t, err)

	err = issueRepo.AddRelation(ctx, relation2)
	require.NoError(t, err)

	relations, err := issueRepo.GetRelations(ctx, issue1.ID)
	require.NoError(t, err)
	require.Len(t, relations, 2)

	err = issueRepo.RemoveRelation(ctx, relation1.Source, relation1.Target, relation1.Kind)
	require.NoError(t, err)

	relations, err = issueRepo.GetRelations(ctx, issue1.ID)
	require.NoError(t, err)
	require.Len(t, relations, 1)

}

func TestIssueRepository_Update(t *testing.T) {
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

	namespaceRepo, err := neo4j.NewNamespaceRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

	projectRepo, err := neo4j.NewProjectRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

	issueRepo, err := neo4j.NewIssueRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

	user := prepareUser(t)
	err = userRepo.Create(ctx, user)
	require.NoError(t, err)

	organization := prepareOrganization(t)
	err = orgRepo.Create(ctx, user.ID, organization)
	require.NoError(t, err)

	namespace, err := model.NewNamespace(testutil.GenerateRandomString(10))
	require.NoError(t, err)

	err = namespaceRepo.Create(ctx, organization.ID, namespace)
	require.NoError(t, err)

	project := prepareProject(t)

	err = projectRepo.Create(ctx, namespace.ID, project)
	require.NoError(t, err)

	issue, err := model.NewIssue(1, "My test epic", model.IssueKindEpic, user.ID)
	require.NoError(t, err)

	err = issueRepo.Create(ctx, project.ID, issue)
	require.NoError(t, err)

	patch := map[string]any{
		"title":       "My updated test epic",
		"description": "My updated test epic description",
	}

	updated, err := issueRepo.Update(ctx, issue.ID, patch)
	require.NoError(t, err)

	assert.Equal(t, patch["title"], updated.Title)
	assert.Equal(t, patch["description"], updated.Description)
	assert.NotNil(t, updated.UpdatedAt)
}

func TestIssueRepository_Delete(t *testing.T) {
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

	namespaceRepo, err := neo4j.NewNamespaceRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

	projectRepo, err := neo4j.NewProjectRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

	issueRepo, err := neo4j.NewIssueRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

	user := prepareUser(t)
	err = userRepo.Create(ctx, user)
	require.NoError(t, err)

	organization := prepareOrganization(t)
	err = orgRepo.Create(ctx, user.ID, organization)
	require.NoError(t, err)

	namespace, err := model.NewNamespace(testutil.GenerateRandomString(10))
	require.NoError(t, err)

	err = namespaceRepo.Create(ctx, organization.ID, namespace)
	require.NoError(t, err)

	project := prepareProject(t)

	err = projectRepo.Create(ctx, namespace.ID, project)
	require.NoError(t, err)

	issue, err := model.NewIssue(1, "My test epic", model.IssueKindEpic, user.ID)
	require.NoError(t, err)

	err = issueRepo.Create(ctx, project.ID, issue)
	require.NoError(t, err)

	_, err = issueRepo.Get(ctx, issue.ID)
	require.NoError(t, err)

	err = issueRepo.Delete(ctx, issue.ID)
	require.NoError(t, err)

	_, err = issueRepo.Get(ctx, issue.ID)
	require.Error(t, err)
}

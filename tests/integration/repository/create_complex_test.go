//go:build integration

package repository

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg/convert"
	"github.com/opcotech/elemo/internal/repository/neo4j"
	testRepo "github.com/opcotech/elemo/internal/testutil/repository"
)

func TestComplex(t *testing.T) {
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

	roleRepo, err := neo4j.NewRoleRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

	orgRepo, err := neo4j.NewOrganizationRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

	nsRepo, err := neo4j.NewNamespaceRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

	projectRepo, err := neo4j.NewProjectRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

	documentRepo, err := neo4j.NewDocumentRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

	todoRepo, err := neo4j.NewTodoRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

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

	assignmentRepo, err := neo4j.NewAssignmentRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

	issueRepo, err := neo4j.NewIssueRepository(
		neo4j.WithDatabase(db),
	)
	require.NoError(t, err)

	gabor := prepareUser(t)
	gabor.Username = "gabor"
	gabor.Email = "gabor@elemo.app"
	gabor.FirstName = "Gabor"
	gabor.LastName = "B"
	gabor.Languages = []model.Language{model.LanguageEN, model.LanguageHU}
	require.NoError(t, userRepo.Create(ctx, gabor))

	kata := prepareUser(t)
	kata.Username = "kata"
	kata.Email = "kata@elemo.app"
	kata.FirstName = "Kata"
	kata.LastName = "D"
	kata.Languages = []model.Language{model.LanguageEN, model.LanguageHU, model.LanguageES}
	require.NoError(t, userRepo.Create(ctx, kata))

	juli := prepareUser(t)
	juli.Username = "juli"
	juli.Email = "juli@elemo.app"
	juli.FirstName = "Juli"
	juli.LastName = "D"
	juli.Languages = []model.Language{model.LanguageEN, model.LanguageHU, model.LanguageES}
	require.NoError(t, userRepo.Create(ctx, juli))

	gaborTodo1, _ := model.NewTodo("Check todos are working", gabor.ID, gabor.ID)
	gaborTodo1.Description = "Check todos are working for Gabor"
	require.NoError(t, todoRepo.Create(ctx, gaborTodo1))

	gaborTodo2, _ := model.NewTodo("Create example test data", gabor.ID, gabor.ID)
	gaborTodo2.Description = "Example data is crucial for testing, create some"
	require.NoError(t, todoRepo.Create(ctx, gaborTodo2))

	kataTodo1, _ := model.NewTodo("Check todos are working", gabor.ID, kata.ID)
	kataTodo1.Description = "Check todos are working for Kata"
	require.NoError(t, todoRepo.Create(ctx, kataTodo1))

	juliTodo1, _ := model.NewTodo("Pet cute doggos", kata.ID, kata.ID)
	juliTodo1.Description = "Go and find some cute doggos to pet"
	require.NoError(t, todoRepo.Create(ctx, juliTodo1))

	adam := prepareUser(t)
	adam.Username = "adam"
	adam.Email = "adam@example.com"
	adam.FirstName = "Adam"
	adam.LastName = "H"
	adam.Languages = []model.Language{model.LanguageEN, model.LanguageHU, model.LanguageES}
	require.NoError(t, userRepo.Create(ctx, adam))

	vera := prepareUser(t)
	vera.Username = "vera"
	vera.Email = "vera@example.com"
	vera.FirstName = "Vera"
	vera.LastName = "M"
	vera.Languages = []model.Language{model.LanguageEN, model.LanguageHU, model.LanguageES, model.LanguageDE}
	require.NoError(t, userRepo.Create(ctx, vera))

	adamTodo1, _ := model.NewTodo("Go for walk with the dogs", vera.ID, adam.ID)
	adamTodo1.Description = "Go for walk with the dogs in the park"
	require.NoError(t, todoRepo.Create(ctx, adamTodo1))

	opcotech := prepareOrganization(t)
	opcotech.Name = "Opcotech"
	opcotech.Email = "info@opcotech.com"
	opcotech.Website = "https://opcotech.com"
	opcotech.Logo = "https://www.opcotech.com/images/logo.png"
	require.NoError(t, orgRepo.Create(ctx, gabor.ID, opcotech))
	require.NoError(t, orgRepo.AddMember(ctx, opcotech.ID, kata.ID))
	require.NoError(t, orgRepo.AddMember(ctx, opcotech.ID, juli.ID))

	example := prepareOrganization(t)
	example.Name = "Example"
	example.Email = "info@example.com"
	example.Website = "https://example.com"
	example.Logo = "https://www.gravatar.com/avatar"
	require.NoError(t, orgRepo.Create(ctx, adam.ID, example))
	require.NoError(t, orgRepo.AddMember(ctx, example.ID, vera.ID))

	opcotechAdmin, _ := model.NewRole("Admins")
	opcotechAdmin.Description = "Admins of Opcotech"
	require.NoError(t, roleRepo.Create(ctx, gabor.ID, opcotech.ID, opcotechAdmin))
	require.NoError(t, roleRepo.AddMember(ctx, opcotechAdmin.ID, kata.ID))

	opcotechDev, _ := model.NewRole("Developers")
	opcotechDev.Description = "Developers of Opcotech"
	require.NoError(t, roleRepo.Create(ctx, gabor.ID, opcotech.ID, opcotechDev))

	opcotechDesigner, _ := model.NewRole("Designers")
	opcotechDesigner.Description = "Designers of Opcotech"
	require.NoError(t, roleRepo.Create(ctx, gabor.ID, opcotech.ID, opcotechDesigner))
	require.NoError(t, roleRepo.AddMember(ctx, opcotechDesigner.ID, kata.ID))
	require.NoError(t, roleRepo.RemoveMember(ctx, opcotechDesigner.ID, gabor.ID))

	opcotechInstructor, _ := model.NewRole("Instructors")
	opcotechInstructor.Description = "Developers of Opcotech"
	require.NoError(t, roleRepo.Create(ctx, gabor.ID, opcotech.ID, opcotechInstructor))
	require.NoError(t, roleRepo.AddMember(ctx, opcotechInstructor.ID, juli.ID))
	require.NoError(t, roleRepo.RemoveMember(ctx, opcotechInstructor.ID, gabor.ID))

	opcotechContractor, _ := model.NewRole("Contractors")
	opcotechContractor.Description = "Contractors of Opcotech"
	require.NoError(t, roleRepo.Create(ctx, gabor.ID, opcotech.ID, opcotechContractor))
	require.NoError(t, roleRepo.AddMember(ctx, opcotechContractor.ID, adam.ID))
	require.NoError(t, roleRepo.RemoveMember(ctx, opcotechContractor.ID, gabor.ID))

	exampleAdmin, _ := model.NewRole("Admins")
	exampleAdmin.Description = "Admins of Example"
	require.NoError(t, roleRepo.Create(ctx, adam.ID, example.ID, exampleAdmin))
	require.NoError(t, roleRepo.AddMember(ctx, exampleAdmin.ID, vera.ID))

	ns, _ := model.NewNamespace("Elemo")
	ns.Description = "The next-generation of project management"
	require.NoError(t, nsRepo.Create(ctx, opcotech.ID, ns))

	webapp, _ := model.NewProject("WEB", "Web application")
	webapp.Description = "Web application for Elemo"
	require.NoError(t, projectRepo.Create(ctx, ns.ID, webapp))

	epic, err := model.NewIssue(1, "Create a new project", model.IssueKindEpic, gabor.ID)
	require.NoError(t, err)
	epic.Description = "New epic description"
	require.NoError(t, issueRepo.Create(ctx, webapp.ID, epic))

	epicAssignment, err := model.NewAssignment(gabor.ID, epic.ID, model.AssignmentKindAssignee)
	require.NoError(t, err)

	epicAssignmentReviewer, err := model.NewAssignment(kata.ID, epic.ID, model.AssignmentKindReviewer)
	require.NoError(t, err)

	require.NoError(t, assignmentRepo.Create(ctx, epicAssignment))
	require.NoError(t, assignmentRepo.Create(ctx, epicAssignmentReviewer))

	issue1, err := model.NewIssue(2, "Create a new project", model.IssueKindTask, gabor.ID)
	require.NoError(t, err)
	issue1.Description = "New task description"
	issue1.Parent = convert.ToPointer(epic.ID)
	require.NoError(t, issueRepo.Create(ctx, webapp.ID, issue1))

	issue1Assignment, err := model.NewAssignment(gabor.ID, issue1.ID, model.AssignmentKindAssignee)
	require.NoError(t, err)

	issue1AssignmentReviewer, err := model.NewAssignment(kata.ID, issue1.ID, model.AssignmentKindReviewer)
	require.NoError(t, err)

	require.NoError(t, assignmentRepo.Create(ctx, issue1Assignment))
	require.NoError(t, assignmentRepo.Create(ctx, issue1AssignmentReviewer))

	issue2, err := model.NewIssue(3, "Create a new project", model.IssueKindTask, gabor.ID)
	require.NoError(t, err)
	issue2.Description = "New task description"
	issue2.Parent = convert.ToPointer(epic.ID)
	require.NoError(t, issueRepo.Create(ctx, webapp.ID, issue2))

	issue2Assignment, err := model.NewAssignment(gabor.ID, issue2.ID, model.AssignmentKindAssignee)
	require.NoError(t, err)

	issue2AssignmentReviewer, err := model.NewAssignment(gabor.ID, issue2.ID, model.AssignmentKindReviewer)
	require.NoError(t, err)

	require.NoError(t, assignmentRepo.Create(ctx, issue2Assignment))
	require.NoError(t, assignmentRepo.Create(ctx, issue2AssignmentReviewer))

	document, _ := model.NewDocument("Project discovery", "file_id", gabor.ID)
	document.Excerpt = "Project discovery document"
	require.NoError(t, documentRepo.Create(ctx, webapp.ID, document))

	comment1, _ := model.NewComment("This is great!", gabor.ID)
	require.NoError(t, commentRepo.Create(ctx, document.ID, comment1))

	comment2, _ := model.NewComment("How would we handle 10k new users per day?", kata.ID)
	require.NoError(t, commentRepo.Create(ctx, document.ID, comment2))

	comment3, _ := model.NewComment("We could use a load balancer and a cache layer.", gabor.ID)
	require.NoError(t, commentRepo.Create(ctx, document.ID, comment3))

	attachment, _ := model.NewAttachment("file_name", "file_id", gabor.ID)
	require.NoError(t, attachmentRepo.Create(ctx, document.ID, attachment))

	discoveryLabel, _ := model.NewLabel("discovery")
	discoveryLabel.Description = "Project discovery"
	require.NoError(t, labelRepo.Create(ctx, discoveryLabel))
	require.NoError(t, labelRepo.AttachTo(ctx, discoveryLabel.ID, document.ID))

	elemoAdmins, _ := model.NewRole("Admins")
	elemoAdmins.Description = "Admins of Elemo web application"
	require.NoError(t, roleRepo.Create(ctx, gabor.ID, webapp.ID, elemoAdmins))
	require.NoError(t, roleRepo.AddMember(ctx, elemoAdmins.ID, kata.ID))

	elemoMembers, _ := model.NewRole("Members")
	elemoMembers.Description = "Members of Elemo web application"
	require.NoError(t, roleRepo.Create(ctx, gabor.ID, webapp.ID, elemoMembers))
	require.NoError(t, roleRepo.AddMember(ctx, elemoMembers.ID, adam.ID))
	require.NoError(t, roleRepo.RemoveMember(ctx, elemoMembers.ID, gabor.ID))
}

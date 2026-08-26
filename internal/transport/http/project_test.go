package http

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg/convert"
	"github.com/opcotech/elemo/internal/pkg/optional"
	"github.com/opcotech/elemo/internal/repository"
	"github.com/opcotech/elemo/internal/service"
	mocksvc "github.com/opcotech/elemo/internal/service/mock"
	"github.com/opcotech/elemo/internal/transport/http/api"
)

func newServiceProject() *service.Project {
	return &service.Project{
		ID:            model.MustNewID(model.ResourceTypeProject),
		Key:           "ENG",
		Name:          "Engineering",
		Description:   "Main engineering project",
		Logo:          "https://example.com/logo.png",
		Status:        model.ProjectStatusActive,
		Teams:         []model.ID{model.MustNewID(model.ResourceTypeRole)},
		DocumentCount: convert.ToPointer(int64(1)),
		IssueCount:    convert.ToPointer(int64(0)),
		CreatedAt:     convert.ToPointer(time.Now().UTC()),
	}
}

func TestNewProjectController(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		c, err := NewProjectController(
			mocksvc.NewMockOrganizationService(ctrl),
			mocksvc.NewMockNamespaceService(ctrl),
			mocksvc.NewMockProjectService(ctrl),
		)
		require.NoError(t, err)
		assert.NotNil(t, c)
	})

	t.Run("missing project service", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		_, err := NewProjectController(
			mocksvc.NewMockOrganizationService(ctrl),
			mocksvc.NewMockNamespaceService(ctrl),
			nil,
		)
		assert.ErrorIs(t, err, ErrNoProjectService)
	})
}

func TestProjectController_V1NamespacesProjectsCreate(t *testing.T) {
	t.Parallel()

	orgID := model.MustNewID(model.ResourceTypeOrganization)
	namespaceID := model.MustNewID(model.ResourceTypeNamespace)
	project := newServiceProject()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ps := mocksvc.NewMockProjectService(ctrl)
		ps.EXPECT().Create(gomock.Any(), namespaceID, service.CreateProjectOpts{
			Key:  "ENG",
			Name: "Engineering",
		}).Return(project, nil)

		c, os, ns := newTestProjectController(t, ctrl, ps)
		stubOrganizationResolve(os, orgID)
		stubNamespaceResolve(ns, orgID, namespaceID)
		resp, err := c.V1NamespacesProjectsCreate(context.Background(), api.V1NamespacesProjectsCreateRequestObject{
			OrganizationRef: orgID.String(),
			NamespaceRef:    namespaceID.String(),
			Body: &api.V1NamespacesProjectsCreateJSONRequestBody{
				Key:  "ENG",
				Name: "Engineering",
			},
		})
		require.NoError(t, err)
		got, ok := resp.(api.V1NamespacesProjectsCreate201JSONResponse)
		require.True(t, ok)
		assert.Equal(t, project.ID.String(), got.Id)
	})

	t.Run("bad namespace id", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		c, os, ns := newTestProjectController(t, ctrl, mocksvc.NewMockProjectService(ctrl))
		stubOrganizationResolve(os, orgID)
		stubNamespaceResolve(ns, orgID, namespaceID)
		resp, err := c.V1NamespacesProjectsCreate(context.Background(), api.V1NamespacesProjectsCreateRequestObject{
			OrganizationRef: "AB",
			Body:            &api.V1NamespacesProjectsCreateJSONRequestBody{Key: "ENG", Name: "Engineering"},
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1NamespacesProjectsCreate400JSONResponse)
		assert.True(t, ok)
	})

	t.Run("invalid status", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		badStatus := api.ProjectStatus("bogus")
		c, os, ns := newTestProjectController(t, ctrl, mocksvc.NewMockProjectService(ctrl))
		stubOrganizationResolve(os, orgID)
		stubNamespaceResolve(ns, orgID, namespaceID)
		resp, err := c.V1NamespacesProjectsCreate(context.Background(), api.V1NamespacesProjectsCreateRequestObject{
			OrganizationRef: orgID.String(),
			NamespaceRef:    namespaceID.String(),
			Body: &api.V1NamespacesProjectsCreateJSONRequestBody{
				Key:    "ENG",
				Name:   "Engineering",
				Status: &badStatus,
			},
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1NamespacesProjectsCreate400JSONResponse)
		assert.True(t, ok)
	})

	t.Run("no permission", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ps := mocksvc.NewMockProjectService(ctrl)
		ps.EXPECT().Create(gomock.Any(), namespaceID, gomock.Any()).Return(nil, service.ErrNoPermission)

		c, os, ns := newTestProjectController(t, ctrl, ps)
		stubOrganizationResolve(os, orgID)
		stubNamespaceResolve(ns, orgID, namespaceID)
		resp, err := c.V1NamespacesProjectsCreate(context.Background(), api.V1NamespacesProjectsCreateRequestObject{
			OrganizationRef: orgID.String(),
			NamespaceRef:    namespaceID.String(),
			Body:            &api.V1NamespacesProjectsCreateJSONRequestBody{Key: "ENG", Name: "Engineering"},
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1NamespacesProjectsCreate403JSONResponse)
		assert.True(t, ok)
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ps := mocksvc.NewMockProjectService(ctrl)
		ps.EXPECT().Create(gomock.Any(), namespaceID, gomock.Any()).Return(nil, errors.Join(service.ErrProjectCreate, repository.ErrNotFound))

		c, os, ns := newTestProjectController(t, ctrl, ps)
		stubOrganizationResolve(os, orgID)
		stubNamespaceResolve(ns, orgID, namespaceID)
		resp, err := c.V1NamespacesProjectsCreate(context.Background(), api.V1NamespacesProjectsCreateRequestObject{
			OrganizationRef: orgID.String(),
			NamespaceRef:    namespaceID.String(),
			Body:            &api.V1NamespacesProjectsCreateJSONRequestBody{Key: "ENG", Name: "Engineering"},
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1NamespacesProjectsCreate404JSONResponse)
		assert.True(t, ok)
	})

	t.Run("key conflict", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ps := mocksvc.NewMockProjectService(ctrl)
		ps.EXPECT().Create(gomock.Any(), namespaceID, gomock.Any()).Return(nil, errors.Join(service.ErrProjectCreate, repository.ErrProjectKeyConflict))

		c, os, ns := newTestProjectController(t, ctrl, ps)
		stubOrganizationResolve(os, orgID)
		stubNamespaceResolve(ns, orgID, namespaceID)
		resp, err := c.V1NamespacesProjectsCreate(context.Background(), api.V1NamespacesProjectsCreateRequestObject{
			OrganizationRef: orgID.String(),
			NamespaceRef:    namespaceID.String(),
			Body:            &api.V1NamespacesProjectsCreateJSONRequestBody{Key: "ENG", Name: "Engineering"},
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1NamespacesProjectsCreate409JSONResponse)
		assert.True(t, ok)
	})
}

func TestProjectController_V1NamespacesProjectsGet(t *testing.T) {
	t.Parallel()

	orgID := model.MustNewID(model.ResourceTypeOrganization)
	namespaceID := model.MustNewID(model.ResourceTypeNamespace)
	project := newServiceProject()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ps := mocksvc.NewMockProjectService(ctrl)
		ps.EXPECT().List(gomock.Any(), namespaceID, gomock.Any()).Return(service.Page[*service.Project]{Items: []*service.Project{project}}, nil)

		c, os, ns := newTestProjectController(t, ctrl, ps)
		stubOrganizationResolve(os, orgID)
		stubNamespaceResolve(ns, orgID, namespaceID)
		resp, err := c.V1NamespacesProjectsGet(context.Background(), api.V1NamespacesProjectsGetRequestObject{
			OrganizationRef: orgID.String(),
			NamespaceRef:    namespaceID.String(),
			Params:          api.V1NamespacesProjectsGetParams{},
		})
		require.NoError(t, err)
		got, ok := resp.(api.V1NamespacesProjectsGet200JSONResponse)
		require.True(t, ok)
		require.Len(t, got.Items, 1)
		assert.Equal(t, project.ID.String(), got.Items[0].Id)
		assert.Equal(t, project.Key, got.Items[0].Key)
		require.NotNil(t, got.Items[0].DocumentCount)
		assert.Equal(t, int64(1), *got.Items[0].DocumentCount)
	})

	t.Run("bad namespace id", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		c, os, ns := newTestProjectController(t, ctrl, mocksvc.NewMockProjectService(ctrl))
		stubOrganizationResolve(os, orgID)
		stubNamespaceResolve(ns, orgID, namespaceID)
		resp, err := c.V1NamespacesProjectsGet(context.Background(), api.V1NamespacesProjectsGetRequestObject{
			OrganizationRef: "AB",
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1NamespacesProjectsGet400JSONResponse)
		assert.True(t, ok)
	})

	t.Run("no permission", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ps := mocksvc.NewMockProjectService(ctrl)
		ps.EXPECT().List(gomock.Any(), namespaceID, gomock.Any()).Return(service.Page[*service.Project]{}, service.ErrNoPermission)

		c, os, ns := newTestProjectController(t, ctrl, ps)
		stubOrganizationResolve(os, orgID)
		stubNamespaceResolve(ns, orgID, namespaceID)
		resp, err := c.V1NamespacesProjectsGet(context.Background(), api.V1NamespacesProjectsGetRequestObject{
			OrganizationRef: orgID.String(),
			NamespaceRef:    namespaceID.String(),
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1NamespacesProjectsGet403JSONResponse)
		assert.True(t, ok)
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ps := mocksvc.NewMockProjectService(ctrl)
		ps.EXPECT().List(gomock.Any(), namespaceID, gomock.Any()).Return(service.Page[*service.Project]{}, errors.Join(service.ErrProjectList, repository.ErrNotFound))

		c, os, ns := newTestProjectController(t, ctrl, ps)
		stubOrganizationResolve(os, orgID)
		stubNamespaceResolve(ns, orgID, namespaceID)
		resp, err := c.V1NamespacesProjectsGet(context.Background(), api.V1NamespacesProjectsGetRequestObject{
			OrganizationRef: orgID.String(),
			NamespaceRef:    namespaceID.String(),
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1NamespacesProjectsGet404JSONResponse)
		assert.True(t, ok)
	})
}

func TestProjectController_V1NamespacesProjectsKeyGet(t *testing.T) {
	t.Parallel()

	orgID := model.MustNewID(model.ResourceTypeOrganization)
	namespaceID := model.MustNewID(model.ResourceTypeNamespace)
	project := newServiceProject()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ps := mocksvc.NewMockProjectService(ctrl)
		ps.EXPECT().GetByKey(gomock.Any(), namespaceID, project.Key).Return(project, nil)

		c, os, ns := newTestProjectController(t, ctrl, ps)
		stubOrganizationResolve(os, orgID)
		stubNamespaceResolve(ns, orgID, namespaceID)
		resp, err := c.V1NamespacesProjectsKeyGet(context.Background(), api.V1NamespacesProjectsKeyGetRequestObject{
			OrganizationRef: orgID.String(),
			NamespaceRef:    namespaceID.String(),
			ProjectKey:      project.Key,
		})
		require.NoError(t, err)
		got, ok := resp.(api.V1NamespacesProjectsKeyGet200JSONResponse)
		require.True(t, ok)
		assert.Equal(t, project.ID.String(), got.Id)
		assert.Equal(t, project.Key, got.Key)
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ps := mocksvc.NewMockProjectService(ctrl)
		ps.EXPECT().GetByKey(gomock.Any(), namespaceID, project.Key).Return(nil, errors.Join(service.ErrProjectGet, repository.ErrNotFound))

		c, os, ns := newTestProjectController(t, ctrl, ps)
		stubOrganizationResolve(os, orgID)
		stubNamespaceResolve(ns, orgID, namespaceID)
		resp, err := c.V1NamespacesProjectsKeyGet(context.Background(), api.V1NamespacesProjectsKeyGetRequestObject{
			OrganizationRef: orgID.String(),
			NamespaceRef:    namespaceID.String(),
			ProjectKey:      project.Key,
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1NamespacesProjectsKeyGet404JSONResponse)
		assert.True(t, ok)
	})
}

func TestProjectController_V1ProjectGet(t *testing.T) {
	t.Parallel()

	project := newServiceProject()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ps := mocksvc.NewMockProjectService(ctrl)
		ps.EXPECT().Get(gomock.Any(), project.ID).Return(project, nil)

		c, _, _ := newTestProjectController(t, ctrl, ps)
		resp, err := c.V1ProjectGet(context.Background(), api.V1ProjectGetRequestObject{ProjectId: project.ID.String()})
		require.NoError(t, err)
		got, ok := resp.(api.V1ProjectGet200JSONResponse)
		require.True(t, ok)
		assert.Equal(t, project.ID.String(), got.Id)
		assert.Equal(t, project.Name, got.Name)
		require.NotNil(t, got.Description)
		assert.Equal(t, project.Description, *got.Description)
	})

	t.Run("bad id", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		c, _, _ := newTestProjectController(t, ctrl, mocksvc.NewMockProjectService(ctrl))
		resp, err := c.V1ProjectGet(context.Background(), api.V1ProjectGetRequestObject{ProjectId: "AB"})
		require.NoError(t, err)
		_, ok := resp.(api.V1ProjectGet400JSONResponse)
		assert.True(t, ok)
	})

	t.Run("no permission", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ps := mocksvc.NewMockProjectService(ctrl)
		ps.EXPECT().Get(gomock.Any(), project.ID).Return(nil, service.ErrNoPermission)

		c, _, _ := newTestProjectController(t, ctrl, ps)
		resp, err := c.V1ProjectGet(context.Background(), api.V1ProjectGetRequestObject{ProjectId: project.ID.String()})
		require.NoError(t, err)
		_, ok := resp.(api.V1ProjectGet403JSONResponse)
		assert.True(t, ok)
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ps := mocksvc.NewMockProjectService(ctrl)
		ps.EXPECT().Get(gomock.Any(), project.ID).Return(nil, errors.Join(service.ErrProjectGet, repository.ErrNotFound))

		c, _, _ := newTestProjectController(t, ctrl, ps)
		resp, err := c.V1ProjectGet(context.Background(), api.V1ProjectGetRequestObject{ProjectId: project.ID.String()})
		require.NoError(t, err)
		_, ok := resp.(api.V1ProjectGet404JSONResponse)
		assert.True(t, ok)
	})
}

func TestProjectController_V1ProjectUpdate(t *testing.T) {
	t.Parallel()

	project := newServiceProject()
	name := "Updated"

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		updated := *project
		updated.Name = name

		ps := mocksvc.NewMockProjectService(ctrl)
		ps.EXPECT().Update(gomock.Any(), project.ID, service.UpdateProjectOpts{
			Name: optional.Some(name),
		}).Return(&updated, nil)

		c, _, _ := newTestProjectController(t, ctrl, ps)
		resp, err := c.V1ProjectUpdate(context.Background(), api.V1ProjectUpdateRequestObject{
			ProjectId: project.ID.String(),
			Body:      &api.V1ProjectUpdateJSONRequestBody{Name: &name},
		})
		require.NoError(t, err)
		got, ok := resp.(api.V1ProjectUpdate200JSONResponse)
		require.True(t, ok)
		assert.Equal(t, name, got.Name)
	})

	t.Run("bad id", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		c, _, _ := newTestProjectController(t, ctrl, mocksvc.NewMockProjectService(ctrl))
		resp, err := c.V1ProjectUpdate(context.Background(), api.V1ProjectUpdateRequestObject{
			ProjectId: "AB",
			Body:      &api.V1ProjectUpdateJSONRequestBody{Name: &name},
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1ProjectUpdate400JSONResponse)
		assert.True(t, ok)
	})

	t.Run("no permission", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ps := mocksvc.NewMockProjectService(ctrl)
		ps.EXPECT().Update(gomock.Any(), project.ID, gomock.Any()).Return(nil, service.ErrNoPermission)

		c, _, _ := newTestProjectController(t, ctrl, ps)
		resp, err := c.V1ProjectUpdate(context.Background(), api.V1ProjectUpdateRequestObject{
			ProjectId: project.ID.String(),
			Body:      &api.V1ProjectUpdateJSONRequestBody{Name: &name},
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1ProjectUpdate403JSONResponse)
		assert.True(t, ok)
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ps := mocksvc.NewMockProjectService(ctrl)
		ps.EXPECT().Update(gomock.Any(), project.ID, gomock.Any()).Return(nil, errors.Join(service.ErrProjectUpdate, repository.ErrNotFound))

		c, _, _ := newTestProjectController(t, ctrl, ps)
		resp, err := c.V1ProjectUpdate(context.Background(), api.V1ProjectUpdateRequestObject{
			ProjectId: project.ID.String(),
			Body:      &api.V1ProjectUpdateJSONRequestBody{Name: &name},
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1ProjectUpdate404JSONResponse)
		assert.True(t, ok)
	})
}

func TestProjectController_V1ProjectDelete(t *testing.T) {
	t.Parallel()

	projectID := model.MustNewID(model.ResourceTypeProject)

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ps := mocksvc.NewMockProjectService(ctrl)
		ps.EXPECT().Delete(gomock.Any(), projectID).Return(nil)

		c, _, _ := newTestProjectController(t, ctrl, ps)
		resp, err := c.V1ProjectDelete(context.Background(), api.V1ProjectDeleteRequestObject{ProjectId: projectID.String()})
		require.NoError(t, err)
		_, ok := resp.(api.V1ProjectDelete204Response)
		assert.True(t, ok)
	})

	t.Run("bad id", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		c, _, _ := newTestProjectController(t, ctrl, mocksvc.NewMockProjectService(ctrl))
		resp, err := c.V1ProjectDelete(context.Background(), api.V1ProjectDeleteRequestObject{ProjectId: "AB"})
		require.NoError(t, err)
		_, ok := resp.(api.V1ProjectDelete400JSONResponse)
		assert.True(t, ok)
	})

	t.Run("no permission", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ps := mocksvc.NewMockProjectService(ctrl)
		ps.EXPECT().Delete(gomock.Any(), projectID).Return(service.ErrNoPermission)

		c, _, _ := newTestProjectController(t, ctrl, ps)
		resp, err := c.V1ProjectDelete(context.Background(), api.V1ProjectDeleteRequestObject{ProjectId: projectID.String()})
		require.NoError(t, err)
		_, ok := resp.(api.V1ProjectDelete403JSONResponse)
		assert.True(t, ok)
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ps := mocksvc.NewMockProjectService(ctrl)
		ps.EXPECT().Delete(gomock.Any(), projectID).Return(errors.Join(service.ErrProjectDelete, repository.ErrNotFound))

		c, _, _ := newTestProjectController(t, ctrl, ps)
		resp, err := c.V1ProjectDelete(context.Background(), api.V1ProjectDeleteRequestObject{ProjectId: projectID.String()})
		require.NoError(t, err)
		_, ok := resp.(api.V1ProjectDelete404JSONResponse)
		assert.True(t, ok)
	})
}

func TestPartialIssueToDTO(t *testing.T) {
	t.Parallel()

	parentID := model.MustNewID(model.ResourceTypeIssue)
	assigneeID := model.MustNewID(model.ResourceTypeUser)
	reporterID := model.MustNewID(model.ResourceTypeUser)
	labelID := model.MustNewID(model.ResourceTypeLabel)
	dueDate := time.Now().UTC()

	issue := &service.PartialIssue{
		ID:        model.MustNewID(model.ResourceTypeIssue),
		Key:       "ENG-7",
		NumericID: 7,
		Parent: &service.PartialIssue{
			ID:          parentID,
			Key:         "ENG-1",
			NumericID:   1,
			Kind:        model.IssueKindEpic,
			Title:       "Parent epic",
			Status:      model.IssueStatusOpen,
			Priority:    model.IssuePriorityHigh,
			Assignments: make([]service.PartialAssignee, 0),
			Labels:      make([]service.PartialLabel, 0),
		},
		Kind:        model.IssueKindStory,
		Title:       "Implement authentication",
		Description: "Add OAuth2 login flow",
		Status:      model.IssueStatusInProgress,
		Priority:    model.IssuePriorityNormal,
		Assignments: []service.PartialAssignee{
			{ID: assigneeID, Kind: model.AssignmentKindAssignee, FirstName: "Ada", LastName: "Lovelace"},
		},
		Labels:     []service.PartialLabel{{ID: labelID, Name: "frontend"}},
		Project:    &service.PartialProject{ID: model.MustNewID(model.ResourceTypeProject), Key: "ENG", Name: "Engineering", Status: model.ProjectStatusActive},
		Namespace:  &service.PartialNamespace{ID: model.MustNewID(model.ResourceTypeNamespace), Name: "Platform"},
		ReportedBy: &service.PartialUser{ID: reporterID},
		DueDate:    &dueDate,
	}

	dto := partialIssueToDTO(issue)
	assert.Equal(t, issue.ID.String(), dto.Id)
	assert.Equal(t, issue.Key, dto.Key)
	assert.Equal(t, int(issue.NumericID), dto.NumericId)
	assert.Equal(t, issue.Title, dto.Title)
	require.NotNil(t, dto.Description)
	assert.Equal(t, issue.Description, *dto.Description)
	assert.Equal(t, api.IssueKind(issue.Kind.String()), dto.Kind)
	assert.Equal(t, api.IssueStatus(issue.Status.String()), dto.Status)
	assert.Equal(t, api.IssuePriority(issue.Priority.String()), dto.Priority)
	require.NotNil(t, dto.ReportedBy)
	assert.Equal(t, reporterID.String(), dto.ReportedBy.Id)
	require.NotNil(t, dto.Parent)
	assert.Equal(t, parentID.String(), dto.Parent.Id)
	assert.Equal(t, "ENG-1", dto.Parent.Key)
	assert.Equal(t, "Parent epic", dto.Parent.Title)
	assert.Equal(t, []api.PartialUser{{
		Id:        assigneeID.String(),
		FirstName: "Ada",
		LastName:  "Lovelace",
	}}, dto.Assignees)
	assert.Empty(t, dto.Reviewers)
	require.NotNil(t, dto.Project)
	assert.Equal(t, "ENG", dto.Project.Key)
	require.NotNil(t, dto.Namespace)
	assert.Equal(t, "Platform", dto.Namespace.Name)
	assert.Equal(t, []api.PartialLabel{{Id: labelID.String(), Name: "frontend"}}, dto.Labels)
	require.NotNil(t, dto.DueDate)
	assert.Equal(t, dueDate, *dto.DueDate)
}

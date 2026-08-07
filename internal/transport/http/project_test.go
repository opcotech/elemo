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
	"github.com/opcotech/elemo/internal/transport/http/api"
)

func newTestProjectController(t *testing.T, ps service.ProjectService) ProjectController {
	t.Helper()
	c, err := NewProjectController(WithProjectService(ps))
	require.NoError(t, err)
	return c
}

func newServiceProject() *service.Project {
	return &service.Project{
		ID:          model.MustNewID(model.ResourceTypeProject),
		Key:         "ENG",
		Name:        "Engineering",
		Description: "Main engineering project",
		Logo:        "https://example.com/logo.png",
		Status:      model.ProjectStatusActive,
		Teams:       []model.ID{model.MustNewID(model.ResourceTypeRole)},
		Documents: []*service.PartialDocument{
			{
				ID:        model.MustNewID(model.ResourceTypeDocument),
				Name:      "Plan",
				Excerpt:   "Overview",
				CreatedBy: model.MustNewID(model.ResourceTypeUser),
				CreatedAt: convert.ToPointer(time.Now().UTC()),
			},
		},
		Issues:    []model.ID{},
		CreatedAt: convert.ToPointer(time.Now().UTC()),
	}
}

func TestNewProjectController(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		c, err := NewProjectController(WithProjectService(service.NewMockProjectService(ctrl)))
		require.NoError(t, err)
		assert.NotNil(t, c)
	})

	t.Run("missing project service", func(t *testing.T) {
		t.Parallel()
		_, err := NewProjectController()
		assert.ErrorIs(t, err, ErrNoProjectService)
	})

	t.Run("nil project service option", func(t *testing.T) {
		t.Parallel()
		_, err := NewProjectController(WithProjectService(nil))
		assert.ErrorIs(t, err, ErrNoProjectService)
	})
}

func TestProjectController_V1NamespacesProjectsCreate(t *testing.T) {
	t.Parallel()

	namespaceID := model.MustNewID(model.ResourceTypeNamespace)
	project := newServiceProject()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ps := service.NewMockProjectService(ctrl)
		ps.EXPECT().Create(gomock.Any(), namespaceID, service.CreateProjectOpts{
			Key:  "ENG",
			Name: "Engineering",
		}).Return(project, nil)

		c := newTestProjectController(t, ps)
		resp, err := c.V1NamespacesProjectsCreate(context.Background(), api.V1NamespacesProjectsCreateRequestObject{
			Id: namespaceID.String(),
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

		c := newTestProjectController(t, service.NewMockProjectService(ctrl))
		resp, err := c.V1NamespacesProjectsCreate(context.Background(), api.V1NamespacesProjectsCreateRequestObject{
			Id:   "not-a-xid",
			Body: &api.V1NamespacesProjectsCreateJSONRequestBody{Key: "ENG", Name: "Engineering"},
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
		c := newTestProjectController(t, service.NewMockProjectService(ctrl))
		resp, err := c.V1NamespacesProjectsCreate(context.Background(), api.V1NamespacesProjectsCreateRequestObject{
			Id: namespaceID.String(),
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

		ps := service.NewMockProjectService(ctrl)
		ps.EXPECT().Create(gomock.Any(), namespaceID, gomock.Any()).Return(nil, service.ErrNoPermission)

		c := newTestProjectController(t, ps)
		resp, err := c.V1NamespacesProjectsCreate(context.Background(), api.V1NamespacesProjectsCreateRequestObject{
			Id:   namespaceID.String(),
			Body: &api.V1NamespacesProjectsCreateJSONRequestBody{Key: "ENG", Name: "Engineering"},
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1NamespacesProjectsCreate403JSONResponse)
		assert.True(t, ok)
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ps := service.NewMockProjectService(ctrl)
		ps.EXPECT().Create(gomock.Any(), namespaceID, gomock.Any()).Return(nil, errors.Join(service.ErrProjectCreate, repository.ErrNotFound))

		c := newTestProjectController(t, ps)
		resp, err := c.V1NamespacesProjectsCreate(context.Background(), api.V1NamespacesProjectsCreateRequestObject{
			Id:   namespaceID.String(),
			Body: &api.V1NamespacesProjectsCreateJSONRequestBody{Key: "ENG", Name: "Engineering"},
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1NamespacesProjectsCreate404JSONResponse)
		assert.True(t, ok)
	})
}

func TestProjectController_V1NamespacesProjectsGet(t *testing.T) {
	t.Parallel()

	namespaceID := model.MustNewID(model.ResourceTypeNamespace)
	project := newServiceProject()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ps := service.NewMockProjectService(ctrl)
		ps.EXPECT().GetAll(gomock.Any(), namespaceID, DefaultOffset, DefaultLimit).Return([]*service.Project{project}, nil)

		c := newTestProjectController(t, ps)
		resp, err := c.V1NamespacesProjectsGet(context.Background(), api.V1NamespacesProjectsGetRequestObject{
			Id:     namespaceID.String(),
			Params: api.V1NamespacesProjectsGetParams{},
		})
		require.NoError(t, err)
		got, ok := resp.(api.V1NamespacesProjectsGet200JSONResponse)
		require.True(t, ok)
		require.Len(t, got, 1)
		assert.Equal(t, project.ID.String(), got[0].Id)
		assert.Equal(t, project.Key, got[0].Key)
		require.Len(t, got[0].Documents, 1)
		assert.Equal(t, "Plan", got[0].Documents[0].Name)
	})

	t.Run("bad namespace id", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		c := newTestProjectController(t, service.NewMockProjectService(ctrl))
		resp, err := c.V1NamespacesProjectsGet(context.Background(), api.V1NamespacesProjectsGetRequestObject{
			Id: "bad",
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1NamespacesProjectsGet400JSONResponse)
		assert.True(t, ok)
	})

	t.Run("no permission", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ps := service.NewMockProjectService(ctrl)
		ps.EXPECT().GetAll(gomock.Any(), namespaceID, DefaultOffset, DefaultLimit).Return(nil, service.ErrNoPermission)

		c := newTestProjectController(t, ps)
		resp, err := c.V1NamespacesProjectsGet(context.Background(), api.V1NamespacesProjectsGetRequestObject{
			Id: namespaceID.String(),
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1NamespacesProjectsGet403JSONResponse)
		assert.True(t, ok)
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ps := service.NewMockProjectService(ctrl)
		ps.EXPECT().GetAll(gomock.Any(), namespaceID, DefaultOffset, DefaultLimit).Return(nil, errors.Join(service.ErrProjectGetAll, repository.ErrNotFound))

		c := newTestProjectController(t, ps)
		resp, err := c.V1NamespacesProjectsGet(context.Background(), api.V1NamespacesProjectsGetRequestObject{
			Id: namespaceID.String(),
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1NamespacesProjectsGet404JSONResponse)
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

		ps := service.NewMockProjectService(ctrl)
		ps.EXPECT().Get(gomock.Any(), project.ID).Return(project, nil)

		c := newTestProjectController(t, ps)
		resp, err := c.V1ProjectGet(context.Background(), api.V1ProjectGetRequestObject{Id: project.ID.String()})
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

		c := newTestProjectController(t, service.NewMockProjectService(ctrl))
		resp, err := c.V1ProjectGet(context.Background(), api.V1ProjectGetRequestObject{Id: "bad"})
		require.NoError(t, err)
		_, ok := resp.(api.V1ProjectGet400JSONResponse)
		assert.True(t, ok)
	})

	t.Run("no permission", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ps := service.NewMockProjectService(ctrl)
		ps.EXPECT().Get(gomock.Any(), project.ID).Return(nil, service.ErrNoPermission)

		c := newTestProjectController(t, ps)
		resp, err := c.V1ProjectGet(context.Background(), api.V1ProjectGetRequestObject{Id: project.ID.String()})
		require.NoError(t, err)
		_, ok := resp.(api.V1ProjectGet403JSONResponse)
		assert.True(t, ok)
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ps := service.NewMockProjectService(ctrl)
		ps.EXPECT().Get(gomock.Any(), project.ID).Return(nil, errors.Join(service.ErrProjectGet, repository.ErrNotFound))

		c := newTestProjectController(t, ps)
		resp, err := c.V1ProjectGet(context.Background(), api.V1ProjectGetRequestObject{Id: project.ID.String()})
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

		ps := service.NewMockProjectService(ctrl)
		ps.EXPECT().Update(gomock.Any(), project.ID, service.UpdateProjectOpts{
			Name: optional.Some(name),
		}).Return(&updated, nil)

		c := newTestProjectController(t, ps)
		resp, err := c.V1ProjectUpdate(context.Background(), api.V1ProjectUpdateRequestObject{
			Id:   project.ID.String(),
			Body: &api.V1ProjectUpdateJSONRequestBody{Name: &name},
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

		c := newTestProjectController(t, service.NewMockProjectService(ctrl))
		resp, err := c.V1ProjectUpdate(context.Background(), api.V1ProjectUpdateRequestObject{
			Id:   "bad",
			Body: &api.V1ProjectUpdateJSONRequestBody{Name: &name},
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1ProjectUpdate400JSONResponse)
		assert.True(t, ok)
	})

	t.Run("no permission", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ps := service.NewMockProjectService(ctrl)
		ps.EXPECT().Update(gomock.Any(), project.ID, gomock.Any()).Return(nil, service.ErrNoPermission)

		c := newTestProjectController(t, ps)
		resp, err := c.V1ProjectUpdate(context.Background(), api.V1ProjectUpdateRequestObject{
			Id:   project.ID.String(),
			Body: &api.V1ProjectUpdateJSONRequestBody{Name: &name},
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1ProjectUpdate403JSONResponse)
		assert.True(t, ok)
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ps := service.NewMockProjectService(ctrl)
		ps.EXPECT().Update(gomock.Any(), project.ID, gomock.Any()).Return(nil, errors.Join(service.ErrProjectUpdate, repository.ErrNotFound))

		c := newTestProjectController(t, ps)
		resp, err := c.V1ProjectUpdate(context.Background(), api.V1ProjectUpdateRequestObject{
			Id:   project.ID.String(),
			Body: &api.V1ProjectUpdateJSONRequestBody{Name: &name},
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

		ps := service.NewMockProjectService(ctrl)
		ps.EXPECT().Delete(gomock.Any(), projectID).Return(nil)

		c := newTestProjectController(t, ps)
		resp, err := c.V1ProjectDelete(context.Background(), api.V1ProjectDeleteRequestObject{Id: projectID.String()})
		require.NoError(t, err)
		_, ok := resp.(api.V1ProjectDelete204Response)
		assert.True(t, ok)
	})

	t.Run("bad id", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		c := newTestProjectController(t, service.NewMockProjectService(ctrl))
		resp, err := c.V1ProjectDelete(context.Background(), api.V1ProjectDeleteRequestObject{Id: "bad"})
		require.NoError(t, err)
		_, ok := resp.(api.V1ProjectDelete400JSONResponse)
		assert.True(t, ok)
	})

	t.Run("no permission", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ps := service.NewMockProjectService(ctrl)
		ps.EXPECT().Delete(gomock.Any(), projectID).Return(service.ErrNoPermission)

		c := newTestProjectController(t, ps)
		resp, err := c.V1ProjectDelete(context.Background(), api.V1ProjectDeleteRequestObject{Id: projectID.String()})
		require.NoError(t, err)
		_, ok := resp.(api.V1ProjectDelete403JSONResponse)
		assert.True(t, ok)
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ps := service.NewMockProjectService(ctrl)
		ps.EXPECT().Delete(gomock.Any(), projectID).Return(errors.Join(service.ErrProjectDelete, repository.ErrNotFound))

		c := newTestProjectController(t, ps)
		resp, err := c.V1ProjectDelete(context.Background(), api.V1ProjectDeleteRequestObject{Id: projectID.String()})
		require.NoError(t, err)
		_, ok := resp.(api.V1ProjectDelete404JSONResponse)
		assert.True(t, ok)
	})
}

func TestPartialProjectToDTO(t *testing.T) {
	t.Parallel()

	p := &service.PartialProject{
		ID:          model.MustNewID(model.ResourceTypeProject),
		Key:         "ENG",
		Name:        "Engineering",
		Description: "desc",
		Logo:        "https://example.com/logo.png",
		Status:      model.ProjectStatusActive,
	}

	dto := partialProjectToDTO(p)
	assert.Equal(t, p.ID.String(), dto.Id)
	assert.Equal(t, p.Key, dto.Key)
	require.NotNil(t, dto.Description)
	assert.Equal(t, p.Description, *dto.Description)
	require.NotNil(t, dto.Logo)
	assert.Equal(t, p.Logo, *dto.Logo)
}

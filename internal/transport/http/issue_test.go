package http

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/opcotech/elemo/internal/license"
	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg/convert"
	"github.com/opcotech/elemo/internal/pkg/optional"
	"github.com/opcotech/elemo/internal/repository"
	"github.com/opcotech/elemo/internal/service"
	mocksvc "github.com/opcotech/elemo/internal/service/mock"
	"github.com/opcotech/elemo/internal/transport/http/api"
)

func newTestIssueController(t *testing.T, is service.IssueService) IssueController {
	t.Helper()
	c, err := NewIssueController(is)
	require.NoError(t, err)
	return c
}

func newServiceIssue() *service.Issue {
	return &service.Issue{
		ID:              model.MustNewID(model.ResourceTypeIssue),
		Key:             "ENG-42",
		NumericID:       42,
		Kind:            model.IssueKindStory,
		Title:           "Implement authentication",
		Description:     "Add OAuth2 login flow",
		Status:          model.IssueStatusOpen,
		Priority:        model.IssuePriorityNormal,
		Resolution:      model.IssueResolutionNone,
		ReportedBy:      &service.PartialUser{ID: model.MustNewID(model.ResourceTypeUser)},
		Assignments:     []service.PartialAssignee{},
		Labels:          []service.PartialLabel{},
		CommentCount:    convert.ToPointer(int64(0)),
		DocumentCount:   convert.ToPointer(int64(0)),
		AttachmentCount: convert.ToPointer(int64(0)),
		WatcherCount:    convert.ToPointer(int64(0)),
		RelationCount:   convert.ToPointer(int64(0)),
		Links:           []model.IssueLink{},
		CreatedAt:       convert.ToPointer(time.Now().UTC()),
	}
}

func newServicePartialIssue(issue *service.Issue) *service.PartialIssue {
	return &service.PartialIssue{
		ID:          issue.ID,
		Key:         issue.Key,
		NumericID:   issue.NumericID,
		Kind:        issue.Kind,
		Title:       issue.Title,
		Description: issue.Description,
		Status:      issue.Status,
		Priority:    issue.Priority,
		Assignments: issue.Assignments,
		Labels:      issue.Labels,
		DueDate:     issue.DueDate,
		StartDate:   issue.StartDate,
		ReportedBy:  issue.ReportedBy,
	}
}

func TestNewIssueController(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		c, err := NewIssueController(mocksvc.NewMockIssueService(ctrl))
		require.NoError(t, err)
		assert.NotNil(t, c)
	})

	t.Run("missing issue service", func(t *testing.T) {
		t.Parallel()
		_, err := NewIssueController(nil)
		assert.ErrorIs(t, err, ErrNoIssueService)
	})
}

func TestIssueController_V1ProjectsIssuesCreate(t *testing.T) {
	t.Parallel()

	projectID := model.MustNewID(model.ResourceTypeProject)
	issue := newServiceIssue()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		is := mocksvc.NewMockIssueService(ctrl)
		is.EXPECT().Create(gomock.Any(), projectID, service.CreateIssueOpts{
			Kind:  model.IssueKindStory,
			Title: "Implement authentication",
		}).Return(issue, nil)

		c := newTestIssueController(t, is)
		resp, err := c.V1ProjectsIssuesCreate(context.Background(), api.V1ProjectsIssuesCreateRequestObject{
			Id: projectID.String(),
			Body: &api.V1ProjectsIssuesCreateJSONRequestBody{
				Kind:  api.IssueKindStory,
				Title: "Implement authentication",
			},
		})
		require.NoError(t, err)
		got, ok := resp.(api.V1ProjectsIssuesCreate201JSONResponse)
		require.True(t, ok)
		assert.Equal(t, issue.ID.String(), got.Id)
		assert.Equal(t, issue.Title, got.Title)
	})

	t.Run("bad project id", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		c := newTestIssueController(t, mocksvc.NewMockIssueService(ctrl))
		resp, err := c.V1ProjectsIssuesCreate(context.Background(), api.V1ProjectsIssuesCreateRequestObject{
			Id:   "not-a-xid",
			Body: &api.V1ProjectsIssuesCreateJSONRequestBody{Kind: api.IssueKindStory, Title: "Implement authentication"},
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1ProjectsIssuesCreate400JSONResponse)
		assert.True(t, ok)
	})

	t.Run("invalid kind", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		c := newTestIssueController(t, mocksvc.NewMockIssueService(ctrl))
		resp, err := c.V1ProjectsIssuesCreate(context.Background(), api.V1ProjectsIssuesCreateRequestObject{
			Id: projectID.String(),
			Body: &api.V1ProjectsIssuesCreateJSONRequestBody{
				Kind:  api.IssueKind("bogus"),
				Title: "Implement authentication",
			},
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1ProjectsIssuesCreate400JSONResponse)
		assert.True(t, ok)
	})

	t.Run("no permission", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		is := mocksvc.NewMockIssueService(ctrl)
		is.EXPECT().Create(gomock.Any(), projectID, gomock.Any()).Return(nil, service.ErrNoPermission)

		c := newTestIssueController(t, is)
		resp, err := c.V1ProjectsIssuesCreate(context.Background(), api.V1ProjectsIssuesCreateRequestObject{
			Id:   projectID.String(),
			Body: &api.V1ProjectsIssuesCreateJSONRequestBody{Kind: api.IssueKindStory, Title: "Implement authentication"},
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1ProjectsIssuesCreate403JSONResponse)
		assert.True(t, ok)
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		is := mocksvc.NewMockIssueService(ctrl)
		is.EXPECT().Create(gomock.Any(), projectID, gomock.Any()).Return(nil, errors.Join(service.ErrIssueCreate, repository.ErrNotFound))

		c := newTestIssueController(t, is)
		resp, err := c.V1ProjectsIssuesCreate(context.Background(), api.V1ProjectsIssuesCreateRequestObject{
			Id:   projectID.String(),
			Body: &api.V1ProjectsIssuesCreateJSONRequestBody{Kind: api.IssueKindStory, Title: "Implement authentication"},
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1ProjectsIssuesCreate404JSONResponse)
		assert.True(t, ok)
	})

	t.Run("license expired", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		is := mocksvc.NewMockIssueService(ctrl)
		is.EXPECT().Create(gomock.Any(), projectID, gomock.Any()).Return(nil, license.ErrLicenseExpired)

		c := newTestIssueController(t, is)
		resp, err := c.V1ProjectsIssuesCreate(context.Background(), api.V1ProjectsIssuesCreateRequestObject{
			Id:   projectID.String(),
			Body: &api.V1ProjectsIssuesCreateJSONRequestBody{Kind: api.IssueKindStory, Title: "Implement authentication"},
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1ProjectsIssuesCreate403JSONResponse)
		assert.True(t, ok)
	})

	t.Run("nil body", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		c := newTestIssueController(t, mocksvc.NewMockIssueService(ctrl))
		resp, err := c.V1ProjectsIssuesCreate(context.Background(), api.V1ProjectsIssuesCreateRequestObject{
			Id:   projectID.String(),
			Body: nil,
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1ProjectsIssuesCreate400JSONResponse)
		assert.True(t, ok)
	})
}

func TestIssueController_V1ProjectsIssuesGet(t *testing.T) {
	t.Parallel()

	projectID := model.MustNewID(model.ResourceTypeProject)
	issue := newServiceIssue()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		is := mocksvc.NewMockIssueService(ctrl)
		is.EXPECT().List(gomock.Any(), projectID, gomock.Any(), gomock.Any()).Return(service.Page[*service.PartialIssue]{Items: []*service.PartialIssue{newServicePartialIssue(issue)}}, nil)

		c := newTestIssueController(t, is)
		resp, err := c.V1ProjectsIssuesGet(context.Background(), api.V1ProjectsIssuesGetRequestObject{
			Id:     projectID.String(),
			Params: api.V1ProjectsIssuesGetParams{},
		})
		require.NoError(t, err)
		got, ok := resp.(api.V1ProjectsIssuesGet200JSONResponse)
		require.True(t, ok)
		require.Len(t, got.Items, 1)
		assert.Equal(t, issue.ID.String(), got.Items[0].Id)
		assert.Equal(t, issue.Title, got.Items[0].Title)
	})

	t.Run("bad project id", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		c := newTestIssueController(t, mocksvc.NewMockIssueService(ctrl))
		resp, err := c.V1ProjectsIssuesGet(context.Background(), api.V1ProjectsIssuesGetRequestObject{
			Id: "bad",
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1ProjectsIssuesGet400JSONResponse)
		assert.True(t, ok)
	})

	t.Run("no permission", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		is := mocksvc.NewMockIssueService(ctrl)
		is.EXPECT().List(gomock.Any(), projectID, gomock.Any(), gomock.Any()).Return(service.Page[*service.PartialIssue]{}, service.ErrNoPermission)

		c := newTestIssueController(t, is)
		resp, err := c.V1ProjectsIssuesGet(context.Background(), api.V1ProjectsIssuesGetRequestObject{
			Id: projectID.String(),
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1ProjectsIssuesGet403JSONResponse)
		assert.True(t, ok)
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		is := mocksvc.NewMockIssueService(ctrl)
		is.EXPECT().List(gomock.Any(), projectID, gomock.Any(), gomock.Any()).Return(service.Page[*service.PartialIssue]{}, errors.Join(service.ErrIssueList, repository.ErrNotFound))

		c := newTestIssueController(t, is)
		resp, err := c.V1ProjectsIssuesGet(context.Background(), api.V1ProjectsIssuesGetRequestObject{
			Id: projectID.String(),
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1ProjectsIssuesGet404JSONResponse)
		assert.True(t, ok)
	})

	t.Run("invalid page size", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		c := newTestIssueController(t, mocksvc.NewMockIssueService(ctrl))
		resp, err := c.V1ProjectsIssuesGet(context.Background(), api.V1ProjectsIssuesGetRequestObject{
			Id: projectID.String(),
			Params: api.V1ProjectsIssuesGetParams{
				PageSize: convert.ToPointer(repository.MaxPageSize + 1),
			},
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1ProjectsIssuesGet400JSONResponse)
		assert.True(t, ok)
	})

	t.Run("invalid cursor", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		is := mocksvc.NewMockIssueService(ctrl)
		is.EXPECT().List(gomock.Any(), projectID, gomock.Any(), gomock.Any()).Return(
			service.Page[*service.PartialIssue]{},
			errors.Join(service.ErrIssueList, repository.ErrInvalidCursor),
		)

		c := newTestIssueController(t, is)
		resp, err := c.V1ProjectsIssuesGet(context.Background(), api.V1ProjectsIssuesGetRequestObject{
			Id: projectID.String(),
			Params: api.V1ProjectsIssuesGetParams{
				PageToken: convert.ToPointer("not-a-token"),
			},
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1ProjectsIssuesGet400JSONResponse)
		assert.True(t, ok)
	})
}

func TestIssueController_V1NamespacesIssuesGet(t *testing.T) {
	t.Parallel()

	namespaceID := model.MustNewID(model.ResourceTypeNamespace)
	issue := newServiceIssue()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		is := mocksvc.NewMockIssueService(ctrl)
		is.EXPECT().ListByNamespace(gomock.Any(), namespaceID, gomock.Any(), gomock.Any()).Return(service.Page[*service.PartialIssue]{Items: []*service.PartialIssue{newServicePartialIssue(issue)}}, nil)

		c := newTestIssueController(t, is)
		resp, err := c.V1NamespacesIssuesGet(context.Background(), api.V1NamespacesIssuesGetRequestObject{
			Id:     namespaceID.String(),
			Params: api.V1NamespacesIssuesGetParams{},
		})
		require.NoError(t, err)
		got, ok := resp.(api.V1NamespacesIssuesGet200JSONResponse)
		require.True(t, ok)
		require.Len(t, got.Items, 1)
		assert.Equal(t, issue.ID.String(), got.Items[0].Id)
		assert.Equal(t, issue.Title, got.Items[0].Title)
	})

	t.Run("bad namespace id", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		c := newTestIssueController(t, mocksvc.NewMockIssueService(ctrl))
		resp, err := c.V1NamespacesIssuesGet(context.Background(), api.V1NamespacesIssuesGetRequestObject{
			Id: "bad",
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1NamespacesIssuesGet400JSONResponse)
		assert.True(t, ok)
	})

	t.Run("no permission", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		is := mocksvc.NewMockIssueService(ctrl)
		is.EXPECT().ListByNamespace(gomock.Any(), namespaceID, gomock.Any(), gomock.Any()).Return(service.Page[*service.PartialIssue]{}, service.ErrNoPermission)

		c := newTestIssueController(t, is)
		resp, err := c.V1NamespacesIssuesGet(context.Background(), api.V1NamespacesIssuesGetRequestObject{
			Id: namespaceID.String(),
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1NamespacesIssuesGet403JSONResponse)
		assert.True(t, ok)
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		is := mocksvc.NewMockIssueService(ctrl)
		is.EXPECT().ListByNamespace(gomock.Any(), namespaceID, gomock.Any(), gomock.Any()).Return(service.Page[*service.PartialIssue]{}, errors.Join(service.ErrIssueList, repository.ErrNotFound))

		c := newTestIssueController(t, is)
		resp, err := c.V1NamespacesIssuesGet(context.Background(), api.V1NamespacesIssuesGetRequestObject{
			Id: namespaceID.String(),
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1NamespacesIssuesGet404JSONResponse)
		assert.True(t, ok)
	})
}

func TestIssueController_V1UsersIssuesGet(t *testing.T) {
	t.Parallel()

	userID := model.MustNewID(model.ResourceTypeUser)
	issue := newServiceIssue()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		is := mocksvc.NewMockIssueService(ctrl)
		is.EXPECT().ListByUser(gomock.Any(), userID, gomock.Any(), gomock.Any()).Return(service.Page[*service.PartialIssue]{Items: []*service.PartialIssue{newServicePartialIssue(issue)}}, nil)

		c := newTestIssueController(t, is)
		resp, err := c.V1UsersIssuesGet(context.Background(), api.V1UsersIssuesGetRequestObject{
			Id:     userID.String(),
			Params: api.V1UsersIssuesGetParams{},
		})
		require.NoError(t, err)
		got, ok := resp.(api.V1UsersIssuesGet200JSONResponse)
		require.True(t, ok)
		require.Len(t, got.Items, 1)
		assert.Equal(t, issue.ID.String(), got.Items[0].Id)
		assert.Equal(t, issue.Title, got.Items[0].Title)
	})

	t.Run("bad user id", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		c := newTestIssueController(t, mocksvc.NewMockIssueService(ctrl))
		resp, err := c.V1UsersIssuesGet(context.Background(), api.V1UsersIssuesGetRequestObject{
			Id: "bad",
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1UsersIssuesGet400JSONResponse)
		assert.True(t, ok)
	})

	t.Run("no permission", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		is := mocksvc.NewMockIssueService(ctrl)
		is.EXPECT().ListByUser(gomock.Any(), userID, gomock.Any(), gomock.Any()).Return(service.Page[*service.PartialIssue]{}, service.ErrNoPermission)

		c := newTestIssueController(t, is)
		resp, err := c.V1UsersIssuesGet(context.Background(), api.V1UsersIssuesGetRequestObject{
			Id: userID.String(),
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1UsersIssuesGet403JSONResponse)
		assert.True(t, ok)
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		is := mocksvc.NewMockIssueService(ctrl)
		is.EXPECT().ListByUser(gomock.Any(), userID, gomock.Any(), gomock.Any()).Return(service.Page[*service.PartialIssue]{}, errors.Join(service.ErrIssueList, repository.ErrNotFound))

		c := newTestIssueController(t, is)
		resp, err := c.V1UsersIssuesGet(context.Background(), api.V1UsersIssuesGetRequestObject{
			Id: userID.String(),
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1UsersIssuesGet404JSONResponse)
		assert.True(t, ok)
	})
}

func TestIssueController_V1IssueGet(t *testing.T) {
	t.Parallel()

	issue := newServiceIssue()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		is := mocksvc.NewMockIssueService(ctrl)
		is.EXPECT().Get(gomock.Any(), issue.ID).Return(issue, nil)

		c := newTestIssueController(t, is)
		resp, err := c.V1IssueGet(context.Background(), api.V1IssueGetRequestObject{Id: issue.ID.String()})
		require.NoError(t, err)
		got, ok := resp.(api.V1IssueGet200JSONResponse)
		require.True(t, ok)
		assert.Equal(t, issue.ID.String(), got.Id)
		assert.Equal(t, issue.Key, got.Key)
		assert.Equal(t, issue.Title, got.Title)
		require.NotNil(t, got.Description)
		assert.Equal(t, issue.Description, *got.Description)
	})

	t.Run("bad id", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		c := newTestIssueController(t, mocksvc.NewMockIssueService(ctrl))
		resp, err := c.V1IssueGet(context.Background(), api.V1IssueGetRequestObject{Id: "bad"})
		require.NoError(t, err)
		_, ok := resp.(api.V1IssueGet400JSONResponse)
		assert.True(t, ok)
	})

	t.Run("no permission", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		is := mocksvc.NewMockIssueService(ctrl)
		is.EXPECT().Get(gomock.Any(), issue.ID).Return(nil, service.ErrNoPermission)

		c := newTestIssueController(t, is)
		resp, err := c.V1IssueGet(context.Background(), api.V1IssueGetRequestObject{Id: issue.ID.String()})
		require.NoError(t, err)
		_, ok := resp.(api.V1IssueGet403JSONResponse)
		assert.True(t, ok)
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		is := mocksvc.NewMockIssueService(ctrl)
		is.EXPECT().Get(gomock.Any(), issue.ID).Return(nil, errors.Join(service.ErrIssueGet, repository.ErrNotFound))

		c := newTestIssueController(t, is)
		resp, err := c.V1IssueGet(context.Background(), api.V1IssueGetRequestObject{Id: issue.ID.String()})
		require.NoError(t, err)
		_, ok := resp.(api.V1IssueGet404JSONResponse)
		assert.True(t, ok)
	})
}

func TestIssueController_V1NamespacesIssuesKeyGet(t *testing.T) {
	t.Parallel()

	issue := newServiceIssue()
	namespaceID := model.MustNewID(model.ResourceTypeNamespace)

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		is := mocksvc.NewMockIssueService(ctrl)
		is.EXPECT().GetByKey(gomock.Any(), namespaceID, issue.Key).Return(issue, nil)

		c := newTestIssueController(t, is)
		resp, err := c.V1NamespacesIssuesKeyGet(context.Background(), api.V1NamespacesIssuesKeyGetRequestObject{
			Id:  namespaceID.String(),
			Key: issue.Key,
		})
		require.NoError(t, err)
		got, ok := resp.(api.V1NamespacesIssuesKeyGet200JSONResponse)
		require.True(t, ok)
		assert.Equal(t, issue.ID.String(), got.Id)
		assert.Equal(t, issue.Key, got.Key)
	})

	t.Run("bad namespace id", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		c := newTestIssueController(t, mocksvc.NewMockIssueService(ctrl))
		resp, err := c.V1NamespacesIssuesKeyGet(context.Background(), api.V1NamespacesIssuesKeyGetRequestObject{
			Id:  "bad",
			Key: issue.Key,
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1NamespacesIssuesKeyGet400JSONResponse)
		assert.True(t, ok)
	})

	t.Run("bad key", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		c := newTestIssueController(t, mocksvc.NewMockIssueService(ctrl))
		resp, err := c.V1NamespacesIssuesKeyGet(context.Background(), api.V1NamespacesIssuesKeyGetRequestObject{
			Id:  namespaceID.String(),
			Key: "bad",
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1NamespacesIssuesKeyGet400JSONResponse)
		assert.True(t, ok)
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		is := mocksvc.NewMockIssueService(ctrl)
		is.EXPECT().GetByKey(gomock.Any(), namespaceID, issue.Key).Return(nil, errors.Join(service.ErrIssueGet, repository.ErrNotFound))

		c := newTestIssueController(t, is)
		resp, err := c.V1NamespacesIssuesKeyGet(context.Background(), api.V1NamespacesIssuesKeyGetRequestObject{
			Id:  namespaceID.String(),
			Key: issue.Key,
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1NamespacesIssuesKeyGet404JSONResponse)
		assert.True(t, ok)
	})
}

func TestIssueController_V1IssueUpdate(t *testing.T) {
	t.Parallel()

	issue := newServiceIssue()
	title := "Updated title"

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		updated := *issue
		updated.Title = title

		is := mocksvc.NewMockIssueService(ctrl)
		is.EXPECT().Update(gomock.Any(), issue.ID, service.UpdateIssueOpts{
			Title: optional.Some(title),
		}).Return(&updated, nil)

		c := newTestIssueController(t, is)
		resp, err := c.V1IssueUpdate(context.Background(), api.V1IssueUpdateRequestObject{
			Id:   issue.ID.String(),
			Body: &api.V1IssueUpdateJSONRequestBody{Title: optional.Some(title)},
		})
		require.NoError(t, err)
		got, ok := resp.(api.V1IssueUpdate200JSONResponse)
		require.True(t, ok)
		assert.Equal(t, title, got.Title)
	})

	t.Run("bad id", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		c := newTestIssueController(t, mocksvc.NewMockIssueService(ctrl))
		resp, err := c.V1IssueUpdate(context.Background(), api.V1IssueUpdateRequestObject{
			Id:   "bad",
			Body: &api.V1IssueUpdateJSONRequestBody{Title: optional.Some(title)},
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1IssueUpdate400JSONResponse)
		assert.True(t, ok)
	})

	t.Run("nil body", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		c := newTestIssueController(t, mocksvc.NewMockIssueService(ctrl))
		resp, err := c.V1IssueUpdate(context.Background(), api.V1IssueUpdateRequestObject{
			Id:   issue.ID.String(),
			Body: nil,
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1IssueUpdate400JSONResponse)
		assert.True(t, ok)
	})

	t.Run("no permission", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		is := mocksvc.NewMockIssueService(ctrl)
		is.EXPECT().Update(gomock.Any(), issue.ID, gomock.Any()).Return(nil, service.ErrNoPermission)

		c := newTestIssueController(t, is)
		resp, err := c.V1IssueUpdate(context.Background(), api.V1IssueUpdateRequestObject{
			Id:   issue.ID.String(),
			Body: &api.V1IssueUpdateJSONRequestBody{Title: optional.Some(title)},
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1IssueUpdate403JSONResponse)
		assert.True(t, ok)
	})

	t.Run("quota exceeded", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		is := mocksvc.NewMockIssueService(ctrl)
		is.EXPECT().Update(gomock.Any(), issue.ID, gomock.Any()).Return(nil, service.ErrQuotaExceeded)

		c := newTestIssueController(t, is)
		resp, err := c.V1IssueUpdate(context.Background(), api.V1IssueUpdateRequestObject{
			Id:   issue.ID.String(),
			Body: &api.V1IssueUpdateJSONRequestBody{Title: optional.Some(title)},
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1IssueUpdate403JSONResponse)
		assert.True(t, ok)
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		is := mocksvc.NewMockIssueService(ctrl)
		is.EXPECT().Update(gomock.Any(), issue.ID, gomock.Any()).Return(nil, errors.Join(service.ErrIssueUpdate, repository.ErrNotFound))

		c := newTestIssueController(t, is)
		resp, err := c.V1IssueUpdate(context.Background(), api.V1IssueUpdateRequestObject{
			Id:   issue.ID.String(),
			Body: &api.V1IssueUpdateJSONRequestBody{Title: optional.Some(title)},
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1IssueUpdate404JSONResponse)
		assert.True(t, ok)
	})

	t.Run("set parent", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		parentID := model.MustNewID(model.ResourceTypeIssue)
		updated := *issue
		updated.Parent = &service.PartialIssue{
			ID:          parentID,
			Kind:        model.IssueKindEpic,
			Title:       "Parent epic",
			Status:      model.IssueStatusOpen,
			Priority:    model.IssuePriorityHigh,
			Assignments: make([]service.PartialAssignee, 0),
			Labels:      make([]service.PartialLabel, 0),
		}

		is := mocksvc.NewMockIssueService(ctrl)
		is.EXPECT().Update(gomock.Any(), issue.ID, service.UpdateIssueOpts{
			Parent: optional.Some(parentID),
		}).Return(&updated, nil)

		c := newTestIssueController(t, is)
		resp, err := c.V1IssueUpdate(context.Background(), api.V1IssueUpdateRequestObject{
			Id:   issue.ID.String(),
			Body: &api.V1IssueUpdateJSONRequestBody{Parent: optional.Some(parentID.String())},
		})
		require.NoError(t, err)
		got, ok := resp.(api.V1IssueUpdate200JSONResponse)
		require.True(t, ok)
		require.NotNil(t, got.Parent)
		assert.Equal(t, parentID.String(), got.Parent.Id)
	})

	t.Run("clear parent", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		updated := *issue
		updated.Parent = nil

		is := mocksvc.NewMockIssueService(ctrl)
		is.EXPECT().Update(gomock.Any(), issue.ID, service.UpdateIssueOpts{
			Parent: optional.Null[model.ID](),
		}).Return(&updated, nil)

		c := newTestIssueController(t, is)
		resp, err := c.V1IssueUpdate(context.Background(), api.V1IssueUpdateRequestObject{
			Id:   issue.ID.String(),
			Body: &api.V1IssueUpdateJSONRequestBody{Parent: optional.Null[string]()},
		})
		require.NoError(t, err)
		got, ok := resp.(api.V1IssueUpdate200JSONResponse)
		require.True(t, ok)
		assert.Nil(t, got.Parent)
	})

	t.Run("self-parent", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		is := mocksvc.NewMockIssueService(ctrl)
		is.EXPECT().Update(gomock.Any(), issue.ID, service.UpdateIssueOpts{
			Parent: optional.Some(issue.ID),
		}).Return(nil, errors.Join(service.ErrIssueUpdate, service.ErrIssueSelfRelation))

		c := newTestIssueController(t, is)
		resp, err := c.V1IssueUpdate(context.Background(), api.V1IssueUpdateRequestObject{
			Id:   issue.ID.String(),
			Body: &api.V1IssueUpdateJSONRequestBody{Parent: optional.Some(issue.ID.String())},
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1IssueUpdate400JSONResponse)
		assert.True(t, ok)
	})
}

func TestIssueController_V1IssueDelete(t *testing.T) {
	t.Parallel()

	issueID := model.MustNewID(model.ResourceTypeIssue)

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		is := mocksvc.NewMockIssueService(ctrl)
		is.EXPECT().Delete(gomock.Any(), issueID).Return(nil)

		c := newTestIssueController(t, is)
		resp, err := c.V1IssueDelete(context.Background(), api.V1IssueDeleteRequestObject{Id: issueID.String()})
		require.NoError(t, err)
		_, ok := resp.(api.V1IssueDelete204Response)
		assert.True(t, ok)
	})

	t.Run("bad id", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		c := newTestIssueController(t, mocksvc.NewMockIssueService(ctrl))
		resp, err := c.V1IssueDelete(context.Background(), api.V1IssueDeleteRequestObject{Id: "bad"})
		require.NoError(t, err)
		_, ok := resp.(api.V1IssueDelete400JSONResponse)
		assert.True(t, ok)
	})

	t.Run("no permission", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		is := mocksvc.NewMockIssueService(ctrl)
		is.EXPECT().Delete(gomock.Any(), issueID).Return(service.ErrNoPermission)

		c := newTestIssueController(t, is)
		resp, err := c.V1IssueDelete(context.Background(), api.V1IssueDeleteRequestObject{Id: issueID.String()})
		require.NoError(t, err)
		_, ok := resp.(api.V1IssueDelete403JSONResponse)
		assert.True(t, ok)
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		is := mocksvc.NewMockIssueService(ctrl)
		is.EXPECT().Delete(gomock.Any(), issueID).Return(errors.Join(service.ErrIssueDelete, repository.ErrNotFound))

		c := newTestIssueController(t, is)
		resp, err := c.V1IssueDelete(context.Background(), api.V1IssueDeleteRequestObject{Id: issueID.String()})
		require.NoError(t, err)
		_, ok := resp.(api.V1IssueDelete404JSONResponse)
		assert.True(t, ok)
	})
}

func TestIssueToDTO(t *testing.T) {
	t.Parallel()

	issue := newServiceIssue()
	assigneeID := model.MustNewID(model.ResourceTypeUser)
	reviewerID := model.MustNewID(model.ResourceTypeUser)
	issue.Assignments = []service.PartialAssignee{
		{ID: assigneeID, Kind: model.AssignmentKindAssignee, FirstName: "Ada", LastName: "Lovelace"},
		{ID: reviewerID, Kind: model.AssignmentKindReviewer, FirstName: "Grace", LastName: "Hopper"},
	}
	project := &service.PartialProject{
		ID:     model.MustNewID(model.ResourceTypeProject),
		Key:    "ENG",
		Name:   "Engineering",
		Status: model.ProjectStatusActive,
	}
	namespace := &service.PartialNamespace{
		ID:   model.MustNewID(model.ResourceTypeNamespace),
		Name: "Platform",
	}
	issue.Project = project
	issue.Namespace = namespace
	parent := &service.PartialIssue{
		ID:          model.MustNewID(model.ResourceTypeIssue),
		Key:         "ENG-1",
		NumericID:   1,
		Kind:        model.IssueKindEpic,
		Title:       "Parent epic",
		Status:      model.IssueStatusOpen,
		Priority:    model.IssuePriorityHigh,
		Assignments: make([]service.PartialAssignee, 0),
		Labels:      make([]service.PartialLabel, 0),
	}
	issue.Parent = parent

	dto := issueToDTO(issue)
	assert.Equal(t, issue.ID.String(), dto.Id)
	assert.Equal(t, issue.Key, dto.Key)
	assert.Equal(t, int(issue.NumericID), dto.NumericId)
	assert.Equal(t, issue.Title, dto.Title)
	assert.Equal(t, []api.PartialUser{{
		Id:        assigneeID.String(),
		FirstName: "Ada",
		LastName:  "Lovelace",
	}}, dto.Assignees)
	assert.Equal(t, []api.PartialUser{{
		Id:        reviewerID.String(),
		FirstName: "Grace",
		LastName:  "Hopper",
	}}, dto.Reviewers)
	require.NotNil(t, dto.Project)
	assert.Equal(t, project.ID.String(), dto.Project.Id)
	assert.Equal(t, project.Key, dto.Project.Key)
	assert.Equal(t, project.Name, dto.Project.Name)
	require.NotNil(t, dto.Namespace)
	assert.Equal(t, namespace.ID.String(), dto.Namespace.Id)
	assert.Equal(t, namespace.Name, dto.Namespace.Name)
	require.NotNil(t, dto.Parent)
	assert.Equal(t, parent.ID.String(), dto.Parent.Id)
	assert.Equal(t, parent.Key, dto.Parent.Key)
	assert.Equal(t, parent.Title, dto.Parent.Title)
	assert.Empty(t, dto.Parent.Reviewers)
	require.NotNil(t, dto.CommentCount)
	assert.Equal(t, int64(0), *dto.CommentCount)
	require.NotNil(t, dto.DocumentCount)
	assert.Equal(t, int64(0), *dto.DocumentCount)

	t.Run("nil created at does not panic", func(t *testing.T) {
		t.Parallel()

		issue := newServiceIssue()
		issue.CreatedAt = nil
		dto := issueToDTO(issue)
		assert.True(t, dto.CreatedAt.IsZero())
	})
}

func TestUpdateIssueJSONRequestBodyToUpdateIssueOpts(t *testing.T) {
	t.Parallel()

	assigneeID := model.MustNewID(model.ResourceTypeUser)
	reviewerID := model.MustNewID(model.ResourceTypeUser)

	t.Run("maps assignees and reviewers", func(t *testing.T) {
		t.Parallel()

		opts, err := updateIssueJSONRequestBodyToUpdateIssueOpts(&api.V1IssueUpdateJSONRequestBody{
			Assignees: optional.Some([]string{assigneeID.String()}),
			Reviewers: optional.Some([]string{reviewerID.String()}),
		})
		require.NoError(t, err)
		require.True(t, opts.Assignees.Defined)
		require.NotNil(t, opts.Assignees.Value)
		assert.Equal(t, []model.ID{assigneeID}, *opts.Assignees.Value)
		require.True(t, opts.Reviewers.Defined)
		require.NotNil(t, opts.Reviewers.Value)
		assert.Equal(t, []model.ID{reviewerID}, *opts.Reviewers.Value)
	})

	t.Run("empty arrays clear assignments", func(t *testing.T) {
		t.Parallel()

		opts, err := updateIssueJSONRequestBodyToUpdateIssueOpts(&api.V1IssueUpdateJSONRequestBody{
			Assignees: optional.Some([]string{}),
			Reviewers: optional.Some([]string{}),
		})
		require.NoError(t, err)
		require.True(t, opts.Assignees.Defined)
		require.NotNil(t, opts.Assignees.Value)
		assert.Empty(t, *opts.Assignees.Value)
		require.True(t, opts.Reviewers.Defined)
		require.NotNil(t, opts.Reviewers.Value)
		assert.Empty(t, *opts.Reviewers.Value)
	})

	t.Run("omitted assignment fields stay undefined", func(t *testing.T) {
		t.Parallel()

		opts, err := updateIssueJSONRequestBodyToUpdateIssueOpts(&api.V1IssueUpdateJSONRequestBody{
			Title: optional.Some("Updated title"),
		})
		require.NoError(t, err)
		assert.False(t, opts.Assignees.Defined)
		assert.False(t, opts.Reviewers.Defined)
	})

	t.Run("maps parent", func(t *testing.T) {
		t.Parallel()

		parentID := model.MustNewID(model.ResourceTypeIssue)
		opts, err := updateIssueJSONRequestBodyToUpdateIssueOpts(&api.V1IssueUpdateJSONRequestBody{
			Parent: optional.Some(parentID.String()),
		})
		require.NoError(t, err)
		require.True(t, opts.Parent.Defined)
		require.NotNil(t, opts.Parent.Value)
		assert.Equal(t, parentID, *opts.Parent.Value)
	})

	t.Run("null parent clears", func(t *testing.T) {
		t.Parallel()

		opts, err := updateIssueJSONRequestBodyToUpdateIssueOpts(&api.V1IssueUpdateJSONRequestBody{
			Parent: optional.Null[string](),
		})
		require.NoError(t, err)
		require.True(t, opts.Parent.Defined)
		assert.Nil(t, opts.Parent.Value)
	})

	t.Run("omitted parent stays undefined", func(t *testing.T) {
		t.Parallel()

		opts, err := updateIssueJSONRequestBodyToUpdateIssueOpts(&api.V1IssueUpdateJSONRequestBody{
			Title: optional.Some("Updated title"),
		})
		require.NoError(t, err)
		assert.False(t, opts.Parent.Defined)
	})
}

func newServiceIssueRelation(issue *service.Issue) *service.IssueRelation {
	return &service.IssueRelation{
		ID:        model.MustNewID(model.ResourceTypeIssueRelation),
		Kind:      model.IssueRelationKindBlocks,
		Direction: service.IssueRelationDirectionOutgoing,
		Related:   newServicePartialIssue(issue),
		CreatedAt: convert.ToPointer(time.Now().UTC()),
	}
}

func TestIssueController_V1IssueRelationsGet(t *testing.T) {
	t.Parallel()

	issue := newServiceIssue()
	relation := newServiceIssueRelation(issue)

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		is := mocksvc.NewMockIssueService(ctrl)
		is.EXPECT().ListRelations(gomock.Any(), issue.ID, gomock.Any()).Return(service.Page[*service.IssueRelation]{
			Items: []*service.IssueRelation{relation},
		}, nil)

		c := newTestIssueController(t, is)
		resp, err := c.V1IssueRelationsGet(context.Background(), api.V1IssueRelationsGetRequestObject{
			Id:     issue.ID.String(),
			Params: api.V1IssueRelationsGetParams{},
		})
		require.NoError(t, err)
		got, ok := resp.(api.V1IssueRelationsGet200JSONResponse)
		require.True(t, ok)
		require.Len(t, got.Items, 1)
		assert.Equal(t, relation.ID.String(), got.Items[0].Id)
		assert.Equal(t, api.IssueRelationKindBlocks, got.Items[0].Kind)
		assert.Equal(t, api.IssueRelationDirectionOutgoing, got.Items[0].Direction)
	})

	t.Run("bad issue id", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		c := newTestIssueController(t, mocksvc.NewMockIssueService(ctrl))
		resp, err := c.V1IssueRelationsGet(context.Background(), api.V1IssueRelationsGetRequestObject{Id: "bad"})
		require.NoError(t, err)
		_, ok := resp.(api.V1IssueRelationsGet400JSONResponse)
		assert.True(t, ok)
	})

	t.Run("no permission", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		is := mocksvc.NewMockIssueService(ctrl)
		is.EXPECT().ListRelations(gomock.Any(), issue.ID, gomock.Any()).Return(service.Page[*service.IssueRelation]{}, service.ErrNoPermission)

		c := newTestIssueController(t, is)
		resp, err := c.V1IssueRelationsGet(context.Background(), api.V1IssueRelationsGetRequestObject{Id: issue.ID.String()})
		require.NoError(t, err)
		_, ok := resp.(api.V1IssueRelationsGet403JSONResponse)
		assert.True(t, ok)
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		is := mocksvc.NewMockIssueService(ctrl)
		is.EXPECT().ListRelations(gomock.Any(), issue.ID, gomock.Any()).Return(
			service.Page[*service.IssueRelation]{},
			errors.Join(service.ErrIssueGetRelations, repository.ErrNotFound),
		)

		c := newTestIssueController(t, is)
		resp, err := c.V1IssueRelationsGet(context.Background(), api.V1IssueRelationsGetRequestObject{Id: issue.ID.String()})
		require.NoError(t, err)
		_, ok := resp.(api.V1IssueRelationsGet404JSONResponse)
		assert.True(t, ok)
	})
}

func TestIssueController_V1IssueRelationsCreate(t *testing.T) {
	t.Parallel()

	issue := newServiceIssue()
	related := newServiceIssue()
	relation := newServiceIssueRelation(related)

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		is := mocksvc.NewMockIssueService(ctrl)
		is.EXPECT().AddRelation(gomock.Any(), issue.ID, related.ID, model.IssueRelationKindBlocks).Return(relation, nil)

		c := newTestIssueController(t, is)
		resp, err := c.V1IssueRelationsCreate(context.Background(), api.V1IssueRelationsCreateRequestObject{
			Id: issue.ID.String(),
			Body: &api.V1IssueRelationsCreateJSONRequestBody{
				RelatedId: related.ID.String(),
				Kind:      api.IssueRelationKindBlocks,
			},
		})
		require.NoError(t, err)
		got, ok := resp.(api.V1IssueRelationsCreate201JSONResponse)
		require.True(t, ok)
		assert.Equal(t, relation.ID.String(), got.Id)
		assert.Equal(t, api.IssueRelationKindBlocks, got.Kind)
	})

	t.Run("self relation", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		is := mocksvc.NewMockIssueService(ctrl)
		is.EXPECT().AddRelation(gomock.Any(), issue.ID, issue.ID, model.IssueRelationKindBlocks).Return(
			nil,
			errors.Join(service.ErrIssueAddRelation, service.ErrIssueSelfRelation),
		)

		c := newTestIssueController(t, is)
		resp, err := c.V1IssueRelationsCreate(context.Background(), api.V1IssueRelationsCreateRequestObject{
			Id: issue.ID.String(),
			Body: &api.V1IssueRelationsCreateJSONRequestBody{
				RelatedId: issue.ID.String(),
				Kind:      api.IssueRelationKindBlocks,
			},
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1IssueRelationsCreate400JSONResponse)
		assert.True(t, ok)
	})

	t.Run("no permission", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		is := mocksvc.NewMockIssueService(ctrl)
		is.EXPECT().AddRelation(gomock.Any(), issue.ID, related.ID, model.IssueRelationKindBlocks).Return(nil, service.ErrNoPermission)

		c := newTestIssueController(t, is)
		resp, err := c.V1IssueRelationsCreate(context.Background(), api.V1IssueRelationsCreateRequestObject{
			Id: issue.ID.String(),
			Body: &api.V1IssueRelationsCreateJSONRequestBody{
				RelatedId: related.ID.String(),
				Kind:      api.IssueRelationKindBlocks,
			},
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1IssueRelationsCreate403JSONResponse)
		assert.True(t, ok)
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		is := mocksvc.NewMockIssueService(ctrl)
		is.EXPECT().AddRelation(gomock.Any(), issue.ID, related.ID, model.IssueRelationKindBlocks).Return(
			nil,
			errors.Join(service.ErrIssueAddRelation, repository.ErrNotFound),
		)

		c := newTestIssueController(t, is)
		resp, err := c.V1IssueRelationsCreate(context.Background(), api.V1IssueRelationsCreateRequestObject{
			Id: issue.ID.String(),
			Body: &api.V1IssueRelationsCreateJSONRequestBody{
				RelatedId: related.ID.String(),
				Kind:      api.IssueRelationKindBlocks,
			},
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1IssueRelationsCreate404JSONResponse)
		assert.True(t, ok)
	})
}

func TestIssueController_V1IssueRelationUpdate(t *testing.T) {
	t.Parallel()

	issue := newServiceIssue()
	related := newServiceIssue()
	relation := newServiceIssueRelation(related)
	relationID := relation.ID

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		is := mocksvc.NewMockIssueService(ctrl)
		is.EXPECT().UpdateRelation(gomock.Any(), issue.ID, relationID, model.IssueRelationKindRelatedTo).Return(relation, nil)

		c := newTestIssueController(t, is)
		resp, err := c.V1IssueRelationUpdate(context.Background(), api.V1IssueRelationUpdateRequestObject{
			Id:         issue.ID.String(),
			RelationId: relationID.String(),
			Body:       &api.V1IssueRelationUpdateJSONRequestBody{Kind: api.IssueRelationKindRelatedTo},
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1IssueRelationUpdate200JSONResponse)
		assert.True(t, ok)
	})

	t.Run("reserved kind", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		is := mocksvc.NewMockIssueService(ctrl)
		is.EXPECT().UpdateRelation(gomock.Any(), issue.ID, relationID, model.IssueRelationKindSubtaskOf).Return(
			nil,
			errors.Join(service.ErrIssueUpdateRelation, service.ErrIssueReservedRelationKind),
		)

		c := newTestIssueController(t, is)
		resp, err := c.V1IssueRelationUpdate(context.Background(), api.V1IssueRelationUpdateRequestObject{
			Id:         issue.ID.String(),
			RelationId: relationID.String(),
			Body:       &api.V1IssueRelationUpdateJSONRequestBody{Kind: api.IssueRelationKindSubtaskOf},
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1IssueRelationUpdate400JSONResponse)
		assert.True(t, ok)
	})

	t.Run("no permission", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		is := mocksvc.NewMockIssueService(ctrl)
		is.EXPECT().UpdateRelation(gomock.Any(), issue.ID, relationID, model.IssueRelationKindRelatedTo).Return(nil, service.ErrNoPermission)

		c := newTestIssueController(t, is)
		resp, err := c.V1IssueRelationUpdate(context.Background(), api.V1IssueRelationUpdateRequestObject{
			Id:         issue.ID.String(),
			RelationId: relationID.String(),
			Body:       &api.V1IssueRelationUpdateJSONRequestBody{Kind: api.IssueRelationKindRelatedTo},
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1IssueRelationUpdate403JSONResponse)
		assert.True(t, ok)
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		is := mocksvc.NewMockIssueService(ctrl)
		is.EXPECT().UpdateRelation(gomock.Any(), issue.ID, relationID, model.IssueRelationKindRelatedTo).Return(
			nil,
			errors.Join(service.ErrIssueUpdateRelation, repository.ErrNotFound),
		)

		c := newTestIssueController(t, is)
		resp, err := c.V1IssueRelationUpdate(context.Background(), api.V1IssueRelationUpdateRequestObject{
			Id:         issue.ID.String(),
			RelationId: relationID.String(),
			Body:       &api.V1IssueRelationUpdateJSONRequestBody{Kind: api.IssueRelationKindRelatedTo},
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1IssueRelationUpdate404JSONResponse)
		assert.True(t, ok)
	})
}

func TestIssueController_V1IssueRelationDelete(t *testing.T) {
	t.Parallel()

	issueID := model.MustNewID(model.ResourceTypeIssue)
	relationID := model.MustNewID(model.ResourceTypeIssueRelation)

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		is := mocksvc.NewMockIssueService(ctrl)
		is.EXPECT().RemoveRelation(gomock.Any(), issueID, relationID).Return(nil)

		c := newTestIssueController(t, is)
		resp, err := c.V1IssueRelationDelete(context.Background(), api.V1IssueRelationDeleteRequestObject{
			Id:         issueID.String(),
			RelationId: relationID.String(),
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1IssueRelationDelete204Response)
		assert.True(t, ok)
	})

	t.Run("bad relation id", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		c := newTestIssueController(t, mocksvc.NewMockIssueService(ctrl))
		resp, err := c.V1IssueRelationDelete(context.Background(), api.V1IssueRelationDeleteRequestObject{
			Id:         issueID.String(),
			RelationId: "bad",
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1IssueRelationDelete400JSONResponse)
		assert.True(t, ok)
	})

	t.Run("no permission", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		is := mocksvc.NewMockIssueService(ctrl)
		is.EXPECT().RemoveRelation(gomock.Any(), issueID, relationID).Return(service.ErrNoPermission)

		c := newTestIssueController(t, is)
		resp, err := c.V1IssueRelationDelete(context.Background(), api.V1IssueRelationDeleteRequestObject{
			Id:         issueID.String(),
			RelationId: relationID.String(),
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1IssueRelationDelete403JSONResponse)
		assert.True(t, ok)
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		is := mocksvc.NewMockIssueService(ctrl)
		is.EXPECT().RemoveRelation(gomock.Any(), issueID, relationID).Return(errors.Join(service.ErrIssueRemoveRelation, repository.ErrNotFound))

		c := newTestIssueController(t, is)
		resp, err := c.V1IssueRelationDelete(context.Background(), api.V1IssueRelationDeleteRequestObject{
			Id:         issueID.String(),
			RelationId: relationID.String(),
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1IssueRelationDelete404JSONResponse)
		assert.True(t, ok)
	})
}

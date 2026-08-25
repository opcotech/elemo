package http

import (
	"context"
	"encoding/json"
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

func newServiceDocument() *service.Document {
	libraryID := model.MustNewID(model.ResourceTypeOrganization)
	return &service.Document{
		ID:        model.MustNewID(model.ResourceTypeDocument),
		Title:     "Project Plan",
		Excerpt:   "Overview of the project plan",
		FileID:    "documents/file",
		CreatedBy: service.PartialUser{ID: model.MustNewID(model.ResourceTypeUser)},
		Library: service.DocumentLibrary{
			ID:   libraryID,
			Type: model.ResourceTypeOrganization,
			Name: "Acme",
		},
		Relations:       []service.DocumentRelation{},
		Labels:          []service.PartialLabel{},
		CommentCount:    convert.ToPointer(int64(0)),
		AttachmentCount: convert.ToPointer(int64(0)),
		CreatedAt:       convert.ToPointer(time.Now().UTC()),
		Content:         []byte("# Project Plan\n\nGoals and timeline."),
	}
}

func newServicePartialDocument() *service.PartialDocument {
	return &service.PartialDocument{
		ID:        model.MustNewID(model.ResourceTypeDocument),
		Title:     "Project Plan",
		Excerpt:   "Overview of the project plan",
		CreatedBy: service.PartialUser{ID: model.MustNewID(model.ResourceTypeUser)},
		CreatedAt: convert.ToPointer(time.Now().UTC()),
	}
}

func TestNewDocumentController(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		c, err := NewDocumentController(
			mocksvc.NewMockOrganizationService(ctrl),
			mocksvc.NewMockNamespaceService(ctrl),
			mocksvc.NewMockDocumentService(ctrl),
		)
		require.NoError(t, err)
		assert.NotNil(t, c)
	})

	t.Run("missing document service", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		_, err := NewDocumentController(
			mocksvc.NewMockOrganizationService(ctrl),
			mocksvc.NewMockNamespaceService(ctrl),
			nil,
		)
		assert.ErrorIs(t, err, ErrNoDocumentService)
	})
}

func TestDocumentController_V1ProjectsDocumentsCreate(t *testing.T) {
	t.Parallel()

	projectID := model.MustNewID(model.ResourceTypeProject)
	doc := newServiceDocument()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ds := mocksvc.NewMockDocumentService(ctrl)
		ds.EXPECT().Create(gomock.Any(), projectID, service.CreateDocumentOpts{
			Title:   "Project Plan",
			Excerpt: "Overview of the project plan",
			Content: []byte("# Project Plan\n\nGoals and timeline."),
		}).Return(doc, nil)

		c, _, _ := newTestDocumentController(t, ctrl, ds)
		resp, err := c.V1ProjectsDocumentsCreate(context.Background(), api.V1ProjectsDocumentsCreateRequestObject{
			ProjectId: projectID.String(),
			Body: &api.V1ProjectsDocumentsCreateJSONRequestBody{
				Title:   "Project Plan",
				Excerpt: optional.Some("Overview of the project plan"),
				Content: optional.Some("# Project Plan\n\nGoals and timeline."),
			},
		})
		require.NoError(t, err)
		got, ok := resp.(api.V1ProjectsDocumentsCreate201JSONResponse)
		require.True(t, ok)
		assert.Equal(t, doc.ID.String(), got.Id)
		assert.Equal(t, doc.Title, got.Title)
		assert.Equal(t, string(doc.Content), got.Content)
	})

	t.Run("bad project id", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		c, _, _ := newTestDocumentController(t, ctrl, mocksvc.NewMockDocumentService(ctrl))
		resp, err := c.V1ProjectsDocumentsCreate(context.Background(), api.V1ProjectsDocumentsCreateRequestObject{
			ProjectId: "AB",
			Body: &api.V1ProjectsDocumentsCreateJSONRequestBody{
				Title:   "Project Plan",
				Content: optional.Some("# Project Plan"),
			},
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1ProjectsDocumentsCreate400JSONResponse)
		assert.True(t, ok)
	})

	t.Run("nil body", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		c, _, _ := newTestDocumentController(t, ctrl, mocksvc.NewMockDocumentService(ctrl))
		resp, err := c.V1ProjectsDocumentsCreate(context.Background(), api.V1ProjectsDocumentsCreateRequestObject{
			ProjectId: projectID.String(),
			Body:      nil,
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1ProjectsDocumentsCreate400JSONResponse)
		assert.True(t, ok)
	})

	t.Run("no permission", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ds := mocksvc.NewMockDocumentService(ctrl)
		ds.EXPECT().Create(gomock.Any(), projectID, gomock.Any()).Return(nil, service.ErrNoPermission)

		c, _, _ := newTestDocumentController(t, ctrl, ds)
		resp, err := c.V1ProjectsDocumentsCreate(context.Background(), api.V1ProjectsDocumentsCreateRequestObject{
			ProjectId: projectID.String(),
			Body: &api.V1ProjectsDocumentsCreateJSONRequestBody{
				Title:   "Project Plan",
				Content: optional.Some("# Project Plan"),
			},
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1ProjectsDocumentsCreate403JSONResponse)
		assert.True(t, ok)
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ds := mocksvc.NewMockDocumentService(ctrl)
		ds.EXPECT().Create(gomock.Any(), projectID, gomock.Any()).Return(nil, errors.Join(service.ErrDocumentCreate, repository.ErrNotFound))

		c, _, _ := newTestDocumentController(t, ctrl, ds)
		resp, err := c.V1ProjectsDocumentsCreate(context.Background(), api.V1ProjectsDocumentsCreateRequestObject{
			ProjectId: projectID.String(),
			Body: &api.V1ProjectsDocumentsCreateJSONRequestBody{
				Title:   "Project Plan",
				Content: optional.Some("# Project Plan"),
			},
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1ProjectsDocumentsCreate404JSONResponse)
		assert.True(t, ok)
	})

	t.Run("quota exceeded", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ds := mocksvc.NewMockDocumentService(ctrl)
		ds.EXPECT().Create(gomock.Any(), projectID, gomock.Any()).Return(nil, service.ErrQuotaExceeded)

		c, _, _ := newTestDocumentController(t, ctrl, ds)
		resp, err := c.V1ProjectsDocumentsCreate(context.Background(), api.V1ProjectsDocumentsCreateRequestObject{
			ProjectId: projectID.String(),
			Body: &api.V1ProjectsDocumentsCreateJSONRequestBody{
				Title:   "Project Plan",
				Content: optional.Some("# Project Plan"),
			},
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1ProjectsDocumentsCreate403JSONResponse)
		assert.True(t, ok)
	})
}

func TestDocumentController_V1NamespacesDocumentsCreate(t *testing.T) {
	t.Parallel()

	orgID := model.MustNewID(model.ResourceTypeOrganization)
	namespaceID := model.MustNewID(model.ResourceTypeNamespace)
	doc := newServiceDocument()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ds := mocksvc.NewMockDocumentService(ctrl)
		ds.EXPECT().Create(gomock.Any(), namespaceID, service.CreateDocumentOpts{
			Title:   "Project Plan",
			Content: []byte("# Project Plan"),
		}).Return(doc, nil)

		c, os, ns := newTestDocumentController(t, ctrl, ds)
		stubOrganizationResolve(os, orgID)
		stubNamespaceResolve(ns, orgID, namespaceID)
		resp, err := c.V1NamespacesDocumentsCreate(context.Background(), api.V1NamespacesDocumentsCreateRequestObject{
			OrganizationRef: orgID.String(),
			NamespaceRef:    namespaceID.String(),
			Body: &api.V1NamespacesDocumentsCreateJSONRequestBody{
				Title:   "Project Plan",
				Content: optional.Some("# Project Plan"),
			},
		})
		require.NoError(t, err)
		got, ok := resp.(api.V1NamespacesDocumentsCreate201JSONResponse)
		require.True(t, ok)
		assert.Equal(t, doc.ID.String(), got.Id)
		assert.Equal(t, doc.Title, got.Title)
	})

	t.Run("bad namespace id", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		c, os, ns := newTestDocumentController(t, ctrl, mocksvc.NewMockDocumentService(ctrl))
		stubOrganizationResolve(os, orgID)
		stubNamespaceResolve(ns, orgID, namespaceID)
		resp, err := c.V1NamespacesDocumentsCreate(context.Background(), api.V1NamespacesDocumentsCreateRequestObject{
			OrganizationRef: "AB",
			Body: &api.V1NamespacesDocumentsCreateJSONRequestBody{
				Title:   "Project Plan",
				Content: optional.Some("# Project Plan"),
			},
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1NamespacesDocumentsCreate400JSONResponse)
		assert.True(t, ok)
	})

	t.Run("nil body", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		c, os, ns := newTestDocumentController(t, ctrl, mocksvc.NewMockDocumentService(ctrl))
		stubOrganizationResolve(os, orgID)
		stubNamespaceResolve(ns, orgID, namespaceID)
		resp, err := c.V1NamespacesDocumentsCreate(context.Background(), api.V1NamespacesDocumentsCreateRequestObject{
			OrganizationRef: orgID.String(),
			NamespaceRef:    namespaceID.String(),
			Body:            nil,
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1NamespacesDocumentsCreate400JSONResponse)
		assert.True(t, ok)
	})

	t.Run("no permission", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ds := mocksvc.NewMockDocumentService(ctrl)
		ds.EXPECT().Create(gomock.Any(), namespaceID, gomock.Any()).Return(nil, service.ErrNoPermission)

		c, os, ns := newTestDocumentController(t, ctrl, ds)
		stubOrganizationResolve(os, orgID)
		stubNamespaceResolve(ns, orgID, namespaceID)
		resp, err := c.V1NamespacesDocumentsCreate(context.Background(), api.V1NamespacesDocumentsCreateRequestObject{
			OrganizationRef: orgID.String(),
			NamespaceRef:    namespaceID.String(),
			Body: &api.V1NamespacesDocumentsCreateJSONRequestBody{
				Title:   "Project Plan",
				Content: optional.Some("# Project Plan"),
			},
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1NamespacesDocumentsCreate403JSONResponse)
		assert.True(t, ok)
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ds := mocksvc.NewMockDocumentService(ctrl)
		ds.EXPECT().Create(gomock.Any(), namespaceID, gomock.Any()).Return(nil, errors.Join(service.ErrDocumentCreate, repository.ErrNotFound))

		c, os, ns := newTestDocumentController(t, ctrl, ds)
		stubOrganizationResolve(os, orgID)
		stubNamespaceResolve(ns, orgID, namespaceID)
		resp, err := c.V1NamespacesDocumentsCreate(context.Background(), api.V1NamespacesDocumentsCreateRequestObject{
			OrganizationRef: orgID.String(),
			NamespaceRef:    namespaceID.String(),
			Body: &api.V1NamespacesDocumentsCreateJSONRequestBody{
				Title:   "Project Plan",
				Content: optional.Some("# Project Plan"),
			},
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1NamespacesDocumentsCreate404JSONResponse)
		assert.True(t, ok)
	})

	t.Run("quota exceeded", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ds := mocksvc.NewMockDocumentService(ctrl)
		ds.EXPECT().Create(gomock.Any(), namespaceID, gomock.Any()).Return(nil, service.ErrQuotaExceeded)

		c, os, ns := newTestDocumentController(t, ctrl, ds)
		stubOrganizationResolve(os, orgID)
		stubNamespaceResolve(ns, orgID, namespaceID)
		resp, err := c.V1NamespacesDocumentsCreate(context.Background(), api.V1NamespacesDocumentsCreateRequestObject{
			OrganizationRef: orgID.String(),
			NamespaceRef:    namespaceID.String(),
			Body: &api.V1NamespacesDocumentsCreateJSONRequestBody{
				Title:   "Project Plan",
				Content: optional.Some("# Project Plan"),
			},
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1NamespacesDocumentsCreate403JSONResponse)
		assert.True(t, ok)
	})
}

func TestDocumentController_V1OrganizationsDocumentsGet(t *testing.T) {
	t.Parallel()

	organizationID := model.MustNewID(model.ResourceTypeOrganization)
	doc := newServicePartialDocument()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ds := mocksvc.NewMockDocumentService(ctrl)
		ds.EXPECT().ListLibrary(gomock.Any(), organizationID, service.LibraryListFilter{}, gomock.Any()).Return(service.Page[*service.PartialDocument]{
			Items: []*service.PartialDocument{doc},
		}, nil)

		c, os, _ := newTestDocumentController(t, ctrl, ds)
		stubOrganizationResolve(os, organizationID)
		resp, err := c.V1OrganizationsDocumentsGet(context.Background(), api.V1OrganizationsDocumentsGetRequestObject{
			OrganizationRef: organizationID.String(),
			Params:          api.V1OrganizationsDocumentsGetParams{},
		})
		require.NoError(t, err)
		got, ok := resp.(api.V1OrganizationsDocumentsGet200JSONResponse)
		require.True(t, ok)
		require.Len(t, got.Items, 1)
		assert.Equal(t, doc.ID.String(), got.Items[0].Id)
		assert.Equal(t, doc.Title, got.Items[0].Title)
	})

	t.Run("bad organization id", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		c, os, _ := newTestDocumentController(t, ctrl, mocksvc.NewMockDocumentService(ctrl))
		stubOrganizationResolve(os, organizationID)
		resp, err := c.V1OrganizationsDocumentsGet(context.Background(), api.V1OrganizationsDocumentsGetRequestObject{
			OrganizationRef: "AB",
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1OrganizationsDocumentsGet400JSONResponse)
		assert.True(t, ok)
	})

	t.Run("no permission", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ds := mocksvc.NewMockDocumentService(ctrl)
		ds.EXPECT().ListLibrary(gomock.Any(), organizationID, service.LibraryListFilter{}, gomock.Any()).Return(service.Page[*service.PartialDocument]{}, service.ErrNoPermission)

		c, os, _ := newTestDocumentController(t, ctrl, ds)
		stubOrganizationResolve(os, organizationID)
		resp, err := c.V1OrganizationsDocumentsGet(context.Background(), api.V1OrganizationsDocumentsGetRequestObject{
			OrganizationRef: organizationID.String(),
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1OrganizationsDocumentsGet403JSONResponse)
		assert.True(t, ok)
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ds := mocksvc.NewMockDocumentService(ctrl)
		ds.EXPECT().ListLibrary(gomock.Any(), organizationID, service.LibraryListFilter{}, gomock.Any()).Return(service.Page[*service.PartialDocument]{}, errors.Join(service.ErrDocumentList, repository.ErrNotFound))

		c, os, _ := newTestDocumentController(t, ctrl, ds)
		stubOrganizationResolve(os, organizationID)
		resp, err := c.V1OrganizationsDocumentsGet(context.Background(), api.V1OrganizationsDocumentsGetRequestObject{
			OrganizationRef: organizationID.String(),
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1OrganizationsDocumentsGet404JSONResponse)
		assert.True(t, ok)
	})

	t.Run("invalid page size", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		c, os, _ := newTestDocumentController(t, ctrl, mocksvc.NewMockDocumentService(ctrl))
		stubOrganizationResolve(os, organizationID)
		resp, err := c.V1OrganizationsDocumentsGet(context.Background(), api.V1OrganizationsDocumentsGetRequestObject{
			OrganizationRef: organizationID.String(),
			Params: api.V1OrganizationsDocumentsGetParams{
				PageSize: convert.ToPointer(repository.MaxPageSize + 1),
			},
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1OrganizationsDocumentsGet400JSONResponse)
		assert.True(t, ok)
	})
}

func TestDocumentController_V1OrganizationsDocumentsCreate(t *testing.T) {
	t.Parallel()

	organizationID := model.MustNewID(model.ResourceTypeOrganization)
	doc := newServiceDocument()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ds := mocksvc.NewMockDocumentService(ctrl)
		ds.EXPECT().Create(gomock.Any(), organizationID, service.CreateDocumentOpts{
			Title:   "Project Plan",
			Excerpt: "Overview of the project plan",
			Content: []byte("# Project Plan\n\nGoals and timeline."),
		}).Return(doc, nil)

		c, os, _ := newTestDocumentController(t, ctrl, ds)
		stubOrganizationResolve(os, organizationID)
		resp, err := c.V1OrganizationsDocumentsCreate(context.Background(), api.V1OrganizationsDocumentsCreateRequestObject{
			OrganizationRef: organizationID.String(),
			Body: &api.V1OrganizationsDocumentsCreateJSONRequestBody{
				Title:   "Project Plan",
				Excerpt: optional.Some("Overview of the project plan"),
				Content: optional.Some("# Project Plan\n\nGoals and timeline."),
			},
		})
		require.NoError(t, err)
		got, ok := resp.(api.V1OrganizationsDocumentsCreate201JSONResponse)
		require.True(t, ok)
		assert.Equal(t, doc.ID.String(), got.Id)
		assert.Equal(t, doc.Title, got.Title)
		assert.Equal(t, string(doc.Content), got.Content)
	})

	t.Run("bad organization id", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		c, os, _ := newTestDocumentController(t, ctrl, mocksvc.NewMockDocumentService(ctrl))
		stubOrganizationResolve(os, organizationID)
		resp, err := c.V1OrganizationsDocumentsCreate(context.Background(), api.V1OrganizationsDocumentsCreateRequestObject{
			OrganizationRef: "AB",
			Body: &api.V1OrganizationsDocumentsCreateJSONRequestBody{
				Title:   "Project Plan",
				Content: optional.Some("# Project Plan"),
			},
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1OrganizationsDocumentsCreate400JSONResponse)
		assert.True(t, ok)
	})

	t.Run("nil body", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		c, os, _ := newTestDocumentController(t, ctrl, mocksvc.NewMockDocumentService(ctrl))
		stubOrganizationResolve(os, organizationID)
		resp, err := c.V1OrganizationsDocumentsCreate(context.Background(), api.V1OrganizationsDocumentsCreateRequestObject{
			OrganizationRef: organizationID.String(),
			Body:            nil,
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1OrganizationsDocumentsCreate400JSONResponse)
		assert.True(t, ok)
	})

	t.Run("no permission", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ds := mocksvc.NewMockDocumentService(ctrl)
		ds.EXPECT().Create(gomock.Any(), organizationID, gomock.Any()).Return(nil, service.ErrNoPermission)

		c, os, _ := newTestDocumentController(t, ctrl, ds)
		stubOrganizationResolve(os, organizationID)
		resp, err := c.V1OrganizationsDocumentsCreate(context.Background(), api.V1OrganizationsDocumentsCreateRequestObject{
			OrganizationRef: organizationID.String(),
			Body: &api.V1OrganizationsDocumentsCreateJSONRequestBody{
				Title:   "Project Plan",
				Content: optional.Some("# Project Plan"),
			},
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1OrganizationsDocumentsCreate403JSONResponse)
		assert.True(t, ok)
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ds := mocksvc.NewMockDocumentService(ctrl)
		ds.EXPECT().Create(gomock.Any(), organizationID, gomock.Any()).Return(nil, errors.Join(service.ErrDocumentCreate, repository.ErrNotFound))

		c, os, _ := newTestDocumentController(t, ctrl, ds)
		stubOrganizationResolve(os, organizationID)
		resp, err := c.V1OrganizationsDocumentsCreate(context.Background(), api.V1OrganizationsDocumentsCreateRequestObject{
			OrganizationRef: organizationID.String(),
			Body: &api.V1OrganizationsDocumentsCreateJSONRequestBody{
				Title:   "Project Plan",
				Content: optional.Some("# Project Plan"),
			},
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1OrganizationsDocumentsCreate404JSONResponse)
		assert.True(t, ok)
	})

	t.Run("quota exceeded", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ds := mocksvc.NewMockDocumentService(ctrl)
		ds.EXPECT().Create(gomock.Any(), organizationID, gomock.Any()).Return(nil, service.ErrQuotaExceeded)

		c, os, _ := newTestDocumentController(t, ctrl, ds)
		stubOrganizationResolve(os, organizationID)
		resp, err := c.V1OrganizationsDocumentsCreate(context.Background(), api.V1OrganizationsDocumentsCreateRequestObject{
			OrganizationRef: organizationID.String(),
			Body: &api.V1OrganizationsDocumentsCreateJSONRequestBody{
				Title:   "Project Plan",
				Content: optional.Some("# Project Plan"),
			},
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1OrganizationsDocumentsCreate403JSONResponse)
		assert.True(t, ok)
	})
}

func TestDocumentController_V1IssuesDocumentsGet(t *testing.T) {
	t.Parallel()

	issueID := model.MustNewID(model.ResourceTypeIssue)
	doc := newServicePartialDocument()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ds := mocksvc.NewMockDocumentService(ctrl)
		ds.EXPECT().ListRelated(gomock.Any(), issueID, gomock.Any()).Return(service.Page[*service.PartialDocument]{
			Items: []*service.PartialDocument{doc},
		}, nil)

		c, _, _ := newTestDocumentController(t, ctrl, ds)
		resp, err := c.V1IssuesDocumentsGet(context.Background(), api.V1IssuesDocumentsGetRequestObject{
			Id:     issueID.String(),
			Params: api.V1IssuesDocumentsGetParams{},
		})
		require.NoError(t, err)
		got, ok := resp.(api.V1IssuesDocumentsGet200JSONResponse)
		require.True(t, ok)
		require.Len(t, got.Items, 1)
		assert.Equal(t, doc.ID.String(), got.Items[0].Id)
		assert.Equal(t, doc.Title, got.Items[0].Title)
	})

	t.Run("bad issue id", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		c, _, _ := newTestDocumentController(t, ctrl, mocksvc.NewMockDocumentService(ctrl))
		resp, err := c.V1IssuesDocumentsGet(context.Background(), api.V1IssuesDocumentsGetRequestObject{
			Id: "not-a-xid",
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1IssuesDocumentsGet400JSONResponse)
		assert.True(t, ok)
	})

	t.Run("no permission", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ds := mocksvc.NewMockDocumentService(ctrl)
		ds.EXPECT().ListRelated(gomock.Any(), issueID, gomock.Any()).Return(service.Page[*service.PartialDocument]{}, service.ErrNoPermission)

		c, _, _ := newTestDocumentController(t, ctrl, ds)
		resp, err := c.V1IssuesDocumentsGet(context.Background(), api.V1IssuesDocumentsGetRequestObject{
			Id: issueID.String(),
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1IssuesDocumentsGet403JSONResponse)
		assert.True(t, ok)
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ds := mocksvc.NewMockDocumentService(ctrl)
		ds.EXPECT().ListRelated(gomock.Any(), issueID, gomock.Any()).Return(service.Page[*service.PartialDocument]{}, errors.Join(service.ErrDocumentList, repository.ErrNotFound))

		c, _, _ := newTestDocumentController(t, ctrl, ds)
		resp, err := c.V1IssuesDocumentsGet(context.Background(), api.V1IssuesDocumentsGetRequestObject{
			Id: issueID.String(),
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1IssuesDocumentsGet404JSONResponse)
		assert.True(t, ok)
	})
}

func TestDocumentController_V1IssuesDocumentsCreate(t *testing.T) {
	t.Parallel()

	issueID := model.MustNewID(model.ResourceTypeIssue)
	doc := newServiceDocument()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ds := mocksvc.NewMockDocumentService(ctrl)
		ds.EXPECT().Create(gomock.Any(), issueID, service.CreateDocumentOpts{
			Title:   "Project Plan",
			Content: []byte("# Project Plan"),
		}).Return(doc, nil)

		c, _, _ := newTestDocumentController(t, ctrl, ds)
		resp, err := c.V1IssuesDocumentsCreate(context.Background(), api.V1IssuesDocumentsCreateRequestObject{
			Id: issueID.String(),
			Body: &api.V1IssuesDocumentsCreateJSONRequestBody{
				Title:   "Project Plan",
				Content: optional.Some("# Project Plan"),
			},
		})
		require.NoError(t, err)
		got, ok := resp.(api.V1IssuesDocumentsCreate201JSONResponse)
		require.True(t, ok)
		assert.Equal(t, doc.ID.String(), got.Id)
		assert.Equal(t, doc.Title, got.Title)
	})

	t.Run("bad issue id", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		c, _, _ := newTestDocumentController(t, ctrl, mocksvc.NewMockDocumentService(ctrl))
		resp, err := c.V1IssuesDocumentsCreate(context.Background(), api.V1IssuesDocumentsCreateRequestObject{
			Id: "not-a-xid",
			Body: &api.V1IssuesDocumentsCreateJSONRequestBody{
				Title:   "Project Plan",
				Content: optional.Some("# Project Plan"),
			},
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1IssuesDocumentsCreate400JSONResponse)
		assert.True(t, ok)
	})

	t.Run("nil body", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		c, _, _ := newTestDocumentController(t, ctrl, mocksvc.NewMockDocumentService(ctrl))
		resp, err := c.V1IssuesDocumentsCreate(context.Background(), api.V1IssuesDocumentsCreateRequestObject{
			Id:   issueID.String(),
			Body: nil,
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1IssuesDocumentsCreate400JSONResponse)
		assert.True(t, ok)
	})

	t.Run("no permission", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ds := mocksvc.NewMockDocumentService(ctrl)
		ds.EXPECT().Create(gomock.Any(), issueID, gomock.Any()).Return(nil, service.ErrNoPermission)

		c, _, _ := newTestDocumentController(t, ctrl, ds)
		resp, err := c.V1IssuesDocumentsCreate(context.Background(), api.V1IssuesDocumentsCreateRequestObject{
			Id: issueID.String(),
			Body: &api.V1IssuesDocumentsCreateJSONRequestBody{
				Title:   "Project Plan",
				Content: optional.Some("# Project Plan"),
			},
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1IssuesDocumentsCreate403JSONResponse)
		assert.True(t, ok)
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ds := mocksvc.NewMockDocumentService(ctrl)
		ds.EXPECT().Create(gomock.Any(), issueID, gomock.Any()).Return(nil, errors.Join(service.ErrDocumentCreate, repository.ErrNotFound))

		c, _, _ := newTestDocumentController(t, ctrl, ds)
		resp, err := c.V1IssuesDocumentsCreate(context.Background(), api.V1IssuesDocumentsCreateRequestObject{
			Id: issueID.String(),
			Body: &api.V1IssuesDocumentsCreateJSONRequestBody{
				Title:   "Project Plan",
				Content: optional.Some("# Project Plan"),
			},
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1IssuesDocumentsCreate404JSONResponse)
		assert.True(t, ok)
	})

	t.Run("quota exceeded", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ds := mocksvc.NewMockDocumentService(ctrl)
		ds.EXPECT().Create(gomock.Any(), issueID, gomock.Any()).Return(nil, service.ErrQuotaExceeded)

		c, _, _ := newTestDocumentController(t, ctrl, ds)
		resp, err := c.V1IssuesDocumentsCreate(context.Background(), api.V1IssuesDocumentsCreateRequestObject{
			Id: issueID.String(),
			Body: &api.V1IssuesDocumentsCreateJSONRequestBody{
				Title:   "Project Plan",
				Content: optional.Some("# Project Plan"),
			},
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1IssuesDocumentsCreate403JSONResponse)
		assert.True(t, ok)
	})
}

func TestDocumentController_V1DocumentGet(t *testing.T) {
	t.Parallel()

	doc := newServiceDocument()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ds := mocksvc.NewMockDocumentService(ctrl)
		ds.EXPECT().Get(gomock.Any(), doc.ID).Return(doc, nil)

		c, _, _ := newTestDocumentController(t, ctrl, ds)
		resp, err := c.V1DocumentGet(context.Background(), api.V1DocumentGetRequestObject{
			Id: doc.ID.String(),
		})
		require.NoError(t, err)
		got, ok := resp.(api.V1DocumentGet200JSONResponse)
		require.True(t, ok)
		assert.Equal(t, doc.ID.String(), got.Id)
		assert.Equal(t, doc.Title, got.Title)
		assert.Equal(t, string(doc.Content), got.Content)
		require.NotNil(t, got.Excerpt)
		assert.Equal(t, doc.Excerpt, *got.Excerpt)
	})

	t.Run("bad document id", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		c, _, _ := newTestDocumentController(t, ctrl, mocksvc.NewMockDocumentService(ctrl))
		resp, err := c.V1DocumentGet(context.Background(), api.V1DocumentGetRequestObject{
			Id: "not-a-xid",
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1DocumentGet400JSONResponse)
		assert.True(t, ok)
	})

	t.Run("no permission", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ds := mocksvc.NewMockDocumentService(ctrl)
		ds.EXPECT().Get(gomock.Any(), doc.ID).Return(nil, service.ErrNoPermission)

		c, _, _ := newTestDocumentController(t, ctrl, ds)
		resp, err := c.V1DocumentGet(context.Background(), api.V1DocumentGetRequestObject{
			Id: doc.ID.String(),
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1DocumentGet403JSONResponse)
		assert.True(t, ok)
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ds := mocksvc.NewMockDocumentService(ctrl)
		ds.EXPECT().Get(gomock.Any(), doc.ID).Return(nil, errors.Join(service.ErrDocumentGet, repository.ErrNotFound))

		c, _, _ := newTestDocumentController(t, ctrl, ds)
		resp, err := c.V1DocumentGet(context.Background(), api.V1DocumentGetRequestObject{
			Id: doc.ID.String(),
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1DocumentGet404JSONResponse)
		assert.True(t, ok)
	})
}

func TestDocumentController_V1DocumentUpdate(t *testing.T) {
	t.Parallel()

	doc := newServiceDocument()
	updated := *doc
	updated.Title = "Updated Plan"
	updated.Content = []byte("# Updated")

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ds := mocksvc.NewMockDocumentService(ctrl)
		ds.EXPECT().Update(gomock.Any(), doc.ID, service.UpdateDocumentOpts{
			Title:   optional.Some("Updated Plan"),
			Content: optional.Some([]byte("# Updated")),
		}).Return(&updated, nil)

		c, _, _ := newTestDocumentController(t, ctrl, ds)
		resp, err := c.V1DocumentUpdate(context.Background(), api.V1DocumentUpdateRequestObject{
			Id: doc.ID.String(),
			Body: &api.V1DocumentUpdateJSONRequestBody{
				Title:   optional.Some("Updated Plan"),
				Content: optional.Some("# Updated"),
			},
		})
		require.NoError(t, err)
		got, ok := resp.(api.V1DocumentUpdate200JSONResponse)
		require.True(t, ok)
		assert.Equal(t, updated.Title, got.Title)
		assert.Equal(t, string(updated.Content), got.Content)
	})

	t.Run("clears folder", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		cleared := *doc
		cleared.Folder = nil

		ds := mocksvc.NewMockDocumentService(ctrl)
		ds.EXPECT().Update(gomock.Any(), doc.ID, service.UpdateDocumentOpts{
			FolderID: optional.Null[model.ID](),
		}).Return(&cleared, nil)

		c, _, _ := newTestDocumentController(t, ctrl, ds)
		resp, err := c.V1DocumentUpdate(context.Background(), api.V1DocumentUpdateRequestObject{
			Id: doc.ID.String(),
			Body: &api.V1DocumentUpdateJSONRequestBody{
				FolderId: optional.Null[string](),
			},
		})
		require.NoError(t, err)
		got, ok := resp.(api.V1DocumentUpdate200JSONResponse)
		require.True(t, ok)
		assert.Nil(t, got.Folder)
	})

	t.Run("moves to folder", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		folderID := model.MustNewID(model.ResourceTypeFolder)
		moved := *doc
		moved.Folder = &service.DocumentFolder{ID: folderID, Name: "Guides"}

		ds := mocksvc.NewMockDocumentService(ctrl)
		ds.EXPECT().Update(gomock.Any(), doc.ID, gomock.Cond(func(opts service.UpdateDocumentOpts) bool {
			return opts.FolderID.Defined &&
				opts.FolderID.Value != nil &&
				opts.FolderID.Value.String() == folderID.String() &&
				opts.FolderID.Value.Type == model.ResourceTypeFolder
		})).Return(&moved, nil)

		c, _, _ := newTestDocumentController(t, ctrl, ds)
		resp, err := c.V1DocumentUpdate(context.Background(), api.V1DocumentUpdateRequestObject{
			Id: doc.ID.String(),
			Body: &api.V1DocumentUpdateJSONRequestBody{
				FolderId: optional.Some(folderID.String()),
			},
		})
		require.NoError(t, err)
		got, ok := resp.(api.V1DocumentUpdate200JSONResponse)
		require.True(t, ok)
		require.NotNil(t, got.Folder)
		assert.Equal(t, folderID.String(), got.Folder.Id)
	})

	t.Run("decodes folder_id from json as a folder id", func(t *testing.T) {
		t.Parallel()
		folderID := model.MustNewID(model.ResourceTypeFolder)
		var body api.V1DocumentUpdateJSONRequestBody
		err := json.Unmarshal([]byte(`{"folder_id":"`+folderID.String()+`"}`), &body)
		require.NoError(t, err)

		opts, err := updateDocumentJSONRequestBodyToUpdateDocumentOpts(&body)
		require.NoError(t, err)
		require.True(t, opts.FolderID.Defined)
		require.NotNil(t, opts.FolderID.Value)
		assert.Equal(t, folderID.String(), opts.FolderID.Value.String())
		assert.Equal(t, model.ResourceTypeFolder, opts.FolderID.Value.Type)
	})

	t.Run("bad document id", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		c, _, _ := newTestDocumentController(t, ctrl, mocksvc.NewMockDocumentService(ctrl))
		resp, err := c.V1DocumentUpdate(context.Background(), api.V1DocumentUpdateRequestObject{
			Id: "not-a-xid",
			Body: &api.V1DocumentUpdateJSONRequestBody{
				Title: optional.Some("Updated Plan"),
			},
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1DocumentUpdate400JSONResponse)
		assert.True(t, ok)
	})

	t.Run("nil body", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		c, _, _ := newTestDocumentController(t, ctrl, mocksvc.NewMockDocumentService(ctrl))
		resp, err := c.V1DocumentUpdate(context.Background(), api.V1DocumentUpdateRequestObject{
			Id:   doc.ID.String(),
			Body: nil,
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1DocumentUpdate400JSONResponse)
		assert.True(t, ok)
	})

	t.Run("no permission", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ds := mocksvc.NewMockDocumentService(ctrl)
		ds.EXPECT().Update(gomock.Any(), doc.ID, gomock.Any()).Return(nil, service.ErrNoPermission)

		c, _, _ := newTestDocumentController(t, ctrl, ds)
		resp, err := c.V1DocumentUpdate(context.Background(), api.V1DocumentUpdateRequestObject{
			Id: doc.ID.String(),
			Body: &api.V1DocumentUpdateJSONRequestBody{
				Title: optional.Some("Updated Plan"),
			},
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1DocumentUpdate403JSONResponse)
		assert.True(t, ok)
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ds := mocksvc.NewMockDocumentService(ctrl)
		ds.EXPECT().Update(gomock.Any(), doc.ID, gomock.Any()).Return(nil, errors.Join(service.ErrDocumentUpdate, repository.ErrNotFound))

		c, _, _ := newTestDocumentController(t, ctrl, ds)
		resp, err := c.V1DocumentUpdate(context.Background(), api.V1DocumentUpdateRequestObject{
			Id: doc.ID.String(),
			Body: &api.V1DocumentUpdateJSONRequestBody{
				Title: optional.Some("Updated Plan"),
			},
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1DocumentUpdate404JSONResponse)
		assert.True(t, ok)
	})
}

func TestDocumentController_V1DocumentDelete(t *testing.T) {
	t.Parallel()

	docID := model.MustNewID(model.ResourceTypeDocument)

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ds := mocksvc.NewMockDocumentService(ctrl)
		ds.EXPECT().Delete(gomock.Any(), docID).Return(nil)

		c, _, _ := newTestDocumentController(t, ctrl, ds)
		resp, err := c.V1DocumentDelete(context.Background(), api.V1DocumentDeleteRequestObject{
			Id: docID.String(),
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1DocumentDelete204Response)
		assert.True(t, ok)
	})

	t.Run("bad document id", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		c, _, _ := newTestDocumentController(t, ctrl, mocksvc.NewMockDocumentService(ctrl))
		resp, err := c.V1DocumentDelete(context.Background(), api.V1DocumentDeleteRequestObject{
			Id: "not-a-xid",
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1DocumentDelete400JSONResponse)
		assert.True(t, ok)
	})

	t.Run("no permission", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ds := mocksvc.NewMockDocumentService(ctrl)
		ds.EXPECT().Delete(gomock.Any(), docID).Return(service.ErrNoPermission)

		c, _, _ := newTestDocumentController(t, ctrl, ds)
		resp, err := c.V1DocumentDelete(context.Background(), api.V1DocumentDeleteRequestObject{
			Id: docID.String(),
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1DocumentDelete403JSONResponse)
		assert.True(t, ok)
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ds := mocksvc.NewMockDocumentService(ctrl)
		ds.EXPECT().Delete(gomock.Any(), docID).Return(errors.Join(service.ErrDocumentDelete, repository.ErrNotFound))

		c, _, _ := newTestDocumentController(t, ctrl, ds)
		resp, err := c.V1DocumentDelete(context.Background(), api.V1DocumentDeleteRequestObject{
			Id: docID.String(),
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1DocumentDelete404JSONResponse)
		assert.True(t, ok)
	})
}

func TestDocumentController_V1ProjectsDocumentsRelate(t *testing.T) {
	t.Parallel()

	projectID := model.MustNewID(model.ResourceTypeProject)
	documentID := model.MustNewID(model.ResourceTypeDocument)

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ds := mocksvc.NewMockDocumentService(ctrl)
		ds.EXPECT().Relate(gomock.Any(), documentID, projectID).Return(nil)

		c, _, _ := newTestDocumentController(t, ctrl, ds)
		resp, err := c.V1ProjectsDocumentsRelate(context.Background(), api.V1ProjectsDocumentsRelateRequestObject{
			ProjectId:  projectID.String(),
			DocumentId: documentID.String(),
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1ProjectsDocumentsRelate204Response)
		assert.True(t, ok)
	})
}

func TestDocumentController_V1ProjectsDocumentsUnrelate(t *testing.T) {
	t.Parallel()

	projectID := model.MustNewID(model.ResourceTypeProject)
	documentID := model.MustNewID(model.ResourceTypeDocument)

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ds := mocksvc.NewMockDocumentService(ctrl)
		ds.EXPECT().Unrelate(gomock.Any(), documentID, projectID).Return(nil)

		c, _, _ := newTestDocumentController(t, ctrl, ds)
		resp, err := c.V1ProjectsDocumentsUnrelate(context.Background(), api.V1ProjectsDocumentsUnrelateRequestObject{
			ProjectId:  projectID.String(),
			DocumentId: documentID.String(),
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1ProjectsDocumentsUnrelate204Response)
		assert.True(t, ok)
	})
}

func TestDocumentToDTO(t *testing.T) {
	t.Parallel()

	createdAt := convert.ToPointer(time.Now().UTC())
	updatedAt := convert.ToPointer(time.Now().UTC())
	docID := model.MustNewID(model.ResourceTypeDocument)
	userID := model.MustNewID(model.ResourceTypeUser)
	labelID := model.MustNewID(model.ResourceTypeLabel)

	t.Run("with excerpt labels and counts", func(t *testing.T) {
		t.Parallel()
		doc := &service.Document{
			ID:      docID,
			Title:   "Project Plan",
			Excerpt: "Overview of the project plan",
			FileID:  "documents/file",
			CreatedBy: service.PartialUser{
				ID: userID,
			},
			Labels: []service.PartialLabel{
				{ID: labelID, Name: "frontend"},
			},
			CommentCount:    convert.ToPointer(int64(3)),
			AttachmentCount: convert.ToPointer(int64(1)),
			CreatedAt:       createdAt,
			UpdatedAt:       updatedAt,
			Content:         []byte("# Project Plan"),
		}

		dto := documentToDTO(doc)
		assert.Equal(t, docID.String(), dto.Id)
		assert.Equal(t, "Project Plan", dto.Title)
		require.NotNil(t, dto.Excerpt)
		assert.Equal(t, "Overview of the project plan", *dto.Excerpt)
		assert.Equal(t, "# Project Plan", dto.Content)
		assert.Equal(t, userID.String(), dto.CreatedBy.Id)
		assert.Equal(t, []api.PartialLabel{{Id: labelID.String(), Name: "frontend"}}, dto.Labels)
		require.NotNil(t, dto.CommentCount)
		assert.Equal(t, int64(3), *dto.CommentCount)
		require.NotNil(t, dto.AttachmentCount)
		assert.Equal(t, int64(1), *dto.AttachmentCount)
		assert.Equal(t, *createdAt, dto.CreatedAt)
		assert.Equal(t, updatedAt, dto.UpdatedAt)
	})

	t.Run("without excerpt labels or counts", func(t *testing.T) {
		t.Parallel()
		doc := &service.Document{
			ID:        docID,
			Title:     "Project Plan",
			CreatedBy: service.PartialUser{ID: userID},
			Content:   []byte("# Project Plan"),
		}

		dto := documentToDTO(doc)
		assert.Equal(t, docID.String(), dto.Id)
		assert.Nil(t, dto.Excerpt)
		assert.Equal(t, "# Project Plan", dto.Content)
		assert.Empty(t, dto.Labels)
		assert.NotNil(t, dto.Labels)
		assert.Nil(t, dto.CommentCount)
		assert.Nil(t, dto.AttachmentCount)
		assert.True(t, dto.CreatedAt.IsZero())
		assert.Nil(t, dto.UpdatedAt)
	})
}

func TestPartialDocumentToDTO(t *testing.T) {
	t.Parallel()

	createdAt := convert.ToPointer(time.Now().UTC())
	docID := model.MustNewID(model.ResourceTypeDocument)
	userID := model.MustNewID(model.ResourceTypeUser)

	t.Run("with excerpt", func(t *testing.T) {
		t.Parallel()
		doc := &service.PartialDocument{
			ID:        docID,
			Title:     "Plan",
			Excerpt:   "Overview",
			CreatedBy: service.PartialUser{ID: userID},
			CreatedAt: createdAt,
			UpdatedAt: createdAt,
		}

		dto := partialDocumentToDTO(doc)
		assert.Equal(t, docID.String(), dto.Id)
		assert.Equal(t, "Plan", dto.Title)
		require.NotNil(t, dto.Excerpt)
		assert.Equal(t, "Overview", *dto.Excerpt)
		assert.Equal(t, userID.String(), dto.CreatedBy.Id)
		assert.Equal(t, createdAt, dto.CreatedAt)
		assert.Equal(t, createdAt, dto.UpdatedAt)
	})

	t.Run("without excerpt", func(t *testing.T) {
		t.Parallel()
		doc := &service.PartialDocument{
			ID:        docID,
			Title:     "Plan",
			CreatedBy: service.PartialUser{ID: userID},
			CreatedAt: nil,
		}

		dto := partialDocumentToDTO(doc)
		assert.Equal(t, docID.String(), dto.Id)
		assert.Nil(t, dto.Excerpt)
		assert.Nil(t, dto.CreatedAt)
		assert.Nil(t, dto.UpdatedAt)
	})
}

package http

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg/convert"
	"github.com/opcotech/elemo/internal/pkg/optional"
	"github.com/opcotech/elemo/internal/service"
	"github.com/opcotech/elemo/internal/transport/http/api"
)

func newTestFolderController(t *testing.T, fs service.FolderService) FolderController {
	t.Helper()
	c, err := NewFolderController(WithFolderService(fs))
	require.NoError(t, err)
	return c
}

func newServiceFolder() *service.Folder {
	libraryID := model.MustNewID(model.ResourceTypeNamespace)
	return &service.Folder{
		ID:   model.MustNewID(model.ResourceTypeFolder),
		Name: "Guides",
		Library: service.DocumentLibrary{
			ID:   libraryID,
			Type: model.ResourceTypeNamespace,
			Name: "Product",
		},
		CreatedBy: service.PartialUser{ID: model.MustNewID(model.ResourceTypeUser)},
		CreatedAt: convert.ToPointer(time.Now().UTC()),
	}
}

func TestNewFolderController(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		c, err := NewFolderController(WithFolderService(service.NewMockFolderService(ctrl)))
		require.NoError(t, err)
		assert.NotNil(t, c)
	})

	t.Run("missing folder service", func(t *testing.T) {
		t.Parallel()
		_, err := NewFolderController()
		assert.ErrorIs(t, err, ErrNoFolderService)
	})
}

func TestFolderController_V1NamespacesFoldersCreate(t *testing.T) {
	t.Parallel()

	namespaceID := model.MustNewID(model.ResourceTypeNamespace)
	folder := newServiceFolder()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		fs := service.NewMockFolderService(ctrl)
		fs.EXPECT().Create(gomock.Any(), namespaceID, service.CreateFolderOpts{Name: "Guides"}).Return(folder, nil)

		c := newTestFolderController(t, fs)
		resp, err := c.V1NamespacesFoldersCreate(context.Background(), api.V1NamespacesFoldersCreateRequestObject{
			Id: namespaceID.String(),
			Body: &api.V1NamespacesFoldersCreateJSONRequestBody{
				Name: "Guides",
			},
		})
		require.NoError(t, err)
		got, ok := resp.(api.V1NamespacesFoldersCreate201JSONResponse)
		require.True(t, ok)
		assert.Equal(t, folder.ID.String(), got.Id)
		assert.Equal(t, "Guides", got.Name)
	})

	t.Run("bad namespace id", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		c := newTestFolderController(t, service.NewMockFolderService(ctrl))
		resp, err := c.V1NamespacesFoldersCreate(context.Background(), api.V1NamespacesFoldersCreateRequestObject{
			Id: "not-a-xid",
			Body: &api.V1NamespacesFoldersCreateJSONRequestBody{
				Name: "Guides",
			},
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1NamespacesFoldersCreate400JSONResponse)
		assert.True(t, ok)
	})
}

func TestFolderController_V1NamespacesFoldersGet(t *testing.T) {
	t.Parallel()

	namespaceID := model.MustNewID(model.ResourceTypeNamespace)
	folder := newServiceFolder()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		fs := service.NewMockFolderService(ctrl)
		fs.EXPECT().List(gomock.Any(), namespaceID, (*model.ID)(nil), gomock.Any()).Return(service.Page[*service.Folder]{
			Items: []*service.Folder{folder},
		}, nil)

		c := newTestFolderController(t, fs)
		resp, err := c.V1NamespacesFoldersGet(context.Background(), api.V1NamespacesFoldersGetRequestObject{
			Id:     namespaceID.String(),
			Params: api.V1NamespacesFoldersGetParams{},
		})
		require.NoError(t, err)
		got, ok := resp.(api.V1NamespacesFoldersGet200JSONResponse)
		require.True(t, ok)
		require.Len(t, got.Items, 1)
		assert.Equal(t, folder.ID.String(), got.Items[0].Id)
	})
}

func TestFolderController_V1FolderGet(t *testing.T) {
	t.Parallel()

	folder := newServiceFolder()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		fs := service.NewMockFolderService(ctrl)
		fs.EXPECT().Get(gomock.Any(), folder.ID).Return(folder, nil)

		c := newTestFolderController(t, fs)
		resp, err := c.V1FolderGet(context.Background(), api.V1FolderGetRequestObject{
			Id: folder.ID.String(),
		})
		require.NoError(t, err)
		got, ok := resp.(api.V1FolderGet200JSONResponse)
		require.True(t, ok)
		assert.Equal(t, folder.ID.String(), got.Id)
	})
}

func TestFolderController_V1FolderDelete(t *testing.T) {
	t.Parallel()

	folderID := model.MustNewID(model.ResourceTypeFolder)

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		fs := service.NewMockFolderService(ctrl)
		fs.EXPECT().Delete(gomock.Any(), folderID).Return(nil)

		c := newTestFolderController(t, fs)
		resp, err := c.V1FolderDelete(context.Background(), api.V1FolderDeleteRequestObject{
			Id: folderID.String(),
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1FolderDelete204Response)
		assert.True(t, ok)
	})
}

func TestFolderController_V1FolderUpdate(t *testing.T) {
	t.Parallel()

	folder := newServiceFolder()
	updated := *folder
	updated.Name = "Architecture"

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		fs := service.NewMockFolderService(ctrl)
		fs.EXPECT().Update(gomock.Any(), folder.ID, service.UpdateFolderOpts{
			Name: optional.Some("Architecture"),
		}).Return(&updated, nil)

		c := newTestFolderController(t, fs)
		resp, err := c.V1FolderUpdate(context.Background(), api.V1FolderUpdateRequestObject{
			Id: folder.ID.String(),
			Body: &api.V1FolderUpdateJSONRequestBody{
				Name: optional.Some("Architecture"),
			},
		})
		require.NoError(t, err)
		got, ok := resp.(api.V1FolderUpdate200JSONResponse)
		require.True(t, ok)
		assert.Equal(t, "Architecture", got.Name)
	})

	t.Run("clears parent", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		cleared := *folder
		cleared.Parent = nil

		fs := service.NewMockFolderService(ctrl)
		fs.EXPECT().Update(gomock.Any(), folder.ID, service.UpdateFolderOpts{
			ParentID: optional.Null[model.ID](),
		}).Return(&cleared, nil)

		c := newTestFolderController(t, fs)
		resp, err := c.V1FolderUpdate(context.Background(), api.V1FolderUpdateRequestObject{
			Id: folder.ID.String(),
			Body: &api.V1FolderUpdateJSONRequestBody{
				ParentId: optional.Null[string](),
			},
		})
		require.NoError(t, err)
		got, ok := resp.(api.V1FolderUpdate200JSONResponse)
		require.True(t, ok)
		assert.Nil(t, got.Parent)
	})
}

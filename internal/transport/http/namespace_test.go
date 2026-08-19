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
	"github.com/opcotech/elemo/internal/service"
	"github.com/opcotech/elemo/internal/transport/http/api"
)

func TestNewNamespaceController(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		c, err := NewNamespaceController(WithNamespaceService(service.NewMockNamespaceService(ctrl)))
		require.NoError(t, err)
		assert.NotNil(t, c)
	})

	t.Run("missing namespace service", func(t *testing.T) {
		t.Parallel()
		_, err := NewNamespaceController()
		assert.ErrorIs(t, err, ErrNoNamespaceService)
	})
}

func TestNamespaceController_V1NamespaceGet(t *testing.T) {
	t.Parallel()

	namespaceID := model.MustNewID(model.ResourceTypeNamespace)
	ns := &service.Namespace{
		ID:            namespaceID,
		Name:          "Engineering",
		Description:   "Engineering team namespace",
		ProjectCount:  convert.ToPointer(int64(1)),
		DocumentCount: convert.ToPointer(int64(1)),
		CreatedAt:     convert.ToPointer(time.Now().UTC()),
	}

	t.Run("success with related resources", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		nsSvc := service.NewMockNamespaceService(ctrl)
		nsSvc.EXPECT().Get(gomock.Any(), namespaceID).Return(ns, nil)

		c, err := NewNamespaceController(WithNamespaceService(nsSvc))
		require.NoError(t, err)

		resp, err := c.V1NamespaceGet(context.Background(), api.V1NamespaceGetRequestObject{Id: namespaceID.String()})
		require.NoError(t, err)
		got, ok := resp.(api.V1NamespaceGet200JSONResponse)
		require.True(t, ok)
		assert.Equal(t, namespaceID.String(), got.Id)
		require.NotNil(t, got.Description)
		assert.Equal(t, ns.Description, *got.Description)
		require.NotNil(t, got.ProjectCount)
		assert.Equal(t, int64(1), *got.ProjectCount)
		require.NotNil(t, got.DocumentCount)
		assert.Equal(t, int64(1), *got.DocumentCount)
	})
}

func TestNamespaceController_V1NamespacesGet(t *testing.T) {
	t.Parallel()

	orgID := model.MustNewID(model.ResourceTypeOrganization)
	namespaceID := model.MustNewID(model.ResourceTypeNamespace)
	ns := &service.AccessibleNamespace{
		Namespace: service.Namespace{
			ID:            namespaceID,
			Name:          "Engineering",
			Description:   "Engineering team namespace",
			ProjectCount:  convert.ToPointer(int64(1)),
			DocumentCount: convert.ToPointer(int64(1)),
			CreatedAt:     convert.ToPointer(time.Now().UTC()),
		},
		Organization: service.PartialOrganization{
			ID:   orgID,
			Name: "ACME",
		},
	}

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		nsSvc := service.NewMockNamespaceService(ctrl)
		nsSvc.EXPECT().ListAccessible(gomock.Any(), gomock.Any()).Return(service.Page[*service.AccessibleNamespace]{
			Items: []*service.AccessibleNamespace{ns},
		}, nil)

		c, err := NewNamespaceController(WithNamespaceService(nsSvc))
		require.NoError(t, err)

		resp, err := c.V1NamespacesGet(context.Background(), api.V1NamespacesGetRequestObject{})
		require.NoError(t, err)
		got, ok := resp.(api.V1NamespacesGet200JSONResponse)
		require.True(t, ok)
		require.Len(t, got.Items, 1)
		assert.Equal(t, namespaceID.String(), got.Items[0].Id)
		assert.Equal(t, orgID.String(), got.Items[0].Organization.Id)
		assert.Equal(t, "ACME", got.Items[0].Organization.Name)
	})
}

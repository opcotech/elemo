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
		ID:          namespaceID,
		Name:        "Engineering",
		Description: "Engineering team namespace",
		Projects: []*service.PartialProject{
			{
				ID:     model.MustNewID(model.ResourceTypeProject),
				Key:    "ENG",
				Name:   "Engineering Project",
				Status: model.ProjectStatusActive,
			},
		},
		Documents: []*service.PartialDocument{
			{
				ID:        model.MustNewID(model.ResourceTypeDocument),
				Name:      "Plan",
				Excerpt:   "Overview",
				CreatedBy: model.MustNewID(model.ResourceTypeUser),
				CreatedAt: convert.ToPointer(time.Now().UTC()),
			},
		},
		CreatedAt: convert.ToPointer(time.Now().UTC()),
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
		require.Len(t, got.Projects, 1)
		assert.Equal(t, "ENG", got.Projects[0].Key)
		require.Len(t, got.Documents, 1)
		assert.Equal(t, "Plan", got.Documents[0].Name)
	})
}

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
	"github.com/opcotech/elemo/internal/repository"
	"github.com/opcotech/elemo/internal/service"
	mocksvc "github.com/opcotech/elemo/internal/service/mock"
	"github.com/opcotech/elemo/internal/transport/http/api"
)

func TestNewNamespaceController(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		c, err := NewNamespaceController(mocksvc.NewMockOrganizationService(ctrl), mocksvc.NewMockNamespaceService(ctrl))
		require.NoError(t, err)
		assert.NotNil(t, c)
	})

	t.Run("missing namespace service", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		_, err := NewNamespaceController(mocksvc.NewMockOrganizationService(ctrl), nil)
		assert.ErrorIs(t, err, ErrNoNamespaceService)
	})
}

func TestNamespaceController_V1NamespaceGet(t *testing.T) {
	t.Parallel()

	orgID := model.MustNewID(model.ResourceTypeOrganization)
	namespaceID := model.MustNewID(model.ResourceTypeNamespace)
	ns := &service.AccessibleNamespace{
		Namespace: service.Namespace{
			ID:            namespaceID,
			Slug:          "engineering",
			Name:          "Engineering",
			Description:   "Engineering team namespace",
			ProjectCount:  convert.ToPointer(int64(1)),
			DocumentCount: convert.ToPointer(int64(1)),
			CreatedAt:     convert.ToPointer(time.Now().UTC()),
		},
		Organization: service.PartialOrganization{
			ID:   orgID,
			Slug: "acme",
			Name: "ACME",
		},
	}

	t.Run("success with related resources", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		nsSvc := mocksvc.NewMockNamespaceService(ctrl)
		nsSvc.EXPECT().GetByRef(gomock.Any(), orgID, namespaceID, "").Return(ns, nil)

		c, os := newTestNamespaceController(t, ctrl, nsSvc)
		stubOrganizationResolve(os, orgID)

		resp, err := c.V1NamespaceGet(context.Background(), api.V1NamespaceGetRequestObject{
			OrganizationRef: orgID.String(),
			NamespaceRef:    namespaceID.String(),
		})
		require.NoError(t, err)
		got, ok := resp.(api.V1NamespaceGet200JSONResponse)
		require.True(t, ok)
		assert.Equal(t, namespaceID.String(), got.Id)
		assert.Equal(t, "engineering", got.Slug)
		require.NotNil(t, got.Description)
		assert.Equal(t, ns.Description, *got.Description)
		require.NotNil(t, got.ProjectCount)
		assert.Equal(t, int64(1), *got.ProjectCount)
		require.NotNil(t, got.DocumentCount)
		assert.Equal(t, int64(1), *got.DocumentCount)
		assert.Equal(t, orgID.String(), got.Organization.Id)
		assert.Equal(t, "acme", got.Organization.Slug)
	})

	t.Run("success by slug", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		nsSvc := mocksvc.NewMockNamespaceService(ctrl)
		nsSvc.EXPECT().GetByRef(gomock.Any(), orgID, model.ID{}, "engineering").Return(ns, nil)

		c, os := newTestNamespaceController(t, ctrl, nsSvc)
		os.EXPECT().Resolve(gomock.Any(), model.ID{}, "acme").Return(&service.Organization{ID: orgID, Slug: "acme"}, nil)

		resp, err := c.V1NamespaceGet(context.Background(), api.V1NamespaceGetRequestObject{
			OrganizationRef: "acme",
			NamespaceRef:    "engineering",
		})
		require.NoError(t, err)
		got, ok := resp.(api.V1NamespaceGet200JSONResponse)
		require.True(t, ok)
		assert.Equal(t, namespaceID.String(), got.Id)
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		nsSvc := mocksvc.NewMockNamespaceService(ctrl)
		nsSvc.EXPECT().GetByRef(gomock.Any(), orgID, namespaceID, "").Return(nil, repository.ErrNotFound)

		c, os := newTestNamespaceController(t, ctrl, nsSvc)
		stubOrganizationResolve(os, orgID)

		resp, err := c.V1NamespaceGet(context.Background(), api.V1NamespaceGetRequestObject{
			OrganizationRef: orgID.String(),
			NamespaceRef:    namespaceID.String(),
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1NamespaceGet404JSONResponse)
		assert.True(t, ok)
	})

	t.Run("mismatched organization", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		otherOrgID := model.MustNewID(model.ResourceTypeOrganization)
		nsSvc := mocksvc.NewMockNamespaceService(ctrl)
		nsSvc.EXPECT().GetByRef(gomock.Any(), otherOrgID, namespaceID, "").Return(nil, repository.ErrNotFound)

		c, os := newTestNamespaceController(t, ctrl, nsSvc)
		stubOrganizationResolve(os, otherOrgID)

		resp, err := c.V1NamespaceGet(context.Background(), api.V1NamespaceGetRequestObject{
			OrganizationRef: otherOrgID.String(),
			NamespaceRef:    namespaceID.String(),
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1NamespaceGet404JSONResponse)
		assert.True(t, ok)
	})

	t.Run("invalid namespace ref", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		c, os := newTestNamespaceController(t, ctrl, mocksvc.NewMockNamespaceService(ctrl))
		stubOrganizationResolve(os, orgID)

		resp, err := c.V1NamespaceGet(context.Background(), api.V1NamespaceGetRequestObject{
			OrganizationRef: orgID.String(),
			NamespaceRef:    "AB",
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1NamespaceGet400JSONResponse)
		assert.True(t, ok)
	})
}

func TestNamespaceController_V1NamespacesGet(t *testing.T) {
	t.Parallel()

	orgID := model.MustNewID(model.ResourceTypeOrganization)
	namespaceID := model.MustNewID(model.ResourceTypeNamespace)
	ns := &service.AccessibleNamespace{
		Namespace: service.Namespace{
			ID:            namespaceID,
			Slug:          "engineering",
			Name:          "Engineering",
			Description:   "Engineering team namespace",
			ProjectCount:  convert.ToPointer(int64(1)),
			DocumentCount: convert.ToPointer(int64(1)),
			CreatedAt:     convert.ToPointer(time.Now().UTC()),
		},
		Organization: service.PartialOrganization{
			ID:   orgID,
			Slug: "acme",
			Name: "ACME",
		},
	}

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		nsSvc := mocksvc.NewMockNamespaceService(ctrl)
		nsSvc.EXPECT().ListAccessible(gomock.Any(), gomock.Any()).Return(service.Page[*service.AccessibleNamespace]{
			Items: []*service.AccessibleNamespace{ns},
		}, nil)

		c, _ := newTestNamespaceController(t, ctrl, nsSvc)

		resp, err := c.V1NamespacesGet(context.Background(), api.V1NamespacesGetRequestObject{})
		require.NoError(t, err)
		got, ok := resp.(api.V1NamespacesGet200JSONResponse)
		require.True(t, ok)
		require.Len(t, got.Items, 1)
		assert.Equal(t, namespaceID.String(), got.Items[0].Id)
		assert.Equal(t, orgID.String(), got.Items[0].Organization.Id)
		assert.Equal(t, "ACME", got.Items[0].Organization.Name)
		assert.Equal(t, "acme", got.Items[0].Organization.Slug)
	})
}

func TestNamespaceController_V1OrganizationsNamespacesCreate(t *testing.T) {
	t.Parallel()

	orgID := model.MustNewID(model.ResourceTypeOrganization)
	namespaceID := model.MustNewID(model.ResourceTypeNamespace)

	t.Run("slug conflict", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)

		nsSvc := mocksvc.NewMockNamespaceService(ctrl)
		nsSvc.EXPECT().Create(gomock.Any(), orgID, service.CreateNamespaceOpts{
			Name: "Platform",
			Slug: "platform",
		}).Return(nil, errors.Join(service.ErrNamespaceCreate, repository.ErrSlugConflict))

		c, os := newTestNamespaceController(t, ctrl, nsSvc)
		stubOrganizationResolve(os, orgID)

		resp, err := c.V1OrganizationsNamespacesCreate(context.Background(), api.V1OrganizationsNamespacesCreateRequestObject{
			OrganizationRef: orgID.String(),
			Body: &api.V1OrganizationsNamespacesCreateJSONRequestBody{
				Name: "Platform",
				Slug: "platform",
			},
		})
		require.NoError(t, err)
		_, ok := resp.(api.V1OrganizationsNamespacesCreate409JSONResponse)
		assert.True(t, ok)
	})

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)

		nsSvc := mocksvc.NewMockNamespaceService(ctrl)
		nsSvc.EXPECT().Create(gomock.Any(), orgID, service.CreateNamespaceOpts{
			Name: "Platform",
			Slug: "platform",
		}).Return(&service.Namespace{ID: namespaceID, Slug: "platform", Name: "Platform"}, nil)

		c, os := newTestNamespaceController(t, ctrl, nsSvc)
		stubOrganizationResolve(os, orgID)

		resp, err := c.V1OrganizationsNamespacesCreate(context.Background(), api.V1OrganizationsNamespacesCreateRequestObject{
			OrganizationRef: orgID.String(),
			Body: &api.V1OrganizationsNamespacesCreateJSONRequestBody{
				Name: "Platform",
				Slug: "platform",
			},
		})
		require.NoError(t, err)
		got, ok := resp.(api.V1OrganizationsNamespacesCreate201JSONResponse)
		require.True(t, ok)
		assert.Equal(t, namespaceID.String(), got.Id)
	})
}

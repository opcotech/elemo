package http

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg/convert"
	"github.com/opcotech/elemo/internal/service"
	"github.com/opcotech/elemo/internal/transport/http/api"
)

func TestOrganizationToDTO(t *testing.T) {
	t.Parallel()

	createdAt := convert.ToPointer(time.Now().UTC())
	updatedAt := convert.ToPointer(time.Now().UTC())
	orgID := model.MustNewID(model.ResourceTypeOrganization)

	t.Run("with projected counts", func(t *testing.T) {
		t.Parallel()

		org := &service.Organization{
			ID:             orgID,
			Name:           "ACME Inc.",
			Email:          "info@example.com",
			Logo:           "https://example.com/logo.png",
			Website:        "https://example.com",
			Status:         model.OrganizationStatusActive,
			MemberCount:    convert.ToPointer(int64(12)),
			TeamCount:      convert.ToPointer(int64(3)),
			NamespaceCount: convert.ToPointer(int64(2)),
			DocumentCount:  convert.ToPointer(int64(4)),
			CreatedAt:      createdAt,
			UpdatedAt:      updatedAt,
		}

		dto := organizationToDTO(org)
		assert.Equal(t, orgID.String(), dto.Id)
		assert.Equal(t, "ACME Inc.", dto.Name)
		require.NotNil(t, dto.MemberCount)
		assert.Equal(t, int64(12), *dto.MemberCount)
		require.NotNil(t, dto.DocumentCount)
		assert.Equal(t, int64(4), *dto.DocumentCount)
		assert.Equal(t, api.OrganizationStatus(org.Status.String()), dto.Status)
	})

	t.Run("without projected counts", func(t *testing.T) {
		t.Parallel()

		org := &service.Organization{
			ID:        orgID,
			Name:      "ACME Inc.",
			Email:     "info@example.com",
			Status:    model.OrganizationStatusActive,
			CreatedAt: createdAt,
		}

		dto := organizationToDTO(org)
		assert.Equal(t, orgID.String(), dto.Id)
		assert.Nil(t, dto.MemberCount)
		assert.Nil(t, dto.DocumentCount)
		assert.Nil(t, dto.TeamCount)
		assert.Nil(t, dto.NamespaceCount)
	})
}

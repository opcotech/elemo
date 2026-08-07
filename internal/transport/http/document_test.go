package http

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg/convert"
	"github.com/opcotech/elemo/internal/service"
)

func TestPartialDocumentToDTO(t *testing.T) {
	t.Parallel()

	createdAt := convert.ToPointer(time.Now().UTC())
	docID := model.MustNewID(model.ResourceTypeDocument)
	userID := model.MustNewID(model.ResourceTypeUser)

	t.Run("with excerpt", func(t *testing.T) {
		t.Parallel()
		doc := &service.PartialDocument{
			ID:        docID,
			Name:      "Plan",
			Excerpt:   "Overview",
			CreatedBy: userID,
			CreatedAt: createdAt,
		}

		dto := partialDocumentToDTO(doc)
		assert.Equal(t, docID.String(), dto.Id)
		assert.Equal(t, "Plan", dto.Name)
		require.NotNil(t, dto.Excerpt)
		assert.Equal(t, "Overview", *dto.Excerpt)
		assert.Equal(t, userID.String(), dto.CreatedBy)
		assert.Equal(t, createdAt, dto.CreatedAt)
	})

	t.Run("without excerpt", func(t *testing.T) {
		t.Parallel()
		doc := &service.PartialDocument{
			ID:        docID,
			Name:      "Plan",
			CreatedBy: userID,
			CreatedAt: nil,
		}

		dto := partialDocumentToDTO(doc)
		assert.Equal(t, docID.String(), dto.Id)
		assert.Nil(t, dto.Excerpt)
		assert.Nil(t, dto.CreatedAt)
	})
}

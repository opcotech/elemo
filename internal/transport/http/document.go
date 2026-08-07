package http

import (
	"github.com/opcotech/elemo/internal/service"
	"github.com/opcotech/elemo/internal/transport/http/api"
)

func partialDocumentToDTO(document *service.PartialDocument) api.PartialDocument {
	nd := api.PartialDocument{
		Id:        document.ID.String(),
		Name:      document.Name,
		CreatedBy: document.CreatedBy.String(),
		CreatedAt: document.CreatedAt,
	}

	if document.Excerpt != "" {
		nd.Excerpt = &document.Excerpt
	}

	return nd
}

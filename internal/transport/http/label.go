package http

import (
	"context"
	"time"

	"github.com/opcotech/elemo/internal/service"
	"github.com/opcotech/elemo/internal/transport/http/api"
)

// LabelController is the controller for label endpoints.
type LabelController interface {
	V1LabelsGet(ctx context.Context, request api.V1LabelsGetRequestObject) (api.V1LabelsGetResponseObject, error)
}

// labelController is the concrete implementation of LabelController.
type labelController struct {
	*baseController
}

func (c *labelController) V1LabelsGet(ctx context.Context, request api.V1LabelsGetRequestObject) (api.V1LabelsGetResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1LabelsGet")
	defer span.End()

	pageParams, err := cursorPageFromParams(request.Params.PageSize, request.Params.PageToken)
	if err != nil {
		return api.V1LabelsGet400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	page, err := c.labelService.List(ctx, pageParams)
	if err != nil {
		if isInvalidPageError(err) {
			return api.V1LabelsGet400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
		}
		return api.V1LabelsGet500JSONResponse{N500JSONResponse: api.N500JSONResponse{
			Message: err.Error(),
		}}, nil
	}

	items := make([]api.Label, len(page.Items))
	for i, label := range page.Items {
		items[i] = labelToDTO(label)
	}

	return api.V1LabelsGet200JSONResponse{
		Items:    items,
		PageInfo: pageInfoToDTO(page.PageInfo),
	}, nil
}

func labelToDTO(label *service.Label) api.Label {
	var description *string
	if label.Description != "" {
		description = &label.Description
	}

	createdAt := time.Time{}
	if label.CreatedAt != nil {
		createdAt = *label.CreatedAt
	}

	return api.Label{
		Id:          label.ID.String(),
		Name:        label.Name,
		Description: description,
		CreatedAt:   createdAt,
		UpdatedAt:   label.UpdatedAt,
	}
}

// NewLabelController creates a new LabelController.
func NewLabelController(opts ...ControllerOption) (LabelController, error) {
	controller, err := newController(opts...)
	if err != nil {
		return nil, err
	}

	if controller.labelService == nil {
		return nil, ErrNoLabelService
	}

	return &labelController{
		baseController: controller,
	}, nil
}

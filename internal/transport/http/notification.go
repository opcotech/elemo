package http

import (
	"context"
	"errors"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg"
	"github.com/opcotech/elemo/internal/service"
	"github.com/opcotech/elemo/internal/transport/http/api"
)

// NotificationController is the controller for the notification endpoints.
type NotificationController interface {
	V1NotificationGet(ctx context.Context, request api.V1NotificationGetRequestObject) (api.V1NotificationGetResponseObject, error)
	V1NotificationsGet(ctx context.Context, request api.V1NotificationsGetRequestObject) (api.V1NotificationsGetResponseObject, error)
	V1NotificationUpdate(ctx context.Context, request api.V1NotificationUpdateRequestObject) (api.V1NotificationUpdateResponseObject, error)
	V1NotificationDelete(ctx context.Context, request api.V1NotificationDeleteRequestObject) (api.V1NotificationDeleteResponseObject, error)
}

// notificationController is the concrete implementation of NotificationController.
type notificationController struct {
	*baseController
}

func (c *notificationController) V1NotificationGet(ctx context.Context, request api.V1NotificationGetRequestObject) (api.V1NotificationGetResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1NotificationGet")
	defer span.End()

	recipientID, ok := ctx.Value(pkg.CtxKeyUserID).(model.ID)
	if !ok {
		return api.V1NotificationGet400JSONResponse{N400JSONResponse: formatBadRequest(model.ErrInvalidID)}, nil
	}

	notificationID, err := model.NewIDFromString(request.Id, model.ResourceTypeNotification.String())
	if err != nil {
		return api.V1NotificationGet400JSONResponse{N400JSONResponse: formatBadRequest(model.ErrInvalidID)}, nil
	}

	notification, err := c.notificationService.Get(ctx, notificationID, recipientID)
	if err != nil {
		if errors.Is(err, service.ErrNoPermission) {
			return api.V1NotificationGet403JSONResponse{N403JSONResponse: permissionDenied}, nil
		}
		if isNotFoundError(err) {
			return api.V1NotificationGet404JSONResponse{N404JSONResponse: notFound}, nil
		}
		return api.V1NotificationGet500JSONResponse{N500JSONResponse: api.N500JSONResponse{
			Message: err.Error(),
		}}, nil
	}

	return api.V1NotificationGet200JSONResponse(notificationToDTO(notification)), nil
}

func (c *notificationController) V1NotificationsGet(ctx context.Context, request api.V1NotificationsGetRequestObject) (api.V1NotificationsGetResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1NotificationsGet")
	defer span.End()

	recipientID, ok := ctx.Value(pkg.CtxKeyUserID).(model.ID)
	if !ok {
		return api.V1NotificationsGet400JSONResponse{N400JSONResponse: formatBadRequest(model.ErrInvalidID)}, nil
	}

	notifications, err := c.notificationService.GetAllByRecipient(ctx,
		recipientID,
		pkg.GetDefaultPtr(request.Params.Offset, DefaultOffset),
		pkg.GetDefaultPtr(request.Params.Limit, DefaultLimit),
	)
	if err != nil {
		if errors.Is(err, service.ErrNoPermission) {
			return api.V1NotificationsGet403JSONResponse{N403JSONResponse: permissionDenied}, nil
		}
		return api.V1NotificationsGet500JSONResponse{N500JSONResponse: api.N500JSONResponse{
			Message: err.Error(),
		}}, nil
	}

	notificationsDTO := make([]api.Notification, len(notifications))
	for i, notification := range notifications {
		notificationsDTO[i] = notificationToDTO(notification)
	}

	return api.V1NotificationsGet200JSONResponse(notificationsDTO), nil
}

func (c *notificationController) V1NotificationUpdate(ctx context.Context, request api.V1NotificationUpdateRequestObject) (api.V1NotificationUpdateResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1NotificationUpdate")
	defer span.End()

	recipientID, ok := ctx.Value(pkg.CtxKeyUserID).(model.ID)
	if !ok {
		return api.V1NotificationUpdate400JSONResponse{N400JSONResponse: formatBadRequest(model.ErrInvalidID)}, nil
	}

	notificationID, err := model.NewIDFromString(request.Id, model.ResourceTypeNotification.String())
	if err != nil {
		return api.V1NotificationUpdate400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	notification, err := c.notificationService.Update(ctx, notificationID, recipientID, request.Body.Read)
	if err != nil {
		if errors.Is(err, service.ErrNoPermission) {
			return api.V1NotificationUpdate403JSONResponse{N403JSONResponse: permissionDenied}, nil
		}
		if isNotFoundError(err) {
			return api.V1NotificationUpdate404JSONResponse{N404JSONResponse: notFound}, nil
		}
		return api.V1NotificationUpdate500JSONResponse{N500JSONResponse: api.N500JSONResponse{
			Message: err.Error(),
		}}, nil
	}

	return api.V1NotificationUpdate200JSONResponse(notificationToDTO(notification)), nil
}

func (c *notificationController) V1NotificationDelete(ctx context.Context, request api.V1NotificationDeleteRequestObject) (api.V1NotificationDeleteResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1NotificationDelete")
	defer span.End()

	recipientID, ok := ctx.Value(pkg.CtxKeyUserID).(model.ID)
	if !ok {
		return api.V1NotificationDelete400JSONResponse{N400JSONResponse: formatBadRequest(model.ErrInvalidID)}, nil
	}

	notificationID, err := model.NewIDFromString(request.Id, model.ResourceTypeNotification.String())
	if err != nil {
		return api.V1NotificationDelete404JSONResponse{N404JSONResponse: notFound}, nil
	}

	if err := c.notificationService.Delete(ctx, notificationID, recipientID); err != nil {
		if errors.Is(err, service.ErrNoPermission) {
			return api.V1NotificationDelete403JSONResponse{N403JSONResponse: permissionDenied}, nil
		}
		if isNotFoundError(err) {
			return api.V1NotificationDelete404JSONResponse{N404JSONResponse: notFound}, nil
		}
		return api.V1NotificationDelete500JSONResponse{N500JSONResponse: api.N500JSONResponse{
			Message: err.Error(),
		}}, nil
	}

	return api.V1NotificationDelete204Response{}, nil
}

// NewNotificationController creates a new NotificationController.
func NewNotificationController(opts ...ControllerOption) (NotificationController, error) {
	c, err := newController(opts...)
	if err != nil {
		return nil, err
	}

	controller := &notificationController{
		baseController: c,
	}

	if controller.notificationService == nil {
		return nil, ErrNoNotificationService
	}

	if controller.userService == nil {
		return nil, ErrNoNotificationService
	}

	return controller, nil
}

func notificationToDTO(notification *model.Notification) api.Notification {
	return api.Notification{
		Id:          notification.ID.String(),
		Title:       notification.Title,
		Description: notification.Description,
		Recipient:   notification.Recipient.String(),
		Read:        notification.Read,
		CreatedAt:   *notification.CreatedAt,
		UpdatedAt:   notification.UpdatedAt,
	}
}

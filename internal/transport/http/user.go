package http

import (
	"context"
	"errors"
	"net/http"

	openapiTypes "github.com/deepmap/oapi-codegen/pkg/types"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg"
	"github.com/opcotech/elemo/internal/pkg/convert"
	"github.com/opcotech/elemo/internal/repository"
	"github.com/opcotech/elemo/internal/service"
	"github.com/opcotech/elemo/internal/transport/http/gen"
)

// UserController is a controller for user endpoints.
type UserController interface {
	GetUser(ctx context.Context, request gen.GetUserRequestObject) (gen.GetUserResponseObject, error)
}

// userController is the concrete implementation of UserController.
type userController struct {
	*baseController
}

func (c *userController) GetUser(ctx context.Context, request gen.GetUserRequestObject) (gen.GetUserResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/GetUser")
	defer span.End()

	var userID model.ID
	var err error

	if request.UserId == "me" {
		var ok bool
		if userID, ok = ctx.Value(pkg.CtxKeyUserID).(model.ID); !ok {
			return gen.GetUser404JSONResponse{N404JSONResponse: notFound}, nil
		}
	} else {
		if userID, err = model.NewIDFromString(request.UserId, model.UserIDType); err != nil {
			return gen.GetUser404JSONResponse{N404JSONResponse: notFound}, nil
		}
	}

	user, err := c.userService.Get(ctx, userID)
	if err != nil {
		if errors.Is(err, service.ErrNoPermission) {
			return gen.GetUser401JSONResponse{N401JSONResponse: permissionDenied}, nil
		}
		if errors.Is(err, repository.ErrNotFound) {
			return gen.GetUser404JSONResponse{N404JSONResponse: notFound}, nil
		}
		return gen.GetUserdefaultJSONResponse{
			Body: gen.HTTPError{
				Message: err.Error(),
			},
			StatusCode: http.StatusInternalServerError,
		}, nil
	}

	return gen.GetUser200JSONResponse(*userToDTO(user)), nil
}

// NewUserController creates a new UserController.
func NewUserController(opts ...ControllerOption) (UserController, error) {
	c, err := newController(opts...)
	if err != nil {
		return nil, err
	}

	controller := &userController{
		baseController: c,
	}

	if controller.userService == nil {
		return nil, ErrNoUserService
	}

	return controller, nil
}

func userToDTO(user *model.User) *gen.User {
	u := &gen.User{
		Id:        convert.ToPointer(user.ID.String()),
		Address:   &user.Address,
		Bio:       &user.Bio,
		Email:     convert.ToPointer(openapiTypes.Email(user.Email)),
		FirstName: &user.FirstName,
		LastName:  &user.LastName,
		Links:     &user.Links,
		Username:  &user.Username,
		Phone:     &user.Phone,
		Picture:   &user.Picture,
		Status:    convert.ToPointer(gen.UserStatus(user.Status.String())),
		Title:     &user.Title,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}

	languages := make([]gen.Language, len(user.Languages))
	for i, language := range user.Languages {
		languages[i] = gen.Language(language.String())
	}

	u.Languages = &languages

	return u
}

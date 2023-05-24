package http

import (
	"context"
	"errors"

	oapiTypes "github.com/deepmap/oapi-codegen/pkg/types"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg"
	"github.com/opcotech/elemo/internal/pkg/convert"
	"github.com/opcotech/elemo/internal/pkg/password"
	"github.com/opcotech/elemo/internal/repository"
	"github.com/opcotech/elemo/internal/service"
	"github.com/opcotech/elemo/internal/transport/http/api"
)

// UserController is a controller for user endpoints.
type UserController interface {
	V1UsersCreate(ctx context.Context, request api.V1UsersCreateRequestObject) (api.V1UsersCreateResponseObject, error)
	V1UserGet(ctx context.Context, request api.V1UserGetRequestObject) (api.V1UserGetResponseObject, error)
	V1UsersGet(ctx context.Context, request api.V1UsersGetRequestObject) (api.V1UsersGetResponseObject, error)
	V1UserUpdate(ctx context.Context, request api.V1UserUpdateRequestObject) (api.V1UserUpdateResponseObject, error)
	V1UserDelete(ctx context.Context, request api.V1UserDeleteRequestObject) (api.V1UserDeleteResponseObject, error)
}

// userController is the concrete implementation of UserController.
type userController struct {
	*baseController
}

func (c *userController) V1UsersCreate(ctx context.Context, request api.V1UsersCreateRequestObject) (api.V1UsersCreateResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1UsersCreate")
	defer span.End()

	user, err := createUserJSONRequestBodyToUser(request.Body)
	if err != nil {
		return api.V1UsersCreate400JSONResponse{N400JSONResponse: badRequest}, nil
	}

	if err := c.userService.Create(ctx, user); err != nil {
		if errors.Is(err, service.ErrNoPermission) {
			return api.V1UsersCreate403JSONResponse{N403JSONResponse: permissionDenied}, nil
		}
		return api.V1UsersCreate500JSONResponse{
			N500JSONResponse: api.N500JSONResponse{
				Message: err.Error(),
			},
		}, nil
	}

	return api.V1UsersCreate201JSONResponse{N201JSONResponse: api.N201JSONResponse{
		Id: user.ID.String(),
	}}, nil
}

func (c *userController) V1UserGet(ctx context.Context, request api.V1UserGetRequestObject) (api.V1UserGetResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1UserGet")
	defer span.End()

	var userID model.ID
	var err error

	if request.Id == "me" {
		var ok bool
		if userID, ok = ctx.Value(pkg.CtxKeyUserID).(model.ID); !ok {
			return api.V1UserGet400JSONResponse{N400JSONResponse: badRequest}, nil
		}
	} else {
		if userID, err = model.NewIDFromString(request.Id, model.ResourceTypeUser.String()); err != nil {
			return api.V1UserGet400JSONResponse{N400JSONResponse: badRequest}, nil
		}
	}

	user, err := c.userService.Get(ctx, userID)
	if err != nil {
		if errors.Is(err, service.ErrNoPermission) {
			return api.V1UserGet403JSONResponse{N403JSONResponse: permissionDenied}, nil
		}
		if errors.Is(err, repository.ErrNotFound) {
			return api.V1UserGet404JSONResponse{N404JSONResponse: notFound}, nil
		}
		return api.V1UserGet500JSONResponse{N500JSONResponse: api.N500JSONResponse{
			Message: err.Error(),
		}}, nil
	}

	return api.V1UserGet200JSONResponse(userToDTO(user)), nil
}

func (c *userController) V1UsersGet(ctx context.Context, request api.V1UsersGetRequestObject) (api.V1UsersGetResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1UsersGet")
	defer span.End()

	users, err := c.userService.GetAll(ctx, pkg.GetDefaultPtr(request.Params.Offset, DefaultOffset), pkg.GetDefaultPtr(request.Params.Limit, DefaultLimit))
	if err != nil {
		if errors.Is(err, service.ErrNoPermission) {
			return api.V1UsersGet403JSONResponse{N403JSONResponse: permissionDenied}, nil
		}
		return api.V1UsersGet500JSONResponse{N500JSONResponse: api.N500JSONResponse{
			Message: err.Error(),
		}}, nil
	}

	usersDTO := make([]api.User, len(users))
	for i, user := range users {
		usersDTO[i] = userToDTO(user)
	}

	return api.V1UsersGet200JSONResponse(usersDTO), nil
}

func (c *userController) V1UserUpdate(ctx context.Context, request api.V1UserUpdateRequestObject) (api.V1UserUpdateResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1UserUpdate")
	defer span.End()

	userID, err := model.NewIDFromString(request.Id, model.ResourceTypeUser.String())
	if err != nil {
		return api.V1UserUpdate404JSONResponse{N404JSONResponse: notFound}, nil
	}

	if request.Body.Password != nil {
		request.Body.Password = convert.ToPointer(password.HashPassword(*request.Body.Password))
	}

	patch := make(map[string]any)
	if err := convert.AnyToAny(request.Body, &patch); err != nil {
		return api.V1UserUpdate400JSONResponse{N400JSONResponse: badRequest}, nil
	}

	user, err := c.userService.Update(ctx, userID, patch)
	if err != nil {
		if errors.Is(err, service.ErrNoPermission) {
			return api.V1UserUpdate403JSONResponse{N403JSONResponse: permissionDenied}, nil
		}
		if errors.Is(err, repository.ErrNotFound) {
			return api.V1UserUpdate404JSONResponse{N404JSONResponse: notFound}, nil
		}
		return api.V1UserUpdate500JSONResponse{N500JSONResponse: api.N500JSONResponse{
			Message: err.Error(),
		}}, nil
	}

	return api.V1UserUpdate200JSONResponse(userToDTO(user)), nil
}

func (c *userController) V1UserDelete(ctx context.Context, request api.V1UserDeleteRequestObject) (api.V1UserDeleteResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1UserDelete")
	defer span.End()

	userID, err := model.NewIDFromString(request.Id, model.ResourceTypeUser.String())
	if err != nil {
		return api.V1UserDelete404JSONResponse{N404JSONResponse: notFound}, nil
	}

	if err := c.userService.Delete(ctx, userID, pkg.GetDefaultPtr(request.Params.Force, false)); err != nil {
		if errors.Is(err, service.ErrNoPermission) {
			return api.V1UserDelete403JSONResponse{N403JSONResponse: permissionDenied}, nil
		}
		if errors.Is(err, repository.ErrNotFound) {
			return api.V1UserDelete404JSONResponse{N404JSONResponse: notFound}, nil
		}
		return api.V1UserDelete500JSONResponse{N500JSONResponse: api.N500JSONResponse{
			Message: err.Error(),
		}}, nil
	}

	return api.V1UserDelete204Response{}, nil
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

func createUserJSONRequestBodyToUser(body *api.V1UsersCreateJSONRequestBody) (*model.User, error) {
	user, err := model.NewUser(body.Username, string(body.Email), password.HashPassword(body.Password))
	if err != nil {
		return nil, err
	}

	user.FirstName = pkg.GetDefaultPtr(body.FirstName, "")
	user.LastName = pkg.GetDefaultPtr(body.LastName, "")
	user.Title = pkg.GetDefaultPtr(body.Title, "")
	user.Picture = pkg.GetDefaultPtr(body.Picture, "")
	user.Bio = pkg.GetDefaultPtr(body.Bio, "")
	user.Address = pkg.GetDefaultPtr(body.Address, "")
	user.Phone = pkg.GetDefaultPtr(body.Phone, "")
	user.Links = pkg.GetDefaultPtr(body.Links, make([]string, 0))

	if body.Languages != nil {
		user.Languages = make([]model.Language, len(*body.Languages))
		for i, language := range *body.Languages {
			var lang model.Language
			if err := lang.UnmarshalText([]byte(language)); err != nil {
				return nil, err
			}
			user.Languages[i] = lang
		}
	}

	return user, nil
}

func userToDTO(user *model.User) api.User {
	u := api.User{
		Id:        user.ID.String(),
		Address:   &user.Address,
		Bio:       &user.Bio,
		Email:     oapiTypes.Email(user.Email),
		FirstName: &user.FirstName,
		LastName:  &user.LastName,
		Links:     &user.Links,
		Username:  user.Username,
		Phone:     &user.Phone,
		Picture:   &user.Picture,
		Status:    api.UserStatus(user.Status.String()),
		Title:     &user.Title,
		Languages: make([]api.Language, len(user.Languages)),
		CreatedAt: *user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}

	for i, language := range user.Languages {
		u.Languages[i] = api.Language(language.String())
	}

	return u
}

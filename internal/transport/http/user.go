package http

import (
	"context"
	"errors"
	"net/http"

	oapiTypes "github.com/deepmap/oapi-codegen/pkg/types"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg"
	"github.com/opcotech/elemo/internal/pkg/convert"
	"github.com/opcotech/elemo/internal/pkg/password"
	"github.com/opcotech/elemo/internal/repository"
	"github.com/opcotech/elemo/internal/service"
	"github.com/opcotech/elemo/internal/transport/http/gen"
)

// UserController is a controller for user endpoints.
type UserController interface {
	CreateUser(ctx context.Context, request gen.CreateUserRequestObject) (gen.CreateUserResponseObject, error)
	GetUser(ctx context.Context, request gen.GetUserRequestObject) (gen.GetUserResponseObject, error)
	GetUsers(ctx context.Context, request gen.GetUsersRequestObject) (gen.GetUsersResponseObject, error)
	UpdateUser(ctx context.Context, request gen.UpdateUserRequestObject) (gen.UpdateUserResponseObject, error)
	DeleteUser(ctx context.Context, request gen.DeleteUserRequestObject) (gen.DeleteUserResponseObject, error)
}

// userController is the concrete implementation of UserController.
type userController struct {
	*baseController
}

func (c *userController) CreateUser(ctx context.Context, request gen.CreateUserRequestObject) (gen.CreateUserResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/CreateUser")
	defer span.End()

	user, err := createUserJSONRequestBodyToUser(request.Body)
	if err != nil {
		return gen.CreateUser400JSONResponse{N400JSONResponse: badRequest}, nil
	}

	if err := c.userService.Create(ctx, user); err != nil {
		if errors.Is(err, service.ErrNoPermission) {
			return gen.CreateUser401JSONResponse{N401JSONResponse: permissionDenied}, nil
		}
		return gen.CreateUserdefaultJSONResponse{
			Body: gen.HTTPError{
				Message: err.Error(),
			},
			StatusCode: http.StatusInternalServerError,
		}, nil
	}

	return gen.CreateUser201JSONResponse{
		UserId: user.ID.String(),
	}, nil
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
		if userID, err = model.NewIDFromString(request.UserId, model.ResourceTypeUser.String()); err != nil {
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

	return gen.GetUser200JSONResponse(userToDTO(user)), nil
}

func (c *userController) GetUsers(ctx context.Context, request gen.GetUsersRequestObject) (gen.GetUsersResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/GetUsers")
	defer span.End()

	users, err := c.userService.GetAll(ctx, pkg.GetDefaultPtr(request.Params.Offset, DefaultOffset), pkg.GetDefaultPtr(request.Params.Limit, DefaultLimit))
	if err != nil {
		return gen.GetUsersdefaultJSONResponse{
			Body: gen.HTTPError{
				Message: err.Error(),
			},
			StatusCode: http.StatusInternalServerError,
		}, nil
	}

	usersDTO := make([]gen.User, len(users))
	for i, user := range users {
		usersDTO[i] = userToDTO(user)
	}

	return gen.GetUsers200JSONResponse(usersDTO), nil
}

func (c *userController) UpdateUser(ctx context.Context, request gen.UpdateUserRequestObject) (gen.UpdateUserResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/UpdateUser")
	defer span.End()

	userID, err := model.NewIDFromString(request.UserId, model.ResourceTypeUser.String())
	if err != nil {
		return gen.UpdateUser404JSONResponse{N404JSONResponse: notFound}, nil
	}

	if request.Body.Password != nil {
		request.Body.Password = convert.ToPointer(password.HashPassword(*request.Body.Password))
	}

	patch := make(map[string]any)
	if err := convert.AnyToAny(request.Body, &patch); err != nil {
		return gen.UpdateUser400JSONResponse{N400JSONResponse: badRequest}, nil
	}

	user, err := c.userService.Update(ctx, userID, patch)
	if err != nil {
		if errors.Is(err, service.ErrNoPermission) {
			return gen.UpdateUser401JSONResponse{N401JSONResponse: permissionDenied}, nil
		}
		if errors.Is(err, repository.ErrNotFound) {
			return gen.UpdateUser404JSONResponse{N404JSONResponse: notFound}, nil
		}
		return gen.UpdateUserdefaultJSONResponse{
			Body: gen.HTTPError{
				Message: err.Error(),
			},
			StatusCode: http.StatusInternalServerError,
		}, nil
	}

	return gen.UpdateUser200JSONResponse(userToDTO(user)), nil
}

func (c *userController) DeleteUser(ctx context.Context, request gen.DeleteUserRequestObject) (gen.DeleteUserResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/DeleteUser")
	defer span.End()

	userID, err := model.NewIDFromString(request.UserId, model.ResourceTypeUser.String())
	if err != nil {
		return gen.DeleteUser400JSONResponse{N400JSONResponse: badRequest}, nil
	}

	if err := c.userService.Delete(ctx, userID, pkg.GetDefaultPtr(request.Params.Force, false)); err != nil {
		if errors.Is(err, service.ErrNoPermission) {
			return gen.DeleteUser401JSONResponse{N401JSONResponse: permissionDenied}, nil
		}
		if errors.Is(err, repository.ErrNotFound) {
			return gen.DeleteUser404JSONResponse{N404JSONResponse: notFound}, nil
		}
		return gen.DeleteUserdefaultJSONResponse{
			Body: gen.HTTPError{
				Message: err.Error(),
			},
			StatusCode: http.StatusInternalServerError,
		}, nil
	}

	return gen.DeleteUser204JSONResponse{}, nil
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

func userToDTO(user *model.User) gen.User {
	u := gen.User{
		Id:        convert.ToPointer(user.ID.String()),
		Address:   &user.Address,
		Bio:       &user.Bio,
		Email:     convert.ToPointer(oapiTypes.Email(user.Email)),
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

func createUserJSONRequestBodyToUser(body *gen.CreateUserJSONRequestBody) (*model.User, error) {
	user := &model.User{
		ID:        model.MustNewNilID(model.ResourceTypeUser),
		Username:  body.Username,
		FirstName: body.FirstName,
		LastName:  body.LastName,
		Email:     string(body.Email),
		Title:     pkg.GetDefaultPtr(body.Title, ""),
		Picture:   pkg.GetDefaultPtr(body.Picture, ""),
		Bio:       pkg.GetDefaultPtr(body.Bio, ""),
		Address:   pkg.GetDefaultPtr(body.Address, ""),
		Phone:     pkg.GetDefaultPtr(body.Phone, ""),
		Links:     pkg.GetDefaultPtr(body.Links, make([]string, 0)),
	}

	if body.Password != nil {
		user.Password = password.HashPassword(*body.Password)
	} else {
		user.Password = password.UnusablePassword
	}

	if body.Status != nil {
		if err := user.Status.UnmarshalText([]byte(*body.Status)); err != nil {
			return nil, err
		}
	} else {
		user.Status = model.UserStatusActive
	}

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

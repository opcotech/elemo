package http

import (
	"context"
	"errors"

	oapiTypes "github.com/oapi-codegen/runtime/types"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg"
	"github.com/opcotech/elemo/internal/pkg/auth"
	"github.com/opcotech/elemo/internal/pkg/convert"
	"github.com/opcotech/elemo/internal/pkg/password"
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
	V1UserRequestPasswordReset(ctx context.Context, request api.V1UserRequestPasswordResetRequestObject) (api.V1UserRequestPasswordResetResponseObject, error)
	V1UserResetPassword(ctx context.Context, request api.V1UserResetPasswordRequestObject) (api.V1UserResetPasswordResponseObject, error)
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
		return api.V1UsersCreate400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
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
			return api.V1UserGet400JSONResponse{N400JSONResponse: formatBadRequest(model.ErrInvalidID)}, nil
		}
	} else {
		if userID, err = model.NewIDFromString(request.Id, model.ResourceTypeUser.String()); err != nil {
			return api.V1UserGet400JSONResponse{N400JSONResponse: formatBadRequest(model.ErrInvalidID)}, nil
		}
	}

	user, err := c.userService.Get(ctx, userID)
	if err != nil {
		if errors.Is(err, service.ErrNoPermission) {
			return api.V1UserGet403JSONResponse{N403JSONResponse: permissionDenied}, nil
		}
		if isNotFoundError(err) {
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

	patch, err := api.ConvertRequestToMap(request.Body)
	if err != nil {
		return api.V1UserUpdate400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	if request.Body.Password != nil && request.Body.NewPassword == nil || request.Body.Password == nil && request.Body.NewPassword != nil {
		return api.V1UserUpdate400JSONResponse{N400JSONResponse: api.N400JSONResponse{
			Message: "The old password and the new password must be provided together",
		}}, nil
	}

	if request.Body.Password != nil && request.Body.NewPassword != nil {
		user, err := c.userService.Get(ctx, userID)
		if err != nil {
			if errors.Is(err, service.ErrNoPermission) {
				return api.V1UserUpdate403JSONResponse{N403JSONResponse: permissionDenied}, nil
			}
			if isNotFoundError(err) {
				return api.V1UserUpdate404JSONResponse{N404JSONResponse: notFound}, nil
			}
			return api.V1UserUpdate500JSONResponse{N500JSONResponse: api.N500JSONResponse{
				Message: err.Error(),
			}}, nil
		}

		if !password.IsPasswordMatching(user.Password, *request.Body.Password) {
			return api.V1UserUpdate400JSONResponse{N400JSONResponse: api.N400JSONResponse{
				Message: "The provided password is incorrect",
			}}, nil
		}

		if password.IsPasswordMatching(user.Password, *request.Body.NewPassword) {
			return api.V1UserUpdate400JSONResponse{N400JSONResponse: api.N400JSONResponse{
				Message: "The new password is the same as the old one",
			}}, nil
		}

		// Update the patch to use the new password hash for the password field
		patch["password"] = convert.ToPointer(password.HashPassword(*request.Body.NewPassword))
		delete(patch, "new_password")
	}

	user, err := c.userService.Update(ctx, userID, patch)
	if err != nil {
		if errors.Is(err, service.ErrNoPermission) {
			return api.V1UserUpdate403JSONResponse{N403JSONResponse: permissionDenied}, nil
		}
		if isNotFoundError(err) {
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
		if isNotFoundError(err) {
			return api.V1UserDelete404JSONResponse{N404JSONResponse: notFound}, nil
		}
		return api.V1UserDelete500JSONResponse{N500JSONResponse: api.N500JSONResponse{
			Message: err.Error(),
		}}, nil
	}

	return api.V1UserDelete204Response{}, nil
}

func (c *userController) V1UserRequestPasswordReset(ctx context.Context, request api.V1UserRequestPasswordResetRequestObject) (api.V1UserRequestPasswordResetResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1UserRequestPasswordReset")
	defer span.End()

	ctx = context.WithValue(ctx, pkg.CtxKeyUserID, pkg.CtxMachineUser)

	if request.Params.Email == "" {
		return api.V1UserRequestPasswordReset400JSONResponse{
			N400JSONResponse: formatBadRequest(errors.New("email is required")),
		}, nil
	}

	user, err := c.userService.GetByEmail(ctx, string(request.Params.Email))
	if err != nil {
		if isNotFoundError(err) {
			return api.V1UserRequestPasswordReset404JSONResponse{N404JSONResponse: notFound}, nil
		}
		return api.V1UserRequestPasswordReset500JSONResponse{N500JSONResponse: api.N500JSONResponse{
			Message: err.Error(),
		}}, nil
	}

	token, err := c.userService.CreateToken(ctx, user.ID, user.Email, model.UserTokenContextResetPassword, nil)
	if err != nil {
		return api.V1UserRequestPasswordReset400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	if err := c.emailService.SendAuthPasswordResetEmail(ctx, user, token); err != nil {
		return api.V1UserRequestPasswordReset500JSONResponse{
			N500JSONResponse: api.N500JSONResponse{
				Message: err.Error(),
			},
		}, nil
	}

	return api.V1UserRequestPasswordReset200Response{}, nil
}

func (c *userController) V1UserResetPassword(ctx context.Context, request api.V1UserResetPasswordRequestObject) (api.V1UserResetPasswordResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1UserResetPassword")
	defer span.End()

	ctx = context.WithValue(ctx, pkg.CtxKeyUserID, pkg.CtxMachineUser)

	tokenData, verifyErr := c.userService.VerifyToken(ctx, request.Body.Token)
	if verifyErr != nil && !errors.Is(verifyErr, service.ErrExpiredToken) {
		return api.V1UserResetPassword400JSONResponse{N400JSONResponse: formatBadRequest(verifyErr)}, nil
	}

	userID, err := model.NewIDFromString(tokenData["user_id"].(string), model.ResourceTypeUser.String())
	if err != nil {
		return api.V1UserResetPassword400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	user, err := c.userService.Get(ctx, userID)
	if err != nil {
		if isNotFoundError(err) {
			return api.V1UserResetPassword404JSONResponse{N404JSONResponse: notFound}, nil
		}
		return api.V1UserResetPassword500JSONResponse{N500JSONResponse: api.N500JSONResponse{
			Message: err.Error(),
		}}, nil
	}

	if verifyErr != nil && errors.Is(verifyErr, service.ErrExpiredToken) {
		if err := c.userService.DeleteToken(ctx, userID, model.UserTokenContextResetPassword); err != nil {
			if !isNotFoundError(err) {
				return api.V1UserResetPassword500JSONResponse{N500JSONResponse: api.N500JSONResponse{
					Message: err.Error(),
				}}, nil
			}
		}

		token, err := c.userService.CreateToken(ctx, user.ID, user.Email, model.UserTokenContextResetPassword, nil)
		if err != nil {
			return api.V1UserResetPassword400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
		}

		if err := c.emailService.SendAuthPasswordResetEmail(ctx, user, token); err != nil {
			return api.V1UserResetPassword500JSONResponse{
				N500JSONResponse: api.N500JSONResponse{
					Message: err.Error(),
				},
			}, nil
		}

		return api.V1UserResetPassword204Response{}, nil
	}

	patch, err := api.ConvertRequestToMap(request.Body)
	if err != nil {
		return api.V1UserResetPassword400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	// Override the password in the patch to hash it
	patch["password"] = auth.HashPassword(request.Body.Password)

	// Set the user ID in context for the update operation
	ctx = context.WithValue(ctx, pkg.CtxKeyUserID, user.ID)

	if _, err = c.userService.Update(ctx, user.ID, patch); err != nil {
		return api.V1UserResetPassword500JSONResponse{
			N500JSONResponse: api.N500JSONResponse{
				Message: err.Error(),
			},
		}, nil
	}

	if err := c.userService.DeleteToken(ctx, userID, model.UserTokenContextResetPassword); err != nil {
		if !isNotFoundError(err) {
			return api.V1UserResetPassword500JSONResponse{N500JSONResponse: api.N500JSONResponse{
				Message: err.Error(),
			}}, nil
		}
	}

	return api.V1UserResetPassword200Response{}, nil
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

	if controller.emailService == nil {
		return nil, ErrNoEmailService
	}

	return controller, nil
}

func createUserJSONRequestBodyToUser(body *api.V1UsersCreateJSONRequestBody) (*model.User, error) {
    user, err := model.NewUser(body.Username, string(body.Email), password.HashPassword(body.Password))
    if err != nil {
        return nil, err
    }

    if body.FirstName == "" {
        return nil, errors.New("FirstName is required")
    }
    if body.LastName == "" {
        return nil, errors.New("LastName is required")
    }

    user.FirstName = body.FirstName
    user.LastName = body.LastName

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
		FirstName: "Test",
		LastName:  "User",
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

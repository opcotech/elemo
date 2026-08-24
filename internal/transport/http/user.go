package http

import (
	"context"
	"errors"
	"net/http"

	oapiTypes "github.com/oapi-codegen/runtime/types"

	"github.com/opcotech/elemo/internal/email"
	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg"
	"github.com/opcotech/elemo/internal/pkg/auth"
	"github.com/opcotech/elemo/internal/pkg/optional"
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
	userService  service.UserService
	emailService service.EmailService
}

func (c *userController) V1UsersCreate(ctx context.Context, request api.V1UsersCreateRequestObject) (api.V1UsersCreateResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1UsersCreate")
	defer span.End()

	opts, err := createUserJSONRequestBodyToCreateUserOpts(request.Body)
	if err != nil {
		return api.V1UsersCreate400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	user, err := c.userService.Create(ctx, opts)
	if err != nil {
		switch classifyServiceError(err) {
		case http.StatusForbidden:
			return api.V1UsersCreate403JSONResponse{N403JSONResponse: permissionDenied}, nil
		default:
			return api.V1UsersCreate500JSONResponse{N500JSONResponse: api.N500JSONResponse{
				Message: err.Error(),
			}}, nil
		}
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
		switch classifyServiceError(err) {
		case http.StatusForbidden:
			return api.V1UserGet403JSONResponse{N403JSONResponse: permissionDenied}, nil
		case http.StatusNotFound:
			return api.V1UserGet404JSONResponse{N404JSONResponse: notFound}, nil
		default:
			return api.V1UserGet500JSONResponse{N500JSONResponse: api.N500JSONResponse{
				Message: err.Error(),
			}}, nil
		}
	}

	return api.V1UserGet200JSONResponse(userToDTO(user)), nil
}

func (c *userController) V1UsersGet(ctx context.Context, request api.V1UsersGetRequestObject) (api.V1UsersGetResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1UsersGet")
	defer span.End()

	pageParams, err := cursorPageFromParams(request.Params.PageSize, request.Params.PageToken)
	if err != nil {
		return api.V1UsersGet400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	page, err := c.userService.List(ctx, pageParams)
	if err != nil {
		switch classifyServiceError(err) {
		case http.StatusBadRequest:
			return api.V1UsersGet400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
		case http.StatusForbidden:
			return api.V1UsersGet403JSONResponse{N403JSONResponse: permissionDenied}, nil
		default:
			return api.V1UsersGet500JSONResponse{N500JSONResponse: api.N500JSONResponse{
				Message: err.Error(),
			}}, nil
		}
	}

	usersDTO := make([]api.User, len(page.Items))
	for i, user := range page.Items {
		usersDTO[i] = userToDTO(user)
	}

	return api.V1UsersGet200JSONResponse{
		Items:    usersDTO,
		PageInfo: pageInfoToDTO(page.PageInfo),
	}, nil
}

func (c *userController) V1UserUpdate(ctx context.Context, request api.V1UserUpdateRequestObject) (api.V1UserUpdateResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1UserUpdate")
	defer span.End()

	userID, err := model.NewIDFromString(request.Id, model.ResourceTypeUser.String())
	if err != nil {
		return api.V1UserUpdate404JSONResponse{N404JSONResponse: notFound}, nil
	}

	if request.Body.Password != nil && request.Body.NewPassword == nil || request.Body.Password == nil && request.Body.NewPassword != nil {
		return api.V1UserUpdate400JSONResponse{N400JSONResponse: api.N400JSONResponse{
			Message: "The old password and the new password must be provided together",
		}}, nil
	}

	opts, err := updateUserJSONRequestBodyToUpdateUserOpts(request.Body)
	if err != nil {
		return api.V1UserUpdate400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	if request.Body.Password != nil && request.Body.NewPassword != nil {
		user, err := c.userService.Get(ctx, userID)
		if err != nil {
			switch classifyServiceError(err) {
			case http.StatusForbidden:
				return api.V1UserUpdate403JSONResponse{N403JSONResponse: permissionDenied}, nil
			case http.StatusNotFound:
				return api.V1UserUpdate404JSONResponse{N404JSONResponse: notFound}, nil
			default:
				return api.V1UserUpdate500JSONResponse{N500JSONResponse: api.N500JSONResponse{
					Message: err.Error(),
				}}, nil
			}
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

		opts.Password = optional.Some(password.HashPassword(*request.Body.NewPassword))
	}

	user, err := c.userService.Update(ctx, userID, opts)
	if err != nil {
		switch classifyServiceError(err) {
		case http.StatusForbidden:
			return api.V1UserUpdate403JSONResponse{N403JSONResponse: permissionDenied}, nil
		case http.StatusNotFound:
			return api.V1UserUpdate404JSONResponse{N404JSONResponse: notFound}, nil
		default:
			return api.V1UserUpdate500JSONResponse{N500JSONResponse: api.N500JSONResponse{
				Message: err.Error(),
			}}, nil
		}
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

	if err := c.userService.Delete(ctx, userID, pkg.DefaultPtr(request.Params.Force, false)); err != nil {
		switch classifyServiceError(err) {
		case http.StatusForbidden:
			return api.V1UserDelete403JSONResponse{N403JSONResponse: permissionDenied}, nil
		case http.StatusNotFound:
			return api.V1UserDelete404JSONResponse{N404JSONResponse: notFound}, nil
		default:
			return api.V1UserDelete500JSONResponse{N500JSONResponse: api.N500JSONResponse{
				Message: err.Error(),
			}}, nil
		}
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
		switch classifyServiceError(err) {
		case http.StatusNotFound:
			return api.V1UserRequestPasswordReset404JSONResponse{N404JSONResponse: notFound}, nil
		default:
			return api.V1UserRequestPasswordReset500JSONResponse{N500JSONResponse: api.N500JSONResponse{
				Message: err.Error(),
			}}, nil
		}
	}

	token, err := c.userService.CreateToken(ctx, user.ID, user.Email, model.UserTokenContextResetPassword, nil)
	if err != nil {
		return api.V1UserRequestPasswordReset400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}

	if err := c.emailService.SendAuthPasswordResetEmail(ctx, email.Recipient{
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
	}, token); err != nil {
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
		switch classifyServiceError(err) {
		case http.StatusNotFound:
			return api.V1UserResetPassword404JSONResponse{N404JSONResponse: notFound}, nil
		default:
			return api.V1UserResetPassword500JSONResponse{N500JSONResponse: api.N500JSONResponse{
				Message: err.Error(),
			}}, nil
		}
	}

	if verifyErr != nil && errors.Is(verifyErr, service.ErrExpiredToken) {
		if err := c.userService.DeleteToken(ctx, userID, model.UserTokenContextResetPassword); err != nil {
			if classifyServiceError(err) != http.StatusNotFound {
				return api.V1UserResetPassword500JSONResponse{N500JSONResponse: api.N500JSONResponse{
					Message: err.Error(),
				}}, nil
			}
		}

		token, err := c.userService.CreateToken(ctx, user.ID, user.Email, model.UserTokenContextResetPassword, nil)
		if err != nil {
			return api.V1UserResetPassword400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
		}

		if err := c.emailService.SendAuthPasswordResetEmail(ctx, email.Recipient{
			Email:     user.Email,
			FirstName: user.FirstName,
			LastName:  user.LastName,
		}, token); err != nil {
			return api.V1UserResetPassword500JSONResponse{
				N500JSONResponse: api.N500JSONResponse{
					Message: err.Error(),
				},
			}, nil
		}

		return api.V1UserResetPassword204Response{}, nil
	}

	// Set the user ID in context for the update operation
	ctx = context.WithValue(ctx, pkg.CtxKeyUserID, user.ID)

	if _, err = c.userService.Update(ctx, user.ID, service.UpdateUserOpts{
		Password: optional.Some(auth.HashPassword(request.Body.Password)),
	}); err != nil {
		return api.V1UserResetPassword500JSONResponse{
			N500JSONResponse: api.N500JSONResponse{
				Message: err.Error(),
			},
		}, nil
	}

	if err := c.userService.DeleteToken(ctx, userID, model.UserTokenContextResetPassword); err != nil {
		if classifyServiceError(err) != http.StatusNotFound {
			return api.V1UserResetPassword500JSONResponse{N500JSONResponse: api.N500JSONResponse{
				Message: err.Error(),
			}}, nil
		}
	}

	return api.V1UserResetPassword200Response{}, nil
}

// NewUserController creates a new UserController.
func NewUserController(
	userService service.UserService,
	emailService service.EmailService,
	opts ...ControllerOption,
) (UserController, error) {
	c, err := newController(opts...)
	if err != nil {
		return nil, err
	}

	if userService == nil {
		return nil, ErrNoUserService
	}

	if emailService == nil {
		return nil, ErrNoEmailService
	}

	return &userController{
		baseController: c,
		userService:    userService,
		emailService:   emailService,
	}, nil
}

func createUserJSONRequestBodyToCreateUserOpts(body *api.V1UsersCreateJSONRequestBody) (service.CreateUserOpts, error) {
	if body.FirstName == "" {
		return service.CreateUserOpts{}, errors.New("FirstName is required")
	}
	if body.LastName == "" {
		return service.CreateUserOpts{}, errors.New("LastName is required")
	}

	opts := service.CreateUserOpts{
		Username:  body.Username,
		FirstName: body.FirstName,
		LastName:  body.LastName,
		Email:     string(body.Email),
		Password:  password.HashPassword(body.Password),
		Status:    model.UserStatusActive,
		Title:     pkg.DefaultPtr(body.Title, ""),
		Picture:   pkg.DefaultPtr(body.Picture, ""),
		Bio:       pkg.DefaultPtr(body.Bio, ""),
		Address:   pkg.DefaultPtr(body.Address, ""),
		Phone:     pkg.DefaultPtr(body.Phone, ""),
		Links:     pkg.DefaultPtr(body.Links, make([]string, 0)),
		Languages: make([]model.Language, 0),
	}

	if body.Languages != nil {
		opts.Languages = make([]model.Language, len(*body.Languages))
		for i, language := range *body.Languages {
			var lang model.Language
			if err := lang.UnmarshalText([]byte(language)); err != nil {
				return service.CreateUserOpts{}, err
			}
			opts.Languages[i] = lang
		}
	}

	return opts, nil
}

func updateUserJSONRequestBodyToUpdateUserOpts(body *api.V1UserUpdateJSONRequestBody) (service.UpdateUserOpts, error) {
	opts := service.UpdateUserOpts{}

	if body.Username != nil {
		opts.Username = optional.Some(*body.Username)
	}
	if body.Email != nil {
		opts.Email = optional.Some(string(*body.Email))
	}
	if body.FirstName != nil {
		opts.FirstName = optional.Some(*body.FirstName)
	}
	if body.LastName != nil {
		opts.LastName = optional.Some(*body.LastName)
	}
	if body.Address.Defined {
		opts.Address = body.Address
	}
	if body.Bio.Defined {
		opts.Bio = body.Bio
	}
	if body.Phone.Defined {
		opts.Phone = body.Phone
	}
	if body.Picture.Defined {
		opts.Picture = body.Picture
	}
	if body.Title.Defined {
		opts.Title = body.Title
	}
	if body.Links != nil {
		opts.Links = optional.Some(*body.Links)
	}
	if body.Status != nil {
		var status model.UserStatus
		if err := status.UnmarshalText([]byte(string(*body.Status))); err != nil {
			return service.UpdateUserOpts{}, err
		}
		opts.Status = optional.Some(status)
	}
	if body.Languages != nil {
		languages := make([]model.Language, len(*body.Languages))
		for i, language := range *body.Languages {
			var lang model.Language
			if err := lang.UnmarshalText([]byte(language)); err != nil {
				return service.UpdateUserOpts{}, err
			}
			languages[i] = lang
		}
		opts.Languages = optional.Some(languages)
	}

	return opts, nil
}

func userToDTO(user *service.User) api.User {
	u := api.User{
		Id:            user.ID.String(),
		Address:       &user.Address,
		Bio:           &user.Bio,
		Email:         oapiTypes.Email(user.Email),
		FirstName:     user.FirstName,
		LastName:      user.LastName,
		Links:         &user.Links,
		Username:      user.Username,
		Phone:         &user.Phone,
		Picture:       &user.Picture,
		Status:        api.UserStatus(user.Status.String()),
		Title:         &user.Title,
		DocumentCount: user.DocumentCount,
		Languages:     make([]api.Language, len(user.Languages)),
		CreatedAt:     *user.CreatedAt,
		UpdatedAt:     user.UpdatedAt,
	}

	for i, language := range user.Languages {
		u.Languages[i] = api.Language(language.String())
	}

	return u
}

package service

import (
	"context"
	"errors"
	"time"

	"github.com/opcotech/elemo/internal/license"
	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg"
	"github.com/opcotech/elemo/internal/pkg/auth"
	"github.com/opcotech/elemo/internal/pkg/optional"
	"github.com/opcotech/elemo/internal/pkg/password"
	"github.com/opcotech/elemo/internal/pkg/validate"
	"github.com/opcotech/elemo/internal/repository"
)

const (
	UserConfirmationDeadline  = 24 * time.Hour
	UserPasswordResetDeadline = 15 * time.Minute
	UserInvitationDeadline    = 7 * 24 * time.Hour
)

// User represents a user returned by the service.
type User struct {
	ID            model.ID
	Username      string
	Email         string
	Password      string
	Status        model.UserStatus
	FirstName     string
	LastName      string
	Picture       string
	Title         string
	Bio           string
	Phone         string
	Address       string
	Links         []string
	Languages     []model.Language
	DocumentCount *int64
	Permissions   []model.ID
	CreatedAt     *time.Time
	UpdatedAt     *time.Time
}

// PartialUser is a lean user used on issue and document reads.
type PartialUser struct {
	ID        model.ID
	FirstName string
	LastName  string
	Picture   string
}

// CreateUserOpts holds the data required to create a user.
type CreateUserOpts struct {
	Username  string           `json:"username" validate:"required,lowercase,min=3,max=50,containsany=0123456789abcdefghijklmnopqrstuvwxyz-_"`
	Email     string           `json:"email" validate:"required,email"`
	Password  string           `json:"password" validate:"required,min=8,max=64"`
	Status    model.UserStatus `json:"status" validate:"omitempty,min=1,max=4"`
	FirstName string           `json:"first_name" validate:"required,min=1,max=50"`
	LastName  string           `json:"last_name" validate:"required,min=1,max=50"`
	Picture   string           `json:"picture" validate:"omitempty,url"`
	Title     string           `json:"title" validate:"omitempty,min=3,max=50"`
	Bio       string           `json:"bio" validate:"omitempty,min=10,max=500"`
	Phone     string           `json:"phone" validate:"omitempty,min=7,max=16"`
	Address   string           `json:"address" validate:"omitempty,min=3,max=500"`
	Links     []string         `json:"links" validate:"omitempty,dive,url"`
	Languages []model.Language `json:"languages" validate:"omitempty,dive"`
}

// Validate validates the create options.
func (o *CreateUserOpts) Validate() error {
	if o.Status == 0 {
		o.Status = model.UserStatusActive
	}
	if err := validate.Struct(o); err != nil {
		return errors.Join(model.ErrInvalidUserDetails, err)
	}
	return nil
}

// UpdateUserOpts holds the fields that can be updated on a user.
// Undefined fields (Defined == false) are left unchanged.
type UpdateUserOpts struct {
	Username  optional.Optional[string]
	Email     optional.Optional[string]
	Password  optional.Optional[string]
	Status    optional.Optional[model.UserStatus]
	FirstName optional.Optional[string]
	LastName  optional.Optional[string]
	Picture   optional.Optional[string]
	Title     optional.Optional[string]
	Bio       optional.Optional[string]
	Phone     optional.Optional[string]
	Address   optional.Optional[string]
	Links     optional.Optional[[]string]
	Languages optional.Optional[[]model.Language]
}

// UserToken represents a user token returned by the service.
type UserToken struct {
	ID        model.ID
	UserID    model.ID
	SentTo    string
	Token     string
	Context   model.UserTokenContext
	CreatedAt *time.Time
}

// CreateUserTokenOpts holds the data required to create a user token.
type CreateUserTokenOpts struct {
	UserID  model.ID               `json:"user_id" validate:"required"`
	SentTo  string                 `json:"sent_to" validate:"required,email"`
	Token   string                 `json:"token" validate:"required,min=60,max=72"`
	Context model.UserTokenContext `json:"context" validate:"required,min=1,max=3"`
}

// Validate validates the create token options.
func (o *CreateUserTokenOpts) Validate() error {
	if err := validate.Struct(o); err != nil {
		return errors.Join(model.ErrInvalidUserToken, err)
	}
	if err := o.UserID.Validate(); err != nil {
		return errors.Join(model.ErrInvalidUserToken, err)
	}
	return nil
}

// UserService serves the business logic of interacting with users in the
// system.
//
//go:generate go tool mockgen -destination=mock/mock_user_gen.go -package=mocksvc . UserService
type UserService interface {
	// Create creates a new user in the system. The user's password is not
	// hashed before being stored in the database. Make sure to hash the
	// password before trying to create the user. If the user already exists,
	// an error is returned.
	Create(ctx context.Context, opts CreateUserOpts) (*User, error)
	// Get returns a user by its ID. If the user does not exist, an error is
	// returned.
	Get(ctx context.Context, id model.ID) (*User, error)
	// GetByEmail returns a user by their email address. If the user does not
	// exist, an error is returned.
	GetByEmail(ctx context.Context, email string) (*User, error)
	// List returns a cursor-paginated page of users in the system.
	List(ctx context.Context, page CursorPage) (Page[*User], error)
	// Update updates a user in the system. If the user does not exist, an
	// error is returned.
	Update(ctx context.Context, id model.ID, opts UpdateUserOpts) (*User, error)
	// Delete updates the user's status to delete and sets the password to
	// pkg.UnusablePassword. This method does not actually delete the user from
	// the database to preserve the user's history and relations unless the
	// force parameter is set to true.
	Delete(ctx context.Context, id model.ID, force bool) error
	// CreateToken creates a user token pair and saves the secret token in the
	// database. If saving the user token is successful, the public token is
	// returned. Any existing tokens are purged.
	CreateToken(ctx context.Context, id model.ID, sendTo string, tokenContext model.UserTokenContext, data map[string]any) (string, error)
	// VerifyToken checks the confirmation token and returns whether the token
	// is valid or not.
	VerifyToken(ctx context.Context, public string) (map[string]any, error)
	// DeleteToken removes a confirmation token, hence prevents
	// token reuse.
	DeleteToken(ctx context.Context, id model.ID, tokenContext model.UserTokenContext) error
}

// userService is the concrete implementation of the UserService interface.
type userService struct {
	runtime
	userRepo       repository.UserRepository
	userTokenRepo  repository.UserTokenRepository
	licenseService LicenseService
}

func userFromRepository(u *repository.User) *User {
	if u == nil {
		return nil
	}
	return &User{
		ID:            u.ID,
		Username:      u.Username,
		Email:         u.Email,
		Password:      u.Password,
		Status:        u.Status,
		FirstName:     u.FirstName,
		LastName:      u.LastName,
		Picture:       u.Picture,
		Title:         u.Title,
		Bio:           u.Bio,
		Phone:         u.Phone,
		Address:       u.Address,
		Links:         u.Links,
		Languages:     u.Languages,
		DocumentCount: u.DocumentCount,
		Permissions:   u.Permissions,
		CreatedAt:     u.CreatedAt,
		UpdatedAt:     u.UpdatedAt,
	}
}

func partialUserFromRepository(u *repository.PartialUser) *PartialUser {
	if u == nil {
		return nil
	}
	return &PartialUser{
		ID:        u.ID,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Picture:   u.Picture,
	}
}

func partialUserValueFromRepository(u repository.PartialUser) PartialUser {
	if mapped := partialUserFromRepository(&u); mapped != nil {
		return *mapped
	}
	return PartialUser{}
}

func (s *userService) Create(ctx context.Context, opts CreateUserOpts) (*User, error) {
	ctx, span := s.tracer.Start(ctx, "service.userService/Create")
	defer span.End()

	if expired, err := s.licenseService.Expired(ctx); expired || err != nil {
		return nil, errors.Join(ErrUserCreate, license.ErrLicenseExpired)
	}

	if err := opts.Validate(); err != nil {
		return nil, errors.Join(ErrUserCreate, err)
	}

	// If the newly created user is not active, e.g. a company is migrating
	// ex-employees, do not check the license quota as that only counts
	// against active users.
	if opts.Status == model.UserStatusActive {
		if ok, err := s.licenseService.WithinThreshold(ctx, license.QuotaUsers); !ok || err != nil {
			return nil, errors.Join(ErrUserCreate, ErrQuotaExceeded)
		}
	}

	user, err := s.userRepo.Create(ctx, repository.CreateUserOpts{
		Username:  opts.Username,
		Email:     opts.Email,
		Password:  opts.Password,
		Status:    opts.Status,
		FirstName: opts.FirstName,
		LastName:  opts.LastName,
		Picture:   opts.Picture,
		Title:     opts.Title,
		Bio:       opts.Bio,
		Phone:     opts.Phone,
		Address:   opts.Address,
		Links:     opts.Links,
		Languages: opts.Languages,
	})
	if err != nil {
		return nil, errors.Join(ErrUserCreate, err)
	}

	return userFromRepository(user), nil
}

func (s *userService) Get(ctx context.Context, id model.ID) (*User, error) {
	ctx, span := s.tracer.Start(ctx, "service.userService/Get")
	defer span.End()

	if err := id.Validate(); err != nil {
		return nil, errors.Join(ErrUserGet, err)
	}

	user, err := s.userRepo.Get(ctx, id, repository.UserDetailProjection())
	if err != nil {
		return nil, errors.Join(ErrUserGet, err)
	}

	return userFromRepository(user), nil
}

func (s *userService) GetByEmail(ctx context.Context, email string) (*User, error) {
	ctx, span := s.tracer.Start(ctx, "service.userService/GetByEmail")
	defer span.End()

	if email == "" {
		return nil, errors.Join(ErrUserGet, ErrInvalidEmail)
	}

	user, err := s.userRepo.GetByEmail(ctx, email, repository.UserDetailProjection())
	if err != nil {
		return nil, errors.Join(ErrUserGet, err)
	}

	return userFromRepository(user), nil
}

func (s *userService) List(ctx context.Context, page CursorPage) (Page[*User], error) {
	ctx, span := s.tracer.Start(ctx, "service.userService/List")
	defer span.End()

	normalized, err := page.Normalize()
	if err != nil {
		return Page[*User]{}, errors.Join(ErrUserList, err)
	}

	if _, err := ctxUserID(ctx); err != nil {
		return Page[*User]{}, errors.Join(ErrUserList, err)
	}

	users, err := s.userRepo.List(ctx, normalized, repository.UserListProjection())
	if err != nil {
		return Page[*User]{}, errors.Join(ErrUserList, err)
	}

	return mapPage(users, userFromRepository), nil
}

func (s *userService) Update(ctx context.Context, id model.ID, opts UpdateUserOpts) (*User, error) {
	ctx, span := s.tracer.Start(ctx, "service.userService/Update")
	defer span.End()

	if expired, err := s.licenseService.Expired(ctx); expired || err != nil {
		return nil, errors.Join(ErrUserUpdate, license.ErrLicenseExpired)
	}

	if err := id.Validate(); err != nil {
		return nil, errors.Join(ErrUserUpdate, err)
	}

	userID, err := ctxUserID(ctx)
	if err != nil {
		return nil, errors.Join(ErrUserUpdate, err)
	}

	if userID != id {
		return nil, errors.Join(ErrUserUpdate, ErrNoPermission)
	}

	// Check if the user is being activated is within the license quota. It
	// could be a possible loophole to activate a previously deleted user to
	// bypass the quota check.
	if opts.Status.Defined && opts.Status.Value != nil && *opts.Status.Value == model.UserStatusActive {
		if ok, err := s.licenseService.WithinThreshold(ctx, license.QuotaUsers); !ok || err != nil {
			return nil, errors.Join(ErrUserUpdate, ErrQuotaExceeded)
		}
	}

	user, err := s.userRepo.Update(ctx, id, repository.UpdateUserOpts{
		Username:  opts.Username,
		Email:     opts.Email,
		Password:  opts.Password,
		Status:    opts.Status,
		FirstName: opts.FirstName,
		LastName:  opts.LastName,
		Picture:   opts.Picture,
		Title:     opts.Title,
		Bio:       opts.Bio,
		Phone:     opts.Phone,
		Address:   opts.Address,
		Links:     opts.Links,
		Languages: opts.Languages,
	})
	if err != nil {
		return nil, errors.Join(ErrUserUpdate, err)
	}

	return userFromRepository(user), nil
}

func (s *userService) Delete(ctx context.Context, id model.ID, force bool) error {
	ctx, span := s.tracer.Start(ctx, "service.userService/Delete")
	defer span.End()

	if expired, err := s.licenseService.Expired(ctx); expired || err != nil {
		return errors.Join(ErrUserUpdate, license.ErrLicenseExpired)
	}

	if err := id.Validate(); err != nil {
		return errors.Join(ErrUserDelete, err)
	}

	userID, err := ctxUserID(ctx)
	if err != nil {
		return errors.Join(ErrUserDelete, err)
	}

	if userID != id {
		return errors.Join(ErrUserDelete, ErrNoPermission)
	}

	if force {
		if err := s.userRepo.Delete(ctx, id); err != nil {
			return errors.Join(ErrUserDelete, err)
		}
	} else {
		if _, err := s.userRepo.Update(ctx, id, repository.UpdateUserOpts{
			Status:   optional.Some(model.UserStatusDeleted),
			Password: optional.Some(password.UnusablePassword),
		}); err != nil {
			return errors.Join(ErrUserDelete, err)
		}
	}

	return nil
}

func (s *userService) CreateToken(ctx context.Context, id model.ID, sendTo string, tokenContext model.UserTokenContext, data map[string]any) (string, error) {
	ctx, span := s.tracer.Start(ctx, "service.userService/CreateToken")
	defer span.End()

	if id.IsNil() {
		return "", errors.Join(ErrUserCreateUserToken, model.ErrInvalidID)
	}

	existingToken, err := s.userTokenRepo.Get(ctx, id, tokenContext)
	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		return "", errors.Join(ErrUserCreateUserToken, err)
	}

	if existingToken != nil {
		if err := s.userTokenRepo.Delete(ctx, existingToken.UserID, existingToken.Context); err != nil {
			return "", errors.Join(ErrUserCreateUserToken, err)
		}
	}

	tokenData := pkg.MergeMaps(data, map[string]any{"user_id": id.String()})
	public, secret, err := auth.GenerateToken(tokenContext.String(), tokenData)
	if err != nil {
		return "", errors.Join(ErrUserCreateUserToken, err)
	}

	createOpts := CreateUserTokenOpts{
		UserID:  id,
		SentTo:  sendTo,
		Token:   secret,
		Context: tokenContext,
	}
	if err := createOpts.Validate(); err != nil {
		return "", errors.Join(ErrUserCreateUserToken, err)
	}

	if _, err := s.userTokenRepo.Create(ctx, repository.CreateUserTokenOpts{
		UserID:  createOpts.UserID,
		SentTo:  createOpts.SentTo,
		Token:   createOpts.Token,
		Context: createOpts.Context,
	}); err != nil {
		return "", errors.Join(ErrUserCreateUserToken, err)
	}

	return public, nil
}

func (s *userService) VerifyToken(ctx context.Context, public string) (map[string]any, error) {
	ctx, span := s.tracer.Start(ctx, "service.userService/VerifyToken")
	defer span.End()

	kind, _, tokenData := auth.SplitToken(public)

	userID, err := model.NewIDFromString(tokenData["user_id"].(string), model.ResourceTypeUser.String())
	if err != nil {
		return nil, errors.Join(ErrUserVerifyToken, ErrInvalidToken)
	}

	var tokenContext model.UserTokenContext
	if err := tokenContext.UnmarshalText([]byte(kind)); err != nil {
		return nil, errors.Join(ErrUserVerifyToken, ErrInvalidToken)
	}

	confirmation, err := s.userTokenRepo.Get(ctx, userID, tokenContext)
	if err != nil {
		return nil, errors.Join(ErrUserVerifyToken, err)
	}

	if !auth.IsTokenMatching(confirmation.Token, public) {
		return nil, errors.Join(ErrUserVerifyToken, ErrInvalidToken)
	}

	var deadline time.Duration
	switch kind {
	case model.UserTokenContextConfirm.String():
		deadline = UserConfirmationDeadline
	case model.UserTokenContextResetPassword.String():
		deadline = UserPasswordResetDeadline
	case model.UserTokenContextInvite.String():
		deadline = UserInvitationDeadline
	default:
		return nil, errors.Join(ErrUserVerifyToken, ErrInvalidToken)
	}

	if time.Now().After(confirmation.CreatedAt.Add(deadline)) {
		return nil, errors.Join(ErrUserVerifyToken, ErrExpiredToken)
	}

	return tokenData, nil
}

func (s *userService) DeleteToken(ctx context.Context, id model.ID, tokenContext model.UserTokenContext) error {
	ctx, span := s.tracer.Start(ctx, "service.userService/DeleteConfirmationToken")
	defer span.End()

	if id.IsNil() {
		return errors.Join(ErrUserDeleteUserToken, model.ErrInvalidID)
	}

	if err := s.userTokenRepo.Delete(ctx, id, tokenContext); err != nil {
		return errors.Join(ErrUserDeleteUserToken, err)
	}

	return nil
}

// NewUserService returns a new instance of the UserService interface.
func NewUserService(
	userRepo repository.UserRepository,
	userTokenRepo repository.UserTokenRepository,
	licenseService LicenseService,
	opts ...Option,
) (UserService, error) {
	rt, err := newRuntime(opts...)
	if err != nil {
		return nil, err
	}

	svc := &userService{
		runtime:        rt,
		userRepo:       userRepo,
		userTokenRepo:  userTokenRepo,
		licenseService: licenseService,
	}

	if svc.userRepo == nil {
		return nil, ErrNoUserRepository
	}

	if svc.userTokenRepo == nil {
		return nil, ErrNoUserTokenRepository
	}

	if svc.licenseService == nil {
		return nil, ErrNoLicenseService
	}

	return svc, nil
}

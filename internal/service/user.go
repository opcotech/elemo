package service

import (
	"context"
	"errors"
	"time"

	"github.com/opcotech/elemo/internal/license"
	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg"
	"github.com/opcotech/elemo/internal/pkg/auth"
	"github.com/opcotech/elemo/internal/pkg/password"
	"github.com/opcotech/elemo/internal/repository"
)

const (
	UserConfirmationDeadline  = 24 * time.Hour
	UserPasswordResetDeadline = 15 * time.Minute
	UserInvitationDeadline    = 7 * 24 * time.Hour
)

// UserService serves the business logic of interacting with users in the
// system.
type UserService interface {
	// Create creates a new user in the system. The user's password is not
	// hashed before being stored in the database. Make sure to hash the
	// password before trying to create the user. If the user already exists,
	// an error is returned.
	Create(ctx context.Context, user *model.User) error
	// Get returns a user by its ID. If the user does not exist, an error is
	// returned.
	Get(ctx context.Context, id model.ID) (*model.User, error)
	// GetByEmail returns a user by their email address. If the user does not
	// exist, an error is returned.
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	// GetAll returns all users in the system. The offset and limit parameters
	// are used to paginate the results. If the offset is greater than the
	// number of users in the system, an empty slice is returned.
	GetAll(ctx context.Context, offset, limit int) ([]*model.User, error)
	// Update updates a user in the system. If the user does not exist, an
	// error is returned.
	Update(ctx context.Context, id model.ID, patch map[string]any) (*model.User, error)
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
	*baseService
}

func (s *userService) Create(ctx context.Context, user *model.User) error {
	ctx, span := s.tracer.Start(ctx, "service.userService/Create")
	defer span.End()

	if expired, err := s.licenseService.Expired(ctx); expired || err != nil {
		return errors.Join(ErrUserCreate, license.ErrLicenseExpired)
	}

	if err := user.Validate(); err != nil {
		return errors.Join(ErrUserCreate, err)
	}

	if !s.permissionService.CtxUserHasPermission(ctx, model.MustNewNilID(model.ResourceTypeUser), model.PermissionKindCreate) {
		return errors.Join(ErrUserCreate, ErrNoPermission)
	}

	// If the newly created user is not active, e.g. a company is migrating
	// ex-employees, do not check the license quota as that only counts
	// against active users.
	if user.Status == model.UserStatusActive {
		if ok, err := s.licenseService.WithinThreshold(ctx, license.QuotaUsers); !ok || err != nil {
			return errors.Join(ErrUserCreate, ErrQuotaExceeded)
		}
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return errors.Join(ErrUserCreate, err)
	}

	return nil
}

func (s *userService) Get(ctx context.Context, id model.ID) (*model.User, error) {
	ctx, span := s.tracer.Start(ctx, "service.userService/Get")
	defer span.End()

	if err := id.Validate(); err != nil {
		return nil, errors.Join(ErrUserGet, err)
	}

	user, err := s.userRepo.Get(ctx, id)
	if err != nil {
		return nil, errors.Join(ErrUserGet, err)
	}

	return user, nil
}

func (s *userService) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	ctx, span := s.tracer.Start(ctx, "service.userService/GetByEmail")
	defer span.End()

	if email == "" {
		return nil, errors.Join(ErrUserGet, ErrInvalidEmail)
	}

	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, errors.Join(ErrUserGet, err)
	}

	return user, nil
}

func (s *userService) GetAll(ctx context.Context, offset, limit int) ([]*model.User, error) {
	ctx, span := s.tracer.Start(ctx, "service.userService/GetAll")
	defer span.End()

	if offset < 0 || limit <= 0 {
		return nil, errors.Join(ErrUserGetAll, ErrInvalidPaginationParams)
	}

	users, err := s.userRepo.GetAll(ctx, offset, limit)
	if err != nil {
		return nil, errors.Join(ErrUserGetAll, err)
	}

	return users, nil
}

func (s *userService) Update(ctx context.Context, id model.ID, patch map[string]any) (*model.User, error) {
	ctx, span := s.tracer.Start(ctx, "service.userService/Update")
	defer span.End()

	if expired, err := s.licenseService.Expired(ctx); expired || err != nil {
		return nil, errors.Join(ErrUserUpdate, license.ErrLicenseExpired)
	}

	if err := id.Validate(); err != nil {
		return nil, errors.Join(ErrUserUpdate, err)
	}

	userID, ok := ctx.Value(pkg.CtxKeyUserID).(model.ID)
	if !ok {
		return nil, errors.Join(ErrUserUpdate, ErrNoUser)
	}

	if userID != id && !s.permissionService.CtxUserHasPermission(ctx, id, model.PermissionKindWrite) {
		return nil, errors.Join(ErrUserUpdate, ErrNoPermission)
	}

	// Check if the user is being activated is within the license quota. It
	// could be a possible loophole to activate a previously deleted user to
	// bypass the quota check.
	if patchStatus, ok := patch["status"]; ok && patchStatus == model.UserStatusActive.String() {
		if ok, err := s.licenseService.WithinThreshold(ctx, license.QuotaUsers); !ok || err != nil {
			return nil, errors.Join(ErrUserUpdate, ErrQuotaExceeded)
		}
	}

	user, err := s.userRepo.Update(ctx, id, patch)
	if err != nil {
		return nil, errors.Join(ErrUserUpdate, err)
	}

	return user, nil
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

	userID, ok := ctx.Value(pkg.CtxKeyUserID).(model.ID)
	if !ok {
		return errors.Join(ErrUserUpdate, ErrNoUser)
	}

	if userID == id || !s.permissionService.CtxUserHasPermission(ctx, id, model.PermissionKindDelete) {
		return errors.Join(ErrUserUpdate, ErrNoPermission)
	}

	if force {
		if err := s.userRepo.Delete(ctx, id); err != nil {
			return errors.Join(ErrUserDelete, err)
		}
	} else {
		patch := map[string]any{
			"status":   model.UserStatusDeleted.String(),
			"password": password.UnusablePassword,
		}

		if _, err := s.userRepo.Update(ctx, id, patch); err != nil {
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

	newToken, err := model.NewUserToken(id, sendTo, secret, tokenContext)
	if err != nil {
		return "", errors.Join(ErrUserCreateUserToken, err)
	}

	if err := s.userTokenRepo.Create(ctx, newToken); err != nil {
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
func NewUserService(opts ...Option) (UserService, error) {
	s, err := newService(opts...)
	if err != nil {
		return nil, err
	}

	svc := &userService{
		baseService: s,
	}

	if svc.userRepo == nil {
		return nil, ErrNoUserRepository
	}

	if svc.userTokenRepo == nil {
		return nil, ErrNoUserTokenRepository
	}

	if svc.permissionService == nil {
		return nil, ErrNoPermissionService
	}

	if svc.licenseService == nil {
		return nil, ErrNoLicenseService
	}

	return svc, nil
}

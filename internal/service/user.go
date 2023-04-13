package service

import (
	"context"
	"errors"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg"
	"github.com/opcotech/elemo/internal/pkg/password"
)

// UserRepository defines the interface for interacting with the user
// repository.
type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	Get(ctx context.Context, id model.ID) (*model.User, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	GetAll(ctx context.Context, offset, limit int) ([]*model.User, error)
	Update(ctx context.Context, id model.ID, patch map[string]any) (*model.User, error)
	Delete(ctx context.Context, id model.ID) error
}

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
}

// userService is the concrete implementation of the UserService interface.
type userService struct {
	*baseService
}

func (s *userService) Create(ctx context.Context, user *model.User) error {
	ctx, span := s.tracer.Start(ctx, "service.userService/Create")
	defer span.End()

	if err := user.Validate(); err != nil {
		return errors.Join(ErrUserCreate, err)
	}

	if !ctxUserPermitted(ctx, s.permissionRepo, model.MustNewNilID(model.ResourceTypeUser), model.PermissionKindCreate) {
		return ErrNoPermission
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

	if err := id.Validate(); err != nil {
		return nil, errors.Join(ErrUserUpdate, err)
	}

	userID, ok := ctx.Value(pkg.CtxKeyUserID).(model.ID)
	if !ok {
		return nil, ErrNoUser
	}

	if userID != id && !ctxUserPermitted(ctx, s.permissionRepo, id, model.PermissionKindWrite) {
		return nil, ErrNoPermission
	}

	if len(patch) == 0 {
		return nil, errors.Join(ErrUserUpdate, ErrNoPatchData)
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

	if err := id.Validate(); err != nil {
		return errors.Join(ErrUserDelete, err)
	}

	userID, ok := ctx.Value(pkg.CtxKeyUserID).(model.ID)
	if !ok {
		return ErrNoUser
	}

	if userID == id || !ctxUserPermitted(ctx, s.permissionRepo, id, model.PermissionKindDelete) {
		return ErrNoPermission
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

	if svc.permissionRepo == nil {
		return nil, ErrNoPermissionRepository
	}

	if svc.licenseService == nil {
		return nil, ErrNoLicenseService
	}

	return svc, nil
}

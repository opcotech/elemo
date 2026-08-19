package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg"
	"github.com/opcotech/elemo/internal/pkg/log"
	"github.com/opcotech/elemo/internal/pkg/optional"
	"github.com/opcotech/elemo/internal/repository"
	"github.com/opcotech/elemo/internal/testutil/mock"
)

func newRepositoryFolder(libraryID, createdBy model.ID) *repository.Folder {
	return &repository.Folder{
		ID:   model.MustNewID(model.ResourceTypeFolder),
		Name: "Guides",
		Library: repository.DocumentLibrary{
			ID:   libraryID,
			Type: libraryID.Type,
			Name: "Library",
		},
		CreatedBy: repository.PartialUser{ID: createdBy},
		CreatedAt: timePtr(time.Now().UTC()),
	}
}

func timePtr(t time.Time) *time.Time {
	return &t
}

func TestNewFolderService(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		svc, err := NewFolderService(
			WithLogger(mock.NewMockLogger(ctrl)),
			WithTracer(mock.NewMockTracer(ctrl)),
			WithFolderRepository(repository.NewMockFolderRepository(ctrl)),
			WithPermissionService(NewMockPermissionService(ctrl)),
		)
		require.NoError(t, err)
		assert.NotNil(t, svc)
	})

	t.Run("no folder repository", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		_, err := NewFolderService(
			WithLogger(mock.NewMockLogger(ctrl)),
			WithTracer(mock.NewMockTracer(ctrl)),
			WithPermissionService(NewMockPermissionService(ctrl)),
		)
		assert.ErrorIs(t, err, ErrNoFolderRepository)
	})

	t.Run("invalid logger", func(t *testing.T) {
		t.Parallel()
		_, err := NewFolderService(WithLogger(nil))
		assert.ErrorIs(t, err, log.ErrNoLogger)
	})
}

func TestFolderService_Create(t *testing.T) {
	t.Parallel()

	libraryID := model.MustNewID(model.ResourceTypeNamespace)
	userID := model.MustNewID(model.ResourceTypeUser)
	opts := CreateFolderOpts{Name: "Guides"}
	repoFolder := newRepositoryFolder(libraryID, userID)

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ctx := context.WithValue(context.Background(), pkg.CtxKeyUserID, userID)
		span := mock.NewMockSpan(ctrl)
		span.EXPECT().End(gomock.Len(0))
		tracer := mock.NewMockTracer(ctrl)
		tracer.EXPECT().Start(ctx, "service.folderService/Create", gomock.Len(0)).Return(ctx, span)

		permSvc := NewMockPermissionService(ctrl)
		permSvc.EXPECT().BootstrapCreator(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		permSvc.EXPECT().CtxUserHas(ctx, libraryID, gomock.Any()).Return(true)

		folderRepo := repository.NewMockFolderRepository(ctrl)
		folderRepo.EXPECT().Create(ctx, repository.CreateFolderOpts{
			Library:   libraryID,
			Name:      opts.Name,
			CreatedBy: userID,
		}).Return(repoFolder, nil)

		s := &folderService{baseService: &baseService{
			logger:            mock.NewMockLogger(ctrl),
			tracer:            tracer,
			folderRepo:        folderRepo,
			permissionService: permSvc,
		}}
		got, err := s.Create(ctx, libraryID, opts)
		require.NoError(t, err)
		assert.Equal(t, repoFolder.ID, got.ID)
		assert.Equal(t, opts.Name, got.Name)
		assert.Equal(t, libraryID, got.Library.ID)
	})

	t.Run("no permission", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ctx := context.WithValue(context.Background(), pkg.CtxKeyUserID, userID)
		span := mock.NewMockSpan(ctrl)
		span.EXPECT().End(gomock.Len(0))
		tracer := mock.NewMockTracer(ctrl)
		tracer.EXPECT().Start(ctx, "service.folderService/Create", gomock.Len(0)).Return(ctx, span)

		permSvc := NewMockPermissionService(ctrl)
		permSvc.EXPECT().BootstrapCreator(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		permSvc.EXPECT().CtxUserHas(ctx, libraryID, gomock.Any()).Return(false)

		s := &folderService{baseService: &baseService{
			logger:            mock.NewMockLogger(ctrl),
			tracer:            tracer,
			permissionService: permSvc,
		}}
		_, err := s.Create(ctx, libraryID, opts)
		assert.ErrorIs(t, err, ErrNoPermission)
	})

	t.Run("invalid library type", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ctx := context.Background()
		span := mock.NewMockSpan(ctrl)
		span.EXPECT().End(gomock.Len(0))
		tracer := mock.NewMockTracer(ctrl)
		tracer.EXPECT().Start(ctx, "service.folderService/Create", gomock.Len(0)).Return(ctx, span)

		s := &folderService{baseService: &baseService{
			logger: mock.NewMockLogger(ctrl),
			tracer: tracer,
		}}
		_, err := s.Create(ctx, model.MustNewID(model.ResourceTypeProject), opts)
		assert.ErrorIs(t, err, model.ErrInvalidID)
	})
}

func TestFolderService_Get(t *testing.T) {
	t.Parallel()

	libraryID := model.MustNewID(model.ResourceTypeOrganization)
	userID := model.MustNewID(model.ResourceTypeUser)
	repoFolder := newRepositoryFolder(libraryID, userID)

	t.Run("success checks folder permission", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ctx := context.Background()
		span := mock.NewMockSpan(ctrl)
		span.EXPECT().End(gomock.Len(0))
		tracer := mock.NewMockTracer(ctrl)
		tracer.EXPECT().Start(ctx, "service.folderService/Get", gomock.Len(0)).Return(ctx, span)

		folderRepo := repository.NewMockFolderRepository(ctrl)
		folderRepo.EXPECT().Get(ctx, repoFolder.ID).Return(repoFolder, nil)

		permSvc := NewMockPermissionService(ctrl)
		permSvc.EXPECT().BootstrapCreator(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		permSvc.EXPECT().CtxUserHas(ctx, repoFolder.ID, gomock.Any()).Return(true)

		s := &folderService{baseService: &baseService{
			logger:            mock.NewMockLogger(ctrl),
			tracer:            tracer,
			folderRepo:        folderRepo,
			permissionService: permSvc,
		}}
		got, err := s.Get(ctx, repoFolder.ID)
		require.NoError(t, err)
		assert.Equal(t, repoFolder.ID, got.ID)
	})
}

func TestFolderService_List(t *testing.T) {
	t.Parallel()

	libraryID := model.MustNewID(model.ResourceTypeNamespace)
	userID := model.MustNewID(model.ResourceTypeUser)
	repoFolder := newRepositoryFolder(libraryID, userID)
	page := CursorPage{Size: 10}

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ctx := context.WithValue(context.Background(), pkg.CtxKeyUserID, userID)
		span := mock.NewMockSpan(ctrl)
		span.EXPECT().End(gomock.Len(0))
		tracer := mock.NewMockTracer(ctrl)
		tracer.EXPECT().Start(ctx, "service.folderService/List", gomock.Len(0)).Return(ctx, span)

		permSvc := NewMockPermissionService(ctrl)
		permSvc.EXPECT().BootstrapCreator(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

		folderRepo := repository.NewMockFolderRepository(ctrl)
		folderRepo.EXPECT().List(ctx, libraryID, (*model.ID)(nil), userID, gomock.Any()).Return(repository.Page[*repository.Folder]{
			Items: []*repository.Folder{repoFolder},
		}, nil)

		s := &folderService{baseService: &baseService{
			logger:            mock.NewMockLogger(ctrl),
			tracer:            tracer,
			folderRepo:        folderRepo,
			permissionService: permSvc,
		}}
		got, err := s.List(ctx, libraryID, nil, page)
		require.NoError(t, err)
		require.Len(t, got.Items, 1)
		assert.Equal(t, repoFolder.ID, got.Items[0].ID)
	})

	t.Run("rejects non-library", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ctx := context.Background()
		span := mock.NewMockSpan(ctrl)
		span.EXPECT().End(gomock.Len(0))
		tracer := mock.NewMockTracer(ctrl)
		tracer.EXPECT().Start(ctx, "service.folderService/List", gomock.Len(0)).Return(ctx, span)

		s := &folderService{baseService: &baseService{
			logger: mock.NewMockLogger(ctrl),
			tracer: tracer,
		}}
		_, err := s.List(ctx, model.MustNewID(model.ResourceTypeProject), nil, page)
		assert.ErrorIs(t, err, model.ErrInvalidID)
	})
}

func TestFolderService_Update(t *testing.T) {
	t.Parallel()

	libraryID := model.MustNewID(model.ResourceTypeNamespace)
	userID := model.MustNewID(model.ResourceTypeUser)
	repoFolder := newRepositoryFolder(libraryID, userID)
	updated := *repoFolder
	updated.Name = "Architecture"

	t.Run("success checks library permission", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ctx := context.Background()
		span := mock.NewMockSpan(ctrl)
		span.EXPECT().End(gomock.Len(0))
		tracer := mock.NewMockTracer(ctrl)
		tracer.EXPECT().Start(ctx, "service.folderService/Update", gomock.Len(0)).Return(ctx, span)

		folderRepo := repository.NewMockFolderRepository(ctrl)
		folderRepo.EXPECT().Get(ctx, repoFolder.ID).Return(repoFolder, nil)
		folderRepo.EXPECT().Update(ctx, repoFolder.ID, repository.UpdateFolderOpts{
			Name: optional.Some("Architecture"),
		}).Return(&updated, nil)

		permSvc := NewMockPermissionService(ctrl)
		permSvc.EXPECT().BootstrapCreator(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		permSvc.EXPECT().CtxUserHas(ctx, libraryID, gomock.Any()).Return(true)

		s := &folderService{baseService: &baseService{
			logger:            mock.NewMockLogger(ctrl),
			tracer:            tracer,
			folderRepo:        folderRepo,
			permissionService: permSvc,
		}}
		got, err := s.Update(ctx, repoFolder.ID, UpdateFolderOpts{Name: optional.Some("Architecture")})
		require.NoError(t, err)
		assert.Equal(t, "Architecture", got.Name)
	})
}

func TestFolderService_Delete(t *testing.T) {
	t.Parallel()

	libraryID := model.MustNewID(model.ResourceTypeNamespace)
	userID := model.MustNewID(model.ResourceTypeUser)
	repoFolder := newRepositoryFolder(libraryID, userID)

	t.Run("success checks library permission", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ctx := context.Background()
		span := mock.NewMockSpan(ctrl)
		span.EXPECT().End(gomock.Len(0))
		tracer := mock.NewMockTracer(ctrl)
		tracer.EXPECT().Start(ctx, "service.folderService/Delete", gomock.Len(0)).Return(ctx, span)

		folderRepo := repository.NewMockFolderRepository(ctrl)
		folderRepo.EXPECT().Get(ctx, repoFolder.ID).Return(repoFolder, nil)
		folderRepo.EXPECT().Delete(ctx, repoFolder.ID).Return(nil)

		permSvc := NewMockPermissionService(ctrl)
		permSvc.EXPECT().BootstrapCreator(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		permSvc.EXPECT().CtxUserHas(ctx, libraryID, gomock.Any()).Return(true)

		s := &folderService{baseService: &baseService{
			logger:            mock.NewMockLogger(ctrl),
			tracer:            tracer,
			folderRepo:        folderRepo,
			permissionService: permSvc,
		}}
		require.NoError(t, s.Delete(ctx, repoFolder.ID))
	})
}

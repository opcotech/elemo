package service_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg"
	"github.com/opcotech/elemo/internal/pkg/log"
	mocklog "github.com/opcotech/elemo/internal/pkg/log/mock"
	"github.com/opcotech/elemo/internal/pkg/optional"
	"github.com/opcotech/elemo/internal/pkg/tracing"
	mocktrace "github.com/opcotech/elemo/internal/pkg/tracing/mock"
	"github.com/opcotech/elemo/internal/repository"
	mockrepo "github.com/opcotech/elemo/internal/repository/mock"
	"github.com/opcotech/elemo/internal/service"
	mocksvc "github.com/opcotech/elemo/internal/service/mock"
	testModel "github.com/opcotech/elemo/internal/testutil/model"
)

type customFieldServiceDeps struct {
	logger            log.Logger
	tracer            tracing.Tracer
	repo              repository.CustomFieldRepository
	permissionService service.PermissionService
}

func newCustomFieldServiceForTest(deps customFieldServiceDeps) service.CustomFieldService {
	if deps.repo == nil {
		deps.repo = mockrepo.NewMockCustomFieldRepository(nil)
	}
	if deps.permissionService == nil {
		deps.permissionService = mocksvc.NewMockPermissionService(nil)
	}
	var opts []service.Option
	if deps.logger != nil {
		opts = append(opts, service.WithLogger(deps.logger))
	}
	if deps.tracer != nil {
		opts = append(opts, service.WithTracer(deps.tracer))
	}
	svc, err := service.NewCustomFieldService(
		deps.repo,
		deps.permissionService,
		opts...,
	)
	if err != nil {
		panic(err)
	}
	return svc
}

func TestNewCustomFieldService(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		svc, err := service.NewCustomFieldService(
			mockrepo.NewMockCustomFieldRepository(ctrl),
			mocksvc.NewMockPermissionService(ctrl),
			service.WithLogger(mocklog.NewMockLogger(ctrl)),
			service.WithTracer(mocktrace.NewMockTracer(ctrl)),
		)
		require.NoError(t, err)
		assert.NotNil(t, svc)
	})

	t.Run("no repository", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		_, err := service.NewCustomFieldService(
			nil,
			mocksvc.NewMockPermissionService(ctrl),
			service.WithLogger(mocklog.NewMockLogger(ctrl)),
			service.WithTracer(mocktrace.NewMockTracer(ctrl)),
		)
		assert.ErrorIs(t, err, service.ErrNoCustomFieldRepository)
	})

	t.Run("invalid logger", func(t *testing.T) {
		t.Parallel()
		_, err := service.NewCustomFieldService(
			mockrepo.NewMockCustomFieldRepository(nil),
			mocksvc.NewMockPermissionService(nil),
			service.WithLogger(nil),
		)
		assert.ErrorIs(t, err, log.ErrNoLogger)
	})

	t.Run("no permission service", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		_, err := service.NewCustomFieldService(
			mockrepo.NewMockCustomFieldRepository(ctrl),
			nil,
			service.WithLogger(mocklog.NewMockLogger(ctrl)),
			service.WithTracer(mocktrace.NewMockTracer(ctrl)),
		)
		assert.ErrorIs(t, err, service.ErrNoPermissionService)
	})
}

func TestCustomFieldService_CreateDefinition(t *testing.T) {
	t.Parallel()

	userID := model.MustNewID(model.ResourceTypeUser)
	projectID := model.MustNewID(model.ResourceTypeProject)
	opts := service.CreateCustomFieldOpts{
		Key:        "story_points",
		Name:       "Story points",
		Kind:       model.CustomFieldKindInteger,
		Scope:      projectID,
		TargetType: model.ResourceTypeIssue,
		Schema:     model.CustomFieldSchema{Integer: &model.CustomFieldIntegerSchema{}},
	}

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ctx := context.WithValue(context.Background(), pkg.CtxKeyUserID, userID)
		span := mocktrace.NewMockSpan(ctrl)
		span.EXPECT().End(gomock.Len(0))
		tracer := mocktrace.NewMockTracer(ctrl)
		tracer.EXPECT().Start(gomock.Any(), "service.customFieldService/CreateDefinition", gomock.Len(0)).Return(ctx, span)

		permSvc := mocksvc.NewMockPermissionService(ctrl)
		permSvc.EXPECT().CtxUserHas(gomock.Any(), projectID, model.ActionCustomFieldManage).Return(true, nil)
		permSvc.EXPECT().ListScopeAncestry(gomock.Any(), projectID).Return([]model.ID{projectID}, nil)

		created := testModel.NewIntegerCustomFieldDefinition(projectID, userID)
		repo := mockrepo.NewMockCustomFieldRepository(ctrl)
		repo.EXPECT().NextSortOrder(gomock.Any(), projectID, model.ResourceTypeIssue).Return(4, nil)
		repo.EXPECT().CreateDefinition(gomock.Any(), gomock.Any()).DoAndReturn(
			func(_ context.Context, def *model.CustomFieldDefinition) (*model.CustomFieldDefinition, error) {
				assert.Equal(t, 4, def.SortOrder)
				return created, nil
			},
		)

		s := newCustomFieldServiceForTest(customFieldServiceDeps{
			logger:            mocklog.NewMockLogger(ctrl),
			tracer:            tracer,
			repo:              repo,
			permissionService: permSvc,
		})
		got, err := s.CreateDefinition(ctx, opts)
		require.NoError(t, err)
		assert.Equal(t, created.ID, got.ID)
	})

	t.Run("explicit order", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ctx := context.WithValue(context.Background(), pkg.CtxKeyUserID, userID)
		span := mocktrace.NewMockSpan(ctrl)
		span.EXPECT().End(gomock.Len(0))
		tracer := mocktrace.NewMockTracer(ctrl)
		tracer.EXPECT().Start(gomock.Any(), "service.customFieldService/CreateDefinition", gomock.Len(0)).Return(ctx, span)

		permSvc := mocksvc.NewMockPermissionService(ctrl)
		permSvc.EXPECT().CtxUserHas(gomock.Any(), projectID, model.ActionCustomFieldManage).Return(true, nil)
		permSvc.EXPECT().ListScopeAncestry(gomock.Any(), projectID).Return([]model.ID{projectID}, nil)

		created := testModel.NewIntegerCustomFieldDefinition(projectID, userID)
		repo := mockrepo.NewMockCustomFieldRepository(ctrl)
		repo.EXPECT().CreateDefinition(gomock.Any(), gomock.Any()).DoAndReturn(
			func(_ context.Context, def *model.CustomFieldDefinition) (*model.CustomFieldDefinition, error) {
				assert.Equal(t, 0, def.SortOrder)
				return created, nil
			},
		)

		s := newCustomFieldServiceForTest(customFieldServiceDeps{
			logger:            mocklog.NewMockLogger(ctrl),
			tracer:            tracer,
			repo:              repo,
			permissionService: permSvc,
		})
		createOpts := opts
		createOpts.SortOrder = optional.Some(0)
		got, err := s.CreateDefinition(ctx, createOpts)
		require.NoError(t, err)
		assert.Equal(t, created.ID, got.ID)
	})

	t.Run("no permission", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ctx := context.WithValue(context.Background(), pkg.CtxKeyUserID, userID)
		span := mocktrace.NewMockSpan(ctrl)
		span.EXPECT().End(gomock.Len(0))
		tracer := mocktrace.NewMockTracer(ctrl)
		tracer.EXPECT().Start(gomock.Any(), "service.customFieldService/CreateDefinition", gomock.Len(0)).Return(ctx, span)

		permSvc := mocksvc.NewMockPermissionService(ctrl)
		permSvc.EXPECT().CtxUserHas(gomock.Any(), projectID, model.ActionCustomFieldManage).Return(false, nil)

		s := newCustomFieldServiceForTest(customFieldServiceDeps{
			logger:            mocklog.NewMockLogger(ctrl),
			tracer:            tracer,
			permissionService: permSvc,
		})
		_, err := s.CreateDefinition(ctx, opts)
		assert.ErrorIs(t, err, service.ErrNoPermission)
	})
}

func TestCustomFieldService_ArchiveDefinition(t *testing.T) {
	t.Parallel()

	userID := model.MustNewID(model.ResourceTypeUser)
	projectID := model.MustNewID(model.ResourceTypeProject)
	def := testModel.NewCustomFieldDefinition(projectID, userID)

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ctx := context.WithValue(context.Background(), pkg.CtxKeyUserID, userID)
		span := mocktrace.NewMockSpan(ctrl)
		span.EXPECT().End(gomock.Len(0)).AnyTimes()
		tracer := mocktrace.NewMockTracer(ctrl)
		tracer.EXPECT().Start(gomock.Any(), "service.customFieldService/UpdateDefinition", gomock.Len(0)).Return(ctx, span)

		permSvc := mocksvc.NewMockPermissionService(ctrl)
		permSvc.EXPECT().CtxUserHas(gomock.Any(), projectID, model.ActionCustomFieldManage).Return(true, nil)

		repo := mockrepo.NewMockCustomFieldRepository(ctrl)
		repo.EXPECT().GetDefinition(gomock.Any(), def.ID).Return(def, nil)
		repo.EXPECT().UpdateDefinition(gomock.Any(), gomock.Any()).DoAndReturn(
			func(_ context.Context, updated *model.CustomFieldDefinition) (*model.CustomFieldDefinition, error) {
				assert.True(t, updated.Archived)
				return updated, nil
			},
		)

		s := newCustomFieldServiceForTest(customFieldServiceDeps{
			logger:            mocklog.NewMockLogger(ctrl),
			tracer:            tracer,
			repo:              repo,
			permissionService: permSvc,
		})
		got, err := s.ArchiveDefinition(ctx, def.ID)
		require.NoError(t, err)
		assert.True(t, got.Archived)
	})
}

func TestCustomFieldService_UpdateDefinition(t *testing.T) {
	t.Parallel()

	userID := model.MustNewID(model.ResourceTypeUser)
	projectID := model.MustNewID(model.ResourceTypeProject)
	def := testModel.NewCustomFieldDefinition(projectID, userID)
	def.SortOrder = 2

	t.Run("order", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ctx := context.WithValue(context.Background(), pkg.CtxKeyUserID, userID)
		span := mocktrace.NewMockSpan(ctrl)
		span.EXPECT().End(gomock.Len(0))
		tracer := mocktrace.NewMockTracer(ctrl)
		tracer.EXPECT().Start(gomock.Any(), "service.customFieldService/UpdateDefinition", gomock.Len(0)).Return(ctx, span)

		permSvc := mocksvc.NewMockPermissionService(ctrl)
		permSvc.EXPECT().CtxUserHas(gomock.Any(), projectID, model.ActionCustomFieldManage).Return(true, nil)

		repo := mockrepo.NewMockCustomFieldRepository(ctrl)
		repo.EXPECT().GetDefinition(gomock.Any(), def.ID).Return(def, nil)
		repo.EXPECT().UpdateDefinition(gomock.Any(), gomock.Any()).DoAndReturn(
			func(_ context.Context, updated *model.CustomFieldDefinition) (*model.CustomFieldDefinition, error) {
				assert.Equal(t, 0, updated.SortOrder)
				return updated, nil
			},
		)

		s := newCustomFieldServiceForTest(customFieldServiceDeps{
			logger:            mocklog.NewMockLogger(ctrl),
			tracer:            tracer,
			repo:              repo,
			permissionService: permSvc,
		})
		got, err := s.UpdateDefinition(ctx, def.ID, service.UpdateCustomFieldOpts{
			SortOrder: optional.Some(0),
		})
		require.NoError(t, err)
		assert.Equal(t, 0, got.SortOrder)
	})
}

func TestCustomFieldService_DeleteDefinition(t *testing.T) {
	t.Parallel()

	userID := model.MustNewID(model.ResourceTypeUser)
	projectID := model.MustNewID(model.ResourceTypeProject)
	def := testModel.NewCustomFieldDefinition(projectID, userID)

	t.Run("in use", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ctx := context.WithValue(context.Background(), pkg.CtxKeyUserID, userID)
		span := mocktrace.NewMockSpan(ctrl)
		span.EXPECT().End(gomock.Len(0))
		tracer := mocktrace.NewMockTracer(ctrl)
		tracer.EXPECT().Start(gomock.Any(), "service.customFieldService/DeleteDefinition", gomock.Len(0)).Return(ctx, span)

		permSvc := mocksvc.NewMockPermissionService(ctrl)
		permSvc.EXPECT().CtxUserHas(gomock.Any(), projectID, model.ActionCustomFieldManage).Return(true, nil)

		repo := mockrepo.NewMockCustomFieldRepository(ctrl)
		repo.EXPECT().GetDefinition(gomock.Any(), def.ID).Return(def, nil)
		repo.EXPECT().CountValues(gomock.Any(), def.ID).Return(int64(2), nil)

		s := newCustomFieldServiceForTest(customFieldServiceDeps{
			logger:            mocklog.NewMockLogger(ctrl),
			tracer:            tracer,
			repo:              repo,
			permissionService: permSvc,
		})
		err := s.DeleteDefinition(ctx, def.ID)
		assert.ErrorIs(t, err, model.ErrCustomFieldInUse)
	})

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ctx := context.WithValue(context.Background(), pkg.CtxKeyUserID, userID)
		span := mocktrace.NewMockSpan(ctrl)
		span.EXPECT().End(gomock.Len(0))
		tracer := mocktrace.NewMockTracer(ctrl)
		tracer.EXPECT().Start(gomock.Any(), "service.customFieldService/DeleteDefinition", gomock.Len(0)).Return(ctx, span)

		permSvc := mocksvc.NewMockPermissionService(ctrl)
		permSvc.EXPECT().CtxUserHas(gomock.Any(), projectID, model.ActionCustomFieldManage).Return(true, nil)

		repo := mockrepo.NewMockCustomFieldRepository(ctrl)
		repo.EXPECT().GetDefinition(gomock.Any(), def.ID).Return(def, nil)
		repo.EXPECT().CountValues(gomock.Any(), def.ID).Return(int64(0), nil)
		repo.EXPECT().DeleteDefinition(gomock.Any(), def.ID).Return(nil)

		s := newCustomFieldServiceForTest(customFieldServiceDeps{
			logger:            mocklog.NewMockLogger(ctrl),
			tracer:            tracer,
			repo:              repo,
			permissionService: permSvc,
		})
		require.NoError(t, s.DeleteDefinition(ctx, def.ID))
	})
}

func TestCustomFieldService_UpdateDefinitionIdentityFrozen(t *testing.T) {
	t.Parallel()

	userID := model.MustNewID(model.ResourceTypeUser)
	projectID := model.MustNewID(model.ResourceTypeProject)
	def := testModel.NewSelectCustomFieldDefinition(projectID, userID)

	t.Run("cannot remove option keys", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ctx := context.WithValue(context.Background(), pkg.CtxKeyUserID, userID)
		span := mocktrace.NewMockSpan(ctrl)
		span.EXPECT().End(gomock.Len(0))
		tracer := mocktrace.NewMockTracer(ctrl)
		tracer.EXPECT().Start(gomock.Any(), "service.customFieldService/UpdateDefinition", gomock.Len(0)).Return(ctx, span)

		permSvc := mocksvc.NewMockPermissionService(ctrl)
		permSvc.EXPECT().CtxUserHas(gomock.Any(), projectID, model.ActionCustomFieldManage).Return(true, nil)

		repo := mockrepo.NewMockCustomFieldRepository(ctrl)
		repo.EXPECT().GetDefinition(gomock.Any(), def.ID).Return(def, nil)

		s := newCustomFieldServiceForTest(customFieldServiceDeps{
			logger:            mocklog.NewMockLogger(ctrl),
			tracer:            tracer,
			repo:              repo,
			permissionService: permSvc,
		})
		_, err := s.UpdateDefinition(ctx, def.ID, service.UpdateCustomFieldOpts{
			Schema: optional.Some(model.CustomFieldSchema{
				Select: &model.CustomFieldSelectSchema{
					Options: []model.CustomFieldOption{{Key: "alpha", Label: "Alpha"}},
				},
			}),
		})
		assert.ErrorIs(t, err, model.ErrCustomFieldOptionInUse)
	})
}

func TestCustomFieldService_ReconcilePending(t *testing.T) {
	t.Parallel()

	resourceID := model.MustNewID(model.ResourceTypeIssue)
	op := repository.CustomFieldOperation{
		ID:         "op-1",
		Kind:       repository.CustomFieldOpStageValues,
		Status:     repository.CustomFieldOpPending,
		ResourceID: resourceID,
	}

	t.Run("commits when the graph resource exists", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ctx := context.Background()
		span := mocktrace.NewMockSpan(ctrl)
		span.EXPECT().End(gomock.Len(0))
		tracer := mocktrace.NewMockTracer(ctrl)
		tracer.EXPECT().Start(gomock.Any(), "service.customFieldService/ReconcilePending", gomock.Len(0)).Return(ctx, span)

		permSvc := mocksvc.NewMockPermissionService(ctrl)
		permSvc.EXPECT().ListScopeAncestry(gomock.Any(), resourceID).Return([]model.ID{resourceID}, nil)

		repo := mockrepo.NewMockCustomFieldRepository(ctrl)
		repo.EXPECT().ListPendingOperations(gomock.Any(), gomock.Any(), 100).Return([]repository.CustomFieldOperation{op}, nil)
		repo.EXPECT().CommitValues(gomock.Any(), resourceID).Return(nil)
		repo.EXPECT().UpdateOperationStatus(gomock.Any(), op.ID, repository.CustomFieldOpCommitted).Return(nil)

		s := newCustomFieldServiceForTest(customFieldServiceDeps{
			logger:            mocklog.NewMockLogger(ctrl),
			tracer:            tracer,
			repo:              repo,
			permissionService: permSvc,
		})
		require.NoError(t, s.ReconcilePending(ctx))
	})

	t.Run("aborts when the graph resource is missing", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ctx := context.Background()
		span := mocktrace.NewMockSpan(ctrl)
		span.EXPECT().End(gomock.Len(0))
		tracer := mocktrace.NewMockTracer(ctrl)
		tracer.EXPECT().Start(gomock.Any(), "service.customFieldService/ReconcilePending", gomock.Len(0)).Return(ctx, span)

		permSvc := mocksvc.NewMockPermissionService(ctrl)
		permSvc.EXPECT().ListScopeAncestry(gomock.Any(), resourceID).Return(nil, repository.ErrNotFound)

		repo := mockrepo.NewMockCustomFieldRepository(ctrl)
		repo.EXPECT().ListPendingOperations(gomock.Any(), gomock.Any(), 100).Return([]repository.CustomFieldOperation{op}, nil)
		repo.EXPECT().AbortValues(gomock.Any(), resourceID).Return(nil)
		repo.EXPECT().UpdateOperationStatus(gomock.Any(), op.ID, repository.CustomFieldOpAborted).Return(nil)

		s := newCustomFieldServiceForTest(customFieldServiceDeps{
			logger:            mocklog.NewMockLogger(ctrl),
			tracer:            tracer,
			repo:              repo,
			permissionService: permSvc,
		})
		require.NoError(t, s.ReconcilePending(ctx))
	})

	t.Run("aborts when ancestry is empty", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ctx := context.Background()
		span := mocktrace.NewMockSpan(ctrl)
		span.EXPECT().End(gomock.Len(0))
		tracer := mocktrace.NewMockTracer(ctrl)
		tracer.EXPECT().Start(gomock.Any(), "service.customFieldService/ReconcilePending", gomock.Len(0)).Return(ctx, span)

		permSvc := mocksvc.NewMockPermissionService(ctrl)
		permSvc.EXPECT().ListScopeAncestry(gomock.Any(), resourceID).Return([]model.ID{}, nil)

		repo := mockrepo.NewMockCustomFieldRepository(ctrl)
		repo.EXPECT().ListPendingOperations(gomock.Any(), gomock.Any(), 100).Return([]repository.CustomFieldOperation{op}, nil)
		repo.EXPECT().AbortValues(gomock.Any(), resourceID).Return(nil)
		repo.EXPECT().UpdateOperationStatus(gomock.Any(), op.ID, repository.CustomFieldOpAborted).Return(nil)

		s := newCustomFieldServiceForTest(customFieldServiceDeps{
			logger:            mocklog.NewMockLogger(ctrl),
			tracer:            tracer,
			repo:              repo,
			permissionService: permSvc,
		})
		require.NoError(t, s.ReconcilePending(ctx))
	})

	t.Run("retries delete resource operations", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ctx := context.Background()
		span := mocktrace.NewMockSpan(ctrl)
		span.EXPECT().End(gomock.Len(0))
		tracer := mocktrace.NewMockTracer(ctrl)
		tracer.EXPECT().Start(gomock.Any(), "service.customFieldService/ReconcilePending", gomock.Len(0)).Return(ctx, span)

		deleteOp := repository.CustomFieldOperation{
			ID:         "op-delete",
			Kind:       repository.CustomFieldOpDeleteResource,
			Status:     repository.CustomFieldOpPending,
			ResourceID: resourceID,
		}
		repo := mockrepo.NewMockCustomFieldRepository(ctrl)
		repo.EXPECT().ListPendingOperations(gomock.Any(), gomock.Any(), 100).Return([]repository.CustomFieldOperation{deleteOp}, nil)
		repo.EXPECT().DeleteForResource(gomock.Any(), resourceID).Return(nil)
		repo.EXPECT().UpdateOperationStatus(gomock.Any(), deleteOp.ID, repository.CustomFieldOpCommitted).Return(nil)

		s := newCustomFieldServiceForTest(customFieldServiceDeps{
			logger: mocklog.NewMockLogger(ctrl),
			tracer: tracer,
			repo:   repo,
		})
		require.NoError(t, s.ReconcilePending(ctx))
	})
}

func TestCustomFieldService_SetValue(t *testing.T) {
	t.Parallel()

	userID := model.MustNewID(model.ResourceTypeUser)
	projectID := model.MustNewID(model.ResourceTypeProject)
	otherProject := model.MustNewID(model.ResourceTypeProject)
	issueID := model.MustNewID(model.ResourceTypeIssue)
	text := "ready"
	value := model.CustomFieldTypedValue{Kind: model.CustomFieldKindText, Text: &text}

	t.Run("rejects a definition outside the resource ancestry", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ctx := context.WithValue(context.Background(), pkg.CtxKeyUserID, userID)
		span := mocktrace.NewMockSpan(ctrl)
		span.EXPECT().End(gomock.Len(0))
		tracer := mocktrace.NewMockTracer(ctrl)
		tracer.EXPECT().Start(gomock.Any(), "service.customFieldService/SetValue", gomock.Len(0)).Return(ctx, span)

		permSvc := mocksvc.NewMockPermissionService(ctrl)
		permSvc.EXPECT().CtxUserHas(gomock.Any(), issueID, model.ActionIssueUpdate).Return(true, nil)
		permSvc.EXPECT().ListScopeAncestry(gomock.Any(), issueID).Return([]model.ID{issueID, projectID}, nil)

		foreign := testModel.NewCustomFieldDefinition(otherProject, userID)
		repo := mockrepo.NewMockCustomFieldRepository(ctrl)
		repo.EXPECT().GetDefinition(gomock.Any(), foreign.ID).Return(foreign, nil)

		s := newCustomFieldServiceForTest(customFieldServiceDeps{
			logger:            mocklog.NewMockLogger(ctrl),
			tracer:            tracer,
			repo:              repo,
			permissionService: permSvc,
		})
		err := s.SetValue(ctx, issueID, foreign.ID, value)
		assert.ErrorIs(t, err, repository.ErrNotFound)
	})

	t.Run("writes when the definition is in ancestry", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ctx := context.WithValue(context.Background(), pkg.CtxKeyUserID, userID)
		span := mocktrace.NewMockSpan(ctrl)
		span.EXPECT().End(gomock.Len(0))
		tracer := mocktrace.NewMockTracer(ctrl)
		tracer.EXPECT().Start(gomock.Any(), "service.customFieldService/SetValue", gomock.Len(0)).Return(ctx, span)

		permSvc := mocksvc.NewMockPermissionService(ctrl)
		permSvc.EXPECT().CtxUserHas(gomock.Any(), issueID, model.ActionIssueUpdate).Return(true, nil)
		permSvc.EXPECT().ListScopeAncestry(gomock.Any(), issueID).Return([]model.ID{issueID, projectID}, nil)

		def := testModel.NewCustomFieldDefinition(projectID, userID)
		repo := mockrepo.NewMockCustomFieldRepository(ctrl)
		repo.EXPECT().GetDefinition(gomock.Any(), def.ID).Return(def, nil)
		repo.EXPECT().ReplaceValues(gomock.Any(), def, issueID, gomock.Any(), true).Return(nil)

		s := newCustomFieldServiceForTest(customFieldServiceDeps{
			logger:            mocklog.NewMockLogger(ctrl),
			tracer:            tracer,
			repo:              repo,
			permissionService: permSvc,
		})
		require.NoError(t, s.SetValue(ctx, issueID, def.ID, value))
	})

	t.Run("rejects archived definition", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ctx := context.WithValue(context.Background(), pkg.CtxKeyUserID, userID)
		span := mocktrace.NewMockSpan(ctrl)
		span.EXPECT().End(gomock.Len(0))
		tracer := mocktrace.NewMockTracer(ctrl)
		tracer.EXPECT().Start(gomock.Any(), "service.customFieldService/SetValue", gomock.Len(0)).Return(ctx, span)

		permSvc := mocksvc.NewMockPermissionService(ctrl)
		permSvc.EXPECT().CtxUserHas(gomock.Any(), issueID, model.ActionIssueUpdate).Return(true, nil)
		permSvc.EXPECT().ListScopeAncestry(gomock.Any(), issueID).Return([]model.ID{issueID, projectID}, nil)

		def := testModel.NewCustomFieldDefinition(projectID, userID)
		def.Archived = true
		repo := mockrepo.NewMockCustomFieldRepository(ctrl)
		repo.EXPECT().GetDefinition(gomock.Any(), def.ID).Return(def, nil)

		s := newCustomFieldServiceForTest(customFieldServiceDeps{
			logger:            mocklog.NewMockLogger(ctrl),
			tracer:            tracer,
			repo:              repo,
			permissionService: permSvc,
		})
		err := s.SetValue(ctx, issueID, def.ID, value)
		assert.ErrorIs(t, err, model.ErrCustomFieldArchived)
	})
}

func TestCustomFieldService_ListEffective(t *testing.T) {
	t.Parallel()

	userID := model.MustNewID(model.ResourceTypeUser)
	projectID := model.MustNewID(model.ResourceTypeProject)
	issueID := model.MustNewID(model.ResourceTypeIssue)

	t.Run("omits archived definitions even with stored values", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ctx := context.WithValue(context.Background(), pkg.CtxKeyUserID, userID)
		span := mocktrace.NewMockSpan(ctrl)
		span.EXPECT().End(gomock.Len(0))
		tracer := mocktrace.NewMockTracer(ctrl)
		tracer.EXPECT().Start(gomock.Any(), "service.customFieldService/ListEffective", gomock.Len(0)).Return(ctx, span)

		permSvc := mocksvc.NewMockPermissionService(ctrl)
		permSvc.EXPECT().CtxUserHas(gomock.Any(), issueID, model.ActionIssueRead).Return(true, nil)
		permSvc.EXPECT().ListScopeAncestry(gomock.Any(), issueID).Return([]model.ID{projectID}, nil)

		archived := testModel.NewCustomFieldDefinition(projectID, userID)
		archived.Archived = true
		active := testModel.NewCustomFieldDefinition(projectID, userID)
		repo := mockrepo.NewMockCustomFieldRepository(ctrl)
		repo.EXPECT().ListDefinitions(gomock.Any(), []model.ID{projectID}, model.ResourceTypeIssue, false).
			Return([]*model.CustomFieldDefinition{archived, active}, nil)
		repo.EXPECT().ListValues(gomock.Any(), issueID, false).Return(nil, nil)

		s := newCustomFieldServiceForTest(customFieldServiceDeps{
			logger:            mocklog.NewMockLogger(ctrl),
			tracer:            tracer,
			repo:              repo,
			permissionService: permSvc,
		})
		entries, err := s.ListEffective(ctx, issueID)
		require.NoError(t, err)
		require.Len(t, entries, 1)
		assert.Equal(t, active.ID, entries[0].Definition.ID)
	})
}

func TestCustomFieldService_StageForResource(t *testing.T) {
	t.Parallel()

	userID := model.MustNewID(model.ResourceTypeUser)
	projectID := model.MustNewID(model.ResourceTypeProject)
	issueID := model.MustNewID(model.ResourceTypeIssue)

	t.Run("does not write when a required field is missing", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ctx := context.Background()
		span := mocktrace.NewMockSpan(ctrl)
		span.EXPECT().End(gomock.Len(0))
		tracer := mocktrace.NewMockTracer(ctrl)
		tracer.EXPECT().Start(gomock.Any(), "service.customFieldService/StageForResource", gomock.Len(0)).Return(ctx, span)

		permSvc := mocksvc.NewMockPermissionService(ctrl)
		permSvc.EXPECT().ListScopeAncestry(gomock.Any(), projectID).Return([]model.ID{projectID}, nil)

		required := testModel.NewCustomFieldDefinition(projectID, userID)
		required.Required = true
		repo := mockrepo.NewMockCustomFieldRepository(ctrl)
		repo.EXPECT().ListDefinitions(gomock.Any(), []model.ID{projectID}, model.ResourceTypeIssue, false).
			Return([]*model.CustomFieldDefinition{required}, nil)

		s := newCustomFieldServiceForTest(customFieldServiceDeps{
			logger:            mocklog.NewMockLogger(ctrl),
			tracer:            tracer,
			repo:              repo,
			permissionService: permSvc,
		})
		err := s.StageForResource(ctx, projectID, issueID, nil)
		assert.ErrorIs(t, err, model.ErrCustomFieldRequired)
	})

	t.Run("skips the operation when there are no writes", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ctx := context.Background()
		span := mocktrace.NewMockSpan(ctrl)
		span.EXPECT().End(gomock.Len(0))
		tracer := mocktrace.NewMockTracer(ctrl)
		tracer.EXPECT().Start(gomock.Any(), "service.customFieldService/StageForResource", gomock.Len(0)).Return(ctx, span)

		permSvc := mocksvc.NewMockPermissionService(ctrl)
		permSvc.EXPECT().ListScopeAncestry(gomock.Any(), projectID).Return([]model.ID{projectID}, nil)

		repo := mockrepo.NewMockCustomFieldRepository(ctrl)
		repo.EXPECT().ListDefinitions(gomock.Any(), []model.ID{projectID}, model.ResourceTypeIssue, false).
			Return([]*model.CustomFieldDefinition{}, nil)

		s := newCustomFieldServiceForTest(customFieldServiceDeps{
			logger:            mocklog.NewMockLogger(ctrl),
			tracer:            tracer,
			repo:              repo,
			permissionService: permSvc,
		})
		require.NoError(t, s.StageForResource(ctx, projectID, issueID, nil))
	})
}

func TestCustomFieldService_GetDefinition(t *testing.T) {
	t.Parallel()

	userID := model.MustNewID(model.ResourceTypeUser)
	projectID := model.MustNewID(model.ResourceTypeProject)
	def := testModel.NewCustomFieldDefinition(projectID, userID)

	t.Run("manage permission", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ctx := context.WithValue(context.Background(), pkg.CtxKeyUserID, userID)
		span := mocktrace.NewMockSpan(ctrl)
		span.EXPECT().End(gomock.Len(0))
		tracer := mocktrace.NewMockTracer(ctrl)
		tracer.EXPECT().Start(gomock.Any(), "service.customFieldService/GetDefinition", gomock.Len(0)).Return(ctx, span)

		permSvc := mocksvc.NewMockPermissionService(ctrl)
		permSvc.EXPECT().CtxUserHas(gomock.Any(), projectID, model.ActionCustomFieldManage).Return(true, nil)

		repo := mockrepo.NewMockCustomFieldRepository(ctrl)
		repo.EXPECT().GetDefinition(gomock.Any(), def.ID).Return(def, nil)

		s := newCustomFieldServiceForTest(customFieldServiceDeps{
			logger:            mocklog.NewMockLogger(ctrl),
			tracer:            tracer,
			repo:              repo,
			permissionService: permSvc,
		})
		got, err := s.GetDefinition(ctx, def.ID)
		require.NoError(t, err)
		assert.Equal(t, def.ID, got.ID)
	})

	t.Run("falls back to read", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ctx := context.WithValue(context.Background(), pkg.CtxKeyUserID, userID)
		span := mocktrace.NewMockSpan(ctrl)
		span.EXPECT().End(gomock.Len(0))
		tracer := mocktrace.NewMockTracer(ctrl)
		tracer.EXPECT().Start(gomock.Any(), "service.customFieldService/GetDefinition", gomock.Len(0)).Return(ctx, span)

		permSvc := mocksvc.NewMockPermissionService(ctrl)
		permSvc.EXPECT().CtxUserHas(gomock.Any(), projectID, model.ActionCustomFieldManage).Return(false, nil)
		permSvc.EXPECT().CtxUserHas(gomock.Any(), projectID, model.ActionProjectRead).Return(true, nil)

		repo := mockrepo.NewMockCustomFieldRepository(ctrl)
		repo.EXPECT().GetDefinition(gomock.Any(), def.ID).Return(def, nil)

		s := newCustomFieldServiceForTest(customFieldServiceDeps{
			logger:            mocklog.NewMockLogger(ctrl),
			tracer:            tracer,
			repo:              repo,
			permissionService: permSvc,
		})
		got, err := s.GetDefinition(ctx, def.ID)
		require.NoError(t, err)
		assert.Equal(t, def.ID, got.ID)
	})
}

func TestCustomFieldService_ListDefinitions(t *testing.T) {
	t.Parallel()

	userID := model.MustNewID(model.ResourceTypeUser)
	projectID := model.MustNewID(model.ResourceTypeProject)
	def := testModel.NewCustomFieldDefinition(projectID, userID)

	t.Run("manage includes archived", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ctx := context.WithValue(context.Background(), pkg.CtxKeyUserID, userID)
		span := mocktrace.NewMockSpan(ctrl)
		span.EXPECT().End(gomock.Len(0))
		tracer := mocktrace.NewMockTracer(ctrl)
		tracer.EXPECT().Start(gomock.Any(), "service.customFieldService/ListDefinitions", gomock.Len(0)).Return(ctx, span)

		permSvc := mocksvc.NewMockPermissionService(ctrl)
		permSvc.EXPECT().CtxUserHas(gomock.Any(), projectID, model.ActionCustomFieldManage).Return(true, nil)
		permSvc.EXPECT().ListScopeAncestry(gomock.Any(), projectID).Return([]model.ID{projectID}, nil)

		repo := mockrepo.NewMockCustomFieldRepository(ctrl)
		repo.EXPECT().ListDefinitions(gomock.Any(), []model.ID{projectID}, model.ResourceTypeIssue, true).
			Return([]*model.CustomFieldDefinition{def}, nil)

		s := newCustomFieldServiceForTest(customFieldServiceDeps{
			logger:            mocklog.NewMockLogger(ctrl),
			tracer:            tracer,
			repo:              repo,
			permissionService: permSvc,
		})
		got, err := s.ListDefinitions(ctx, projectID, model.ResourceTypeIssue, true)
		require.NoError(t, err)
		require.Len(t, got, 1)
	})

	t.Run("read forces include archived false", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ctx := context.WithValue(context.Background(), pkg.CtxKeyUserID, userID)
		span := mocktrace.NewMockSpan(ctrl)
		span.EXPECT().End(gomock.Len(0))
		tracer := mocktrace.NewMockTracer(ctrl)
		tracer.EXPECT().Start(gomock.Any(), "service.customFieldService/ListDefinitions", gomock.Len(0)).Return(ctx, span)

		permSvc := mocksvc.NewMockPermissionService(ctrl)
		permSvc.EXPECT().CtxUserHas(gomock.Any(), projectID, model.ActionCustomFieldManage).Return(false, nil)
		permSvc.EXPECT().CtxUserHas(gomock.Any(), projectID, model.ActionProjectRead).Return(true, nil)
		permSvc.EXPECT().ListScopeAncestry(gomock.Any(), projectID).Return([]model.ID{projectID}, nil)

		repo := mockrepo.NewMockCustomFieldRepository(ctrl)
		repo.EXPECT().ListDefinitions(gomock.Any(), []model.ID{projectID}, model.ResourceTypeIssue, false).
			Return([]*model.CustomFieldDefinition{def}, nil)

		s := newCustomFieldServiceForTest(customFieldServiceDeps{
			logger:            mocklog.NewMockLogger(ctrl),
			tracer:            tracer,
			repo:              repo,
			permissionService: permSvc,
		})
		got, err := s.ListDefinitions(ctx, projectID, model.ResourceTypeIssue, true)
		require.NoError(t, err)
		require.Len(t, got, 1)
	})
}

func TestCustomFieldService_DeleteValue(t *testing.T) {
	t.Parallel()

	userID := model.MustNewID(model.ResourceTypeUser)
	projectID := model.MustNewID(model.ResourceTypeProject)
	issueID := model.MustNewID(model.ResourceTypeIssue)

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ctx := context.WithValue(context.Background(), pkg.CtxKeyUserID, userID)
		span := mocktrace.NewMockSpan(ctrl)
		span.EXPECT().End(gomock.Len(0))
		tracer := mocktrace.NewMockTracer(ctrl)
		tracer.EXPECT().Start(gomock.Any(), "service.customFieldService/DeleteValue", gomock.Len(0)).Return(ctx, span)

		permSvc := mocksvc.NewMockPermissionService(ctrl)
		permSvc.EXPECT().CtxUserHas(gomock.Any(), issueID, model.ActionIssueUpdate).Return(true, nil)
		permSvc.EXPECT().ListScopeAncestry(gomock.Any(), issueID).Return([]model.ID{issueID, projectID}, nil)

		def := testModel.NewCustomFieldDefinition(projectID, userID)
		repo := mockrepo.NewMockCustomFieldRepository(ctrl)
		repo.EXPECT().GetDefinition(gomock.Any(), def.ID).Return(def, nil)
		repo.EXPECT().DeleteValues(gomock.Any(), def.ID, issueID).Return(nil)

		s := newCustomFieldServiceForTest(customFieldServiceDeps{
			logger:            mocklog.NewMockLogger(ctrl),
			tracer:            tracer,
			repo:              repo,
			permissionService: permSvc,
		})
		require.NoError(t, s.DeleteValue(ctx, issueID, def.ID))
	})

	t.Run("required field", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ctx := context.WithValue(context.Background(), pkg.CtxKeyUserID, userID)
		span := mocktrace.NewMockSpan(ctrl)
		span.EXPECT().End(gomock.Len(0))
		tracer := mocktrace.NewMockTracer(ctrl)
		tracer.EXPECT().Start(gomock.Any(), "service.customFieldService/DeleteValue", gomock.Len(0)).Return(ctx, span)

		permSvc := mocksvc.NewMockPermissionService(ctrl)
		permSvc.EXPECT().CtxUserHas(gomock.Any(), issueID, model.ActionIssueUpdate).Return(true, nil)
		permSvc.EXPECT().ListScopeAncestry(gomock.Any(), issueID).Return([]model.ID{issueID, projectID}, nil)

		def := testModel.NewCustomFieldDefinition(projectID, userID)
		def.Required = true
		repo := mockrepo.NewMockCustomFieldRepository(ctrl)
		repo.EXPECT().GetDefinition(gomock.Any(), def.ID).Return(def, nil)

		s := newCustomFieldServiceForTest(customFieldServiceDeps{
			logger:            mocklog.NewMockLogger(ctrl),
			tracer:            tracer,
			repo:              repo,
			permissionService: permSvc,
		})
		err := s.DeleteValue(ctx, issueID, def.ID)
		assert.ErrorIs(t, err, model.ErrCustomFieldRequired)
	})
}

func TestCustomFieldService_Search(t *testing.T) {
	t.Parallel()

	userID := model.MustNewID(model.ResourceTypeUser)
	projectID := model.MustNewID(model.ResourceTypeProject)
	def := testModel.NewIntegerCustomFieldDefinition(projectID, userID)
	allowed := model.MustNewID(model.ResourceTypeIssue)
	denied := model.MustNewID(model.ResourceTypeIssue)
	n := int64(8)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.WithValue(context.Background(), pkg.CtxKeyUserID, userID)
	span := mocktrace.NewMockSpan(ctrl)
	span.EXPECT().End(gomock.Len(0))
	tracer := mocktrace.NewMockTracer(ctrl)
	tracer.EXPECT().Start(gomock.Any(), "service.customFieldService/Search", gomock.Len(0)).Return(ctx, span)

	permSvc := mocksvc.NewMockPermissionService(ctrl)
	permSvc.EXPECT().CtxUserHas(gomock.Any(), allowed, model.ActionIssueRead).Return(true, nil)
	permSvc.EXPECT().CtxUserHas(gomock.Any(), denied, model.ActionIssueRead).Return(false, nil)

	repo := mockrepo.NewMockCustomFieldRepository(ctrl)
	repo.EXPECT().GetDefinition(gomock.Any(), def.ID).Return(def, nil)
	repo.EXPECT().Search(gomock.Any(), def.ID, gomock.Any(), 10).Return([]model.ID{allowed, denied}, nil)

	s := newCustomFieldServiceForTest(customFieldServiceDeps{
		logger:            mocklog.NewMockLogger(ctrl),
		tracer:            tracer,
		repo:              repo,
		permissionService: permSvc,
	})
	got, err := s.Search(ctx, service.CustomFieldSearchQuery{
		DefinitionID: def.ID,
		Predicate:    repository.CustomFieldPredicate{Op: repository.CustomFieldPredEq, Integer: &n},
		Limit:        10,
	})
	require.NoError(t, err)
	assert.Equal(t, []model.ID{allowed}, got)
}

func TestCustomFieldService_CommitAbortDeleteForResource(t *testing.T) {
	t.Parallel()

	resourceID := model.MustNewID(model.ResourceTypeIssue)

	t.Run("commit", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo := mockrepo.NewMockCustomFieldRepository(ctrl)
		repo.EXPECT().CommitValues(gomock.Any(), resourceID).Return(nil)
		repo.EXPECT().UpdatePendingOperations(gomock.Any(), resourceID, repository.CustomFieldOpCommitted).Return(nil)

		s := newCustomFieldServiceForTest(customFieldServiceDeps{
			logger: mocklog.NewMockLogger(ctrl),
			repo:   repo,
		})
		require.NoError(t, s.CommitForResource(context.Background(), resourceID))
	})

	t.Run("abort", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo := mockrepo.NewMockCustomFieldRepository(ctrl)
		repo.EXPECT().AbortValues(gomock.Any(), resourceID).Return(nil)
		repo.EXPECT().UpdatePendingOperations(gomock.Any(), resourceID, repository.CustomFieldOpAborted).Return(nil)

		s := newCustomFieldServiceForTest(customFieldServiceDeps{
			logger: mocklog.NewMockLogger(ctrl),
			repo:   repo,
		})
		require.NoError(t, s.AbortForResource(context.Background(), resourceID))
	})

	t.Run("delete success", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo := mockrepo.NewMockCustomFieldRepository(ctrl)
		repo.EXPECT().DeleteForResource(gomock.Any(), resourceID).Return(nil)

		s := newCustomFieldServiceForTest(customFieldServiceDeps{
			logger: mocklog.NewMockLogger(ctrl),
			repo:   repo,
		})
		require.NoError(t, s.DeleteForResource(context.Background(), resourceID))
	})

	t.Run("delete records pending operation on failure", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo := mockrepo.NewMockCustomFieldRepository(ctrl)
		repo.EXPECT().DeleteForResource(gomock.Any(), resourceID).Return(assert.AnError)
		repo.EXPECT().CreateOperation(gomock.Any(), gomock.Any()).Return(&repository.CustomFieldOperation{}, nil)

		s := newCustomFieldServiceForTest(customFieldServiceDeps{
			logger: mocklog.NewMockLogger(ctrl),
			repo:   repo,
		})
		err := s.DeleteForResource(context.Background(), resourceID)
		assert.ErrorIs(t, err, service.ErrCustomFieldDelete)
		assert.ErrorIs(t, err, assert.AnError)
	})
}

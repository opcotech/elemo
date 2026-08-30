package service_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/opcotech/elemo/internal/license"
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
	licenseService    service.LicenseService
}

func newCustomFieldServiceForTest(deps customFieldServiceDeps) service.CustomFieldService {
	if deps.repo == nil {
		deps.repo = mockrepo.NewMockCustomFieldRepository(nil)
	}
	if deps.permissionService == nil {
		deps.permissionService = mocksvc.NewMockPermissionService(nil)
	}
	if deps.licenseService == nil {
		deps.licenseService = mocksvc.NewMockLicenseService(nil)
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
		deps.licenseService,
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
			mocksvc.NewMockLicenseService(ctrl),
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
			mocksvc.NewMockLicenseService(ctrl),
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
			mocksvc.NewMockLicenseService(nil),
			service.WithLogger(nil),
		)
		assert.ErrorIs(t, err, log.ErrNoLogger)
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
		tracer.EXPECT().Start(ctx, "service.customFieldService/CreateDefinition", gomock.Len(0)).Return(ctx, span)

		licenseSvc := mocksvc.NewMockLicenseService(ctrl)
		licenseSvc.EXPECT().HasFeature(ctx, license.FeatureCustomFields).Return(true, nil)

		permSvc := mocksvc.NewMockPermissionService(ctrl)
		permSvc.EXPECT().CtxUserHas(ctx, projectID, model.ActionCustomFieldManage).Return(true, nil)
		permSvc.EXPECT().ListScopeAncestry(ctx, projectID).Return([]model.ID{projectID}, nil)

		created := testModel.NewIntegerCustomFieldDefinition(projectID, userID)
		repo := mockrepo.NewMockCustomFieldRepository(ctrl)
		repo.EXPECT().NextSortOrder(ctx, projectID, model.ResourceTypeIssue).Return(4, nil)
		repo.EXPECT().CreateDefinition(ctx, gomock.Any()).DoAndReturn(
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
			licenseService:    licenseSvc,
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
		tracer.EXPECT().Start(ctx, "service.customFieldService/CreateDefinition", gomock.Len(0)).Return(ctx, span)

		licenseSvc := mocksvc.NewMockLicenseService(ctrl)
		licenseSvc.EXPECT().HasFeature(ctx, license.FeatureCustomFields).Return(true, nil)

		permSvc := mocksvc.NewMockPermissionService(ctrl)
		permSvc.EXPECT().CtxUserHas(ctx, projectID, model.ActionCustomFieldManage).Return(true, nil)
		permSvc.EXPECT().ListScopeAncestry(ctx, projectID).Return([]model.ID{projectID}, nil)

		created := testModel.NewIntegerCustomFieldDefinition(projectID, userID)
		repo := mockrepo.NewMockCustomFieldRepository(ctrl)
		repo.EXPECT().CreateDefinition(ctx, gomock.Any()).DoAndReturn(
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
			licenseService:    licenseSvc,
		})
		createOpts := opts
		createOpts.SortOrder = optional.Some(0)
		got, err := s.CreateDefinition(ctx, createOpts)
		require.NoError(t, err)
		assert.Equal(t, created.ID, got.ID)
	})

	t.Run("feature unavailable", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ctx := context.WithValue(context.Background(), pkg.CtxKeyUserID, userID)
		span := mocktrace.NewMockSpan(ctrl)
		span.EXPECT().End(gomock.Len(0))
		tracer := mocktrace.NewMockTracer(ctrl)
		tracer.EXPECT().Start(ctx, "service.customFieldService/CreateDefinition", gomock.Len(0)).Return(ctx, span)

		licenseSvc := mocksvc.NewMockLicenseService(ctrl)
		licenseSvc.EXPECT().HasFeature(ctx, license.FeatureCustomFields).Return(false, nil)

		s := newCustomFieldServiceForTest(customFieldServiceDeps{
			logger:         mocklog.NewMockLogger(ctrl),
			tracer:         tracer,
			licenseService: licenseSvc,
		})
		_, err := s.CreateDefinition(ctx, opts)
		assert.ErrorIs(t, err, service.ErrFeatureDisabled)
	})

	t.Run("no permission", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ctx := context.WithValue(context.Background(), pkg.CtxKeyUserID, userID)
		span := mocktrace.NewMockSpan(ctrl)
		span.EXPECT().End(gomock.Len(0))
		tracer := mocktrace.NewMockTracer(ctrl)
		tracer.EXPECT().Start(ctx, "service.customFieldService/CreateDefinition", gomock.Len(0)).Return(ctx, span)

		licenseSvc := mocksvc.NewMockLicenseService(ctrl)
		licenseSvc.EXPECT().HasFeature(ctx, license.FeatureCustomFields).Return(true, nil)

		permSvc := mocksvc.NewMockPermissionService(ctrl)
		permSvc.EXPECT().CtxUserHas(ctx, projectID, model.ActionCustomFieldManage).Return(false, nil)

		s := newCustomFieldServiceForTest(customFieldServiceDeps{
			logger:            mocklog.NewMockLogger(ctrl),
			tracer:            tracer,
			permissionService: permSvc,
			licenseService:    licenseSvc,
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
		tracer.EXPECT().Start(ctx, "service.customFieldService/UpdateDefinition", gomock.Len(0)).Return(ctx, span)

		licenseSvc := mocksvc.NewMockLicenseService(ctrl)
		licenseSvc.EXPECT().HasFeature(ctx, license.FeatureCustomFields).Return(true, nil)

		permSvc := mocksvc.NewMockPermissionService(ctrl)
		permSvc.EXPECT().CtxUserHas(ctx, projectID, model.ActionCustomFieldManage).Return(true, nil)

		repo := mockrepo.NewMockCustomFieldRepository(ctrl)
		repo.EXPECT().GetDefinition(ctx, def.ID).Return(def, nil)
		repo.EXPECT().UpdateDefinition(ctx, gomock.Any()).DoAndReturn(
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
			licenseService:    licenseSvc,
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
		tracer.EXPECT().Start(ctx, "service.customFieldService/UpdateDefinition", gomock.Len(0)).Return(ctx, span)

		licenseSvc := mocksvc.NewMockLicenseService(ctrl)
		licenseSvc.EXPECT().HasFeature(ctx, license.FeatureCustomFields).Return(true, nil)

		permSvc := mocksvc.NewMockPermissionService(ctrl)
		permSvc.EXPECT().CtxUserHas(ctx, projectID, model.ActionCustomFieldManage).Return(true, nil)

		repo := mockrepo.NewMockCustomFieldRepository(ctrl)
		repo.EXPECT().GetDefinition(ctx, def.ID).Return(def, nil)
		repo.EXPECT().UpdateDefinition(ctx, gomock.Any()).DoAndReturn(
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
			licenseService:    licenseSvc,
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
		tracer.EXPECT().Start(ctx, "service.customFieldService/DeleteDefinition", gomock.Len(0)).Return(ctx, span)

		licenseSvc := mocksvc.NewMockLicenseService(ctrl)
		licenseSvc.EXPECT().HasFeature(ctx, license.FeatureCustomFields).Return(true, nil)

		permSvc := mocksvc.NewMockPermissionService(ctrl)
		permSvc.EXPECT().CtxUserHas(ctx, projectID, model.ActionCustomFieldManage).Return(true, nil)

		repo := mockrepo.NewMockCustomFieldRepository(ctrl)
		repo.EXPECT().GetDefinition(ctx, def.ID).Return(def, nil)
		repo.EXPECT().CountValues(ctx, def.ID).Return(int64(2), nil)

		s := newCustomFieldServiceForTest(customFieldServiceDeps{
			logger:            mocklog.NewMockLogger(ctrl),
			tracer:            tracer,
			repo:              repo,
			permissionService: permSvc,
			licenseService:    licenseSvc,
		})
		err := s.DeleteDefinition(ctx, def.ID)
		assert.ErrorIs(t, err, model.ErrCustomFieldInUse)
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
		tracer.EXPECT().Start(ctx, "service.customFieldService/UpdateDefinition", gomock.Len(0)).Return(ctx, span)

		licenseSvc := mocksvc.NewMockLicenseService(ctrl)
		licenseSvc.EXPECT().HasFeature(ctx, license.FeatureCustomFields).Return(true, nil)

		permSvc := mocksvc.NewMockPermissionService(ctrl)
		permSvc.EXPECT().CtxUserHas(ctx, projectID, model.ActionCustomFieldManage).Return(true, nil)

		repo := mockrepo.NewMockCustomFieldRepository(ctrl)
		repo.EXPECT().GetDefinition(ctx, def.ID).Return(def, nil)

		s := newCustomFieldServiceForTest(customFieldServiceDeps{
			logger:            mocklog.NewMockLogger(ctrl),
			tracer:            tracer,
			repo:              repo,
			permissionService: permSvc,
			licenseService:    licenseSvc,
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
		tracer.EXPECT().Start(ctx, "service.customFieldService/ReconcilePending", gomock.Len(0)).Return(ctx, span)

		permSvc := mocksvc.NewMockPermissionService(ctrl)
		permSvc.EXPECT().ListScopeAncestry(ctx, resourceID).Return([]model.ID{resourceID}, nil)

		repo := mockrepo.NewMockCustomFieldRepository(ctrl)
		repo.EXPECT().ListPendingOperations(ctx, gomock.Any(), 100).Return([]repository.CustomFieldOperation{op}, nil)
		repo.EXPECT().CommitValues(ctx, resourceID).Return(nil)
		repo.EXPECT().UpdateOperationStatus(ctx, op.ID, repository.CustomFieldOpCommitted).Return(nil)

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
		tracer.EXPECT().Start(ctx, "service.customFieldService/ReconcilePending", gomock.Len(0)).Return(ctx, span)

		permSvc := mocksvc.NewMockPermissionService(ctrl)
		permSvc.EXPECT().ListScopeAncestry(ctx, resourceID).Return(nil, repository.ErrNotFound)

		repo := mockrepo.NewMockCustomFieldRepository(ctrl)
		repo.EXPECT().ListPendingOperations(ctx, gomock.Any(), 100).Return([]repository.CustomFieldOperation{op}, nil)
		repo.EXPECT().AbortValues(ctx, resourceID).Return(nil)
		repo.EXPECT().UpdateOperationStatus(ctx, op.ID, repository.CustomFieldOpAborted).Return(nil)

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
		tracer.EXPECT().Start(ctx, "service.customFieldService/ReconcilePending", gomock.Len(0)).Return(ctx, span)

		permSvc := mocksvc.NewMockPermissionService(ctrl)
		permSvc.EXPECT().ListScopeAncestry(ctx, resourceID).Return([]model.ID{}, nil)

		repo := mockrepo.NewMockCustomFieldRepository(ctrl)
		repo.EXPECT().ListPendingOperations(ctx, gomock.Any(), 100).Return([]repository.CustomFieldOperation{op}, nil)
		repo.EXPECT().AbortValues(ctx, resourceID).Return(nil)
		repo.EXPECT().UpdateOperationStatus(ctx, op.ID, repository.CustomFieldOpAborted).Return(nil)

		s := newCustomFieldServiceForTest(customFieldServiceDeps{
			logger:            mocklog.NewMockLogger(ctrl),
			tracer:            tracer,
			repo:              repo,
			permissionService: permSvc,
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
		tracer.EXPECT().Start(ctx, "service.customFieldService/SetValue", gomock.Len(0)).Return(ctx, span)

		licenseSvc := mocksvc.NewMockLicenseService(ctrl)
		licenseSvc.EXPECT().HasFeature(ctx, license.FeatureCustomFields).Return(true, nil)

		permSvc := mocksvc.NewMockPermissionService(ctrl)
		permSvc.EXPECT().CtxUserHas(ctx, issueID, model.ActionIssueUpdate).Return(true, nil)
		permSvc.EXPECT().ListScopeAncestry(ctx, issueID).Return([]model.ID{issueID, projectID}, nil)

		foreign := testModel.NewCustomFieldDefinition(otherProject, userID)
		repo := mockrepo.NewMockCustomFieldRepository(ctrl)
		repo.EXPECT().GetDefinition(ctx, foreign.ID).Return(foreign, nil)

		s := newCustomFieldServiceForTest(customFieldServiceDeps{
			logger:            mocklog.NewMockLogger(ctrl),
			tracer:            tracer,
			repo:              repo,
			permissionService: permSvc,
			licenseService:    licenseSvc,
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
		tracer.EXPECT().Start(ctx, "service.customFieldService/SetValue", gomock.Len(0)).Return(ctx, span)

		licenseSvc := mocksvc.NewMockLicenseService(ctrl)
		licenseSvc.EXPECT().HasFeature(ctx, license.FeatureCustomFields).Return(true, nil)

		permSvc := mocksvc.NewMockPermissionService(ctrl)
		permSvc.EXPECT().CtxUserHas(ctx, issueID, model.ActionIssueUpdate).Return(true, nil)
		permSvc.EXPECT().ListScopeAncestry(ctx, issueID).Return([]model.ID{issueID, projectID}, nil)

		def := testModel.NewCustomFieldDefinition(projectID, userID)
		repo := mockrepo.NewMockCustomFieldRepository(ctrl)
		repo.EXPECT().GetDefinition(ctx, def.ID).Return(def, nil)
		repo.EXPECT().ReplaceValues(ctx, def, issueID, gomock.Any(), true).Return(nil)

		s := newCustomFieldServiceForTest(customFieldServiceDeps{
			logger:            mocklog.NewMockLogger(ctrl),
			tracer:            tracer,
			repo:              repo,
			permissionService: permSvc,
			licenseService:    licenseSvc,
		})
		require.NoError(t, s.SetValue(ctx, issueID, def.ID, value))
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
		tracer.EXPECT().Start(ctx, "service.customFieldService/ListEffective", gomock.Len(0)).Return(ctx, span)

		permSvc := mocksvc.NewMockPermissionService(ctrl)
		permSvc.EXPECT().CtxUserHas(ctx, issueID, model.ActionIssueRead).Return(true, nil)
		permSvc.EXPECT().ListScopeAncestry(ctx, issueID).Return([]model.ID{projectID}, nil)

		archived := testModel.NewCustomFieldDefinition(projectID, userID)
		archived.Archived = true
		active := testModel.NewCustomFieldDefinition(projectID, userID)
		repo := mockrepo.NewMockCustomFieldRepository(ctrl)
		repo.EXPECT().ListDefinitions(ctx, []model.ID{projectID}, model.ResourceTypeIssue, false).
			Return([]*model.CustomFieldDefinition{archived, active}, nil)
		repo.EXPECT().ListValues(ctx, issueID, false).Return(nil, nil)

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
		tracer.EXPECT().Start(ctx, "service.customFieldService/StageForResource", gomock.Len(0)).Return(ctx, span)

		licenseSvc := mocksvc.NewMockLicenseService(ctrl)
		licenseSvc.EXPECT().HasFeature(ctx, license.FeatureCustomFields).Return(true, nil)

		permSvc := mocksvc.NewMockPermissionService(ctrl)
		permSvc.EXPECT().ListScopeAncestry(ctx, projectID).Return([]model.ID{projectID}, nil)

		required := testModel.NewCustomFieldDefinition(projectID, userID)
		required.Required = true
		repo := mockrepo.NewMockCustomFieldRepository(ctrl)
		repo.EXPECT().ListDefinitions(ctx, []model.ID{projectID}, model.ResourceTypeIssue, false).
			Return([]*model.CustomFieldDefinition{required}, nil)

		s := newCustomFieldServiceForTest(customFieldServiceDeps{
			logger:            mocklog.NewMockLogger(ctrl),
			tracer:            tracer,
			repo:              repo,
			permissionService: permSvc,
			licenseService:    licenseSvc,
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
		tracer.EXPECT().Start(ctx, "service.customFieldService/StageForResource", gomock.Len(0)).Return(ctx, span)

		licenseSvc := mocksvc.NewMockLicenseService(ctrl)
		licenseSvc.EXPECT().HasFeature(ctx, license.FeatureCustomFields).Return(true, nil)

		permSvc := mocksvc.NewMockPermissionService(ctrl)
		permSvc.EXPECT().ListScopeAncestry(ctx, projectID).Return([]model.ID{projectID}, nil)

		repo := mockrepo.NewMockCustomFieldRepository(ctrl)
		repo.EXPECT().ListDefinitions(ctx, []model.ID{projectID}, model.ResourceTypeIssue, false).
			Return([]*model.CustomFieldDefinition{}, nil)

		s := newCustomFieldServiceForTest(customFieldServiceDeps{
			logger:            mocklog.NewMockLogger(ctrl),
			tracer:            tracer,
			repo:              repo,
			permissionService: permSvc,
			licenseService:    licenseSvc,
		})
		require.NoError(t, s.StageForResource(ctx, projectID, issueID, nil))
	})
}

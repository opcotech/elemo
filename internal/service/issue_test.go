package service_test

import (
	"context"
	"testing"
	"time"

	mocklog "github.com/opcotech/elemo/internal/pkg/log/mock"
	mocktrace "github.com/opcotech/elemo/internal/pkg/tracing/mock"
	mockrepo "github.com/opcotech/elemo/internal/repository/mock"
	"github.com/opcotech/elemo/internal/service"
	mocksvc "github.com/opcotech/elemo/internal/service/mock"

	"go.uber.org/mock/gomock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/opcotech/elemo/internal/license"
	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg"
	"github.com/opcotech/elemo/internal/pkg/convert"
	"github.com/opcotech/elemo/internal/pkg/log"
	"github.com/opcotech/elemo/internal/pkg/optional"
	"github.com/opcotech/elemo/internal/pkg/tracing"
	"github.com/opcotech/elemo/internal/repository"
	testModel "github.com/opcotech/elemo/internal/testutil/model"
)

func newCreateIssueOpts() service.CreateIssueOpts {
	return service.CreateIssueOpts{
		Kind:        model.IssueKindStory,
		Title:       "test issue title",
		Description: "test description for issue",
	}
}

type issueServiceDeps struct {
	logger             log.Logger
	tracer             tracing.Tracer
	issueRepo          repository.IssueRepository
	assignmentRepo     repository.AssignmentRepository
	labelRepo          repository.LabelRepository
	permissionService  service.PermissionService
	licenseService     service.LicenseService
	searchService      service.SearchService
	customFieldService service.CustomFieldService
}

func newIssueServiceForTest(deps issueServiceDeps) service.IssueService {
	if deps.issueRepo == nil {
		deps.issueRepo = mockrepo.NewMockIssueRepository(nil)
	}
	if deps.assignmentRepo == nil {
		deps.assignmentRepo = mockrepo.NewMockAssignmentRepository(nil)
	}
	if deps.labelRepo == nil {
		deps.labelRepo = mockrepo.NewMockLabelRepository(nil)
	}
	if deps.permissionService == nil {
		deps.permissionService = mocksvc.NewMockPermissionService(nil)
	}
	if deps.licenseService == nil {
		deps.licenseService = mocksvc.NewMockLicenseService(nil)
	}
	if deps.searchService == nil {
		deps.searchService = mocksvc.NewMockSearchService(nil)
	}
	if deps.customFieldService == nil {
		deps.customFieldService = nopCustomFieldService{}
	}
	var opts []service.Option
	if deps.logger != nil {
		opts = append(opts, service.WithLogger(deps.logger))
	}
	if deps.tracer != nil {
		opts = append(opts, service.WithTracer(deps.tracer))
	}
	svc, err := service.NewIssueService(
		deps.issueRepo,
		deps.assignmentRepo,
		deps.labelRepo,
		deps.permissionService,
		deps.licenseService,
		deps.searchService,
		deps.customFieldService,
		opts...,
	)
	if err != nil {
		panic(err)
	}
	return svc
}

func TestNewIssueService(t *testing.T) {
	tests := []struct {
		name    string
		build   func(ctrl *gomock.Controller) (service.IssueService, error)
		wantErr error
	}{
		{
			name: "new issue service",
			build: func(ctrl *gomock.Controller) (service.IssueService, error) {
				return service.NewIssueService(
					mockrepo.NewMockIssueRepository(nil),
					mockrepo.NewMockAssignmentRepository(nil),
					mockrepo.NewMockLabelRepository(nil),
					mocksvc.NewMockPermissionService(nil),
					mocksvc.NewMockLicenseService(nil),
					mocksvc.NewMockSearchService(nil),
					mocksvc.NewMockCustomFieldService(nil),
					service.WithLogger(mocklog.NewMockLogger(ctrl)),
					service.WithTracer(mocktrace.NewMockTracer(ctrl)),
				)
			},
		},
		{
			name: "new issue service with invalid options",
			build: func(_ *gomock.Controller) (service.IssueService, error) {
				return service.NewIssueService(
					mockrepo.NewMockIssueRepository(nil),
					mockrepo.NewMockAssignmentRepository(nil),
					mockrepo.NewMockLabelRepository(nil),
					mocksvc.NewMockPermissionService(nil),
					mocksvc.NewMockLicenseService(nil),
					mocksvc.NewMockSearchService(nil),
					nil,
					service.WithLogger(nil),
				)
			},
			wantErr: log.ErrNoLogger,
		},
		{
			name: "new issue service with no issue repository",
			build: func(ctrl *gomock.Controller) (service.IssueService, error) {
				return service.NewIssueService(
					nil,
					mockrepo.NewMockAssignmentRepository(nil),
					mockrepo.NewMockLabelRepository(nil),
					mocksvc.NewMockPermissionService(nil),
					mocksvc.NewMockLicenseService(nil),
					mocksvc.NewMockSearchService(nil),
					nil,
					service.WithLogger(mocklog.NewMockLogger(ctrl)),
					service.WithTracer(mocktrace.NewMockTracer(ctrl)),
				)
			},
			wantErr: service.ErrNoIssueRepository,
		},
		{
			name: "new issue service with no assignment repository",
			build: func(ctrl *gomock.Controller) (service.IssueService, error) {
				return service.NewIssueService(
					mockrepo.NewMockIssueRepository(nil),
					nil,
					mockrepo.NewMockLabelRepository(nil),
					mocksvc.NewMockPermissionService(nil),
					mocksvc.NewMockLicenseService(nil),
					mocksvc.NewMockSearchService(nil),
					nil,
					service.WithLogger(mocklog.NewMockLogger(ctrl)),
					service.WithTracer(mocktrace.NewMockTracer(ctrl)),
				)
			},
			wantErr: service.ErrNoAssignmentRepository,
		},
		{
			name: "new issue service with no label repository",
			build: func(ctrl *gomock.Controller) (service.IssueService, error) {
				return service.NewIssueService(
					mockrepo.NewMockIssueRepository(nil),
					mockrepo.NewMockAssignmentRepository(nil),
					nil,
					mocksvc.NewMockPermissionService(nil),
					mocksvc.NewMockLicenseService(nil),
					mocksvc.NewMockSearchService(nil),
					nil,
					service.WithLogger(mocklog.NewMockLogger(ctrl)),
					service.WithTracer(mocktrace.NewMockTracer(ctrl)),
				)
			},
			wantErr: service.ErrNoLabelRepository,
		},
		{
			name: "new issue service with no permission service",
			build: func(ctrl *gomock.Controller) (service.IssueService, error) {
				return service.NewIssueService(
					mockrepo.NewMockIssueRepository(nil),
					mockrepo.NewMockAssignmentRepository(nil),
					mockrepo.NewMockLabelRepository(nil),
					nil,
					mocksvc.NewMockLicenseService(nil),
					mocksvc.NewMockSearchService(nil),
					nil,
					service.WithLogger(mocklog.NewMockLogger(ctrl)),
					service.WithTracer(mocktrace.NewMockTracer(ctrl)),
				)
			},
			wantErr: service.ErrNoPermissionService,
		},
		{
			name: "new issue service with no license service",
			build: func(ctrl *gomock.Controller) (service.IssueService, error) {
				return service.NewIssueService(
					mockrepo.NewMockIssueRepository(nil),
					mockrepo.NewMockAssignmentRepository(nil),
					mockrepo.NewMockLabelRepository(nil),
					mocksvc.NewMockPermissionService(nil),
					nil,
					mocksvc.NewMockSearchService(nil),
					nil,
					service.WithLogger(mocklog.NewMockLogger(ctrl)),
					service.WithTracer(mocktrace.NewMockTracer(ctrl)),
				)
			},
			wantErr: service.ErrNoLicenseService,
		},
		{
			name: "new issue service with no search service",
			build: func(ctrl *gomock.Controller) (service.IssueService, error) {
				return service.NewIssueService(
					mockrepo.NewMockIssueRepository(nil),
					mockrepo.NewMockAssignmentRepository(nil),
					mockrepo.NewMockLabelRepository(nil),
					mocksvc.NewMockPermissionService(nil),
					mocksvc.NewMockLicenseService(nil),
					nil,
					nil,
					service.WithLogger(mocklog.NewMockLogger(ctrl)),
					service.WithTracer(mocktrace.NewMockTracer(ctrl)),
				)
			},
			wantErr: service.ErrNoSearchService,
		},
		{
			name: "new issue service with no custom field service",
			build: func(ctrl *gomock.Controller) (service.IssueService, error) {
				return service.NewIssueService(
					mockrepo.NewMockIssueRepository(nil),
					mockrepo.NewMockAssignmentRepository(nil),
					mockrepo.NewMockLabelRepository(nil),
					mocksvc.NewMockPermissionService(nil),
					mocksvc.NewMockLicenseService(nil),
					mocksvc.NewMockSearchService(nil),
					nil,
					service.WithLogger(mocklog.NewMockLogger(ctrl)),
					service.WithTracer(mocktrace.NewMockTracer(ctrl)),
				)
			},
			wantErr: service.ErrNoCustomFieldService,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			got, err := tt.build(ctrl)
			require.ErrorIs(t, err, tt.wantErr)
			if tt.wantErr == nil {
				assert.NotNil(t, got)
			}
		})
	}
}

func TestIssueService_Create(t *testing.T) {
	projectID := model.MustNewID(model.ResourceTypeProject)
	userID := model.MustNewID(model.ResourceTypeUser)
	parentID := model.MustNewID(model.ResourceTypeIssue)
	opts := newCreateIssueOpts()
	repoIssue := testModel.NewRepositoryIssue(userID)

	type fields struct {
		baseService func(ctrl *gomock.Controller, ctx context.Context, projectID model.ID, opts service.CreateIssueOpts) issueServiceDeps
	}
	type args struct {
		ctx       context.Context
		projectID model.ID
		opts      service.CreateIssueOpts
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr error
	}{
		{
			name: "create issue",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, projectID model.ID, opts service.CreateIssueOpts) issueServiceDeps {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.issueService/Create", gomock.Len(0)).Return(ctx, span)

					creatorID := ctx.Value(pkg.CtxKeyUserID).(model.ID)
					issueRepo := mockrepo.NewMockIssueRepository(ctrl)
					issueRepo.EXPECT().Create(ctx, gomock.Cond(func(o repository.CreateIssueOpts) bool {
						return o.ProjectID == projectID &&
							o.Kind == opts.Kind &&
							o.Title == opts.Title &&
							o.Description == opts.Description &&
							o.Status == model.IssueStatusOpen &&
							o.Priority == model.IssuePriorityNormal &&
							o.Resolution == model.IssueResolutionNone &&
							o.ReportedBy == creatorID
					})).Return(repoIssue, nil)

					permSvc := mocksvc.NewMockPermissionService(ctrl)
					permSvc.EXPECT().BootstrapCreator(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
					permSvc.EXPECT().CtxUserHas(ctx, projectID, gomock.Any()).Return(true, nil)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return issueServiceDeps{
						searchService:     mockSearchIndex(ctrl),
						logger:            mocklog.NewMockLogger(ctrl),
						tracer:            tracer,
						issueRepo:         issueRepo,
						permissionService: permSvc,
						licenseService:    licenseSvc,
					}
				},
			},
			args: args{
				ctx:       context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				projectID: projectID,
				opts:      opts,
			},
		},
		{
			name: "create issue with defaults overridden",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, projectID model.ID, opts service.CreateIssueOpts) issueServiceDeps {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.issueService/Create", gomock.Len(0)).Return(ctx, span)

					creatorID := ctx.Value(pkg.CtxKeyUserID).(model.ID)
					issueRepo := mockrepo.NewMockIssueRepository(ctrl)
					issueRepo.EXPECT().Create(ctx, gomock.Cond(func(o repository.CreateIssueOpts) bool {
						return o.ProjectID == projectID &&
							o.Status == opts.Status &&
							o.Priority == opts.Priority &&
							o.Resolution == opts.Resolution &&
							o.ReportedBy == creatorID
					})).Return(repoIssue, nil)

					permSvc := mocksvc.NewMockPermissionService(ctrl)
					permSvc.EXPECT().BootstrapCreator(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
					permSvc.EXPECT().CtxUserHas(ctx, projectID, gomock.Any()).Return(true, nil)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return issueServiceDeps{
						searchService:     mockSearchIndex(ctrl),
						logger:            mocklog.NewMockLogger(ctrl),
						tracer:            tracer,
						issueRepo:         issueRepo,
						permissionService: permSvc,
						licenseService:    licenseSvc,
					}
				},
			},
			args: args{
				ctx:       context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				projectID: projectID,
				opts: service.CreateIssueOpts{
					Kind:        model.IssueKindBug,
					Title:       "bug title here",
					Description: "detailed bug description",
					Status:      model.IssueStatusInProgress,
					Priority:    model.IssuePriorityHigh,
					Resolution:  model.IssueResolutionNone,
				},
			},
		},
		{
			name: "create issue with license expired",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, _ service.CreateIssueOpts) issueServiceDeps {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.issueService/Create", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(true, nil)

					return issueServiceDeps{
						searchService:  mocksvc.NewMockSearchService(ctrl),
						logger:         mocklog.NewMockLogger(ctrl),
						tracer:         tracer,
						licenseService: licenseSvc,
					}
				},
			},
			args: args{
				ctx:       context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				projectID: projectID,
				opts:      opts,
			},
			wantErr: license.ErrLicenseExpired,
		},
		{
			name: "create issue with no permission",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, projectID model.ID, _ service.CreateIssueOpts) issueServiceDeps {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.issueService/Create", gomock.Len(0)).Return(ctx, span)

					permSvc := mocksvc.NewMockPermissionService(ctrl)
					permSvc.EXPECT().BootstrapCreator(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
					permSvc.EXPECT().CtxUserHas(ctx, projectID, gomock.Any()).Return(false, nil)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return issueServiceDeps{
						searchService:     mocksvc.NewMockSearchService(ctrl),
						logger:            mocklog.NewMockLogger(ctrl),
						tracer:            tracer,
						permissionService: permSvc,
						licenseService:    licenseSvc,
					}
				},
			},
			args: args{
				ctx:       context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				projectID: projectID,
				opts:      opts,
			},
			wantErr: service.ErrNoPermission,
		},
		{
			name: "create issue with invalid project ID",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, _ service.CreateIssueOpts) issueServiceDeps {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.issueService/Create", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return issueServiceDeps{
						searchService:  mocksvc.NewMockSearchService(ctrl),
						logger:         mocklog.NewMockLogger(ctrl),
						tracer:         tracer,
						licenseService: licenseSvc,
					}
				},
			},
			args: args{
				ctx:       context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				projectID: model.ID{},
				opts:      opts,
			},
			wantErr: model.ErrInvalidID,
		},
		{
			name: "create issue with invalid opts",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, _ service.CreateIssueOpts) issueServiceDeps {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.issueService/Create", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return issueServiceDeps{
						searchService:  mocksvc.NewMockSearchService(ctrl),
						logger:         mocklog.NewMockLogger(ctrl),
						tracer:         tracer,
						licenseService: licenseSvc,
					}
				},
			},
			args: args{
				ctx:       context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				projectID: projectID,
				opts:      service.CreateIssueOpts{Kind: model.IssueKindStory, Title: "ab"},
			},
			wantErr: model.ErrInvalidIssueDetails,
		},
		{
			name: "create issue with repository error",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, projectID model.ID, _ service.CreateIssueOpts) issueServiceDeps {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.issueService/Create", gomock.Len(0)).Return(ctx, span)

					issueRepo := mockrepo.NewMockIssueRepository(ctrl)
					issueRepo.EXPECT().Create(ctx, gomock.Any()).Return(nil, repository.ErrIssueCreate)

					permSvc := mocksvc.NewMockPermissionService(ctrl)
					permSvc.EXPECT().BootstrapCreator(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
					permSvc.EXPECT().CtxUserHas(ctx, projectID, gomock.Any()).Return(true, nil)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return issueServiceDeps{
						searchService:     mocksvc.NewMockSearchService(ctrl),
						logger:            mocklog.NewMockLogger(ctrl),
						tracer:            tracer,
						issueRepo:         issueRepo,
						permissionService: permSvc,
						licenseService:    licenseSvc,
					}
				},
			},
			args: args{
				ctx:       context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				projectID: projectID,
				opts:      opts,
			},
			wantErr: repository.ErrIssueCreate,
		},
		{
			name: "create issue with no user ID in context",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, projectID model.ID, _ service.CreateIssueOpts) issueServiceDeps {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.issueService/Create", gomock.Len(0)).Return(ctx, span)

					permSvc := mocksvc.NewMockPermissionService(ctrl)
					permSvc.EXPECT().BootstrapCreator(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
					permSvc.EXPECT().CtxUserHas(ctx, projectID, gomock.Any()).Return(true, nil)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return issueServiceDeps{
						searchService:     mocksvc.NewMockSearchService(ctrl),
						logger:            mocklog.NewMockLogger(ctrl),
						tracer:            tracer,
						permissionService: permSvc,
						licenseService:    licenseSvc,
					}
				},
			},
			args: args{
				ctx:       context.Background(),
				projectID: projectID,
				opts:      opts,
			},
			wantErr: service.ErrNoUser,
		},
		{
			name: "create issue with inaccessible parent",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, projectID model.ID, opts service.CreateIssueOpts) issueServiceDeps {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.issueService/Create", gomock.Len(0)).Return(ctx, span)

					permSvc := mocksvc.NewMockPermissionService(ctrl)
					permSvc.EXPECT().BootstrapCreator(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
					permSvc.EXPECT().CtxUserHas(ctx, projectID, gomock.Any()).Return(true, nil)
					permSvc.EXPECT().CtxUserHas(ctx, *opts.Parent, gomock.Any()).Return(false, nil)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return issueServiceDeps{
						searchService:     mocksvc.NewMockSearchService(ctrl),
						logger:            mocklog.NewMockLogger(ctrl),
						tracer:            tracer,
						permissionService: permSvc,
						licenseService:    licenseSvc,
					}
				},
			},
			args: args{
				ctx:       context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				projectID: projectID,
				opts: service.CreateIssueOpts{
					Kind:        model.IssueKindStory,
					Title:       "test issue title",
					Description: "test description for issue",
					Parent:      &parentID,
				},
			},
			wantErr: service.ErrNoPermission,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			s := newIssueServiceForTest(tt.fields.baseService(ctrl, tt.args.ctx, tt.args.projectID, tt.args.opts))

			_, err := s.Create(tt.args.ctx, tt.args.projectID, tt.args.opts)
			if tt.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestIssueService_CreateCustomFields(t *testing.T) {
	t.Parallel()

	projectID := model.MustNewID(model.ResourceTypeProject)
	userID := model.MustNewID(model.ResourceTypeUser)
	repoIssue := testModel.NewRepositoryIssue(userID)
	opts := newCreateIssueOpts()
	ctx := context.WithValue(context.Background(), pkg.CtxKeyUserID, userID)

	t.Run("stages and commits custom fields", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		span := mocktrace.NewMockSpan(ctrl)
		span.EXPECT().End(gomock.Len(0))
		tracer := mocktrace.NewMockTracer(ctrl)
		tracer.EXPECT().Start(ctx, "service.issueService/Create", gomock.Len(0)).Return(ctx, span)

		issueRepo := mockrepo.NewMockIssueRepository(ctrl)
		issueRepo.EXPECT().Create(ctx, gomock.Any()).Return(repoIssue, nil)

		permSvc := mocksvc.NewMockPermissionService(ctrl)
		permSvc.EXPECT().CtxUserHas(ctx, projectID, model.ActionIssueCreate).Return(true, nil)
		permSvc.EXPECT().BootstrapCreator(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

		licenseSvc := mocksvc.NewMockLicenseService(ctrl)
		licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

		cfSvc := mocksvc.NewMockCustomFieldService(ctrl)
		cfSvc.EXPECT().StageForResource(ctx, projectID, gomock.Any(), opts.CustomFields).Return(nil)
		cfSvc.EXPECT().CommitForResource(ctx, gomock.Any()).Return(nil)

		s := newIssueServiceForTest(issueServiceDeps{
			searchService:      mockSearchIndex(ctrl),
			logger:             mocklog.NewMockLogger(ctrl),
			tracer:             tracer,
			issueRepo:          issueRepo,
			permissionService:  permSvc,
			licenseService:     licenseSvc,
			customFieldService: cfSvc,
		})
		_, err := s.Create(ctx, projectID, opts)
		require.NoError(t, err)
	})

	t.Run("aborts staged values when graph create fails", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		span := mocktrace.NewMockSpan(ctrl)
		span.EXPECT().End(gomock.Len(0))
		tracer := mocktrace.NewMockTracer(ctrl)
		tracer.EXPECT().Start(ctx, "service.issueService/Create", gomock.Len(0)).Return(ctx, span)

		issueRepo := mockrepo.NewMockIssueRepository(ctrl)
		issueRepo.EXPECT().Create(ctx, gomock.Any()).Return(nil, repository.ErrNotFound)

		permSvc := mocksvc.NewMockPermissionService(ctrl)
		permSvc.EXPECT().CtxUserHas(ctx, projectID, model.ActionIssueCreate).Return(true, nil)

		licenseSvc := mocksvc.NewMockLicenseService(ctrl)
		licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

		cfSvc := mocksvc.NewMockCustomFieldService(ctrl)
		cfSvc.EXPECT().StageForResource(ctx, projectID, gomock.Any(), opts.CustomFields).Return(nil)
		cfSvc.EXPECT().AbortForResource(ctx, gomock.Any()).Return(nil)

		s := newIssueServiceForTest(issueServiceDeps{
			searchService:      mocksvc.NewMockSearchService(ctrl),
			logger:             mocklog.NewMockLogger(ctrl),
			tracer:             tracer,
			issueRepo:          issueRepo,
			permissionService:  permSvc,
			licenseService:     licenseSvc,
			customFieldService: cfSvc,
		})
		_, err := s.Create(ctx, projectID, opts)
		require.ErrorIs(t, err, repository.ErrNotFound)
	})

	t.Run("bootstraps even if commit fails", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		span := mocktrace.NewMockSpan(ctrl)
		span.EXPECT().End(gomock.Len(0))
		tracer := mocktrace.NewMockTracer(ctrl)
		tracer.EXPECT().Start(ctx, "service.issueService/Create", gomock.Len(0)).Return(ctx, span)

		issueRepo := mockrepo.NewMockIssueRepository(ctrl)
		issueRepo.EXPECT().Create(ctx, gomock.Any()).Return(repoIssue, nil)

		permSvc := mocksvc.NewMockPermissionService(ctrl)
		permSvc.EXPECT().CtxUserHas(ctx, projectID, model.ActionIssueCreate).Return(true, nil)
		permSvc.EXPECT().BootstrapCreator(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

		licenseSvc := mocksvc.NewMockLicenseService(ctrl)
		licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

		cfSvc := mocksvc.NewMockCustomFieldService(ctrl)
		cfSvc.EXPECT().StageForResource(ctx, projectID, gomock.Any(), opts.CustomFields).Return(nil)
		cfSvc.EXPECT().CommitForResource(ctx, gomock.Any()).Return(assert.AnError)

		logger := mocklog.NewMockLogger(ctrl)
		logger.EXPECT().Warn(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

		s := newIssueServiceForTest(issueServiceDeps{
			searchService:      mockSearchIndex(ctrl),
			logger:             logger,
			tracer:             tracer,
			issueRepo:          issueRepo,
			permissionService:  permSvc,
			licenseService:     licenseSvc,
			customFieldService: cfSvc,
		})
		_, err := s.Create(ctx, projectID, opts)
		require.NoError(t, err)
	})

	t.Run("stage error does not create issue", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		span := mocktrace.NewMockSpan(ctrl)
		span.EXPECT().End(gomock.Len(0))
		tracer := mocktrace.NewMockTracer(ctrl)
		tracer.EXPECT().Start(ctx, "service.issueService/Create", gomock.Len(0)).Return(ctx, span)

		permSvc := mocksvc.NewMockPermissionService(ctrl)
		permSvc.EXPECT().CtxUserHas(ctx, projectID, model.ActionIssueCreate).Return(true, nil)

		licenseSvc := mocksvc.NewMockLicenseService(ctrl)
		licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

		cfSvc := mocksvc.NewMockCustomFieldService(ctrl)
		cfSvc.EXPECT().StageForResource(ctx, projectID, gomock.Any(), opts.CustomFields).Return(assert.AnError)

		s := newIssueServiceForTest(issueServiceDeps{
			searchService:      mocksvc.NewMockSearchService(ctrl),
			logger:             mocklog.NewMockLogger(ctrl),
			tracer:             tracer,
			issueRepo:          mockrepo.NewMockIssueRepository(ctrl),
			permissionService:  permSvc,
			licenseService:     licenseSvc,
			customFieldService: cfSvc,
		})
		_, err := s.Create(ctx, projectID, opts)
		require.ErrorIs(t, err, service.ErrIssueCreate)
		require.ErrorIs(t, err, assert.AnError)
	})

	t.Run("joins abort error when graph create fails", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		span := mocktrace.NewMockSpan(ctrl)
		span.EXPECT().End(gomock.Len(0))
		tracer := mocktrace.NewMockTracer(ctrl)
		tracer.EXPECT().Start(ctx, "service.issueService/Create", gomock.Len(0)).Return(ctx, span)

		issueRepo := mockrepo.NewMockIssueRepository(ctrl)
		issueRepo.EXPECT().Create(ctx, gomock.Any()).Return(nil, repository.ErrNotFound)

		permSvc := mocksvc.NewMockPermissionService(ctrl)
		permSvc.EXPECT().CtxUserHas(ctx, projectID, model.ActionIssueCreate).Return(true, nil)

		licenseSvc := mocksvc.NewMockLicenseService(ctrl)
		licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

		cfSvc := mocksvc.NewMockCustomFieldService(ctrl)
		cfSvc.EXPECT().StageForResource(ctx, projectID, gomock.Any(), opts.CustomFields).Return(nil)
		cfSvc.EXPECT().AbortForResource(ctx, gomock.Any()).Return(assert.AnError)

		s := newIssueServiceForTest(issueServiceDeps{
			searchService:      mocksvc.NewMockSearchService(ctrl),
			logger:             mocklog.NewMockLogger(ctrl),
			tracer:             tracer,
			issueRepo:          issueRepo,
			permissionService:  permSvc,
			licenseService:     licenseSvc,
			customFieldService: cfSvc,
		})
		_, err := s.Create(ctx, projectID, opts)
		require.ErrorIs(t, err, service.ErrIssueCreate)
		require.ErrorIs(t, err, repository.ErrNotFound)
		require.ErrorIs(t, err, assert.AnError)
	})
}

func TestIssueService_Get(t *testing.T) {
	issueID := model.MustNewID(model.ResourceTypeIssue)
	userID := model.MustNewID(model.ResourceTypeUser)
	repoIssue := testModel.NewRepositoryIssue(userID)
	repoIssue.ID = issueID
	want := service.IssueFromRepository(repoIssue)

	type fields struct {
		baseService func(ctrl *gomock.Controller, ctx context.Context, id model.ID) issueServiceDeps
	}
	type args struct {
		ctx context.Context
		id  model.ID
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *service.Issue
		wantErr error
	}{
		{
			name: "get issue",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) issueServiceDeps {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.issueService/Get", gomock.Len(0)).Return(ctx, span)

					issueRepo := mockrepo.NewMockIssueRepository(ctrl)
					issueRepo.EXPECT().Get(ctx, id, repository.IssueDetailProjection()).Return(repoIssue, nil)

					permSvc := mocksvc.NewMockPermissionService(ctrl)
					permSvc.EXPECT().BootstrapCreator(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
					permSvc.EXPECT().CtxUserHas(ctx, id, gomock.Any()).Return(true, nil)

					return issueServiceDeps{
						searchService:     mocksvc.NewMockSearchService(ctrl),
						logger:            mocklog.NewMockLogger(ctrl),
						tracer:            tracer,
						issueRepo:         issueRepo,
						permissionService: permSvc,
					}
				},
			},
			args: args{
				ctx: context.Background(),
				id:  issueID,
			},
			want: want,
		},
		{
			name: "get issue with no permission",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) issueServiceDeps {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.issueService/Get", gomock.Len(0)).Return(ctx, span)

					permSvc := mocksvc.NewMockPermissionService(ctrl)
					permSvc.EXPECT().BootstrapCreator(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
					permSvc.EXPECT().CtxUserHas(ctx, id, gomock.Any()).Return(false, nil)

					return issueServiceDeps{
						searchService:     mocksvc.NewMockSearchService(ctrl),
						logger:            mocklog.NewMockLogger(ctrl),
						tracer:            tracer,
						permissionService: permSvc,
					}
				},
			},
			args: args{
				ctx: context.Background(),
				id:  issueID,
			},
			wantErr: service.ErrNoPermission,
		},
		{
			name: "get issue with invalid ID",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID) issueServiceDeps {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.issueService/Get", gomock.Len(0)).Return(ctx, span)

					return issueServiceDeps{
						searchService: mocksvc.NewMockSearchService(ctrl),
						logger:        mocklog.NewMockLogger(ctrl),
						tracer:        tracer,
					}
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.ID{},
			},
			wantErr: model.ErrInvalidID,
		},
		{
			name: "get issue with repository error",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) issueServiceDeps {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.issueService/Get", gomock.Len(0)).Return(ctx, span)

					issueRepo := mockrepo.NewMockIssueRepository(ctrl)
					issueRepo.EXPECT().Get(ctx, id, repository.IssueDetailProjection()).Return(nil, repository.ErrIssueRead)

					permSvc := mocksvc.NewMockPermissionService(ctrl)
					permSvc.EXPECT().BootstrapCreator(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
					permSvc.EXPECT().CtxUserHas(ctx, id, gomock.Any()).Return(true, nil)

					return issueServiceDeps{
						searchService:     mocksvc.NewMockSearchService(ctrl),
						logger:            mocklog.NewMockLogger(ctrl),
						tracer:            tracer,
						issueRepo:         issueRepo,
						permissionService: permSvc,
					}
				},
			},
			args: args{
				ctx: context.Background(),
				id:  issueID,
			},
			wantErr: repository.ErrIssueRead,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			s := newIssueServiceForTest(tt.fields.baseService(ctrl, tt.args.ctx, tt.args.id))

			got, err := s.Get(tt.args.ctx, tt.args.id)
			if tt.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestIssueService_GetByKey(t *testing.T) {
	issueID := model.MustNewID(model.ResourceTypeIssue)
	namespaceID := model.MustNewID(model.ResourceTypeNamespace)
	userID := model.MustNewID(model.ResourceTypeUser)
	repoIssue := testModel.NewRepositoryIssue(userID)
	repoIssue.ID = issueID
	repoIssue.Key = "ENG-42"
	repoIssue.NumericID = 42
	want := service.IssueFromRepository(repoIssue)

	type fields struct {
		baseService func(ctrl *gomock.Controller, ctx context.Context, namespaceID model.ID, key string) issueServiceDeps
	}
	type args struct {
		ctx         context.Context
		namespaceID model.ID
		key         string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *service.Issue
		wantErr error
	}{
		{
			name: "get issue by key",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, namespaceID model.ID, key string) issueServiceDeps {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.issueService/GetByKey", gomock.Len(0)).Return(ctx, span)

					issueRepo := mockrepo.NewMockIssueRepository(ctrl)
					issueRepo.EXPECT().GetByKey(ctx, namespaceID, key, repository.IssueDetailProjection()).Return(repoIssue, nil)

					permSvc := mocksvc.NewMockPermissionService(ctrl)
					permSvc.EXPECT().BootstrapCreator(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
					permSvc.EXPECT().CtxUserHas(ctx, issueID, gomock.Any()).Return(true, nil)

					return issueServiceDeps{
						searchService:     mocksvc.NewMockSearchService(ctrl),
						logger:            mocklog.NewMockLogger(ctrl),
						tracer:            tracer,
						issueRepo:         issueRepo,
						permissionService: permSvc,
					}
				},
			},
			args: args{
				ctx:         context.Background(),
				namespaceID: namespaceID,
				key:         "ENG-42",
			},
			want: want,
		},
		{
			name: "get issue by key with no permission",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, namespaceID model.ID, key string) issueServiceDeps {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.issueService/GetByKey", gomock.Len(0)).Return(ctx, span)

					issueRepo := mockrepo.NewMockIssueRepository(ctrl)
					issueRepo.EXPECT().GetByKey(ctx, namespaceID, key, repository.IssueDetailProjection()).Return(repoIssue, nil)

					permSvc := mocksvc.NewMockPermissionService(ctrl)
					permSvc.EXPECT().BootstrapCreator(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
					permSvc.EXPECT().CtxUserHas(ctx, issueID, gomock.Any()).Return(false, nil)

					return issueServiceDeps{
						searchService:     mocksvc.NewMockSearchService(ctrl),
						logger:            mocklog.NewMockLogger(ctrl),
						tracer:            tracer,
						issueRepo:         issueRepo,
						permissionService: permSvc,
					}
				},
			},
			args: args{
				ctx:         context.Background(),
				namespaceID: namespaceID,
				key:         "ENG-42",
			},
			wantErr: service.ErrNoPermission,
		},
		{
			name: "get issue with invalid namespace ID",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, _ string) issueServiceDeps {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.issueService/GetByKey", gomock.Len(0)).Return(ctx, span)

					return issueServiceDeps{
						searchService: mocksvc.NewMockSearchService(ctrl),
						logger:        mocklog.NewMockLogger(ctrl),
						tracer:        tracer,
					}
				},
			},
			args: args{
				ctx:         context.Background(),
				namespaceID: model.ID{},
				key:         "ENG-42",
			},
			wantErr: model.ErrInvalidID,
		},
		{
			name: "get issue with invalid key",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, _ string) issueServiceDeps {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.issueService/GetByKey", gomock.Len(0)).Return(ctx, span)

					return issueServiceDeps{
						searchService: mocksvc.NewMockSearchService(ctrl),
						logger:        mocklog.NewMockLogger(ctrl),
						tracer:        tracer,
					}
				},
			},
			args: args{
				ctx:         context.Background(),
				namespaceID: namespaceID,
				key:         "bad",
			},
			wantErr: model.ErrInvalidIssueDetails,
		},
		{
			name: "get issue by key with repository error",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, namespaceID model.ID, key string) issueServiceDeps {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.issueService/GetByKey", gomock.Len(0)).Return(ctx, span)

					issueRepo := mockrepo.NewMockIssueRepository(ctrl)
					issueRepo.EXPECT().GetByKey(ctx, namespaceID, key, repository.IssueDetailProjection()).Return(nil, repository.ErrIssueRead)

					return issueServiceDeps{
						searchService: mocksvc.NewMockSearchService(ctrl),
						logger:        mocklog.NewMockLogger(ctrl),
						tracer:        tracer,
						issueRepo:     issueRepo,
					}
				},
			},
			args: args{
				ctx:         context.Background(),
				namespaceID: namespaceID,
				key:         "ENG-42",
			},
			wantErr: repository.ErrIssueRead,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			s := newIssueServiceForTest(tt.fields.baseService(ctrl, tt.args.ctx, tt.args.namespaceID, tt.args.key))

			got, err := s.GetByKey(tt.args.ctx, tt.args.namespaceID, tt.args.key)
			if tt.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestListGrantCoversRoot(t *testing.T) {
	t.Parallel()

	root := model.MustNewID(model.ResourceTypeProject)
	namespace := model.MustNewID(model.ResourceTypeNamespace)
	org := model.MustNewID(model.ResourceTypeOrganization)
	other := model.MustNewID(model.ResourceTypeProject)
	ancestry := []model.ID{root, namespace, org}

	assert.True(t, service.ListGrantCoversRoot([]model.ID{org}, ancestry))
	assert.True(t, service.ListGrantCoversRoot([]model.ID{root}, ancestry))
	assert.False(t, service.ListGrantCoversRoot([]model.ID{other}, ancestry))
	assert.False(t, service.ListGrantCoversRoot(nil, ancestry))
	assert.False(t, service.ListGrantCoversRoot([]model.ID{org}, nil))
}

func TestIssueService_List(t *testing.T) {
	projectID := model.MustNewID(model.ResourceTypeProject)
	userID := model.MustNewID(model.ResourceTypeUser)
	scopeIDs := []model.ID{projectID}
	repoIssueA := testModel.NewRepositoryIssue(userID)
	repoIssueB := testModel.NewRepositoryIssue(userID)
	repoIssues := []*repository.PartialIssue{
		{
			ID:          repoIssueA.ID,
			Key:         repoIssueA.Key,
			NumericID:   repoIssueA.NumericID,
			Parent:      repoIssueA.Parent,
			Kind:        repoIssueA.Kind,
			Title:       repoIssueA.Title,
			Status:      repoIssueA.Status,
			Priority:    repoIssueA.Priority,
			Assignments: repoIssueA.Assignments,
			Labels:      repoIssueA.Labels,
			DueDate:     repoIssueA.DueDate,
		},
		{
			ID:          repoIssueB.ID,
			Key:         repoIssueB.Key,
			NumericID:   repoIssueB.NumericID,
			Parent:      repoIssueB.Parent,
			Kind:        repoIssueB.Kind,
			Title:       repoIssueB.Title,
			Status:      repoIssueB.Status,
			Priority:    repoIssueB.Priority,
			Assignments: repoIssueB.Assignments,
			Labels:      repoIssueB.Labels,
			DueDate:     repoIssueB.DueDate,
		},
	}
	want := service.Page[*service.PartialIssue]{Items: service.PartialIssuesFromRepository(repoIssues)}

	type fields struct {
		baseService func(ctrl *gomock.Controller, ctx context.Context, projectID model.ID) issueServiceDeps
	}
	type args struct {
		ctx       context.Context
		projectID model.ID
		page      service.CursorPage
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    service.Page[*service.PartialIssue]
		wantErr error
	}{
		{
			name: "list issues",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, projectID model.ID) issueServiceDeps {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.issueService/List", gomock.Len(0)).Return(ctx, span)

					issueRepo := mockrepo.NewMockIssueRepository(ctrl)
					issueRepo.EXPECT().ListForProject(ctx, repository.IssueListQuery{
						ProjectID:  projectID,
						ActorID:    userID,
						Action:     model.ActionIssueRead,
						ScopeIDs:   nil,
						SortField:  repository.IssueListSortFieldRank,
						Page:       repository.CursorPage{Size: 10},
						Order:      repository.SortDirectionAsc,
						Projection: repository.IssueListForProjectProjection(),
					}).Return(repository.Page[*repository.PartialIssue]{Items: repoIssues}, nil)

					permSvc := mocksvc.NewMockPermissionService(ctrl)
					permSvc.EXPECT().BootstrapCreator(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
					permSvc.EXPECT().CtxUserListGrantScopes(ctx, model.ActionIssueRead).Return(scopeIDs, nil)
					permSvc.EXPECT().ListScopeAncestry(ctx, projectID).Return([]model.ID{projectID}, nil)

					return issueServiceDeps{
						searchService:     mocksvc.NewMockSearchService(ctrl),
						logger:            mocklog.NewMockLogger(ctrl),
						tracer:            tracer,
						issueRepo:         issueRepo,
						permissionService: permSvc,
					}
				},
			},
			args: args{
				ctx:       context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				projectID: projectID,
				page:      service.CursorPage{Size: 10},
			},
			want: want,
		},
		{
			name: "list issues with grant that does not cover the project",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, projectID model.ID) issueServiceDeps {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.issueService/List", gomock.Len(0)).Return(ctx, span)

					otherProjectID := model.MustNewID(model.ResourceTypeProject)
					issueRepo := mockrepo.NewMockIssueRepository(ctrl)
					issueRepo.EXPECT().ListForProject(ctx, repository.IssueListQuery{
						ProjectID:  projectID,
						ActorID:    userID,
						Action:     model.ActionIssueRead,
						ScopeIDs:   []model.ID{otherProjectID},
						SortField:  repository.IssueListSortFieldRank,
						Page:       repository.CursorPage{Size: 10},
						Order:      repository.SortDirectionAsc,
						Projection: repository.IssueListForProjectProjection(),
					}).Return(repository.Page[*repository.PartialIssue]{Items: repoIssues}, nil)

					permSvc := mocksvc.NewMockPermissionService(ctrl)
					permSvc.EXPECT().BootstrapCreator(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
					permSvc.EXPECT().CtxUserListGrantScopes(ctx, model.ActionIssueRead).Return([]model.ID{otherProjectID}, nil)
					permSvc.EXPECT().ListScopeAncestry(ctx, projectID).Return([]model.ID{projectID}, nil)

					return issueServiceDeps{
						searchService:     mocksvc.NewMockSearchService(ctrl),
						logger:            mocklog.NewMockLogger(ctrl),
						tracer:            tracer,
						issueRepo:         issueRepo,
						permissionService: permSvc,
					}
				},
			},
			args: args{
				ctx:       context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				projectID: projectID,
				page:      service.CursorPage{Size: 10},
			},
			want: want,
		},
		{
			name: "list issues with no user",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID) issueServiceDeps {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.issueService/List", gomock.Len(0)).Return(ctx, span)

					return issueServiceDeps{
						searchService: mocksvc.NewMockSearchService(ctrl),
						logger:        mocklog.NewMockLogger(ctrl),
						tracer:        tracer,
					}
				},
			},
			args: args{
				ctx:       context.Background(),
				projectID: projectID,
				page:      service.CursorPage{Size: 10},
			},
			wantErr: service.ErrNoUser,
		},
		{
			name: "list issues with invalid project ID",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID) issueServiceDeps {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.issueService/List", gomock.Len(0)).Return(ctx, span)

					return issueServiceDeps{
						searchService: mocksvc.NewMockSearchService(ctrl),
						logger:        mocklog.NewMockLogger(ctrl),
						tracer:        tracer,
					}
				},
			},
			args: args{
				ctx:       context.Background(),
				projectID: model.ID{},
				page:      service.CursorPage{Size: 10},
			},
			wantErr: model.ErrInvalidID,
		},
		{
			name: "list issues with invalid page size",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID) issueServiceDeps {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.issueService/List", gomock.Len(0)).Return(ctx, span)

					return issueServiceDeps{
						searchService: mocksvc.NewMockSearchService(ctrl),
						logger:        mocklog.NewMockLogger(ctrl),
						tracer:        tracer,
					}
				},
			},
			args: args{
				ctx:       context.Background(),
				projectID: projectID,
				page:      service.CursorPage{Size: -1},
			},
			wantErr: repository.ErrInvalidPageSize,
		},
		{
			name: "list issues with repository error",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, projectID model.ID) issueServiceDeps {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.issueService/List", gomock.Len(0)).Return(ctx, span)

					issueRepo := mockrepo.NewMockIssueRepository(ctrl)
					issueRepo.EXPECT().ListForProject(ctx, repository.IssueListQuery{
						ProjectID:  projectID,
						ActorID:    userID,
						Action:     model.ActionIssueRead,
						ScopeIDs:   nil,
						SortField:  repository.IssueListSortFieldRank,
						Page:       repository.CursorPage{Size: 10},
						Order:      repository.SortDirectionAsc,
						Projection: repository.IssueListForProjectProjection(),
					}).Return(repository.Page[*repository.PartialIssue]{}, repository.ErrIssueRead)

					permSvc := mocksvc.NewMockPermissionService(ctrl)
					permSvc.EXPECT().BootstrapCreator(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
					permSvc.EXPECT().CtxUserListGrantScopes(ctx, model.ActionIssueRead).Return(scopeIDs, nil)
					permSvc.EXPECT().ListScopeAncestry(ctx, projectID).Return([]model.ID{projectID}, nil)

					return issueServiceDeps{
						searchService:     mocksvc.NewMockSearchService(ctrl),
						logger:            mocklog.NewMockLogger(ctrl),
						tracer:            tracer,
						issueRepo:         issueRepo,
						permissionService: permSvc,
					}
				},
			},
			args: args{
				ctx:       context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				projectID: projectID,
				page:      service.CursorPage{Size: 10},
			},
			wantErr: repository.ErrIssueRead,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			s := newIssueServiceForTest(tt.fields.baseService(ctrl, tt.args.ctx, tt.args.projectID))

			got, err := s.List(tt.args.ctx, tt.args.projectID, tt.args.page, service.IssueListOptions{})
			if tt.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestIssueService_ListByNamespace(t *testing.T) {
	namespaceID := model.MustNewID(model.ResourceTypeNamespace)
	userID := model.MustNewID(model.ResourceTypeUser)
	scopeIDs := []model.ID{namespaceID}
	repoIssueA := testModel.NewRepositoryIssue(userID)
	repoIssueB := testModel.NewRepositoryIssue(userID)
	repoIssues := []*repository.PartialIssue{
		{
			ID:          repoIssueA.ID,
			Key:         repoIssueA.Key,
			NumericID:   repoIssueA.NumericID,
			Parent:      repoIssueA.Parent,
			Kind:        repoIssueA.Kind,
			Title:       repoIssueA.Title,
			Status:      repoIssueA.Status,
			Priority:    repoIssueA.Priority,
			Assignments: repoIssueA.Assignments,
			Labels:      repoIssueA.Labels,
			DueDate:     repoIssueA.DueDate,
		},
		{
			ID:          repoIssueB.ID,
			Key:         repoIssueB.Key,
			NumericID:   repoIssueB.NumericID,
			Parent:      repoIssueB.Parent,
			Kind:        repoIssueB.Kind,
			Title:       repoIssueB.Title,
			Status:      repoIssueB.Status,
			Priority:    repoIssueB.Priority,
			Assignments: repoIssueB.Assignments,
			Labels:      repoIssueB.Labels,
			DueDate:     repoIssueB.DueDate,
		},
	}
	want := service.Page[*service.PartialIssue]{Items: service.PartialIssuesFromRepository(repoIssues)}

	type fields struct {
		baseService func(ctrl *gomock.Controller, ctx context.Context, namespaceID model.ID) issueServiceDeps
	}
	type args struct {
		ctx         context.Context
		namespaceID model.ID
		page        service.CursorPage
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    service.Page[*service.PartialIssue]
		wantErr error
	}{
		{
			name: "list issues",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, namespaceID model.ID) issueServiceDeps {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.issueService/ListByNamespace", gomock.Len(0)).Return(ctx, span)

					issueRepo := mockrepo.NewMockIssueRepository(ctrl)
					issueRepo.EXPECT().ListForNamespace(ctx, repository.IssueListForNamespaceQuery{
						NamespaceID: namespaceID,
						ActorID:     userID,
						Action:      model.ActionIssueRead,
						ScopeIDs:    nil,
						SortField:   repository.IssueListSortFieldRank,
						Page:        repository.CursorPage{Size: 10},
						Order:       repository.SortDirectionAsc,
						Projection:  repository.IssueListForNamespaceProjection(),
					}).Return(repository.Page[*repository.PartialIssue]{Items: repoIssues}, nil)

					permSvc := mocksvc.NewMockPermissionService(ctrl)
					permSvc.EXPECT().BootstrapCreator(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
					permSvc.EXPECT().CtxUserListGrantScopes(ctx, model.ActionIssueRead).Return(scopeIDs, nil)
					permSvc.EXPECT().ListScopeAncestry(ctx, namespaceID).Return([]model.ID{namespaceID}, nil)

					return issueServiceDeps{
						searchService:     mocksvc.NewMockSearchService(ctrl),
						logger:            mocklog.NewMockLogger(ctrl),
						tracer:            tracer,
						issueRepo:         issueRepo,
						permissionService: permSvc,
					}
				},
			},
			args: args{
				ctx:         context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				namespaceID: namespaceID,
				page:        service.CursorPage{Size: 10},
			},
			want: want,
		},
		{
			name: "list issues with no user",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID) issueServiceDeps {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.issueService/ListByNamespace", gomock.Len(0)).Return(ctx, span)

					return issueServiceDeps{
						searchService: mocksvc.NewMockSearchService(ctrl),
						logger:        mocklog.NewMockLogger(ctrl),
						tracer:        tracer,
					}
				},
			},
			args: args{
				ctx:         context.Background(),
				namespaceID: namespaceID,
				page:        service.CursorPage{Size: 10},
			},
			wantErr: service.ErrNoUser,
		},
		{
			name: "list issues with invalid namespace ID",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID) issueServiceDeps {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.issueService/ListByNamespace", gomock.Len(0)).Return(ctx, span)

					return issueServiceDeps{
						searchService: mocksvc.NewMockSearchService(ctrl),
						logger:        mocklog.NewMockLogger(ctrl),
						tracer:        tracer,
					}
				},
			},
			args: args{
				ctx:         context.Background(),
				namespaceID: model.ID{},
				page:        service.CursorPage{Size: 10},
			},
			wantErr: model.ErrInvalidID,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			s := newIssueServiceForTest(tt.fields.baseService(ctrl, tt.args.ctx, tt.args.namespaceID))

			got, err := s.ListByNamespace(tt.args.ctx, tt.args.namespaceID, tt.args.page, service.IssueListOptions{})
			if tt.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestIssueService_ListByUser(t *testing.T) {
	userID := model.MustNewID(model.ResourceTypeUser)
	otherUserID := model.MustNewID(model.ResourceTypeUser)
	scopeIDs := []model.ID{model.MustNewID(model.ResourceTypeProject)}
	repoIssueA := testModel.NewRepositoryIssue(userID)
	repoIssueB := testModel.NewRepositoryIssue(userID)
	repoIssues := []*repository.PartialIssue{
		{
			ID:          repoIssueA.ID,
			Key:         repoIssueA.Key,
			NumericID:   repoIssueA.NumericID,
			Parent:      repoIssueA.Parent,
			Kind:        repoIssueA.Kind,
			Title:       repoIssueA.Title,
			Status:      repoIssueA.Status,
			Priority:    repoIssueA.Priority,
			Assignments: repoIssueA.Assignments,
			Labels:      repoIssueA.Labels,
			DueDate:     repoIssueA.DueDate,
		},
		{
			ID:          repoIssueB.ID,
			Key:         repoIssueB.Key,
			NumericID:   repoIssueB.NumericID,
			Parent:      repoIssueB.Parent,
			Kind:        repoIssueB.Kind,
			Title:       repoIssueB.Title,
			Status:      repoIssueB.Status,
			Priority:    repoIssueB.Priority,
			Assignments: repoIssueB.Assignments,
			Labels:      repoIssueB.Labels,
			DueDate:     repoIssueB.DueDate,
		},
	}
	want := service.Page[*service.PartialIssue]{Items: service.PartialIssuesFromRepository(repoIssues)}

	type fields struct {
		baseService func(ctrl *gomock.Controller, ctx context.Context, userID model.ID) issueServiceDeps
	}
	type args struct {
		ctx    context.Context
		userID model.ID
		page   service.CursorPage
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    service.Page[*service.PartialIssue]
		wantErr error
	}{
		{
			name: "list own issues",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, userID model.ID) issueServiceDeps {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.issueService/ListByUser", gomock.Len(0)).Return(ctx, span)

					issueRepo := mockrepo.NewMockIssueRepository(ctrl)
					issueRepo.EXPECT().ListForUser(ctx, repository.IssueListForUserQuery{
						UserID:     userID,
						ActorID:    userID,
						Action:     model.ActionIssueRead,
						ScopeIDs:   scopeIDs,
						SortField:  repository.IssueListSortFieldRank,
						Page:       repository.CursorPage{Size: 10},
						Order:      repository.SortDirectionAsc,
						Projection: repository.IssueListForUserProjection(),
					}).Return(repository.Page[*repository.PartialIssue]{Items: repoIssues}, nil)

					permSvc := mocksvc.NewMockPermissionService(ctrl)
					permSvc.EXPECT().CtxUserListGrantScopes(ctx, model.ActionIssueRead).Return(scopeIDs, nil)

					return issueServiceDeps{
						searchService:     mocksvc.NewMockSearchService(ctrl),
						logger:            mocklog.NewMockLogger(ctrl),
						tracer:            tracer,
						issueRepo:         issueRepo,
						permissionService: permSvc,
					}
				},
			},
			args: args{
				ctx:    context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				userID: userID,
				page:   service.CursorPage{Size: 10},
			},
			want: want,
		},
		{
			name: "list another user's issues",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID) issueServiceDeps {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.issueService/ListByUser", gomock.Len(0)).Return(ctx, span)

					return issueServiceDeps{
						searchService: mocksvc.NewMockSearchService(ctrl),
						logger:        mocklog.NewMockLogger(ctrl),
						tracer:        tracer,
					}
				},
			},
			args: args{
				ctx:    context.WithValue(context.Background(), pkg.CtxKeyUserID, otherUserID),
				userID: userID,
				page:   service.CursorPage{Size: 10},
			},
			wantErr: service.ErrNoPermission,
		},
		{
			name: "list another user's issues with no permission",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID) issueServiceDeps {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.issueService/ListByUser", gomock.Len(0)).Return(ctx, span)

					return issueServiceDeps{
						searchService: mocksvc.NewMockSearchService(ctrl),
						logger:        mocklog.NewMockLogger(ctrl),
						tracer:        tracer,
					}
				},
			},
			args: args{
				ctx:    context.WithValue(context.Background(), pkg.CtxKeyUserID, otherUserID),
				userID: userID,
				page:   service.CursorPage{Size: 10},
			},
			wantErr: service.ErrNoPermission,
		},
		{
			name: "list issues with no user",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID) issueServiceDeps {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.issueService/ListByUser", gomock.Len(0)).Return(ctx, span)

					return issueServiceDeps{
						searchService: mocksvc.NewMockSearchService(ctrl),
						logger:        mocklog.NewMockLogger(ctrl),
						tracer:        tracer,
					}
				},
			},
			args: args{
				ctx:    context.Background(),
				userID: userID,
				page:   service.CursorPage{Size: 10},
			},
			wantErr: service.ErrNoUser,
		},
		{
			name: "list issues with invalid user ID",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID) issueServiceDeps {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.issueService/ListByUser", gomock.Len(0)).Return(ctx, span)

					return issueServiceDeps{
						searchService: mocksvc.NewMockSearchService(ctrl),
						logger:        mocklog.NewMockLogger(ctrl),
						tracer:        tracer,
					}
				},
			},
			args: args{
				ctx:    context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				userID: model.ID{},
				page:   service.CursorPage{Size: 10},
			},
			wantErr: model.ErrInvalidID,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			s := newIssueServiceForTest(tt.fields.baseService(ctrl, tt.args.ctx, tt.args.userID))

			got, err := s.ListByUser(tt.args.ctx, tt.args.userID, tt.args.page, service.IssueListOptions{})
			if tt.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestIssueService_Update(t *testing.T) {
	issueID := model.MustNewID(model.ResourceTypeIssue)
	userID := model.MustNewID(model.ResourceTypeUser)
	repoIssue := testModel.NewRepositoryIssue(userID)
	repoIssue.ID = issueID
	repoIssue.Title = "updated title"
	want := service.IssueFromRepository(repoIssue)
	opts := service.UpdateIssueOpts{
		Title: optional.Some("updated title"),
	}

	type fields struct {
		baseService func(ctrl *gomock.Controller, ctx context.Context, id model.ID, opts service.UpdateIssueOpts) issueServiceDeps
	}
	type args struct {
		ctx  context.Context
		id   model.ID
		opts service.UpdateIssueOpts
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *service.Issue
		wantErr error
	}{
		{
			name: "update issue",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, opts service.UpdateIssueOpts) issueServiceDeps {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.issueService/Update", gomock.Len(0)).Return(ctx, span)

					issueRepo := mockrepo.NewMockIssueRepository(ctrl)
					issueRepo.EXPECT().Update(ctx, id, repository.UpdateIssueOpts{
						Title: opts.Title,
					}, repository.IssueDetailProjection()).Return(repoIssue, nil)

					permSvc := mocksvc.NewMockPermissionService(ctrl)
					permSvc.EXPECT().BootstrapCreator(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
					permSvc.EXPECT().CtxUserHas(ctx, id, gomock.Any()).Return(true, nil)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return issueServiceDeps{
						searchService:     mockSearchIndex(ctrl),
						logger:            mocklog.NewMockLogger(ctrl),
						tracer:            tracer,
						issueRepo:         issueRepo,
						permissionService: permSvc,
						licenseService:    licenseSvc,
					}
				},
			},
			args: args{
				ctx:  context.Background(),
				id:   issueID,
				opts: opts,
			},
			want: want,
		},
		{
			name: "update issue with license expired",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, _ service.UpdateIssueOpts) issueServiceDeps {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.issueService/Update", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(true, nil)

					return issueServiceDeps{
						searchService:  mocksvc.NewMockSearchService(ctrl),
						logger:         mocklog.NewMockLogger(ctrl),
						tracer:         tracer,
						licenseService: licenseSvc,
					}
				},
			},
			args: args{
				ctx:  context.Background(),
				id:   issueID,
				opts: opts,
			},
			wantErr: license.ErrLicenseExpired,
		},
		{
			name: "update issue with no permission",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ service.UpdateIssueOpts) issueServiceDeps {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.issueService/Update", gomock.Len(0)).Return(ctx, span)

					permSvc := mocksvc.NewMockPermissionService(ctrl)
					permSvc.EXPECT().BootstrapCreator(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
					permSvc.EXPECT().CtxUserHas(ctx, id, gomock.Any()).Return(false, nil)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return issueServiceDeps{
						searchService:     mocksvc.NewMockSearchService(ctrl),
						logger:            mocklog.NewMockLogger(ctrl),
						tracer:            tracer,
						permissionService: permSvc,
						licenseService:    licenseSvc,
					}
				},
			},
			args: args{
				ctx:  context.Background(),
				id:   issueID,
				opts: opts,
			},
			wantErr: service.ErrNoPermission,
		},
		{
			name: "update issue with invalid ID",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, _ service.UpdateIssueOpts) issueServiceDeps {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.issueService/Update", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return issueServiceDeps{
						searchService:  mocksvc.NewMockSearchService(ctrl),
						logger:         mocklog.NewMockLogger(ctrl),
						tracer:         tracer,
						licenseService: licenseSvc,
					}
				},
			},
			args: args{
				ctx:  context.Background(),
				id:   model.ID{},
				opts: opts,
			},
			wantErr: model.ErrInvalidID,
		},
		{
			name: "update issue with repository error",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, opts service.UpdateIssueOpts) issueServiceDeps {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.issueService/Update", gomock.Len(0)).Return(ctx, span)

					issueRepo := mockrepo.NewMockIssueRepository(ctrl)
					issueRepo.EXPECT().Update(ctx, id, repository.UpdateIssueOpts{
						Title: opts.Title,
					}, repository.IssueDetailProjection()).Return(nil, repository.ErrIssueUpdate)

					permSvc := mocksvc.NewMockPermissionService(ctrl)
					permSvc.EXPECT().BootstrapCreator(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
					permSvc.EXPECT().CtxUserHas(ctx, id, gomock.Any()).Return(true, nil)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return issueServiceDeps{
						searchService:     mocksvc.NewMockSearchService(ctrl),
						logger:            mocklog.NewMockLogger(ctrl),
						tracer:            tracer,
						issueRepo:         issueRepo,
						permissionService: permSvc,
						licenseService:    licenseSvc,
					}
				},
			},
			args: args{
				ctx:  context.Background(),
				id:   issueID,
				opts: opts,
			},
			wantErr: repository.ErrIssueUpdate,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			s := newIssueServiceForTest(tt.fields.baseService(ctrl, tt.args.ctx, tt.args.id, tt.args.opts))

			got, err := s.Update(tt.args.ctx, tt.args.id, tt.args.opts)
			if tt.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestIssueService_UpdateAssignments(t *testing.T) {
	issueID := model.MustNewID(model.ResourceTypeIssue)
	userID := model.MustNewID(model.ResourceTypeUser)
	assigneeID := model.MustNewID(model.ResourceTypeUser)
	reviewerID := model.MustNewID(model.ResourceTypeUser)
	repoIssue := testModel.NewRepositoryIssue(userID)
	repoIssue.ID = issueID

	t.Run("sync assignee edges", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ctx := context.Background()
		span := mocktrace.NewMockSpan(ctrl)
		span.EXPECT().End(gomock.Len(0))
		tracer := mocktrace.NewMockTracer(ctrl)
		tracer.EXPECT().Start(ctx, "service.issueService/Update", gomock.Len(0)).Return(ctx, span)

		existingAssignmentID := model.MustNewID(model.ResourceTypeAssignment)
		staleAssigneeID := model.MustNewID(model.ResourceTypeUser)

		issueRepo := mockrepo.NewMockIssueRepository(ctrl)
		issueRepo.EXPECT().Update(ctx, issueID, repository.UpdateIssueOpts{}, repository.IssueDetailProjection()).Return(repoIssue, nil)
		updated := *repoIssue
		updated.Assignments = []repository.PartialAssignee{
			{ID: assigneeID, Kind: model.AssignmentKindAssignee},
		}
		issueRepo.EXPECT().Get(ctx, issueID, repository.IssueDetailProjection()).Return(&updated, nil)

		assignmentRepo := mockrepo.NewMockAssignmentRepository(ctrl)
		assignmentRepo.EXPECT().ListByResource(ctx, issueID, repository.CursorPage{Size: service.AssignmentSyncPageSize}, repository.AssignmentListProjection()).Return(repository.Page[*repository.Assignment]{Items: []*repository.Assignment{
			{
				ID:       existingAssignmentID,
				Kind:     model.AssignmentKindAssignee,
				User:     staleAssigneeID,
				Resource: issueID,
			},
			{
				ID:       model.MustNewID(model.ResourceTypeAssignment),
				Kind:     model.AssignmentKindReviewer,
				User:     staleAssigneeID,
				Resource: issueID,
			},
		}}, nil)
		assignmentRepo.EXPECT().Delete(ctx, existingAssignmentID).Return(nil)
		assignmentRepo.EXPECT().Create(ctx, repository.CreateAssignmentOpts{
			Kind:     model.AssignmentKindAssignee,
			User:     assigneeID,
			Resource: issueID,
		}).Return(&repository.Assignment{}, nil)

		permSvc := mocksvc.NewMockPermissionService(ctrl)
		permSvc.EXPECT().BootstrapCreator(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		permSvc.EXPECT().CtxUserHas(ctx, issueID, gomock.Any()).Return(true, nil)

		licenseSvc := mocksvc.NewMockLicenseService(ctrl)
		licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

		s := newIssueServiceForTest(issueServiceDeps{
			searchService:     mockSearchIndex(ctrl),
			logger:            mocklog.NewMockLogger(ctrl),
			tracer:            tracer,
			issueRepo:         issueRepo,
			assignmentRepo:    assignmentRepo,
			labelRepo:         mockrepo.NewMockLabelRepository(nil),
			permissionService: permSvc,
			licenseService:    licenseSvc,
		})

		got, err := s.Update(ctx, issueID, service.UpdateIssueOpts{
			Assignees: optional.Some([]model.ID{assigneeID}),
		})
		require.NoError(t, err)
		assert.Equal(t, []service.PartialAssignee{
			{ID: assigneeID, Kind: model.AssignmentKindAssignee},
		}, got.Assignments)
	})

	t.Run("sync reviewer edges", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ctx := context.Background()
		span := mocktrace.NewMockSpan(ctrl)
		span.EXPECT().End(gomock.Len(0))
		tracer := mocktrace.NewMockTracer(ctrl)
		tracer.EXPECT().Start(ctx, "service.issueService/Update", gomock.Len(0)).Return(ctx, span)

		issueRepo := mockrepo.NewMockIssueRepository(ctrl)
		issueRepo.EXPECT().Update(ctx, issueID, repository.UpdateIssueOpts{}, repository.IssueDetailProjection()).Return(repoIssue, nil)
		updated := *repoIssue
		updated.Assignments = []repository.PartialAssignee{
			{ID: reviewerID, Kind: model.AssignmentKindReviewer},
		}
		issueRepo.EXPECT().Get(ctx, issueID, repository.IssueDetailProjection()).Return(&updated, nil)

		assignmentRepo := mockrepo.NewMockAssignmentRepository(ctrl)
		assignmentRepo.EXPECT().ListByResource(ctx, issueID, repository.CursorPage{Size: service.AssignmentSyncPageSize}, repository.AssignmentListProjection()).Return(repository.Page[*repository.Assignment]{Items: []*repository.Assignment{}}, nil)
		assignmentRepo.EXPECT().Create(ctx, repository.CreateAssignmentOpts{
			Kind:     model.AssignmentKindReviewer,
			User:     reviewerID,
			Resource: issueID,
		}).Return(&repository.Assignment{}, nil)

		permSvc := mocksvc.NewMockPermissionService(ctrl)
		permSvc.EXPECT().BootstrapCreator(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		permSvc.EXPECT().CtxUserHas(ctx, issueID, gomock.Any()).Return(true, nil)

		licenseSvc := mocksvc.NewMockLicenseService(ctrl)
		licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

		s := newIssueServiceForTest(issueServiceDeps{
			searchService:     mockSearchIndex(ctrl),
			logger:            mocklog.NewMockLogger(ctrl),
			tracer:            tracer,
			issueRepo:         issueRepo,
			assignmentRepo:    assignmentRepo,
			labelRepo:         mockrepo.NewMockLabelRepository(nil),
			permissionService: permSvc,
			licenseService:    licenseSvc,
		})

		got, err := s.Update(ctx, issueID, service.UpdateIssueOpts{
			Reviewers: optional.Some([]model.ID{reviewerID}),
		})
		require.NoError(t, err)
		assert.Equal(t, []service.PartialAssignee{
			{ID: reviewerID, Kind: model.AssignmentKindReviewer},
		}, got.Assignments)
	})

	t.Run("reject multiple assignees without license feature", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ctx := context.Background()
		span := mocktrace.NewMockSpan(ctrl)
		span.EXPECT().End(gomock.Len(0))
		tracer := mocktrace.NewMockTracer(ctrl)
		tracer.EXPECT().Start(ctx, "service.issueService/Update", gomock.Len(0)).Return(ctx, span)

		permSvc := mocksvc.NewMockPermissionService(ctrl)
		permSvc.EXPECT().BootstrapCreator(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		permSvc.EXPECT().CtxUserHas(ctx, issueID, gomock.Any()).Return(true, nil)

		licenseSvc := mocksvc.NewMockLicenseService(ctrl)
		licenseSvc.EXPECT().Expired(ctx).Return(false, nil)
		licenseSvc.EXPECT().HasFeature(ctx, license.FeatureMultipleAssignees).Return(false, nil)

		s := newIssueServiceForTest(issueServiceDeps{
			searchService:     mocksvc.NewMockSearchService(ctrl),
			logger:            mocklog.NewMockLogger(ctrl),
			tracer:            tracer,
			permissionService: permSvc,
			licenseService:    licenseSvc,
		})

		_, err := s.Update(ctx, issueID, service.UpdateIssueOpts{
			Assignees: optional.Some([]model.ID{
				model.MustNewID(model.ResourceTypeUser),
				model.MustNewID(model.ResourceTypeUser),
			}),
		})
		require.Error(t, err)
		assert.ErrorIs(t, err, service.ErrQuotaExceeded)
	})

	t.Run("allow multiple assignees with license feature", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ctx := context.Background()
		span := mocktrace.NewMockSpan(ctrl)
		span.EXPECT().End(gomock.Len(0))
		tracer := mocktrace.NewMockTracer(ctrl)
		tracer.EXPECT().Start(ctx, "service.issueService/Update", gomock.Len(0)).Return(ctx, span)

		assigneeA := model.MustNewID(model.ResourceTypeUser)
		assigneeB := model.MustNewID(model.ResourceTypeUser)

		issueRepo := mockrepo.NewMockIssueRepository(ctrl)
		issueRepo.EXPECT().Update(ctx, issueID, repository.UpdateIssueOpts{}, repository.IssueDetailProjection()).Return(repoIssue, nil)
		updated := *repoIssue
		updated.Assignments = []repository.PartialAssignee{
			{ID: assigneeA, Kind: model.AssignmentKindAssignee},
			{ID: assigneeB, Kind: model.AssignmentKindAssignee},
		}
		issueRepo.EXPECT().Get(ctx, issueID, repository.IssueDetailProjection()).Return(&updated, nil)

		assignmentRepo := mockrepo.NewMockAssignmentRepository(ctrl)
		assignmentRepo.EXPECT().ListByResource(ctx, issueID, repository.CursorPage{Size: service.AssignmentSyncPageSize}, repository.AssignmentListProjection()).Return(repository.Page[*repository.Assignment]{Items: []*repository.Assignment{}}, nil)
		assignmentRepo.EXPECT().Create(ctx, repository.CreateAssignmentOpts{
			Kind:     model.AssignmentKindAssignee,
			User:     assigneeA,
			Resource: issueID,
		}).Return(&repository.Assignment{}, nil)
		assignmentRepo.EXPECT().Create(ctx, repository.CreateAssignmentOpts{
			Kind:     model.AssignmentKindAssignee,
			User:     assigneeB,
			Resource: issueID,
		}).Return(&repository.Assignment{}, nil)

		permSvc := mocksvc.NewMockPermissionService(ctrl)
		permSvc.EXPECT().BootstrapCreator(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		permSvc.EXPECT().CtxUserHas(ctx, issueID, gomock.Any()).Return(true, nil)

		licenseSvc := mocksvc.NewMockLicenseService(ctrl)
		licenseSvc.EXPECT().Expired(ctx).Return(false, nil)
		licenseSvc.EXPECT().HasFeature(ctx, license.FeatureMultipleAssignees).Return(true, nil)

		s := newIssueServiceForTest(issueServiceDeps{
			searchService:     mockSearchIndex(ctrl),
			logger:            mocklog.NewMockLogger(ctrl),
			tracer:            tracer,
			issueRepo:         issueRepo,
			assignmentRepo:    assignmentRepo,
			labelRepo:         mockrepo.NewMockLabelRepository(nil),
			permissionService: permSvc,
			licenseService:    licenseSvc,
		})

		got, err := s.Update(ctx, issueID, service.UpdateIssueOpts{
			Assignees: optional.Some([]model.ID{assigneeA, assigneeB}),
		})
		require.NoError(t, err)
		assert.Equal(t, []service.PartialAssignee{
			{ID: assigneeA, Kind: model.AssignmentKindAssignee},
			{ID: assigneeB, Kind: model.AssignmentKindAssignee},
		}, got.Assignments)
	})
}

func TestIssueService_UpdateParent(t *testing.T) {
	issueID := model.MustNewID(model.ResourceTypeIssue)
	userID := model.MustNewID(model.ResourceTypeUser)
	parentID := model.MustNewID(model.ResourceTypeIssue)
	repoIssue := testModel.NewRepositoryIssue(userID)
	repoIssue.ID = issueID

	t.Run("set parent", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ctx := context.Background()
		span := mocktrace.NewMockSpan(ctrl)
		span.EXPECT().End(gomock.Len(0))
		tracer := mocktrace.NewMockTracer(ctrl)
		tracer.EXPECT().Start(ctx, "service.issueService/Update", gomock.Len(0)).Return(ctx, span)

		parent := &repository.PartialIssue{
			ID:          parentID,
			Kind:        model.IssueKindEpic,
			Title:       "parent",
			Status:      model.IssueStatusOpen,
			Priority:    model.IssuePriorityNormal,
			Assignments: make([]repository.PartialAssignee, 0),
			Labels:      make([]repository.PartialLabel, 0),
		}

		issueRepo := mockrepo.NewMockIssueRepository(ctrl)
		issueRepo.EXPECT().Update(ctx, issueID, repository.UpdateIssueOpts{}, repository.IssueDetailProjection()).Return(repoIssue, nil)
		issueRepo.EXPECT().AddRelation(ctx, repository.CreateIssueRelationOpts{
			Source: issueID,
			Target: parentID,
			Kind:   model.IssueRelationKindSubtaskOf,
		}).Return(&repository.IssueRelation{}, nil)
		updated := *repoIssue
		updated.Parent = parent
		issueRepo.EXPECT().Get(ctx, issueID, repository.IssueDetailProjection()).Return(&updated, nil)

		permSvc := mocksvc.NewMockPermissionService(ctrl)
		permSvc.EXPECT().BootstrapCreator(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		permSvc.EXPECT().CtxUserHas(ctx, issueID, gomock.Any()).Return(true, nil)
		permSvc.EXPECT().CtxUserHas(ctx, parentID, gomock.Any()).Return(true, nil)

		licenseSvc := mocksvc.NewMockLicenseService(ctrl)
		licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

		s := newIssueServiceForTest(issueServiceDeps{
			searchService:     mockSearchIndex(ctrl),
			logger:            mocklog.NewMockLogger(ctrl),
			tracer:            tracer,
			issueRepo:         issueRepo,
			permissionService: permSvc,
			licenseService:    licenseSvc,
		})

		got, err := s.Update(ctx, issueID, service.UpdateIssueOpts{
			Parent: optional.Some(parentID),
		})
		require.NoError(t, err)
		require.NotNil(t, got.Parent)
		assert.Equal(t, parentID, got.Parent.ID)
	})

	t.Run("clear parent", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ctx := context.Background()
		span := mocktrace.NewMockSpan(ctrl)
		span.EXPECT().End(gomock.Len(0))
		tracer := mocktrace.NewMockTracer(ctrl)
		tracer.EXPECT().Start(ctx, "service.issueService/Update", gomock.Len(0)).Return(ctx, span)

		currentParentID := model.MustNewID(model.ResourceTypeIssue)
		current := *repoIssue
		current.Parent = &repository.PartialIssue{
			ID:          currentParentID,
			Kind:        model.IssueKindEpic,
			Title:       "parent",
			Status:      model.IssueStatusOpen,
			Priority:    model.IssuePriorityNormal,
			Assignments: make([]repository.PartialAssignee, 0),
			Labels:      make([]repository.PartialLabel, 0),
		}

		issueRepo := mockrepo.NewMockIssueRepository(ctrl)
		issueRepo.EXPECT().Update(ctx, issueID, repository.UpdateIssueOpts{}, repository.IssueDetailProjection()).Return(&current, nil)
		issueRepo.EXPECT().RemoveRelation(ctx, issueID, currentParentID, model.IssueRelationKindSubtaskOf).Return(nil)
		cleared := current
		cleared.Parent = nil
		issueRepo.EXPECT().Get(ctx, issueID, repository.IssueDetailProjection()).Return(&cleared, nil)

		permSvc := mocksvc.NewMockPermissionService(ctrl)
		permSvc.EXPECT().BootstrapCreator(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		permSvc.EXPECT().CtxUserHas(ctx, issueID, gomock.Any()).Return(true, nil)

		licenseSvc := mocksvc.NewMockLicenseService(ctrl)
		licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

		s := newIssueServiceForTest(issueServiceDeps{
			searchService:     mockSearchIndex(ctrl),
			logger:            mocklog.NewMockLogger(ctrl),
			tracer:            tracer,
			issueRepo:         issueRepo,
			permissionService: permSvc,
			licenseService:    licenseSvc,
		})

		got, err := s.Update(ctx, issueID, service.UpdateIssueOpts{
			Parent: optional.Null[model.ID](),
		})
		require.NoError(t, err)
		assert.Nil(t, got.Parent)
	})

	t.Run("reject self-parent", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ctx := context.Background()
		span := mocktrace.NewMockSpan(ctrl)
		span.EXPECT().End(gomock.Len(0))
		tracer := mocktrace.NewMockTracer(ctrl)
		tracer.EXPECT().Start(ctx, "service.issueService/Update", gomock.Len(0)).Return(ctx, span)

		permSvc := mocksvc.NewMockPermissionService(ctrl)
		permSvc.EXPECT().BootstrapCreator(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		permSvc.EXPECT().CtxUserHas(ctx, issueID, gomock.Any()).Return(true, nil)

		licenseSvc := mocksvc.NewMockLicenseService(ctrl)
		licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

		s := newIssueServiceForTest(issueServiceDeps{
			searchService:     mocksvc.NewMockSearchService(ctrl),
			logger:            mocklog.NewMockLogger(ctrl),
			tracer:            tracer,
			permissionService: permSvc,
			licenseService:    licenseSvc,
		})

		_, err := s.Update(ctx, issueID, service.UpdateIssueOpts{
			Parent: optional.Some(issueID),
		})
		require.Error(t, err)
		assert.ErrorIs(t, err, service.ErrIssueSelfRelation)
	})

	t.Run("no permission on new parent", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ctx := context.Background()
		span := mocktrace.NewMockSpan(ctrl)
		span.EXPECT().End(gomock.Len(0))
		tracer := mocktrace.NewMockTracer(ctrl)
		tracer.EXPECT().Start(ctx, "service.issueService/Update", gomock.Len(0)).Return(ctx, span)

		permSvc := mocksvc.NewMockPermissionService(ctrl)
		permSvc.EXPECT().BootstrapCreator(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		permSvc.EXPECT().CtxUserHas(ctx, issueID, gomock.Any()).Return(true, nil)
		permSvc.EXPECT().CtxUserHas(ctx, parentID, gomock.Any()).Return(false, nil)

		licenseSvc := mocksvc.NewMockLicenseService(ctrl)
		licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

		s := newIssueServiceForTest(issueServiceDeps{
			searchService:     mocksvc.NewMockSearchService(ctrl),
			logger:            mocklog.NewMockLogger(ctrl),
			tracer:            tracer,
			permissionService: permSvc,
			licenseService:    licenseSvc,
		})

		_, err := s.Update(ctx, issueID, service.UpdateIssueOpts{
			Parent: optional.Some(parentID),
		})
		require.Error(t, err)
		assert.ErrorIs(t, err, service.ErrNoPermission)
	})
}

func TestIssueService_Delete(t *testing.T) {
	issueID := model.MustNewID(model.ResourceTypeIssue)

	type fields struct {
		baseService func(ctrl *gomock.Controller, ctx context.Context, id model.ID) issueServiceDeps
	}
	type args struct {
		ctx context.Context
		id  model.ID
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr error
	}{
		{
			name: "delete issue",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) issueServiceDeps {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.issueService/Delete", gomock.Len(0)).Return(ctx, span)

					issueRepo := mockrepo.NewMockIssueRepository(ctrl)
					issueRepo.EXPECT().Delete(ctx, id).Return(nil)

					permSvc := mocksvc.NewMockPermissionService(ctrl)
					permSvc.EXPECT().BootstrapCreator(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
					permSvc.EXPECT().CtxUserHas(ctx, id, gomock.Any()).Return(true, nil)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return issueServiceDeps{
						searchService:     mockSearchDelete(ctrl),
						logger:            mocklog.NewMockLogger(ctrl),
						tracer:            tracer,
						issueRepo:         issueRepo,
						permissionService: permSvc,
						licenseService:    licenseSvc,
					}
				},
			},
			args: args{
				ctx: context.Background(),
				id:  issueID,
			},
		},
		{
			name: "delete issue with license expired",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID) issueServiceDeps {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.issueService/Delete", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(true, nil)

					return issueServiceDeps{
						searchService:  mocksvc.NewMockSearchService(ctrl),
						logger:         mocklog.NewMockLogger(ctrl),
						tracer:         tracer,
						licenseService: licenseSvc,
					}
				},
			},
			args: args{
				ctx: context.Background(),
				id:  issueID,
			},
			wantErr: license.ErrLicenseExpired,
		},
		{
			name: "delete issue with no permission",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) issueServiceDeps {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.issueService/Delete", gomock.Len(0)).Return(ctx, span)

					permSvc := mocksvc.NewMockPermissionService(ctrl)
					permSvc.EXPECT().BootstrapCreator(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
					permSvc.EXPECT().CtxUserHas(ctx, id, gomock.Any()).Return(false, nil)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return issueServiceDeps{
						searchService:     mocksvc.NewMockSearchService(ctrl),
						logger:            mocklog.NewMockLogger(ctrl),
						tracer:            tracer,
						permissionService: permSvc,
						licenseService:    licenseSvc,
					}
				},
			},
			args: args{
				ctx: context.Background(),
				id:  issueID,
			},
			wantErr: service.ErrNoPermission,
		},
		{
			name: "delete issue with invalid ID",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID) issueServiceDeps {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.issueService/Delete", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return issueServiceDeps{
						searchService:  mocksvc.NewMockSearchService(ctrl),
						logger:         mocklog.NewMockLogger(ctrl),
						tracer:         tracer,
						licenseService: licenseSvc,
					}
				},
			},
			args: args{
				ctx: context.Background(),
				id:  model.ID{},
			},
			wantErr: model.ErrInvalidID,
		},
		{
			name: "delete issue with repository error",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) issueServiceDeps {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.issueService/Delete", gomock.Len(0)).Return(ctx, span)

					issueRepo := mockrepo.NewMockIssueRepository(ctrl)
					issueRepo.EXPECT().Delete(ctx, id).Return(repository.ErrIssueDelete)

					permSvc := mocksvc.NewMockPermissionService(ctrl)
					permSvc.EXPECT().BootstrapCreator(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
					permSvc.EXPECT().CtxUserHas(ctx, id, gomock.Any()).Return(true, nil)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return issueServiceDeps{
						searchService:     mocksvc.NewMockSearchService(ctrl),
						logger:            mocklog.NewMockLogger(ctrl),
						tracer:            tracer,
						issueRepo:         issueRepo,
						permissionService: permSvc,
						licenseService:    licenseSvc,
					}
				},
			},
			args: args{
				ctx: context.Background(),
				id:  issueID,
			},
			wantErr: repository.ErrIssueDelete,
		},
		{
			name: "warns when custom field delete fails",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) issueServiceDeps {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.issueService/Delete", gomock.Len(0)).Return(ctx, span)

					issueRepo := mockrepo.NewMockIssueRepository(ctrl)
					issueRepo.EXPECT().Delete(ctx, id).Return(nil)

					permSvc := mocksvc.NewMockPermissionService(ctrl)
					permSvc.EXPECT().BootstrapCreator(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
					permSvc.EXPECT().CtxUserHas(ctx, id, gomock.Any()).Return(true, nil)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					cfSvc := mocksvc.NewMockCustomFieldService(ctrl)
					cfSvc.EXPECT().DeleteForResource(ctx, id).Return(assert.AnError)

					logger := mocklog.NewMockLogger(ctrl)
					logger.EXPECT().Warn(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

					return issueServiceDeps{
						searchService:      mockSearchDelete(ctrl),
						logger:             logger,
						tracer:             tracer,
						issueRepo:          issueRepo,
						permissionService:  permSvc,
						licenseService:     licenseSvc,
						customFieldService: cfSvc,
					}
				},
			},
			args: args{
				ctx: context.Background(),
				id:  issueID,
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			s := newIssueServiceForTest(tt.fields.baseService(ctrl, tt.args.ctx, tt.args.id))

			err := s.Delete(tt.args.ctx, tt.args.id)
			if tt.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestIssueService_ListRelations(t *testing.T) {
	issueID := model.MustNewID(model.ResourceTypeIssue)
	relatedID := model.MustNewID(model.ResourceTypeIssue)
	item := &repository.IssueRelationItem{
		ID:        model.MustNewID(model.ResourceTypeIssueRelation),
		Kind:      model.IssueRelationKindBlocks,
		Source:    issueID,
		Target:    relatedID,
		Related:   &repository.PartialIssue{ID: relatedID, Key: "ENG-1", NumericID: 1, Kind: model.IssueKindTask, Title: "related", Status: model.IssueStatusOpen, Priority: model.IssuePriorityNormal, Assignments: []repository.PartialAssignee{}, Labels: []repository.PartialLabel{}},
		CreatedAt: convert.ToPointer(time.Now().UTC()),
	}

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		span := mocktrace.NewMockSpan(ctrl)
		span.EXPECT().End(gomock.Len(0))
		tracer := mocktrace.NewMockTracer(ctrl)
		tracer.EXPECT().Start(gomock.Any(), "service.issueService/ListRelations", gomock.Len(0)).Return(context.Background(), span)

		issueRepo := mockrepo.NewMockIssueRepository(ctrl)
		issueRepo.EXPECT().ListRelations(gomock.Any(), repository.IssueRelationListQuery{
			IssueID: issueID,
			Page:    repository.CursorPage{Size: 100},
		}).Return(repository.Page[*repository.IssueRelationItem]{Items: []*repository.IssueRelationItem{item}}, nil)

		permSvc := mocksvc.NewMockPermissionService(ctrl)
		permSvc.EXPECT().BootstrapCreator(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		permSvc.EXPECT().CtxUserHas(gomock.Any(), issueID, gomock.Any()).Return(true, nil)

		s := newIssueServiceForTest(issueServiceDeps{
			searchService:     mocksvc.NewMockSearchService(ctrl),
			logger:            mocklog.NewMockLogger(ctrl),
			tracer:            tracer,
			issueRepo:         issueRepo,
			permissionService: permSvc,
		})
		got, err := s.ListRelations(context.Background(), issueID, service.CursorPage{Size: 100})
		require.NoError(t, err)
		require.Len(t, got.Items, 1)
		assert.Equal(t, service.IssueRelationDirectionOutgoing, got.Items[0].Direction)
		assert.Equal(t, relatedID, got.Items[0].Related.ID)
	})

	t.Run("incoming direction", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		incoming := *item
		incoming.Source = relatedID
		incoming.Target = issueID

		span := mocktrace.NewMockSpan(ctrl)
		span.EXPECT().End(gomock.Len(0))
		tracer := mocktrace.NewMockTracer(ctrl)
		tracer.EXPECT().Start(gomock.Any(), "service.issueService/ListRelations", gomock.Len(0)).Return(context.Background(), span)

		issueRepo := mockrepo.NewMockIssueRepository(ctrl)
		issueRepo.EXPECT().ListRelations(gomock.Any(), gomock.Any()).Return(
			repository.Page[*repository.IssueRelationItem]{Items: []*repository.IssueRelationItem{&incoming}},
			nil,
		)

		permSvc := mocksvc.NewMockPermissionService(ctrl)
		permSvc.EXPECT().BootstrapCreator(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		permSvc.EXPECT().CtxUserHas(gomock.Any(), issueID, gomock.Any()).Return(true, nil)

		s := newIssueServiceForTest(issueServiceDeps{
			searchService:     mocksvc.NewMockSearchService(ctrl),
			logger:            mocklog.NewMockLogger(ctrl),
			tracer:            tracer,
			issueRepo:         issueRepo,
			permissionService: permSvc,
		})
		got, err := s.ListRelations(context.Background(), issueID, service.CursorPage{Size: 100})
		require.NoError(t, err)
		require.Len(t, got.Items, 1)
		assert.Equal(t, service.IssueRelationDirectionIncoming, got.Items[0].Direction)
	})

	t.Run("no permission", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		span := mocktrace.NewMockSpan(ctrl)
		span.EXPECT().End(gomock.Len(0))
		tracer := mocktrace.NewMockTracer(ctrl)
		tracer.EXPECT().Start(gomock.Any(), "service.issueService/ListRelations", gomock.Len(0)).Return(context.Background(), span)

		permSvc := mocksvc.NewMockPermissionService(ctrl)
		permSvc.EXPECT().BootstrapCreator(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		permSvc.EXPECT().CtxUserHas(gomock.Any(), issueID, gomock.Any()).Return(false, nil)

		s := newIssueServiceForTest(issueServiceDeps{
			searchService:     mocksvc.NewMockSearchService(ctrl),
			logger:            mocklog.NewMockLogger(ctrl),
			tracer:            tracer,
			permissionService: permSvc,
		})
		_, err := s.ListRelations(context.Background(), issueID, service.CursorPage{Size: 100})
		assert.ErrorIs(t, err, service.ErrNoPermission)
	})
}

func TestIssueService_AddRelation(t *testing.T) {
	issueID := model.MustNewID(model.ResourceTypeIssue)
	relatedID := model.MustNewID(model.ResourceTypeIssue)
	createdAt := convert.ToPointer(time.Now().UTC())
	created := &repository.IssueRelation{
		ID:        model.MustNewID(model.ResourceTypeIssueRelation),
		Source:    issueID,
		Target:    relatedID,
		Kind:      model.IssueRelationKindBlocks,
		CreatedAt: createdAt,
	}
	relatedIssue := testModel.NewRepositoryIssue(model.MustNewID(model.ResourceTypeUser))
	relatedIssue.ID = relatedID

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		span := mocktrace.NewMockSpan(ctrl)
		span.EXPECT().End(gomock.Len(0))
		tracer := mocktrace.NewMockTracer(ctrl)
		tracer.EXPECT().Start(gomock.Any(), "service.issueService/AddRelation", gomock.Len(0)).Return(context.Background(), span)

		issueRepo := mockrepo.NewMockIssueRepository(ctrl)
		issueRepo.EXPECT().AddRelation(gomock.Any(), repository.CreateIssueRelationOpts{
			Source: issueID,
			Target: relatedID,
			Kind:   model.IssueRelationKindBlocks,
		}).Return(created, nil)
		issueRepo.EXPECT().Get(gomock.Any(), relatedID, repository.IssueProjection{}).Return(relatedIssue, nil)

		permSvc := mocksvc.NewMockPermissionService(ctrl)
		permSvc.EXPECT().BootstrapCreator(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		permSvc.EXPECT().CtxUserHas(gomock.Any(), issueID, gomock.Any()).Return(true, nil)
		permSvc.EXPECT().CtxUserHas(gomock.Any(), relatedID, gomock.Any()).Return(true, nil)

		licenseSvc := mocksvc.NewMockLicenseService(ctrl)
		licenseSvc.EXPECT().Expired(gomock.Any()).Return(false, nil)

		s := newIssueServiceForTest(issueServiceDeps{
			searchService:     mocksvc.NewMockSearchService(ctrl),
			logger:            mocklog.NewMockLogger(ctrl),
			tracer:            tracer,
			issueRepo:         issueRepo,
			permissionService: permSvc,
			licenseService:    licenseSvc,
		})
		got, err := s.AddRelation(context.Background(), issueID, relatedID, model.IssueRelationKindBlocks)
		require.NoError(t, err)
		assert.Equal(t, created.ID, got.ID)
		assert.Equal(t, service.IssueRelationDirectionOutgoing, got.Direction)
		assert.Equal(t, relatedID, got.Related.ID)
	})

	t.Run("self relation", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		span := mocktrace.NewMockSpan(ctrl)
		span.EXPECT().End(gomock.Len(0))
		tracer := mocktrace.NewMockTracer(ctrl)
		tracer.EXPECT().Start(gomock.Any(), "service.issueService/AddRelation", gomock.Len(0)).Return(context.Background(), span)

		licenseSvc := mocksvc.NewMockLicenseService(ctrl)
		licenseSvc.EXPECT().Expired(gomock.Any()).Return(false, nil)

		s := newIssueServiceForTest(issueServiceDeps{
			searchService:  mocksvc.NewMockSearchService(ctrl),
			logger:         mocklog.NewMockLogger(ctrl),
			tracer:         tracer,
			licenseService: licenseSvc,
		})
		_, err := s.AddRelation(context.Background(), issueID, issueID, model.IssueRelationKindBlocks)
		assert.ErrorIs(t, err, service.ErrIssueSelfRelation)
	})

	t.Run("reserved subtask of", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		span := mocktrace.NewMockSpan(ctrl)
		span.EXPECT().End(gomock.Len(0))
		tracer := mocktrace.NewMockTracer(ctrl)
		tracer.EXPECT().Start(gomock.Any(), "service.issueService/AddRelation", gomock.Len(0)).Return(context.Background(), span)

		licenseSvc := mocksvc.NewMockLicenseService(ctrl)
		licenseSvc.EXPECT().Expired(gomock.Any()).Return(false, nil)

		s := newIssueServiceForTest(issueServiceDeps{
			searchService:  mocksvc.NewMockSearchService(ctrl),
			logger:         mocklog.NewMockLogger(ctrl),
			tracer:         tracer,
			licenseService: licenseSvc,
		})
		_, err := s.AddRelation(context.Background(), issueID, relatedID, model.IssueRelationKindSubtaskOf)
		assert.ErrorIs(t, err, service.ErrIssueReservedRelationKind)
	})

	t.Run("reserved depends on", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		span := mocktrace.NewMockSpan(ctrl)
		span.EXPECT().End(gomock.Len(0))
		tracer := mocktrace.NewMockTracer(ctrl)
		tracer.EXPECT().Start(gomock.Any(), "service.issueService/AddRelation", gomock.Len(0)).Return(context.Background(), span)

		licenseSvc := mocksvc.NewMockLicenseService(ctrl)
		licenseSvc.EXPECT().Expired(gomock.Any()).Return(false, nil)

		s := newIssueServiceForTest(issueServiceDeps{
			searchService:  mocksvc.NewMockSearchService(ctrl),
			logger:         mocklog.NewMockLogger(ctrl),
			tracer:         tracer,
			licenseService: licenseSvc,
		})
		_, err := s.AddRelation(context.Background(), issueID, relatedID, model.IssueRelationKindDependsOn)
		assert.ErrorIs(t, err, service.ErrIssueReservedRelationKind)
	})

	t.Run("no write permission", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		span := mocktrace.NewMockSpan(ctrl)
		span.EXPECT().End(gomock.Len(0))
		tracer := mocktrace.NewMockTracer(ctrl)
		tracer.EXPECT().Start(gomock.Any(), "service.issueService/AddRelation", gomock.Len(0)).Return(context.Background(), span)

		permSvc := mocksvc.NewMockPermissionService(ctrl)
		permSvc.EXPECT().BootstrapCreator(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		permSvc.EXPECT().CtxUserHas(gomock.Any(), issueID, gomock.Any()).Return(false, nil)

		licenseSvc := mocksvc.NewMockLicenseService(ctrl)
		licenseSvc.EXPECT().Expired(gomock.Any()).Return(false, nil)

		s := newIssueServiceForTest(issueServiceDeps{
			searchService:     mocksvc.NewMockSearchService(ctrl),
			logger:            mocklog.NewMockLogger(ctrl),
			tracer:            tracer,
			permissionService: permSvc,
			licenseService:    licenseSvc,
		})
		_, err := s.AddRelation(context.Background(), issueID, relatedID, model.IssueRelationKindBlocks)
		assert.ErrorIs(t, err, service.ErrNoPermission)
	})
}

func TestIssueService_UpdateRelation(t *testing.T) {
	issueID := model.MustNewID(model.ResourceTypeIssue)
	relatedID := model.MustNewID(model.ResourceTypeIssue)
	relationID := model.MustNewID(model.ResourceTypeIssueRelation)
	createdAt := convert.ToPointer(time.Now().UTC())
	relatedIssue := testModel.NewRepositoryIssue(model.MustNewID(model.ResourceTypeUser))
	relatedIssue.ID = relatedID

	t.Run("incoming kind change becomes outgoing", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		span := mocktrace.NewMockSpan(ctrl)
		span.EXPECT().End(gomock.Len(0))
		tracer := mocktrace.NewMockTracer(ctrl)
		tracer.EXPECT().Start(gomock.Any(), "service.issueService/UpdateRelation", gomock.Len(0)).Return(context.Background(), span)

		existing := &repository.IssueRelation{
			ID:     relationID,
			Source: relatedID,
			Target: issueID,
			Kind:   model.IssueRelationKindBlocks,
		}
		created := &repository.IssueRelation{
			ID:        model.MustNewID(model.ResourceTypeIssueRelation),
			Source:    issueID,
			Target:    relatedID,
			Kind:      model.IssueRelationKindRelatedTo,
			CreatedAt: createdAt,
		}

		issueRepo := mockrepo.NewMockIssueRepository(ctrl)
		issueRepo.EXPECT().GetRelation(gomock.Any(), relationID).Return(existing, nil)
		issueRepo.EXPECT().RemoveRelationByID(gomock.Any(), relationID).Return(nil)
		issueRepo.EXPECT().AddRelation(gomock.Any(), repository.CreateIssueRelationOpts{
			Source: issueID,
			Target: relatedID,
			Kind:   model.IssueRelationKindRelatedTo,
		}).Return(created, nil)
		issueRepo.EXPECT().Get(gomock.Any(), relatedID, repository.IssueProjection{}).Return(relatedIssue, nil)

		permSvc := mocksvc.NewMockPermissionService(ctrl)
		permSvc.EXPECT().BootstrapCreator(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		permSvc.EXPECT().CtxUserHas(gomock.Any(), issueID, gomock.Any()).Return(true, nil)
		permSvc.EXPECT().CtxUserHas(gomock.Any(), relatedID, gomock.Any()).Return(true, nil)

		licenseSvc := mocksvc.NewMockLicenseService(ctrl)
		licenseSvc.EXPECT().Expired(gomock.Any()).Return(false, nil)

		s := newIssueServiceForTest(issueServiceDeps{
			searchService:     mocksvc.NewMockSearchService(ctrl),
			logger:            mocklog.NewMockLogger(ctrl),
			tracer:            tracer,
			issueRepo:         issueRepo,
			permissionService: permSvc,
			licenseService:    licenseSvc,
		})
		got, err := s.UpdateRelation(context.Background(), issueID, relationID, model.IssueRelationKindRelatedTo)
		require.NoError(t, err)
		assert.Equal(t, created.ID, got.ID)
		assert.Equal(t, service.IssueRelationDirectionOutgoing, got.Direction)
		assert.Equal(t, model.IssueRelationKindRelatedTo, got.Kind)
	})

	t.Run("relation does not belong to issue", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		span := mocktrace.NewMockSpan(ctrl)
		span.EXPECT().End(gomock.Len(0))
		tracer := mocktrace.NewMockTracer(ctrl)
		tracer.EXPECT().Start(gomock.Any(), "service.issueService/UpdateRelation", gomock.Len(0)).Return(context.Background(), span)

		issueRepo := mockrepo.NewMockIssueRepository(ctrl)
		issueRepo.EXPECT().GetRelation(gomock.Any(), relationID).Return(&repository.IssueRelation{
			ID:     relationID,
			Source: model.MustNewID(model.ResourceTypeIssue),
			Target: model.MustNewID(model.ResourceTypeIssue),
			Kind:   model.IssueRelationKindBlocks,
		}, nil)

		permSvc := mocksvc.NewMockPermissionService(ctrl)
		permSvc.EXPECT().BootstrapCreator(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		permSvc.EXPECT().CtxUserHas(gomock.Any(), issueID, gomock.Any()).Return(true, nil)

		licenseSvc := mocksvc.NewMockLicenseService(ctrl)
		licenseSvc.EXPECT().Expired(gomock.Any()).Return(false, nil)

		s := newIssueServiceForTest(issueServiceDeps{
			searchService:     mocksvc.NewMockSearchService(ctrl),
			logger:            mocklog.NewMockLogger(ctrl),
			tracer:            tracer,
			issueRepo:         issueRepo,
			permissionService: permSvc,
			licenseService:    licenseSvc,
		})
		_, err := s.UpdateRelation(context.Background(), issueID, relationID, model.IssueRelationKindRelatedTo)
		assert.ErrorIs(t, err, repository.ErrNotFound)
	})

	t.Run("reserved subtask of", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		span := mocktrace.NewMockSpan(ctrl)
		span.EXPECT().End(gomock.Len(0))
		tracer := mocktrace.NewMockTracer(ctrl)
		tracer.EXPECT().Start(gomock.Any(), "service.issueService/UpdateRelation", gomock.Len(0)).Return(context.Background(), span)

		licenseSvc := mocksvc.NewMockLicenseService(ctrl)
		licenseSvc.EXPECT().Expired(gomock.Any()).Return(false, nil)

		s := newIssueServiceForTest(issueServiceDeps{
			searchService:  mocksvc.NewMockSearchService(ctrl),
			logger:         mocklog.NewMockLogger(ctrl),
			tracer:         tracer,
			licenseService: licenseSvc,
		})
		_, err := s.UpdateRelation(context.Background(), issueID, relationID, model.IssueRelationKindSubtaskOf)
		assert.ErrorIs(t, err, service.ErrIssueReservedRelationKind)
	})

	t.Run("reserved depends on", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		span := mocktrace.NewMockSpan(ctrl)
		span.EXPECT().End(gomock.Len(0))
		tracer := mocktrace.NewMockTracer(ctrl)
		tracer.EXPECT().Start(gomock.Any(), "service.issueService/UpdateRelation", gomock.Len(0)).Return(context.Background(), span)

		licenseSvc := mocksvc.NewMockLicenseService(ctrl)
		licenseSvc.EXPECT().Expired(gomock.Any()).Return(false, nil)

		s := newIssueServiceForTest(issueServiceDeps{
			searchService:  mocksvc.NewMockSearchService(ctrl),
			logger:         mocklog.NewMockLogger(ctrl),
			tracer:         tracer,
			licenseService: licenseSvc,
		})
		_, err := s.UpdateRelation(context.Background(), issueID, relationID, model.IssueRelationKindDependsOn)
		assert.ErrorIs(t, err, service.ErrIssueReservedRelationKind)
	})
}

func TestIssueService_RemoveRelation(t *testing.T) {
	issueID := model.MustNewID(model.ResourceTypeIssue)
	relatedID := model.MustNewID(model.ResourceTypeIssue)
	relationID := model.MustNewID(model.ResourceTypeIssueRelation)

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		span := mocktrace.NewMockSpan(ctrl)
		span.EXPECT().End(gomock.Len(0))
		tracer := mocktrace.NewMockTracer(ctrl)
		tracer.EXPECT().Start(gomock.Any(), "service.issueService/RemoveRelation", gomock.Len(0)).Return(context.Background(), span)

		issueRepo := mockrepo.NewMockIssueRepository(ctrl)
		issueRepo.EXPECT().GetRelation(gomock.Any(), relationID).Return(&repository.IssueRelation{
			ID:     relationID,
			Source: issueID,
			Target: relatedID,
			Kind:   model.IssueRelationKindBlocks,
		}, nil)
		issueRepo.EXPECT().RemoveRelationByID(gomock.Any(), relationID).Return(nil)

		permSvc := mocksvc.NewMockPermissionService(ctrl)
		permSvc.EXPECT().BootstrapCreator(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		permSvc.EXPECT().CtxUserHas(gomock.Any(), issueID, gomock.Any()).Return(true, nil)
		permSvc.EXPECT().CtxUserHas(gomock.Any(), relatedID, gomock.Any()).Return(true, nil)

		licenseSvc := mocksvc.NewMockLicenseService(ctrl)
		licenseSvc.EXPECT().Expired(gomock.Any()).Return(false, nil)

		s := newIssueServiceForTest(issueServiceDeps{
			searchService:     mocksvc.NewMockSearchService(ctrl),
			logger:            mocklog.NewMockLogger(ctrl),
			tracer:            tracer,
			issueRepo:         issueRepo,
			permissionService: permSvc,
			licenseService:    licenseSvc,
		})
		require.NoError(t, s.RemoveRelation(context.Background(), issueID, relationID))
	})

	t.Run("no permission on related", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		span := mocktrace.NewMockSpan(ctrl)
		span.EXPECT().End(gomock.Len(0))
		tracer := mocktrace.NewMockTracer(ctrl)
		tracer.EXPECT().Start(gomock.Any(), "service.issueService/RemoveRelation", gomock.Len(0)).Return(context.Background(), span)

		issueRepo := mockrepo.NewMockIssueRepository(ctrl)
		issueRepo.EXPECT().GetRelation(gomock.Any(), relationID).Return(&repository.IssueRelation{
			ID:     relationID,
			Source: issueID,
			Target: relatedID,
			Kind:   model.IssueRelationKindBlocks,
		}, nil)

		permSvc := mocksvc.NewMockPermissionService(ctrl)
		permSvc.EXPECT().BootstrapCreator(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		permSvc.EXPECT().CtxUserHas(gomock.Any(), issueID, gomock.Any()).Return(true, nil)
		permSvc.EXPECT().CtxUserHas(gomock.Any(), relatedID, gomock.Any()).Return(false, nil)

		licenseSvc := mocksvc.NewMockLicenseService(ctrl)
		licenseSvc.EXPECT().Expired(gomock.Any()).Return(false, nil)

		s := newIssueServiceForTest(issueServiceDeps{
			searchService:     mocksvc.NewMockSearchService(ctrl),
			logger:            mocklog.NewMockLogger(ctrl),
			tracer:            tracer,
			issueRepo:         issueRepo,
			permissionService: permSvc,
			licenseService:    licenseSvc,
		})
		err := s.RemoveRelation(context.Background(), issueID, relationID)
		assert.ErrorIs(t, err, service.ErrNoPermission)
	})
}

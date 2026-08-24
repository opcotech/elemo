package service_test

import (
	"context"
	"strings"
	"testing"

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
	"github.com/opcotech/elemo/internal/pkg/log"
	"github.com/opcotech/elemo/internal/pkg/optional"
	"github.com/opcotech/elemo/internal/pkg/tracing"
	"github.com/opcotech/elemo/internal/repository"
	testModel "github.com/opcotech/elemo/internal/testutil/model"
)

func newCreateDocumentOpts() service.CreateDocumentOpts {
	return service.CreateDocumentOpts{
		Title:   "test document",
		Excerpt: "test excerpt for the document",
		Content: []byte("document body"),
	}
}

func matchDocumentFileID() gomock.Matcher {
	return gomock.Cond(func(path string) bool {
		return strings.HasPrefix(path, service.DocumentFilePrefix)
	})
}

type documentServiceDeps struct {
	logger            log.Logger
	tracer            tracing.Tracer
	documentRepo      repository.DocumentRepository
	permissionService service.PermissionService
	licenseService    service.LicenseService
	staticFileService service.StaticFileService
	searchService     service.SearchService
}

func newDocumentServiceForTest(deps documentServiceDeps) service.DocumentService {
	if deps.documentRepo == nil {
		deps.documentRepo = mockrepo.NewMockDocumentRepository(nil)
	}
	if deps.licenseService == nil {
		deps.licenseService = mocksvc.NewMockLicenseService(nil)
	}
	if deps.permissionService == nil {
		deps.permissionService = mocksvc.NewMockPermissionService(nil)
	}
	if deps.staticFileService == nil {
		deps.staticFileService = mocksvc.NewMockStaticFileService(nil)
	}
	if deps.searchService == nil {
		deps.searchService = mocksvc.NewMockSearchService(nil)
	}
	var opts []service.Option
	if deps.logger != nil {
		opts = append(opts, service.WithLogger(deps.logger))
	}
	if deps.tracer != nil {
		opts = append(opts, service.WithTracer(deps.tracer))
	}
	svc, err := service.NewDocumentService(
		deps.documentRepo,
		deps.licenseService,
		deps.permissionService,
		deps.staticFileService,
		deps.searchService,
		opts...,
	)
	if err != nil {
		panic(err)
	}
	return svc
}

func TestNewDocumentService(t *testing.T) {
	tests := []struct {
		name    string
		build   func(ctrl *gomock.Controller) (service.DocumentService, error)
		wantErr error
	}{
		{
			name: "new document service",
			build: func(ctrl *gomock.Controller) (service.DocumentService, error) {
				return service.NewDocumentService(
					mockrepo.NewMockDocumentRepository(nil),
					mocksvc.NewMockLicenseService(nil),
					mocksvc.NewMockPermissionService(nil),
					mocksvc.NewMockStaticFileService(nil),
					mocksvc.NewMockSearchService(nil),
					service.WithLogger(mocklog.NewMockLogger(ctrl)),
					service.WithTracer(mocktrace.NewMockTracer(ctrl)),
				)
			},
		},
		{
			name: "new document service with invalid options",
			build: func(_ *gomock.Controller) (service.DocumentService, error) {
				return service.NewDocumentService(
					mockrepo.NewMockDocumentRepository(nil),
					mocksvc.NewMockLicenseService(nil),
					mocksvc.NewMockPermissionService(nil),
					mocksvc.NewMockStaticFileService(nil),
					mocksvc.NewMockSearchService(nil),
					service.WithLogger(nil),
				)
			},
			wantErr: log.ErrNoLogger,
		},
		{
			name: "new document service with no document repository",
			build: func(ctrl *gomock.Controller) (service.DocumentService, error) {
				return service.NewDocumentService(
					nil,
					mocksvc.NewMockLicenseService(nil),
					mocksvc.NewMockPermissionService(nil),
					mocksvc.NewMockStaticFileService(nil),
					mocksvc.NewMockSearchService(nil),
					service.WithLogger(mocklog.NewMockLogger(ctrl)),
					service.WithTracer(mocktrace.NewMockTracer(ctrl)),
				)
			},
			wantErr: service.ErrNoDocumentRepository,
		},
		{
			name: "new document service with no permission service",
			build: func(ctrl *gomock.Controller) (service.DocumentService, error) {
				return service.NewDocumentService(
					mockrepo.NewMockDocumentRepository(nil),
					mocksvc.NewMockLicenseService(nil),
					nil,
					mocksvc.NewMockStaticFileService(nil),
					mocksvc.NewMockSearchService(nil),
					service.WithLogger(mocklog.NewMockLogger(ctrl)),
					service.WithTracer(mocktrace.NewMockTracer(ctrl)),
				)
			},
			wantErr: service.ErrNoPermissionService,
		},
		{
			name: "new document service with no license service",
			build: func(ctrl *gomock.Controller) (service.DocumentService, error) {
				return service.NewDocumentService(
					mockrepo.NewMockDocumentRepository(nil),
					nil,
					mocksvc.NewMockPermissionService(nil),
					mocksvc.NewMockStaticFileService(nil),
					mocksvc.NewMockSearchService(nil),
					service.WithLogger(mocklog.NewMockLogger(ctrl)),
					service.WithTracer(mocktrace.NewMockTracer(ctrl)),
				)
			},
			wantErr: service.ErrNoLicenseService,
		},
		{
			name: "new document service with no static file service",
			build: func(ctrl *gomock.Controller) (service.DocumentService, error) {
				return service.NewDocumentService(
					mockrepo.NewMockDocumentRepository(nil),
					mocksvc.NewMockLicenseService(nil),
					mocksvc.NewMockPermissionService(nil),
					nil,
					mocksvc.NewMockSearchService(nil),
					service.WithLogger(mocklog.NewMockLogger(ctrl)),
					service.WithTracer(mocktrace.NewMockTracer(ctrl)),
				)
			},
			wantErr: service.ErrNoStaticFileService,
		},
		{
			name: "new document service with no search service",
			build: func(ctrl *gomock.Controller) (service.DocumentService, error) {
				return service.NewDocumentService(
					mockrepo.NewMockDocumentRepository(nil),
					mocksvc.NewMockLicenseService(nil),
					mocksvc.NewMockPermissionService(nil),
					mocksvc.NewMockStaticFileService(nil),
					nil,
					service.WithLogger(mocklog.NewMockLogger(ctrl)),
					service.WithTracer(mocktrace.NewMockTracer(ctrl)),
				)
			},
			wantErr: service.ErrNoSearchService,
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

func TestDocumentService_Create(t *testing.T) {
	belongsTo := model.MustNewID(model.ResourceTypeOrganization)
	userID := model.MustNewID(model.ResourceTypeUser)
	opts := newCreateDocumentOpts()

	type fields struct {
		baseService func(ctrl *gomock.Controller, ctx context.Context, userID, belongsTo model.ID, opts service.CreateDocumentOpts) documentServiceDeps
	}
	type args struct {
		ctx       context.Context
		belongsTo model.ID
		opts      service.CreateDocumentOpts
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr error
	}{
		{
			name: "create document",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, userID, belongsTo model.ID, opts service.CreateDocumentOpts) documentServiceDeps {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.documentService/Create", gomock.Len(0)).Return(ctx, span)

					staticFileSvc := mocksvc.NewMockStaticFileService(ctrl)
					staticFileSvc.EXPECT().Create(ctx, matchDocumentFileID(), opts.Content).Return(nil)

					documentRepo := mockrepo.NewMockDocumentRepository(ctrl)
					documentRepo.EXPECT().Create(ctx, gomock.Cond(func(got repository.CreateDocumentOpts) bool {
						return got.Library == belongsTo &&
							got.Title == opts.Title &&
							got.Excerpt == opts.Excerpt &&
							got.CreatedBy == userID &&
							strings.HasPrefix(got.FileID, service.DocumentFilePrefix)
					})).Return(testModel.NewRepositoryDocument(userID), nil)

					permSvc := mocksvc.NewMockPermissionService(ctrl)
					permSvc.EXPECT().BootstrapCreator(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
					permSvc.EXPECT().CtxUserHas(ctx, belongsTo, gomock.Any()).Return(true, nil)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)
					licenseSvc.EXPECT().WithinThreshold(ctx, license.QuotaDocuments).Return(true, nil)

					return documentServiceDeps{
						searchService:     mockSearchIndex(ctrl),
						logger:            mocklog.NewMockLogger(ctrl),
						tracer:            tracer,
						documentRepo:      documentRepo,
						permissionService: permSvc,
						licenseService:    licenseSvc,
						staticFileService: staticFileSvc,
					}
				},
			},
			args: args{
				ctx:       context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				belongsTo: belongsTo,
				opts:      opts,
			},
		},
		{
			name: "create document with license expired",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, _ model.ID, _ service.CreateDocumentOpts) documentServiceDeps {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.documentService/Create", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(true, nil)

					return documentServiceDeps{
						searchService:  mocksvc.NewMockSearchService(ctrl),
						logger:         mocklog.NewMockLogger(ctrl),
						tracer:         tracer,
						licenseService: licenseSvc,
					}
				},
			},
			args: args{
				ctx:       context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				belongsTo: belongsTo,
				opts:      opts,
			},
			wantErr: license.ErrLicenseExpired,
		},
		{
			name: "create document with quota exceeded",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, belongsTo model.ID, _ service.CreateDocumentOpts) documentServiceDeps {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.documentService/Create", gomock.Len(0)).Return(ctx, span)

					permSvc := mocksvc.NewMockPermissionService(ctrl)
					permSvc.EXPECT().BootstrapCreator(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
					permSvc.EXPECT().CtxUserHas(ctx, belongsTo, gomock.Any()).Return(true, nil)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)
					licenseSvc.EXPECT().WithinThreshold(ctx, license.QuotaDocuments).Return(false, nil)

					return documentServiceDeps{
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
				belongsTo: belongsTo,
				opts:      opts,
			},
			wantErr: service.ErrQuotaExceeded,
		},
		{
			name: "create document with invalid belongsTo",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, _ model.ID, _ service.CreateDocumentOpts) documentServiceDeps {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.documentService/Create", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return documentServiceDeps{
						searchService:  mocksvc.NewMockSearchService(ctrl),
						logger:         mocklog.NewMockLogger(ctrl),
						tracer:         tracer,
						licenseService: licenseSvc,
					}
				},
			},
			args: args{
				ctx:       context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				belongsTo: model.ID{},
				opts:      opts,
			},
			wantErr: model.ErrInvalidID,
		},
		{
			name: "create document with invalid details",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, _ model.ID, _ service.CreateDocumentOpts) documentServiceDeps {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.documentService/Create", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return documentServiceDeps{
						searchService:  mocksvc.NewMockSearchService(ctrl),
						logger:         mocklog.NewMockLogger(ctrl),
						tracer:         tracer,
						licenseService: licenseSvc,
					}
				},
			},
			args: args{
				ctx:       context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				belongsTo: belongsTo,
				opts:      service.CreateDocumentOpts{Title: "ab", Content: []byte("body")},
			},
			wantErr: model.ErrInvalidDocumentDetails,
		},
		{
			name: "create document with empty content",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, userID, belongsTo model.ID, opts service.CreateDocumentOpts) documentServiceDeps {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.documentService/Create", gomock.Len(0)).Return(ctx, span)

					staticFileSvc := mocksvc.NewMockStaticFileService(ctrl)
					staticFileSvc.EXPECT().Create(ctx, matchDocumentFileID(), opts.Content).Return(nil)

					documentRepo := mockrepo.NewMockDocumentRepository(ctrl)
					documentRepo.EXPECT().Create(ctx, gomock.Cond(func(got repository.CreateDocumentOpts) bool {
						return got.Library == belongsTo &&
							got.Title == opts.Title &&
							got.Excerpt == opts.Excerpt &&
							got.CreatedBy == userID &&
							strings.HasPrefix(got.FileID, service.DocumentFilePrefix)
					})).Return(testModel.NewRepositoryDocument(userID), nil)

					permSvc := mocksvc.NewMockPermissionService(ctrl)
					permSvc.EXPECT().BootstrapCreator(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
					permSvc.EXPECT().CtxUserHas(ctx, belongsTo, gomock.Any()).Return(true, nil)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)
					licenseSvc.EXPECT().WithinThreshold(ctx, license.QuotaDocuments).Return(true, nil)

					return documentServiceDeps{
						searchService:     mockSearchIndex(ctrl),
						logger:            mocklog.NewMockLogger(ctrl),
						tracer:            tracer,
						documentRepo:      documentRepo,
						permissionService: permSvc,
						licenseService:    licenseSvc,
						staticFileService: staticFileSvc,
					}
				},
			},
			args: args{
				ctx:       context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				belongsTo: belongsTo,
				opts: service.CreateDocumentOpts{
					Title: "test document",
				},
			},
		},
		{
			name: "create document with no permission",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, belongsTo model.ID, _ service.CreateDocumentOpts) documentServiceDeps {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.documentService/Create", gomock.Len(0)).Return(ctx, span)

					permSvc := mocksvc.NewMockPermissionService(ctrl)
					permSvc.EXPECT().BootstrapCreator(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
					permSvc.EXPECT().CtxUserHas(ctx, belongsTo, gomock.Any()).Return(false, nil)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return documentServiceDeps{
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
				belongsTo: belongsTo,
				opts:      opts,
			},
			wantErr: service.ErrNoPermission,
		},
		{
			name: "create document with no user ID in context",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, belongsTo model.ID, _ service.CreateDocumentOpts) documentServiceDeps {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.documentService/Create", gomock.Len(0)).Return(ctx, span)

					permSvc := mocksvc.NewMockPermissionService(ctrl)
					permSvc.EXPECT().BootstrapCreator(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
					permSvc.EXPECT().CtxUserHas(ctx, belongsTo, gomock.Any()).Return(true, nil)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)
					licenseSvc.EXPECT().WithinThreshold(ctx, license.QuotaDocuments).Return(true, nil)

					return documentServiceDeps{
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
				belongsTo: belongsTo,
				opts:      opts,
			},
			wantErr: service.ErrNoUser,
		},
		{
			name: "create document with static file error",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, belongsTo model.ID, opts service.CreateDocumentOpts) documentServiceDeps {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.documentService/Create", gomock.Len(0)).Return(ctx, span)

					staticFileSvc := mocksvc.NewMockStaticFileService(ctrl)
					staticFileSvc.EXPECT().Create(ctx, matchDocumentFileID(), opts.Content).Return(service.ErrStaticFileCreate)

					permSvc := mocksvc.NewMockPermissionService(ctrl)
					permSvc.EXPECT().BootstrapCreator(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
					permSvc.EXPECT().CtxUserHas(ctx, belongsTo, gomock.Any()).Return(true, nil)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)
					licenseSvc.EXPECT().WithinThreshold(ctx, license.QuotaDocuments).Return(true, nil)

					return documentServiceDeps{
						searchService:     mocksvc.NewMockSearchService(ctrl),
						logger:            mocklog.NewMockLogger(ctrl),
						tracer:            tracer,
						permissionService: permSvc,
						licenseService:    licenseSvc,
						staticFileService: staticFileSvc,
					}
				},
			},
			args: args{
				ctx:       context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				belongsTo: belongsTo,
				opts:      opts,
			},
			wantErr: service.ErrStaticFileCreate,
		},
		{
			name: "create document with repository error",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, userID, belongsTo model.ID, opts service.CreateDocumentOpts) documentServiceDeps {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.documentService/Create", gomock.Len(0)).Return(ctx, span)

					staticFileSvc := mocksvc.NewMockStaticFileService(ctrl)
					staticFileSvc.EXPECT().Create(ctx, matchDocumentFileID(), opts.Content).Return(nil)
					staticFileSvc.EXPECT().Delete(ctx, matchDocumentFileID()).Return(nil)

					documentRepo := mockrepo.NewMockDocumentRepository(ctrl)
					documentRepo.EXPECT().Create(ctx, gomock.Cond(func(got repository.CreateDocumentOpts) bool {
						return got.Library == belongsTo &&
							got.Title == opts.Title &&
							got.Excerpt == opts.Excerpt &&
							got.CreatedBy == userID &&
							strings.HasPrefix(got.FileID, service.DocumentFilePrefix)
					})).Return(nil, repository.ErrDocumentCreate)

					permSvc := mocksvc.NewMockPermissionService(ctrl)
					permSvc.EXPECT().BootstrapCreator(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
					permSvc.EXPECT().CtxUserHas(ctx, belongsTo, gomock.Any()).Return(true, nil)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)
					licenseSvc.EXPECT().WithinThreshold(ctx, license.QuotaDocuments).Return(true, nil)

					return documentServiceDeps{
						searchService:     mocksvc.NewMockSearchService(ctrl),
						logger:            mocklog.NewMockLogger(ctrl),
						tracer:            tracer,
						documentRepo:      documentRepo,
						permissionService: permSvc,
						licenseService:    licenseSvc,
						staticFileService: staticFileSvc,
					}
				},
			},
			args: args{
				ctx:       context.WithValue(context.Background(), pkg.CtxKeyUserID, userID),
				belongsTo: belongsTo,
				opts:      opts,
			},
			wantErr: repository.ErrDocumentCreate,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			userID, _ := tt.args.ctx.Value(pkg.CtxKeyUserID).(model.ID)
			s := newDocumentServiceForTest(tt.fields.baseService(ctrl, tt.args.ctx, userID, tt.args.belongsTo, tt.args.opts))

			_, err := s.Create(tt.args.ctx, tt.args.belongsTo, tt.args.opts)
			if tt.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestDocumentService_Get(t *testing.T) {
	documentID := model.MustNewID(model.ResourceTypeDocument)
	userID := model.MustNewID(model.ResourceTypeUser)
	repoDocument := testModel.NewRepositoryDocument(userID)
	repoDocument.ID = documentID
	content := []byte("document body")
	want := service.DocumentFromRepository(repoDocument, content)

	type fields struct {
		baseService func(ctrl *gomock.Controller, ctx context.Context, id model.ID) documentServiceDeps
	}
	type args struct {
		ctx context.Context
		id  model.ID
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *service.Document
		wantErr error
	}{
		{
			name: "get document",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) documentServiceDeps {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.documentService/Get", gomock.Len(0)).Return(ctx, span)

					documentRepo := mockrepo.NewMockDocumentRepository(ctrl)
					documentRepo.EXPECT().Get(ctx, id, repository.DocumentDetailProjection()).Return(repoDocument, nil)

					staticFileSvc := mocksvc.NewMockStaticFileService(ctrl)
					staticFileSvc.EXPECT().Get(ctx, repoDocument.FileID).Return(content, nil)

					permSvc := mocksvc.NewMockPermissionService(ctrl)
					permSvc.EXPECT().BootstrapCreator(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
					permSvc.EXPECT().CtxUserHas(ctx, id, gomock.Any()).Return(true, nil)

					return documentServiceDeps{
						searchService:     mocksvc.NewMockSearchService(ctrl),
						logger:            mocklog.NewMockLogger(ctrl),
						tracer:            tracer,
						documentRepo:      documentRepo,
						permissionService: permSvc,
						staticFileService: staticFileSvc,
					}
				},
			},
			args: args{
				ctx: context.Background(),
				id:  documentID,
			},
			want: want,
		},
		{
			name: "get document with invalid ID",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID) documentServiceDeps {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.documentService/Get", gomock.Len(0)).Return(ctx, span)

					return documentServiceDeps{
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
			name: "get document with no permission",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) documentServiceDeps {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.documentService/Get", gomock.Len(0)).Return(ctx, span)

					documentRepo := mockrepo.NewMockDocumentRepository(ctrl)
					documentRepo.EXPECT().Get(ctx, id, repository.DocumentDetailProjection()).Return(repoDocument, nil)

					permSvc := mocksvc.NewMockPermissionService(ctrl)
					permSvc.EXPECT().BootstrapCreator(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
					permSvc.EXPECT().CtxUserHas(ctx, id, gomock.Any()).Return(false, nil)

					return documentServiceDeps{
						searchService:     mocksvc.NewMockSearchService(ctrl),
						logger:            mocklog.NewMockLogger(ctrl),
						tracer:            tracer,
						documentRepo:      documentRepo,
						permissionService: permSvc,
					}
				},
			},
			args: args{
				ctx: context.Background(),
				id:  documentID,
			},
			wantErr: service.ErrNoPermission,
		},
		{
			name: "get document with repository error",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) documentServiceDeps {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.documentService/Get", gomock.Len(0)).Return(ctx, span)

					documentRepo := mockrepo.NewMockDocumentRepository(ctrl)
					documentRepo.EXPECT().Get(ctx, id, repository.DocumentDetailProjection()).Return(nil, repository.ErrDocumentRead)

					return documentServiceDeps{
						searchService: mocksvc.NewMockSearchService(ctrl),
						logger:        mocklog.NewMockLogger(ctrl),
						tracer:        tracer,
						documentRepo:  documentRepo,
					}
				},
			},
			args: args{
				ctx: context.Background(),
				id:  documentID,
			},
			wantErr: repository.ErrDocumentRead,
		},
		{
			name: "get document with static file error",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) documentServiceDeps {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.documentService/Get", gomock.Len(0)).Return(ctx, span)

					documentRepo := mockrepo.NewMockDocumentRepository(ctrl)
					documentRepo.EXPECT().Get(ctx, id, repository.DocumentDetailProjection()).Return(repoDocument, nil)

					staticFileSvc := mocksvc.NewMockStaticFileService(ctrl)
					staticFileSvc.EXPECT().Get(ctx, repoDocument.FileID).Return(nil, service.ErrStaticFileGet)

					permSvc := mocksvc.NewMockPermissionService(ctrl)
					permSvc.EXPECT().BootstrapCreator(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
					permSvc.EXPECT().CtxUserHas(ctx, id, gomock.Any()).Return(true, nil)

					return documentServiceDeps{
						searchService:     mocksvc.NewMockSearchService(ctrl),
						logger:            mocklog.NewMockLogger(ctrl),
						tracer:            tracer,
						documentRepo:      documentRepo,
						permissionService: permSvc,
						staticFileService: staticFileSvc,
					}
				},
			},
			args: args{
				ctx: context.Background(),
				id:  documentID,
			},
			wantErr: service.ErrStaticFileGet,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			s := newDocumentServiceForTest(tt.fields.baseService(ctrl, tt.args.ctx, tt.args.id))

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

func TestDocumentService_Update(t *testing.T) {
	documentID := model.MustNewID(model.ResourceTypeDocument)
	userID := model.MustNewID(model.ResourceTypeUser)
	repoDocument := testModel.NewRepositoryDocument(userID)
	repoDocument.ID = documentID
	content := []byte("document body")
	updatedContent := []byte("updated document body")
	titleOpts := service.UpdateDocumentOpts{Title: optional.Some("Updated Title")}
	contentOpts := service.UpdateDocumentOpts{Content: optional.Some(updatedContent)}
	folderID := model.MustNewID(model.ResourceTypeFolder)
	folderOpts := service.UpdateDocumentOpts{FolderID: optional.Some(folderID)}
	movedRepoDocument := *repoDocument
	movedRepoDocument.Folder = &repository.DocumentFolder{ID: folderID, Name: "Guides"}

	type fields struct {
		baseService func(ctrl *gomock.Controller, ctx context.Context, id model.ID, opts service.UpdateDocumentOpts) documentServiceDeps
	}
	type args struct {
		ctx  context.Context
		id   model.ID
		opts service.UpdateDocumentOpts
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *service.Document
		wantErr error
	}{
		{
			name: "update document title",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, opts service.UpdateDocumentOpts) documentServiceDeps {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.documentService/Update", gomock.Len(0)).Return(ctx, span)

					documentRepo := mockrepo.NewMockDocumentRepository(ctrl)
					documentRepo.EXPECT().Get(ctx, id, repository.DocumentDetailProjection()).Return(repoDocument, nil)
					documentRepo.EXPECT().Update(ctx, id, repository.UpdateDocumentOpts{
						Title:   opts.Title,
						Excerpt: opts.Excerpt,
					}).Return(repoDocument, nil)

					staticFileSvc := mocksvc.NewMockStaticFileService(ctrl)
					staticFileSvc.EXPECT().Get(ctx, repoDocument.FileID).Return(content, nil)

					permSvc := mocksvc.NewMockPermissionService(ctrl)
					permSvc.EXPECT().BootstrapCreator(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
					permSvc.EXPECT().CtxUserHas(ctx, id, gomock.Any()).Return(true, nil)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return documentServiceDeps{
						searchService:     mockSearchIndex(ctrl),
						logger:            mocklog.NewMockLogger(ctrl),
						tracer:            tracer,
						documentRepo:      documentRepo,
						permissionService: permSvc,
						licenseService:    licenseSvc,
						staticFileService: staticFileSvc,
					}
				},
			},
			args: args{
				ctx:  context.Background(),
				id:   documentID,
				opts: titleOpts,
			},
			want: service.DocumentFromRepository(repoDocument, content),
		},
		{
			name: "update document title with missing static file",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, opts service.UpdateDocumentOpts) documentServiceDeps {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.documentService/Update", gomock.Len(0)).Return(ctx, span)

					documentRepo := mockrepo.NewMockDocumentRepository(ctrl)
					documentRepo.EXPECT().Get(ctx, id, repository.DocumentDetailProjection()).Return(repoDocument, nil)
					documentRepo.EXPECT().Update(ctx, id, repository.UpdateDocumentOpts{
						Title:   opts.Title,
						Excerpt: opts.Excerpt,
					}).Return(repoDocument, nil)

					staticFileSvc := mocksvc.NewMockStaticFileService(ctrl)
					staticFileSvc.EXPECT().Get(ctx, repoDocument.FileID).Return(nil, repository.ErrNotFound)

					permSvc := mocksvc.NewMockPermissionService(ctrl)
					permSvc.EXPECT().BootstrapCreator(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
					permSvc.EXPECT().CtxUserHas(ctx, id, gomock.Any()).Return(true, nil)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return documentServiceDeps{
						searchService:     mockSearchIndex(ctrl),
						logger:            mocklog.NewMockLogger(ctrl),
						tracer:            tracer,
						documentRepo:      documentRepo,
						permissionService: permSvc,
						licenseService:    licenseSvc,
						staticFileService: staticFileSvc,
					}
				},
			},
			args: args{
				ctx:  context.Background(),
				id:   documentID,
				opts: titleOpts,
			},
			want: service.DocumentFromRepository(repoDocument, []byte{}),
		},
		{
			name: "update document content",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ service.UpdateDocumentOpts) documentServiceDeps {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.documentService/Update", gomock.Len(0)).Return(ctx, span)

					documentRepo := mockrepo.NewMockDocumentRepository(ctrl)
					documentRepo.EXPECT().Get(ctx, id, repository.DocumentDetailProjection()).Return(repoDocument, nil)

					staticFileSvc := mocksvc.NewMockStaticFileService(ctrl)
					staticFileSvc.EXPECT().Update(ctx, repoDocument.FileID, updatedContent).Return(nil)
					staticFileSvc.EXPECT().Get(ctx, repoDocument.FileID).Return(updatedContent, nil)

					permSvc := mocksvc.NewMockPermissionService(ctrl)
					permSvc.EXPECT().BootstrapCreator(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
					permSvc.EXPECT().CtxUserHas(ctx, id, gomock.Any()).Return(true, nil)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return documentServiceDeps{
						searchService:     mockSearchIndex(ctrl),
						logger:            mocklog.NewMockLogger(ctrl),
						tracer:            tracer,
						documentRepo:      documentRepo,
						permissionService: permSvc,
						licenseService:    licenseSvc,
						staticFileService: staticFileSvc,
					}
				},
			},
			args: args{
				ctx:  context.Background(),
				id:   documentID,
				opts: contentOpts,
			},
			want: service.DocumentFromRepository(repoDocument, updatedContent),
		},
		{
			name: "move document to folder",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, opts service.UpdateDocumentOpts) documentServiceDeps {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0)).Times(2)

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.documentService/Update", gomock.Len(0)).Return(ctx, span)
					tracer.EXPECT().Start(ctx, "service.documentService/MoveToFolder", gomock.Len(0)).Return(ctx, span)

					documentRepo := mockrepo.NewMockDocumentRepository(ctrl)
					documentRepo.EXPECT().Get(ctx, id, repository.DocumentDetailProjection()).Return(repoDocument, nil).Times(2)
					documentRepo.EXPECT().MoveToFolder(ctx, id, opts.FolderID.Value).Return(&movedRepoDocument, nil)

					staticFileSvc := mocksvc.NewMockStaticFileService(ctrl)
					staticFileSvc.EXPECT().Get(ctx, movedRepoDocument.FileID).Return(content, nil)

					permSvc := mocksvc.NewMockPermissionService(ctrl)
					permSvc.EXPECT().BootstrapCreator(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
					permSvc.EXPECT().CtxUserHas(ctx, id, gomock.Any()).Return(true, nil).Times(2)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return documentServiceDeps{
						searchService:     mockSearchIndex(ctrl),
						logger:            mocklog.NewMockLogger(ctrl),
						tracer:            tracer,
						documentRepo:      documentRepo,
						permissionService: permSvc,
						licenseService:    licenseSvc,
						staticFileService: staticFileSvc,
					}
				},
			},
			args: args{
				ctx:  context.Background(),
				id:   documentID,
				opts: folderOpts,
			},
			want: service.DocumentFromRepository(&movedRepoDocument, content),
		},
		{
			name: "update document with license expired",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID, _ service.UpdateDocumentOpts) documentServiceDeps {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.documentService/Update", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(true, nil)

					return documentServiceDeps{
						searchService:  mocksvc.NewMockSearchService(ctrl),
						logger:         mocklog.NewMockLogger(ctrl),
						tracer:         tracer,
						licenseService: licenseSvc,
					}
				},
			},
			args: args{
				ctx:  context.Background(),
				id:   documentID,
				opts: titleOpts,
			},
			wantErr: license.ErrLicenseExpired,
		},
		{
			name: "update document with no permission",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ service.UpdateDocumentOpts) documentServiceDeps {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.documentService/Update", gomock.Len(0)).Return(ctx, span)

					documentRepo := mockrepo.NewMockDocumentRepository(ctrl)
					documentRepo.EXPECT().Get(ctx, id, repository.DocumentDetailProjection()).Return(repoDocument, nil)

					permSvc := mocksvc.NewMockPermissionService(ctrl)
					permSvc.EXPECT().BootstrapCreator(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
					permSvc.EXPECT().CtxUserHas(ctx, id, gomock.Any()).Return(false, nil)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return documentServiceDeps{
						searchService:     mocksvc.NewMockSearchService(ctrl),
						logger:            mocklog.NewMockLogger(ctrl),
						tracer:            tracer,
						documentRepo:      documentRepo,
						permissionService: permSvc,
						licenseService:    licenseSvc,
					}
				},
			},
			args: args{
				ctx:  context.Background(),
				id:   documentID,
				opts: titleOpts,
			},
			wantErr: service.ErrNoPermission,
		},
		{
			name: "update document with static file error",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, _ service.UpdateDocumentOpts) documentServiceDeps {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.documentService/Update", gomock.Len(0)).Return(ctx, span)

					documentRepo := mockrepo.NewMockDocumentRepository(ctrl)
					documentRepo.EXPECT().Get(ctx, id, repository.DocumentDetailProjection()).Return(repoDocument, nil)

					staticFileSvc := mocksvc.NewMockStaticFileService(ctrl)
					staticFileSvc.EXPECT().Update(ctx, repoDocument.FileID, updatedContent).Return(service.ErrStaticFileUpdate)

					permSvc := mocksvc.NewMockPermissionService(ctrl)
					permSvc.EXPECT().BootstrapCreator(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
					permSvc.EXPECT().CtxUserHas(ctx, id, gomock.Any()).Return(true, nil)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return documentServiceDeps{
						searchService:     mocksvc.NewMockSearchService(ctrl),
						logger:            mocklog.NewMockLogger(ctrl),
						tracer:            tracer,
						documentRepo:      documentRepo,
						permissionService: permSvc,
						licenseService:    licenseSvc,
						staticFileService: staticFileSvc,
					}
				},
			},
			args: args{
				ctx:  context.Background(),
				id:   documentID,
				opts: contentOpts,
			},
			wantErr: service.ErrStaticFileUpdate,
		},
		{
			name: "update document with repository error",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID, opts service.UpdateDocumentOpts) documentServiceDeps {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.documentService/Update", gomock.Len(0)).Return(ctx, span)

					documentRepo := mockrepo.NewMockDocumentRepository(ctrl)
					documentRepo.EXPECT().Get(ctx, id, repository.DocumentDetailProjection()).Return(repoDocument, nil)
					documentRepo.EXPECT().Update(ctx, id, repository.UpdateDocumentOpts{
						Title:   opts.Title,
						Excerpt: opts.Excerpt,
					}).Return(nil, repository.ErrDocumentUpdate)

					permSvc := mocksvc.NewMockPermissionService(ctrl)
					permSvc.EXPECT().BootstrapCreator(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
					permSvc.EXPECT().CtxUserHas(ctx, id, gomock.Any()).Return(true, nil)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return documentServiceDeps{
						searchService:     mocksvc.NewMockSearchService(ctrl),
						logger:            mocklog.NewMockLogger(ctrl),
						tracer:            tracer,
						documentRepo:      documentRepo,
						permissionService: permSvc,
						licenseService:    licenseSvc,
					}
				},
			},
			args: args{
				ctx:  context.Background(),
				id:   documentID,
				opts: titleOpts,
			},
			wantErr: repository.ErrDocumentUpdate,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			s := newDocumentServiceForTest(tt.fields.baseService(ctrl, tt.args.ctx, tt.args.id, tt.args.opts))

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

func TestDocumentService_Delete(t *testing.T) {
	documentID := model.MustNewID(model.ResourceTypeDocument)
	userID := model.MustNewID(model.ResourceTypeUser)
	repoDocument := testModel.NewRepositoryDocument(userID)
	repoDocument.ID = documentID

	type fields struct {
		baseService func(ctrl *gomock.Controller, ctx context.Context, id model.ID) documentServiceDeps
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
			name: "delete document",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) documentServiceDeps {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.documentService/Delete", gomock.Len(0)).Return(ctx, span)

					documentRepo := mockrepo.NewMockDocumentRepository(ctrl)
					documentRepo.EXPECT().Get(ctx, id, repository.DocumentDetailProjection()).Return(repoDocument, nil)
					documentRepo.EXPECT().Delete(ctx, id).Return(nil)

					staticFileSvc := mocksvc.NewMockStaticFileService(ctrl)
					staticFileSvc.EXPECT().Delete(ctx, repoDocument.FileID).Return(nil)

					permSvc := mocksvc.NewMockPermissionService(ctrl)
					permSvc.EXPECT().BootstrapCreator(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
					permSvc.EXPECT().CtxUserHas(ctx, id, gomock.Any()).Return(true, nil)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return documentServiceDeps{
						searchService:     mockSearchDelete(ctrl),
						logger:            mocklog.NewMockLogger(ctrl),
						tracer:            tracer,
						documentRepo:      documentRepo,
						permissionService: permSvc,
						licenseService:    licenseSvc,
						staticFileService: staticFileSvc,
					}
				},
			},
			args: args{
				ctx: context.Background(),
				id:  documentID,
			},
		},
		{
			name: "delete document with license expired",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, _ model.ID) documentServiceDeps {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.documentService/Delete", gomock.Len(0)).Return(ctx, span)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(true, nil)

					return documentServiceDeps{
						searchService:  mocksvc.NewMockSearchService(ctrl),
						logger:         mocklog.NewMockLogger(ctrl),
						tracer:         tracer,
						licenseService: licenseSvc,
					}
				},
			},
			args: args{
				ctx: context.Background(),
				id:  documentID,
			},
			wantErr: license.ErrLicenseExpired,
		},
		{
			name: "delete document with no permission",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) documentServiceDeps {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.documentService/Delete", gomock.Len(0)).Return(ctx, span)

					documentRepo := mockrepo.NewMockDocumentRepository(ctrl)
					documentRepo.EXPECT().Get(ctx, id, repository.DocumentDetailProjection()).Return(repoDocument, nil)

					permSvc := mocksvc.NewMockPermissionService(ctrl)
					permSvc.EXPECT().BootstrapCreator(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
					permSvc.EXPECT().CtxUserHas(ctx, id, gomock.Any()).Return(false, nil)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return documentServiceDeps{
						searchService:     mocksvc.NewMockSearchService(ctrl),
						logger:            mocklog.NewMockLogger(ctrl),
						tracer:            tracer,
						documentRepo:      documentRepo,
						permissionService: permSvc,
						licenseService:    licenseSvc,
					}
				},
			},
			args: args{
				ctx: context.Background(),
				id:  documentID,
			},
			wantErr: service.ErrNoPermission,
		},
		{
			name: "delete document with repository error",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) documentServiceDeps {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.documentService/Delete", gomock.Len(0)).Return(ctx, span)

					documentRepo := mockrepo.NewMockDocumentRepository(ctrl)
					documentRepo.EXPECT().Get(ctx, id, repository.DocumentDetailProjection()).Return(repoDocument, nil)
					documentRepo.EXPECT().Delete(ctx, id).Return(repository.ErrDocumentDelete)

					permSvc := mocksvc.NewMockPermissionService(ctrl)
					permSvc.EXPECT().BootstrapCreator(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
					permSvc.EXPECT().CtxUserHas(ctx, id, gomock.Any()).Return(true, nil)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return documentServiceDeps{
						searchService:     mocksvc.NewMockSearchService(ctrl),
						logger:            mocklog.NewMockLogger(ctrl),
						tracer:            tracer,
						documentRepo:      documentRepo,
						permissionService: permSvc,
						licenseService:    licenseSvc,
					}
				},
			},
			args: args{
				ctx: context.Background(),
				id:  documentID,
			},
			wantErr: repository.ErrDocumentDelete,
		},
		{
			name: "delete document with static file error after graph delete",
			fields: fields{
				baseService: func(ctrl *gomock.Controller, ctx context.Context, id model.ID) documentServiceDeps {
					span := mocktrace.NewMockSpan(ctrl)
					span.EXPECT().End(gomock.Len(0))

					tracer := mocktrace.NewMockTracer(ctrl)
					tracer.EXPECT().Start(ctx, "service.documentService/Delete", gomock.Len(0)).Return(ctx, span)

					documentRepo := mockrepo.NewMockDocumentRepository(ctrl)
					documentRepo.EXPECT().Get(ctx, id, repository.DocumentDetailProjection()).Return(repoDocument, nil)
					documentRepo.EXPECT().Delete(ctx, id).Return(nil)

					staticFileSvc := mocksvc.NewMockStaticFileService(ctrl)
					staticFileSvc.EXPECT().Delete(ctx, repoDocument.FileID).Return(service.ErrStaticFileDelete)

					permSvc := mocksvc.NewMockPermissionService(ctrl)
					permSvc.EXPECT().BootstrapCreator(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
					permSvc.EXPECT().CtxUserHas(ctx, id, gomock.Any()).Return(true, nil)

					licenseSvc := mocksvc.NewMockLicenseService(ctrl)
					licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

					return documentServiceDeps{
						searchService:     mocksvc.NewMockSearchService(ctrl),
						logger:            mocklog.NewMockLogger(ctrl),
						tracer:            tracer,
						documentRepo:      documentRepo,
						permissionService: permSvc,
						licenseService:    licenseSvc,
						staticFileService: staticFileSvc,
					}
				},
			},
			args: args{
				ctx: context.Background(),
				id:  documentID,
			},
			wantErr: service.ErrStaticFileDelete,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			s := newDocumentServiceForTest(tt.fields.baseService(ctrl, tt.args.ctx, tt.args.id))

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

func TestDocumentService_ListLibrary(t *testing.T) {
	t.Parallel()

	libraryID := model.MustNewID(model.ResourceTypeNamespace)
	userID := model.MustNewID(model.ResourceTypeUser)
	repoDoc := testModel.NewRepositoryDocument(userID)
	page := service.CursorPage{Size: 10}

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ctx := context.WithValue(context.Background(), pkg.CtxKeyUserID, userID)
		span := mocktrace.NewMockSpan(ctrl)
		span.EXPECT().End(gomock.Len(0))
		tracer := mocktrace.NewMockTracer(ctrl)
		tracer.EXPECT().Start(ctx, "service.documentService/ListLibrary", gomock.Len(0)).Return(ctx, span)

		permSvc := mocksvc.NewMockPermissionService(ctrl)
		permSvc.EXPECT().BootstrapCreator(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		permSvc.EXPECT().CtxUserListGrantScopes(ctx, model.ActionDocumentRead).Return([]model.ID{libraryID}, nil)
		permSvc.EXPECT().ListScopeAncestry(ctx, libraryID).Return([]model.ID{libraryID}, nil)

		documentRepo := mockrepo.NewMockDocumentRepository(ctrl)
		documentRepo.EXPECT().ListLibrary(ctx, libraryID, userID, nil, repository.LibraryListFilter{}, gomock.Any(), repository.DocumentSummaryProjection()).Return(repository.Page[*repository.Document]{
			Items: []*repository.Document{repoDoc},
		}, nil)

		s := newDocumentServiceForTest(documentServiceDeps{
			searchService:     mocksvc.NewMockSearchService(ctrl),
			logger:            mocklog.NewMockLogger(ctrl),
			tracer:            tracer,
			documentRepo:      documentRepo,
			permissionService: permSvc,
		})
		got, err := s.ListLibrary(ctx, libraryID, service.LibraryListFilter{}, page)
		require.NoError(t, err)
		require.Len(t, got.Items, 1)
		assert.Equal(t, repoDoc.ID, got.Items[0].ID)
	})

	t.Run("rejects non-library", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ctx := context.WithValue(context.Background(), pkg.CtxKeyUserID, userID)
		span := mocktrace.NewMockSpan(ctrl)
		span.EXPECT().End(gomock.Len(0))
		tracer := mocktrace.NewMockTracer(ctrl)
		tracer.EXPECT().Start(ctx, "service.documentService/ListLibrary", gomock.Len(0)).Return(ctx, span)

		s := newDocumentServiceForTest(documentServiceDeps{
			searchService: mocksvc.NewMockSearchService(ctrl),
			logger:        mocklog.NewMockLogger(ctrl),
			tracer:        tracer,
		})
		_, err := s.ListLibrary(ctx, model.MustNewID(model.ResourceTypeProject), service.LibraryListFilter{}, page)
		assert.ErrorIs(t, err, model.ErrInvalidID)
	})
}

func TestDocumentService_ListRelated(t *testing.T) {
	t.Parallel()

	projectID := model.MustNewID(model.ResourceTypeProject)
	userID := model.MustNewID(model.ResourceTypeUser)
	repoDoc := testModel.NewRepositoryDocument(userID)

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ctx := context.WithValue(context.Background(), pkg.CtxKeyUserID, userID)
		span := mocktrace.NewMockSpan(ctrl)
		span.EXPECT().End(gomock.Len(0))
		tracer := mocktrace.NewMockTracer(ctrl)
		tracer.EXPECT().Start(ctx, "service.documentService/ListRelated", gomock.Len(0)).Return(ctx, span)

		permSvc := mocksvc.NewMockPermissionService(ctrl)
		permSvc.EXPECT().BootstrapCreator(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		permSvc.EXPECT().CtxUserHas(ctx, projectID, gomock.Any()).Return(true, nil)

		documentRepo := mockrepo.NewMockDocumentRepository(ctrl)
		documentRepo.EXPECT().ListRelated(ctx, projectID, userID, gomock.Any(), repository.DocumentSummaryProjection()).Return(repository.Page[*repository.Document]{
			Items: []*repository.Document{repoDoc},
		}, nil)

		s := newDocumentServiceForTest(documentServiceDeps{
			searchService:     mocksvc.NewMockSearchService(ctrl),
			logger:            mocklog.NewMockLogger(ctrl),
			tracer:            tracer,
			documentRepo:      documentRepo,
			permissionService: permSvc,
		})
		got, err := s.ListRelated(ctx, projectID, service.CursorPage{Size: 10})
		require.NoError(t, err)
		require.Len(t, got.Items, 1)
		assert.Equal(t, repoDoc.ID, got.Items[0].ID)
	})
}

func TestDocumentService_Relate(t *testing.T) {
	t.Parallel()

	userID := model.MustNewID(model.ResourceTypeUser)
	repoDoc := testModel.NewRepositoryDocument(userID)
	projectID := model.MustNewID(model.ResourceTypeProject)

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ctx := context.Background()
		span := mocktrace.NewMockSpan(ctrl)
		span.EXPECT().End(gomock.Len(0))
		tracer := mocktrace.NewMockTracer(ctrl)
		tracer.EXPECT().Start(ctx, "service.documentService/Relate", gomock.Len(0)).Return(ctx, span)

		documentRepo := mockrepo.NewMockDocumentRepository(ctrl)
		documentRepo.EXPECT().Get(ctx, repoDoc.ID, repository.DocumentDetailProjection()).Return(repoDoc, nil)
		documentRepo.EXPECT().Relate(ctx, repoDoc.ID, projectID).Return(nil)

		permSvc := mocksvc.NewMockPermissionService(ctrl)
		permSvc.EXPECT().BootstrapCreator(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		permSvc.EXPECT().CtxUserHas(ctx, repoDoc.ID, gomock.Any()).Return(true, nil)
		permSvc.EXPECT().CtxUserHas(ctx, projectID, gomock.Any()).Return(true, nil)

		s := newDocumentServiceForTest(documentServiceDeps{
			searchService:     mocksvc.NewMockSearchService(ctrl),
			logger:            mocklog.NewMockLogger(ctrl),
			tracer:            tracer,
			documentRepo:      documentRepo,
			permissionService: permSvc,
		})
		require.NoError(t, s.Relate(ctx, repoDoc.ID, projectID))
	})
}

func TestDocumentService_Unrelate(t *testing.T) {
	t.Parallel()

	userID := model.MustNewID(model.ResourceTypeUser)
	repoDoc := testModel.NewRepositoryDocument(userID)
	projectID := model.MustNewID(model.ResourceTypeProject)

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ctx := context.Background()
		span := mocktrace.NewMockSpan(ctrl)
		span.EXPECT().End(gomock.Len(0))
		tracer := mocktrace.NewMockTracer(ctrl)
		tracer.EXPECT().Start(ctx, "service.documentService/Unrelate", gomock.Len(0)).Return(ctx, span)

		documentRepo := mockrepo.NewMockDocumentRepository(ctrl)
		documentRepo.EXPECT().Get(ctx, repoDoc.ID, repository.DocumentDetailProjection()).Return(repoDoc, nil)
		documentRepo.EXPECT().Unrelate(ctx, repoDoc.ID, projectID).Return(nil)

		permSvc := mocksvc.NewMockPermissionService(ctrl)
		permSvc.EXPECT().BootstrapCreator(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		permSvc.EXPECT().CtxUserHas(ctx, repoDoc.ID, gomock.Any()).Return(true, nil)
		permSvc.EXPECT().CtxUserHas(ctx, projectID, gomock.Any()).Return(true, nil)

		s := newDocumentServiceForTest(documentServiceDeps{
			searchService:     mocksvc.NewMockSearchService(ctrl),
			logger:            mocklog.NewMockLogger(ctrl),
			tracer:            tracer,
			documentRepo:      documentRepo,
			permissionService: permSvc,
		})
		require.NoError(t, s.Unrelate(ctx, repoDoc.ID, projectID))
	})
}

func TestDocumentService_MoveLibrary(t *testing.T) {
	t.Parallel()

	userID := model.MustNewID(model.ResourceTypeUser)
	repoDoc := testModel.NewRepositoryDocument(userID)
	newLibrary := model.MustNewID(model.ResourceTypeNamespace)
	moved := *repoDoc
	moved.Library = repository.DocumentLibrary{
		ID:   newLibrary,
		Type: model.ResourceTypeNamespace,
		Name: "Other",
	}
	content := []byte("document body")

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ctx := context.Background()
		span := mocktrace.NewMockSpan(ctrl)
		span.EXPECT().End(gomock.Len(0))
		tracer := mocktrace.NewMockTracer(ctrl)
		tracer.EXPECT().Start(ctx, "service.documentService/MoveLibrary", gomock.Len(0)).Return(ctx, span)

		documentRepo := mockrepo.NewMockDocumentRepository(ctrl)
		documentRepo.EXPECT().ResolveLibrary(ctx, newLibrary).Return(newLibrary, nil)
		documentRepo.EXPECT().Get(ctx, repoDoc.ID, repository.DocumentDetailProjection()).Return(repoDoc, nil)
		documentRepo.EXPECT().MoveLibrary(ctx, repoDoc.ID, newLibrary).Return(&moved, nil)

		permSvc := mocksvc.NewMockPermissionService(ctrl)
		permSvc.EXPECT().BootstrapCreator(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		permSvc.EXPECT().CtxUserHas(ctx, repoDoc.ID, gomock.Any()).Return(true, nil)
		permSvc.EXPECT().CtxUserHas(ctx, newLibrary, gomock.Any()).Return(true, nil)

		staticFileSvc := mocksvc.NewMockStaticFileService(ctrl)
		staticFileSvc.EXPECT().Get(ctx, moved.FileID).Return(content, nil)

		s := newDocumentServiceForTest(documentServiceDeps{
			searchService:     mockSearchIndex(ctrl),
			logger:            mocklog.NewMockLogger(ctrl),
			tracer:            tracer,
			documentRepo:      documentRepo,
			permissionService: permSvc,
			staticFileService: staticFileSvc,
		})
		got, err := s.MoveLibrary(ctx, repoDoc.ID, newLibrary)
		require.NoError(t, err)
		assert.Equal(t, newLibrary, got.Library.ID)
	})

	t.Run("missing static file", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ctx := context.Background()
		span := mocktrace.NewMockSpan(ctrl)
		span.EXPECT().End(gomock.Len(0))
		tracer := mocktrace.NewMockTracer(ctrl)
		tracer.EXPECT().Start(ctx, "service.documentService/MoveLibrary", gomock.Len(0)).Return(ctx, span)

		documentRepo := mockrepo.NewMockDocumentRepository(ctrl)
		documentRepo.EXPECT().ResolveLibrary(ctx, newLibrary).Return(newLibrary, nil)
		documentRepo.EXPECT().Get(ctx, repoDoc.ID, repository.DocumentDetailProjection()).Return(repoDoc, nil)
		documentRepo.EXPECT().MoveLibrary(ctx, repoDoc.ID, newLibrary).Return(&moved, nil)

		permSvc := mocksvc.NewMockPermissionService(ctrl)
		permSvc.EXPECT().BootstrapCreator(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		permSvc.EXPECT().CtxUserHas(ctx, repoDoc.ID, gomock.Any()).Return(true, nil)
		permSvc.EXPECT().CtxUserHas(ctx, newLibrary, gomock.Any()).Return(true, nil)

		staticFileSvc := mocksvc.NewMockStaticFileService(ctrl)
		staticFileSvc.EXPECT().Get(ctx, repoDoc.FileID).Return(nil, repository.ErrNotFound)

		s := newDocumentServiceForTest(documentServiceDeps{
			searchService:     mockSearchIndex(ctrl),
			logger:            mocklog.NewMockLogger(ctrl),
			tracer:            tracer,
			documentRepo:      documentRepo,
			permissionService: permSvc,
			staticFileService: staticFileSvc,
		})
		got, err := s.MoveLibrary(ctx, repoDoc.ID, newLibrary)
		require.NoError(t, err)
		assert.Equal(t, newLibrary, got.Library.ID)
		assert.Empty(t, got.Content)
	})

	t.Run("rejects non-library", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ctx := context.Background()
		span := mocktrace.NewMockSpan(ctrl)
		span.EXPECT().End(gomock.Len(0))
		tracer := mocktrace.NewMockTracer(ctrl)
		tracer.EXPECT().Start(ctx, "service.documentService/MoveLibrary", gomock.Len(0)).Return(ctx, span)

		s := newDocumentServiceForTest(documentServiceDeps{
			searchService: mocksvc.NewMockSearchService(ctrl),
			logger:        mocklog.NewMockLogger(ctrl),
			tracer:        tracer,
		})
		_, err := s.MoveLibrary(ctx, repoDoc.ID, model.MustNewID(model.ResourceTypeProject))
		assert.ErrorIs(t, err, model.ErrInvalidID)
	})
}

func TestDocumentService_MoveToFolder(t *testing.T) {
	t.Parallel()

	userID := model.MustNewID(model.ResourceTypeUser)
	repoDoc := testModel.NewRepositoryDocument(userID)
	folderID := model.MustNewID(model.ResourceTypeFolder)
	moved := *repoDoc
	moved.Folder = &repository.DocumentFolder{ID: folderID, Name: "Guides"}
	content := []byte("document body")

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ctx := context.Background()
		span := mocktrace.NewMockSpan(ctrl)
		span.EXPECT().End(gomock.Len(0))
		tracer := mocktrace.NewMockTracer(ctrl)
		tracer.EXPECT().Start(ctx, "service.documentService/MoveToFolder", gomock.Len(0)).Return(ctx, span)

		documentRepo := mockrepo.NewMockDocumentRepository(ctrl)
		documentRepo.EXPECT().Get(ctx, repoDoc.ID, repository.DocumentDetailProjection()).Return(repoDoc, nil)
		documentRepo.EXPECT().MoveToFolder(ctx, repoDoc.ID, &folderID).Return(&moved, nil)

		permSvc := mocksvc.NewMockPermissionService(ctrl)
		permSvc.EXPECT().BootstrapCreator(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		permSvc.EXPECT().CtxUserHas(ctx, repoDoc.ID, gomock.Any()).Return(true, nil)

		staticFileSvc := mocksvc.NewMockStaticFileService(ctrl)
		staticFileSvc.EXPECT().Get(ctx, moved.FileID).Return(content, nil)

		s := newDocumentServiceForTest(documentServiceDeps{
			searchService:     mocksvc.NewMockSearchService(ctrl),
			logger:            mocklog.NewMockLogger(ctrl),
			tracer:            tracer,
			documentRepo:      documentRepo,
			permissionService: permSvc,
			staticFileService: staticFileSvc,
		})
		got, err := s.MoveToFolder(ctx, repoDoc.ID, &folderID)
		require.NoError(t, err)
		require.NotNil(t, got.Folder)
		assert.Equal(t, folderID, got.Folder.ID)
	})

	t.Run("missing static file", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ctx := context.Background()
		span := mocktrace.NewMockSpan(ctrl)
		span.EXPECT().End(gomock.Len(0))
		tracer := mocktrace.NewMockTracer(ctrl)
		tracer.EXPECT().Start(ctx, "service.documentService/MoveToFolder", gomock.Len(0)).Return(ctx, span)

		documentRepo := mockrepo.NewMockDocumentRepository(ctrl)
		documentRepo.EXPECT().Get(ctx, repoDoc.ID, repository.DocumentDetailProjection()).Return(repoDoc, nil)
		documentRepo.EXPECT().MoveToFolder(ctx, repoDoc.ID, &folderID).Return(&moved, nil)

		permSvc := mocksvc.NewMockPermissionService(ctrl)
		permSvc.EXPECT().BootstrapCreator(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		permSvc.EXPECT().CtxUserHas(ctx, repoDoc.ID, gomock.Any()).Return(true, nil)

		staticFileSvc := mocksvc.NewMockStaticFileService(ctrl)
		staticFileSvc.EXPECT().Get(ctx, repoDoc.FileID).Return(nil, repository.ErrNotFound)

		s := newDocumentServiceForTest(documentServiceDeps{
			searchService:     mocksvc.NewMockSearchService(ctrl),
			logger:            mocklog.NewMockLogger(ctrl),
			tracer:            tracer,
			documentRepo:      documentRepo,
			permissionService: permSvc,
			staticFileService: staticFileSvc,
		})
		got, err := s.MoveToFolder(ctx, repoDoc.ID, &folderID)
		require.NoError(t, err)
		require.NotNil(t, got.Folder)
		assert.Equal(t, folderID, got.Folder.ID)
		assert.Empty(t, got.Content)
	})

	t.Run("clears folder", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ctx := context.Background()
		span := mocktrace.NewMockSpan(ctrl)
		span.EXPECT().End(gomock.Len(0))
		tracer := mocktrace.NewMockTracer(ctrl)
		tracer.EXPECT().Start(ctx, "service.documentService/MoveToFolder", gomock.Len(0)).Return(ctx, span)

		cleared := *repoDoc
		cleared.Folder = nil

		documentRepo := mockrepo.NewMockDocumentRepository(ctrl)
		documentRepo.EXPECT().Get(ctx, repoDoc.ID, repository.DocumentDetailProjection()).Return(repoDoc, nil)
		documentRepo.EXPECT().MoveToFolder(ctx, repoDoc.ID, (*model.ID)(nil)).Return(&cleared, nil)

		permSvc := mocksvc.NewMockPermissionService(ctrl)
		permSvc.EXPECT().BootstrapCreator(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		permSvc.EXPECT().CtxUserHas(ctx, repoDoc.ID, gomock.Any()).Return(true, nil)

		staticFileSvc := mocksvc.NewMockStaticFileService(ctrl)
		staticFileSvc.EXPECT().Get(ctx, cleared.FileID).Return(content, nil)

		s := newDocumentServiceForTest(documentServiceDeps{
			searchService:     mocksvc.NewMockSearchService(ctrl),
			logger:            mocklog.NewMockLogger(ctrl),
			tracer:            tracer,
			documentRepo:      documentRepo,
			permissionService: permSvc,
			staticFileService: staticFileSvc,
		})
		got, err := s.MoveToFolder(ctx, repoDoc.ID, nil)
		require.NoError(t, err)
		assert.Nil(t, got.Folder)
	})
}

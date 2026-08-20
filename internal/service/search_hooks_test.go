package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg"
	"github.com/opcotech/elemo/internal/repository"
	"github.com/opcotech/elemo/internal/testutil/mock"
	testModel "github.com/opcotech/elemo/internal/testutil/model"
)

func TestNamespaceService_Create_IndexesSearch(t *testing.T) {
	t.Parallel()

	orgID := model.MustNewID(model.ResourceTypeOrganization)
	userID := model.MustNewID(model.ResourceTypeUser)
	opts := newCreateNamespaceOpts()
	ns := testModel.NewRepositoryNamespace()
	ctx := context.WithValue(context.Background(), pkg.CtxKeyUserID, userID)

	ctrl := gomock.NewController(t)
	span := mock.NewMockSpan(ctrl)
	span.EXPECT().End(gomock.Len(0))
	tracer := mock.NewMockTracer(ctrl)
	tracer.EXPECT().Start(ctx, "service.namespaceService/Create", gomock.Len(0)).Return(ctx, span)

	namespaceRepo := repository.NewMockNamespaceRepository(ctrl)
	namespaceRepo.EXPECT().Create(ctx, repository.CreateNamespaceOpts{
		Name:        opts.Name,
		Description: opts.Description,
		CreatorID:   userID,
		OrgID:       orgID,
	}).Return(ns, nil)

	permSvc := NewMockPermissionService(ctrl)
	permSvc.EXPECT().CtxUserHas(ctx, orgID, gomock.Any()).Return(true)
	permSvc.EXPECT().BootstrapCreator(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

	licenseSvc := mock.NewMockLicenseService(ctrl)
	licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

	searchSvc := NewMockSearchService(ctrl)
	searchSvc.EXPECT().EnqueueIndex(ctx, ns.ID).Return(nil)

	svc := &namespaceService{baseService: &baseService{
		logger:            mock.NewMockLogger(ctrl),
		tracer:            tracer,
		namespaceRepo:     namespaceRepo,
		permissionService: permSvc,
		licenseService:    licenseSvc,
		searchService:     searchSvc,
	}}
	got, err := svc.Create(ctx, orgID, opts)
	require.NoError(t, err)
	assert.Equal(t, ns.ID, got.ID)
}

func TestNamespaceService_Delete_DeletesSearchByScope(t *testing.T) {
	t.Parallel()

	id := model.MustNewID(model.ResourceTypeNamespace)
	ctx := context.Background()

	ctrl := gomock.NewController(t)
	span := mock.NewMockSpan(ctrl)
	span.EXPECT().End(gomock.Len(0))
	tracer := mock.NewMockTracer(ctrl)
	tracer.EXPECT().Start(ctx, "service.namespaceService/Delete", gomock.Len(0)).Return(ctx, span)

	namespaceRepo := repository.NewMockNamespaceRepository(ctrl)
	namespaceRepo.EXPECT().Delete(ctx, id).Return(nil)

	permSvc := NewMockPermissionService(ctrl)
	permSvc.EXPECT().CtxUserHas(ctx, id, gomock.Any()).Return(true)

	licenseSvc := mock.NewMockLicenseService(ctrl)
	licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

	searchSvc := NewMockSearchService(ctrl)
	searchSvc.EXPECT().DeleteByScope(ctx, id).Return(nil)

	svc := &namespaceService{baseService: &baseService{
		logger:            mock.NewMockLogger(ctrl),
		tracer:            tracer,
		namespaceRepo:     namespaceRepo,
		permissionService: permSvc,
		licenseService:    licenseSvc,
		searchService:     searchSvc,
	}}
	require.NoError(t, svc.Delete(ctx, id))
}

func TestIssueService_Delete_DeletesSearch(t *testing.T) {
	t.Parallel()

	id := model.MustNewID(model.ResourceTypeIssue)
	ctx := context.Background()

	ctrl := gomock.NewController(t)
	span := mock.NewMockSpan(ctrl)
	span.EXPECT().End(gomock.Len(0))
	tracer := mock.NewMockTracer(ctrl)
	tracer.EXPECT().Start(ctx, "service.issueService/Delete", gomock.Len(0)).Return(ctx, span)

	issueRepo := repository.NewMockIssueRepository(ctrl)
	issueRepo.EXPECT().Delete(ctx, id).Return(nil)

	permSvc := NewMockPermissionService(ctrl)
	permSvc.EXPECT().CtxUserHas(ctx, id, gomock.Any()).Return(true)

	licenseSvc := mock.NewMockLicenseService(ctrl)
	licenseSvc.EXPECT().Expired(ctx).Return(false, nil)

	searchSvc := NewMockSearchService(ctrl)
	searchSvc.EXPECT().Delete(ctx, id).Return(nil)

	svc := &issueService{baseService: &baseService{
		logger:            mock.NewMockLogger(ctrl),
		tracer:            tracer,
		issueRepo:         issueRepo,
		permissionService: permSvc,
		licenseService:    licenseSvc,
		searchService:     searchSvc,
	}}
	require.NoError(t, svc.Delete(ctx, id))
}

package service_test

import (
	"context"
	"testing"

	mocklog "github.com/opcotech/elemo/internal/pkg/log/mock"
	mocktrace "github.com/opcotech/elemo/internal/pkg/tracing/mock"
	mockrepo "github.com/opcotech/elemo/internal/repository/mock"
	"github.com/opcotech/elemo/internal/service"
	mocksvc "github.com/opcotech/elemo/internal/service/mock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg"
	"github.com/opcotech/elemo/internal/repository"
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
	span := mocktrace.NewMockSpan(ctrl)
	span.EXPECT().End(gomock.Len(0))
	tracer := mocktrace.NewMockTracer(ctrl)
	tracer.EXPECT().Start(gomock.Any(), "service.namespaceService/Create", gomock.Len(0)).Return(ctx, span)

	namespaceRepo := mockrepo.NewMockNamespaceRepository(ctrl)
	namespaceRepo.EXPECT().Create(gomock.Any(), repository.CreateNamespaceOpts{
		Name:        opts.Name,
		Slug:        opts.Slug,
		Description: opts.Description,
		CreatorID:   userID,
		OrgID:       orgID,
	}).Return(ns, nil)

	permSvc := mocksvc.NewMockPermissionService(ctrl)
	permSvc.EXPECT().CtxUserHas(gomock.Any(), orgID, gomock.Any()).Return(true, nil)
	permSvc.EXPECT().BootstrapCreator(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

	searchSvc := mocksvc.NewMockSearchService(ctrl)
	searchSvc.EXPECT().EnqueueIndex(gomock.Any(), ns.ID).Return(nil)

	svc := func() service.NamespaceService {
		svc, err := service.NewNamespaceService(
			namespaceRepo,
			permSvc,
			searchSvc,
			service.WithLogger(mocklog.NewMockLogger(ctrl)),
			service.WithTracer(tracer),
		)
		if err != nil {
			panic(err)
		}
		return svc
	}()
	got, err := svc.Create(ctx, orgID, opts)
	require.NoError(t, err)
	assert.Equal(t, ns.ID, got.ID)
}

func TestNamespaceService_Delete_DeletesSearchByScope(t *testing.T) {
	t.Parallel()

	id := model.MustNewID(model.ResourceTypeNamespace)
	ctx := context.Background()

	ctrl := gomock.NewController(t)
	span := mocktrace.NewMockSpan(ctrl)
	span.EXPECT().End(gomock.Len(0))
	tracer := mocktrace.NewMockTracer(ctrl)
	tracer.EXPECT().Start(gomock.Any(), "service.namespaceService/Delete", gomock.Len(0)).Return(ctx, span)

	namespaceRepo := mockrepo.NewMockNamespaceRepository(ctrl)
	namespaceRepo.EXPECT().Delete(gomock.Any(), id).Return(nil)

	permSvc := mocksvc.NewMockPermissionService(ctrl)
	permSvc.EXPECT().CtxUserHas(gomock.Any(), id, gomock.Any()).Return(true, nil)

	searchSvc := mocksvc.NewMockSearchService(ctrl)
	searchSvc.EXPECT().DeleteByScope(gomock.Any(), id).Return(nil)

	svc := func() service.NamespaceService {
		svc, err := service.NewNamespaceService(
			namespaceRepo,
			permSvc,
			searchSvc,
			service.WithLogger(mocklog.NewMockLogger(ctrl)),
			service.WithTracer(tracer),
		)
		if err != nil {
			panic(err)
		}
		return svc
	}()
	require.NoError(t, svc.Delete(ctx, id))
}

func TestIssueService_Delete_DeletesSearch(t *testing.T) {
	t.Parallel()

	id := model.MustNewID(model.ResourceTypeIssue)
	ctx := context.Background()

	ctrl := gomock.NewController(t)
	span := mocktrace.NewMockSpan(ctrl)
	span.EXPECT().End(gomock.Len(0))
	tracer := mocktrace.NewMockTracer(ctrl)
	tracer.EXPECT().Start(gomock.Any(), "service.issueService/Delete", gomock.Len(0)).Return(ctx, span)

	issueRepo := mockrepo.NewMockIssueRepository(ctrl)
	issueRepo.EXPECT().Delete(gomock.Any(), id).Return(nil)

	permSvc := mocksvc.NewMockPermissionService(ctrl)
	permSvc.EXPECT().CtxUserHas(gomock.Any(), id, gomock.Any()).Return(true, nil)

	searchSvc := mocksvc.NewMockSearchService(ctrl)
	searchSvc.EXPECT().Delete(gomock.Any(), id).Return(nil)

	svc := func() service.IssueService {
		svc, err := service.NewIssueService(
			issueRepo,
			mockrepo.NewMockAssignmentRepository(ctrl),
			mockrepo.NewMockLabelRepository(ctrl),
			permSvc,
			searchSvc,
			nopCustomFieldService{},
			service.WithLogger(mocklog.NewMockLogger(ctrl)),
			service.WithTracer(tracer),
		)
		if err != nil {
			panic(err)
		}
		return svc
	}()
	require.NoError(t, svc.Delete(ctx, id))
}

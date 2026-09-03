package http

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/service"
	mocksvc "github.com/opcotech/elemo/internal/service/mock"
)

func stubOrganizationResolve(os *mocksvc.MockOrganizationService, orgID model.ID) {
	os.EXPECT().Resolve(gomock.Any(), orgID, "").Return(&service.Organization{ID: orgID, Slug: "acme", Name: "ACME"}, nil).AnyTimes()
}

func stubNamespaceResolve(ns *mocksvc.MockNamespaceService, orgID, nsID model.ID) {
	ns.EXPECT().Resolve(gomock.Any(), orgID, nsID, "").Return(&service.Namespace{ID: nsID, Slug: "platform", Name: "Platform"}, nil).AnyTimes()
}

func newTestDocumentController(t *testing.T, ctrl *gomock.Controller, ds service.DocumentService) (DocumentController, *mocksvc.MockOrganizationService, *mocksvc.MockNamespaceService) {
	t.Helper()
	os := mocksvc.NewMockOrganizationService(ctrl)
	ns := mocksvc.NewMockNamespaceService(ctrl)
	c, err := NewDocumentController(os, ns, ds)
	require.NoError(t, err)
	return c, os, ns
}

func newTestFolderController(t *testing.T, ctrl *gomock.Controller, fs service.FolderService) (FolderController, *mocksvc.MockOrganizationService, *mocksvc.MockNamespaceService) {
	t.Helper()
	os := mocksvc.NewMockOrganizationService(ctrl)
	ns := mocksvc.NewMockNamespaceService(ctrl)
	c, err := NewFolderController(os, ns, fs)
	require.NoError(t, err)
	return c, os, ns
}

func newTestIssueController(t *testing.T, ctrl *gomock.Controller, is service.IssueService) (IssueController, *mocksvc.MockOrganizationService, *mocksvc.MockNamespaceService) {
	t.Helper()
	os := mocksvc.NewMockOrganizationService(ctrl)
	ns := mocksvc.NewMockNamespaceService(ctrl)
	c, err := NewIssueController(os, ns, is)
	require.NoError(t, err)
	return c, os, ns
}

func newTestProjectController(t *testing.T, ctrl *gomock.Controller, ps service.ProjectService) (ProjectController, *mocksvc.MockOrganizationService, *mocksvc.MockNamespaceService) {
	t.Helper()
	os := mocksvc.NewMockOrganizationService(ctrl)
	ns := mocksvc.NewMockNamespaceService(ctrl)
	c, err := NewProjectController(os, ns, ps)
	require.NoError(t, err)
	return c, os, ns
}

func newTestNamespaceController(t *testing.T, ctrl *gomock.Controller, nsSvc service.NamespaceService) (NamespaceController, *mocksvc.MockOrganizationService) {
	t.Helper()
	os := mocksvc.NewMockOrganizationService(ctrl)
	c, err := NewNamespaceController(os, nsSvc)
	require.NoError(t, err)
	return c, os
}

func newTestCustomFieldController(t *testing.T, svc service.CustomFieldService) CustomFieldController {
	t.Helper()
	c, err := NewCustomFieldController(svc)
	require.NoError(t, err)
	return c
}

func newTestPluginController(t *testing.T, svc service.PluginService) PluginController {
	t.Helper()
	c, err := NewPluginController(svc)
	require.NoError(t, err)
	return c
}

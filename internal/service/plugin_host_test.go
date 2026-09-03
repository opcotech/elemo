package service_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/opcotech/elemo/internal/config"
	"github.com/opcotech/elemo/internal/license"
	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg"
	mocklog "github.com/opcotech/elemo/internal/pkg/log/mock"
	mocktrace "github.com/opcotech/elemo/internal/pkg/tracing/mock"
	elemoplugin "github.com/opcotech/elemo/internal/plugin"
	"github.com/opcotech/elemo/internal/repository"
	mockrepo "github.com/opcotech/elemo/internal/repository/mock"
	"github.com/opcotech/elemo/internal/service"
	mocksvc "github.com/opcotech/elemo/internal/service/mock"
)

type pluginHostHarness struct {
	ctx      context.Context
	userID   model.ID
	orgID    model.ID
	issueID  model.ID
	svc      service.PluginService
	repo     *mockrepo.MockPluginRepository
	extRepo  *mockrepo.MockExtensionRepository
	issues   *mocksvc.MockIssueService
	projects *mocksvc.MockProjectService
	users    *mocksvc.MockUserService
}

func hostPluginManifest() model.PluginManifest {
	return model.PluginManifest{
		SchemaVersion: 1,
		ID:            "com.elemo.timetracking",
		Name:          "Time Tracking",
		Version:       "1.1.0",
		Requires:      model.PluginRequires{PluginAPI: "^1"},
		Frontend:      &model.PluginFrontendDecl{Entry: "frontend/index.js"},
		Capabilities: []model.PluginCapability{
			model.CapabilityIssuesRead,
			model.CapabilityIssuesUpdate,
			model.CapabilityProjectsRead,
			model.CapabilityUsersRead,
			model.CapabilityPermissionsCheck,
			model.CapabilityPluginStorageRead,
			model.CapabilityPluginStorageWrite,
			model.CapabilityGraphRead,
			model.CapabilityGraphWrite,
			model.CapabilityEventsPublish,
		},
		Graph: &model.PluginGraphSchema{
			Nodes: []model.PluginGraphNodeDecl{
				{
					Kind:  "TimeEntry",
					Scope: model.PluginGraphNodeScope{Parent: "Issue"},
					Properties: []model.PluginGraphPropertyDecl{
						{Name: "seconds", Type: model.PluginGraphPropertyTypeInteger, Required: true},
						{Name: "note", Type: model.PluginGraphPropertyTypeStr},
						{Name: "user_id", Type: model.PluginGraphPropertyTypeStr},
					},
				},
			},
			Relations: []model.PluginGraphRelationDecl{
				{
					Kind:        "LOGGED_ON",
					From:        "TimeEntry",
					To:          "Issue",
					Cardinality: model.PluginGraphCardinalityManyToOne,
				},
				{
					Kind:        "LOGGED_BY",
					From:        "TimeEntry",
					To:          "User",
					Cardinality: model.PluginGraphCardinalityManyToOne,
				},
			},
		},
	}
}

func newPluginHostHarness(t *testing.T, manifest model.PluginManifest) pluginHostHarness {
	t.Helper()
	ctrl := gomock.NewController(t)
	userID := model.MustNewID(model.ResourceTypeUser)
	orgID := model.MustNewID(model.ResourceTypeOrganization)
	issueID := model.MustNewID(model.ResourceTypeIssue)
	ctx := context.WithValue(context.Background(), pkg.CtxKeyUserID, userID)

	span := mocktrace.NewMockSpan(ctrl)
	span.EXPECT().End(gomock.Len(0)).AnyTimes()
	tracer := mocktrace.NewMockTracer(ctrl)
	tracer.EXPECT().Start(gomock.Any(), gomock.Any(), gomock.Len(0)).Return(ctx, span).AnyTimes()

	logger := mocklog.NewMockLogger(ctrl)
	logger.EXPECT().Warn(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	lic := mocksvc.NewMockLicenseService(ctrl)
	perm := mocksvc.NewMockPermissionService(ctrl)
	repo := mockrepo.NewMockPluginRepository(ctrl)
	extRepo := mockrepo.NewMockExtensionRepository(ctrl)
	issues := mocksvc.NewMockIssueService(ctrl)
	projects := mocksvc.NewMockProjectService(ctrl)
	users := mocksvc.NewMockUserService(ctrl)

	svc, err := service.NewPluginService(
		config.PluginConfig{Directory: t.TempDir()},
		repo,
		extRepo,
		perm,
		lic,
		issues,
		projects,
		users,
		nil,
		service.WithLogger(logger),
		service.WithTracer(tracer),
	)
	require.NoError(t, err)

	lic.EXPECT().HasFeature(gomock.Any(), license.FeaturePlugins).Return(true, nil).AnyTimes()
	repo.EXPECT().GetInstallation(gomock.Any(), manifest.ID).Return(&model.PluginInstallation{
		PluginID: manifest.ID,
		Version:  manifest.Version,
		Status:   model.PluginStatusActive,
		Manifest: manifest,
	}, nil).AnyTimes()
	perm.EXPECT().ListScopeAncestry(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, resource model.ID) ([]model.ID, error) {
			return []model.ID{resource}, nil
		},
	).AnyTimes()
	repo.EXPECT().ListActivationsByScope(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, ancestry []model.ID) ([]*model.PluginActivation, error) {
			act := &model.PluginActivation{
				PluginID: manifest.ID,
				Enabled:  true,
				Config:   json.RawMessage(`{"k":"v"}`),
			}
			if len(ancestry) > 0 {
				act.ScopeID = ancestry[0]
			}
			return []*model.PluginActivation{act}, nil
		},
	).AnyTimes()
	perm.EXPECT().CtxUserHas(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil).AnyTimes()

	return pluginHostHarness{
		ctx:      ctx,
		userID:   userID,
		orgID:    orgID,
		issueID:  issueID,
		svc:      svc,
		repo:     repo,
		extRepo:  extRepo,
		issues:   issues,
		projects: projects,
		users:    users,
	}
}

func mustJSON(t *testing.T, v any) json.RawMessage {
	t.Helper()
	b, err := json.Marshal(v)
	require.NoError(t, err)
	return b
}

func TestPluginHost_Call(t *testing.T) {
	t.Parallel()

	manifest := hostPluginManifest()
	require.NoError(t, manifest.Validate())

	t.Run("plugin.config.get", func(t *testing.T) {
		t.Parallel()
		h := newPluginHostHarness(t, manifest)
		resp, err := service.CallPluginHost(h.ctx, h.svc, manifest.ID, elemoplugin.HostRequest{
			Method:  "plugin.config.get",
			ScopeID: h.orgID.Composite(),
		})
		require.NoError(t, err)
		require.True(t, resp.OK, resp.Error)
		assert.JSONEq(t, `{"k":"v"}`, string(resp.Data))
	})

	t.Run("issues.get", func(t *testing.T) {
		t.Parallel()
		h := newPluginHostHarness(t, manifest)
		issue := &service.Issue{
			ID:    h.issueID,
			Key:   "TT-1",
			Title: "Track time",
			Project: &service.PartialProject{
				ID: model.MustNewID(model.ResourceTypeProject),
			},
		}
		h.issues.EXPECT().Get(h.ctx, h.issueID).Return(issue, nil)
		resp, err := service.CallPluginHost(h.ctx, h.svc, manifest.ID, elemoplugin.HostRequest{
			Method:  "issues.get",
			Payload: mustJSON(t, map[string]string{"id": h.issueID.String()}),
		})
		require.NoError(t, err)
		require.True(t, resp.OK, resp.Error)
		assert.Contains(t, string(resp.Data), issue.Key)
	})

	t.Run("issues.list", func(t *testing.T) {
		t.Parallel()
		h := newPluginHostHarness(t, manifest)
		projectID := model.MustNewID(model.ResourceTypeProject)
		h.issues.EXPECT().List(h.ctx, projectID, service.CursorPage{Size: 50}, service.IssueListOptions{}).Return(service.Page[*service.PartialIssue]{
			Items: []*service.PartialIssue{{ID: h.issueID, Key: "TT-1", Title: "Track time"}},
		}, nil)
		resp, err := service.CallPluginHost(h.ctx, h.svc, manifest.ID, elemoplugin.HostRequest{
			Method:  "issues.list",
			Payload: mustJSON(t, map[string]string{"projectId": projectID.String()}),
		})
		require.NoError(t, err)
		require.True(t, resp.OK, resp.Error)
	})

	t.Run("issues.update", func(t *testing.T) {
		t.Parallel()
		h := newPluginHostHarness(t, manifest)
		h.issues.EXPECT().Update(h.ctx, h.issueID, gomock.Any()).Return(&service.Issue{
			ID:    h.issueID,
			Title: "Updated",
		}, nil)
		resp, err := service.CallPluginHost(h.ctx, h.svc, manifest.ID, elemoplugin.HostRequest{
			Method:  "issues.update",
			Payload: mustJSON(t, map[string]string{"id": h.issueID.String(), "title": "Updated"}),
		})
		require.NoError(t, err)
		require.True(t, resp.OK, resp.Error)
	})

	t.Run("projects.get", func(t *testing.T) {
		t.Parallel()
		h := newPluginHostHarness(t, manifest)
		projectID := model.MustNewID(model.ResourceTypeProject)
		h.projects.EXPECT().Get(h.ctx, projectID).Return(&service.Project{
			ID:   projectID,
			Key:  "TT",
			Name: "Time",
		}, nil)
		resp, err := service.CallPluginHost(h.ctx, h.svc, manifest.ID, elemoplugin.HostRequest{
			Method:  "projects.get",
			Payload: mustJSON(t, map[string]string{"id": projectID.String()}),
		})
		require.NoError(t, err)
		require.True(t, resp.OK, resp.Error)
	})

	t.Run("projects.list", func(t *testing.T) {
		t.Parallel()
		h := newPluginHostHarness(t, manifest)
		nsID := model.MustNewID(model.ResourceTypeNamespace)
		h.projects.EXPECT().List(h.ctx, nsID, service.CursorPage{Size: 50}).Return(service.Page[*service.Project]{
			Items: []*service.Project{{ID: model.MustNewID(model.ResourceTypeProject), Key: "TT", Name: "Time"}},
		}, nil)
		resp, err := service.CallPluginHost(h.ctx, h.svc, manifest.ID, elemoplugin.HostRequest{
			Method:  "projects.list",
			Payload: mustJSON(t, map[string]string{"namespaceId": nsID.String()}),
		})
		require.NoError(t, err)
		require.True(t, resp.OK, resp.Error)
	})

	t.Run("users.get", func(t *testing.T) {
		t.Parallel()
		h := newPluginHostHarness(t, manifest)
		h.users.EXPECT().Get(h.ctx, h.userID).Return(&service.User{
			ID:        h.userID,
			FirstName: "Ada",
			LastName:  "Lovelace",
		}, nil)
		resp, err := service.CallPluginHost(h.ctx, h.svc, manifest.ID, elemoplugin.HostRequest{
			Method:  "users.get",
			Payload: mustJSON(t, map[string]string{"id": h.userID.String()}),
		})
		require.NoError(t, err)
		require.True(t, resp.OK, resp.Error)
	})

	t.Run("permissions.check", func(t *testing.T) {
		t.Parallel()
		h := newPluginHostHarness(t, manifest)
		resp, err := service.CallPluginHost(h.ctx, h.svc, manifest.ID, elemoplugin.HostRequest{
			Method: "permissions.check",
			Payload: mustJSON(t, map[string]string{
				"resourceId":   h.issueID.String(),
				"resourceType": model.ResourceTypeIssue.String(),
				"action":       model.ActionIssueRead.String(),
			}),
		})
		require.NoError(t, err)
		require.True(t, resp.OK, resp.Error)
		assert.JSONEq(t, `{"allowed":true}`, string(resp.Data))
	})

	t.Run("plugin.storage.get", func(t *testing.T) {
		t.Parallel()
		h := newPluginHostHarness(t, manifest)
		h.repo.EXPECT().GetStorage(h.ctx, manifest.ID, h.orgID, "timer").Return(&model.PluginStorageEntry{
			PluginID: manifest.ID,
			ScopeID:  h.orgID,
			Key:      "timer",
			Value:    []byte(`{"running":true}`),
		}, nil)
		resp, err := service.CallPluginHost(h.ctx, h.svc, manifest.ID, elemoplugin.HostRequest{
			Method:  "plugin.storage.get",
			ScopeID: h.orgID.Composite(),
			Payload: mustJSON(t, map[string]string{"key": "timer"}),
		})
		require.NoError(t, err)
		require.True(t, resp.OK, resp.Error)
		assert.JSONEq(t, `{"running":true}`, string(resp.Data))
	})

	t.Run("plugin.storage.set", func(t *testing.T) {
		t.Parallel()
		h := newPluginHostHarness(t, manifest)
		h.repo.EXPECT().SetStorage(h.ctx, gomock.Any()).Return(&model.PluginStorageEntry{}, nil)
		resp, err := service.CallPluginHost(h.ctx, h.svc, manifest.ID, elemoplugin.HostRequest{
			Method:  "plugin.storage.set",
			ScopeID: h.orgID.Composite(),
			Payload: json.RawMessage(`{"key":"timer","value":{"running":true}}`),
		})
		require.NoError(t, err)
		require.True(t, resp.OK, resp.Error)
	})

	t.Run("plugin.storage.delete", func(t *testing.T) {
		t.Parallel()
		h := newPluginHostHarness(t, manifest)
		h.repo.EXPECT().DeleteStorage(h.ctx, manifest.ID, h.orgID, "timer").Return(nil)
		resp, err := service.CallPluginHost(h.ctx, h.svc, manifest.ID, elemoplugin.HostRequest{
			Method:  "plugin.storage.delete",
			ScopeID: h.orgID.Composite(),
			Payload: mustJSON(t, map[string]string{"key": "timer"}),
		})
		require.NoError(t, err)
		require.True(t, resp.OK, resp.Error)
	})

	t.Run("plugin.storage.list", func(t *testing.T) {
		t.Parallel()
		h := newPluginHostHarness(t, manifest)
		h.repo.EXPECT().ListStorage(h.ctx, manifest.ID, h.orgID).Return([]*model.PluginStorageEntry{
			{Key: "timer"},
		}, nil)
		resp, err := service.CallPluginHost(h.ctx, h.svc, manifest.ID, elemoplugin.HostRequest{
			Method:  "plugin.storage.list",
			ScopeID: h.orgID.Composite(),
		})
		require.NoError(t, err)
		require.True(t, resp.OK, resp.Error)
		assert.JSONEq(t, `["timer"]`, string(resp.Data))
	})

	t.Run("graph.nodes.create", func(t *testing.T) {
		t.Parallel()
		h := newPluginHostHarness(t, manifest)
		created, err := model.NewExtension(manifest.ID, "TimeEntry", map[string]any{"seconds": int64(12)})
		require.NoError(t, err)
		created.Parent = &h.issueID
		h.extRepo.EXPECT().Create(h.ctx, gomock.Any()).Return(created, nil)
		resp, err := service.CallPluginHost(h.ctx, h.svc, manifest.ID, elemoplugin.HostRequest{
			Method: "graph.nodes.create",
			Payload: mustJSON(t, map[string]any{
				"kind":       "TimeEntry",
				"parentId":   h.issueID.Composite(),
				"parentType": model.ResourceTypeIssue.String(),
				"properties": map[string]any{"seconds": 12},
			}),
		})
		require.NoError(t, err)
		require.True(t, resp.OK, resp.Error)
	})

	t.Run("graph.nodes.get", func(t *testing.T) {
		t.Parallel()
		h := newPluginHostHarness(t, manifest)
		ext, err := model.NewExtension(manifest.ID, "TimeEntry", map[string]any{"seconds": int64(12)})
		require.NoError(t, err)
		h.extRepo.EXPECT().Get(h.ctx, manifest.ID, ext.ID).Return(ext, nil)
		resp, err := service.CallPluginHost(h.ctx, h.svc, manifest.ID, elemoplugin.HostRequest{
			Method:  "graph.nodes.get",
			Payload: mustJSON(t, map[string]string{"id": ext.ID.String()}),
		})
		require.NoError(t, err)
		require.True(t, resp.OK, resp.Error)
	})

	t.Run("graph.nodes.update", func(t *testing.T) {
		t.Parallel()
		h := newPluginHostHarness(t, manifest)
		ext, err := model.NewExtension(manifest.ID, "TimeEntry", map[string]any{"seconds": int64(12)})
		require.NoError(t, err)
		h.extRepo.EXPECT().Get(h.ctx, manifest.ID, ext.ID).Return(ext, nil)
		h.extRepo.EXPECT().Update(h.ctx, manifest.ID, ext.ID, gomock.Any()).Return(ext, nil)
		resp, err := service.CallPluginHost(h.ctx, h.svc, manifest.ID, elemoplugin.HostRequest{
			Method: "graph.nodes.update",
			Payload: mustJSON(t, map[string]any{
				"id":         ext.ID.String(),
				"properties": map[string]any{"seconds": 90},
			}),
		})
		require.NoError(t, err)
		require.True(t, resp.OK, resp.Error)
	})

	t.Run("graph.nodes.delete", func(t *testing.T) {
		t.Parallel()
		h := newPluginHostHarness(t, manifest)
		ext, err := model.NewExtension(manifest.ID, "TimeEntry", map[string]any{"seconds": int64(12)})
		require.NoError(t, err)
		h.extRepo.EXPECT().Get(h.ctx, manifest.ID, ext.ID).Return(ext, nil)
		h.extRepo.EXPECT().Delete(h.ctx, manifest.ID, ext.ID).Return(nil)
		resp, err := service.CallPluginHost(h.ctx, h.svc, manifest.ID, elemoplugin.HostRequest{
			Method:  "graph.nodes.delete",
			Payload: mustJSON(t, map[string]string{"id": ext.ID.String()}),
		})
		require.NoError(t, err)
		require.True(t, resp.OK, resp.Error)
	})

	t.Run("graph.nodes.list", func(t *testing.T) {
		t.Parallel()
		h := newPluginHostHarness(t, manifest)
		h.extRepo.EXPECT().List(h.ctx, gomock.Any()).Return(repository.Page[*model.Extension]{}, nil)
		resp, err := service.CallPluginHost(h.ctx, h.svc, manifest.ID, elemoplugin.HostRequest{
			Method: "graph.nodes.list",
			Payload: mustJSON(t, map[string]any{
				"kind":      "TimeEntry",
				"scopeId":   h.issueID.Composite(),
				"scopeType": model.ResourceTypeIssue.String(),
				"pageSize":  10,
			}),
		})
		require.NoError(t, err)
		require.True(t, resp.OK, resp.Error)
	})

	t.Run("graph.nodes.move", func(t *testing.T) {
		t.Parallel()
		h := newPluginHostHarness(t, manifest)
		ext, err := model.NewExtension(manifest.ID, "TimeEntry", map[string]any{"seconds": int64(12)})
		require.NoError(t, err)
		newParent := model.MustNewID(model.ResourceTypeIssue)
		h.extRepo.EXPECT().Get(h.ctx, manifest.ID, ext.ID).Return(ext, nil)
		h.extRepo.EXPECT().Move(h.ctx, gomock.Any()).Return(ext, nil)
		resp, err := service.CallPluginHost(h.ctx, h.svc, manifest.ID, elemoplugin.HostRequest{
			Method: "graph.nodes.move",
			Payload: mustJSON(t, map[string]string{
				"id":         ext.ID.String(),
				"parentId":   newParent.String(),
				"parentType": model.ResourceTypeIssue.String(),
			}),
		})
		require.NoError(t, err)
		require.True(t, resp.OK, resp.Error)
	})

	t.Run("graph.relations.create", func(t *testing.T) {
		t.Parallel()
		h := newPluginHostHarness(t, manifest)
		from, err := model.NewExtension(manifest.ID, "TimeEntry", map[string]any{"seconds": int64(12)})
		require.NoError(t, err)
		h.extRepo.EXPECT().Get(h.ctx, manifest.ID, from.ID).Return(from, nil)
		h.extRepo.EXPECT().CountRelations(h.ctx, manifest.ID, "LOGGED_BY", from.ID, h.userID).Return(int64(0), int64(0), nil)
		h.extRepo.EXPECT().CreateRelation(h.ctx, gomock.Any()).Return(&model.ExtensionRelation{
			ID:   "rel-1",
			Kind: "LOGGED_BY",
			From: from.ID,
			To:   h.userID,
		}, nil)
		resp, err := service.CallPluginHost(h.ctx, h.svc, manifest.ID, elemoplugin.HostRequest{
			Method: "graph.relations.create",
			Payload: mustJSON(t, map[string]string{
				"kind":     "LOGGED_BY",
				"fromId":   from.ID.String(),
				"fromType": model.ResourceTypeExtension.String(),
				"toId":     h.userID.String(),
				"toType":   model.ResourceTypeUser.String(),
			}),
		})
		require.NoError(t, err)
		require.True(t, resp.OK, resp.Error)
	})

	t.Run("graph.relations.delete", func(t *testing.T) {
		t.Parallel()
		h := newPluginHostHarness(t, manifest)
		h.extRepo.EXPECT().DeleteRelation(h.ctx, manifest.ID, "rel-1").Return(nil)
		resp, err := service.CallPluginHost(h.ctx, h.svc, manifest.ID, elemoplugin.HostRequest{
			Method:  "graph.relations.delete",
			Payload: mustJSON(t, map[string]string{"id": "rel-1"}),
		})
		require.NoError(t, err)
		require.True(t, resp.OK, resp.Error)
	})

	t.Run("graph.relations.list", func(t *testing.T) {
		t.Parallel()
		h := newPluginHostHarness(t, manifest)
		h.extRepo.EXPECT().ListRelations(
			h.ctx,
			manifest.ID,
			"LOGGED_ON",
			h.issueID,
			model.PluginGraphRelationDirectionIncoming,
			gomock.Any(),
		).Return(repository.Page[*model.ExtensionRelation]{}, nil)
		resp, err := service.CallPluginHost(h.ctx, h.svc, manifest.ID, elemoplugin.HostRequest{
			Method: "graph.relations.list",
			Payload: mustJSON(t, map[string]any{
				"kind":      "LOGGED_ON",
				"nodeId":    h.issueID.String(),
				"nodeType":  model.ResourceTypeIssue.String(),
				"direction": "incoming",
				"pageSize":  10,
			}),
		})
		require.NoError(t, err)
		require.True(t, resp.OK, resp.Error)
	})

	t.Run("events.publish", func(t *testing.T) {
		t.Parallel()
		h := newPluginHostHarness(t, manifest)
		resp, err := service.CallPluginHost(h.ctx, h.svc, manifest.ID, elemoplugin.HostRequest{
			Method: "events.publish",
		})
		require.NoError(t, err)
		require.False(t, resp.OK)
		assert.Contains(t, resp.Error, "reserved")
	})

	t.Run("unknown method", func(t *testing.T) {
		t.Parallel()
		h := newPluginHostHarness(t, manifest)
		resp, err := service.CallPluginHost(h.ctx, h.svc, manifest.ID, elemoplugin.HostRequest{Method: "neo4j.query"})
		require.NoError(t, err)
		require.False(t, resp.OK)
		assert.Contains(t, resp.Error, elemoplugin.ErrUnknownHostMethod.Error())
	})

	t.Run("missing capability", func(t *testing.T) {
		t.Parallel()
		denied := hostPluginManifest()
		denied.Capabilities = nil
		h := newPluginHostHarness(t, denied)
		resp, err := service.CallPluginHost(h.ctx, h.svc, denied.ID, elemoplugin.HostRequest{
			Method:  "issues.get",
			Payload: mustJSON(t, map[string]string{"id": h.issueID.String()}),
		})
		require.NoError(t, err)
		require.False(t, resp.OK)
		assert.Contains(t, resp.Error, elemoplugin.ErrCapabilityDenied.Error())
	})

	t.Run("bad scope", func(t *testing.T) {
		t.Parallel()
		h := newPluginHostHarness(t, manifest)
		resp, err := service.CallPluginHost(h.ctx, h.svc, manifest.ID, elemoplugin.HostRequest{
			Method:  "plugin.storage.get",
			ScopeID: "not-a-valid-id",
			Payload: mustJSON(t, map[string]string{"key": "timer"}),
		})
		require.NoError(t, err)
		require.False(t, resp.OK)
	})

	t.Run("empty storage scope", func(t *testing.T) {
		t.Parallel()
		h := newPluginHostHarness(t, manifest)
		resp, err := service.CallPluginHost(h.ctx, h.svc, manifest.ID, elemoplugin.HostRequest{
			Method:  "plugin.storage.get",
			Payload: mustJSON(t, map[string]string{"key": "timer"}),
		})
		require.NoError(t, err)
		require.False(t, resp.OK)
	})
}

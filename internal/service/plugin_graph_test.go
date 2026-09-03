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

func timeTrackingGraphManifest() model.PluginManifest {
	return model.PluginManifest{
		SchemaVersion: 1,
		ID:            "com.elemo.timetracking",
		Name:          "Time Tracking",
		Version:       "1.1.0",
		Requires:      model.PluginRequires{PluginAPI: "^1"},
		Frontend:      &model.PluginFrontendDecl{Entry: "frontend/index.js"},
		Capabilities: []model.PluginCapability{
			model.CapabilityIssuesRead,
			model.CapabilityUsersRead,
			model.CapabilityGraphRead,
			model.CapabilityGraphWrite,
		},
		Slots: []model.PluginUISlot{
			model.PluginSlotIssueSidebar,
			model.PluginSlotIssueActivity,
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

type pluginGraphHarness struct {
	ctx     context.Context
	userID  model.ID
	issueID model.ID
	extRepo *mockrepo.MockExtensionRepository
	svc     service.PluginService
}

func newPluginGraphHarness(t *testing.T, manifest model.PluginManifest) pluginGraphHarness {
	t.Helper()
	return newPluginGraphHarnessWithConfig(t, manifest, nil)
}

func newPluginGraphHarnessWithConfig(
	t *testing.T,
	manifest model.PluginManifest,
	activationConfig json.RawMessage,
) pluginGraphHarness {
	t.Helper()
	ctrl := gomock.NewController(t)
	userID := model.MustNewID(model.ResourceTypeUser)
	issueID := model.MustNewID(model.ResourceTypeIssue)
	ctx := context.WithValue(context.Background(), pkg.CtxKeyUserID, userID)

	span := mocktrace.NewMockSpan(ctrl)
	span.EXPECT().End(gomock.Len(0)).AnyTimes()
	tracer := mocktrace.NewMockTracer(ctrl)
	tracer.EXPECT().Start(gomock.Any(), gomock.Any(), gomock.Len(0)).Return(ctx, span).AnyTimes()

	lic := mocksvc.NewMockLicenseService(ctrl)
	perm := mocksvc.NewMockPermissionService(ctrl)
	repo := mockrepo.NewMockPluginRepository(ctrl)
	extRepo := mockrepo.NewMockExtensionRepository(ctrl)

	svc, err := service.NewPluginService(
		config.PluginConfig{Directory: t.TempDir()},
		repo,
		extRepo,
		perm,
		lic,
		nil, nil, nil, nil,
		service.WithLogger(mocklog.NewMockLogger(ctrl)),
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
				Config:   activationConfig,
			}
			if len(ancestry) > 0 {
				act.ScopeID = ancestry[0]
			}
			return []*model.PluginActivation{act}, nil
		},
	).AnyTimes()
	perm.EXPECT().CtxUserHas(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil).AnyTimes()

	return pluginGraphHarness{
		ctx:     ctx,
		userID:  userID,
		issueID: issueID,
		extRepo: extRepo,
		svc:     svc,
	}
}

func TestPluginService_CreateNodeStampsUserID(t *testing.T) {
	t.Parallel()

	manifest := timeTrackingGraphManifest()
	require.NoError(t, manifest.Validate())
	h := newPluginGraphHarness(t, manifest)

	h.extRepo.EXPECT().Create(h.ctx, gomock.Any()).DoAndReturn(
		func(_ context.Context, opts repository.CreateExtensionOpts) (*model.Extension, error) {
			assert.Equal(t, h.userID.String(), opts.Properties["user_id"])
			assert.Equal(t, int64(60), opts.Properties["seconds"])
			ext, err := model.NewExtension(manifest.ID, opts.Kind, opts.Properties)
			require.NoError(t, err)
			parent := h.issueID
			ext.Parent = &parent
			return ext, nil
		},
	)

	created, err := h.svc.CreateNode(h.ctx, manifest.ID, service.CreateExtensionNodeOpts{
		Kind:   "TimeEntry",
		Parent: h.issueID,
		Properties: map[string]any{
			"seconds": 60,
			"user_id": "spoofed-user",
		},
	})
	require.NoError(t, err)
	require.NotNil(t, created.Parent)
	assert.Equal(t, h.issueID, *created.Parent)
	assert.Equal(t, h.userID.String(), created.Properties["user_id"])
}

func TestPluginService_ListNodesReturnsParent(t *testing.T) {
	t.Parallel()

	manifest := timeTrackingGraphManifest()
	require.NoError(t, manifest.Validate())
	h := newPluginGraphHarness(t, manifest)

	ext, err := model.NewExtension(manifest.ID, "TimeEntry", map[string]any{"seconds": int64(15)})
	require.NoError(t, err)
	parent := h.issueID
	ext.Parent = &parent

	h.extRepo.EXPECT().List(h.ctx, gomock.Any()).DoAndReturn(
		func(_ context.Context, filter repository.ListExtensionFilter) (repository.Page[*model.Extension], error) {
			assert.Equal(t, "TimeEntry", filter.Kind)
			assert.Equal(t, h.issueID, filter.Scope)
			assert.NotNil(t, filter.Page.Token)
			assert.Equal(t, "cursor-1", *filter.Page.Token)
			return repository.Page[*model.Extension]{Items: []*model.Extension{ext}}, nil
		},
	)

	token := "cursor-1"
	page, err := h.svc.ListNodes(h.ctx, manifest.ID, service.ListExtensionNodeOpts{
		Kind:  "TimeEntry",
		Scope: h.issueID,
		Page:  repository.CursorPage{Size: 50, Token: &token},
	})
	require.NoError(t, err)
	require.Len(t, page.Items, 1)
	require.NotNil(t, page.Items[0].Parent)
	assert.Equal(t, h.issueID, *page.Items[0].Parent)
}

func TestPluginService_MoveNodeRetargetsDomainEdges(t *testing.T) {
	t.Parallel()

	manifest := timeTrackingGraphManifest()
	require.NoError(t, manifest.Validate())
	h := newPluginGraphHarness(t, manifest)
	target := model.MustNewID(model.ResourceTypeIssue)

	existing, err := model.NewExtension(manifest.ID, "TimeEntry", map[string]any{
		"seconds": int64(90),
		"user_id": h.userID.String(),
	})
	require.NoError(t, err)
	oldParent := h.issueID
	existing.Parent = &oldParent

	relType, err := elemoplugin.RelationType(manifest.ID, "LOGGED_ON")
	require.NoError(t, err)

	h.extRepo.EXPECT().Get(h.ctx, manifest.ID, existing.ID).Return(existing, nil)
	h.extRepo.EXPECT().Move(h.ctx, gomock.Any()).DoAndReturn(
		func(_ context.Context, opts repository.MoveExtensionOpts) (*model.Extension, error) {
			assert.Equal(t, existing.ID, opts.ID)
			assert.Equal(t, target, opts.Parent)
			assert.Equal(t, []string{relType}, opts.RelationTypes)
			moved := *existing
			moved.Parent = &target
			return &moved, nil
		},
	)

	moved, err := h.svc.MoveNode(h.ctx, manifest.ID, existing.ID, target)
	require.NoError(t, err)
	require.NotNil(t, moved.Parent)
	assert.Equal(t, target, *moved.Parent)
}

func TestPluginService_CreateRelationToCallerUser(t *testing.T) {
	t.Parallel()

	manifest := timeTrackingGraphManifest()
	require.NoError(t, manifest.Validate())
	h := newPluginGraphHarness(t, manifest)

	from, err := model.NewExtension(manifest.ID, "TimeEntry", map[string]any{
		"seconds": int64(30),
		"user_id": h.userID.String(),
	})
	require.NoError(t, err)

	h.extRepo.EXPECT().Get(h.ctx, manifest.ID, from.ID).Return(from, nil)
	h.extRepo.EXPECT().CountRelations(h.ctx, manifest.ID, "LOGGED_BY", from.ID, h.userID).Return(int64(0), int64(0), nil)
	h.extRepo.EXPECT().CreateRelation(h.ctx, gomock.Any()).DoAndReturn(
		func(_ context.Context, opts repository.CreateExtensionRelationOpts) (*model.ExtensionRelation, error) {
			assert.Equal(t, "LOGGED_BY", opts.Kind)
			assert.Equal(t, from.ID, opts.From)
			assert.Equal(t, h.userID, opts.To)
			return &model.ExtensionRelation{
				ID:   "rel-1",
				Kind: opts.Kind,
				From: opts.From,
				To:   opts.To,
			}, nil
		},
	)

	rel, err := h.svc.CreateRelation(h.ctx, manifest.ID, service.CreateExtensionRelationOpts{
		Kind: "LOGGED_BY",
		From: from.ID,
		To:   h.userID,
	})
	require.NoError(t, err)
	assert.Equal(t, h.userID, rel.To)
}

func TestPluginService_CreateRelationEnforcesCardinality(t *testing.T) {
	t.Parallel()

	manifest := timeTrackingGraphManifest()
	require.NoError(t, manifest.Validate())
	h := newPluginGraphHarness(t, manifest)

	from, err := model.NewExtension(manifest.ID, "TimeEntry", map[string]any{
		"seconds": int64(30),
		"user_id": h.userID.String(),
	})
	require.NoError(t, err)

	h.extRepo.EXPECT().Get(h.ctx, manifest.ID, from.ID).Return(from, nil)
	h.extRepo.EXPECT().CountRelations(h.ctx, manifest.ID, "LOGGED_BY", from.ID, h.userID).Return(int64(1), int64(0), nil)

	_, err = h.svc.CreateRelation(h.ctx, manifest.ID, service.CreateExtensionRelationOpts{
		Kind: "LOGGED_BY",
		From: from.ID,
		To:   h.userID,
	})
	require.ErrorIs(t, err, model.ErrPluginRelationCardinality)
}

func TestPluginService_ListRelationsPassesDirection(t *testing.T) {
	t.Parallel()

	manifest := timeTrackingGraphManifest()
	require.NoError(t, manifest.Validate())
	h := newPluginGraphHarness(t, manifest)

	h.extRepo.EXPECT().ListRelations(
		h.ctx,
		manifest.ID,
		"LOGGED_ON",
		h.issueID,
		model.PluginGraphRelationDirectionIncoming,
		gomock.Any(),
	).Return(repository.Page[*model.ExtensionRelation]{}, nil)

	_, err := h.svc.ListRelations(h.ctx, manifest.ID, service.ListExtensionRelationOpts{
		Kind:      "LOGGED_ON",
		Node:      h.issueID,
		Direction: model.PluginGraphRelationDirectionIncoming,
	})
	require.NoError(t, err)
}

func TestPluginService_CreateRelationRejectsOtherUser(t *testing.T) {
	t.Parallel()

	manifest := timeTrackingGraphManifest()
	require.NoError(t, manifest.Validate())
	h := newPluginGraphHarness(t, manifest)

	from, err := model.NewExtension(manifest.ID, "TimeEntry", map[string]any{
		"seconds": int64(30),
	})
	require.NoError(t, err)
	other := model.MustNewID(model.ResourceTypeUser)

	h.extRepo.EXPECT().Get(h.ctx, manifest.ID, from.ID).Return(from, nil)
	h.extRepo.EXPECT().CountRelations(h.ctx, manifest.ID, "LOGGED_BY", from.ID, other).Return(int64(0), int64(0), nil)

	_, err = h.svc.CreateRelation(h.ctx, manifest.ID, service.CreateExtensionRelationOpts{
		Kind: "LOGGED_BY",
		From: from.ID,
		To:   other,
	})
	require.ErrorIs(t, err, service.ErrNoPermission)
}

func TestAssertAdditiveGraph_AllowsOptionalUserID(t *testing.T) {
	t.Parallel()

	oldSchema := &model.PluginGraphSchema{
		Nodes: []model.PluginGraphNodeDecl{
			{
				Kind:  "TimeEntry",
				Scope: model.PluginGraphNodeScope{Parent: "Issue"},
				Properties: []model.PluginGraphPropertyDecl{
					{Name: "seconds", Type: model.PluginGraphPropertyTypeInteger, Required: true},
					{Name: "note", Type: model.PluginGraphPropertyTypeStr},
				},
			},
		},
		Relations: []model.PluginGraphRelationDecl{
			{Kind: "LOGGED_ON", From: "TimeEntry", To: "Issue"},
		},
	}
	newSchema := timeTrackingGraphManifest().Graph

	err := service.AssertAdditiveGraph(oldSchema, newSchema)
	require.NoError(t, err)

	stripped := &model.PluginGraphSchema{
		Nodes: []model.PluginGraphNodeDecl{
			{
				Kind:  "TimeEntry",
				Scope: model.PluginGraphNodeScope{Parent: "Issue"},
				Properties: []model.PluginGraphPropertyDecl{
					{Name: "seconds", Type: model.PluginGraphPropertyTypeInteger, Required: true},
				},
			},
		},
	}
	err = service.AssertAdditiveGraph(oldSchema, stripped)
	require.ErrorIs(t, err, service.ErrPluginSchemaNotAdditive)
}

func accountingGraphManifest() model.PluginManifest {
	return model.PluginManifest{
		SchemaVersion: 1,
		ID:            "com.elemo.accounting",
		Name:          "Accounting",
		Version:       "1.0.0",
		Requires:      model.PluginRequires{PluginAPI: "^1"},
		Frontend:      &model.PluginFrontendDecl{Entry: "frontend/index.js"},
		Capabilities: []model.PluginCapability{
			model.CapabilityIssuesRead,
			model.CapabilityProjectsRead,
			model.CapabilityGraphRead,
			model.CapabilityGraphWrite,
		},
		Config: []model.PluginConfigFieldDecl{
			{Name: "time_source", Type: model.PluginConfigFieldTypeGraphBinding, Foreign: "LoggedTime"},
		},
		Graph: &model.PluginGraphSchema{
			Foreign: []model.PluginGraphForeignDecl{
				{
					Name:   "LoggedTime",
					Parent: "Issue",
					Properties: []model.PluginGraphPropertyDecl{
						{Name: "seconds", Type: model.PluginGraphPropertyTypeInteger, Required: true},
					},
				},
			},
			Nodes: []model.PluginGraphNodeDecl{
				{
					Kind:  "Budget",
					Scope: model.PluginGraphNodeScope{Parent: "Organization"},
					Properties: []model.PluginGraphPropertyDecl{
						{Name: "seconds", Type: model.PluginGraphPropertyTypeInteger, Required: true},
					},
				},
			},
			Relations: []model.PluginGraphRelationDecl{
				{
					Kind:        "COUNTED_AGAINST",
					From:        "LoggedTime",
					To:          "Budget",
					Cardinality: model.PluginGraphCardinalityManyToOne,
				},
			},
		},
	}
}

func TestPluginService_CreateRelationForeignBinding(t *testing.T) {
	t.Parallel()

	manifest := accountingGraphManifest()
	require.NoError(t, manifest.Validate())
	binding := json.RawMessage(`{"time_source":{"plugin_id":"com.elemo.timetracking","kind":"TimeEntry"}}`)
	h := newPluginGraphHarnessWithConfig(t, manifest, binding)

	from, err := model.NewExtension("com.elemo.timetracking", "TimeEntry", map[string]any{
		"seconds": int64(60),
	})
	require.NoError(t, err)
	to, err := model.NewExtension(manifest.ID, "Budget", map[string]any{
		"seconds": int64(3600),
	})
	require.NoError(t, err)

	h.extRepo.EXPECT().Get(h.ctx, "com.elemo.timetracking", from.ID).Return(from, nil)
	h.extRepo.EXPECT().Get(h.ctx, manifest.ID, to.ID).Return(to, nil)
	h.extRepo.EXPECT().CountRelations(h.ctx, manifest.ID, "COUNTED_AGAINST", from.ID, to.ID).Return(int64(0), int64(0), nil)
	h.extRepo.EXPECT().CreateRelation(h.ctx, gomock.Any()).DoAndReturn(
		func(_ context.Context, opts repository.CreateExtensionRelationOpts) (*model.ExtensionRelation, error) {
			assert.Equal(t, "COUNTED_AGAINST", opts.Kind)
			assert.Equal(t, from.ID, opts.From)
			assert.Equal(t, to.ID, opts.To)
			return &model.ExtensionRelation{ID: "rel-foreign", Kind: opts.Kind, From: opts.From, To: opts.To}, nil
		},
	)

	rel, err := h.svc.CreateRelation(h.ctx, manifest.ID, service.CreateExtensionRelationOpts{
		Kind: "COUNTED_AGAINST",
		From: from.ID,
		To:   to.ID,
	})
	require.NoError(t, err)
	assert.Equal(t, "rel-foreign", rel.ID)
}

func TestPluginService_CreateRelationRejectsUnboundForeign(t *testing.T) {
	t.Parallel()

	manifest := accountingGraphManifest()
	require.NoError(t, manifest.Validate())
	h := newPluginGraphHarness(t, manifest)

	from, err := model.NewExtension("com.elemo.timetracking", "TimeEntry", map[string]any{
		"seconds": int64(60),
	})
	require.NoError(t, err)
	to, err := model.NewExtension(manifest.ID, "Budget", map[string]any{
		"seconds": int64(3600),
	})
	require.NoError(t, err)

	_, err = h.svc.CreateRelation(h.ctx, manifest.ID, service.CreateExtensionRelationOpts{
		Kind: "COUNTED_AGAINST",
		From: from.ID,
		To:   to.ID,
	})
	require.ErrorIs(t, err, model.ErrPluginGraphBinding)
}

func TestPluginService_GetNode(t *testing.T) {
	t.Parallel()

	manifest := timeTrackingGraphManifest()
	require.NoError(t, manifest.Validate())
	h := newPluginGraphHarness(t, manifest)

	ext, err := model.NewExtension(manifest.ID, "TimeEntry", map[string]any{
		"seconds": int64(12),
		"user_id": h.userID.String(),
	})
	require.NoError(t, err)
	ext.Parent = &h.issueID

	h.extRepo.EXPECT().Get(h.ctx, manifest.ID, ext.ID).Return(ext, nil)

	got, err := h.svc.GetNode(h.ctx, manifest.ID, ext.ID, "")
	require.NoError(t, err)
	assert.Equal(t, ext.ID, got.ID)
}

func TestPluginService_GetNodeNotFound(t *testing.T) {
	t.Parallel()

	manifest := timeTrackingGraphManifest()
	require.NoError(t, manifest.Validate())
	h := newPluginGraphHarness(t, manifest)
	id := model.MustNewID(model.ResourceTypeExtension)

	h.extRepo.EXPECT().Get(h.ctx, manifest.ID, id).Return(nil, repository.ErrNotFound)

	_, err := h.svc.GetNode(h.ctx, manifest.ID, id, "")
	require.ErrorIs(t, err, repository.ErrNotFound)
}

func TestPluginService_UpdateNode(t *testing.T) {
	t.Parallel()

	manifest := timeTrackingGraphManifest()
	require.NoError(t, manifest.Validate())
	h := newPluginGraphHarness(t, manifest)

	ext, err := model.NewExtension(manifest.ID, "TimeEntry", map[string]any{
		"seconds": int64(12),
		"user_id": h.userID.String(),
	})
	require.NoError(t, err)

	h.extRepo.EXPECT().Get(h.ctx, manifest.ID, ext.ID).Return(ext, nil)
	h.extRepo.EXPECT().Update(h.ctx, manifest.ID, ext.ID, gomock.Any()).DoAndReturn(
		func(_ context.Context, _ string, _ model.ID, opts repository.UpdateExtensionOpts) (*model.Extension, error) {
			assert.Equal(t, int64(90), opts.Properties["seconds"])
			updated := *ext
			updated.Properties = opts.Properties
			return &updated, nil
		},
	)

	got, err := h.svc.UpdateNode(h.ctx, manifest.ID, ext.ID, map[string]any{"seconds": 90})
	require.NoError(t, err)
	assert.Equal(t, int64(90), got.Properties["seconds"])
}

func TestPluginService_DeleteNode(t *testing.T) {
	t.Parallel()

	manifest := timeTrackingGraphManifest()
	require.NoError(t, manifest.Validate())
	h := newPluginGraphHarness(t, manifest)

	ext, err := model.NewExtension(manifest.ID, "TimeEntry", map[string]any{
		"seconds": int64(12),
		"user_id": h.userID.String(),
	})
	require.NoError(t, err)

	h.extRepo.EXPECT().Get(h.ctx, manifest.ID, ext.ID).Return(ext, nil)
	h.extRepo.EXPECT().Delete(h.ctx, manifest.ID, ext.ID).Return(nil)

	require.NoError(t, h.svc.DeleteNode(h.ctx, manifest.ID, ext.ID))
}

func TestPluginService_DeleteRelation(t *testing.T) {
	t.Parallel()

	manifest := timeTrackingGraphManifest()
	require.NoError(t, manifest.Validate())
	h := newPluginGraphHarness(t, manifest)

	h.extRepo.EXPECT().DeleteRelation(h.ctx, manifest.ID, "rel-1").Return(nil)
	require.NoError(t, h.svc.DeleteRelation(h.ctx, manifest.ID, "rel-1"))
}

func TestPluginService_DeleteRelationNotFound(t *testing.T) {
	t.Parallel()

	manifest := timeTrackingGraphManifest()
	require.NoError(t, manifest.Validate())
	h := newPluginGraphHarness(t, manifest)

	h.extRepo.EXPECT().DeleteRelation(h.ctx, manifest.ID, "missing").Return(repository.ErrNotFound)
	err := h.svc.DeleteRelation(h.ctx, manifest.ID, "missing")
	require.ErrorIs(t, err, repository.ErrNotFound)
}

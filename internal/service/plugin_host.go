package service

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/opcotech/elemo/internal/config"
	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg/event"
	"github.com/opcotech/elemo/internal/pkg/optional"
	elemoplugin "github.com/opcotech/elemo/internal/plugin"
	"github.com/opcotech/elemo/internal/repository"
)

type pluginHost struct {
	plugins    *pluginService
	issues     IssueService
	projects   ProjectService
	users      UserService
	permission PermissionService
}

func (h *pluginHost) Call(ctx context.Context, pluginID string, req elemoplugin.HostRequest) (elemoplugin.HostResponse, error) {
	scope, err := parseScopeID(req.ScopeID)
	if err != nil && req.ScopeID != "" {
		return elemoplugin.HostError(err), nil
	}

	// Active plugins may read their own activation config without plugin.storage.read.
	if req.Method == "plugin.config.get" {
		if _, err := h.plugins.repo.GetInstallation(ctx, pluginID); err != nil {
			return elemoplugin.HostError(err), nil
		}
		return h.configGet(ctx, pluginID, scope)
	}

	capability, err := elemoplugin.CapabilityForMethod(req.Method)
	if err != nil {
		return elemoplugin.HostError(err), nil
	}
	inst, err := h.plugins.repo.GetInstallation(ctx, pluginID)
	if err != nil {
		return elemoplugin.HostError(err), nil
	}
	if err := elemoplugin.RequireCapability(inst.Manifest, capability); err != nil {
		return elemoplugin.HostError(err), nil
	}

	switch req.Method {
	case "issues.get":
		return h.issuesGet(ctx, req.Payload)
	case "issues.list":
		return h.issuesList(ctx, req.Payload)
	case "issues.update":
		return h.issuesUpdate(ctx, req.Payload)
	case "projects.get":
		return h.projectsGet(ctx, req.Payload)
	case "projects.list":
		return h.projectsList(ctx, req.Payload)
	case "users.get":
		return h.usersGet(ctx, req.Payload)
	case "permissions.check":
		return h.permissionsCheck(ctx, req.Payload)
	case "plugin.storage.get":
		return h.storageGet(ctx, pluginID, scope, req.Payload)
	case "plugin.storage.set":
		return h.storageSet(ctx, pluginID, scope, req.Payload)
	case "plugin.storage.delete":
		return h.storageDelete(ctx, pluginID, scope, req.Payload)
	case "plugin.storage.list":
		return h.storageList(ctx, pluginID, scope)
	case "graph.nodes.create":
		return h.graphCreate(ctx, pluginID, req.Payload)
	case "graph.nodes.get":
		return h.graphGet(ctx, pluginID, req.Payload)
	case "graph.nodes.update":
		return h.graphUpdate(ctx, pluginID, req.Payload)
	case "graph.nodes.delete":
		return h.graphDelete(ctx, pluginID, req.Payload)
	case "graph.nodes.list":
		return h.graphList(ctx, pluginID, req.Payload)
	case "graph.nodes.move":
		return h.graphMove(ctx, pluginID, req.Payload)
	case "graph.relations.create":
		return h.graphRelCreate(ctx, pluginID, req.Payload)
	case "graph.relations.delete":
		return h.graphRelDelete(ctx, pluginID, req.Payload)
	case "graph.relations.list":
		return h.graphRelList(ctx, pluginID, req.Payload)
	case "events.publish":
		return elemoplugin.HostError(errors.New("events.publish is reserved")), nil
	default:
		return elemoplugin.HostError(elemoplugin.ErrUnknownHostMethod), nil
	}
}

func (h *pluginHost) issuesGet(ctx context.Context, payload json.RawMessage) (elemoplugin.HostResponse, error) {
	var body struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(payload, &body); err != nil {
		return elemoplugin.HostError(err), nil
	}
	id, err := model.NewIDFromString(body.ID, model.ResourceTypeIssue.String())
	if err != nil {
		return elemoplugin.HostError(err), nil
	}
	issue, err := h.issues.Get(ctx, id)
	if err != nil {
		return elemoplugin.HostError(err), nil
	}
	out := map[string]any{
		"id":    issue.ID.String(),
		"key":   issue.Key,
		"title": issue.Title,
	}
	if issue.Project != nil {
		out["projectId"] = issue.Project.ID.String()
	}
	return elemoplugin.HostOK(out)
}

func (h *pluginHost) issuesList(ctx context.Context, payload json.RawMessage) (elemoplugin.HostResponse, error) {
	var body struct {
		ProjectID string `json:"projectId"`
	}
	if err := json.Unmarshal(payload, &body); err != nil {
		return elemoplugin.HostError(err), nil
	}
	projectID, err := model.NewIDFromString(body.ProjectID, model.ResourceTypeProject.String())
	if err != nil {
		return elemoplugin.HostError(err), nil
	}
	page, err := h.issues.List(ctx, projectID, CursorPage{Size: 50}, IssueListOptions{})
	if err != nil {
		return elemoplugin.HostError(err), nil
	}
	items := make([]map[string]any, 0, len(page.Items))
	for _, issue := range page.Items {
		item := map[string]any{"id": issue.ID.String(), "key": issue.Key, "title": issue.Title}
		if issue.Project != nil {
			item["projectId"] = issue.Project.ID.String()
		}
		items = append(items, item)
	}
	return elemoplugin.HostOK(items)
}

func (h *pluginHost) issuesUpdate(ctx context.Context, payload json.RawMessage) (elemoplugin.HostResponse, error) {
	var body struct {
		ID    string `json:"id"`
		Title string `json:"title"`
	}
	if err := json.Unmarshal(payload, &body); err != nil {
		return elemoplugin.HostError(err), nil
	}
	id, err := model.NewIDFromString(body.ID, model.ResourceTypeIssue.String())
	if err != nil {
		return elemoplugin.HostError(err), nil
	}
	opts := UpdateIssueOpts{}
	if body.Title != "" {
		opts.Title = optional.Some(body.Title)
	}
	issue, err := h.issues.Update(ctx, id, opts)
	if err != nil {
		return elemoplugin.HostError(err), nil
	}
	return elemoplugin.HostOK(map[string]any{"id": issue.ID.String(), "title": issue.Title})
}

func (h *pluginHost) projectsGet(ctx context.Context, payload json.RawMessage) (elemoplugin.HostResponse, error) {
	var body struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(payload, &body); err != nil {
		return elemoplugin.HostError(err), nil
	}
	id, err := model.NewIDFromString(body.ID, model.ResourceTypeProject.String())
	if err != nil {
		return elemoplugin.HostError(err), nil
	}
	project, err := h.projects.Get(ctx, id)
	if err != nil {
		return elemoplugin.HostError(err), nil
	}
	return elemoplugin.HostOK(map[string]any{"id": project.ID.String(), "key": project.Key, "name": project.Name})
}

func (h *pluginHost) projectsList(ctx context.Context, payload json.RawMessage) (elemoplugin.HostResponse, error) {
	var body struct {
		NamespaceID string `json:"namespaceId"`
	}
	if err := json.Unmarshal(payload, &body); err != nil {
		return elemoplugin.HostError(err), nil
	}
	ns, err := model.NewIDFromString(body.NamespaceID, model.ResourceTypeNamespace.String())
	if err != nil {
		return elemoplugin.HostError(err), nil
	}
	page, err := h.projects.List(ctx, ns, CursorPage{Size: 50})
	if err != nil {
		return elemoplugin.HostError(err), nil
	}
	items := make([]map[string]any, 0, len(page.Items))
	for _, p := range page.Items {
		items = append(items, map[string]any{"id": p.ID.String(), "key": p.Key, "name": p.Name})
	}
	return elemoplugin.HostOK(items)
}

func (h *pluginHost) usersGet(ctx context.Context, payload json.RawMessage) (elemoplugin.HostResponse, error) {
	var body struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(payload, &body); err != nil {
		return elemoplugin.HostError(err), nil
	}
	id, err := model.NewIDFromString(body.ID, model.ResourceTypeUser.String())
	if err != nil {
		return elemoplugin.HostError(err), nil
	}
	user, err := h.users.Get(ctx, id)
	if err != nil {
		return elemoplugin.HostError(err), nil
	}
	return elemoplugin.HostOK(map[string]any{
		"id":         user.ID.String(),
		"first_name": user.FirstName,
		"last_name":  user.LastName,
	})
}

func (h *pluginHost) permissionsCheck(ctx context.Context, payload json.RawMessage) (elemoplugin.HostResponse, error) {
	var body struct {
		ResourceID   string `json:"resourceId"`
		ResourceType string `json:"resourceType"`
		Action       string `json:"action"`
	}
	if err := json.Unmarshal(payload, &body); err != nil {
		return elemoplugin.HostError(err), nil
	}
	id, err := model.NewIDFromString(body.ResourceID, body.ResourceType)
	if err != nil {
		return elemoplugin.HostError(err), nil
	}
	var action model.Action
	if err := action.UnmarshalText([]byte(body.Action)); err != nil {
		return elemoplugin.HostError(err), nil
	}
	ok, err := h.permission.CtxUserHas(ctx, id, action)
	if err != nil {
		return elemoplugin.HostError(err), nil
	}
	return elemoplugin.HostOK(map[string]any{"allowed": ok})
}

func (h *pluginHost) storageGet(ctx context.Context, pluginID string, scope model.ID, payload json.RawMessage) (elemoplugin.HostResponse, error) {
	if err := requireScopeRead(ctx, h.permission, scope); err != nil {
		return elemoplugin.HostError(err), nil
	}
	var body struct {
		Key string `json:"key"`
	}
	if err := json.Unmarshal(payload, &body); err != nil {
		return elemoplugin.HostError(err), nil
	}
	entry, err := h.plugins.repo.GetStorage(ctx, pluginID, scope, body.Key)
	if err != nil {
		return elemoplugin.HostError(err), nil
	}
	return elemoplugin.HostResponse{OK: true, Data: entry.Value}, nil
}

func (h *pluginHost) storageSet(ctx context.Context, pluginID string, scope model.ID, payload json.RawMessage) (elemoplugin.HostResponse, error) {
	if err := requireScopeUpdate(ctx, h.permission, scope); err != nil {
		return elemoplugin.HostError(err), nil
	}
	var body struct {
		Key   string          `json:"key"`
		Value json.RawMessage `json:"value"`
	}
	if err := json.Unmarshal(payload, &body); err != nil {
		return elemoplugin.HostError(err), nil
	}
	_, err := h.plugins.repo.SetStorage(ctx, &model.PluginStorageEntry{
		PluginID: pluginID, ScopeID: scope, Key: body.Key, Value: body.Value,
	})
	if err != nil {
		return elemoplugin.HostError(err), nil
	}
	return elemoplugin.HostOK(nil)
}

func (h *pluginHost) storageDelete(ctx context.Context, pluginID string, scope model.ID, payload json.RawMessage) (elemoplugin.HostResponse, error) {
	if err := requireScopeUpdate(ctx, h.permission, scope); err != nil {
		return elemoplugin.HostError(err), nil
	}
	var body struct {
		Key string `json:"key"`
	}
	if err := json.Unmarshal(payload, &body); err != nil {
		return elemoplugin.HostError(err), nil
	}
	if err := h.plugins.repo.DeleteStorage(ctx, pluginID, scope, body.Key); err != nil {
		return elemoplugin.HostError(err), nil
	}
	return elemoplugin.HostOK(nil)
}

func (h *pluginHost) storageList(ctx context.Context, pluginID string, scope model.ID) (elemoplugin.HostResponse, error) {
	if err := requireScopeRead(ctx, h.permission, scope); err != nil {
		return elemoplugin.HostError(err), nil
	}
	entries, err := h.plugins.repo.ListStorage(ctx, pluginID, scope)
	if err != nil {
		return elemoplugin.HostError(err), nil
	}
	keys := make([]string, 0, len(entries))
	for _, e := range entries {
		keys = append(keys, e.Key)
	}
	return elemoplugin.HostOK(keys)
}

func requireScopeRead(ctx context.Context, perm PermissionService, scope model.ID) error {
	if scope.IsNil() {
		return ErrNoPermission
	}
	action, ok := model.ReadActionFor(scope.Type)
	if !ok {
		return ErrNoPermission
	}
	return requireAction(ctx, perm, scope, action)
}

func requireScopeUpdate(ctx context.Context, perm PermissionService, scope model.ID) error {
	if scope.IsNil() {
		return ErrNoPermission
	}
	action, ok := model.UpdateActionFor(scope.Type)
	if !ok {
		return ErrNoPermission
	}
	return requireAction(ctx, perm, scope, action)
}

func (h *pluginHost) graphCreate(ctx context.Context, pluginID string, payload json.RawMessage) (elemoplugin.HostResponse, error) {
	var body struct {
		Kind       string         `json:"kind"`
		ParentID   string         `json:"parentId"`
		ParentType string         `json:"parentType"`
		Properties map[string]any `json:"properties"`
	}
	if err := json.Unmarshal(payload, &body); err != nil {
		return elemoplugin.HostError(err), nil
	}
	parent, err := parseTypedID(body.ParentID, body.ParentType)
	if err != nil {
		return elemoplugin.HostError(err), nil
	}
	ext, err := h.plugins.CreateNode(ctx, pluginID, CreateExtensionNodeOpts{
		Kind: body.Kind, Parent: parent, Properties: body.Properties,
	})
	if err != nil {
		return elemoplugin.HostError(err), nil
	}
	return elemoplugin.HostOK(extensionHostJSON(ext))
}

func (h *pluginHost) configGet(ctx context.Context, pluginID string, scope model.ID) (elemoplugin.HostResponse, error) {
	if scope.IsNil() {
		return elemoplugin.HostError(model.ErrInvalidID), nil
	}
	cfg, err := h.plugins.GetConfig(ctx, pluginID, scope)
	if err != nil {
		return elemoplugin.HostError(err), nil
	}
	var values any
	if err := json.Unmarshal(cfg, &values); err != nil {
		values = map[string]any{}
	}
	return elemoplugin.HostOK(values)
}

func (h *pluginHost) graphGet(ctx context.Context, pluginID string, payload json.RawMessage) (elemoplugin.HostResponse, error) {
	var body struct {
		ID            string `json:"id"`
		OwnerPluginID string `json:"ownerPluginId"`
	}
	if err := json.Unmarshal(payload, &body); err != nil {
		return elemoplugin.HostError(err), nil
	}
	id, err := model.NewIDFromString(body.ID, model.ResourceTypeExtension.String())
	if err != nil {
		return elemoplugin.HostError(err), nil
	}
	ext, err := h.plugins.GetNode(ctx, pluginID, id, body.OwnerPluginID)
	if err != nil {
		return elemoplugin.HostError(err), nil
	}
	return elemoplugin.HostOK(extensionHostJSON(ext))
}

func (h *pluginHost) graphUpdate(ctx context.Context, pluginID string, payload json.RawMessage) (elemoplugin.HostResponse, error) {
	var body struct {
		ID         string         `json:"id"`
		Properties map[string]any `json:"properties"`
	}
	if err := json.Unmarshal(payload, &body); err != nil {
		return elemoplugin.HostError(err), nil
	}
	id, err := model.NewIDFromString(body.ID, model.ResourceTypeExtension.String())
	if err != nil {
		return elemoplugin.HostError(err), nil
	}
	ext, err := h.plugins.UpdateNode(ctx, pluginID, id, body.Properties)
	if err != nil {
		return elemoplugin.HostError(err), nil
	}
	return elemoplugin.HostOK(extensionHostJSON(ext))
}

func (h *pluginHost) graphDelete(ctx context.Context, pluginID string, payload json.RawMessage) (elemoplugin.HostResponse, error) {
	var body struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(payload, &body); err != nil {
		return elemoplugin.HostError(err), nil
	}
	id, err := model.NewIDFromString(body.ID, model.ResourceTypeExtension.String())
	if err != nil {
		return elemoplugin.HostError(err), nil
	}
	if err := h.plugins.DeleteNode(ctx, pluginID, id); err != nil {
		return elemoplugin.HostError(err), nil
	}
	return elemoplugin.HostOK(nil)
}

func (h *pluginHost) graphList(ctx context.Context, pluginID string, payload json.RawMessage) (elemoplugin.HostResponse, error) {
	var body struct {
		Kind          string         `json:"kind"`
		ScopeID       string         `json:"scopeId"`
		ScopeType     string         `json:"scopeType"`
		Equals        map[string]any `json:"equals"`
		OwnerPluginID string         `json:"ownerPluginId"`
		PageSize      int            `json:"pageSize"`
	}
	if err := json.Unmarshal(payload, &body); err != nil {
		return elemoplugin.HostError(err), nil
	}
	scope, err := parseTypedID(body.ScopeID, body.ScopeType)
	if err != nil {
		return elemoplugin.HostError(err), nil
	}
	page, err := h.plugins.ListNodes(ctx, pluginID, ListExtensionNodeOpts{
		Kind:          body.Kind,
		Scope:         scope,
		Equals:        body.Equals,
		OwnerPluginID: body.OwnerPluginID,
		Page:          repository.CursorPage{Size: body.PageSize},
	})
	if err != nil {
		return elemoplugin.HostError(err), nil
	}
	return elemoplugin.HostOK(extensionsHostJSON(page.Items))
}

func (h *pluginHost) graphMove(ctx context.Context, pluginID string, payload json.RawMessage) (elemoplugin.HostResponse, error) {
	var body struct {
		ID         string `json:"id"`
		ParentID   string `json:"parentId"`
		ParentType string `json:"parentType"`
	}
	if err := json.Unmarshal(payload, &body); err != nil {
		return elemoplugin.HostError(err), nil
	}
	id, err := model.NewIDFromString(body.ID, model.ResourceTypeExtension.String())
	if err != nil {
		return elemoplugin.HostError(err), nil
	}
	parent, err := model.NewIDFromString(body.ParentID, body.ParentType)
	if err != nil {
		return elemoplugin.HostError(err), nil
	}
	ext, err := h.plugins.MoveNode(ctx, pluginID, id, parent)
	if err != nil {
		return elemoplugin.HostError(err), nil
	}
	return elemoplugin.HostOK(extensionHostJSON(ext))
}

func (h *pluginHost) graphRelCreate(ctx context.Context, pluginID string, payload json.RawMessage) (elemoplugin.HostResponse, error) {
	var body struct {
		Kind     string `json:"kind"`
		FromID   string `json:"fromId"`
		FromType string `json:"fromType"`
		ToID     string `json:"toId"`
		ToType   string `json:"toType"`
	}
	if err := json.Unmarshal(payload, &body); err != nil {
		return elemoplugin.HostError(err), nil
	}
	from, err := model.NewIDFromString(body.FromID, body.FromType)
	if err != nil {
		return elemoplugin.HostError(err), nil
	}
	to, err := model.NewIDFromString(body.ToID, body.ToType)
	if err != nil {
		return elemoplugin.HostError(err), nil
	}
	rel, err := h.plugins.CreateRelation(ctx, pluginID, CreateExtensionRelationOpts{
		Kind: body.Kind, From: from, To: to,
	})
	if err != nil {
		return elemoplugin.HostError(err), nil
	}
	return elemoplugin.HostOK(relationHostJSON(rel))
}

func (h *pluginHost) graphRelDelete(ctx context.Context, pluginID string, payload json.RawMessage) (elemoplugin.HostResponse, error) {
	var body struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(payload, &body); err != nil {
		return elemoplugin.HostError(err), nil
	}
	if err := h.plugins.DeleteRelation(ctx, pluginID, body.ID); err != nil {
		return elemoplugin.HostError(err), nil
	}
	return elemoplugin.HostOK(nil)
}

func (h *pluginHost) graphRelList(ctx context.Context, pluginID string, payload json.RawMessage) (elemoplugin.HostResponse, error) {
	var body struct {
		Kind      string `json:"kind"`
		NodeID    string `json:"nodeId"`
		NodeType  string `json:"nodeType"`
		Direction string `json:"direction"`
		PageSize  int    `json:"pageSize"`
	}
	if err := json.Unmarshal(payload, &body); err != nil {
		return elemoplugin.HostError(err), nil
	}
	node, err := model.NewIDFromString(body.NodeID, body.NodeType)
	if err != nil {
		return elemoplugin.HostError(err), nil
	}
	direction := model.PluginGraphRelationDirectionOutgoing
	if body.Direction != "" {
		if err := direction.UnmarshalText([]byte(body.Direction)); err != nil {
			return elemoplugin.HostError(err), nil
		}
	}
	page, err := h.plugins.ListRelations(ctx, pluginID, ListExtensionRelationOpts{
		Kind:      body.Kind,
		Node:      node,
		Direction: direction,
		Page:      repository.CursorPage{Size: body.PageSize},
	})
	if err != nil {
		return elemoplugin.HostError(err), nil
	}
	return elemoplugin.HostOK(relationsHostJSON(page.Items))
}

func extensionHostJSON(ext *model.Extension) map[string]any {
	if ext == nil {
		return nil
	}
	props := ext.Properties
	if props == nil {
		props = map[string]any{}
	}
	out := map[string]any{
		"id":         ext.ID.String(),
		"plugin_id":  ext.PluginID,
		"kind":       ext.Kind,
		"properties": props,
	}
	if ext.Parent != nil {
		out["parent_id"] = ext.Parent.String()
		out["parent_type"] = ext.Parent.Label()
	}
	if ext.CreatedAt != nil {
		out["created_at"] = ext.CreatedAt
	}
	if ext.UpdatedAt != nil {
		out["updated_at"] = ext.UpdatedAt
	}
	return out
}

func extensionsHostJSON(items []*model.Extension) []map[string]any {
	out := make([]map[string]any, 0, len(items))
	for _, ext := range items {
		if payload := extensionHostJSON(ext); payload != nil {
			out = append(out, payload)
		}
	}
	return out
}

func relationHostJSON(rel *model.ExtensionRelation) map[string]any {
	if rel == nil {
		return nil
	}
	return map[string]any{
		"id":        rel.ID,
		"kind":      rel.Kind,
		"from":      rel.From.String(),
		"from_type": rel.From.Type.String(),
		"to":        rel.To.String(),
		"to_type":   rel.To.Type.String(),
	}
}

func relationsHostJSON(items []*model.ExtensionRelation) []map[string]any {
	out := make([]map[string]any, 0, len(items))
	for _, rel := range items {
		if payload := relationHostJSON(rel); payload != nil {
			out = append(out, payload)
		}
	}
	return out
}

// NewPluginService constructs the plugin platform and host adapter.
func NewPluginService(
	conf config.PluginConfig,
	repo repository.PluginRepository,
	extensionRepo repository.ExtensionRepository,
	permissionService PermissionService,
	licenseService LicenseService,
	issueService IssueService,
	projectService ProjectService,
	userService UserService,
	bus *event.Bus,
	opts ...Option,
) (PluginService, error) {
	rt, err := newRuntime(opts...)
	if err != nil {
		return nil, err
	}
	if conf.Directory == "" {
		conf.Directory = "/var/lib/elemo/plugins"
	}
	if repo == nil {
		return nil, ErrNoPluginRepository
	}
	if extensionRepo == nil {
		return nil, ErrNoExtensionRepository
	}
	if permissionService == nil {
		return nil, ErrNoPermissionService
	}
	if licenseService == nil {
		return nil, ErrNoLicenseService
	}

	svc := &pluginService{
		runtime:           rt,
		conf:              conf,
		repo:              repo,
		extensionRepo:     extensionRepo,
		permissionService: permissionService,
		licenseService:    licenseService,
	}
	if bus != nil {
		svc.eventBus = bus
		svc.bus = bus
	}

	host := &pluginHost{
		plugins:    svc,
		issues:     issueService,
		projects:   projectService,
		users:      userService,
		permission: permissionService,
	}
	svc.host = host
	svc.registry = elemoplugin.NewRegistry(elemoplugin.NewWazeroRuntime(host, conf.ExecutionTimeout))
	svc.bindEvents()
	return svc, nil
}

package http

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"time"

	"github.com/opcotech/elemo/internal/model"
	elemoplugin "github.com/opcotech/elemo/internal/plugin"
	"github.com/opcotech/elemo/internal/service"
	"github.com/opcotech/elemo/internal/transport/http/api"
)

const maxPluginPackageBytes = 32 << 20

// PluginController is the controller for plugin endpoints.
type PluginController interface {
	V1PluginsGet(ctx context.Context, request api.V1PluginsGetRequestObject) (api.V1PluginsGetResponseObject, error)
	V1PluginsCreate(ctx context.Context, request api.V1PluginsCreateRequestObject) (api.V1PluginsCreateResponseObject, error)
	V1PluginsFrontendGet(ctx context.Context, request api.V1PluginsFrontendGetRequestObject) (api.V1PluginsFrontendGetResponseObject, error)
	V1PluginGet(ctx context.Context, request api.V1PluginGetRequestObject) (api.V1PluginGetResponseObject, error)
	V1PluginDelete(ctx context.Context, request api.V1PluginDeleteRequestObject) (api.V1PluginDeleteResponseObject, error)
	V1PluginEnable(ctx context.Context, request api.V1PluginEnableRequestObject) (api.V1PluginEnableResponseObject, error)
	V1PluginDisable(ctx context.Context, request api.V1PluginDisableRequestObject) (api.V1PluginDisableResponseObject, error)
	V1PluginConfigGet(ctx context.Context, request api.V1PluginConfigGetRequestObject) (api.V1PluginConfigGetResponseObject, error)
	V1PluginConfigPatch(ctx context.Context, request api.V1PluginConfigPatchRequestObject) (api.V1PluginConfigPatchResponseObject, error)
	V1PluginUpgrade(ctx context.Context, request api.V1PluginUpgradeRequestObject) (api.V1PluginUpgradeResponseObject, error)
	V1PluginInvoke(ctx context.Context, request api.V1PluginInvokeRequestObject) (api.V1PluginInvokeResponseObject, error)
	V1PluginGraphNodesGet(ctx context.Context, request api.V1PluginGraphNodesGetRequestObject) (api.V1PluginGraphNodesGetResponseObject, error)
	V1PluginGraphNodesCreate(ctx context.Context, request api.V1PluginGraphNodesCreateRequestObject) (api.V1PluginGraphNodesCreateResponseObject, error)
	V1PluginGraphNodeGet(ctx context.Context, request api.V1PluginGraphNodeGetRequestObject) (api.V1PluginGraphNodeGetResponseObject, error)
	V1PluginGraphNodeUpdate(ctx context.Context, request api.V1PluginGraphNodeUpdateRequestObject) (api.V1PluginGraphNodeUpdateResponseObject, error)
	V1PluginGraphNodeDelete(ctx context.Context, request api.V1PluginGraphNodeDeleteRequestObject) (api.V1PluginGraphNodeDeleteResponseObject, error)
	V1PluginGraphNodeMove(ctx context.Context, request api.V1PluginGraphNodeMoveRequestObject) (api.V1PluginGraphNodeMoveResponseObject, error)
	V1PluginGraphRelationsGet(ctx context.Context, request api.V1PluginGraphRelationsGetRequestObject) (api.V1PluginGraphRelationsGetResponseObject, error)
	V1PluginGraphRelationsCreate(ctx context.Context, request api.V1PluginGraphRelationsCreateRequestObject) (api.V1PluginGraphRelationsCreateResponseObject, error)
	V1PluginGraphRelationDelete(ctx context.Context, request api.V1PluginGraphRelationDeleteRequestObject) (api.V1PluginGraphRelationDeleteResponseObject, error)
	ServePluginAsset(w http.ResponseWriter, r *http.Request)
}

type pluginController struct {
	*baseController
	pluginService service.PluginService
}

func (c *pluginController) V1PluginsGet(
	ctx context.Context,
	request api.V1PluginsGetRequestObject,
) (api.V1PluginsGetResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1PluginsGet")
	defer span.End()

	if request.Params.ScopeId != nil || request.Params.ScopeType != nil {
		if request.Params.ScopeId == nil || request.Params.ScopeType == nil {
			return api.V1PluginsGet400JSONResponse{N400JSONResponse: formatBadRequest(errors.New("scope_id and scope_type are required together"))}, nil
		}
		scope, err := resourceIDFromAPI(*request.Params.ScopeType, *request.Params.ScopeId)
		if err != nil {
			return api.V1PluginsGet400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
		}
		items, err := c.pluginService.ListManaged(ctx, scope)
		if err != nil {
			switch classifyServiceError(err) {
			case http.StatusBadRequest:
				return api.V1PluginsGet400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
			case http.StatusForbidden:
				return api.V1PluginsGet403JSONResponse{N403JSONResponse: permissionDenied}, nil
			default:
				return api.V1PluginsGet500JSONResponse{N500JSONResponse: api.N500JSONResponse{Message: err.Error()}}, nil
			}
		}
		out := make([]api.Plugin, len(items))
		for i, item := range items {
			out[i] = pluginToDTO(item.Installation, item.Enabled, item.Config)
		}
		return api.V1PluginsGet200JSONResponse(out), nil
	}

	list, err := c.pluginService.List(ctx)
	if err != nil {
		switch classifyServiceError(err) {
		case http.StatusForbidden:
			return api.V1PluginsGet403JSONResponse{N403JSONResponse: permissionDenied}, nil
		default:
			return api.V1PluginsGet500JSONResponse{N500JSONResponse: api.N500JSONResponse{Message: err.Error()}}, nil
		}
	}
	out := make([]api.Plugin, len(list))
	for i, inst := range list {
		out[i] = pluginToDTO(inst, nil, nil)
	}
	return api.V1PluginsGet200JSONResponse(out), nil
}

func (c *pluginController) V1PluginsCreate(
	ctx context.Context,
	request api.V1PluginsCreateRequestObject,
) (api.V1PluginsCreateResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1PluginsCreate")
	defer span.End()

	zipBytes, err := readPluginPackage(request.Body)
	if err != nil {
		return api.V1PluginsCreate400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}
	inst, err := c.pluginService.Install(ctx, zipBytes)
	if err != nil {
		switch classifyServiceError(err) {
		case http.StatusBadRequest:
			return api.V1PluginsCreate400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
		case http.StatusForbidden:
			return api.V1PluginsCreate403JSONResponse{N403JSONResponse: permissionDenied}, nil
		case http.StatusConflict:
			return api.V1PluginsCreate409JSONResponse{N409JSONResponse: api.N409JSONResponse{Message: err.Error()}}, nil
		default:
			return api.V1PluginsCreate500JSONResponse{N500JSONResponse: api.N500JSONResponse{Message: err.Error()}}, nil
		}
	}
	return api.V1PluginsCreate201JSONResponse(pluginToDTO(inst, nil, nil)), nil
}

func (c *pluginController) V1PluginsFrontendGet(
	ctx context.Context,
	request api.V1PluginsFrontendGetRequestObject,
) (api.V1PluginsFrontendGetResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1PluginsFrontendGet")
	defer span.End()

	scope, err := resourceIDFromAPI(request.Params.ScopeType, request.Params.ScopeId)
	if err != nil {
		return api.V1PluginsFrontendGet400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}
	list, err := c.pluginService.ListFrontend(ctx, scope)
	if err != nil {
		switch classifyServiceError(err) {
		case http.StatusBadRequest:
			return api.V1PluginsFrontendGet400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
		case http.StatusForbidden:
			return api.V1PluginsFrontendGet403JSONResponse{N403JSONResponse: permissionDenied}, nil
		case http.StatusNotFound:
			return api.V1PluginsFrontendGet404JSONResponse{N404JSONResponse: notFound}, nil
		default:
			return api.V1PluginsFrontendGet500JSONResponse{N500JSONResponse: api.N500JSONResponse{Message: err.Error()}}, nil
		}
	}
	out := make([]api.FrontendPlugin, len(list))
	for i, p := range list {
		item := api.FrontendPlugin{
			Id:         p.ID,
			Version:    p.Version,
			Entrypoint: p.Entrypoint,
			Slots:      make([]api.PluginUISlot, len(p.Slots)),
		}
		if p.Module != "" {
			item.Module = &p.Module
		}
		for j, slot := range p.Slots {
			item.Slots[j] = api.PluginUISlot(slot)
		}
		out[i] = item
	}
	return api.V1PluginsFrontendGet200JSONResponse(out), nil
}

func (c *pluginController) V1PluginGet(
	ctx context.Context,
	request api.V1PluginGetRequestObject,
) (api.V1PluginGetResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1PluginGet")
	defer span.End()

	inst, err := c.pluginService.Get(ctx, request.PluginId)
	if err != nil {
		switch classifyServiceError(err) {
		case http.StatusForbidden:
			return api.V1PluginGet403JSONResponse{N403JSONResponse: permissionDenied}, nil
		case http.StatusNotFound:
			return api.V1PluginGet404JSONResponse{N404JSONResponse: notFound}, nil
		default:
			return api.V1PluginGet500JSONResponse{N500JSONResponse: api.N500JSONResponse{Message: err.Error()}}, nil
		}
	}
	return api.V1PluginGet200JSONResponse(pluginToDTO(inst, nil, nil)), nil
}

func (c *pluginController) V1PluginDelete(
	ctx context.Context,
	request api.V1PluginDeleteRequestObject,
) (api.V1PluginDeleteResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1PluginDelete")
	defer span.End()

	if err := c.pluginService.Uninstall(ctx, request.PluginId); err != nil {
		switch classifyServiceError(err) {
		case http.StatusForbidden:
			return api.V1PluginDelete403JSONResponse{N403JSONResponse: permissionDenied}, nil
		case http.StatusNotFound:
			return api.V1PluginDelete404JSONResponse{N404JSONResponse: notFound}, nil
		default:
			return api.V1PluginDelete500JSONResponse{N500JSONResponse: api.N500JSONResponse{Message: err.Error()}}, nil
		}
	}
	return api.V1PluginDelete204Response{}, nil
}

func (c *pluginController) V1PluginEnable(
	ctx context.Context,
	request api.V1PluginEnableRequestObject,
) (api.V1PluginEnableResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1PluginEnable")
	defer span.End()

	if request.Body == nil {
		return api.V1PluginEnable400JSONResponse{N400JSONResponse: formatBadRequest(errors.New("missing body"))}, nil
	}
	scope, err := resourceIDFromAPI(request.Body.ScopeType, request.Body.ScopeId)
	if err != nil {
		return api.V1PluginEnable400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}
	var config json.RawMessage
	if request.Body.Config != nil {
		config, err = json.Marshal(request.Body.Config)
		if err != nil {
			return api.V1PluginEnable400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
		}
	}
	if err := c.pluginService.Enable(ctx, request.PluginId, scope, config); err != nil {
		switch classifyServiceError(err) {
		case http.StatusBadRequest:
			return api.V1PluginEnable400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
		case http.StatusForbidden:
			return api.V1PluginEnable403JSONResponse{N403JSONResponse: permissionDenied}, nil
		case http.StatusNotFound:
			return api.V1PluginEnable404JSONResponse{N404JSONResponse: notFound}, nil
		default:
			return api.V1PluginEnable500JSONResponse{N500JSONResponse: api.N500JSONResponse{Message: err.Error()}}, nil
		}
	}
	return api.V1PluginEnable204Response{}, nil
}

func (c *pluginController) V1PluginDisable(
	ctx context.Context,
	request api.V1PluginDisableRequestObject,
) (api.V1PluginDisableResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1PluginDisable")
	defer span.End()

	if request.Body == nil {
		return api.V1PluginDisable400JSONResponse{N400JSONResponse: formatBadRequest(errors.New("missing body"))}, nil
	}
	scope, err := resourceIDFromAPI(request.Body.ScopeType, request.Body.ScopeId)
	if err != nil {
		return api.V1PluginDisable400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}
	if err := c.pluginService.Disable(ctx, request.PluginId, scope); err != nil {
		switch classifyServiceError(err) {
		case http.StatusBadRequest:
			return api.V1PluginDisable400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
		case http.StatusForbidden:
			return api.V1PluginDisable403JSONResponse{N403JSONResponse: permissionDenied}, nil
		case http.StatusNotFound:
			return api.V1PluginDisable404JSONResponse{N404JSONResponse: notFound}, nil
		default:
			return api.V1PluginDisable500JSONResponse{N500JSONResponse: api.N500JSONResponse{Message: err.Error()}}, nil
		}
	}
	return api.V1PluginDisable204Response{}, nil
}

func (c *pluginController) V1PluginConfigGet(
	ctx context.Context,
	request api.V1PluginConfigGetRequestObject,
) (api.V1PluginConfigGetResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1PluginConfigGet")
	defer span.End()

	scope, err := resourceIDFromAPI(request.Params.ScopeType, request.Params.ScopeId)
	if err != nil {
		return api.V1PluginConfigGet400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}
	raw, err := c.pluginService.GetManagedConfig(ctx, request.PluginId, scope)
	if err != nil {
		switch classifyServiceError(err) {
		case http.StatusBadRequest:
			return api.V1PluginConfigGet400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
		case http.StatusForbidden:
			return api.V1PluginConfigGet403JSONResponse{N403JSONResponse: permissionDenied}, nil
		case http.StatusNotFound:
			return api.V1PluginConfigGet404JSONResponse{N404JSONResponse: notFound}, nil
		default:
			return api.V1PluginConfigGet500JSONResponse{N500JSONResponse: api.N500JSONResponse{Message: err.Error()}}, nil
		}
	}
	return api.V1PluginConfigGet200JSONResponse{Config: rawConfigMap(raw)}, nil
}

func (c *pluginController) V1PluginConfigPatch(
	ctx context.Context,
	request api.V1PluginConfigPatchRequestObject,
) (api.V1PluginConfigPatchResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1PluginConfigPatch")
	defer span.End()

	if request.Body == nil {
		return api.V1PluginConfigPatch400JSONResponse{N400JSONResponse: formatBadRequest(errors.New("missing body"))}, nil
	}
	scope, err := resourceIDFromAPI(request.Params.ScopeType, request.Params.ScopeId)
	if err != nil {
		return api.V1PluginConfigPatch400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}
	raw, err := json.Marshal(request.Body.Config)
	if err != nil {
		return api.V1PluginConfigPatch400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}
	if err := c.pluginService.SetConfig(ctx, request.PluginId, scope, raw); err != nil {
		switch classifyServiceError(err) {
		case http.StatusBadRequest:
			return api.V1PluginConfigPatch400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
		case http.StatusForbidden:
			return api.V1PluginConfigPatch403JSONResponse{N403JSONResponse: permissionDenied}, nil
		case http.StatusNotFound:
			return api.V1PluginConfigPatch404JSONResponse{N404JSONResponse: notFound}, nil
		default:
			return api.V1PluginConfigPatch500JSONResponse{N500JSONResponse: api.N500JSONResponse{Message: err.Error()}}, nil
		}
	}
	return api.V1PluginConfigPatch200JSONResponse{Config: request.Body.Config}, nil
}

func (c *pluginController) V1PluginUpgrade(
	ctx context.Context,
	request api.V1PluginUpgradeRequestObject,
) (api.V1PluginUpgradeResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1PluginUpgrade")
	defer span.End()

	zipBytes, err := readPluginPackage(request.Body)
	if err != nil {
		return api.V1PluginUpgrade400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}
	inst, err := c.pluginService.Upgrade(ctx, request.PluginId, zipBytes)
	if err != nil {
		switch classifyServiceError(err) {
		case http.StatusBadRequest:
			return api.V1PluginUpgrade400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
		case http.StatusForbidden:
			return api.V1PluginUpgrade403JSONResponse{N403JSONResponse: permissionDenied}, nil
		case http.StatusNotFound:
			return api.V1PluginUpgrade404JSONResponse{N404JSONResponse: notFound}, nil
		default:
			return api.V1PluginUpgrade500JSONResponse{N500JSONResponse: api.N500JSONResponse{Message: err.Error()}}, nil
		}
	}
	return api.V1PluginUpgrade200JSONResponse(pluginToDTO(inst, nil, nil)), nil
}

func (c *pluginController) V1PluginInvoke(
	ctx context.Context,
	request api.V1PluginInvokeRequestObject,
) (api.V1PluginInvokeResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1PluginInvoke")
	defer span.End()

	if request.Body == nil {
		return api.V1PluginInvoke400JSONResponse{N400JSONResponse: formatBadRequest(errors.New("missing body"))}, nil
	}
	var payload json.RawMessage
	if request.Body.Payload != nil {
		payload, _ = json.Marshal(request.Body.Payload)
	}
	resp, err := c.pluginService.Invoke(ctx, request.PluginId, elemoplugin.InvokeRequest{
		Function: request.Body.Function,
		ScopeID:  request.Body.ScopeId,
		Payload:  payload,
	})
	if err != nil {
		switch classifyServiceError(err) {
		case http.StatusBadRequest:
			return api.V1PluginInvoke400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
		case http.StatusForbidden:
			return api.V1PluginInvoke403JSONResponse{N403JSONResponse: permissionDenied}, nil
		case http.StatusNotFound:
			return api.V1PluginInvoke404JSONResponse{N404JSONResponse: notFound}, nil
		default:
			return api.V1PluginInvoke500JSONResponse{N500JSONResponse: api.N500JSONResponse{Message: err.Error()}}, nil
		}
	}
	out := api.PluginInvokeResult{Ok: resp.OK}
	if resp.Error != "" {
		out.Error = &resp.Error
	}
	if len(resp.Data) > 0 {
		var data any
		if err := json.Unmarshal(resp.Data, &data); err == nil {
			out.Data = data
		}
	}
	return api.V1PluginInvoke200JSONResponse(out), nil
}

func (c *pluginController) V1PluginGraphNodesGet(
	ctx context.Context,
	request api.V1PluginGraphNodesGetRequestObject,
) (api.V1PluginGraphNodesGetResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1PluginGraphNodesGet")
	defer span.End()

	scope, err := resourceIDFromAPI(request.Params.ScopeType, request.Params.ScopeId)
	if err != nil {
		return api.V1PluginGraphNodesGet400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}
	page, err := cursorPageFromParams(request.Params.PageSize, request.Params.PageToken)
	if err != nil {
		return api.V1PluginGraphNodesGet400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}
	equals, err := parseEqualsFilter(request.Params.Equals)
	if err != nil {
		return api.V1PluginGraphNodesGet400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}
	ownerPluginID := ""
	if request.Params.OwnerPluginId != nil {
		ownerPluginID = *request.Params.OwnerPluginId
	}
	result, err := c.pluginService.ListNodes(ctx, request.PluginId, service.ListExtensionNodeOpts{
		Kind:          request.Params.Kind,
		Scope:         scope,
		Equals:        equals,
		OwnerPluginID: ownerPluginID,
		Page:          page,
	})
	if err != nil {
		switch classifyServiceError(err) {
		case http.StatusBadRequest:
			return api.V1PluginGraphNodesGet400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
		case http.StatusForbidden:
			return api.V1PluginGraphNodesGet403JSONResponse{N403JSONResponse: permissionDenied}, nil
		case http.StatusNotFound:
			return api.V1PluginGraphNodesGet404JSONResponse{N404JSONResponse: notFound}, nil
		default:
			return api.V1PluginGraphNodesGet500JSONResponse{N500JSONResponse: api.N500JSONResponse{Message: err.Error()}}, nil
		}
	}
	items := make([]api.ExtensionNode, len(result.Items))
	for i, n := range result.Items {
		items[i] = extensionToDTO(n)
	}
	return api.V1PluginGraphNodesGet200JSONResponse{
		Items:    items,
		PageInfo: pageInfoToDTO(result.PageInfo),
	}, nil
}

func (c *pluginController) V1PluginGraphNodesCreate(
	ctx context.Context,
	request api.V1PluginGraphNodesCreateRequestObject,
) (api.V1PluginGraphNodesCreateResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1PluginGraphNodesCreate")
	defer span.End()

	if request.Body == nil {
		return api.V1PluginGraphNodesCreate400JSONResponse{N400JSONResponse: formatBadRequest(errors.New("missing body"))}, nil
	}
	parent, err := resourceIDFromAPI(request.Body.ParentType, request.Body.ParentId)
	if err != nil {
		return api.V1PluginGraphNodesCreate400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}
	opts := service.CreateExtensionNodeOpts{
		Kind:       request.Body.Kind,
		Parent:     parent,
		Properties: mapOrEmpty(request.Body.Properties),
	}
	if request.Body.Relation != nil {
		to, err := resourceIDFromAPI(request.Body.Relation.ToType, request.Body.Relation.ToId)
		if err != nil {
			return api.V1PluginGraphNodesCreate400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
		}
		opts.Relation = &service.CreateExtensionRelationOpts{Kind: request.Body.Relation.Kind, To: to}
	}
	created, err := c.pluginService.CreateNode(ctx, request.PluginId, opts)
	if err != nil {
		switch classifyServiceError(err) {
		case http.StatusBadRequest:
			return api.V1PluginGraphNodesCreate400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
		case http.StatusForbidden:
			return api.V1PluginGraphNodesCreate403JSONResponse{N403JSONResponse: permissionDenied}, nil
		case http.StatusNotFound:
			return api.V1PluginGraphNodesCreate404JSONResponse{N404JSONResponse: notFound}, nil
		case http.StatusConflict:
			return api.V1PluginGraphNodesCreate409JSONResponse{N409JSONResponse: api.N409JSONResponse{Message: err.Error()}}, nil
		default:
			return api.V1PluginGraphNodesCreate500JSONResponse{N500JSONResponse: api.N500JSONResponse{Message: err.Error()}}, nil
		}
	}
	return api.V1PluginGraphNodesCreate201JSONResponse(extensionToDTO(created)), nil
}

func (c *pluginController) V1PluginGraphNodeGet(
	ctx context.Context,
	request api.V1PluginGraphNodeGetRequestObject,
) (api.V1PluginGraphNodeGetResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1PluginGraphNodeGet")
	defer span.End()

	id, err := model.NewIDFromString(request.Id, model.ResourceTypeExtension.String())
	if err != nil {
		return api.V1PluginGraphNodeGet404JSONResponse{N404JSONResponse: notFound}, nil
	}
	ownerPluginID := ""
	if request.Params.OwnerPluginId != nil {
		ownerPluginID = *request.Params.OwnerPluginId
	}
	node, err := c.pluginService.GetNode(ctx, request.PluginId, id, ownerPluginID)
	if err != nil {
		switch classifyServiceError(err) {
		case http.StatusForbidden:
			return api.V1PluginGraphNodeGet403JSONResponse{N403JSONResponse: permissionDenied}, nil
		case http.StatusNotFound:
			return api.V1PluginGraphNodeGet404JSONResponse{N404JSONResponse: notFound}, nil
		default:
			return api.V1PluginGraphNodeGet500JSONResponse{N500JSONResponse: api.N500JSONResponse{Message: err.Error()}}, nil
		}
	}
	return api.V1PluginGraphNodeGet200JSONResponse(extensionToDTO(node)), nil
}

func (c *pluginController) V1PluginGraphNodeUpdate(
	ctx context.Context,
	request api.V1PluginGraphNodeUpdateRequestObject,
) (api.V1PluginGraphNodeUpdateResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1PluginGraphNodeUpdate")
	defer span.End()

	if request.Body == nil {
		return api.V1PluginGraphNodeUpdate400JSONResponse{N400JSONResponse: formatBadRequest(errors.New("missing body"))}, nil
	}
	id, err := model.NewIDFromString(request.Id, model.ResourceTypeExtension.String())
	if err != nil {
		return api.V1PluginGraphNodeUpdate400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}
	updated, err := c.pluginService.UpdateNode(ctx, request.PluginId, id, request.Body.Properties)
	if err != nil {
		switch classifyServiceError(err) {
		case http.StatusBadRequest:
			return api.V1PluginGraphNodeUpdate400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
		case http.StatusForbidden:
			return api.V1PluginGraphNodeUpdate403JSONResponse{N403JSONResponse: permissionDenied}, nil
		case http.StatusNotFound:
			return api.V1PluginGraphNodeUpdate404JSONResponse{N404JSONResponse: notFound}, nil
		default:
			return api.V1PluginGraphNodeUpdate500JSONResponse{N500JSONResponse: api.N500JSONResponse{Message: err.Error()}}, nil
		}
	}
	return api.V1PluginGraphNodeUpdate200JSONResponse(extensionToDTO(updated)), nil
}

func (c *pluginController) V1PluginGraphNodeDelete(
	ctx context.Context,
	request api.V1PluginGraphNodeDeleteRequestObject,
) (api.V1PluginGraphNodeDeleteResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1PluginGraphNodeDelete")
	defer span.End()

	id, err := model.NewIDFromString(request.Id, model.ResourceTypeExtension.String())
	if err != nil {
		return api.V1PluginGraphNodeDelete404JSONResponse{N404JSONResponse: notFound}, nil
	}
	if err := c.pluginService.DeleteNode(ctx, request.PluginId, id); err != nil {
		switch classifyServiceError(err) {
		case http.StatusForbidden:
			return api.V1PluginGraphNodeDelete403JSONResponse{N403JSONResponse: permissionDenied}, nil
		case http.StatusNotFound:
			return api.V1PluginGraphNodeDelete404JSONResponse{N404JSONResponse: notFound}, nil
		default:
			return api.V1PluginGraphNodeDelete500JSONResponse{N500JSONResponse: api.N500JSONResponse{Message: err.Error()}}, nil
		}
	}
	return api.V1PluginGraphNodeDelete204Response{}, nil
}

func (c *pluginController) V1PluginGraphNodeMove(
	ctx context.Context,
	request api.V1PluginGraphNodeMoveRequestObject,
) (api.V1PluginGraphNodeMoveResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1PluginGraphNodeMove")
	defer span.End()

	if request.Body == nil {
		return api.V1PluginGraphNodeMove400JSONResponse{N400JSONResponse: formatBadRequest(errors.New("missing body"))}, nil
	}
	id, err := model.NewIDFromString(request.Id, model.ResourceTypeExtension.String())
	if err != nil {
		return api.V1PluginGraphNodeMove404JSONResponse{N404JSONResponse: notFound}, nil
	}
	parent, err := resourceIDFromAPI(request.Body.ParentType, request.Body.ParentId)
	if err != nil {
		return api.V1PluginGraphNodeMove400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}
	moved, err := c.pluginService.MoveNode(ctx, request.PluginId, id, parent)
	if err != nil {
		switch classifyServiceError(err) {
		case http.StatusBadRequest:
			return api.V1PluginGraphNodeMove400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
		case http.StatusForbidden:
			return api.V1PluginGraphNodeMove403JSONResponse{N403JSONResponse: permissionDenied}, nil
		case http.StatusNotFound:
			return api.V1PluginGraphNodeMove404JSONResponse{N404JSONResponse: notFound}, nil
		default:
			return api.V1PluginGraphNodeMove500JSONResponse{N500JSONResponse: api.N500JSONResponse{Message: err.Error()}}, nil
		}
	}
	return api.V1PluginGraphNodeMove200JSONResponse(extensionToDTO(moved)), nil
}

func (c *pluginController) V1PluginGraphRelationsGet(
	ctx context.Context,
	request api.V1PluginGraphRelationsGetRequestObject,
) (api.V1PluginGraphRelationsGetResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1PluginGraphRelationsGet")
	defer span.End()

	node, err := resourceIDFromAPI(request.Params.NodeType, request.Params.NodeId)
	if err != nil {
		return api.V1PluginGraphRelationsGet400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}
	page, err := cursorPageFromParams(request.Params.PageSize, request.Params.PageToken)
	if err != nil {
		return api.V1PluginGraphRelationsGet400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}
	direction := model.PluginGraphRelationDirectionOutgoing
	if request.Params.Direction != nil {
		if err := direction.UnmarshalText([]byte(*request.Params.Direction)); err != nil {
			return api.V1PluginGraphRelationsGet400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
		}
	}
	result, err := c.pluginService.ListRelations(ctx, request.PluginId, service.ListExtensionRelationOpts{
		Kind:      request.Params.Kind,
		Node:      node,
		Direction: direction,
		Page:      page,
	})
	if err != nil {
		switch classifyServiceError(err) {
		case http.StatusBadRequest:
			return api.V1PluginGraphRelationsGet400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
		case http.StatusForbidden:
			return api.V1PluginGraphRelationsGet403JSONResponse{N403JSONResponse: permissionDenied}, nil
		case http.StatusNotFound:
			return api.V1PluginGraphRelationsGet404JSONResponse{N404JSONResponse: notFound}, nil
		default:
			return api.V1PluginGraphRelationsGet500JSONResponse{N500JSONResponse: api.N500JSONResponse{Message: err.Error()}}, nil
		}
	}
	items := make([]api.ExtensionRelation, len(result.Items))
	for i, rel := range result.Items {
		items[i] = relationToDTO(rel)
	}
	return api.V1PluginGraphRelationsGet200JSONResponse{
		Items:    items,
		PageInfo: pageInfoToDTO(result.PageInfo),
	}, nil
}

func (c *pluginController) V1PluginGraphRelationsCreate(
	ctx context.Context,
	request api.V1PluginGraphRelationsCreateRequestObject,
) (api.V1PluginGraphRelationsCreateResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1PluginGraphRelationsCreate")
	defer span.End()

	if request.Body == nil {
		return api.V1PluginGraphRelationsCreate400JSONResponse{N400JSONResponse: formatBadRequest(errors.New("missing body"))}, nil
	}
	from, err := resourceIDFromAPI(request.Body.FromType, request.Body.FromId)
	if err != nil {
		return api.V1PluginGraphRelationsCreate400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}
	to, err := resourceIDFromAPI(request.Body.ToType, request.Body.ToId)
	if err != nil {
		return api.V1PluginGraphRelationsCreate400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
	}
	created, err := c.pluginService.CreateRelation(ctx, request.PluginId, service.CreateExtensionRelationOpts{
		Kind: request.Body.Kind, From: from, To: to,
	})
	if err != nil {
		switch classifyServiceError(err) {
		case http.StatusBadRequest:
			return api.V1PluginGraphRelationsCreate400JSONResponse{N400JSONResponse: formatBadRequest(err)}, nil
		case http.StatusForbidden:
			return api.V1PluginGraphRelationsCreate403JSONResponse{N403JSONResponse: permissionDenied}, nil
		case http.StatusNotFound:
			return api.V1PluginGraphRelationsCreate404JSONResponse{N404JSONResponse: notFound}, nil
		case http.StatusConflict:
			return api.V1PluginGraphRelationsCreate409JSONResponse{N409JSONResponse: api.N409JSONResponse{Message: err.Error()}}, nil
		default:
			return api.V1PluginGraphRelationsCreate500JSONResponse{N500JSONResponse: api.N500JSONResponse{Message: err.Error()}}, nil
		}
	}
	return api.V1PluginGraphRelationsCreate201JSONResponse(relationToDTO(created)), nil
}

func (c *pluginController) V1PluginGraphRelationDelete(
	ctx context.Context,
	request api.V1PluginGraphRelationDeleteRequestObject,
) (api.V1PluginGraphRelationDeleteResponseObject, error) {
	ctx, span := c.tracer.Start(ctx, "transport.http.handler/V1PluginGraphRelationDelete")
	defer span.End()

	if err := c.pluginService.DeleteRelation(ctx, request.PluginId, request.Id); err != nil {
		switch classifyServiceError(err) {
		case http.StatusForbidden:
			return api.V1PluginGraphRelationDelete403JSONResponse{N403JSONResponse: permissionDenied}, nil
		case http.StatusNotFound:
			return api.V1PluginGraphRelationDelete404JSONResponse{N404JSONResponse: notFound}, nil
		default:
			return api.V1PluginGraphRelationDelete500JSONResponse{N500JSONResponse: api.N500JSONResponse{Message: err.Error()}}, nil
		}
	}
	return api.V1PluginGraphRelationDelete204Response{}, nil
}

func pluginToDTO(inst *model.PluginInstallation, enabled *bool, config json.RawMessage) api.Plugin {
	if inst == nil {
		return api.Plugin{}
	}
	caps := make([]api.PluginCapability, len(inst.Manifest.Capabilities))
	for i, capability := range inst.Manifest.Capabilities {
		caps[i] = api.PluginCapability(capability)
	}
	slots := make([]api.PluginUISlot, len(inst.Manifest.Slots))
	for i, slot := range inst.Manifest.Slots {
		slots[i] = api.PluginUISlot(slot)
	}
	created := time.Time{}
	if inst.CreatedAt != nil {
		created = *inst.CreatedAt
	}
	schema := configSchemaToDTO(inst.Manifest.Config)
	graph := graphSummaryToDTO(inst.Manifest.Graph)
	out := api.Plugin{
		Id:           inst.ID,
		PluginId:     inst.PluginID,
		Name:         inst.Manifest.Name,
		Version:      inst.Version,
		Status:       api.PluginStatus(inst.Status.String()),
		Capabilities: caps,
		Slots:        slots,
		Enabled:      enabled,
		ConfigSchema: schema,
		Graph:        graph,
		CreatedAt:    created,
		UpdatedAt:    inst.UpdatedAt,
	}
	if inst.Error != "" {
		out.Error = &inst.Error
	}
	if cfg := rawConfigMap(config); cfg != nil {
		out.Config = &cfg
	}
	return out
}

func rawConfigMap(raw json.RawMessage) map[string]any {
	if len(raw) == 0 || string(raw) == "null" {
		return map[string]any{}
	}
	var values map[string]any
	if err := json.Unmarshal(raw, &values); err != nil || values == nil {
		return map[string]any{}
	}
	return values
}

func parseEqualsFilter(raw *string) (map[string]any, error) {
	if raw == nil || strings.TrimSpace(*raw) == "" {
		return nil, nil
	}
	var equals map[string]any
	if err := json.Unmarshal([]byte(*raw), &equals); err != nil {
		return nil, fmt.Errorf("equals must be a JSON object")
	}
	return equals, nil
}

func configSchemaToDTO(fields []model.PluginConfigFieldDecl) *[]api.PluginConfigField {
	if len(fields) == 0 {
		return nil
	}
	out := make([]api.PluginConfigField, len(fields))
	for i, field := range fields {
		item := api.PluginConfigField{
			Name: field.Name,
			Type: api.PluginConfigFieldType(field.Type.String()),
		}
		if field.Required {
			required := true
			item.Required = &required
		}
		if field.Foreign != "" {
			foreign := field.Foreign
			item.Foreign = &foreign
		}
		out[i] = item
	}
	return &out
}

func graphSummaryToDTO(schema *model.PluginGraphSchema) *api.PluginGraphSummary {
	if schema == nil {
		return nil
	}
	out := api.PluginGraphSummary{}
	if len(schema.Nodes) > 0 {
		nodes := make([]api.PluginGraphKindSummary, len(schema.Nodes))
		for i, node := range schema.Nodes {
			nodes[i] = api.PluginGraphKindSummary{
				Kind:       node.Kind,
				Parent:     node.Scope.Parent,
				Properties: propertySummaryToDTO(node.Properties),
			}
		}
		out.Nodes = &nodes
	}
	if len(schema.Foreign) > 0 {
		foreign := make([]api.PluginGraphForeignSummary, len(schema.Foreign))
		for i, decl := range schema.Foreign {
			foreign[i] = api.PluginGraphForeignSummary{
				Name:       decl.Name,
				Parent:     decl.Parent,
				Properties: propertySummaryToDTO(decl.Properties),
			}
		}
		out.Foreign = &foreign
	}
	return &out
}

func propertySummaryToDTO(props []model.PluginGraphPropertyDecl) *[]api.PluginGraphPropertySummary {
	if len(props) == 0 {
		return nil
	}
	out := make([]api.PluginGraphPropertySummary, len(props))
	for i, p := range props {
		item := api.PluginGraphPropertySummary{Name: p.Name, Type: p.Type.String()}
		if p.Required {
			required := true
			item.Required = &required
		}
		out[i] = item
	}
	return &out
}

func extensionToDTO(ext *model.Extension) api.ExtensionNode {
	if ext == nil {
		return api.ExtensionNode{}
	}
	props := ext.Properties
	if props == nil {
		props = map[string]any{}
	}
	out := api.ExtensionNode{
		Id:         ext.ID.String(),
		PluginId:   ext.PluginID,
		Kind:       ext.Kind,
		Properties: props,
		CreatedAt:  ext.CreatedAt,
		UpdatedAt:  ext.UpdatedAt,
	}
	if ext.Parent != nil {
		parentID := ext.Parent.String()
		parentType := api.ResourceType(ext.Parent.Label())
		out.ParentId = &parentID
		out.ParentType = &parentType
	}
	return out
}

func relationToDTO(rel *model.ExtensionRelation) api.ExtensionRelation {
	if rel == nil {
		return api.ExtensionRelation{}
	}
	return api.ExtensionRelation{
		Id:        rel.ID,
		Kind:      rel.Kind,
		From:      rel.From.String(),
		FromType:  api.ResourceType(rel.From.Label()),
		To:        rel.To.String(),
		ToType:    api.ResourceType(rel.To.Label()),
		CreatedAt: rel.CreatedAt,
	}
}

func mapOrEmpty(in *map[string]any) map[string]any {
	if in == nil {
		return map[string]any{}
	}
	return *in
}

func readPluginPackage(reader *multipart.Reader) ([]byte, error) {
	if reader == nil {
		return nil, errors.New("missing plugin package")
	}
	for {
		part, err := reader.NextPart()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, err
		}
		name := part.FormName()
		if name != "package" {
			_ = part.Close()
			continue
		}
		data, err := io.ReadAll(io.LimitReader(part, maxPluginPackageBytes+1))
		_ = part.Close()
		if err != nil {
			return nil, err
		}
		if len(data) > maxPluginPackageBytes {
			return nil, elemoplugin.ErrPackageTooLarge
		}
		return data, nil
	}
	return nil, errors.New("missing plugin package")
}

// NewPluginController creates a PluginController.
func NewPluginController(pluginService service.PluginService, opts ...ControllerOption) (PluginController, error) {
	c, err := newController(opts...)
	if err != nil {
		return nil, err
	}
	if pluginService == nil {
		return nil, ErrNoPluginService
	}
	return &pluginController{
		baseController: c,
		pluginService:  pluginService,
	}, nil
}

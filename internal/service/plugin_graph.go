package service

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg"
	"github.com/opcotech/elemo/internal/pkg/event"
	elemoplugin "github.com/opcotech/elemo/internal/plugin"
	"github.com/opcotech/elemo/internal/repository"
)

func (s *pluginService) CreateNode(
	ctx context.Context,
	pluginID string,
	opts CreateExtensionNodeOpts,
) (*model.Extension, error) {
	ctx, span := s.tracer.Start(ctx, "service.pluginService/CreateNode")
	defer span.End()

	if err := s.requireFeature(ctx, ErrPluginGraph); err != nil {
		return nil, err
	}
	manifest, err := s.requireGraphWrite(ctx, pluginID, opts.Parent)
	if err != nil {
		return nil, err
	}
	node, ok := manifest.Graph.NodeKind(opts.Kind)
	if !ok {
		return nil, errors.Join(ErrPluginGraph, model.ErrPluginGraphSchema)
	}
	if err := validateParentKind(node, opts.Parent); err != nil {
		return nil, errors.Join(ErrPluginGraph, err)
	}
	props, err := coerceProperties(node, opts.Properties, true)
	if err != nil {
		return nil, errors.Join(ErrPluginGraph, err)
	}
	stampUserID(ctx, node, props)
	if err := requireAction(ctx, s.permissionService, opts.Parent, model.ActionExtensionCreate); err != nil {
		return nil, errors.Join(ErrPluginGraph, err)
	}

	created, err := s.extensionRepo.Create(ctx, repository.CreateExtensionOpts{
		PluginID:   pluginID,
		Kind:       opts.Kind,
		Parent:     opts.Parent,
		Properties: props,
	})
	if err != nil {
		return nil, errors.Join(ErrPluginGraph, err)
	}
	if opts.Relation != nil {
		opts.Relation.From = created.ID
		if _, err := s.CreateRelation(ctx, pluginID, *opts.Relation); err != nil {
			_ = s.extensionRepo.Delete(ctx, pluginID, created.ID)
			return nil, err
		}
	}
	s.publishExtensionEvent(ctx, model.PluginEventExtensionCreated, pluginID, created)
	return created, nil
}

func (s *pluginService) GetNode(ctx context.Context, pluginID string, id model.ID, ownerPluginID string) (*model.Extension, error) {
	ctx, span := s.tracer.Start(ctx, "service.pluginService/GetNode")
	defer span.End()

	if err := s.requireFeature(ctx, ErrPluginGraph); err != nil {
		return nil, err
	}
	owner := pluginID
	if ownerPluginID != "" && ownerPluginID != pluginID {
		if err := s.requireForeignAccess(ctx, pluginID, ownerPluginID, id); err != nil {
			return nil, err
		}
		owner = ownerPluginID
	} else if _, err := s.requireGraphRead(ctx, pluginID, id); err != nil {
		return nil, err
	}
	if err := requireAction(ctx, s.permissionService, id, model.ActionExtensionRead); err != nil {
		return nil, errors.Join(ErrPluginGraph, err)
	}
	ext, err := s.extensionRepo.Get(ctx, owner, id)
	if err != nil {
		return nil, errors.Join(ErrPluginGraph, err)
	}
	return ext, nil
}

func (s *pluginService) UpdateNode(
	ctx context.Context,
	pluginID string,
	id model.ID,
	properties map[string]any,
) (*model.Extension, error) {
	ctx, span := s.tracer.Start(ctx, "service.pluginService/UpdateNode")
	defer span.End()

	if err := s.requireFeature(ctx, ErrPluginGraph); err != nil {
		return nil, err
	}
	manifest, err := s.requireGraphWrite(ctx, pluginID, id)
	if err != nil {
		return nil, err
	}
	if err := requireAction(ctx, s.permissionService, id, model.ActionExtensionUpdate); err != nil {
		return nil, errors.Join(ErrPluginGraph, err)
	}
	existing, err := s.extensionRepo.Get(ctx, pluginID, id)
	if err != nil {
		return nil, errors.Join(ErrPluginGraph, err)
	}
	node, ok := manifest.Graph.NodeKind(existing.Kind)
	if !ok {
		return nil, errors.Join(ErrPluginGraph, model.ErrPluginGraphSchema)
	}
	props, err := coerceProperties(node, properties, false)
	if err != nil {
		return nil, errors.Join(ErrPluginGraph, err)
	}
	updated, err := s.extensionRepo.Update(ctx, pluginID, id, repository.UpdateExtensionOpts{Properties: props})
	if err != nil {
		return nil, errors.Join(ErrPluginGraph, err)
	}
	s.publishExtensionEvent(ctx, model.PluginEventExtensionUpdated, pluginID, updated)
	return updated, nil
}

func (s *pluginService) DeleteNode(ctx context.Context, pluginID string, id model.ID) error {
	ctx, span := s.tracer.Start(ctx, "service.pluginService/DeleteNode")
	defer span.End()

	if err := s.requireFeature(ctx, ErrPluginGraph); err != nil {
		return err
	}
	if _, err := s.requireGraphWrite(ctx, pluginID, id); err != nil {
		return err
	}
	if err := requireAction(ctx, s.permissionService, id, model.ActionExtensionDelete); err != nil {
		return errors.Join(ErrPluginGraph, err)
	}
	existing, getErr := s.extensionRepo.Get(ctx, pluginID, id)
	if err := s.extensionRepo.Delete(ctx, pluginID, id); err != nil {
		return errors.Join(ErrPluginGraph, err)
	}
	if getErr == nil {
		s.publishExtensionEvent(ctx, model.PluginEventExtensionDeleted, pluginID, existing)
	}
	return nil
}

func (s *pluginService) ListNodes(
	ctx context.Context,
	pluginID string,
	opts ListExtensionNodeOpts,
) (repository.Page[*model.Extension], error) {
	ctx, span := s.tracer.Start(ctx, "service.pluginService/ListNodes")
	defer span.End()

	if err := s.requireFeature(ctx, ErrPluginGraph); err != nil {
		return repository.Page[*model.Extension]{}, err
	}
	owner := pluginID
	if opts.OwnerPluginID != "" && opts.OwnerPluginID != pluginID {
		if err := s.requireForeignKindAccess(ctx, pluginID, opts.OwnerPluginID, opts.Kind, opts.Scope); err != nil {
			return repository.Page[*model.Extension]{}, err
		}
		owner = opts.OwnerPluginID
	} else if _, err := s.requireGraphRead(ctx, pluginID, opts.Scope); err != nil {
		return repository.Page[*model.Extension]{}, err
	}
	if err := requireAction(ctx, s.permissionService, opts.Scope, model.ActionExtensionRead); err != nil {
		return repository.Page[*model.Extension]{}, errors.Join(ErrPluginGraph, err)
	}
	page, err := s.extensionRepo.List(ctx, repository.ListExtensionFilter{
		PluginID: owner,
		Kind:     opts.Kind,
		Scope:    opts.Scope,
		Equals:   opts.Equals,
		Page:     opts.Page,
	})
	if err != nil {
		return repository.Page[*model.Extension]{}, errors.Join(ErrPluginGraph, err)
	}
	return page, nil
}

func (s *pluginService) MoveNode(ctx context.Context, pluginID string, id, parent model.ID) (*model.Extension, error) {
	ctx, span := s.tracer.Start(ctx, "service.pluginService/MoveNode")
	defer span.End()

	if err := s.requireFeature(ctx, ErrPluginGraph); err != nil {
		return nil, err
	}
	manifest, err := s.requireGraphWrite(ctx, pluginID, id)
	if err != nil {
		return nil, err
	}
	if _, err := s.requireGraphWrite(ctx, pluginID, parent); err != nil {
		return nil, err
	}
	if err := requireAction(ctx, s.permissionService, id, model.ActionExtensionUpdate); err != nil {
		return nil, errors.Join(ErrPluginGraph, err)
	}
	if err := requireAction(ctx, s.permissionService, parent, model.ActionExtensionCreate); err != nil {
		return nil, errors.Join(ErrPluginGraph, err)
	}
	if err := requireUpdateOrExtension(ctx, s.permissionService, parent); err != nil {
		return nil, errors.Join(ErrPluginGraph, err)
	}
	existing, err := s.extensionRepo.Get(ctx, pluginID, id)
	if err != nil {
		return nil, errors.Join(ErrPluginGraph, err)
	}
	node, ok := manifest.Graph.NodeKind(existing.Kind)
	if !ok {
		return nil, errors.Join(ErrPluginGraph, model.ErrPluginGraphSchema)
	}
	if err := validateParentKind(node, parent); err != nil {
		return nil, errors.Join(ErrPluginGraph, err)
	}
	relTypes := parentDomainRelationTypes(manifest, existing.Kind, parent.Label())
	moved, err := s.extensionRepo.Move(ctx, repository.MoveExtensionOpts{
		PluginID:      pluginID,
		ID:            id,
		Parent:        parent,
		RelationTypes: relTypes,
	})
	if err != nil {
		return nil, errors.Join(ErrPluginGraph, err)
	}
	return moved, nil
}

func parentDomainRelationTypes(manifest model.PluginManifest, fromKind, parentLabel string) []string {
	if manifest.Graph == nil {
		return nil
	}
	out := make([]string, 0)
	for _, rel := range manifest.Graph.Relations {
		if rel.From != fromKind || rel.To != parentLabel {
			continue
		}
		if rel.Cardinality != model.PluginGraphCardinalityManyToOne &&
			rel.Cardinality != model.PluginGraphCardinalityOneToOne {
			continue
		}
		relType, err := elemoplugin.RelationType(manifest.ID, rel.Kind)
		if err != nil {
			continue
		}
		out = append(out, relType)
	}
	return out
}

func stampUserID(ctx context.Context, node model.PluginGraphNodeDecl, props map[string]any) {
	declared := false
	for _, p := range node.Properties {
		if p.Name == "user_id" {
			declared = true
			break
		}
	}
	if !declared {
		return
	}
	if uid := pkg.CtxUserID(ctx); uid != "" {
		props["user_id"] = uid
	}
}

func (s *pluginService) CreateRelation(
	ctx context.Context,
	pluginID string,
	opts CreateExtensionRelationOpts,
) (*model.ExtensionRelation, error) {
	ctx, span := s.tracer.Start(ctx, "service.pluginService/CreateRelation")
	defer span.End()

	if err := s.requireFeature(ctx, ErrPluginGraph); err != nil {
		return nil, err
	}
	manifest, err := s.requireGraphWrite(ctx, pluginID, opts.From)
	if err != nil {
		return nil, err
	}
	rel, ok := manifest.Graph.RelationKind(opts.Kind)
	if !ok {
		return nil, errors.Join(ErrPluginGraph, model.ErrPluginGraphSchema)
	}
	if err := s.validateRelationEnds(ctx, pluginID, manifest, rel, opts.From, opts.To); err != nil {
		return nil, errors.Join(ErrPluginGraph, err)
	}
	if err := s.enforceCardinality(ctx, pluginID, rel, opts.From, opts.To); err != nil {
		return nil, errors.Join(ErrPluginGraph, err)
	}
	if err := authorizeRelationEnds(ctx, s.permissionService, manifest, rel, opts.From, opts.To); err != nil {
		return nil, errors.Join(ErrPluginGraph, err)
	}
	created, err := s.extensionRepo.CreateRelation(ctx, repository.CreateExtensionRelationOpts{
		PluginID: pluginID,
		Kind:     opts.Kind,
		From:     opts.From,
		To:       opts.To,
	})
	if err != nil {
		return nil, errors.Join(ErrPluginGraph, err)
	}
	return created, nil
}

func (s *pluginService) DeleteRelation(ctx context.Context, pluginID, relID string) error {
	ctx, span := s.tracer.Start(ctx, "service.pluginService/DeleteRelation")
	defer span.End()

	if err := s.requireFeature(ctx, ErrPluginGraph); err != nil {
		return err
	}
	inst, err := s.repo.GetInstallation(ctx, pluginID)
	if err != nil {
		return errors.Join(ErrPluginGraph, err)
	}
	if err := elemopluginRequireCap(inst.Manifest, model.CapabilityGraphWrite); err != nil {
		return errors.Join(ErrPluginGraph, err)
	}
	if err := s.extensionRepo.DeleteRelation(ctx, pluginID, relID); err != nil {
		return errors.Join(ErrPluginGraph, err)
	}
	return nil
}

func (s *pluginService) ListRelations(
	ctx context.Context,
	pluginID string,
	opts ListExtensionRelationOpts,
) (repository.Page[*model.ExtensionRelation], error) {
	ctx, span := s.tracer.Start(ctx, "service.pluginService/ListRelations")
	defer span.End()

	if err := s.requireFeature(ctx, ErrPluginGraph); err != nil {
		return repository.Page[*model.ExtensionRelation]{}, err
	}
	if _, err := s.requireGraphRead(ctx, pluginID, opts.Node); err != nil {
		return repository.Page[*model.ExtensionRelation]{}, err
	}
	action, ok := model.ReadActionFor(opts.Node.Type)
	if !ok {
		action = model.ActionExtensionRead
	}
	if err := requireAction(ctx, s.permissionService, opts.Node, action); err != nil {
		return repository.Page[*model.ExtensionRelation]{}, errors.Join(ErrPluginGraph, err)
	}
	out, err := s.extensionRepo.ListRelations(ctx, pluginID, opts.Kind, opts.Node, opts.Direction, opts.Page)
	if err != nil {
		return repository.Page[*model.ExtensionRelation]{}, errors.Join(ErrPluginGraph, err)
	}
	return out, nil
}

func (s *pluginService) requireGraphRead(ctx context.Context, pluginID string, scope model.ID) (model.PluginManifest, error) {
	inst, err := s.repo.GetInstallation(ctx, pluginID)
	if err != nil {
		return model.PluginManifest{}, errors.Join(ErrPluginGraph, err)
	}
	if err := elemopluginRequireCap(inst.Manifest, model.CapabilityGraphRead); err != nil {
		return model.PluginManifest{}, errors.Join(ErrPluginGraph, err)
	}
	if err := s.requireActive(ctx, pluginID, scope); err != nil {
		return model.PluginManifest{}, errors.Join(ErrPluginGraph, err)
	}
	return inst.Manifest, nil
}

func (s *pluginService) requireGraphWrite(ctx context.Context, pluginID string, scope model.ID) (model.PluginManifest, error) {
	inst, err := s.repo.GetInstallation(ctx, pluginID)
	if err != nil {
		return model.PluginManifest{}, errors.Join(ErrPluginGraph, err)
	}
	if err := elemopluginRequireCap(inst.Manifest, model.CapabilityGraphWrite); err != nil {
		return model.PluginManifest{}, errors.Join(ErrPluginGraph, err)
	}
	if err := s.requireActive(ctx, pluginID, scope); err != nil {
		return model.PluginManifest{}, errors.Join(ErrPluginGraph, err)
	}
	if inst.Manifest.Graph == nil {
		return model.PluginManifest{}, errors.Join(ErrPluginGraph, model.ErrPluginGraphSchema)
	}
	return inst.Manifest, nil
}

func elemopluginRequireCap(manifest model.PluginManifest, capability model.PluginCapability) error {
	if !manifest.HasCapability(capability) {
		return ErrNoPermission
	}
	return nil
}

func validateParentKind(node model.PluginGraphNodeDecl, parent model.ID) error {
	if rt, ok := model.CoreScopeType(node.Scope.Parent); ok {
		if parent.Type != rt {
			return fmt.Errorf("parent type %s does not match %s", parent.Label(), node.Scope.Parent)
		}
		return nil
	}
	if parent.Type != model.ResourceTypeExtension {
		return fmt.Errorf("parent must be an extension of kind %s", node.Scope.Parent)
	}
	return nil
}

func authorizeRelationEnds(
	ctx context.Context,
	perm PermissionService,
	manifest model.PluginManifest,
	rel model.PluginGraphRelationDecl,
	from, to model.ID,
) error {
	if err := requireUpdateOrExtension(ctx, perm, from); err != nil {
		return err
	}
	if to.Type == model.ResourceTypeUser {
		if pkg.CtxUserID(ctx) != to.String() {
			return ErrNoPermission
		}
		return nil
	}
	if to.Type == model.ResourceTypeExtension && isOwnPluginKind(manifest, rel.To) {
		return requireAction(ctx, perm, to, model.ActionExtensionRead)
	}
	return requireUpdateOrExtension(ctx, perm, to)
}

func isOwnPluginKind(manifest model.PluginManifest, name string) bool {
	if manifest.Graph == nil {
		return false
	}
	_, ok := manifest.Graph.NodeKind(name)
	return ok
}

func (s *pluginService) validateRelationEnds(
	ctx context.Context,
	pluginID string,
	manifest model.PluginManifest,
	rel model.PluginGraphRelationDecl,
	from, to model.ID,
) error {
	if err := s.matchRelationEndpoint(ctx, pluginID, manifest, rel.From, from); err != nil {
		return err
	}
	return s.matchRelationEndpoint(ctx, pluginID, manifest, rel.To, to)
}

func (s *pluginService) matchRelationEndpoint(
	ctx context.Context,
	pluginID string,
	manifest model.PluginManifest,
	declared string,
	id model.ID,
) error {
	if rt, ok := model.CoreScopeType(declared); ok {
		if id.Type != rt {
			return errors.Join(model.ErrPluginGraphSchema, fmt.Errorf("relation endpoint type %s does not match %s", id.Label(), declared))
		}
		return nil
	}
	if id.Type != model.ResourceTypeExtension {
		return errors.Join(model.ErrPluginGraphSchema, fmt.Errorf("relation endpoint %s must be an extension of kind %s", id.Label(), declared))
	}
	if _, ok := manifest.Graph.NodeKind(declared); ok {
		ext, err := s.extensionRepo.Get(ctx, pluginID, id)
		if err != nil {
			return err
		}
		if ext.Kind != declared {
			return errors.Join(model.ErrPluginGraphSchema, fmt.Errorf("extension kind %s does not match %s", ext.Kind, declared))
		}
		return nil
	}
	foreign, ok := manifest.Graph.ForeignKind(declared)
	if !ok {
		return errors.Join(model.ErrPluginGraphSchema, fmt.Errorf("undeclared relation endpoint %s", declared))
	}
	act, err := s.nearestActivation(ctx, pluginID, id)
	if err != nil {
		return err
	}
	binding, ok := model.GraphBinding(act.Config, manifest.Config, foreign.Name)
	if !ok {
		return errors.Join(model.ErrPluginGraphBinding, fmt.Errorf("foreign %s is not bound", foreign.Name))
	}
	ext, err := s.extensionRepo.Get(ctx, binding.PluginID, id)
	if err != nil {
		return err
	}
	if ext.Kind != binding.Kind {
		return errors.Join(model.ErrPluginGraphBinding, fmt.Errorf("extension kind %s does not match bound kind %s", ext.Kind, binding.Kind))
	}
	return nil
}

func (s *pluginService) enforceCardinality(
	ctx context.Context,
	pluginID string,
	rel model.PluginGraphRelationDecl,
	from, to model.ID,
) error {
	switch rel.Cardinality {
	case model.PluginGraphCardinalityManyToMany:
		return nil
	case model.PluginGraphCardinalityOneToMany, model.PluginGraphCardinalityManyToOne, model.PluginGraphCardinalityOneToOne:
	default:
		return model.ErrPluginRelationCardinality
	}
	outgoing, incoming, err := s.extensionRepo.CountRelations(ctx, pluginID, rel.Kind, from, to)
	if err != nil {
		return err
	}
	if rel.Cardinality == model.PluginGraphCardinalityManyToOne || rel.Cardinality == model.PluginGraphCardinalityOneToOne {
		if outgoing > 0 {
			return model.ErrPluginRelationCardinality
		}
	}
	if rel.Cardinality == model.PluginGraphCardinalityOneToMany || rel.Cardinality == model.PluginGraphCardinalityOneToOne {
		if incoming > 0 {
			return model.ErrPluginRelationCardinality
		}
	}
	return nil
}

func (s *pluginService) requireForeignAccess(ctx context.Context, pluginID, ownerPluginID string, id model.ID) error {
	if _, err := s.requireGraphRead(ctx, pluginID, id); err != nil {
		return err
	}
	ext, err := s.extensionRepo.Get(ctx, ownerPluginID, id)
	if err != nil {
		return errors.Join(ErrPluginGraph, err)
	}
	return s.requireForeignKindAccess(ctx, pluginID, ownerPluginID, ext.Kind, id)
}

func (s *pluginService) requireForeignKindAccess(ctx context.Context, pluginID, ownerPluginID, kind string, scope model.ID) error {
	if _, err := s.requireGraphRead(ctx, pluginID, scope); err != nil {
		return err
	}
	if err := s.requireActive(ctx, ownerPluginID, scope); err != nil {
		return errors.Join(ErrPluginGraph, model.ErrPluginNotActive)
	}
	act, err := s.nearestActivation(ctx, pluginID, scope)
	if err != nil {
		return errors.Join(ErrPluginGraph, err)
	}
	inst, err := s.repo.GetInstallation(ctx, pluginID)
	if err != nil {
		return errors.Join(ErrPluginGraph, err)
	}
	if !model.BindingMatches(act.Config, inst.Manifest.Config, ownerPluginID, kind) {
		return errors.Join(ErrPluginGraph, model.ErrPluginGraphBinding)
	}
	return nil
}

func (s *pluginService) publishExtensionEvent(ctx context.Context, typ model.PluginEventType, pluginID string, ext *model.Extension) {
	if ext == nil {
		return
	}
	payload := map[string]any{
		"plugin_id": pluginID,
		"kind":      ext.Kind,
		"id":        ext.ID.String(),
	}
	if ext.Parent != nil {
		payload["parent_id"] = ext.Parent.String()
		payload["parent_type"] = ext.Parent.Label()
		payload["scope_id"] = ext.Parent.Composite()
	}
	publishDomainEvent(ctx, s.eventBus, s.logger, event.Event{
		Type:     typ,
		Resource: ext.ID,
		Payload:  payload,
	})
}

func requireUpdateOrExtension(ctx context.Context, perm PermissionService, id model.ID) error {
	if id.Type == model.ResourceTypeExtension {
		return requireAction(ctx, perm, id, model.ActionExtensionUpdate)
	}
	action, ok := model.UpdateActionFor(id.Type)
	if !ok {
		return ErrNoPermission
	}
	return requireAction(ctx, perm, id, action)
}

func coerceProperties(node model.PluginGraphNodeDecl, in map[string]any, requireRequired bool) (map[string]any, error) {
	out := make(map[string]any, len(in))
	declared := make(map[string]model.PluginGraphPropertyDecl, len(node.Properties))
	for _, p := range node.Properties {
		declared[p.Name] = p
	}
	for k := range in {
		if _, ok := declared[k]; !ok {
			return nil, fmt.Errorf("undeclared property %s", k)
		}
	}
	for _, p := range node.Properties {
		v, ok := in[p.Name]
		if !ok {
			if requireRequired && p.Required {
				return nil, fmt.Errorf("missing required property %s", p.Name)
			}
			continue
		}
		coerced, err := coerceProperty(p.Type, v)
		if err != nil {
			return nil, fmt.Errorf("property %s: %w", p.Name, err)
		}
		out[p.Name] = coerced
	}
	return out, nil
}

func coerceProperty(t model.PluginGraphPropertyType, v any) (any, error) {
	switch t {
	case model.PluginGraphPropertyTypeStr:
		s, ok := v.(string)
		if !ok {
			return nil, errors.New("expected string")
		}
		return s, nil
	case model.PluginGraphPropertyTypeInteger:
		switch n := v.(type) {
		case int:
			return int64(n), nil
		case int64:
			return n, nil
		case float64:
			return int64(n), nil
		case string:
			i, err := strconv.ParseInt(n, 10, 64)
			return i, err
		default:
			return nil, errors.New("expected integer")
		}
	case model.PluginGraphPropertyTypeDecimal:
		switch n := v.(type) {
		case string:
			return n, nil
		case float64:
			return strconv.FormatFloat(n, 'f', -1, 64), nil
		default:
			return nil, errors.New("expected decimal")
		}
	case model.PluginGraphPropertyTypeBoolean:
		b, ok := v.(bool)
		if !ok {
			return nil, errors.New("expected boolean")
		}
		return b, nil
	case model.PluginGraphPropertyTypeDate:
		s, ok := v.(string)
		if !ok {
			return nil, errors.New("expected date")
		}
		if _, err := time.Parse("2006-01-02", s); err != nil {
			return nil, err
		}
		return s, nil
	case model.PluginGraphPropertyTypeDateTime:
		s, ok := v.(string)
		if !ok {
			return nil, errors.New("expected datetime")
		}
		if _, err := time.Parse(time.RFC3339, strings.TrimSpace(s)); err != nil {
			return nil, err
		}
		return s, nil
	default:
		return nil, errors.New("unsupported property type")
	}
}

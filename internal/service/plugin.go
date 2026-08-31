package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/opcotech/elemo/internal/config"
	"github.com/opcotech/elemo/internal/license"
	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg"
	"github.com/opcotech/elemo/internal/pkg/event"
	"github.com/opcotech/elemo/internal/pkg/log"
	"github.com/opcotech/elemo/internal/pkg/safepath"
	elemoplugin "github.com/opcotech/elemo/internal/plugin"
	"github.com/opcotech/elemo/internal/repository"
)

// FrontendPlugin is the discovery payload for enabled frontend plugins.
type FrontendPlugin struct {
	ID         string
	Version    string
	Entrypoint string
	Module     string
	Slots      []model.PluginUISlot
}

// PluginListItem is an installation plus optional scoped activation.
type PluginListItem struct {
	Installation *model.PluginInstallation
	Enabled      *bool
	Config       json.RawMessage
}

// CreateExtensionNodeOpts is the public graph create payload.
type CreateExtensionNodeOpts struct {
	Kind       string
	Parent     model.ID
	Properties map[string]any
	Relation   *CreateExtensionRelationOpts
}

type CreateExtensionRelationOpts struct {
	Kind string
	From model.ID
	To   model.ID
}

type ListExtensionNodeOpts struct {
	Kind          string
	Scope         model.ID
	Equals        map[string]any
	OwnerPluginID string
	Page          repository.CursorPage
}

type ListExtensionRelationOpts struct {
	Kind      string
	Node      model.ID
	Direction model.PluginGraphRelationDirection
	Page      repository.CursorPage
}

//go:generate go tool mockgen -destination=mock/mock_plugin_gen.go -package=mocksvc . PluginService
type PluginService interface {
	Install(ctx context.Context, zip []byte) (*model.PluginInstallation, error)
	Upgrade(ctx context.Context, pluginID string, zip []byte) (*model.PluginInstallation, error)
	Uninstall(ctx context.Context, pluginID string) error
	Enable(ctx context.Context, pluginID string, scope model.ID, config json.RawMessage) error
	Disable(ctx context.Context, pluginID string, scope model.ID) error
	Get(ctx context.Context, pluginID string) (*model.PluginInstallation, error)
	List(ctx context.Context) ([]*model.PluginInstallation, error)
	ListManaged(ctx context.Context, scope model.ID) ([]PluginListItem, error)
	ListFrontend(ctx context.Context, scope model.ID) ([]FrontendPlugin, error)
	GetConfig(ctx context.Context, pluginID string, scope model.ID) (json.RawMessage, error)
	GetManagedConfig(ctx context.Context, pluginID string, scope model.ID) (json.RawMessage, error)
	SetConfig(ctx context.Context, pluginID string, scope model.ID, config json.RawMessage) error
	Invoke(ctx context.Context, pluginID string, req elemoplugin.InvokeRequest) (elemoplugin.InvokeResponse, error)
	AssetPath(ctx context.Context, pluginID, version, rel string) (string, error)
	Restore(ctx context.Context) error

	CreateNode(ctx context.Context, pluginID string, opts CreateExtensionNodeOpts) (*model.Extension, error)
	GetNode(ctx context.Context, pluginID string, id model.ID, ownerPluginID string) (*model.Extension, error)
	UpdateNode(ctx context.Context, pluginID string, id model.ID, properties map[string]any) (*model.Extension, error)
	DeleteNode(ctx context.Context, pluginID string, id model.ID) error
	ListNodes(ctx context.Context, pluginID string, opts ListExtensionNodeOpts) (repository.Page[*model.Extension], error)
	MoveNode(ctx context.Context, pluginID string, id, parent model.ID) (*model.Extension, error)
	CreateRelation(ctx context.Context, pluginID string, opts CreateExtensionRelationOpts) (*model.ExtensionRelation, error)
	DeleteRelation(ctx context.Context, pluginID, relID string) error
	ListRelations(ctx context.Context, pluginID string, opts ListExtensionRelationOpts) (repository.Page[*model.ExtensionRelation], error)
}

type pluginService struct {
	runtime
	conf              config.PluginConfig
	repo              repository.PluginRepository
	extensionRepo     repository.ExtensionRepository
	permissionService PermissionService
	licenseService    LicenseService
	registry          *elemoplugin.Registry
	host              elemoplugin.Host
	bus               *event.Bus
	subscriptions     []*event.Subscription
	eventWG           sync.WaitGroup
}

func (s *pluginService) requireFeature(ctx context.Context, wrap error) error {
	ok, err := s.licenseService.HasFeature(ctx, license.FeaturePlugins)
	if err != nil {
		return errors.Join(wrap, err)
	}
	if !ok {
		return errors.Join(wrap, ErrFeatureDisabled)
	}
	return nil
}

func (s *pluginService) Install(ctx context.Context, zip []byte) (*model.PluginInstallation, error) {
	ctx, span := s.tracer.Start(ctx, "service.pluginService/Install")
	defer span.End()

	if err := s.requireFeature(ctx, ErrPluginInstall); err != nil {
		return nil, err
	}
	if err := requireAction(ctx, s.permissionService, model.InstallationID(), model.ActionPluginInstall); err != nil {
		return nil, errors.Join(ErrPluginInstall, err)
	}

	tmp, err := os.MkdirTemp("", "elemo-plugin-*")
	if err != nil {
		return nil, errors.Join(ErrPluginInstall, err)
	}
	defer os.RemoveAll(tmp)

	pkg, err := elemoplugin.ExtractZip(zip, tmp)
	if err != nil {
		return nil, errors.Join(ErrPluginInstall, err)
	}
	if _, err := s.repo.GetInstallation(ctx, pkg.Manifest.ID); err == nil {
		return nil, errors.Join(ErrPluginInstall, repository.ErrPluginConflict)
	} else if !errors.Is(err, repository.ErrNotFound) {
		return nil, errors.Join(ErrPluginInstall, err)
	}

	dest := elemoplugin.InstallDirectory(s.conf.Directory, pkg.Manifest.ID, pkg.Manifest.Version)
	if err := os.MkdirAll(filepath.Dir(dest), 0o750); err != nil {
		return nil, errors.Join(ErrPluginInstall, err)
	}
	if err := os.RemoveAll(dest); err != nil {
		return nil, errors.Join(ErrPluginInstall, err)
	}
	if err := os.Rename(pkg.Root, dest); err != nil {
		if copyErr := copyDir(pkg.Root, dest); copyErr != nil {
			return nil, errors.Join(ErrPluginInstall, err, copyErr)
		}
	}
	pkg.Root = dest

	wasm, err := elemoplugin.ReadWASM(pkg)
	if err != nil {
		return nil, errors.Join(ErrPluginInstall, err)
	}

	inst := &model.PluginInstallation{
		PluginID: pkg.Manifest.ID,
		Version:  pkg.Manifest.Version,
		Status:   model.PluginStatusInstalled,
		Manifest: pkg.Manifest,
	}
	saved, err := s.repo.UpsertInstallation(ctx, inst)
	if err != nil {
		return nil, errors.Join(ErrPluginInstall, err)
	}
	if err := s.registry.Put(ctx, elemoplugin.LoadedPlugin{
		ID:       saved.PluginID,
		Version:  saved.Version,
		Manifest: saved.Manifest,
		Root:     dest,
		Status:   model.PluginStatusInstalled,
	}, wasm); err != nil {
		saved.Status = model.PluginStatusFailed
		saved.Error = err.Error()
		_, _ = s.repo.UpsertInstallation(ctx, saved)
		return saved, errors.Join(ErrPluginInstall, err)
	}
	return saved, nil
}

func (s *pluginService) Upgrade(ctx context.Context, pluginID string, zip []byte) (*model.PluginInstallation, error) {
	ctx, span := s.tracer.Start(ctx, "service.pluginService/Upgrade")
	defer span.End()

	if err := s.requireFeature(ctx, ErrPluginUpgrade); err != nil {
		return nil, err
	}
	if err := requireAction(ctx, s.permissionService, model.InstallationID(), model.ActionPluginInstall); err != nil {
		return nil, errors.Join(ErrPluginUpgrade, err)
	}
	current, err := s.repo.GetInstallation(ctx, pluginID)
	if err != nil {
		return nil, errors.Join(ErrPluginUpgrade, err)
	}

	tmp, err := os.MkdirTemp("", "elemo-plugin-*")
	if err != nil {
		return nil, errors.Join(ErrPluginUpgrade, err)
	}
	defer os.RemoveAll(tmp)
	pkg, err := elemoplugin.ExtractZip(zip, tmp)
	if err != nil {
		return nil, errors.Join(ErrPluginUpgrade, err)
	}
	if pkg.Manifest.ID != pluginID {
		return nil, errors.Join(ErrPluginUpgrade, model.ErrInvalidPluginID)
	}
	if err := assertAdditiveGraph(current.Manifest.Graph, pkg.Manifest.Graph); err != nil {
		return nil, errors.Join(ErrPluginUpgrade, err)
	}

	_ = s.registry.Stop(ctx, pluginID, s.conf.ExecutionTimeout)

	dest := elemoplugin.InstallDirectory(s.conf.Directory, pkg.Manifest.ID, pkg.Manifest.Version)
	_ = os.RemoveAll(dest)
	if err := os.MkdirAll(filepath.Dir(dest), 0o750); err != nil {
		return nil, errors.Join(ErrPluginUpgrade, err)
	}
	if err := os.Rename(pkg.Root, dest); err != nil {
		if copyErr := copyDir(pkg.Root, dest); copyErr != nil {
			return nil, errors.Join(ErrPluginUpgrade, err, copyErr)
		}
	}
	wasm, err := elemoplugin.ReadWASM(elemoplugin.ExtractedPackage{Root: dest, Manifest: pkg.Manifest})
	if err != nil {
		return nil, errors.Join(ErrPluginUpgrade, err)
	}

	current.Version = pkg.Manifest.Version
	current.Manifest = pkg.Manifest
	current.Status = model.PluginStatusInstalled
	current.Error = ""
	saved, err := s.repo.UpsertInstallation(ctx, current)
	if err != nil {
		return nil, errors.Join(ErrPluginUpgrade, err)
	}
	if err := s.registry.Put(ctx, elemoplugin.LoadedPlugin{
		ID: saved.PluginID, Version: saved.Version, Manifest: saved.Manifest, Root: dest, Status: model.PluginStatusInstalled,
	}, wasm); err != nil {
		return nil, errors.Join(ErrPluginUpgrade, err)
	}
	acts, err := s.repo.ListActivations(ctx, pluginID)
	if err != nil {
		return saved, nil
	}
	for _, act := range acts {
		if act.Enabled {
			_ = s.registry.Start(ctx, pluginID)
			break
		}
	}
	return saved, nil
}

func (s *pluginService) Uninstall(ctx context.Context, pluginID string) error {
	ctx, span := s.tracer.Start(ctx, "service.pluginService/Uninstall")
	defer span.End()

	if err := s.requireFeature(ctx, ErrPluginUninstall); err != nil {
		return err
	}
	if err := requireAction(ctx, s.permissionService, model.InstallationID(), model.ActionPluginInstall); err != nil {
		return errors.Join(ErrPluginUninstall, err)
	}
	inst, err := s.repo.GetInstallation(ctx, pluginID)
	if err != nil {
		return errors.Join(ErrPluginUninstall, err)
	}
	_ = s.registry.Remove(ctx, pluginID)
	if err := s.extensionRepo.DeleteByPlugin(ctx, pluginID); err != nil {
		s.logger.Warn(ctx, "failed to delete plugin graph", log.WithError(err))
	}
	if err := s.repo.DeleteActivations(ctx, pluginID); err != nil {
		return errors.Join(ErrPluginUninstall, err)
	}
	if err := s.repo.DeleteStorageForPlugin(ctx, pluginID); err != nil {
		return errors.Join(ErrPluginUninstall, err)
	}
	if err := s.repo.DeleteInstallation(ctx, pluginID); err != nil {
		return errors.Join(ErrPluginUninstall, err)
	}
	_ = elemoplugin.RemoveVersion(s.conf.Directory, pluginID, inst.Version)
	return nil
}

func (s *pluginService) Enable(ctx context.Context, pluginID string, scope model.ID, config json.RawMessage) error {
	ctx, span := s.tracer.Start(ctx, "service.pluginService/Enable")
	defer span.End()

	if err := s.requireFeature(ctx, ErrPluginEnable); err != nil {
		return err
	}
	if err := requireAction(ctx, s.permissionService, scope, model.ActionPluginManage); err != nil {
		return errors.Join(ErrPluginEnable, err)
	}
	inst, err := s.repo.GetInstallation(ctx, pluginID)
	if err != nil {
		return errors.Join(ErrPluginEnable, err)
	}
	act := &model.PluginActivation{PluginID: pluginID, ScopeID: scope, Enabled: true}
	if existing, err := s.repo.GetActivation(ctx, pluginID, scope); err == nil {
		act.Config = existing.Config
		act.CreatedAt = existing.CreatedAt
	} else if !errors.Is(err, repository.ErrNotFound) {
		return errors.Join(ErrPluginEnable, err)
	}
	if len(config) > 0 && string(config) != "null" {
		act.Config = config
	}
	if err := s.validateActivationConfig(ctx, inst.Manifest, act.Config, scope); err != nil {
		return errors.Join(ErrPluginEnable, err)
	}
	if err := s.registry.Start(ctx, pluginID); err != nil {
		return errors.Join(ErrPluginEnable, err)
	}
	if _, err := s.repo.UpsertActivation(ctx, act); err != nil {
		return errors.Join(ErrPluginEnable, err)
	}
	return nil
}

func (s *pluginService) Disable(ctx context.Context, pluginID string, scope model.ID) error {
	ctx, span := s.tracer.Start(ctx, "service.pluginService/Disable")
	defer span.End()

	if err := s.requireFeature(ctx, ErrPluginDisable); err != nil {
		return err
	}
	if err := requireAction(ctx, s.permissionService, scope, model.ActionPluginManage); err != nil {
		return errors.Join(ErrPluginDisable, err)
	}
	act := &model.PluginActivation{PluginID: pluginID, ScopeID: scope, Enabled: false}
	if existing, err := s.repo.GetActivation(ctx, pluginID, scope); err == nil {
		act.Config = existing.Config
		act.CreatedAt = existing.CreatedAt
	} else if !errors.Is(err, repository.ErrNotFound) {
		return errors.Join(ErrPluginDisable, err)
	}
	if _, err := s.repo.UpsertActivation(ctx, act); err != nil {
		return errors.Join(ErrPluginDisable, err)
	}
	acts, err := s.repo.ListActivations(ctx, pluginID)
	if err != nil {
		return errors.Join(ErrPluginDisable, err)
	}
	for _, act := range acts {
		if act.Enabled {
			return nil
		}
	}
	return s.registry.Stop(ctx, pluginID, s.conf.ExecutionTimeout)
}

func (s *pluginService) Get(ctx context.Context, pluginID string) (*model.PluginInstallation, error) {
	ctx, span := s.tracer.Start(ctx, "service.pluginService/Get")
	defer span.End()

	if err := s.requireFeature(ctx, ErrPluginGet); err != nil {
		return nil, err
	}
	if err := requireAction(ctx, s.permissionService, model.InstallationID(), model.ActionPluginInstall); err != nil {
		return nil, errors.Join(ErrPluginGet, err)
	}
	inst, err := s.repo.GetInstallation(ctx, pluginID)
	if err != nil {
		return nil, errors.Join(ErrPluginGet, err)
	}
	if loaded, ok := s.registry.Get(pluginID); ok {
		inst.Status = loaded.Status
		inst.Error = loaded.Error
	}
	return inst, nil
}

func (s *pluginService) List(ctx context.Context) ([]*model.PluginInstallation, error) {
	ctx, span := s.tracer.Start(ctx, "service.pluginService/List")
	defer span.End()

	if err := s.requireFeature(ctx, ErrPluginList); err != nil {
		return nil, err
	}
	if err := requireAction(ctx, s.permissionService, model.InstallationID(), model.ActionPluginInstall); err != nil {
		return nil, errors.Join(ErrPluginList, err)
	}
	list, err := s.repo.ListInstallations(ctx)
	if err != nil {
		return nil, errors.Join(ErrPluginList, err)
	}
	for _, inst := range list {
		if loaded, ok := s.registry.Get(inst.PluginID); ok {
			inst.Status = loaded.Status
			inst.Error = loaded.Error
		}
	}
	return list, nil
}

func (s *pluginService) ListManaged(ctx context.Context, scope model.ID) ([]PluginListItem, error) {
	ctx, span := s.tracer.Start(ctx, "service.pluginService/ListManaged")
	defer span.End()

	if err := s.requireFeature(ctx, ErrPluginList); err != nil {
		return nil, err
	}
	if err := requireAction(ctx, s.permissionService, scope, model.ActionPluginManage); err != nil {
		return nil, errors.Join(ErrPluginList, err)
	}
	list, err := s.repo.ListInstallations(ctx)
	if err != nil {
		return nil, errors.Join(ErrPluginList, err)
	}
	out := make([]PluginListItem, 0, len(list))
	for _, inst := range list {
		if loaded, ok := s.registry.Get(inst.PluginID); ok {
			inst.Status = loaded.Status
			inst.Error = loaded.Error
		}
		item := PluginListItem{Installation: inst}
		act, err := s.repo.GetActivation(ctx, inst.PluginID, scope)
		if err == nil {
			enabled := act.Enabled
			item.Enabled = &enabled
			item.Config = act.Config
		} else if !errors.Is(err, repository.ErrNotFound) {
			return nil, errors.Join(ErrPluginList, err)
		}
		out = append(out, item)
	}
	return out, nil
}

func (s *pluginService) ListFrontend(ctx context.Context, scope model.ID) ([]FrontendPlugin, error) {
	ctx, span := s.tracer.Start(ctx, "service.pluginService/ListFrontend")
	defer span.End()

	if err := s.requireFeature(ctx, ErrPluginList); err != nil {
		return nil, err
	}
	read, ok := model.ReadActionFor(scope.Type)
	if !ok {
		return nil, errors.Join(ErrPluginList, model.ErrInvalidResourceType)
	}
	if err := requireAction(ctx, s.permissionService, scope, read); err != nil {
		return nil, errors.Join(ErrPluginList, err)
	}
	ancestry, err := s.permissionService.ListScopeAncestry(ctx, scope)
	if err != nil {
		return nil, errors.Join(ErrPluginList, err)
	}
	acts, err := s.repo.ListActivationsByScope(ctx, ancestry)
	if err != nil {
		return nil, errors.Join(ErrPluginList, err)
	}
	seen := make(map[string]struct{})
	var out []FrontendPlugin
	for _, act := range acts {
		if _, dup := seen[act.PluginID]; dup {
			continue
		}
		seen[act.PluginID] = struct{}{}
		inst, err := s.repo.GetInstallation(ctx, act.PluginID)
		if err != nil {
			continue
		}
		if inst.Manifest.Frontend == nil {
			continue
		}
		loaded, ok := s.registry.Get(act.PluginID)
		if !ok || !loaded.Status.ServesFrontend() {
			continue
		}
		entry := inst.Manifest.Frontend.Entry
		if entry == "" {
			entry = model.PluginFrontendEntryDefault
		}
		out = append(out, FrontendPlugin{
			ID:         inst.PluginID,
			Version:    inst.Version,
			Entrypoint: entry,
			Module:     inst.Manifest.Frontend.Module,
			Slots:      inst.Manifest.Slots,
		})
	}
	return out, nil
}

func (s *pluginService) GetConfig(ctx context.Context, pluginID string, scope model.ID) (json.RawMessage, error) {
	ctx, span := s.tracer.Start(ctx, "service.pluginService/GetConfig")
	defer span.End()

	if err := s.requireFeature(ctx, ErrPluginGet); err != nil {
		return nil, err
	}
	act, err := s.nearestActivation(ctx, pluginID, scope)
	if err != nil {
		return nil, errors.Join(ErrPluginGet, err)
	}
	if len(act.Config) == 0 {
		return json.RawMessage("{}"), nil
	}
	return act.Config, nil
}

func (s *pluginService) GetManagedConfig(ctx context.Context, pluginID string, scope model.ID) (json.RawMessage, error) {
	ctx, span := s.tracer.Start(ctx, "service.pluginService/GetManagedConfig")
	defer span.End()

	if err := s.requireFeature(ctx, ErrPluginGet); err != nil {
		return nil, err
	}
	if err := requireAction(ctx, s.permissionService, scope, model.ActionPluginManage); err != nil {
		return nil, errors.Join(ErrPluginGet, err)
	}
	act, err := s.repo.GetActivation(ctx, pluginID, scope)
	if errors.Is(err, repository.ErrNotFound) {
		return json.RawMessage("{}"), nil
	}
	if err != nil {
		return nil, errors.Join(ErrPluginGet, err)
	}
	if len(act.Config) == 0 {
		return json.RawMessage("{}"), nil
	}
	return act.Config, nil
}

func (s *pluginService) SetConfig(ctx context.Context, pluginID string, scope model.ID, config json.RawMessage) error {
	ctx, span := s.tracer.Start(ctx, "service.pluginService/SetConfig")
	defer span.End()

	if err := s.requireFeature(ctx, ErrPluginConfig); err != nil {
		return err
	}
	if err := requireAction(ctx, s.permissionService, scope, model.ActionPluginManage); err != nil {
		return errors.Join(ErrPluginConfig, err)
	}
	inst, err := s.repo.GetInstallation(ctx, pluginID)
	if err != nil {
		return errors.Join(ErrPluginConfig, err)
	}
	act, err := s.repo.GetActivation(ctx, pluginID, scope)
	if err != nil {
		return errors.Join(ErrPluginConfig, err)
	}
	if err := s.validateActivationConfig(ctx, inst.Manifest, config, scope); err != nil {
		return errors.Join(ErrPluginConfig, err)
	}
	act.Config = config
	if _, err := s.repo.UpsertActivation(ctx, act); err != nil {
		return errors.Join(ErrPluginConfig, err)
	}
	return nil
}

func (s *pluginService) Invoke(
	ctx context.Context,
	pluginID string,
	req elemoplugin.InvokeRequest,
) (elemoplugin.InvokeResponse, error) {
	ctx, span := s.tracer.Start(ctx, "service.pluginService/Invoke")
	defer span.End()

	if err := s.requireFeature(ctx, ErrPluginInvoke); err != nil {
		return elemoplugin.InvokeResponse{}, err
	}
	scope, err := parseScopeID(req.ScopeID)
	if err != nil {
		return elemoplugin.InvokeResponse{}, errors.Join(ErrPluginInvoke, err)
	}
	if err := s.requireActive(ctx, pluginID, scope); err != nil {
		return elemoplugin.InvokeResponse{}, errors.Join(ErrPluginInvoke, err)
	}
	req.UserID = pkg.CtxUserID(ctx)
	body, err := elemoplugin.EncodeJSON(req)
	if err != nil {
		return elemoplugin.InvokeResponse{}, errors.Join(ErrPluginInvoke, err)
	}
	out, err := s.registry.Call(ctx, pluginID, req.Function, body)
	if err != nil {
		if errors.Is(err, elemoplugin.ErrPluginNotLoaded) || errors.Is(err, elemoplugin.ErrInvokeDisabled) {
			return elemoplugin.InvokeResponse{}, errors.Join(ErrPluginInvoke, repository.ErrNotFound)
		}
		return elemoplugin.InvokeResponse{}, errors.Join(ErrPluginInvoke, err)
	}
	var resp elemoplugin.InvokeResponse
	if err := elemoplugin.DecodeJSON(out, &resp); err != nil {
		return elemoplugin.InvokeResponse{OK: true, Data: out}, nil
	}
	if !resp.OK {
		s.logger.Warn(ctx, "plugin invoke failed", log.WithMetadata(map[string]any{
			"plugin_id": pluginID,
			"function":  req.Function,
			"error":     resp.Error,
		}))
	}
	return resp, nil
}

func (s *pluginService) AssetPath(ctx context.Context, pluginID, version, rel string) (string, error) {
	ctx, span := s.tracer.Start(ctx, "service.pluginService/AssetPath")
	defer span.End()
	if err := ctx.Err(); err != nil {
		return "", errors.Join(ErrPluginAsset, err)
	}

	loaded, ok := s.registry.Get(pluginID)
	if !ok || loaded.Version != version {
		return "", errors.Join(ErrPluginAsset, repository.ErrNotFound)
	}
	if !loaded.Status.ServesFrontend() {
		return "", errors.Join(ErrPluginAsset, repository.ErrNotFound)
	}
	path, err := safepath.Normalize(loaded.Root, rel)
	if err != nil {
		return "", errors.Join(ErrPluginAsset, err)
	}
	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		return "", errors.Join(ErrPluginAsset, repository.ErrNotFound)
	}
	return path, nil
}

func (s *pluginService) Restore(ctx context.Context) error {
	list, err := s.repo.ListInstallations(ctx)
	if err != nil {
		return err
	}
	for _, inst := range list {
		root := elemoplugin.InstallDirectory(s.conf.Directory, inst.PluginID, inst.Version)
		pkg := elemoplugin.ExtractedPackage{Root: root, Manifest: inst.Manifest}
		wasm, err := elemoplugin.ReadWASM(pkg)
		if err != nil {
			s.logger.Warn(ctx, "failed to restore plugin wasm", log.WithError(err))
			wasm = nil
		}
		if err := s.registry.Put(ctx, elemoplugin.LoadedPlugin{
			ID: inst.PluginID, Version: inst.Version, Manifest: inst.Manifest, Root: root, Status: model.PluginStatusInstalled,
		}, wasm); err != nil {
			s.logger.Warn(ctx, "failed to load plugin", log.WithError(err))
		}
		acts, err := s.repo.ListActivations(ctx, inst.PluginID)
		if err != nil {
			s.logger.Warn(ctx, "failed to list plugin activations", log.WithError(err))
			continue
		}
		for _, act := range acts {
			if act.Enabled {
				if err := s.registry.Start(ctx, inst.PluginID); err != nil {
					s.logger.Warn(ctx, "failed to start plugin", log.WithError(err))
				}
				break
			}
		}
	}
	return nil
}

func (s *pluginService) nearestActivation(ctx context.Context, pluginID string, scope model.ID) (*model.PluginActivation, error) {
	ancestry, err := s.permissionService.ListScopeAncestry(ctx, scope)
	if err != nil {
		return nil, err
	}
	acts, err := s.repo.ListActivationsByScope(ctx, ancestry)
	if err != nil {
		return nil, err
	}
	byScope := make(map[string]*model.PluginActivation, len(acts))
	for _, act := range acts {
		if act.PluginID == pluginID && act.Enabled {
			byScope[act.ScopeID.Composite()] = act
		}
	}
	for _, id := range ancestry {
		if act, ok := byScope[id.Composite()]; ok {
			return act, nil
		}
	}
	return nil, repository.ErrNotFound
}

func (s *pluginService) requireActive(ctx context.Context, pluginID string, scope model.ID) error {
	if _, err := s.nearestActivation(ctx, pluginID, scope); err != nil {
		return err
	}
	return nil
}

func (s *pluginService) validateActivationConfig(
	ctx context.Context,
	manifest model.PluginManifest,
	config json.RawMessage,
	scope model.ID,
) error {
	if len(manifest.Config) == 0 {
		return nil
	}
	values := map[string]json.RawMessage{}
	if len(config) > 0 && string(config) != "null" {
		if err := json.Unmarshal(config, &values); err != nil {
			return errors.Join(model.ErrPluginConfigInvalid, err)
		}
	}
	for _, field := range manifest.Config {
		raw, ok := values[field.Name]
		if !ok || len(raw) == 0 || string(raw) == "null" {
			if field.Required {
				return errors.Join(model.ErrPluginConfigInvalid, fmt.Errorf("missing required config field %s", field.Name))
			}
			continue
		}
		switch field.Type {
		case model.PluginConfigFieldTypeStr:
			var s string
			if err := json.Unmarshal(raw, &s); err != nil {
				return errors.Join(model.ErrPluginConfigInvalid, fmt.Errorf("config field %s: expected string", field.Name))
			}
		case model.PluginConfigFieldTypeInteger:
			var n json.Number
			if err := json.Unmarshal(raw, &n); err != nil {
				var i int64
				if err := json.Unmarshal(raw, &i); err != nil {
					return errors.Join(model.ErrPluginConfigInvalid, fmt.Errorf("config field %s: expected integer", field.Name))
				}
			}
		case model.PluginConfigFieldTypeBoolean:
			var b bool
			if err := json.Unmarshal(raw, &b); err != nil {
				return errors.Join(model.ErrPluginConfigInvalid, fmt.Errorf("config field %s: expected boolean", field.Name))
			}
		case model.PluginConfigFieldTypeGraphBinding:
			var binding model.PluginGraphBindingValue
			if err := json.Unmarshal(raw, &binding); err != nil || binding.PluginID == "" || binding.Kind == "" {
				return errors.Join(model.ErrPluginConfigInvalid, fmt.Errorf("config field %s: expected plugin_id and kind", field.Name))
			}
			foreign, ok := manifest.Graph.ForeignKind(field.Foreign)
			if !ok {
				return errors.Join(model.ErrPluginConfigInvalid, fmt.Errorf("config field %s: unknown foreign %s", field.Name, field.Foreign))
			}
			owner, err := s.repo.GetInstallation(ctx, binding.PluginID)
			if err != nil {
				return errors.Join(model.ErrPluginGraphBinding, err)
			}
			kind, ok := owner.Manifest.Graph.NodeKind(binding.Kind)
			if !ok {
				return errors.Join(model.ErrPluginGraphBinding, fmt.Errorf("plugin %s has no kind %s", binding.PluginID, binding.Kind))
			}
			if !foreign.MatchesKind(kind) {
				return errors.Join(model.ErrPluginGraphBinding, fmt.Errorf("kind %s does not match foreign %s", binding.Kind, field.Foreign))
			}
			if err := s.requireActive(ctx, binding.PluginID, scope); err != nil {
				return errors.Join(model.ErrPluginGraphBinding, err)
			}
		}
	}
	return nil
}

func (s *pluginService) bindEvents() {
	if s.bus == nil {
		return
	}
	for _, topic := range []model.PluginEventType{
		model.PluginEventIssueCreated,
		model.PluginEventIssueUpdated,
		model.PluginEventIssueDeleted,
		model.PluginEventProjectCreated,
		model.PluginEventProjectUpdated,
		model.PluginEventExtensionCreated,
		model.PluginEventExtensionUpdated,
		model.PluginEventExtensionDeleted,
	} {
		topic := topic
		sub, err := s.bus.Subscribe(topic, func(ctx context.Context, evt event.Event) error {
			s.enqueuePluginEvent(ctx, evt)
			return nil
		})
		if err != nil {
			continue
		}
		s.subscriptions = append(s.subscriptions, sub)
	}
}

func (s *pluginService) enqueuePluginEvent(ctx context.Context, evt event.Event) {
	cloned := evt
	if evt.Payload != nil {
		cloned.Payload = maps.Clone(evt.Payload)
	}
	user := ctx.Value(pkg.CtxKeyUserID)
	s.eventWG.Add(1)
	go func() {
		defer s.eventWG.Done()
		timeout := s.conf.ExecutionTimeout
		if timeout <= 0 {
			timeout = 5 * time.Second
		}
		bg, cancel := context.WithTimeout(context.WithoutCancel(ctx), timeout)
		defer cancel()
		defer func() {
			if rec := recover(); rec != nil {
				s.logger.Error(bg, "plugin event handler panicked",
					log.WithValue(fmt.Sprint(rec)))
			}
		}()
		if user != nil {
			bg = context.WithValue(bg, pkg.CtxKeyUserID, user)
		}
		s.dispatchEvent(bg, cloned)
	}()
}

func (s *pluginService) dispatchEvent(ctx context.Context, evt event.Event) {
	timeout := s.conf.ExecutionTimeout
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	payload, _ := elemoplugin.EncodeJSON(evt)
	invoke := elemoplugin.InvokeRequest{Function: "onEvent", Payload: payload}
	if scope, ok := evt.Payload["scope_id"].(string); ok {
		invoke.ScopeID = scope
	}
	if uid := pkg.CtxUserID(ctx); uid != "" {
		invoke.UserID = uid
	}
	body, err := elemoplugin.EncodeJSON(invoke)
	if err != nil {
		return
	}
	for _, loaded := range s.registry.List() {
		if loaded.Status != model.PluginStatusActive {
			continue
		}
		want := false
		for _, t := range loaded.Manifest.Events {
			if t == evt.Type {
				want = true
				break
			}
		}
		if !want {
			continue
		}
		if isExtensionEvent(evt.Type) && !s.shouldDeliverExtensionEvent(ctx, loaded, evt) {
			continue
		}
		_, err := s.registry.Call(ctx, loaded.ID, "onEvent", body)
		if err != nil {
			s.logger.Warn(ctx, "plugin event handler failed", log.WithError(err))
		}
	}
}

func isExtensionEvent(t model.PluginEventType) bool {
	switch t {
	case model.PluginEventExtensionCreated, model.PluginEventExtensionUpdated, model.PluginEventExtensionDeleted:
		return true
	default:
		return false
	}
}

func (s *pluginService) shouldDeliverExtensionEvent(ctx context.Context, loaded elemoplugin.LoadedPlugin, evt event.Event) bool {
	owner, _ := evt.Payload["plugin_id"].(string)
	kind, _ := evt.Payload["kind"].(string)
	if owner == loaded.ID {
		return true
	}
	if loaded.Manifest.Graph == nil || len(loaded.Manifest.Graph.Foreign) == 0 {
		return true
	}
	act, err := s.nearestActivation(ctx, loaded.ID, evt.Resource)
	if err != nil {
		return false
	}
	return model.BindingMatches(act.Config, loaded.Manifest.Config, owner, kind)
}

func parseScopeID(raw string) (model.ID, error) {
	if raw == "" {
		return model.ID{}, model.ErrInvalidID
	}
	if id, err := model.ParseCompositeID(raw); err == nil {
		return id, nil
	}
	return model.NewIDFromString(raw, model.ResourceTypeOrganization.String())
}

func parseTypedID(raw, typ string) (model.ID, error) {
	if raw == "" {
		return model.ID{}, model.ErrInvalidID
	}
	if id, err := model.ParseCompositeID(raw); err == nil {
		if typ != "" && id.Type.String() != typ {
			return model.ID{}, model.ErrInvalidID
		}
		return id, nil
	}
	return model.NewIDFromString(raw, typ)
}

func copyDir(src, dest string) error {
	if err := os.MkdirAll(dest, 0o750); err != nil {
		return err
	}
	srcRoot, err := os.OpenRoot(src)
	if err != nil {
		return err
	}
	defer srcRoot.Close()
	dstRoot, err := os.OpenRoot(dest)
	if err != nil {
		return err
	}
	defer dstRoot.Close()

	return fs.WalkDir(srcRoot.FS(), ".", func(rel string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}
		if d.IsDir() {
			return dstRoot.Mkdir(rel, 0o750)
		}
		data, err := srcRoot.ReadFile(rel)
		if err != nil {
			return err
		}
		return dstRoot.WriteFile(rel, data, 0o600)
	})
}

func assertAdditiveGraph(oldS, newS *model.PluginGraphSchema) error {
	if oldS == nil {
		return nil
	}
	if newS == nil {
		return ErrPluginSchemaNotAdditive
	}
	for _, oldNode := range oldS.Nodes {
		next, ok := newS.NodeKind(oldNode.Kind)
		if !ok {
			return ErrPluginSchemaNotAdditive
		}
		oldProps := map[string]model.PluginGraphPropertyDecl{}
		for _, p := range oldNode.Properties {
			oldProps[p.Name] = p
		}
		newProps := map[string]model.PluginGraphPropertyDecl{}
		for _, p := range next.Properties {
			newProps[p.Name] = p
		}
		for name, p := range oldProps {
			np, ok := newProps[name]
			if !ok || np.Type != p.Type {
				return ErrPluginSchemaNotAdditive
			}
			if p.Required && !np.Required {
				return ErrPluginSchemaNotAdditive
			}
		}
	}
	for _, oldRel := range oldS.Relations {
		if _, ok := newS.RelationKind(oldRel.Kind); !ok {
			return ErrPluginSchemaNotAdditive
		}
	}
	for _, oldForeign := range oldS.Foreign {
		next, ok := newS.ForeignKind(oldForeign.Name)
		if !ok || next.Parent != oldForeign.Parent {
			return ErrPluginSchemaNotAdditive
		}
	}
	return nil
}

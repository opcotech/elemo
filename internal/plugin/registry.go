package plugin

import (
	"context"
	"sync"
	"time"

	"github.com/opcotech/elemo/internal/model"
)

// LoadedPlugin is a registry snapshot. Callers must not mutate it.
type LoadedPlugin struct {
	ID       string
	Version  string
	Manifest model.PluginManifest
	Root     string
	Status   model.PluginStatus
	Error    string
}

type registryEntry struct {
	loaded   LoadedPlugin
	wasm     []byte
	inFlight sync.WaitGroup
}

// Registry is an in-memory, concurrency-safe catalog of installed plugins.
type Registry struct {
	mu      sync.RWMutex
	entries map[string]*registryEntry
	runtime Runtime
}

// NewRegistry returns an empty registry backed by runtime.
func NewRegistry(runtime Runtime) *Registry {
	if runtime == nil {
		runtime = NoopRuntime{}
	}
	return &Registry{
		entries: make(map[string]*registryEntry),
		runtime: runtime,
	}
}

func (r *Registry) Runtime() Runtime {
	return r.runtime
}

// Put records an installation. WASM is loaded but not started.
func (r *Registry) Put(ctx context.Context, loaded LoadedPlugin, wasm []byte) error {
	if err := loaded.Manifest.Validate(); err != nil {
		return err
	}
	if loaded.Status == model.PluginStatusUnknown {
		loaded.Status = model.PluginStatusInstalled
	}

	r.mu.Lock()
	prev := r.entries[loaded.ID]
	entry := &registryEntry{loaded: loaded, wasm: wasm}
	r.entries[loaded.ID] = entry
	r.mu.Unlock()

	if prev != nil {
		prev.inFlight.Wait()
		_ = r.runtime.Stop(ctx, loaded.ID)
	}

	return r.runtime.Load(ctx, Plugin{ID: loaded.ID, Version: loaded.Version, WASM: wasm})
}

// Get returns a copy of the loaded plugin.
func (r *Registry) Get(pluginID string) (LoadedPlugin, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	entry, ok := r.entries[pluginID]
	if !ok {
		return LoadedPlugin{}, false
	}
	return entry.loaded, true
}

// List returns all loaded plugins in unspecified order.
func (r *Registry) List() []LoadedPlugin {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]LoadedPlugin, 0, len(r.entries))
	for _, entry := range r.entries {
		out = append(out, entry.loaded)
	}
	return out
}

// Start starts the WASM guest. The registry lock is not held during Start.
func (r *Registry) Start(ctx context.Context, pluginID string) error {
	entry := r.lookup(pluginID)
	if entry == nil {
		return ErrPluginNotLoaded
	}
	r.setStatus(pluginID, model.PluginStatusStarting, "")
	if err := r.runtime.Start(ctx, pluginID); err != nil {
		r.setStatus(pluginID, model.PluginStatusFailed, err.Error())
		return err
	}
	r.setStatus(pluginID, model.PluginStatusActive, "")
	return nil
}

// Stop stops the WASM guest after in-flight calls drain or timeout.
func (r *Registry) Stop(ctx context.Context, pluginID string, drain time.Duration) error {
	entry := r.lookup(pluginID)
	if entry == nil {
		return nil
	}
	r.setStatus(pluginID, model.PluginStatusDisabling, "")
	done := make(chan struct{})
	go func() {
		entry.inFlight.Wait()
		close(done)
	}()
	timer := time.NewTimer(drain)
	defer timer.Stop()
	select {
	case <-done:
	case <-timer.C:
	case <-ctx.Done():
	}
	err := r.runtime.Stop(ctx, pluginID)
	if err != nil {
		r.setStatus(pluginID, model.PluginStatusFailed, err.Error())
		return err
	}
	r.setStatus(pluginID, model.PluginStatusDisabled, "")
	return nil
}

// Remove unloads a plugin after Stop.
func (r *Registry) Remove(ctx context.Context, pluginID string) error {
	if err := r.Stop(ctx, pluginID, 2*time.Second); err != nil {
		return err
	}
	r.mu.Lock()
	delete(r.entries, pluginID)
	r.mu.Unlock()
	return nil
}

// Call invokes a WASM function without holding the registry lock.
func (r *Registry) Call(ctx context.Context, pluginID, function string, input []byte) ([]byte, error) {
	entry := r.lookup(pluginID)
	if entry == nil {
		return nil, ErrPluginNotLoaded
	}
	if entry.loaded.Status != model.PluginStatusActive {
		return nil, ErrInvokeDisabled
	}
	entry.inFlight.Add(1)
	defer entry.inFlight.Done()
	return r.runtime.Call(ctx, pluginID, function, input)
}

func (r *Registry) lookup(pluginID string) *registryEntry {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.entries[pluginID]
}

func (r *Registry) setStatus(pluginID string, status model.PluginStatus, errMsg string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if entry, ok := r.entries[pluginID]; ok {
		entry.loaded.Status = status
		entry.loaded.Error = errMsg
	}
}

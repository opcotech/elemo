package plugin

import (
	"context"
	"errors"
)

var (
	ErrPluginNotLoaded  = errors.New("plugin is not loaded")
	ErrPluginNotStarted = errors.New("plugin is not started")
	ErrRuntimeTrap      = errors.New("plugin wasm trap")
	ErrRuntimeTimeout   = errors.New("plugin execution timed out")
	ErrNotReactor       = errors.New("plugin wasm is not a wasi reactor")
)

// Plugin is the runtime load descriptor.
type Plugin struct {
	ID      string
	Version string
	WASM    []byte
}

// Runtime loads and executes WASM guests. Implementations must isolate traps
// and honor ctx cancellation.
type Runtime interface {
	Load(ctx context.Context, plugin Plugin) error
	Start(ctx context.Context, pluginID string) error
	Stop(ctx context.Context, pluginID string) error
	Call(ctx context.Context, pluginID, function string, input []byte) ([]byte, error)
}

// NoopRuntime is a Runtime that accepts empty (frontend-only) plugins.
type NoopRuntime struct{}

func (NoopRuntime) Load(_ context.Context, plugin Plugin) error {
	if len(plugin.WASM) > 0 {
		return errors.New("noop runtime cannot load wasm")
	}
	return nil
}

func (NoopRuntime) Start(context.Context, string) error { return nil }

func (NoopRuntime) Stop(context.Context, string) error { return nil }

func (NoopRuntime) Call(context.Context, string, string, []byte) ([]byte, error) {
	return nil, ErrPluginNotStarted
}

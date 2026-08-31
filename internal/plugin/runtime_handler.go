package plugin

import (
	"context"
	"sync"
)

// HandlerRuntime is a test Runtime that dispatches to Go handlers.
type HandlerRuntime struct {
	mu       sync.Mutex
	handlers map[string]func(function string, input []byte) ([]byte, error)
	started  map[string]bool
	loaded   map[string][]byte
}

func NewHandlerRuntime() *HandlerRuntime {
	return &HandlerRuntime{
		handlers: make(map[string]func(string, []byte) ([]byte, error)),
		started:  make(map[string]bool),
		loaded:   make(map[string][]byte),
	}
}

func (r *HandlerRuntime) SetHandler(pluginID string, h func(function string, input []byte) ([]byte, error)) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.handlers[pluginID] = h
}

func (r *HandlerRuntime) Load(_ context.Context, plugin Plugin) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.loaded[plugin.ID] = plugin.WASM
	return nil
}

func (r *HandlerRuntime) Start(_ context.Context, pluginID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.started[pluginID] = true
	return nil
}

func (r *HandlerRuntime) Stop(_ context.Context, pluginID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.started[pluginID] = false
	return nil
}

func (r *HandlerRuntime) Call(_ context.Context, pluginID, function string, input []byte) ([]byte, error) {
	r.mu.Lock()
	h := r.handlers[pluginID]
	started := r.started[pluginID]
	r.mu.Unlock()
	if !started {
		return nil, ErrPluginNotStarted
	}
	if h == nil {
		return []byte(`{"ok":true}`), nil
	}
	return h(function, input)
}

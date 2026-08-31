package plugin

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

const PluginAPIVersion = "1"

var (
	ErrUnknownHostMethod = errors.New("unknown host method")
	ErrCapabilityDenied  = errors.New("plugin capability denied")
	ErrInvokeDisabled    = errors.New("plugin is not enabled")
	ErrRuntimeCall       = errors.New("plugin call failed")
)

// InvokeRequest is the versioned JSON body for POST /invoke and WASM dispatch.
type InvokeRequest struct {
	Function string          `json:"function"`
	ScopeID  string          `json:"scopeId"`
	UserID   string          `json:"userId,omitempty"`
	Payload  json.RawMessage `json:"payload"`
}

// InvokeResponse is the versioned JSON result of a plugin function.
type InvokeResponse struct {
	OK    bool            `json:"ok"`
	Error string          `json:"error,omitempty"`
	Data  json.RawMessage `json:"data,omitempty"`
}

// HostRequest is a capability-gated call from a WASM guest to Elemo.
type HostRequest struct {
	Method  string          `json:"method"`
	ScopeID string          `json:"scopeId"`
	Payload json.RawMessage `json:"payload"`
}

// HostResponse is returned to the WASM guest.
type HostResponse struct {
	OK    bool            `json:"ok"`
	Error string          `json:"error,omitempty"`
	Data  json.RawMessage `json:"data,omitempty"`
}

// Host executes Plugin API methods. Implemented in internal/service.
type Host interface {
	Call(ctx context.Context, pluginID string, req HostRequest) (HostResponse, error)
}

func EncodeJSON(v any) ([]byte, error) {
	return json.Marshal(v)
}

func DecodeJSON(data []byte, v any) error {
	if len(data) == 0 {
		return nil
	}
	return json.Unmarshal(data, v)
}

func HostError(err error) HostResponse {
	if err == nil {
		return HostResponse{OK: true}
	}
	return HostResponse{OK: false, Error: err.Error()}
}

func HostOK(v any) (HostResponse, error) {
	if v == nil {
		return HostResponse{OK: true}, nil
	}
	data, err := json.Marshal(v)
	if err != nil {
		return HostResponse{}, err
	}
	return HostResponse{OK: true, Data: data}, nil
}

func InvokeError(err error) InvokeResponse {
	if err == nil {
		return InvokeResponse{OK: true}
	}
	return InvokeResponse{OK: false, Error: err.Error()}
}

func FormatTimeout(d time.Duration) string {
	return fmt.Sprintf("%s timeout", d)
}

// EncodeGuestRequest builds the JSON envelope written into guest linear
// memory. If input is already a full InvokeRequest (function set), it is
// forwarded so ScopeID is preserved.
func EncodeGuestRequest(function string, input []byte) ([]byte, error) {
	var req InvokeRequest
	if len(input) > 0 && json.Unmarshal(input, &req) == nil && req.Function != "" {
		return EncodeJSON(req)
	}
	var payload json.RawMessage
	if len(input) > 0 && json.Valid(input) {
		payload = json.RawMessage(input)
	}
	return EncodeJSON(InvokeRequest{Function: function, Payload: payload})
}

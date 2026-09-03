// Package plugin is the public SDK for Elemo WASM plugins.
//
// Plugins compile with
// CGO_ENABLED=0 GOOS=wasip1 GOARCH=wasm go build -buildmode=c-shared
// and must not import github.com/opcotech/elemo/internal.
package plugin

import "encoding/json"

const (
	APIVersion = "1"
	maxIO      = 1 << 20
)

// Handler dispatches a named plugin function.
type Handler func(req Request) ([]byte, error)

// Request is the JSON envelope Elemo sends to elemo_call.
type Request struct {
	Function string          `json:"function"`
	ScopeID  string          `json:"scopeId"`
	UserID   string          `json:"userId,omitempty"`
	Payload  json.RawMessage `json:"payload"`
}

// Response is the JSON envelope returned to Elemo.
type Response struct {
	OK    bool            `json:"ok"`
	Error string          `json:"error,omitempty"`
	Data  json.RawMessage `json:"data,omitempty"`
}

// HostRequest is sent to Elemo via host_call.
type HostRequest struct {
	Method  string          `json:"method"`
	ScopeID string          `json:"scopeId"`
	Payload json.RawMessage `json:"payload"`
}

// HostResponse is returned from Elemo host_call.
type HostResponse struct {
	OK    bool            `json:"ok"`
	Error string          `json:"error,omitempty"`
	Data  json.RawMessage `json:"data,omitempty"`
}

// Register sets the plugin function handler. Call from init.
func Register(h Handler) {
	setHandler(h)
}

// Reply encodes a successful plugin response.
func Reply(v any) ([]byte, error) {
	if v == nil {
		return mustJSON(Response{OK: true}), nil
	}
	data, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return json.Marshal(Response{OK: true, Data: data})
}

func mustJSON(v any) []byte {
	b, _ := json.Marshal(v)
	return b
}

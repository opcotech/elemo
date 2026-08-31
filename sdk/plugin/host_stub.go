//go:build !wasip1

package plugin

import (
	"encoding/json"
	"errors"
)

// HostFn is the host-call implementation used outside WASI. Tests replace it
// to stub Elemo host methods. The default returns an error so the SDK still
// compiles in the Elemo module.
var HostFn = func(_, _ string, _ any) (json.RawMessage, error) {
	return nil, errors.New("plugin host calls require GOOS=wasip1")
}

// Host is only available inside a WASI plugin unless HostFn is replaced.
func Host(method, scopeID string, payload any) (json.RawMessage, error) {
	return HostFn(method, scopeID, payload)
}

func setHandler(_ Handler) {}

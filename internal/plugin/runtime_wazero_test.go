package plugin_test

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/opcotech/elemo/internal/plugin"
)

func TestWazeroRuntimeIntegration_StartInitializesGoReactor(t *testing.T) {
	if testing.Short() {
		t.Skip("compiles a wasip1 guest")
	}

	wasm := buildWASIGuest(t, true)
	ctx := context.Background()
	rt := plugin.NewWazeroRuntime(nil, 15*time.Second)
	require.NoError(t, rt.Load(ctx, plugin.Plugin{
		ID:   "com.elemo.testguest",
		WASM: wasm,
	}))
	require.NoError(t, rt.Start(ctx, "com.elemo.testguest"))

	out, err := rt.Call(ctx, "com.elemo.testguest", "ping", nil)
	require.NoError(t, err)

	var resp plugin.InvokeResponse
	require.NoError(t, plugin.DecodeJSON(out, &resp))
	require.True(t, resp.OK)
	require.JSONEq(t, `{"pong":true}`, string(resp.Data))

	require.NoError(t, rt.Stop(ctx, "com.elemo.testguest"))
}

type recordingHost struct {
	calls []plugin.HostRequest
}

func (h *recordingHost) Call(_ context.Context, _ string, req plugin.HostRequest) (plugin.HostResponse, error) {
	h.calls = append(h.calls, req)
	switch req.Method {
	case "plugin.storage.get":
		return plugin.HostOK(map[string]any{"running": false})
	case "plugin.storage.set", "plugin.storage.delete":
		return plugin.HostOK(nil)
	default:
		return plugin.HostError(errUnexpectedMethod(req.Method)), nil
	}
}

type unexpectedMethodError string

func (e unexpectedMethodError) Error() string { return "unexpected host method " + string(e) }

func errUnexpectedMethod(method string) error {
	return unexpectedMethodError(method)
}

func TestWazeroRuntimeIntegration_GuestHostCall(t *testing.T) {
	if testing.Short() {
		t.Skip("compiles a wasip1 guest")
	}

	wasm := buildWASIGuest(t, true)
	ctx := context.Background()
	host := &recordingHost{}
	rt := plugin.NewWazeroRuntime(host, 15*time.Second)
	require.NoError(t, rt.Load(ctx, plugin.Plugin{
		ID:   "com.elemo.testguest",
		WASM: wasm,
	}))
	require.NoError(t, rt.Start(ctx, "com.elemo.testguest"))

	out, err := rt.Call(ctx, "com.elemo.testguest", "now", nil)
	require.NoError(t, err)
	var nowResp plugin.InvokeResponse
	require.NoError(t, plugin.DecodeJSON(out, &nowResp))
	require.True(t, nowResp.OK, string(out))

	body, err := plugin.EncodeJSON(plugin.InvokeRequest{
		Function: "hostping",
		ScopeID:  "Namespace:01hmguestscope0000000",
	})
	require.NoError(t, err)
	out, err = rt.Call(ctx, "com.elemo.testguest", "hostping", body)
	require.NoError(t, err, "host_call from elemo_call")

	var resp plugin.InvokeResponse
	require.NoError(t, plugin.DecodeJSON(out, &resp))
	require.True(t, resp.OK, "guest error: %s", resp.Error)
	require.Len(t, host.calls, 1)
	require.Equal(t, "plugin.storage.get", host.calls[0].Method)
	require.Equal(t, "Namespace:01hmguestscope0000000", host.calls[0].ScopeID)

	require.NoError(t, rt.Stop(ctx, "com.elemo.testguest"))
}

func TestWazeroRuntimeIntegration_SerializesConcurrentCalls(t *testing.T) {
	if testing.Short() {
		t.Skip("compiles a wasip1 guest")
	}

	wasm := buildWASIGuest(t, true)
	ctx := context.Background()
	host := &recordingHost{}
	rt := plugin.NewWazeroRuntime(host, 15*time.Second)
	require.NoError(t, rt.Load(ctx, plugin.Plugin{
		ID:   "com.elemo.testguest",
		WASM: wasm,
	}))
	require.NoError(t, rt.Start(ctx, "com.elemo.testguest"))

	const n = 16
	var wg sync.WaitGroup
	errs := make(chan error, n)
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
			out, err := rt.Call(ctx, "com.elemo.testguest", "ping", nil)
			if err != nil {
				errs <- err
				return
			}
			var resp plugin.InvokeResponse
			if err := plugin.DecodeJSON(out, &resp); err != nil {
				errs <- err
				return
			}
			if !resp.OK {
				errs <- errors.New(resp.Error)
			}
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		require.NoError(t, err)
	}

	require.NoError(t, rt.Stop(ctx, "com.elemo.testguest"))
}

func TestWazeroRuntimeIntegration_ReinstantiatesAfterGuestExit(t *testing.T) {
	if testing.Short() {
		t.Skip("compiles a wasip1 guest")
	}

	wasm := buildWASIGuest(t, true)
	ctx := context.Background()
	rt := plugin.NewWazeroRuntime(nil, 15*time.Second)
	require.NoError(t, rt.Load(ctx, plugin.Plugin{
		ID:   "com.elemo.testguest",
		WASM: wasm,
	}))
	require.NoError(t, rt.Start(ctx, "com.elemo.testguest"))

	_, err := rt.Call(ctx, "com.elemo.testguest", "crash", nil)
	require.Error(t, err)
	require.ErrorIs(t, err, plugin.ErrRuntimeTrap)

	out, err := rt.Call(ctx, "com.elemo.testguest", "ping", nil)
	require.NoError(t, err)
	var resp plugin.InvokeResponse
	require.NoError(t, plugin.DecodeJSON(out, &resp))
	require.True(t, resp.OK, string(out))

	require.NoError(t, rt.Stop(ctx, "com.elemo.testguest"))
}

func TestWazeroRuntimeIntegration_TimeTrackingTimerStart(t *testing.T) {
	if testing.Short() {
		t.Skip("compiles the timetracking plugin")
	}

	wasm := buildTimeTrackingWASM(t)
	ctx := context.Background()
	host := &recordingHost{}
	rt := plugin.NewWazeroRuntime(host, 15*time.Second)
	require.NoError(t, rt.Load(ctx, plugin.Plugin{
		ID:   "com.elemo.timetracking",
		WASM: wasm,
	}))
	require.NoError(t, rt.Start(ctx, "com.elemo.timetracking"))

	body, err := plugin.EncodeJSON(plugin.InvokeRequest{
		Function: "timer.start",
		ScopeID:  "Project:01hmtimerstart000000",
		Payload:  []byte(`{"issueId":"01hmissue000000000000"}`),
	})
	require.NoError(t, err)

	out, err := rt.Call(ctx, "com.elemo.timetracking", "timer.start", body)
	require.NoError(t, err, "timer.start")

	var resp plugin.InvokeResponse
	require.NoError(t, plugin.DecodeJSON(out, &resp))
	require.True(t, resp.OK, "guest error: %s", resp.Error)
	require.Len(t, host.calls, 1)
	require.Equal(t, "plugin.storage.set", host.calls[0].Method)

	require.NoError(t, rt.Stop(ctx, "com.elemo.timetracking"))
}

func buildTimeTrackingWASM(t *testing.T) []byte {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	require.True(t, ok)
	pluginDir := filepath.Join(filepath.Dir(file), "..", "..", "plugins", "timetracking")
	out := filepath.Join(t.TempDir(), "plugin.wasm")

	cmd := exec.CommandContext(t.Context(), "go", "build", "-buildmode=c-shared", "-o", out, ".")
	cmd.Dir = pluginDir
	cmd.Env = append(os.Environ(), "GOOS=wasip1", "GOARCH=wasm", "CGO_ENABLED=0")
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, string(output))

	wasm, err := os.ReadFile(out)
	require.NoError(t, err)
	require.NotEmpty(t, wasm)
	return wasm
}

func TestWazeroRuntimeIntegration_LoadRejectsCommandModule(t *testing.T) {
	if testing.Short() {
		t.Skip("compiles a wasip1 guest")
	}

	wasm := buildWASIGuest(t, false)
	ctx := context.Background()
	rt := plugin.NewWazeroRuntime(nil, 15*time.Second)
	err := rt.Load(ctx, plugin.Plugin{
		ID:   "com.elemo.testguest",
		WASM: wasm,
	})
	require.ErrorIs(t, err, plugin.ErrNotReactor)

	err = rt.Start(ctx, "com.elemo.testguest")
	require.ErrorIs(t, err, plugin.ErrNotReactor)
}

func buildWASIGuest(t *testing.T, reactor bool) []byte {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	require.True(t, ok)
	guestDir := filepath.Join(filepath.Dir(file), "testdata", "wasip1guest")
	out := filepath.Join(t.TempDir(), "plugin.wasm")

	args := []string{"build", "-o", out, "."}
	if reactor {
		args = []string{"build", "-buildmode=c-shared", "-o", out, "."}
	}
	cmd := exec.CommandContext(t.Context(), "go", args...)
	cmd.Dir = guestDir
	cmd.Env = append(os.Environ(), "GOOS=wasip1", "GOARCH=wasm", "CGO_ENABLED=0")
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, string(output))

	wasm, err := os.ReadFile(out)
	require.NoError(t, err)
	require.NotEmpty(t, wasm)
	return wasm
}

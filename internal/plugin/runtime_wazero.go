package plugin

import (
	"bytes"
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"math"
	"sync"
	"time"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

const (
	hostModuleName   = "elemo"
	exportCall       = "elemo_call"
	exportStart      = "elemo_start"
	exportStop       = "elemo_stop"
	exportInPtr      = "elemo_in_ptr"
	exportOutPtr     = "elemo_out_ptr"
	exportInitialize = "_initialize"
	maxGuestIO       = 1 << 20
)

type wazeroInstance struct {
	runtime wazero.Runtime
	module  api.Module
	started bool
	stderr  bytes.Buffer
}

// WazeroRuntime executes GOOS=wasip1 guests behind Runtime.
type WazeroRuntime struct {
	mu       sync.Mutex
	timeout  time.Duration
	host     Host
	compiled map[string][]byte
	live     map[string]*wazeroInstance
	callMu   map[string]*sync.Mutex
}

// NewWazeroRuntime returns a wazero-backed Runtime. timeout is applied to
// every Call/Start/Stop in addition to ctx.
func NewWazeroRuntime(host Host, timeout time.Duration) *WazeroRuntime {
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	return &WazeroRuntime{
		timeout:  timeout,
		host:     host,
		compiled: make(map[string][]byte),
		live:     make(map[string]*wazeroInstance),
		callMu:   make(map[string]*sync.Mutex),
	}
}

func (r *WazeroRuntime) Load(_ context.Context, plugin Plugin) error {
	if plugin.ID == "" {
		return errors.New("plugin id is required")
	}
	if len(plugin.WASM) == 0 {
		r.mu.Lock()
		delete(r.compiled, plugin.ID)
		r.mu.Unlock()
		return nil
	}
	copied := make([]byte, len(plugin.WASM))
	copy(copied, plugin.WASM)
	err := requireWASIReactor(copied)
	r.mu.Lock()
	r.compiled[plugin.ID] = copied
	r.mu.Unlock()
	return err
}

func (r *WazeroRuntime) pluginCallMu(pluginID string) *sync.Mutex {
	r.mu.Lock()
	defer r.mu.Unlock()
	m, ok := r.callMu[pluginID]
	if !ok {
		m = &sync.Mutex{}
		r.callMu[pluginID] = m
	}
	return m
}

func (r *WazeroRuntime) Start(ctx context.Context, pluginID string) error {
	ctx, cancel := withTimeout(ctx, r.timeout)
	defer cancel()

	mu := r.pluginCallMu(pluginID)
	mu.Lock()
	defer mu.Unlock()
	_, err := r.ensureLive(ctx, pluginID, false)
	return err
}

func (r *WazeroRuntime) Stop(ctx context.Context, pluginID string) error {
	ctx, cancel := withTimeout(ctx, r.timeout)
	defer cancel()

	mu := r.pluginCallMu(pluginID)
	mu.Lock()
	defer mu.Unlock()
	return r.stopLive(ctx, pluginID)
}

func (r *WazeroRuntime) Call(ctx context.Context, pluginID, function string, input []byte) (out []byte, err error) {
	if len(input) > maxGuestIO {
		return nil, errors.New("plugin input is too large")
	}
	ctx, cancel := withTimeout(ctx, r.timeout)
	defer cancel()

	defer func() {
		if rec := recover(); rec != nil {
			err = fmt.Errorf("%w: panic: %v", ErrRuntimeTrap, rec)
		}
	}()

	body, err := EncodeGuestRequest(function, input)
	if err != nil {
		return nil, err
	}

	// WASI reactors are not thread-safe, so Calls for one plugin serialize.
	// The lock is not reentrant: host_call must not invoke Runtime.Call for
	// the same plugin on this goroutine. Plugin domain events are enqueued
	// asynchronously for that reason.
	mu := r.pluginCallMu(pluginID)
	mu.Lock()
	defer mu.Unlock()

	inst, err := r.ensureLive(ctx, pluginID, true)
	if err != nil {
		return nil, err
	}

	out, err = r.callGuest(ctx, inst, body)
	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return nil, ErrRuntimeTimeout
		}
		if msg := bytes.TrimSpace(inst.stderr.Bytes()); len(msg) > 0 {
			return nil, errors.Join(ErrRuntimeTrap, err, errors.New(string(truncateBytes(msg, 512))))
		}
		return nil, errors.Join(ErrRuntimeTrap, err)
	}
	return out, nil
}

func (r *WazeroRuntime) ensureLive(ctx context.Context, pluginID string, required bool) (*wazeroInstance, error) {
	r.mu.Lock()
	inst := r.live[pluginID]
	wasm := r.compiled[pluginID]
	if inst != nil && inst.module != nil && !inst.module.IsClosed() {
		r.mu.Unlock()
		return inst, nil
	}
	if inst != nil {
		delete(r.live, pluginID)
		r.mu.Unlock()
		_ = closeInstance(ctx, inst)
	} else {
		r.mu.Unlock()
	}

	if len(wasm) == 0 {
		if required {
			return nil, ErrPluginNotStarted
		}
		return nil, nil
	}
	if err := requireWASIReactor(wasm); err != nil {
		return nil, err
	}

	inst, err := r.instantiate(ctx, pluginID, wasm)
	if err != nil {
		return nil, err
	}
	if fn := inst.module.ExportedFunction(exportStart); fn != nil {
		if _, err := callExport(ctx, fn); err != nil {
			_ = closeInstance(ctx, inst)
			return nil, errors.Join(ErrRuntimeTrap, err)
		}
	}
	inst.started = true

	r.mu.Lock()
	if old := r.live[pluginID]; old != nil {
		r.mu.Unlock()
		_ = closeInstance(ctx, inst)
		if old.module != nil && !old.module.IsClosed() {
			return old, nil
		}
		return nil, ErrPluginNotStarted
	}
	r.live[pluginID] = inst
	r.mu.Unlock()
	return inst, nil
}

func (r *WazeroRuntime) stopLive(ctx context.Context, pluginID string) error {
	r.mu.Lock()
	inst := r.live[pluginID]
	delete(r.live, pluginID)
	r.mu.Unlock()
	if inst == nil {
		return nil
	}

	var callErr error
	if inst.started && inst.module != nil && !inst.module.IsClosed() {
		if fn := inst.module.ExportedFunction(exportStop); fn != nil {
			_, callErr = callExport(ctx, fn)
		}
	}
	closeErr := closeInstance(ctx, inst)
	if callErr != nil {
		return errors.Join(ErrRuntimeTrap, callErr, closeErr)
	}
	return closeErr
}

func closeInstance(ctx context.Context, inst *wazeroInstance) error {
	if inst == nil {
		return nil
	}
	var moduleErr, runtimeErr error
	if inst.module != nil {
		moduleErr = inst.module.Close(ctx)
	}
	if inst.runtime != nil {
		runtimeErr = inst.runtime.Close(ctx)
	}
	return errors.Join(moduleErr, runtimeErr)
}

func truncateBytes(b []byte, n int) []byte {
	if len(b) <= n {
		return b
	}
	return b[:n]
}

func (r *WazeroRuntime) instantiate(ctx context.Context, pluginID string, wasm []byte) (*wazeroInstance, error) {
	rt := wazero.NewRuntimeWithConfig(ctx, wazero.NewRuntimeConfig().WithCloseOnContextDone(true))
	if _, err := wasi_snapshot_preview1.Instantiate(ctx, rt); err != nil {
		_ = rt.Close(ctx)
		return nil, err
	}
	if err := r.buildHostModule(ctx, rt, pluginID); err != nil {
		_ = rt.Close(ctx)
		return nil, err
	}

	inst := &wazeroInstance{}
	// GOOS=wasip1 + //go:wasmexport is a WASI reactor. wazero must run
	// _initialize before any other export or the Go runtime traps.
	mod, err := rt.InstantiateWithConfig(ctx, wasm, wazero.NewModuleConfig().
		WithName(pluginID).
		WithStdout(io.Discard).
		WithStderr(&inst.stderr).
		WithSysWalltime().
		WithSysNanotime().
		WithSysNanosleep().
		WithRandSource(rand.Reader).
		WithStartFunctions(exportInitialize))
	if err != nil {
		_ = rt.Close(ctx)
		return nil, errors.Join(ErrRuntimeTrap, err)
	}
	inst.runtime = rt
	inst.module = mod
	return inst, nil
}

func (r *WazeroRuntime) buildHostModule(ctx context.Context, rt wazero.Runtime, pluginID string) error {
	_, err := rt.NewHostModuleBuilder(hostModuleName).
		NewFunctionBuilder().
		WithFunc(func(ctx context.Context, m api.Module, inPtr, inLen, outPtr, outCap uint32) (result int32) {
			defer func() {
				if rec := recover(); rec != nil {
					result = -1
				}
			}()
			return r.hostCall(ctx, m, pluginID, inPtr, inLen, outPtr, outCap)
		}).
		Export("host_call").
		Instantiate(ctx)
	return err
}

func (r *WazeroRuntime) hostCall(
	ctx context.Context,
	m api.Module,
	pluginID string,
	inPtr, inLen, outPtr, outCap uint32,
) int32 {
	if r.host == nil {
		return -1
	}
	mem := m.Memory()
	in, ok := mem.Read(inPtr, inLen)
	if !ok {
		return -1
	}
	copied := make([]byte, len(in))
	copy(copied, in)

	var req HostRequest
	if err := DecodeJSON(copied, &req); err != nil {
		return writeGuest(mem, outPtr, outCap, mustJSON(HostError(err)))
	}
	resp, err := r.host.Call(ctx, pluginID, req)
	if err != nil {
		resp = HostError(err)
	}
	body, err := EncodeJSON(resp)
	if err != nil {
		return -1
	}
	return writeGuest(mem, outPtr, outCap, body)
}

func (r *WazeroRuntime) callGuest(ctx context.Context, inst *wazeroInstance, input []byte) ([]byte, error) {
	mod := inst.module
	inPtrFn := mod.ExportedFunction(exportInPtr)
	outPtrFn := mod.ExportedFunction(exportOutPtr)
	callFn := mod.ExportedFunction(exportCall)
	if inPtrFn == nil || outPtrFn == nil || callFn == nil {
		return nil, fmt.Errorf("plugin is missing required exports")
	}

	inPtrs, err := callExport(ctx, inPtrFn)
	if err != nil {
		return nil, err
	}
	outPtrs, err := callExport(ctx, outPtrFn)
	if err != nil {
		return nil, err
	}
	inPtr, err := uint64ToUint32(inPtrs[0])
	if err != nil {
		return nil, err
	}
	outPtr, err := uint64ToUint32(outPtrs[0])
	if err != nil {
		return nil, err
	}
	mem := mod.Memory()
	if !mem.Write(inPtr, input) {
		return nil, errors.New("failed to write plugin input")
	}
	results, err := callFn.Call(ctx, uint64(len(input)))
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, errors.New("plugin call returned no result")
	}
	outLen, err := uint64ToUint32(results[0])
	if err != nil {
		return nil, err
	}
	if outLen > maxGuestIO {
		return nil, errors.New("plugin output is too large")
	}
	out, ok := mem.Read(outPtr, outLen)
	if !ok {
		return nil, errors.New("failed to read plugin output")
	}
	copied := make([]byte, len(out))
	copy(copied, out)
	return copied, nil
}

func callExport(ctx context.Context, fn api.Function) ([]uint64, error) {
	results, err := fn.Call(ctx)
	if err != nil {
		return nil, err
	}
	return results, nil
}

func writeGuest(mem api.Memory, ptr, capacity uint32, data []byte) int32 {
	if len(data) > math.MaxInt32 || uint64(len(data)) > uint64(capacity) {
		return -1
	}
	if !mem.Write(ptr, data) {
		return -1
	}
	return int32(len(data)) //nolint:gosec // len(data) is bounded by math.MaxInt32
}

func uint64ToUint32(v uint64) (uint32, error) {
	if v > math.MaxUint32 {
		return 0, fmt.Errorf("value %d exceeds uint32", v)
	}
	return uint32(v), nil //nolint:gosec // v is bounded by math.MaxUint32
}

func mustJSON(v any) []byte {
	b, _ := EncodeJSON(v)
	return b
}

func withTimeout(ctx context.Context, d time.Duration) (context.Context, context.CancelFunc) {
	if _, ok := ctx.Deadline(); ok {
		return context.WithCancel(ctx)
	}
	return context.WithTimeout(ctx, d)
}

func requireWASIReactor(wasm []byte) error {
	if !wasmHasFunctionExport(wasm, exportInitialize) {
		return fmt.Errorf("%w: missing %s export", ErrNotReactor, exportInitialize)
	}
	return nil
}

func wasmHasFunctionExport(wasm []byte, name string) bool {
	if len(wasm) < 8 || string(wasm[:4]) != "\x00asm" {
		return false
	}
	const exportSectionID = 7
	off := 8
	for off < len(wasm) {
		id := wasm[off]
		off++
		size, next, ok := readULEB128(wasm, off)
		if !ok {
			return false
		}
		off = next
		end := off + int(size)
		if end < off || end > len(wasm) {
			return false
		}
		if id == exportSectionID {
			return exportSectionHasFunc(wasm[off:end], name)
		}
		off = end
	}
	return false
}

func exportSectionHasFunc(section []byte, name string) bool {
	count, off, ok := readULEB128(section, 0)
	if !ok {
		return false
	}
	for i := uint32(0); i < count; i++ {
		nlen, next, ok := readULEB128(section, off)
		if !ok {
			return false
		}
		off = next
		end := off + int(nlen)
		if end < off || end > len(section) {
			return false
		}
		exportName := string(section[off:end])
		off = end
		if off >= len(section) {
			return false
		}
		kind := section[off]
		off++
		_, next, ok = readULEB128(section, off)
		if !ok {
			return false
		}
		off = next
		const funcExport = 0
		if kind == funcExport && exportName == name {
			return true
		}
	}
	return false
}

func readULEB128(b []byte, off int) (uint32, int, bool) {
	var result uint32
	var shift uint
	for {
		if off >= len(b) {
			return 0, off, false
		}
		c := b[off]
		off++
		result |= uint32(c&0x7f) << shift
		if c&0x80 == 0 {
			return result, off, true
		}
		shift += 7
		if shift >= 32 {
			return 0, off, false
		}
	}
}

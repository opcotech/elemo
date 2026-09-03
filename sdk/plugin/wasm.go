//go:build wasip1

package plugin

import (
	"encoding/json"
	"unsafe"
)

var (
	inBuf   [maxIO]byte
	outBuf  [maxIO]byte
	hostIn  [maxIO]byte
	hostOut [maxIO]byte
)

var handler Handler

func setHandler(h Handler) {
	handler = h
}

func dispatch(req Request) []byte {
	if handler == nil {
		return mustJSON(Response{OK: false, Error: "no handler"})
	}
	data, err := handler(req)
	if err != nil {
		return mustJSON(Response{OK: false, Error: err.Error()})
	}
	if len(data) == 0 {
		return mustJSON(Response{OK: true})
	}
	return data
}

type hostError string

func (e hostError) Error() string { return string(e) }

func errHost(msg string) error {
	if msg == "" {
		msg = "host error"
	}
	return hostError(msg)
}

//go:wasmexport elemo_in_ptr
func inPtr() uint32 {
	return uint32(uintptr(unsafe.Pointer(&inBuf[0])))
}

//go:wasmexport elemo_out_ptr
func outPtr() uint32 {
	return uint32(uintptr(unsafe.Pointer(&outBuf[0])))
}

//go:wasmexport elemo_start
func start() int32 {
	_ = dispatch(Request{Function: "start"})
	return 0
}

//go:wasmexport elemo_stop
func stop() int32 {
	_ = dispatch(Request{Function: "stop"})
	return 0
}

//go:wasmexport elemo_call
func call(n uint32) uint32 {
	if n > maxIO {
		return writeOut(mustJSON(Response{OK: false, Error: "input too large"}))
	}
	raw := make([]byte, n)
	copy(raw, inBuf[:n])

	var req Request
	if err := json.Unmarshal(raw, &req); err != nil {
		return writeOut(mustJSON(Response{OK: false, Error: err.Error()}))
	}
	return writeOut(dispatch(req))
}

func writeOut(data []byte) uint32 {
	if len(data) > maxIO {
		data = mustJSON(Response{OK: false, Error: "output too large"})
	}
	copy(outBuf[:], data)
	return uint32(len(data))
}

//go:wasmimport elemo host_call
func hostCall(inPtr, inLen, outPtr, outCap uint32) int32

// Host invokes a versioned Elemo Plugin API method.
func Host(method, scopeID string, payload any) (json.RawMessage, error) {
	var raw json.RawMessage
	if payload != nil {
		b, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}
		raw = b
	}
	req := HostRequest{Method: method, ScopeID: scopeID, Payload: raw}
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	if len(body) > maxIO {
		return nil, errHost("host request too large")
	}
	copy(hostIn[:], body)
	n := hostCall(
		uint32(uintptr(unsafe.Pointer(&hostIn[0]))),
		uint32(len(body)),
		uint32(uintptr(unsafe.Pointer(&hostOut[0]))),
		uint32(len(hostOut)),
	)
	if n < 0 {
		return nil, errHost("host call failed")
	}
	var resp HostResponse
	if err := json.Unmarshal(hostOut[:n], &resp); err != nil {
		return nil, err
	}
	if !resp.OK {
		return nil, errHost(resp.Error)
	}
	return resp.Data, nil
}

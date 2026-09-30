package http

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/goccy/go-json"

	"github.com/opcotech/elemo/internal/entitlement"
	"github.com/opcotech/elemo/internal/pkg/convert"
	"github.com/opcotech/elemo/internal/transport/http/api"
)

var (
	notFound = api.N404JSONResponse{
		Message: "The requested resource was not found",
	}
	permissionDenied = api.N403JSONResponse{
		Message: "The requested operation is forbidden",
	}
	seatLimitReached = api.N409JSONResponse{
		Message: "Licensed seat limit reached",
		Code:    convert.ToPointer(api.HTTPErrorCodeSeatLimitReached),
	}
	activationDenied = api.N409JSONResponse{
		Message: "User activation is not entitled",
		Code:    convert.ToPointer(api.HTTPErrorCodeActivationDenied),
	}
	mutationDenied = api.N409JSONResponse{
		Message: "Installation is read-only",
		Code:    convert.ToPointer(api.HTTPErrorCodeEntitlementReadOnly),
	}
)

func entitlementConflict(err error) api.N409JSONResponse {
	switch {
	case errors.Is(err, entitlement.ErrMutationDenied):
		return mutationDenied
	case errors.Is(err, entitlement.ErrActivationDenied):
		return activationDenied
	case errors.Is(err, entitlement.ErrSeatLimitReached):
		return seatLimitReached
	default:
		return api.N409JSONResponse{Message: err.Error()}
	}
}

func formatBadRequest(err error) api.N400JSONResponse {
	return api.N400JSONResponse{
		Message: fmt.Sprintf("The provided input is invalid. %s", err.Error()),
	}
}

func setCommonHeaders(w http.ResponseWriter) {
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("X-XSS-Protection", "1; mode=block")
}

func mustWrite(w http.ResponseWriter, data []byte) {
	if _, err := w.Write(data); err != nil {
		panic(err)
	}
}

// WriteJSONResponse writes the JSON response to the response writer.
func WriteJSONResponse(w http.ResponseWriter, response any, status int) {
	w.Header().Set("Content-Type", "application/json")

	resp, err := json.Marshal(response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		mustWrite(w, []byte(errors.Join(convert.ErrMarshal, err).Error()))
	}

	w.WriteHeader(status)
	mustWrite(w, resp)
}

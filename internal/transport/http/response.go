package http

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/goccy/go-json"

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
)

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

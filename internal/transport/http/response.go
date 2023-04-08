package http

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/opcotech/elemo/internal/pkg/convert"
	"github.com/opcotech/elemo/internal/transport/http/gen"
)

var (
	notFound = gen.N404JSONResponse{
		Message: "The requested resource was not found",
	}
	permissionDenied = gen.N401JSONResponse{
		Message: "Permission denied",
	}
)

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

package http

import (
	"context"
	"net/http"

	"github.com/opcotech/elemo/internal/pkg/log"
)

// httpError is a replacement of the default http.Error function. It wraps the
// error in a HTTPErrorResponse, logs the message, and returns it in as a JSON
// response.
func httpError(ctx context.Context, w http.ResponseWriter, err error, status int) {
	if status >= 500 {
		log.Error(ctx, err)
	} else {
		log.Warn(ctx, err.Error())
	}

	w.Header().Set("X-Content-Type-Options", "nosniff")
	WriteJSONResponse(w, err, status)
}

// httpErrorStruct logs the error and returns it in as a JSON response with the
// given status code.
func httpErrorStruct(ctx context.Context, w http.ResponseWriter, err error, errStruct any, status int) {
	if status >= 500 {
		log.Error(ctx, err)
	} else {
		log.Warn(ctx, err.Error())
	}

	w.Header().Set("X-Content-Type-Options", "nosniff")
	WriteJSONResponse(w, errStruct, status)
}

package http

import (
	"context"
	"errors"
	"net/http"

	"github.com/opcotech/elemo/internal/pkg/log"
	"github.com/opcotech/elemo/internal/pkg/tracing"
	"github.com/opcotech/elemo/internal/repository"
)

const (
	ErrorCodeUnknown int = iota
	ErrorCodeEmailExists
	ErrorCodePasswordStrength
)

// ErrorResponse wraps an error message that is returned by the API.
type ErrorResponse struct {
	Code  int    `json:"code"`
	Error string `json:"error"`
}

func logError(ctx context.Context, err error, status int) {
	if status >= 500 {
		log.Error(ctx, err, log.WithTraceID(tracing.GetTraceIDFromCtx(ctx)))
	} else {
		log.Warn(ctx, err.Error(), log.WithTraceID(tracing.GetTraceIDFromCtx(ctx)))
	}
}

func getErrorCode(code int) int {
	return 1000 + code
}

// httpError is a replacement of the default http.Error function. It wraps the
// error in a ErrorResponse, logs the message, and returns it in as a JSON
// response.
func httpError(ctx context.Context, w http.ResponseWriter, err error, status int) {
	logError(ctx, err, status)

	setCommonHeaders(w)
	w.Header().Set("X-Robots-Tag", "noindex")
	WriteJSONResponse(w, ErrorResponse{Code: getErrorCode(ErrorCodeUnknown), Error: err.Error()}, status)
}

// httpErrorStruct logs the error and returns it in as a JSON response with the
// given status code.
func httpErrorStruct(ctx context.Context, w http.ResponseWriter, err error, errStruct any, status int) {
	logError(ctx, err, status)

	setCommonHeaders(w)
	w.Header().Set("X-Robots-Tag", "noindex")
	WriteJSONResponse(w, errStruct, status)
}

// isNotFoundError returns true if the error is related to a not found
// resource, regardless if it does not exist or not found in the given
// workspace.
func isNotFoundError(err error) bool {
	return errors.Is(err, repository.ErrNotFound)
}

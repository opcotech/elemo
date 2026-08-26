package http

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/opcotech/elemo/internal/license"
	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg/log"
	"github.com/opcotech/elemo/internal/pkg/validate"
	"github.com/opcotech/elemo/internal/repository"
	"github.com/opcotech/elemo/internal/service"
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
	errorCode := getErrorCode(ErrorCodeUnknown)
	if status >= 500 {
		log.Error(ctx, err,
			log.WithEventType("http.error.server"),
			log.WithErrorCode(fmt.Sprintf("%d", errorCode)),
			log.WithStatus(status),
		)
	} else {
		log.Warn(ctx, err.Error(),
			log.WithEventType("http.error.client"),
			log.WithErrorCode(fmt.Sprintf("%d", errorCode)),
			log.WithStatus(status),
		)
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

func isInvalidPageError(err error) bool {
	return errors.Is(err, repository.ErrInvalidPageSize) || errors.Is(err, repository.ErrInvalidCursor)
}

func isClientValidationError(err error) bool {
	return isInvalidPageError(err) ||
		errors.Is(err, service.ErrNoUser) ||
		errors.Is(err, service.ErrInvalidEmail) ||
		errors.Is(err, service.ErrInvalidToken) ||
		errors.Is(err, service.ErrExpiredToken) ||
		errors.Is(err, service.ErrOrganizationMemberAlreadyExists) ||
		errors.Is(err, service.ErrOrganizationMemberInvalidStatus) ||
		errors.Is(err, service.ErrIssueSelfRelation) ||
		errors.Is(err, service.ErrIssueReservedRelationKind) ||
		errors.Is(err, repository.ErrFolderNameConflict) ||
		errors.Is(err, repository.ErrFolderCycle) ||
		errors.Is(err, model.ErrInvalidID) ||
		errors.Is(err, model.ErrInvalidResourceType) ||
		errors.Is(err, model.ErrInvalidIssueDetails) ||
		errors.Is(err, model.ErrInvalidIssueRelationKind) ||
		errors.Is(err, model.ErrInvalidDocumentDetails) ||
		errors.Is(err, model.ErrInvalidFolderDetails) ||
		errors.Is(err, model.ErrInvalidProjectDetails) ||
		errors.Is(err, model.ErrInvalidNamespaceDetails) ||
		errors.Is(err, model.ErrInvalidOrganizationDetails) ||
		errors.Is(err, model.ErrInvalidRoleDetails) ||
		errors.Is(err, model.ErrInvalidTeamDetails) ||
		errors.Is(err, model.ErrInvalidTodoDetails) ||
		errors.Is(err, model.ErrInvalidUserDetails) ||
		errors.Is(err, model.ErrInvalidGrant) ||
		errors.Is(err, model.ErrInvalidAction) ||
		errors.Is(err, model.ErrNotAPrincipal) ||
		errors.Is(err, validate.ErrInvalidSlug) ||
		errors.Is(err, validate.ErrReservedSlug) ||
		errors.Is(err, validate.ErrXIDShapedSlug) ||
		errors.Is(err, validate.ErrInvalidProjectKey) ||
		errors.Is(err, validate.ErrReservedProjectKey) ||
		errors.Is(err, validate.ErrInvalidRef)
}

func isConflictError(err error) bool {
	return errors.Is(err, repository.ErrSlugConflict) ||
		errors.Is(err, repository.ErrProjectKeyConflict)
}

func isForbiddenError(err error) bool {
	return errors.Is(err, service.ErrNoPermission) ||
		errors.Is(err, license.ErrLicenseExpired) ||
		errors.Is(err, service.ErrQuotaExceeded) ||
		errors.Is(err, model.ErrPrivilegeEscalation)
}

// classifyServiceError maps service/repository errors to HTTP status codes.
// Handlers wrap the result in generated OpenAPI response types.
func classifyServiceError(err error) int {
	switch {
	case isConflictError(err):
		return http.StatusConflict
	case isClientValidationError(err):
		return http.StatusBadRequest
	case isForbiddenError(err):
		return http.StatusForbidden
	case isNotFoundError(err):
		return http.StatusNotFound
	default:
		return http.StatusInternalServerError
	}
}

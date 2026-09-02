package http

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/opcotech/elemo/internal/license"
	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg/log"
	mocklog "github.com/opcotech/elemo/internal/pkg/log/mock"
	"github.com/opcotech/elemo/internal/repository"
	"github.com/opcotech/elemo/internal/service"
	testHttp "github.com/opcotech/elemo/internal/testutil/http"
	"github.com/opcotech/elemo/internal/transport/http/api"
)

func TestHTTPError(t *testing.T) {
	type args struct {
		err    error
		status int
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "HTTP error with status 400",
			args: args{
				err:    errors.New("bad request"),
				status: http.StatusBadRequest,
			},
		},
		{
			name: "HTTP error with status 500",
			args: args{
				err:    errors.New("bad request"),
				status: http.StatusInternalServerError,
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			r, err := http.NewRequestWithContext(context.Background(), "GET", "/", nil)
			require.NoError(t, err)

			logger := mocklog.NewMockLogger(ctrl)
			if tt.args.status >= 500 {
				logger.EXPECT().Log(gomock.Any(), log.LevelError, tt.args.err.Error(), gomock.Any()).Return()
			} else {
				logger.EXPECT().Log(gomock.Any(), log.LevelWarn, tt.args.err.Error(), gomock.Any()).Return()
			}

			ctx := log.WithContext(context.Background(), logger)

			rr := testHttp.ExecuteRequest(r, func(w http.ResponseWriter, _ *http.Request) {
				httpError(ctx, w, tt.args.err, tt.args.status)
			})

			testHttp.CheckResponseCode(t, tt.args.status, rr.Code)
		})
	}
}

func TestHTTPErrorStruct(t *testing.T) {
	type args struct {
		err    error
		status int
	}
	tests := []struct {
		name string
		args args
		want api.HTTPError
	}{
		{
			name: "HTTP error with status 400",
			args: args{
				err:    errors.New("bad request"),
				status: http.StatusBadRequest,
			},
			want: api.HTTPError{
				Message: "Forbidden",
			},
		},
		{
			name: "HTTP error with status 500",
			args: args{
				err:    errors.New("internal server error"),
				status: http.StatusInternalServerError,
			},
			want: api.HTTPError{
				Message: "Server error",
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			r, err := http.NewRequestWithContext(context.Background(), "GET", "/", nil)
			require.NoError(t, err)

			logger := mocklog.NewMockLogger(ctrl)
			if tt.args.status >= 500 {
				logger.EXPECT().Log(gomock.Any(), log.LevelError, tt.args.err.Error(), gomock.Any()).Return()
			} else {
				logger.EXPECT().Log(gomock.Any(), log.LevelWarn, tt.args.err.Error(), gomock.Any()).Return()
			}

			ctx := log.WithContext(context.Background(), logger)

			rr := testHttp.ExecuteRequest(r, func(w http.ResponseWriter, _ *http.Request) {
				httpErrorStruct(ctx, w, tt.args.err, &tt.want, tt.args.status)
			})

			testHttp.CheckResponseCode(t, tt.args.status, rr.Code)
			testHttp.CheckResponseBody(t, rr.Body, &tt.want, &api.HTTPError{})
		})
	}
}

func TestClassifyServiceError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		err    error
		status int
	}{
		{name: "invalid page size", err: repository.ErrInvalidPageSize, status: http.StatusBadRequest},
		{name: "invalid cursor", err: repository.ErrInvalidCursor, status: http.StatusBadRequest},
		{name: "self relation", err: service.ErrIssueSelfRelation, status: http.StatusBadRequest},
		{name: "invalid issue details", err: model.ErrInvalidIssueDetails, status: http.StatusBadRequest},
		{name: "invalid folder details", err: model.ErrInvalidFolderDetails, status: http.StatusBadRequest},
		{name: "invalid project details", err: model.ErrInvalidProjectDetails, status: http.StatusBadRequest},
		{name: "invalid id", err: model.ErrInvalidID, status: http.StatusBadRequest},
		{name: "no user", err: service.ErrNoUser, status: http.StatusBadRequest},
		{name: "member already exists", err: service.ErrOrganizationMemberAlreadyExists, status: http.StatusBadRequest},
		{name: "member invalid status", err: service.ErrOrganizationMemberInvalidStatus, status: http.StatusBadRequest},
		{name: "folder name conflict", err: repository.ErrFolderNameConflict, status: http.StatusBadRequest},
		{name: "folder cycle", err: repository.ErrFolderCycle, status: http.StatusBadRequest},
		{name: "wrapped validation", err: errors.Join(service.ErrProjectList, model.ErrInvalidID), status: http.StatusBadRequest},
		{name: "privilege escalation", err: model.ErrPrivilegeEscalation, status: http.StatusForbidden},
		{name: "no permission", err: service.ErrNoPermission, status: http.StatusForbidden},
		{name: "invalid grant", err: model.ErrInvalidGrant, status: http.StatusBadRequest},
		{name: "wrapped permission", err: errors.Join(service.ErrProjectGet, service.ErrNoPermission), status: http.StatusForbidden},
		{name: "license expired", err: license.ErrLicenseExpired, status: http.StatusForbidden},
		{name: "quota exceeded", err: service.ErrQuotaExceeded, status: http.StatusForbidden},
		{name: "feature disabled", err: service.ErrFeatureDisabled, status: http.StatusForbidden},
		{name: "not found", err: repository.ErrNotFound, status: http.StatusNotFound},
		{name: "slug conflict", err: repository.ErrSlugConflict, status: http.StatusConflict},
		{name: "custom field key conflict", err: repository.ErrCustomFieldKeyConflict, status: http.StatusConflict},
		{name: "invalid custom field details", err: model.ErrInvalidCustomFieldDetails, status: http.StatusBadRequest},
		{name: "custom field required", err: model.ErrCustomFieldRequired, status: http.StatusBadRequest},
		{name: "wrapped slug conflict", err: errors.Join(service.ErrOrganizationCreate, repository.ErrSlugConflict), status: http.StatusConflict},
		{name: "wrapped not found", err: errors.Join(service.ErrProjectGet, repository.ErrNotFound), status: http.StatusNotFound},
		{name: "unknown", err: errors.New("boom"), status: http.StatusInternalServerError},
		{name: "wrapped unknown", err: errors.Join(service.ErrProjectGet, errors.New("boom")), status: http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.status, classifyServiceError(tt.err))
		})
	}
}

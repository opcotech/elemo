package http

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"

	"github.com/opcotech/elemo/internal/pkg/log"
	testHttp "github.com/opcotech/elemo/internal/testutil/http"
	"github.com/opcotech/elemo/internal/testutil/mock"
	"github.com/opcotech/elemo/internal/transport/http/gen"
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

			r, err := http.NewRequestWithContext(context.Background(), "GET", "/", nil)
			require.NoError(t, err)

			logger := new(mock.Logger)
			if tt.args.status >= 500 {
				logger.On("Log", zapcore.ErrorLevel, tt.args.err.Error(), []zapcore.Field{
					log.WithError(tt.args.err),
				}).Return()
			} else {
				logger.On("Log", zapcore.WarnLevel, tt.args.err.Error(), []zapcore.Field(nil)).Return()
			}

			ctx := log.WithContext(context.Background(), logger)

			rr := testHttp.ExecuteRequest(r, func(w http.ResponseWriter, r *http.Request) {
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
		want gen.HTTPError
	}{
		{
			name: "HTTP error with status 400",
			args: args{
				err:    errors.New("bad request"),
				status: http.StatusBadRequest,
			},
			want: gen.HTTPError{
				Message: "Forbidden",
			},
		},
		{
			name: "HTTP error with status 500",
			args: args{
				err:    errors.New("internal server error"),
				status: http.StatusInternalServerError,
			},
			want: gen.HTTPError{
				Message: "Server error",
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r, err := http.NewRequestWithContext(context.Background(), "GET", "/", nil)
			require.NoError(t, err)

			logger := new(mock.Logger)
			if tt.args.status >= 500 {
				logger.On("Log", zapcore.ErrorLevel, tt.args.err.Error(), []zapcore.Field{
					log.WithError(tt.args.err),
				}).Return()
			} else {
				logger.On("Log", zapcore.WarnLevel, tt.args.err.Error(), []zapcore.Field(nil)).Return()
			}

			ctx := log.WithContext(context.Background(), logger)

			rr := testHttp.ExecuteRequest(r, func(w http.ResponseWriter, r *http.Request) {
				httpErrorStruct(ctx, w, tt.args.err, &tt.want, tt.args.status)
			})

			testHttp.CheckResponseCode(t, tt.args.status, rr.Code)
			testHttp.CheckResponseBody(t, rr.Body, &tt.want, &gen.HTTPError{})
		})
	}
}

package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"

	"github.com/opcotech/elemo/internal/pkg"
	"github.com/opcotech/elemo/internal/pkg/log"
	"github.com/opcotech/elemo/internal/testutil/mock"
)

func TestWithContextObject(t *testing.T) {
	testObj := "test-value"

	request, err := http.NewRequestWithContext(context.Background(), "GET", "/", nil)
	require.NoError(t, err)

	wrappedFunc := http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		require.Equal(t, testObj, r.Context().Value(pkg.CtxKey("test")).(string))
	})

	WithContextObject("test", testObj)(wrappedFunc).ServeHTTP(httptest.NewRecorder(), request)
}

func TestWithRequestLogger(t *testing.T) {
	logger := new(mock.Logger)
	logger.On("Log", zapcore.InfoLevel, "serve http request", mock.Anything).Return()

	ctx := log.WithContext(context.Background(), logger)

	request, err := http.NewRequestWithContext(ctx, "GET", "/", nil)
	require.NoError(t, err)

	wrappedFunc := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {})
	WithRequestLogger(wrappedFunc).ServeHTTP(httptest.NewRecorder(), request)

	logger.AssertExpectations(t)
}

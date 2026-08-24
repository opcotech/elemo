package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/opcotech/elemo/internal/pkg"
	"github.com/opcotech/elemo/internal/pkg/log"
	mocklog "github.com/opcotech/elemo/internal/pkg/log/mock"
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
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := mocklog.NewMockLogger(ctrl)
	logger.EXPECT().Log(gomock.Any(), log.LevelInfo, "serve http request", gomock.Any()).Return()

	ctx := log.WithContext(context.Background(), logger)

	request, err := http.NewRequestWithContext(ctx, "GET", "/", nil)
	require.NoError(t, err)

	wrappedFunc := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {})
	WithRequestLogger(wrappedFunc).ServeHTTP(httptest.NewRecorder(), request)
}

func TestWithPrometheusMetrics_usesChiRoutePattern(t *testing.T) {
	router := chi.NewRouter()
	router.Use(WithPrometheusMetrics)
	router.Get("/v1/projects/{projectId}/issues", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/v1/projects/proj-1/issues", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	handlers := httpDurationHandlerLabels(t)
	require.Contains(t, handlers, "/v1/projects/{projectId}/issues")
	require.NotContains(t, handlers, "/v1/projects/proj-1/issues")
}

func httpDurationHandlerLabels(t *testing.T) []string {
	t.Helper()

	families, err := prometheus.DefaultGatherer.Gather()
	require.NoError(t, err)

	handlers := make([]string, 0)
	for _, family := range families {
		if family.GetName() != "http_request_duration_seconds" {
			continue
		}
		for _, metric := range family.GetMetric() {
			for _, label := range metric.GetLabel() {
				if label.GetName() == "handler" {
					handlers = append(handlers, label.GetValue())
				}
			}
		}
	}
	return handlers
}

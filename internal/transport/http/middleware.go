package http

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"runtime"
	"strings"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-oauth2/oauth2/v4"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	httpMetricsProm "github.com/slok/go-http-metrics/metrics/prometheus"
	httpMetricsMiddleware "github.com/slok/go-http-metrics/middleware"
	httpMetricsMiddlewareStd "github.com/slok/go-http-metrics/middleware/std"

	"github.com/opcotech/elemo/internal/pkg/log"
	"github.com/opcotech/elemo/internal/pkg/tracing"
)

const (
	ctxKeyUserID ctxKey = "userID"
)

// ctxKey is the type alias for the context key.
type ctxKey string

type ctxCallbackFunc func(w http.ResponseWriter, r *http.Request) any

func getMiddlewareName(fn func(next http.Handler) http.Handler) (string, string) {
	cache := make(map[uintptr][]string)

	fnPtr := reflect.ValueOf(fn).Pointer()

	if res, ok := cache[fnPtr]; ok {
		return res[0], res[1]
	}

	path := runtime.FuncForPC(fnPtr).Name()
	parts := strings.Split(path, ".")
	name := parts[len(parts)-1]
	cache[fnPtr] = append(cache[fnPtr], name, path)

	return name, path
}

func withContextObject(ctxKey ctxKey, cb ctxCallbackFunc) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r = r.WithContext(context.WithValue(r.Context(), ctxKey, cb(w, r)))
			next.ServeHTTP(w, r)
		})
	}
}

// WithContextObject returns a middleware that adds any value to the context
// associated with the given key.
func WithContextObject(key ctxKey, value any) func(next http.Handler) http.Handler {
	return withContextObject(key, func(w http.ResponseWriter, r *http.Request) any {
		return value
	})
}

func WithOtelTracer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		otelhttp.NewHandler(next, r.URL.Path).ServeHTTP(w, r)
	})
}

func WithPrometheusMetrics(next http.Handler) http.Handler {
	return httpMetricsMiddlewareStd.Handler("", httpMetricsMiddleware.New(httpMetricsMiddleware.Config{
		Service:  "elemo",
		Recorder: httpMetricsProm.NewRecorder(httpMetricsProm.Config{}),
	}), next)
}

// WithTracedMiddleware returns an HTTP middleware that traces the middleware
// execution by creating a new span and passing the context to the next
// handler.
func WithTracedMiddleware(tracer trace.Tracer, middleware func(next http.Handler) http.Handler) func(next http.Handler) http.Handler {
	if tracer == nil {
		tracer = tracing.NoopTracer()
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			name, path := getMiddlewareName(middleware)
			ctx, span := tracer.Start(r.Context(), fmt.Sprintf("transport.http.middleware/%s", name))
			defer span.End()

			span.SetAttributes(attribute.KeyValue{
				Key:   "middleware.path",
				Value: attribute.StringValue(path),
			})

			middleware(next).ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// WithRequestLogger returns a middleware that logs the request.
//
// The middleware depends on WithLogger. To use this middleware, you must call
// both of those first.
func WithRequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wrappedWriter := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

		currentTime := time.Now()
		defer func(ctx context.Context, w middleware.WrapResponseWriter, r *http.Request, t time.Time) {
			log.Info(ctx, "serve http request",
				log.WithProtocol(r.Proto),
				log.WithMethod(r.Method),
				log.WithPath(r.URL.Path),
				log.WithRequestID(middleware.GetReqID(ctx)),
				log.WithRemoteAddr(r.RemoteAddr),
				log.WithUserAgent(r.UserAgent()),
				log.WithSize(int64(w.BytesWritten())),
				log.WithStatus(w.Status()),
				log.WithDuration(time.Since(t).Seconds()),
				log.WithAction(log.ActionHTTPRequestHandle),
			)
		}(r.Context(), wrappedWriter, r, currentTime)

		next.ServeHTTP(wrappedWriter, r)
	})
}

// WithUserKey returns a middleware that adds the user Key to the context,
// parsed from the Authorization header if present. Otherwise, an empty string
// is added.
func WithUserKey(tokenValidator func(r *http.Request) (oauth2.TokenInfo, error)) func(next http.Handler) http.Handler {
	return withContextObject(ctxKeyUserID, func(w http.ResponseWriter, r *http.Request) any {
		if info, _ := tokenValidator(r); info != nil {
			return info.GetUserID()
		}

		return ""
	})
}

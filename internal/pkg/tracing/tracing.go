package tracing

import (
	"context"
	"errors"
	"fmt"

	"go.opentelemetry.io/otel"
	otlptrace "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/embedded"
	nooptrace "go.opentelemetry.io/otel/trace/noop"

	"github.com/opcotech/elemo/internal/config"
	"github.com/opcotech/elemo/internal/model"
)

var (
	ErrNoTracer        = errors.New("no tracer") // the tracer is missing
	noopTracerProvider = nooptrace.NewTracerProvider()
	noopTracer         = noopTracerProvider.Tracer("github.com/opcotech/elemo")
)

// Tracer re-defines the tracing.Tracer interface as the
type Tracer interface {
	embedded.Tracer
	Start(ctx context.Context, spanName string, opts ...trace.SpanStartOption) (context.Context, trace.Span)
}

// NewTracerProvider creates a new tracer provider.
func NewTracerProvider(ctx context.Context, version *model.VersionInfo, service string, cfg *config.TracingConfig) (trace.TracerProvider, error) {
	exporter, err := otlptrace.New(
		ctx,
		otlptrace.WithEndpoint(cfg.CollectorEndpoint),
		otlptrace.WithInsecure(),
	)
	if err != nil {
		return nil, err
	}

	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(fmt.Sprintf("%s-%s", cfg.ServiceName, service)),
			semconv.ServiceVersionKey.String(version.Version),
		)),
		sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.TraceIDRatioBased(cfg.TraceRatio))),
		sdktrace.WithBatcher(exporter),
	)
	otel.SetTracerProvider(tracerProvider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	return tracerProvider, err
}

// NoopTracer returns a noop tracer.
func NoopTracer() Tracer {
	return noopTracer
}

// GetTraceID extracts the trace ID from the span context.
func GetTraceID(span trace.Span) string {
	if span == nil {
		return ""
	}
	return span.SpanContext().TraceID().String()
}

// GetTraceIDFromCtx extracts the trace ID from the context.
func GetTraceIDFromCtx(ctx context.Context) string {
	return GetTraceID(trace.SpanFromContext(ctx))
}

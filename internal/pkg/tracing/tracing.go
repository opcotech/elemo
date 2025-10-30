package tracing

import (
	"context"
	"errors"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	otlptrace "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
	nooptrace "go.opentelemetry.io/otel/trace/noop"

	"github.com/opcotech/elemo/internal/config"
	"github.com/opcotech/elemo/internal/model"
)

var (
	ErrNoTracer        = errors.New("no tracer") // the tracer is missing
	noopTracerProvider = nooptrace.NewTracerProvider()
	noopTracer         = noopTracerProvider.Tracer("github.com/opcotech/elemo")
)

// Span re-defines the trace.Span interface to avoid embedding issues
//
//go:generate mockgen -destination ../../testutil/mock/span_gen.go -package mock -mock_names Span=MockSpan github.com/opcotech/elemo/internal/pkg/tracing Span
type Span interface {
	End(options ...trace.SpanEndOption)
	AddEvent(name string, options ...trace.EventOption)
	AddLink(link trace.Link)
	IsRecording() bool
	RecordError(err error, options ...trace.EventOption)
	SpanContext() trace.SpanContext
	SetStatus(code codes.Code, description string)
	SetName(name string)
	SetAttributes(kv ...attribute.KeyValue)
	TracerProvider() trace.TracerProvider
}

// Tracer re-defines the tracing.Tracer interface
//
//go:generate mockgen -destination ../../testutil/mock/tracer_gen.go -package mock -mock_names Tracer=MockTracer github.com/opcotech/elemo/internal/pkg/tracing Tracer
type Tracer interface {
	Start(ctx context.Context, spanName string, opts ...trace.SpanStartOption) (context.Context, Span)
}

// WrapTracer wraps an OpenTelemetry tracer to implement our custom Tracer interface
func WrapTracer(otelTracer trace.Tracer) Tracer {
	return &tracerWrapper{tracer: otelTracer}
}

// tracerWrapper wraps the OpenTelemetry tracer to implement our custom Tracer interface
type tracerWrapper struct {
	tracer trace.Tracer
}

func (w *tracerWrapper) Start(ctx context.Context, spanName string, opts ...trace.SpanStartOption) (context.Context, Span) {
	ctx, span := w.tracer.Start(ctx, spanName, opts...)
	return ctx, &spanWrapper{span: span}
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
	return &noopTracerWrapper{tracer: noopTracer}
}

// noopTracerWrapper wraps the OpenTelemetry noop tracer to implement our custom Tracer interface
type noopTracerWrapper struct {
	tracer trace.Tracer
}

func (w *noopTracerWrapper) Start(ctx context.Context, spanName string, opts ...trace.SpanStartOption) (context.Context, Span) {
	ctx, span := w.tracer.Start(ctx, spanName, opts...)
	return ctx, &spanWrapper{span: span}
}

// spanWrapper wraps the OpenTelemetry span to implement our custom Span interface
type spanWrapper struct {
	span trace.Span
}

func (w *spanWrapper) End(options ...trace.SpanEndOption) {
	w.span.End(options...)
}

func (w *spanWrapper) AddEvent(name string, options ...trace.EventOption) {
	w.span.AddEvent(name, options...)
}

func (w *spanWrapper) AddLink(link trace.Link) {
	w.span.AddLink(link)
}

func (w *spanWrapper) IsRecording() bool {
	return w.span.IsRecording()
}

func (w *spanWrapper) RecordError(err error, options ...trace.EventOption) {
	w.span.RecordError(err, options...)
}

func (w *spanWrapper) SpanContext() trace.SpanContext {
	return w.span.SpanContext()
}

func (w *spanWrapper) SetStatus(code codes.Code, description string) {
	w.span.SetStatus(code, description)
}

func (w *spanWrapper) SetName(name string) {
	w.span.SetName(name)
}

func (w *spanWrapper) SetAttributes(kv ...attribute.KeyValue) {
	w.span.SetAttributes(kv...)
}

func (w *spanWrapper) TracerProvider() trace.TracerProvider {
	return w.span.TracerProvider()
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

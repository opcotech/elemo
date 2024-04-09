package mock

import (
	"context"

	"github.com/stretchr/testify/mock"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/embedded"

	"github.com/opcotech/elemo/internal/pkg/tracing"
)

type TracerProvider struct {
	mock.Mock
}

func (m *TracerProvider) Tracer(name string, options ...trace.TracerOption) tracing.Tracer {
	args := m.Called(name, options)
	return args.Get(0).(tracing.Tracer)
}

type Tracer struct {
	mock.Mock
	embedded.Tracer
}

func (m *Tracer) Start(ctx context.Context, spanName string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	args := m.Called(ctx, spanName, opts)
	return args.Get(0).(context.Context), args.Get(1).(trace.Span)
}

type Span struct {
	mock.Mock
	embedded.Span
}

func (m *Span) AddLink(link trace.Link) {
	m.Called(link)
}

func (m *Span) End(options ...trace.SpanEndOption) {
	m.Called(options)
}

func (m *Span) AddEvent(name string, options ...trace.EventOption) {
	m.Called(name, options)
}

func (m *Span) IsRecording() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *Span) RecordError(err error, options ...trace.EventOption) {
	m.Called(err, options)
}

func (m *Span) SpanContext() trace.SpanContext {
	args := m.Called()
	return args.Get(0).(trace.SpanContext)
}

func (m *Span) SetStatus(code codes.Code, description string) {
	m.Called(code, description)
}

func (m *Span) SetName(name string) {
	m.Called(name)
}

func (m *Span) SetAttributes(kv ...attribute.KeyValue) {
	m.Called(kv)
}

func (m *Span) TracerProvider() trace.TracerProvider {
	args := m.Called()
	return args.Get(0).(trace.TracerProvider)
}

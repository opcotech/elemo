package service

import (
	"context"

	"github.com/opcotech/elemo/internal/entitlement"
	"github.com/opcotech/elemo/internal/pkg/event"
	"github.com/opcotech/elemo/internal/pkg/log"
	"github.com/opcotech/elemo/internal/pkg/tracing"
)

type mutationAuthorizedKey struct{}

// Option defines a configuration option for the service runtime.
type Option func(*runtime) error

// WithLogger sets the logger for the service runtime.
func WithLogger(logger log.Logger) Option {
	return func(r *runtime) error {
		if logger == nil {
			return log.ErrNoLogger
		}

		r.logger = logger
		return nil
	}
}

// WithTracer sets the tracer for the service runtime.
func WithTracer(tracer tracing.Tracer) Option {
	return func(r *runtime) error {
		if tracer == nil {
			return tracing.ErrNoTracer
		}

		r.tracer = tracer
		return nil
	}
}

type runtime struct {
	logger    log.Logger
	tracer    tracing.Tracer
	eventBus  EventPublisher
	mutations entitlement.MutationPolicy
}

// EventPublisher is the subset of the in-process bus used by domain services.
type EventPublisher interface {
	Publish(ctx context.Context, event event.Event) error
}

// WithEventBus sets the in-process event publisher. A nil bus is a no-op.
func WithEventBus(bus EventPublisher) Option {
	return func(r *runtime) error {
		r.eventBus = bus
		return nil
	}
}

// WithMutationPolicy sets the entitlement write gate. A nil policy is rejected.
func WithMutationPolicy(policy entitlement.MutationPolicy) Option {
	return func(r *runtime) error {
		if policy == nil {
			return entitlement.ErrNoMutationPolicy
		}
		r.mutations = policy
		return nil
	}
}

func newRuntime(opts ...Option) (runtime, error) {
	r := runtime{
		logger:    log.DefaultLogger(),
		tracer:    tracing.NoopTracer(),
		mutations: entitlement.Unrestricted(),
	}

	for _, opt := range opts {
		if err := opt(&r); err != nil {
			return runtime{}, err
		}
	}

	return r, nil
}

// requireMutation authorizes a user-initiated write for this operation. Nested
// service calls reuse the outer decision from ctx instead of re-evaluating the
// clock-based policy after an earlier write.
func (r runtime) requireMutation(ctx context.Context) (context.Context, error) {
	if ctx.Value(mutationAuthorizedKey{}) != nil {
		return ctx, nil
	}
	if r.mutations == nil {
		return ctx, entitlement.ErrNoMutationPolicy
	}
	if err := r.mutations.AllowsMutation(ctx); err != nil {
		return ctx, err
	}
	return context.WithValue(ctx, mutationAuthorizedKey{}, true), nil
}

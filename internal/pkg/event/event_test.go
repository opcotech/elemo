package event_test

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/opcotech/elemo/internal/model"
	"github.com/opcotech/elemo/internal/pkg/event"
)

func TestBus_PublishSubscribe(t *testing.T) {
	t.Parallel()

	bus := event.NewBus()
	var got event.Event
	sub, err := bus.Subscribe(model.PluginEventIssueCreated, func(_ context.Context, evt event.Event) error {
		got = evt
		return nil
	})
	require.NoError(t, err)
	require.NotNil(t, sub)

	resource := model.MustNewID(model.ResourceTypeIssue)
	require.NoError(t, bus.Publish(context.Background(), event.Event{
		Type:     model.PluginEventIssueCreated,
		Resource: resource,
	}))
	assert.Equal(t, resource, got.Resource)
	assert.False(t, got.EmittedAt.IsZero())

	sub.Unsubscribe()
	got = event.Event{}
	require.NoError(t, bus.Publish(context.Background(), event.Event{
		Type:     model.PluginEventIssueCreated,
		Resource: resource,
	}))
	assert.True(t, got.Resource.IsNil())
}

func TestBus_SubscribeValidation(t *testing.T) {
	t.Parallel()

	bus := event.NewBus()
	_, err := bus.Subscribe(model.PluginEventType("nope"), func(context.Context, event.Event) error { return nil })
	require.ErrorIs(t, err, event.ErrUnknownTopic)

	_, err = bus.Subscribe(model.PluginEventIssueUpdated, nil)
	require.ErrorIs(t, err, event.ErrNoHandler)
}

func TestBus_HandlerErrorDoesNotStopOthers(t *testing.T) {
	t.Parallel()

	bus := event.NewBus()
	sentinel := errors.New("handler failed")
	var second atomic.Bool

	_, err := bus.Subscribe(model.PluginEventProjectCreated, func(context.Context, event.Event) error {
		return sentinel
	})
	require.NoError(t, err)
	_, err = bus.Subscribe(model.PluginEventProjectCreated, func(context.Context, event.Event) error {
		second.Store(true)
		return nil
	})
	require.NoError(t, err)

	err = bus.Publish(context.Background(), event.Event{Type: model.PluginEventProjectCreated})
	require.ErrorIs(t, err, sentinel)
	assert.True(t, second.Load())
}

func TestBus_ConcurrentPublish(t *testing.T) {
	t.Parallel()

	bus := event.NewBus()
	var n atomic.Int64
	_, err := bus.Subscribe(model.PluginEventIssueUpdated, func(context.Context, event.Event) error {
		n.Add(1)
		return nil
	})
	require.NoError(t, err)

	var wg sync.WaitGroup
	for range 32 {
		wg.Go(func() {
			_ = bus.Publish(context.Background(), event.Event{Type: model.PluginEventIssueUpdated})
		})
	}
	wg.Wait()
	assert.Equal(t, int64(32), n.Load())
}

func TestBus_NilPublish(t *testing.T) {
	t.Parallel()

	var bus *event.Bus
	require.NoError(t, bus.Publish(context.Background(), event.Event{Type: model.PluginEventIssueDeleted}))
}

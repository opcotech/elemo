package event

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/opcotech/elemo/internal/model"
)

var (
	ErrNoHandler    = errors.New("no event handler")
	ErrUnknownTopic = errors.New("unknown event topic")
)

// Event is a domain event published to in-process subscribers.
type Event struct {
	Type      model.PluginEventType `json:"type"`
	Resource  model.ID              `json:"resource"`
	Actor     model.ID              `json:"actor,omitempty"`
	Payload   map[string]any        `json:"payload,omitempty"`
	EmittedAt time.Time             `json:"emitted_at"`
}

// Handler consumes a single event. Failure is the caller's to log.
type Handler func(ctx context.Context, event Event) error

// Subscription is a disposable handler registration.
type Subscription struct {
	id    uint64
	topic model.PluginEventType
	bus   *Bus
}

// Unsubscribe removes the handler. It is idempotent.
func (s *Subscription) Unsubscribe() {
	if s == nil || s.bus == nil {
		return
	}
	s.bus.unsubscribe(s.topic, s.id)
	s.bus = nil
}

// Bus is an in-process pub/sub. It is concurrency-safe. It does not persist,
// retry, or fan out over the network.
type Bus struct {
	mu       sync.RWMutex
	nextID   uint64
	handlers map[model.PluginEventType]map[uint64]Handler
}

// NewBus returns an empty event bus.
func NewBus() *Bus {
	return &Bus{
		handlers: make(map[model.PluginEventType]map[uint64]Handler),
	}
}

// Subscribe registers handler for topic. The handler is invoked synchronously
// on Publish; the publisher does not hold the bus lock while calling it.
func (b *Bus) Subscribe(topic model.PluginEventType, handler Handler) (*Subscription, error) {
	if b == nil {
		return nil, ErrNoHandler
	}
	if !topic.Valid() {
		return nil, ErrUnknownTopic
	}
	if handler == nil {
		return nil, ErrNoHandler
	}

	b.mu.Lock()
	defer b.mu.Unlock()
	b.nextID++
	id := b.nextID
	if b.handlers[topic] == nil {
		b.handlers[topic] = make(map[uint64]Handler)
	}
	b.handlers[topic][id] = handler
	return &Subscription{id: id, topic: topic, bus: b}, nil
}

func (b *Bus) unsubscribe(topic model.PluginEventType, id uint64) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if handlers, ok := b.handlers[topic]; ok {
		delete(handlers, id)
		if len(handlers) == 0 {
			delete(b.handlers, topic)
		}
	}
}

// Publish delivers event to current subscribers. Handlers run sequentially.
// A handler error does not stop remaining handlers; errors are joined.
func (b *Bus) Publish(ctx context.Context, event Event) error {
	if b == nil {
		return nil
	}
	if event.EmittedAt.IsZero() {
		event.EmittedAt = time.Now().UTC()
	}

	b.mu.RLock()
	handlers := b.handlers[event.Type]
	snapshot := make([]Handler, 0, len(handlers))
	for _, h := range handlers {
		snapshot = append(snapshot, h)
	}
	b.mu.RUnlock()

	var joined error
	for _, handler := range snapshot {
		if err := handler(ctx, event); err != nil {
			joined = errors.Join(joined, err)
		}
	}
	return joined
}

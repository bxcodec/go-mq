package mq

import (
	"context"
	"time"
)

type contextKey string

const (
	contextKeyMaxOutstanding = contextKey("maxOutstanding")
)

// WithMaxOutstanding decorate ctx with maxOutstanding value.
func WithMaxOutstanding(ctx context.Context, v int) context.Context {
	return context.WithValue(ctx, contextKeyMaxOutstanding, v)
}

// MaxOutstanding value from ctx.
func MaxOutstanding(ctx context.Context) (int, bool) {
	v, ok := ctx.Value(contextKeyMaxOutstanding).(int)
	return v, ok
}

// Subscriber is the interface that wraps the basic Subscribe method.
//
// Subscribe subscribes to a channel for receive messages. The message will be
// passed to h Handler. Message handling need to be Ack or Nack.
// Leave the Nack might block the receiving until timeout
// (the behaviour will differ based on implementation).
type Subscriber interface {
	Subscribe(ctx context.Context, channel string, h Handler) error
}

// Publisher is the interface that wraps the basic Publish method.
//
// Publish publishes the msg to specific topic.
type Publisher interface {
	Publish(topic string, msg []byte) PublishResult
}

// PublishResult is results of the publish.
type PublishResult interface {
	Get(context.Context) (id string, err error)
	Ready() <-chan struct{}
}

// Message represent the message from server.
//
// Method Ack or Nack should be invoked.
type Message interface {
	ID() string
	Body() []byte
	Timestamp() time.Time
	Ack() error
	Nack() error
}

// Handler is the message handler.
type Handler interface {
	Handle(Message)
}

// HandlerFunc is the function adapter of Handler.
type HandlerFunc func(Message)

// Handle invoke f(msg).
func (f HandlerFunc) Handle(m Message) {
	f(m)
}

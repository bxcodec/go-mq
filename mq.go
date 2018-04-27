package mq

import (
	"context"
	"time"
)

type contextKey string

const (
	contextKeyMaxOutstanding = contextKey("maxOutstanding")
	contextKeyNumWorkers     = contextKey("numWorkers")
)

// ContextWithMaxOutstanding decorate ctx with maxOutstanding value.
func ContextWithMaxOutstanding(ctx context.Context, v int) context.Context {
	return context.WithValue(ctx, contextKeyMaxOutstanding, v)
}

// MaxOutstandingFromContext get value of maxOutstanding from context.
func MaxOutstandingFromContext(ctx context.Context) (int, bool) {
	v, ok := ctx.Value(contextKeyMaxOutstanding).(int)
	return v, ok
}

// ContextWithNumWorkers decorate ctx with numWorkers value.
func ContextWithNumWorkers(ctx context.Context, v int) context.Context {
	return context.WithValue(ctx, contextKeyNumWorkers, v)
}

// NumWorkersFromContext get value of numWorkers from context.
func NumWorkersFromContext(ctx context.Context) (int, bool) {
	v, ok := ctx.Value(contextKeyNumWorkers).(int)
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

// AckHandler handle message and ack automatically.
type AckHandler interface {
	// HandleAck handle message that automatically ack when nil error returned.
	HandleAck(Message) error
}

// AckErrNotifier provides method NotifyAckErr, invokes when Ack failed.
type AckErrNotifier interface {
	NotifyAckErr(Message, error)
}

// NackErrNotifier provides method NotifyNackErr, invokes when Nack failed.
type NackErrNotifier interface {
	NotifyNackErr(Message, error)
}

// AutoAck creates Handler from AckHandler.
//
// AutoAck will handle the message, when err returned by AckHandler then it will m.Nack() otherwise m.Ack().
// To know the successful of Ack and Nack, the h should also implement the AckErrNotifier and NackErrNotifier. Implementing these interfaces are optional.
func AutoAck(h AckHandler) Handler {
	return HandlerFunc(func(m Message) {
		if err := h.HandleAck(m); err != nil {
			if err := m.Nack(); err != nil {
				notifyNackErr(h, m, err)
			}
		}

		if err := m.Ack(); err != nil {
			notifyAckErr(h, m, err)
		}
	})
}

func notifyAckErr(v interface{}, m Message, err error) {
	if n, ok := v.(AckErrNotifier); ok {
		n.NotifyAckErr(m, err)
	}
}

func notifyNackErr(v interface{}, m Message, err error) {
	if n, ok := v.(NackErrNotifier); ok {
		n.NotifyNackErr(m, err)
	}
}

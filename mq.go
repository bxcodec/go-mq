package mq

import (
	"context"
	"time"
)

// Subscriber is the interface that wraps the basic Subscribe method.
//
// Subscribe subscribes to a channel for receive messages. The message will be
// passed to h Handler. Message handling need to be Ack or Nack.
// Leave the Nack might block the receiving until timeout
// (the behaviour will be differ based on implementation).
type Subscriber interface {
	Subscribe(ctx context.Context, channel string, h Handler) error
}

// Publisher is the interface that wraps the basic Publish method.
//
// Publish publishes the msg to specific topic.
type Publisher interface {
	Publish(topic string, msg []byte) error
}

// Closer is the interface that wraps the basic Close method.
type Closer interface {
	Close() error
}

// SubscribePublisher is the interface that groups the basic Subscribe
// and Publish methods.
type SubscribePublisher interface {
	Subscriber
	Publisher
}

// SubscribePublishCloser is the interface that groups the basic Subscribe, Publish and Close methods.
type SubscribePublishCloser interface {
	Subscriber
	Publisher
	Closer
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

// Receiver is the interface that wraps the Receive method.
//
// It accept ctx for cancellation.
// Receive method will blocks until it canceled or error returned.
type Receiver interface {
	Receive(ctx context.Context, handler func(Message)) error
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

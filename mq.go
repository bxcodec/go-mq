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
// (the behaviour will differ based on implementation).
type Subscriber interface {
	Subscribe(ctx context.Context, channel string, h Handler) error
}

// Publisher is the interface that wraps the basic Publish method.
//
// Publish publishes the msg to specific topic.
type Publisher interface {
	Publish(topic string, msg []byte) error
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

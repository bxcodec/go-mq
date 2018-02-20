package pubsub

import (
	"time"

	"cloud.google.com/go/pubsub"
)

type message struct {
	v *pubsub.Message
}

func newMessage(m *pubsub.Message) *message {
	return &message{v: m}
}

func (m *message) ID() string {
	return m.v.ID
}

func (m *message) Body() []byte {
	return m.v.Data
}

func (m *message) Timestamp() time.Time {
	return m.v.PublishTime
}

func (m *message) Ack() error {
	m.v.Ack()
	return nil
}

func (m *message) Nack() error {
	m.v.Nack()
	return nil
}

package pubsub

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"

	"cloud.google.com/go/pubsub"
	mq "github.com/KurioApp/go-mq"
	"google.golang.org/api/option"
)

const (
	stateReady int32 = iota
	stateClosing
	stateClosed
)

// PubSub represents the PubSub (client).
type PubSub struct {
	client *pubsub.Client
	state  int32
	wg     sync.WaitGroup

	mu     sync.Mutex
	topics map[string]*pubsub.Topic
}

// New constructs new PubSub.
func New(ctx context.Context, projectID string, opts ...option.ClientOption) (*PubSub, error) {
	client, err := pubsub.NewClient(ctx, projectID, opts...)
	if err != nil {
		return nil, err
	}

	return &PubSub{
		client: client,
		topics: make(map[string]*pubsub.Topic),
	}, nil
}

// Subscribe to specific channel.
func (p *PubSub) Subscribe(ctx context.Context, channel string, h mq.Handler) error {
	if atomic.LoadInt32(&p.state) != stateReady {
		return errors.New("pubsub: not in ready state")
	}

	subs := p.client.Subscription(channel)

	p.wg.Add(1)
	defer p.wg.Done()
	return subs.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
		select {
		case <-ctx.Done():
			return
		default:
		}

		m := newMessage(msg)
		h.Handle(m)
	})
}

// Publish message to specific topic.
func (p *PubSub) Publish(topicID string, msg []byte) error {
	if atomic.LoadInt32(&p.state) != stateReady {
		return errors.New("pusbub: not in ready state")
	}

	topic := p.topicOf(topicID)

	p.wg.Add(1)
	defer p.wg.Done()
	res := topic.Publish(context.Background(), &pubsub.Message{Data: msg})
	_, err := res.Get(context.Background())
	return err
}

func (p *PubSub) topicOf(id string) *pubsub.Topic {
	p.mu.Lock()
	defer p.mu.Unlock()
	t, found := p.topics[id]
	if !found {
		t = p.client.Topic(id)
		p.topics[id] = t
	}
	return t
}

func (p *PubSub) stopAllTopics() {
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, v := range p.topics {
		v.Stop()
	}
}

// EnsureSubscription ensures the subscription exists.
func (p *PubSub) EnsureSubscription(ctx context.Context, topic *pubsub.Topic, id string) (*pubsub.Subscription, error) {
	if atomic.LoadInt32(&p.state) != stateReady {
		return nil, errors.New("pusbub: not in ready state")
	}

	subs := p.client.Subscription(id)
	exists, err := subs.Exists(ctx)
	if err != nil {
		return nil, err
	}

	if !exists {
		return p.client.CreateSubscription(ctx, id, pubsub.SubscriptionConfig{Topic: topic})
	}

	return subs, nil
}

// EnsureTopic ensures the topic exists.
func (p *PubSub) EnsureTopic(ctx context.Context, id string) (*pubsub.Topic, error) {
	if atomic.LoadInt32(&p.state) != stateReady {
		return nil, errors.New("pusbub: not in ready state")
	}

	topic := p.client.Topic(id)
	exists, err := topic.Exists(ctx)
	if err != nil {
		return nil, err
	}

	if !exists {
		return p.client.CreateTopic(ctx, id)
	}

	return topic, err
}

// Close the connection with server.
func (p *PubSub) Close() error {
	if !atomic.CompareAndSwapInt32(&p.state, stateReady, stateClosing) {
		return errors.New("pubsub: not in ready state")
	}
	defer atomic.StoreInt32(&p.state, stateClosed)

	p.stopAllTopics()
	p.wg.Wait()
	return p.client.Close()
}
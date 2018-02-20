package pubsub_test

import (
	"context"
	"flag"
	"log"
	"testing"
	"time"

	"github.com/KurioApp/go-mq/pubsub"

	mq "github.com/KurioApp/go-mq"
	uuid "github.com/satori/go.uuid"
	"google.golang.org/api/option"
	tilde "gopkg.in/mattes/go-expand-tilde.v1"

	gpubsub "cloud.google.com/go/pubsub"
)

var (
	flagProjectID       = flag.String("gcp.project-id", "", "Google Cloud Project ID")
	flagTopicID         = flag.String("gcp.topic-id", "", "Google Cloud Topic ID")
	flagSubscriptionID  = flag.String("gcp.subscription-id", "", "Google Cloud Subscription ID")
	flagCredentialsFile = flag.String("gcp.credentials-file", "", "Google Cloud Credentials File")
	flagConnectTimeout  = flag.Duration("gcp.connect-timeout", 5*time.Second, "Google Cloud connect timeout")
)

func TestSubscribe(t *testing.T) {
	if testing.Short() {
		t.Skip("Require non-short mode")
	}

	fix := setup(t)
	defer fix.tearDown()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	msgc := make(chan string)
	go func() {
		err := fix.pubSub.Subscribe(ctx, *flagSubscriptionID, mq.HandlerFunc(func(msg mq.Message) {
			msgc <- string(msg.Body())
			if err := msg.Ack(); err != nil {
				panic(err)
			}
		}))

		if err != nil {
			panic(err)
		}
	}()

	// Submit job
	text := uuid.NewV4().String()
	res := fix.topic.Publish(context.Background(), &gpubsub.Message{Data: []byte(text)})
	if _, err := res.Get(context.Background()); err != nil {
		t.Fatal(err)
	}

	log.Println("Message published")

	// Wait for the job and assert
	if err := waitForMessage(msgc, text, 10*time.Second); err != nil {
		t.Fatal(err)
	}

	log.Println("Match")
}

func TestPublish(t *testing.T) {
	if testing.Short() {
		t.Skip("Require non-short mode")
	}

	fix := setup(t)
	defer fix.tearDown()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	msgc := make(chan string)
	go func() {
		err := fix.subs.Receive(ctx, func(ctx context.Context, msg *gpubsub.Message) {
			msgc <- string(msg.Data)
			msg.Ack()
		})

		if err != nil {
			panic(err)
		}
	}()

	// Submit job
	text := uuid.NewV4().String()
	if err := fix.pubSub.Publish(*flagTopicID, []byte(text)); err != nil {
		t.Fatal(err)
	}

	log.Println("Message published")

	// Wait for the job and assert
	if err := waitForMessage(msgc, text, 10*time.Second); err != nil {
		t.Fatal(err)
	}

	log.Println("Match")
}

func waitForMessage(msgc <-chan string, expectMsg string, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	for {
		select {
		case msg := <-msgc:
			if msg == expectMsg {
				return nil
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

type fixture struct {
	t      *testing.T
	pubSub *pubsub.PubSub
	topic  *gpubsub.Topic
	subs   *gpubsub.Subscription
}

func (f *fixture) tearDown() {
	if err := deleteSubscription(f.subs); err != nil {
		f.t.Error("Fail to delete subscription:", err)
	}

	log.Println("Subscription removed")

	f.topic.Stop()
	if err := deleteTopic(f.topic); err != nil {
		f.t.Error("Fail to delete topic:", err)
	}

	log.Println("Topic removed")

	if err := f.pubSub.Close(); err != nil {
		f.t.Error("Fail to close pubSub:", err)
	}

	log.Println("PubSub closed")
}

func setup(t *testing.T) *fixture {
	credsFile, err := tilde.Expand(*flagCredentialsFile)
	if err != nil {
		t.Fatal(err)
	}

	connect := func() (*pubsub.PubSub, error) {
		ctx, cancel := context.WithTimeout(context.Background(), *flagConnectTimeout)
		defer cancel()

		return pubsub.New(ctx, *flagProjectID, option.WithCredentialsFile(credsFile))
	}

	pubSub, err := connect()
	if err != nil {
		t.Fatal(err)
	}

	log.Println("Connected")

	topic, err := ensureTopic(pubSub, *flagTopicID)
	if err != nil {
		t.Fatal(err)
	}

	log.Println("Topic ready")

	subs, err := ensureSubscription(pubSub, topic, *flagSubscriptionID)
	if err != nil {
		t.Fatal(err)
	}

	log.Println("Subscription ready")

	return &fixture{
		t:      t,
		pubSub: pubSub,
		topic:  topic,
		subs:   subs,
	}
}

func ensureTopic(p *pubsub.PubSub, topicID string) (*gpubsub.Topic, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return p.EnsureTopic(ctx, topicID)
}

func ensureSubscription(p *pubsub.PubSub, topic *gpubsub.Topic, subscriptionID string) (*gpubsub.Subscription, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return p.EnsureSubscription(ctx, topic, subscriptionID)
}

func deleteTopic(topic *gpubsub.Topic) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return topic.Delete(ctx)
}

func deleteSubscription(subs *gpubsub.Subscription) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return subs.Delete(ctx)
}

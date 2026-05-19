package publisher

import (
	"context"
	"fmt"
	"sync"

	"cloud.google.com/go/pubsub"
)

// Publisher struct to manage Pub/Sub topics and client
type Publisher struct {
	client *pubsub.Client
	topics map[string]*pubsub.Topic
	mu     sync.Mutex
}

// Creates a new Publisher instance
func New(ctx context.Context, projectID string) (*Publisher, error) {
	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("pubsub.NewClient: %w", err)
	}
	return &Publisher{
		client: client,
		topics: make(map[string]*pubsub.Topic),
	}, nil
}

// Closes the Pub/Sub client
func (p *Publisher) Close() error {
	return p.client.Close()
}

// Publishes a message to a specific topic
func (p *Publisher) Publish(ctx context.Context, topicID string, data []byte) error {
	topic := p.topic(topicID)
	result := topic.Publish(ctx, &pubsub.Message{Data: data})
	if _, err := result.Get(ctx); err != nil {
		return fmt.Errorf("publish to %s: %w", topicID, err)
	}
	return nil
}

// Gets a topic from the cache or creates a new one
func (p *Publisher) topic(topicID string) *pubsub.Topic {
	p.mu.Lock()
	defer p.mu.Unlock()
	if t, ok := p.topics[topicID]; ok {
		return t
	}
	t := p.client.Topic(topicID)
	p.topics[topicID] = t
	return t
}
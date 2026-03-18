package publisher

import (
	"context"
	"cloud.google.com/go/pubsub"
)

func PublishMsg(ctx context.Context, projectID, topicID string, data []byte) error {
	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		return err
	}
	defer client.Close()

	topic := client.Topic(topicID)
	result := topic.Publish(ctx, &pubsub.Message{Data: data})
	_, err = result.Get(ctx)
	return err
}
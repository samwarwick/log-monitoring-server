package main

import (
	"context"
	"encoding/json"
	"fmt"

	"cloud.google.com/go/pubsub"
)

func publish(topicID string, msg Message) error {
	ctx := context.Background()

	t := client.Topic(topicID)

	msgBytes, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}

	result := t.Publish(ctx, &pubsub.Message{
		Data: []byte(msgBytes),
	})
	id, err := result.Get(ctx)
	if err != nil {
		return fmt.Errorf("publish(): result.Get: %v", err)
	}
	fmt.Printf("Published message (ID=%v)\n", id)
	return nil
}

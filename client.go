package main

import (
	"context"
	"fmt"
	"os"

	"cloud.google.com/go/pubsub"
	"cloud.google.com/go/pubsub/pstest"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
)

// https://pkg.go.dev/cloud.google.com/go/pubsub#Client
var client *pubsub.Client

// TODO: how to clean-up all these deferred Close() functions?
// https://stackoverflow.com/questions/62441316/golang-grpc-cant-keep-alive-the-client-connection-is-closing

func getMockClient(projectID string) (*pubsub.Client, error) {
	ctx := context.Background()
	srv := pstest.NewServer()
	// defer srv.Close()

	conn, err := grpc.Dial(srv.Addr, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	// defer conn.Close()

	client, err := pubsub.NewClient(ctx, projectID, option.WithGRPCConn(conn))

	if err != nil {
		return nil, err
	}
	// defer client.Close()

	return client, nil
}

func getEmulatorClient(projectID string) (*pubsub.Client, error) {
	ctx := context.Background()

	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("pubsub: NewClient: %v", err)
	}

	return client, nil
}

func configurePubsub(topicName, subscriptionName string) error {
	ctx := context.Background()

	// https://pkg.go.dev/cloud.google.com/go/pubsub#Topic
	topic := client.Topic(topicName)
	exists, err := topic.Exists(ctx)
	if !exists {
		topic, err = client.CreateTopic(ctx, topicName)
		if err != nil {
			return fmt.Errorf("pubsub: CreateTopic: %v", err)
		}
		fmt.Printf("Created topic: %s (ID=%v)\n", topic.String(), topic.ID())
	}

	// https://pkg.go.dev/cloud.google.com/go/pubsub#Subscription
	subscription := client.Subscription(subscriptionName)
	exists, err = subscription.Exists(ctx)
	if !exists {
		subscription, err = client.CreateSubscription(ctx, subscriptionName, pubsub.SubscriptionConfig{
			Topic: topic,
		})
		if err != nil {
			return fmt.Errorf("pubsub: CreateSubscription: %v", err)
		}
		fmt.Printf("Created subscription: %s (ID=%v)\n", subscription.String(), subscription.ID())
	}

	return nil
}

func configureClient(isEmulator bool) {
	var projectID = "lms"
	var err error
	if isEmulator {
		if os.Getenv("PUBSUB_EMULATOR_HOST") == "" {
			fmt.Println("PUBSUB_EMULATOR_HOST environment variable not defined")
			os.Exit(1)
		}
		client, err = getEmulatorClient(projectID)
	} else {
		client, err = getMockClient(projectID)
	}
	if err != nil {
		fmt.Printf("ERROR: get client%s\n", err.Error())
		return
	}
	fmt.Printf("Client created (emulator=%t)\n", isEmulator)

	err = configurePubsub("log-topic", "log-sub")
	if err != nil {
		fmt.Printf("ERROR: configurePubsub() %s \n", err.Error())
	}
}

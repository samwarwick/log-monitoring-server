package main

import (
	"context"
	"encoding/json"
	"fmt"
	"sync/atomic"
	"time"

	"cloud.google.com/go/pubsub"
)

type MonitorConfig struct {
	Duration             time.Duration
	MessageBatchSize     int
	FlushIntervalSeconds time.Duration
}

var messageQueue []*pubsub.Message

var logStore = LogStore{name: "service"}

func flusher(config MonitorConfig) {
	startTime := time.Now()
	fmt.Printf("Log flusher initialised. Duration=%s, Interval=%s, Batch size=%d\n", config.Duration, config.FlushIntervalSeconds, config.MessageBatchSize)
	ticker := time.NewTicker(config.FlushIntervalSeconds)
	for range ticker.C {
		if time.Now().Sub(startTime) >= config.Duration {
			fmt.Printf("Stopping flusher at %s\n", time.Now())
			ticker.Stop()
			break
		}
		flushMessages(config.MessageBatchSize)
	}
}

// TODO: Can we be sure that messageQueue thread-safe?
//       i.e. what happens if subscribe() adds to the queue while we flush it?
func flushMessages(batchSize int) {
	fmt.Printf("Flushing %d queued messages at %s\n", len(messageQueue), time.Now())
	var messageBatch []*pubsub.Message
	for _, message := range messageQueue {
		fmt.Printf("Message ID: %s\n", message.ID)
		messageBatch = append(messageBatch, message)
		if len(messageBatch) == batchSize {
			fmt.Printf("Flush %d messages\n", batchSize)
			storeMessages(messageBatch)
			messageQueue = deleteFlushedMessages(messageQueue, messageBatch)
			messageBatch = messageBatch[:0]
		}
	}
	if len(messageBatch) > 0 {
		fmt.Printf("Flush remaining %d message(s)\n", len(messageBatch))
		storeMessages(messageBatch)
		messageQueue = deleteFlushedMessages(messageQueue, messageBatch)
	}
	// fmt.Printf("Service Log\nRecords: %d\nDetails:%v\n", len(logStore.LogEntries()), logStore.LogEntries())
	if len(messageQueue) > 0 {
		panic("Message queue not fully flushed!")
	}
}

func storeMessages(messages []*pubsub.Message) {
	fmt.Printf("Storing %d log messages\n", len(messages))
	var logMessages []Message
	for _, message := range messages {
		var logMessage Message
		json.Unmarshal([]byte(string(message.Data)), &logMessage)
		logMessages = append(logMessages, logMessage)
	}
	logStore.AddMessages(logMessages)
}

func contains(messages []*pubsub.Message, id string) bool {
	for _, v := range messages {
		if v.ID == id {
			return true
		}
	}

	return false
}

// TODO: Is there a more efficient way to do this? There does not seem to be a simple way to delete elements from a slice in go.
func deleteFlushedMessages(queue []*pubsub.Message, flushedMessages []*pubsub.Message) []*pubsub.Message {
	var reducedMessages []*pubsub.Message
	for _, message := range queue {
		if !contains(flushedMessages, message.ID) {
			reducedMessages = append(reducedMessages, message)
		}
	}
	return reducedMessages
}

// https://pkg.go.dev/cloud.google.com/go/pubsub#Subscription.Receive
func subscribe(config MonitorConfig, subscriptionID string) (int32, error) {
	subscription := client.Subscription(subscriptionID)

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, config.Duration)
	defer cancel()

	if config.Duration > 0 && config.FlushIntervalSeconds > 0 {
		go flusher(config)
	}

	var received int32

	err := subscription.Receive(ctx, func(_ context.Context, msg *pubsub.Message) {
		//fmt.Fprintf(config.DebugOutput, "Received message (ID=%v): %q\n", msg.ID, string(msg.Data))
		var pubsubMessage Message
		json.Unmarshal([]byte(string(msg.Data)), &pubsubMessage)
		fmt.Printf("Received message (ID=%v). Service=%s, Payload=%s, Severity=%s\n", msg.ID, pubsubMessage.ServiceName, pubsubMessage.Payload, pubsubMessage.Severity)
		atomic.AddInt32(&received, 1)
		if config.MessageBatchSize > 1 {
			msg.Ack() // TODO: we should probably only Ack after the message has been succesfully written to the DB. Attempts to do this had unwanted side effects!
			messageQueue = append(messageQueue, msg)
			// fmt.Fprintf(config.DebugOutput, "Queued messages: %d\n", len(messageQueue))
		} else {
			msg.Ack()
		}
	})

	if err != nil {
		return received, fmt.Errorf("subscription.Receive: %v", err)
	}
	fmt.Printf("Received %d messages\n", received)

	return received, nil
}

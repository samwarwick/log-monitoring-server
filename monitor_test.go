package main

import (
	"fmt"
	"testing"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/stretchr/testify/assert"
)

func testMessage() Message {
	return Message{
		ServiceName: "Admin",
		Payload:     "Hello world",
		Severity:    Info.String(),
		Timestamp:   time.Now(),
	}
}

func publishTestMessages(messageCount int) {
	message := testMessage()

	for i := 1; i <= messageCount; i++ {
		message.Payload = fmt.Sprintf("Hello world %d", i)
		err := publish("log-topic", message)
		if err != nil {
			fmt.Printf("ERROR: publish() %s\n", err.Error())
		}
	}
}

func setup() {
	configureClient(false)
	logStore.Init()
}

func TestReceiveSingleMessage(t *testing.T) {
	setup()

	err := publish("log-topic", testMessage())
	if err != nil {
		t.Fatalf("publish() %s\n", err.Error())
	}

	var messageCount int32
	config := MonitorConfig{Duration: 3 * time.Second}
	messageCount, err = subscribe(config, "log-sub")
	if err != nil {
		t.Fatalf("subscribe() %s \n", err.Error())
	}
	if messageCount != 1 {
		t.Fatalf("subscribe() expected 1 message, returned %d\n", messageCount)
	}
}

func TestReceiveBatchSingleMessage(t *testing.T) {
	setup()

	err := publish("log-topic", testMessage())
	if err != nil {
		t.Fatalf("publish() %s \n", err.Error())
	}

	var messageCount int32
	config := MonitorConfig{Duration: 3 * time.Second, MessageBatchSize: 3, FlushIntervalSeconds: 2 * time.Second}
	messageCount, err = subscribe(config, "log-sub")
	if err != nil {
		t.Fatalf("subscribe() %s \n", err.Error())
	}
	if messageCount != 1 {
		t.Fatalf("subscribe() expected 1 message, returned %d\n", messageCount)
	}

	assert.Equal(t, 1, len(logStore.LogEntries()), "Message should be stored")
}

func TestReceiveBatchMessagesExceedBatchSize(t *testing.T) {
	setup()

	testMessageCount := 5
	publishTestMessages(testMessageCount)

	config := MonitorConfig{Duration: 5 * time.Second, MessageBatchSize: 2, FlushIntervalSeconds: 2 * time.Second}
	messageCount, err := subscribe(config, "log-sub")
	if err != nil {
		t.Fatalf("subscribe() %s\n", err.Error())
	}
	if int(messageCount) != testMessageCount {
		t.Fatalf("subscribe() expected 1 message, returned %d\n", messageCount)
	}

	assert.Equal(t, testMessageCount, len(logStore.LogEntries()), "All messages should be stored")

	// TODO: we also need to test that the messages were queued in 3 separate batches
	//       this could be achieved by grouping the logservice entries by the created_at time
}

func TestFlushMessagesLessThanBatchSize(t *testing.T) {
	// https://pkg.go.dev/cloud.google.com/go/internal/pubsub#Message
	messageQueue = append(messageQueue, &pubsub.Message{ID: "1", Data: []byte("Message 1")})
	messageQueue = append(messageQueue, &pubsub.Message{ID: "2", Data: []byte("Message 2")})
	flushMessages(5)
	assert.Equal(t, 0, len(messageQueue), "Message queue should be empty")
}

func TestFlushMessagesExceedBatchSize(t *testing.T) {
	// https://pkg.go.dev/cloud.google.com/go/internal/pubsub#Message
	messageQueue = append(messageQueue, &pubsub.Message{ID: "1", Data: []byte("Message 1")})
	messageQueue = append(messageQueue, &pubsub.Message{ID: "2", Data: []byte("Message 2")})
	messageQueue = append(messageQueue, &pubsub.Message{ID: "3", Data: []byte("Message 3")})
	messageQueue = append(messageQueue, &pubsub.Message{ID: "4", Data: []byte("Message 4")})
	messageQueue = append(messageQueue, &pubsub.Message{ID: "5", Data: []byte("Message 5")})
	flushMessages(2)
	assert.Equal(t, 0, len(messageQueue), "Message queue should be empty")

	// TODO: we should test that the messages were flushed in 3 separate batches
}

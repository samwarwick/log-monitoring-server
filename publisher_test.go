package main

import (
	"testing"
	"time"
)

var message = Message{
	ServiceName: "Admin",
	Payload:     "Hello world",
	Severity:    Info.String(),
	Timestamp:   time.Now(),
}

func TestPublishSingleMessage(t *testing.T) {
	configureClient(false)

	err := publish("log-topic", message)
	if err != nil {
		t.Fatalf("publish() %s\n", err.Error())
	}
}

func TestPublishToInvalidTopic(t *testing.T) {
	configureClient(false)

	err := publish("no-such-topic", message)
	if err == nil {
		t.Fatalf("publish() invalid topic did not throw an error\n")
	}
}

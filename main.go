package main

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

var config = MonitorConfig{Duration: 15 * time.Second}

func publishMessages(messageCount int) {
	message := Message{
		ServiceName: "Admin",
		Payload:     "Hello world",
		Severity:    Info.String(),
		Timestamp:   time.Now(),
	}

	for i := 1; i <= messageCount; i++ {
		message.Payload = fmt.Sprintf("Hello world %d", i)
		err := publish("log-topic", message)
		if err != nil {
			fmt.Printf("ERROR: publish() %s\n", err.Error())
		}
	}
}

func randomServiceName() string {
	services := []string{
		"alpha",
		"bravo",
		"charlie",
	}
	return services[rand.Intn(len(services))]
}

func publishSimulatedMessages(duration time.Duration) {
	fmt.Printf("Sending random messages for %s\n", duration)
	startTime := time.Now()
	minDelayMs := 50
	maxDelayMs := 2000

	message := Message{
		ServiceName: "Admin",
		Payload:     "Hello world",
		Severity:    Info.String(),
		Timestamp:   time.Now(),
	}

	messageCount := 0
	for time.Now().Sub(startTime) < duration {
		messageCount++
		message.ServiceName = randomServiceName()
		message.Payload = fmt.Sprintf("Hello from %s (message #%d)", message.ServiceName, messageCount)
		message.Severity = Severity(rand.Intn(5)).String()
		err := publish("log-topic", message)
		if err != nil {
			fmt.Printf("ERROR: publish() %s\n", err.Error())
		}
		time.Sleep(time.Duration(rand.Intn(maxDelayMs-minDelayMs)+minDelayMs) * time.Millisecond)
	}

	fmt.Printf("%d messages sent\n", messageCount)
}

func configureMonitor() {
	duration := os.Getenv("DURATION")
	intVar, err := strconv.Atoi(duration)
	if err == nil {
		config.Duration = time.Duration(intVar) * time.Second
	}

	batchSize := os.Getenv("BATCHSIZE")
	intVar, err = strconv.Atoi(batchSize)
	if err == nil {
		config.MessageBatchSize = intVar
	}

	flushInterval := os.Getenv("FLUSHINTERVAL")
	intVar, err = strconv.Atoi(flushInterval)
	if err == nil {
		config.FlushIntervalSeconds = time.Duration(intVar) * time.Second
	}
}

func subscribeMessages() {
	configureMonitor()

	_, err := subscribe(config, "log-sub")
	if err != nil {
		fmt.Printf("ERROR: subscribe() %s\n", err.Error())
	}
}

func main() {
	fmt.Printf("Log Monitoring Server (LMS)\n\n")

	if len(os.Args) == 1 {
		usage()
		os.Exit(1)
	}

	var err error
	messageCount := 3
	showServiceLog := false

	switch os.Args[1] {
	case "mock", "emu":
		configureClient(os.Args[1] == "emu")
		publishMessages(messageCount)
		subscribeMessages()
		showServiceLog = true
	case "pub":
		configureClient(true)
		if len(os.Args) > 2 {
			messageCount, err = strconv.Atoi(os.Args[2])
			if err != nil {
				fmt.Printf("Invalid message count -- %s \n", os.Args[2])
			}
		}
		publishMessages(messageCount)
	case "sim":
		configureClient(true)
		configureMonitor()
		publishSimulatedMessages(config.Duration)
	case "sub", "mon", "lms":
		configureClient(true)
		subscribeMessages()
		showServiceLog = true
	default:
		fmt.Println("Invalid command")
		os.Exit(1)
	}

	if showServiceLog && config.MessageBatchSize > 1 {
		fmt.Println("Service Log:")
		logStore.LogCsv()
		fmt.Printf("%d record(s)\n", len(logStore.logs))
	}

	os.Exit(0)
}

func usage() {
	fmt.Printf("Usage:\n\n    %s <command> [arguments]\n\n", filepath.Base(os.Args[0]))
	fmt.Printf("Commands:\n\n")
	fmt.Println("    mock            Publish and receive test mesages using pstest mock")
	fmt.Println("    emu             Publish and receive test mesages using PubSub Emulator (docker)")
	fmt.Println("    pub [qty]       Publish qty of test messages")
	fmt.Println("    mon             Run LMS monitor for DURATION")
	fmt.Printf("    sim             Run service simulator, sending random messages for DURATION\n\n")
	fmt.Printf("Enviroment variables:\n\n")
	fmt.Println("    PUBSUB_EMULATOR_HOST  Url and port of PubSub emulator e.g. localhost:8681")
	fmt.Println("    DURATION              Time (seconds) to run LMS and simulator")
	fmt.Println("    BATCHSIZE             Number of messages per batch")
	fmt.Printf("    FLUSHINTERVAL         Time interval (seconds) to flush queued messages to DB\n\n")
}

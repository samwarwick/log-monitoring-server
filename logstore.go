package main

import (
	"fmt"
	"time"
)

type LogStore struct {
	name     string
	logs     []ServiceLog
	severity []ServiceSeverity
}

func (s *LogStore) Init() {
	s.logs = []ServiceLog{}
	s.severity = []ServiceSeverity{}
}

func (s *LogStore) Name() string {
	return s.name
}

func (s *LogStore) LogEntries() []ServiceLog {
	return s.logs
}

func (s *LogStore) AddLogEntry(entry ServiceLog) {
	_ = entry.severity.String()
	entry.created_at = time.Now()
	fmt.Println("Adding", entry)
	s.logs = append(s.logs, entry)
}

func (s *LogStore) AddMessages(messages []Message) {
	for _, message := range messages {
		severity, _ := ParseSeverityString(message.Severity) // TODO improve this
		logEntry := ServiceLog{
			service_name: message.ServiceName,
			payload:      message.Payload,
			severity:     severity,
			timestamp:    message.Timestamp,
			created_at:   time.Now(),
		}
		fmt.Printf("Storing message: %v\n", logEntry)
		s.logs = append(s.logs, logEntry)
	}
}

func (s *LogStore) LogCsv() {
	fmt.Println("\"service_name\",\"payload\",\"severity\",\"timestamp\",\"created_at\"")
	for _, r := range s.logs {
		fmt.Printf("\"%s\",\"%s\",\"%s\",\"%s\",\"%s\"\n", r.service_name, r.payload, r.severity, r.timestamp, r.created_at)
	}
}

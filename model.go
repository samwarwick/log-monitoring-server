package main

import (
	"errors"
	"time"
)

type Message struct {
	ServiceName string    `json:"service_name"`
	Payload     string    `json:"payload"`
	Severity    string    `json:"severity"`
	Timestamp   time.Time `json:"timestamp"`
}

type ServiceLog struct {
	service_name string
	payload      string
	severity     Severity
	timestamp    time.Time
	created_at   time.Time
}

type ServiceSeverity struct {
	service_name string
	severity     Severity
	count        int32
	created_at   time.Time
}

type Severity int

const (
	Debug Severity = iota
	Info
	Warning
	Error
	Fatal
)

func (s Severity) String() string {
	return []string{"debug", "info", "warn", "error", "fatal"}[s]
}

func ParseSeverityString(str string) (Severity, error) {
	switch str {
	case "debug":
		return Debug, nil
	case "info":
		return Info, nil
	case "warn":
		return Warning, nil
	case "error":
		return Error, nil
	case "fatal":
		return Fatal, nil
	}
	return Fatal, errors.New("Invalid severity")
}

package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// go test -run TestLog*

func TestLogStoreName(t *testing.T) {
	db := LogStore{name: "test"}
	assert.Equal(t, "test", db.Name(), "Datastore name")
}

func TestLogStoreEntries(t *testing.T) {
	db := LogStore{name: "test"}
	db.AddLogEntry(ServiceLog{service_name: "alpha", payload: "Hello world", severity: Info})
	db.AddLogEntry(ServiceLog{service_name: "beta", payload: "Goodbye", severity: Debug})
	rows := len(db.LogEntries())
	assert.Equal(t, 2, rows, "Numbers of rows")
	assert.Equal(t, "info", (db.logs[0]).severity.String(), "Severity string")
	db.LogCsv()
}

# log-monitoring-server

An attempt at the Log Monitoring Server (LMS) technical challenge using the Go language and Google Pub/Sub messaging service.


## Overview

The overall objective of this exercise is to create a prototype event-driven log monitoring server (LMS). The majority of the requirements have been implemented (see Requirements below for details).

Please note that this exercise been tackled without any prior experience using the Go language or Google Pub/Sub.

The solution has been tested on Windows 10 and Ubuntu 20 (under WSL2).


### Tech Stack

* [go](https://go.dev/) (1.18)
* [GCP Pub/Sub](https://cloud.google.com/pubsub)
* [pstest](https://pkg.go.dev/cloud.google.com/go/pubsub/pstest) (fake PubSub service)
* [Docker](https://www.docker.com/)


## Quickstart

1. Ensure `go` version 1.18 or later is installed. Confirm with `go version`.

2. Run `./make.sh`. This will run the unit tests and build the LMS CLI (see Testing below)

3. Run `./lms mock`. This confirms that LMS CLI has been built successfully.

Note that Docker is required to run the integration/E2E tests.


## Requirements

Status summary of the key requirements discerned from the exercise documentation.

| Requirement | Status |
| --- |  --- |
| Multiple services can push logs to LMS | Done |
| Log messages are sent to a Google Pub/Sub topic| Done |
| Log messages are persisted in MySQL | Outstanding. A mock in-memory datastore has been used instead. |
| service_severity table must be kept in sync with service_logs | Outstanding. service_severity table not implemented |
| Logs are inserted in batches | Done |
| Batch size is configurable via an environment variable | Done. BATCHSIZE env var |
| Logs are flushed at regular intervals | Done |
| Flush interval is configurable via an environment variable | Done. FLUSHINTERVAL env var |
| DCL must not get overloaded with unacknowledged requests | Not tested. Implementation has focused on happy-path |
| E2E test proves published data is replicated in both tables | service_severity table not implemented |
| E2E test runs for at least 1 minute | Done |
| E2E test logs message count and batch sizes etc. | Done |
| Unit tests | Done |
| Makefile with proper build targets | Partially implemented with [make.sh](./make.sh) |
| README | Done |


## Solution Notes

### Main Files

Go source code

| File | Description
| --- |  --- |
| [client.go](client.go) | Create and configure PubSub client
| [logstore.go](logstore.go) | Mock log database in-memory
| [main.go](main.go) | LMS CLI
| [model.go](model.go) | PubSub message and database models
| [monitor.go](monitor.go) | LMS monitor and PubSub message subscriber
| [publisher.go](publisher.go) | PubSub message publisher

Bash scripts

| File | Description
| --- |  --- |
| [e2e_test.sh](e2e_test.sh) | Run full E2E test and save logs in `output` folder
| [make.sh](make.sh) | Build script for LMS CLI. Runs unit tests


### Message Flusher

It has been assumed that all outstanding messages should be flushed in every cycle. e.g. if there are 25 messages and the batch-size is 10, the cycle will flush 3 times; 10, 10 and 5. Outstanding messages are not deferred until the next cycle.

The flusher has been implemented using a simple go [ticker](https://gobyexample.com/tickers). This seems to be holding-up OK for the POC but is probably not an optimal approach. e.g. concerns over flushing the queue while the subscriber continues to add messages to it.

### Message Acknowledgement

To Ack or not to Ack? My original intent was to defer sending the Pub/Sub Ack message until the message had been successfully flushed. This is because we don't want to risk losing the message in the event of a failed flush. However, I encountered some issues with blocking processes when the ack was deferred so I reverted to ack upon receipt.


### Mocked Datastore

The solution does not implement the MySQL database. Instead a lightweight in-memory mock database store is used ([logstore.go](logstore.go)). This has the advantage of enabling the core functionality to be tested at unit-level instead of more complex and time-consuming integration tests. The solution has been designed in such a way that a real database provider can be plugged-in later. 


### Exception Handling

The solution has focused on the functional happy-path. For a complete implementation full exception handling needs to be implemented e.g. for invalid message schemas and content.


### Logging the Logger

LMS itself requires logging functionality. This is to aid debugging any issues and to confirm the functional correctness of the overall LMS solution.
For now all messages are sent to stdout in the form of print statements. A better solution would be to implement a standard logging API such as the go [logger](https://pkg.go.dev/google.golang.org/api/logging/v2).


### Other

'TODO' comments have been added in the source code concerning a few other minor points not already mentioned above.

There could possibly be issues with idempotency. e.g. if the same message is received more than once we need to ensure the database write can handle it.


## Testing

### Static Analysis

Static analysis is performed with `go vet` as part of the make script.

### Unit Tests

Unit tests exist for the essential functionality.
The tests use the built-in go testing library and [testify](https://github.com/stretchr/testify) assert.
GCP Pub/Sub is mocked with [pstest](https://pkg.go.dev/cloud.google.com/go/pubsub/pstest).
Test coverage is produced with `go test -cover`. Currently at 52%.

### Integration and E2E Tests

The LMS command-line interface (CLI) provides comprehensive functionality for running integration and E2E tests on the final solution. Running `lms` without any arguments displays the full usage:

```
Log Monitoring Server (LMS)

Usage:

    lms.exe <command> [arguments]

Commands:

    mock            Publish and receive test messages using pstest mock
    emu             Publish and receive test messages using PubSub Emulator (docker)
    pub [qty]       Publish qty of test messages
    mon             Run LMS monitor for DURATION
    sim             Run service simulator, sending random messages for DURATION

Environment variables:

    PUBSUB_EMULATOR_HOST  Url and port of PubSub emulator e.g. localhost:8681
    DURATION              Time (seconds) to run LMS and simulator
    BATCHSIZE             Number of messages per batch
    FLUSHINTERVAL         Time interval (seconds) to flush queued messages to DB
```

LMS requires the dockerised GCP PubSub emulator to be running. Start the emulator with:

```docker run --rm -ti -p 8681:8681 -e PUBSUB_PROJECT1=lms,log-topic:log-sub messagebird/gcloud-pubsub-emulator:latest```

Every LMS client session requires the PUBSUB_EMULATOR_HOST environment variable to be set, pointing to the emulator URL and port. e.g:

```export PUBSUB_EMULATOR_HOST=localhost:8681```

The script [e2e_test.sh](e2e_test.sh) has been provided to perform a full E2E test run lasting 2 minutes. 

The messages are batched in blocks of 10 and flushed at 20 second intervals. The LMS command is:

```DURATION=120 BATCHSIZE=10 FLUSHINTERVAL=20 ./lms mon```

Test messages are sent to LMS in parallel by using the LMS simulator command. This runs for 90 seconds, sending random messages at varying intervals. The command is:

```DURATION=90 ./lms sim```

Output is recorded in time-stamped logs within the `output` folder.

`logger_e2e-*.txt` contains the output from the service message simulator.

`monitor_e2e-*.txt` contains the output from the LMS monitor. The database log messages are dumped in CSV format at the end.

The published (logger) message count should always match the subscribed (monitor) message count.

See the [results](results) folder for an example test run. 

**NOTE**: The requirements mandate a default batch size of 5000 messages and a cycle time of 1 minute. For simplicity and efficiency much lower values have been used for testing purposes.

## Additional Thoughts 

### Address Outstanding Requirements

* Implement the database layer with MySQL. Use [dockertest](https://github.com/ory/dockertest)

* Implement the service_severity table in the database layer. It is important to maintain the data integrity of the service_logs and service_severity tables during updates. The database updates could be combined in a single [MySQL Transaction](https://www.mysqltutorial.org/mysql-transaction.aspx).

* Ensure DCL does not get overloaded with unacknowledged requests. This could be tested by adjusting the message delay in the LMS CLI simulator.

* Implement a proto-type data access layer client. Possibly via a REST API.


### Additional Testing

* Test against a live GCP Pub/Sub service instance.

* Increase unit testing coverage. Aim for about 80%.

* Further static analysis e.g. with Sonarqube.

* Non-functional testing. Test with large batch sizes and higher frequency of concurrent messages.


### Other

* Implement CI/CD with GitHub Actions. Run steps in [make.sh](make.sh) including unit tests.


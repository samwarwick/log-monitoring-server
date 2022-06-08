#!/bin/bash

export PUBSUB_EMULATOR_HOST=localhost:8681

echo "Running E2E LMS test"

start_time=`date +%Y-%m-%d-%H%M`

mkdir -p output

(time DURATION=120 BATCHSIZE=10 FLUSHINTERVAL=20 ./lms mon) 2>&1 | tee output/monitor_e2e_$start_time.txt &

(time DURATION=90 ./lms sim) 2>&1 | tee output/logger_e2e_$start_time.txt &

wait < <(jobs -p)
echo "Finished"
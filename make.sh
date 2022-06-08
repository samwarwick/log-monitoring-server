#!/bin/bash

echo "Analyze"
go vet

echo "Test"
go test -cover

echo "Build"
go build

echo "Done"
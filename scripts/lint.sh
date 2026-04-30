#!/bin/bash

if ! command -v golangci-lint &> /dev/null
then
    echo "golangci-lint CLI could not be found. Installing..."
    curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin
fi

golangci-lint -v run ./...

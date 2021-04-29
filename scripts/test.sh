#!/bin/bash
set -e -x

go test $(go list ./... ) -v -covermode=count -coverprofile=coverage.out -p 1
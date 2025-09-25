#!/bin/bash

go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
"$(go env GOPATH)/bin/sqlc" generate

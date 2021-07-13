# Development

## Overview

* Get dependencies: `go mod download`
* Build locally: `go build -v -o dynatrace-service ./cmd/`
* Run tests: `go test -race -v ./...`
* Run local: `ENV=local ./dynatrace-service`

## Debugging

Remote debugging is supported using [Skaffold](https://skaffold.dev/) via `skaffold debug`, which starts a [Delve](https://github.com/go-delve/delve) instance prior to running the service.

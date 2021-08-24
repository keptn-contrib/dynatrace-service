# Development

## Overview

* Get dependencies: `go mod download`
* Build locally: `go build -v -o dynatrace-service ./cmd/`
* Run tests: `go test -race -v ./...`
* Run local: `ENV=local ./dynatrace-service`

## Debugging

Remote debugging is supported using [Skaffold](https://skaffold.dev/) via `skaffold debug`, which starts a [Delve](https://github.com/go-delve/delve) instance prior to running the service.

## Setting the log output level

The minimum log level of messages emitted by the service may be set using the `LOG_LEVEL_DYNATRACE_SERVICE` environment variable. The following levels are supported: `panic`, `fatal`, `error`,`warn` (or `warning`), `info`, `debug` and `trace`. By default the minimum level is set to `info`, meaning that info, warning, error, fatal and panic messages are emitted.

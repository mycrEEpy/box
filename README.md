# box

`box` is an opinionated & minimalistic application framework for Go.

## Usage

See [examples](examples/main.go).

## Functionality

`box` provides the following functionality:

* Logger based on slog which automatically uses the JSON handler in Kubernetes
* Optional web server based on Echo with liveness/readiness & metrics endpoints

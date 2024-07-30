# box

[![Go Reference](https://pkg.go.dev/badge/github.com/mycreepy/box.svg)](https://pkg.go.dev/github.com/mycreepy/box)
[![Go Report Card](https://goreportcard.com/badge/github.com/mycreepy/box?style=flat-square)](https://goreportcard.com/report/github.com/mycreepy/box)

`box` is an opinionated & minimalistic application framework for Go.

## Functionality

`box` provides the following functionality:

* Logger based on slog which automatically uses the JSON handler in Kubernetes
* Optional web server based on Echo with liveness/readiness & metrics endpoints

## Usage

See [examples](examples/main.go).

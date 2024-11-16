# box

[![Go Reference](https://pkg.go.dev/badge/github.com/mycreepy/box.svg)](https://pkg.go.dev/github.com/mycreepy/box)
[![Go Report Card](https://goreportcard.com/badge/github.com/mycreepy/box?style=flat-square)](https://goreportcard.com/report/github.com/mycreepy/box)
[![Go Build & Test](https://github.com/mycrEEpy/box/actions/workflows/build.yml/badge.svg)](https://github.com/mycrEEpy/box/actions/workflows/build.yml)
[![Go Coverage](https://github.com/mycreepy/box/wiki/coverage.svg)](https://raw.githack.com/wiki/mycreepy/box/coverage.html)

![Gopher in a box](https://i.ibb.co/wzgHfC1/box-gopher.png)

`box` is an opinionated & minimalistic application framework for Go.

## Functionality

`box` provides the following functionality:

* Context which is canceled when the SIGINT or SIGTERM signal is received
* Logger based on slog which automatically uses the JSON handler in Kubernetes
* Optional web server based on Echo with liveness/readiness & metrics endpoints
* Configuration of GOMAXPROCS & GOMEMLIMIT to match cgroup limits in Kubernetes.

## Usage

See [examples](examples/main.go).

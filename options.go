package box

import (
	"net/http"
	"os"
	"runtime/trace"

	"github.com/labstack/echo-contrib/echoprometheus"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"gopkg.in/yaml.v3"
)

// Option is a modifier function which can alter the provided functionality of a Box.
type Option func(*Box)

// WithConfig configures the Box with the given Config.
func WithConfig(config Config) Option {
	return func(box *Box) {
		box.Config = config
	}
}

// WithConfigFromPath reads a configuration file from the given path, decodes it from YAML and calls WithConfig.
// Panics if the file can not be opened or decoded.
func WithConfigFromPath(path string) Option {
	return func(box *Box) {
		file, err := os.Open(path)
		if err != nil {
			panic(err)
		}

		defer file.Close()

		var wrapper configWrapper

		err = yaml.NewDecoder(file).Decode(&wrapper)
		if err != nil {
			panic(err)
		}

		WithConfig(wrapper.Config)(box)
	}
}

// WithGlobalLogger sets the global slog logger to the Box's logger.
func WithGlobalLogger() Option {
	return func(box *Box) {
		box.loggerGlobal = true
	}
}

// WithWebServer enables the web server functionality provided by WebServer.
func WithWebServer() Option {
	return func(box *Box) {
		box.WebServer = &WebServer{
			Echo: echo.New(),
			defaultLivenessProbe: func(c echo.Context) error {
				return c.NoContent(http.StatusOK)
			},
			defaultReadinessProbe: func(c echo.Context) error {
				return c.NoContent(http.StatusOK)
			},
		}

		box.WebServer.HideBanner = true
		box.WebServer.HidePort = true

		if box.Config.ListenAddress == "" {
			box.Config.ListenAddress = ":8000"
		}

		box.WebServer.Use(middleware.Recover(), echoprometheus.NewMiddleware("box_webserver"))

		box.WebServer.GET("/metrics", echoprometheus.NewHandler())
		box.WebServer.GET("/healthz", box.WebServer.defaultLivenessProbe)
		box.WebServer.GET("/readyz", box.WebServer.defaultReadinessProbe)
	}
}

// WithLivenessProbe allows to override the default liveness probe of the WebServer.
func WithLivenessProbe(probe func(c echo.Context) error) Option {
	return func(box *Box) {
		if box.WebServer == nil {
			WithWebServer()(box)
		}

		box.WebServer.GET("/healthz", probe)
	}
}

// WithReadinessProbe allows to override the default readiness probe of the WebServer.
func WithReadinessProbe(probe func(c echo.Context) error) Option {
	return func(box *Box) {
		if box.WebServer == nil {
			WithWebServer()(box)
		}

		box.WebServer.GET("/readyz", probe)
	}
}

// WithTraceFlightRecorder creates a tracing flight recorder and exposes it on /tracez.
func WithTraceFlightRecorder(cfg trace.FlightRecorderConfig) Option {
	return func(box *Box) {
		if box.WebServer == nil {
			WithWebServer()(box)
		}

		box.flightRecorder = trace.NewFlightRecorder(cfg)

		box.WebServer.GET("/tracez", func(c echo.Context) error {
			if box.flightRecorder == nil || !box.flightRecorder.Enabled() {
				return c.NoContent(http.StatusServiceUnavailable)
			}

			box.flightRecorderMut.Lock()
			defer box.flightRecorderMut.Unlock()

			_, err := box.flightRecorder.WriteTo(c.Response().Writer)
			return err
		})
	}
}

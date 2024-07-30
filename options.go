package box

import (
	"net/http"
	"os"

	"github.com/labstack/echo-contrib/echoprometheus"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"gopkg.in/yaml.v3"
)

type Option func(*Box)

func WithConfig(config Config) Option {
	return func(box *Box) {
		box.Config = config
		box.Logger = setupLogger(box.Config.LogLevel)
	}
}

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

func WithLivenessProbe(probe func(c echo.Context) error) Option {
	return func(box *Box) {
		if box.WebServer == nil {
			WithWebServer()(box)
		}

		box.WebServer.GET("/healthz", probe)
	}
}

func WithReadinessProbe(probe func(c echo.Context) error) Option {
	return func(box *Box) {
		if box.WebServer == nil {
			WithWebServer()(box)
		}

		box.WebServer.GET("/readyz", probe)
	}
}

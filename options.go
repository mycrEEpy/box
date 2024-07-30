package box

import (
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"gopkg.in/yaml.v3"
)

type Option func(*Box)

func WithConfig(config Config) Option {
	return func(box *Box) {
		box.Config = config
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

		box.Config = wrapper.Config
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

		box.WebServer.Use(middleware.Recover())

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

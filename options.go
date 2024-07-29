package box

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/labstack/echo/v4"
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

		switch {
		case strings.HasSuffix(path, ".yaml"), strings.HasSuffix(path, ".yml"):
			err = yaml.NewDecoder(file).Decode(&wrapper)
		case strings.HasSuffix(path, ".json"):
			err = json.NewDecoder(file).Decode(&wrapper)
		default:
			err = fmt.Errorf("unsupported file type: %s", path)
		}

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

		box.WebServer.Echo.HideBanner = true
		box.WebServer.Echo.HidePort = true

		if box.Config.ListenAddress == "" {
			box.Config.ListenAddress = ":8000"
		}

		box.WebServer.Echo.GET("/healthz", box.WebServer.defaultLivenessProbe)
		box.WebServer.Echo.GET("/readyz", box.WebServer.defaultReadinessProbe)
	}
}

func WithLivenessProbe(probe func(c echo.Context) error) Option {
	return func(box *Box) {
		if box.WebServer == nil {
			WithWebServer()(box)
		}

		box.WebServer.Echo.GET("/healthz", probe)
	}
}

func WithReadinessProbe(probe func(c echo.Context) error) Option {
	return func(box *Box) {
		if box.WebServer == nil {
			WithWebServer()(box)
		}

		box.WebServer.Echo.GET("/readyz", probe)
	}
}

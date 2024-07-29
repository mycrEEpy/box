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
		box.Config = &config
	}
}

func WithConfigFromPath(path string) Option {
	return func(box *Box) {
		file, err := os.Open(path)
		if err != nil {
			panic(err)
		}

		defer file.Close()

		switch {
		case strings.HasSuffix(path, ".yaml"), strings.HasSuffix(path, ".yml"):
			err = yaml.NewDecoder(file).Decode(box.Config)
		case strings.HasSuffix(path, ".json"):
			err = json.NewDecoder(file).Decode(box.Config)
		default:
			err = fmt.Errorf("unsupported file type: %s", path)
		}

		if err != nil {
			panic(err)
		}
	}
}

func WithWebServer() Option {
	return func(box *Box) {
		box.WebServer = &WebServer{
			Echo: echo.New(),
		}

		box.WebServer.Echo.HideBanner = true
		box.WebServer.Echo.HidePort = true

		if box.Config.ListenAddress == "" {
			box.Config.ListenAddress = ":8000"
		}

		if box.WebServer.LivenessProbe == nil {
			box.WebServer.LivenessProbe = func(c echo.Context) error {
				return c.NoContent(http.StatusOK)
			}
		}

		if box.WebServer.ReadinessProbe == nil {
			box.WebServer.ReadinessProbe = func(c echo.Context) error {
				return c.NoContent(http.StatusOK)
			}
		}

		box.WebServer.Echo.GET("/healthz", box.WebServer.LivenessProbe)
		box.WebServer.Echo.GET("/readyz", box.WebServer.ReadinessProbe)
	}
}

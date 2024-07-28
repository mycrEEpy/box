package main

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/mycreepy/box"
)

type App struct {
	*box.Box
}

func (app *App) helloWorld(c echo.Context) error {
	return c.String(http.StatusOK, "Hello, World!")
}

func main() {
	app := App{
		Box: box.New(
			box.WithConfigFromPath("./config.yml"),
			box.WithWebServer(),
		),
	}

	app.WebServer.GET("/", app.helloWorld)

	slog.Info("starting webserver", slog.String("listenAddress", app.Config.ListenAddress))

	err := app.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		panic(err)
	}
}

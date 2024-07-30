package box

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
)

var (
	DefaultConfig = Config{
		ListenAddress: ":8000",
	}
)

type Box struct {
	Config    Config
	Logger    *slog.Logger
	WebServer *WebServer
}

type WebServer struct {
	*echo.Echo
	defaultLivenessProbe  func(c echo.Context) error
	defaultReadinessProbe func(c echo.Context) error
}

type Config struct {
	ListenAddress string `json:"listenAddress" yaml:"listenAddress"`
}

type configWrapper struct {
	Config Config `json:"box" yaml:"box"`
}

func New(options ...Option) *Box {
	box := &Box{
		Config: DefaultConfig,
		Logger: setupLogger(),
	}

	for _, option := range options {
		option(box)
	}

	return box
}

func setupLogger() *slog.Logger {
	if isRunningInKubernetes() {
		return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}

	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
}

func (box *Box) ListenAndServe() error {
	if box.WebServer == nil {
		return errors.New("web server has not been initialized")
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	go func() {
		<-ctx.Done()

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer shutdownCancel()

		_ = box.WebServer.Echo.Shutdown(shutdownCtx)
	}()

	return box.WebServer.Echo.Start(box.Config.ListenAddress)
}

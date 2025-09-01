package box

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"runtime/trace"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
)

// DefaultConfig is the default Config for a Box.
var DefaultConfig = Config{
	ListenAddress: ":8000",
}

// Box is the main struct of this package which should be embedded into other structs.
type Box struct {
	Config Config

	Context       context.Context
	cancelContext context.CancelFunc

	Logger       *slog.Logger
	loggerGlobal bool

	WebServer *WebServer

	flightRecorderMut sync.Mutex
	flightRecorder    *trace.FlightRecorder
}

// WebServer provides the web server functionality of Box by embedding an Echo instance.
type WebServer struct {
	*echo.Echo
	defaultLivenessProbe  func(c echo.Context) error
	defaultReadinessProbe func(c echo.Context) error
}

// Config is the configuration struct for a Box.
type Config struct {
	LogLevel      string `yaml:"logLevel"`
	ListenAddress string `yaml:"listenAddress"`
	TLSCertFile   string `yaml:"tlsCertFile"`
	TLSKeyFile    string `yaml:"tlsKeyFile"`
}

type configWrapper struct {
	Config Config `json:"box" yaml:"box"`
}

// New constructs a new Box with various Option parameters and a Context,
// which is canceled when the SIGINT or SIGTERM signals are received.
func New(options ...Option) *Box {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)

	box := &Box{
		Context:       ctx,
		cancelContext: cancel,
	}

	WithConfig(DefaultConfig)(box)

	setupBoxWithFlags(box)

	for _, option := range options {
		option(box)
	}

	box.Logger = setupLogger(box.Config.LogLevel)

	if box.loggerGlobal {
		slog.SetDefault(box.Logger)
	}

	if box.flightRecorder != nil {
		err := box.flightRecorder.Start()
		if err != nil {
			panic(fmt.Errorf("failed to start trace flight recorder: %w", err))
		}

		go func() {
			<-box.Context.Done()
			box.flightRecorder.Stop()
		}()
	}

	return box
}

func setupLogger(levelStr string) *slog.Logger {
	if isRunningInKubernetes() {
		return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: parseLogLevel(levelStr, slog.LevelInfo),
		}))
	}

	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: parseLogLevel(levelStr, slog.LevelDebug),
	}))
}

func parseLogLevel(levelStr string, defaultLevel slog.Level) slog.Level {
	switch strings.ToLower(levelStr) {
	case "":
		return defaultLevel
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		panic(fmt.Errorf("unknown log level: %s", levelStr))
	}
}

// CancelContext cancels the Context of the Box.
func (box *Box) CancelContext() {
	box.cancelContext()
}

// ListenAndServe starts the listener of the WebServer and blocks until the Context of the Box is canceled.
// If TLSCertFile & TLSKeyFile are set, it starts an HTTPS server, otherwise an HTTP server.
// If the WebServer is not initialized, it returns an error.
func (box *Box) ListenAndServe() error {
	if box.WebServer == nil {
		return errors.New("web server has not been initialized")
	}

	defer box.cancelContext()

	go func() {
		<-box.Context.Done()

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer shutdownCancel()

		_ = box.WebServer.Echo.Shutdown(shutdownCtx)
	}()

	if len(box.Config.TLSCertFile) > 0 && len(box.Config.TLSKeyFile) > 0 {
		return box.WebServer.Echo.StartTLS(box.Config.ListenAddress, box.Config.TLSCertFile, box.Config.TLSKeyFile)
	}

	return box.WebServer.Echo.Start(box.Config.ListenAddress)
}

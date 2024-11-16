package box

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/KimMachineGun/automemlimit/memlimit"
	"github.com/labstack/echo/v4"
	"go.uber.org/automaxprocs/maxprocs"
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

	onceCpu sync.Once
	onceMem sync.Once
}

// WebServer provides the web server functionality of Box by embedding an Echo instance.
type WebServer struct {
	*echo.Echo
	defaultLivenessProbe  func(c echo.Context) error
	defaultReadinessProbe func(c echo.Context) error
}

// Config is the configuration struct for a Box.
type Config struct {
	LogLevel      string  `yaml:"logLevel"`
	ListenAddress string  `yaml:"listenAddress"`
	TLSCertFile   string  `yaml:"tlsCertFile"`
	TLSKeyFile    string  `yaml:"tlsKeyFile"`
	CpuMinThreads int     `yaml:"cpuMinThreads"`
	MemLimitRatio float64 `yaml:"memLimitRatio"`
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

	err := setupCgroupLimits(box.Config.CpuMinThreads, box.Config.MemLimitRatio)
	if err != nil {
		panic(err)
	}

	return box
}

func setupCgroupLimits(minThreads int, memLimitRatio float64) error {
	if isRunningInKubernetes() {
		_, err := maxprocs.Set(maxprocs.Min(minThreads))
		if err != nil {
			return err
		}

		_, err = memlimit.SetGoMemLimit(memLimitRatio)
		if err != nil {
			return err
		}
	}

	return nil
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

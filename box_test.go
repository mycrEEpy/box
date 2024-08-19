package box_test

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/mycreepy/box"
	"github.com/prometheus/client_golang/prometheus"
)

func TestNew(t *testing.T) {
	b := box.New(box.WithConfig(box.DefaultConfig))
	if b == nil {
		t.Error("box.New() returned nil")
		return
	}
}

func TestMustRegisterFlags(t *testing.T) {
	// Should not panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("MustRegisterFlags() panicked: %v", r)
		}
	}()

	box.MustRegisterFlags()
}

func TestWithConfig(t *testing.T) {
	b := box.New(box.WithConfig(box.Config{
		LogLevel: "warn",
	}))
	if b == nil {
		t.Error("box.New() returned nil")
		return
	}

	if b.Config.LogLevel != "warn" {
		t.Errorf("LogLevel should be 'warn' but got '%s'", b.Config.LogLevel)
	}
}

func TestWithConfigFromPath(t *testing.T) {
	b := box.New(box.WithConfigFromPath("testdata/config.yml"))
	if b == nil {
		t.Error("box.New() returned nil")
		return
	}

	if b.Config.LogLevel != "error" {
		t.Errorf("LogLevel should be 'error' but got '%s'", b.Config.LogLevel)
	}

	if !b.Logger.Enabled(context.Background(), slog.LevelError) {
		t.Error("logger should log in error level")
	}
}

func TestWithGlobalLogger(t *testing.T) {
	b := box.New(box.WithGlobalLogger())
	if b == nil {
		t.Error("box.New() returned nil")
		return
	}

	if slog.Default() != b.Logger {
		t.Errorf("slog.Default() should be '%+v' but got '%+v'", b.Logger, slog.Default())
	}
}

func TestLoggerInKubernetes(t *testing.T) {
	t.Setenv("KUBERNETES_SERVICE_HOST", "localhost")

	b := box.New(box.WithConfig(box.Config{LogLevel: "debug"}))
	if b == nil {
		t.Error("box.New() returned nil")
		return
	}

	if !b.Logger.Enabled(context.Background(), slog.LevelDebug) {
		t.Error("logger should log in debug level")
	}
}

func TestWithWebServer(t *testing.T) {
	prometheus.DefaultRegisterer = prometheus.NewRegistry()

	b := box.New(box.WithWebServer())
	if b == nil {
		t.Error("box.New() returned nil")
		return
	}

	if b.WebServer == nil {
		t.Error("WebServer is nil")
		return
	}
}

func TestWithReadinessProbe(t *testing.T) {
	prometheus.DefaultRegisterer = prometheus.NewRegistry()

	probe := func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	}

	b := box.New(box.WithReadinessProbe(probe))
	if b == nil {
		t.Error("box.New() returned nil")
		return
	}

	if b.WebServer == nil {
		t.Error("WebServer is nil")
		return
	}
}

func TestWithLivenessProbe(t *testing.T) {
	prometheus.DefaultRegisterer = prometheus.NewRegistry()

	probe := func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	}

	b := box.New(box.WithLivenessProbe(probe))
	if b == nil {
		t.Error("box.New() returned nil")
		return
	}

	if b.WebServer == nil {
		t.Error("WebServer is nil")
		return
	}
}

func TestBox_ListenAndServe(t *testing.T) {
	prometheus.DefaultRegisterer = prometheus.NewRegistry()

	b := box.New(
		box.WithConfig(box.Config{ListenAddress: "localhost:8000"}),
		box.WithWebServer(),
	)
	if b == nil {
		t.Error("box.New() returned nil")
		return
	}

	go func() {
		err := b.ListenAndServe()
		if !errors.Is(err, http.ErrServerClosed) {
			t.Errorf("ListenAndServe should have stopped gracefully but did not: %s", err)
			return
		}
	}()

	time.Sleep(500 * time.Millisecond)

	resp, err := http.Get("http://localhost:8000/healthz")
	if err != nil {
		t.Errorf("http.Get returned an error: %s", err)
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("http.Get returned wrong status code: %d", resp.StatusCode)
		return
	}

	b.CancelContext()

	time.Sleep(500 * time.Millisecond)
}

func TestBox_CancelContext(t *testing.T) {
	b := box.New()
	if b == nil {
		t.Error("box.New() returned nil")
		return
	}

	b.CancelContext()

	<-b.Context.Done()
}

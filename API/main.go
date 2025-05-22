package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
)

func main() {
	if err := run(); err != nil {
		log.Fatalln(err)
	}
}

func run() (err error) {
	// Capture interrupt signal
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	// Set up OpenTelemetry.
	otelShutdown, err := setupOTelSDK(ctx)
	if err != nil {
		return
	}
	// Handle shutdown properly
	defer func() {
		err = errors.Join(err, otelShutdown(context.Background()))
	}()

	// Setup Echo Server
	e := echo.New()

	// Middleware configurations
	// e.Use(middleware.Logger())
	e.Use(otelecho.Middleware("api"))

	// Routing
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "hello world")
	})

	// Channel for capturing server error
	srvErr := make(chan error, 1)

	// Server listen inside a routine
	go func() {
		srvErr <- e.Start(":3000")
	}()

	select {
	case err = <-srvErr:
		return
	case <-ctx.Done():
		stop()
	}

	err = e.Shutdown(context.Background())
	return
}

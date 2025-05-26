package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/labstack/echo/v4"
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

	// Setup Echo Server
	e := echo.New()

	// Middleware configurations
	// e.Use(middleware.Logger())

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

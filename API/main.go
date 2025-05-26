package main

import (
	"apm-learning/telemetry"
	"context"
	"encoding/json"
	"errors"
	"fmt"
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

	// Set up OpenTelemetry.
	otelShutdown, err := telemetry.SetupOTelSDK(ctx)
	if err != nil {
		return
	}
	// Handle shutdown properly
	defer func() {
		err = errors.Join(err, otelShutdown(context.Background()))
	}()

	// Create a new Echo instance
	e := echo.New()

	// Add a health check endpoint that will be skipped by the tracer
	// Routing
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "hello world")
	})

	e.GET("/health", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	exportRoutes(e)

	// Start the server
	// Channel for capturing server error
	srvErr := make(chan error, 1)

	// Server listen inside a routine
	go func() {
		srvErr <- e.Start(":8000")
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

func exportRoutes(e *echo.Echo) {
	// Extract routes
	routes := e.Routes()
	println("exporting routes")

	// Prepare a simplified structure for JSON
	type Route struct {
		Method string `json:"method"`
		Path   string `json:"path"`
		Name   string `json:"name"`
	}

	var routeList []Route
	for _, r := range routes {
		routeList = append(routeList, Route{
			Method: r.Method,
			Path:   r.Path,
			Name:   r.Name,
		})
	}

	// Serialize to JSON
	data, err := json.MarshalIndent(routeList, "", "  ")
	if err != nil {
		println("Failed to marshal routes: %v", err)
		return
	}

	// Save to file
	err = os.WriteFile("routes.json", data, 0644)
	if err != nil {
		println("Failed to write file: %v", err)
		return
	}

	fmt.Println("Routes saved to routes.json")
}

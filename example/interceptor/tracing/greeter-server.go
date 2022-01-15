// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.
package main

import (
	"context"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/rookie-ninja/rk-echo/interceptor/context"
	"github.com/rookie-ninja/rk-echo/interceptor/log/zap"
	"github.com/rookie-ninja/rk-echo/interceptor/tracing/telemetry"
	"github.com/rookie-ninja/rk-entry/entry"
	rkmidtrace "github.com/rookie-ninja/rk-entry/middleware/tracing"
	"net/http"
)

// In this example, we will start a new echo server with tracing interceptor enabled.
// Listen on port of 8080 with GET /rk/v1/greeter?name=<xxx>.
func main() {
	// ****************************************
	// ********** Create Exporter *************
	// ****************************************

	// Export trace to stdout
	exporter := rkmidtrace.NewFileExporter("stdout")

	// Export trace to local file system
	// exporter := rkmidtrace.NewFileExporter("logs/trace.log")

	// Export trace to jaeger agent
	// exporter := rkmidtrace.NewJaegerExporter(jaeger.WithAgentEndpoint())

	// ********************************************
	// ********** Enable interceptors *************
	// ********************************************
	interceptors := []echo.MiddlewareFunc{
		rkecholog.Interceptor(),
		rkechotrace.Interceptor(
			// Entry name and entry type will be used for distinguishing interceptors. Recommended.
			// rkmidtrace.WithEntryNameAndType("greeter", "echo"),
			//
			// Provide an exporter.
			rkmidtrace.WithExporter(exporter),
			//
			// Provide propagation.TextMapPropagator
			// rkmidtrace.WithPropagator(<propagator>),
			//
			// Provide SpanProcessor
			// rkmidtrace.WithSpanProcessor(<span processor>),
			//
			// Provide TracerProvider
			// rkmidtrace.WithTracerProvider(<trace provider>),
		),
	}

	// 1: Create echo server
	server := startGreeterServer(interceptors...)
	defer server.Shutdown(context.TODO())

	// 2: Wait for ctrl-C to shutdown server
	rkentry.GlobalAppCtx.WaitForShutdownSig()
}

// Start echo server.
func startGreeterServer(interceptors ...echo.MiddlewareFunc) *echo.Echo {
	server := echo.New()
	server.Use(interceptors...)
	server.GET("/rk/v1/greeter", Greeter)

	go func() {
		if err := server.Start(":8080"); err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()

	return server
}

// GreeterResponse Response of Greeter.
type GreeterResponse struct {
	Message string
}

// Greeter Handler.
func Greeter(ctx echo.Context) error {
	rkechoctx.GetLogger(ctx).Info("Received request from client.")

	ctx.JSON(http.StatusOK, &GreeterResponse{
		Message: fmt.Sprintf("Hello %s!", ctx.QueryParam("name")),
	})

	return nil
}

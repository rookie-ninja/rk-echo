// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.
package main

import (
	"context"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rookie-ninja/rk-echo/interceptor/context"
	"github.com/rookie-ninja/rk-echo/interceptor/metrics/prom"
	"github.com/rookie-ninja/rk-entry/entry"
	rkmidmetrics "github.com/rookie-ninja/rk-entry/middleware/metrics"
	"github.com/rookie-ninja/rk-prom"
	"net/http"
)

// In this example, we will start a new echo server with metrics interceptor enabled.
// Listen on port of 8080 with GET /rk/v1/greeter?name=<xxx>.
func main() {
	// Override app name which would replace namespace value in prometheus.
	// rkentry.GlobalAppCtx.GetAppInfoEntry().AppName = "newApp"

	// ********************************************
	// ********** Enable interceptors *************
	// ********************************************
	interceptors := []echo.MiddlewareFunc{
		rkechometrics.Interceptor(
			// Entry name and entry type will be used for distinguishing interceptors. Recommended.
			rkmidmetrics.WithEntryNameAndType("greeter", "echo"),
			//
			// Provide new prometheus registerer.
			// Default value is prometheus.DefaultRegisterer
			rkmidmetrics.WithRegisterer(prometheus.NewRegistry()),
		),
	}

	// 1: Start prometheus client
	// By default, we will start prometheus client at localhost:1608/metrics
	promEntry := rkprom.RegisterPromEntry()
	promEntry.Bootstrap(context.Background())
	defer promEntry.Interrupt(context.Background())

	// 2: Create echo server
	server := startGreeterServer(interceptors...)
	defer server.Shutdown(context.TODO())

	// 3: Wait for ctrl-C to shutdown server
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
	// ******************************************
	// ********** rpc-scoped logger *************
	// ******************************************
	//
	// RequestId will be printed if enabled by bellow codes.
	// 1: Enable rkechometa.Interceptor() in server side.
	// 2: rkechoctx.SetHeaderToClient(ctx, rkechoctx.RequestIdKey, rkcommon.GenerateRequestId())
	//
	rkechoctx.GetLogger(ctx).Info("Received request from client.")

	ctx.JSON(http.StatusOK, &GreeterResponse{
		Message: fmt.Sprintf("Hello %s!", ctx.QueryParam("name")),
	})

	return nil
}

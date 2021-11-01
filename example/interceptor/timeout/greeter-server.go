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
	"github.com/rookie-ninja/rk-echo/interceptor/panic"
	"github.com/rookie-ninja/rk-echo/interceptor/timeout"
	"github.com/rookie-ninja/rk-entry/entry"
	"net/http"
	"time"
)

// In this example, we will start a new gin server with rate limit interceptor enabled.
// Listen on port of 8080 with GET /rk/v1/greeter?name=<xxx>.
func main() {
	// ******************************************************
	// ********** Override App name and version *************
	// ******************************************************
	//
	// rkentry.GlobalAppCtx.GetAppInfoEntry().AppName = "demo-app"
	// rkentry.GlobalAppCtx.GetAppInfoEntry().Version = "demo-version"

	// ********************************************
	// ********** Enable interceptors *************
	// ********************************************
	interceptors := []echo.MiddlewareFunc{
		rkechopanic.Interceptor(),
		rkecholog.Interceptor(),
		rkechotimeout.Interceptor(
		// Entry name and entry type will be used for distinguishing interceptors. Recommended.
		//rkechotimeout.WithEntryNameAndType("greeter", "echo"),
		//
		// Provide timeout and response handler, a default one would be assigned with http.StatusRequestTimeout
		// This option impact all routes
		//rkechotimeout.WithTimeoutAndResp(time.Second, nil),
		//
		// Provide timeout and response handler by path, a default one would be assigned with http.StatusRequestTimeout
		//rkechotimeout.WithTimeoutAndRespByPath("/rk/v1/healthy", time.Second, nil),
		),
	}

	// 1: Create gin server
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
	// ******************************************
	// ********** rpc-scoped logger *************
	// ******************************************
	//
	// RequestId will be printed if enabled by bellow codes.
	// 1: Enable rkechometa.Interceptor() in server side.
	// 2: rkechoctx.SetHeaderToClient(ctx, rkechoctx.RequestIdKey, rkcommon.GenerateRequestId())
	//
	rkechoctx.GetLogger(ctx).Info("Received request from client.")

	// Set request id with X-Request-Id to outgoing headers.
	// rkechoctx.SetHeaderToClient(ctx, rkechoctx.RequestIdKey, "this-is-my-request-id-overridden")

	// Sleep for 5 seconds waiting to be timed out by interceptor
	time.Sleep(10 * time.Second)

	ctx.JSON(http.StatusOK, &GreeterResponse{
		Message: fmt.Sprintf("Hello %s!", ctx.QueryParam("name")),
	})

	return nil
}

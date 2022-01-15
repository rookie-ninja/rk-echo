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
	"github.com/rookie-ninja/rk-echo/interceptor/ratelimit"
	"github.com/rookie-ninja/rk-entry/entry"
	"net/http"
)

// In this example, we will start a new echo server with rate limit interceptor enabled.
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
		rkecholog.Interceptor(),
		rkecholimit.Interceptor(
		// Entry name and entry type will be used for distinguishing interceptors. Recommended.
		// rkmidlimit.WithEntryNameAndType("greeter", "echo"),
		//
		// Provide algorithm, rkmidlimit.LeakyBucket and rkmidlimit.TokenBucket was available, default is TokenBucket.
		//rkmidlimit.WithAlgorithm(rkmidlimit.LeakyBucket),
		//
		// Provide request per second, if provide value of zero, then no requests will be pass through and user will receive an error with
		// resource exhausted.
		//rkmidlimit.WithReqPerSec(10),
		//
		// Provide request per second with path name.
		// The name should be full path name. if provide value of zero,
		// then no requests will be pass through and user will receive an error with resource exhausted.
		//rkmidlimit.WithReqPerSecByPath("/rk/v1/greeter", 0),
		//
		// Provide user function of limiter
		//rkmidlimit.WithGlobalLimiter(func() error {
		//	 return nil
		//}),
		//
		// Provide user function of limiter by path name.
		// The name should be full path name.
		//rkmidlimit.WithLimiterByPath("/rk/v1/greeter", func() error {
		//	 return nil
		//}),
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
	// ******************************************
	// ********** rpc-scoped logger *************
	// ******************************************
	//
	// RequestId will be printed if enabled by bellow codes.
	// 1: Enable rkechometa.Interceptor() in server side.
	// 2: rkechoctx.AddHeaderToClient(ctx, rkechoctx.RequestIdKey, rkcommon.GenerateRequestId())
	//
	rkechoctx.GetLogger(ctx).Info("Received request from client.")

	// Set request id with X-Request-Id to outgoing headers.
	// rkechoctx.SetHeaderToClient(ctx, rkechoctx.RequestIdKey, "this-is-my-request-id-overridden")

	ctx.JSON(http.StatusOK, &GreeterResponse{
		Message: fmt.Sprintf("Hello %s!", ctx.QueryParam("name")),
	})

	return nil
}

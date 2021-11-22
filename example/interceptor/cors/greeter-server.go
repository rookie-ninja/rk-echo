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
	rkechocors "github.com/rookie-ninja/rk-echo/interceptor/cors"
	"github.com/rookie-ninja/rk-entry/entry"
	"net/http"
)

// In this example, we will start a new echo server with cors interceptor enabled.
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
		rkechocors.Interceptor(
			// Entry name and entry type will be used for distinguishing interceptors. Recommended.
			rkechocors.WithEntryNameAndType("greeter", "echo"),
			// Provide skipper function
			// rkechocors.WithSkipper(func(e echo.Context) bool {
			//     return false
			// }),
			// Bellow section is for CORS policy.
			// Please refer https://developer.mozilla.org/en-US/docs/Web/HTTP/CORS for details.
			// Provide allowed origins
			rkechocors.WithAllowOrigins("http://localhost:*"),
			// Whether to allow credentials
			// rkechocors.WithAllowCredentials(true),
			// Provide expose headers
			// rkechocors.WithExposeHeaders(""),
			// Provide max age
			// rkechocors.WithMaxAge(1),
			// Provide allowed headers
			// rkechocors.WithAllowHeaders(""),
			// Provide allowed headers
			// rkechocors.WithAllowMethods(""),
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
	// 2: rkechoctx.SetHeaderToClient(ctx, rkechoctx.RequestIdKey, rkcommon.GenerateRequestId())
	//
	rkechoctx.GetLogger(ctx).Info("Received request from client.")

	return ctx.JSON(http.StatusOK, &GreeterResponse{
		Message: fmt.Sprintf("Hello %s!", ctx.QueryParam("name")),
	})
}

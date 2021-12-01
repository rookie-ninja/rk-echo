// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"github.com/labstack/echo/v4"
	"github.com/rookie-ninja/rk-echo/interceptor/context"
	rkechosec "github.com/rookie-ninja/rk-echo/interceptor/secure"
	"github.com/rookie-ninja/rk-entry/entry"
	"net/http"
)

// In this example, we will start a new echo server with secure interceptor enabled.
// Listen on port of 8080 with GET /rk/v1/greeter.
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
		rkechosec.Interceptor(
			// Required, entry name and entry type will be used for distinguishing interceptors. Recommended.
			rkechosec.WithEntryNameAndType("greeter", "echo"),
			//
			// X-XSS-Protection header value.
			// Optional. Default value "1; mode=block".
			//rkechosec.WithXSSProtection("my-value"),
			//
			// X-Content-Type-Options header value.
			// Optional. Default value "nosniff".
			//rkechosec.WithContentTypeNosniff("my-value"),
			//
			// X-Frame-Options header value.
			// Optional. Default value "SAMEORIGIN".
			//rkechosec.WithXFrameOptions("my-value"),
			//
			// Optional, Strict-Transport-Security header value.
			//rkechosec.WithHSTSMaxAge(1),
			//
			// Optional, excluding subdomains of HSTS, default is false
			//rkechosec.WithHSTSExcludeSubdomains(true),
			//
			// Optional, enabling HSTS preload, default is false
			//rkechosec.WithHSTSPreloadEnabled(true),
			//
			// Content-Security-Policy header value.
			// Optional. Default value "".
			//rkechosec.WithContentSecurityPolicy("my-value"),
			//
			// Content-Security-Policy-Report-Only header value.
			// Optional. Default value false.
			//rkechosec.WithCSPReportOnly(true),
			//
			// Referrer-Policy header value.
			// Optional. Default value "".
			//rkechosec.WithReferrerPolicy("my-value"),
			//
			// Ignoring path prefix.
			//rkechosec.WithIgnorePrefix("/rk/v1"),
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
	rkechoctx.GetLogger(ctx).Info("Received request from client.")

	ctx.JSON(http.StatusOK, &GreeterResponse{
		Message: "Received message!",
	})

	return nil
}

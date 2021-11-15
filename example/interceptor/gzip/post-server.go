// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"fmt"
	"github.com/labstack/echo/v4"
	rkechogzip "github.com/rookie-ninja/rk-echo/interceptor/gzip"
	"github.com/rookie-ninja/rk-entry/entry"
	"io"
	"net/http"
	"strings"
)

// In this example, we will start a new echo server with gzip interceptor enabled.
// Listen on port of 8080 with POST /rk/v1/post.
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
		rkechogzip.Interceptor(
		// Entry name and entry type will be used for distinguishing interceptors. Recommended.
		// rkechogzip.WithEntryNameAndType("greeter", "echo"),
		//
		// Provide level of compression.
		// Available options are
		// - NoCompression
		// - BestSpeed
		// - BestCompression
		// - DefaultCompression
		// - HuffmanOnly
		//rkechogzip.WithLevel(rkechogzip.DefaultCompression),
		//
		// Provide skipper function
		//rkechogzip.WithSkipper(func(e echo.Context) bool {
		//	return false
		//}),
		),
	}

	// 1: Create echo server
	server := startPostServer(interceptors...)
	defer server.Shutdown(context.TODO())

	// 2: Wait for ctrl-C to shutdown server
	rkentry.GlobalAppCtx.WaitForShutdownSig()
}

// Start echo server.
func startPostServer(interceptors ...echo.MiddlewareFunc) *echo.Echo {
	server := echo.New()
	server.Use(interceptors...)
	server.POST("/rk/v1/post", post)

	go func() {
		if err := server.Start(":8080"); err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()

	return server
}

// PostResponse Response of Post.
type PostResponse struct {
	Message string
}

// post Handler.
func post(ctx echo.Context) error {
	buf := new(strings.Builder)
	io.Copy(buf, ctx.Request().Body)

	ctx.JSON(http.StatusOK, &PostResponse{
		Message: fmt.Sprintf("Received %s!", buf.String()),
	})

	return nil
}

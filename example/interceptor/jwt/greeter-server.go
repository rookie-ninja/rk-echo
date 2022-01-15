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
	"github.com/rookie-ninja/rk-echo/interceptor/jwt"
	"github.com/rookie-ninja/rk-entry/entry"
	rkmidjwt "github.com/rookie-ninja/rk-entry/middleware/jwt"
	"net/http"
)

// In this example, we will start a new echo server with jwt interceptor enabled.
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
		//rkecholog.Interceptor(),
		rkechojwt.Interceptor(
			// Entry name and entry type will be used for distinguishing interceptors. Recommended.
			// rkginjwt.WithEntryNameAndType("greeter", "gin"),
			//
			// Required, provide signing key.
			rkmidjwt.WithSigningKey([]byte("my-secret")),
			//
			// rkmidjwt.WithIgnorePrefix(""),
			//
			// Optional, provide skipper function
			//rkmidjwt.WithSkipper(func(e *gin.Context) bool {
			//	return true
			//}),
			//
			// Optional, provide token parse function, default one will be assigned.
			//rkmidjwt.WithParseTokenFunc(func(auth string, ctx *gin.Context) (*jwt.Token, error) {
			//	return nil, nil
			//}),
			//
			// Optional, provide key function, default one will be assigned.
			//rkmidjwt.WithKeyFunc(func(token *jwt.Token) (interface{}, error) {
			//	return nil, nil
			//}),
			//
			// Optional, default is Bearer
			//rkmidjwt.WithAuthScheme("Bearer"),
			//
			// Optional
			//rkmidjwt.WithTokenLookup("header:my-jwt-header-key"),
			//
			// Optional, default is HS256
			//rkmidjwt.WithSigningAlgorithm(rkginjwt.AlgorithmHS256),
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
		Message: fmt.Sprintf("Is token valid:%v!", rkechoctx.GetJwtToken(ctx).Valid),
	})

	return nil
}

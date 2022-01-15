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
	"github.com/rookie-ninja/rk-entry/entry"
	rkmidlog "github.com/rookie-ninja/rk-entry/middleware/log"
	"net/http"
)

// In this example, we will start a new echo server with log interceptor enabled.
// Listen on port of 8080 with GET /rk/v1/greeter?name=<xxx>.
func main() {
	// ********************************************
	// ********** Enable interceptors *************
	// ********************************************
	interceptors := []echo.MiddlewareFunc{
		//rkechometa.Interceptor(),
		rkecholog.Interceptor(
			// Entry name and entry type will be used for distinguishing interceptors. Recommended.
			rkmidlog.WithEntryNameAndType("greeter", "echo"),
			//
			// Zap logger would be logged as JSON format.
			// rkmidlog.WithZapLoggerEncoding("json"),
			//
			// Event logger would be logged as JSON format.
			// rkmidlog.WithEventLoggerEncoding("json"),
			//
			// Zap logger would be logged to specified path.
			// rkmidlog.WithZapLoggerOutputPaths("logs/server-zap.log"),
			//
			// Event logger would be logged to specified path.
			// rkmidlog.WithEventLoggerOutputPaths("logs/server-event.log"),
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

// Greeter Handler for greeter.
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

	// *******************************************
	// ********** rpc-scoped event  *************
	// *******************************************
	//
	// Get rkquery.Event which would be printed as soon as request finish.
	// User can call any Add/Set/Get functions on rkquery.Event
	//
	// rkechoctx.GetEvent(ctx).AddPair("rk-key", "rk-value")

	// *********************************************
	// ********** Get incoming headers *************
	// *********************************************
	//
	// Read headers sent from client.
	//
	//for k, v := range rkechoctx.GetIncomingHeaders(ctx) {
	//	 fmt.Println(fmt.Sprintf("%s: %s", k, v))
	//}

	// *********************************************************
	// ********** Add headers will send to client **************
	// *********************************************************
	//
	// Send headers to client with this function
	//
	//rkechoctx.AddHeaderToClient(ctx, "from-server", "value")

	// ***********************************************
	// ********** Get and log request id *************
	// ***********************************************
	//
	// RequestId will be printed on both client and server side.
	//
	//rkechoctx.SetHeaderToClient(ctx, rkechoctx.RequestIdKey, rkcommon.GenerateRequestId())

	ctx.JSON(http.StatusOK, &GreeterResponse{
		Message: fmt.Sprintf("Hello %s!", ctx.QueryParam("name")),
	})

	return nil
}

// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.
package main

import (
	"context"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/rookie-ninja/rk-echo/interceptor/auth"
	"github.com/rookie-ninja/rk-echo/interceptor/context"
	"github.com/rookie-ninja/rk-echo/interceptor/log/zap"
	"github.com/rookie-ninja/rk-entry/entry"
	"net/http"
)

// In this example, we will start a new echo server with auth interceptor enabled.
// Listen on port of 8080 with GET /rk/v1/greeter?name=<xxx>.
func main() {
	// ********************************************
	// ********** Enable interceptors *************
	// ********************************************
	interceptors := []echo.MiddlewareFunc{
		rkecholog.Interceptor(),
		rkechoauth.Interceptor(
			// rkechoauth.WithIgnorePrefix("/rk/v1/greeter"),
			rkechoauth.WithBasicAuth("", "rk-user:rk-pass"),
			rkechoauth.WithApiKeyAuth("rk-api-key"),
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
	validateCtx(ctx)

	ctx.JSON(http.StatusOK, &GreeterResponse{
		Message: fmt.Sprintf("Hello %s!", ctx.QueryParam("name")),
	})

	return nil
}

func validateCtx(ctx echo.Context) {
	// 1: get incoming headers
	printIndex("[1]: get incoming headers")
	prettyHeader(rkechoctx.GetIncomingHeaders(ctx))

	// 2: add header to client
	printIndex("[2]: add header to client")
	rkechoctx.AddHeaderToClient(ctx, "add-key", "add-value")

	// 3: set header to client
	printIndex("[3]: set header to client")
	rkechoctx.SetHeaderToClient(ctx, "set-key", "set-value")

	// 4: get event
	printIndex("[4]: get event")
	rkechoctx.GetEvent(ctx).SetCounter("my-counter", 1)

	// 5: get logger
	printIndex("[5]: get logger")
	rkechoctx.GetLogger(ctx).Info("error msg")

	// 6: get request id
	printIndex("[6]: get request id")
	fmt.Println(rkechoctx.GetRequestId(ctx))

	// 7: get trace id
	printIndex("[7]: get trace id")
	fmt.Println(rkechoctx.GetTraceId(ctx))

	// 8: get entry name
	printIndex("[8]: get entry name")
	fmt.Println(rkechoctx.GetEntryName(ctx))

	// 9: get trace span
	printIndex("[9]: get trace span")
	fmt.Println(rkechoctx.GetTraceSpan(ctx))

	// 10: get tracer
	printIndex("[10]: get tracer")
	fmt.Println(rkechoctx.GetTracer(ctx))

	// 11: get trace provider
	printIndex("[11]: get trace provider")
	fmt.Println(rkechoctx.GetTracerProvider(ctx))

	// 12: get tracer propagator
	printIndex("[12]: get tracer propagator")
	fmt.Println(rkechoctx.GetTracerPropagator(ctx))

	// 13: inject span
	printIndex("[13]: inject span")
	req := &http.Request{}
	rkechoctx.InjectSpanToHttpRequest(ctx, req)

	// 14: new trace span
	printIndex("[14]: new trace span")
	fmt.Println(rkechoctx.NewTraceSpan(ctx, "my-span"))

	// 15: end trace span
	printIndex("[15]: end trace span")
	rkechoctx.EndTraceSpan(ctx, rkechoctx.NewTraceSpan(ctx, "my-span"), true)
}

func printIndex(key string) {
	fmt.Println(fmt.Sprintf("%s", key))
}

func prettyHeader(header http.Header) {
	for k, v := range header {
		fmt.Println(fmt.Sprintf("%s:%s", k, v))
	}
}

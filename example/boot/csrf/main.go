// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.
package main

import (
	"context"
	_ "embed"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/rookie-ninja/rk-echo/boot"
	"github.com/rookie-ninja/rk-echo/middleware/context"
	"github.com/rookie-ninja/rk-entry/v2/entry"
	"net/http"
)

//go:embed boot.yaml
var boot []byte

func main() {
	// Bootstrap preload entries
	rkentry.BootstrapPreloadEntryYAML(boot)

	// Bootstrap gin entry from boot config
	res := rkecho.RegisterEchoEntryYAML(boot)

	// Register GET and POST method of /rk/v1/greeter
	entry := res["greeter"].(*rkecho.EchoEntry)
	entry.Echo.GET("/rk/v1/greeter", Greeter)
	entry.Echo.POST("/rk/v1/greeter", Greeter)

	// Bootstrap echo entry
	res["greeter"].Bootstrap(context.Background())

	// Wait for shutdown signal
	rkentry.GlobalAppCtx.WaitForShutdownSig()

	// Interrupt echo entry
	res["greeter"].Interrupt(context.Background())
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

	return ctx.JSON(http.StatusOK, &GreeterResponse{
		Message: fmt.Sprintf("CSRF token:%v", rkechoctx.GetCsrfToken(ctx)),
	})
}

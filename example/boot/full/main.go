// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.
package main

import (
	"context"
	"github.com/labstack/echo/v4"
	"github.com/rookie-ninja/rk-echo/boot"
	"github.com/rookie-ninja/rk-entry/entry"
	"net/http"
)

func main() {
	// Bootstrap basic entries from boot config.
	rkentry.RegisterInternalEntriesFromConfig("example/boot/full/boot.yaml")

	// Bootstrap gin entry from boot config
	res := rkecho.RegisterEchoEntriesWithConfig("example/boot/full/boot.yaml")

	//res["greeter"].(*rkecho.EchoEntry).Echo.GET("/hello", hello)
	//res["greeter"].(*rkecho.EchoEntry).Echo.HideBanner = true
	//res["greeter"].(*rkecho.EchoEntry).Echo.HidePort = true
	//

	// Bootstrap gin entry
	res["greeter"].Bootstrap(context.Background())

	// Wait for shutdown signal
	rkentry.GlobalAppCtx.WaitForShutdownSig()

	// Interrupt gin entry
	res["greeter"].Interrupt(context.Background())
}

// Handler
func hello(c echo.Context) error {
	return c.String(http.StatusOK, "Hello, World!")
}
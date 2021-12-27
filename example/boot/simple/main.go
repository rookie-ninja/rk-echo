// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.
package main

import (
	"context"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/rookie-ninja/rk-echo/boot"
	"github.com/rookie-ninja/rk-entry/entry"
	"net/http"
)

// @title RK Swagger for Echo
// @version 1.0
// @description This is a greeter service with rk-boot.
func main() {
	// Bootstrap basic entries from boot config.
	rkentry.RegisterInternalEntriesFromConfig("example/boot/simple/boot.yaml")

	// Bootstrap echo entry from boot config
	res := rkecho.RegisterEchoEntriesWithConfig("example/boot/simple/boot.yaml")

	// Get EchoEntry
	echoEntry := res["greeter"].(*rkecho.EchoEntry)
	// Use *echo.Echo adding handler.
	echoEntry.Echo.GET("/v1/greeter", Greeter)

	// Bootstrap echo entry
	res["greeter"].Bootstrap(context.Background())

	// Wait for shutdown signal
	rkentry.GlobalAppCtx.WaitForShutdownSig()

	// Interrupt echo entry
	res["greeter"].Interrupt(context.Background())
}

// @Summary Greeter service
// @Id 1
// @version 1.0
// @produce application/json
// @Param name query string true "Input name"
// @Success 200 {object} GreeterResponse
// @Router /v1/greeter [get]
func Greeter(ctx echo.Context) error {
	return ctx.JSON(http.StatusOK, &GreeterResponse{
		Message: fmt.Sprintf("Hello %s!", ctx.QueryParam("name")),
	})
}

// Response.
type GreeterResponse struct {
	Message string
}

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
	"io"
	"net/http"
	"strings"
)

func main() {
	// Bootstrap basic entries from boot config.
	rkentry.RegisterInternalEntriesFromConfig("example/boot/gzip/boot.yaml")

	// Bootstrap echo entry from boot config
	res := rkecho.RegisterEchoEntriesWithConfig("example/boot/gzip/boot.yaml")

	// Register post method
	res["greeter"].(*rkecho.EchoEntry).Echo.POST("/rk/v1/post", post)

	// Bootstrap echo entry
	res["greeter"].Bootstrap(context.Background())

	// Wait for shutdown signal
	rkentry.GlobalAppCtx.WaitForShutdownSig()

	// Interrupt echo entry
	res["greeter"].Interrupt(context.Background())
}

// PostResponse Response of Post.
type PostResponse struct {
	ReceivedMessage string
}

// post Handler.
func post(ctx echo.Context) error {
	buf := new(strings.Builder)
	io.Copy(buf, ctx.Request().Body)

	ctx.JSON(http.StatusOK, &PostResponse{
		ReceivedMessage: fmt.Sprintf("%s", buf.String()),
	})

	return nil
}

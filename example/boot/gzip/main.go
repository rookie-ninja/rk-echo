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
	"github.com/rookie-ninja/rk-entry/v2/entry"
	"io"
	"net/http"
	"strings"
)

//go:embed boot.yaml
var boot []byte

func main() {
	// Bootstrap preload entries
	rkentry.BootstrapPreloadEntryYAML(boot)

	// Bootstrap gin entry from boot config
	res := rkecho.RegisterEchoEntryYAML(boot)

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

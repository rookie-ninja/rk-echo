// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

// Package rkechopanic is a middleware of echo framework for recovering from panic
package rkechopanic

import (
	"github.com/labstack/echo/v4"
	"github.com/rookie-ninja/rk-echo/middleware/context"
	"github.com/rookie-ninja/rk-entry/v2/error"
	"github.com/rookie-ninja/rk-entry/v2/middleware"
	"github.com/rookie-ninja/rk-entry/v2/middleware/panic"
	"net/http"
)

// Interceptor returns a echo.MiddlewareFunc (middleware)
func Interceptor(opts ...rkmidpanic.Option) echo.MiddlewareFunc {
	set := rkmidpanic.NewOptionSet(opts...)

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			ctx.Set(rkmid.EntryNameKey.String(), set.GetEntryName())

			handlerFunc := func(resp rkerror.ErrorInterface) {
				ctx.JSON(http.StatusInternalServerError, resp)
			}
			beforeCtx := set.BeforeCtx(rkechoctx.GetEvent(ctx), rkechoctx.GetLogger(ctx), handlerFunc)
			set.Before(beforeCtx)

			defer beforeCtx.Output.DeferFunc()

			return next(ctx)
		}
	}
}

// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

// Package rkecholog is a middleware for echo framework for logging RPC.
package rkecholog

import (
	"github.com/labstack/echo/v4"
	"github.com/rookie-ninja/rk-echo/middleware/context"
	"github.com/rookie-ninja/rk-entry/v2/middleware"
	"github.com/rookie-ninja/rk-entry/v2/middleware/log"
	"strconv"
)

// Middleware returns a echo.MiddlewareFunc (middleware) that logs requests using uber-go/zap.
func Middleware(opts ...rkmidlog.Option) echo.MiddlewareFunc {
	set := rkmidlog.NewOptionSet(opts...)

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			ctx.Set(rkmid.EntryNameKey.String(), set.GetEntryName())

			// call before
			beforeCtx := set.BeforeCtx(ctx.Request())
			set.Before(beforeCtx)

			ctx.Set(rkmid.EventKey.String(), beforeCtx.Output.Event)
			ctx.Set(rkmid.LoggerKey.String(), beforeCtx.Output.Logger)

			err := next(ctx)

			// call after
			afterCtx := set.AfterCtx(
				rkechoctx.GetRequestId(ctx),
				rkechoctx.GetTraceId(ctx),
				strconv.Itoa(ctx.Response().Status))
			set.After(beforeCtx, afterCtx)

			return err
		}
	}
}

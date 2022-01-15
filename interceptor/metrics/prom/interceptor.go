// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

// Package rkechometrics is a middleware for echo framework which record prometheus metrics for RPC
package rkechometrics

import (
	"github.com/labstack/echo/v4"
	"github.com/rookie-ninja/rk-entry/middleware"
	"github.com/rookie-ninja/rk-entry/middleware/metrics"
	"strconv"
)

// Interceptor create a new prometheus metrics interceptor with options.
func Interceptor(opts ...rkmidmetrics.Option) echo.MiddlewareFunc {
	set := rkmidmetrics.NewOptionSet(opts...)

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			ctx.Set(rkmid.EntryNameKey.String(), set.GetEntryName())

			beforeCtx := set.BeforeCtx(ctx.Request())
			set.Before(beforeCtx)

			err := next(ctx)

			afterCtx := set.AfterCtx(strconv.Itoa(ctx.Response().Status))
			set.After(beforeCtx, afterCtx)

			return err
		}
	}
}

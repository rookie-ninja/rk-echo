// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

// Package rkechoprom is a middleware for echo framework which record prometheus metrics for RPC
package rkechoprom

import (
	"github.com/labstack/echo/v4"
	"github.com/rookie-ninja/rk-entry/v2/middleware"
	"github.com/rookie-ninja/rk-entry/v2/middleware/prom"
	"strconv"
)

// Middleware create a new prometheus metrics interceptor with options.
func Middleware(opts ...rkmidprom.Option) echo.MiddlewareFunc {
	set := rkmidprom.NewOptionSet(opts...)

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

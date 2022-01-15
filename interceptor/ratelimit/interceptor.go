// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

// Package rkecholimit is a middleware of echo framework for adding rate limit in RPC response
package rkecholimit

import (
	"github.com/labstack/echo/v4"
	"github.com/rookie-ninja/rk-entry/middleware"
	"github.com/rookie-ninja/rk-entry/middleware/ratelimit"
)

// Interceptor Add rate limit interceptors.
func Interceptor(opts ...rkmidlimit.Option) echo.MiddlewareFunc {
	set := rkmidlimit.NewOptionSet(opts...)

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			ctx.Set(rkmid.EntryNameKey.String(), set.GetEntryName())

			beforeCtx := set.BeforeCtx(ctx.Request())
			set.Before(beforeCtx)

			if beforeCtx.Output.ErrResp != nil {
				return ctx.JSON(beforeCtx.Output.ErrResp.Err.Code, beforeCtx.Output.ErrResp)
			}

			return next(ctx)
		}
	}
}

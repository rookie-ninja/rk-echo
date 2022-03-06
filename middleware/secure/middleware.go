// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

// Package rkechosec is a middleware of echo framework for adding secure headers in RPC response
package rkechosec

import (
	"github.com/labstack/echo/v4"
	"github.com/rookie-ninja/rk-entry/v2/middleware"
	"github.com/rookie-ninja/rk-entry/v2/middleware/secure"
)

// Middleware Add security interceptors.
//
// Mainly copied from bellow.
// https://github.com/labstack/echo/blob/master/middleware/secure.go
func Middleware(opts ...rkmidsec.Option) echo.MiddlewareFunc {
	set := rkmidsec.NewOptionSet(opts...)

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			ctx.Set(rkmid.EntryNameKey.String(), set.GetEntryName())

			// case 1: return to user if error occur
			beforeCtx := set.BeforeCtx(ctx.Request())
			set.Before(beforeCtx)

			for k, v := range beforeCtx.Output.HeadersToReturn {
				ctx.Response().Header().Set(k, v)
			}

			return next(ctx)
		}
	}
}

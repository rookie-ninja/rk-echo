// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

// Package rkechojwt is a middleware for echo framework which validating jwt token for RPC
package rkechojwt

import (
	"github.com/labstack/echo/v4"
	"github.com/rookie-ninja/rk-entry/v2/middleware"
	"github.com/rookie-ninja/rk-entry/v2/middleware/jwt"
)

// Middleware Add jwt interceptors.
//
// Mainly copied from bellow.
// https://github.com/labstack/echo/blob/master/middleware/jwt.go
func Middleware(opts ...rkmidjwt.Option) echo.MiddlewareFunc {
	set := rkmidjwt.NewOptionSet(opts...)

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			ctx.Set(rkmid.EntryNameKey.String(), set.GetEntryName())

			beforeCtx := set.BeforeCtx(ctx.Request(), nil)
			set.Before(beforeCtx)

			// case 1: error response
			if beforeCtx.Output.ErrResp != nil {
				return ctx.JSON(beforeCtx.Output.ErrResp.Code(),
					beforeCtx.Output.ErrResp)
			}

			// insert into context
			ctx.Set(rkmid.JwtTokenKey.String(), beforeCtx.Output.JwtToken)

			return next(ctx)
		}
	}
}

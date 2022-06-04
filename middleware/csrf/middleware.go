// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

// Package rkechocsrf is a middleware for echo framework which validating csrf token for RPC
package rkechocsrf

import (
	"github.com/labstack/echo/v4"
	"github.com/rookie-ninja/rk-entry/v2/middleware"
	"github.com/rookie-ninja/rk-entry/v2/middleware/csrf"
	"net/http"
)

// Middleware Add csrf interceptors.
//
// Mainly copied from bellow.
// https://github.com/labstack/echo/blob/master/middleware/csrf.go
func Middleware(opts ...rkmidcsrf.Option) echo.MiddlewareFunc {
	set := rkmidcsrf.NewOptionSet(opts...)

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			ctx.Set(rkmid.EntryNameKey.String(), set.GetEntryName())

			beforeCtx := set.BeforeCtx(ctx.Request())
			set.Before(beforeCtx)

			if beforeCtx.Output.ErrResp != nil {
				return ctx.JSON(beforeCtx.Output.ErrResp.Code(), beforeCtx.Output.ErrResp)
			}

			for _, v := range beforeCtx.Output.VaryHeaders {
				ctx.Response().Header().Add(rkmid.HeaderVary, v)
			}

			if beforeCtx.Output.Cookie != nil {
				http.SetCookie(ctx.Response(), beforeCtx.Output.Cookie)
			}

			// store token in the context
			ctx.Set(rkmid.CsrfTokenKey.String(), beforeCtx.Input.Token)

			return next(ctx)
		}
	}
}

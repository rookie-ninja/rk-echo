// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

// Package rkechocors is a CORS middleware for echo framework
package rkechocors

import (
	"github.com/labstack/echo/v4"
	rkmid "github.com/rookie-ninja/rk-entry/middleware"
	rkmidcors "github.com/rookie-ninja/rk-entry/middleware/cors"
	"net/http"
)

// Interceptor Add cors interceptors.
//
// Mainly copied and modified from bellow.
// https://github.com/labstack/echo/blob/master/middleware/cors.go
func Interceptor(opts ...rkmidcors.Option) echo.MiddlewareFunc {
	set := rkmidcors.NewOptionSet(opts...)

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			ctx.Set(rkmid.EntryNameKey.String(), set.GetEntryName())

			beforeCtx := set.BeforeCtx(ctx.Request())
			set.Before(beforeCtx)

			for k, v := range beforeCtx.Output.HeadersToReturn {
				ctx.Response().Header().Set(k, v)
			}

			for _, v := range beforeCtx.Output.HeaderVary {
				ctx.Response().Header().Add(rkmid.HeaderVary, v)
			}

			// case 1: with abort
			if beforeCtx.Output.Abort {
				return ctx.NoContent(http.StatusNoContent)
			}

			// case 2: call next
			return next(ctx)
		}
	}
}

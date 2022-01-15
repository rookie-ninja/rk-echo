// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

// Package rkechometa is a middleware of echo framework for adding metadata in RPC response
package rkechometa

import (
	"github.com/labstack/echo/v4"
	"github.com/rookie-ninja/rk-echo/interceptor/context"
	rkmid "github.com/rookie-ninja/rk-entry/middleware"
	rkmidmeta "github.com/rookie-ninja/rk-entry/middleware/meta"
)

// Interceptor will add common headers as extension style in http response.
func Interceptor(opts ...rkmidmeta.Option) echo.MiddlewareFunc {
	set := rkmidmeta.NewOptionSet(opts...)

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			ctx.Set(rkmid.EntryNameKey.String(), set.GetEntryName())

			beforeCtx := set.BeforeCtx(rkechoctx.GetEvent(ctx))
			set.Before(beforeCtx)

			ctx.Set(rkmid.HeaderRequestId, beforeCtx.Output.RequestId)

			for k, v := range beforeCtx.Output.HeadersToReturn {
				ctx.Response().Header().Set(k, v)
			}

			return next(ctx)
		}
	}
}

// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

// Package rkechotimeout is a middleware of echo framework for timing out request in RPC response
package rkechotimeout

import (
	"github.com/labstack/echo/v4"
	"github.com/rookie-ninja/rk-echo/interceptor"
)

// Interceptor Add timeout interceptors.
func Interceptor(opts ...Option) echo.MiddlewareFunc {
	set := newOptionSet(opts...)

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			ctx.Set(rkechointer.RpcEntryNameKey, set.EntryName)

			return set.Tick(ctx, next)
		}
	}
}

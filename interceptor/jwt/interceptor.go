// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

// Package rkechojwt is a middleware for echo framework which validating jwt token for RPC
package rkechojwt

import (
	"github.com/labstack/echo/v4"
	"github.com/rookie-ninja/rk-common/error"
	"github.com/rookie-ninja/rk-echo/interceptor"
	"net/http"
)

// Interceptor Add jwt interceptors.
//
// Mainly copied from bellow.
// https://github.com/labstack/echo/blob/master/middleware/jwt.go
func Interceptor(opts ...Option) echo.MiddlewareFunc {
	set := newOptionSet(opts...)

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			ctx.Set(rkechointer.RpcEntryNameKey, set.EntryName)

			if set.Skipper(ctx) {
				return next(ctx)
			}

			// extract token from extractor
			var auth string
			var err error
			for _, extractor := range set.extractors {
				// Extract token from extractor, if it's not fail break the loop and
				// set auth
				auth, err = extractor(ctx)
				if err == nil {
					break
				}
			}

			if err != nil {
				return ctx.JSON(http.StatusUnauthorized, rkerror.New(
					rkerror.WithHttpCode(http.StatusUnauthorized),
					rkerror.WithMessage("invalid or expired jwt"),
					rkerror.WithDetails(err)))
			}

			// parse token
			token, err := set.ParseTokenFunc(auth, ctx)

			if err != nil {
				return ctx.JSON(http.StatusUnauthorized, rkerror.New(
					rkerror.WithHttpCode(http.StatusUnauthorized),
					rkerror.WithMessage("invalid or expired jwt"),
					rkerror.WithDetails(err)))
			}

			// insert into context
			ctx.Set(rkechointer.RpcJwtTokenKey, token)

			return next(ctx)
		}
	}
}

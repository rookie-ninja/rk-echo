// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

// Package rkechocors is a CORS middleware for echo framework
package rkechocors

import (
	"github.com/labstack/echo/v4"
	"github.com/rookie-ninja/rk-echo/interceptor"
	"net/http"
	"strconv"
	"strings"
)

// Interceptor Add cors interceptors.
//
// Mainly copied and modified from bellow.
// https://github.com/labstack/echo/blob/master/middleware/cors.go
func Interceptor(opts ...Option) echo.MiddlewareFunc {
	set := newOptionSet(opts...)

	allowMethods := strings.Join(set.AllowMethods, ",")
	allowHeaders := strings.Join(set.AllowHeaders, ",")
	exposeHeaders := strings.Join(set.ExposeHeaders, ",")
	maxAge := strconv.Itoa(set.MaxAge)

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			ctx.Set(rkechointer.RpcEntryNameKey, set.EntryName)

			if set.Skipper(ctx) {
				return next(ctx)
			}

			originHeader := ctx.Request().Header.Get(echo.HeaderOrigin)
			preflight := ctx.Request().Method == http.MethodOptions

			// 1: if no origin header was provided, we will return 204 if request is not a OPTION method
			if originHeader == "" {
				// 1.1: if not a preflight request, then pass through
				if !preflight {
					return next(ctx)
				}

				// 1.2: if it is a preflight request, then return with 204
				return ctx.NoContent(http.StatusNoContent)
			}

			// 2: origin not allowed, we will return 204 if request is not a OPTION method
			if !set.isOriginAllowed(originHeader) {
				// 2.1: if not a preflight request, then pass through
				if !preflight {
					return ctx.NoContent(http.StatusFound)
				}

				// 2.2: if it is a preflight request, then return with 204
				return ctx.NoContent(http.StatusNoContent)
			}

			// 3: not a OPTION method
			if !preflight {
				ctx.Response().Header().Set(echo.HeaderAccessControlAllowOrigin, originHeader)
				// 3.1: add Access-Control-Allow-Credentials
				if set.AllowCredentials {
					ctx.Response().Header().Set(echo.HeaderAccessControlAllowCredentials, "true")
				}
				// 3.2: add Access-Control-Expose-Headers
				if exposeHeaders != "" {
					ctx.Response().Header().Set(echo.HeaderAccessControlExposeHeaders, exposeHeaders)
				}
				return next(ctx)
			}

			// 4: preflight request, return 204
			// add related headers including:
			//
			// - Vary
			// - Access-Control-Allow-Origin
			// - Access-Control-Allow-Methods
			// - Access-Control-Allow-Credentials
			// - Access-Control-Allow-Headers
			// - Access-Control-Max-Age
			ctx.Response().Header().Add(echo.HeaderVary, echo.HeaderAccessControlRequestMethod)
			ctx.Response().Header().Add(echo.HeaderVary, echo.HeaderAccessControlRequestHeaders)
			ctx.Response().Header().Set(echo.HeaderAccessControlAllowOrigin, originHeader)
			ctx.Response().Header().Set(echo.HeaderAccessControlAllowMethods, allowMethods)

			// 4.1: Access-Control-Allow-Credentials
			if set.AllowCredentials {
				ctx.Response().Header().Set(echo.HeaderAccessControlAllowCredentials, "true")
			}

			// 4.2: Access-Control-Allow-Headers
			if allowHeaders != "" {
				ctx.Response().Header().Set(echo.HeaderAccessControlAllowHeaders, allowHeaders)
			} else {
				h := ctx.Request().Header.Get(echo.HeaderAccessControlRequestHeaders)
				if h != "" {
					ctx.Response().Header().Set(echo.HeaderAccessControlAllowHeaders, h)
				}
			}
			if set.MaxAge > 0 {
				// 4.3: Access-Control-Max-Age
				ctx.Response().Header().Set(echo.HeaderAccessControlMaxAge, maxAge)
			}

			return ctx.NoContent(http.StatusNoContent)
		}
	}
}

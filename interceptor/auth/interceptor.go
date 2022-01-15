// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

// Package rkechoauth is auth middleware for echo framework
package rkechoauth

import (
	"github.com/labstack/echo/v4"
	rkmid "github.com/rookie-ninja/rk-entry/middleware"
	rkmidauth "github.com/rookie-ninja/rk-entry/middleware/auth"
)

// Interceptor validate bellow authorization.
//
// 1: Basic Auth: The client sends HTTP requests with the Authorization header that contains the word Basic, followed by a space and a base64-encoded(non-encrypted) string username: password.
// 2: Bearer Token: Commonly known as token authentication. It is an HTTP authentication scheme that involves security tokens called bearer tokens.
// 3: API key: An API key is a token that a client provides when making API calls. With API key auth, you send a key-value pair to the API in the request headers.
func Interceptor(opts ...rkmidauth.Option) echo.MiddlewareFunc {
	set := rkmidauth.NewOptionSet(opts...)

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			ctx.Set(rkmid.EntryNameKey.String(), set.GetEntryName())

			// case 1: return to user if error occur
			beforeCtx := set.BeforeCtx(ctx.Request())
			set.Before(beforeCtx)

			if beforeCtx.Output.ErrResp != nil {
				for k, v := range beforeCtx.Output.HeadersToReturn {
					ctx.Response().Header().Set(k, v)
				}
				return ctx.JSON(beforeCtx.Output.ErrResp.Err.Code, beforeCtx.Output.ErrResp)
			}

			return next(ctx)
		}
	}
}

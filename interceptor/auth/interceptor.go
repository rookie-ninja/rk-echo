// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

// Package rkechoauth is auth middleware for echo framework
package rkechoauth

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/rookie-ninja/rk-common/error"
	"github.com/rookie-ninja/rk-echo/interceptor"
	"net/http"
	"strings"
)

// Interceptor validate bellow authorization.
//
// 1: Basic Auth: The client sends HTTP requests with the Authorization header that contains the word Basic, followed by a space and a base64-encoded(non-encrypted) string username: password.
// 2: Bearer Token: Commonly known as token authentication. It is an HTTP authentication scheme that involves security tokens called bearer tokens.
// 3: API key: An API key is a token that a client provides when making API calls. With API key auth, you send a key-value pair to the API in the request headers.
func Interceptor(opts ...Option) echo.MiddlewareFunc {
	set := newOptionSet(opts...)

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			ctx.Set(rkechointer.RpcEntryNameKey, set.EntryName)

			err := before(ctx, set)

			if err != nil {
				return err
			}

			return next(ctx)
		}
	}
}

func before(ctx echo.Context, set *optionSet) error {
	if !set.ShouldAuth(ctx) {
		return nil
	}

	authHeader := ctx.Request().Header.Get(rkechointer.RpcAuthorizationHeaderKey)
	apiKeyHeader := ctx.Request().Header.Get(rkechointer.RpcApiKeyHeaderKey)

	if len(authHeader) > 0 {
		// Contains auth header
		// Basic auth type
		tokens := strings.SplitN(authHeader, " ", 2)
		if len(tokens) != 2 {
			resp := rkerror.New(
				rkerror.WithHttpCode(http.StatusUnauthorized),
				rkerror.WithMessage("Invalid Basic Auth format"))
			ctx.JSON(http.StatusUnauthorized, resp)
			return resp.Err
		}
		if !set.Authorized(tokens[0], tokens[1]) {
			if tokens[0] == typeBasic {
				ctx.Response().Header().Set("WWW-Authenticate", fmt.Sprintf(`%s realm="%s"`, typeBasic, set.BasicRealm))
			}

			resp := rkerror.New(
				rkerror.WithHttpCode(http.StatusUnauthorized),
				rkerror.WithMessage("Invalid credential"))

			ctx.JSON(http.StatusUnauthorized, resp)
			return resp.Err
		}
	} else if len(apiKeyHeader) > 0 {
		// Contains api key
		if !set.Authorized(typeApiKey, apiKeyHeader) {
			resp := rkerror.New(
				rkerror.WithHttpCode(http.StatusUnauthorized),
				rkerror.WithMessage("Invalid X-API-Key"))

			ctx.JSON(http.StatusUnauthorized, resp)

			return resp.Err
		}
	} else {
		authHeaders := []string{}
		if len(set.BasicAccounts) > 0 {
			ctx.Response().Header().Set("WWW-Authenticate", fmt.Sprintf(`%s realm="%s"`, typeBasic, set.BasicRealm))
			authHeaders = append(authHeaders, "Basic Auth")
		}
		if len(set.ApiKey) > 0 {
			authHeaders = append(authHeaders, "X-API-Key")
		}

		errMsg := fmt.Sprintf("Missing authorization, provide one of bellow auth header:[%s]", strings.Join(authHeaders, ","))

		resp := rkerror.New(
			rkerror.WithHttpCode(http.StatusUnauthorized),
			rkerror.WithMessage(errMsg))

		ctx.JSON(http.StatusUnauthorized, resp)

		return resp.Err
	}

	return nil
}

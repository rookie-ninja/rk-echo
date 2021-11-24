// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.
package rkechocors

import (
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

const originHeaderValue = "http://ut-origin"

var userHandler = func(ctx echo.Context) error {
	return ctx.String(http.StatusOK, "")
}

func TestInterceptor(t *testing.T) {
	defer assertNotPanic(t)

	// with skipper
	ctx := newCtx(http.MethodGet)
	handler := Interceptor(WithSkipper(func(context echo.Context) bool {
		return true
	}))
	assert.Nil(t, handler(userHandler)(ctx))
	assert.Equal(t, http.StatusOK, ctx.Response().Status)

	// with empty option, all request will be passed
	ctx = newCtx(http.MethodGet)
	handler = Interceptor()
	assert.Nil(t, handler(userHandler)(ctx))
	assert.Equal(t, http.StatusOK, ctx.Response().Status)

	// match 1.1
	ctx = newCtx(http.MethodGet)
	ctx.Request().Header.Del(echo.HeaderOrigin)
	handler = Interceptor()
	assert.Nil(t, handler(userHandler)(ctx))
	assert.Equal(t, http.StatusOK, ctx.Response().Status)

	// match 1.2
	ctx = newCtx(http.MethodOptions)
	ctx.Request().Header.Del(echo.HeaderOrigin)
	handler = Interceptor()
	assert.Nil(t, handler(userHandler)(ctx))
	assert.Equal(t, http.StatusNoContent, ctx.Response().Status)

	// match 2
	ctx = newCtx(http.MethodOptions)
	handler = Interceptor(WithAllowOrigins("http://do-not-pass-through"))
	assert.Nil(t, handler(userHandler)(ctx))
	assert.Equal(t, http.StatusNoContent, ctx.Response().Status)

	// match 3
	ctx = newCtx(http.MethodGet)
	handler = Interceptor()
	assert.Nil(t, handler(userHandler)(ctx))
	assert.Equal(t, http.StatusOK, ctx.Response().Status)
	assert.Equal(t, originHeaderValue, ctx.Response().Header().Get(echo.HeaderAccessControlAllowOrigin))

	// match 3.1
	ctx = newCtx(http.MethodGet)
	handler = Interceptor(WithAllowCredentials(true))
	assert.Nil(t, handler(userHandler)(ctx))
	assert.Equal(t, http.StatusOK, ctx.Response().Status)
	assert.Equal(t, originHeaderValue, ctx.Response().Header().Get(echo.HeaderAccessControlAllowOrigin))
	assert.Equal(t, "true", ctx.Response().Header().Get(echo.HeaderAccessControlAllowCredentials))

	// match 3.2
	ctx = newCtx(http.MethodGet)
	handler = Interceptor(
		WithAllowCredentials(true),
		WithExposeHeaders("expose"))
	assert.Nil(t, handler(userHandler)(ctx))
	assert.Equal(t, http.StatusOK, ctx.Response().Status)
	assert.Equal(t, originHeaderValue, ctx.Response().Header().Get(echo.HeaderAccessControlAllowOrigin))
	assert.Equal(t, "true", ctx.Response().Header().Get(echo.HeaderAccessControlAllowCredentials))
	assert.Equal(t, "expose", ctx.Response().Header().Get(echo.HeaderAccessControlExposeHeaders))

	// match 4
	ctx = newCtx(http.MethodOptions)
	handler = Interceptor()
	assert.Nil(t, handler(userHandler)(ctx))
	assert.Equal(t, http.StatusNoContent, ctx.Response().Status)
	assert.Equal(t, []string{
		echo.HeaderAccessControlRequestMethod,
		echo.HeaderAccessControlRequestHeaders,
	}, ctx.Response().Header().Values(echo.HeaderVary))
	assert.Equal(t, originHeaderValue, ctx.Response().Header().Get(echo.HeaderAccessControlAllowOrigin))
	assert.NotEmpty(t, ctx.Response().Header().Get(echo.HeaderAccessControlAllowMethods))

	// match 4.1
	ctx = newCtx(http.MethodOptions)
	handler = Interceptor(WithAllowCredentials(true))
	assert.Nil(t, handler(userHandler)(ctx))
	assert.Equal(t, http.StatusNoContent, ctx.Response().Status)
	assert.Equal(t, []string{
		echo.HeaderAccessControlRequestMethod,
		echo.HeaderAccessControlRequestHeaders,
	}, ctx.Response().Header().Values(echo.HeaderVary))
	assert.Equal(t, originHeaderValue, ctx.Response().Header().Get(echo.HeaderAccessControlAllowOrigin))
	assert.NotEmpty(t, ctx.Response().Header().Get(echo.HeaderAccessControlAllowMethods))
	assert.Equal(t, "true", ctx.Response().Header().Get(echo.HeaderAccessControlAllowCredentials))

	// match 4.2
	ctx = newCtx(http.MethodOptions)
	handler = Interceptor(WithAllowHeaders("ut-header"))
	assert.Nil(t, handler(userHandler)(ctx))
	assert.Equal(t, http.StatusNoContent, ctx.Response().Status)
	assert.Equal(t, []string{
		echo.HeaderAccessControlRequestMethod,
		echo.HeaderAccessControlRequestHeaders,
	}, ctx.Response().Header().Values(echo.HeaderVary))
	assert.Equal(t, originHeaderValue, ctx.Response().Header().Get(echo.HeaderAccessControlAllowOrigin))
	assert.NotEmpty(t, ctx.Response().Header().Get(echo.HeaderAccessControlAllowMethods))
	assert.Equal(t, "ut-header", ctx.Response().Header().Get(echo.HeaderAccessControlAllowHeaders))

	// match 4.3
	ctx = newCtx(http.MethodOptions)
	handler = Interceptor(WithMaxAge(1))
	assert.Nil(t, handler(userHandler)(ctx))
	assert.Equal(t, http.StatusNoContent, ctx.Response().Status)
	assert.Equal(t, []string{
		echo.HeaderAccessControlRequestMethod,
		echo.HeaderAccessControlRequestHeaders,
	}, ctx.Response().Header().Values(echo.HeaderVary))
	assert.Equal(t, originHeaderValue, ctx.Response().Header().Get(echo.HeaderAccessControlAllowOrigin))
	assert.NotEmpty(t, ctx.Response().Header().Get(echo.HeaderAccessControlAllowMethods))
	assert.Equal(t, "1", ctx.Response().Header().Get(echo.HeaderAccessControlMaxAge))
}

func newCtx(method string) echo.Context {
	req := httptest.NewRequest(method, "/ut-path", nil)
	req.Header = http.Header{}
	req.Header.Set(echo.HeaderOrigin, originHeaderValue)

	resp := httptest.NewRecorder()
	return echo.New().NewContext(req, resp)
}

func assertNotPanic(t *testing.T) {
	if r := recover(); r != nil {
		// Expect panic to be called with non nil error
		assert.True(t, false)
	} else {
		// This should never be called in case of a bug
		assert.True(t, true)
	}
}

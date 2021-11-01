// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkecholimit

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestInterceptor_WithoutOptions(t *testing.T) {
	defer assertNotPanic(t)

	handler := Interceptor()

	ctx := newCtx()
	ctx.Request().URL.Path = "/ut-path"

	f := handler(defaultMiddlewareFunc)

	assert.Nil(t, f(ctx))
}

func TestInterceptor_WithTokenBucket(t *testing.T) {
	defer assertNotPanic(t)

	handler := Interceptor(
		WithAlgorithm(TokenBucket),
		WithReqPerSec(1),
		WithReqPerSecByPath("ut-path", 1))

	ctx := newCtx()
	ctx.Request().URL.Path = "/ut-path"

	f := handler(defaultMiddlewareFunc)

	assert.Nil(t, f(ctx))
}

func TestInterceptor_WithLeakyBucket(t *testing.T) {
	defer assertNotPanic(t)

	handler := Interceptor(
		WithAlgorithm(LeakyBucket),
		WithReqPerSec(1),
		WithReqPerSecByPath("ut-path", 1))

	ctx := newCtx()
	ctx.Request().URL.Path = "/ut-path"

	f := handler(defaultMiddlewareFunc)

	assert.Nil(t, f(ctx))
}

func TestInterceptor_WithUserLimiter(t *testing.T) {
	defer assertNotPanic(t)

	handler := Interceptor(
		WithGlobalLimiter(func(ctx echo.Context) error {
			return fmt.Errorf("ut-error")
		}),
		WithLimiterByPath("/ut-path", func(ctx echo.Context) error {
			return fmt.Errorf("ut-error")
		}))

	ctx := newCtx()
	ctx.Request().URL.Path = "/ut-path"

	f := handler(defaultMiddlewareFunc)

	assert.NotNil(t, f(ctx))

	assert.Equal(t, http.StatusTooManyRequests, ctx.Response().Status)
}

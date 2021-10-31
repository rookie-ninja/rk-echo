// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

// Package rkechometa is a middleware of echo framework for adding metadata in RPC response
package rkechometa

import (
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

var defaultMiddlewareFunc = func(context echo.Context) error {
	return nil
}

func newCtx() echo.Context {
	return echo.New().NewContext(
		httptest.NewRequest(http.MethodGet, "/ut-path", nil),
		httptest.NewRecorder())
}

func TestInterceptor(t *testing.T) {
	handler := Interceptor(
		WithEntryNameAndType("ut-entry", "ut-type"))

	ctx := newCtx()

	f := handler(defaultMiddlewareFunc)

	assert.Nil(t, f(ctx))

	assert.NotEmpty(t, ctx.Response().Writer.Header().Get("X-RK-App-Name"))
	assert.Empty(t, ctx.Response().Writer.Header().Get("X-RK-App-Version"))
	assert.NotEmpty(t, ctx.Response().Writer.Header().Get("X-RK-App-Unix-Time"))
	assert.NotEmpty(t, ctx.Response().Writer.Header().Get("X-RK-Received-Time"))
}

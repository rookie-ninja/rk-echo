// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkecholog

import (
	"github.com/labstack/echo/v4"
	"github.com/rookie-ninja/rk-echo/interceptor/context"
	"github.com/rookie-ninja/rk-entry/entry"
	"github.com/rookie-ninja/rk-query"
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

func TestInterceptor_WithShouldNotLog(t *testing.T) {
	defer assertNotPanic(t)

	ctx := newCtx()
	ctx.Request().URL.Path = "/rk/v1/assets"

	handler := Interceptor(
		WithEntryNameAndType("ut-entry", "ut-type"),
		WithZapLoggerEntry(rkentry.NoopZapLoggerEntry()),
		WithEventLoggerEntry(rkentry.NoopEventLoggerEntry()))

	f := handler(defaultMiddlewareFunc)

	assert.Nil(t, f(ctx))
}

func TestInterceptor_HappyCase(t *testing.T) {
	defer assertNotPanic(t)

	ctx := newCtx()

	handler := Interceptor(
		WithEntryNameAndType("ut-entry", "ut-type"),
		WithZapLoggerEntry(rkentry.NoopZapLoggerEntry()),
		WithEventLoggerEntry(rkentry.NoopEventLoggerEntry()))

	ctx.Response().Writer.Header().Set(rkechoctx.RequestIdKey, "ut-request-id")
	ctx.Response().Writer.Header().Set(rkechoctx.TraceIdKey, "ut-trace-id")

	f := handler(defaultMiddlewareFunc)

	assert.Nil(t, f(ctx))

	event := rkechoctx.GetEvent(ctx)

	assert.NotEmpty(t, event.GetRemoteAddr())
	assert.NotEmpty(t, event.ListPayloads())
	assert.NotEmpty(t, event.GetOperation())
	assert.NotEmpty(t, event.GetRequestId())
	assert.NotEmpty(t, event.GetTraceId())
	assert.NotEmpty(t, event.GetResCode())
	assert.Equal(t, rkquery.Ended, event.GetEventStatus())
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

// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkechotrace

import (
	"bytes"
	"context"
	"github.com/labstack/echo/v4"
	"github.com/rookie-ninja/rk-entry/middleware"
	"github.com/rookie-ninja/rk-entry/middleware/tracing"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/trace"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

var userHandler = func(ctx echo.Context) error {
	return ctx.String(http.StatusOK, "")
}

func TestInterceptor(t *testing.T) {
	defer assertNotPanic(t)

	beforeCtx := rkmidtrace.NewBeforeCtx()
	afterCtx := rkmidtrace.NewAfterCtx()
	mock := rkmidtrace.NewOptionSetMock(beforeCtx, afterCtx, nil, nil, nil)
	beforeCtx.Output.NewCtx = context.TODO()

	// case 1: with error response
	inter := Interceptor(rkmidtrace.WithMockOptionSet(mock))
	ctx, _ := newCtx()

	inter(userHandler)(ctx)

	// case 2: happy case
	noopTracerProvider := trace.NewNoopTracerProvider()
	_, span := noopTracerProvider.Tracer("rk-trace-noop").Start(ctx.Request().Context(), "noop-span")
	beforeCtx.Output.Span = span

	inter(userHandler)(ctx)

	spanFromCtx := ctx.Get(rkmid.SpanKey.String())
	assert.Equal(t, span, spanFromCtx)
}

func newCtx() (echo.Context, *httptest.ResponseRecorder) {
	var buf bytes.Buffer
	req := httptest.NewRequest(http.MethodGet, "/ut-path", &buf)
	resp := httptest.NewRecorder()
	return echo.New().NewContext(req, resp), resp
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

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

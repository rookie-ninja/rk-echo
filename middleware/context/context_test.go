// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkechoctx

import (
	"context"
	"github.com/golang-jwt/jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/rookie-ninja/rk-entry/v2/middleware"
	"github.com/rookie-ninja/rk-logger"
	"github.com/rookie-ninja/rk-query"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newCtx() echo.Context {
	return echo.New().NewContext(
		httptest.NewRequest(http.MethodGet, "/ut-path", nil),
		httptest.NewRecorder())
}

func TestGetIncomingHeaders(t *testing.T) {
	ctx := newCtx()
	ctx.Request().Header.Set("ut-key", "ut-value")

	assert.Len(t, GetIncomingHeaders(ctx), 1)
	assert.Equal(t, "ut-value", GetIncomingHeaders(ctx).Get("ut-key"))
}

func TestGormCtx(t *testing.T) {
	ctx := newCtx()
	assert.NotNil(t, GormCtx(ctx))
}

func TestAddHeaderToClient(t *testing.T) {
	defer assertNotPanic(t)

	ctx := newCtx()

	// With nil context
	AddHeaderToClient(nil, "", "")

	// With nil writer
	AddHeaderToClient(ctx, "", "")

	// Happy case
	AddHeaderToClient(ctx, "ut-key", "ut-value")
	assert.Equal(t, "ut-value", ctx.Response().Header().Get("ut-key"))
}

func TestSetHeaderToClient(t *testing.T) {
	defer assertNotPanic(t)

	ctx := newCtx()

	// With nil context
	SetHeaderToClient(nil, "", "")

	// With nil writer
	SetHeaderToClient(ctx, "", "")

	// Happy case
	SetHeaderToClient(ctx, "ut-key", "ut-value")
	assert.Equal(t, "ut-value", ctx.Response().Header().Get("ut-key"))
}

func TestGetEvent(t *testing.T) {
	// With nil context
	assert.Equal(t, noopEvent, GetEvent(nil))

	// With no event in context
	ctx := newCtx()
	assert.Equal(t, noopEvent, GetEvent(ctx))

	// Happy case
	event := rkquery.NewEventFactory().CreateEventNoop()
	ctx.Set(rkmid.EventKey.String(), event)
	assert.Equal(t, event, GetEvent(ctx))
}

func TestGetLogger(t *testing.T) {
	// With nil context
	assert.Equal(t, rklogger.NoopLogger, GetLogger(nil))

	ctx := newCtx()

	// With no logger in context
	assert.Equal(t, rklogger.NoopLogger, GetLogger(ctx))

	// Happy case
	// Add request id and trace id
	ctx.Response().Writer.Header().Set(rkmid.HeaderRequestId, "ut-request-id")
	ctx.Response().Writer.Header().Set(rkmid.HeaderTraceId, "ut-trace-id")
	ctx.Set(rkmid.LoggerKey.String(), rklogger.NoopLogger)

	assert.Equal(t, rklogger.NoopLogger, GetLogger(ctx))
}

func TestGetRequestId(t *testing.T) {
	// With nil context
	assert.Empty(t, GetRequestId(nil))

	ctx := newCtx()

	// With no requestId in context
	assert.Empty(t, GetRequestId(ctx))

	// Happy case
	ctx.Response().Writer.Header().Set(rkmid.HeaderRequestId, "ut-request-id")
	assert.Equal(t, "ut-request-id", GetRequestId(ctx))
}

func TestGetTraceId(t *testing.T) {
	// With nil context
	assert.Empty(t, GetTraceId(nil))

	ctx := newCtx()

	// With no traceId in context
	assert.Empty(t, GetTraceId(ctx))

	// Happy case
	ctx.Response().Writer.Header().Set(rkmid.HeaderTraceId, "ut-trace-id")
	assert.Equal(t, "ut-trace-id", GetTraceId(ctx))
}

func TestGetEntryName(t *testing.T) {
	// With nil context
	assert.Empty(t, GetEntryName(nil))

	ctx := newCtx()

	// With no entry name in context
	assert.Empty(t, GetEntryName(ctx))

	// Happy case
	ctx.Set(rkmid.EntryNameKey.String(), "ut-entry-name")
	assert.Equal(t, "ut-entry-name", GetEntryName(ctx))
}

func TestGetTraceSpan(t *testing.T) {
	ctx := newCtx()
	ctx.SetRequest(ctx.Request().WithContext(context.TODO()))

	// With nil context
	assert.NotNil(t, GetTraceSpan(nil))

	// With no span in context
	assert.NotNil(t, GetTraceSpan(ctx))

	// Happy case
	_, span := noopTracerProvider.Tracer("ut-trace").Start(ctx.Request().Context(), "noop-span")
	ctx.Set(rkmid.SpanKey.String(), span)
	assert.Equal(t, span, GetTraceSpan(ctx))
}

func TestGetTracer(t *testing.T) {
	ctx := newCtx()
	ctx.SetRequest(ctx.Request().WithContext(context.TODO()))

	// With nil context
	assert.NotNil(t, GetTracer(nil))

	// With no tracer in context
	assert.NotNil(t, GetTracer(ctx))

	// Happy case
	tracer := noopTracerProvider.Tracer("ut-trace")
	ctx.Set(rkmid.TracerKey.String(), tracer)
	assert.Equal(t, tracer, GetTracer(ctx))
}

func TestGetTracerProvider(t *testing.T) {
	ctx := newCtx()
	ctx.SetRequest(ctx.Request().WithContext(context.TODO()))

	// With nil context
	assert.NotNil(t, GetTracerProvider(nil))

	// With no tracer provider in context
	assert.NotNil(t, GetTracerProvider(ctx))

	// Happy case
	provider := trace.NewNoopTracerProvider()
	ctx.Set(rkmid.TracerProviderKey.String(), provider)
	assert.Equal(t, provider, GetTracerProvider(ctx))
}

func TestGetTracerPropagator(t *testing.T) {
	ctx := newCtx()
	ctx.SetRequest(ctx.Request().WithContext(context.TODO()))

	// With nil context
	assert.Nil(t, GetTracerPropagator(nil))

	// With no tracer propagator in context
	assert.Nil(t, GetTracerPropagator(ctx))

	// Happy case
	prop := propagation.NewCompositeTextMapPropagator()
	ctx.Set(rkmid.PropagatorKey.String(), prop)
	assert.Equal(t, prop, GetTracerPropagator(ctx))
}

func TestInjectSpanToHttpRequest(t *testing.T) {
	defer assertNotPanic(t)

	// With nil context and request
	InjectSpanToHttpRequest(nil, nil)

	// Happy case
	ctx := newCtx()
	ctx.SetRequest(ctx.Request().WithContext(context.TODO()))

	prop := propagation.NewCompositeTextMapPropagator()
	ctx.Set(rkmid.PropagatorKey.String(), prop)
	InjectSpanToHttpRequest(ctx, &http.Request{
		Header: http.Header{},
	})
}

func TestNewTraceSpan(t *testing.T) {
	ctx := newCtx()
	ctx.SetRequest(ctx.Request().WithContext(context.TODO()))

	assert.NotNil(t, NewTraceSpan(ctx, "ut-span"))
}

func TestEndTraceSpan(t *testing.T) {
	defer assertNotPanic(t)

	ctx := newCtx()
	ctx.SetRequest(ctx.Request().WithContext(context.TODO()))

	// With success
	span := GetTraceSpan(ctx)
	EndTraceSpan(ctx, span, true)

	// With failure
	span = GetTraceSpan(ctx)
	EndTraceSpan(ctx, span, false)
}

func TestGetJwtToken(t *testing.T) {
	defer assertNotPanic(t)

	// with nil
	assert.Nil(t, GetJwtToken(nil))

	// With failure
	ctx := newCtx()
	assert.Nil(t, GetJwtToken(ctx))

	// With success
	ctx.Set(rkmid.JwtTokenKey.String(), &jwt.Token{})
	assert.NotNil(t, GetJwtToken(ctx))
}

func TestGetCsrfToken(t *testing.T) {
	defer assertNotPanic(t)

	// with nil
	assert.Empty(t, GetCsrfToken(nil))

	// With failure
	ctx := newCtx()
	assert.Empty(t, GetCsrfToken(ctx))

	// With success
	ctx.Set(rkmid.CsrfTokenKey.String(), "value")
	assert.Equal(t, "value", GetCsrfToken(ctx))
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

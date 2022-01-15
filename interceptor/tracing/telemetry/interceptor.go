// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

// Package rkechotrace is aa middleware of echo framework for recording trace info of RPC
package rkechotrace

import (
	"github.com/labstack/echo/v4"
	"github.com/rookie-ninja/rk-echo/interceptor/context"
	"github.com/rookie-ninja/rk-entry/middleware"
	"github.com/rookie-ninja/rk-entry/middleware/tracing"
)

// Interceptor create a interceptor with opentelemetry.
func Interceptor(opts ...rkmidtrace.Option) echo.MiddlewareFunc {
	set := rkmidtrace.NewOptionSet(opts...)

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			ctx.Set(rkmid.EntryNameKey.String(), set.GetEntryName())
			ctx.Set(rkmid.TracerKey.String(), set.GetTracer())
			ctx.Set(rkmid.TracerProviderKey.String(), set.GetProvider())
			ctx.Set(rkmid.PropagatorKey.String(), set.GetPropagator())

			beforeCtx := set.BeforeCtx(ctx.Request(), false)
			set.Before(beforeCtx)

			// create request with new context
			ctx.SetRequest(ctx.Request().WithContext(beforeCtx.Output.NewCtx))

			// add to context
			if beforeCtx.Output.Span != nil {
				traceId := beforeCtx.Output.Span.SpanContext().TraceID().String()
				rkechoctx.GetEvent(ctx).SetTraceId(traceId)
				ctx.Response().Header().Set(rkmid.HeaderTraceId, traceId)
				ctx.Set(rkmid.SpanKey.String(), beforeCtx.Output.Span)
			}

			err := next(ctx)

			afterCtx := set.AfterCtx(ctx.Response().Status, "")
			set.After(beforeCtx, afterCtx)

			return err
		}
	}
}

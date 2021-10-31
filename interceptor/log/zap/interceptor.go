// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

// Package rkecholog is a middleware for echo framework for logging RPC.
package rkecholog

import (
	"github.com/labstack/echo/v4"
	"github.com/rookie-ninja/rk-echo/interceptor"
	"github.com/rookie-ninja/rk-echo/interceptor/context"
	"github.com/rookie-ninja/rk-entry/entry"
	"github.com/rookie-ninja/rk-query"
	"go.uber.org/zap"
	"strconv"
	"time"
)

// Interceptor returns a gin.HandlerFunc (middleware) that logs requests using uber-go/zap.
func Interceptor(opts ...Option) echo.MiddlewareFunc {
	set := newOptionSet(opts...)

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			ctx.Set(rkechointer.RpcEntryNameKey, set.EntryName)

			before(ctx, set)

			err := next(ctx)

			after(ctx)

			return err
		}
	}
}

func before(ctx echo.Context, set *optionSet) {
	var event rkquery.Event
	if rkechointer.ShouldLog(ctx) {
		event = set.eventLoggerEntry.GetEventFactory().CreateEvent(
			rkquery.WithZapLogger(set.eventLoggerOverride),
			rkquery.WithEncoding(set.eventLoggerEncoding),
			rkquery.WithAppName(rkentry.GlobalAppCtx.GetAppInfoEntry().AppName),
			rkquery.WithAppVersion(rkentry.GlobalAppCtx.GetAppInfoEntry().Version),
			rkquery.WithEntryName(set.EntryName),
			rkquery.WithEntryType(set.EntryType))
	} else {
		event = set.eventLoggerEntry.GetEventFactory().CreateEventNoop()
	}

	event.SetStartTime(time.Now())

	remoteIp, remotePort := rkechointer.GetRemoteAddressSet(ctx)
	// handle remote address
	event.SetRemoteAddr(remoteIp + ":" + remotePort)

	payloads := []zap.Field{
		zap.String("apiPath", ctx.Request().URL.Path),
		zap.String("apiMethod", ctx.Request().Method),
		zap.String("apiQuery", ctx.Request().URL.RawQuery),
		zap.String("apiProtocol", ctx.Request().Proto),
		zap.String("userAgent", ctx.Request().UserAgent()),
	}

	// handle payloads
	event.AddPayloads(payloads...)

	// handle operation
	event.SetOperation(ctx.Request().URL.Path)

	ctx.Set(rkechointer.RpcEventKey, event)
	ctx.Set(rkechointer.RpcLoggerKey, set.ZapLogger)
}

func after(ctx echo.Context) {
	event := rkechoctx.GetEvent(ctx)

	if requestId := rkechoctx.GetRequestId(ctx); len(requestId) > 0 {
		event.SetEventId(requestId)
		event.SetRequestId(requestId)
	}

	if traceId := rkechoctx.GetTraceId(ctx); len(traceId) > 0 {
		event.SetTraceId(traceId)
	}

	event.SetResCode(strconv.Itoa(ctx.Response().Status))
	event.SetEndTime(time.Now())
	event.Finish()
}

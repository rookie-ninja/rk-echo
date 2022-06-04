// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

// Package rkechotimeout is a middleware of echo framework for timing out request in RPC response
package rkechotimeout

import (
	"bytes"
	"github.com/labstack/echo/v4"
	"github.com/rookie-ninja/rk-echo/middleware/context"
	"github.com/rookie-ninja/rk-entry/v2/middleware"
	"github.com/rookie-ninja/rk-entry/v2/middleware/timeout"
	"net/http"
)

// Middleware Add timeout interceptors.
func Middleware(opts ...rkmidtimeout.Option) echo.MiddlewareFunc {
	set := rkmidtimeout.NewOptionSet(opts...)

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			ctx.Set(rkmid.EntryNameKey.String(), set.GetEntryName())

			// case 1: return to user if error occur
			beforeCtx := set.BeforeCtx(ctx.Request(), rkechoctx.GetEvent(ctx))
			toCtx := &timeoutCtx{
				echoCtx:  ctx,
				nextFunc: next,
				before:   beforeCtx,
			}
			// assign handlers
			beforeCtx.Input.InitHandler = initHandler(toCtx)
			beforeCtx.Input.NextHandler = nextHandler(toCtx)
			beforeCtx.Input.PanicHandler = panicHandler(toCtx)
			beforeCtx.Input.FinishHandler = finishHandler(toCtx)
			beforeCtx.Input.TimeoutHandler = timeoutHandler(toCtx)

			// call before
			set.Before(beforeCtx)

			beforeCtx.Output.WaitFunc()

			return toCtx.nextError
		}
	}
}

type timeoutCtx struct {
	bufPool   *bufferPool
	buffer    *bytes.Buffer
	oldW      http.ResponseWriter
	newW      *writer
	echoCtx   echo.Context
	before    *rkmidtimeout.BeforeCtx
	nextFunc  echo.HandlerFunc
	nextError error
}

func timeoutHandler(ctx *timeoutCtx) func() {
	return func() {
		ctx.newW.mu.Lock()
		defer ctx.newW.mu.Unlock()

		ctx.newW.timeout = true

		// free buffer
		ctx.newW.FreeBuffer()
		ctx.bufPool.Put(ctx.buffer)

		// switch to original writer
		ctx.echoCtx.Response().Writer = ctx.oldW

		// write timed out response
		ctx.echoCtx.JSON(ctx.before.Output.TimeoutErrResp.Code(), ctx.before.Output.TimeoutErrResp)

		// switch back to new writer since user code may still want to write to it.
		// Panic may occur if we ignore this step.
		ctx.echoCtx.Response().Writer = ctx.newW
	}
}

func finishHandler(ctx *timeoutCtx) func() {
	return func() {
		ctx.newW.mu.Lock()
		defer ctx.newW.mu.Unlock()

		// copy headers and code
		dst := ctx.newW.ResponseWriter.Header()
		for k, vv := range ctx.newW.Header() {
			dst[k] = vv
		}
		ctx.newW.ResponseWriter.WriteHeader(ctx.newW.code)

		// copy contents
		if _, err := ctx.newW.ResponseWriter.Write(ctx.buffer.Bytes()); err != nil {
			panic(err)
		}

		// free buffer
		ctx.newW.FreeBuffer()
		ctx.bufPool.Put(ctx.buffer)
	}
}

func panicHandler(ctx *timeoutCtx) func() {
	return func() {
		ctx.newW.FreeBuffer()
		ctx.echoCtx.Response().Writer = ctx.oldW
	}
}

func nextHandler(ctx *timeoutCtx) func() {
	return func() {
		ctx.nextError = ctx.nextFunc(ctx.echoCtx)
	}
}

func initHandler(ctx *timeoutCtx) func() {
	// create a buffer pool and new writer
	// Why?
	//
	// We may face the case that request timed out while user code is writing to response writer.
	// So, we create a new writer with mutex lock and ignore contents user code writers if timed out .
	ctx.bufPool = &bufferPool{}
	ctx.buffer = ctx.bufPool.Get()
	ctx.oldW = ctx.echoCtx.Response().Writer
	ctx.newW = newWriter(ctx.oldW, ctx.buffer)

	return func() {
		ctx.echoCtx.Response().Writer = ctx.newW
	}
}

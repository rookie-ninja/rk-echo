// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

// Package rkechometrics is a middleware for echo framework which record prometheus metrics for RPC
package rkechometrics

import (
	"github.com/labstack/echo/v4"
	"github.com/rookie-ninja/rk-echo/interceptor"
	"time"
)

// Interceptor create a new prometheus metrics interceptor with options.
func Interceptor(opts ...Option) echo.MiddlewareFunc {
	set := newOptionSet(opts...)

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			ctx.Set(rkechointer.RpcEntryNameKey, set.EntryName)

			// start timer
			startTime := time.Now()

			err := next(ctx)

			// end timer
			elapsed := time.Now().Sub(startTime)

			// ignoring /rk/v1/assets, /rk/v1/tv and /sw/ path while logging since these are internal APIs.
			if rkechointer.ShouldLog(ctx) {
				if durationMetrics := GetServerDurationMetrics(ctx); durationMetrics != nil {
					durationMetrics.Observe(float64(elapsed.Nanoseconds()))
				}

				if resCodeMetrics := GetServerResCodeMetrics(ctx); resCodeMetrics != nil {
					resCodeMetrics.Inc()
				}
			}

			return err
		}
	}
}

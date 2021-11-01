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

func TestWithEntryNameAndType(t *testing.T) {
	defer assertNotPanic(t)

	set := newOptionSet(
		WithEntryNameAndType("ut-entry", "ut-type"))

	assert.Equal(t, "ut-entry", set.EntryName)
	assert.Equal(t, "ut-type", set.EntryType)
	assert.Len(t, set.limiter, 1)

	// Should be noop limiter
	ctx := newCtx()
	ctx.Request().URL.Path = "/ut-path"

	set.getLimiter("")(ctx)
}

func TestWithReqPerSec(t *testing.T) {
	// With non-zero
	set := newOptionSet(
		WithReqPerSec(1))

	assert.Equal(t, 1, set.reqPerSec)
	assert.Len(t, set.limiter, 1)

	// Should be token based limiter
	ctx := newCtx()
	ctx.Request().URL.Path = "/ut-path"
	set.getLimiter("")(ctx)

	// With zero
	set = newOptionSet(
		WithReqPerSec(0))

	assert.Equal(t, 0, set.reqPerSec)
	assert.Len(t, set.limiter, 1)

	// should be zero rate limiter which returns error
	ctx = newCtx()
	ctx.Request().URL.Path = "/ut-path"
	assert.NotNil(t, set.getLimiter("")(ctx))
}

func TestWithReqPerSecByPath(t *testing.T) {
	// with non-zero
	set := newOptionSet(
		WithReqPerSecByPath("/ut-path", 1))

	assert.Equal(t, 1, set.reqPerSecByPath["/ut-path"])
	assert.NotNil(t, set.limiter["/ut-path"])

	// Should be token based limiter
	ctx := newCtx()
	ctx.Request().URL.Path = "/ut-path"
	set.getLimiter("/ut-path")(ctx)

	// With zero
	set = newOptionSet(
		WithReqPerSecByPath("/ut-path", 0))

	assert.Equal(t, 0, set.reqPerSecByPath["/ut-path"])
	assert.NotNil(t, set.limiter["/ut-path"])

	// should be zero rate limiter which returns error
	ctx = newCtx()
	ctx.Request().URL.Path = "/ut-path"
	assert.NotNil(t, set.getLimiter("/ut-path")(ctx))
}

func TestWithAlgorithm(t *testing.T) {
	defer assertNotPanic(t)

	// With invalid algorithm
	set := newOptionSet(
		WithAlgorithm("invalid-algo"))

	// should be noop limiter
	assert.Len(t, set.limiter, 1)
	set.getLimiter("")

	// With leaky bucket non zero
	set = newOptionSet(
		WithAlgorithm(LeakyBucket),
		WithReqPerSec(1),
		WithReqPerSecByPath("ut-method", 1))

	// should be leaky bucket
	assert.Len(t, set.limiter, 2)
	ctx := newCtx()
	ctx.Request().URL.Path = "/ut-path"
	set.getLimiter("")(ctx)
	set.getLimiter("ut-path")(ctx)
}

func TestWithGlobalLimiter(t *testing.T) {
	set := newOptionSet(
		WithGlobalLimiter(func(ctx echo.Context) error {
			return fmt.Errorf("ut error")
		}))

	assert.Len(t, set.limiter, 1)
	ctx := newCtx()
	ctx.Request().URL.Path = "/ut-path"
	assert.NotNil(t, set.getLimiter("")(ctx))
}

func TestWithLimiterByPath(t *testing.T) {
	set := newOptionSet(
		WithLimiterByPath("/ut-path", func(ctx echo.Context) error {
			return fmt.Errorf("ut error")
		}))

	assert.Len(t, set.limiter, 2)

	ctx := newCtx()
	ctx.Request().URL.Path = "/ut-path"
	assert.NotNil(t, set.getLimiter("/ut-path")(ctx))
}

func TestOptionSet_Wait(t *testing.T) {
	defer assertNotPanic(t)

	// With user limiter
	set := newOptionSet(
		WithGlobalLimiter(func(echo.Context) error {
			return nil
		}))

	ctx := newCtx()
	ctx.Request().URL.Path = "/ut-path"
	set.Wait(ctx)

	// With token bucket and global limiter
	set = newOptionSet(
		WithAlgorithm(TokenBucket))

	set.Wait(ctx)

	// With token bucket and limiter by method
	set = newOptionSet(
		WithAlgorithm(TokenBucket),
		WithReqPerSecByPath("/ut-path", 100))

	set.Wait(ctx)

	// With leaky bucket and global limiter
	set = newOptionSet(
		WithAlgorithm(LeakyBucket))

	set.Wait(ctx)

	// With leaky bucket and limiter by method
	set = newOptionSet(
		WithAlgorithm(LeakyBucket),
		WithReqPerSecByPath("/ut-path", 100))

	set.Wait(ctx)

	// Without any configuration
	set = newOptionSet()
	set.Wait(ctx)
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

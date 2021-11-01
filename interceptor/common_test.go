// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkechointer

import (
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func newCtx() echo.Context {
	return echo.New().NewContext(
		httptest.NewRequest(http.MethodGet, "/ut-path", nil),
		httptest.NewRecorder())
}

func TestGetRemoteAddressSet(t *testing.T) {
	// With nil context
	ip, port := GetRemoteAddressSet(nil)
	assert.Equal(t, "0.0.0.0", ip)
	assert.Equal(t, "0", port)

	// With nil Request in context
	ctx := newCtx()
	assert.Equal(t, "0.0.0.0", ip)
	assert.Equal(t, "0", port)

	// With x-forwarded-for equals to ::1
	ctx = newCtx()
	ctx.SetRequest(&http.Request{
		RemoteAddr: "1.1.1.1:1",
		Header:     http.Header{},
	})
	ctx.Request().Header.Set("x-forwarded-for", "::1")
	ip, port = GetRemoteAddressSet(ctx)

	assert.Equal(t, "localhost", ip)
	assert.Equal(t, "1", port)

	// Happy case
	ctx = newCtx()
	ctx.SetRequest(&http.Request{
		RemoteAddr: "1.1.1.1:1",
		Header:     http.Header{},
	})
	ip, port = GetRemoteAddressSet(ctx)

	assert.Equal(t, "1.1.1.1", ip)
	assert.Equal(t, "1", port)
}

func TestShouldLog(t *testing.T) {
	// With nil context
	assert.False(t, ShouldLog(nil))

	// With nil request in context
	ctx := newCtx()
	ctx.SetRequest(nil)
	assert.False(t, ShouldLog(ctx))

	// With ignoring path
	ctx = newCtx()
	ctx.SetRequest(&http.Request{
		URL: &url.URL{
			Path: "/rk/v1/assets",
		},
	})
	assert.False(t, ShouldLog(ctx))

	ctx.SetRequest(&http.Request{
		URL: &url.URL{
			Path: "/rk/v1/tv",
		},
	})
	assert.False(t, ShouldLog(ctx))

	ctx.SetRequest(&http.Request{
		URL: &url.URL{
			Path: "/sw/",
		},
	})
	assert.False(t, ShouldLog(ctx))

	// Expect true
	ctx.SetRequest(&http.Request{
		URL: &url.URL{
			Path: "ut-path",
		},
	})
	assert.True(t, ShouldLog(ctx))
}

// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkechosec

import (
	"bytes"
	"crypto/tls"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

var userHandler = func(ctx echo.Context) error {
	return ctx.String(http.StatusOK, "")
}

func newCtx() (echo.Context, *httptest.ResponseRecorder) {
	var buf bytes.Buffer
	req := httptest.NewRequest(http.MethodPost, "/ut-path", &buf)
	resp := httptest.NewRecorder()
	return echo.New().NewContext(req, resp), resp
}

func TestInterceptor(t *testing.T) {
	defer assertNotPanic(t)

	// with skipper
	ctx, resp := newCtx()
	handler := Interceptor(WithSkipper(func(context echo.Context) bool {
		return true
	}))
	assert.Nil(t, handler(userHandler)(ctx))
	assert.Equal(t, http.StatusOK, resp.Code)

	// without options
	ctx, resp = newCtx()
	handler = Interceptor()
	assert.Nil(t, handler(userHandler)(ctx))
	assert.Equal(t, http.StatusOK, resp.Code)
	containsHeader(t, ctx,
		headerXXSSProtection,
		headerXContentTypeOptions,
		headerXFrameOptions)

	// with options
	ctx, resp = newCtx()
	ctx.Request().TLS = &tls.ConnectionState{}
	handler = Interceptor(
		WithXSSProtection("ut-xss"),
		WithContentTypeNosniff("ut-sniff"),
		WithXFrameOptions("ut-frame"),
		WithHSTSMaxAge(10),
		WithHSTSExcludeSubdomains(true),
		WithHSTSPreloadEnabled(true),
		WithContentSecurityPolicy("ut-policy"),
		WithCSPReportOnly(true),
		WithReferrerPolicy("ut-ref"),
		WithIgnorePrefix("ut-prefix"))
	assert.Nil(t, handler(userHandler)(ctx))
	assert.Equal(t, http.StatusOK, resp.Code)
	containsHeader(t, ctx,
		headerXXSSProtection,
		headerXContentTypeOptions,
		headerXFrameOptions,
		headerStrictTransportSecurity,
		headerContentSecurityPolicyReportOnly,
		headerReferrerPolicy)
}

func containsHeader(t *testing.T, ctx echo.Context, headers ...string) {
	for _, v := range headers {
		assert.Contains(t, ctx.Response().Header(), v)
	}
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

// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkechoauth

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/rookie-ninja/rk-echo/interceptor"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

var defaultMiddlewareFunc = func(context echo.Context) error {
	return nil
}

func TestInterceptor_WithIgnoringPath(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/ut-ignore-path", nil)
	res := httptest.NewRecorder()
	ctx := e.NewContext(req, res)

	handler := Interceptor(
		WithEntryNameAndType("ut-entry", "ut-type"),
		WithBasicAuth("ut-realm", "user:pass"),
		WithApiKeyAuth("ut-api-key"),
		WithIgnorePrefix("/ut-ignore-path"))

	f := handler(defaultMiddlewareFunc)

	assert.Nil(t, f(ctx))
}

func TestInterceptor_WithBasicAuth_Invalid(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/ut-path", nil)
	res := httptest.NewRecorder()
	ctx := e.NewContext(req, res)

	handler := Interceptor(
		WithEntryNameAndType("ut-entry", "ut-type"),
		WithBasicAuth("ut-realm", "user:pass"))

	// set invalid auth header
	req.Header.Set(rkechointer.RpcAuthorizationHeaderKey, "invalid")

	f := handler(defaultMiddlewareFunc)

	assert.NotNil(t, f(ctx))
	assert.Equal(t, http.StatusUnauthorized, ctx.Response().Status)
}

func TestInterceptor_WithBasicAuth_InvalidBasicAuth(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/ut-path", nil)
	res := httptest.NewRecorder()
	ctx := e.NewContext(req, res)

	handler := Interceptor(
		WithEntryNameAndType("ut-entry", "ut-type"),
		WithBasicAuth("ut-realm", "user:pass"))

	// set invalid auth header
	req.Header.Set(rkechointer.RpcAuthorizationHeaderKey, fmt.Sprintf("%s invalid", typeBasic))

	f := handler(defaultMiddlewareFunc)

	assert.NotNil(t, f(ctx))
	assert.Equal(t, http.StatusUnauthorized, ctx.Response().Status)
}

func TestInterceptor_WithApiKey_Invalid(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/ut-path", nil)
	res := httptest.NewRecorder()
	ctx := e.NewContext(req, res)

	handler := Interceptor(
		WithEntryNameAndType("ut-entry", "ut-type"),
		WithApiKeyAuth("ut-api-key"))

	// set invalid auth header
	req.Header.Set(rkechointer.RpcApiKeyHeaderKey, "invalid")

	f := handler(defaultMiddlewareFunc)

	assert.NotNil(t, f(ctx))
	assert.Equal(t, http.StatusUnauthorized, ctx.Response().Status)
}

func TestInterceptor_MissingAuth(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/ut-path", nil)
	res := httptest.NewRecorder()
	ctx := e.NewContext(req, res)

	handler := Interceptor(
		WithEntryNameAndType("ut-entry", "ut-type"),
		WithApiKeyAuth("ut-api-key"))

	f := handler(defaultMiddlewareFunc)

	assert.NotNil(t, f(ctx))
	assert.Equal(t, http.StatusUnauthorized, ctx.Response().Status)
}

func TestInterceptor_HappyCase(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/ut-ignore-path", nil)
	res := httptest.NewRecorder()
	ctx := e.NewContext(req, res)

	handler := Interceptor(
		WithEntryNameAndType("ut-entry", "ut-type"),
		//WithBasicAuth("ut-realm", "user:pass"),
		WithApiKeyAuth("ut-api-key"))

	req.Header.Set(rkechointer.RpcApiKeyHeaderKey, "ut-api-key")

	f := handler(defaultMiddlewareFunc)

	assert.Nil(t, f(ctx))
}

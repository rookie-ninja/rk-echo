// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkechocsrf

import (
	"bytes"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

var userHandler = func(ctx echo.Context) error {
	return ctx.String(http.StatusOK, "")
}

func TestInterceptor(t *testing.T) {
	defer assertNotPanic(t)

	// match 1
	ctx, resp := newCtx(http.MethodGet)
	handler := Interceptor(WithSkipper(func(context echo.Context) bool {
		return true
	}))
	assert.Nil(t, handler(userHandler)(ctx))
	assert.Equal(t, http.StatusOK, resp.Code)

	// match 2.1
	ctx, resp = newCtx(http.MethodGet)
	handler = Interceptor()
	assert.Nil(t, handler(userHandler)(ctx))
	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Contains(t, resp.Header().Get("Set-Cookie"), "_csrf")

	// match 2.2
	ctx, resp = newCtx(http.MethodGet)
	ctx.Request().AddCookie(&http.Cookie{
		Name:  "_csrf",
		Value: "ut-csrf-token",
	})
	handler = Interceptor()
	assert.Nil(t, handler(userHandler)(ctx))
	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Contains(t, resp.Header().Get("Set-Cookie"), "_csrf")

	// match 3.1
	ctx, resp = newCtx(http.MethodGet)
	handler = Interceptor()
	assert.Nil(t, handler(userHandler)(ctx))
	assert.Equal(t, http.StatusOK, resp.Code)

	// match 3.2
	ctx, resp = newCtx(http.MethodPost)
	handler = Interceptor()
	assert.Nil(t, handler(userHandler)(ctx))
	assert.Equal(t, http.StatusBadRequest, resp.Code)

	// match 3.3
	ctx, resp = newCtx(http.MethodPost)
	ctx.Request().Header.Set(headerXCSRFToken, "ut-csrf-token")
	handler = Interceptor()
	assert.Nil(t, handler(userHandler)(ctx))
	assert.Equal(t, http.StatusForbidden, resp.Code)

	// match 4.1
	ctx, resp = newCtx(http.MethodGet)
	handler = Interceptor(
		WithCookiePath("ut-path"))
	assert.Nil(t, handler(userHandler)(ctx))
	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Contains(t, resp.Header().Get("Set-Cookie"), "ut-path")

	// match 4.2
	ctx, resp = newCtx(http.MethodGet)
	handler = Interceptor(
		WithCookieDomain("ut-domain"))
	assert.Nil(t, handler(userHandler)(ctx))
	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Contains(t, resp.Header().Get("Set-Cookie"), "ut-domain")

	// match 4.3
	ctx, resp = newCtx(http.MethodGet)
	handler = Interceptor(
		WithCookieSameSite(http.SameSiteStrictMode))
	assert.Nil(t, handler(userHandler)(ctx))
	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Contains(t, resp.Header().Get("Set-Cookie"), "Strict")
}

func newCtx(method string) (echo.Context, *httptest.ResponseRecorder) {
	var buf bytes.Buffer
	req := httptest.NewRequest(method, "/ut-path", &buf)
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

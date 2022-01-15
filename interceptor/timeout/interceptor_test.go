// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkechotimeout

import (
	"fmt"
	"github.com/labstack/echo/v4"
	rkmidtimeout "github.com/rookie-ninja/rk-entry/middleware/timeout"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func sleepH(ctx echo.Context) error {
	time.Sleep(time.Second)
	ctx.JSON(http.StatusOK, "{}")
	return nil
}

func panicH(ctx echo.Context) error {
	panic(fmt.Errorf("ut panic"))
}

func returnH(ctx echo.Context) error {
	ctx.JSON(http.StatusOK, "{}")
	return nil
}

var customResponse = func(ctx echo.Context) error {
	return fmt.Errorf("custom error")
}

func getEcho(path string, handler echo.HandlerFunc, middleware echo.MiddlewareFunc) *echo.Echo {
	e := echo.New()
	e.Use(middleware)
	e.GET(path, handler)
	return e
}

func TestInterceptor_WithTimeout(t *testing.T) {
	// with global timeout response
	e := getEcho("/", sleepH, Interceptor(
		rkmidtimeout.WithTimeout(time.Nanosecond)))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	e.ServeHTTP(w, req)
	assert.Equal(t, http.StatusRequestTimeout, w.Code)

	// with path
	e = getEcho("/ut-path", sleepH, Interceptor(
		rkmidtimeout.WithTimeoutByPath("/ut-path", time.Nanosecond)))

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/ut-path", nil)
	e.ServeHTTP(w, req)

	assert.Equal(t, http.StatusRequestTimeout, w.Code)
}

func TestInterceptor_WithPanic(t *testing.T) {
	defer assertPanic(t)

	r := getEcho("/", panicH, Interceptor(
		rkmidtimeout.WithTimeout(time.Minute)))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	r.ServeHTTP(w, req)
}

func TestInterceptor_HappyCase(t *testing.T) {
	// Let's add two routes /timeout and /happy
	// We expect interceptor acts as the name describes
	r := echo.New()
	r.Use(Interceptor(
		rkmidtimeout.WithTimeoutByPath("/timeout", time.Nanosecond),
		rkmidtimeout.WithTimeoutByPath("/happy", time.Minute)))

	r.GET("/timeout", sleepH)
	r.GET("/happy", returnH)

	// timeout on /timeout
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/timeout", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusRequestTimeout, w.Code)

	// OK on /happy
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/happy", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func assertPanic(t *testing.T) {
	if r := recover(); r != nil {
		// Expect panic to be called with non nil error
		assert.True(t, true)
	} else {
		// This should never be called in case of a bug
		assert.True(t, false)
	}
}

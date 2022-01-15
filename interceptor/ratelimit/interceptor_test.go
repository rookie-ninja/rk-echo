// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkecholimit

import (
	"bytes"
	"github.com/labstack/echo/v4"
	"github.com/rookie-ninja/rk-common/error"
	"github.com/rookie-ninja/rk-entry/middleware/ratelimit"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

var userHandler = func(ctx echo.Context) error {
	return ctx.String(http.StatusOK, "")
}

func TestInterceptor(t *testing.T) {
	defer assertNotPanic(t)

	beforeCtx := rkmidlimit.NewBeforeCtx()
	mock := rkmidlimit.NewOptionSetMock(beforeCtx)

	// case 1: with error response
	inter := Interceptor(rkmidlimit.WithMockOptionSet(mock))
	ctx, w := newCtx()
	// assign any of error response
	beforeCtx.Output.ErrResp = rkerror.New(rkerror.WithHttpCode(http.StatusTooManyRequests))
	inter(userHandler)(ctx)
	assert.Equal(t, http.StatusTooManyRequests, w.Code)

	// case 2: happy case
	ctx, w = newCtx()
	beforeCtx.Output.ErrResp = nil
	inter(userHandler)(ctx)
	assert.Equal(t, http.StatusOK, w.Code)
}

func newCtx() (echo.Context, *httptest.ResponseRecorder) {
	var buf bytes.Buffer
	req := httptest.NewRequest(http.MethodGet, "/ut-path", &buf)
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

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

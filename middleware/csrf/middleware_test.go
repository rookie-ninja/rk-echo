// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkechocsrf

import (
	"bytes"
	"github.com/labstack/echo/v4"
	"github.com/rookie-ninja/rk-entry/v2/error"
	"github.com/rookie-ninja/rk-entry/v2/middleware"
	"github.com/rookie-ninja/rk-entry/v2/middleware/csrf"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

var userHandler = func(ctx echo.Context) error {
	return ctx.String(http.StatusOK, "")
}

func TestMiddleware(t *testing.T) {
	defer assertNotPanic(t)

	beforeCtx := rkmidcsrf.NewBeforeCtx()
	mock := rkmidcsrf.NewOptionSetMock(beforeCtx)

	// case 1: with error response
	inter := Middleware(rkmidcsrf.WithMockOptionSet(mock))
	ctx, w := newCtx()

	// assign any of error response
	beforeCtx.Output.ErrResp = rkerror.NewForbidden("")
	inter(userHandler)(ctx)
	assert.Equal(t, http.StatusForbidden, w.Code)

	// case 2: happy case
	beforeCtx.Output.ErrResp = nil
	beforeCtx.Output.VaryHeaders = []string{"value"}
	beforeCtx.Output.Cookie = &http.Cookie{}
	ctx, w = newCtx()
	inter(userHandler)(ctx)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotEmpty(t, w.Header().Get(rkmid.HeaderVary))
	assert.NotNil(t, w.Header().Get("Set-Cookie"))
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

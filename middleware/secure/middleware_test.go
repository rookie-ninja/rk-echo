// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkechosec

import (
	"bytes"
	"github.com/labstack/echo/v4"
	"github.com/rookie-ninja/rk-entry/v2/middleware/secure"
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

	beforeCtx := rkmidsec.NewBeforeCtx()
	mock := rkmidsec.NewOptionSetMock(beforeCtx)

	// case 1: with error response
	inter := Middleware(rkmidsec.WithMockOptionSet(mock))
	ctx, w := newCtx()
	// assign any of error response
	beforeCtx.Output.HeadersToReturn["key"] = "value"
	inter(userHandler)(ctx)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "value", w.Header().Get("key"))
}

func newCtx() (echo.Context, *httptest.ResponseRecorder) {
	var buf bytes.Buffer
	req := httptest.NewRequest(http.MethodPost, "/ut-path", &buf)
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

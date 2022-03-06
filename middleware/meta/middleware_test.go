// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkechometa

import (
	"bytes"
	"github.com/labstack/echo/v4"
	"github.com/rookie-ninja/rk-entry/v2/entry"
	"github.com/rookie-ninja/rk-entry/v2/middleware/meta"
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

	beforeCtx := rkmidmeta.NewBeforeCtx()
	mock := rkmidmeta.NewOptionSetMock(beforeCtx)

	inter := Middleware(rkmidmeta.WithMockOptionSet(mock))
	ctx, w := newCtx()

	beforeCtx.Input.Event = rkentry.EventEntryNoop.CreateEventNoop()
	beforeCtx.Output.HeadersToReturn["key"] = "value"

	inter(userHandler)(ctx)

	assert.Equal(t, "value", w.Header().Get("key"))
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

// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkecholog

import (
	"bytes"
	"github.com/labstack/echo/v4"
	rkentry "github.com/rookie-ninja/rk-entry/v2/entry"
	"github.com/rookie-ninja/rk-entry/v2/middleware"
	"github.com/rookie-ninja/rk-entry/v2/middleware/log"
	"github.com/rookie-ninja/rk-query"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
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

	beforeCtx := rkmidlog.NewBeforeCtx()
	afterCtx := rkmidlog.NewAfterCtx()
	mock := rkmidlog.NewOptionSetMock(beforeCtx, afterCtx)
	inter := Middleware(rkmidlog.WithMockOptionSet(mock))
	ctx, w := newCtx()

	// happy case
	event := rkentry.EventEntryNoop.CreateEventNoop()
	logger := rkentry.LoggerEntryNoop.Logger
	beforeCtx.Output.Event = event
	beforeCtx.Output.Logger = logger

	inter(userHandler)(ctx)

	eventFromCtx := ctx.Get(rkmid.EventKey.String())
	loggerFromCtx := ctx.Get(rkmid.LoggerKey.String())
	assert.Equal(t, event, eventFromCtx.(rkquery.Event))
	assert.Equal(t, logger, loggerFromCtx.(*zap.Logger))

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

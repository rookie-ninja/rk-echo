// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkechoauth

import (
	"bytes"
	"github.com/labstack/echo/v4"
	rkerror "github.com/rookie-ninja/rk-common/error"
	rkmidauth "github.com/rookie-ninja/rk-entry/middleware/auth"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

var userFunc = func(context echo.Context) error {
	return nil
}

func newCtx() (echo.Context, *httptest.ResponseRecorder) {
	var buf bytes.Buffer
	req := httptest.NewRequest(http.MethodGet, "/ut-path", &buf)
	resp := httptest.NewRecorder()
	return echo.New().NewContext(req, resp), resp
}

func TestInterceptor(t *testing.T) {
	beforeCtx := rkmidauth.NewBeforeCtx()
	mock := rkmidauth.NewOptionSetMock(beforeCtx)

	// case 1: with error response
	inter := Interceptor(rkmidauth.WithMockOptionSet(mock))
	ctx, w := newCtx()
	// assign any of error response
	beforeCtx.Output.ErrResp = rkerror.New(rkerror.WithHttpCode(http.StatusUnauthorized))
	beforeCtx.Output.HeadersToReturn["key"] = "value"
	inter(userFunc)(ctx)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Equal(t, "value", w.Header().Get("key"))

	// case 2: happy case
	beforeCtx.Output.ErrResp = nil
	ctx, w = newCtx()
	inter(userFunc)(ctx)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

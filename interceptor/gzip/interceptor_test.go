// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkechogzip

import (
	"bytes"
	"compress/gzip"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newCtx(encode bool) (echo.Context, *httptest.ResponseRecorder) {
	var buf bytes.Buffer
	req := httptest.NewRequest(http.MethodPost, "/ut-path", &buf)

	if encode {
		zw := gzip.NewWriter(&buf)
		zw.Write([]byte("ut-string"))
		zw.Flush()
		zw.Close()
		req.Header.Set(echo.HeaderContentEncoding, gzipEncoding)
		req.Header.Set(echo.HeaderAcceptEncoding, gzipEncoding)
	} else {
		buf.WriteString("ut-string")
	}

	resp := httptest.NewRecorder()

	return echo.New().NewContext(req, resp), resp
}

func TestInterceptor(t *testing.T) {
	defer assertNotPanic(t)

	// With skipper
	ctx, _ := newCtx(false)
	handler := Interceptor(WithSkipper(func(context echo.Context) bool {
		return true
	}))

	f := handler(func(ctx echo.Context) error {
		buf := new(bytes.Buffer)
		buf.ReadFrom(ctx.Request().Body)

		return ctx.String(http.StatusOK, buf.String())
	})

	assert.Nil(t, f(ctx))
	recorder := ctx.Response().Writer.(*httptest.ResponseRecorder)
	assert.Equal(t, "ut-string", recorder.Body.String())

	// without skipper
	ctx, recorder = newCtx(true)
	handler = Interceptor()

	f = handler(func(ctx echo.Context) error {
		buf := new(bytes.Buffer)
		buf.ReadFrom(ctx.Request().Body)

		return ctx.String(http.StatusOK, buf.String())
	})

	assert.Nil(t, f(ctx))
	zr, _ := gzip.NewReader(recorder.Body)

	var res bytes.Buffer

	io.Copy(&res, zr)
	assert.Equal(t, "ut-string", res.String())

	// with empty response
	ctx, recorder = newCtx(true)
	handler = Interceptor()

	f = handler(func(ctx echo.Context) error {
		return ctx.String(http.StatusOK, "")
	})

	assert.Nil(t, f(ctx))
}

// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkechojwt

import (
	"errors"
	"github.com/golang-jwt/jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"net/http"
	"strings"
	"testing"
)

var userHandler = func(ctx echo.Context) error {
	return ctx.String(http.StatusOK, "")
}

func TestInterceptor(t *testing.T) {
	defer assertNotPanic(t)

	// with skipper
	ctx, resp := newCtx()
	handler := Interceptor(WithSkipper(func(context echo.Context) bool {
		return true
	}))
	assert.Nil(t, handler(userHandler)(ctx))
	assert.Equal(t, http.StatusOK, resp.Code)

	// without options
	ctx, resp = newCtx()
	handler = Interceptor()
	assert.Nil(t, handler(userHandler)(ctx))
	assert.Equal(t, http.StatusUnauthorized, resp.Code)

	// with parse token error
	parseTokenErrFunc := func(auth string, c echo.Context) (*jwt.Token, error) {
		return nil, errors.New("ut-error")
	}
	ctx, resp = newCtx()
	ctx.Request().Header.Set(headerAuthorization, strings.Join([]string{"Bearer", "ut-auth"}, " "))
	handler = Interceptor(
		WithParseTokenFunc(parseTokenErrFunc))
	assert.Nil(t, handler(userHandler)(ctx))
	assert.Equal(t, http.StatusUnauthorized, resp.Code)

	// happy case
	parseTokenErrFunc = func(auth string, c echo.Context) (*jwt.Token, error) {
		return &jwt.Token{}, nil
	}
	ctx, resp = newCtx()
	ctx.Request().Header.Set(headerAuthorization, strings.Join([]string{"Bearer", "ut-auth"}, " "))
	handler = Interceptor(
		WithParseTokenFunc(parseTokenErrFunc))
	assert.Nil(t, handler(userHandler)(ctx))
	assert.Equal(t, http.StatusOK, resp.Code)
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

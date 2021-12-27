// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkecho

import (
	"context"
	"github.com/rookie-ninja/rk-entry/entry"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewTvEntry(t *testing.T) {
	entry := NewTvEntry(
		WithEventLoggerEntryTv(rkentry.NoopEventLoggerEntry()),
		WithZapLoggerEntryTv(rkentry.NoopZapLoggerEntry()))

	assert.Equal(t, TvEntryNameDefault, entry.GetName())
	assert.Equal(t, TvEntryType, entry.GetType())
	assert.Equal(t, TvEntryDescription, entry.GetDescription())
	assert.NotEmpty(t, entry.String())
	assert.Nil(t, entry.UnmarshalJSON(nil))
}

func TestTvEntry_Bootstrap(t *testing.T) {
	entry := NewTvEntry(
		WithEventLoggerEntryTv(rkentry.NoopEventLoggerEntry()),
		WithZapLoggerEntryTv(rkentry.NoopZapLoggerEntry()))

	entry.Bootstrap(context.TODO())
}

func TestTvEntry_Interrupt(t *testing.T) {
	entry := NewTvEntry(
		WithEventLoggerEntryTv(rkentry.NoopEventLoggerEntry()),
		WithZapLoggerEntryTv(rkentry.NoopZapLoggerEntry()))

	entry.Interrupt(context.TODO())
}

func TestTvEntry_TV(t *testing.T) {
	entry := NewTvEntry(
		WithEventLoggerEntryTv(rkentry.NoopEventLoggerEntry()),
		WithZapLoggerEntryTv(rkentry.NoopZapLoggerEntry()))
	entry.Bootstrap(context.TODO())

	defer assertNotPanic(t)
	// With nil context
	entry.TV(nil)

	// With all paths
	ctx, recorder := newCtx()
	ctx.SetParamNames("*")
	ctx.SetParamValues("/")
	entry.TV(ctx)
	assert.NotEmpty(t, recorder.Body.String())

	// apis
	ctx, recorder = newCtx()
	ctx.SetParamNames("*")
	ctx.SetParamValues("/apis")
	entry.TV(ctx)
	assert.NotEmpty(t, recorder.Body.String())

	// entries
	ctx, recorder = newCtx()
	ctx.SetParamNames("*")
	ctx.SetParamValues("/entries")
	entry.TV(ctx)
	assert.NotEmpty(t, recorder.Body.String())

	// configs
	ctx, recorder = newCtx()
	ctx.SetParamNames("*")
	ctx.SetParamValues("/configs")
	entry.TV(ctx)
	assert.NotEmpty(t, recorder.Body.String())

	// certs
	ctx, recorder = newCtx()
	ctx.SetParamNames("*")
	ctx.SetParamValues("/certs")
	entry.TV(ctx)
	assert.NotEmpty(t, recorder.Body.String())

	// os
	ctx, recorder = newCtx()
	ctx.SetParamNames("*")
	ctx.SetParamValues("/os")
	entry.TV(ctx)
	assert.NotEmpty(t, recorder.Body.String())

	// env
	ctx, recorder = newCtx()
	ctx.SetParamNames("*")
	ctx.SetParamValues("/env")
	entry.TV(ctx)
	assert.NotEmpty(t, recorder.Body.String())

	// prometheus
	ctx, recorder = newCtx()
	ctx.SetParamNames("*")
	ctx.SetParamValues("/prometheus")
	entry.TV(ctx)
	assert.NotEmpty(t, recorder.Body.String())

	// logs
	ctx, recorder = newCtx()
	ctx.SetParamNames("*")
	ctx.SetParamValues("/logs")
	entry.TV(ctx)
	assert.NotEmpty(t, recorder.Body.String())

	// deps
	ctx, recorder = newCtx()
	ctx.SetParamNames("*")
	ctx.SetParamValues("/deps")
	entry.TV(ctx)
	assert.NotEmpty(t, recorder.Body.String())

	// license
	ctx, recorder = newCtx()
	ctx.SetParamNames("*")
	ctx.SetParamValues("/license")
	entry.TV(ctx)
	assert.NotEmpty(t, recorder.Body.String())

	// info
	ctx, recorder = newCtx()
	ctx.SetParamNames("*")
	ctx.SetParamValues("/info")
	entry.TV(ctx)
	assert.NotEmpty(t, recorder.Body.String())

	// git
	ctx, recorder = newCtx()
	ctx.SetParamNames("*")
	ctx.SetParamValues("/git")
	entry.TV(ctx)
	assert.NotEmpty(t, recorder.Body.String())

	// unknown
	ctx, recorder = newCtx()
	ctx.SetParamNames("*")
	ctx.SetParamValues("/unknown")
	entry.TV(ctx)
	assert.NotEmpty(t, recorder.Body.String())
}

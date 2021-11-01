// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkechometrics

import (
	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rookie-ninja/rk-echo/interceptor"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

var defaultMiddlewareFunc = func(context echo.Context) error {
	return nil
}

func newCtx() echo.Context {
	return echo.New().NewContext(
		httptest.NewRequest(http.MethodGet, "/ut-path", nil),
		httptest.NewRecorder())
}

func TestWithEntryNameAndType(t *testing.T) {
	set := newOptionSet(
		WithEntryNameAndType("ut-entry", "ut-type"))

	assert.Equal(t, "ut-entry", set.EntryName)
	assert.Equal(t, "ut-type", set.EntryType)

	defer clearAllMetrics()
}

func TestWithRegisterer(t *testing.T) {
	reg := prometheus.NewRegistry()
	set := newOptionSet(
		WithRegisterer(reg))

	assert.Equal(t, reg, set.registerer)

	defer clearAllMetrics()
}

func TestGetOptionSet(t *testing.T) {
	// With nil context
	assert.Nil(t, getOptionSet(nil))

	ctx := newCtx()

	// Happy case
	ctx.Set(rkechointer.RpcEntryNameKey, "ut-entry")
	set := newOptionSet()
	optionsMap["ut-entry"] = set
	assert.Equal(t, set, getOptionSet(ctx))

	defer clearAllMetrics()
}

func TestGetServerMetricsSet(t *testing.T) {
	reg := prometheus.NewRegistry()
	set := newOptionSet(
		WithEntryNameAndType("ut-entry", "ut-type"),
		WithRegisterer(reg))

	ctx := newCtx()

	ctx.Set(rkechointer.RpcEntryNameKey, "ut-entry")
	assert.Equal(t, set.MetricsSet, GetServerMetricsSet(ctx))

	defer clearAllMetrics()
}

func TestListServerMetricsSets(t *testing.T) {
	reg := prometheus.NewRegistry()
	newOptionSet(
		WithEntryNameAndType("ut-entry", "ut-type"),
		WithRegisterer(reg))

	ctx := newCtx()
	ctx.Set(rkechointer.RpcEntryNameKey, "ut-entry")
	assert.Len(t, ListServerMetricsSets(), 1)

	defer clearAllMetrics()
}

func TestGetServerResCodeMetrics(t *testing.T) {
	// With nil context
	assert.Nil(t, GetServerResCodeMetrics(nil))

	// Happy case
	reg := prometheus.NewRegistry()
	newOptionSet(
		WithEntryNameAndType("ut-entry", "ut-type"),
		WithRegisterer(reg))

	ctx := newCtx()

	ctx.Set(rkechointer.RpcEntryNameKey, "ut-entry")

	assert.NotNil(t, GetServerResCodeMetrics(ctx))

	defer clearAllMetrics()
}

func TestGetServerErrorMetrics(t *testing.T) {
	// With nil context
	assert.Nil(t, GetServerErrorMetrics(nil))

	ctx := newCtx()

	// Happy case
	reg := prometheus.NewRegistry()
	newOptionSet(
		WithEntryNameAndType("ut-entry", "ut-type"),
		WithRegisterer(reg))

	ctx.Set(rkechointer.RpcEntryNameKey, "ut-entry")

	assert.NotNil(t, GetServerErrorMetrics(ctx))

	defer clearAllMetrics()
}

func TestGetServerDurationMetrics(t *testing.T) {
	// With nil context
	assert.Nil(t, GetServerDurationMetrics(nil))

	// Happy case
	reg := prometheus.NewRegistry()
	newOptionSet(
		WithEntryNameAndType("ut-entry", "ut-type"),
		WithRegisterer(reg))

	ctx := newCtx()
	ctx.Set(rkechointer.RpcEntryNameKey, "ut-entry")

	assert.NotNil(t, GetServerDurationMetrics(ctx))

	defer clearAllMetrics()
}

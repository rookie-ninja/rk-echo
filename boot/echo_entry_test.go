// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkecho

import (
	"context"
	"github.com/labstack/echo/v4"
	rkecholog "github.com/rookie-ninja/rk-echo/interceptor/log/zap"
	rkechometrics "github.com/rookie-ninja/rk-echo/interceptor/metrics/prom"
	"github.com/rookie-ninja/rk-entry/entry"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net"
	"os"
	"path"
	"strconv"
	"strings"
	"testing"
	"time"
)

const (
	defaultBootConfigStr = `
---
echo:
  - name: greeter
    port: 8080
    enabled: true
    sw:
      enabled: true
      path: "sw"
    commonService:
      enabled: true
    tv:
      enabled: true
    prom:
      enabled: true
      pusher:
        enabled: false
    interceptors:
      loggingZap:
        enabled: true
      metricsProm:
        enabled: true
      auth:
        enabled: true
        basic:
          - "user:pass"
      meta:
        enabled: true
      tracingTelemetry:
        enabled: true
      ratelimit:
        enabled: true
      timeout:
        enabled: true
  - name: greeter2
    port: 2008
    enabled: true
    sw:
      enabled: true
      path: "sw"
    commonService:
      enabled: true
    tv:
      enabled: true
    interceptors:
      loggingZap:
        enabled: true
      metricsProm:
        enabled: true
      auth:
        enabled: true
        basic:
          - "user:pass"
`
)

func TestWithZapLoggerEntryEcho_HappyCase(t *testing.T) {
	loggerEntry := rkentry.NoopZapLoggerEntry()
	entry := RegisterEchoEntry()

	option := WithZapLoggerEntryEcho(loggerEntry)
	option(entry)

	assert.Equal(t, loggerEntry, entry.ZapLoggerEntry)
}

func TestWithEventLoggerEntryEcho_HappyCase(t *testing.T) {
	entry := RegisterEchoEntry()

	eventLoggerEntry := rkentry.NoopEventLoggerEntry()

	option := WithEventLoggerEntryEcho(eventLoggerEntry)
	option(entry)

	assert.Equal(t, eventLoggerEntry, entry.EventLoggerEntry)
}

func TestWithInterceptorsEcho_WithNilInterceptorList(t *testing.T) {
	entry := RegisterEchoEntry()

	option := WithInterceptorsEcho(nil)
	option(entry)

	assert.NotNil(t, entry.Interceptors)
}

func TestWithInterceptorsEcho_HappyCase(t *testing.T) {
	entry := RegisterEchoEntry()

	loggingInterceptor := rkecholog.Interceptor()
	metricsInterceptor := rkechometrics.Interceptor()

	interceptors := []echo.MiddlewareFunc{
		loggingInterceptor,
		metricsInterceptor,
	}

	option := WithInterceptorsEcho(interceptors...)
	option(entry)

	assert.NotNil(t, entry.Interceptors)
	// should contains logging, metrics and panic interceptor
	// where panic interceptor is inject by default
	assert.Len(t, entry.Interceptors, 3)
}

func TestWithCommonServiceEntryEcho_WithEntry(t *testing.T) {
	entry := RegisterEchoEntry()

	option := WithCommonServiceEntryEcho(NewCommonServiceEntry())
	option(entry)

	assert.NotNil(t, entry.CommonServiceEntry)
}

func TestWithCommonServiceEntryEcho_WithoutEntry(t *testing.T) {
	entry := RegisterEchoEntry()

	assert.Nil(t, entry.CommonServiceEntry)
}

func TestWithTVEntryEcho_WithEntry(t *testing.T) {
	entry := RegisterEchoEntry()

	option := WithTVEntryEcho(NewTvEntry())
	option(entry)

	assert.NotNil(t, entry.TvEntry)
}

func TestWithTVEntry_WithoutEntry(t *testing.T) {
	entry := RegisterEchoEntry()

	assert.Nil(t, entry.TvEntry)
}

func TestWithCertEntryEcho_HappyCase(t *testing.T) {
	entry := RegisterEchoEntry()
	certEntry := &rkentry.CertEntry{}

	option := WithCertEntryEcho(certEntry)
	option(entry)

	assert.Equal(t, entry.CertEntry, certEntry)
}

func TestWithSWEntryEcho_HappyCase(t *testing.T) {
	entry := RegisterEchoEntry()
	sw := NewSwEntry()

	option := WithSwEntryEcho(sw)
	option(entry)

	assert.Equal(t, entry.SwEntry, sw)
}

func TestWithPortEcho_HappyCase(t *testing.T) {
	entry := RegisterEchoEntry()
	port := uint64(1111)

	option := WithPortEcho(port)
	option(entry)

	assert.Equal(t, entry.Port, port)
}

func TestWithNameEcho_HappyCase(t *testing.T) {
	entry := RegisterEchoEntry()
	name := "unit-test-entry"

	option := WithNameEcho(name)
	option(entry)

	assert.Equal(t, entry.EntryName, name)
}

func TestRegisterEchoEntriesWithConfig_WithInvalidConfigFilePath(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			// expect panic to be called with non nil error
			assert.True(t, true)
		} else {
			// this should never be called in case of a bug
			assert.True(t, false)
		}
	}()

	RegisterEchoEntriesWithConfig("/invalid-path")
}

func TestRegisterEchoEntriesWithConfig_WithNilFactory(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			// expect panic to be called with non nil error
			assert.True(t, false)
		} else {
			// this should never be called in case of a bug
			assert.True(t, true)
		}
	}()

	// write config file in unit test temp directory
	tempDir := path.Join(t.TempDir(), "boot.yaml")
	assert.Nil(t, ioutil.WriteFile(tempDir, []byte(defaultBootConfigStr), os.ModePerm))
	entries := RegisterEchoEntriesWithConfig(tempDir)
	assert.NotNil(t, entries)
	assert.Len(t, entries, 2)
}

func TestRegisterEchoEntriesWithConfig_HappyCase(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			// expect panic to be called with non nil error
			assert.True(t, false)
		} else {
			// this should never be called in case of a bug
			assert.True(t, true)
		}
	}()

	// write config file in unit test temp directory
	tempDir := path.Join(t.TempDir(), "boot.yaml")
	assert.Nil(t, ioutil.WriteFile(tempDir, []byte(defaultBootConfigStr), os.ModePerm))
	entries := RegisterEchoEntriesWithConfig(tempDir)
	assert.NotNil(t, entries)
	assert.Len(t, entries, 2)

	// validate entry element based on boot.yaml config defined in defaultBootConfigStr
	greeter := entries["greeter"].(*EchoEntry)
	assert.NotNil(t, greeter)
	assert.Equal(t, uint64(8080), greeter.Port)
	assert.NotNil(t, greeter.SwEntry)
	assert.NotNil(t, greeter.CommonServiceEntry)
	assert.NotNil(t, greeter.TvEntry)
	// logging, metrics, auth and panic interceptor should be included
	assert.True(t, len(greeter.Interceptors) > 0)

	greeter2 := entries["greeter2"].(*EchoEntry)
	assert.NotNil(t, greeter2)
	assert.Equal(t, uint64(2008), greeter2.Port)
	assert.NotNil(t, greeter2.SwEntry)
	assert.NotNil(t, greeter2.CommonServiceEntry)
	assert.NotNil(t, greeter2.TvEntry)
	// logging, metrics, auth and panic interceptor should be included
	assert.Len(t, greeter2.Interceptors, 4)
}

func TestRegisterEchoEntry_WithZapLoggerEntry(t *testing.T) {
	loggerEntry := rkentry.NoopZapLoggerEntry()
	entry := RegisterEchoEntry(WithZapLoggerEntryEcho(loggerEntry))
	assert.Equal(t, loggerEntry, entry.ZapLoggerEntry)
}

func TestRegisterEchoEntry_WithEventLoggerEntry(t *testing.T) {
	loggerEntry := rkentry.NoopEventLoggerEntry()

	entry := RegisterEchoEntry(WithEventLoggerEntryEcho(loggerEntry))
	assert.Equal(t, loggerEntry, entry.EventLoggerEntry)
}

func TestNewEchoEntry_WithInterceptors(t *testing.T) {
	loggingInterceptor := rkecholog.Interceptor()
	entry := RegisterEchoEntry(WithInterceptorsEcho(loggingInterceptor))
	assert.Len(t, entry.Interceptors, 2)
}

func TestNewEchoEntry_WithCommonServiceEntry(t *testing.T) {
	entry := RegisterEchoEntry(WithCommonServiceEntryEcho(NewCommonServiceEntry()))
	assert.NotNil(t, entry.CommonServiceEntry)
}

func TestNewEchoEntry_WithTVEntry(t *testing.T) {
	entry := RegisterEchoEntry(WithTVEntryEcho(NewTvEntry()))
	assert.NotNil(t, entry.TvEntry)
}

func TestNewEchoEntry_WithCertStore(t *testing.T) {
	certEntry := &rkentry.CertEntry{}

	entry := RegisterEchoEntry(WithCertEntryEcho(certEntry))
	assert.Equal(t, certEntry, entry.CertEntry)
}

func TestNewEchoEntry_WithSWEntry(t *testing.T) {
	sw := NewSwEntry()
	entry := RegisterEchoEntry(WithSwEntryEcho(sw))
	assert.Equal(t, sw, entry.SwEntry)
}

func TestNewEchoEntry_WithPort(t *testing.T) {
	entry := RegisterEchoEntry(WithPortEcho(8080))
	assert.Equal(t, uint64(8080), entry.Port)
}

func TestNewEchoEntry_WithName(t *testing.T) {
	entry := RegisterEchoEntry(WithNameEcho("unit-test-greeter"))
	assert.Equal(t, "unit-test-greeter", entry.GetName())
}

func TestNewEchoEntry_WithDefaultValue(t *testing.T) {
	entry := RegisterEchoEntry()
	assert.True(t, strings.HasPrefix(entry.GetName(), "EchoServer-"))
	assert.NotNil(t, entry.ZapLoggerEntry)
	assert.NotNil(t, entry.EventLoggerEntry)
	assert.Len(t, entry.Interceptors, 1)
	assert.NotNil(t, entry.Echo)
	assert.Nil(t, entry.SwEntry)
	assert.Nil(t, entry.CertEntry)
	assert.False(t, entry.IsSwEnabled())
	assert.False(t, entry.IsTlsEnabled())
	assert.Nil(t, entry.CommonServiceEntry)
	assert.Nil(t, entry.TvEntry)
	assert.Equal(t, "EchoEntry", entry.GetType())
}

func TestEchoEntry_GetName_HappyCase(t *testing.T) {
	entry := RegisterEchoEntry(WithNameEcho("unit-test-entry"))
	assert.Equal(t, "unit-test-entry", entry.GetName())
}

func TestEchoEntry_GetType_HappyCase(t *testing.T) {
	assert.Equal(t, "EchoEntry", RegisterEchoEntry().GetType())
}

func TestEchoEntry_String_HappyCase(t *testing.T) {
	assert.NotEmpty(t, RegisterEchoEntry().String())
}

func TestEchoEntry_IsSwEnabled_ExpectTrue(t *testing.T) {
	sw := NewSwEntry()
	entry := RegisterEchoEntry(WithSwEntryEcho(sw))
	assert.True(t, entry.IsSwEnabled())
}

func TestEchoEntry_IsSwEnabled_ExpectFalse(t *testing.T) {
	entry := RegisterEchoEntry()
	assert.False(t, entry.IsSwEnabled())
}

func TestEchoEntry_IsTlsEnabled_ExpectTrue(t *testing.T) {
	certEntry := &rkentry.CertEntry{
		Store: &rkentry.CertStore{},
	}

	entry := RegisterEchoEntry(WithCertEntryEcho(certEntry))
	assert.True(t, entry.IsTlsEnabled())
}

func TestEchoEntry_IsTlsEnabled_ExpectFalse(t *testing.T) {
	entry := RegisterEchoEntry()
	assert.False(t, entry.IsTlsEnabled())
}

func TestEchoEntry_GetEcho_HappyCase(t *testing.T) {
	entry := RegisterEchoEntry()
	assert.NotNil(t, entry.Echo)
	assert.NotNil(t, entry.Echo.Server)
}

func TestEchoEntry_Bootstrap_WithSwagger(t *testing.T) {
	sw := NewSwEntry(
		WithPathSw("sw"),
		WithZapLoggerEntrySw(rkentry.NoopZapLoggerEntry()),
		WithEventLoggerEntrySw(rkentry.NoopEventLoggerEntry()))
	entry := RegisterEchoEntry(
		WithNameEcho("unit-test-entry"),
		WithPortEcho(8080),
		WithZapLoggerEntryEcho(rkentry.NoopZapLoggerEntry()),
		WithEventLoggerEntryEcho(rkentry.NoopEventLoggerEntry()),
		WithSwEntryEcho(sw))

	go entry.Bootstrap(context.Background())
	time.Sleep(time.Second)
	// endpoint should be accessible with 8080 port
	validateServerIsUp(t, entry.Port)
	assert.Len(t, entry.Echo.Routes(), 3)

	entry.Interrupt(context.Background())
	time.Sleep(time.Second)
}

func TestEchoEntry_Bootstrap_WithoutSwagger(t *testing.T) {
	entry := RegisterEchoEntry(
		WithNameEcho("unit-test-entry"),
		WithPortEcho(8080),
		WithZapLoggerEntryEcho(rkentry.NoopZapLoggerEntry()),
		WithEventLoggerEntryEcho(rkentry.NoopEventLoggerEntry()))

	go entry.Bootstrap(context.Background())
	time.Sleep(time.Second)
	// endpoint should be accessible with 8080 port
	validateServerIsUp(t, entry.Port)
	assert.Empty(t, entry.Echo.Routes())

	entry.Interrupt(context.Background())
	time.Sleep(time.Second)
}

func TestEchoEntry_Bootstrap_WithoutTLS(t *testing.T) {
	entry := RegisterEchoEntry(
		WithNameEcho("unit-test-entry"),
		WithPortEcho(8080),
		WithZapLoggerEntryEcho(rkentry.NoopZapLoggerEntry()),
		WithEventLoggerEntryEcho(rkentry.NoopEventLoggerEntry()))

	go entry.Bootstrap(context.Background())
	time.Sleep(time.Second)
	// endpoint should be accessible with 8080 port
	validateServerIsUp(t, entry.Port)

	entry.Interrupt(context.Background())
	time.Sleep(time.Second)
}

func TestEchoEntry_Shutdown_WithBootstrap(t *testing.T) {
	defer assertNotPanic(t)

	entry := RegisterEchoEntry(
		WithNameEcho("unit-test-entry"),
		WithPortEcho(8080),
		WithZapLoggerEntryEcho(rkentry.NoopZapLoggerEntry()),
		WithEventLoggerEntryEcho(rkentry.NoopEventLoggerEntry()))

	go entry.Bootstrap(context.Background())
	time.Sleep(time.Second)
	// endpoint should be accessible with 8080 port
	validateServerIsUp(t, entry.Port)

	entry.Interrupt(context.Background())
	time.Sleep(time.Second)
}

func TestEchoEntry_Shutdown_WithoutBootstrap(t *testing.T) {
	defer assertNotPanic(t)

	entry := RegisterEchoEntry(
		WithNameEcho("unit-test-entry"),
		WithPortEcho(8080),
		WithZapLoggerEntryEcho(rkentry.NoopZapLoggerEntry()),
		WithEventLoggerEntryEcho(rkentry.NoopEventLoggerEntry()))

	entry.Interrupt(context.Background())
}

func validateServerIsUp(t *testing.T, port uint64) {
	conn, err := net.DialTimeout("tcp", net.JoinHostPort("0.0.0.0", strconv.FormatUint(port, 10)), time.Second)
	assert.Nil(t, err)
	assert.NotNil(t, conn)
	if conn != nil {
		assert.Nil(t, conn.Close())
	}
}

// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkecho

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus"
	rkechometa "github.com/rookie-ninja/rk-echo/interceptor/meta"
	rkechometrics "github.com/rookie-ninja/rk-echo/interceptor/metrics/prom"
	rkentry "github.com/rookie-ninja/rk-entry/entry"
	rkmidmetrics "github.com/rookie-ninja/rk-entry/middleware/metrics"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"math/big"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"strconv"
	"testing"
	"time"
)

const (
	defaultBootConfigStr = `
---
echo:
 - name: greeter
   port: 1949
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
     cors:
       enabled: true
     jwt:
       enabled: true
     secure:
       enabled: true
     csrf:
       enabled: true
     gzip:
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
 - name: greeter3
   port: 2022
   enabled: false
`
)

func TestGetEchoEntry(t *testing.T) {
	// expect nil
	assert.Nil(t, GetEchoEntry("entry-name"))

	// happy case
	echoEntry := RegisterEchoEntry(WithName("ut"))
	assert.Equal(t, echoEntry, GetEchoEntry("ut"))

	rkentry.GlobalAppCtx.RemoveEntry("ut")
}

func TestRegisterEchoEntry(t *testing.T) {
	// without options
	entry := RegisterEchoEntry()
	assert.NotNil(t, entry)
	assert.NotEmpty(t, entry.GetName())
	assert.NotEmpty(t, entry.GetType())
	assert.NotEmpty(t, entry.GetDescription())
	assert.NotEmpty(t, entry.String())
	rkentry.GlobalAppCtx.RemoveEntry(entry.GetName())

	// with options
	entry = RegisterEchoEntry(
		WithZapLoggerEntry(nil),
		WithEventLoggerEntry(nil),
		WithCommonServiceEntry(rkentry.RegisterCommonServiceEntry()),
		WithTvEntry(rkentry.RegisterTvEntry()),
		WithStaticFileHandlerEntry(rkentry.RegisterStaticFileHandlerEntry()),
		WithCertEntry(rkentry.RegisterCertEntry()),
		WithSwEntry(rkentry.RegisterSwEntry()),
		WithPort(8080),
		WithName("ut-entry"),
		WithDescription("ut-desc"),
		WithPromEntry(rkentry.RegisterPromEntry()))

	assert.NotEmpty(t, entry.GetName())
	assert.NotEmpty(t, entry.GetType())
	assert.NotEmpty(t, entry.GetDescription())
	assert.NotEmpty(t, entry.String())
	assert.True(t, entry.IsSwEnabled())
	assert.True(t, entry.IsStaticFileHandlerEnabled())
	assert.True(t, entry.IsPromEnabled())
	assert.True(t, entry.IsCommonServiceEnabled())
	assert.True(t, entry.IsTvEnabled())
	assert.True(t, entry.IsTlsEnabled())

	bytes, err := entry.MarshalJSON()
	assert.NotEmpty(t, bytes)
	assert.Nil(t, err)
	assert.Nil(t, entry.UnmarshalJSON([]byte{}))
}

func TestEchoEntry_AddInterceptor(t *testing.T) {
	defer assertNotPanic(t)
	entry := RegisterEchoEntry()
	inter := rkechometa.Interceptor()
	entry.AddInterceptor(inter)
}

func TestEchoEntry_Bootstrap(t *testing.T) {
	defer assertNotPanic(t)

	// without enable sw, static, prom, common, tv, tls
	entry := RegisterEchoEntry(WithPort(8080))
	entry.Bootstrap(context.TODO())
	validateServerIsUp(t, 8080, entry.IsTlsEnabled())
	assert.Empty(t, entry.Echo.Routes())

	entry.Interrupt(context.TODO())

	// with enable sw, static, prom, common, tv, tls
	certEntry := rkentry.RegisterCertEntry()
	certEntry.Store.ServerCert, certEntry.Store.ServerKey = generateCerts()

	entry = RegisterEchoEntry(
		WithPort(8080),
		WithCommonServiceEntry(rkentry.RegisterCommonServiceEntry()),
		WithTvEntry(rkentry.RegisterTvEntry()),
		WithStaticFileHandlerEntry(rkentry.RegisterStaticFileHandlerEntry()),
		WithCertEntry(certEntry),
		WithSwEntry(rkentry.RegisterSwEntry()),
		WithPromEntry(rkentry.RegisterPromEntry()))
	entry.Bootstrap(context.TODO())
	validateServerIsUp(t, 8080, entry.IsTlsEnabled())
	assert.NotEmpty(t, entry.Echo.Routes())

	entry.Interrupt(context.TODO())
}

func TestEchoEntry_startServer_InvalidTls(t *testing.T) {
	defer assertPanic(t)

	// with invalid tls
	entry := RegisterEchoEntry(
		WithPort(8080),
		WithCertEntry(rkentry.RegisterCertEntry()))
	event := rkentry.NoopEventLoggerEntry().GetEventFactory().CreateEventNoop()
	logger := rkentry.NoopZapLoggerEntry().GetLogger()

	entry.startServer(event, logger)
}

func TestEchoEntry_startServer_TlsServerFail(t *testing.T) {
	defer assertPanic(t)

	certEntry := rkentry.RegisterCertEntry()
	certEntry.Store.ServerCert, certEntry.Store.ServerKey = generateCerts()

	// let's give an invalid port
	entry := RegisterEchoEntry(
		WithPort(808080),
		WithCertEntry(certEntry))

	event := rkentry.NoopEventLoggerEntry().GetEventFactory().CreateEventNoop()
	logger := rkentry.NoopZapLoggerEntry().GetLogger()

	entry.startServer(event, logger)
}

func TestEchoEntry_startServer_ServerFail(t *testing.T) {
	defer assertPanic(t)

	// let's give an invalid port
	entry := RegisterEchoEntry(
		WithPort(808080))

	event := rkentry.NoopEventLoggerEntry().GetEventFactory().CreateEventNoop()
	logger := rkentry.NoopZapLoggerEntry().GetLogger()

	entry.startServer(event, logger)
}

func TestRegisterEchoEntriesWithConfig(t *testing.T) {
	assertNotPanic(t)

	// write config file in unit test temp directory
	tempDir := path.Join(t.TempDir(), "boot.yaml")
	assert.Nil(t, ioutil.WriteFile(tempDir, []byte(defaultBootConfigStr), os.ModePerm))
	entries := RegisterEchoEntriesWithConfig(tempDir)
	assert.NotNil(t, entries)
	assert.Len(t, entries, 2)

	// validate entry element based on boot.yaml config defined in defaultBootConfigStr
	greeter := entries["greeter"].(*EchoEntry)
	assert.NotNil(t, greeter)

	greeter2 := entries["greeter2"].(*EchoEntry)
	assert.NotNil(t, greeter2)

	greeter3 := entries["greeter3"]
	assert.Nil(t, greeter3)
}

func TestEchoEntry_constructSwUrl(t *testing.T) {
	// happy case
	writer := httptest.NewRecorder()
	ctx := echo.New().NewContext(&http.Request{
		Host: "8.8.8.8:1111",
	}, writer)

	path := "ut-sw"
	port := 1111

	sw := rkentry.RegisterSwEntry(rkentry.WithPathSw(path), rkentry.WithPortSw(uint64(port)))
	entry := RegisterEchoEntry(WithSwEntry(sw), WithPort(uint64(port)))

	assert.Equal(t, fmt.Sprintf("http://8.8.8.8:%s/%s/", strconv.Itoa(port), path), entry.constructSwUrl(ctx))

	// with tls
	ctx.Request().TLS = &tls.ConnectionState{}
	assert.Equal(t, fmt.Sprintf("https://8.8.8.8:%s/%s/", strconv.Itoa(port), path), entry.constructSwUrl(ctx))

	// without swagger
	entry = RegisterEchoEntry(WithPort(uint64(port)))
	assert.Equal(t, "N/A", entry.constructSwUrl(ctx))
}

func TestEchoEntry_API(t *testing.T) {
	defer assertNotPanic(t)

	req := httptest.NewRequest(http.MethodGet, "/ut-test", nil)
	writer := httptest.NewRecorder()
	ctx := echo.New().NewContext(req, writer)

	entry := RegisterEchoEntry(
		WithCommonServiceEntry(rkentry.RegisterCommonServiceEntry()),
		WithName("unit-test"))

	entry.Echo.GET("/ut-test", func(c echo.Context) error {
		return nil
	})

	entry.Apis(ctx)
	assert.Equal(t, 200, writer.Code)
	assert.NotEmpty(t, writer.Body.String())

	entry.Interrupt(context.TODO())
}

func TestEchoEntry_Req_HappyCase(t *testing.T) {
	defer assertNotPanic(t)

	req := httptest.NewRequest(http.MethodGet, "/ut-test", nil)
	writer := httptest.NewRecorder()
	ctx := echo.New().NewContext(req, writer)

	entry := RegisterEchoEntry(
		WithCommonServiceEntry(rkentry.RegisterCommonServiceEntry()),
		WithPort(8080),
		WithName("ut"))

	entry.AddInterceptor(rkechometrics.Interceptor(
		rkmidmetrics.WithEntryNameAndType("ut", "Echo"),
		rkmidmetrics.WithRegisterer(prometheus.NewRegistry())))

	entry.Bootstrap(context.TODO())

	entry.Req(ctx)
	assert.Equal(t, 200, writer.Code)
	assert.NotEmpty(t, writer.Body.String())

	entry.Interrupt(context.TODO())
}

func TestEchoEntry_Req_WithEmpty(t *testing.T) {
	defer assertNotPanic(t)

	req := httptest.NewRequest(http.MethodGet, "/req", nil)
	writer := httptest.NewRecorder()
	ctx := echo.New().NewContext(req, writer)

	entry := RegisterEchoEntry(
		WithCommonServiceEntry(rkentry.RegisterCommonServiceEntry()),
		WithPort(8080),
		WithName("ut"))

	entry.AddInterceptor(rkechometrics.Interceptor(
		rkmidmetrics.WithRegisterer(prometheus.NewRegistry())))

	entry.Bootstrap(context.TODO())

	entry.Req(ctx)
	assert.Equal(t, 200, writer.Code)
	assert.NotEmpty(t, writer.Body.String())

	entry.Interrupt(context.TODO())
}

func TestEchoEntry_TV(t *testing.T) {
	defer assertNotPanic(t)

	entry := RegisterEchoEntry(
		WithCommonServiceEntry(rkentry.RegisterCommonServiceEntry()),
		WithTvEntry(rkentry.RegisterTvEntry()),
		WithPort(8080),
		WithName("ut"))

	entry.AddInterceptor(rkechometrics.Interceptor(
		rkmidmetrics.WithEntryNameAndType("ut", "Echo")))

	entry.Bootstrap(context.TODO())

	// for /api
	req := httptest.NewRequest(http.MethodGet, "/apis", nil)
	writer := httptest.NewRecorder()
	ctx := echo.New().NewContext(req, writer)
	ctx.SetParamNames("*")
	ctx.SetParamValues("/apis")

	entry.TV(ctx)
	assert.Equal(t, 200, writer.Code)
	assert.NotEmpty(t, writer.Body.String())

	// for default
	req = httptest.NewRequest(http.MethodGet, "/other", nil)
	writer = httptest.NewRecorder()
	ctx = echo.New().NewContext(req, writer)
	ctx.SetParamNames("*")
	ctx.SetParamValues("/other")

	entry.TV(ctx)
	assert.Equal(t, 200, writer.Code)
	assert.NotEmpty(t, writer.Body.String())

	entry.Interrupt(context.TODO())
}

func generateCerts() ([]byte, []byte) {
	// Create certs and return as []byte
	ca := &x509.Certificate{
		Subject: pkix.Name{
			Organization: []string{"Fake cert."},
		},
		SerialNumber:          big.NewInt(42),
		NotAfter:              time.Now().Add(2 * time.Hour),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}

	// Create a Private Key
	key, _ := rsa.GenerateKey(rand.Reader, 4096)

	// Use CA Cert to sign a CSR and create a Public Cert
	csr := &key.PublicKey
	cert, _ := x509.CreateCertificate(rand.Reader, ca, ca, csr, key)

	// Convert keys into pem.Block
	c := &pem.Block{Type: "CERTIFICATE", Bytes: cert}
	k := &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)}

	return pem.EncodeToMemory(c), pem.EncodeToMemory(k)
}

func validateServerIsUp(t *testing.T, port uint64, isTls bool) {
	// sleep for 2 seconds waiting server startup
	time.Sleep(2 * time.Second)

	if !isTls {
		conn, err := net.DialTimeout("tcp", net.JoinHostPort("0.0.0.0", strconv.FormatUint(port, 10)), time.Second)
		assert.Nil(t, err)
		assert.NotNil(t, conn)
		if conn != nil {
			assert.Nil(t, conn.Close())
		}
		return
	}

	tlsConf := &tls.Config{
		InsecureSkipVerify: true,
	}

	tlsConn, err := tls.Dial("tcp", net.JoinHostPort("0.0.0.0", strconv.FormatUint(port, 10)), tlsConf)
	assert.Nil(t, err)
	assert.NotNil(t, tlsConn)
	if tlsConn != nil {
		assert.Nil(t, tlsConn.Close())
	}
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

func assertPanic(t *testing.T) {
	if r := recover(); r != nil {
		// Expect panic to be called with non nil error
		assert.True(t, true)
	} else {
		// This should never be called in case of a bug
		assert.True(t, false)
	}
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

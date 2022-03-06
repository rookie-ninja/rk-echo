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
	"github.com/rookie-ninja/rk-echo/middleware/meta"
	"github.com/rookie-ninja/rk-entry/v2/entry"
	"github.com/stretchr/testify/assert"
	"math/big"
	"net"
	"os"
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
   docs:
     enabled: true
   prom:
     enabled: true
     pusher:
       enabled: false
   middleware:
     logging:
       enabled: true
     prom:
       enabled: true
     auth:
       enabled: true
       basic:
         - "user:pass"
     meta:
       enabled: true
     trace:
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
   docs:
     enabled: true
   middleware:
     logging:
       enabled: true
     prom:
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

	rkentry.GlobalAppCtx.RemoveEntry(echoEntry)
}

func TestRegisterEchoEntry(t *testing.T) {
	// without options
	entry := RegisterEchoEntry()
	assert.NotNil(t, entry)
	assert.NotEmpty(t, entry.GetName())
	assert.NotEmpty(t, entry.GetType())
	assert.NotEmpty(t, entry.GetDescription())
	assert.NotEmpty(t, entry.String())
	rkentry.GlobalAppCtx.RemoveEntry(entry)

	// with options
	commonServiceEntry := rkentry.RegisterCommonServiceEntry(&rkentry.BootCommonService{
		Enabled: true,
	})
	staticEntry := rkentry.RegisterStaticFileHandlerEntry(&rkentry.BootStaticFileHandler{
		Enabled: true,
	})
	certEntry := rkentry.RegisterCertEntry(&rkentry.BootCert{
		Cert: []*rkentry.BootCertE{
			{
				Name: "ut-cert",
			},
		},
	})
	swEntry := rkentry.RegisterSWEntry(&rkentry.BootSW{
		Enabled: true,
	})
	promEntry := rkentry.RegisterPromEntry(&rkentry.BootProm{
		Enabled: true,
	})
	docsEntry := rkentry.RegisterDocsEntry(&rkentry.BootDocs{
		Enabled: true,
	})

	// with options
	entry = RegisterEchoEntry(
		WithLoggerEntry(rkentry.LoggerEntryNoop),
		WithEventEntry(rkentry.EventEntryNoop),
		WithCommonServiceEntry(commonServiceEntry),
		WithStaticFileHandlerEntry(staticEntry),
		WithCertEntry(certEntry[0]),
		WithSwEntry(swEntry),
		WithDocsEntry(docsEntry),
		WithPromEntry(promEntry),
		WithPort(8080),
		WithName("ut-entry"),
		WithDescription("ut-desc"))

	assert.NotEmpty(t, entry.GetName())
	assert.NotEmpty(t, entry.GetType())
	assert.NotEmpty(t, entry.GetDescription())
	assert.NotEmpty(t, entry.String())
	assert.True(t, entry.IsSwEnabled())
	assert.True(t, entry.IsStaticFileHandlerEnabled())
	assert.True(t, entry.IsPromEnabled())
	assert.True(t, entry.IsCommonServiceEnabled())
	assert.True(t, entry.IsDocsEnabled())
	assert.False(t, entry.IsTlsEnabled())

	bytes, err := entry.MarshalJSON()
	assert.NotEmpty(t, bytes)
	assert.Nil(t, err)
	assert.Nil(t, entry.UnmarshalJSON([]byte{}))
}

func TestEchoEntry_AddMiddleware(t *testing.T) {
	defer assertNotPanic(t)
	entry := RegisterEchoEntry()
	inter := rkechometa.Middleware()
	entry.AddMiddleware(inter)
}

func TestEchoEntry_Bootstrap(t *testing.T) {
	defer assertNotPanic(t)

	// without enable sw, static, prom, common, tv, tls
	entry := RegisterEchoEntry(WithPort(8080))
	entry.Bootstrap(context.TODO())
	validateServerIsUp(t, 8080, entry.IsTlsEnabled())
	assert.Empty(t, entry.Echo.Routes())

	entry.Interrupt(context.TODO())

	// with options
	commonServiceEntry := rkentry.RegisterCommonServiceEntry(&rkentry.BootCommonService{
		Enabled: true,
	})
	staticEntry := rkentry.RegisterStaticFileHandlerEntry(&rkentry.BootStaticFileHandler{
		Enabled: true,
	})
	swEntry := rkentry.RegisterSWEntry(&rkentry.BootSW{
		Enabled: true,
	})
	promEntry := rkentry.RegisterPromEntry(&rkentry.BootProm{
		Enabled: true,
	})
	docsEntry := rkentry.RegisterDocsEntry(&rkentry.BootDocs{
		Enabled: true,
	})

	entry = RegisterEchoEntry(
		WithPort(8080),
		WithCommonServiceEntry(commonServiceEntry),
		WithDocsEntry(docsEntry),
		WithStaticFileHandlerEntry(staticEntry),
		WithSwEntry(swEntry),
		WithPromEntry(promEntry))
	entry.Bootstrap(context.TODO())
	validateServerIsUp(t, 8080, entry.IsTlsEnabled())
	assert.NotEmpty(t, entry.Echo.Routes())

	entry.Interrupt(context.TODO())
}

func TestEchoEntry_startServer_TlsServerFail(t *testing.T) {
	defer assertPanic(t)

	certEntry := rkentry.RegisterCertEntry(&rkentry.BootCert{
		Cert: []*rkentry.BootCertE{
			{
				Name: "ut-cert",
			},
		},
	})[0]
	certificate, _ := tls.X509KeyPair(generateCerts())
	certEntry.Certificate = &certificate

	// let's give an invalid port
	entry := RegisterEchoEntry(
		WithPort(808080),
		WithCertEntry(certEntry))

	event := rkentry.EventEntryNoop.EventFactory.CreateEventNoop()
	logger := rkentry.LoggerEntryNoop.Logger

	entry.startServer(event, logger)
}

func TestEchoEntry_startServer_ServerFail(t *testing.T) {
	defer assertPanic(t)

	// let's give an invalid port
	entry := RegisterEchoEntry(
		WithPort(808080))

	event := rkentry.EventEntryNoop.CreateEventNoop()
	logger := rkentry.LoggerEntryNoop.Logger

	entry.startServer(event, logger)
}

func TestRegisterEchoEntryYAML(t *testing.T) {
	assertNotPanic(t)

	entries := RegisterEchoEntryYAML([]byte(defaultBootConfigStr))
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

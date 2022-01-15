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
	httptest "github.com/stretchr/testify/http"
	"io/ioutil"
	"net/http"
	"testing"
)

func TestNewOptionSet(t *testing.T) {
	// without options
	set := newOptionSet()
	assert.NotEmpty(t, set.EntryName)
	assert.False(t, set.Skipper(echo.New().NewContext(nil, nil)))
	assert.Equal(t, DefaultCompression, set.Level)
	assert.NotNil(t, set.decompressPool)
	assert.NotNil(t, set.compressPool)

	// with level
	set = newOptionSet(
		WithEntryNameAndType("ut-name", "ut-type"),
		WithLevel(NoCompression),
		WithSkipper(func(context echo.Context) bool {
			return true
		}))
	assert.Equal(t, NoCompression, set.Level)
}

func TestNewCompressPool(t *testing.T) {
	// with DefaultCompression
	pool := newCompressPool(DefaultCompression)
	assert.NotNil(t, pool.delegate.Get())

	// with NoCompression
	pool = newCompressPool(NoCompression)
	assert.NotNil(t, pool.delegate.Get())

	// with DefaultCompression
	pool = newCompressPool(BestSpeed)
	assert.NotNil(t, pool.delegate.Get())

	// with DefaultCompression
	pool = newCompressPool(BestCompression)
	assert.NotNil(t, pool.delegate.Get())

	// with DefaultCompression
	pool = newCompressPool(DefaultCompression)
	assert.NotNil(t, pool.delegate.Get())

	// with DefaultCompression
	pool = newCompressPool(HuffmanOnly)
	assert.NotNil(t, pool.delegate.Get())

	// with DefaultCompression
	pool = newCompressPool("invalid")
	assert.NotNil(t, pool.delegate.Get())
}

func TestCompressPool_Get(t *testing.T) {
	pool := newCompressPool(DefaultCompression)
	assert.NotNil(t, pool.Get())
}

func TestCompressPool_Put(t *testing.T) {
	defer assertNotPanic(t)

	pool := newCompressPool(DefaultCompression)
	// put different types of value
	pool.Put(nil)
	pool.Put("string")
	pool.Put(1)
}

func TestDecompressPool_Get(t *testing.T) {
	pool := newDecompressPool()
	assert.NotNil(t, pool.Get())
}

func TestDecompressPool_Put(t *testing.T) {
	defer assertNotPanic(t)

	pool := newDecompressPool()
	// put different types of value
	pool.Put(nil)
	pool.Put("string")
	pool.Put(1)
}

func TestGzipResponseWriter(t *testing.T) {
	defer assertNotPanic(t)

	// WriteHeader() write header with http.StatusNoContent
	rw := httptest.TestResponseWriter{}
	w := new(bytes.Buffer)
	gzipRW := newGzipResponseWriter(w, &rw)
	gzipRW.WriteHeader(http.StatusNoContent)
	assert.Empty(t, rw.Header().Get(echo.HeaderContentEncoding))
	assert.Empty(t, rw.Header().Get(echo.HeaderContentLength))
	assert.Equal(t, http.StatusNoContent, rw.StatusCode)

	// WriteHeader() write header with other status code
	rw = httptest.TestResponseWriter{}
	w = new(bytes.Buffer)
	gzipRW = newGzipResponseWriter(w, &rw)
	gzipRW.WriteHeader(http.StatusOK)
	assert.Equal(t, http.StatusOK, rw.StatusCode)

	// Write() without Content-Type
	rw = httptest.TestResponseWriter{}
	w = new(bytes.Buffer)
	gzipRW = newGzipResponseWriter(w, &rw)
	gzipRW.Write([]byte("ut-message"))
	assert.NotEmpty(t, rw.Header().Get(echo.HeaderContentType))
	assert.NotEmpty(t, w.String())

	// Write() with Content-Type
	rw = httptest.TestResponseWriter{}
	w = new(bytes.Buffer)
	gzipRW = newGzipResponseWriter(w, &rw)
	rw.Header().Set(echo.HeaderContentType, "ut-type")
	gzipRW.Write([]byte("ut-message"))
	assert.NotEmpty(t, rw.Header().Get(echo.HeaderContentType))
	assert.NotEmpty(t, w.String())

	// Flush() with type of gzip.Writer
	rw = httptest.TestResponseWriter{}
	gw, _ := gzip.NewWriterLevel(ioutil.Discard, gzip.DefaultCompression)
	gzipRW = newGzipResponseWriter(gw, &rw)
	gzipRW.Flush()
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

// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkechogzip

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"github.com/labstack/echo/v4"
	rkmid "github.com/rookie-ninja/rk-entry/v2/middleware"
	"github.com/rs/xid"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"sync"
)

const (
	// GzipEncoding encoding type of gzip
	gzipEncoding = "gzip"
	// NoCompression copied from gzip.NoCompression
	NoCompression = "noCompression"
	// BestSpeed copied from gzip.BestSpeed
	BestSpeed = "bestSpeed"
	// BestCompression copied from gzip.BestCompression
	BestCompression = "bestCompression"
	// DefaultCompression copied from gzip.DefaultCompression
	DefaultCompression = "defaultCompression"
	// HuffmanOnly copied from gzip.HuffmanOnly
	HuffmanOnly = "huffmanOnly"
)

// Interceptor would distinguish auth set based on.
var (
	optionsMap     = make(map[string]*optionSet)
	defaultSkipper = func(echo.Context) bool {
		return false
	}
)

// Create new optionSet with rpc type nad options.
func newOptionSet(opts ...Option) *optionSet {
	set := &optionSet{
		EntryName:      xid.New().String(),
		EntryType:      "",
		Skipper:        defaultSkipper,
		Level:          DefaultCompression,
		decompressPool: newDecompressPool(),
	}

	for i := range opts {
		opts[i](set)
	}

	// create a new compressPool
	set.compressPool = newCompressPool(set.Level)

	if _, ok := optionsMap[set.EntryName]; !ok {
		optionsMap[set.EntryName] = set
	}

	return set
}

// Options which is used while initializing extension interceptor
type optionSet struct {
	EntryName      string
	EntryType      string
	Skipper        Skipper
	Level          string
	ignorePrefix   []string
	decompressPool *decompressPool
	compressPool   *compressPool
}

// ShouldIgnore determine whether auth should be ignored based on path
func (set *optionSet) ShouldIgnore(ctx echo.Context) bool {
	if ctx != nil && ctx.Request().URL != nil {
		for i := range set.ignorePrefix {
			if strings.HasPrefix(ctx.Request().URL.Path, set.ignorePrefix[i]) {
				return true
			}
		}

		return rkmid.ShouldIgnoreGlobal(ctx.Request().URL.Path)
	}

	return false
}

// Option if for middleware options while creating middleware
type Option func(*optionSet)

// WithEntryNameAndType provide entry name and entry type.
func WithEntryNameAndType(entryName, entryType string) Option {
	return func(opt *optionSet) {
		opt.EntryName = entryName
		opt.EntryType = entryType
	}
}

// WithLevel provide level of compressing.
func WithLevel(level string) Option {
	return func(opt *optionSet) {
		opt.Level = level
	}
}

// WithSkipper provide skipper.
func WithSkipper(skip Skipper) Option {
	return func(opt *optionSet) {
		opt.Skipper = skip
	}
}

// WithPathToIgnore provide path prefix to ignore middleware
func WithPathToIgnore(prefix ...string) Option {
	return func(opt *optionSet) {
		opt.ignorePrefix = append(opt.ignorePrefix, prefix...)
	}
}

// sync.Pool is the delegate of this pool
type compressPool struct {
	delegate *sync.Pool
}

// Create a new compress pool
func newCompressPool(level string) *compressPool {
	levelLowerCase := strings.ToLower(level)

	levelInt := gzip.DefaultCompression

	switch levelLowerCase {
	case strings.ToLower(NoCompression):
		levelInt = gzip.NoCompression
	case strings.ToLower(BestSpeed):
		levelInt = gzip.BestSpeed
	case strings.ToLower(BestCompression):
		levelInt = gzip.BestCompression
	case strings.ToLower(DefaultCompression):
		levelInt = gzip.DefaultCompression
	case strings.ToLower(HuffmanOnly):
		levelInt = gzip.HuffmanOnly
	default:
		levelInt = gzip.DefaultCompression
	}

	return &compressPool{
		delegate: &sync.Pool{
			New: func() interface{} {
				// Ok to ignore error because of above switch statement
				writer, _ := gzip.NewWriterLevel(ioutil.Discard, levelInt)
				return writer
			},
		},
	}
}

// Get item gzip.Writer from pool
func (p *compressPool) Get() *gzip.Writer {
	// assert no error
	raw := p.delegate.Get()

	switch raw.(type) {
	case *gzip.Writer:
		return raw.(*gzip.Writer)
	}

	return nil
}

// Put item gzip.Writer back to pool
func (p *compressPool) Put(x interface{}) {
	p.delegate.Put(x)
}

// sync.Pool is the delegate of this pool
type decompressPool struct {
	delegate *sync.Pool
}

// Create a new decompress pool
func newDecompressPool() *decompressPool {
	pool := &sync.Pool{
		New: func() interface{} {
			// In order to create a gzip.Reader, we need to pass a bytes with format gzip.
			// Create a gzip.Writer is the easiest way to achieve this goal.
			writer, _ := gzip.NewWriterLevel(ioutil.Discard, gzip.DefaultCompression)
			b := new(bytes.Buffer)
			writer.Reset(b)
			writer.Flush()
			writer.Close()

			// Create a reader, ignoring error since we created a empty writer
			reader, _ := gzip.NewReader(bytes.NewReader(b.Bytes()))
			return reader
		},
	}

	return &decompressPool{
		delegate: pool,
	}
}

// Get item gzip.Reader from pool
func (p *decompressPool) Get() *gzip.Reader {
	// assert no error
	raw := p.delegate.Get()

	switch raw.(type) {
	case *gzip.Reader:
		return raw.(*gzip.Reader)
	}

	return nil
}

// Put item gzip.Reader back to pool
func (p *decompressPool) Put(x interface{}) {
	p.delegate.Put(x)
}

// Skipper default skipper will always return false
type Skipper func(echo.Context) bool

// Copied from https://github.com/labstack/echo/blob/master/middleware/compress.go
//
// Why not use middleware.GzipWithConfig directly?
//
// rk-echo support multi-entries of echo framework. In order to match rk-echo architecture,
// we need to modify some of logic in middleware.
type gzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

func newGzipResponseWriter(w io.Writer, rw http.ResponseWriter) *gzipResponseWriter {
	return &gzipResponseWriter{
		Writer:         w,
		ResponseWriter: rw,
	}
}

// WriteHeader writes header into http.ResponseWriter
func (w *gzipResponseWriter) WriteHeader(code int) {
	if code == http.StatusNoContent {
		w.ResponseWriter.Header().Del(echo.HeaderContentEncoding)
	}
	w.Header().Del(echo.HeaderContentLength)
	w.ResponseWriter.WriteHeader(code)
}

// Write writes bytes into gzipWriter.
func (w *gzipResponseWriter) Write(b []byte) (int, error) {
	if w.Header().Get(echo.HeaderContentType) == "" {
		w.Header().Set(echo.HeaderContentType, http.DetectContentType(b))
	}

	return w.Writer.Write(b)
}

// Flush flushes contents in http.ResponseWriter.
func (w *gzipResponseWriter) Flush() {
	w.Writer.(*gzip.Writer).Flush()
	if flusher, ok := w.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

// Hijack hijack http.ResponseWriter
func (w *gzipResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return w.ResponseWriter.(http.Hijacker).Hijack()
}

// Push pushes target to http.ResponseWriter
func (w *gzipResponseWriter) Push(target string, opts *http.PushOptions) error {
	if p, ok := w.ResponseWriter.(http.Pusher); ok {
		return p.Push(target, opts)
	}
	return http.ErrNotSupported
}

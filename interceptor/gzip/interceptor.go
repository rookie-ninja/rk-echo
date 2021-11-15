// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkechogzip

import (
	"bytes"
	"github.com/labstack/echo/v4"
	rkerror "github.com/rookie-ninja/rk-common/error"
	"github.com/rookie-ninja/rk-echo/interceptor"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
)

// Interceptor Add gzip compress and decompress interceptors.
//
// Mainly copied from bellow.
// https://github.com/labstack/echo/blob/master/middleware/decompress.go
// https://github.com/labstack/echo/blob/master/middleware/compress.go
func Interceptor(opts ...Option) echo.MiddlewareFunc {
	set := newOptionSet(opts...)

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			ctx.Set(rkechointer.RpcEntryNameKey, set.EntryName)

			if set.Skipper(ctx) {
				return next(ctx)
			}

			// deal with request decompression
			switch ctx.Request().Header.Get(echo.HeaderContentEncoding) {
			case gzipEncoding:
				gzipReader := set.decompressPool.Get()

				// make gzipReader to read from original request body
				if err := gzipReader.Reset(ctx.Request().Body); err != nil {
					// return reader back to sync.Pool
					set.decompressPool.Put(gzipReader)

					// body is empty, keep on going
					if err == io.EOF {
						return next(ctx)
					}

					return rkerror.New(
						rkerror.WithHttpCode(http.StatusInternalServerError),
						rkerror.WithDetails(err)).Err
				}

				// create a buffer and copy decompressed data into it via gzipReader
				var buf bytes.Buffer
				if _, err := io.Copy(&buf, gzipReader); err != nil {
					return rkerror.New(
						rkerror.WithHttpCode(http.StatusInternalServerError),
						rkerror.WithDetails(err)).Err
				}

				// close both gzipReader and original reader in request body
				gzipReader.Close()
				ctx.Request().Body.Close()
				set.decompressPool.Put(gzipReader)

				// assign decompressed buffer to request
				ctx.Request().Body = ioutil.NopCloser(&buf)
			}

			// deal with response compression
			ctx.Response().Header().Add(echo.HeaderVary, echo.HeaderAcceptEncoding)
			// gzip is one of expected encoding type from request
			if strings.Contains(ctx.Request().Header.Get(echo.HeaderAcceptEncoding), gzipEncoding) {
				// set to response header
				ctx.Response().Header().Set(echo.HeaderContentEncoding, gzipEncoding)

				// create gzip writer
				gzipWriter := set.compressPool.Get()

				// reset writer of gzip writer to original writer from response
				originalWriter := ctx.Response().Writer
				gzipWriter.Reset(originalWriter)

				// defer func
				defer func() {
					if ctx.Response().Size == 0 {
						// remove encoding header if response is empty
						if ctx.Response().Header().Get(echo.HeaderContentEncoding) == gzipEncoding {
							ctx.Response().Header().Del(echo.HeaderContentEncoding)
						}
						// we have to reset response to it's pristine state when
						// nothing is written to body or error is returned.
						ctx.Response().Writer = originalWriter

						// reset to empty
						gzipWriter.Reset(ioutil.Discard)
					}

					// close gzipWriter
					gzipWriter.Close()

					// put gzipWriter back to pool
					set.compressPool.Put(gzipWriter)
				}()

				// assign new writer to response
				ctx.Response().Writer = newGzipResponseWriter(gzipWriter, originalWriter)
			}

			return next(ctx)
		}
	}
}

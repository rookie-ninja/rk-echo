// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkechosec

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/rookie-ninja/rk-echo/interceptor"
)

// Interceptor Add security interceptors.
//
// Mainly copied from bellow.
// https://github.com/labstack/echo/blob/master/middleware/secure.go
func Interceptor(opts ...Option) echo.MiddlewareFunc {
	set := newOptionSet(opts...)

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			ctx.Set(rkechointer.RpcEntryNameKey, set.EntryName)

			if set.Skipper(ctx) {
				return next(ctx)
			}

			req := ctx.Request()
			res := ctx.Response()

			// Add X-XSS-Protection header
			if set.XSSProtection != "" {
				res.Header().Set(headerXXSSProtection, set.XSSProtection)
			}

			// Add X-Content-Type-Options header
			if set.ContentTypeNosniff != "" {
				res.Header().Set(headerXContentTypeOptions, set.ContentTypeNosniff)
			}

			// Add X-Frame-Options header
			if set.XFrameOptions != "" {
				res.Header().Set(headerXFrameOptions, set.XFrameOptions)
			}

			// Add Strict-Transport-Security header
			if (ctx.IsTLS() || (req.Header.Get(headerXForwardedProto) == "https")) && set.HSTSMaxAge != 0 {
				subdomains := ""
				if !set.HSTSExcludeSubdomains {
					subdomains = "; includeSubdomains"
				}
				if set.HSTSPreloadEnabled {
					subdomains = fmt.Sprintf("%s; preload", subdomains)
				}
				res.Header().Set(headerStrictTransportSecurity, fmt.Sprintf("max-age=%d%s", set.HSTSMaxAge, subdomains))
			}

			// Add Content-Security-Policy-Report-Only or Content-Security-Policy header
			if set.ContentSecurityPolicy != "" {
				if set.CSPReportOnly {
					res.Header().Set(headerContentSecurityPolicyReportOnly, set.ContentSecurityPolicy)
				} else {
					res.Header().Set(headerContentSecurityPolicy, set.ContentSecurityPolicy)
				}
			}

			// Add Referrer-Policy header
			if set.ReferrerPolicy != "" {
				res.Header().Set(headerReferrerPolicy, set.ReferrerPolicy)
			}

			return next(ctx)
		}
	}
}

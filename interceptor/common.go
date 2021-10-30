// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

// Package rkechointer provides common utility functions for middleware of echo framework
package rkechointer

import (
	"github.com/labstack/echo/v4"
	"github.com/rookie-ninja/rk-common/common"
	"go.uber.org/zap"
	"net"
	"strings"
)

var (
	// Realm environment variable
	Realm = zap.String("realm", rkcommon.GetEnvValueOrDefault("REALM", "*"))
	// Region environment variable
	Region = zap.String("region", rkcommon.GetEnvValueOrDefault("REGION", "*"))
	// AZ environment variable
	AZ = zap.String("az", rkcommon.GetEnvValueOrDefault("AZ", "*"))
	// Domain environment variable
	Domain = zap.String("domain", rkcommon.GetEnvValueOrDefault("DOMAIN", "*"))
	// LocalIp read local IP from localhost
	LocalIp = zap.String("localIp", rkcommon.GetLocalIP())
	// LocalHostname read hostname from localhost
	LocalHostname = zap.String("localHostname", rkcommon.GetLocalHostname())
)

const (
	// RpcEntryNameKey entry name key
	RpcEntryNameKey = "echoEntryName"
	// RpcEntryNameValue entry name
	RpcEntryNameValue = "echo"
	// RpcEntryTypeValue entry type
	RpcEntryTypeValue = "echo"
	// RpcEventKey event key
	RpcEventKey = "echoEvent"
	// RpcLoggerKey logger key
	RpcLoggerKey = "echoLogger"
	// RpcTracerKey tracer key
	RpcTracerKey = "echoTracer"
	// RpcSpanKey span key
	RpcSpanKey = "echoSpan"
	// RpcTracerProviderKey trace provider key
	RpcTracerProviderKey = "echoTracerProvider"
	// RpcPropagatorKey propagator key
	RpcPropagatorKey = "echoPropagator"
	// RpcAuthorizationHeaderKey auth key
	RpcAuthorizationHeaderKey = "authorization"
	// RpcApiKeyHeaderKey api auth key
	RpcApiKeyHeaderKey = "X-API-Key"
)

// GetRemoteAddressSet returns remote endpoint information set including IP, Port.
// We will do as best as we can to determine it.
// If fails, then just return default ones.
func GetRemoteAddressSet(ctx echo.Context) (remoteIp, remotePort string) {
	remoteIp, remotePort = "0.0.0.0", "0"

	if ctx == nil || ctx.Request == nil {
		return
	}

	var err error
	if remoteIp, remotePort, err = net.SplitHostPort(ctx.Request().RemoteAddr); err != nil {
		return
	}

	forwardedRemoteIp := ctx.Request().Header.Get("x-forwarded-for")

	// Deal with forwarded remote ip
	if len(forwardedRemoteIp) > 0 {
		if forwardedRemoteIp == "::1" {
			forwardedRemoteIp = "localhost"
		}

		remoteIp = forwardedRemoteIp
	}

	if remoteIp == "::1" {
		remoteIp = "localhost"
	}

	return remoteIp, remotePort
}

// ShouldLog determines whether should log the RPC
func ShouldLog(ctx echo.Context) bool {
	if ctx == nil || ctx.Request() == nil {
		return false
	}

	// ignoring /rk/v1/assets, /rk/v1/tv and /sw/ path while logging since these are internal APIs.
	if strings.HasPrefix(ctx.Request().URL.Path, "/rk/v1/assets") ||
		strings.HasPrefix(ctx.Request().URL.Path, "/rk/v1/tv") ||
		strings.HasPrefix(ctx.Request().URL.Path, "/sw/") {
		return false
	}

	return true
}

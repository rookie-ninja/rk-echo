// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

// Package rkecho an implementation of rkentry.Entry which could be used start restful server with echo framework
package rkecho

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rookie-ninja/rk-common/common"
	"github.com/rookie-ninja/rk-echo/interceptor/auth"
	rkechoctx "github.com/rookie-ninja/rk-echo/interceptor/context"
	"github.com/rookie-ninja/rk-echo/interceptor/cors"
	"github.com/rookie-ninja/rk-echo/interceptor/csrf"
	"github.com/rookie-ninja/rk-echo/interceptor/gzip"
	"github.com/rookie-ninja/rk-echo/interceptor/jwt"
	"github.com/rookie-ninja/rk-echo/interceptor/log/zap"
	"github.com/rookie-ninja/rk-echo/interceptor/meta"
	"github.com/rookie-ninja/rk-echo/interceptor/metrics/prom"
	rkechopanic "github.com/rookie-ninja/rk-echo/interceptor/panic"
	"github.com/rookie-ninja/rk-echo/interceptor/ratelimit"
	"github.com/rookie-ninja/rk-echo/interceptor/secure"
	"github.com/rookie-ninja/rk-echo/interceptor/timeout"
	"github.com/rookie-ninja/rk-echo/interceptor/tracing/telemetry"
	"github.com/rookie-ninja/rk-entry/entry"
	rkmidauth "github.com/rookie-ninja/rk-entry/middleware/auth"
	rkmidcors "github.com/rookie-ninja/rk-entry/middleware/cors"
	rkmidcsrf "github.com/rookie-ninja/rk-entry/middleware/csrf"
	rkmidjwt "github.com/rookie-ninja/rk-entry/middleware/jwt"
	rkmidlog "github.com/rookie-ninja/rk-entry/middleware/log"
	rkmidmeta "github.com/rookie-ninja/rk-entry/middleware/meta"
	rkmidmetrics "github.com/rookie-ninja/rk-entry/middleware/metrics"
	rkmidpanic "github.com/rookie-ninja/rk-entry/middleware/panic"
	rkmidlimit "github.com/rookie-ninja/rk-entry/middleware/ratelimit"
	rkmidsec "github.com/rookie-ninja/rk-entry/middleware/secure"
	rkmidtimeout "github.com/rookie-ninja/rk-entry/middleware/timeout"
	rkmidtrace "github.com/rookie-ninja/rk-entry/middleware/tracing"
	"github.com/rookie-ninja/rk-query"
	"go.uber.org/zap"
	"net/http"
	"path"
	"strconv"
	"strings"
)

const (
	// EchoEntryType type of entry
	EchoEntryType = "EchoEntry"
	// EchoEntryDescription description of entry
	EchoEntryDescription = "Internal RK entry which helps to bootstrap with Echo framework."
)

// This must be declared in order to register registration function into rk context
// otherwise, rk-boot won't able to bootstrap echo entry automatically from boot config file
func init() {
	rkentry.RegisterEntryRegFunc(RegisterEchoEntriesWithConfig)
}

// BootConfig boot config which is for echo entry.
type BootConfig struct {
	Echo []struct {
		Enabled     bool   `yaml:"enabled" json:"enabled"`
		Name        string `yaml:"name" json:"name"`
		Port        uint64 `yaml:"port" json:"port"`
		Description string `yaml:"description" json:"description"`
		Cert        struct {
			Ref string `yaml:"ref" json:"ref"`
		} `yaml:"cert" json:"cert"`
		SW            rkentry.BootConfigSw            `yaml:"sw" json:"sw"`
		CommonService rkentry.BootConfigCommonService `yaml:"commonService" json:"commonService"`
		TV            rkentry.BootConfigTv            `yaml:"tv" json:"tv"`
		Prom          rkentry.BootConfigProm          `yaml:"prom" json:"prom"`
		Static        rkentry.BootConfigStaticHandler `yaml:"static" json:"static"`
		Interceptors  struct {
			LoggingZap  rkmidlog.BootConfig     `yaml:"loggingZap" json:"loggingZap"`
			MetricsProm rkmidmetrics.BootConfig `yaml:"metricsProm" json:"metricsProm"`
			Auth        rkmidauth.BootConfig    `yaml:"auth" json:"auth"`
			Cors        rkmidcors.BootConfig    `yaml:"cors" json:"cors"`
			Meta        rkmidmeta.BootConfig    `yaml:"meta" json:"meta"`
			Jwt         rkmidjwt.BootConfig     `yaml:"jwt" json:"jwt"`
			Secure      rkmidsec.BootConfig     `yaml:"secure" json:"secure"`
			RateLimit   rkmidlimit.BootConfig   `yaml:"rateLimit" json:"rateLimit"`
			Csrf        rkmidcsrf.BootConfig    `yaml:"csrf" yaml:"csrf"`
			Gzip        struct {
				Enabled bool   `yaml:"enabled" json:"enabled"`
				Level   string `yaml:"level" json:"level"`
			} `yaml:"gzip" json:"gzip"`
			Timeout          rkmidtimeout.BootConfig `yaml:"timeout" json:"timeout"`
			TracingTelemetry rkmidtrace.BootConfig   `yaml:"tracingTelemetry" json:"tracingTelemetry"`
		} `yaml:"interceptors" json:"interceptors"`
		Logger struct {
			ZapLogger struct {
				Ref string `yaml:"ref" json:"ref"`
			} `yaml:"zapLogger" json:"zapLogger"`
			EventLogger struct {
				Ref string `yaml:"ref" json:"ref"`
			} `yaml:"eventLogger" json:"eventLogger"`
		} `yaml:"logger" json:"logger"`
	} `yaml:"echo" json:"echo"`
}

// EchoEntry implements rkentry.Entry interface.
type EchoEntry struct {
	EntryName          string                          `json:"entryName" yaml:"entryName"`
	EntryType          string                          `json:"entryType" yaml:"entryType"`
	EntryDescription   string                          `json:"-" yaml:"-"`
	ZapLoggerEntry     *rkentry.ZapLoggerEntry         `json:"-" yaml:"-"`
	EventLoggerEntry   *rkentry.EventLoggerEntry       `json:"-" yaml:"-"`
	Port               uint64                          `json:"port" yaml:"port"`
	CertEntry          *rkentry.CertEntry              `json:"-" yaml:"-"`
	SwEntry            *rkentry.SwEntry                `json:"-" yaml:"-"`
	CommonServiceEntry *rkentry.CommonServiceEntry     `json:"-" yaml:"-"`
	Echo               *echo.Echo                      `json:"-" yaml:"-"`
	PromEntry          *rkentry.PromEntry              `json:"-" yaml:"-"`
	StaticFileEntry    *rkentry.StaticFileHandlerEntry `json:"-" yaml:"-"`
	TvEntry            *rkentry.TvEntry                `json:"-" yaml:"-"`
}

// RegisterEchoEntriesWithConfig register echo entries with provided config file (Must YAML file).
//
// Currently, support two ways to provide config file path.
// 1: With function parameters
// 2: With command line flag "--rkboot" described in rkcommon.BootConfigPathFlagKey (Will override function parameter if exists)
// Command line flag has high priority which would override function parameter
//
// Error handling:
// Process will shutdown if any errors occur with rkcommon.ShutdownWithError function
//
// Override elements in config file:
// We learned from HELM source code which would override elements in YAML file with "--set" flag followed with comma
// separated key/value pairs.
//
// We are using "--rkset" described in rkcommon.BootConfigOverrideKey in order to distinguish with user flags
// Example of common usage: ./binary_file --rkset "key1=val1,key2=val2"
// Example of nested map:   ./binary_file --rkset "outer.inner.key=val"
// Example of slice:        ./binary_file --rkset "outer[0].key=val"
func RegisterEchoEntriesWithConfig(configFilePath string) map[string]rkentry.Entry {
	res := make(map[string]rkentry.Entry)

	// 1: Decode config map into boot config struct
	config := &BootConfig{}
	rkcommon.UnmarshalBootConfig(configFilePath, config)

	// 2: Init echo entries with boot config
	for i := range config.Echo {
		element := config.Echo[i]
		if !element.Enabled {
			continue
		}

		name := element.Name

		zapLoggerEntry := rkentry.GlobalAppCtx.GetZapLoggerEntry(element.Logger.ZapLogger.Ref)
		if zapLoggerEntry == nil {
			zapLoggerEntry = rkentry.GlobalAppCtx.GetZapLoggerEntryDefault()
		}

		eventLoggerEntry := rkentry.GlobalAppCtx.GetEventLoggerEntry(element.Logger.EventLogger.Ref)
		if eventLoggerEntry == nil {
			eventLoggerEntry = rkentry.GlobalAppCtx.GetEventLoggerEntryDefault()
		}

		// Register swagger entry
		swEntry := rkentry.RegisterSwEntryWithConfig(&element.SW, element.Name, element.Port,
			zapLoggerEntry, eventLoggerEntry, element.CommonService.Enabled)

		// Register prometheus entry
		promRegistry := prometheus.NewRegistry()
		promEntry := rkentry.RegisterPromEntryWithConfig(&element.Prom, element.Name, element.Port,
			zapLoggerEntry, eventLoggerEntry, promRegistry)

		// Register common service entry
		commonServiceEntry := rkentry.RegisterCommonServiceEntryWithConfig(&element.CommonService, element.Name,
			zapLoggerEntry, eventLoggerEntry)

		// Register TV entry
		tvEntry := rkentry.RegisterTvEntryWithConfig(&element.TV, element.Name,
			zapLoggerEntry, eventLoggerEntry)

		// Register static file handler
		staticEntry := rkentry.RegisterStaticFileHandlerEntryWithConfig(&element.Static, element.Name,
			zapLoggerEntry, eventLoggerEntry)

		inters := make([]echo.MiddlewareFunc, 0)

		// logging middlewares
		if element.Interceptors.LoggingZap.Enabled {
			inters = append(inters, rkecholog.Interceptor(
				rkmidlog.ToOptions(&element.Interceptors.LoggingZap, element.Name, EchoEntryType,
					zapLoggerEntry, eventLoggerEntry)...))
		}

		// metrics middleware
		if element.Interceptors.MetricsProm.Enabled {
			inters = append(inters, rkechometrics.Interceptor(
				rkmidmetrics.ToOptions(&element.Interceptors.MetricsProm, element.Name, EchoEntryType,
					promRegistry, rkmidmetrics.LabelerTypeHttp)...))
		}

		// tracing middleware
		if element.Interceptors.TracingTelemetry.Enabled {
			inters = append(inters, rkechotrace.Interceptor(
				rkmidtrace.ToOptions(&element.Interceptors.TracingTelemetry, element.Name, EchoEntryType)...))
		}

		// jwt middleware
		if element.Interceptors.Jwt.Enabled {
			inters = append(inters, rkechojwt.Interceptor(
				rkmidjwt.ToOptions(&element.Interceptors.Jwt, element.Name, EchoEntryType)...))
		}

		// secure middleware
		if element.Interceptors.Secure.Enabled {
			inters = append(inters, rkechosec.Interceptor(
				rkmidsec.ToOptions(&element.Interceptors.Secure, element.Name, EchoEntryType)...))
		}

		// csrf middleware
		if element.Interceptors.Csrf.Enabled {
			inters = append(inters, rkechocsrf.Interceptor(
				rkmidcsrf.ToOptions(&element.Interceptors.Csrf, element.Name, EchoEntryType)...))
		}

		// cors middleware
		if element.Interceptors.Cors.Enabled {
			inters = append(inters, rkechocors.Interceptor(
				rkmidcors.ToOptions(&element.Interceptors.Cors, element.Name, EchoEntryType)...))
		}

		// gzip middleware
		if element.Interceptors.Gzip.Enabled {
			opts := []rkechogzip.Option{
				rkechogzip.WithEntryNameAndType(element.Name, EchoEntryType),
				rkechogzip.WithLevel(element.Interceptors.Gzip.Level),
			}

			inters = append(inters, rkechogzip.Interceptor(opts...))
		}

		// meta middleware
		if element.Interceptors.Meta.Enabled {
			inters = append(inters, rkechometa.Interceptor(
				rkmidmeta.ToOptions(&element.Interceptors.Meta, element.Name, EchoEntryType)...))
		}

		// auth middlewares
		if element.Interceptors.Auth.Enabled {
			inters = append(inters, rkechoauth.Interceptor(
				rkmidauth.ToOptions(&element.Interceptors.Auth, element.Name, EchoEntryType)...))
		}

		// timeout middlewares
		if element.Interceptors.Timeout.Enabled {
			inters = append(inters, rkechotimeout.Interceptor(
				rkmidtimeout.ToOptions(&element.Interceptors.Timeout, element.Name, EchoEntryType)...))
		}

		// rate limit middleware
		if element.Interceptors.RateLimit.Enabled {
			inters = append(inters, rkecholimit.Interceptor(
				rkmidlimit.ToOptions(&element.Interceptors.RateLimit, element.Name, EchoEntryType)...))
		}

		certEntry := rkentry.GlobalAppCtx.GetCertEntry(element.Cert.Ref)

		entry := RegisterEchoEntry(
			WithName(name),
			WithDescription(element.Description),
			WithPort(element.Port),
			WithZapLoggerEntry(zapLoggerEntry),
			WithEventLoggerEntry(eventLoggerEntry),
			WithCertEntry(certEntry),
			WithPromEntry(promEntry),
			WithTvEntry(tvEntry),
			WithCommonServiceEntry(commonServiceEntry),
			WithSwEntry(swEntry),
			WithStaticFileHandlerEntry(staticEntry))

		entry.AddInterceptor(inters...)

		res[name] = entry
	}

	return res
}

// RegisterEchoEntry register EchoEntry with options.
func RegisterEchoEntry(opts ...EchoEntryOption) *EchoEntry {
	entry := &EchoEntry{
		EntryType:        EchoEntryType,
		EntryDescription: EchoEntryDescription,
		Port:             8080,
	}

	for i := range opts {
		opts[i](entry)
	}

	if entry.ZapLoggerEntry == nil {
		entry.ZapLoggerEntry = rkentry.GlobalAppCtx.GetZapLoggerEntryDefault()
	}

	if entry.EventLoggerEntry == nil {
		entry.EventLoggerEntry = rkentry.GlobalAppCtx.GetEventLoggerEntryDefault()
	}

	if len(entry.EntryName) < 1 {
		entry.EntryName = "EchoServer-" + strconv.FormatUint(entry.Port, 10)
	}

	if entry.Echo == nil {
		entry.Echo = echo.New()
		entry.Echo.HidePort = true
		entry.Echo.HideBanner = true
	}

	// insert panic interceptor
	entry.Echo.Use(rkechopanic.Interceptor(
		rkmidpanic.WithEntryNameAndType(entry.EntryName, entry.EntryType)))

	rkentry.GlobalAppCtx.AddEntry(entry)

	return entry
}

// GetName Get entry name.
func (entry *EchoEntry) GetName() string {
	return entry.EntryName
}

// GetType Get entry type.
func (entry *EchoEntry) GetType() string {
	return entry.EntryType
}

// GetDescription Get description of entry.
func (entry *EchoEntry) GetDescription() string {
	return entry.EntryDescription
}

// Bootstrap EchoEntry.
func (entry *EchoEntry) Bootstrap(ctx context.Context) {
	event, logger := entry.logBasicInfo("Bootstrap")

	// Is swagger enabled?
	if entry.IsSwEnabled() {
		// Register swagger path into Router.
		entry.Echo.GET(strings.TrimSuffix(entry.SwEntry.Path, "/"), func(ctx echo.Context) error {
			ctx.Redirect(http.StatusTemporaryRedirect, entry.SwEntry.Path)
			return nil
		})
		entry.Echo.GET(path.Join(entry.SwEntry.Path, "*"), echo.WrapHandler(entry.SwEntry.ConfigFileHandler()))
		entry.Echo.GET(path.Join(entry.SwEntry.AssetsFilePath, "*"), echo.WrapHandler(entry.SwEntry.AssetsFileHandler()))
		entry.SwEntry.Bootstrap(ctx)
	}

	// Is static file handler enabled?
	if entry.IsStaticFileHandlerEnabled() {
		// Register path into Router.
		entry.Echo.GET(strings.TrimSuffix(entry.StaticFileEntry.Path, "/"), func(ctx echo.Context) error {
			ctx.Redirect(http.StatusTemporaryRedirect, entry.StaticFileEntry.Path)
			return nil
		})

		// Register path into Router.
		entry.Echo.GET(path.Join(entry.StaticFileEntry.Path, "*"), echo.WrapHandler(entry.StaticFileEntry.GetFileHandler()))
		entry.StaticFileEntry.Bootstrap(ctx)
	}

	// Is prometheus enabled?
	if entry.IsPromEnabled() {
		// Register prom path into Router.
		entry.Echo.GET(entry.PromEntry.Path, echo.WrapHandler(promhttp.HandlerFor(entry.PromEntry.Gatherer, promhttp.HandlerOpts{})))

		// don't start with http handler, we will handle it by ourselves
		entry.PromEntry.Bootstrap(ctx)
	}

	// Is common service enabled?
	if entry.IsCommonServiceEnabled() {
		// Register common service path into Router.
		entry.Echo.GET(entry.CommonServiceEntry.HealthyPath, echo.WrapHandler(http.HandlerFunc(entry.CommonServiceEntry.Healthy)))
		entry.Echo.GET(entry.CommonServiceEntry.GcPath, echo.WrapHandler(http.HandlerFunc(entry.CommonServiceEntry.Gc)))
		entry.Echo.GET(entry.CommonServiceEntry.InfoPath, echo.WrapHandler(http.HandlerFunc(entry.CommonServiceEntry.Info)))
		entry.Echo.GET(entry.CommonServiceEntry.ConfigsPath, echo.WrapHandler(http.HandlerFunc(entry.CommonServiceEntry.Configs)))
		entry.Echo.GET(entry.CommonServiceEntry.SysPath, echo.WrapHandler(http.HandlerFunc(entry.CommonServiceEntry.Sys)))
		entry.Echo.GET(entry.CommonServiceEntry.EntriesPath, echo.WrapHandler(http.HandlerFunc(entry.CommonServiceEntry.Entries)))
		entry.Echo.GET(entry.CommonServiceEntry.CertsPath, echo.WrapHandler(http.HandlerFunc(entry.CommonServiceEntry.Certs)))
		entry.Echo.GET(entry.CommonServiceEntry.LogsPath, echo.WrapHandler(http.HandlerFunc(entry.CommonServiceEntry.Logs)))
		entry.Echo.GET(entry.CommonServiceEntry.DepsPath, echo.WrapHandler(http.HandlerFunc(entry.CommonServiceEntry.Deps)))
		entry.Echo.GET(entry.CommonServiceEntry.LicensePath, echo.WrapHandler(http.HandlerFunc(entry.CommonServiceEntry.License)))
		entry.Echo.GET(entry.CommonServiceEntry.ReadmePath, echo.WrapHandler(http.HandlerFunc(entry.CommonServiceEntry.Readme)))
		entry.Echo.GET(entry.CommonServiceEntry.GitPath, echo.WrapHandler(http.HandlerFunc(entry.CommonServiceEntry.Git)))

		// swagger doc already generated at rkentry.CommonService
		// follow bellow actions
		entry.Echo.GET(entry.CommonServiceEntry.ApisPath, entry.Apis)
		entry.Echo.GET(entry.CommonServiceEntry.ReqPath, entry.Req)

		// Bootstrap common service entry.
		entry.CommonServiceEntry.Bootstrap(ctx)
	}

	// Is TV enabled?
	if entry.IsTvEnabled() {
		// Bootstrap TV entry.
		entry.Echo.GET(strings.TrimSuffix(entry.TvEntry.BasePath, "/"), func(ctx echo.Context) error {
			ctx.Redirect(http.StatusTemporaryRedirect, entry.TvEntry.BasePath)
			return nil
		})
		entry.Echo.GET(path.Join(entry.TvEntry.BasePath, "*"), entry.TV)
		entry.Echo.GET(path.Join(entry.TvEntry.AssetsFilePath, "*"), echo.WrapHandler(entry.TvEntry.AssetsFileHandler()))

		entry.TvEntry.Bootstrap(ctx)
	}

	// Start echo server
	go entry.startServer(event, logger)

	entry.EventLoggerEntry.GetEventHelper().Finish(event)
}

// Interrupt EchoEntry.
func (entry *EchoEntry) Interrupt(ctx context.Context) {
	event, logger := entry.logBasicInfo("Interrupt")

	if entry.IsSwEnabled() {
		// Interrupt swagger entry
		entry.SwEntry.Interrupt(ctx)
	}

	if entry.IsStaticFileHandlerEnabled() {
		// Interrupt entry
		entry.StaticFileEntry.Interrupt(ctx)
	}

	if entry.IsPromEnabled() {
		// Interrupt prometheus entry
		entry.PromEntry.Interrupt(ctx)
	}

	if entry.IsCommonServiceEnabled() {
		// Interrupt common service entry
		entry.CommonServiceEntry.Interrupt(ctx)
	}

	if entry.IsTvEnabled() {
		// Interrupt common service entry
		entry.TvEntry.Interrupt(ctx)
	}

	if entry.Echo != nil {
		if err := entry.Echo.Shutdown(context.Background()); err != nil && err != http.ErrServerClosed {
			event.AddErr(err)
			logger.Warn("Error occurs while stopping echo-server.", event.ListPayloads()...)
		}
	}

	entry.EventLoggerEntry.GetEventHelper().Finish(event)
}

// String Stringfy entry.
func (entry *EchoEntry) String() string {
	bytes, _ := json.Marshal(entry)
	return string(bytes)
}

// ***************** Stringfy *****************

// MarshalJSON Marshal entry.
func (entry *EchoEntry) MarshalJSON() ([]byte, error) {
	m := map[string]interface{}{
		"entryName":          entry.EntryName,
		"entryType":          entry.EntryType,
		"entryDescription":   entry.EntryDescription,
		"eventLoggerEntry":   entry.EventLoggerEntry.GetName(),
		"zapLoggerEntry":     entry.ZapLoggerEntry.GetName(),
		"port":               entry.Port,
		"swEntry":            entry.SwEntry,
		"commonServiceEntry": entry.CommonServiceEntry,
		"promEntry":          entry.PromEntry,
		"tvEntry":            entry.TvEntry,
	}

	if entry.CertEntry != nil {
		m["certEntry"] = entry.CertEntry.GetName()
	}

	interceptorsStr := make([]string, 0)
	m["interceptors"] = &interceptorsStr

	return json.Marshal(&m)
}

// UnmarshalJSON Not supported.
func (entry *EchoEntry) UnmarshalJSON([]byte) error {
	return nil
}

// ***************** Public functions *****************

// GetEchoEntry Get EchoEntry from rkentry.GlobalAppCtx.
func GetEchoEntry(name string) *EchoEntry {
	entryRaw := rkentry.GlobalAppCtx.GetEntry(name)
	if entryRaw == nil {
		return nil
	}

	entry, _ := entryRaw.(*EchoEntry)
	return entry
}

// AddInterceptor Add interceptors.
// This function should be called before Bootstrap() called.
func (entry *EchoEntry) AddInterceptor(inters ...echo.MiddlewareFunc) {
	entry.Echo.Use(inters...)
}

// IsTlsEnabled Is TLS enabled?
func (entry *EchoEntry) IsTlsEnabled() bool {
	return entry.CertEntry != nil && entry.CertEntry.Store != nil
}

// IsSwEnabled Is swagger entry enabled?
func (entry *EchoEntry) IsSwEnabled() bool {
	return entry.SwEntry != nil
}

// IsCommonServiceEnabled Is common service entry enabled?
func (entry *EchoEntry) IsCommonServiceEnabled() bool {
	return entry.CommonServiceEntry != nil
}

// IsTvEnabled Is TV entry enabled?
func (entry *EchoEntry) IsTvEnabled() bool {
	return entry.TvEntry != nil
}

// IsPromEnabled Is prometheus entry enabled?
func (entry *EchoEntry) IsPromEnabled() bool {
	return entry.PromEntry != nil
}

// IsStaticFileHandlerEnabled Is static file handler entry enabled?
func (entry *EchoEntry) IsStaticFileHandlerEnabled() bool {
	return entry.StaticFileEntry != nil
}

// ***************** Helper function *****************

// Add basic fields into event.
func (entry *EchoEntry) logBasicInfo(operation string) (rkquery.Event, *zap.Logger) {
	event := entry.EventLoggerEntry.GetEventHelper().Start(
		operation,
		rkquery.WithEntryName(entry.GetName()),
		rkquery.WithEntryType(entry.GetType()))
	logger := entry.ZapLoggerEntry.GetLogger().With(
		zap.String("eventId", event.GetEventId()),
		zap.String("entryName", entry.EntryName))

	// add general info
	event.AddPayloads(
		zap.Uint64("echoPort", entry.Port))

	// add SwEntry info
	if entry.IsSwEnabled() {
		event.AddPayloads(
			zap.Bool("swEnabled", true),
			zap.String("swPath", entry.SwEntry.Path))
	}

	// add CommonServiceEntry info
	if entry.IsCommonServiceEnabled() {
		event.AddPayloads(
			zap.Bool("commonServiceEnabled", true),
			zap.String("commonServicePathPrefix", "/rk/v1/"))
	}

	// add TvEntry info
	if entry.IsTvEnabled() {
		event.AddPayloads(
			zap.Bool("tvEnabled", true),
			zap.String("tvPath", "/rk/v1/tv/"))
	}

	// add PromEntry info
	if entry.IsPromEnabled() {
		event.AddPayloads(
			zap.Bool("promEnabled", true),
			zap.Uint64("promPort", entry.PromEntry.Port),
			zap.String("promPath", entry.PromEntry.Path))
	}

	// add StaticFileHandlerEntry info
	if entry.IsStaticFileHandlerEnabled() {
		event.AddPayloads(
			zap.Bool("staticFileHandlerEnabled", true),
			zap.String("staticFileHandlerPath", entry.StaticFileEntry.Path))
	}

	// add tls info
	if entry.IsTlsEnabled() {
		event.AddPayloads(
			zap.Bool("tlsEnabled", true))
	}

	logger.Info(fmt.Sprintf("%s echoEntry", operation))

	return event, logger
}

// Start server
// We move the code here for testability
func (entry *EchoEntry) startServer(event rkquery.Event, logger *zap.Logger) {
	if entry.Echo != nil {
		// If TLS was enabled, we need to load server certificate and key and start http server with ListenAndServeTLS()
		if entry.IsTlsEnabled() {
			err := entry.Echo.StartTLS(
				":"+strconv.FormatUint(entry.Port, 10),
				entry.CertEntry.Store.ServerCert,
				entry.CertEntry.Store.ServerKey)

			if err != nil && err != http.ErrServerClosed {
				event.AddErr(err)
				logger.Error("Error occurs while starting echo server with tls.", event.ListPayloads()...)
				rkcommon.ShutdownWithError(err)
			}
		} else {
			err := entry.Echo.Start(":" + strconv.FormatUint(entry.Port, 10))

			if err != nil && err != http.ErrServerClosed {
				event.AddErr(err)
				logger.Error("Error occurs while starting echo server.", event.ListPayloads()...)
				rkcommon.ShutdownWithError(err)
			}
		}
	}
}

// ***************** Common Service Extension API *****************

// Apis list apis
func (entry *EchoEntry) Apis(ctx echo.Context) error {
	ctx.Response().Header().Set("Access-Control-Allow-Origin", "*")

	return ctx.JSON(http.StatusOK, entry.doApis(ctx))
}

// Req handler
func (entry *EchoEntry) Req(ctx echo.Context) error {
	return ctx.JSON(http.StatusOK, entry.doReq(ctx))
}

// TV handler
func (entry *EchoEntry) TV(ctx echo.Context) error {
	logger := rkechoctx.GetLogger(ctx)

	switch item := ctx.Param("*"); item {
	case "apis":
		buf := entry.TvEntry.ExecuteTemplate("apis", entry.doApis(ctx), logger)
		ctx.HTMLBlob(http.StatusOK, buf.Bytes())
	default:
		buf := entry.TvEntry.Action(item, logger)
		ctx.HTMLBlob(http.StatusOK, buf.Bytes())
	}

	return nil
}

// Helper function for APIs call
func (entry *EchoEntry) doApis(ctx echo.Context) *rkentry.ApisResponse {
	res := &rkentry.ApisResponse{
		Entries: make([]*rkentry.ApisResponseElement, 0),
	}

	routes := entry.Echo.Routes()
	for j := range routes {
		info := routes[j]

		entry := &rkentry.ApisResponseElement{
			EntryName: entry.GetName(),
			Method:    info.Method,
			Path:      info.Path,
			Port:      entry.Port,
			SwUrl:     entry.constructSwUrl(ctx),
		}
		res.Entries = append(res.Entries, entry)
	}

	return res
}

// Construct swagger URL based on IP and scheme
func (entry *EchoEntry) constructSwUrl(ctx echo.Context) string {
	if entry == nil || entry.SwEntry == nil {
		return "N/A"
	}

	originalURL := fmt.Sprintf("localhost:%d", entry.Port)
	if ctx != nil && ctx.Request() != nil && len(ctx.Request().Host) > 0 {
		originalURL = ctx.Request().Host
	}

	scheme := "http"
	if ctx != nil && ctx.Request() != nil && ctx.Request().TLS != nil {
		scheme = "https"
	}

	return fmt.Sprintf("%s://%s%s", scheme, originalURL, entry.SwEntry.Path)
}

// Helper function for Req call
func (entry *EchoEntry) doReq(ctx echo.Context) *rkentry.ReqResponse {
	metricsSet := rkmidmetrics.GetServerMetricsSet(entry.GetName())
	if metricsSet == nil {
		return &rkentry.ReqResponse{
			Metrics: make([]*rkentry.ReqMetricsRK, 0),
		}
	}

	vector := metricsSet.GetSummary(rkmidmetrics.MetricsNameElapsedNano)
	if vector == nil {
		return &rkentry.ReqResponse{
			Metrics: make([]*rkentry.ReqMetricsRK, 0),
		}
	}

	reqMetrics := rkentry.NewPromMetricsInfo(vector)

	// Fill missed metrics
	apis := make([]string, 0)

	routes := entry.Echo.Routes()
	for j := range routes {
		info := routes[j]
		apis = append(apis, info.Path)
	}

	// Add empty metrics into result
	for i := range apis {
		if !entry.containsMetrics(apis[i], reqMetrics) {
			reqMetrics = append(reqMetrics, &rkentry.ReqMetricsRK{
				RestPath: apis[i],
				ResCode:  make([]*rkentry.ResCodeRK, 0),
			})
		}
	}

	return &rkentry.ReqResponse{
		Metrics: reqMetrics,
	}
}

// Is metrics from prometheus contains particular api?
func (entry *EchoEntry) containsMetrics(api string, metrics []*rkentry.ReqMetricsRK) bool {
	for i := range metrics {
		if metrics[i].RestPath == api {
			return true
		}
	}

	return false
}

// ***************** Options *****************

// EchoEntryOption Echo entry option.
type EchoEntryOption func(*EchoEntry)

// WithName provide name.
func WithName(name string) EchoEntryOption {
	return func(entry *EchoEntry) {
		entry.EntryName = name
	}
}

// WithDescription provide name.
func WithDescription(description string) EchoEntryOption {
	return func(entry *EchoEntry) {
		entry.EntryDescription = description
	}
}

// WithPort provide port.
func WithPort(port uint64) EchoEntryOption {
	return func(entry *EchoEntry) {
		entry.Port = port
	}
}

// WithZapLoggerEntry provide rkentry.ZapLoggerEntry.
func WithZapLoggerEntry(zapLogger *rkentry.ZapLoggerEntry) EchoEntryOption {
	return func(entry *EchoEntry) {
		entry.ZapLoggerEntry = zapLogger
	}
}

// WithEventLoggerEntry provide rkentry.EventLoggerEntry.
func WithEventLoggerEntry(eventLogger *rkentry.EventLoggerEntry) EchoEntryOption {
	return func(entry *EchoEntry) {
		entry.EventLoggerEntry = eventLogger
	}
}

// WithCertEntry provide rkentry.CertEntry.
func WithCertEntry(certEntry *rkentry.CertEntry) EchoEntryOption {
	return func(entry *EchoEntry) {
		entry.CertEntry = certEntry
	}
}

// WithSwEntry provide SwEntry.
func WithSwEntry(sw *rkentry.SwEntry) EchoEntryOption {
	return func(entry *EchoEntry) {
		entry.SwEntry = sw
	}
}

// WithCommonServiceEntry provide CommonServiceEntry.
func WithCommonServiceEntry(commonServiceEntry *rkentry.CommonServiceEntry) EchoEntryOption {
	return func(entry *EchoEntry) {
		entry.CommonServiceEntry = commonServiceEntry
	}
}

// WithPromEntry provide PromEntry.
func WithPromEntry(prom *rkentry.PromEntry) EchoEntryOption {
	return func(entry *EchoEntry) {
		entry.PromEntry = prom
	}
}

// WithStaticFileHandlerEntry provide StaticFileHandlerEntry.
func WithStaticFileHandlerEntry(staticEntry *rkentry.StaticFileHandlerEntry) EchoEntryOption {
	return func(entry *EchoEntry) {
		entry.StaticFileEntry = staticEntry
	}
}

// WithTvEntry provide TvEntry.
func WithTvEntry(tvEntry *rkentry.TvEntry) EchoEntryOption {
	return func(entry *EchoEntry) {
		entry.TvEntry = tvEntry
	}
}

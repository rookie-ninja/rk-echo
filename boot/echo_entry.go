// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

// Package rkecho an implementation of rkentry.Entry which could be used start restful server with echo framework
package rkecho

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rookie-ninja/rk-echo/middleware/auth"
	"github.com/rookie-ninja/rk-echo/middleware/cors"
	"github.com/rookie-ninja/rk-echo/middleware/csrf"
	"github.com/rookie-ninja/rk-echo/middleware/gzip"
	"github.com/rookie-ninja/rk-echo/middleware/jwt"
	"github.com/rookie-ninja/rk-echo/middleware/log"
	"github.com/rookie-ninja/rk-echo/middleware/meta"
	"github.com/rookie-ninja/rk-echo/middleware/panic"
	rkechoprom "github.com/rookie-ninja/rk-echo/middleware/prom"
	"github.com/rookie-ninja/rk-echo/middleware/ratelimit"
	"github.com/rookie-ninja/rk-echo/middleware/secure"
	"github.com/rookie-ninja/rk-echo/middleware/timeout"
	"github.com/rookie-ninja/rk-echo/middleware/tracing"
	"github.com/rookie-ninja/rk-entry/v2/entry"
	"github.com/rookie-ninja/rk-entry/v2/middleware"
	"github.com/rookie-ninja/rk-entry/v2/middleware/auth"
	"github.com/rookie-ninja/rk-entry/v2/middleware/cors"
	"github.com/rookie-ninja/rk-entry/v2/middleware/csrf"
	"github.com/rookie-ninja/rk-entry/v2/middleware/jwt"
	"github.com/rookie-ninja/rk-entry/v2/middleware/log"
	"github.com/rookie-ninja/rk-entry/v2/middleware/meta"
	"github.com/rookie-ninja/rk-entry/v2/middleware/panic"
	"github.com/rookie-ninja/rk-entry/v2/middleware/prom"
	"github.com/rookie-ninja/rk-entry/v2/middleware/ratelimit"
	"github.com/rookie-ninja/rk-entry/v2/middleware/secure"
	"github.com/rookie-ninja/rk-entry/v2/middleware/timeout"
	"github.com/rookie-ninja/rk-entry/v2/middleware/tracing"
	"github.com/rookie-ninja/rk-query"
	"go.uber.org/zap"
	"net/http"
	"path"
	"strconv"
	"strings"
	"sync"
)

const (
	// EchoEntryType type of entry
	EchoEntryType = "EchoEntry"
)

// This must be declared in order to register registration function into rk context
// otherwise, rk-boot won't able to bootstrap echo entry automatically from boot config file
func init() {
	rkentry.RegisterEntryRegFunc(RegisterEchoEntryYAML)
}

// BootEcho boot config which is for echo entry.
type BootEcho struct {
	Echo []struct {
		Enabled       bool                          `yaml:"enabled" json:"enabled"`
		Name          string                        `yaml:"name" json:"name"`
		Port          uint64                        `yaml:"port" json:"port"`
		Description   string                        `yaml:"description" json:"description"`
		SW            rkentry.BootSW                `yaml:"sw" json:"sw"`
		Docs          rkentry.BootDocs              `yaml:"docs" json:"docs"`
		CommonService rkentry.BootCommonService     `yaml:"commonService" json:"commonService"`
		Prom          rkentry.BootProm              `yaml:"prom" json:"prom"`
		CertEntry     string                        `yaml:"certEntry" json:"certEntry"`
		LoggerEntry   string                        `yaml:"loggerEntry" json:"loggerEntry"`
		EventEntry    string                        `yaml:"eventEntry" json:"eventEntry"`
		Static        rkentry.BootStaticFileHandler `yaml:"static" json:"static"`
		Middleware    struct {
			Ignore    []string                `yaml:"ignore" json:"ignore"`
			Logging   rkmidlog.BootConfig     `yaml:"logging" json:"logging"`
			Prom      rkmidprom.BootConfig    `yaml:"prom" json:"prom"`
			Auth      rkmidauth.BootConfig    `yaml:"auth" json:"auth"`
			Cors      rkmidcors.BootConfig    `yaml:"cors" json:"cors"`
			Meta      rkmidmeta.BootConfig    `yaml:"meta" json:"meta"`
			Jwt       rkmidjwt.BootConfig     `yaml:"jwt" json:"jwt"`
			Secure    rkmidsec.BootConfig     `yaml:"secure" json:"secure"`
			RateLimit rkmidlimit.BootConfig   `yaml:"rateLimit" json:"rateLimit"`
			Csrf      rkmidcsrf.BootConfig    `yaml:"csrf" yaml:"csrf"`
			Timeout   rkmidtimeout.BootConfig `yaml:"timeout" json:"timeout"`
			Trace     rkmidtrace.BootConfig   `yaml:"trace" json:"trace"`
			Gzip      struct {
				Enabled bool     `yaml:"enabled" json:"enabled"`
				Ignore  []string `yaml:"ignore" json:"ignore"`
				Level   string   `yaml:"level" json:"level"`
			} `yaml:"gzip" json:"gzip"`
		} `yaml:"middleware" json:"middleware"`
	} `yaml:"echo" json:"echo"`
}

// EchoEntry implements rkentry.Entry interface.
type EchoEntry struct {
	entryName          string                          `json:"entryName" yaml:"entryName"`
	entryType          string                          `json:"entryType" yaml:"entryType"`
	entryDescription   string                          `json:"-" yaml:"-"`
	Echo               *echo.Echo                      `json:"-" yaml:"-"`
	Port               uint64                          `json:"-" yaml:"-"`
	LoggerEntry        *rkentry.LoggerEntry            `json:"-" yaml:"-"`
	EventEntry         *rkentry.EventEntry             `json:"-" yaml:"-"`
	SwEntry            *rkentry.SWEntry                `json:"-" yaml:"-"`
	DocsEntry          *rkentry.DocsEntry              `json:"-" yaml:"-"`
	CommonServiceEntry *rkentry.CommonServiceEntry     `json:"-" yaml:"-"`
	PromEntry          *rkentry.PromEntry              `json:"-" yaml:"-"`
	StaticFileEntry    *rkentry.StaticFileHandlerEntry `json:"-" yaml:"-"`
	CertEntry          *rkentry.CertEntry              `json:"-" yaml:"-"`
	bootstrapLogOnce   sync.Once                       `json:"-" yaml:"-"`
}

// RegisterEchoEntryYAML register echo entries with provided config file (Must YAML file).
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
func RegisterEchoEntryYAML(raw []byte) map[string]rkentry.Entry {
	res := make(map[string]rkentry.Entry)

	// 1: Decode config map into boot config struct
	config := &BootEcho{}
	rkentry.UnmarshalBootYAML(raw, config)

	// 2: Init echo entries with boot config
	for i := range config.Echo {
		element := config.Echo[i]
		if !element.Enabled {
			continue
		}

		name := element.Name

		// logger entry
		loggerEntry := rkentry.GlobalAppCtx.GetLoggerEntry(element.LoggerEntry)
		if loggerEntry == nil {
			loggerEntry = rkentry.LoggerEntryStdout
		}

		// event entry
		eventEntry := rkentry.GlobalAppCtx.GetEventEntry(element.EventEntry)
		if eventEntry == nil {
			eventEntry = rkentry.EventEntryStdout
		}

		// cert entry
		certEntry := rkentry.GlobalAppCtx.GetCertEntry(element.CertEntry)

		// Register swagger entry
		swEntry := rkentry.RegisterSWEntry(&element.SW, rkentry.WithNameSWEntry(element.Name))

		// Register docs entry
		docsEntry := rkentry.RegisterDocsEntry(&element.Docs, rkentry.WithNameDocsEntry(element.Name))

		// Register prometheus entry
		promRegistry := prometheus.NewRegistry()
		promEntry := rkentry.RegisterPromEntry(&element.Prom, rkentry.WithRegistryPromEntry(promRegistry))

		// Register common service entry
		commonServiceEntry := rkentry.RegisterCommonServiceEntry(&element.CommonService)

		// Register static file handler
		staticEntry := rkentry.RegisterStaticFileHandlerEntry(&element.Static, rkentry.WithNameStaticFileHandlerEntry(element.Name))

		inters := make([]echo.MiddlewareFunc, 0)

		// add global path ignorance
		rkmid.AddPathToIgnoreGlobal(element.Middleware.Ignore...)

		// logging middlewares
		if element.Middleware.Logging.Enabled {
			inters = append(inters, rkecholog.Middleware(
				rkmidlog.ToOptions(&element.Middleware.Logging, element.Name, EchoEntryType,
					loggerEntry, eventEntry)...))
		}

		// insert panic interceptor
		inters = append(inters, rkechopanic.Interceptor(
			rkmidpanic.WithEntryNameAndType(element.Name, EchoEntryType)))

		// prom middleware
		if element.Middleware.Prom.Enabled {
			inters = append(inters, rkechoprom.Middleware(
				rkmidprom.ToOptions(&element.Middleware.Prom, element.Name, EchoEntryType,
					promRegistry, rkmidprom.LabelerTypeHttp)...))
		}

		// tracing middleware
		if element.Middleware.Trace.Enabled {
			inters = append(inters, rkechotrace.Middleware(
				rkmidtrace.ToOptions(&element.Middleware.Trace, element.Name, EchoEntryType)...))
		}

		// jwt middleware
		if element.Middleware.Jwt.Enabled {
			inters = append(inters, rkechojwt.Middleware(
				rkmidjwt.ToOptions(&element.Middleware.Jwt, element.Name, EchoEntryType)...))
		}

		// secure middleware
		if element.Middleware.Secure.Enabled {
			inters = append(inters, rkechosec.Middleware(
				rkmidsec.ToOptions(&element.Middleware.Secure, element.Name, EchoEntryType)...))
		}

		// csrf middleware
		if element.Middleware.Csrf.Enabled {
			inters = append(inters, rkechocsrf.Middleware(
				rkmidcsrf.ToOptions(&element.Middleware.Csrf, element.Name, EchoEntryType)...))
		}

		// cors middleware
		if element.Middleware.Cors.Enabled {
			inters = append(inters, rkechocors.Middleware(
				rkmidcors.ToOptions(&element.Middleware.Cors, element.Name, EchoEntryType)...))
		}

		// gzip middleware
		if element.Middleware.Gzip.Enabled {
			opts := []rkechogzip.Option{
				rkechogzip.WithEntryNameAndType(element.Name, EchoEntryType),
				rkechogzip.WithLevel(element.Middleware.Gzip.Level),
			}

			inters = append(inters, rkechogzip.Middleware(opts...))
		}

		// meta middleware
		if element.Middleware.Meta.Enabled {
			inters = append(inters, rkechometa.Middleware(
				rkmidmeta.ToOptions(&element.Middleware.Meta, element.Name, EchoEntryType)...))
		}

		// auth middlewares
		if element.Middleware.Auth.Enabled {
			inters = append(inters, rkechoauth.Middleware(
				rkmidauth.ToOptions(&element.Middleware.Auth, element.Name, EchoEntryType)...))
		}

		// timeout middlewares
		if element.Middleware.Timeout.Enabled {
			inters = append(inters, rkechotimeout.Middleware(
				rkmidtimeout.ToOptions(&element.Middleware.Timeout, element.Name, EchoEntryType)...))
		}

		// rate limit middleware
		if element.Middleware.RateLimit.Enabled {
			inters = append(inters, rkecholimit.Middleware(
				rkmidlimit.ToOptions(&element.Middleware.RateLimit, element.Name, EchoEntryType)...))
		}

		entry := RegisterEchoEntry(
			WithName(name),
			WithDescription(element.Description),
			WithPort(element.Port),
			WithLoggerEntry(loggerEntry),
			WithEventEntry(eventEntry),
			WithSwEntry(swEntry),
			WithDocsEntry(docsEntry),
			WithPromEntry(promEntry),
			WithCommonServiceEntry(commonServiceEntry),
			WithCertEntry(certEntry),
			WithStaticFileHandlerEntry(staticEntry))

		entry.AddMiddleware(inters...)

		res[name] = entry
	}

	return res
}

// RegisterEchoEntry register EchoEntry with options.
func RegisterEchoEntry(opts ...EchoEntryOption) *EchoEntry {
	entry := &EchoEntry{
		entryType:        EchoEntryType,
		entryDescription: "Internal RK entry which helps to bootstrap with Echo framework.",
		LoggerEntry:      rkentry.NewLoggerEntryStdout(),
		EventEntry:       rkentry.NewEventEntryStdout(),
		Port:             8080,
	}

	for i := range opts {
		opts[i](entry)
	}

	if len(entry.entryName) < 1 {
		entry.entryName = "echo-" + strconv.FormatUint(entry.Port, 10)
	}

	if entry.Echo == nil {
		entry.Echo = echo.New()
		entry.Echo.HidePort = true
		entry.Echo.HideBanner = true
	}

	// add entry name and entry type into loki syncer if enabled
	entry.LoggerEntry.AddEntryLabelToLokiSyncer(entry)
	entry.EventEntry.AddEntryLabelToLokiSyncer(entry)

	rkentry.GlobalAppCtx.AddEntry(entry)

	return entry
}

// GetName Get entry name.
func (entry *EchoEntry) GetName() string {
	return entry.entryName
}

// GetType Get entry type.
func (entry *EchoEntry) GetType() string {
	return entry.entryType
}

// GetDescription Get description of entry.
func (entry *EchoEntry) GetDescription() string {
	return entry.entryDescription
}

// Bootstrap EchoEntry.
func (entry *EchoEntry) Bootstrap(ctx context.Context) {
	event, logger := entry.logBasicInfo("Bootstrap", ctx)

	// Is common service enabled?
	if entry.IsCommonServiceEnabled() {
		// Register common service path into Router.
		entry.Echo.GET(entry.CommonServiceEntry.ReadyPath, echo.WrapHandler(http.HandlerFunc(entry.CommonServiceEntry.Ready)))
		entry.Echo.GET(entry.CommonServiceEntry.AlivePath, echo.WrapHandler(http.HandlerFunc(entry.CommonServiceEntry.Alive)))
		entry.Echo.GET(entry.CommonServiceEntry.GcPath, echo.WrapHandler(http.HandlerFunc(entry.CommonServiceEntry.Gc)))
		entry.Echo.GET(entry.CommonServiceEntry.InfoPath, echo.WrapHandler(http.HandlerFunc(entry.CommonServiceEntry.Info)))

		// Bootstrap common service entry.
		entry.CommonServiceEntry.Bootstrap(ctx)
	}

	// Is swagger enabled?
	if entry.IsSwEnabled() {
		// Register swagger path into Router.
		entry.Echo.GET(strings.TrimSuffix(entry.SwEntry.Path, "/"), func(ctx echo.Context) error {
			ctx.Redirect(http.StatusTemporaryRedirect, entry.SwEntry.Path)
			return nil
		})
		entry.Echo.GET(path.Join(entry.SwEntry.Path, "*"), echo.WrapHandler(entry.SwEntry.ConfigFileHandler()))
		entry.SwEntry.Bootstrap(ctx)
	}

	// Is Docs enabled?
	if entry.IsDocsEnabled() {
		// Bootstrap Docs entry.
		entry.Echo.GET(strings.TrimSuffix(entry.DocsEntry.Path, "/"), func(ctx echo.Context) error {
			ctx.Redirect(http.StatusTemporaryRedirect, entry.DocsEntry.Path)
			return nil
		})
		entry.Echo.GET(path.Join(entry.DocsEntry.Path, "*"), echo.WrapHandler(entry.DocsEntry.ConfigFileHandler()))

		entry.DocsEntry.Bootstrap(ctx)
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

	// Start echo server
	go entry.startServer(event, logger)

	entry.bootstrapLogOnce.Do(func() {
		// Print link and logging message
		scheme := "http"
		if entry.IsTlsEnabled() {
			scheme = "https"
		}

		if entry.IsSwEnabled() {
			entry.LoggerEntry.Info(fmt.Sprintf("SwaggerEntry: %s://localhost:%d%s", scheme, entry.Port, entry.SwEntry.Path))
		}
		if entry.IsDocsEnabled() {
			entry.LoggerEntry.Info(fmt.Sprintf("DocsEntry: %s://localhost:%d%s", scheme, entry.Port, entry.DocsEntry.Path))
		}
		if entry.IsPromEnabled() {
			entry.LoggerEntry.Info(fmt.Sprintf("PromEntry: %s://localhost:%d%s", scheme, entry.Port, entry.PromEntry.Path))
		}
		if entry.IsStaticFileHandlerEnabled() {
			entry.LoggerEntry.Info(fmt.Sprintf("StaticFileHandlerEntry: %s://localhost:%d%s", scheme, entry.Port, entry.StaticFileEntry.Path))
		}
		if entry.IsCommonServiceEnabled() {
			handlers := []string{
				fmt.Sprintf("%s://localhost:%d%s", scheme, entry.Port, entry.CommonServiceEntry.ReadyPath),
				fmt.Sprintf("%s://localhost:%d%s", scheme, entry.Port, entry.CommonServiceEntry.AlivePath),
				fmt.Sprintf("%s://localhost:%d%s", scheme, entry.Port, entry.CommonServiceEntry.InfoPath),
			}

			entry.LoggerEntry.Info(fmt.Sprintf("CommonSreviceEntry: %s", strings.Join(handlers, ", ")))
		}
		entry.EventEntry.Finish(event)
	})
}

// Interrupt EchoEntry.
func (entry *EchoEntry) Interrupt(ctx context.Context) {
	event, logger := entry.logBasicInfo("Interrupt", ctx)

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

	if entry.IsDocsEnabled() {
		entry.DocsEntry.Interrupt(ctx)
	}

	if entry.Echo != nil {
		if err := entry.Echo.Shutdown(context.Background()); err != nil && err != http.ErrServerClosed {
			event.AddErr(err)
			logger.Warn("Error occurs while stopping echo-server.", event.ListPayloads()...)
		}
	}

	entry.EventEntry.Finish(event)

	rkentry.GlobalAppCtx.RemoveEntry(entry)
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
		"name":                   entry.entryName,
		"type":                   entry.entryType,
		"description":            entry.entryDescription,
		"port":                   entry.Port,
		"swEntry":                entry.SwEntry,
		"docsEntry":              entry.DocsEntry,
		"commonServiceEntry":     entry.CommonServiceEntry,
		"promEntry":              entry.PromEntry,
		"staticFileHandlerEntry": entry.StaticFileEntry,
	}

	if entry.IsTlsEnabled() {
		m["certEntry"] = entry.CertEntry
	}

	return json.Marshal(&m)
}

// UnmarshalJSON Not supported.
func (entry *EchoEntry) UnmarshalJSON([]byte) error {
	return nil
}

// ***************** Public functions *****************

// GetEchoEntry Get EchoEntry from rkentry.GlobalAppCtx.
func GetEchoEntry(name string) *EchoEntry {
	entryRaw := rkentry.GlobalAppCtx.GetEntry(EchoEntryType, name)
	if entryRaw == nil {
		return nil
	}

	entry, _ := entryRaw.(*EchoEntry)
	return entry
}

// AddMiddleware Add interceptors.
// This function should be called before Bootstrap() called.
func (entry *EchoEntry) AddMiddleware(inters ...echo.MiddlewareFunc) {
	entry.Echo.Use(inters...)
}

// IsTlsEnabled Is TLS enabled?
func (entry *EchoEntry) IsTlsEnabled() bool {
	return entry.CertEntry != nil && entry.CertEntry.Certificate != nil
}

// IsSwEnabled Is swagger entry enabled?
func (entry *EchoEntry) IsSwEnabled() bool {
	return entry.SwEntry != nil
}

// IsCommonServiceEnabled Is common service entry enabled?
func (entry *EchoEntry) IsCommonServiceEnabled() bool {
	return entry.CommonServiceEntry != nil
}

// IsDocsEnabled Is docs entry enabled?
func (entry *EchoEntry) IsDocsEnabled() bool {
	return entry.DocsEntry != nil
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
func (entry *EchoEntry) logBasicInfo(operation string, ctx context.Context) (rkquery.Event, *zap.Logger) {
	event := entry.EventEntry.Start(
		operation,
		rkquery.WithEntryName(entry.GetName()),
		rkquery.WithEntryType(entry.GetType()))

	// extract eventId if exists
	if val := ctx.Value("eventId"); val != nil {
		if id, ok := val.(string); ok {
			event.SetEventId(id)
		}
	}

	logger := entry.LoggerEntry.With(
		zap.String("eventId", event.GetEventId()),
		zap.String("entryName", entry.entryName),
		zap.String("entryType", entry.entryType))

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

	// add DocsEntry info
	if entry.IsDocsEnabled() {
		event.AddPayloads(
			zap.Bool("docsEnabled", true),
			zap.String("docsPath", entry.DocsEntry.Path))
	}

	// add PromEntry info
	if entry.IsPromEnabled() {
		event.AddPayloads(
			zap.Bool("promEnabled", true),
			zap.Uint64("promPort", entry.Port),
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

	logger.Info(fmt.Sprintf("%s EchoEntry", operation))

	return event, logger
}

// Start server
// We move the code here for testability
func (entry *EchoEntry) startServer(event rkquery.Event, logger *zap.Logger) {
	if entry.Echo != nil {
		// If TLS was enabled, we need to load server certificate and key and start http server with ListenAndServeTLS()
		if entry.IsTlsEnabled() {
			entry.Echo.TLSServer = &http.Server{
				Addr:      "0.0.0.0:" + strconv.FormatUint(entry.Port, 10),
				Handler:   entry.Echo,
				TLSConfig: &tls.Config{Certificates: []tls.Certificate{*entry.CertEntry.Certificate}},
			}

			err := entry.Echo.TLSServer.ListenAndServe()

			if err != nil && err != http.ErrServerClosed {
				logger.Error("Error occurs while starting echo server with tls.", event.ListPayloads()...)
				entry.bootstrapLogOnce.Do(func() {
					entry.EventEntry.FinishWithCond(event, false)
				})
				rkentry.ShutdownWithError(err)
			}
		} else {
			err := entry.Echo.Start(":" + strconv.FormatUint(entry.Port, 10))

			if err != nil && err != http.ErrServerClosed {
				logger.Error("Error occurs while starting echo server.", event.ListPayloads()...)
				entry.bootstrapLogOnce.Do(func() {
					entry.EventEntry.FinishWithCond(event, false)
				})
				rkentry.ShutdownWithError(err)
			}
		}
	}
}

// ***************** Options *****************

// EchoEntryOption Echo entry option.
type EchoEntryOption func(*EchoEntry)

// WithName provide name.
func WithName(name string) EchoEntryOption {
	return func(entry *EchoEntry) {
		entry.entryName = name
	}
}

// WithDescription provide name.
func WithDescription(description string) EchoEntryOption {
	return func(entry *EchoEntry) {
		entry.entryDescription = description
	}
}

// WithPort provide port.
func WithPort(port uint64) EchoEntryOption {
	return func(entry *EchoEntry) {
		entry.Port = port
	}
}

// WithLoggerEntry provide rkentry.LoggerEntry.
func WithLoggerEntry(zapLogger *rkentry.LoggerEntry) EchoEntryOption {
	return func(entry *EchoEntry) {
		entry.LoggerEntry = zapLogger
	}
}

// WithEventEntry provide rkentry.EventEntry.
func WithEventEntry(eventLogger *rkentry.EventEntry) EchoEntryOption {
	return func(entry *EchoEntry) {
		entry.EventEntry = eventLogger
	}
}

// WithCertEntry provide rkentry.CertEntry.
func WithCertEntry(certEntry *rkentry.CertEntry) EchoEntryOption {
	return func(entry *EchoEntry) {
		entry.CertEntry = certEntry
	}
}

// WithSwEntry provide rkentry.SWEntry.
func WithSwEntry(sw *rkentry.SWEntry) EchoEntryOption {
	return func(entry *EchoEntry) {
		entry.SwEntry = sw
	}
}

// WithCommonServiceEntry provide rkentry.CommonServiceEntry.
func WithCommonServiceEntry(commonServiceEntry *rkentry.CommonServiceEntry) EchoEntryOption {
	return func(entry *EchoEntry) {
		entry.CommonServiceEntry = commonServiceEntry
	}
}

// WithPromEntry provide rkentry.PromEntry.
func WithPromEntry(prom *rkentry.PromEntry) EchoEntryOption {
	return func(entry *EchoEntry) {
		entry.PromEntry = prom
	}
}

// WithStaticFileHandlerEntry provide rkentry.StaticFileHandlerEntry.
func WithStaticFileHandlerEntry(staticEntry *rkentry.StaticFileHandlerEntry) EchoEntryOption {
	return func(entry *EchoEntry) {
		entry.StaticFileEntry = staticEntry
	}
}

// WithDocsEntry provide rkentry.DocsEntry.
func WithDocsEntry(docs *rkentry.DocsEntry) EchoEntryOption {
	return func(entry *EchoEntry) {
		entry.DocsEntry = docs
	}
}

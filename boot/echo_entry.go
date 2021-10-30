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
	"github.com/rookie-ninja/rk-common/common"
	rkechopanic "github.com/rookie-ninja/rk-echo/interceptor/panic"
	"github.com/rookie-ninja/rk-entry/entry"
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

var bootstrapEventIdKey = eventIdKey{}

type eventIdKey struct{}

// This must be declared in order to register registration function into rk context
// otherwise, rk-boot won't able to bootstrap echo entry automatically from boot config file
func init() {
	rkentry.RegisterEntryRegFunc(RegisterEchoEntriesWithConfig)
}

// BootConfigEcho boot config which is for echo entry.
//
// 1: Echo.Enabled: Enable echo entry, default is true.
// 2: Echo.Name: Name of echo entry, should be unique globally.
// 3: Echo.Port: Port of echo entry.
type BootConfigEcho struct {
	Echo []struct {
		Enabled     bool   `yaml:"enabled" json:"enabled"`
		Name        string `yaml:"name" json:"name"`
		Port        uint64 `yaml:"port" json:"port"`
		Description string `yaml:"description" json:"description"`
		Cert        struct {
			Ref string `yaml:"ref" json:"ref"`
		} `yaml:"cert" json:"cert"`
		SW            BootConfigSw            `yaml:"sw" json:"sw"`
		CommonService BootConfigCommonService `yaml:"commonService" json:"commonService"`
		Logger        struct {
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
	EntryName          string                    `json:"entryName" yaml:"entryName"`
	EntryType          string                    `json:"entryType" yaml:"entryType"`
	EntryDescription   string                    `json:"entryDescription" yaml:"entryDescription"`
	ZapLoggerEntry     *rkentry.ZapLoggerEntry   `json:"zapLoggerEntry" yaml:"zapLoggerEntry"`
	EventLoggerEntry   *rkentry.EventLoggerEntry `json:"eventLoggerEntry" yaml:"eventLoggerEntry"`
	Port               uint64                    `json:"port" yaml:"port"`
	CertEntry          *rkentry.CertEntry        `json:"certEntry" yaml:"certEntry"`
	SwEntry            *SwEntry                  `json:"swEntry" yaml:"swEntry"`
	CommonServiceEntry *CommonServiceEntry       `json:"commonServiceEntry" yaml:"commonServiceEntry"`
	Echo               *echo.Echo                `json:"-" yaml:"-"`
	Interceptors       []echo.MiddlewareFunc     `json:"-" yaml:"-"`
}

// EchoEntryOption Echo entry option.
type EchoEntryOption func(*EchoEntry)

// WithNameEcho provide name.
func WithNameEcho(name string) EchoEntryOption {
	return func(entry *EchoEntry) {
		entry.EntryName = name
	}
}

// WithDescriptionEcho provide name.
func WithDescriptionEcho(description string) EchoEntryOption {
	return func(entry *EchoEntry) {
		entry.EntryDescription = description
	}
}

// WithPortEcho provide port.
func WithPortEcho(port uint64) EchoEntryOption {
	return func(entry *EchoEntry) {
		entry.Port = port
	}
}

// WithZapLoggerEntryEcho provide rkentry.ZapLoggerEntry.
func WithZapLoggerEntryEcho(zapLogger *rkentry.ZapLoggerEntry) EchoEntryOption {
	return func(entry *EchoEntry) {
		entry.ZapLoggerEntry = zapLogger
	}
}

// WithEventLoggerEntryEcho provide rkentry.EventLoggerEntry.
func WithEventLoggerEntryEcho(eventLogger *rkentry.EventLoggerEntry) EchoEntryOption {
	return func(entry *EchoEntry) {
		entry.EventLoggerEntry = eventLogger
	}
}

// WithCertEntryEcho provide rkentry.CertEntry.
func WithCertEntryEcho(certEntry *rkentry.CertEntry) EchoEntryOption {
	return func(entry *EchoEntry) {
		entry.CertEntry = certEntry
	}
}

// WithSwEntryEcho provide SwEntry.
func WithSwEntryEcho(sw *SwEntry) EchoEntryOption {
	return func(entry *EchoEntry) {
		entry.SwEntry = sw
	}
}

// WithCommonServiceEntryEcho provide CommonServiceEntry.
func WithCommonServiceEntryEcho(commonServiceEntry *CommonServiceEntry) EchoEntryOption {
	return func(entry *EchoEntry) {
		entry.CommonServiceEntry = commonServiceEntry
	}
}

// WithInterceptorsEcho provide user interceptors.
func WithInterceptorsEcho(inters ...echo.MiddlewareFunc) EchoEntryOption {
	return func(entry *EchoEntry) {
		if entry.Interceptors == nil {
			entry.Interceptors = make([]echo.MiddlewareFunc, 0)
		}

		entry.Interceptors = append(entry.Interceptors, inters...)
	}
}

// GetEchoEntry Get EchoEntry from rkentry.GlobalAppCtx.
func GetEchoEntry(name string) *EchoEntry {
	entryRaw := rkentry.GlobalAppCtx.GetEntry(name)
	if entryRaw == nil {
		return nil
	}

	entry, _ := entryRaw.(*EchoEntry)
	return entry
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
	config := &BootConfigEcho{}
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

		// Did we enabled swagger?
		var swEntry *SwEntry
		if element.SW.Enabled {
			// Init swagger custom headers from config
			headers := make(map[string]string, 0)
			for i := range element.SW.Headers {
				header := element.SW.Headers[i]
				tokens := strings.Split(header, ":")
				if len(tokens) == 2 {
					headers[tokens[0]] = tokens[1]
				}
			}

			swEntry = NewSwEntry(
				WithNameSw(fmt.Sprintf("%s-sw", element.Name)),
				WithZapLoggerEntrySw(zapLoggerEntry),
				WithEventLoggerEntrySw(eventLoggerEntry),
				WithEnableCommonServiceSw(element.CommonService.Enabled),
				WithPortSw(element.Port),
				WithPathSw(element.SW.Path),
				WithJsonPathSw(element.SW.JsonPath),
				WithHeadersSw(headers))
		}

		// Did we enabled common service?
		var commonServiceEntry *CommonServiceEntry
		if element.CommonService.Enabled {
			commonServiceEntry = NewCommonServiceEntry(
				WithNameCommonService(fmt.Sprintf("%s-commonService", element.Name)),
				WithZapLoggerEntryCommonService(zapLoggerEntry),
				WithEventLoggerEntryCommonService(eventLoggerEntry))
		}

		certEntry := rkentry.GlobalAppCtx.GetCertEntry(element.Cert.Ref)

		entry := RegisterEchoEntry(
			WithNameEcho(name),
			WithDescriptionEcho(element.Description),
			WithPortEcho(element.Port),
			WithZapLoggerEntryEcho(zapLoggerEntry),
			WithEventLoggerEntryEcho(eventLoggerEntry),
			WithCertEntryEcho(certEntry),
			WithCommonServiceEntryEcho(commonServiceEntry),
			WithSwEntryEcho(swEntry))

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

	// insert panic interceptor
	entry.Interceptors = append(entry.Interceptors, rkechopanic.Interceptor(
		rkechopanic.WithEntryNameAndType(entry.EntryName, entry.EntryType)))

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

	// Default interceptor should be at front
	entry.Echo.Use(entry.Interceptors...)

	rkentry.GlobalAppCtx.AddEntry(entry)

	return entry
}

// Bootstrap EchoEntry.
func (entry *EchoEntry) Bootstrap(ctx context.Context) {
	event := entry.EventLoggerEntry.GetEventHelper().Start(
		"bootstrap",
		rkquery.WithEntryName(entry.EntryName),
		rkquery.WithEntryType(entry.EntryType))

	entry.logBasicInfo(event)

	ctx = context.WithValue(context.Background(), bootstrapEventIdKey, event.GetEventId())
	logger := entry.ZapLoggerEntry.GetLogger().With(zap.String("eventId", event.GetEventId()))

	// Is swagger enabled?
	if entry.IsSwEnabled() {
		// Register swagger path into Router.
		entry.Echo.GET(path.Join(entry.SwEntry.Path, "*"), entry.SwEntry.ConfigFileHandler())
		entry.Echo.GET("/rk/v1/assets/sw/*", entry.SwEntry.AssetsFileHandler())

		// Bootstrap swagger entry.
		entry.SwEntry.Bootstrap(ctx)
	}

	// Is common service enabled?
	if entry.IsCommonServiceEnabled() {
		// Register common service path into Router.
		entry.Echo.GET("/rk/v1/healthy", entry.CommonServiceEntry.Healthy)
		entry.Echo.GET("/rk/v1/gc", entry.CommonServiceEntry.Gc)
		entry.Echo.GET("/rk/v1/info", entry.CommonServiceEntry.Info)
		entry.Echo.GET("/rk/v1/configs", entry.CommonServiceEntry.Configs)
		entry.Echo.GET("/rk/v1/apis", entry.CommonServiceEntry.Apis)
		entry.Echo.GET("/rk/v1/sys", entry.CommonServiceEntry.Sys)
		entry.Echo.GET("/rk/v1/req", entry.CommonServiceEntry.Req)
		entry.Echo.GET("/rk/v1/entries", entry.CommonServiceEntry.Entries)
		entry.Echo.GET("/rk/v1/certs", entry.CommonServiceEntry.Certs)
		entry.Echo.GET("/rk/v1/logs", entry.CommonServiceEntry.Logs)
		entry.Echo.GET("/rk/v1/deps", entry.CommonServiceEntry.Deps)
		entry.Echo.GET("/rk/v1/license", entry.CommonServiceEntry.License)
		entry.Echo.GET("/rk/v1/readme", entry.CommonServiceEntry.Readme)
		entry.Echo.GET("/rk/v1/git", entry.CommonServiceEntry.Git)

		// Bootstrap common service entry.
		entry.CommonServiceEntry.Bootstrap(ctx)
	}

	logger.Info("Bootstrapping EchoEntry.", event.ListPayloads()...)
	go func(echoEntry *EchoEntry) {
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
	}(entry)

	entry.EventLoggerEntry.GetEventHelper().Finish(event)
}

// Interrupt EchoEntry.
func (entry *EchoEntry) Interrupt(ctx context.Context) {
	event := entry.EventLoggerEntry.GetEventHelper().Start(
		"interrupt",
		rkquery.WithEntryName(entry.EntryName),
		rkquery.WithEntryType(entry.EntryType))

	ctx = context.WithValue(context.Background(), bootstrapEventIdKey, event.GetEventId())
	logger := entry.ZapLoggerEntry.GetLogger().With(zap.String("eventId", event.GetEventId()))

	entry.logBasicInfo(event)

	logger.Info("Interrupting EchoEntry.", event.ListPayloads()...)

	entry.EventLoggerEntry.GetEventHelper().Finish(event)
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

// String Stringfy entry.
func (entry *EchoEntry) String() string {
	bytes, _ := json.Marshal(entry)
	return string(bytes)
}

// MarshalJSON Marshal entry.
func (entry *EchoEntry) MarshalJSON() ([]byte, error) {
	m := map[string]interface{}{
		"entryName":        entry.EntryName,
		"entryType":        entry.EntryType,
		"entryDescription": entry.EntryDescription,
	}

	return json.Marshal(&m)
}

// UnmarshalJSON Not supported.
func (entry *EchoEntry) UnmarshalJSON([]byte) error {
	return nil
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


// Add basic fields into event.
func (entry *EchoEntry) logBasicInfo(event rkquery.Event) {
	event.AddPayloads(
		zap.String("entryName", entry.EntryName),
		zap.String("entryType", entry.EntryType),
		zap.Uint64("port", entry.Port),
	)
}

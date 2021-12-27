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
	"github.com/markbates/pkger"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rookie-ninja/rk-common/common"
	"github.com/rookie-ninja/rk-echo/interceptor/auth"
	"github.com/rookie-ninja/rk-echo/interceptor/cors"
	"github.com/rookie-ninja/rk-echo/interceptor/csrf"
	"github.com/rookie-ninja/rk-echo/interceptor/gzip"
	"github.com/rookie-ninja/rk-echo/interceptor/jwt"
	"github.com/rookie-ninja/rk-echo/interceptor/log/zap"
	"github.com/rookie-ninja/rk-echo/interceptor/meta"
	"github.com/rookie-ninja/rk-echo/interceptor/metrics/prom"
	"github.com/rookie-ninja/rk-echo/interceptor/panic"
	"github.com/rookie-ninja/rk-echo/interceptor/ratelimit"
	"github.com/rookie-ninja/rk-echo/interceptor/secure"
	"github.com/rookie-ninja/rk-echo/interceptor/timeout"
	"github.com/rookie-ninja/rk-echo/interceptor/tracing/telemetry"
	"github.com/rookie-ninja/rk-entry/entry"
	"github.com/rookie-ninja/rk-prom"
	"github.com/rookie-ninja/rk-query"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.uber.org/zap"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"time"
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
// 4: Echo.Cert.Ref: Reference of rkentry.CertEntry.
// 5: Echo.SW: See BootConfigSW for details.
// 6: Echo.CommonService: See BootConfigCommonService for details.
// 7: Echo.TV: See BootConfigTv for details.
// 8: Echo.Prom: See BootConfigProm for details.
// 9: Echo.Interceptors.LoggingZap.Enabled: Enable zap logging interceptor.
// 10: Echo.Interceptors.MetricsProm.Enable: Enable prometheus interceptor.
// 11: Echo.Interceptors.auth.Enabled: Enable basic auth.
// 12: Echo.Interceptors.auth.Basic: Credential for basic auth, scheme: <user:pass>
// 13: Echo.Interceptors.auth.ApiKey: Credential for X-API-Key.
// 14: Echo.Interceptors.auth.igorePrefix: List of paths that will be ignored.
// 15: Echo.Interceptors.Extension.Enabled: Enable extension interceptor.
// 16: Echo.Interceptors.Extension.Prefix: Prefix of extension header key.
// 17: Echo.Interceptors.TracingTelemetry.Enabled: Enable tracing interceptor with opentelemetry.
// 18: Echo.Interceptors.TracingTelemetry.Exporter.File.Enabled: Enable file exporter which support type of stdout and local file.
// 19: Echo.Interceptors.TracingTelemetry.Exporter.File.OutputPath: Output path of file exporter, stdout and file path is supported.
// 20: Echo.Interceptors.TracingTelemetry.Exporter.Jaeger.Enabled: Enable jaeger exporter.
// 21: Echo.Interceptors.TracingTelemetry.Exporter.Jaeger.AgentEndpoint: Specify jeager agent endpoint, localhost:6832 would be used by default.
// 22: Echo.Interceptors.RateLimit.Enabled: Enable rate limit interceptor.
// 23: Echo.Interceptors.RateLimit.Algorithm: Algorithm of rate limiter.
// 24: Echo.Interceptors.RateLimit.ReqPerSec: Request per second.
// 25: Echo.Interceptors.RateLimit.Paths.path: Name of full path.
// 26: Echo.Interceptors.RateLimit.Paths.ReqPerSec: Request per second by path.
// 27: Echo.Interceptors.Timeout.Enabled: Enable timeout interceptor.
// 28: Echo.Interceptors.Timeout.TimeoutMs: Timeout in milliseconds.
// 29: Echo.Interceptors.Timeout.Paths.path: Name of full path.
// 30: Echo.Interceptors.Timeout.Paths.TimeoutMs: Timeout in milliseconds by path.
// 31: Echo.Logger.ZapLogger.Ref: Zap logger reference, see rkentry.ZapLoggerEntry for details.
// 32: Echo.Logger.EventLogger.Ref: Event logger reference, see rkentry.EventLoggerEntry for details.
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
		TV            BootConfigTv            `yaml:"tv" json:"tv"`
		Prom          BootConfigProm          `yaml:"prom" json:"prom"`
		Static        BootConfigStaticHandler `yaml:"static" json:"static"`
		Interceptors  struct {
			LoggingZap struct {
				Enabled                bool     `yaml:"enabled" json:"enabled"`
				ZapLoggerEncoding      string   `yaml:"zapLoggerEncoding" json:"zapLoggerEncoding"`
				ZapLoggerOutputPaths   []string `yaml:"zapLoggerOutputPaths" json:"zapLoggerOutputPaths"`
				EventLoggerEncoding    string   `yaml:"eventLoggerEncoding" json:"eventLoggerEncoding"`
				EventLoggerOutputPaths []string `yaml:"eventLoggerOutputPaths" json:"eventLoggerOutputPaths"`
			} `yaml:"loggingZap" json:"loggingZap"`
			MetricsProm struct {
				Enabled bool `yaml:"enabled" json:"enabled"`
			} `yaml:"metricsProm" json:"metricsProm"`
			Auth struct {
				Enabled      bool     `yaml:"enabled" json:"enabled"`
				IgnorePrefix []string `yaml:"ignorePrefix" json:"ignorePrefix"`
				Basic        []string `yaml:"basic" json:"basic"`
				ApiKey       []string `yaml:"apiKey" json:"apiKey"`
			} `yaml:"auth" json:"auth"`
			Cors struct {
				Enabled          bool     `yaml:"enabled" json:"enabled"`
				AllowOrigins     []string `yaml:"allowOrigins" json:"allowOrigins"`
				AllowCredentials bool     `yaml:"allowCredentials" json:"allowCredentials"`
				AllowHeaders     []string `yaml:"allowHeaders" json:"allowHeaders"`
				AllowMethods     []string `yaml:"allowMethods" json:"allowMethods"`
				ExposeHeaders    []string `yaml:"exposeHeaders" json:"exposeHeaders"`
				MaxAge           int      `yaml:"maxAge" json:"maxAge"`
			} `yaml:"cors" json:"cors"`
			Meta struct {
				Enabled bool   `yaml:"enabled" json:"enabled"`
				Prefix  string `yaml:"prefix" json:"prefix"`
			} `yaml:"meta" json:"meta"`
			Jwt struct {
				Enabled      bool     `yaml:"enabled" json:"enabled"`
				IgnorePrefix []string `yaml:"ignorePrefix" json:"ignorePrefix"`
				SigningKey   string   `yaml:"signingKey" json:"signingKey"`
				SigningKeys  []string `yaml:"signingKeys" json:"signingKeys"`
				SigningAlgo  string   `yaml:"signingAlgo" json:"signingAlgo"`
				TokenLookup  string   `yaml:"tokenLookup" json:"tokenLookup"`
				AuthScheme   string   `yaml:"authScheme" json:"authScheme"`
			} `yaml:"jwt" json:"jwt"`
			Secure struct {
				Enabled               bool     `yaml:"enabled" json:"enabled"`
				IgnorePrefix          []string `yaml:"ignorePrefix" json:"ignorePrefix"`
				XssProtection         string   `yaml:"xssProtection" json:"xssProtection"`
				ContentTypeNosniff    string   `yaml:"contentTypeNosniff" json:"contentTypeNosniff"`
				XFrameOptions         string   `yaml:"xFrameOptions" json:"xFrameOptions"`
				HstsMaxAge            int      `yaml:"hstsMaxAge" json:"hstsMaxAge"`
				HstsExcludeSubdomains bool     `yaml:"hstsExcludeSubdomains" json:"hstsExcludeSubdomains"`
				HstsPreloadEnabled    bool     `yaml:"hstsPreloadEnabled" json:"hstsPreloadEnabled"`
				ContentSecurityPolicy string   `yaml:"contentSecurityPolicy" json:"contentSecurityPolicy"`
				CspReportOnly         bool     `yaml:"cspReportOnly" json:"cspReportOnly"`
				ReferrerPolicy        string   `yaml:"referrerPolicy" json:"referrerPolicy"`
			} `yaml:"secure" json:"secure"`
			Csrf struct {
				Enabled        bool     `yaml:"enabled" json:"enabled"`
				IgnorePrefix   []string `yaml:"ignorePrefix" json:"ignorePrefix"`
				TokenLength    int      `yaml:"tokenLength" json:"tokenLength"`
				TokenLookup    string   `yaml:"tokenLookup" json:"tokenLookup"`
				CookieName     string   `yaml:"cookieName" json:"cookieName"`
				CookieDomain   string   `yaml:"cookieDomain" json:"cookieDomain"`
				CookiePath     string   `yaml:"cookiePath" json:"cookiePath"`
				CookieMaxAge   int      `yaml:"cookieMaxAge" json:"cookieMaxAge"`
				CookieHttpOnly bool     `yaml:"cookieHttpOnly" json:"cookieHttpOnly"`
				CookieSameSite string   `yaml:"cookieSameSite" json:"cookieSameSite"`
			} `yaml:"csrf" yaml:"csrf"`
			Gzip struct {
				Enabled bool   `yaml:"enabled" json:"enabled"`
				Level   string `yaml:"level" json:"level"`
			} `yaml:"gzip" json:"gzip"`
			RateLimit struct {
				Enabled   bool   `yaml:"enabled" json:"enabled"`
				Algorithm string `yaml:"algorithm" json:"algorithm"`
				ReqPerSec int    `yaml:"reqPerSec" json:"reqPerSec"`
				Paths     []struct {
					Path      string `yaml:"path" json:"path"`
					ReqPerSec int    `yaml:"reqPerSec" json:"reqPerSec"`
				} `yaml:"paths" json:"paths"`
			} `yaml:"rateLimit" json:"rateLimit"`
			Timeout struct {
				Enabled   bool `yaml:"enabled" json:"enabled"`
				TimeoutMs int  `yaml:"timeoutMs" json:"timeoutMs"`
				Paths     []struct {
					Path      string `yaml:"path" json:"path"`
					TimeoutMs int    `yaml:"timeoutMs" json:"timeoutMs"`
				} `yaml:"paths" json:"paths"`
			} `yaml:"timeout" json:"timeout"`
			TracingTelemetry struct {
				Enabled  bool `yaml:"enabled" json:"enabled"`
				Exporter struct {
					File struct {
						Enabled    bool   `yaml:"enabled" json:"enabled"`
						OutputPath string `yaml:"outputPath" json:"outputPath"`
					} `yaml:"file" json:"file"`
					Jaeger struct {
						Agent struct {
							Enabled bool   `yaml:"enabled" json:"enabled"`
							Host    string `yaml:"host" json:"host"`
							Port    int    `yaml:"port" json:"port"`
						} `yaml:"agent" json:"agent"`
						Collector struct {
							Enabled  bool   `yaml:"enabled" json:"enabled"`
							Endpoint string `yaml:"endpoint" json:"endpoint"`
							Username string `yaml:"username" json:"username"`
							Password string `yaml:"password" json:"password"`
						} `yaml:"collector" json:"collector"`
					} `yaml:"jaeger" json:"jaeger"`
				} `yaml:"exporter" json:"exporter"`
			} `yaml:"tracingTelemetry" json:"tracingTelemetry"`
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
	PromEntry          *PromEntry                `json:"promEntry" yaml:"promEntry"`
	StaticFileEntry    *StaticFileHandlerEntry   `json:"staticFileHandlerEntry" yaml:"staticFileHandlerEntry"`
	TvEntry            *TvEntry                  `json:"tvEntry" yaml:"tvEntry"`
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

// WithPromEntryEcho provide PromEntry.
func WithPromEntryEcho(prom *PromEntry) EchoEntryOption {
	return func(entry *EchoEntry) {
		entry.PromEntry = prom
	}
}

// WithStaticFileHandlerEntryEcho provide StaticFileHandlerEntry.
func WithStaticFileHandlerEntryEcho(staticEntry *StaticFileHandlerEntry) EchoEntryOption {
	return func(entry *EchoEntry) {
		entry.StaticFileEntry = staticEntry
	}
}

// WithTVEntryEcho provide TvEntry.
func WithTVEntryEcho(tvEntry *TvEntry) EchoEntryOption {
	return func(entry *EchoEntry) {
		entry.TvEntry = tvEntry
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

		promRegistry := prometheus.NewRegistry()
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

		// Did we enabled prometheus?
		var promEntry *PromEntry
		if element.Prom.Enabled {
			var pusher *rkprom.PushGatewayPusher
			if element.Prom.Pusher.Enabled {
				certEntry := rkentry.GlobalAppCtx.GetCertEntry(element.Prom.Pusher.Cert.Ref)
				var certStore *rkentry.CertStore

				if certEntry != nil {
					certStore = certEntry.Store
				}

				pusher, _ = rkprom.NewPushGatewayPusher(
					rkprom.WithIntervalMSPusher(time.Duration(element.Prom.Pusher.IntervalMs)*time.Millisecond),
					rkprom.WithRemoteAddressPusher(element.Prom.Pusher.RemoteAddress),
					rkprom.WithJobNamePusher(element.Prom.Pusher.JobName),
					rkprom.WithBasicAuthPusher(element.Prom.Pusher.BasicAuth),
					rkprom.WithZapLoggerEntryPusher(zapLoggerEntry),
					rkprom.WithEventLoggerEntryPusher(eventLoggerEntry),
					rkprom.WithCertStorePusher(certStore))
			}

			promRegistry.Register(prometheus.NewGoCollector())
			promEntry = NewPromEntry(
				WithNameProm(fmt.Sprintf("%s-prom", element.Name)),
				WithPortProm(element.Port),
				WithPathProm(element.Prom.Path),
				WithZapLoggerEntryProm(zapLoggerEntry),
				WithPromRegistryProm(promRegistry),
				WithEventLoggerEntryProm(eventLoggerEntry),
				WithPusherProm(pusher))

			if promEntry.Pusher != nil {
				promEntry.Pusher.SetGatherer(promEntry.Gatherer)
			}
		}

		inters := make([]echo.MiddlewareFunc, 0)

		// Did we enabled logging interceptor?
		if element.Interceptors.LoggingZap.Enabled {
			opts := []rkecholog.Option{
				rkecholog.WithEntryNameAndType(element.Name, EchoEntryType),
				rkecholog.WithEventLoggerEntry(eventLoggerEntry),
				rkecholog.WithZapLoggerEntry(zapLoggerEntry),
			}

			if strings.ToLower(element.Interceptors.LoggingZap.ZapLoggerEncoding) == "json" {
				opts = append(opts, rkecholog.WithZapLoggerEncoding(rkecholog.ENCODING_JSON))
			}

			if strings.ToLower(element.Interceptors.LoggingZap.EventLoggerEncoding) == "json" {
				opts = append(opts, rkecholog.WithEventLoggerEncoding(rkecholog.ENCODING_JSON))
			}

			if len(element.Interceptors.LoggingZap.ZapLoggerOutputPaths) > 0 {
				opts = append(opts, rkecholog.WithZapLoggerOutputPaths(element.Interceptors.LoggingZap.ZapLoggerOutputPaths...))
			}

			if len(element.Interceptors.LoggingZap.EventLoggerOutputPaths) > 0 {
				opts = append(opts, rkecholog.WithEventLoggerOutputPaths(element.Interceptors.LoggingZap.EventLoggerOutputPaths...))
			}

			inters = append(inters, rkecholog.Interceptor(opts...))
		}

		// Did we enabled metrics interceptor?
		if element.Interceptors.MetricsProm.Enabled {
			opts := []rkechometrics.Option{
				rkechometrics.WithRegisterer(promRegistry),
				rkechometrics.WithEntryNameAndType(element.Name, EchoEntryType),
			}

			inters = append(inters, rkechometrics.Interceptor(opts...))
		}

		// Did we enabled tracing interceptor?
		if element.Interceptors.TracingTelemetry.Enabled {
			var exporter trace.SpanExporter

			if element.Interceptors.TracingTelemetry.Exporter.File.Enabled {
				exporter = rkechotrace.CreateFileExporter(element.Interceptors.TracingTelemetry.Exporter.File.OutputPath)
			}

			if element.Interceptors.TracingTelemetry.Exporter.Jaeger.Agent.Enabled {
				opts := make([]jaeger.AgentEndpointOption, 0)
				if len(element.Interceptors.TracingTelemetry.Exporter.Jaeger.Agent.Host) > 0 {
					opts = append(opts,
						jaeger.WithAgentHost(element.Interceptors.TracingTelemetry.Exporter.Jaeger.Agent.Host))
				}
				if element.Interceptors.TracingTelemetry.Exporter.Jaeger.Agent.Port > 0 {
					opts = append(opts,
						jaeger.WithAgentPort(
							fmt.Sprintf("%d", element.Interceptors.TracingTelemetry.Exporter.Jaeger.Agent.Port)))
				}

				exporter = rkechotrace.CreateJaegerExporter(jaeger.WithAgentEndpoint(opts...))
			}

			if element.Interceptors.TracingTelemetry.Exporter.Jaeger.Collector.Enabled {
				opts := []jaeger.CollectorEndpointOption{
					jaeger.WithUsername(element.Interceptors.TracingTelemetry.Exporter.Jaeger.Collector.Username),
					jaeger.WithPassword(element.Interceptors.TracingTelemetry.Exporter.Jaeger.Collector.Password),
				}

				if len(element.Interceptors.TracingTelemetry.Exporter.Jaeger.Collector.Endpoint) > 0 {
					opts = append(opts, jaeger.WithEndpoint(element.Interceptors.TracingTelemetry.Exporter.Jaeger.Collector.Endpoint))
				}

				exporter = rkechotrace.CreateJaegerExporter(jaeger.WithCollectorEndpoint(opts...))
			}

			opts := []rkechotrace.Option{
				rkechotrace.WithEntryNameAndType(element.Name, EchoEntryType),
				rkechotrace.WithExporter(exporter),
			}

			inters = append(inters, rkechotrace.Interceptor(opts...))
		}

		// Did we enabled jwt interceptor?
		if element.Interceptors.Jwt.Enabled {
			var signingKey []byte
			if len(element.Interceptors.Jwt.SigningKey) > 0 {
				signingKey = []byte(element.Interceptors.Jwt.SigningKey)
			}

			opts := []rkechojwt.Option{
				rkechojwt.WithEntryNameAndType(element.Name, EchoEntryType),
				rkechojwt.WithSigningKey(signingKey),
				rkechojwt.WithSigningAlgorithm(element.Interceptors.Jwt.SigningAlgo),
				rkechojwt.WithTokenLookup(element.Interceptors.Jwt.TokenLookup),
				rkechojwt.WithAuthScheme(element.Interceptors.Jwt.AuthScheme),
				rkechojwt.WithIgnorePrefix(element.Interceptors.Jwt.IgnorePrefix...),
			}

			for _, v := range element.Interceptors.Jwt.SigningKeys {
				tokens := strings.SplitN(v, ":", 2)
				if len(tokens) == 2 {
					opts = append(opts, rkechojwt.WithSigningKeys(tokens[0], tokens[1]))
				}
			}

			inters = append(inters, rkechojwt.Interceptor(opts...))
		}

		// Did we enabled secure interceptor?
		if element.Interceptors.Secure.Enabled {
			opts := []rkechosec.Option{
				rkechosec.WithEntryNameAndType(element.Name, EchoEntryType),
				rkechosec.WithXSSProtection(element.Interceptors.Secure.XssProtection),
				rkechosec.WithContentTypeNosniff(element.Interceptors.Secure.ContentTypeNosniff),
				rkechosec.WithXFrameOptions(element.Interceptors.Secure.XFrameOptions),
				rkechosec.WithHSTSMaxAge(element.Interceptors.Secure.HstsMaxAge),
				rkechosec.WithHSTSExcludeSubdomains(element.Interceptors.Secure.HstsExcludeSubdomains),
				rkechosec.WithHSTSPreloadEnabled(element.Interceptors.Secure.HstsPreloadEnabled),
				rkechosec.WithContentSecurityPolicy(element.Interceptors.Secure.ContentSecurityPolicy),
				rkechosec.WithCSPReportOnly(element.Interceptors.Secure.CspReportOnly),
				rkechosec.WithReferrerPolicy(element.Interceptors.Secure.ReferrerPolicy),
				rkechosec.WithIgnorePrefix(element.Interceptors.Secure.IgnorePrefix...),
			}

			inters = append(inters, rkechosec.Interceptor(opts...))
		}

		// Did we enabled csrf interceptor?
		if element.Interceptors.Csrf.Enabled {
			opts := []rkechocsrf.Option{
				rkechocsrf.WithEntryNameAndType(element.Name, EchoEntryType),
				rkechocsrf.WithTokenLength(element.Interceptors.Csrf.TokenLength),
				rkechocsrf.WithTokenLookup(element.Interceptors.Csrf.TokenLookup),
				rkechocsrf.WithCookieName(element.Interceptors.Csrf.CookieName),
				rkechocsrf.WithCookieDomain(element.Interceptors.Csrf.CookieDomain),
				rkechocsrf.WithCookiePath(element.Interceptors.Csrf.CookiePath),
				rkechocsrf.WithCookieMaxAge(element.Interceptors.Csrf.CookieMaxAge),
				rkechocsrf.WithCookieHTTPOnly(element.Interceptors.Csrf.CookieHttpOnly),
				rkechocsrf.WithIgnorePrefix(element.Interceptors.Csrf.IgnorePrefix...),
			}

			// convert to string to cookie same sites
			sameSite := http.SameSiteDefaultMode

			switch strings.ToLower(element.Interceptors.Csrf.CookieSameSite) {
			case "lax":
				sameSite = http.SameSiteLaxMode
			case "strict":
				sameSite = http.SameSiteStrictMode
			case "none":
				sameSite = http.SameSiteNoneMode
			default:
				sameSite = http.SameSiteDefaultMode
			}

			opts = append(opts, rkechocsrf.WithCookieSameSite(sameSite))

			inters = append(inters, rkechocsrf.Interceptor(opts...))
		}

		// Did we enabled cors interceptor?
		if element.Interceptors.Cors.Enabled {
			opts := []rkechocors.Option{
				rkechocors.WithEntryNameAndType(element.Name, EchoEntryType),
				rkechocors.WithAllowOrigins(element.Interceptors.Cors.AllowOrigins...),
				rkechocors.WithAllowCredentials(element.Interceptors.Cors.AllowCredentials),
				rkechocors.WithExposeHeaders(element.Interceptors.Cors.ExposeHeaders...),
				rkechocors.WithMaxAge(element.Interceptors.Cors.MaxAge),
				rkechocors.WithAllowHeaders(element.Interceptors.Cors.AllowHeaders...),
				rkechocors.WithAllowMethods(element.Interceptors.Cors.AllowMethods...),
			}

			inters = append(inters, rkechocors.Interceptor(opts...))
		}

		// Did we enabled gzip interceptor?
		if element.Interceptors.Gzip.Enabled {
			opts := []rkechogzip.Option{
				rkechogzip.WithEntryNameAndType(element.Name, EchoEntryType),
				rkechogzip.WithLevel(element.Interceptors.Gzip.Level),
			}

			inters = append(inters, rkechogzip.Interceptor(opts...))
		}

		// Did we enabled meta interceptor?
		if element.Interceptors.Meta.Enabled {
			opts := []rkechometa.Option{
				rkechometa.WithEntryNameAndType(element.Name, EchoEntryType),
				rkechometa.WithPrefix(element.Interceptors.Meta.Prefix),
			}

			inters = append(inters, rkechometa.Interceptor(opts...))
		}

		// Did we enabled auth interceptor?
		if element.Interceptors.Auth.Enabled {
			opts := make([]rkechoauth.Option, 0)
			opts = append(opts,
				rkechoauth.WithEntryNameAndType(element.Name, EchoEntryType),
				rkechoauth.WithBasicAuth(element.Name, element.Interceptors.Auth.Basic...),
				rkechoauth.WithApiKeyAuth(element.Interceptors.Auth.ApiKey...))

			// Add exceptional path
			if swEntry != nil {
				opts = append(opts, rkechoauth.WithIgnorePrefix(strings.TrimSuffix(swEntry.Path, "/")))
			}

			opts = append(opts, rkechoauth.WithIgnorePrefix("/rk/v1/assets"))
			opts = append(opts, rkechoauth.WithIgnorePrefix(element.Interceptors.Auth.IgnorePrefix...))

			inters = append(inters, rkechoauth.Interceptor(opts...))
		}

		// Did we enabled timeout interceptor?
		// This should be in front of rate limit interceptor since rate limit may block over the threshold of timeout.
		if element.Interceptors.Timeout.Enabled {
			opts := make([]rkechotimeout.Option, 0)
			opts = append(opts,
				rkechotimeout.WithEntryNameAndType(element.Name, EchoEntryType))

			timeout := time.Duration(element.Interceptors.Timeout.TimeoutMs) * time.Millisecond
			opts = append(opts, rkechotimeout.WithTimeoutAndResp(timeout, nil))

			for i := range element.Interceptors.Timeout.Paths {
				e := element.Interceptors.Timeout.Paths[i]
				timeout := time.Duration(e.TimeoutMs) * time.Millisecond
				opts = append(opts, rkechotimeout.WithTimeoutAndRespByPath(e.Path, timeout, nil))
			}

			inters = append(inters, rkechotimeout.Interceptor(opts...))
		}

		// Did we enabled rate limit interceptor?
		if element.Interceptors.RateLimit.Enabled {
			opts := make([]rkecholimit.Option, 0)
			opts = append(opts,
				rkecholimit.WithEntryNameAndType(element.Name, EchoEntryType))

			if len(element.Interceptors.RateLimit.Algorithm) > 0 {
				opts = append(opts, rkecholimit.WithAlgorithm(element.Interceptors.RateLimit.Algorithm))
			}
			opts = append(opts, rkecholimit.WithReqPerSec(element.Interceptors.RateLimit.ReqPerSec))

			for i := range element.Interceptors.RateLimit.Paths {
				e := element.Interceptors.RateLimit.Paths[i]
				opts = append(opts, rkecholimit.WithReqPerSecByPath(e.Path, e.ReqPerSec))
			}

			inters = append(inters, rkecholimit.Interceptor(opts...))
		}

		// Did we enabled common service?
		var commonServiceEntry *CommonServiceEntry
		if element.CommonService.Enabled {
			commonServiceEntry = NewCommonServiceEntry(
				WithNameCommonService(fmt.Sprintf("%s-commonService", element.Name)),
				WithZapLoggerEntryCommonService(zapLoggerEntry),
				WithEventLoggerEntryCommonService(eventLoggerEntry))
		}

		// Did we enabled tv?
		var tvEntry *TvEntry
		if element.TV.Enabled {
			tvEntry = NewTvEntry(
				WithNameTv(fmt.Sprintf("%s-tv", element.Name)),
				WithZapLoggerEntryTv(zapLoggerEntry),
				WithEventLoggerEntryTv(eventLoggerEntry))
		}

		// DId we enabled static file handler?
		var staticEntry *StaticFileHandlerEntry
		if element.Static.Enabled {
			var fs http.FileSystem
			switch element.Static.SourceType {
			case "pkger":
				fs = pkger.Dir(element.Static.SourcePath)
			case "local":
				if !filepath.IsAbs(element.Static.SourcePath) {
					wd, _ := os.Getwd()
					element.Static.SourcePath = path.Join(wd, element.Static.SourcePath)
				}
				fs = http.Dir(element.Static.SourcePath)
			}

			staticEntry = NewStaticFileHandlerEntry(
				WithZapLoggerEntryStatic(zapLoggerEntry),
				WithEventLoggerEntryStatic(eventLoggerEntry),
				WithNameStatic(fmt.Sprintf("%s-static", element.Name)),
				WithPathStatic(element.Static.Path),
				WithFileSystemStatic(fs))
		}

		certEntry := rkentry.GlobalAppCtx.GetCertEntry(element.Cert.Ref)

		entry := RegisterEchoEntry(
			WithNameEcho(name),
			WithDescriptionEcho(element.Description),
			WithPortEcho(element.Port),
			WithZapLoggerEntryEcho(zapLoggerEntry),
			WithEventLoggerEntryEcho(eventLoggerEntry),
			WithCertEntryEcho(certEntry),
			WithPromEntryEcho(promEntry),
			WithTVEntryEcho(tvEntry),
			WithCommonServiceEntryEcho(commonServiceEntry),
			WithSwEntryEcho(swEntry),
			WithStaticFileHandlerEntryEcho(staticEntry),
			WithInterceptorsEcho(inters...))

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

	rkentry.GlobalAppCtx.AddEntry(entry)

	return entry
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
		entry.Echo.GET(path.Join(entry.SwEntry.Path, "*"), entry.SwEntry.ConfigFileHandler())
		entry.Echo.GET("/rk/v1/assets/sw/*", entry.SwEntry.AssetsFileHandler())

		// Bootstrap swagger entry.
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
		entry.Echo.GET(path.Join(entry.StaticFileEntry.Path, "*"), entry.StaticFileEntry.GetFileHandler())

		// Bootstrap entry.
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

	// Is TV enabled?
	if entry.IsTvEnabled() {
		// Bootstrap TV entry.
		entry.Echo.GET("/rk/v1/tv", func(ctx echo.Context) error {
			ctx.Redirect(http.StatusTemporaryRedirect, "/rk/v1/tv/")
			return nil
		})
		entry.Echo.GET("/rk/v1/tv/*", entry.TvEntry.TV)
		entry.Echo.GET("/rk/v1/assets/tv/*", entry.TvEntry.AssetsFileHandler())

		entry.TvEntry.Bootstrap(ctx)
	}

	// Default interceptor should be at front
	entry.Echo.Use(entry.Interceptors...)

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

	for i := range entry.Interceptors {
		element := entry.Interceptors[i]
		interceptorsStr = append(interceptorsStr,
			path.Base(runtime.FuncForPC(reflect.ValueOf(element).Pointer()).Name()))
	}

	return json.Marshal(&m)
}

// UnmarshalJSON Not supported.
func (entry *EchoEntry) UnmarshalJSON([]byte) error {
	return nil
}

// AddInterceptor Add interceptors.
// This function should be called before Bootstrap() called.
func (entry *EchoEntry) AddInterceptor(inters ...echo.MiddlewareFunc) {
	entry.Interceptors = append(entry.Interceptors, inters...)
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

// Add basic fields into event.
func (entry *EchoEntry) logBasicInfo(operation string) (rkquery.Event, *zap.Logger) {
	event := entry.EventLoggerEntry.GetEventHelper().Start(
		operation,
		rkquery.WithEntryName(entry.GetName()),
		rkquery.WithEntryType(entry.GetType()))
	logger := entry.ZapLoggerEntry.GetLogger().With(
		zap.String("eventId", event.GetEventId()),
		zap.String("entryName", entry.EntryName))

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

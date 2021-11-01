# rk-echo
[![build](https://github.com/rookie-ninja/rk-echo/actions/workflows/ci.yml/badge.svg)](https://github.com/rookie-ninja/rk-echo/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/rookie-ninja/rk-echo)](https://goreportcard.com/report/github.com/rookie-ninja/rk-echo)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

Interceptor & bootstrapper designed for echo framework. Currently, supports bellow functionalities.

| Name | Description |
| ---- | ---- |
| Start with YAML | Start service with YAML config. |
| Start with code | Start service from code. |
| Echo Service | Echo service. |
| Swagger Service | Swagger UI. |
| Common Service | List of common API available on Echo. |
| TV Service | A Web UI shows application and environment information. |
| Metrics interceptor | Collect RPC metrics and export as prometheus client. |
| Log interceptor | Log every RPC requests as event with rk-query. |
| Trace interceptor | Collect RPC trace and export it to stdout, file or jaeger. |
| Panic interceptor | Recover from panic for RPC requests and log it. |
| Meta interceptor | Send application metadata as header to client. |
| Auth interceptor | Support [Basic Auth] and [API Key] authorization types. |
| RateLimit interceptor | Limiting RPC rate |
| Timeout interceptor | Timing out request by configuration. |

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->
**Table of Contents**  *generated with [DocToc](https://github.com/thlorenz/doctoc)*

- [Installation](#installation)
- [Quick Start](#quick-start)
  - [Start Echo Service](#start-echo-service)
  - [Output](#output)
    - [Echo Service](#echo-service)
    - [Swagger Service](#swagger-service)
    - [TV Service](#tv-service)
    - [Metrics](#metrics)
    - [Logging](#logging)
    - [Meta](#meta)
- [YAML Config](#yaml-config)
  - [Echo Service](#echo-service-1)
  - [Common Service](#common-service)
  - [Swagger Service](#swagger-service-1)
  - [Prom Client](#prom-client)
  - [TV Service](#tv-service-1)
  - [Interceptors](#interceptors)
    - [Log](#log)
    - [Metrics](#metrics-1)
    - [Auth](#auth)
    - [Meta](#meta-1)
    - [Tracing](#tracing)
    - [RateLimit](#ratelimit)
    - [Timeout](#timeout)
  - [Development Status: Stable](#development-status-stable)
  - [Contributing](#contributing)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

## Installation
`go get -u github.com/rookie-ninja/rk-echo`

## Quick Start
Bootstrapper can be used with YAML config. In the bellow example, we will start bellow services automatically.
- Echo Service
- Swagger Service
- Common Service
- TV Service
- Metrics
- Logging
- Meta

Please refer example at [example/boot/simple](example/boot/simple).

### Start Echo Service

- [boot.yaml](example/boot/simple/boot.yaml)

```yaml
---
echo:
  - name: greeter                     # Required
    port: 8080                        # Required
    enabled: true                     # Required
    tv:
      enabled: true                   # Optional, default: false
    prom:
      enabled: true                   # Optional, default: false
    sw:                               # Optional
      enabled: true                   # Optional, default: false
    commonService:                    # Optional
      enabled: true                   # Optional, default: false
    interceptors:
      loggingZap:
        enabled: true
      metricsProm:
        enabled: true
      meta:
        enabled: true
```

- [main.go](example/boot/simple/main.go)

```go
func main() {
	// Bootstrap basic entries from boot config.
	rkentry.RegisterInternalEntriesFromConfig("example/boot/simple/boot.yaml")

	// Bootstrap gin entry from boot config
	res := rkecho.RegisterEchoEntriesWithConfig("example/boot/simple/boot.yaml")

	// Bootstrap gin entry
	res["greeter"].Bootstrap(context.Background())

	// Wait for shutdown signal
	rkentry.GlobalAppCtx.WaitForShutdownSig()

	// Interrupt gin entry
	res["greeter"].Interrupt(context.Background())
}
```

```go
$ go run main.go
```

### Output
#### Echo Service
Try to test Echo Service with [curl](https://curl.se/)
```shell script
# Curl to common service
$ curl localhost:8080/rk/v1/healthy
{"healthy":true}
```

#### Swagger Service
By default, we could access swagger UI at [/sw].
- http://localhost:8080/sw

![sw](docs/img/simple-sw.png)

#### TV Service
By default, we could access TV at [/tv].

![tv](docs/img/simple-tv.png)

#### Metrics
By default, we could access prometheus client at [/metrics]
- http://localhost:8080/metrics

![prom](docs/img/simple-prom.png)

#### Logging
By default, we enable zap logger and event logger with console encoding type.
```shell script
2021-11-01T23:28:14.555+0800    INFO    boot/sw_entry.go:201    Bootstrapping SwEntry.  {"eventId": "29b411e0-7e44-4e67-aeb3-95682d2b0a2d", "entryName": "greeter-sw", "entryType": "SwEntry", "jsonPath": "", "path": "/sw/", "port": 8080}
2021-11-01T23:28:14.555+0800    INFO    boot/prom_entry.go:207  Bootstrapping promEntry.        {"eventId": "29b411e0-7e44-4e67-aeb3-95682d2b0a2d", "entryName": "greeter-prom", "entryType": "PromEntry", "entryDescription": "Internal RK entry which implements prometheus client with Echo framework.", "path": "/metrics", "port": 8080}
2021-11-01T23:28:14.555+0800    INFO    boot/common_service_entry.go:156        Bootstrapping CommonServiceEntry.       {"eventId": "29b411e0-7e44-4e67-aeb3-95682d2b0a2d", "entryName": "greeter-commonService", "entryType": "CommonServiceEntry"}
2021-11-01T23:28:14.557+0800    INFO    boot/tv_entry.go:213    Bootstrapping tvEntry.  {"eventId": "29b411e0-7e44-4e67-aeb3-95682d2b0a2d", "entryName": "greeter-tv", "entryType": "TvEntry", "path": "/rk/v1/tv/*item"}
2021-11-01T23:28:14.557+0800    INFO    boot/echo_entry.go:694  Bootstrapping EchoEntry.        {"eventId": "29b411e0-7e44-4e67-aeb3-95682d2b0a2d", "entryName": "greeter", "entryType": "EchoEntry", "port": 8080}
```
```shell script
------------------------------------------------------------------------
endTime=2021-11-01T23:28:14.555288+08:00
startTime=2021-11-01T23:28:14.555234+08:00
elapsedNano=54231
timezone=CST
ids={"eventId":"29b411e0-7e44-4e67-aeb3-95682d2b0a2d"}
app={"appName":"rk-echo","appVersion":"master-e4538d7","entryName":"greeter-sw","entryType":"SwEntry"}
env={"arch":"amd64","az":"*","domain":"*","hostname":"lark.local","localIP":"192.168.1.104","os":"darwin","realm":"*","region":"*"}
payloads={"entryName":"greeter-sw","entryType":"SwEntry","jsonPath":"","path":"/sw/","port":8080}
error={}
counters={}
pairs={}
timing={}
remoteAddr=localhost
operation=bootstrap
resCode=OK
eventStatus=Ended
EOE
...
------------------------------------------------------------------------
endTime=2021-11-01T23:28:14.55714+08:00
startTime=2021-11-01T23:28:14.555172+08:00
elapsedNano=1968399
timezone=CST
ids={"eventId":"29b411e0-7e44-4e67-aeb3-95682d2b0a2d"}
app={"appName":"rk-echo","appVersion":"master-e4538d7","entryName":"greeter","entryType":"EchoEntry"}
env={"arch":"amd64","az":"*","domain":"*","hostname":"lark.local","localIP":"192.168.1.104","os":"darwin","realm":"*","region":"*"}
payloads={"entryName":"greeter","entryType":"EchoEntry","port":8080}
error={}
counters={}
pairs={}
timing={}
remoteAddr=localhost
operation=bootstrap
resCode=OK
eventStatus=Ended
EOE
```

#### Meta
By default, we will send back some metadata to client including gateway with headers.
```shell script
$ curl -vs localhost:8080/rk/v1/healthy
...
< HTTP/1.1 200 OK
< Content-Type: application/json; charset=utf-8
< X-Request-Id: 3332e575-43d8-4bfe-84dd-45b5fc5fb104
< X-Rk-App-Name: rk-echo
< X-Rk-App-Unix-Time: 2021-06-25T01:30:45.143869+08:00
< X-Rk-App-Version: master-xxx
< X-Rk-Received-Time: 2021-06-25T01:30:45.143869+08:00
< X-Trace-Id: 65b9aa7a9705268bba492fdf4a0e5652
< Date: Thu, 24 Jun 2021 17:30:45 GMT
...
```

## YAML Config
Available configuration
User can start multiple echo servers at the same time. Please make sure use different port and name.

### Echo Service
| name | description | type | default value |
| ------ | ------ | ------ | ------ |
| echo.name | The name of echo server | string | N/A |
| echo.port | The port of echo server | integer | nil, server won't start |
| echo.enabled | Enable echo entry or not | bool | false |
| echo.description | Description of echo entry. | string | "" |
| echo.cert.ref | Reference of cert entry declared in [cert entry](https://github.com/rookie-ninja/rk-entry#certentry) | string | "" |
| echo.logger.zapLogger.ref | Reference of zapLoggerEntry declared in [zapLoggerEntry](https://github.com/rookie-ninja/rk-entry#zaploggerentry) | string | "" |
| echo.logger.eventLogger.ref | Reference of eventLoggerEntry declared in [eventLoggerEntry](https://github.com/rookie-ninja/rk-entry#eventloggerentry) | string | "" |

### Common Service
| Path | Description |
| ---- | ---- |
| /rk/v1/apis | List APIs in current EchoEntry. |
| /rk/v1/certs | List CertEntry. |
| /rk/v1/configs | List ConfigEntry. |
| /rk/v1/deps | List dependencies related application, entire contents of go.mod file would be returned. |
| /rk/v1/entries | List all Entries. |
| /rk/v1/gc | Trigger GC |
| /rk/v1/healthy | Get application healthy status. |
| /rk/v1/info | Get application and process info. |
| /rk/v1/license | Get license related application, entire contents of LICENSE file would be returned. |
| /rk/v1/logs | List logger related entries. |
| /rk/v1/git | Get git information. |
| /rk/v1/readme | Get contents of README file. |
| /rk/v1/req | List prometheus metrics of requests. |
| /rk/v1/sys | Get OS stat. |
| /rk/v1/tv | Get HTML page of /tv. |

| name | description | type | default value |
| ------ | ------ | ------ | ------ |
| echo.commonService.enabled | Enable embedded common service | boolean | false |

### Swagger Service
| name | description | type | default value |
| ------ | ------ | ------ | ------ |
| echo.sw.enabled | Enable swagger service over echo server | boolean | false |
| echo.sw.path | The path access swagger service from web | string | /sw |
| echo.sw.jsonPath | Where the swagger.json files are stored locally | string | "" |
| echo.sw.headers | Headers would be sent to caller as scheme of [key:value] | []string | [] |

### Prom Client
| name | description | type | default value |
| ------ | ------ | ------ | ------ |
| echo.prom.enabled | Enable prometheus | boolean | false |
| echo.prom.path | Path of prometheus | string | /metrics |
| echo.prom.pusher.enabled | Enable prometheus pusher | bool | false |
| echo.prom.pusher.jobName | Job name would be attached as label while pushing to remote pushgateway | string | "" |
| echo.prom.pusher.remoteAddress | PushGateWay address, could be form of http://x.x.x.x or x.x.x.x | string | "" |
| echo.prom.pusher.intervalMs | Push interval in milliseconds | string | 1000 |
| echo.prom.pusher.basicAuth | Basic auth used to interact with remote pushgateway, form of [user:pass] | string | "" |
| echo.prom.pusher.cert.ref | Reference of rkentry.CertEntry | string | "" |

### TV Service
| name | description | type | default value |
| ------ | ------ | ------ | ------ |
| echo.tv.enabled | Enable RK TV | boolean | false |

### Interceptors
#### Log
| name | description | type | default value |
| ------ | ------ | ------ | ------ |
| echo.interceptors.loggingZap.enabled | Enable log interceptor | boolean | false |
| echo.interceptors.loggingZap.zapLoggerEncoding | json or console | string | console |
| echo.interceptors.loggingZap.zapLoggerOutputPaths | Output paths | []string | stdout |
| echo.interceptors.loggingZap.eventLoggerEncoding | json or console | string | console |
| echo.interceptors.loggingZap.eventLoggerOutputPaths | Output paths | []string | false |

We will log two types of log for every RPC call.
- zapLogger

Contains user printed logging with requestId or traceId.

- eventLogger

Contains per RPC metadata, response information, environment information and etc.

| Field | Description |
| ---- | ---- |
| endTime | As name described |
| startTime | As name described |
| elapsedNano | Elapsed time for RPC in nanoseconds |
| timezone | As name described |
| ids | Contains three different ids(eventId, requestId and traceId). If meta interceptor was enabled or event.SetRequestId() was called by user, then requestId would be attached. eventId would be the same as requestId if meta interceptor was enabled. If trace interceptor was enabled, then traceId would be attached. |
| app | Contains [appName, appVersion](https://github.com/rookie-ninja/rk-entry#appinfoentry), entryName, entryType. |
| env | Contains arch, az, domain, hostname, localIP, os, realm, region. realm, region, az, domain were retrieved from environment variable named as REALM, REGION, AZ and DOMAIN. "*" means empty environment variable.|
| payloads | Contains RPC related metadata |
| error | Contains errors if occur |
| counters | Set by calling event.SetCounter() by user. |
| pairs | Set by calling event.AddPair() by user. |
| timing | Set by calling event.StartTimer() and event.EndTimer() by user. |
| remoteAddr |  As name described |
| operation | RPC method name |
| resCode | Response code of RPC |
| eventStatus | Ended or InProgress |

- example

```shell script
------------------------------------------------------------------------
endTime=2021-11-01T23:31:01.706614+08:00
startTime=2021-11-01T23:31:01.706335+08:00
elapsedNano=278966
timezone=CST
ids={"eventId":"61cae46e-ea98-47b5-8a39-1090d015e09a","requestId":"61cae46e-ea98-47b5-8a39-1090d015e09a"}
app={"appName":"rk-echo","appVersion":"master-e4538d7","entryName":"greeter","entryType":"EchoEntry"}
env={"arch":"amd64","az":"*","domain":"*","hostname":"lark.local","localIP":"192.168.1.104","os":"darwin","realm":"*","region":"*"}
payloads={"apiMethod":"GET","apiPath":"/rk/v1/healthy","apiProtocol":"HTTP/1.1","apiQuery":"","userAgent":"curl/7.64.1"}
error={}
counters={}
pairs={}
timing={}
remoteAddr=localhost:54376
operation=/rk/v1/healthy
resCode=200
eventStatus=Ended
EOE
```

#### Metrics
| name | description | type | default value |
| ------ | ------ | ------ | ------ |
| echo.interceptors.metricsProm.enabled | Enable metrics interceptor | boolean | false |

#### Auth
Enable the server side auth. codes.Unauthenticated would be returned to client if not authorized with user defined credential.

| name | description | type | default value |
| ------ | ------ | ------ | ------ |
| echo.interceptors.auth.enabled | Enable auth interceptor | boolean | false |
| echo.interceptors.auth.basic | Basic auth credentials as scheme of <user:pass> | []string | [] |
| echo.interceptors.auth.apiKey | API key auth | []string | [] |
| echo.interceptors.auth.ignorePrefix | The paths of prefix that will be ignored by interceptor | []string | [] |

#### Meta
Send application metadata as header to client.

| name | description | type | default value |
| ------ | ------ | ------ | ------ |
| echo.interceptors.meta.enabled | Enable meta interceptor | boolean | false |
| echo.interceptors.meta.prefix | Header key was formed as X-<Prefix>-XXX | string | RK |

#### Tracing
| name | description | type | default value |
| ------ | ------ | ------ | ------ |
| echo.interceptors.tracingTelemetry.enabled | Enable tracing interceptor | boolean | false |
| echo.interceptors.tracingTelemetry.exporter.file.enabled | Enable file exporter | boolean | RK |
| echo.interceptors.tracingTelemetry.exporter.file.outputPath | Export tracing info to files | string | stdout |
| echo.interceptors.tracingTelemetry.exporter.jaeger.agent.enabled | Export tracing info to jaeger agent | boolean | false |
| echo.interceptors.tracingTelemetry.exporter.jaeger.agent.host | As name described | string | localhost |
| echo.interceptors.tracingTelemetry.exporter.jaeger.agent.port | As name described | int | 6831 |
| echo.interceptors.tracingTelemetry.exporter.jaeger.collector.enabled | Export tracing info to jaeger collector | boolean | false |
| echo.interceptors.tracingTelemetry.exporter.jaeger.collector.endpoint | As name described | string | http://localhost:16368/api/trace |
| echo.interceptors.tracingTelemetry.exporter.jaeger.collector.username | As name described | string | "" |
| echo.interceptors.tracingTelemetry.exporter.jaeger.collector.password | As name described | string | "" |

#### RateLimit
| name | description | type | default value |
| ------ | ------ | ------ | ------ |
| echo.interceptors.rateLimit.enabled | Enable rate limit interceptor | boolean | false |
| echo.interceptors.rateLimit.algorithm | Provide algorithm, tokenBucket and leakyBucket are available options | string | tokenBucket |
| echo.interceptors.rateLimit.reqPerSec | Request per second globally | int | 0 |
| echo.interceptors.rateLimit.paths.path | Full path | string | "" |
| echo.interceptors.rateLimit.paths.reqPerSec | Request per second by full path | int | 0 |

#### Timeout
| name | description | type | default value |
| ------ | ------ | ------ | ------ |
| echo.interceptors.timeout.enabled | Enable timeout interceptor | boolean | false |
| echo.interceptors.timeout.timeoutMs | Global timeout in milliseconds. | int | 5000 |
| echo.interceptors.timeout.paths.path | Full path | string | "" |
| echo.interceptors.timeout.paths.timeoutMs | Timeout in milliseconds by full path | int | 5000 |

### Development Status: Stable

### Contributing
We encourage and support an active, healthy community of contributors &mdash;
including you! Details are in the [contribution guide](CONTRIBUTING.md) and
the [code of conduct](CODE_OF_CONDUCT.md). The rk maintainers keep an eye on
issues and pull requests, but you can also report any negative conduct to
lark@rkdev.info.

Released under the [Apache 2.0 License](LICENSE).


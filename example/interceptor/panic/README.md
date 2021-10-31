# Panic interceptor
In this example, we will try to create echo server with panic interceptor enabled.

Panic interceptor will add do the bellow actions.
- Recover from panic
- Convert interface to standard rkerror.ErrorResp style of error
- Set resCode to 500
- Print stacktrace
- Set [panic:1] into event as counters
- Add error into event

**Please make sure panic interceptor to be added at last in chain of interceptors.**

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->
**Table of Contents**  *generated with [DocToc](https://github.com/thlorenz/doctoc)*

- [Quick start](#quick-start)
  - [Code](#code)
- [Example](#example)
  - [Start server](#start-server)
  - [Output](#output)
  - [Code](#code-1)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

## Quick start
Get rk-echo package from the remote repository.

```go
go get -u github.com/rookie-ninja/rk-echo
```
### Code
```go
import     "github.com/rookie-ninja/rk-echo/interceptor/panic"
```
```go
    // ********************************************
    // ********** Enable interceptors *************
    // ********************************************
	interceptors := []echo.MiddlewareFunc{
        rkechopanic.Interceptor(),
    }
```

## Example
We will enable log interceptor to monitor RPC.

### Start server
```shell script
$ go run greeter-server.go
```

### Output
- Server side log (zap & event)
```shell script
2021-11-01T05:02:43.550+0800    ERROR   panic/interceptor.go:36 panic occurs:
goroutine 51 [running]:
...
main.Greeter(0x4ccb1e8, 0xc00026c140, 0x0, 0x0)
        /Users/dongxuny/workspace/rk/rk-echo/example/interceptor/panic/greeter-server.go:69 +0x39
...
created by net/http.(*Server).Serve
        /usr/local/Cellar/go/1.16.3/libexec/src/net/http/server.go:3013 +0x39b
        {"error": "[Internal Server Error] Panic manually!"}
```
```shell script
------------------------------------------------------------------------
endTime=2021-11-01T05:02:43.551205+08:00
startTime=2021-11-01T05:02:43.550745+08:00
elapsedNano=460179
timezone=CST
ids={"eventId":"1d0bf11f-1b0d-46be-9573-99d730885bb6"}
app={"appName":"rk","appVersion":"","entryName":"echo","entryType":"echo"}
env={"arch":"amd64","az":"*","domain":"*","hostname":"lark.local","localIP":"192.168.101.5","os":"darwin","realm":"*","region":"*"}
payloads={"apiMethod":"GET","apiPath":"/rk/v1/greeter","apiProtocol":"HTTP/1.1","apiQuery":"name=rk-dev","userAgent":"curl/7.64.1"}
error={"[Internal Server Error] Panic manually!":1}
counters={"panic":1}
pairs={}
timing={}
remoteAddr=localhost:57093
operation=/rk/v1/greeter
resCode=500
eventStatus=Ended
EOE
```
- Client side
```shell script
$ curl "localhost:8080/rk/v1/greeter?name=rk-dev"
{"error":{"code":500,"status":"Internal Server Error","message":"Panic manually!","details":[]}}
```

### Code
- [greeter-server.go](greeter-server.go)

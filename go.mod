module github.com/rookie-ninja/rk-echo

go 1.16

require (
	github.com/alecthomas/template v0.0.0-20190718012654-fb15b899a751
	github.com/juju/ratelimit v1.0.1
	github.com/labstack/echo/v4 v4.6.1
	github.com/markbates/pkger v0.17.1
	github.com/prometheus/client_golang v1.10.0
	github.com/rookie-ninja/rk-common v1.2.1
	github.com/rookie-ninja/rk-entry v1.0.3
	github.com/rookie-ninja/rk-logger v1.2.3
	github.com/rookie-ninja/rk-prom v1.1.3
	github.com/rookie-ninja/rk-query v1.2.4
	github.com/stretchr/testify v1.7.0
	github.com/swaggo/swag v1.7.4
	go.opentelemetry.io/contrib v1.1.0
	go.opentelemetry.io/otel v1.1.0
	go.opentelemetry.io/otel/exporters/jaeger v1.1.0
	go.opentelemetry.io/otel/exporters/stdout/stdouttrace v1.1.0
	go.opentelemetry.io/otel/sdk v1.1.0
	go.opentelemetry.io/otel/trace v1.1.0
	go.uber.org/ratelimit v0.2.0
	go.uber.org/zap v1.16.0
)

package rkechotrace

import (
	"github.com/labstack/echo/v4"
	"github.com/rookie-ninja/rk-echo/interceptor"
	"github.com/rookie-ninja/rk-echo/interceptor/context"
	"github.com/rookie-ninja/rk-entry/entry"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	oteltrace "go.opentelemetry.io/otel/trace"
)

// Interceptor create a interceptor with opentelemetry.
func Interceptor(opts ...Option) echo.MiddlewareFunc {
	set := newOptionSet(opts...)

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			ctx.Set(rkechointer.RpcEntryNameKey, set.EntryName)
			ctx.Set(rkechointer.RpcTracerKey, set.Tracer)
			ctx.Set(rkechointer.RpcTracerProviderKey, set.Provider)
			ctx.Set(rkechointer.RpcPropagatorKey, set.Propagator)

			span := before(ctx, set)
			defer span.End()

			err := next(ctx)

			after(ctx, span)

			return err
		}
	}
}

func before(ctx echo.Context, set *optionSet) oteltrace.Span {
	opts := []oteltrace.SpanStartOption{
		oteltrace.WithAttributes(semconv.NetAttributesFromHTTPRequest("tcp", ctx.Request())...),
		oteltrace.WithAttributes(semconv.EndUserAttributesFromHTTPRequest(ctx.Request())...),
		oteltrace.WithAttributes(semconv.HTTPServerAttributesFromHTTPRequest(rkentry.GlobalAppCtx.GetAppInfoEntry().AppName, ctx.Path(), ctx.Request())...),
		oteltrace.WithAttributes(localeToAttributes()...),
		oteltrace.WithSpanKind(oteltrace.SpanKindServer),
	}

	// 1: extract tracing info from request header
	spanCtx := oteltrace.SpanContextFromContext(
		set.Propagator.Extract(ctx.Request().Context(), propagation.HeaderCarrier(ctx.Request().Header)))

	spanName := ctx.Path()
	if len(spanName) < 1 {
		spanName = "rk-span-default"
	}

	// 2: start new span
	newRequestCtx, span := set.Tracer.Start(
		oteltrace.ContextWithRemoteSpanContext(ctx.Request().Context(), spanCtx),
		spanName, opts...)
	// 2.1: pass the span through the request context
	ctx.SetRequest(ctx.Request().WithContext(newRequestCtx))

	// 3: read trace id, tracer, traceProvider, propagator and logger into event data and echo context
	rkechoctx.GetEvent(ctx).SetTraceId(span.SpanContext().TraceID().String())
	ctx.Response().Header().Set(rkechoctx.TraceIdKey, span.SpanContext().TraceID().String())

	ctx.Set(rkechointer.RpcSpanKey, span)
	return span
}

func after(ctx echo.Context, span oteltrace.Span) {
	attrs := semconv.HTTPAttributesFromHTTPStatusCode(ctx.Response().Status)
	spanStatus, spanMessage := semconv.SpanStatusFromHTTPStatusCode(ctx.Response().Status)
	span.SetAttributes(attrs...)
	span.SetStatus(spanStatus, spanMessage)
}

// Convert locale information into attributes.
func localeToAttributes() []attribute.KeyValue {
	res := []attribute.KeyValue{
		attribute.String(rkechointer.Realm.Key, rkechointer.Realm.String),
		attribute.String(rkechointer.Region.Key, rkechointer.Region.String),
		attribute.String(rkechointer.AZ.Key, rkechointer.AZ.String),
		attribute.String(rkechointer.Domain.Key, rkechointer.Domain.String),
	}

	return res
}

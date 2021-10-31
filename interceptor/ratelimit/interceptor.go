package rkecholimit

import (
	"github.com/labstack/echo/v4"
	"github.com/rookie-ninja/rk-common/error"
	"github.com/rookie-ninja/rk-echo/interceptor"
	"github.com/rookie-ninja/rk-echo/interceptor/context"
	"net/http"
)

// Interceptor Add rate limit interceptors.
func Interceptor(opts ...Option) echo.MiddlewareFunc {
	set := newOptionSet(opts...)

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			ctx.Set(rkechointer.RpcEntryNameKey, set.EntryName)

			event := rkechoctx.GetEvent(ctx)

			if duration, err := set.Wait(ctx, ctx.Request().URL.Path); err != nil {
				event.SetCounter("rateLimitWaitMs", duration.Milliseconds())
				event.AddErr(err)

				resp := rkerror.New(
					rkerror.WithHttpCode(http.StatusTooManyRequests),
					rkerror.WithDetails(err))

				ctx.JSON(http.StatusTooManyRequests, resp)

				return resp.Err
			}

			return next(ctx)
		}
	}
}

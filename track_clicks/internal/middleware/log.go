package middleware

import (
	"log/slog"
	"time"

	"github.com/labstack/echo/v4"
)

// SlogLogger — middleware для логирования запросов через slog
func SlogLogger() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			req := c.Request()
			res := c.Response()

			err := next(c)

			duration := time.Since(start)

			slog.Info("http request",
				"method", req.Method,
				"path", req.URL.Path,
				"status", res.Status,
				"latency", duration.String(),
				"request_id", res.Header().Get(echo.HeaderXRequestID),
				"error", err,
			)
			return err
		}
	}
}

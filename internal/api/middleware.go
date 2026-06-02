package api

import (
	"time"

	"github/M2A96/Monopoly.git/log"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

const (
	HeaderRequestID     = "X-Request-ID"
	HeaderCorrelationID = "X-Correlation-ID"

	ContextKeyRequestID     = "request_id"
	ContextKeyCorrelationID = "correlation_id"
)

// RequestIDMiddleware generates or propagates a request ID for every request.
func RequestIDMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			rid := c.Request().Header.Get(HeaderRequestID)
			if rid == "" {
				rid = uuid.New().String()
			}
			c.Set(ContextKeyRequestID, rid)
			c.Response().Header().Set(HeaderRequestID, rid)
			return next(c)
		}
	}
}

// CorrelationIDMiddleware propagates or generates a correlation ID for distributed tracing.
func CorrelationIDMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			cid := c.Request().Header.Get(HeaderCorrelationID)
			if cid == "" {
				cid = uuid.New().String()
			}
			c.Set(ContextKeyCorrelationID, cid)
			c.Response().Header().Set(HeaderCorrelationID, cid)
			return next(c)
		}
	}
}

// StructuredLogMiddleware logs every HTTP request with required fields:
// timestamp, request_id, correlation_id, service_name.
func StructuredLogMiddleware(logger log.RuntimeLogger, serviceName string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()

			requestID, _ := c.Get(ContextKeyRequestID).(string)
			correlationID, _ := c.Get(ContextKeyCorrelationID).(string)

			fields := map[string]any{
				"service_name":   serviceName,
				"request_id":     requestID,
				"correlation_id": correlationID,
				"method":         c.Request().Method,
				"path":           c.Request().URL.Path,
				"remote_addr":    c.Request().RemoteAddr,
			}

			logger.WithFields(fields).Info("request started")

			err := next(c)

			fields["status"] = c.Response().Status
			fields["latency_ms"] = time.Since(start).Milliseconds()

			if err != nil {
				fields["error"] = err.Error()
				logger.WithFields(fields).Error("request failed")
			} else {
				logger.WithFields(fields).Info("request completed")
			}

			return err
		}
	}
}

// RequestIDFromContext extracts the request ID from an echo.Context.
func RequestIDFromContext(c echo.Context) string {
	rid, _ := c.Get(ContextKeyRequestID).(string)
	return rid
}

// CorrelationIDFromContext extracts the correlation ID from an echo.Context.
func CorrelationIDFromContext(c echo.Context) string {
	cid, _ := c.Get(ContextKeyCorrelationID).(string)
	return cid
}

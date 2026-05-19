package api

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

type healthHandler struct {
	sqlDB *sql.DB
}

var _ Handler = (*healthHandler)(nil)

// NewHealthHandler creates a simple health/readiness handler for the server.
func NewHealthHandler(
	sqlDB *sql.DB,
) *healthHandler {
	return &healthHandler{
		sqlDB: sqlDB,
	}
}

// RegisterRoutes implements Handler.
func (h *healthHandler) RegisterRoutes(e *echo.Echo) {
	e.GET("/healthz", h.Health)
	e.GET("/readyz", h.Ready)
}

func (h *healthHandler) Health(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{
		"status": "ok",
	})
}

func (h *healthHandler) Ready(c echo.Context) error {
	if h.sqlDB == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{
			"status": "database unavailable",
		})
	}

	ctx, cancel := context.WithTimeout(c.Request().Context(), 2*time.Second)
	defer cancel()

	if err := h.sqlDB.PingContext(ctx); err != nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{
			"status": "database unavailable",
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"status": "ready",
	})
}

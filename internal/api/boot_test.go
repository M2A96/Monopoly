package api_test

import (
	"testing"

	"github/M2A96/Monopoly.git/internal/api"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestBootHandlersCanRegisterRoutes(t *testing.T) {
	e := echo.New()

	healthHandler := api.NewHealthHandler(nil)
	healthHandler.RegisterRoutes(e)

	routes := e.Routes()
	assert.Greater(t, len(routes), 0, "at least one route should be registered")
}

func TestAllRoutePathsHaveAPIV1Prefix(t *testing.T) {
	e := echo.New()

	healthHandler := api.NewHealthHandler(nil)
	healthHandler.RegisterRoutes(e)

	for _, r := range e.Routes() {
		if r.Path != "/health" && r.Path != "/ready" && r.Path != "/healthz" && r.Path != "/readyz" {
			assert.Contains(t, r.Path, "/api/v1/", "route %s should have /api/v1/ prefix", r.Path)
		}
	}
}

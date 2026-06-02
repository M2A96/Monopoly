package api_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github/M2A96/Monopoly.git/internal/api"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestHealthHandler_Health(t *testing.T) {
	e := echo.New()
	h := api.NewHealthHandler(nil)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.Health(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestHealthHandler_Ready_NoDB(t *testing.T) {
	e := echo.New()
	h := api.NewHealthHandler(nil)

	req := httptest.NewRequest(http.MethodGet, "/ready", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.Ready(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
}

func TestHealthHandler_RegisterRoutes(t *testing.T) {
	e := echo.New()
	h := api.NewHealthHandler(nil)
	h.RegisterRoutes(e)

	routes := e.Routes()
	paths := make(map[string]bool)
	for _, r := range routes {
		paths[r.Path] = true
	}

	assert.True(t, paths["/health"], "should have /health route")
	assert.True(t, paths["/ready"], "should have /ready route")
	assert.True(t, paths["/healthz"], "should have /healthz route")
	assert.True(t, paths["/readyz"], "should have /readyz route")
}

package api_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github/M2A96/Monopoly.git/internal/api"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestRequestIDMiddleware_GeneratesID(t *testing.T) {
	e := echo.New()
	e.Use(api.RequestIDMiddleware())
	e.GET("/test", func(c echo.Context) error {
		rid := api.RequestIDFromContext(c)
		assert.NotEmpty(t, rid)
		return c.String(http.StatusOK, rid)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.NotEmpty(t, rec.Header().Get(api.HeaderRequestID))
}

func TestRequestIDMiddleware_PropagatesExistingID(t *testing.T) {
	e := echo.New()
	e.Use(api.RequestIDMiddleware())

	existingID := "existing-request-id-123"
	var capturedID string

	e.GET("/test", func(c echo.Context) error {
		capturedID = api.RequestIDFromContext(c)
		return c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set(api.HeaderRequestID, existingID)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, existingID, capturedID)
	assert.Equal(t, existingID, rec.Header().Get(api.HeaderRequestID))
}

func TestCorrelationIDMiddleware_GeneratesID(t *testing.T) {
	e := echo.New()
	e.Use(api.CorrelationIDMiddleware())
	e.GET("/test", func(c echo.Context) error {
		cid := api.CorrelationIDFromContext(c)
		assert.NotEmpty(t, cid)
		return c.String(http.StatusOK, cid)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.NotEmpty(t, rec.Header().Get(api.HeaderCorrelationID))
}

func TestCorrelationIDMiddleware_PropagatesExistingID(t *testing.T) {
	e := echo.New()
	e.Use(api.CorrelationIDMiddleware())

	existingCID := "my-correlation-id-abc"
	var capturedCID string

	e.GET("/test", func(c echo.Context) error {
		capturedCID = api.CorrelationIDFromContext(c)
		return c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set(api.HeaderCorrelationID, existingCID)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, existingCID, capturedCID)
}

func TestStructuredLogMiddleware_LogsRequest(t *testing.T) {
	e := echo.New()
	logger := &noopLogger{}
	e.Use(api.RequestIDMiddleware())
	e.Use(api.CorrelationIDMiddleware())
	e.Use(api.StructuredLogMiddleware(logger, "test-service"))
	e.GET("/test", func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

// internal/api/property_handlers.go
package api

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/gob"
	"net/http"
	"strconv"

	"github/M2A96/Monopoly.git/config"
	"github/M2A96/Monopoly.git/internal/core/domain/bo"
	"github/M2A96/Monopoly.git/internal/core/ports/input"
	"github/M2A96/Monopoly.git/log"
	"github/M2A96/Monopoly.git/object"
	"github/M2A96/Monopoly.git/object/dao"
	"github/M2A96/Monopoly.git/util"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type (
	PropertyHandler interface {
		GetProperty(c echo.Context) error
		ListProperties(c echo.Context) error
		CreateProperty(c echo.Context) error
		UpdateProperty(c echo.Context) error
		DeleteProperty(c echo.Context) error
		BuyProperty(c echo.Context) error
		MortgageProperty(c echo.Context) error
		AddHouse(c echo.Context) error
		AddHotel(c echo.Context) error
	}

	propertyHandler struct {
		handler[input.PropertyServicer]
	}

	propertyHandlerOptioner = handlerOptioner[input.PropertyServicer]
)

var (
	_ Handler         = (*propertyHandler)(nil)
	_ PropertyHandler = (*propertyHandler)(nil)
)

// NewPropertyHandler creates a new property handler instance
func NewPropertyHandler(
	configConfigger config.Configger,
	logRuntimeLogger log.RuntimeLogger,
	objectUUIDer object.UUIDer,
	traceTracer trace.Tracer,
	optioner ...propertyHandlerOptioner,
) *propertyHandler {
	return &propertyHandler{
		handler: *NewHandler[input.PropertyServicer](
			configConfigger,
			logRuntimeLogger,
			objectUUIDer,
			traceTracer,
			optioner...,
		),
	}
}

// WithPropertyHandlerPropertyServicer sets the property service for the handler
func WithPropertyHandlerPropertyServicer(
	servicePropertyServicer input.PropertyServicer,
) propertyHandlerOptioner {
	return WithHandlerServicer[input.PropertyServicer](servicePropertyServicer)
}

// CreateProperty implements PropertyHandler
func (h *propertyHandler) CreateProperty(c echo.Context) error {
	ctx := c.Request().Context()
	ctxWT, ctxWTCancelFunc := context.WithTimeout(ctx, object.NUMHTTPServerTimeout)
	defer ctxWTCancelFunc()

	var traceSpan trace.Span
	ctxWT, traceSpan = h.GetTracer().Start(
		ctxWT,
		"CreateProperty",
		trace.WithSpanKind(trace.SpanKindServer),
	)
	defer traceSpan.End()

	utilRuntimeContext := util.NewRuntimeContext(ctxWT)
	utilSpanContext := util.NewSpanContext(traceSpan)
	fields := map[string]any{
		"name":   "CreateProperty",
		"rt_ctx": utilRuntimeContext,
		"sp_ctx": utilSpanContext,
		"config": h.GetConfigger(),
	}

	h.GetRuntimeLogger().
		WithFields(fields).
		Info(object.URIEmpty)

	// Parse request body
	var property bo.Propertyer
	if err := c.Bind(&property); err != nil {
		h.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error("Failed to parse request body")
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, "Failed to parse request body")

		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request format",
		})
	}

	// Create property using service
	newPropertyId, err := h.GetServicer().CreateProperty(ctxWT, property)
	if err != nil {
		h.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error("Failed to create property")
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, "Failed to create property")

		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to create property",
		})
	}

	h.GetRuntimeLogger().
		WithFields(fields).
		WithField(object.URIFieldResponse, newPropertyId).
		Debug(object.URIEmpty)

	return c.JSON(http.StatusCreated, newPropertyId)
}

// GetProperty implements PropertyHandler
func (h *propertyHandler) GetProperty(c echo.Context) error {
	ctx := c.Request().Context()
	ctxWT, ctxWTCancelFunc := context.WithTimeout(ctx, object.NUMHTTPServerTimeout)
	defer ctxWTCancelFunc()

	var traceSpan trace.Span
	ctxWT, traceSpan = h.GetTracer().Start(
		ctxWT,
		"GetProperty",
		trace.WithSpanKind(trace.SpanKindServer),
	)
	defer traceSpan.End()

	utilRuntimeContext := util.NewRuntimeContext(ctxWT)
	utilSpanContext := util.NewSpanContext(traceSpan)
	fields := map[string]any{
		"name":   "GetProperty",
		"rt_ctx": utilRuntimeContext,
		"sp_ctx": utilSpanContext,
		"config": h.GetConfigger(),
	}

	h.GetRuntimeLogger().
		WithFields(fields).
		Info(object.URIEmpty)

	// Parse UUID from path parameter
	idStr := c.Param("id")
	id, err := h.GetUUIDer().Parse(idStr)
	if err != nil {
		h.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error("Invalid property ID format")
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, "Invalid property ID format")

		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid property ID format",
		})
	}

	h.GetRuntimeLogger().
		WithFields(fields).
		WithField(object.URIFieldID, id).
		Debug(object.URIEmpty)

	// Get property using service
	property, err := h.GetServicer().Get(ctxWT, id)
	if err != nil {
		h.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error("Failed to retrieve property")
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, "Failed to retrieve property")

		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "Property not found",
		})
	}

	return c.JSON(http.StatusOK, property)
}

// ListProperties implements PropertyHandler
func (h *propertyHandler) ListProperties(c echo.Context) error {
	ctx := c.Request().Context()
	ctxWT, ctxWTCancelFunc := context.WithTimeout(ctx, object.NUMHTTPServerTimeout)
	defer ctxWTCancelFunc()

	var traceSpan trace.Span
	ctxWT, traceSpan = h.GetTracer().Start(
		ctxWT,
		"ListProperties",
		trace.WithSpanKind(trace.SpanKindServer),
	)
	defer traceSpan.End()

	utilRuntimeContext := util.NewRuntimeContext(ctxWT)
	utilSpanContext := util.NewSpanContext(traceSpan)
	fields := map[string]any{
		"name":   "ListProperties",
		"rt_ctx": utilRuntimeContext,
		"sp_ctx": utilSpanContext,
		"config": h.GetConfigger(),
	}

	h.GetRuntimeLogger().
		WithFields(fields).
		Info(object.URIEmpty)

	daoCursor := dao.NewCursor(0)

	pageToken := c.QueryParam("page_token")

	if pageToken != object.URIEmpty {
		h.GetRuntimeLogger().
			WithFields(fields).
			Debug(`pageToken != object.URIEmpty`)

		cursorGobEncoderEncodeBase64URLEncodingDecode, errBase64URLEncodingDecode := base64.URLEncoding.DecodeString(
			pageToken,
		)
		if errBase64URLEncodingDecode != nil {
			h.GetRuntimeLogger().
				WithFields(fields).
				WithField(object.URIFieldError, errBase64URLEncodingDecode).
				Error(object.ErrBase64Decode.Error())
			traceSpan.RecordError(errBase64URLEncodingDecode)
			traceSpan.SetStatus(codes.Error, object.ErrBase64Decode.Error())

			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "Invalid page token",
			})
		}

		h.GetRuntimeLogger().
			WithFields(fields).
			WithField(
				object.URIFieldCursorGobEncoderEncodeBase64URLEncodingDecode,
				cursorGobEncoderEncodeBase64URLEncodingDecode,
			).
			Debug(object.URIEmpty)

		bytesBuffer := bytes.NewBuffer(cursorGobEncoderEncodeBase64URLEncodingDecode)
		gobDecoder := gob.NewDecoder(bytesBuffer)

		if err := gobDecoder.Decode(daoCursor); err != nil {
			h.GetRuntimeLogger().
				WithFields(fields).
				WithField(object.URIFieldError, err).
				Error(object.ErrGobDecoderDecode.Error())
			traceSpan.RecordError(err)
			traceSpan.SetStatus(codes.Error, object.ErrGobDecoderDecode.Error())

			return err
		}
	}

	h.GetRuntimeLogger().
		WithFields(fields).
		WithField(object.URIFieldDAOCursor, daoCursor).
		Debug(object.URIEmpty)

	pageSize, err := strconv.Atoi(c.QueryParam("page_size"))
	if err != nil {
		h.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error("Invalid page size")
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, "Invalid page size")
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid page size",
		})
	}

	daoPagination := dao.NewPagination(daoCursor, uint32(pageSize))

	h.GetRuntimeLogger().
		WithFields(fields).
		WithField(object.URIFieldDAOPagination, daoPagination).
		Debug(object.URIEmpty)

	// Get game ID from path parameter if available
	gameIDStr := c.Param("game_id")
	var propertyFilter dao.PropertyFilter

	if gameIDStr != "" {
		_, err = h.GetUUIDer().Parse(gameIDStr)
		if err != nil {
			h.GetRuntimeLogger().
				WithFields(fields).
				WithField(object.URIFieldError, err).
				Error("Invalid game ID format")
			traceSpan.RecordError(err)
			traceSpan.SetStatus(codes.Error, "Invalid game ID format")
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "Invalid game ID format",
			})
		}
		propertyFilter = dao.NewPropertyFilter(
			[]uuid.UUID{},
			object.URIEmpty,
			object.URIEmpty,
			uuid.Nil,
			0,
			false,
			false,
		)
	} else {
		propertyFilter = dao.NewPropertyFilter(
			[]uuid.UUID{},
			object.URIEmpty,
			object.URIEmpty,
			uuid.Nil,
			0,
			false,
			false,
		)
	}

	h.GetRuntimeLogger().
		WithFields(fields).
		WithField(object.URIFieldDAOPropertyFilter, propertyFilter).
		Debug(object.URIEmpty)

	properties, cursor, err := h.GetServicer().List(
		ctxWT,
		daoPagination,
		propertyFilter,
	)
	if err != nil {
		h.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error("Failed to list properties")
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, "Failed to list properties")

		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to list properties",
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"properties": properties,
		"cursor":     cursor,
	})
}

// UpdateProperty implements PropertyHandler
func (h *propertyHandler) UpdateProperty(c echo.Context) error {
	ctx := c.Request().Context()
	ctxWT, ctxWTCancelFunc := context.WithTimeout(ctx, object.NUMHTTPServerTimeout)
	defer ctxWTCancelFunc()

	var traceSpan trace.Span
	ctxWT, traceSpan = h.GetTracer().Start(
		ctxWT,
		"UpdateProperty",
		trace.WithSpanKind(trace.SpanKindServer),
	)
	defer traceSpan.End()

	utilRuntimeContext := util.NewRuntimeContext(ctxWT)
	utilSpanContext := util.NewSpanContext(traceSpan)
	fields := map[string]any{
		"name":   "UpdateProperty",
		"rt_ctx": utilRuntimeContext,
		"sp_ctx": utilSpanContext,
		"config": h.GetConfigger(),
	}

	h.GetRuntimeLogger().
		WithFields(fields).
		Info(object.URIEmpty)

	// Parse property ID from path parameter
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error("Invalid property ID format")
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, "Invalid property ID format")

		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid property ID format",
		})
	}

	// Parse request body
	var property bo.Propertyer
	if err := c.Bind(&property); err != nil {
		h.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error("Failed to parse request body")
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, "Failed to parse request body")

		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request format",
		})
	}

	// Update property using service
	if err := h.GetServicer().UpdateProperty(ctxWT, id, property); err != nil {
		h.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error("Failed to update property")
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, "Failed to update property")

		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to update property",
		})
	}

	return c.NoContent(http.StatusOK)
}

// DeleteProperty implements PropertyHandler
func (h *propertyHandler) DeleteProperty(c echo.Context) error {
	ctx := c.Request().Context()
	ctxWT, ctxWTCancelFunc := context.WithTimeout(ctx, object.NUMHTTPServerTimeout)
	defer ctxWTCancelFunc()

	var traceSpan trace.Span
	ctxWT, traceSpan = h.GetTracer().Start(
		ctxWT,
		"DeleteProperty",
		trace.WithSpanKind(trace.SpanKindServer),
	)
	defer traceSpan.End()

	utilRuntimeContext := util.NewRuntimeContext(ctxWT)
	utilSpanContext := util.NewSpanContext(traceSpan)
	fields := map[string]any{
		"name":   "DeleteProperty",
		"rt_ctx": utilRuntimeContext,
		"sp_ctx": utilSpanContext,
		"config": h.GetConfigger(),
	}

	h.GetRuntimeLogger().
		WithFields(fields).
		Info(object.URIEmpty)

	// Parse UUID from path parameter
	idStr := c.Param("id")
	id, err := h.GetUUIDer().Parse(idStr)
	if err != nil {
		h.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error("Invalid property ID format")
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, "Invalid property ID format")

		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid property ID format",
		})
	}

	// Delete property using service
	if err := h.GetServicer().DeleteProperty(
		ctxWT,
		id,
	); err != nil {
		h.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error("Failed to delete property")
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, "Failed to delete property")

		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to delete property",
		})
	}

	return c.NoContent(http.StatusOK)
}

// BuyProperty implements PropertyHandler
func (h *propertyHandler) BuyProperty(c echo.Context) error {
	ctx := c.Request().Context()
	ctxWT, ctxWTCancelFunc := context.WithTimeout(ctx, object.NUMHTTPServerTimeout)
	defer ctxWTCancelFunc()

	var traceSpan trace.Span
	ctxWT, traceSpan = h.GetTracer().Start(
		ctxWT,
		"BuyProperty",
		trace.WithSpanKind(trace.SpanKindServer),
	)
	defer traceSpan.End()

	utilRuntimeContext := util.NewRuntimeContext(ctxWT)
	utilSpanContext := util.NewSpanContext(traceSpan)
	fields := map[string]any{
		"name":   "BuyProperty",
		"rt_ctx": utilRuntimeContext,
		"sp_ctx": utilSpanContext,
		"config": h.GetConfigger(),
	}

	h.GetRuntimeLogger().
		WithFields(fields).
		Info(object.URIEmpty)

	// Parse property ID from path parameter
	propertyIDStr := c.Param("property_id")
	propertyID, err := h.GetUUIDer().Parse(propertyIDStr)
	if err != nil {
		h.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error("Invalid property ID format")
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, "Invalid property ID format")

		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid property ID format",
		})
	}

	// Parse player ID from request body
	type buyPropertyRequest struct {
		PlayerID string `json:"player_id"`
	}

	var req buyPropertyRequest
	if err := c.Bind(&req); err != nil {
		h.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error("Failed to parse request body")
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, "Failed to parse request body")

		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request format",
		})
	}

	playerID, err := h.GetUUIDer().Parse(req.PlayerID)
	if err != nil {
		h.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error("Invalid player ID format")
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, "Invalid player ID format")

		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid player ID format",
		})
	}

	// Buy property using service
	if err := h.GetServicer().BuyProperty(ctxWT, propertyID, playerID); err != nil {
		h.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error("Failed to buy property")
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, "Failed to buy property")

		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to buy property",
		})
	}

	return c.NoContent(http.StatusOK)
}

// MortgageProperty implements PropertyHandler
func (h *propertyHandler) MortgageProperty(c echo.Context) error {
	ctx := c.Request().Context()
	ctxWT, ctxWTCancelFunc := context.WithTimeout(ctx, object.NUMHTTPServerTimeout)
	defer ctxWTCancelFunc()

	var traceSpan trace.Span
	ctxWT, traceSpan = h.GetTracer().Start(
		ctxWT,
		"MortgageProperty",
		trace.WithSpanKind(trace.SpanKindServer),
	)
	defer traceSpan.End()

	utilRuntimeContext := util.NewRuntimeContext(ctxWT)
	utilSpanContext := util.NewSpanContext(traceSpan)
	fields := map[string]any{
		"name":   "MortgageProperty",
		"rt_ctx": utilRuntimeContext,
		"sp_ctx": utilSpanContext,
		"config": h.GetConfigger(),
	}

	h.GetRuntimeLogger().
		WithFields(fields).
		Info(object.URIEmpty)

	// Parse property ID from path parameter
	propertyIDStr := c.Param("property_id")
	propertyID, err := h.GetUUIDer().Parse(propertyIDStr)
	if err != nil {
		h.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error("Invalid property ID format")
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, "Invalid property ID format")

		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid property ID format",
		})
	}

	// Parse mortgage status from request body
	type mortgagePropertyRequest struct {
		Mortgaged bool `json:"mortgaged"`
	}

	var req mortgagePropertyRequest
	if err := c.Bind(&req); err != nil {
		h.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error("Failed to parse request body")
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, "Failed to parse request body")

		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request format",
		})
	}

	// Mortgage property using service
	if err := h.GetServicer().MortgageProperty(ctxWT, propertyID, req.Mortgaged); err != nil {
		h.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error("Failed to mortgage property")
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, "Failed to mortgage property")

		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to mortgage property",
		})
	}

	return c.NoContent(http.StatusOK)
}

// AddHouse implements PropertyHandler
func (h *propertyHandler) AddHouse(c echo.Context) error {
	ctx := c.Request().Context()
	ctxWT, ctxWTCancelFunc := context.WithTimeout(ctx, object.NUMHTTPServerTimeout)
	defer ctxWTCancelFunc()

	var traceSpan trace.Span
	ctxWT, traceSpan = h.GetTracer().Start(
		ctxWT,
		"AddHouse",
		trace.WithSpanKind(trace.SpanKindServer),
	)
	defer traceSpan.End()

	utilRuntimeContext := util.NewRuntimeContext(ctxWT)
	utilSpanContext := util.NewSpanContext(traceSpan)
	fields := map[string]any{
		"name":   "AddHouse",
		"rt_ctx": utilRuntimeContext,
		"sp_ctx": utilSpanContext,
		"config": h.GetConfigger(),
	}

	h.GetRuntimeLogger().
		WithFields(fields).
		Info(object.URIEmpty)

	// Parse property ID from path parameter
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error("Invalid property ID format")
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, "Invalid property ID format")

		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid property ID format",
		})
	}

	// Add house using service
	if err := h.GetServicer().AddHouse(ctxWT, id); err != nil {
		h.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error("Failed to add house to property")
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, "Failed to add house to property")

		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to add house to property",
		})
	}

	return c.NoContent(http.StatusOK)
}

// AddHotel implements PropertyHandler
func (h *propertyHandler) AddHotel(c echo.Context) error {
	ctx := c.Request().Context()
	ctxWT, ctxWTCancelFunc := context.WithTimeout(ctx, object.NUMHTTPServerTimeout)
	defer ctxWTCancelFunc()

	var traceSpan trace.Span
	ctxWT, traceSpan = h.GetTracer().Start(
		ctxWT,
		"AddHotel",
		trace.WithSpanKind(trace.SpanKindServer),
	)
	defer traceSpan.End()

	utilRuntimeContext := util.NewRuntimeContext(ctxWT)
	utilSpanContext := util.NewSpanContext(traceSpan)
	fields := map[string]any{
		"name":   "AddHotel",
		"rt_ctx": utilRuntimeContext,
		"sp_ctx": utilSpanContext,
		"config": h.GetConfigger(),
	}

	h.GetRuntimeLogger().
		WithFields(fields).
		Info(object.URIEmpty)

	// Parse property ID from path parameter
	idStr := c.Param("id")
	id, err := h.GetUUIDer().Parse(idStr)
	if err != nil {
		h.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error("Invalid property ID format")
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, "Invalid property ID format")

		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid property ID format",
		})
	}

	// Add hotel using service
	if err := h.GetServicer().AddHotel(ctxWT, id); err != nil {
		h.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error("Failed to add hotel to property")
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, "Failed to add hotel to property")

		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to add hotel to property",
		})
	}

	return c.NoContent(http.StatusOK)
}

// RegisterRoutes implements Handler.
func (h *propertyHandler) RegisterRoutes(e *echo.Echo) {
	e.POST("/properties", h.CreateProperty)
	e.GET("/properties/:id", h.GetProperty)
	e.GET("/properties", h.ListProperties)
	e.GET("/games/:game_id/properties", h.ListProperties)
	e.PUT("/properties/:id", h.UpdateProperty)
	e.DELETE("/properties/:id", h.DeleteProperty)
	e.POST("/properties/:property_id/buy", h.BuyProperty)
	e.PUT("/properties/:property_id/mortgage", h.MortgageProperty)
	e.POST("/properties/:id/houses", h.AddHouse)
	e.POST("/properties/:id/hotels", h.AddHotel)
}

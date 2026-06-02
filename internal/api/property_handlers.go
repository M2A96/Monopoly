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

func WithPropertyHandlerPropertyServicer(
	servicePropertyServicer input.PropertyServicer,
) propertyHandlerOptioner {
	return WithHandlerServicer[input.PropertyServicer](servicePropertyServicer)
}

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

	newPropertyID, err := h.GetServicer().CreateProperty(ctxWT, property)
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
		WithField(object.URIFieldResponse, newPropertyID).
		Debug(object.URIEmpty)

	return c.JSON(http.StatusCreated, map[string]string{
		"id": newPropertyID.String(),
	})
}

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
		cursorBytes, errDecode := base64.URLEncoding.DecodeString(pageToken)
		if errDecode != nil {
			h.GetRuntimeLogger().
				WithFields(fields).
				WithField(object.URIFieldError, errDecode).
				Error(object.ErrBase64Decode.Error())
			traceSpan.RecordError(errDecode)
			traceSpan.SetStatus(codes.Error, object.ErrBase64Decode.Error())

			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "Invalid page token",
			})
		}

		if err := gob.NewDecoder(bytes.NewBuffer(cursorBytes)).Decode(daoCursor); err != nil {
			h.GetRuntimeLogger().
				WithFields(fields).
				WithField(object.URIFieldError, err).
				Error(object.ErrGobDecoderDecode.Error())
			traceSpan.RecordError(err)
			traceSpan.SetStatus(codes.Error, object.ErrGobDecoderDecode.Error())

			return err
		}
	}

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

	propertyFilter := dao.NewPropertyFilter(
		[]uuid.UUID{},
		object.URIEmpty,
		object.URIEmpty,
		uuid.Nil,
		0,
		false,
		false,
	)

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

func (h *propertyHandler) RegisterRoutes(e *echo.Echo) {
	e.POST("/api/v1/properties", h.CreateProperty)
	e.GET("/api/v1/properties/:id", h.GetProperty)
	e.GET("/api/v1/properties", h.ListProperties)
}

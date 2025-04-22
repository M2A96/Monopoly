// internal/api/trade_request_handlers.go
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
	TradeRequestHandler interface {
		GetTradeRequest(c echo.Context) error
		ListTradeRequests(c echo.Context) error
		CreateTradeRequest(c echo.Context) error
		UpdateTradeRequest(c echo.Context) error
		DeleteTradeRequest(c echo.Context) error
		AcceptTradeRequest(c echo.Context) error
		RejectTradeRequest(c echo.Context) error
		GetTradeRequestsByGameID(c echo.Context) error
		GetTradeRequestsByPlayerID(c echo.Context) error
	}

	tradeRequestHandler struct {
		handler[input.TradeRequestServicer]
	}

	tradeRequestHandlerOptioner = handlerOptioner[input.TradeRequestServicer]
)

var (
	_ Handler             = (*tradeRequestHandler)(nil)
	_ TradeRequestHandler = (*tradeRequestHandler)(nil)
)

// NewTradeRequestHandler creates a new trade request handler instance
func NewTradeRequestHandler(
	configConfigger config.Configger,
	logRuntimeLogger log.RuntimeLogger,
	objectUUIDer object.UUIDer,
	traceTracer trace.Tracer,
	optioner ...tradeRequestHandlerOptioner,
) *tradeRequestHandler {
	return &tradeRequestHandler{
		handler: *NewHandler[input.TradeRequestServicer](
			configConfigger,
			logRuntimeLogger,
			objectUUIDer,
			traceTracer,
			optioner...,
		),
	}
}

// WithTradeRequestHandlerTradeRequestServicer sets the trade request service for the handler
func WithTradeRequestHandlerTradeRequestServicer(
	serviceTradeRequestServicer input.TradeRequestServicer,
) tradeRequestHandlerOptioner {
	return WithHandlerServicer[input.TradeRequestServicer](serviceTradeRequestServicer)
}

// CreateTradeRequest implements TradeRequestHandler
func (h *tradeRequestHandler) CreateTradeRequest(c echo.Context) error {
	ctx := c.Request().Context()
	ctxWT, ctxWTCancelFunc := context.WithTimeout(ctx, object.NUMHTTPServerTimeout)
	defer ctxWTCancelFunc()

	var traceSpan trace.Span
	ctxWT, traceSpan = h.GetTracer().Start(
		ctxWT,
		"CreateTradeRequest",
		trace.WithSpanKind(trace.SpanKindServer),
	)
	defer traceSpan.End()

	utilRuntimeContext := util.NewRuntimeContext(ctxWT)
	utilSpanContext := util.NewSpanContext(traceSpan)
	fields := map[string]any{
		"name":   "CreateTradeRequest",
		"rt_ctx": utilRuntimeContext,
		"sp_ctx": utilSpanContext,
		"config": h.GetConfigger(),
	}

	h.GetRuntimeLogger().
		WithFields(fields).
		Info(object.URIEmpty)

	// Parse request body
	var tradeRequest bo.TradeRequester
	if err := c.Bind(&tradeRequest); err != nil {
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

	// Create trade request using service
	newTradeRequestId, err := h.GetServicer().CreateTradeRequest(ctxWT, tradeRequest)
	if err != nil {
		h.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error("Failed to create trade request")
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, "Failed to create trade request")

		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to create trade request",
		})
	}

	// Convert UUID to base62 for shorter representation
	base62Id := util.UUIDToBase62(newTradeRequestId)

	h.GetRuntimeLogger().
		WithFields(fields).
		WithField(object.URIFieldResponse, newTradeRequestId).
		WithField("base62_id", base62Id).
		Debug(object.URIEmpty)

	return c.JSON(http.StatusCreated, map[string]string{
		"id": base62Id,
	})
}

// GetTradeRequest implements TradeRequestHandler
func (h *tradeRequestHandler) GetTradeRequest(c echo.Context) error {
	ctx := c.Request().Context()
	ctxWT, ctxWTCancelFunc := context.WithTimeout(ctx, object.NUMHTTPServerTimeout)
	defer ctxWTCancelFunc()

	var traceSpan trace.Span
	ctxWT, traceSpan = h.GetTracer().Start(
		ctxWT,
		"GetTradeRequest",
		trace.WithSpanKind(trace.SpanKindServer),
	)
	defer traceSpan.End()

	utilRuntimeContext := util.NewRuntimeContext(ctxWT)
	utilSpanContext := util.NewSpanContext(traceSpan)
	fields := map[string]any{
		"name":   "GetTradeRequest",
		"rt_ctx": utilRuntimeContext,
		"sp_ctx": utilSpanContext,
		"config": h.GetConfigger(),
	}

	h.GetRuntimeLogger().
		WithFields(fields).
		Info(object.URIEmpty)

	// Parse ID from path parameter - could be UUID or base62 encoded
	idStr := c.Param("id")
	var id uuid.UUID
	var err error

	// Try parsing as standard UUID first
	id, err = h.GetUUIDer().Parse(idStr)
	if err != nil {
		// If not a standard UUID, try parsing as base62
		id, err = util.Base62ToUUID(idStr)
		if err != nil {
			h.GetRuntimeLogger().
				WithFields(fields).
				WithField(object.URIFieldError, err).
				Error("Invalid trade request ID format - not a valid UUID or base62 ID")
			traceSpan.RecordError(err)
			traceSpan.SetStatus(codes.Error, object.ErrUUIDerParse.Error())

			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "Invalid trade request ID format",
			})
		}
	}

	h.GetRuntimeLogger().
		WithFields(fields).
		WithField(object.URIFieldID, id).
		Debug(object.URIEmpty)

	// Get trade request using service
	tradeRequest, err := h.GetServicer().Get(ctxWT, id)
	if err != nil {
		h.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error("Failed to retrieve trade request")
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, "Failed to retrieve trade request")

		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "Trade request not found",
		})
	}

	return c.JSON(http.StatusOK, tradeRequest)
}

// ListTradeRequests implements TradeRequestHandler
func (h *tradeRequestHandler) ListTradeRequests(c echo.Context) error {
	ctx := c.Request().Context()
	ctxWT, ctxWTCancelFunc := context.WithTimeout(ctx, object.NUMHTTPServerTimeout)
	defer ctxWTCancelFunc()

	var traceSpan trace.Span
	ctxWT, traceSpan = h.GetTracer().Start(
		ctxWT,
		"ListTradeRequests",
		trace.WithSpanKind(trace.SpanKindServer),
	)
	defer traceSpan.End()

	utilRuntimeContext := util.NewRuntimeContext(ctxWT)
	utilSpanContext := util.NewSpanContext(traceSpan)
	fields := map[string]any{
		"name":   "ListTradeRequests",
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

	// Create trade filter based on query parameters
	// This is a simplified example - adjust based on your actual filter requirements
	daoTradeFilter := dao.NewTradeFilter(
		[]uuid.UUID{},
		uuid.Nil, // Sender ID
		uuid.Nil, // Receiver ID
		"",
		uuid.Nil, // Status
	)

	h.GetRuntimeLogger().
		WithFields(fields).
		WithField("dao_trade_filter", daoTradeFilter).
		Debug(object.URIEmpty)

	tradeRequests, cursor, err := h.GetServicer().List(
		ctxWT,
		daoPagination,
		daoTradeFilter,
	)
	if err != nil {
		h.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error("Failed to list trade requests")
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, "Failed to list trade requests")

		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to list trade requests",
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"trade_requests": tradeRequests,
		"cursor":         cursor,
	})
}

// UpdateTradeRequest implements TradeRequestHandler
func (h *tradeRequestHandler) UpdateTradeRequest(c echo.Context) error {
	ctx := c.Request().Context()
	ctxWT, ctxWTCancelFunc := context.WithTimeout(ctx, object.NUMHTTPServerTimeout)
	defer ctxWTCancelFunc()

	var traceSpan trace.Span
	ctxWT, traceSpan = h.GetTracer().Start(
		ctxWT,
		"UpdateTradeRequest",
		trace.WithSpanKind(trace.SpanKindServer),
	)
	defer traceSpan.End()

	utilRuntimeContext := util.NewRuntimeContext(ctxWT)
	utilSpanContext := util.NewSpanContext(traceSpan)
	fields := map[string]any{
		"name":   "UpdateTradeRequest",
		"rt_ctx": utilRuntimeContext,
		"sp_ctx": utilSpanContext,
		"config": h.GetConfigger(),
	}

	h.GetRuntimeLogger().
		WithFields(fields).
		Info(object.URIEmpty)

	// Parse ID from path parameter - could be UUID or base62 encoded
	idStr := c.Param("id")
	var id uuid.UUID
	var err error

	// Try parsing as standard UUID first
	id, err = h.GetUUIDer().Parse(idStr)
	if err != nil {
		// If not a standard UUID, try parsing as base62
		id, err = util.Base62ToUUID(idStr)
		if err != nil {
			h.GetRuntimeLogger().
				WithFields(fields).
				WithField(object.URIFieldError, err).
				Error("Invalid trade request ID format - not a valid UUID or base62 ID")
			traceSpan.RecordError(err)
			traceSpan.SetStatus(codes.Error, object.ErrUUIDerParse.Error())

			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "Invalid trade request ID format",
			})
		}
	}

	h.GetRuntimeLogger().
		WithFields(fields).
		WithField(object.URIFieldID, id).
		Debug(object.URIEmpty)

	// Parse request body
	var tradeRequest bo.TradeRequester
	if err := c.Bind(&tradeRequest); err != nil {
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

	// Update trade request using service
	err = h.GetServicer().UpdateTradeRequest(ctxWT, id, tradeRequest)
	if err != nil {
		h.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error("Failed to update trade request")
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, "Failed to update trade request")

		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to update trade request",
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Trade request updated successfully",
	})
}

// DeleteTradeRequest implements TradeRequestHandler
func (h *tradeRequestHandler) DeleteTradeRequest(c echo.Context) error {
	ctx := c.Request().Context()
	ctxWT, ctxWTCancelFunc := context.WithTimeout(ctx, object.NUMHTTPServerTimeout)
	defer ctxWTCancelFunc()

	var traceSpan trace.Span
	ctxWT, traceSpan = h.GetTracer().Start(
		ctxWT,
		"DeleteTradeRequest",
		trace.WithSpanKind(trace.SpanKindServer),
	)
	defer traceSpan.End()

	utilRuntimeContext := util.NewRuntimeContext(ctxWT)
	utilSpanContext := util.NewSpanContext(traceSpan)
	fields := map[string]any{
		"name":   "DeleteTradeRequest",
		"rt_ctx": utilRuntimeContext,
		"sp_ctx": utilSpanContext,
		"config": h.GetConfigger(),
	}

	h.GetRuntimeLogger().
		WithFields(fields).
		Info(object.URIEmpty)

	// Parse ID from path parameter - could be UUID or base62 encoded
	idStr := c.Param("id")
	var id uuid.UUID
	var err error

	// Try parsing as standard UUID first
	id, err = h.GetUUIDer().Parse(idStr)
	if err != nil {
		// If not a standard UUID, try parsing as base62
		id, err = util.Base62ToUUID(idStr)
		if err != nil {
			h.GetRuntimeLogger().
				WithFields(fields).
				WithField(object.URIFieldError, err).
				Error("Invalid trade request ID format - not a valid UUID or base62 ID")
			traceSpan.RecordError(err)
			traceSpan.SetStatus(codes.Error, object.ErrUUIDerParse.Error())

			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "Invalid trade request ID format",
			})
		}
	}

	h.GetRuntimeLogger().
		WithFields(fields).
		WithField(object.URIFieldID, id).
		Debug(object.URIEmpty)

	// Delete trade request using service
	err = h.GetServicer().DeleteTradeRequest(ctxWT, id)
	if err != nil {
		h.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error("Failed to delete trade request")
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, "Failed to delete trade request")

		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to delete trade request",
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Trade request deleted successfully",
	})
}

// AcceptTradeRequest implements TradeRequestHandler.
func (h *tradeRequestHandler) AcceptTradeRequest(c echo.Context) error {
	ctx := c.Request().Context()
	ctxWT, ctxWTCancelFunc := context.WithTimeout(ctx, object.NUMHTTPServerTimeout)
	defer ctxWTCancelFunc()

	var traceSpan trace.Span
	ctxWT, traceSpan = h.GetTracer().Start(
		ctxWT,
		"AcceptTradeRequest",
		trace.WithSpanKind(trace.SpanKindServer),
	)
	defer traceSpan.End()

	utilRuntimeContext := util.NewRuntimeContext(ctxWT)
	utilSpanContext := util.NewSpanContext(traceSpan)
	fields := map[string]any{
		"name":   "AcceptTradeRequest",
		"rt_ctx": utilRuntimeContext,
		"sp_ctx": utilSpanContext,
		"config": h.GetConfigger(),
	}

	h.GetRuntimeLogger().
		WithFields(fields).
		Info(object.URIEmpty)

	// Parse ID from path parameter - could be UUID or base62 encoded
	idStr := c.Param("id")
	var id uuid.UUID
	var err error

	// Try parsing as standard UUID first
	id, err = h.GetUUIDer().Parse(idStr)
	if err != nil {
		// If not a standard UUID, try parsing as base62
		id, err = util.Base62ToUUID(idStr)
		if err != nil {
			h.GetRuntimeLogger().
				WithFields(fields).
				WithField(object.URIFieldError, err).
				Error("Invalid trade request ID format - not a valid UUID or base62 ID")
			traceSpan.RecordError(err)
			traceSpan.SetStatus(codes.Error, object.ErrUUIDerParse.Error())

			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "Invalid trade request ID format",
			})
		}
	}

	h.GetRuntimeLogger().
		WithFields(fields).
		WithField(object.URIFieldID, id).
		Debug(object.URIEmpty)

	// Accept trade request using service
	err = h.GetServicer().AcceptTradeRequest(ctxWT, id)
	if err != nil {
		h.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error("Failed to accept trade request")
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, "Failed to accept trade request")

		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to accept trade request",
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Trade request accepted successfully",
	})
}

// RejectTradeRequest implements TradeRequestHandler
func (h *tradeRequestHandler) RejectTradeRequest(c echo.Context) error {
	ctx := c.Request().Context()
	ctxWT, ctxWTCancelFunc := context.WithTimeout(ctx, object.NUMHTTPServerTimeout)
	defer ctxWTCancelFunc()

	var traceSpan trace.Span
	ctxWT, traceSpan = h.GetTracer().Start(
		ctxWT,
		"RejectTradeRequest",
		trace.WithSpanKind(trace.SpanKindServer),
	)
	defer traceSpan.End()

	utilRuntimeContext := util.NewRuntimeContext(ctxWT)
	utilSpanContext := util.NewSpanContext(traceSpan)
	fields := map[string]any{
		"name":   "RejectTradeRequest",
		"rt_ctx": utilRuntimeContext,
		"sp_ctx": utilSpanContext,
		"config": h.GetConfigger(),
	}

	h.GetRuntimeLogger().
		WithFields(fields).
		Info(object.URIEmpty)

	// Parse ID from path parameter - could be UUID or base62 encoded
	idStr := c.Param("id")
	var id uuid.UUID
	var err error

	// Try parsing as standard UUID first
	id, err = h.GetUUIDer().Parse(idStr)
	if err != nil {
		// If not a standard UUID, try parsing as base62
		id, err = util.Base62ToUUID(idStr)
		if err != nil {
			h.GetRuntimeLogger().
				WithFields(fields).
				WithField(object.URIFieldError, err).
				Error("Invalid trade request ID format - not a valid UUID or base62 ID")
			traceSpan.RecordError(err)
			traceSpan.SetStatus(codes.Error, object.ErrUUIDerParse.Error())

			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "Invalid trade request ID format",
			})
		}
	}

	h.GetRuntimeLogger().
		WithFields(fields).
		WithField(object.URIFieldID, id).
		Debug(object.URIEmpty)

	// Reject trade request using service
	err = h.GetServicer().RejectTradeRequest(ctxWT, id)
	if err != nil {
		h.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error("Failed to reject trade request")
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, "Failed to reject trade request")

		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to reject trade request",
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Trade request rejected successfully",
	})
}

// GetTradeRequestsByGameID implements TradeRequestHandler
func (h *tradeRequestHandler) GetTradeRequestsByGameID(c echo.Context) error {
	ctx := c.Request().Context()
	ctxWT, ctxWTCancelFunc := context.WithTimeout(ctx, object.NUMHTTPServerTimeout)
	defer ctxWTCancelFunc()

	var traceSpan trace.Span
	ctxWT, traceSpan = h.GetTracer().Start(
		ctxWT,
		"GetTradeRequestsByGameID",
		trace.WithSpanKind(trace.SpanKindServer),
	)
	defer traceSpan.End()

	utilRuntimeContext := util.NewRuntimeContext(ctxWT)
	utilSpanContext := util.NewSpanContext(traceSpan)
	fields := map[string]any{
		"name":   "GetTradeRequestsByGameID",
		"rt_ctx": utilRuntimeContext,
		"sp_ctx": utilSpanContext,
		"config": h.GetConfigger(),
	}

	h.GetRuntimeLogger().
		WithFields(fields).
		Info(object.URIEmpty)

	// Parse game ID from path parameter
	gameIDStr := c.Param("game_id")
	gameID, err := h.GetUUIDer().Parse(gameIDStr)
	if err != nil {
		h.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error("Invalid game ID format")
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, object.ErrUUIDerParse.Error())

		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid game ID format",
		})
	}

	h.GetRuntimeLogger().
		WithFields(fields).
		WithField("game_id", gameID).
		Debug(object.URIEmpty)

	// Handle pagination
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

	// Get trade requests by game ID
	tradeRequests, cursor, err := h.GetServicer().GetTradeRequestsByGameID(ctxWT, gameID, daoPagination)
	if err != nil {
		h.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error("Failed to retrieve trade requests for game")
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, "Failed to retrieve trade requests for game")

		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to retrieve trade requests for game",
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"trade_requests": tradeRequests,
		"cursor":         cursor,
	})
}

// GetTradeRequestsByPlayerID implements TradeRequestHandler
func (h *tradeRequestHandler) GetTradeRequestsByPlayerID(c echo.Context) error {
	ctx := c.Request().Context()
	ctxWT, ctxWTCancelFunc := context.WithTimeout(ctx, object.NUMHTTPServerTimeout)
	defer ctxWTCancelFunc()

	var traceSpan trace.Span
	ctxWT, traceSpan = h.GetTracer().Start(
		ctxWT,
		"GetTradeRequestsByPlayerID",
		trace.WithSpanKind(trace.SpanKindServer),
	)
	defer traceSpan.End()

	utilRuntimeContext := util.NewRuntimeContext(ctxWT)
	utilSpanContext := util.NewSpanContext(traceSpan)
	fields := map[string]any{
		"name":   "GetTradeRequestsByPlayerID",
		"rt_ctx": utilRuntimeContext,
		"sp_ctx": utilSpanContext,
		"config": h.GetConfigger(),
	}

	h.GetRuntimeLogger().
		WithFields(fields).
		Info(object.URIEmpty)

	// Parse player ID from path parameter
	playerIDStr := c.Param("player_id")
	playerID, err := h.GetUUIDer().Parse(playerIDStr)
	if err != nil {
		h.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error("Invalid player ID format")
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, object.ErrUUIDerParse.Error())

		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid player ID format",
		})
	}

	h.GetRuntimeLogger().
		WithFields(fields).
		WithField("player_id", playerID).
		Debug(object.URIEmpty)

	// Handle pagination
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

	// Get trade requests by player ID
	tradeRequests, cursor, err := h.GetServicer().GetTradeRequestsByPlayerID(ctxWT, playerID, daoPagination)
	if err != nil {
		h.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error("Failed to retrieve trade requests for player")
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, "Failed to retrieve trade requests for player")

		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to retrieve trade requests for player",
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"trade_requests": tradeRequests,
		"cursor":         cursor,
	})
}

// RegisterRoutes implements Handler
func (h *tradeRequestHandler) RegisterRoutes(e *echo.Echo) {
	tradeRequestGroup := e.Group("/trade-requests")

	// Single trade request routes
	tradeRequestGroup.GET("/:id", h.GetTradeRequest)
	tradeRequestGroup.PUT("/:id", h.UpdateTradeRequest)
	tradeRequestGroup.DELETE("/:id", h.DeleteTradeRequest)
	tradeRequestGroup.POST("/:id/accept", h.AcceptTradeRequest)
	tradeRequestGroup.POST("/:id/reject", h.RejectTradeRequest)

	// Collection routes
	tradeRequestGroup.GET("", h.ListTradeRequests)
	tradeRequestGroup.POST("", h.CreateTradeRequest)

	// Game-specific routes
	tradeRequestGroup.GET("/game/:game_id", h.GetTradeRequestsByGameID)

	// Player-specific routes
	tradeRequestGroup.GET("/player/:player_id", h.GetTradeRequestsByPlayerID)
}

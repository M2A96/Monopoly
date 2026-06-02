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

func WithTradeRequestHandlerTradeRequestServicer(
	serviceTradeRequestServicer input.TradeRequestServicer,
) tradeRequestHandlerOptioner {
	return WithHandlerServicer[input.TradeRequestServicer](serviceTradeRequestServicer)
}

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

	newTradeRequestID, err := h.GetServicer().CreateTradeRequest(ctxWT, tradeRequest)
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

	base62ID := util.UUIDToBase62(newTradeRequestID)

	h.GetRuntimeLogger().
		WithFields(fields).
		WithField(object.URIFieldResponse, newTradeRequestID).
		WithField("base62_id", base62ID).
		Debug(object.URIEmpty)

	return c.JSON(http.StatusCreated, map[string]string{
		"id": base62ID,
	})
}

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

	idStr := c.Param("id")
	var id uuid.UUID
	var err error

	id, err = h.GetUUIDer().Parse(idStr)
	if err != nil {
		id, err = util.Base62ToUUID(idStr)
		if err != nil {
			h.GetRuntimeLogger().
				WithFields(fields).
				WithField(object.URIFieldError, err).
				Error("Invalid trade request ID format")
			traceSpan.RecordError(err)
			traceSpan.SetStatus(codes.Error, object.ErrUUIDerParse.Error())

			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "Invalid trade request ID format",
			})
		}
	}

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

	daoTradeFilter := dao.NewTradeFilter(
		[]uuid.UUID{},
		uuid.Nil,
		uuid.Nil,
		"",
		uuid.Nil,
	)

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

	daoCursor := dao.NewCursor(0)
	pageToken := c.QueryParam("page_token")

	if pageToken != object.URIEmpty {
		cursorBytes, errDecode := base64.URLEncoding.DecodeString(pageToken)
		if errDecode != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid page token"})
		}

		if err := gob.NewDecoder(bytes.NewBuffer(cursorBytes)).Decode(daoCursor); err != nil {
			return err
		}
	}

	pageSize, err := strconv.Atoi(c.QueryParam("page_size"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid page size"})
	}

	daoPagination := dao.NewPagination(daoCursor, uint32(pageSize))

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

	daoCursor := dao.NewCursor(0)
	pageToken := c.QueryParam("page_token")

	if pageToken != object.URIEmpty {
		cursorBytes, errDecode := base64.URLEncoding.DecodeString(pageToken)
		if errDecode != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid page token"})
		}

		if err := gob.NewDecoder(bytes.NewBuffer(cursorBytes)).Decode(daoCursor); err != nil {
			return err
		}
	}

	pageSize, err := strconv.Atoi(c.QueryParam("page_size"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid page size"})
	}

	daoPagination := dao.NewPagination(daoCursor, uint32(pageSize))

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

func (h *tradeRequestHandler) RegisterRoutes(e *echo.Echo) {
	g := e.Group("/api/v1/trade-requests")
	g.POST("", h.CreateTradeRequest)
	g.GET("", h.ListTradeRequests)
	g.GET("/:id", h.GetTradeRequest)

	e.GET("/api/v1/games/:game_id/trade-requests", h.GetTradeRequestsByGameID)
	e.GET("/api/v1/players/:player_id/trade-requests", h.GetTradeRequestsByPlayerID)
}

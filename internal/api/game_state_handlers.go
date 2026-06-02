package api

import (
	"context"
	"net/http"
	"strconv"

	"github/M2A96/Monopoly.git/config"
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
	GameStateHandler interface {
		GetCurrentState(c echo.Context) error
		GetGameStateHistory(c echo.Context) error
	}

	gameStateHandler struct {
		handler[input.GameStateServicer]
	}

	gameStateHandlerOptioner = handlerOptioner[input.GameStateServicer]
)

var (
	_ Handler          = (*gameStateHandler)(nil)
	_ GameStateHandler = (*gameStateHandler)(nil)
)

func NewGameStateHandler(
	configConfigger config.Configger,
	logRuntimeLogger log.RuntimeLogger,
	objectUUIDer object.UUIDer,
	traceTracer trace.Tracer,
	optioner ...gameStateHandlerOptioner,
) *gameStateHandler {
	return &gameStateHandler{
		handler: *NewHandler[input.GameStateServicer](
			configConfigger,
			logRuntimeLogger,
			objectUUIDer,
			traceTracer,
			optioner...,
		),
	}
}

func WithGameStateHandlerGameStateServicer(
	serviceGameStateServicer input.GameStateServicer,
) gameStateHandlerOptioner {
	return WithHandlerServicer[input.GameStateServicer](serviceGameStateServicer)
}

func (h *gameStateHandler) GetCurrentState(c echo.Context) error {
	ctx := c.Request().Context()
	ctxWT, ctxWTCancelFunc := context.WithTimeout(ctx, object.NUMHTTPServerTimeout)
	defer ctxWTCancelFunc()

	var traceSpan trace.Span
	ctxWT, traceSpan = h.GetTracer().Start(
		ctxWT,
		"GetCurrentState",
		trace.WithSpanKind(trace.SpanKindServer),
	)
	defer traceSpan.End()

	utilRuntimeContext := util.NewRuntimeContext(ctxWT)
	utilSpanContext := util.NewSpanContext(traceSpan)
	fields := map[string]any{
		"name":   "GetCurrentState",
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
				Error("Invalid game ID format")
			traceSpan.RecordError(err)
			traceSpan.SetStatus(codes.Error, object.ErrUUIDerParse.Error())

			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "Invalid game ID format",
			})
		}
	}

	gameState, err := h.GetServicer().GetCurrentState(ctxWT, id)
	if err != nil {
		h.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error("Failed to get current game state")
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, "Failed to get current game state")

		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to get current game state",
		})
	}

	return c.JSON(http.StatusOK, gameState)
}

func (h *gameStateHandler) GetGameStateHistory(c echo.Context) error {
	ctx := c.Request().Context()
	ctxWT, ctxWTCancelFunc := context.WithTimeout(ctx, object.NUMHTTPServerTimeout)
	defer ctxWTCancelFunc()

	var traceSpan trace.Span
	ctxWT, traceSpan = h.GetTracer().Start(
		ctxWT,
		"GetGameStateHistory",
		trace.WithSpanKind(trace.SpanKindServer),
	)
	defer traceSpan.End()

	utilRuntimeContext := util.NewRuntimeContext(ctxWT)
	utilSpanContext := util.NewSpanContext(traceSpan)
	fields := map[string]any{
		"name":   "GetGameStateHistory",
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
				Error("Invalid game ID format")
			traceSpan.RecordError(err)
			traceSpan.SetStatus(codes.Error, object.ErrUUIDerParse.Error())

			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "Invalid game ID format",
			})
		}
	}

	offset := 0
	if cursorParam := c.QueryParam("cursor"); cursorParam != "" {
		if v, err := strconv.Atoi(cursorParam); err == nil && v >= 0 {
			offset = v
		}
	}

	limit := 10
	if limitParam := c.QueryParam("limit"); limitParam != "" {
		if v, err := strconv.Atoi(limitParam); err == nil && v > 0 {
			limit = v
		}
	}

	daoCursor := dao.NewCursor(uint32(offset))
	pagination := dao.NewPagination(daoCursor, uint32(limit))

	filter := dao.NewGameStateFilter(id, nil, nil)
	gameStates, nextCursor, err := h.GetServicer().GetGameStateHistory(ctxWT, filter, pagination)
	if err != nil {
		h.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error("Failed to get game state history")
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, "Failed to get game state history")

		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to get game state history",
		})
	}

	response := map[string]interface{}{
		"game_states": gameStates,
		"pagination": map[string]interface{}{
			"limit":  limit,
			"cursor": daoCursor,
		},
	}

	if nextCursor != nil {
		response["pagination"].(map[string]interface{})["next_cursor"] = nextCursor
	}

	return c.JSON(http.StatusOK, response)
}

func (h *gameStateHandler) RegisterRoutes(e *echo.Echo) {
	e.GET("/api/v1/games/:id/state/current", h.GetCurrentState)
	e.GET("/api/v1/games/:id/states", h.GetGameStateHistory)
}

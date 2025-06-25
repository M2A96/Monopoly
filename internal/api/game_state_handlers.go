// internal/api/game_state_handlers.go
package api

import (
	"context"
	"net/http"

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
	GameStateHandler interface {
		GetCurrentState(c echo.Context) error
		SaveGameState(c echo.Context) error
		GetGameStateHistory(c echo.Context) error
		RestoreGameState(c echo.Context) error
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

// NewGameStateHandler creates a new game state handler instance
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

// WithGameStateHandlerGameStateServicer sets the game state servicer for the handler
func WithGameStateHandlerGameStateServicer(
	serviceGameStateServicer input.GameStateServicer,
) gameStateHandlerOptioner {
	return WithHandlerServicer[input.GameStateServicer](serviceGameStateServicer)
}

// GetCurrentState retrieves the current state of a game
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
				Error("Invalid game ID format - not a valid UUID or base62 ID")
			traceSpan.RecordError(err)
			traceSpan.SetStatus(codes.Error, object.ErrUUIDerParse.Error())

			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "Invalid game ID format",
			})
		}
	}

	h.GetRuntimeLogger().
		WithFields(fields).
		WithField(object.URIFieldID, id).
		Debug(object.URIEmpty)

	// Call service to get current game state
	gameState, err := h.GetServicer().GetCurrentState(ctxWT, id)
	if err != nil {
		h.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error("Failed to get current game state")
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, "Failed to get current game state")

		statusCode := http.StatusInternalServerError
		if err == object.ErrNotFound {
			statusCode = http.StatusNotFound
		}

		return c.JSON(statusCode, map[string]string{
			"error": "Failed to get current game state",
		})
	}

	h.GetRuntimeLogger().
		WithFields(fields).
		WithField(object.URIFieldResponse, gameState).
		Debug(object.URIEmpty)

	return c.JSON(http.StatusOK, gameState)
}

// SaveGameState saves the current state of a game
func (h *gameStateHandler) SaveGameState(c echo.Context) error {
	ctx := c.Request().Context()
	ctxWT, ctxWTCancelFunc := context.WithTimeout(ctx, object.NUMHTTPServerTimeout)
	defer ctxWTCancelFunc()

	var traceSpan trace.Span
	ctxWT, traceSpan = h.GetTracer().Start(
		ctxWT,
		"SaveGameState",
		trace.WithSpanKind(trace.SpanKindServer),
	)
	defer traceSpan.End()

	utilRuntimeContext := util.NewRuntimeContext(ctxWT)
	utilSpanContext := util.NewSpanContext(traceSpan)
	fields := map[string]any{
		"name":   "SaveGameState",
		"rt_ctx": utilRuntimeContext,
		"sp_ctx": utilSpanContext,
		"config": h.GetConfigger(),
	}

	h.GetRuntimeLogger().
		WithFields(fields).
		Info(object.URIEmpty)

	// Parse request body
	var gameState bo.GameStater
	if err := c.Bind(&gameState); err != nil {
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

	// Save game state using service
	err := h.GetServicer().SaveGameState(ctxWT, gameState)
	if err != nil {
		h.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error("Failed to save game state")
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, "Failed to save game state")

		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to save game state",
		})
	}

	h.GetRuntimeLogger().
		WithFields(fields).
		WithField("game_id", gameState.GetGame().GetID()).
		Debug("Game state saved successfully")

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Game state saved successfully",
	})
}

// GetGameStateHistory retrieves the history of game states
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
				Error("Invalid game ID format - not a valid UUID or base62 ID")
			traceSpan.RecordError(err)
			traceSpan.SetStatus(codes.Error, object.ErrUUIDerParse.Error())

			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "Invalid game ID format",
			})
		}
	}

	// Parse pagination parameters
	limit := 10 // Default limit
	if limitParam := c.QueryParam("limit"); limitParam != "" {
		limitInt, err := util.ParseInt(limitParam)
		if err == nil && limitInt > 0 {
			limit = limitInt
		}
	}

	cursor := c.QueryParam("cursor")
	pagination := dao.NewPagination(limit, cursor)

	h.GetRuntimeLogger().
		WithFields(fields).
		WithField(object.URIFieldID, id).
		WithField("pagination", pagination).
		Debug(object.URIEmpty)

	// Call service to get game state history
	gameStates, nextCursor, err := h.GetServicer().GetGameStateHistory(ctxWT, id, pagination)
	if err != nil {
		h.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error("Failed to get game state history")
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, "Failed to get game state history")

		statusCode := http.StatusInternalServerError
		if err == object.ErrNotFound {
			statusCode = http.StatusNotFound
		}

		return c.JSON(statusCode, map[string]string{
			"error": "Failed to get game state history",
		})
	}

	// Prepare response with pagination info
	response := map[string]interface{}{
		"game_states": gameStates,
		"pagination": map[string]interface{}{
			"limit": limit,
		},
	}

	// Add next cursor if available
	if nextCursor != nil && nextCursor.GetCursor() != "" {
		response["pagination"].(map[string]interface{})["next_cursor"] = nextCursor.GetCursor()
	}

	h.GetRuntimeLogger().
		WithFields(fields).
		WithField(object.URIFieldResponse, response).
		Debug(object.URIEmpty)

	return c.JSON(http.StatusOK, response)
}

// RestoreGameState restores a game to a previous state
func (h *gameStateHandler) RestoreGameState(c echo.Context) error {
	ctx := c.Request().Context()
	ctxWT, ctxWTCancelFunc := context.WithTimeout(ctx, object.NUMHTTPServerTimeout)
	defer ctxWTCancelFunc()

	var traceSpan trace.Span
	ctxWT, traceSpan = h.GetTracer().Start(
		ctxWT,
		"RestoreGameState",
		trace.WithSpanKind(trace.SpanKindServer),
	)
	defer traceSpan.End()

	utilRuntimeContext := util.NewRuntimeContext(ctxWT)
	utilSpanContext := util.NewSpanContext(traceSpan)
	fields := map[string]any{
		"name":   "RestoreGameState",
		"rt_ctx": utilRuntimeContext,
		"sp_ctx": utilSpanContext,
		"config": h.GetConfigger(),
	}

	h.GetRuntimeLogger().
		WithFields(fields).
		Info(object.URIEmpty)

	// Parse game ID from path parameter
	gameIDStr := c.Param("id")
	var gameID uuid.UUID
	var err error

	// Try parsing as standard UUID first
	gameID, err = h.GetUUIDer().Parse(gameIDStr)
	if err != nil {
		// If not a standard UUID, try parsing as base62
		gameID, err = util.Base62ToUUID(gameIDStr)
		if err != nil {
			h.GetRuntimeLogger().
				WithFields(fields).
				WithField(object.URIFieldError, err).
				Error("Invalid game ID format - not a valid UUID or base62 ID")
			traceSpan.RecordError(err)
			traceSpan.SetStatus(codes.Error, object.ErrUUIDerParse.Error())

			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "Invalid game ID format",
			})
		}
	}

	// Parse state ID from path parameter
	stateIDStr := c.Param("state_id")
	var stateID uuid.UUID

	// Try parsing as standard UUID first
	stateID, err = h.GetUUIDer().Parse(stateIDStr)
	if err != nil {
		// If not a standard UUID, try parsing as base62
		stateID, err = util.Base62ToUUID(stateIDStr)
		if err != nil {
			h.GetRuntimeLogger().
				WithFields(fields).
				WithField(object.URIFieldError, err).
				Error("Invalid state ID format - not a valid UUID or base62 ID")
			traceSpan.RecordError(err)
			traceSpan.SetStatus(codes.Error, object.ErrUUIDerParse.Error())

			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "Invalid state ID format",
			})
		}
	}

	h.GetRuntimeLogger().
		WithFields(fields).
		WithField("game_id", gameID).
		WithField("state_id", stateID).
		Debug(object.URIEmpty)

	// Call service to restore game state
	err = h.GetServicer().RestoreGameState(ctxWT, gameID, stateID)
	if err != nil {
		h.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error("Failed to restore game state")
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, "Failed to restore game state")

		statusCode := http.StatusInternalServerError
		if err == object.ErrNotFound {
			statusCode = http.StatusNotFound
		}

		return c.JSON(statusCode, map[string]string{
			"error": "Failed to restore game state",
		})
	}

	h.GetRuntimeLogger().
		WithFields(fields).
		WithField("game_id", gameID).
		WithField("state_id", stateID).
		Debug("Game state restored successfully")

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Game state restored successfully",
	})
}

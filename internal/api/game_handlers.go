// internal/api/game_handlers.go
package api

import (
	"context"
	"net/http"

	"github/M2A96/Monopoly.git/config"
	"github/M2A96/Monopoly.git/internal/core/domain/bo"
	"github/M2A96/Monopoly.git/internal/core/ports/input"
	"github/M2A96/Monopoly.git/log"
	"github/M2A96/Monopoly.git/object"
	"github/M2A96/Monopoly.git/util"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type (
	GameHandler interface {
		GetGame(c echo.Context) error
		StartGame(c echo.Context) error
		GetGameState(c echo.Context) error
		CreateGame(c echo.Context) error
	}
	gameHandler struct {
		handler[input.GameServicer]
	}

	gameHandlerOptioner = handlerOptioner[input.GameServicer]
)

var (
	_ Handler     = (*gameHandler)(nil)
	_ GameHandler = (*gameHandler)(nil)
)

// NewGameHandler is a function.
func NewGameHandler(
	configConfigger config.Configger,
	logRuntimeLogger log.RuntimeLogger,
	objectUUIDer object.UUIDer,
	traceTracer trace.Tracer,
	optioner ...gameHandlerOptioner,
) *gameHandler {
	return &gameHandler{
		handler: *NewHandler[input.GameServicer](
			configConfigger,
			logRuntimeLogger,
			objectUUIDer,
			traceTracer,
			optioner...,
		),
	}
}

// WithGameHandlerGameServicer is a function.
func WithGameHandlerGameServicer(
	serviceGameServicer input.GameServicer,
) gameHandlerOptioner {
	return WithHandlerServicer[input.GameServicer](serviceGameServicer)
}

// CreateGame implements GameHandler.
func (h *gameHandler) CreateGame(c echo.Context) error {
	ctx := c.Request().Context()
	ctxWT, ctxWTCancelFunc := context.WithTimeout(ctx, object.NUMHTTPServerTimeout)
	defer ctxWTCancelFunc()

	var traceSpan trace.Span
	ctxWT, traceSpan = h.GetTracer().Start(
		ctxWT,
		"CreateGame",
		trace.WithSpanKind(trace.SpanKindServer),
	)
	defer traceSpan.End()

	utilRuntimeContext := util.NewRuntimeContext(ctxWT)
	utilSpanContext := util.NewSpanContext(traceSpan)
	fields := map[string]any{
		"name":   "CreateGame",
		"rt_ctx": utilRuntimeContext,
		"sp_ctx": utilSpanContext,
		"config": h.GetConfigger(),
	}

	h.GetRuntimeLogger().
		WithFields(fields).
		Info(object.URIEmpty)

	// Parse request body
	var game bo.Gamer
	if err := c.Bind(&game); err != nil {
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

	// Create game using service
	newGameId, err := h.GetServicer().CreateGame(ctxWT, game)
	if err != nil {
		h.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error(object.ErrGameServiceCreate.Error())
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, object.ErrGameServiceCreate.Error())

		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to create game",
		})
	}

	h.GetRuntimeLogger().
		WithFields(fields).
		WithField(object.URIFieldResponse, newGameId).
		Debug(object.URIEmpty)

	return c.JSON(http.StatusCreated, newGameId)
}

// GetGame implements GameHandler.
func (h *gameHandler) GetGame(c echo.Context) error {
	ctx := c.Request().Context()
	ctxWT, ctxWTCancelFunc := context.WithTimeout(ctx, object.NUMHTTPServerTimeout)
	defer ctxWTCancelFunc()

	var traceSpan trace.Span
	ctxWT, traceSpan = h.GetTracer().Start(
		ctxWT,
		"GetGame",
		trace.WithSpanKind(trace.SpanKindServer),
	)
	defer traceSpan.End()

	utilRuntimeContext := util.NewRuntimeContext(ctxWT)
	utilSpanContext := util.NewSpanContext(traceSpan)
	fields := map[string]any{
		"name":   "GetGame",
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
			Error(object.ErrUUIDerParse.Error())
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, object.ErrUUIDerParse.Error())

		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid game ID format",
		})
	}

	h.GetRuntimeLogger().
		WithFields(fields).
		WithField(object.URIFieldID, id).
		Debug(object.URIEmpty)

	// Call service to get game
	game, err := h.GetServicer().Get(ctxWT, id)
	if err != nil {
		h.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error(object.ErrGameServiceGet.Error())
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, object.ErrGameServiceGet.Error())

		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to retrieve game",
		})
	}

	// }
	newGame := bo.NewGamer(
		game.GetID(),
		game.GetName(),
		game.GetStatus(),
		game.GetCurrentPlayerID(),
		game.GetWinnerID(),
	)

	h.GetRuntimeLogger().
		WithFields(fields).
		WithField("response", newGame).
		Debug(object.URIEmpty)

	return c.JSON(http.StatusOK, newGame.GetMap())
}

// GetGameState implements GameHandler.
func (h *gameHandler) GetGameState(c echo.Context) error {
	ctx := c.Request().Context()
	ctxWT, ctxWTCancelFunc := context.WithTimeout(ctx, object.NUMHTTPServerTimeout)
	defer ctxWTCancelFunc()

	var traceSpan trace.Span
	ctxWT, traceSpan = h.GetTracer().Start(
		ctxWT,
		"GetGameState",
		trace.WithSpanKind(trace.SpanKindServer),
	)
	defer traceSpan.End()

	utilRuntimeContext := util.NewRuntimeContext(ctxWT)
	utilSpanContext := util.NewSpanContext(traceSpan)
	fields := map[string]any{
		"name":   "GetGameState",
		"rt_ctx": utilRuntimeContext,
		"sp_ctx": utilSpanContext,
		"config": h.GetConfigger(),
	}

	h.GetRuntimeLogger().
		WithFields(fields).
		Info(object.URIEmpty)

	// Parse game ID
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error(object.ErrUUIDerParse.Error())
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, object.ErrUUIDerParse.Error())

		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid game ID format",
		})
	}

	// Get game state using service
	gameState, err := h.GetServicer().GetGameState(ctxWT, id)
	if err != nil {
		h.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error(object.ErrGameServiceGetState.Error())
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, object.ErrGameServiceGetState.Error())

		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to retrieve game state",
		})
	}

	boGame := bo.NewGamer(
		gameState.GetID(),
		gameState.GetName(),
		gameState.GetStatus(),
		gameState.GetCurrentPlayerID(),
		gameState.GetWinnerID(),
	)

	h.GetRuntimeLogger().
		WithFields(fields).
		WithField(object.URIFieldResponse, boGame).
		Debug(object.URIEmpty)

	return c.JSON(http.StatusOK, boGame.GetStatus())
}

// StartGame implements GameHandler.
func (h *gameHandler) StartGame(c echo.Context) error {
	ctx := c.Request().Context()
	ctxWT, ctxWTCancelFunc := context.WithTimeout(ctx, object.NUMHTTPServerTimeout)
	defer ctxWTCancelFunc()

	var traceSpan trace.Span
	ctxWT, traceSpan = h.GetTracer().Start(
		ctxWT,
		"StartGame",
		trace.WithSpanKind(trace.SpanKindServer),
	)
	defer traceSpan.End()

	utilRuntimeContext := util.NewRuntimeContext(ctxWT)
	utilSpanContext := util.NewSpanContext(traceSpan)
	fields := map[string]any{
		"name":   "StartGame",
		"rt_ctx": utilRuntimeContext,
		"sp_ctx": utilSpanContext,
		"config": h.GetConfigger(),
	}

	h.GetRuntimeLogger().
		WithFields(fields).
		Info(object.URIEmpty)

	// Parse game ID
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
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

	// Start game using service
	err = h.GetServicer().StartGame(ctxWT, id)
	if err != nil {
		h.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error(object.ErrGameServiceStart.Error())
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, object.ErrGameServiceStart.Error())

		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to start game",
		})
	}

	h.GetRuntimeLogger().
		WithFields(fields).
		Debug(object.URIEmpty)

	return c.JSON(http.StatusOK, map[string]string{
		"status": "Game started successfully",
	})
}

// RegisterRoutes implements Handler.
func (h *gameHandler) RegisterRoutes(e *echo.Echo) {
	e.GET("/api/v1/games/:id", h.GetGame)
	e.POST("/api/v1/games/:id/start", h.StartGame)
	e.GET("/api/v1/games/:id/state", h.GetGameState)
	e.POST("/api/v1/games", h.CreateGame)
}

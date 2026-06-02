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
	PlayerHandler interface {
		GetPlayer(c echo.Context) error
		ListPlayers(c echo.Context) error
		CreatePlayer(c echo.Context) error
	}

	playerHandler struct {
		handler[input.PlayerServicer]
	}

	playerHandlerOptioner = handlerOptioner[input.PlayerServicer]
)

var (
	_ Handler       = (*playerHandler)(nil)
	_ PlayerHandler = (*playerHandler)(nil)
)

func NewPlayerHandler(
	configConfigger config.Configger,
	logRuntimeLogger log.RuntimeLogger,
	objectUUIDer object.UUIDer,
	traceTracer trace.Tracer,
	optioner ...playerHandlerOptioner,
) *playerHandler {
	return &playerHandler{
		handler: *NewHandler[input.PlayerServicer](
			configConfigger,
			logRuntimeLogger,
			objectUUIDer,
			traceTracer,
			optioner...,
		),
	}
}

func WithPlayerHandlerPlayerServicer(
	servicePlayerServicer input.PlayerServicer,
) playerHandlerOptioner {
	return WithHandlerServicer[input.PlayerServicer](servicePlayerServicer)
}

func (h *playerHandler) CreatePlayer(c echo.Context) error {
	ctx := c.Request().Context()
	ctxWT, ctxWTCancelFunc := context.WithTimeout(ctx, object.NUMHTTPServerTimeout)
	defer ctxWTCancelFunc()

	var traceSpan trace.Span
	ctxWT, traceSpan = h.GetTracer().Start(
		ctxWT,
		"CreatePlayer",
		trace.WithSpanKind(trace.SpanKindServer),
	)
	defer traceSpan.End()

	utilRuntimeContext := util.NewRuntimeContext(ctxWT)
	utilSpanContext := util.NewSpanContext(traceSpan)
	fields := map[string]any{
		"name":   "CreatePlayer",
		"rt_ctx": utilRuntimeContext,
		"sp_ctx": utilSpanContext,
		"config": h.GetConfigger(),
	}

	h.GetRuntimeLogger().
		WithFields(fields).
		Info(object.URIEmpty)

	var player bo.Player
	if err := c.Bind(&player); err != nil {
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

	newPlayerID, err := h.GetServicer().CreatePlayer(ctxWT, player)
	if err != nil {
		h.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error("Failed to create player")
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, "Failed to create player")

		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to create player",
		})
	}

	h.GetRuntimeLogger().
		WithFields(fields).
		WithField(object.URIFieldResponse, newPlayerID).
		Debug(object.URIEmpty)

	return c.JSON(http.StatusCreated, map[string]string{
		"id": newPlayerID.String(),
	})
}

func (h *playerHandler) GetPlayer(c echo.Context) error {
	ctx := c.Request().Context()
	ctxWT, ctxWTCancelFunc := context.WithTimeout(ctx, object.NUMHTTPServerTimeout)
	defer ctxWTCancelFunc()

	var traceSpan trace.Span
	ctxWT, traceSpan = h.GetTracer().Start(
		ctxWT,
		"GetPlayer",
		trace.WithSpanKind(trace.SpanKindServer),
	)
	defer traceSpan.End()

	utilRuntimeContext := util.NewRuntimeContext(ctxWT)
	utilSpanContext := util.NewSpanContext(traceSpan)
	fields := map[string]any{
		"name":   "GetPlayer",
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
			Error("Invalid player ID format")
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, "Invalid player ID format")

		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid player ID format",
		})
	}

	h.GetRuntimeLogger().
		WithFields(fields).
		WithField(object.URIFieldID, id).
		Debug(object.URIEmpty)

	player, err := h.GetServicer().Get(ctxWT, id)
	if err != nil {
		h.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error("Failed to retrieve player")
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, "Failed to retrieve player")

		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "Player not found",
		})
	}

	return c.JSON(http.StatusOK, player)
}

func (h *playerHandler) ListPlayers(c echo.Context) error {
	ctx := c.Request().Context()
	ctxWT, ctxWTCancelFunc := context.WithTimeout(ctx, object.NUMHTTPServerTimeout)
	defer ctxWTCancelFunc()

	var traceSpan trace.Span
	ctxWT, traceSpan = h.GetTracer().Start(
		ctxWT,
		"ListPlayers",
		trace.WithSpanKind(trace.SpanKindServer),
	)
	defer traceSpan.End()

	utilRuntimeContext := util.NewRuntimeContext(ctxWT)
	utilSpanContext := util.NewSpanContext(traceSpan)
	fields := map[string]any{
		"name":   "ListPlayers",
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

	gameIDStr := c.Param("game_id")
	gameID, err := h.GetUUIDer().Parse(gameIDStr)
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

	daoPlayerFilter := dao.NewPlayerFilter(
		[]uuid.UUID{},
		"",
		gameID,
		0,
		0,
		false,
		0,
		false,
	)

	players, cursor, err := h.GetServicer().List(
		ctxWT,
		daoPagination,
		daoPlayerFilter,
	)
	if err != nil {
		h.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error("Failed to list players")
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, "Failed to list players")

		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to list players",
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"players": players,
		"cursor":  cursor,
	})
}

func (h *playerHandler) RegisterRoutes(e *echo.Echo) {
	e.POST("/api/v1/players", h.CreatePlayer)
	e.GET("/api/v1/players/:id", h.GetPlayer)
	e.GET("/api/v1/games/:game_id/players", h.ListPlayers)
}

// internal/api/player_handlers.go
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
		UpdatePlayer(c echo.Context) error
		DeletePlayer(c echo.Context) error
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

// NewPlayerHandler creates a new player handler instance
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

// WithPlayerHandlerPlayerServicer sets the player service for the handler
func WithPlayerHandlerPlayerServicer(
	servicePlayerServicer input.PlayerServicer,
) playerHandlerOptioner {
	return WithHandlerServicer[input.PlayerServicer](servicePlayerServicer)
}

// CreatePlayer implements PlayerHandler
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

	// Parse request body
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

	// Create player using service
	newPlayerId, err := h.GetServicer().CreatePlayer(ctxWT, player)
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
		WithField(object.URIFieldResponse, newPlayerId).
		Debug(object.URIEmpty)

	return c.JSON(http.StatusCreated, newPlayerId)
}

// GetPlayer implements PlayerHandler
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

	// Parse UUID from path parameter
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

	// Get player using service
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

// ListPlayers implements PlayerHandler
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
		h.GetRuntimeLogger().
			WithFields(fields).
			Debug(`apiv1GameServiceListRequest.GetPageToken() != object.URIEmpty`)

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
				"error": "Invalid page size",
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

	daoPlyerFilter := dao.NewPlayerFilter(
		[]uuid.UUID{},
		"",
		gameID,
		0,
		0,
		false,
		0,
		false,
	)

	h.GetRuntimeLogger().
		WithFields(fields).
		WithField(object.URIFieldDAOPlayerFilter, daoPlyerFilter).
		Debug(object.URIEmpty)

	players, cursor, err := h.GetServicer().List(
		ctxWT,
		daoPagination,
		daoPlyerFilter,
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

// UpdatePlayer implements PlayerHandler
func (h *playerHandler) UpdatePlayer(c echo.Context) error {
	ctx := c.Request().Context()
	ctxWT, ctxWTCancelFunc := context.WithTimeout(ctx, object.NUMHTTPServerTimeout)
	defer ctxWTCancelFunc()

	var traceSpan trace.Span
	ctxWT, traceSpan = h.GetTracer().Start(
		ctxWT,
		"UpdatePlayer",
		trace.WithSpanKind(trace.SpanKindServer),
	)
	defer traceSpan.End()

	utilRuntimeContext := util.NewRuntimeContext(ctxWT)
	utilSpanContext := util.NewSpanContext(traceSpan)
	fields := map[string]any{
		"name":   "UpdatePlayer",
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
			Error("Invalid player ID format")
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, "Invalid player ID format")

		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid player ID format",
		})
	}

	// Parse request body
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

	// Update player using service
	if err := h.GetServicer().UpdatePlayer(ctxWT, id, player); err != nil {
		h.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error("Failed to update player")
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, "Failed to update player")

		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to update player",
		})
	}

	return c.NoContent(http.StatusOK)
}

// DeletePlayer implements PlayerHandler
func (h *playerHandler) DeletePlayer(c echo.Context) error {
	ctx := c.Request().Context()
	ctxWT, ctxWTCancelFunc := context.WithTimeout(ctx, object.NUMHTTPServerTimeout)
	defer ctxWTCancelFunc()

	var traceSpan trace.Span
	ctxWT, traceSpan = h.GetTracer().Start(
		ctxWT,
		"DeletePlayer",
		trace.WithSpanKind(trace.SpanKindServer),
	)
	defer traceSpan.End()

	utilRuntimeContext := util.NewRuntimeContext(ctxWT)
	utilSpanContext := util.NewSpanContext(traceSpan)
	fields := map[string]any{
		"name":   "DeletePlayer",
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
			Error("Invalid player ID format")
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, "Invalid player ID format")

		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid player ID format",
		})
	}

	// Delete player using service
	if err := h.GetServicer().DeletePlayer(
		ctxWT,
		id,
	); err != nil {
		h.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error(object.ErrPlayerHandlerDelete.Error())
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, object.ErrPlayerHandlerDelete.Error())

		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to delete player",
		})
	}

	return c.NoContent(http.StatusOK)
}

// RegisterRoutes implements Handler.
func (h *playerHandler) RegisterRoutes(e *echo.Echo) {
	e.POST("/players", h.CreatePlayer)
	e.GET("/players/:id", h.GetPlayer)
	e.GET("/players", h.ListPlayers)
	e.PUT("/players/:id", h.UpdatePlayer)
	e.DELETE("/players/:id", h.DeletePlayer)
}

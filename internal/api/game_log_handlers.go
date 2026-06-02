// internal/api/game_log_handlers.go
package api

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/gob"
	"net/http"
	"strconv"
	"time"

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
	// GameLogHandler defines the interface for game log operations
	GameLogHandler interface {
		GetGameLog(c echo.Context) error
		ListGameLogs(c echo.Context) error
		CreateGameLog(c echo.Context) error
		GetGameLogsByGameID(c echo.Context) error
		GetGameLogsByPlayerID(c echo.Context) error
		GetGameLogsByTimeRange(c echo.Context) error
	}

	gameLogHandler struct {
		handler[input.GameLogServicer]
	}

	gameLogHandlerOptioner = handlerOptioner[input.GameLogServicer]
)

var (
	_ Handler        = (*gameLogHandler)(nil)
	_ GameLogHandler = (*gameLogHandler)(nil)
)

// NewGameLogHandler creates a new game log handler instance
func NewGameLogHandler(
	configConfigger config.Configger,
	logRuntimeLogger log.RuntimeLogger,
	objectUUIDer object.UUIDer,
	traceTracer trace.Tracer,
	optioner ...gameLogHandlerOptioner,
) *gameLogHandler {
	return &gameLogHandler{
		handler: *NewHandler[input.GameLogServicer](
			configConfigger,
			logRuntimeLogger,
			objectUUIDer,
			traceTracer,
			optioner...,
		),
	}
}

// WithGameLogHandlerGameLogServicer sets the game log service for the handler
func WithGameLogHandlerGameLogServicer(
	serviceGameLogServicer input.GameLogServicer,
) gameLogHandlerOptioner {
	return WithHandlerServicer[input.GameLogServicer](serviceGameLogServicer)
}

// CreateGameLog implements GameLogHandler
func (h *gameLogHandler) CreateGameLog(c echo.Context) error {
	ctx := c.Request().Context()
	ctxWT, ctxWTCancelFunc := context.WithTimeout(ctx, object.NUMHTTPServerTimeout)
	defer ctxWTCancelFunc()

	var traceSpan trace.Span
	ctxWT, traceSpan = h.GetTracer().Start(
		ctxWT,
		"CreateGameLog",
		trace.WithSpanKind(trace.SpanKindServer),
	)
	defer traceSpan.End()

	utilRuntimeContext := util.NewRuntimeContext(ctxWT)
	utilSpanContext := util.NewSpanContext(traceSpan)
	fields := map[string]any{
		"name":   "CreateGameLog",
		"rt_ctx": utilRuntimeContext,
		"sp_ctx": utilSpanContext,
		"config": h.GetConfigger(),
	}

	h.GetRuntimeLogger().
		WithFields(fields).
		Info(object.URIEmpty)

	// Parse request body
	var gameLog bo.GameLogger
	if err := c.Bind(&gameLog); err != nil {
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

	// Create game log using service
	newGameLogId, err := h.GetServicer().CreateGameLog(ctxWT, gameLog)
	if err != nil {
		h.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error("Failed to create game log")
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, "Failed to create game log")

		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to create game log",
		})
	}

	// Convert UUID to base62 for shorter representation
	base62Id := util.UUIDToBase62(newGameLogId)

	h.GetRuntimeLogger().
		WithFields(fields).
		WithField(object.URIFieldResponse, newGameLogId).
		WithField("base62_id", base62Id).
		Debug(object.URIEmpty)

	return c.JSON(http.StatusCreated, map[string]string{
		"id": base62Id,
	})
}

// GetGameLog implements GameLogHandler
func (h *gameLogHandler) GetGameLog(c echo.Context) error {
	ctx := c.Request().Context()
	ctxWT, ctxWTCancelFunc := context.WithTimeout(ctx, object.NUMHTTPServerTimeout)
	defer ctxWTCancelFunc()

	var traceSpan trace.Span
	ctxWT, traceSpan = h.GetTracer().Start(
		ctxWT,
		"GetGameLog",
		trace.WithSpanKind(trace.SpanKindServer),
	)
	defer traceSpan.End()

	utilRuntimeContext := util.NewRuntimeContext(ctxWT)
	utilSpanContext := util.NewSpanContext(traceSpan)
	fields := map[string]any{
		"name":   "GetGameLog",
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
				Error("Invalid game log ID format - not a valid UUID or base62 ID")
			traceSpan.RecordError(err)
			traceSpan.SetStatus(codes.Error, "Invalid game log ID format")

			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "Invalid game log ID format",
			})
		}
	}

	h.GetRuntimeLogger().
		WithFields(fields).
		WithField(object.URIFieldID, id).
		Debug(object.URIEmpty)

	// Call service to get game log
	gameLog, err := h.GetServicer().Get(ctxWT, id)
	if err != nil {
		h.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error("Failed to retrieve game log")
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, "Failed to retrieve game log")

		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "Game log not found",
		})
	}

	h.GetRuntimeLogger().
		WithFields(fields).
		WithField(object.URIFieldResponse, gameLog).
		Debug(object.URIEmpty)

	return c.JSON(http.StatusOK, gameLog)
}

// ListGameLogs implements GameLogHandler
func (h *gameLogHandler) ListGameLogs(c echo.Context) error {
	ctx := c.Request().Context()
	ctxWT, ctxWTCancelFunc := context.WithTimeout(ctx, object.NUMHTTPServerTimeout)
	defer ctxWTCancelFunc()

	var traceSpan trace.Span
	ctxWT, traceSpan = h.GetTracer().Start(
		ctxWT,
		"ListGameLogs",
		trace.WithSpanKind(trace.SpanKindServer),
	)
	defer traceSpan.End()

	utilRuntimeContext := util.NewRuntimeContext(ctxWT)
	utilSpanContext := util.NewSpanContext(traceSpan)
	fields := map[string]any{
		"name":   "ListGameLogs",
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

	daoGameLogFilter := dao.NewGameLogFilter(
		[]uuid.UUID{},
		gameID,
		uuid.Nil,
		"",
		nil,
		nil,
	)

	h.GetRuntimeLogger().
		WithFields(fields).
		WithField(object.URIFieldDAOPGameLogFilter, daoGameLogFilter).
		Debug(object.URIEmpty)

	// Get game logs
	gameLogs, cursor, err := h.GetServicer().List(
		ctxWT,
		daoPagination,
		daoGameLogFilter,
	)
	if err != nil {
		h.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error("Failed to list game logs")
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, "Failed to list game logs")

		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to list game logs",
		})
	}

	// Prepare response
	response := map[string]interface{}{
		"game_logs": gameLogs,
		"cursor":    cursor,
	}

	h.GetRuntimeLogger().
		WithFields(fields).
		WithField(object.URIFieldResponse, response).
		Debug(object.URIEmpty)

	return c.JSON(http.StatusOK, response)
}

// GetGameLogsByGameID implements GameLogHandler
func (h *gameLogHandler) GetGameLogsByGameID(c echo.Context) error {
	ctx := c.Request().Context()
	ctxWT, ctxWTCancelFunc := context.WithTimeout(ctx, object.NUMHTTPServerTimeout)
	defer ctxWTCancelFunc()

	var traceSpan trace.Span
	ctxWT, traceSpan = h.GetTracer().Start(
		ctxWT,
		"GetGameLogsByGameID",
		trace.WithSpanKind(trace.SpanKindServer),
	)
	defer traceSpan.End()

	utilRuntimeContext := util.NewRuntimeContext(ctxWT)
	utilSpanContext := util.NewSpanContext(traceSpan)
	fields := map[string]any{
		"name":   "GetGameLogsByGameID",
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

	daoGameLogFilter := dao.NewGameLogFilter(
		[]uuid.UUID{},
		gameID,
		uuid.Nil,
		"",
		nil,
		nil,
	)

	h.GetRuntimeLogger().
		WithFields(fields).
		WithField(object.URIFieldDAOPGameLogFilter, daoGameLogFilter).
		Debug(object.URIEmpty)

	// Get game logs by game ID
	gameLogs, cursor, err := h.GetServicer().GetGameLogsByGameID(ctxWT,
		gameID,
		daoPagination,
	)
	if err != nil {
		h.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error("Failed to get game logs by game ID")
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, "Failed to get game logs by game ID")

		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to get game logs by game ID",
		})
	}

	// Prepare response
	response := map[string]interface{}{
		"game_logs": gameLogs,
		"cursor":    cursor,
	}

	h.GetRuntimeLogger().
		WithFields(fields).
		WithField(object.URIFieldResponse, response).
		Debug(object.URIEmpty)

	return c.JSON(http.StatusOK, response)
}

// GetGameLogsByPlayerID implements GameLogHandler
func (h *gameLogHandler) GetGameLogsByPlayerID(c echo.Context) error {
	ctx := c.Request().Context()
	ctxWT, ctxWTCancelFunc := context.WithTimeout(ctx, object.NUMHTTPServerTimeout)
	defer ctxWTCancelFunc()

	var traceSpan trace.Span
	ctxWT, traceSpan = h.GetTracer().Start(
		ctxWT,
		"GetGameLogsByPlayerID",
		trace.WithSpanKind(trace.SpanKindServer),
	)
	defer traceSpan.End()

	utilRuntimeContext := util.NewRuntimeContext(ctxWT)
	utilSpanContext := util.NewSpanContext(traceSpan)
	fields := map[string]any{
		"name":   "GetGameLogsByPlayerID",
		"rt_ctx": utilRuntimeContext,
		"sp_ctx": utilSpanContext,
		"config": h.GetConfigger(),
	}

	h.GetRuntimeLogger().
		WithFields(fields).
		Info(object.URIEmpty)

	// Parse player ID from path parameter
	playerIDStr := c.Param("player_id")
	var playerID uuid.UUID
	var err error

	// Try parsing as standard UUID first
	playerID, err = h.GetUUIDer().Parse(playerIDStr)
	if err != nil {
		// If not a standard UUID, try parsing as base62
		playerID, err = util.Base62ToUUID(playerIDStr)
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
	}
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

	// Get game logs by player ID
	gameLogs, cursor, err := h.GetServicer().GetGameLogsByPlayerID(
		ctxWT,
		playerID,
		daoPagination,
	)
	if err != nil {
		h.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error("Failed to get game logs by player ID")
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, "Failed to get game logs by player ID")

		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to get game logs by player ID",
		})
	}

	// Prepare response
	response := map[string]interface{}{
		"game_logs": gameLogs,
		"cursor":    cursor,
	}

	h.GetRuntimeLogger().
		WithFields(fields).
		WithField(object.URIFieldResponse, response).
		Debug(object.URIEmpty)

	return c.JSON(http.StatusOK, response)
}

// GetGameLogsByTimeRange implements GameLogHandler
func (h *gameLogHandler) GetGameLogsByTimeRange(c echo.Context) error {
	ctx := c.Request().Context()
	ctxWT, ctxWTCancelFunc := context.WithTimeout(ctx, object.NUMHTTPServerTimeout)
	defer ctxWTCancelFunc()

	var traceSpan trace.Span
	ctxWT, traceSpan = h.GetTracer().Start(
		ctxWT,
		"GetGameLogsByTimeRange",
		trace.WithSpanKind(trace.SpanKindServer),
	)
	defer traceSpan.End()

	utilRuntimeContext := util.NewRuntimeContext(ctxWT)
	utilSpanContext := util.NewSpanContext(traceSpan)
	fields := map[string]any{
		"name":   "GetGameLogsByTimeRange",
		"rt_ctx": utilRuntimeContext,
		"sp_ctx": utilSpanContext,
		"config": h.GetConfigger(),
	}

	h.GetRuntimeLogger().
		WithFields(fields).
		Info(object.URIEmpty)

	// Parse time range parameters
	startTimeStr := c.QueryParam("start_time")
	endTimeStr := c.QueryParam("end_time")

	var startTime, endTime time.Time
	var err error

	// Parse start time if provided
	if startTimeStr != "" {
		startTime, err = time.Parse(time.RFC3339, startTimeStr)
		if err != nil {
			h.GetRuntimeLogger().
				WithFields(fields).
				WithField(object.URIFieldError, err).
				Error("Invalid start time format")
			traceSpan.RecordError(err)
			traceSpan.SetStatus(codes.Error, "Invalid start time format")

			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "Invalid start time format. Use RFC3339 format (e.g., 2023-01-01T00:00:00Z)",
			})
		}
	} else {
		// Default to 24 hours ago if not provided
		startTime = time.Now().Add(-24 * time.Hour)
	}

	// Parse end time if provided
	if endTimeStr != "" {
		endTime, err = time.Parse(time.RFC3339, endTimeStr)
		if err != nil {
			h.GetRuntimeLogger().
				WithFields(fields).
				WithField(object.URIFieldError, err).
				Error("Invalid end time format")
			traceSpan.RecordError(err)
			traceSpan.SetStatus(codes.Error, "Invalid end time format")

			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "Invalid end time format. Use RFC3339 format (e.g., 2023-01-01T00:00:00Z)",
			})
		}
	} else {
		// Default to current time if not provided
		endTime = time.Now()
	}

	// Validate time range
	if startTime.After(endTime) {
		h.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error("Start time is after end time")
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, "Start time is after end time")

		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Start time must be before end time",
		})
	}

	// Parse player ID from path parameter
	gameIDStr := c.Param("game_id")
	var gameID uuid.UUID

	// Try parsing as standard UUID first
	gameID, err = h.GetUUIDer().Parse(gameIDStr)
	if err != nil {
		// If not a standard UUID, try parsing as base62
		gameID, err = util.Base62ToUUID(gameIDStr)
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
	}

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

	// Get game logs by time range
	gameLogs, cursor, err := h.GetServicer().GetGameLogsByTimeRange(
		ctxWT,
		gameID,
		startTime,
		endTime,
		daoPagination,
	)
	if err != nil {
		h.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error("Failed to get game logs by time range")
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, "Failed to get game logs by time range")

		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to get game logs by time range",
		})
	}

	// Prepare response
	response := map[string]interface{}{
		"game_logs": gameLogs,
		"cursor":    cursor,
	}

	h.GetRuntimeLogger().
		WithFields(fields).
		WithField(object.URIFieldResponse, response).
		Debug(object.URIEmpty)

	return c.JSON(http.StatusOK, response)
}

// RegisterRoutes implements Handler.
func (h *gameLogHandler) RegisterRoutes(e *echo.Echo) {
	// Register game log routes
	e.POST("/api/v1/game-logs", h.CreateGameLog)
	e.GET("/api/v1/game-logs/:id", h.GetGameLog)
	e.GET("/api/v1/game-logs", h.ListGameLogs)
	e.GET("/api/v1/games/:game_id/logs", h.GetGameLogsByGameID)
	e.GET("/api/v1/players/:player_id/logs", h.GetGameLogsByPlayerID)
	e.GET("/api/v1/game-logs/time-range", h.GetGameLogsByTimeRange)
}

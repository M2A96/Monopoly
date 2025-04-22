// internal/core/domain/services/game_log_service.go
package services

import (
	"context"
	"database/sql"
	"github/M2A96/Monopoly.git/config"
	"github/M2A96/Monopoly.git/internal/core/domain/bo"
	"github/M2A96/Monopoly.git/internal/core/ports/input"
	"github/M2A96/Monopoly.git/internal/core/ports/output"
	"github/M2A96/Monopoly.git/log"
	"github/M2A96/Monopoly.git/object"
	"github/M2A96/Monopoly.git/object/dao"
	"github/M2A96/Monopoly.git/util"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// GameLogService implements the input.GameLogServicer interface
type (
	gameLogService struct {
		configConfigger               config.Configger
		logRuntimeLogger              log.RuntimeLogger
		objectUUIDer                  object.UUIDer
		repositoryGameLogRepositorier output.GameLogRepositorier
		objectTimer                   object.Timer
		traceTracer                   trace.Tracer
	}

	gameLogServiceOptioner interface {
		apply(*gameLogService)
	}

	gameLogServiceOptionerFunc func(*gameLogService)
)

var (
	_ input.GameLogServicer         = (*gameLogService)(nil)
	_ config.GetConfigger           = (*gameLogService)(nil)
	_ log.GetRuntimeLogger          = (*gameLogService)(nil)
	_ object.GetUUIDer              = (*gameLogService)(nil)
	_ output.GetGameLogRepositorier = (*gameLogService)(nil)
	_ object.GetTimer               = (*gameLogService)(nil)
	_ util.GetTracer                = (*gameLogService)(nil)
)

// NewGameLogService creates a new game log service instance
func NewGameLogService(
	configConfigger config.Configger,
	logRuntimeLogger log.RuntimeLogger,
	objectUUIDer object.UUIDer,
	traceTracer trace.Tracer,
	optioners ...gameLogServiceOptioner,
) *gameLogService {
	gameLogService := &gameLogService{
		configConfigger:               configConfigger,
		logRuntimeLogger:              logRuntimeLogger,
		objectUUIDer:                  objectUUIDer,
		repositoryGameLogRepositorier: nil,
		objectTimer:                   nil,
		traceTracer:                   traceTracer,
	}

	return gameLogService.WithOptioners(optioners...)
}

// WithGameLogServiceGameLogRepositorier sets the game log repository for the service
func WithGameLogServiceGameLogRepositorier(
	repositoryGameLogRepositorier output.GameLogRepositorier,
) gameLogServiceOptioner {
	return gameLogServiceOptionerFunc(func(
		service *gameLogService,
	) {
		service.repositoryGameLogRepositorier = repositoryGameLogRepositorier
	})
}

// WithGameLogServiceTimer sets the timer for the service
func WithGameLogServiceTimer(
	objectTimer object.Timer,
) gameLogServiceOptioner {
	return gameLogServiceOptionerFunc(func(
		service *gameLogService,
	) {
		service.objectTimer = objectTimer
	})
}

// GetTracer implements util.GetTracer
func (service *gameLogService) GetTracer() trace.Tracer {
	return service.traceTracer
}

// GetGameLogRepositorier implements output.GetGameLogRepositorier
func (service *gameLogService) GetGameLogRepositorier() output.GameLogRepositorier {
	return service.repositoryGameLogRepositorier
}

// GetUUIDer implements object.GetUUIDer
func (service *gameLogService) GetUUIDer() object.UUIDer {
	return service.objectUUIDer
}

// GetRuntimeLogger implements log.GetRuntimeLogger
func (service *gameLogService) GetRuntimeLogger() log.RuntimeLogger {
	return service.logRuntimeLogger
}

// GetConfigger implements config.GetConfigger
func (service *gameLogService) GetConfigger() config.Configger {
	return service.configConfigger
}

// GetTimer implements object.GetTimer
func (service *gameLogService) GetTimer() object.Timer {
	return service.objectTimer
}

// Get implements input.GameLogServicer
func (service *gameLogService) Get(
	ctx context.Context,
	id uuid.UUID,
) (bo.GameLogger, error) {
	var traceSpan trace.Span

	ctx, traceSpan = service.GetTracer().Start(
		ctx,
		"Get",
		trace.WithSpanKind(trace.SpanKindInternal),
	)
	defer traceSpan.End()

	utilRuntimeContext := util.NewRuntimeContext(ctx)
	utilSpanContext := util.NewSpanContext(traceSpan)
	fields := map[string]any{
		"name":   "Get",
		"rt_ctx": utilRuntimeContext,
		"sp_ctx": utilSpanContext,
		"config": service.GetConfigger(),
		"id":     id,
	}

	service.GetRuntimeLogger().
		WithFields(fields).
		Info(object.URIEmpty)

	cudIDer := dao.NewCUDID(map[string]uuid.UUID{"id": id})

	service.GetRuntimeLogger().
		WithFields(fields).
		WithField(object.URIFieldCUDIDer, cudIDer).
		Debug(object.URIEmpty)

	daoGameLogger, err := service.GetGameLogRepositorier().Read(
		ctx,
		cudIDer,
	)
	if err != nil {
		service.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error("Failed to read game log")
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, "Failed to read game log")

		return nil, err
	}

	service.GetRuntimeLogger().
		WithFields(fields).
		WithField(object.URIFieldDAOGameLogger, daoGameLogger).
		Debug(object.URIEmpty)

	boGameLogger := bo.NewGameLog(
		daoGameLogger.GetCUDIDer().GetID()["id"],
		daoGameLogger.GetGameID(),
		daoGameLogger.GetPlayerID(),
		daoGameLogger.GetAction(),
		daoGameLogger.GetDescription(),
		daoGameLogger.GetTimestamp(),
	)

	service.GetRuntimeLogger().
		WithFields(fields).
		WithField(object.URIFieldBOGameLogger, boGameLogger).
		Debug(object.URIEmpty)

	return boGameLogger, nil
}

// List implements input.GameLogServicer
func (service *gameLogService) List(
	ctx context.Context,
	daoPaginationer dao.Paginationer,
	daoGameLogFilter dao.GameLogFilter,
) ([]bo.GameLogger, dao.Cursorer, error) {
	var traceSpan trace.Span

	ctx, traceSpan = service.GetTracer().Start(
		ctx,
		"List",
		trace.WithSpanKind(trace.SpanKindInternal),
	)
	defer traceSpan.End()

	utilRuntimeContext := util.NewRuntimeContext(ctx)
	utilSpanContext := util.NewSpanContext(traceSpan)
	fields := map[string]any{
		"name":                "List",
		"rt_ctx":              utilRuntimeContext,
		"sp_ctx":              utilSpanContext,
		"config":              service.GetConfigger(),
		"dao_paginationer":    daoPaginationer,
		"dao_game_log_filter": daoGameLogFilter,
	}

	service.GetRuntimeLogger().
		WithFields(fields).
		Info(object.URIEmpty)

	daoGameLoggers, daoCursorer, err := service.GetGameLogRepositorier().
		ReadList(ctx, daoPaginationer, daoGameLogFilter)
	if err != nil {
		service.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error("Failed to read game log list")
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, "Failed to read game log list")

		return nil, nil, err
	}

	service.GetRuntimeLogger().
		WithFields(fields).
		WithField(object.URIFieldDAOGameLoggers, daoGameLoggers).
		WithField(object.URIFieldDAOCursorer, daoCursorer).
		Debug(object.URIEmpty)

	boGameLoggers := make([]bo.GameLogger, 0, len(daoGameLoggers))

	for key, daoGameLogger := range daoGameLoggers {
		service.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldKey, key).
			WithField(object.URIFieldDAOGameLogger, daoGameLogger).
			Debug(object.URIEmpty)

		boGameLoggers = append(boGameLoggers, bo.NewGameLog(
			daoGameLogger.GetCUDIDer().GetID()["id"],
			daoGameLogger.GetGameID(),
			daoGameLogger.GetPlayerID(),
			daoGameLogger.GetAction(),
			daoGameLogger.GetDescription(),
			daoGameLogger.GetTimestamp(),
		))
	}

	service.GetRuntimeLogger().
		WithFields(fields).
		WithField(object.URIFieldBOGameLoggers, boGameLoggers).
		Debug(object.URIEmpty)

	return boGameLoggers, daoCursorer, nil
}

// CreateGameLog implements input.GameLogServicer
func (service *gameLogService) CreateGameLog(
	ctx context.Context,
	boGameLogger bo.GameLogger,
) (uuid.UUID, error) {
	var traceSpan trace.Span

	ctx, traceSpan = service.GetTracer().Start(
		ctx,
		"CreateGameLog",
		trace.WithSpanKind(trace.SpanKindInternal),
	)
	defer traceSpan.End()

	utilRuntimeContext := util.NewRuntimeContext(ctx)
	utilSpanContext := util.NewSpanContext(traceSpan)
	fields := map[string]any{
		"name":           "CreateGameLog",
		"rt_ctx":         utilRuntimeContext,
		"sp_ctx":         utilSpanContext,
		"config":         service.GetConfigger(),
		"bo_game_logger": boGameLogger,
	}

	service.GetRuntimeLogger().
		WithFields(fields).
		Info(object.URIEmpty)

	nowUTC := service.GetTimer().NowUTC()

	// If timestamp is not provided, use current time
	timestamp := boGameLogger.GetTimestamp()
	if timestamp.IsZero() {
		timestamp = nowUTC
	}

	daoGameLogger := dao.NewGameLog(
		uuid.Nil,
		boGameLogger.GetGameID(),
		boGameLogger.GetPlayerID(),
		boGameLogger.GetAction(),
		boGameLogger.GetDescription(),
		timestamp,
		nowUTC,
		nowUTC,
		sql.NullTime{
			Time:  time.Time{},
			Valid: false,
		},
	)

	service.GetRuntimeLogger().
		WithFields(fields).
		WithField(object.URIFieldDAOGameLogger, daoGameLogger).
		Debug(object.URIEmpty)

	id, err := service.GetGameLogRepositorier().Create(ctx, daoGameLogger)
	if err != nil {
		service.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error("Failed to create game log")
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, "Failed to create game log")

		return uuid.Nil, err
	}

	service.GetRuntimeLogger().
		WithFields(fields).
		WithField(object.URIFieldID, id).
		Debug(object.URIEmpty)

	return id.GetID()["id"], nil
}

// GetGameLogsByGameID implements input.GameLogServicer
func (service *gameLogService) GetGameLogsByGameID(
	ctx context.Context,
	gameID uuid.UUID,
	daoPaginationer dao.Paginationer,
) ([]bo.GameLogger, dao.Cursorer, error) {
	var traceSpan trace.Span

	ctx, traceSpan = service.GetTracer().Start(
		ctx,
		"GetGameLogsByGameID",
		trace.WithSpanKind(trace.SpanKindInternal),
	)
	defer traceSpan.End()

	utilRuntimeContext := util.NewRuntimeContext(ctx)
	utilSpanContext := util.NewSpanContext(traceSpan)
	fields := map[string]any{
		"name":             "GetGameLogsByGameID",
		"rt_ctx":           utilRuntimeContext,
		"sp_ctx":           utilSpanContext,
		"config":           service.GetConfigger(),
		"game_id":          gameID,
		"dao_paginationer": daoPaginationer,
	}

	service.GetRuntimeLogger().
		WithFields(fields).
		Info(object.URIEmpty)

	// Create a filter for game logs by game ID
	gameLogFilter := dao.NewGameLogFilter(
		[]uuid.UUID{},
		gameID,
		uuid.Nil,
		"", // No specific action filter
		nil,
		nil,
	)

	service.GetRuntimeLogger().
		WithFields(fields).
		WithField("game_log_filter", gameLogFilter).
		Debug(object.URIEmpty)

	// Call repository to list game logs with the filter
	daoGameLoggers, daoCursorer, err := service.GetGameLogRepositorier().
		ReadList(
			ctx,
			daoPaginationer,
			gameLogFilter,
		)
	if err != nil {
		service.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error("Failed to list game logs by game ID")
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, "Failed to list game logs by game ID")

		return nil, nil, err
	}

	service.GetRuntimeLogger().
		WithFields(fields).
		WithField("dao_game_loggers_count", len(daoGameLoggers)).
		WithField("dao_cursorer", daoCursorer).
		Debug(object.URIEmpty)

	// Convert DAO game logs to business objects
	boGameLoggers := make([]bo.GameLogger, 0, len(daoGameLoggers))
	for _, daoGameLogger := range daoGameLoggers {
		boGameLogger := bo.NewGameLog(
			daoGameLogger.GetCUDIDer().GetID()["id"],
			daoGameLogger.GetGameID(),
			daoGameLogger.GetPlayerID(),
			daoGameLogger.GetAction(),
			daoGameLogger.GetDescription(),
			daoGameLogger.GetTimestamp(),
		)
		boGameLoggers = append(boGameLoggers, boGameLogger)
	}

	service.GetRuntimeLogger().
		WithFields(fields).
		WithField("bo_game_loggers_count", len(boGameLoggers)).
		Debug(object.URIEmpty)

	return boGameLoggers, daoCursorer, nil
}

// GetGameLogsByPlayerID implements input.GameLogServicer.
func (service *gameLogService) GetGameLogsByPlayerID(
	ctx context.Context,
	playerID uuid.UUID,
	daoPaginationer dao.Paginationer,
) ([]bo.GameLogger, dao.Cursorer, error) {
	var traceSpan trace.Span

	ctx, traceSpan = service.GetTracer().Start(
		ctx,
		"GetGameLogsByPlayerID",
		trace.WithSpanKind(trace.SpanKindInternal),
	)
	defer traceSpan.End()

	utilRuntimeContext := util.NewRuntimeContext(ctx)
	utilSpanContext := util.NewSpanContext(traceSpan)
	fields := map[string]any{
		"name":             "GetGameLogsByPlayerID",
		"rt_ctx":           utilRuntimeContext,
		"sp_ctx":           utilSpanContext,
		"config":           service.GetConfigger(),
		"player_id":        playerID,
		"dao_paginationer": daoPaginationer,
	}

	service.GetRuntimeLogger().
		WithFields(fields).
		Info(object.URIEmpty)

	// Create a filter for game logs by player ID
	// Note: There appears to be a custom implementation of NewGameLogFilter in this project
	// that differs from the dao package definition
	gameLogFilter := dao.NewGameLogFilter(
		[]uuid.UUID{},
		uuid.Nil,
		playerID,
		"", // No specific action filter
		nil,
		nil,
	)

	service.GetRuntimeLogger().
		WithFields(fields).
		WithField("game_log_filter", gameLogFilter).
		Debug(object.URIEmpty)

	// Call repository to list game logs with the filter
	daoGameLoggers, daoCursorer, err := service.GetGameLogRepositorier().ReadList(
		ctx,
		daoPaginationer,
		gameLogFilter,
	)
	if err != nil {
		service.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error("Failed to list game logs by player ID")
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, "Failed to list game logs by player ID")

		return nil, nil, err
	}

	service.GetRuntimeLogger().
		WithFields(fields).
		WithField("dao_game_loggers_count", len(daoGameLoggers)).
		WithField("dao_cursorer", daoCursorer).
		Debug(object.URIEmpty)

	// Convert DAO game logs to business objects
	boGameLoggers := make([]bo.GameLogger, 0, len(daoGameLoggers))
	for _, daoGameLogger := range daoGameLoggers {
		boGameLogger := bo.NewGameLog(
			daoGameLogger.GetCUDIDer().GetID()["id"],
			daoGameLogger.GetGameID(),
			daoGameLogger.GetPlayerID(),
			daoGameLogger.GetAction(),
			daoGameLogger.GetDescription(),
			daoGameLogger.GetTimestamp(),
		)
		boGameLoggers = append(boGameLoggers, boGameLogger)
	}

	service.GetRuntimeLogger().
		WithFields(fields).
		WithField("bo_game_loggers_count", len(boGameLoggers)).
		Debug(object.URIEmpty)

	return boGameLoggers, daoCursorer, nil
}

// GetGameLogsByTimeRange implements input.GameLogServicer.
func (service *gameLogService) GetGameLogsByTimeRange(
	ctx context.Context,
	gameID uuid.UUID,
	startTime time.Time,
	endTime time.Time,
	daoPaginationer dao.Paginationer,
) ([]bo.GameLogger, dao.Cursorer, error) {
	var traceSpan trace.Span

	ctx, traceSpan = service.GetTracer().Start(
		ctx,
		"GetGameLogsByTimeRange",
		trace.WithSpanKind(trace.SpanKindInternal),
	)
	defer traceSpan.End()

	utilRuntimeContext := util.NewRuntimeContext(ctx)
	utilSpanContext := util.NewSpanContext(traceSpan)
	fields := map[string]any{
		"name":             "GetGameLogsByTimeRange",
		"rt_ctx":           utilRuntimeContext,
		"sp_ctx":           utilSpanContext,
		"config":           service.GetConfigger(),
		"game_id":          gameID,
		"start_time":       startTime,
		"end_time":         endTime,
		"dao_paginationer": daoPaginationer,
	}

	service.GetRuntimeLogger().
		WithFields(fields).
		Info(object.URIEmpty)

	// Create a filter for game logs by game ID and time range
	gameLogFilter := dao.NewGameLogFilter(
		[]uuid.UUID{},
		gameID,
		uuid.Nil,
		"", // No specific action filter
		&startTime,
		&endTime,
	)

	service.GetRuntimeLogger().
		WithFields(fields).
		WithField("game_log_filter", gameLogFilter).
		Debug(object.URIEmpty)

	// Call repository to list game logs with the filter
	daoGameLoggers, daoCursorer, err := service.GetGameLogRepositorier().ReadList(
		ctx,
		daoPaginationer,
		gameLogFilter,
	)
	if err != nil {
		service.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error("Failed to list game logs by time range")
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, "Failed to list game logs by time range")

		return nil, nil, err
	}

	service.GetRuntimeLogger().
		WithFields(fields).
		WithField("dao_game_loggers_count", len(daoGameLoggers)).
		WithField("dao_cursorer", daoCursorer).
		Debug(object.URIEmpty)

	// Convert DAO game logs to business objects
	boGameLoggers := make([]bo.GameLogger, 0, len(daoGameLoggers))
	for _, daoGameLogger := range daoGameLoggers {
		boGameLogger := bo.NewGameLog(
			daoGameLogger.GetCUDIDer().GetID()["id"],
			daoGameLogger.GetGameID(),
			daoGameLogger.GetPlayerID(),
			daoGameLogger.GetAction(),
			daoGameLogger.GetDescription(),
			daoGameLogger.GetTimestamp(),
		)
		boGameLoggers = append(boGameLoggers, boGameLogger)
	}

	service.GetRuntimeLogger().
		WithFields(fields).
		WithField("bo_game_loggers_count", len(boGameLoggers)).
		Debug(object.URIEmpty)

	return boGameLoggers, daoCursorer, nil
}

// WithOptioners applies the provided optioners to the service
func (service *gameLogService) WithOptioners(
	optioners ...gameLogServiceOptioner,
) *gameLogService {
	newService := service.clone()

	for _, optioner := range optioners {
		optioner.apply(newService)
	}

	return newService
}

func (service *gameLogService) clone() *gameLogService {
	newService := *service

	return &newService
}

func (optionerFunc gameLogServiceOptionerFunc) apply(
	service *gameLogService,
) {
	optionerFunc(service)
}

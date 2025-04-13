package repository

import (
	"context"
	"database/sql"
	"fmt"
	"github/M2A96/Monopoly.git/config"
	"github/M2A96/Monopoly.git/internal/core/ports/output"
	"github/M2A96/Monopoly.git/log"
	"github/M2A96/Monopoly.git/object"
	"github/M2A96/Monopoly.git/object/dao"
	"github/M2A96/Monopoly.git/util"
	"time"

	"github.com/samber/lo"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
)

type (
	gameLogRepository struct {
		repository
	}

	gameLogRepositoryOptioner = repositoryOptioner
)

// Ensure gameLogRepository implements output.GameLogRepositorier
var _ output.GameLogRepositorier = (*gameLogRepository)(nil)

// NewGameLogRepository creates a new game log repository instance
func NewGameLogRepository(
	configConfigger config.Configger,
	logRuntimeLogger log.RuntimeLogger,
	objectUUIDer object.UUIDer,
	traceTracer trace.Tracer,
	optioners ...gameLogRepositoryOptioner,
) *gameLogRepository {
	return &gameLogRepository{
		repository: *NewRepository(
			configConfigger,
			logRuntimeLogger,
			objectUUIDer,
			traceTracer,
			optioners...,
		),
	}
}

// WithGameLogRepositoryTimer adds a timer to the game log repository
func WithGameLogRepositoryTimer(
	objectTimer object.Timer,
) gameLogRepositoryOptioner {
	return WithRepositoryTimer(objectTimer)
}

// WithGameLogRepositoryDB adds a database connection to the game log repository
func WithGameLogRepositoryDB(
	gormDB *gorm.DB,
) gameLogRepositoryOptioner {
	return WithRepositoryDB(gormDB, object.URITableGameLog)
}

// Create implements the Create method of the GameLogRepositorier interface
func (repository *gameLogRepository) Create(
	ctx context.Context,
	daoGameLog dao.GameLogger,
) (dao.CUDIDer, error) {
	var traceSpan trace.Span

	ctx, traceSpan = repository.GetTracer().Start(
		ctx,
		"Create",
		trace.WithSpanKind(trace.SpanKindClient),
	)
	defer traceSpan.End()

	utilRuntimeContext := util.NewRuntimeContext(ctx)
	utilSpanContext := util.NewSpanContext(traceSpan)
	fields := map[string]any{
		"name":         "Create",
		"rt_ctx":       utilRuntimeContext,
		"sp_ctx":       utilSpanContext,
		"config":       repository.GetConfigger(),
		"dao_game_log": daoGameLog,
	}

	repository.GetRuntimeLogger().
		WithFields(fields).
		Info(object.URIEmpty)

	id, err := repository.GetUUIDer().NewRandom()
	if err != nil {
		repository.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error(object.ErrUUIDerNewRandom.Error())
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, object.ErrUUIDerNewRandom.Error())

		return nil, err
	}

	repository.GetRuntimeLogger().
		WithFields(fields).
		WithField(object.URIFieldID, id).
		Debug(object.URIEmpty)

	nowUTC := repository.GetTimer().NowUTC()

	repository.GetRuntimeLogger().
		WithFields(fields).
		WithField(object.URIFieldNowUTC, nowUTC).
		Debug(object.URIEmpty)

	newGameLog := dao.NewGameLog(
		id,
		daoGameLog.GetGameID(),
		daoGameLog.GetPlayerID(),
		daoGameLog.GetAction(),
		daoGameLog.GetDescription(),
		daoGameLog.GetTimestamp(),
		nowUTC,
		nowUTC,
		sql.NullTime{
			Time:  time.Time{},
			Valid: false,
		},
	)

	repository.GetRuntimeLogger().
		WithFields(fields).
		WithField("dao_game_log", newGameLog).
		Debug(object.URIEmpty)

	gormDB := repository.GetDB().
		WithContext(ctx).
		Create(newGameLog.GetMap())
	if err = gormDB.Error; err != nil {
		repository.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error("Failed to create game log")
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, "Failed to create game log")

		return nil, err
	}

	return newGameLog.GetCUDIDer(), nil
}

// Read implements the Read method of the GameLogRepositorier interface
func (repository *gameLogRepository) Read(
	ctx context.Context,
	id dao.CUDIDer,
) (dao.GameLogger, error) {
	var traceSpan trace.Span

	ctx, traceSpan = repository.GetTracer().Start(
		ctx,
		"Read",
		trace.WithSpanKind(trace.SpanKindClient),
	)
	defer traceSpan.End()

	utilRuntimeContext := util.NewRuntimeContext(ctx)
	utilSpanContext := util.NewSpanContext(traceSpan)
	fields := map[string]any{
		"name":   "Read",
		"rt_ctx": utilRuntimeContext,
		"sp_ctx": utilSpanContext,
		"config": repository.GetConfigger(),
		"id":     id.GetID(),
	}

	repository.GetRuntimeLogger().
		WithFields(fields).
		Info(object.URIEmpty)

	result := map[string]any{}

	gormDB := repository.GetDB().
		WithContext(ctx).
		Where(lo.Assign(
			id.GetMap(),
			map[string]any{
				"deleted_at": nil,
			},
		)).
		Select(fmt.Sprintf(`%s.*`, object.URITableGameLog)).
		Find(result)
	if err := gormDB.Error; err != nil {
		repository.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error("Failed to read game log")
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, "Failed to read game log")

		return nil, err
	}

	if len(result) == 0 {
		repository.GetRuntimeLogger().
			WithFields(fields).
			Error("Game log not found")
		traceSpan.SetStatus(codes.Error, "Game log not found")

		return nil, fmt.Errorf("game log not found")
	}

	gameLog, err := dao.NewGameLogFromMap(
		repository.GetUUIDer(),
		result,
	)
	if err != nil {
		repository.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error("Failed to create game log from map")
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, "Failed to create game log from map")

		return nil, err
	}

	return gameLog, nil
}

// ReadList implements the ReadList method of the GameLogRepositorier interface
func (repository *gameLogRepository) ReadList(
	ctx context.Context,
	pagination dao.Paginationer,
	filter dao.GameLogFilter,
) ([]dao.GameLogger, dao.Cursorer, error) {
	var traceSpan trace.Span

	ctx, traceSpan = repository.GetTracer().Start(
		ctx,
		"ReadList",
		trace.WithSpanKind(trace.SpanKindClient),
	)
	defer traceSpan.End()

	utilRuntimeContext := util.NewRuntimeContext(ctx)
	utilSpanContext := util.NewSpanContext(traceSpan)
	fields := map[string]any{
		"name":       "ReadList",
		"rt_ctx":     utilRuntimeContext,
		"sp_ctx":     utilSpanContext,
		"config":     repository.GetConfigger(),
		"pagination": pagination,
		"filter":     filter,
	}

	repository.GetRuntimeLogger().
		WithFields(fields).
		Info(object.URIEmpty)

	result := make([]map[string]any, 0, pagination.GetLimit()+1)

	gormDB := repository.GetDB().
		WithContext(ctx).
		Scopes(
			filter.Filter,
			pagination.Pagination(object.URITableGameLog),
		).
		Where(map[string]any{
			"deleted_at": nil,
		}).
		Select(fmt.Sprintf(`%s.*`, object.URITableGameLog)).
		Find(&result)
	if err := gormDB.Error; err != nil {
		repository.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error("Failed to read game log list")
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, "Failed to read game log list")

		return nil, nil, err
	}

	repository.GetRuntimeLogger().
		WithFields(fields).
		WithField(object.URIFieldResult, result).
		Debug(object.URIEmpty)

	daoGameLogs := make([]dao.GameLogger, 0, pagination.GetLimit())

	for key, value := range result {
		repository.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldKey, key).
			WithField(object.URIFieldValue, value).
			Debug(object.URIEmpty)

		if uint32(key) == pagination.GetLimit() {
			repository.GetRuntimeLogger().
				WithFields(fields).
				Debug(`uint32(key) == pagination.GetLimit()`)

			break
		}

		daoGameLog, err := dao.NewGameLogFromMap(repository.GetUUIDer(), value)
		if err != nil {
			repository.GetRuntimeLogger().
				WithFields(fields).
				WithField(object.URIFieldError, err).
				Error("Failed to create game log from map")
			traceSpan.RecordError(err)
			traceSpan.SetStatus(codes.Error, "Failed to create game log from map")

			return nil, nil, err
		}

		repository.GetRuntimeLogger().
			WithFields(fields).
			WithField("dao_game_log", daoGameLog).
			Debug(object.URIEmpty)

		daoGameLogs = append(daoGameLogs, daoGameLog)
	}

	repository.GetRuntimeLogger().
		WithFields(fields).
		WithField("dao_game_logs", daoGameLogs).
		Debug(object.URIEmpty)

	var daoCursorer dao.Cursorer

	if pagination.GetLimit() < uint32(len(result)) {
		repository.GetRuntimeLogger().
			WithFields(fields).
			Debug(`pagination.GetLimit() < uint32(len(result))`)

		daoCursorer = dao.NewCursor(
			pagination.GetCursorer().GetOffset() + pagination.GetLimit(),
		)
	}

	repository.GetRuntimeLogger().
		WithFields(fields).
		WithField(object.URIFieldDAOCursorer, daoCursorer).
		Debug(object.URIEmpty)

	return daoGameLogs, daoCursorer, nil
}

// Update implements the Update method of the GameLogRepositorier interface
func (repository *gameLogRepository) Update(
	ctx context.Context,
	gameLog dao.GameLogger,
) (time.Time, error) {
	var traceSpan trace.Span

	ctx, traceSpan = repository.GetTracer().Start(
		ctx,
		"Update",
		trace.WithSpanKind(trace.SpanKindClient),
	)
	defer traceSpan.End()

	utilRuntimeContext := util.NewRuntimeContext(ctx)
	utilSpanContext := util.NewSpanContext(traceSpan)
	fields := map[string]any{
		"name":         "Update",
		"rt_ctx":       utilRuntimeContext,
		"sp_ctx":       utilSpanContext,
		"config":       repository.GetConfigger(),
		"dao_game_log": gameLog,
	}

	repository.GetRuntimeLogger().
		WithFields(fields).
		Info(object.URIEmpty)

	nowUTC := repository.GetTimer().NowUTC()

	repository.GetRuntimeLogger().
		WithFields(fields).
		WithField(object.URIFieldNowUTC, nowUTC).
		Debug(object.URIEmpty)

	daoGameLog := dao.NewGameLog(
		gameLog.GetCUDIDer().GetID()["id"],
		gameLog.GetGameID(),
		gameLog.GetPlayerID(),
		gameLog.GetAction(),
		gameLog.GetDescription(),
		gameLog.GetTimestamp(),
		gameLog.GetCUDer().GetCreatedAt(),
		nowUTC,
		sql.NullTime{
			Time:  time.Time{},
			Valid: false,
		},
	)

	gormDB := repository.GetDB().
		WithContext(ctx).
		Where(lo.Assign(
			daoGameLog.GetCUDIDer().GetMap(),
			map[string]any{
				"deleted_at": nil,
			},
		)).
		Updates(lo.Assign(
			daoGameLog.GetMap(),
			map[string]any{
				"updated_at": nowUTC,
			},
		))
	if err := gormDB.Error; err != nil {
		repository.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error("Failed to update game log")
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, "Failed to update game log")

		return time.Time{}, err
	}

	if gormDB.RowsAffected == 0 {
		repository.GetRuntimeLogger().
			WithFields(fields).
			Error("Game log not found for update")
		traceSpan.SetStatus(codes.Error, "Game log not found for update")

		return time.Time{}, fmt.Errorf("game log not found for update")
	}

	return nowUTC, nil
}

// Delete implements the Delete method of the GameLogRepositorier interface
func (repository *gameLogRepository) Delete(
	ctx context.Context,
	id dao.CUDIDer,
) (time.Time, error) {
	var traceSpan trace.Span

	ctx, traceSpan = repository.GetTracer().Start(
		ctx,
		"Delete",
		trace.WithSpanKind(trace.SpanKindClient),
	)
	defer traceSpan.End()

	utilRuntimeContext := util.NewRuntimeContext(ctx)
	utilSpanContext := util.NewSpanContext(traceSpan)
	fields := map[string]any{
		"name":   "Delete",
		"rt_ctx": utilRuntimeContext,
		"sp_ctx": utilSpanContext,
		"config": repository.GetConfigger(),
		"id":     id.GetID(),
	}

	repository.GetRuntimeLogger().
		WithFields(fields).
		Info(object.URIEmpty)

	nowUTC := repository.GetTimer().NowUTC()

	repository.GetRuntimeLogger().
		WithFields(fields).
		WithField(object.URIFieldNowUTC, nowUTC).
		Debug(object.URIEmpty)

	gormDB := repository.GetDB().
		WithContext(ctx).
		Where(lo.Assign(
			id.GetMap(),
			map[string]any{
				"deleted_at": nil,
			},
		)).
		Updates(map[string]any{
			"deleted_at": nowUTC,
		})
	if err := gormDB.Error; err != nil {
		repository.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error("Failed to delete game log")
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, "Failed to delete game log")

		return time.Time{}, err
	}

	if gormDB.RowsAffected == 0 {
		repository.GetRuntimeLogger().
			WithFields(fields).
			Error("Game log not found for delete")
		traceSpan.SetStatus(codes.Error, "Game log not found for delete")

		return time.Time{}, fmt.Errorf("game log not found for delete")
	}

	return nowUTC, nil
}

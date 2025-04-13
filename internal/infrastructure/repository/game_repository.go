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
	gameRepository struct {
		repository
	}

	gameRepositoryOptioner = repositoryOptioner
)

// Ensure gameRepository implements output.GameRepositorier
var _ output.GameRepositorier = (*gameRepository)(nil)

// NewGameRepository creates a new game repository instance
func NewGameRepository(
	configConfigger config.Configger,
	logRuntimeLogger log.RuntimeLogger,
	objectUUIDer object.UUIDer,
	traceTracer trace.Tracer,
	optioners ...gameRepositoryOptioner,
) *gameRepository {
	return &gameRepository{
		repository: *NewRepository(
			configConfigger,
			logRuntimeLogger,
			objectUUIDer,
			traceTracer,
			optioners...,
		),
	}
}

// WithGameRepositoryTimer adds a timer to the game repository
func WithGameRepositoryTimer(
	objectTimer object.Timer,
) gameRepositoryOptioner {
	return WithRepositoryTimer(objectTimer)
}

// WithGameRepositoryDB adds a database connection to the game repository
func WithGameRepositoryDB(
	gormDB *gorm.DB,
) gameRepositoryOptioner {
	return WithRepositoryDB(gormDB, object.URITableGame)
}

// Create implements the Create method of the GameRepositorier interface
func (repository *gameRepository) Create(
	ctx context.Context,
	daoGame dao.Gamer,
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
		"name":     "Create",
		"rt_ctx":   utilRuntimeContext,
		"sp_ctx":   utilSpanContext,
		"config":   repository.GetConfigger(),
		"dao_game": daoGame,
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

	newGame := dao.NewGame(
		nowUTC,
		nowUTC,
		sql.NullTime{
			Time:  time.Time{},
			Valid: false,
		},
		id,
		daoGame.GetName(),
		daoGame.GetStatus(),
		daoGame.GetCurrentPlayerID(),
		daoGame.GetWinnerID(),
	)

	repository.GetRuntimeLogger().
		WithFields(fields).
		WithField(object.URIFieldDAOGame, newGame).
		Debug(object.URIEmpty)

	gormDB := repository.GetDB().
		WithContext(ctx).
		Create(newGame.GetMap())
	if err = gormDB.Error; err != nil {
		repository.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error(object.ErrGameRepositoryCreate.Error())
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, object.ErrGameRepositoryCreate.Error())

		return nil, err
	}

	return newGame.GetCUDIDer(), nil
}

// Read implements the Read method of the GameRepositorier interface
func (repository *gameRepository) Read(
	ctx context.Context,
	id dao.CUDIDer,
) (dao.Gamer, error) {
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
		Select(fmt.Sprintf(`%s.*`, object.URITableGame)).
		Find(result)
	if err := gormDB.Error; err != nil {
		repository.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error(object.ErrGameRepositoryRead.Error())
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, object.ErrGameRepositoryRead.Error())

		return nil, err
	}

	if gormDB.RowsAffected == 0 {
		repository.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, object.ErrGameRepositoryRead).
			Error(object.ErrGameRepositoryRead.Error())
		traceSpan.RecordError(object.ErrGameRepositoryRead)
		traceSpan.SetStatus(codes.Error, object.ErrGameRepositoryRead.Error())

		return nil, object.ErrGameRepositoryRead
	}

	repository.GetRuntimeLogger().
		WithFields(fields).
		WithField(object.URIFieldResult, result).
		Debug(object.URIEmpty)

	newGame, err := dao.NewGamerFromMap(repository.GetUUIDer(), result)
	if err != nil {
		repository.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error(object.ErrDAONewGameFromMap.Error())
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, object.ErrDAONewGameFromMap.Error())

		return nil, err
	}
	repository.GetRuntimeLogger().
		WithFields(fields).
		WithField(object.URIFieldDAOGame, newGame).
		Debug(object.URIEmpty)

	return newGame, nil
}

// ReadList implements the ReadList method of the GameRepositorier interface
func (repository *gameRepository) ReadList(
	ctx context.Context,
	pagination dao.Paginationer,
	filter dao.GameFilter,
) ([]dao.Gamer, dao.Cursorer, error) {
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
			pagination.Pagination(object.URITableGame),
		).
		Where(map[string]any{
			"deleted_at": nil,
		}).
		Select(fmt.Sprintf(`%s.*`, object.URITableGame)).
		Find(&result)
	if err := gormDB.Error; err != nil {
		repository.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error(object.ErrGameRepositoryReadList.Error())
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, object.ErrGameRepositoryReadList.Error())

		return nil, nil, err
	}

	repository.GetRuntimeLogger().
		WithFields(fields).
		WithField(object.URIFieldResult, result).
		Debug(object.URIEmpty)

	daoGamers := make([]dao.Gamer, 0, pagination.GetLimit())

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

		daoGamer, err := dao.NewGamerFromMap(repository.GetUUIDer(), value)
		if err != nil {
			repository.GetRuntimeLogger().
				WithFields(fields).
				WithField(object.URIFieldError, err).
				Error(object.ErrDAONewGameFromMap.Error())
			traceSpan.RecordError(err)
			traceSpan.SetStatus(codes.Error, object.ErrDAONewGameFromMap.Error())

			return nil, nil, err
		}

		repository.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldDAOGame, daoGamer).
			Debug(object.URIEmpty)

		daoGamers = append(daoGamers, daoGamer)
	}

	repository.GetRuntimeLogger().
		WithFields(fields).
		WithField(object.URIFieldDAOGames, daoGamers).
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

	return daoGamers, daoCursorer, nil
}

// Update implements the Update method of the GameRepositorier interface
func (repository *gameRepository) Update(
	ctx context.Context,
	game dao.Gamer,
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
		"name":     "Update",
		"rt_ctx":   utilRuntimeContext,
		"sp_ctx":   utilSpanContext,
		"config":   repository.GetConfigger(),
		"game_id":  game.GetCUDIDer().GetID(),
		"dao_game": game,
	}

	repository.GetRuntimeLogger().
		WithFields(fields).
		Info(object.URIEmpty)

	nowUTC := repository.GetTimer().NowUTC()

	repository.GetRuntimeLogger().
		WithFields(fields).
		WithField(object.URIFieldNowUTC, nowUTC).
		Debug(object.URIEmpty)

	daoGame := dao.NewGame(
		game.GetCUDer().GetCreatedAt(),
		nowUTC,
		sql.NullTime{
			Time:  time.Time{},
			Valid: false,
		},
		game.GetCUDIDer().GetID()["id"],
		game.GetName(),
		game.GetStatus(),
		game.GetCurrentPlayerID(),
		game.GetWinnerID(),
	)

	repository.GetRuntimeLogger().
		WithFields(fields).
		WithField(object.URIFieldDAOGame, daoGame).
		Debug(object.URIEmpty)

	// Update the game in the database
	gormDB := repository.GetDB().
		WithContext(ctx).
		Where(lo.Assign(
			daoGame.GetCUDIDer().GetMap(),
			map[string]any{
				"deleted_at": nil,
			},
		)).
		Updates(daoGame.GetMap())
	if err := gormDB.Error; err != nil {
		repository.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error(object.ErrGameRepositoryUpdate.Error())
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, object.ErrGameRepositoryUpdate.Error())

		return time.Time{}, err
	}

	return nowUTC, nil
}

// Delete implements the Delete method of the GameRepositorier interface
func (repository *gameRepository) Delete(
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

	// Delete the game from the database
	gormDB := repository.GetDB().
		WithContext(ctx).
		Where("id = ?", id.GetID()).
		Delete(nil)
	if err := gormDB.Error; err != nil {
		repository.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error(object.ErrGameRepositoryDelete.Error())
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, object.ErrGameRepositoryDelete.Error())

		return time.Time{}, err
	}

	return nowUTC, nil
}

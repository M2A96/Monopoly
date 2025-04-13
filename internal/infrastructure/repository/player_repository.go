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
	playerRepository struct {
		repository
	}

	playerRepositoryOptioner = repositoryOptioner
)

// Ensure playerRepository implements output.PlayerRepositorier
var _ output.PlayerRepositorier = (*playerRepository)(nil)

// NewPlayerRepository creates a new player repository instance
func NewPlayerRepository(
	configConfigger config.Configger,
	logRuntimeLogger log.RuntimeLogger,
	objectUUIDer object.UUIDer,
	traceTracer trace.Tracer,
	optioners ...playerRepositoryOptioner,
) *playerRepository {
	return &playerRepository{
		repository: *NewRepository(
			configConfigger,
			logRuntimeLogger,
			objectUUIDer,
			traceTracer,
			optioners...,
		),
	}
}

// WithPlayerRepositoryTimer adds a timer to the player repository
func WithPlayerRepositoryTimer(
	objectTimer object.Timer,
) playerRepositoryOptioner {
	return WithRepositoryTimer(objectTimer)
}

// WithPlayerRepositoryDB adds a database connection to the player repository
func WithPlayerRepositoryDB(
	gormDB *gorm.DB,
) playerRepositoryOptioner {
	return WithRepositoryDB(gormDB, object.URITablePlayer)
}

// Create implements the Create method of the PlayerRepositorier interface
func (repository *playerRepository) Create(
	ctx context.Context,
	daoPlayer dao.Player,
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
		"name":       "Create",
		"rt_ctx":     utilRuntimeContext,
		"sp_ctx":     utilSpanContext,
		"config":     repository.GetConfigger(),
		"dao_player": daoPlayer,
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

	newPlayer := dao.NewPlayer(
		id,
		daoPlayer.GetGameID(),
		daoPlayer.GetName(),
		daoPlayer.GetBalance(),
		daoPlayer.GetPosition(),
		daoPlayer.GetInJail(),
		daoPlayer.GetJailTurns(),
		daoPlayer.GetBankrupt(),
		nowUTC,
		nowUTC,
		sql.NullTime{
			Time:  time.Time{},
			Valid: false,
		},
	)

	repository.GetRuntimeLogger().
		WithFields(fields).
		WithField(object.URIFieldDAOPlayer, newPlayer).
		Debug(object.URIEmpty)

	gormDB := repository.GetDB().
		WithContext(ctx).
		Create(newPlayer.GetMap())
	if err = gormDB.Error; err != nil {
		repository.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error(object.ErrPlayerRepositoryCreate.Error())
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, object.ErrPlayerRepositoryCreate.Error())

		return nil, err
	}

	return newPlayer.GetCUDIDer(), nil
}

// Read implements the Read method of the PlayerRepositorier interface
func (repository *playerRepository) Read(
	ctx context.Context,
	id dao.CUDIDer,
) (dao.Player, error) {
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
		Select(fmt.Sprintf(`%s.*`, object.URITablePlayer)).
		Find(result)
	if err := gormDB.Error; err != nil {
		repository.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error(object.ErrPlayerRepositoryRead.Error())
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, object.ErrPlayerRepositoryRead.Error())

		return nil, err
	}

	if len(result) == 0 {
		repository.GetRuntimeLogger().
			WithFields(fields).
			Error(object.ErrPlayerRepositoryReadNotFound.Error())
		traceSpan.SetStatus(codes.Error, object.ErrPlayerRepositoryReadNotFound.Error())

		return nil, object.ErrPlayerRepositoryReadNotFound
	}

	player, err := dao.NewPlayerFromMap(
		repository.GetUUIDer(),
		result,
	)
	if err != nil {
		repository.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error(object.ErrPlayerRepositoryRead.Error())
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, object.ErrPlayerRepositoryRead.Error())

		return nil, err
	}

	return player, nil
}

// ReadList implements the ReadList method of the PlayerRepositorier interface
func (repository *playerRepository) ReadList(
	ctx context.Context,
	pagination dao.Paginationer,
	filter dao.PlayerFilter,
) ([]dao.Player, dao.Cursorer, error) {
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
			pagination.Pagination(object.URITablePlayer),
		).
		Where(map[string]any{
			"deleted_at": nil,
		}).
		Select(fmt.Sprintf(`%s.*`, object.URITablePlayer)).
		Find(&result)
	if err := gormDB.Error; err != nil {
		repository.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error("Failed to read property list")
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, "Failed to read property list")

		return nil, nil, err
	}

	repository.GetRuntimeLogger().
		WithFields(fields).
		WithField(object.URIFieldResult, result).
		Debug(object.URIEmpty)

	daoPlayers := make([]dao.Player, 0, pagination.GetLimit())

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

		daoPlayer, err := dao.NewPlayerFromMap(repository.GetUUIDer(), value)
		if err != nil {
			repository.GetRuntimeLogger().
				WithFields(fields).
				WithField(object.URIFieldError, err).
				Error("Failed to create property from map")
			traceSpan.RecordError(err)
			traceSpan.SetStatus(codes.Error, "Failed to create property from map")

			return nil, nil, err
		}

		repository.GetRuntimeLogger().
			WithFields(fields).
			WithField("dao_property", daoPlayer).
			Debug(object.URIEmpty)

		daoPlayers = append(daoPlayers, daoPlayer)
	}

	repository.GetRuntimeLogger().
		WithFields(fields).
		WithField("dao_properties", daoPlayers).
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

	return daoPlayers, daoCursorer, nil
}

// Update implements the Update method of the PlayerRepositorier interface
func (repository *playerRepository) Update(
	ctx context.Context,
	player dao.Player,
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
		"name":       "Update",
		"rt_ctx":     utilRuntimeContext,
		"sp_ctx":     utilSpanContext,
		"config":     repository.GetConfigger(),
		"dao_player": player,
	}

	repository.GetRuntimeLogger().
		WithFields(fields).
		Info(object.URIEmpty)

	nowUTC := repository.GetTimer().NowUTC()

	repository.GetRuntimeLogger().
		WithFields(fields).
		WithField(object.URIFieldNowUTC, nowUTC).
		Debug(object.URIEmpty)

	daoPlayer := dao.NewPlayer(
		player.GetCUDIDer().GetID()["id"],
		player.GetGameID(),
		player.GetName(),
		player.GetBalance(),
		player.GetPosition(),
		player.GetInJail(),
		player.GetJailTurns(),
		player.GetBankrupt(),
		player.GetCUDer().GetCreatedAt(),
		nowUTC,
		sql.NullTime{
			Time:  time.Time{},
			Valid: false,
		},
	)

	gormDB := repository.GetDB().
		WithContext(ctx).
		Where(lo.Assign(
			daoPlayer.GetCUDIDer().GetMap(),
			map[string]any{
				"deleted_at": nil,
			},
		)).
		Updates(lo.Assign(
			daoPlayer.GetMap(),
			map[string]any{
				"updated_at": nowUTC,
			},
		))
	if err := gormDB.Error; err != nil {
		repository.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error(object.ErrPlayerRepositoryUpdate.Error())
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, object.ErrPlayerRepositoryUpdate.Error())

		return time.Time{}, err
	}

	if gormDB.RowsAffected == 0 {
		repository.GetRuntimeLogger().
			WithFields(fields).
			Error(object.ErrPlayerRepositoryUpdateNotFound.Error())
		traceSpan.SetStatus(codes.Error, object.ErrPlayerRepositoryUpdateNotFound.Error())

		return time.Time{}, object.ErrPlayerRepositoryUpdateNotFound
	}

	return nowUTC, nil
}

// Delete implements the Delete method of the PlayerRepositorier interface
func (repository *playerRepository) Delete(
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
			Error(object.ErrPlayerRepositoryDelete.Error())
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, object.ErrPlayerRepositoryDelete.Error())

		return time.Time{}, err
	}

	if gormDB.RowsAffected == 0 {
		repository.GetRuntimeLogger().
			WithFields(fields).
			Error(object.ErrPlayerRepositoryNotFound.Error())
		traceSpan.SetStatus(codes.Error, object.ErrPlayerRepositoryNotFound.Error())

		return time.Time{}, object.ErrPlayerRepositoryNotFound
	}

	return nowUTC, nil
}

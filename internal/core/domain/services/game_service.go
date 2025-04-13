// internal/services/game_service.go
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

// GameService implements the input.GameService interface
type (
	gameService struct {
		configConfigger            config.Configger
		logRuntimeLogger           log.RuntimeLogger
		objectUUIDer               object.UUIDer
		repositoryGameRepositorier output.GameRepositorier
		serviceServicers           Servicers
		traceTracer                trace.Tracer
	}

	gameServiceOptioner interface {
		apply(*gameService)
	}

	gameServiceOptionerFunc func(*gameService)
)

var (
	_ input.GameServicer         = (*gameService)(nil)
	_ GetServicers               = (*gameService)(nil)
	_ WithServicers              = (*gameService)(nil)
	_ config.GetConfigger        = (*gameService)(nil)
	_ log.GetRuntimeLogger       = (*gameService)(nil)
	_ object.GetUUIDer           = (*gameService)(nil)
	_ output.GetGameRepositorier = (*gameService)(nil)
	_ util.GetTracer             = (*gameService)(nil)
)

// NewGameService is a function.
func NewGameService(
	configConfigger config.Configger,
	logRuntimeLogger log.RuntimeLogger,
	objectUUIDer object.UUIDer,
	traceTracer trace.Tracer,
	optioners ...gameServiceOptioner,
) *gameService {
	gameService := &gameService{
		configConfigger:            configConfigger,
		logRuntimeLogger:           logRuntimeLogger,
		objectUUIDer:               objectUUIDer,
		repositoryGameRepositorier: nil,
		serviceServicers:           nil,
		traceTracer:                traceTracer,
	}

	return gameService.WithOptioners(optioners...)
}

// WithGameServiceGameRepositorier is a function.
func WithGameServiceGameRepositorier(
	repositoryGameRepositorier output.GameRepositorier,
) gameServiceOptioner {
	return gameServiceOptionerFunc(func(
		config *gameService,
	) {
		config.repositoryGameRepositorier = repositoryGameRepositorier
	})
}

// GetTracer implements util.GetTracer.
func (g *gameService) GetTracer() trace.Tracer {
	return g.traceTracer
}

// GetAddressRepositorier implements output.GetGameRepositorier.
func (g *gameService) GetGameRepositorier() output.GameRepositorier {
	return g.repositoryGameRepositorier
}

// GetUUIDer implements object.GetUUIDer.
func (g *gameService) GetUUIDer() object.UUIDer {
	return g.objectUUIDer
}

// GetRuntimeLogger implements log.GetRuntimeLogger.
func (g *gameService) GetRuntimeLogger() log.RuntimeLogger {
	return g.logRuntimeLogger
}

// GetConfigger implements config.GetConfigger.
func (g *gameService) GetConfigger() config.Configger {
	return g.configConfigger
}

// GetServicers implements input.GetServicers.
func (g *gameService) GetServicers() Servicers {
	return g.serviceServicers
}

// CreateGame implements input.GameServicer.
func (service *gameService) CreateGame(
	ctx context.Context,
	boGamer bo.Gamer,
) (uuid.UUID, error) {
	var traceSpan trace.Span

	ctx, traceSpan = service.GetTracer().Start(
		ctx,
		"CreateGame",
		trace.WithSpanKind(trace.SpanKindInternal),
	)
	defer traceSpan.End()

	utilRuntimeContext := util.NewRuntimeContext(ctx)
	utilSpanContext := util.NewSpanContext(traceSpan)
	fields := map[string]any{
		"name":     "Create",
		"rt_ctx":   utilRuntimeContext,
		"sp_ctx":   utilSpanContext,
		"config":   service.GetConfigger(),
		"bo_gamer": boGamer,
	}

	service.GetRuntimeLogger().
		WithFields(fields).
		Info(object.URIEmpty)

	daoGame := dao.NewGame(
		time.Time{},
		time.Time{},
		sql.NullTime{
			Time:  time.Time{},
			Valid: false,
		},
		uuid.Nil,
		boGamer.GetName(),
		boGamer.GetStatus(),
		boGamer.GetCurrentPlayerID(),
		boGamer.GetWinnerID(),
	)

	service.GetRuntimeLogger().
		WithFields(fields).
		WithField(object.URIFieldDAOGame, daoGame).
		Debug(object.URIEmpty)

	id, err := service.GetGameRepositorier().Create(ctx, daoGame)
	if err != nil {
		service.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error(object.ErrGameServiceCreate.Error())
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, object.ErrGameServiceCreate.Error())

		return uuid.Nil, err
	}

	service.GetRuntimeLogger().
		WithFields(fields).
		WithField(object.URIFieldID, id).
		Debug(object.URIEmpty)

	return id.GetID()["id"], nil
}

// Get implements input.GameServicer.
func (service *gameService) Get(
	ctx context.Context,
	gameID uuid.UUID,
) (bo.Gamer, error) {
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
		"name":    "Get",
		"rt_ctx":  utilRuntimeContext,
		"sp_ctx":  utilSpanContext,
		"config":  service.GetConfigger(),
		"game_id": gameID,
	}

	service.GetRuntimeLogger().
		WithFields(fields).
		Info(object.URIEmpty)

	daoGamer, err := service.GetGameRepositorier().Read(
		ctx,
		dao.NewCUDID(map[string]uuid.UUID{"id": gameID}),
	)
	if err != nil {
		service.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error(object.ErrGameRepositoryRead.Error())
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, object.ErrGameRepositoryRead.Error())

		return nil, err
	}

	service.GetRuntimeLogger().
		WithFields(fields).
		WithField(object.URIFieldDAOGame, daoGamer).
		Debug(object.URIEmpty)

	boGame := bo.NewGamer(
		daoGamer.GetCUDIDer().GetID()["id"],
		daoGamer.GetName(),
		daoGamer.GetStatus(),
		daoGamer.GetCurrentPlayerID(),
		daoGamer.GetWinnerID(),
	)

	service.GetRuntimeLogger().
		WithFields(fields).
		WithField(object.URIFieldBOGame, boGame).
		Debug(object.URIEmpty)

	return boGame, nil
}

// GetGameState implements input.GameServicer.
func (service *gameService) GetGameState(
	ctx context.Context,
	gameID uuid.UUID,
) (bo.Gamer, error) {
	var traceSpan trace.Span

	ctx, traceSpan = service.GetTracer().Start(
		ctx,
		"GetGameState",
		trace.WithSpanKind(trace.SpanKindInternal),
	)
	defer traceSpan.End()

	utilRuntimeContext := util.NewRuntimeContext(ctx)
	utilSpanContext := util.NewSpanContext(traceSpan)
	fields := map[string]any{
		"name":    "GetGameState",
		"rt_ctx":  utilRuntimeContext,
		"sp_ctx":  utilSpanContext,
		"config":  service.GetConfigger(),
		"game_id": gameID,
	}

	service.GetRuntimeLogger().
		WithFields(fields).
		Info(object.URIEmpty)

	return service.Get(ctx, gameID)
}

// List implements input.GameServicer.
func (service *gameService) List(
	ctx context.Context,
	daoPaginationer dao.Paginationer,
	daoGameFilterer dao.GameFilter,
) ([]bo.Gamer, dao.Cursorer, error) {
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
		"name":              "List",
		"rt_ctx":            utilRuntimeContext,
		"sp_ctx":            utilSpanContext,
		"config":            service.GetConfigger(),
		"dao_paginationer":  daoPaginationer,
		"dao_game_filterer": daoGameFilterer,
	}

	service.GetRuntimeLogger().
		WithFields(fields).
		Info(object.URIEmpty)

	daoGamers, daoCursorer, err := service.GetGameRepositorier().
		ReadList(ctx, daoPaginationer, daoGameFilterer)
	if err != nil {
		service.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error(object.ErrGameRepositoryReadList.Error())
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, object.ErrGameRepositoryReadList.Error())

		return nil, nil, err
	}

	service.GetRuntimeLogger().
		WithFields(fields).
		WithField(object.URIFieldDAOGamers, daoGamers).
		WithField(object.URIFieldDAOCursorer, daoCursorer).
		Debug(object.URIEmpty)

	boGamers := make([]bo.Gamer, 0, len(daoGamers))

	for key, daoGamer := range daoGamers {
		service.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldKey, key).
			WithField(object.URIFieldDAOGame, daoGamer).
			Debug(object.URIEmpty)

		boGamers = append(boGamers, bo.NewGamer(
			daoGamer.GetCUDIDer().GetID()["id"],
			daoGamer.GetName(),
			daoGamer.GetStatus(),
			daoGamer.GetCurrentPlayerID(),
			daoGamer.GetWinnerID(),
		))
	}

	service.GetRuntimeLogger().
		WithFields(fields).
		WithField(object.URIFieldBOGamers, boGamers).
		Debug(object.URIEmpty)

	return boGamers, daoCursorer, nil
}

// StartGame implements input.GameServicer.
func (service *gameService) StartGame(
	ctx context.Context,
	id uuid.UUID,
) error {
	var traceSpan trace.Span

	ctx, traceSpan = service.GetTracer().Start(
		ctx,
		"StartGame",
		trace.WithSpanKind(trace.SpanKindInternal),
	)
	defer traceSpan.End()

	utilRuntimeContext := util.NewRuntimeContext(ctx)
	utilSpanContext := util.NewSpanContext(traceSpan)
	fields := map[string]any{
		"name":    "StartGame",
		"rt_ctx":  utilRuntimeContext,
		"sp_ctx":  utilSpanContext,
		"config":  service.GetConfigger(),
		"game_id": id,
	}

	service.GetRuntimeLogger().
		WithFields(fields).
		Info(object.URIEmpty)

	// Get the current game state
	game, err := service.Get(ctx, id)
	if err != nil {
		service.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error(object.ErrGameServiceGet.Error())
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, object.ErrGameServiceGet.Error())
		return err
	}

	// Create updated game with in_progress status
	updatedGame := bo.NewGamer(
		game.GetID(),
		game.GetName(),
		"in_progress",
		game.GetCurrentPlayerID(),
		game.GetWinnerID(),
	)

	// Update the game in repository
	daoGame := dao.NewGame(
		time.Time{},
		time.Time{},
		sql.NullTime{},
		updatedGame.GetID(),
		updatedGame.GetName(),
		updatedGame.GetStatus(),
		updatedGame.GetCurrentPlayerID(),
		updatedGame.GetWinnerID(),
	)

	_, err = service.GetGameRepositorier().Update(
		ctx,
		daoGame,
	)
	if err != nil {
		service.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error(object.ErrGameRepositoryUpdate.Error())
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, object.ErrGameRepositoryUpdate.Error())
		return err
	}

	return nil
}

// WithServicers implements input.GameServicer.
func (g *gameService) WithServicers(
	servicers Servicers,
) {
	g.serviceServicers = servicers
}

// WithOptioners is a function.
func (service *gameService) WithOptioners(
	optioners ...gameServiceOptioner,
) *gameService {
	newService := service.clone()
	for _, optioner := range optioners {
		optioner.apply(newService)
	}

	return newService
}

func (service *gameService) clone() *gameService {
	newService := service

	return newService
}

func (optionerFunc gameServiceOptionerFunc) apply(
	service *gameService,
) {
	optionerFunc(service)
}

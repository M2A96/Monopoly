package services

import (
	"context"
	"github/M2A96/Monopoly.git/config"
	"github/M2A96/Monopoly.git/internal/core/domain/bo"
	"github/M2A96/Monopoly.git/internal/core/ports/input"
	"github/M2A96/Monopoly.git/internal/core/ports/output"
	"github/M2A96/Monopoly.git/log"
	"github/M2A96/Monopoly.git/object"
	"github/M2A96/Monopoly.git/object/dao"
	"github/M2A96/Monopoly.git/util"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type (
	gameStateService struct {
		configConfigger                 config.Configger
		logRuntimeLogger                log.RuntimeLogger
		objectUUIDer                    object.UUIDer
		repositoryGameStateRepositorier output.GameStateRepositorier
		serviceServicers                Servicers
		traceTracer                     trace.Tracer
	}

	gameStateServiceOptioner interface {
		apply(*gameStateService)
	}

	gameStateServiceOptionerFunc func(*gameStateService)
)

var (
	_ input.GameStateServicer         = (*gameStateService)(nil)
	_ GetServicers                    = (*gameStateService)(nil)
	_ WithServicers                   = (*gameStateService)(nil)
	_ config.GetConfigger             = (*gameStateService)(nil)
	_ log.GetRuntimeLogger            = (*gameStateService)(nil)
	_ object.GetUUIDer                = (*gameStateService)(nil)
	_ output.GetGameStateRepositorier = (*gameStateService)(nil)
	_ util.GetTracer                  = (*gameStateService)(nil)
)

func NewGameStateService(
	configConfigger config.Configger,
	logRuntimeLogger log.RuntimeLogger,
	objectUUIDer object.UUIDer,
	traceTracer trace.Tracer,
	optioners ...gameStateServiceOptioner,
) *gameStateService {
	gameStateService := &gameStateService{
		configConfigger:                 configConfigger,
		logRuntimeLogger:                logRuntimeLogger,
		objectUUIDer:                    objectUUIDer,
		repositoryGameStateRepositorier: nil,
		serviceServicers:                nil,
		traceTracer:                     traceTracer,
	}

	return gameStateService.WithOptioners(optioners...)
}

func (service *gameStateService) WithServicers(
	servicers Servicers,
) {
	service.serviceServicers = servicers
}

func WithGameStateServiceGameStateRepositorier(
	repositoryGameStateRepositorier output.GameStateRepositorier,
) gameStateServiceOptioner {
	return gameStateServiceOptionerFunc(func(
		service *gameStateService,
	) {
		service.repositoryGameStateRepositorier = repositoryGameStateRepositorier
	})
}

func WithGameStateServiceServicers(
	serviceServicers Servicers,
) gameStateServiceOptioner {
	return gameStateServiceOptionerFunc(func(
		service *gameStateService,
	) {
		service.serviceServicers = serviceServicers
	})
}

func (service *gameStateService) WithOptioners(
	optioners ...gameStateServiceOptioner,
) *gameStateService {
	newService := service.clone()

	for _, optioner := range optioners {
		optioner.apply(newService)
	}

	return newService
}

func (service *gameStateService) GetTracer() trace.Tracer {
	return service.traceTracer
}

func (service *gameStateService) GetGameStateRepositorier() output.GameStateRepositorier {
	return service.repositoryGameStateRepositorier
}

func (service *gameStateService) GetUUIDer() object.UUIDer {
	return service.objectUUIDer
}

func (service *gameStateService) GetRuntimeLogger() log.RuntimeLogger {
	return service.logRuntimeLogger
}

func (service *gameStateService) GetConfigger() config.Configger {
	return service.configConfigger
}

func (service *gameStateService) GetServicers() Servicers {
	return service.serviceServicers
}

func (service *gameStateService) SetServicers(serviceServicers Servicers) {
	service.serviceServicers = serviceServicers
}

func (service *gameStateService) GetGameStateServicer() input.GameStateServicer {
	return service
}

func (service *gameStateService) GetCurrentState(
	ctx context.Context,
	gameID uuid.UUID,
) (bo.GameStater, error) {
	var traceSpan trace.Span

	ctx, traceSpan = service.GetTracer().Start(
		ctx,
		"GetCurrentState",
		trace.WithSpanKind(trace.SpanKindInternal),
	)
	defer traceSpan.End()

	utilRuntimeContext := util.NewRuntimeContext(ctx)
	utilSpanContext := util.NewSpanContext(traceSpan)
	fields := map[string]any{
		"name":    "GetCurrentState",
		"rt_ctx":  utilRuntimeContext,
		"sp_ctx":  utilSpanContext,
		"config":  service.GetConfigger(),
		"game_id": gameID,
	}

	service.GetRuntimeLogger().
		WithFields(fields).
		Info(object.URIEmpty)

	daoGameState, err := service.GetGameStateRepositorier().
		Read(
			ctx,
			dao.NewCUDID(map[string]uuid.UUID{"id": gameID}),
		)
	if err != nil {
		service.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error(object.ErrGameStateServiceGetCurrentState.Error())
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, object.ErrGameStateServiceGetCurrentState.Error())

		return nil, err
	}

	service.GetRuntimeLogger().
		WithFields(fields).
		WithField(object.URIFieldDAOGameState, daoGameState).
		Debug(object.URIEmpty)

	daoGamer := daoGameState.GetGame()
	boPlayers := make([]bo.Player, 0, len(daoGameState.GetPlayers()))
	boProperties := make([]bo.Propertyer, 0, len(daoGameState.GetProperties()))

	boGame := bo.NewGamer(
		daoGamer.GetCUDIDer().GetID()["id"],
		daoGamer.GetName(),
		daoGamer.GetStatus(),
		daoGamer.GetCurrentPlayerID(),
		daoGamer.GetWinnerID(),
	)

	for key, daoPlayer := range daoGameState.GetPlayers() {
		service.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldKey, key).
			WithField(object.URIFieldDAOPlayer, daoPlayer).
			Debug(object.URIEmpty)

		boPlayers = append(boPlayers, bo.NewPlayer(
			daoPlayer.GetCUDIDer().GetID()["id"],
			daoPlayer.GetGameID(),
			daoPlayer.GetName(),
			daoPlayer.GetBalance(),
			daoPlayer.GetPosition(),
			daoPlayer.GetInJail(),
			daoPlayer.GetJailTurns(),
			daoPlayer.GetBankrupt(),
		))
	}

	for key, daoProperty := range daoGameState.GetProperties() {
		service.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldKey, key).
			WithField(object.URIFieldDAOProperty, daoProperty).
			Debug(object.URIEmpty)

		boProperties = append(boProperties, bo.NewProperty(
			daoProperty.GetCUDIDer().GetID()["id"],
			daoProperty.GetName(),
			daoProperty.GetColorGroup(),
			daoProperty.GetPrice(),
			daoProperty.GetHousePrice(),
			daoProperty.GetHotelPrice(),
			daoProperty.GetRent(),
			daoProperty.GetRentWithColorSet(),
			daoProperty.GetRentWith1House(),
			daoProperty.GetRentWith2Houses(),
			daoProperty.GetRentWith3Houses(),
			daoProperty.GetRentWith4Houses(),
			daoProperty.GetRentWithHotel(),
			daoProperty.GetMortgageValue(),
			daoProperty.GetOwnerID(),
			daoProperty.GetHouses(),
			daoProperty.GetHasHotel(),
			daoProperty.GetMortgaged(),
		))
	}

	boGameState := bo.NewGameState(
		boGame,
		boPlayers,
		boProperties,
	)

	service.GetRuntimeLogger().
		WithFields(fields).
		WithField(object.URIFieldBOGame, boGame).
		Debug(object.URIEmpty)

	return boGameState, nil
}

func (service *gameStateService) GetGameStateHistory(
	ctx context.Context,
	daoGameStateFilter dao.GameStateFilter,
	daoPagination dao.Paginationer,
) ([]bo.GameStater, dao.Cursorer, error) {
	var traceSpan trace.Span

	ctx, traceSpan = service.GetTracer().Start(
		ctx,
		"GetGameStateHistory",
		trace.WithSpanKind(trace.SpanKindInternal),
	)
	defer traceSpan.End()

	utilRuntimeContext := util.NewRuntimeContext(ctx)
	utilSpanContext := util.NewSpanContext(traceSpan)
	fields := map[string]any{
		"name":                  "GetGameStateHistory",
		"rt_ctx":                utilRuntimeContext,
		"sp_ctx":                utilSpanContext,
		"config":                service.GetConfigger(),
		"dao_game_state_filter": daoGameStateFilter,
		"dao_pagination":        daoPagination,
	}

	service.GetRuntimeLogger().
		WithFields(fields).
		Info(object.URIEmpty)

	daoGameStates, daoCursor, err := service.GetGameStateRepositorier().
		ReadList(
			ctx,
			daoPagination,
			daoGameStateFilter,
		)
	if err != nil {
		service.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error(object.ErrGameStateServiceGetGameStateHistory.Error())
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, object.ErrGameStateServiceGetGameStateHistory.Error())

		return nil, daoCursor, err
	}

	boGameStates := make([]bo.GameStater, 0, len(daoGameStates))

	for key, daoGameState := range daoGameStates {
		service.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldKey, key).
			WithField(object.URIFieldDAOGameState, daoGameState).
			Debug(object.URIEmpty)

		boPlayers := make([]bo.Player, 0, len(daoGameState.GetPlayers()))
		boProperties := make([]bo.Propertyer, 0, len(daoGameState.GetProperties()))
		daoGame := daoGameState.GetGame()

		boGame := bo.NewGamer(
			daoGame.GetCUDIDer().GetID()["id"],
			daoGameState.GetGame().GetName(),
			daoGameState.GetGame().GetStatus(),
			daoGameState.GetGame().GetCurrentPlayerID(),
			daoGameState.GetGame().GetWinnerID(),
		)

		for key, daoPlayer := range daoGameState.GetPlayers() {
			service.GetRuntimeLogger().
				WithFields(fields).
				WithField(object.URIFieldKey, key).
				WithField(object.URIFieldDAOPlayer, daoPlayer).
				Debug(object.URIEmpty)

			boPlayers = append(boPlayers, bo.NewPlayer(
				daoPlayer.GetCUDIDer().GetID()["id"],
				daoPlayer.GetGameID(),
				daoPlayer.GetName(),
				daoPlayer.GetBalance(),
				daoPlayer.GetPosition(),
				daoPlayer.GetInJail(),
				daoPlayer.GetJailTurns(),
				daoPlayer.GetBankrupt(),
			))
		}

		for key, daoProperty := range daoGameState.GetProperties() {
			service.GetRuntimeLogger().
				WithFields(fields).
				WithField(object.URIFieldKey, key).
				WithField(object.URIFieldDAOProperty, daoProperty).
				Debug(object.URIEmpty)

			boProperties = append(boProperties, bo.NewProperty(
				daoProperty.GetCUDIDer().GetID()["id"],
				daoProperty.GetName(),
				daoProperty.GetColorGroup(),
				daoProperty.GetPrice(),
				daoProperty.GetHousePrice(),
				daoProperty.GetHotelPrice(),
				daoProperty.GetRent(),
				daoProperty.GetRentWithColorSet(),
				daoProperty.GetRentWith1House(),
				daoProperty.GetRentWith2Houses(),
				daoProperty.GetRentWith3Houses(),
				daoProperty.GetRentWith4Houses(),
				daoProperty.GetRentWithHotel(),
				daoProperty.GetMortgageValue(),
				daoProperty.GetOwnerID(),
				daoProperty.GetHouses(),
				daoProperty.GetHasHotel(),
				daoProperty.GetMortgaged(),
			))
		}

		boGameStates = append(boGameStates, bo.NewGameState(
			boGame,
			boPlayers,
			boProperties,
		))
	}

	return boGameStates, daoCursor, nil
}

func (service *gameStateService) clone() *gameStateService {
	newService := *service

	return &newService
}

func (optionerFunc gameStateServiceOptionerFunc) apply(
	service *gameStateService,
) {
	optionerFunc(service)
}

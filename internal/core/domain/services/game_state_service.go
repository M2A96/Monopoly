// internal/core/domain/services/game_state_service.go
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

// GameStateService implements the input.GameStateServicer interface
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

// NewGameStateService creates a new game state service instance
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

// WithServicers implements WithServicers.
func (service *gameStateService) WithServicers(
	servicers Servicers,
) {
	service.serviceServicers = servicers
}

// WithGameStateServiceGameStateRepositorier sets the game state repository for the service
func WithGameStateServiceGameStateRepositorier(
	repositoryGameStateRepositorier output.GameStateRepositorier,
) gameStateServiceOptioner {
	return gameStateServiceOptionerFunc(func(
		service *gameStateService,
	) {
		service.repositoryGameStateRepositorier = repositoryGameStateRepositorier
	})
}

// WithGameStateServiceServicers sets the servicers for the service
func WithGameStateServiceServicers(
	serviceServicers Servicers,
) gameStateServiceOptioner {
	return gameStateServiceOptionerFunc(func(
		service *gameStateService,
	) {
		service.serviceServicers = serviceServicers
	})
}

// WithOptioners applies the optioners to the service
func (service *gameStateService) WithOptioners(
	optioners ...gameStateServiceOptioner,
) *gameStateService {
	newService := service.clone()

	for _, optioner := range optioners {
		optioner.apply(newService)
	}

	return newService
}

// GetTracer implements util.GetTracer
func (service *gameStateService) GetTracer() trace.Tracer {
	return service.traceTracer
}

// GetGameStateRepositorier implements output.GetGameStateRepositorier
func (service *gameStateService) GetGameStateRepositorier() output.GameStateRepositorier {
	return service.repositoryGameStateRepositorier
}

// GetUUIDer implements object.GetUUIDer
func (service *gameStateService) GetUUIDer() object.UUIDer {
	return service.objectUUIDer
}

// GetRuntimeLogger implements log.GetRuntimeLogger
func (service *gameStateService) GetRuntimeLogger() log.RuntimeLogger {
	return service.logRuntimeLogger
}

// GetConfigger implements config.GetConfigger
func (service *gameStateService) GetConfigger() config.Configger {
	return service.configConfigger
}

// GetServicers implements GetServicers
func (service *gameStateService) GetServicers() Servicers {
	return service.serviceServicers
}

// SetServicers implements WithServicers
func (service *gameStateService) SetServicers(serviceServicers Servicers) {
	service.serviceServicers = serviceServicers
}

// GetGameStateServicer implements input.GetGameStateServicer
func (service *gameStateService) GetGameStateServicer() input.GameStateServicer {
	return service
}

// GetCurrentState retrieves the current state of a game
func (service *gameStateService) GetCurrentState(
	ctx context.Context,
	gameID uuid.UUID,
) (bo.GameStater, error) {
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

	// Get the latest game state from the repository
	// The service should not be creating these objects - they should be passed in or handled by a factory
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

	service.GetRuntimeLogger().
		WithFields(fields).
		WithField(object.URIFieldBOGame, boGame).
		Debug(object.URIEmpty)

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

// SaveGameState saves the current state of a game
func (service *gameStateService) SaveGameState(
	ctx context.Context,
	boGameState bo.GameStater,
) error {
	var traceSpan trace.Span

	ctx, traceSpan = service.GetTracer().Start(
		ctx,
		"SaveGameState",
		trace.WithSpanKind(trace.SpanKindInternal),
	)
	defer traceSpan.End()

	utilRuntimeContext := util.NewRuntimeContext(ctx)
	utilSpanContext := util.NewSpanContext(traceSpan)
	fields := map[string]any{
		"name":          "SaveGameState",
		"rt_ctx":        utilRuntimeContext,
		"sp_ctx":        utilSpanContext,
		"config":        service.GetConfigger(),
		"bo_game_state": boGameState,
	}

	service.GetRuntimeLogger().
		WithFields(fields).
		Info(object.URIEmpty)

	boGame := boGameState.GetGame()
	daoGame := dao.NewGame(
		time.Time{},
		time.Time{},
		sql.NullTime{
			Time:  time.Time{},
			Valid: false,
		},
		boGame.GetID(),
		boGame.GetName(),
		boGame.GetStatus(),
		boGame.GetCurrentPlayerID(),
		boGame.GetWinnerID(),
	)

	service.GetRuntimeLogger().
		WithFields(fields).
		WithField(object.URIFieldBOGame, boGame).
		Debug(object.URIEmpty)

	daoPlayers := make([]dao.Player, 0, len(boGameState.GetPlayers()))

	for key, boPlayer := range boGameState.GetPlayers() {
		service.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldKey, key).
			WithField(object.URIFieldBOPlayer, boPlayer).
			Debug(object.URIEmpty)

		daoPlayers = append(daoPlayers, dao.NewPlayer(
			uuid.Nil,
			boPlayer.GetGameID(),
			boPlayer.GetName(),
			boPlayer.GetBalance(),
			boPlayer.GetBalance(),
			boPlayer.GetInJail(),
			boPlayer.GetJailTurns(),
			boPlayer.GetBankrupt(),
			time.Time{},
			time.Time{},
			sql.NullTime{
				Time:  time.Time{},
				Valid: false,
			},
		))
	}

	daoProperties := make([]dao.Propertyer, 0, len(boGameState.GetProperties()))
	for key, boProperty := range boGameState.GetProperties() {
		service.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldKey, key).
			WithField(object.URIFieldBOProperty, boProperty).
			Debug(object.URIEmpty)

		daoProperties = append(daoProperties, dao.NewProperty(
			uuid.Nil,
			boProperty.GetName(),
			boProperty.GetColorGroup(),
			boProperty.GetPrice(),
			boProperty.GetHousePrice(),
			boProperty.GetHotelPrice(),
			boProperty.GetRent(),
			boProperty.GetRentWithColorSet(),
			boProperty.GetRentWith1House(),
			boProperty.GetRentWith2Houses(),
			boProperty.GetRentWith3Houses(),
			boProperty.GetRentWith4Houses(),
			boProperty.GetRentWithHotel(),
			boProperty.GetMortgageValue(),
			boProperty.GetOwnerID(),
			boProperty.GetHouses(),
			boProperty.GetHasHotel(),
			boProperty.GetMortgaged(),
			time.Time{},
			time.Time{},
			sql.NullTime{
				Time:  time.Time{},
				Valid: false,
			},
		))
	}

	daoGameState := dao.NewGameState(
		daoGame,
		daoPlayers,
		daoProperties,
	)

	service.GetRuntimeLogger().
		WithFields(fields).
		WithField(object.URIFieldDAOGameState, daoGameState).
		Debug(object.URIEmpty)

	id, err := service.GetGameStateRepositorier().
		Create(
			ctx,
			daoGameState,
		)
	if err != nil {
		service.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error(object.ErrGameStateServiceSaveGameState.Error())
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, object.ErrGameRepositoryCreate.Error())
		return err
	}
	service.GetRuntimeLogger().
		WithFields(fields).
		WithField(object.URIFieldID, id).
		Debug(object.URIEmpty)

	return nil
}

// GetGameStateHistory retrieves the history of game states
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
		"name":                  "SaveGameState",
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

// RestoreGameState restores a game to a previous state
// RestoreGameState restores a game to a previous state
func (service *gameStateService) RestoreGameState(
	ctx context.Context,
	gameID uuid.UUID,
	stateID uuid.UUID,
) error {
	var traceSpan trace.Span

	ctx, traceSpan = service.GetTracer().Start(
		ctx,
		"RestoreGameState",
		trace.WithSpanKind(trace.SpanKindInternal),
	)
	defer traceSpan.End()

	utilRuntimeContext := util.NewRuntimeContext(ctx)
	utilSpanContext := util.NewSpanContext(traceSpan)
	fields := map[string]any{
		"name":     "RestoreGameState",
		"rt_ctx":   utilRuntimeContext,
		"sp_ctx":   utilSpanContext,
		"config":   service.GetConfigger(),
		"game_id":  gameID,
		"state_id": stateID,
	}

	service.GetRuntimeLogger().
		WithFields(fields).
		Info(object.URIEmpty)

	// Get the specific game state
	daoGameState, err := service.GetGameStateRepositorier().Read(
		ctx,
		dao.NewCUDID(map[string]uuid.UUID{"id": stateID}),
	)
	if err != nil {
		service.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error("Failed to retrieve game state")
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, "Failed to retrieve game state")

		return err
	}

	service.GetRuntimeLogger().
		WithFields(fields).
		WithField(object.URIFieldDAOGameState, daoGameState).
		Debug(object.URIEmpty)

	// Verify that the state belongs to the specified game
	daoGame := daoGameState.GetGame()
	if daoGame.GetCUDIDer().GetID()["id"] != gameID {
		err := sql.ErrNoRows
		service.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			WithField("state_game_id", daoGame.GetCUDIDer().GetID()["id"]).
			Error("Game state does not belong to the specified game")
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, "Game state does not belong to the specified game")

		return err
	}

	// Convert DAO to BO objects
	boPlayers := make([]bo.Player, 0, len(daoGameState.GetPlayers()))
	boProperties := make([]bo.Propertyer, 0, len(daoGameState.GetProperties()))

	boGame := bo.NewGamer(
		daoGame.GetCUDIDer().GetID()["id"],
		daoGame.GetName(),
		daoGame.GetStatus(),
		daoGame.GetCurrentPlayerID(),
		daoGame.GetWinnerID(),
	)

	service.GetRuntimeLogger().
		WithFields(fields).
		WithField(object.URIFieldBOGame, boGame).
		Debug(object.URIEmpty)

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

	// Create a new game state with the restored state
	boGameState := bo.NewGameState(
		boGame,
		boPlayers,
		boProperties,
	)

	service.GetRuntimeLogger().
		WithFields(fields).
		WithField(object.URIFieldBOGameState, boGameState).
		Debug(object.URIEmpty)

	// Update the current game with the restored state
	err = service.SaveGameState(ctx, boGameState)
	if err != nil {
		service.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error("Failed to save restored game state")
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, "Failed to save restored game state")

		return err
	}

	service.GetRuntimeLogger().
		WithFields(fields).
		Info("Game state restored successfully")

	return nil
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

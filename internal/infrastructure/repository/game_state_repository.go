package repository

import (
	"context"
	"fmt"
	"github/M2A96/Monopoly.git/config"
	"github/M2A96/Monopoly.git/internal/core/ports/output"
	"github/M2A96/Monopoly.git/log"
	"github/M2A96/Monopoly.git/object"
	"github/M2A96/Monopoly.git/object/dao"
	"github/M2A96/Monopoly.git/util"
	"time"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
)

type (
	gameStateRepository struct {
		repository
	}

	gameStateRepositoryOptioner = repositoryOptioner
)

// Ensure gameStateRepository implements output.GameStateRepositorier
var _ output.GameStateRepositorier = (*gameStateRepository)(nil)

// NewGameStateRepository creates a new game state repository instance
func NewGameStateRepository(
	configConfigger config.Configger,
	logRuntimeLogger log.RuntimeLogger,
	objectUUIDer object.UUIDer,
	traceTracer trace.Tracer,
	optioners ...gameStateRepositoryOptioner,
) *gameStateRepository {
	return &gameStateRepository{
		repository: *NewRepository(
			configConfigger,
			logRuntimeLogger,
			objectUUIDer,
			traceTracer,
			optioners...,
		),
	}
}

// WithGameStateRepositoryTimer adds a timer to the game state repository
func WithGameStateRepositoryTimer(
	objectTimer object.Timer,
) gameStateRepositoryOptioner {
	return WithRepositoryTimer(objectTimer)
}

// WithGameStateRepositoryDB adds a database connection to the game state repository
func WithGameStateRepositoryDB(
	gormDB *gorm.DB,
) gameStateRepositoryOptioner {
	return WithRepositoryDB(gormDB, object.URITableGameState)
}

// Create implements the Create method of the GameStateRepositorier interface
func (repository *gameStateRepository) Create(
	ctx context.Context,
	daoGameState dao.GameStater,
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
		"name":           "Create",
		"rt_ctx":         utilRuntimeContext,
		"sp_ctx":         utilSpanContext,
		"config":         repository.GetConfigger(),
		"dao_game_state": daoGameState,
	}

	repository.GetRuntimeLogger().
		WithFields(fields).
		Info(object.URIEmpty)

	// Game state is a composite object, so we need to handle it differently
	// We'll return the game's ID as the identifier for the game state
	gameID := daoGameState.GetGame().GetCUDIDer()

	repository.GetRuntimeLogger().
		WithFields(fields).
		WithField(object.URIFieldID, gameID).
		Debug(object.URIEmpty)

	return gameID, nil
}

// Read implements the Read method of the GameStateRepositorier interface
func (repository *gameStateRepository) Read(
	ctx context.Context,
	id dao.CUDIDer,
) (dao.GameStater, error) {
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

	// For game state, we need to fetch the game, players, and properties
	// First, fetch the game
	gameResult := map[string]any{}
	gormDB := repository.GetDB().
		WithContext(ctx).
		Where(map[string]any{
			"id":         id.GetID()["id"],
			"deleted_at": nil,
		}).
		Select(fmt.Sprintf(`%s.*`, object.URITableGame)).
		Find(&gameResult)
	if err := gormDB.Error; err != nil {
		repository.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error("Failed to read game for game state")
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, "Failed to read game for game state")

		return nil, err
	}

	if len(gameResult) == 0 {
		repository.GetRuntimeLogger().
			WithFields(fields).
			Error("Game not found for game state")
		traceSpan.SetStatus(codes.Error, "Game not found for game state")

		return nil, fmt.Errorf("game not found for game state")
	}

	game, err := dao.NewGamerFromMap(
		repository.GetUUIDer(),
		gameResult,
	)
	if err != nil {
		repository.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error("Failed to create game from map")
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, "Failed to create game from map")

		return nil, err
	}

	// Next, fetch the players for this game
	playerResults := make([]map[string]any, 0)
	gormDB = repository.GetDB().
		WithContext(ctx).
		Where(map[string]any{
			"game_id":    id.GetID()["id"],
			"deleted_at": nil,
		}).
		Select(fmt.Sprintf(`%s.*`, object.URITablePlayer)).
		Find(&playerResults)
	if err := gormDB.Error; err != nil {
		repository.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error("Failed to read players for game state")
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, "Failed to read players for game state")

		return nil, err
	}

	players := make([]dao.Player, 0, len(playerResults))
	for _, playerResult := range playerResults {
		player, err := dao.NewPlayerFromMap(repository.GetUUIDer(), playerResult)
		if err != nil {
			repository.GetRuntimeLogger().
				WithFields(fields).
				WithField(object.URIFieldError, err).
				Error("Failed to create player from map")
			traceSpan.RecordError(err)
			traceSpan.SetStatus(codes.Error, "Failed to create player from map")

			return nil, err
		}
		players = append(players, player)
	}

	// Finally, fetch the properties for this game
	propertyResults := make([]map[string]any, 0)
	gormDB = repository.GetDB().
		WithContext(ctx).
		Where(map[string]any{
			"game_id":    id.GetID()["id"],
			"deleted_at": nil,
		}).
		Select(fmt.Sprintf(`%s.*`, object.URITableProperty)).
		Find(&propertyResults)
	if err := gormDB.Error; err != nil {
		repository.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error("Failed to read properties for game state")
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, "Failed to read properties for game state")

		return nil, err
	}

	properties := make([]dao.Propertyer, 0, len(propertyResults))
	for _, propertyResult := range propertyResults {
		property, err := dao.NewPropertyFromMap(repository.GetUUIDer(), propertyResult)
		if err != nil {
			repository.GetRuntimeLogger().
				WithFields(fields).
				WithField(object.URIFieldError, err).
				Error("Failed to create property from map")
			traceSpan.RecordError(err)
			traceSpan.SetStatus(codes.Error, "Failed to create property from map")

			return nil, err
		}
		properties = append(properties, property)
	}

	// Create the game state with all components
	gameState := dao.NewGameState(game, players, properties)

	return gameState, nil
}

// ReadList implements the ReadList method of the GameStateRepositorier interface
func (repository *gameStateRepository) ReadList(
	ctx context.Context,
	pagination dao.Paginationer,
	filter dao.GameStateFilter,
) ([]dao.GameStater, dao.Cursorer, error) {
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

	// For game states, we need to fetch games first, then for each game fetch its players and properties
	// First, fetch the games with pagination and filtering
	gameResults := make([]map[string]any, 0, pagination.GetLimit()+1)
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
		Find(&gameResults)
	if err := gormDB.Error; err != nil {
		repository.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error("Failed to read games for game state list")
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, "Failed to read games for game state list")

		return nil, nil, err
	}

	repository.GetRuntimeLogger().
		WithFields(fields).
		WithField(object.URIFieldResult, gameResults).
		Debug(object.URIEmpty)

	// Process only up to the limit
	actualLimit := pagination.GetLimit()
	if uint32(len(gameResults)) > actualLimit {
		gameResults = gameResults[:actualLimit]
	}

	// Create game states for each game
	daoGameStates := make([]dao.GameStater, 0, len(gameResults))
	for _, gameResult := range gameResults {
		// Create the game
		game, err := dao.NewGamerFromMap(repository.GetUUIDer(), gameResult)
		if err != nil {
			repository.GetRuntimeLogger().
				WithFields(fields).
				WithField(object.URIFieldError, err).
				Error("Failed to create game from map")
			traceSpan.RecordError(err)
			traceSpan.SetStatus(codes.Error, "Failed to create game from map")

			return nil, nil, err
		}

		// Fetch players for this game
		playerResults := make([]map[string]any, 0)
		gormDB = repository.GetDB().
			WithContext(ctx).
			Where(map[string]any{
				"game_id":    gameResult["id"],
				"deleted_at": nil,
			}).
			Select(fmt.Sprintf(`%s.*`, object.URITablePlayer)).
			Find(&playerResults)
		if err := gormDB.Error; err != nil {
			repository.GetRuntimeLogger().
				WithFields(fields).
				WithField(object.URIFieldError, err).
				Error("Failed to read players for game state")
			traceSpan.RecordError(err)
			traceSpan.SetStatus(codes.Error, "Failed to read players for game state")

			return nil, nil, err
		}

		players := make([]dao.Player, 0, len(playerResults))
		for _, playerResult := range playerResults {
			player, err := dao.NewPlayerFromMap(repository.GetUUIDer(), playerResult)
			if err != nil {
				repository.GetRuntimeLogger().
					WithFields(fields).
					WithField(object.URIFieldError, err).
					Error("Failed to create player from map")
				traceSpan.RecordError(err)
				traceSpan.SetStatus(codes.Error, "Failed to create player from map")

				return nil, nil, err
			}
			players = append(players, player)
		}

		// Fetch properties for this game
		propertyResults := make([]map[string]any, 0)
		gormDB = repository.GetDB().
			WithContext(ctx).
			Where(map[string]any{
				"game_id":    gameResult["id"],
				"deleted_at": nil,
			}).
			Select(fmt.Sprintf(`%s.*`, object.URITableProperty)).
			Find(&propertyResults)
		if err := gormDB.Error; err != nil {
			repository.GetRuntimeLogger().
				WithFields(fields).
				WithField(object.URIFieldError, err).
				Error("Failed to read properties for game state")
			traceSpan.RecordError(err)
			traceSpan.SetStatus(codes.Error, "Failed to read properties for game state")

			return nil, nil, err
		}

		properties := make([]dao.Propertyer, 0, len(propertyResults))
		for _, propertyResult := range propertyResults {
			property, err := dao.NewPropertyFromMap(repository.GetUUIDer(), propertyResult)
			if err != nil {
				repository.GetRuntimeLogger().
					WithFields(fields).
					WithField(object.URIFieldError, err).
					Error("Failed to create property from map")
				traceSpan.RecordError(err)
				traceSpan.SetStatus(codes.Error, "Failed to create property from map")

				return nil, nil, err
			}
			properties = append(properties, property)
		}

		// Create the game state
		gameState := dao.NewGameState(game, players, properties)
		daoGameStates = append(daoGameStates, gameState)
	}

	repository.GetRuntimeLogger().
		WithFields(fields).
		WithField("dao_game_states", daoGameStates).
		Debug(object.URIEmpty)

	// Handle pagination
	var daoCursorer dao.Cursorer
	if uint32(len(gameResults)) > pagination.GetLimit() {
		repository.GetRuntimeLogger().
			WithFields(fields).
			Debug(`uint32(len(gameResults)) > pagination.GetLimit()`)

		daoCursorer = dao.NewCursor(
			pagination.GetCursorer().GetOffset() + pagination.GetLimit(),
		)
	}

	repository.GetRuntimeLogger().
		WithFields(fields).
		WithField(object.URIFieldDAOCursorer, daoCursorer).
		Debug(object.URIEmpty)

	return daoGameStates, daoCursorer, nil
}

// Update implements the Update method of the GameStateRepositorier interface
// Note: Game state is a composite object, so update will update the game, players, and properties
func (repository *gameStateRepository) Update(
	ctx context.Context,
	gameState dao.GameStater,
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
		"name":           "Update",
		"rt_ctx":         utilRuntimeContext,
		"sp_ctx":         utilSpanContext,
		"config":         repository.GetConfigger(),
		"dao_game_state": gameState,
	}

	repository.GetRuntimeLogger().
		WithFields(fields).
		Info(object.URIEmpty)

	// For game state update, we need to update the game
	// The players and properties should be updated through their respective repositories
	// Here we'll just update the game
	nowUTC := repository.GetTimer().NowUTC()

	repository.GetRuntimeLogger().
		WithFields(fields).
		WithField(object.URIFieldNowUTC, nowUTC).
		Debug(object.URIEmpty)

	// Update the game
	game := gameState.GetGame()
	gormDB := repository.GetDB().
		WithContext(ctx).
		Where(map[string]any{
			"id":         game.GetCUDIDer().GetID()["id"],
			"deleted_at": nil,
		}).
		Updates(map[string]any{
			"name":              game.GetName(),
			"status":            game.GetStatus(),
			"current_player_id": game.GetCurrentPlayerID(),
			"winner_id":         game.GetWinnerID(),
			"updated_at":        nowUTC,
		})
	if err := gormDB.Error; err != nil {
		repository.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error("Failed to update game for game state")
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, "Failed to update game for game state")

		return time.Time{}, err
	}

	if gormDB.RowsAffected == 0 {
		repository.GetRuntimeLogger().
			WithFields(fields).
			Error("Game not found for game state update")
		traceSpan.SetStatus(codes.Error, "Game not found for game state update")

		return time.Time{}, fmt.Errorf("game not found for game state update")
	}

	return nowUTC, nil
}

// Delete implements the Delete method of the GameStateRepositorier interface
// Note: Game state is a composite object, so delete will mark the game as deleted
func (repository *gameStateRepository) Delete(
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

	// For game state delete, we'll mark the game as deleted
	nowUTC := repository.GetTimer().NowUTC()

	repository.GetRuntimeLogger().
		WithFields(fields).
		WithField(object.URIFieldNowUTC, nowUTC).
		Debug(object.URIEmpty)

	// Mark the game as deleted
	gormDB := repository.GetDB().
		WithContext(ctx).
		Where(map[string]any{
			"id":         id.GetID()["id"],
			"deleted_at": nil,
		}).
		Updates(map[string]any{
			"deleted_at": nowUTC,
		})
	if err := gormDB.Error; err != nil {
		repository.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error("Failed to delete game for game state")
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, "Failed to delete game for game state")

		return time.Time{}, err
	}

	if gormDB.RowsAffected == 0 {
		repository.GetRuntimeLogger().
			WithFields(fields).
			Error("Game not found for game state delete")
		traceSpan.SetStatus(codes.Error, "Game not found for game state delete")

		return time.Time{}, fmt.Errorf("game not found for game state delete")
	}

	return nowUTC, nil
}

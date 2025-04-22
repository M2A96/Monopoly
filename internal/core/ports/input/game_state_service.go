// internal/ports/input/game_state_service.go
package input

import (
	"context"
	"github/M2A96/Monopoly.git/internal/core/domain/bo"
	"github/M2A96/Monopoly.git/object/dao"

	"github.com/google/uuid"
)

// GameStateService defines the interface for game state operations
type (
	GameStateServicer interface {
		// GetCurrentState retrieves the current state of a game
		GetCurrentState(
			context.Context,
			uuid.UUID,
		) (bo.GameStater, error)

		// SaveGameState saves the current state of a game
		SaveGameState(
			context.Context,
			bo.GameStater,
		) error

		// GetGameStateHistory retrieves the history of game states
		GetGameStateHistory(
			context.Context,
			uuid.UUID,
			dao.Paginationer,
		) ([]bo.GameStater, dao.Cursorer, error)

		// RestoreGameState restores a game to a previous state
		RestoreGameState(
			context.Context,
			uuid.UUID,
			uuid.UUID, // state ID
		) error
	}

	// GetGameStateServicer is an interface for retrieving the game state service
	GetGameStateServicer interface {
		// GetGameStateServicer returns the game state service implementation
		GetGameStateServicer() GameStateServicer
	}
)

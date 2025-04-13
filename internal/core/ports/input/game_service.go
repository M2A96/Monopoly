// internal/ports/input/game_service.go
package input

import (
	"context"
	"github/M2A96/Monopoly.git/internal/core/domain/bo"
	"github/M2A96/Monopoly.git/object/dao"

	"github.com/google/uuid"
)

// GameService defines the interface for game operations
type (
	GameServicer interface {
		// Get is a function.
		Get(
			context.Context,
			uuid.UUID,
		) (bo.Gamer, error)

		// List is a function.
		List(
			context.Context,
			dao.Paginationer,
			dao.GameFilter,
		) ([]bo.Gamer, dao.Cursorer, error)

		// CreateGame creates a new game with the given name
		CreateGame(
			context.Context,
			bo.Gamer,
		) (uuid.UUID, error)

		// StartGame changes the status of a game to "in_progress"
		StartGame(
			context.Context,
			uuid.UUID,
		) error

		// GetGameState retrieves the current state of a game
		GetGameState(
			context context.Context,
			id uuid.UUID,
		) (bo.Gamer, error)
	}

	// GetGameServicer is an interface.
	GetGameServicer interface {
		// GetGameServicer is a function.
		GetGameServicer() GameServicer
	}
)

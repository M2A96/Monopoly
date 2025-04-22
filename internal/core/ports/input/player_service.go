// internal/ports/input/player_service.go
package input

import (
	"context"
	"github/M2A96/Monopoly.git/internal/core/domain/bo"
	"github/M2A96/Monopoly.git/object/dao"

	"github.com/google/uuid"
)

// PlayerService defines the interface for player operations
type (
	PlayerServicer interface {
		// Get retrieves a player by ID
		Get(
			context.Context,
			uuid.UUID,
		) (bo.Player, error)

		// List retrieves a list of players with pagination
		List(
			context.Context,
			dao.Paginationer,
			dao.PlayerFilter,
		) ([]bo.Player, dao.Cursorer, error)

		// CreatePlayer creates a new player
		CreatePlayer(
			context.Context,
			bo.Player,
		) (uuid.UUID, error)

		// UpdatePlayer updates an existing player
		UpdatePlayer(
			context.Context,
			uuid.UUID,
			bo.Player,
		) error

		// DeletePlayer removes a player from the game
		DeletePlayer(
			context.Context,
			uuid.UUID,
		) error
	}

	// GetPlayerServicer is an interface for retrieving the player service
	GetPlayerServicer interface {
		// GetPlayerServicer returns the player service implementation
		GetPlayerServicer() PlayerServicer
	}
)

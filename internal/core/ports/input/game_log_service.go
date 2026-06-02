// internal/ports/input/game_log_service.go
package input

import (
	"context"
	"github/M2A96/Monopoly.git/internal/core/domain/bo"
	"github/M2A96/Monopoly.git/object/dao"
	"time"

	"github.com/google/uuid"
)

// GameLogService defines the interface for game log operations
type (
	GameLogServicer interface {
		// Get retrieves a game log by ID
		Get(
			context.Context,
			uuid.UUID,
		) (bo.GameLogger, error)

		// List retrieves a list of game logs with pagination
		List(
			context.Context,
			dao.Paginationer,
			dao.GameLogFilter,
		) ([]bo.GameLogger, dao.Cursorer, error)

		// CreateGameLog creates a new game log entry
		CreateGameLog(
			context.Context,
			bo.GameLogger,
		) (uuid.UUID, error)

		// GetGameLogsByGameID retrieves all logs for a specific game
		GetGameLogsByGameID(
			context.Context,
			uuid.UUID,
			dao.Paginationer,
		) ([]bo.GameLogger, dao.Cursorer, error)

		// GetGameLogsByPlayerID retrieves all logs for a specific player
		GetGameLogsByPlayerID(
			context.Context,
			uuid.UUID,
			dao.Paginationer,
		) ([]bo.GameLogger, dao.Cursorer, error)

		// GetGameLogsByTimeRange retrieves logs within a specific time range
		GetGameLogsByTimeRange(
			context.Context,
			uuid.UUID,
			time.Time,
			time.Time,
			dao.Paginationer,
		) ([]bo.GameLogger, dao.Cursorer, error)
	}

	// GetGameLogServicer is an interface for retrieving the game log service
	GetGameLogServicer interface {
		// GetGameLogServicer returns the game log service implementation
		GetGameLogServicer() GameLogServicer
	}
)

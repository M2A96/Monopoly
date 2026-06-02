package input

import (
	"context"
	"github/M2A96/Monopoly.git/internal/core/domain/bo"
	"github/M2A96/Monopoly.git/object/dao"

	"github.com/google/uuid"
)

type (
	GameStateServicer interface {
		GetCurrentState(
			context.Context,
			uuid.UUID,
		) (bo.GameStater, error)

		GetGameStateHistory(
			context.Context,
			dao.GameStateFilter,
			dao.Paginationer,
		) ([]bo.GameStater, dao.Cursorer, error)
	}

	GetGameStateServicer interface {
		GetGameStateServicer() GameStateServicer
	}
)

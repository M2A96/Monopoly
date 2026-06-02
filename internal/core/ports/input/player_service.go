package input

import (
	"context"
	"github/M2A96/Monopoly.git/internal/core/domain/bo"
	"github/M2A96/Monopoly.git/object/dao"

	"github.com/google/uuid"
)

type (
	PlayerServicer interface {
		Get(
			context.Context,
			uuid.UUID,
		) (bo.Player, error)

		List(
			context.Context,
			dao.Paginationer,
			dao.PlayerFilter,
		) ([]bo.Player, dao.Cursorer, error)

		CreatePlayer(
			context.Context,
			bo.Player,
		) (uuid.UUID, error)
	}

	GetPlayerServicer interface {
		GetPlayerServicer() PlayerServicer
	}
)

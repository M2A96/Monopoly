package input

import (
	"context"
	"github/M2A96/Monopoly.git/internal/core/domain/bo"
	"github/M2A96/Monopoly.git/object/dao"

	"github.com/google/uuid"
)

type (
	TradeRequestServicer interface {
		Get(
			context.Context,
			uuid.UUID,
		) (bo.TradeRequester, error)

		List(
			context.Context,
			dao.Paginationer,
			dao.TradeFilter,
		) ([]bo.TradeRequester, dao.Cursorer, error)

		CreateTradeRequest(
			context.Context,
			bo.TradeRequester,
		) (uuid.UUID, error)

		GetTradeRequestsByGameID(
			context.Context,
			uuid.UUID,
			dao.Paginationer,
		) ([]bo.TradeRequester, dao.Cursorer, error)

		GetTradeRequestsByPlayerID(
			context.Context,
			uuid.UUID,
			dao.Paginationer,
		) ([]bo.TradeRequester, dao.Cursorer, error)
	}

	GetTradeRequestServicer interface {
		GetTradeRequestServicer() TradeRequestServicer
	}
)

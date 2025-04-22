// internal/ports/input/trade_request_service.go
package input

import (
	"context"
	"github/M2A96/Monopoly.git/internal/core/domain/bo"
	"github/M2A96/Monopoly.git/object/dao"

	"github.com/google/uuid"
)

// TradeRequestService defines the interface for trade request operations
type (
	TradeRequestServicer interface {
		// Get retrieves a trade request by ID
		Get(
			context.Context,
			uuid.UUID,
		) (bo.TradeRequester, error)

		// List retrieves a list of trade requests with pagination
		List(
			context.Context,
			dao.Paginationer,
			dao.TradeFilter,
		) ([]bo.TradeRequester, dao.Cursorer, error)

		// CreateTradeRequest creates a new trade request
		CreateTradeRequest(
			context.Context,
			bo.TradeRequester,
		) (uuid.UUID, error)

		// UpdateTradeRequest updates an existing trade request
		UpdateTradeRequest(
			context.Context,
			uuid.UUID,
			bo.TradeRequester,
		) error

		// DeleteTradeRequest removes a trade request
		DeleteTradeRequest(
			context.Context,
			uuid.UUID,
		) error

		// AcceptTradeRequest accepts a trade request
		AcceptTradeRequest(
			context.Context,
			uuid.UUID,
		) error

		// RejectTradeRequest rejects a trade request
		RejectTradeRequest(
			context.Context,
			uuid.UUID,
		) error

		// GetTradeRequestsByGameID retrieves all trade requests for a specific game
		GetTradeRequestsByGameID(
			context.Context,
			uuid.UUID,
			dao.Paginationer,
		) ([]bo.TradeRequester, dao.Cursorer, error)

		// GetTradeRequestsByPlayerID retrieves all trade requests for a specific player
		GetTradeRequestsByPlayerID(
			context.Context,
			uuid.UUID,
			dao.Paginationer,
		) ([]bo.TradeRequester, dao.Cursorer, error)
	}

	// GetTradeRequestServicer is an interface for retrieving the trade request service
	GetTradeRequestServicer interface {
		// GetTradeRequestServicer returns the trade request service implementation
		GetTradeRequestServicer() TradeRequestServicer
	}
)

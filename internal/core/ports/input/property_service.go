// internal/ports/input/property_service.go
package input

import (
	"context"
	"github/M2A96/Monopoly.git/internal/core/domain/bo"
	"github/M2A96/Monopoly.git/object/dao"

	"github.com/google/uuid"
)

// PropertyService defines the interface for property operations
type (
	PropertyServicer interface {
		// Get retrieves a property by ID
		Get(
			context.Context,
			uuid.UUID,
		) (bo.Propertyer, error)

		// List retrieves a list of properties with pagination
		List(
			context.Context,
			dao.Paginationer,
			dao.PropertyFilter,
		) ([]bo.Propertyer, dao.Cursorer, error)

		// CreateProperty creates a new property
		CreateProperty(
			context.Context,
			bo.Propertyer,
		) (uuid.UUID, error)

		// UpdateProperty updates an existing property
		UpdateProperty(
			context.Context,
			int,
			bo.Propertyer,
		) error

		// DeleteProperty removes a property from the game
		DeleteProperty(
			context.Context,
			uuid.UUID,
		) error

		// BuyProperty assigns a property to a player
		BuyProperty(
			context.Context,
			uuid.UUID,
			uuid.UUID,
		) error

		// MortgageProperty toggles the mortgage status of a property
		MortgageProperty(
			context.Context,
			uuid.UUID,
			bool,
		) error

		// AddHouse adds a house to a property
		AddHouse(
			context.Context,
			int,
		) error

		// AddHotel adds a hotel to a property
		AddHotel(
			context.Context,
			uuid.UUID,
		) error
	}

	// GetPropertyServicer is an interface for retrieving the property service
	GetPropertyServicer interface {
		// GetPropertyServicer returns the property service implementation
		GetPropertyServicer() PropertyServicer
	}
)

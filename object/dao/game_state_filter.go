package dao

import (
	"encoding/json"
	"fmt"

	"github/M2A96/Monopoly.git/object"

	"gorm.io/gorm"

	"github.com/google/uuid"
)

type (
	// GameStateFilter is an interface.
	GameStateFilter interface {
		Filterer
		// GetGameID is a function.
		GetGameID() uuid.UUID
		GetPlayerIDs() []uuid.UUID
		GetPropertyIDs() []int
	}

	gameStateFilter struct {
		Filterer
		gameID      uuid.UUID
		playerIDs   []uuid.UUID
		propertyIDs []int
	}
)

var (
	_ GameStateFilter  = (*gameStateFilter)(nil)
	_ Filterer         = (*gameStateFilter)(nil)
	_ json.Marshaler   = (*gameStateFilter)(nil)
	_ object.GetMapper = (*gameStateFilter)(nil)
)

func NewGameStateFilter(
	gameID uuid.UUID,
	playerIDs []uuid.UUID,
	propertyIDs []int,
) GameStateFilter {
	return &gameStateFilter{
		gameID:      gameID,
		playerIDs:   playerIDs,
		propertyIDs: propertyIDs,
	}
}

// GetMap implements object.GetMapper.
func (gs *gameStateFilter) GetMap() map[string]any {
	return map[string]any{
		"game_id":      gs.GetGameID(),
		"player_ids":   gs.GetPlayerIDs(),
		"property_ids": gs.GetPropertyIDs(),
	}
}

// MarshalJSON implements json.Marshaler.
func (gs *gameStateFilter) MarshalJSON() ([]byte, error) {
	return json.Marshal(gs.GetMap())
}

// Filter implements GameStateFilter.
// Subtle: this method shadows the method (Filterer).Filter of gameStateFilter.Filterer.
func (gs *gameStateFilter) Filter(
	gormDB *gorm.DB,
) *gorm.DB {
	if gs.GetGameID() != uuid.Nil {
		gormDB = gormDB.
			Where(fmt.Sprintf(`%[1]s.game_id = ?`, "game_state"), gs.GetGameID())
	}

	if len(gs.GetPlayerIDs()) != 0 {
		gormDB = gormDB.
			Where(fmt.Sprintf(`%[1]s.player_id IN ?`, "game_state_player"), gs.GetPlayerIDs())
	}

	if len(gs.GetPropertyIDs()) != 0 {
		gormDB = gormDB.
			Where(fmt.Sprintf(`%[1]s.property_id IN ?`, "game_state_property"), gs.GetPropertyIDs())
	}

	return gormDB
}

// GetGameID implements GameStateFilter.
func (gs *gameStateFilter) GetGameID() uuid.UUID {
	return gs.gameID
}

// GetPlayerIDs implements GameStateFilter.
func (gs *gameStateFilter) GetPlayerIDs() []uuid.UUID {
	return gs.playerIDs
}

// GetPropertyIDs implements GameStateFilter.
func (gs *gameStateFilter) GetPropertyIDs() []int {
	return gs.propertyIDs
}

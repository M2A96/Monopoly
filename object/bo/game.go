package bo

import (
	"encoding/json"
	"github/M2A96/Monopoly.git/object"

	"github.com/google/uuid"
)

type (
	Gamer interface {
		GetName() string
		GetStatus() string
		GetCurrentPlayerID() uuid.UUID
		GetWinnerID() uuid.UUID
	}

	game struct {
		name            string
		status          string
		currentPlayerID uuid.UUID
		winnerID        uuid.UUID
	}
)

var (
	_ Gamer            = (*game)(nil)
	_ object.GetMapper = (*game)(nil)
	_ json.Marshaler   = (*game)(nil)
)

func NewGamer(
	name string,
	status string,
	currentPlayerID uuid.UUID,
	winnerID uuid.UUID,
) *game {
	return &game{
		name:            name,
		status:          status,
		currentPlayerID: currentPlayerID,
		winnerID:        winnerID,
	}
}

// MarshalJSON implements json.Marshaler.
func (g *game) MarshalJSON() ([]byte, error) {
	return json.Marshal(g.GetMap())
}

// GetMap implements object.GetMapper.
func (g *game) GetMap() map[string]any {
	return map[string]any{
		"name":              g.GetName(),
		"status":            g.GetStatus(),
		"current_player_id": g.GetCurrentPlayerID(),
		"winner_id":         g.GetWinnerID(),
	}
}

// GetCurrentPlayerID implements Gamer.
func (g *game) GetCurrentPlayerID() uuid.UUID {
	return g.currentPlayerID
}

// GetName implements Gamer.
func (g *game) GetName() string {
	return g.name
}

// GetStatus implements Gamer.
func (g *game) GetStatus() string {
	return g.GetStatus()
}

// GetWinnerID implements Gamer.
func (g *game) GetWinnerID() uuid.UUID {
	return g.GetWinnerID()
}

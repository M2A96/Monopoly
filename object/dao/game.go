package dao

import (
	"database/sql"
	"encoding/json"
	"github/M2A96/Monopoly.git/object"
	"time"

	"github.com/google/uuid"
	"github.com/samber/lo"
)

type (
	Gamer interface {
		// GetCUDer is a function.
		GetCUDer() CUDer
		// GetCUDIDer is a function.
		GetCUDIDer() CUDIDer
		GetName() string
		GetStatus() string
		GetCreatedAt() time.Time
		GetUpdatedAt() time.Time
		GetCurrentPlayerID() uuid.UUID
		GetWinnerID() uuid.UUID
	}

	// Game represents a Monopoly game
	game struct {
		cuder           CUDer
		cudIDer         CUDIDer
		name            string
		status          string
		createdAt       time.Time
		updatedAt       time.Time
		currentPlayerID uuid.UUID
		winnerID        uuid.UUID
	}
)

var (
	_ Gamer            = (*game)(nil)
	_ object.GetMapper = (*game)(nil)
	_ json.Marshaler   = (*game)(nil)
)

// GetCreatedAt implements Gamer.
func (g *game) GetCreatedAt() time.Time {
	return g.createdAt
}

// GetCurrentPlayerID implementsgGamer.
func (g *game) GetCurrentPlayerID() uuid.UUID {
	return g.currentPlayerID
}

// GetName implements Gamer.
func (g *game) GetName() string {
	return g.name
}

// GetStatus implements Gamer.
func (g *game) GetStatus() string {
	return g.status
}

// GetUpdatedAt implements Gamer.
func (g *game) GetUpdatedAt() time.Time {
	return g.updatedAt
}

// GetWinnerID implements Gamer.
func (g *game) GetWinnerID() uuid.UUID {
	return g.winnerID
}

// GetCUDIDer implements Gamer.
func (g *game) GetCUDIDer() CUDIDer {
	return g.cudIDer
}

// GetCUDer implements Gamer.
func (g *game) GetCUDer() CUDer {
	return g.cuder
}

func NewGame(
	createdAt time.Time,
	updatedAt time.Time,
	deletedAt sql.NullTime,
	id uuid.UUID,
	name string,
	status string,
	currentPlayerID uuid.UUID,
	winnerID uuid.UUID,
) *game {
	return &game{
		cuder:           NewCUD(createdAt, updatedAt, deletedAt),
		cudIDer:         NewCUDID(map[string]uuid.UUID{"id": id}),
		name:            name,
		status:          status,
		createdAt:       createdAt,
		updatedAt:       updatedAt,
		currentPlayerID: currentPlayerID,
		winnerID:        winnerID,
	}
}

func NewGamerFromMap(
	uuider object.UUIDer,
	value map[string]any,
) (Gamer, error) {
	cuder, err := NewCUDerFromMap(value)
	if err != nil {
		return nil, err
	}

	cudIDer, err := NewCUDIDerFromMap(uuider, value)
	if err != nil {
		return nil, err
	}

	name, ok := value["name"].(string)
	if !ok {
		return nil, object.ErrTypeAssertion
	}

	status, ok := value["status"].(string)
	if !ok {
		return nil, object.ErrTypeAssertion
	}

	currentPlayerOd, ok := value["current_player"].(uuid.UUID)
	if !ok {
		return nil, object.ErrTypeAssertion
	}

	winnerId, ok := value["winner"].(uuid.UUID)
	if !ok {
		return nil, object.ErrTypeAssertion
	}

	return &game{
		cuder:           cuder,
		cudIDer:         cudIDer,
		name:            name,
		status:          status,
		createdAt:       cuder.GetCreatedAt(),
		updatedAt:       cuder.GetUpdatedAt(),
		currentPlayerID: currentPlayerOd,
		winnerID:        winnerId,
	}, nil

}

// MarshalJSON implements json.Marshaler.
func (g *game) MarshalJSON() ([]byte, error) {
	return json.Marshal(g.GetMap())
}

// GetMap implements object.GetMapper.
func (g *game) GetMap() map[string]any {
	return lo.Assign(
		g.cuder.GetMap(),
		g.cudIDer.GetMap(),
		map[string]any{
			"name":           g.GetName(),
			"status":         g.GetStatus(),
			"created_at":     g.GetCreatedAt(),
			"updated_at":     g.GetUpdatedAt(),
			"current_player": g.GetCurrentPlayerID(),
			"winner":         g.GetWinnerID(),
		})
}

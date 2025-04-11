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
	GameLogger interface {
		// GetCUDer is a function.
		GetCUDer() CUDer
		// GetCUDIDer is a function.
		GetCUDIDer() CUDIDer
		GetGameID() uuid.UUID
		GetPlayerID() uuid.UUID
		GetAction() string
		GetDescription() string
		GetTimestamp() time.Time
	}

	// gameLog represents a log entry for game events
	gameLog struct {
		cuder       CUDer
		cudIDer     CUDIDer
		gameID      uuid.UUID
		playerID    uuid.UUID
		action      string
		description string
		timestamp   time.Time
	}
)

var (
	_ GameLogger       = (*gameLog)(nil)
	_ object.GetMapper = (*gameLog)(nil)
	_ json.Marshaler   = (*gameLog)(nil)
)

// GetGameID implements GameLogger.
func (gl *gameLog) GetGameID() uuid.UUID {
	return gl.gameID
}

// GetPlayerID implements GameLogger.
func (gl *gameLog) GetPlayerID() uuid.UUID {
	return gl.playerID
}

// GetAction implements GameLogger.
func (gl *gameLog) GetAction() string {
	return gl.action
}

// GetDescription implements GameLogger.
func (gl *gameLog) GetDescription() string {
	return gl.description
}

// GetTimestamp implements GameLogger.
func (gl *gameLog) GetTimestamp() time.Time {
	return gl.timestamp
}

// GetCUDIDer implements GameLogger.
func (gl *gameLog) GetCUDIDer() CUDIDer {
	return gl.cudIDer
}

// GetCUDer implements GameLogger.
func (gl *gameLog) GetCUDer() CUDer {
	return gl.cuder
}

func NewGameLog(
	id uuid.UUID,
	gameID uuid.UUID,
	playerID uuid.UUID,
	action string,
	description string,
	timestamp time.Time,
	createdAt time.Time,
	updatedAt time.Time,
	deletedAt sql.NullTime,
) *gameLog {
	return &gameLog{
		cuder:       NewCUD(createdAt, updatedAt, deletedAt),
		cudIDer:     NewCUDID(map[string]uuid.UUID{"id": id}),
		gameID:      gameID,
		playerID:    playerID,
		action:      action,
		description: description,
		timestamp:   timestamp,
	}
}

func (gl *gameLog) NewGameLogFromMap(
	uuider object.UUIDer,
	value map[string]any,
) (GameLogger, error) {
	cuder, err := NewCUDerFromMap(value)
	if err != nil {
		return nil, err
	}

	cudIDer, err := NewCUDIDerFromMap(uuider, value)
	if err != nil {
		return nil, err
	}

	gameID, ok := value["game_id"].(uuid.UUID)
	if !ok {
		return nil, object.ErrTypeAssertion
	}

	playerID, ok := value["player_id"].(uuid.UUID)
	if !ok {
		return nil, object.ErrTypeAssertion
	}

	action, ok := value["action"].(string)
	if !ok {
		return nil, object.ErrTypeAssertion
	}

	description, ok := value["description"].(string)
	if !ok {
		return nil, object.ErrTypeAssertion
	}

	timestamp, ok := value["timestamp"].(time.Time)
	if !ok {
		return nil, object.ErrTypeAssertion
	}

	return &gameLog{
		cuder:       cuder,
		cudIDer:     cudIDer,
		gameID:      gameID,
		playerID:    playerID,
		action:      action,
		description: description,
		timestamp:   timestamp,
	}, nil
}

// MarshalJSON implements json.Marshaler.
func (gl *gameLog) MarshalJSON() ([]byte, error) {
	return json.Marshal(gl.GetMap())
}

// GetMap implements object.GetMapper.
func (gl *gameLog) GetMap() map[string]any {
	return lo.Assign(
		gl.cuder.GetMap(),
		gl.cudIDer.GetMap(),
		map[string]any{
			"game_id":     gl.GetGameID(),
			"player_id":   gl.GetPlayerID(),
			"action":      gl.GetAction(),
			"description": gl.GetDescription(),
			"timestamp":   gl.GetTimestamp(),
		})
}

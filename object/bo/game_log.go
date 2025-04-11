package bo

import (
	"encoding/json"
	"github/M2A96/Monopoly.git/object"
	"time"

	"github.com/google/uuid"
	"github.com/samber/lo"
)

type (
	GameLogger interface {
		GetGameID() uuid.UUID
		GetPlayerID() uuid.UUID
		GetAction() string
		GetDescription() string
		GetTimestamp() time.Time
	}

	// gameLog represents a log entry for game events
	gameLog struct {
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

func NewGameLog(
	id uuid.UUID,
	gameID uuid.UUID,
	playerID uuid.UUID,
	action string,
	description string,
	timestamp time.Time,
) *gameLog {
	return &gameLog{
		gameID:      gameID,
		playerID:    playerID,
		action:      action,
		description: description,
		timestamp:   timestamp,
	}
}

// MarshalJSON implements json.Marshaler.
func (gl *gameLog) MarshalJSON() ([]byte, error) {
	return json.Marshal(gl.GetMap())
}

// GetMap implements object.GetMapper.
func (gl *gameLog) GetMap() map[string]any {
	return lo.Assign(
		map[string]any{
			"game_id":     gl.GetGameID(),
			"player_id":   gl.GetPlayerID(),
			"action":      gl.GetAction(),
			"description": gl.GetDescription(),
			"timestamp":   gl.GetTimestamp(),
		})
}

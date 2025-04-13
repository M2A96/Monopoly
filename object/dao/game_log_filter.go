package dao

import (
	"encoding/json"
	"fmt"
	"time"

	"github/M2A96/Monopoly.git/object"

	"gorm.io/gorm"

	"github.com/google/uuid"
)

type (
	// GameLogFilter is an interface.
	GameLogFilter interface {
		Filterer
		// GetIDs is a function.
		GetIDs() []uuid.UUID
		GetGameID() uuid.UUID
		GetPlayerID() uuid.UUID
		GetAction() string
		GetTimestampStart() *time.Time
		GetTimestampEnd() *time.Time
	}

	gameLogFilter struct {
		Filterer
		ids            []uuid.UUID
		gameID         uuid.UUID
		playerID       uuid.UUID
		action         string
		timestampStart *time.Time
		timestampEnd   *time.Time
	}
)

var (
	_ GameLogFilter    = (*gameLogFilter)(nil)
	_ Filterer         = (*gameLogFilter)(nil)
	_ json.Marshaler   = (*gameLogFilter)(nil)
	_ object.GetMapper = (*gameLogFilter)(nil)
)

func NewGameLogFilter(
	ids []uuid.UUID,
	gameID uuid.UUID,
	playerID uuid.UUID,
	action string,
	timestampStart *time.Time,
	timestampEnd *time.Time,
) GameLogFilter {
	return &gameLogFilter{
		ids:            ids,
		gameID:         gameID,
		playerID:       playerID,
		action:         action,
		timestampStart: timestampStart,
		timestampEnd:   timestampEnd,
	}
}

// GetMap implements object.GetMapper.
func (gl *gameLogFilter) GetMap() map[string]any {
	return map[string]any{
		"ids":             gl.GetIDs(),
		"game_id":         gl.GetGameID(),
		"player_id":       gl.GetPlayerID(),
		"action":          gl.GetAction(),
		"timestamp_start": gl.GetTimestampStart(),
		"timestamp_end":   gl.GetTimestampEnd(),
	}
}

// MarshalJSON implements json.Marshaler.
func (gl *gameLogFilter) MarshalJSON() ([]byte, error) {
	return json.Marshal(gl.GetMap())
}

// Filter implements GameLogFilter.
// Subtle: this method shadows the method (Filterer).Filter of gameLogFilter.Filterer.
func (gl *gameLogFilter) Filter(
	gormDB *gorm.DB,
) *gorm.DB {
	if len(gl.GetIDs()) != 0 {
		gormDB = gormDB.
			Where(fmt.Sprintf(`%[1]s.id IN ?`, "game_log"), gl.GetIDs())
	}

	if gl.GetGameID() != uuid.Nil {
		gormDB = gormDB.
			Where(fmt.Sprintf(`%[1]s.game_id = ?`, "game_log"), gl.GetGameID())
	}

	if gl.GetPlayerID() != uuid.Nil {
		gormDB = gormDB.
			Where(fmt.Sprintf(`%[1]s.player_id = ?`, "game_log"), gl.GetPlayerID())
	}

	if gl.GetAction() != "" {
		gormDB = gormDB.
			Where(fmt.Sprintf(`%[1]s.action = ?`, "game_log"), gl.GetAction())
	}

	if gl.GetTimestampStart() != nil {
		gormDB = gormDB.
			Where(fmt.Sprintf(`%[1]s.timestamp >= ?`, "game_log"), gl.GetTimestampStart())
	}

	if gl.GetTimestampEnd() != nil {
		gormDB = gormDB.
			Where(fmt.Sprintf(`%[1]s.timestamp <= ?`, "game_log"), gl.GetTimestampEnd())
	}

	return gormDB
}

// GetIDs implements GameLogFilter.
func (gl *gameLogFilter) GetIDs() []uuid.UUID {
	return gl.ids
}

// GetGameID implements GameLogFilter.
func (gl *gameLogFilter) GetGameID() uuid.UUID {
	return gl.gameID
}

// GetPlayerID implements GameLogFilter.
func (gl *gameLogFilter) GetPlayerID() uuid.UUID {
	return gl.playerID
}

// GetAction implements GameLogFilter.
func (gl *gameLogFilter) GetAction() string {
	return gl.action
}

// GetTimestampStart implements GameLogFilter.
func (gl *gameLogFilter) GetTimestampStart() *time.Time {
	return gl.timestampStart
}

// GetTimestampEnd implements GameLogFilter.
func (gl *gameLogFilter) GetTimestampEnd() *time.Time {
	return gl.timestampEnd
}

package dao

import (
	"encoding/json"
	"fmt"

	"github/M2A96/Monopoly.git/object"

	"gorm.io/gorm"

	"github.com/google/uuid"
)

type (
	// GameFilter is an interface.
	GameFilter interface {
		Filterer
		// GetIDs is a function.
		GetIDs() []uuid.UUID
		GetName() string
		GetStatus() string
		GetCurrentPlayerID() uuid.UUID
		GetWinnerID() uuid.UUID
	}

	gameFilter struct {
		Filterer
		ids             []uuid.UUID
		name            string
		status          string
		currentPlayerID uuid.UUID
		winnerID        uuid.UUID
	}
)

var (
	_ GameFilter       = (*gameFilter)(nil)
	_ Filterer         = (*gameFilter)(nil)
	_ json.Marshaler   = (*gameFilter)(nil)
	_ object.GetMapper = (*gameFilter)(nil)
)

func NewGameFilter(
	ids []uuid.UUID,
	name string,
	status string,
	currentPlayerID uuid.UUID,
	winnerID uuid.UUID,
) GameFilter {
	return &gameFilter{
		ids:             ids,
		name:            name,
		status:          status,
		currentPlayerID: currentPlayerID,
		winnerID:        winnerID,
	}
}

// GetMap implements object.GetMapper.
func (g *gameFilter) GetMap() map[string]any {
	return map[string]any{
		object.URIGameFilterFieldIDs:             g.GetIDs(),
		object.URIGameFilterFieldName:            g.GetName(),
		object.URIGameFilterFieldStatus:          g.GetStatus(),
		object.URIGameFilterFieldCurrentPlayerID: g.GetCurrentPlayerID(),
		object.URIGameFilterFieldWinnerID:        g.GetWinnerID(),
	}
}

// MarshalJSON implements json.Marshaler.
func (g *gameFilter) MarshalJSON() ([]byte, error) {
	return json.Marshal(g.GetMap())
}

// Filter implements GameFilter.
// Subtle: this method shadows the method (Filterer).Filter of gameFilter.Filterer.
func (g *gameFilter) Filter(
	gormDB *gorm.DB,
) *gorm.DB {
	if len(g.GetIDs()) != 0 {
		gormDB = gormDB.
			Where(fmt.Sprintf(`%[1]s.id IN ?`, object.URITableGame), g.GetIDs())
	}

	if g.GetName() != "" {
		gormDB = gormDB.
			Where(fmt.Sprintf(`%[1]s.name = ?`, object.URITableGame), g.GetName())
	}

	if g.GetStatus() != "" {
		gormDB = gormDB.
			Where(fmt.Sprintf(`%[1]s.status = ?`, object.URITableGame), g.GetStatus())
	}

	if g.GetCurrentPlayerID() != uuid.Nil {
		gormDB = gormDB.
			Where(fmt.Sprintf(`%[1]s.current_player_id = ?`, object.URITableGame), g.GetCurrentPlayerID())
	}

	if g.GetWinnerID() != uuid.Nil {
		gormDB = gormDB.
			Where(fmt.Sprintf(`%[1]s.winner_id = ?`, object.URITableGame), g.GetWinnerID())
	}

	return gormDB
}

// GetCurrentPlayerID implements GameFilter.
func (g *gameFilter) GetCurrentPlayerID() uuid.UUID {
	return g.currentPlayerID
}

// GetIDs implements GameFilter.
func (g *gameFilter) GetIDs() []uuid.UUID {
	return g.ids
}

// GetName implements GameFilter.
func (g *gameFilter) GetName() string {
	return g.name
}

// GetStatus implements GameFilter.
func (g *gameFilter) GetStatus() string {
	return g.status
}

// GetWinnerID implements GameFilter.
func (g *gameFilter) GetWinnerID() uuid.UUID {
	return g.winnerID
}

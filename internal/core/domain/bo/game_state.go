package bo

import (
	"encoding/json"
	"github/M2A96/Monopoly.git/object"
)

type (
	GameStater interface {
		// GetGame returns the game associated with this game state
		GetGame() Gamer
		// GetPlayers returns the players in the game
		GetPlayers() []Player
		// GetProperties returns the properties in the game
		GetProperties() []Propertyer
	}

	// gameState represents the current state of a Monopoly game
	gameState struct {
		game       Gamer        `json:"game"`
		players    []Player     `json:"players,omitempty"`
		properties []Propertyer `json:"properties,omitempty"`
	}
)

var (
	_ GameStater       = (*gameState)(nil)
	_ object.GetMapper = (*gameState)(nil)
	_ json.Marshaler   = (*gameState)(nil)
)

// GetGame implements GameStater.
func (gs *gameState) GetGame() Gamer {
	return gs.game
}

// GetPlayers implements GameStater.
func (gs *gameState) GetPlayers() []Player {
	return gs.players
}

// GetProperties implements GameStater.
func (gs *gameState) GetProperties() []Propertyer {
	return gs.properties
}

// NewGameState creates a new game state
func NewGameState(
	game Gamer,
	players []Player,
	properties []Propertyer,
) *gameState {
	return &gameState{
		game:       game,
		players:    players,
		properties: properties,
	}
}

// MarshalJSON implements json.Marshaler.
func (gs *gameState) MarshalJSON() ([]byte, error) {
	return json.Marshal(gs.GetMap())
}

// GetMap implements object.GetMapper.
func (gs *gameState) GetMap() map[string]any {
	return map[string]any{
		"game":       gs.game,
		"players":    gs.players,
		"properties": gs.properties,
	}
}

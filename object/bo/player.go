package bo

import (
	"encoding/json"
	"github/M2A96/Monopoly.git/object"

	"github.com/google/uuid"
	"github.com/samber/lo"
)

type (
	Player interface {
		GetName() string
		GetBalance() int
		GetPosition() int
		GetInJail() bool
		GetJailTurns() int
		GetBankrupt() bool
		GetGameID() uuid.UUID
	}

	// player represents a player in the game
	player struct {
		name      string
		balance   int
		position  int
		inJail    bool
		jailTurns int
		bankrupt  bool
		gameID    uuid.UUID
	}
)

var (
	_ Player           = (*player)(nil)
	_ object.GetMapper = (*player)(nil)
	_ json.Marshaler   = (*player)(nil)
)

// GetName implements Player.
func (p *player) GetName() string {
	return p.name
}

// GetBalance implements Player.
func (p *player) GetBalance() int {
	return p.balance
}

// GetPosition implements Player.
func (p *player) GetPosition() int {
	return p.position
}

// GetInJail implements Player.
func (p *player) GetInJail() bool {
	return p.inJail
}

// GetJailTurns implements Player.
func (p *player) GetJailTurns() int {
	return p.jailTurns
}

// GetBankrupt implements Player.
func (p *player) GetBankrupt() bool {
	return p.bankrupt
}

// GetGameID implements Player.
func (p *player) GetGameID() uuid.UUID {
	return p.gameID
}

func NewPlayer(
	id uuid.UUID,
	gameID uuid.UUID,
	name string,
	balance int,
	position int,
	inJail bool,
	jailTurns int,
	bankrupt bool,
) *player {
	return &player{
		name:      name,
		balance:   balance,
		position:  position,
		inJail:    inJail,
		jailTurns: jailTurns,
		bankrupt:  bankrupt,
		gameID:    gameID,
	}
}

// MarshalJSON implements json.Marshaler.
func (p *player) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.GetMap())
}

// GetMap implements object.GetMapper.
func (p *player) GetMap() map[string]any {
	return lo.Assign(
		map[string]any{
			"name":       p.GetName(),
			"balance":    p.GetBalance(),
			"position":   p.GetPosition(),
			"in_jail":    p.GetInJail(),
			"jail_turns": p.GetJailTurns(),
			"bankrupt":   p.GetBankrupt(),
			"game_id":    p.GetGameID(),
		})
}

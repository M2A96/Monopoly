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
	Player interface {
		// GetCUDer is a function.
		GetCUDer() CUDer
		// GetCUDIDer is a function.
		GetCUDIDer() CUDIDer
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
		cuder     CUDer
		cudIDer   CUDIDer
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

// GetCUDIDer implements Player.
func (p *player) GetCUDIDer() CUDIDer {
	return p.cudIDer
}

// GetCUDer implements Player.
func (p *player) GetCUDer() CUDer {
	return p.cuder
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
	createdAt time.Time,
	updatedAt time.Time,
	deletedAt sql.NullTime,
) *player {
	return &player{
		cuder:     NewCUD(createdAt, updatedAt, deletedAt),
		cudIDer:   NewCUDID(map[string]uuid.UUID{"id": id}),
		name:      name,
		balance:   balance,
		position:  position,
		inJail:    inJail,
		jailTurns: jailTurns,
		bankrupt:  bankrupt,
		gameID:    gameID,
	}
}

func (p *player) NewPlayerFromMap(
	uuider object.UUIDer,
	value map[string]any,
) (Player, error) {
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

	balance, ok := value["balance"].(int)
	if !ok {
		return nil, object.ErrTypeAssertion
	}

	position, ok := value["position"].(int)
	if !ok {
		return nil, object.ErrTypeAssertion
	}

	inJail, ok := value["in_jail"].(bool)
	if !ok {
		return nil, object.ErrTypeAssertion
	}

	jailTurns, ok := value["jail_turns"].(int)
	if !ok {
		return nil, object.ErrTypeAssertion
	}

	bankrupt, ok := value["bankrupt"].(bool)
	if !ok {
		return nil, object.ErrTypeAssertion
	}

	gameIDStr, ok := value["game_id"].(string)
	if !ok {
		return nil, object.ErrTypeAssertion
	}

	gameID, err := uuider.Parse(gameIDStr)
	if err != nil {
		return nil, err
	}

	return &player{
		cuder:     cuder,
		cudIDer:   cudIDer,
		name:      name,
		balance:   balance,
		position:  position,
		inJail:    inJail,
		jailTurns: jailTurns,
		bankrupt:  bankrupt,
		gameID:    gameID,
	}, nil
}

// MarshalJSON implements json.Marshaler.
func (p *player) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.GetMap())
}

// GetMap implements object.GetMapper.
func (p *player) GetMap() map[string]any {
	return lo.Assign(
		p.cuder.GetMap(),
		p.cudIDer.GetMap(),
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

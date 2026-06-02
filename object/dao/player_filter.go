package dao

import (
	"encoding/json"
	"fmt"

	"github/M2A96/Monopoly.git/object"

	"gorm.io/gorm"

	"github.com/google/uuid"
)

type (
	// PlayerFilter is an interface.
	PlayerFilter interface {
		Filterer
		// GetIDs is a function.
		GetIDs() []uuid.UUID
		GetName() string
		GetGameID() uuid.UUID
		GetBalance() int
		GetPosition() int
		GetInJail() bool
		GetJailTurns() int
		GetBankrupt() bool
	}

	playerFilter struct {
		Filterer
		ids       []uuid.UUID
		name      string
		gameID    uuid.UUID
		balance   int
		position  int
		inJail    bool
		jailTurns int
		bankrupt  bool
	}
)

var (
	_ PlayerFilter     = (*playerFilter)(nil)
	_ Filterer         = (*playerFilter)(nil)
	_ json.Marshaler   = (*playerFilter)(nil)
	_ object.GetMapper = (*playerFilter)(nil)
)

func NewPlayerFilter(
	ids []uuid.UUID,
	name string,
	gameID uuid.UUID,
	balance int,
	position int,
	inJail bool,
	jailTurns int,
	bankrupt bool,
) PlayerFilter {
	return &playerFilter{
		ids:       ids,
		name:      name,
		gameID:    gameID,
		balance:   balance,
		position:  position,
		inJail:    inJail,
		jailTurns: jailTurns,
		bankrupt:  bankrupt,
	}
}

// GetMap implements object.GetMapper.
func (p *playerFilter) GetMap() map[string]any {
	return map[string]any{
		object.URIPlayerFilterFieldIDs:       p.GetIDs(),
		object.URIPlayerFilterFieldName:      p.GetName(),
		object.URIPlayerFilterFieldGameID:    p.GetGameID(),
		object.URIPlayerFilterFieldBalance:   p.GetBalance(),
		object.URIPlayerFilterFieldPosition:  p.GetPosition(),
		object.URIPlayerFilterFieldInJail:    p.GetInJail(),
		object.URIPlayerFilterFieldJailTurns: p.GetJailTurns(),
		object.URIPlayerFilterFieldBankrupt:  p.GetBankrupt(),
	}
}

// MarshalJSON implements json.Marshaler.
func (p *playerFilter) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.GetMap())
}

// Filter implements PlayerFilter.
// Subtle: this method shadows the method (Filterer).Filter of playerFilter.Filterer.
func (p *playerFilter) Filter(
	gormDB *gorm.DB,
) *gorm.DB {
	if len(p.GetIDs()) != 0 {
		gormDB = gormDB.
			Where(fmt.Sprintf(`%[1]s.id IN ?`, object.URITablePlayer), p.GetIDs())
	}

	if p.GetName() != "" {
		gormDB = gormDB.
			Where(fmt.Sprintf(`%[1]s.name = ?`, object.URITablePlayer), p.GetName())
	}

	if p.GetGameID() != uuid.Nil {
		gormDB = gormDB.
			Where(fmt.Sprintf(`%[1]s.game_id = ?`, object.URITablePlayer), p.GetGameID())
	}

	if p.GetBalance() != 0 {
		gormDB = gormDB.
			Where(fmt.Sprintf(`%[1]s.balance = ?`, object.URITablePlayer), p.GetBalance())
	}

	if p.GetPosition() != 0 {
		gormDB = gormDB.
			Where(fmt.Sprintf(`%[1]s.position = ?`, object.URITablePlayer), p.GetPosition())
	}

	if p.GetInJail() {
		gormDB = gormDB.
			Where(fmt.Sprintf(`%[1]s.in_jail = ?`, object.URITablePlayer), p.GetInJail())
	}

	if p.GetJailTurns() != 0 {
		gormDB = gormDB.
			Where(fmt.Sprintf(`%[1]s.jail_turns = ?`, object.URITablePlayer), p.GetJailTurns())
	}

	if p.GetBankrupt() {
		gormDB = gormDB.
			Where(fmt.Sprintf(`%[1]s.bankrupt = ?`, object.URITablePlayer), p.GetBankrupt())
	}

	return gormDB
}

// GetIDs implements PlayerFilter.
func (p *playerFilter) GetIDs() []uuid.UUID {
	return p.ids
}

// GetName implements PlayerFilter.
func (p *playerFilter) GetName() string {
	return p.name
}

// GetGameID implements PlayerFilter.
func (p *playerFilter) GetGameID() uuid.UUID {
	return p.gameID
}

// GetBalance implements PlayerFilter.
func (p *playerFilter) GetBalance() int {
	return p.balance
}

// GetPosition implements PlayerFilter.
func (p *playerFilter) GetPosition() int {
	return p.position
}

// GetInJail implements PlayerFilter.
func (p *playerFilter) GetInJail() bool {
	return p.inJail
}

// GetJailTurns implements PlayerFilter.
func (p *playerFilter) GetJailTurns() int {
	return p.jailTurns
}

// GetBankrupt implements PlayerFilter.
func (p *playerFilter) GetBankrupt() bool {
	return p.bankrupt
}

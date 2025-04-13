package dao

import (
	"encoding/json"
	"fmt"

	"github/M2A96/Monopoly.git/object"

	"gorm.io/gorm"

	"github.com/google/uuid"
)

type (
	// TradeFilter is an interface.
	TradeFilter interface {
		Filterer
		// GetIDs is a function.
		GetIDs() []uuid.UUID
		GetSenderID() uuid.UUID
		GetReceiverID() uuid.UUID
		GetStatus() string
		GetGameID() uuid.UUID
	}

	tradeFilter struct {
		Filterer
		ids        []uuid.UUID
		senderID   uuid.UUID
		receiverID uuid.UUID
		status     string
		gameID     uuid.UUID
	}
)

var (
	_ TradeFilter      = (*tradeFilter)(nil)
	_ Filterer         = (*tradeFilter)(nil)
	_ json.Marshaler   = (*tradeFilter)(nil)
	_ object.GetMapper = (*tradeFilter)(nil)
)

func NewTradeFilter(
	ids []uuid.UUID,
	senderID uuid.UUID,
	receiverID uuid.UUID,
	status string,
	gameID uuid.UUID,
) TradeFilter {
	return &tradeFilter{
		ids:        ids,
		senderID:   senderID,
		receiverID: receiverID,
		status:     status,
		gameID:     gameID,
	}
}

// GetMap implements object.GetMapper.
func (t *tradeFilter) GetMap() map[string]any {
	return map[string]any{
		"ids":         t.GetIDs(),
		"sender_id":   t.GetSenderID(),
		"receiver_id": t.GetReceiverID(),
		"status":      t.GetStatus(),
		"game_id":     t.GetGameID(),
	}
}

// MarshalJSON implements json.Marshaler.
func (t *tradeFilter) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.GetMap())
}

// Filter implements TradeFilter.
// Subtle: this method shadows the method (Filterer).Filter of tradeFilter.Filterer.
func (t *tradeFilter) Filter(
	gormDB *gorm.DB,
) *gorm.DB {
	if len(t.GetIDs()) != 0 {
		gormDB = gormDB.
			Where(fmt.Sprintf(`%[1]s.id IN ?`, "trade_request"), t.GetIDs())
	}

	if t.GetSenderID() != uuid.Nil {
		gormDB = gormDB.
			Where(fmt.Sprintf(`%[1]s.sender_id = ?`, "trade_request"), t.GetSenderID())
	}

	if t.GetReceiverID() != uuid.Nil {
		gormDB = gormDB.
			Where(fmt.Sprintf(`%[1]s.receiver_id = ?`, "trade_request"), t.GetReceiverID())
	}

	if t.GetStatus() != "" {
		gormDB = gormDB.
			Where(fmt.Sprintf(`%[1]s.status = ?`, "trade_request"), t.GetStatus())
	}

	if t.GetGameID() != uuid.Nil {
		gormDB = gormDB.
			Where(fmt.Sprintf(`%[1]s.game_id = ?`, "trade_request"), t.GetGameID())
	}

	return gormDB
}

// GetIDs implements TradeFilter.
func (t *tradeFilter) GetIDs() []uuid.UUID {
	return t.ids
}

// GetSenderID implements TradeFilter.
func (t *tradeFilter) GetSenderID() uuid.UUID {
	return t.senderID
}

// GetReceiverID implements TradeFilter.
func (t *tradeFilter) GetReceiverID() uuid.UUID {
	return t.receiverID
}

// GetStatus implements TradeFilter.
func (t *tradeFilter) GetStatus() string {
	return t.status
}

// GetGameID implements TradeFilter.
func (t *tradeFilter) GetGameID() uuid.UUID {
	return t.gameID
}

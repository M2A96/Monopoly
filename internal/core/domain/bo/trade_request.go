package bo

import (
	"encoding/json"
	"github/M2A96/Monopoly.git/object"

	"github.com/google/uuid"
	"github.com/samber/lo"
)

type (
	TradeRequester interface {
		GetSenderID() uuid.UUID
		GetReceiverID() uuid.UUID
		GetOfferingMoney() int
		GetRequestingMoney() int
		GetOfferingProperties() []int
		GetRequestingProperties() []int
		GetStatus() string
		GetGameID() uuid.UUID
	}

	// tradeRequest represents a trade proposal between players
	tradeRequest struct {
		senderID             uuid.UUID
		receiverID           uuid.UUID
		offeringMoney        int
		requestingMoney      int
		offeringProperties   []int
		requestingProperties []int
		status               string // e.g., "pending", "accepted", "rejected"
		gameID               uuid.UUID
	}
)

var (
	_ TradeRequester   = (*tradeRequest)(nil)
	_ object.GetMapper = (*tradeRequest)(nil)
	_ json.Marshaler   = (*tradeRequest)(nil)
)

// GetSenderID implements TradeRequester.
func (tr *tradeRequest) GetSenderID() uuid.UUID {
	return tr.senderID
}

// GetReceiverID implements TradeRequester.
func (tr *tradeRequest) GetReceiverID() uuid.UUID {
	return tr.receiverID
}

// GetOfferingMoney implements TradeRequester.
func (tr *tradeRequest) GetOfferingMoney() int {
	return tr.offeringMoney
}

// GetRequestingMoney implements TradeRequester.
func (tr *tradeRequest) GetRequestingMoney() int {
	return tr.requestingMoney
}

// GetOfferingProperties implements TradeRequester.
func (tr *tradeRequest) GetOfferingProperties() []int {
	return tr.offeringProperties
}

// GetRequestingProperties implements TradeRequester.
func (tr *tradeRequest) GetRequestingProperties() []int {
	return tr.requestingProperties
}

// GetStatus implements TradeRequester.
func (tr *tradeRequest) GetStatus() string {
	return tr.status
}

// GetGameID implements TradeRequester.
func (tr *tradeRequest) GetGameID() uuid.UUID {
	return tr.gameID
}

func NewTradeRequest(
	id uuid.UUID,
	gameID uuid.UUID,
	senderID uuid.UUID,
	receiverID uuid.UUID,
	offeringMoney int,
	requestingMoney int,
	offeringProperties []int,
	requestingProperties []int,
	status string,
) *tradeRequest {
	return &tradeRequest{
		senderID:             senderID,
		receiverID:           receiverID,
		offeringMoney:        offeringMoney,
		requestingMoney:      requestingMoney,
		offeringProperties:   offeringProperties,
		requestingProperties: requestingProperties,
		status:               status,
		gameID:               gameID,
	}
}

// MarshalJSON implements json.Marshaler.
func (tr *tradeRequest) MarshalJSON() ([]byte, error) {
	return json.Marshal(tr.GetMap())
}

// GetMap implements object.GetMapper.
func (tr *tradeRequest) GetMap() map[string]any {
	return lo.Assign(
		map[string]any{
			"sender_id":             tr.GetSenderID(),
			"receiver_id":           tr.GetReceiverID(),
			"offering_money":        tr.GetOfferingMoney(),
			"requesting_money":      tr.GetRequestingMoney(),
			"offering_properties":   tr.GetOfferingProperties(),
			"requesting_properties": tr.GetRequestingProperties(),
			"status":                tr.GetStatus(),
			"game_id":               tr.GetGameID(),
		})
}

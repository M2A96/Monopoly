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
	TradeRequester interface {
		// GetCUDer is a function.
		GetCUDer() CUDer
		// GetCUDIDer is a function.
		GetCUDIDer() CUDIDer
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
		cuder                CUDer
		cudIDer              CUDIDer
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

// GetCUDIDer implements TradeRequester.
func (tr *tradeRequest) GetCUDIDer() CUDIDer {
	return tr.cudIDer
}

// GetCUDer implements TradeRequester.
func (tr *tradeRequest) GetCUDer() CUDer {
	return tr.cuder
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
	createdAt time.Time,
	updatedAt time.Time,
	deletedAt sql.NullTime,
) *tradeRequest {
	return &tradeRequest{
		cuder:                NewCUD(createdAt, updatedAt, deletedAt),
		cudIDer:              NewCUDID(map[string]uuid.UUID{"id": id}),
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

func NewTradeRequestFromMap(
	uuider object.UUIDer,
	value map[string]any,
) (TradeRequester, error) {
	cuder, err := NewCUDerFromMap(value)
	if err != nil {
		return nil, err
	}

	cudIDer, err := NewCUDIDerFromMap(uuider, value)
	if err != nil {
		return nil, err
	}

	senderIDStr, ok := value["sender_id"].(string)
	if !ok {
		return nil, object.ErrTypeAssertion
	}

	senderID, err := uuider.Parse(senderIDStr)
	if err != nil {
		return nil, err
	}

	receiverIDStr, ok := value["receiver_id"].(string)
	if !ok {
		return nil, object.ErrTypeAssertion
	}

	receiverID, err := uuider.Parse(receiverIDStr)
	if err != nil {
		return nil, err
	}

	offeringMoney, ok := value["offering_money"].(int)
	if !ok {
		return nil, object.ErrTypeAssertion
	}

	requestingMoney, ok := value["requesting_money"].(int)
	if !ok {
		return nil, object.ErrTypeAssertion
	}

	offeringPropertiesAny, ok := value["offering_properties"].([]any)
	if !ok {
		return nil, object.ErrTypeAssertion
	}

	offeringProperties := make([]int, len(offeringPropertiesAny))
	for i, v := range offeringPropertiesAny {
		propID, ok := v.(int)
		if !ok {
			return nil, object.ErrTypeAssertion
		}
		offeringProperties[i] = propID
	}

	requestingPropertiesAny, ok := value["requesting_properties"].([]any)
	if !ok {
		return nil, object.ErrTypeAssertion
	}

	requestingProperties := make([]int, len(requestingPropertiesAny))
	for i, v := range requestingPropertiesAny {
		propID, ok := v.(int)
		if !ok {
			return nil, object.ErrTypeAssertion
		}
		requestingProperties[i] = propID
	}

	status, ok := value["status"].(string)
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

	return &tradeRequest{
		cuder:                cuder,
		cudIDer:              cudIDer,
		senderID:             senderID,
		receiverID:           receiverID,
		offeringMoney:        offeringMoney,
		requestingMoney:      requestingMoney,
		offeringProperties:   offeringProperties,
		requestingProperties: requestingProperties,
		status:               status,
		gameID:               gameID,
	}, nil
}

// MarshalJSON implements json.Marshaler.
func (tr *tradeRequest) MarshalJSON() ([]byte, error) {
	return json.Marshal(tr.GetMap())
}

// GetMap implements object.GetMapper.
func (tr *tradeRequest) GetMap() map[string]any {
	return lo.Assign(
		tr.cuder.GetMap(),
		tr.cudIDer.GetMap(),
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

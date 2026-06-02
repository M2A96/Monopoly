package models

import "github.com/google/uuid"

type ActionWindowType string

const (
	ActionWaitingForRoll                 ActionWindowType = "WAITING_FOR_ROLL"
	ActionWaitingForPurchaseDecision     ActionWindowType = "WAITING_FOR_PURCHASE_DECISION"
	ActionWaitingForTradeResponse        ActionWindowType = "WAITING_FOR_TRADE_RESPONSE"
	ActionWaitingForAuctionBid           ActionWindowType = "WAITING_FOR_AUCTION_BID"
	ActionWaitingForJailDecision         ActionWindowType = "WAITING_FOR_JAIL_DECISION"
	ActionWaitingForBankruptcyResolution ActionWindowType = "WAITING_FOR_BANKRUPTCY_RESOLUTION"
	ActionWaitingForCardEffect           ActionWindowType = "WAITING_FOR_CARD_EFFECT"
)

// ActionWindow describes the expected response from one or more players
// at a specific point in the game flow.
type ActionWindow struct {
	Type       ActionWindowType
	PlayerID   uuid.UUID // player whose response is awaited
	SpaceIndex int       // relevant space (purchase decision, auction, etc.)
	CardID     string    // relevant card ID, if the action is card-triggered
}

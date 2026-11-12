package events

import (
	"time"

	"github.com/google/uuid"
)

type EventType string

const (
	EventGameCreated             EventType = "GAME_CREATED"
	EventGameStarted             EventType = "GAME_STARTED"
	EventGameEnded               EventType = "GAME_ENDED"
	EventPlayerJoined            EventType = "PLAYER_JOINED"
	EventTurnStarted             EventType = "TURN_STARTED"
	EventTurnEnded               EventType = "TURN_ENDED"
	EventDiceRolled              EventType = "DICE_ROLLED"
	EventPlayerMoved             EventType = "PLAYER_MOVED"
	EventPassedGo                EventType = "PASSED_GO"
	EventLandedOnSpace           EventType = "LANDED_ON_SPACE"
	EventPropertyPurchased       EventType = "PROPERTY_PURCHASED"
	EventAuctionStarted          EventType = "AUCTION_STARTED"
	EventAuctionBidPlaced        EventType = "AUCTION_BID_PLACED"
	EventAuctionEnded            EventType = "AUCTION_ENDED"
	EventRentCharged             EventType = "RENT_CHARGED"
	EventHouseBuilt              EventType = "HOUSE_BUILT"
	EventHotelBuilt              EventType = "HOTEL_BUILT"
	EventHouseSold               EventType = "HOUSE_SOLD"
	EventHotelSold               EventType = "HOTEL_SOLD"
	EventPropertyMortgaged       EventType = "PROPERTY_MORTGAGED"
	EventPropertyUnmortgaged     EventType = "PROPERTY_UNMORTGAGED"
	EventSentToJail              EventType = "SENT_TO_JAIL"
	EventReleasedFromJail        EventType = "RELEASED_FROM_JAIL"
	EventTradeProposed           EventType = "TRADE_PROPOSED"
	EventTradeAccepted           EventType = "TRADE_ACCEPTED"
	EventTradeRejected           EventType = "TRADE_REJECTED"
	EventBankruptcyDeclared      EventType = "BANKRUPTCY_DECLARED"
	EventChanceCardDrawn         EventType = "CHANCE_CARD_DRAWN"
	EventCommunityChestCardDrawn EventType = "COMMUNITY_CHEST_CARD_DRAWN"
	EventTaxPaid                 EventType = "TAX_PAID"
	EventAuctionPassed           EventType = "AUCTION_PASSED"
	EventDebtIncurred            EventType = "DEBT_INCURRED"
	EventDebtPaid                EventType = "DEBT_PAID"
	EventPlayerResigned          EventType = "PLAYER_RESIGNED"
)

// Event is the base interface for all domain events.
type Event interface {
	GetType() EventType
	GetGameID() uuid.UUID
	GetSequenceNumber() int64
	GetOccurredAt() time.Time
}

// BaseEvent contains common fields embedded in every event.
type BaseEvent struct {
	Type           EventType
	GameID         uuid.UUID
	SequenceNumber int64
	OccurredAt     time.Time
}

func (e BaseEvent) GetType() EventType       { return e.Type }
func (e BaseEvent) GetGameID() uuid.UUID     { return e.GameID }
func (e BaseEvent) GetSequenceNumber() int64 { return e.SequenceNumber }
func (e BaseEvent) GetOccurredAt() time.Time { return e.OccurredAt }

// Stamper allows the reducer registry to stamp base fields on events
// without knowing the concrete type. All event types inherit this via BaseEvent embedding.
type Stamper interface {
	StampBase(gameID uuid.UUID, seq int64)
}

func (e *BaseEvent) StampBase(gameID uuid.UUID, seq int64) {
	e.GameID = gameID
	e.SequenceNumber = seq
}

func StampEvents(evts []Event, gameID uuid.UUID, counter *int64) {
	for _, evt := range evts {
		if s, ok := evt.(Stamper); ok {
			*counter++
			s.StampBase(gameID, *counter)
		}
	}
}

// --- Game lifecycle ---

type GameCreatedEvent struct {
	BaseEvent
	Name string
}

type GameStartedEvent struct {
	BaseEvent
}

type GameEndedEvent struct {
	BaseEvent
	WinnerID uuid.UUID
}

type PlayerJoinedEvent struct {
	BaseEvent
	PlayerID uuid.UUID
	Name     string
	Order    int // turn order index
}

// --- Turn flow ---

type TurnStartedEvent struct {
	BaseEvent
	PlayerID   uuid.UUID
	TurnNumber int
}

type TurnEndedEvent struct {
	BaseEvent
	PlayerID   uuid.UUID
	TurnNumber int
}

// --- Dice and movement ---

type DiceRolledEvent struct {
	BaseEvent
	PlayerID uuid.UUID
	Die1     int
	Die2     int
}

type PlayerMovedEvent struct {
	BaseEvent
	PlayerID     uuid.UUID
	FromPosition int
	ToPosition   int
}

// PassedGoEvent is emitted whenever a player's movement crosses or lands on Go.
type PassedGoEvent struct {
	BaseEvent
	PlayerID uuid.UUID
	Amount   int // always 200 in standard rules
}

type LandedOnSpaceEvent struct {
	BaseEvent
	PlayerID   uuid.UUID
	SpaceIndex int
}

// --- Property ---

type PropertyPurchasedEvent struct {
	BaseEvent
	PlayerID   uuid.UUID
	SpaceIndex int
	Price      int
}

type RentChargedEvent struct {
	BaseEvent
	PayerID    uuid.UUID
	OwnerID    uuid.UUID
	SpaceIndex int
	Amount     int
}

type HouseBuiltEvent struct {
	BaseEvent
	PlayerID   uuid.UUID
	SpaceIndex int
	Cost       int
}

type HotelBuiltEvent struct {
	BaseEvent
	PlayerID   uuid.UUID
	SpaceIndex int
	Cost       int
}

type HouseSoldEvent struct {
	BaseEvent
	PlayerID   uuid.UUID
	SpaceIndex int
	Proceeds   int
}

type HotelSoldEvent struct {
	BaseEvent
	PlayerID   uuid.UUID
	SpaceIndex int
	Proceeds   int
}

type PropertyMortgagedEvent struct {
	BaseEvent
	PlayerID   uuid.UUID
	SpaceIndex int
	Amount     int // mortgage value received
}

type PropertyUnmortgagedEvent struct {
	BaseEvent
	PlayerID   uuid.UUID
	SpaceIndex int
	Cost       int // redemption cost paid
}

// --- Auction ---

type AuctionStartedEvent struct {
	BaseEvent
	SpaceIndex int
}

type AuctionBidPlacedEvent struct {
	BaseEvent
	PlayerID uuid.UUID
	Amount   int
}

type AuctionEndedEvent struct {
	BaseEvent
	WinnerID   uuid.UUID
	SpaceIndex int
	Amount     int
}

// --- Jail ---

type SentToJailEvent struct {
	BaseEvent
	PlayerID uuid.UUID
}

type ReleasedFromJailEvent struct {
	BaseEvent
	PlayerID uuid.UUID
	Method   string // "DOUBLES" | "PAYMENT" | "CARD"
}

// --- Trade ---

type TradeProposedEvent struct {
	BaseEvent
	TradeID         uuid.UUID
	ProposerID      uuid.UUID
	RecipientID     uuid.UUID
	OfferedMoney    int
	OfferedSpaces   []int
	RequestedMoney  int
	RequestedSpaces []int
}

type TradeAcceptedEvent struct {
	BaseEvent
	TradeID uuid.UUID
}

type TradeRejectedEvent struct {
	BaseEvent
	TradeID uuid.UUID
}

// --- Bankruptcy ---

type BankruptcyDeclaredEvent struct {
	BaseEvent
	PlayerID uuid.UUID
	// CreditorID is the player who is owed money; uuid.Nil means the bank.
	CreditorID uuid.UUID
}

// --- Cards ---

type ChanceCardDrawnEvent struct {
	BaseEvent
	PlayerID uuid.UUID
	CardID   string
}

type CommunityChestCardDrawnEvent struct {
	BaseEvent
	PlayerID uuid.UUID
	CardID   string
}

// --- Tax ---

type TaxPaidEvent struct {
	BaseEvent
	PlayerID   uuid.UUID
	SpaceIndex int
	Amount     int
}

// --- Phase 3: auctions, debts, resignation ---

type AuctionPassedEvent struct {
	BaseEvent
	PlayerID uuid.UUID
}

// DebtIncurredEvent: a charge the debtor could not cover. CreditorID uuid.Nil = bank.
type DebtIncurredEvent struct {
	BaseEvent
	DebtorID   uuid.UUID
	CreditorID uuid.UUID
	Amount     int
	Reason     string
}

type DebtPaidEvent struct {
	BaseEvent
	DebtorID   uuid.UUID
	CreditorID uuid.UUID
	Amount     int
}

type PlayerResignedEvent struct {
	BaseEvent
	PlayerID uuid.UUID
	Reason   string
}

package commands

import "github.com/google/uuid"

type CommandType string

const (
	CommandCreateGame           CommandType = "CREATE_GAME"
	CommandJoinGame             CommandType = "JOIN_GAME"
	CommandStartGame            CommandType = "START_GAME"
	CommandRollDice             CommandType = "ROLL_DICE"
	CommandBuyProperty          CommandType = "BUY_PROPERTY"
	CommandDeclinePurchase      CommandType = "DECLINE_PURCHASE"
	CommandEndTurn              CommandType = "END_TURN"
	CommandBuildHouse           CommandType = "BUILD_HOUSE"
	CommandBuildHotel           CommandType = "BUILD_HOTEL"
	CommandSellHouse            CommandType = "SELL_HOUSE"
	CommandSellHotel            CommandType = "SELL_HOTEL"
	CommandMortgageProperty     CommandType = "MORTGAGE_PROPERTY"
	CommandUnmortgageProperty   CommandType = "UNMORTGAGE_PROPERTY"
	CommandTradePropose         CommandType = "TRADE_PROPOSE"
	CommandTradeAccept          CommandType = "TRADE_ACCEPT"
	CommandTradeReject          CommandType = "TRADE_REJECT"
	CommandAuctionBid           CommandType = "AUCTION_BID"
	CommandPayJailFine          CommandType = "PAY_JAIL_FINE"
	CommandUseGetOutOfJailCard  CommandType = "USE_GET_OUT_OF_JAIL_CARD"
	CommandDeclareBankruptcy    CommandType = "DECLARE_BANKRUPTCY"
)

// Command is the base interface for all player-issued game commands.
// Commands are intentions — they can fail validation.
type Command interface {
	GetType() CommandType
	GetGameID() uuid.UUID
	GetPlayerID() uuid.UUID
}

// BaseCommand contains fields common to every command.
type BaseCommand struct {
	GameID   uuid.UUID
	PlayerID uuid.UUID
}

func (c BaseCommand) GetGameID() uuid.UUID   { return c.GameID }
func (c BaseCommand) GetPlayerID() uuid.UUID { return c.PlayerID }

// --- Game lifecycle ---

type CreateGameCommand struct {
	BaseCommand
	Name string
}

func (c CreateGameCommand) GetType() CommandType { return CommandCreateGame }

type JoinGameCommand struct {
	BaseCommand
	Name string
}

func (c JoinGameCommand) GetType() CommandType { return CommandJoinGame }

type StartGameCommand struct {
	BaseCommand
}

func (c StartGameCommand) GetType() CommandType { return CommandStartGame }

// --- Turn actions ---

// RollDiceCommand carries externally injected dice values.
// The runtime generates randomness and injects it here so reducers stay pure.
type RollDiceCommand struct {
	BaseCommand
	Die1 int
	Die2 int
}

func (c RollDiceCommand) GetType() CommandType { return CommandRollDice }

type BuyPropertyCommand struct {
	BaseCommand
}

func (c BuyPropertyCommand) GetType() CommandType { return CommandBuyProperty }

type DeclinePurchaseCommand struct {
	BaseCommand
}

func (c DeclinePurchaseCommand) GetType() CommandType { return CommandDeclinePurchase }

type EndTurnCommand struct {
	BaseCommand
}

func (c EndTurnCommand) GetType() CommandType { return CommandEndTurn }

// --- Buildings ---

type BuildHouseCommand struct {
	BaseCommand
	SpaceIndex int
}

func (c BuildHouseCommand) GetType() CommandType { return CommandBuildHouse }

type BuildHotelCommand struct {
	BaseCommand
	SpaceIndex int
}

func (c BuildHotelCommand) GetType() CommandType { return CommandBuildHotel }

type SellHouseCommand struct {
	BaseCommand
	SpaceIndex int
}

func (c SellHouseCommand) GetType() CommandType { return CommandSellHouse }

type SellHotelCommand struct {
	BaseCommand
	SpaceIndex int
}

func (c SellHotelCommand) GetType() CommandType { return CommandSellHotel }

// --- Mortgage ---

type MortgagePropertyCommand struct {
	BaseCommand
	SpaceIndex int
}

func (c MortgagePropertyCommand) GetType() CommandType { return CommandMortgageProperty }

type UnmortgagePropertyCommand struct {
	BaseCommand
	SpaceIndex int
}

func (c UnmortgagePropertyCommand) GetType() CommandType { return CommandUnmortgageProperty }

// --- Trade ---

type TradeProposeCommand struct {
	BaseCommand
	RecipientID     uuid.UUID
	OfferedMoney    int
	OfferedSpaces   []int
	RequestedMoney  int
	RequestedSpaces []int
}

func (c TradeProposeCommand) GetType() CommandType { return CommandTradePropose }

type TradeAcceptCommand struct {
	BaseCommand
	TradeID uuid.UUID
}

func (c TradeAcceptCommand) GetType() CommandType { return CommandTradeAccept }

type TradeRejectCommand struct {
	BaseCommand
	TradeID uuid.UUID
}

func (c TradeRejectCommand) GetType() CommandType { return CommandTradeReject }

// --- Auction ---

type AuctionBidCommand struct {
	BaseCommand
	Amount int
}

func (c AuctionBidCommand) GetType() CommandType { return CommandAuctionBid }

// --- Jail ---

type PayJailFineCommand struct {
	BaseCommand
}

func (c PayJailFineCommand) GetType() CommandType { return CommandPayJailFine }

type UseGetOutOfJailCardCommand struct {
	BaseCommand
}

func (c UseGetOutOfJailCardCommand) GetType() CommandType { return CommandUseGetOutOfJailCard }

// --- Bankruptcy ---

type DeclareBankruptcyCommand struct {
	BaseCommand
}

func (c DeclareBankruptcyCommand) GetType() CommandType { return CommandDeclareBankruptcy }

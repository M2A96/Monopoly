package validation

import (
	"fmt"

	"github.com/google/uuid"
	"github/M2A96/Monopoly.git/internal/game/commands"
	"github/M2A96/Monopoly.git/internal/game/models"
)

type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error on %s: %s", e.Field, e.Message)
}

type Validator struct{}

func NewValidator() *Validator {
	return &Validator{}
}

func (v *Validator) Validate(state models.GameState, cmd commands.Command) error {
	if err := validateGameState(state); err != nil {
		return err
	}

	switch c := cmd.(type) {
	case *commands.CreateGameCommand:
		return v.validateCreateGame(c)
	case *commands.JoinGameCommand:
		return v.validateJoinGame(state, c)
	case *commands.StartGameCommand:
		return v.validateStartGame(state)
	case *commands.RollDiceCommand:
		return v.validateRollDice(state, c)
	case *commands.BuyPropertyCommand:
		return v.validateBuyProperty(state, c)
	case *commands.DeclinePurchaseCommand:
		return v.validateDeclinePurchase(state, c)
	case *commands.EndTurnCommand:
		return v.validateEndTurn(state, c)
	case *commands.BuildHouseCommand:
		return v.validateBuildHouse(state, c)
	case *commands.BuildHotelCommand:
		return v.validateBuildHotel(state, c)
	case *commands.SellHouseCommand:
		return v.validateSellHouse(state, c)
	case *commands.SellHotelCommand:
		return v.validateSellHotel(state, c)
	case *commands.MortgagePropertyCommand:
		return v.validateMortgageProperty(state, c)
	case *commands.UnmortgagePropertyCommand:
		return v.validateUnmortgageProperty(state, c)
	case *commands.TradeProposeCommand:
		return v.validateTradePropose(state, c)
	case *commands.TradeAcceptCommand:
		return v.validateTradeAccept(state, c)
	case *commands.TradeRejectCommand:
		return v.validateTradeReject(state, c)
	case *commands.AuctionBidCommand:
		return v.validateAuctionBid(state, c)
	case *commands.PayJailFineCommand:
		return v.validatePayJailFine(state, c)
	case *commands.UseGetOutOfJailCardCommand:
		return v.validateUseGetOutOfJailCard(state, c)
	case *commands.DeclareBankruptcyCommand:
		return v.validateDeclareBankruptcy(state, c)
	default:
		return fmt.Errorf("unknown command type")
	}
}

func validateGameState(state models.GameState) error {
	if state.Ownerships == nil {
		return &ValidationError{Field: "state", Message: "ownerships map not initialized"}
	}
	return nil
}

func (v *Validator) validateCreateGame(c *commands.CreateGameCommand) error {
	if c.Name == "" {
		return &ValidationError{Field: "name", Message: "game name required"}
	}
	return nil
}

func (v *Validator) validateJoinGame(state models.GameState, c *commands.JoinGameCommand) error {
	if state.Status != models.GameStatusWaiting {
		return &ValidationError{Field: "status", Message: "can only join waiting games"}
	}
	if len(state.Players) >= 8 {
		return &ValidationError{Field: "players", Message: "game is full"}
	}
	for _, p := range state.Players {
		if p.PlayerID == c.PlayerID {
			return &ValidationError{Field: "player", Message: "already in game"}
		}
	}
	return nil
}

func (v *Validator) validateStartGame(state models.GameState) error {
	if state.Status != models.GameStatusWaiting {
		return &ValidationError{Field: "status", Message: "game not in waiting status"}
	}
	if len(state.Players) < 2 {
		return &ValidationError{Field: "players", Message: "need at least 2 players"}
	}
	return nil
}

func (v *Validator) validateRollDice(state models.GameState, c *commands.RollDiceCommand) error {
	if state.Status != models.GameStatusActive {
		return &ValidationError{Field: "status", Message: "game not active"}
	}
	if c.PlayerID != state.Turn.ActivePlayerID {
		return &ValidationError{Field: "turn", Message: "not your turn"}
	}
	if state.Turn.Phase != models.TurnPhasePreRoll {
		return &ValidationError{Field: "phase", Message: "cannot roll in current phase"}
	}
	player := findPlayer(state, c.PlayerID)
	if player == nil {
		return &ValidationError{Field: "player", Message: "not in game"}
	}
	if player.IsBankrupt {
		return &ValidationError{Field: "player", Message: "bankrupt"}
	}
	if c.Die1 < 1 || c.Die1 > 6 || c.Die2 < 1 || c.Die2 > 6 {
		return &ValidationError{Field: "dice", Message: "invalid dice values"}
	}
	return nil
}

func (v *Validator) validateBuyProperty(state models.GameState, c *commands.BuyPropertyCommand) error {
	if state.Status != models.GameStatusActive {
		return &ValidationError{Field: "status", Message: "game not active"}
	}
	if state.Turn.Phase != models.TurnPhaseAction || state.Turn.ActionWindow == nil {
		return &ValidationError{Field: "phase", Message: "no action pending"}
	}
	if state.Turn.ActionWindow.Type != models.ActionWaitingForPurchaseDecision {
		return &ValidationError{Field: "action", Message: "no purchase decision pending"}
	}
	if state.Turn.ActionWindow.PlayerID != c.PlayerID {
		return &ValidationError{Field: "player", Message: "not your purchase decision"}
	}
	return nil
}

func (v *Validator) validateDeclinePurchase(state models.GameState, c *commands.DeclinePurchaseCommand) error {
	if state.Status != models.GameStatusActive {
		return &ValidationError{Field: "status", Message: "game not active"}
	}
	if state.Turn.Phase != models.TurnPhaseAction || state.Turn.ActionWindow == nil {
		return &ValidationError{Field: "phase", Message: "no action pending"}
	}
	if state.Turn.ActionWindow.Type != models.ActionWaitingForPurchaseDecision {
		return &ValidationError{Field: "action", Message: "no purchase decision pending"}
	}
	if state.Turn.ActionWindow.PlayerID != c.PlayerID {
		return &ValidationError{Field: "player", Message: "not your purchase decision"}
	}
	return nil
}

func (v *Validator) validateEndTurn(state models.GameState, c *commands.EndTurnCommand) error {
	if state.Status != models.GameStatusActive {
		return &ValidationError{Field: "status", Message: "game not active"}
	}
	if c.PlayerID != state.Turn.ActivePlayerID {
		return &ValidationError{Field: "turn", Message: "not your turn"}
	}
	player := findPlayer(state, c.PlayerID)
	if player == nil {
		return &ValidationError{Field: "player", Message: "not in game"}
	}
	if player.IsBankrupt {
		return &ValidationError{Field: "player", Message: "bankrupt"}
	}
	return nil
}

func (v *Validator) validateBuildHouse(state models.GameState, c *commands.BuildHouseCommand) error {
	if state.Status != models.GameStatusActive {
		return &ValidationError{Field: "status", Message: "game not active"}
	}
	o, ok := state.Ownerships[c.SpaceIndex]
	if !ok || o.OwnerID != c.PlayerID {
		return &ValidationError{Field: "ownership", Message: "not your property"}
	}
	if o.Mortgaged {
		return &ValidationError{Field: "mortgaged", Message: "property is mortgaged"}
	}
	if o.Buildings >= 4 {
		return &ValidationError{Field: "buildings", Message: "use build hotel instead"}
	}
	return nil
}

func (v *Validator) validateBuildHotel(state models.GameState, c *commands.BuildHotelCommand) error {
	if state.Status != models.GameStatusActive {
		return &ValidationError{Field: "status", Message: "game not active"}
	}
	o, ok := state.Ownerships[c.SpaceIndex]
	if !ok || o.OwnerID != c.PlayerID {
		return &ValidationError{Field: "ownership", Message: "not your property"}
	}
	if o.Buildings != 4 {
		return &ValidationError{Field: "buildings", Message: "need 4 houses first"}
	}
	return nil
}

func (v *Validator) validateSellHouse(state models.GameState, c *commands.SellHouseCommand) error {
	if state.Status != models.GameStatusActive {
		return &ValidationError{Field: "status", Message: "game not active"}
	}
	o, ok := state.Ownerships[c.SpaceIndex]
	if !ok || o.OwnerID != c.PlayerID {
		return &ValidationError{Field: "ownership", Message: "not your property"}
	}
	if o.Buildings <= 0 {
		return &ValidationError{Field: "buildings", Message: "no houses to sell"}
	}
	return nil
}

func (v *Validator) validateSellHotel(state models.GameState, c *commands.SellHotelCommand) error {
	if state.Status != models.GameStatusActive {
		return &ValidationError{Field: "status", Message: "game not active"}
	}
	o, ok := state.Ownerships[c.SpaceIndex]
	if !ok || o.OwnerID != c.PlayerID {
		return &ValidationError{Field: "ownership", Message: "not your property"}
	}
	if o.Buildings != models.HotelLevel {
		return &ValidationError{Field: "buildings", Message: "no hotel to sell"}
	}
	return nil
}

func (v *Validator) validateMortgageProperty(state models.GameState, c *commands.MortgagePropertyCommand) error {
	if state.Status != models.GameStatusActive {
		return &ValidationError{Field: "status", Message: "game not active"}
	}
	o, ok := state.Ownerships[c.SpaceIndex]
	if !ok || o.OwnerID != c.PlayerID {
		return &ValidationError{Field: "ownership", Message: "not your property"}
	}
	if o.Mortgaged {
		return &ValidationError{Field: "mortgaged", Message: "already mortgaged"}
	}
	if o.Buildings > 0 {
		return &ValidationError{Field: "buildings", Message: "must sell buildings first"}
	}
	return nil
}

func (v *Validator) validateUnmortgageProperty(state models.GameState, c *commands.UnmortgagePropertyCommand) error {
	if state.Status != models.GameStatusActive {
		return &ValidationError{Field: "status", Message: "game not active"}
	}
	o, ok := state.Ownerships[c.SpaceIndex]
	if !ok || o.OwnerID != c.PlayerID {
		return &ValidationError{Field: "ownership", Message: "not your property"}
	}
	if !o.Mortgaged {
		return &ValidationError{Field: "mortgaged", Message: "not mortgaged"}
	}
	return nil
}

func (v *Validator) validateTradePropose(state models.GameState, c *commands.TradeProposeCommand) error {
	if state.Status != models.GameStatusActive {
		return &ValidationError{Field: "status", Message: "game not active"}
	}
	if c.PlayerID == c.RecipientID {
		return &ValidationError{Field: "player", Message: "cannot trade with yourself"}
	}
	return nil
}

func (v *Validator) validateTradeAccept(state models.GameState, c *commands.TradeAcceptCommand) error {
	if state.Status != models.GameStatusActive {
		return &ValidationError{Field: "status", Message: "game not active"}
	}
	return nil
}

func (v *Validator) validateTradeReject(state models.GameState, c *commands.TradeRejectCommand) error {
	if state.Status != models.GameStatusActive {
		return &ValidationError{Field: "status", Message: "game not active"}
	}
	return nil
}

func (v *Validator) validateAuctionBid(state models.GameState, c *commands.AuctionBidCommand) error {
	if state.Status != models.GameStatusActive {
		return &ValidationError{Field: "status", Message: "game not active"}
	}
	if state.Auction == nil {
		return &ValidationError{Field: "auction", Message: "no active auction"}
	}
	return nil
}

func (v *Validator) validatePayJailFine(state models.GameState, c *commands.PayJailFineCommand) error {
	if state.Status != models.GameStatusActive {
		return &ValidationError{Field: "status", Message: "game not active"}
	}
	player := findPlayer(state, c.PlayerID)
	if player == nil {
		return &ValidationError{Field: "player", Message: "not in game"}
	}
	if !player.IsInJail {
		return &ValidationError{Field: "jail", Message: "not in jail"}
	}
	return nil
}

func (v *Validator) validateUseGetOutOfJailCard(state models.GameState, c *commands.UseGetOutOfJailCardCommand) error {
	if state.Status != models.GameStatusActive {
		return &ValidationError{Field: "status", Message: "game not active"}
	}
	player := findPlayer(state, c.PlayerID)
	if player == nil {
		return &ValidationError{Field: "player", Message: "not in game"}
	}
	if !player.IsInJail {
		return &ValidationError{Field: "jail", Message: "not in jail"}
	}
	if player.GetOutOfJailCards <= 0 {
		return &ValidationError{Field: "cards", Message: "no get out of jail free cards"}
	}
	return nil
}

func (v *Validator) validateDeclareBankruptcy(state models.GameState, c *commands.DeclareBankruptcyCommand) error {
	if state.Status != models.GameStatusActive {
		return &ValidationError{Field: "status", Message: "game not active"}
	}
	player := findPlayer(state, c.PlayerID)
	if player == nil {
		return &ValidationError{Field: "player", Message: "not in game"}
	}
	if player.IsBankrupt {
		return &ValidationError{Field: "player", Message: "already bankrupt"}
	}
	return nil
}

// findPlayer returns a read-only view of the player; validators never mutate state.
func findPlayer(state models.GameState, id uuid.UUID) *models.PlayerState {
	for i := range state.Players {
		if state.Players[i].PlayerID == id {
			return &state.Players[i]
		}
	}
	return nil
}

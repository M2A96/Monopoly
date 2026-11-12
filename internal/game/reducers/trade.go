package reducers

import (
	"fmt"

	"github.com/google/uuid"
	"github/M2A96/Monopoly.git/internal/game/commands"
	"github/M2A96/Monopoly.git/internal/game/events"
	"github/M2A96/Monopoly.git/internal/game/models"
)

func reduceTradePropose(state models.GameState, cmd commands.Command) (models.GameState, []events.Event, error) {
	c, ok := cmd.(*commands.TradeProposeCommand)
	if !ok {
		return state, nil, fmt.Errorf("expected TradeProposeCommand, got %T", cmd)
	}

	if state.Status != models.GameStatusActive {
		return state, nil, fmt.Errorf("game not active")
	}

	if c.PlayerID == c.RecipientID {
		return state, nil, fmt.Errorf("cannot trade with yourself")
	}

	proposer := findPlayer(&state, c.PlayerID)
	if proposer == nil || proposer.IsBankrupt {
		return state, nil, fmt.Errorf("proposer not valid")
	}

	recipient := findPlayer(&state, c.RecipientID)
	if recipient == nil || recipient.IsBankrupt {
		return state, nil, fmt.Errorf("recipient not valid")
	}

	if proposer.Money < c.OfferedMoney {
		return state, nil, fmt.Errorf("proposer has insufficient funds")
	}

	for _, idx := range c.OfferedSpaces {
		if !isOwner(&state, c.PlayerID, idx) {
			return state, nil, fmt.Errorf("proposer does not own space %d", idx)
		}
		o := state.Ownerships[idx]
		if o.Buildings > 0 {
			return state, nil, fmt.Errorf("cannot trade property with buildings")
		}
	}

	for _, idx := range c.RequestedSpaces {
		if !isOwner(&state, c.RecipientID, idx) {
			return state, nil, fmt.Errorf("recipient does not own space %d", idx)
		}
		o := state.Ownerships[idx]
		if o.Buildings > 0 {
			return state, nil, fmt.Errorf("cannot trade property with buildings")
		}
	}

	if recipient.Money < c.RequestedMoney {
		return state, nil, fmt.Errorf("recipient has insufficient funds")
	}

	state = copyState(state)
	proposer = findPlayer(&state, c.PlayerID)
	recipient = findPlayer(&state, c.RecipientID)

	// Derived from game ID + trade count so replay yields the same ID.
	tradeID := uuid.NewSHA1(state.GameID, []byte(fmt.Sprintf("trade:%d", len(state.Trades))))
	trade := models.Trade{
		TradeID:         tradeID,
		ProposerID:      c.PlayerID,
		RecipientID:     c.RecipientID,
		OfferedMoney:    c.OfferedMoney,
		OfferedSpaces:   c.OfferedSpaces,
		RequestedMoney:  c.RequestedMoney,
		RequestedSpaces: c.RequestedSpaces,
		Status:          models.TradeStatusPending,
	}
	state.Trades = append(state.Trades, trade)

	recipient.PendingActions = append(recipient.PendingActions, models.ActionWindow{
		Type:     models.ActionWaitingForTradeResponse,
		PlayerID: c.RecipientID,
	})

	return state, []events.Event{
		&events.TradeProposedEvent{
			BaseEvent:       events.BaseEvent{Type: events.EventTradeProposed},
			TradeID:         tradeID,
			ProposerID:      c.PlayerID,
			RecipientID:     c.RecipientID,
			OfferedMoney:    c.OfferedMoney,
			OfferedSpaces:   c.OfferedSpaces,
			RequestedMoney:  c.RequestedMoney,
			RequestedSpaces: c.RequestedSpaces,
		},
	}, nil
}

func reduceTradeAccept(state models.GameState, cmd commands.Command) (models.GameState, []events.Event, error) {
	c, ok := cmd.(*commands.TradeAcceptCommand)
	if !ok {
		return state, nil, fmt.Errorf("expected TradeAcceptCommand, got %T", cmd)
	}

	if state.Status != models.GameStatusActive {
		return state, nil, fmt.Errorf("game not active")
	}

	tradeIdx := -1
	for i, t := range state.Trades {
		if t.TradeID == c.TradeID {
			tradeIdx = i
			break
		}
	}

	if tradeIdx == -1 {
		return state, nil, fmt.Errorf("trade not found")
	}

	trade := &state.Trades[tradeIdx]
	if trade.Status != models.TradeStatusPending {
		return state, nil, fmt.Errorf("trade is not pending")
	}

	if trade.RecipientID != c.PlayerID {
		return state, nil, fmt.Errorf("you are not the recipient of this trade")
	}

	state = copyState(state)
	trade = &state.Trades[tradeIdx]

	proposer := findPlayer(&state, trade.ProposerID)
	recipient := findPlayer(&state, trade.RecipientID)

	if proposer.Money < trade.OfferedMoney || recipient.Money < trade.RequestedMoney {
		return state, nil, fmt.Errorf("insufficient funds for trade")
	}

	proposer.Money -= trade.OfferedMoney
	recipient.Money += trade.OfferedMoney
	recipient.Money -= trade.RequestedMoney
	proposer.Money += trade.RequestedMoney

	for _, idx := range trade.OfferedSpaces {
		o := state.Ownerships[idx]
		o.OwnerID = trade.RecipientID
		state.Ownerships[idx] = o
		removeSpaceIndex(proposer, idx)
		addSpaceIndex(recipient, idx)
	}

	for _, idx := range trade.RequestedSpaces {
		o := state.Ownerships[idx]
		o.OwnerID = trade.ProposerID
		state.Ownerships[idx] = o
		removeSpaceIndex(recipient, idx)
		addSpaceIndex(proposer, idx)
	}

	trade.Status = models.TradeStatusAccepted

	for i := range state.Players {
		newPending := state.Players[i].PendingActions[:0]
		for _, pw := range state.Players[i].PendingActions {
			if pw.Type != models.ActionWaitingForTradeResponse {
				newPending = append(newPending, pw)
			}
		}
		state.Players[i].PendingActions = newPending
	}

	return state, []events.Event{
		&events.TradeAcceptedEvent{
			BaseEvent: events.BaseEvent{Type: events.EventTradeAccepted},
			TradeID:   c.TradeID,
		},
	}, nil
}

func reduceTradeReject(state models.GameState, cmd commands.Command) (models.GameState, []events.Event, error) {
	c, ok := cmd.(*commands.TradeRejectCommand)
	if !ok {
		return state, nil, fmt.Errorf("expected TradeRejectCommand, got %T", cmd)
	}

	if state.Status != models.GameStatusActive {
		return state, nil, fmt.Errorf("game not active")
	}

	tradeIdx := -1
	for i, t := range state.Trades {
		if t.TradeID == c.TradeID {
			tradeIdx = i
			break
		}
	}

	if tradeIdx == -1 {
		return state, nil, fmt.Errorf("trade not found")
	}

	trade := &state.Trades[tradeIdx]
	if trade.Status != models.TradeStatusPending {
		return state, nil, fmt.Errorf("trade is not pending")
	}

	if trade.RecipientID != c.PlayerID {
		return state, nil, fmt.Errorf("you are not the recipient")
	}

	state = copyState(state)
	trade = &state.Trades[tradeIdx]
	trade.Status = models.TradeStatusRejected

	for i := range state.Players {
		newPending := state.Players[i].PendingActions[:0]
		for _, pw := range state.Players[i].PendingActions {
			if pw.Type != models.ActionWaitingForTradeResponse {
				newPending = append(newPending, pw)
			}
		}
		state.Players[i].PendingActions = newPending
	}

	return state, []events.Event{
		&events.TradeRejectedEvent{
			BaseEvent: events.BaseEvent{Type: events.EventTradeRejected},
			TradeID:   c.TradeID,
		},
	}, nil
}

// validateJailAction: jail exits by fine or card happen on the player's own turn, before rolling.
func validateJailAction(state models.GameState, playerID uuid.UUID) (*models.PlayerState, error) {
	if state.Status != models.GameStatusActive {
		return nil, fmt.Errorf("game not active")
	}
	player := findPlayer(&state, playerID)
	if player == nil {
		return nil, fmt.Errorf("player not in game")
	}
	if !player.IsInJail {
		return nil, fmt.Errorf("player is not in jail")
	}
	if state.Turn.ActivePlayerID != playerID {
		return nil, fmt.Errorf("not your turn")
	}
	if state.Turn.Phase != models.TurnPhasePreRoll {
		return nil, fmt.Errorf("can only leave jail before rolling")
	}
	return player, nil
}

func reducePayJailFine(state models.GameState, cmd commands.Command) (models.GameState, []events.Event, error) {
	c, ok := cmd.(*commands.PayJailFineCommand)
	if !ok {
		return state, nil, fmt.Errorf("expected PayJailFineCommand, got %T", cmd)
	}

	player, err := validateJailAction(state, c.PlayerID)
	if err != nil {
		return state, nil, err
	}
	if player.Money < jailFine {
		return state, nil, fmt.Errorf("insufficient funds for jail fine")
	}

	state = copyState(state)
	player = findPlayer(&state, c.PlayerID)
	player.Money -= jailFine
	release(player)

	return state, []events.Event{releasedEvent(c.PlayerID, "PAYMENT")}, nil
}

func reduceUseGetOutOfJailCard(state models.GameState, cmd commands.Command) (models.GameState, []events.Event, error) {
	c, ok := cmd.(*commands.UseGetOutOfJailCardCommand)
	if !ok {
		return state, nil, fmt.Errorf("expected UseGetOutOfJailCardCommand, got %T", cmd)
	}

	player, err := validateJailAction(state, c.PlayerID)
	if err != nil {
		return state, nil, err
	}
	if player.GetOutOfJailCards <= 0 {
		return state, nil, fmt.Errorf("no Get Out of Jail Free cards")
	}

	state = copyState(state)
	player = findPlayer(&state, c.PlayerID)
	player.GetOutOfJailCards--
	// Return the card to its deck so it can be drawn again.
	if held := heldJailCards(&state, c.PlayerID); len(held) > 0 {
		delete(state.CardDecks.JailCardHolders, held[0])
	}
	release(player)

	return state, []events.Event{releasedEvent(c.PlayerID, "CARD")}, nil
}

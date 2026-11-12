package reducers

import (
	"sort"

	"github.com/google/uuid"
	"github/M2A96/Monopoly.git/internal/game/events"
	"github/M2A96/Monopoly.git/internal/game/models"
	"github/M2A96/Monopoly.git/pkg/board"
)

// bank is the creditor/recipient ID used for payments to or from the bank.
var bank = uuid.Nil

// charge moves amount from debtor to creditor (bank when creditor == uuid.Nil).
// If the debtor cannot cover it, nothing moves and a pending Debt is recorded instead;
// the debtor must then raise funds and PayDebt, or DeclareBankruptcy.
func charge(state *models.GameState, debtorID, creditorID uuid.UUID, amount int, reason string) []events.Event {
	if amount <= 0 {
		return nil
	}
	debtor := findPlayer(state, debtorID)
	if debtor == nil || debtor.IsBankrupt {
		return nil
	}
	if debtor.Money >= amount {
		transfer(state, debtor, creditorID, amount)
		return nil
	}
	state.Debts = append(state.Debts, models.Debt{
		DebtorID:   debtorID,
		CreditorID: creditorID,
		Amount:     amount,
		Reason:     reason,
	})
	return []events.Event{&events.DebtIncurredEvent{
		BaseEvent:  events.BaseEvent{Type: events.EventDebtIncurred},
		DebtorID:   debtorID,
		CreditorID: creditorID,
		Amount:     amount,
		Reason:     reason,
	}}
}

func transfer(state *models.GameState, debtor *models.PlayerState, creditorID uuid.UUID, amount int) {
	debtor.Money -= amount
	if creditorID == bank {
		return
	}
	if creditor := findPlayer(state, creditorID); creditor != nil && !creditor.IsBankrupt {
		creditor.Money += amount
	}
}

func firstDebtOf(state *models.GameState, playerID uuid.UUID) int {
	for i, d := range state.Debts {
		if d.DebtorID == playerID {
			return i
		}
	}
	return -1
}

// settlePhase derives the turn phase after any state change that may open or close
// an obligation. Priority: debts, then auction, then a pending purchase decision.
func settlePhase(state *models.GameState) {
	if state.Status != models.GameStatusActive {
		return
	}
	if len(state.Debts) > 0 {
		state.Turn.Phase = models.TurnPhaseAction
		state.Turn.ActionWindow = &models.ActionWindow{
			Type:     models.ActionWaitingForBankruptcyResolution,
			PlayerID: state.Debts[0].DebtorID,
		}
		return
	}
	if state.Auction != nil {
		state.Turn.Phase = models.TurnPhaseAction
		state.Turn.ActionWindow = &models.ActionWindow{
			Type:       models.ActionWaitingForAuctionBid,
			PlayerID:   state.Auction.CurrentBidderID(),
			SpaceIndex: state.Auction.SpaceIndex,
		}
		return
	}
	if w := state.Turn.ActionWindow; w != nil && w.Type == models.ActionWaitingForPurchaseDecision {
		state.Turn.Phase = models.TurnPhaseAction
		return
	}
	state.Turn.ActionWindow = nil
	switch state.Turn.Phase {
	case models.TurnPhaseLand, models.TurnPhaseAction:
		state.Turn.Phase = phaseAfterResolution(state)
	}
}

// advanceTurn hands the turn to the next non-bankrupt player after fromID.
func advanceTurn(state *models.GameState, fromID uuid.UUID) []events.Event {
	nextID := nextActivePlayer(state, fromID)
	if nextID == uuid.Nil {
		return nil
	}
	turnNumber := state.Turn.Number
	if findPlayerIndex(state, nextID) <= findPlayerIndex(state, fromID) {
		turnNumber++
	}
	evts := []events.Event{&events.TurnEndedEvent{
		BaseEvent:  events.BaseEvent{Type: events.EventTurnEnded},
		PlayerID:   fromID,
		TurnNumber: state.Turn.Number,
	}}
	state.Turn = models.Turn{
		Number:         turnNumber,
		ActivePlayerID: nextID,
		Phase:          models.TurnPhasePreRoll,
	}
	return append(evts, &events.TurnStartedEvent{
		BaseEvent:  events.BaseEvent{Type: events.EventTurnStarted},
		PlayerID:   nextID,
		TurnNumber: turnNumber,
	})
}

// drawCard takes the next card from a cycling deck, skipping held jail cards.
func drawCard(deck []models.Card, index *int, holders map[string]uuid.UUID) (models.Card, bool) {
	n := len(deck)
	for tries := 0; tries < n; tries++ {
		card := deck[*index%n]
		*index = (*index + 1) % n
		if _, held := holders[card.ID]; held {
			continue
		}
		return card, true
	}
	return models.Card{}, false
}

// heldJailCards returns the jail card IDs held by a player, sorted for determinism.
func heldJailCards(state *models.GameState, playerID uuid.UUID) []string {
	var ids []string
	for id, holder := range state.CardDecks.JailCardHolders {
		if holder == playerID {
			ids = append(ids, id)
		}
	}
	sort.Strings(ids)
	return ids
}

// eliminate removes a player from play, transferring everything to creditorID
// (a player, or the bank when uuid.Nil). Buildings are always sold to the bank first.
func eliminate(state *models.GameState, playerID, creditorID uuid.UUID) []events.Event {
	player := findPlayer(state, playerID)
	creditor := findPlayer(state, creditorID)
	if creditor != nil && creditor.IsBankrupt {
		creditor, creditorID = nil, bank
	}

	cash := player.Money
	owned := append([]int(nil), player.OwnedSpaceIndexes...)
	sort.Ints(owned)
	for _, idx := range owned {
		o, ok := state.Ownerships[idx]
		if !ok {
			continue
		}
		space := board.SpaceByIndex[idx]
		if o.Buildings == models.HotelLevel {
			state.Bank.Hotels++
			cash += space.HouseCost * models.HotelLevel / 2
		} else if o.Buildings > 0 {
			state.Bank.Houses += o.Buildings
			cash += space.HouseCost * o.Buildings / 2
		}
		o.Buildings = 0
		if creditor != nil {
			o.OwnerID = creditorID
			state.Ownerships[idx] = o
			addSpaceIndex(creditor, idx)
		} else {
			delete(state.Ownerships, idx)
		}
	}
	if creditor != nil {
		creditor.Money += cash
	}

	for _, id := range heldJailCards(state, playerID) {
		if creditor != nil {
			state.CardDecks.JailCardHolders[id] = creditorID
			creditor.GetOutOfJailCards++
		} else {
			delete(state.CardDecks.JailCardHolders, id)
		}
	}

	player.Money = 0
	player.OwnedSpaceIndexes = nil
	player.GetOutOfJailCards = 0
	player.IsInJail = false
	player.JailState = models.JailState{}
	player.IsBankrupt = true

	// Debts owed by the player are settled by the elimination; debts owed to them go to the bank.
	kept := state.Debts[:0]
	for _, d := range state.Debts {
		if d.DebtorID == playerID {
			continue
		}
		if d.CreditorID == playerID {
			d.CreditorID = bank
		}
		kept = append(kept, d)
	}
	state.Debts = kept
	if len(state.Debts) == 0 {
		state.Debts = nil
	}

	for i := range state.Trades {
		t := &state.Trades[i]
		if t.Status == models.TradeStatusPending && (t.ProposerID == playerID || t.RecipientID == playerID) {
			t.Status = models.TradeStatusCanceled
		}
	}

	var evts []events.Event
	if state.Auction != nil {
		evts = append(evts, leaveAuction(state, playerID)...)
	}

	if activePlayers(state) <= 1 {
		var winnerID uuid.UUID
		for _, p := range state.Players {
			if !p.IsBankrupt {
				winnerID = p.PlayerID
				break
			}
		}
		state.Status = models.GameStatusFinished
		state.Auction = nil
		state.Debts = nil
		state.Turn.ActionWindow = nil
		return append(evts, &events.GameEndedEvent{
			BaseEvent: events.BaseEvent{Type: events.EventGameEnded},
			WinnerID:  winnerID,
		})
	}

	if state.Turn.ActivePlayerID == playerID {
		state.Auction = nil
		evts = append(evts, advanceTurn(state, playerID)...)
	}
	settlePhase(state)
	return evts
}

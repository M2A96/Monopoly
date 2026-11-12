package models

import "github.com/google/uuid"

type GameStatus string

const (
	GameStatusWaiting  GameStatus = "WAITING"
	GameStatusActive   GameStatus = "ACTIVE"
	GameStatusFinished GameStatus = "FINISHED"
)

// GameState is the canonical authoritative runtime model.
// All gameplay mutations happen exclusively through GameState.
type GameState struct {
	GameID     uuid.UUID
	Status     GameStatus
	Players    []PlayerState
	Turn       Turn
	Bank       Bank
	Ownerships map[int]Ownership // spaceIndex -> Ownership
	CardDecks  CardDecks
	Auction    *Auction
	Trades     []Trade
	Debts      []Debt // unpaid charges, resolved in order

	// EventSequence is the sequence number of the last event emitted for this game.
	// It lives in state (not in the reducer registry) so replay reproduces identical numbering.
	EventSequence int64
}

// Clone returns a deep copy: no slice, map or pointer is shared with the original.
// Reducers clone before mutating, and the runtime clones state handed to subscribers.
func (state GameState) Clone() GameState {
	cp := state
	cp.Players = make([]PlayerState, len(state.Players))
	for i := range state.Players {
		cp.Players[i] = state.Players[i]
		cp.Players[i].OwnedSpaceIndexes = make([]int, len(state.Players[i].OwnedSpaceIndexes))
		copy(cp.Players[i].OwnedSpaceIndexes, state.Players[i].OwnedSpaceIndexes)
		cp.Players[i].PendingActions = make([]ActionWindow, len(state.Players[i].PendingActions))
		copy(cp.Players[i].PendingActions, state.Players[i].PendingActions)
	}
	cp.Ownerships = make(map[int]Ownership, len(state.Ownerships))
	for k, v := range state.Ownerships {
		cp.Ownerships[k] = v
	}
	cp.Trades = make([]Trade, len(state.Trades))
	copy(cp.Trades, state.Trades)
	if state.Debts != nil {
		cp.Debts = make([]Debt, len(state.Debts))
		copy(cp.Debts, state.Debts)
	}
	if state.Auction != nil {
		a := *state.Auction
		a.Bids = make(map[uuid.UUID]int, len(state.Auction.Bids))
		for k, v := range state.Auction.Bids {
			a.Bids[k] = v
		}
		a.Participants = append([]uuid.UUID(nil), state.Auction.Participants...)
		cp.Auction = &a
	}
	if state.Turn.ActionWindow != nil {
		w := *state.Turn.ActionWindow
		cp.Turn.ActionWindow = &w
	}
	if state.CardDecks.JailCardHolders != nil {
		cp.CardDecks.JailCardHolders = make(map[string]uuid.UUID, len(state.CardDecks.JailCardHolders))
		for k, v := range state.CardDecks.JailCardHolders {
			cp.CardDecks.JailCardHolders[k] = v
		}
	}
	cp.CardDecks.Chance = make([]Card, len(state.CardDecks.Chance))
	copy(cp.CardDecks.Chance, state.CardDecks.Chance)
	cp.CardDecks.CommunityChest = make([]Card, len(state.CardDecks.CommunityChest))
	copy(cp.CardDecks.CommunityChest, state.CardDecks.CommunityChest)
	return cp
}

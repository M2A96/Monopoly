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

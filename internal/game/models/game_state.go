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
}

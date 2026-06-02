package models

import "github.com/google/uuid"

type PlayerState struct {
	PlayerID          uuid.UUID
	Name              string
	Money             int
	Position          int // 0-39 board space index
	IsInJail          bool
	JailState         JailState
	IsBankrupt        bool
	OwnedSpaceIndexes []int // convenience cache; authoritative data is in GameState.Ownerships
	GetOutOfJailCards int   // count of Get Out of Jail Free cards held
	PendingActions    []ActionWindow
}

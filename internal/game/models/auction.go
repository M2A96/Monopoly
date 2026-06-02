package models

import "github.com/google/uuid"

// Auction represents an active property auction.
// An auction starts when a player lands on an unowned property and declines to buy it.
type Auction struct {
	SpaceIndex int
	Bids       map[uuid.UUID]int // playerID -> current bid amount
	HighestBid int
	WinnerID   uuid.UUID
}

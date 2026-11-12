package models

import "github.com/google/uuid"

// Auction represents an active property auction.
// An auction starts when a player lands on an unowned property and declines to buy it.
//
// Bidding is round-robin so the outcome is deterministic: Participants holds the players
// still in the auction in turn order, and CurrentBidder indexes the one who must act next.
type Auction struct {
	SpaceIndex      int
	Bids            map[uuid.UUID]int // playerID -> highest bid placed by that player
	HighestBid      int
	HighestBidderID uuid.UUID // uuid.Nil until the first bid
	Participants    []uuid.UUID
	CurrentBidder   int
	WinnerID        uuid.UUID
}

// CurrentBidderID returns the player who must bid or pass next.
func (a *Auction) CurrentBidderID() uuid.UUID {
	if a == nil || len(a.Participants) == 0 {
		return uuid.Nil
	}
	return a.Participants[a.CurrentBidder%len(a.Participants)]
}

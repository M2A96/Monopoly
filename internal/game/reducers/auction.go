package reducers

import (
	"fmt"

	"github.com/google/uuid"
	"github/M2A96/Monopoly.git/internal/game/commands"
	"github/M2A96/Monopoly.git/internal/game/events"
	"github/M2A96/Monopoly.git/internal/game/models"
)

// startAuction opens a round-robin auction for spaceIndex. Every non-bankrupt player
// takes part, in turn order, starting with startID (the player who declined).
func startAuction(state *models.GameState, spaceIndex int, startID uuid.UUID) []events.Event {
	start := findPlayerIndex(state, startID)
	var participants []uuid.UUID
	for i := 0; i < len(state.Players); i++ {
		p := state.Players[(start+i)%len(state.Players)]
		if !p.IsBankrupt {
			participants = append(participants, p.PlayerID)
		}
	}
	state.Auction = &models.Auction{
		SpaceIndex:   spaceIndex,
		Bids:         make(map[uuid.UUID]int),
		Participants: participants,
	}
	state.Turn.ActionWindow = nil
	settlePhase(state)
	return []events.Event{&events.AuctionStartedEvent{
		BaseEvent:  events.BaseEvent{Type: events.EventAuctionStarted},
		SpaceIndex: spaceIndex,
	}}
}

func validateAuctionTurn(state models.GameState, playerID uuid.UUID) error {
	if state.Status != models.GameStatusActive {
		return fmt.Errorf("game not active")
	}
	if state.Auction == nil {
		return fmt.Errorf("no active auction")
	}
	if len(state.Debts) > 0 {
		return fmt.Errorf("pending debts must be resolved first")
	}
	if state.Auction.CurrentBidderID() != playerID {
		return fmt.Errorf("not your turn to bid")
	}
	return nil
}

func reduceAuctionBid(state models.GameState, cmd commands.Command) (models.GameState, []events.Event, error) {
	c, ok := cmd.(*commands.AuctionBidCommand)
	if !ok {
		return state, nil, fmt.Errorf("expected AuctionBidCommand, got %T", cmd)
	}
	if err := validateAuctionTurn(state, c.PlayerID); err != nil {
		return state, nil, err
	}
	if c.Amount <= state.Auction.HighestBid {
		return state, nil, fmt.Errorf("bid must be higher than current highest bid (%d)", state.Auction.HighestBid)
	}
	player := findPlayer(&state, c.PlayerID)
	if player.Money < c.Amount {
		return state, nil, fmt.Errorf("insufficient funds: have %d, bid %d", player.Money, c.Amount)
	}

	state = copyState(state)
	a := state.Auction
	a.Bids[c.PlayerID] = c.Amount
	a.HighestBid = c.Amount
	a.HighestBidderID = c.PlayerID
	advanceBidder(a, a.CurrentBidder+1)

	evts := []events.Event{&events.AuctionBidPlacedEvent{
		BaseEvent: events.BaseEvent{Type: events.EventAuctionBidPlaced},
		PlayerID:  c.PlayerID,
		Amount:    c.Amount,
	}}
	evts = append(evts, maybeCloseAuction(&state)...)
	return state, evts, nil
}

func reduceAuctionPass(state models.GameState, cmd commands.Command) (models.GameState, []events.Event, error) {
	c, ok := cmd.(*commands.AuctionPassCommand)
	if !ok {
		return state, nil, fmt.Errorf("expected AuctionPassCommand, got %T", cmd)
	}
	if err := validateAuctionTurn(state, c.PlayerID); err != nil {
		return state, nil, err
	}
	if state.Auction.HighestBidderID == c.PlayerID {
		return state, nil, fmt.Errorf("highest bidder cannot pass")
	}

	state = copyState(state)
	evts := []events.Event{&events.AuctionPassedEvent{
		BaseEvent: events.BaseEvent{Type: events.EventAuctionPassed},
		PlayerID:  c.PlayerID,
	}}
	evts = append(evts, leaveAuction(&state, c.PlayerID)...)
	return state, evts, nil
}

// leaveAuction removes a participant (pass or elimination) and closes the auction if decided.
func leaveAuction(state *models.GameState, playerID uuid.UUID) []events.Event {
	a := state.Auction
	idx := -1
	for i, id := range a.Participants {
		if id == playerID {
			idx = i
			break
		}
	}
	if idx == -1 {
		return nil
	}
	cur := a.CurrentBidder
	a.Participants = append(a.Participants[:idx], a.Participants[idx+1:]...)
	if idx < cur {
		cur--
	}

	if a.HighestBidderID == playerID {
		// An eliminated highest bidder's bid is void; fall back to the best remaining bid.
		a.HighestBid, a.HighestBidderID = 0, uuid.Nil
		for _, id := range a.Participants {
			if b := a.Bids[id]; b > a.HighestBid {
				a.HighestBid, a.HighestBidderID = b, id
			}
		}
	}
	advanceBidder(a, cur)
	return maybeCloseAuction(state)
}

// advanceBidder points CurrentBidder at position from (wrapping), skipping the highest
// bidder while anyone else is still in: the leader never has to outbid themselves.
func advanceBidder(a *models.Auction, from int) {
	n := len(a.Participants)
	if n == 0 {
		a.CurrentBidder = 0
		return
	}
	a.CurrentBidder = ((from % n) + n) % n
	if n > 1 && a.Participants[a.CurrentBidder] == a.HighestBidderID {
		a.CurrentBidder = (a.CurrentBidder + 1) % n
	}
}

// maybeCloseAuction ends the auction when only the highest bidder is left,
// or when everyone has passed without bidding (the property stays with the bank).
func maybeCloseAuction(state *models.GameState) []events.Event {
	a := state.Auction
	won := len(a.Participants) == 1 && a.Participants[0] == a.HighestBidderID
	if !won && len(a.Participants) > 0 {
		settlePhase(state)
		return nil
	}

	state.Auction = nil
	ended := &events.AuctionEndedEvent{
		BaseEvent:  events.BaseEvent{Type: events.EventAuctionEnded},
		SpaceIndex: a.SpaceIndex,
	}
	evts := []events.Event{ended}
	if won {
		ended.WinnerID = a.HighestBidderID
		ended.Amount = a.HighestBid
		winner := findPlayer(state, a.HighestBidderID)
		state.Ownerships[a.SpaceIndex] = models.Ownership{SpaceIndex: a.SpaceIndex, OwnerID: winner.PlayerID}
		addSpaceIndex(winner, a.SpaceIndex)
		evts = append(evts, charge(state, winner.PlayerID, bank, a.HighestBid, "AUCTION")...)
	}
	settlePhase(state)
	return evts
}

package replay_test

import (
	"sort"
	"testing"

	"github.com/google/uuid"
	"github/M2A96/Monopoly.git/internal/game/commands"
	"github/M2A96/Monopoly.git/internal/game/models"
	"github/M2A96/Monopoly.git/internal/game/reducers"
	"github/M2A96/Monopoly.git/pkg/board"
	"github/M2A96/Monopoly.git/pkg/dice"
)

// bot drives a game the way the runtime will: it reads state, picks commands with a
// fixed policy, and injects dice from a seeded provider. Candidates are tried in a
// fixed order and the first one the engine accepts is played, so games are reproducible.
type bot struct {
	r     *reducers.Registry
	dice  dice.Provider
	state models.GameState
	cmds  []commands.Command
}

func newBot(seed int64, setup []commands.Command) (*bot, error) {
	b := &bot{r: reducers.NewRegistry(), dice: dice.NewProvider(seed), state: emptyState()}
	for _, c := range setup {
		if !b.try(c) {
			return nil, errSetup
		}
	}
	return b, nil
}

type setupErr struct{}

func (setupErr) Error() string { return "setup command rejected" }

var errSetup error = setupErr{}

func (b *bot) try(cmd commands.Command) bool {
	next, _, err := b.r.Dispatch(b.state, cmd)
	if err != nil {
		return false
	}
	b.state = next
	b.cmds = append(b.cmds, cmd)
	return true
}

func (b *bot) first(cands ...commands.Command) bool {
	for _, c := range cands {
		if b.try(c) {
			return true
		}
	}
	return false
}

// step plays one command. It returns false if the bot is stuck (a bug in the engine or policy).
func (b *bot) step() bool {
	s := b.state
	cb := func(id uuid.UUID) commands.BaseCommand { return commands.BaseCommand{GameID: s.GameID, PlayerID: id} }

	if len(s.Debts) > 0 {
		d := s.Debts[0]
		cands := []commands.Command{&commands.PayDebtCommand{BaseCommand: cb(d.DebtorID)}}
		cands = append(cands, raiseFunds(s, d.DebtorID, cb(d.DebtorID))...)
		cands = append(cands, &commands.DeclareBankruptcyCommand{BaseCommand: cb(d.DebtorID)})
		return b.first(cands...)
	}

	if a := s.Auction; a != nil {
		id := a.CurrentBidderID()
		p := player(s, id)
		bid := a.HighestBid + 10
		if bid <= board.SpaceByIndex[a.SpaceIndex].Price && bid <= p.Money-100 {
			return b.first(&commands.AuctionBidCommand{BaseCommand: cb(id), Amount: bid},
				&commands.AuctionPassCommand{BaseCommand: cb(id)})
		}
		return b.first(&commands.AuctionPassCommand{BaseCommand: cb(id)})
	}

	active := s.Turn.ActivePlayerID
	p := player(s, active)

	if w := s.Turn.ActionWindow; w != nil && w.Type == models.ActionWaitingForPurchaseDecision {
		if p.Money-board.SpaceByIndex[w.SpaceIndex].Price >= 150 {
			return b.first(&commands.BuyPropertyCommand{BaseCommand: cb(active)},
				&commands.DeclinePurchaseCommand{BaseCommand: cb(active)})
		}
		return b.first(&commands.DeclinePurchaseCommand{BaseCommand: cb(active)})
	}

	switch s.Turn.Phase {
	case models.TurnPhasePreRoll:
		if p.Money > 600 {
			for _, idx := range sortedOwned(p) {
				if b.try(&commands.BuildHouseCommand{BaseCommand: cb(active), SpaceIndex: idx}) ||
					b.try(&commands.BuildHotelCommand{BaseCommand: cb(active), SpaceIndex: idx}) {
					return true
				}
			}
		}
		if p.IsInJail {
			if b.first(&commands.UseGetOutOfJailCardCommand{BaseCommand: cb(active)}) {
				return true
			}
		}
		d1, d2 := b.dice.Roll()
		return b.try(&commands.RollDiceCommand{BaseCommand: cb(active), Die1: d1, Die2: d2})
	case models.TurnPhaseEnd:
		return b.try(&commands.EndTurnCommand{BaseCommand: cb(active)})
	}
	return false
}

// raiseFunds lists sell/mortgage commands in a fixed order: sell hotels, then houses, then mortgage.
func raiseFunds(s models.GameState, id uuid.UUID, base commands.BaseCommand) []commands.Command {
	p := player(s, id)
	var cands []commands.Command
	for _, idx := range sortedOwned(p) {
		cands = append(cands, &commands.SellHotelCommand{BaseCommand: base, SpaceIndex: idx})
	}
	for _, idx := range sortedOwned(p) {
		cands = append(cands, &commands.SellHouseCommand{BaseCommand: base, SpaceIndex: idx})
	}
	for _, idx := range sortedOwned(p) {
		cands = append(cands, &commands.MortgagePropertyCommand{BaseCommand: base, SpaceIndex: idx})
	}
	return cands
}

func sortedOwned(p *models.PlayerState) []int {
	idx := append([]int(nil), p.OwnedSpaceIndexes...)
	sort.Ints(idx)
	return idx
}

func player(s models.GameState, id uuid.UUID) *models.PlayerState {
	for i := range s.Players {
		if s.Players[i].PlayerID == id {
			return &s.Players[i]
		}
	}
	return nil
}

// checkInvariants verifies rules that must hold after every command.
func checkInvariants(t *testing.T, s models.GameState) {
	t.Helper()
	houses, hotels := s.Bank.Houses, s.Bank.Hotels
	for idx, o := range s.Ownerships {
		if o.Buildings == models.HotelLevel {
			hotels++
		} else {
			houses += o.Buildings
		}
		owner := player(s, o.OwnerID)
		if owner == nil || owner.IsBankrupt {
			t.Fatalf("space %d owned by missing or bankrupt player", idx)
		}
		found := false
		for _, i := range owner.OwnedSpaceIndexes {
			found = found || i == idx
		}
		if !found {
			t.Fatalf("space %d missing from owner's cache", idx)
		}
	}
	if houses != models.BankStartingHouses || hotels != models.BankStartingHotels {
		t.Fatalf("building supply not conserved: %d houses, %d hotels", houses, hotels)
	}
	for _, p := range s.Players {
		if p.Money < 0 {
			t.Fatalf("player %s has negative balance %d", p.PlayerID, p.Money)
		}
		for _, i := range p.OwnedSpaceIndexes {
			if o, ok := s.Ownerships[i]; !ok || o.OwnerID != p.PlayerID {
				t.Fatalf("player %s cache lists space %d they do not own", p.PlayerID, i)
			}
		}
		if p.IsBankrupt && (p.Money != 0 || len(p.OwnedSpaceIndexes) > 0) {
			t.Fatalf("bankrupt player %s still holds assets", p.PlayerID)
		}
		held := 0
		for _, holder := range s.CardDecks.JailCardHolders {
			if holder == p.PlayerID {
				held++
			}
		}
		if held != p.GetOutOfJailCards {
			t.Fatalf("player %s: %d jail cards counted, %d held", p.PlayerID, p.GetOutOfJailCards, held)
		}
	}
}

// playGame runs the bot until the game finishes or maxCmds commands have been played.
func playGame(t *testing.T, seed int64, setup []commands.Command, maxCmds int) *bot {
	t.Helper()
	b, err := newBot(seed, setup)
	if err != nil {
		t.Fatal(err)
	}
	for len(b.cmds) < maxCmds && b.state.Status == models.GameStatusActive {
		if !b.step() {
			t.Fatalf("seed %d: bot stuck after %d commands (phase %s, window %+v)",
				seed, len(b.cmds), b.state.Turn.Phase, b.state.Turn.ActionWindow)
		}
		checkInvariants(t, b.state)
	}
	return b
}

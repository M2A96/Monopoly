package reducers_test

import (
	"testing"

	"github.com/google/uuid"
	"github/M2A96/Monopoly.git/internal/game/commands"
	"github/M2A96/Monopoly.git/internal/game/events"
	"github/M2A96/Monopoly.git/internal/game/models"
	"github/M2A96/Monopoly.git/internal/game/reducers"
)

// --- helpers ---

type game struct {
	t     *testing.T
	r     *reducers.Registry
	state models.GameState
	ids   []uuid.UUID
}

func newGame(t *testing.T, players int) *game {
	t.Helper()
	g := &game{t: t, r: reducers.NewRegistry(), state: newTestState()}
	g.must(&commands.CreateGameCommand{BaseCommand: commands.BaseCommand{PlayerID: uuid.New()}, Name: "Rules"})
	for i := 0; i < players; i++ {
		id := uuid.New()
		g.ids = append(g.ids, id)
		g.must(&commands.JoinGameCommand{BaseCommand: commands.BaseCommand{PlayerID: id}, Name: "P"})
	}
	g.must(&commands.StartGameCommand{BaseCommand: commands.BaseCommand{PlayerID: g.ids[0]}})
	return g
}

func (g *game) do(cmd commands.Command) ([]events.Event, error) {
	next, evts, err := g.r.Dispatch(g.state, cmd)
	if err == nil {
		g.state = next
	}
	return evts, err
}

func (g *game) must(cmd commands.Command) []events.Event {
	g.t.Helper()
	evts, err := g.do(cmd)
	if err != nil {
		g.t.Fatalf("%s: %v", cmd.GetType(), err)
	}
	return evts
}

func (g *game) reject(cmd commands.Command, why string) {
	g.t.Helper()
	if _, err := g.do(cmd); err == nil {
		g.t.Fatalf("%s should be rejected: %s", cmd.GetType(), why)
	}
}

func (g *game) player(i int) *models.PlayerState { return findPlayerState(&g.state, g.ids[i]) }

func (g *game) own(i, space, buildings int) {
	g.state.Ownerships[space] = models.Ownership{SpaceIndex: space, OwnerID: g.ids[i], Buildings: buildings}
	addSpaceIndex(g.player(i), space)
}

func base(id uuid.UUID) commands.BaseCommand { return commands.BaseCommand{PlayerID: id} }

func roll(id uuid.UUID, d1, d2 int) *commands.RollDiceCommand {
	return &commands.RollDiceCommand{BaseCommand: base(id), Die1: d1, Die2: d2}
}

// --- 3.1a turn enforcement ---

func TestEndTurnBeforeRollRejected(t *testing.T) {
	g := newGame(t, 2)
	g.reject(&commands.EndTurnCommand{BaseCommand: base(g.ids[0])}, "player has not rolled")
}

func TestEndTurnWithPendingPurchaseRejected(t *testing.T) {
	g := newGame(t, 2)
	g.must(roll(g.ids[0], 1, 2)) // Baltic Avenue, unowned
	g.reject(&commands.EndTurnCommand{BaseCommand: base(g.ids[0])}, "purchase decision pending")
}

func TestJailExitOnlyByActivePlayerBeforeRoll(t *testing.T) {
	g := newGame(t, 2)
	g.player(1).IsInJail = true
	g.player(1).GetOutOfJailCards = 1
	g.reject(&commands.PayJailFineCommand{BaseCommand: base(g.ids[1])}, "not p2's turn")
	g.reject(&commands.UseGetOutOfJailCardCommand{BaseCommand: base(g.ids[1])}, "not p2's turn")

	g.player(0).IsInJail = true
	g.must(roll(g.ids[0], 1, 2)) // failed doubles attempt -> END
	g.reject(&commands.PayJailFineCommand{BaseCommand: base(g.ids[0])}, "already rolled")
}

func TestEliminatedPlayerCannotAct(t *testing.T) {
	g := newGame(t, 3)
	g.must(&commands.ResignCommand{BaseCommand: base(g.ids[2])})
	g.reject(&commands.TradeProposeCommand{BaseCommand: base(g.ids[2]), RecipientID: g.ids[0], OfferedMoney: 1}, "resigned")
}

// --- 3.3a property rules ---

func TestUnaffordablePropertyStillOffered(t *testing.T) {
	g := newGame(t, 2)
	g.player(0).Money = 10
	g.must(roll(g.ids[0], 1, 2))
	if w := g.state.Turn.ActionWindow; w == nil || w.Type != models.ActionWaitingForPurchaseDecision {
		t.Fatalf("expected purchase decision, got %+v", w)
	}
	g.reject(&commands.BuyPropertyCommand{BaseCommand: base(g.ids[0])}, "insufficient funds")
	evts := g.must(&commands.DeclinePurchaseCommand{BaseCommand: base(g.ids[0])})
	if !findEvtByType(evts, events.EventAuctionStarted) {
		t.Fatal("expected auction after declining")
	}
}

func TestNearestRailroadCardChargesDoubleRent(t *testing.T) {
	g := newGame(t, 2)
	g.own(1, 15, 0) // Pennsylvania Railroad, one railroad owned -> $25
	g.state.CardDecks.Chance = []models.Card{{ID: "CH04", Effect: models.CardEffectMoveToNearestRailroad, DoubleTax: true}}
	g.state.CardDecks.ChanceIndex = 0

	g.must(roll(g.ids[0], 3, 4)) // Chance at 7 -> nearest railroad 15
	if got := g.player(0).Position; got != 15 {
		t.Fatalf("expected position 15, got %d", got)
	}
	if got := g.player(1).Money; got != 1500+50 {
		t.Errorf("expected owner to receive double rent $50, got %d", got-1500)
	}
}

func TestNearestUtilityCardChargesTenTimesDice(t *testing.T) {
	g := newGame(t, 2)
	g.own(1, 12, 0) // Electric Company
	g.state.CardDecks.Chance = []models.Card{{ID: "CH06", Effect: models.CardEffectMoveToNearestUtility}}

	g.must(roll(g.ids[0], 3, 4)) // dice total 7
	if got := g.player(1).Money - 1500; got != 70 {
		t.Errorf("expected 10x dice = $70, got $%d", got)
	}
}

func TestCardDeckCyclesAndSkipsHeldJailCard(t *testing.T) {
	g := newGame(t, 2)
	g.state.CardDecks.Chance = []models.Card{
		{ID: "CH08", Effect: models.CardEffectGetOutOfJail},
		{ID: "CH07", Effect: models.CardEffectCollect, Amount: 50},
	}
	g.state.CardDecks.ChanceIndex = 0

	drawAgain := func() {
		g.player(0).Position = 0
		g.state.Turn.Phase = models.TurnPhasePreRoll
		g.must(roll(g.ids[0], 3, 4)) // Chance at 7
	}

	g.must(roll(g.ids[0], 3, 4)) // draws the jail card
	if g.player(0).GetOutOfJailCards != 1 {
		t.Fatal("expected p1 to hold a jail card")
	}
	drawAgain() // CH07
	drawAgain() // deck wraps, CH08 is held -> skipped -> CH07
	if got := g.player(0).Money; got != 1500+100 {
		t.Errorf("expected the dividend twice ($1600), got %d", got)
	}
}

// --- 3.4a building rules ---

func TestEvenBuildFromZero(t *testing.T) {
	g := newGame(t, 2)
	g.own(0, 1, 0)
	g.own(0, 3, 0)
	g.must(&commands.BuildHouseCommand{BaseCommand: base(g.ids[0]), SpaceIndex: 1})
	g.reject(&commands.BuildHouseCommand{BaseCommand: base(g.ids[0]), SpaceIndex: 1}, "uneven: 2 vs 0")
}

func TestHotelNeedsFullGroupAtFourHouses(t *testing.T) {
	g := newGame(t, 2)
	g.own(0, 1, 4)
	g.own(0, 3, 3)
	g.reject(&commands.BuildHotelCommand{BaseCommand: base(g.ids[0]), SpaceIndex: 1}, "Baltic has 3 houses")
}

func TestNoBuildingWithMortgageInGroup(t *testing.T) {
	g := newGame(t, 2)
	g.own(0, 1, 0)
	g.own(0, 3, 0)
	o := g.state.Ownerships[3]
	o.Mortgaged = true
	g.state.Ownerships[3] = o
	g.reject(&commands.BuildHouseCommand{BaseCommand: base(g.ids[0]), SpaceIndex: 1}, "Baltic mortgaged")
}

func TestEvenSellRule(t *testing.T) {
	g := newGame(t, 2)
	g.own(0, 1, 2)
	g.own(0, 3, 1)
	g.reject(&commands.SellHouseCommand{BaseCommand: base(g.ids[0]), SpaceIndex: 3}, "sell from most-built first")
	g.must(&commands.SellHouseCommand{BaseCommand: base(g.ids[0]), SpaceIndex: 1})
}

func TestSellHotelNeedsHousesInBank(t *testing.T) {
	g := newGame(t, 2)
	g.own(0, 1, models.HotelLevel)
	g.own(0, 3, models.HotelLevel)
	g.state.Bank.Houses = 3
	g.reject(&commands.SellHotelCommand{BaseCommand: base(g.ids[0]), SpaceIndex: 1}, "bank short of houses")
}

func TestMortgageBlockedByBuildingsInGroup(t *testing.T) {
	g := newGame(t, 2)
	g.own(0, 1, 0)
	g.own(0, 3, 1)
	g.reject(&commands.MortgagePropertyCommand{BaseCommand: base(g.ids[0]), SpaceIndex: 1}, "Baltic has a house")
}

// --- 3.5a jail rules ---

func TestJailCardReturnsToDeck(t *testing.T) {
	g := newGame(t, 2)
	g.state.CardDecks.Chance = []models.Card{{ID: "CH08", Effect: models.CardEffectGetOutOfJail}}
	g.must(roll(g.ids[0], 3, 4))
	if g.state.CardDecks.JailCardHolders["CH08"] != g.ids[0] {
		t.Fatal("expected CH08 held by p1")
	}
	g.must(&commands.EndTurnCommand{BaseCommand: base(g.ids[0])})
	g.must(roll(g.ids[1], 1, 2))
	g.must(&commands.DeclinePurchaseCommand{BaseCommand: base(g.ids[1])})
	g.must(&commands.AuctionPassCommand{BaseCommand: base(g.ids[1])})
	g.must(&commands.AuctionPassCommand{BaseCommand: base(g.ids[0])})
	g.must(&commands.EndTurnCommand{BaseCommand: base(g.ids[1])})

	g.player(0).IsInJail = true
	g.must(&commands.UseGetOutOfJailCardCommand{BaseCommand: base(g.ids[0])})
	if _, held := g.state.CardDecks.JailCardHolders["CH08"]; held {
		t.Error("expected CH08 returned to the deck")
	}
}

func TestForcedJailFineCanCreateDebt(t *testing.T) {
	g := newGame(t, 2)
	p := g.player(0)
	p.IsInJail, p.Position, p.Money = true, 10, 20
	p.JailState.TurnsInJail = 2
	g.must(roll(g.ids[0], 1, 3)) // third failure: forced fine, moves to 14 (Virginia, unowned)
	if len(g.state.Debts) != 1 || g.state.Debts[0].Amount != 50 {
		t.Fatalf("expected $50 jail-fine debt, got %+v", g.state.Debts)
	}
	if g.player(0).IsInJail {
		t.Error("expected player released after third attempt")
	}
}

// --- 3.6a auction rules ---

func TestAuctionHighestBidderWins(t *testing.T) {
	g := newGame(t, 2)
	g.must(roll(g.ids[0], 1, 2)) // Baltic
	g.must(&commands.DeclinePurchaseCommand{BaseCommand: base(g.ids[0])})

	g.reject(&commands.AuctionBidCommand{BaseCommand: base(g.ids[1]), Amount: 10}, "p1 bids first")
	g.must(&commands.AuctionBidCommand{BaseCommand: base(g.ids[0]), Amount: 10})
	g.reject(&commands.AuctionPassCommand{BaseCommand: base(g.ids[0])}, "highest bidder / not their turn")
	g.must(&commands.AuctionBidCommand{BaseCommand: base(g.ids[1]), Amount: 20})
	evts := g.must(&commands.AuctionPassCommand{BaseCommand: base(g.ids[0])})

	ended, ok := findEvt(evts, events.EventAuctionEnded).(*events.AuctionEndedEvent)
	if !ok || ended.WinnerID != g.ids[1] || ended.Amount != 20 {
		t.Fatalf("expected p2 to win for $20, got %+v", ended)
	}
	if g.state.Ownerships[3].OwnerID != g.ids[1] || g.player(1).Money != 1480 {
		t.Errorf("expected p2 to own Baltic and pay $20")
	}
	if g.state.Auction != nil || g.state.Turn.Phase != models.TurnPhaseEnd {
		t.Errorf("expected auction closed and p1's turn at END, got phase %s", g.state.Turn.Phase)
	}
}

func TestAuctionAllPassLeavesPropertyWithBank(t *testing.T) {
	g := newGame(t, 2)
	g.must(roll(g.ids[0], 1, 2))
	g.must(&commands.DeclinePurchaseCommand{BaseCommand: base(g.ids[0])})
	g.must(&commands.AuctionPassCommand{BaseCommand: base(g.ids[0])})
	evts := g.must(&commands.AuctionPassCommand{BaseCommand: base(g.ids[1])})
	if !findEvtByType(evts, events.EventAuctionEnded) {
		t.Fatal("expected AuctionEnded")
	}
	if _, owned := g.state.Ownerships[3]; owned {
		t.Error("property should remain unowned")
	}
}

func TestAuctionAfterDoublesReturnsToRoll(t *testing.T) {
	g := newGame(t, 2)
	g.must(roll(g.ids[0], 3, 3)) // Oriental Avenue, doubles
	g.must(&commands.DeclinePurchaseCommand{BaseCommand: base(g.ids[0])})
	g.must(&commands.AuctionPassCommand{BaseCommand: base(g.ids[0])})
	g.must(&commands.AuctionPassCommand{BaseCommand: base(g.ids[1])})
	if g.state.Turn.Phase != models.TurnPhasePreRoll {
		t.Errorf("expected PRE_ROLL after doubles, got %s", g.state.Turn.Phase)
	}
}

// --- 3.7a bankruptcy rules ---

func TestRaiseFundsThenPayDebt(t *testing.T) {
	g := newGame(t, 2)
	g.own(1, 1, 0)
	g.own(1, 3, models.HotelLevel) // Baltic hotel: $450
	g.own(0, 39, 0)                // Boardwalk, mortgage value $200
	g.player(0).Money = 300

	g.must(roll(g.ids[0], 1, 2))
	if len(g.state.Debts) != 1 || g.state.Debts[0].CreditorID != g.ids[1] {
		t.Fatalf("expected debt to p2, got %+v", g.state.Debts)
	}
	g.reject(&commands.PayDebtCommand{BaseCommand: base(g.ids[0])}, "only $300 of $450")
	g.must(&commands.MortgagePropertyCommand{BaseCommand: base(g.ids[0]), SpaceIndex: 39})
	evts := g.must(&commands.PayDebtCommand{BaseCommand: base(g.ids[0])})

	if !findEvtByType(evts, events.EventDebtPaid) {
		t.Error("expected DebtPaid event")
	}
	if g.player(0).Money != 50 || g.player(1).Money != 1950 {
		t.Errorf("expected p1 $50 / p2 $1950, got %d / %d", g.player(0).Money, g.player(1).Money)
	}
	if g.state.Turn.Phase != models.TurnPhaseEnd {
		t.Errorf("expected END after paying, got %s", g.state.Turn.Phase)
	}
}

func TestBankruptcyToPlayerTransfersAssetsAndPassesTurn(t *testing.T) {
	g := newGame(t, 3)
	g.own(1, 1, 0)
	g.own(1, 3, models.HotelLevel)
	g.own(0, 1+5, 0) // Oriental, will transfer
	g.own(0, 8, 0)   // Vermont
	o := g.state.Ownerships[8]
	o.Mortgaged = true
	g.state.Ownerships[8] = o
	g.own(0, 9, 0) // Connecticut: light-blue set complete with 6 and 8
	g.state.Ownerships[6] = models.Ownership{SpaceIndex: 6, OwnerID: g.ids[0]}
	g.player(0).Money = 100

	g.must(roll(g.ids[0], 1, 2))
	evts := g.must(&commands.DeclareBankruptcyCommand{BaseCommand: base(g.ids[0])})
	if !findEvtByType(evts, events.EventBankruptcyDeclared) {
		t.Fatal("expected BankruptcyDeclared")
	}

	if !g.player(0).IsBankrupt || len(g.player(0).OwnedSpaceIndexes) != 0 {
		t.Error("expected p1 eliminated with no properties")
	}
	if g.state.Ownerships[6].OwnerID != g.ids[1] || !g.state.Ownerships[8].Mortgaged {
		t.Error("expected p2 to receive properties, mortgages kept")
	}
	if g.player(1).Money != 1500+100 {
		t.Errorf("expected p2 to receive p1's $100, got %d", g.player(1).Money-1500)
	}
	if g.state.Turn.ActivePlayerID != g.ids[1] || g.state.Turn.Phase != models.TurnPhasePreRoll {
		t.Errorf("expected turn to pass to p2 in PRE_ROLL, got %s/%s", g.state.Turn.ActivePlayerID, g.state.Turn.Phase)
	}
	if g.state.Status != models.GameStatusActive {
		t.Error("game should continue with 2 players")
	}
	if len(g.state.Debts) != 0 {
		t.Error("debt should be settled by bankruptcy")
	}
}

func TestBankruptcySellsBuildingsBackToBank(t *testing.T) {
	g := newGame(t, 2)
	g.own(0, 1, 2)
	g.own(0, 3, 2)
	g.state.Bank.Houses -= 4
	g.player(0).Money = 10
	g.must(roll(g.ids[0], 1, 3)) // Income Tax $200 -> debt to bank
	g.must(&commands.DeclareBankruptcyCommand{BaseCommand: base(g.ids[0])})
	if g.state.Bank.Houses != models.BankStartingHouses {
		t.Errorf("expected all houses back in the bank, got %d", g.state.Bank.Houses)
	}
}

func TestCollectFromPlayersCreatesDebtForPoorPlayer(t *testing.T) {
	g := newGame(t, 3)
	g.state.CardDecks.CommunityChest = []models.Card{{ID: "CC10", Effect: models.CardEffectCollectFromPlayers, Amount: 10}}
	g.player(2).Money = 5
	g.must(roll(g.ids[0], 1, 1)) // Community Chest at 2 (doubles)
	if len(g.state.Debts) != 1 || g.state.Debts[0].DebtorID != g.ids[2] {
		t.Fatalf("expected one debt from p3, got %+v", g.state.Debts)
	}
	if w := g.state.Turn.ActionWindow; w == nil || w.PlayerID != g.ids[2] {
		t.Fatalf("expected resolution window for p3, got %+v", w)
	}
	g.reject(roll(g.ids[0], 2, 3), "debt pending")
	g.must(&commands.DeclareBankruptcyCommand{BaseCommand: base(g.ids[2])})
	if g.state.Turn.ActivePlayerID != g.ids[0] || g.state.Turn.Phase != models.TurnPhasePreRoll {
		t.Errorf("expected p1 to continue (doubles), got %s/%s", g.state.Turn.ActivePlayerID, g.state.Turn.Phase)
	}
}

// --- 3.8a win conditions ---

func TestResignationTimeoutEndsGame(t *testing.T) {
	g := newGame(t, 3)
	g.must(&commands.ResignCommand{BaseCommand: base(g.ids[1]), Reason: commands.ResignVoluntary})
	if g.state.Status != models.GameStatusActive {
		t.Fatal("game should continue with 2 players")
	}
	evts := g.must(&commands.ResignCommand{BaseCommand: base(g.ids[0]), Reason: commands.ResignTimeout})
	ended, ok := findEvt(evts, events.EventGameEnded).(*events.GameEndedEvent)
	if !ok || ended.WinnerID != g.ids[2] {
		t.Fatalf("expected p3 to win, got %+v", ended)
	}
	if g.state.Status != models.GameStatusFinished {
		t.Error("expected FINISHED")
	}
	g.reject(roll(g.ids[2], 1, 2), "game finished")
}

func TestResignUnknownReasonRejected(t *testing.T) {
	g := newGame(t, 2)
	g.reject(&commands.ResignCommand{BaseCommand: base(g.ids[0]), Reason: "BORED"}, "unknown reason")
}

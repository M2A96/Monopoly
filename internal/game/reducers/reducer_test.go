package reducers_test

import (
	"strings"
	"testing"

	"github.com/google/uuid"
	"github/M2A96/Monopoly.git/internal/game/commands"
	"github/M2A96/Monopoly.git/internal/game/events"
	"github/M2A96/Monopoly.git/internal/game/models"
	"github/M2A96/Monopoly.git/internal/game/reducers"
	"github/M2A96/Monopoly.git/internal/game/replay"
)

func newTestState() models.GameState {
	return models.GameState{
		GameID:     uuid.New(),
		Status:     models.GameStatusWaiting,
		Players:    nil,
		Bank:       models.NewBank(),
		Ownerships: make(map[int]models.Ownership),
		CardDecks: models.CardDecks{
			Chance:         models.DefaultChanceDeck(),
			CommunityChest: models.DefaultCommunityChestDeck(),
		},
		Auction: nil,
		Trades:  nil,
		Turn:    models.Turn{},
	}
}

func newActiveGame(t *testing.T) (models.GameState, uuid.UUID, uuid.UUID, uuid.UUID) {
	t.Helper()
	registry := reducers.NewRegistry()
	state := newTestState()

	p1 := uuid.New()
	p2 := uuid.New()
	gameName := "Test Game"

	var err error
	state, _, err = registry.Dispatch(state, &commands.CreateGameCommand{
		BaseCommand: commands.BaseCommand{PlayerID: p1},
		Name:        gameName,
	})
	if err != nil {
		t.Fatalf("create game: %v", err)
	}

	state, _, err = registry.Dispatch(state, &commands.JoinGameCommand{
		BaseCommand: commands.BaseCommand{PlayerID: p1},
		Name:        "Alice",
	})
	if err != nil {
		t.Fatalf("join game p1: %v", err)
	}

	state, _, err = registry.Dispatch(state, &commands.JoinGameCommand{
		BaseCommand: commands.BaseCommand{PlayerID: p2},
		Name:        "Bob",
	})
	if err != nil {
		t.Fatalf("join game p2: %v", err)
	}

	state, _, err = registry.Dispatch(state, &commands.StartGameCommand{
		BaseCommand: commands.BaseCommand{PlayerID: p1},
	})
	if err != nil {
		t.Fatalf("start game: %v", err)
	}

	return state, state.GameID, p1, p2
}

func findEvt(evts []events.Event, eventType events.EventType) events.Event {
	for _, e := range evts {
		if e.GetType() == eventType {
			return e
		}
	}
	return nil
}

func findEvtByType(evts []events.Event, eventType events.EventType) bool {
	return findEvt(evts, eventType) != nil
}

func TestCreateGame(t *testing.T) {
	registry := reducers.NewRegistry()
	state := newTestState()
	playerID := uuid.New()

	newState, evts, err := registry.Dispatch(state, &commands.CreateGameCommand{
		BaseCommand: commands.BaseCommand{PlayerID: playerID},
		Name:        "My Game",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if newState.Status != models.GameStatusWaiting {
		t.Errorf("expected WAITING status, got %s", newState.Status)
	}
	if len(evts) != 1 {
		t.Fatalf("expected 1 event, got %d", len(evts))
	}
	if evts[0].GetType() != events.EventGameCreated {
		t.Errorf("expected GameCreated event, got %s", evts[0].GetType())
	}
	if newState.Ownerships == nil {
		t.Error("expected Ownerships map to be initialized")
	}
	if newState.Bank.Houses != 32 || newState.Bank.Hotels != 12 {
		t.Errorf("expected 32 houses and 12 hotels, got %d and %d", newState.Bank.Houses, newState.Bank.Hotels)
	}
}

func TestJoinGame(t *testing.T) {
	registry := reducers.NewRegistry()
	state := newTestState()
	p1 := uuid.New()
	p2 := uuid.New()

	state, _, _ = registry.Dispatch(state, &commands.CreateGameCommand{
		BaseCommand: commands.BaseCommand{PlayerID: p1},
		Name:        "Game",
	})

	newState, evts, err := registry.Dispatch(state, &commands.JoinGameCommand{
		BaseCommand: commands.BaseCommand{PlayerID: p1},
		Name:        "Alice",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(newState.Players) != 1 {
		t.Fatalf("expected 1 player, got %d", len(newState.Players))
	}
	if newState.Players[0].Money != 1500 {
		t.Errorf("expected $1500 starting money, got %d", newState.Players[0].Money)
	}
	if !findEvtByType(evts, events.EventPlayerJoined) {
		t.Error("expected PlayerJoined event")
	}

	newState, _, err = registry.Dispatch(newState, &commands.JoinGameCommand{
		BaseCommand: commands.BaseCommand{PlayerID: p2},
		Name:        "Bob",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(newState.Players) != 2 {
		t.Fatalf("expected 2 players, got %d", len(newState.Players))
	}

	_, _, err = registry.Dispatch(newState, &commands.JoinGameCommand{
		BaseCommand: commands.BaseCommand{PlayerID: p1},
		Name:        "Alice",
	})
	if err == nil {
		t.Error("expected error for duplicate player")
	}
}

func TestJoinGameFull(t *testing.T) {
	registry := reducers.NewRegistry()
	state := newTestState()
	state, _, _ = registry.Dispatch(state, &commands.CreateGameCommand{
		BaseCommand: commands.BaseCommand{PlayerID: uuid.New()},
		Name:        "Game",
	})

	for i := 0; i < 8; i++ {
		var err error
		state, _, err = registry.Dispatch(state, &commands.JoinGameCommand{
			BaseCommand: commands.BaseCommand{PlayerID: uuid.New()},
			Name:        "Player",
		})
		if err != nil {
			t.Fatalf("join %d: %v", i, err)
		}
	}

	_, _, err := registry.Dispatch(state, &commands.JoinGameCommand{
		BaseCommand: commands.BaseCommand{PlayerID: uuid.New()},
		Name:        "Extra",
	})
	if err == nil {
		t.Error("expected error for full game")
	}
}

func TestStartGame(t *testing.T) {
	registry := reducers.NewRegistry()
	state := newTestState()
	p1 := uuid.New()
	p2 := uuid.New()

	state, _, _ = registry.Dispatch(state, &commands.CreateGameCommand{
		BaseCommand: commands.BaseCommand{PlayerID: p1},
		Name:        "Game",
	})
	state, _, _ = registry.Dispatch(state, &commands.JoinGameCommand{
		BaseCommand: commands.BaseCommand{PlayerID: p1},
		Name:        "Alice",
	})
	state, _, _ = registry.Dispatch(state, &commands.JoinGameCommand{
		BaseCommand: commands.BaseCommand{PlayerID: p2},
		Name:        "Bob",
	})

	newState, evts, err := registry.Dispatch(state, &commands.StartGameCommand{
		BaseCommand: commands.BaseCommand{PlayerID: p1},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if newState.Status != models.GameStatusActive {
		t.Errorf("expected ACTIVE status, got %s", newState.Status)
	}
	if newState.Turn.ActivePlayerID != p1 {
		t.Errorf("expected first player to be active")
	}
	if newState.Turn.Phase != models.TurnPhasePreRoll {
		t.Errorf("expected PRE_ROLL phase, got %s", newState.Turn.Phase)
	}
	if !findEvtByType(evts, events.EventGameStarted) {
		t.Error("expected GameStarted event")
	}
	if !findEvtByType(evts, events.EventTurnStarted) {
		t.Error("expected TurnStarted event")
	}
}

func TestStartGameTooFewPlayers(t *testing.T) {
	registry := reducers.NewRegistry()
	state := newTestState()
	p1 := uuid.New()

	state, _, _ = registry.Dispatch(state, &commands.CreateGameCommand{
		BaseCommand: commands.BaseCommand{PlayerID: p1},
		Name:        "Game",
	})
	state, _, _ = registry.Dispatch(state, &commands.JoinGameCommand{
		BaseCommand: commands.BaseCommand{PlayerID: p1},
		Name:        "Alice",
	})

	_, _, err := registry.Dispatch(state, &commands.StartGameCommand{
		BaseCommand: commands.BaseCommand{PlayerID: p1},
	})
	if err == nil {
		t.Error("expected error for too few players")
	}
}

func TestRollDice(t *testing.T) {
	state, _, p1, _ := newActiveGame(t)
	registry := reducers.NewRegistry()

	newState, evts, err := registry.Dispatch(state, &commands.RollDiceCommand{
		BaseCommand: commands.BaseCommand{PlayerID: p1},
		Die1:        2,
		Die2:        3,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !findEvtByType(evts, events.EventDiceRolled) {
		t.Error("expected DiceRolled event")
	}
	if !findEvtByType(evts, events.EventPlayerMoved) {
		t.Error("expected PlayerMoved event")
	}

	player := findPlayerState(&newState, p1)
	if player.Position != 5 {
		t.Errorf("expected position 5, got %d", player.Position)
	}
	if newState.Turn.ActionWindow == nil || newState.Turn.ActionWindow.Type != models.ActionWaitingForPurchaseDecision {
		t.Errorf("expected purchase decision for unowned railroad")
	}
}

func TestRollDicePassedGo(t *testing.T) {
	state, _, p1, _ := newActiveGame(t)
	registry := reducers.NewRegistry()

	player := findPlayerState(&state, p1)
	player.Position = 38

	state.Players[0] = *player

	newState, evts, err := registry.Dispatch(state, &commands.RollDiceCommand{
		BaseCommand: commands.BaseCommand{PlayerID: p1},
		Die1:        5,
		Die2:        6,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !findEvtByType(evts, events.EventPassedGo) {
		t.Error("expected PassedGo event")
	}

	player = findPlayerState(&newState, p1)
	if player.Money != 1500+200 {
		t.Errorf("expected $1700 after passing GO, got %d", player.Money)
	}
}

func TestRollDiceDoubles(t *testing.T) {
	state, _, p1, _ := newActiveGame(t)
	registry := reducers.NewRegistry()

	newState, evts, err := registry.Dispatch(state, &commands.RollDiceCommand{
		BaseCommand: commands.BaseCommand{PlayerID: p1},
		Die1:        3,
		Die2:        3,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_ = evts

	diceEvt := findEvt(evts, events.EventDiceRolled)
	if diceEvt == nil {
		t.Fatal("expected DiceRolled event")
	}

	// Doubles land on Oriental Avenue (6): resolve the purchase, then the player rolls again.
	newState, _, err = registry.Dispatch(newState, &commands.BuyPropertyCommand{
		BaseCommand: commands.BaseCommand{PlayerID: p1},
	})
	if err != nil {
		t.Fatalf("buy after doubles: %v", err)
	}
	if newState.Turn.Phase != models.TurnPhasePreRoll {
		t.Fatalf("expected PRE_ROLL after doubles, got %s", newState.Turn.Phase)
	}

	nextState, _, err := registry.Dispatch(newState, &commands.RollDiceCommand{
		BaseCommand: commands.BaseCommand{PlayerID: p1},
		Die1:        2,
		Die2:        1,
	})
	if err != nil {
		t.Fatalf("second roll after doubles: %v", err)
	}

	player := findPlayerState(&nextState, p1)
	expectedPos := (6 + 3) % 40
	if player.Position != expectedPos {
		t.Errorf("expected position %d after doubles roll, got %d", expectedPos, player.Position)
	}
}

func TestRollDiceNotYourTurn(t *testing.T) {
	state, _, _, p2 := newActiveGame(t)
	registry := reducers.NewRegistry()

	_, _, err := registry.Dispatch(state, &commands.RollDiceCommand{
		BaseCommand: commands.BaseCommand{PlayerID: p2},
		Die1:        1,
		Die2:        2,
	})
	if err == nil {
		t.Error("expected error for wrong player")
	}
}

func TestBuyProperty(t *testing.T) {
	state, _, p1, _ := newActiveGame(t)
	registry := reducers.NewRegistry()

	state, _, _ = registry.Dispatch(state, &commands.RollDiceCommand{
		BaseCommand: commands.BaseCommand{PlayerID: p1},
		Die1:        1,
		Die2:        2,
	})

	if state.Turn.ActionWindow == nil || state.Turn.ActionWindow.Type != models.ActionWaitingForPurchaseDecision {
		t.Fatalf("expected purchase decision, got phase=%s actionWindow=%v", state.Turn.Phase, state.Turn.ActionWindow)
	}

	newState, evts, err := registry.Dispatch(state, &commands.BuyPropertyCommand{
		BaseCommand: commands.BaseCommand{PlayerID: p1},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !findEvtByType(evts, events.EventPropertyPurchased) {
		t.Error("expected PropertyPurchased event")
	}

	player := findPlayerState(&newState, p1)
	if player.Money != 1500-60 {
		t.Errorf("expected $1440, got %d", player.Money)
	}

	if _, ok := newState.Ownerships[3]; !ok {
		t.Error("expected property 3 to be owned")
	}
	if newState.Turn.Phase != models.TurnPhaseEnd {
		t.Errorf("expected END phase after buying on a non-doubles roll, got %s", newState.Turn.Phase)
	}
}

func TestDeclinePurchaseStartsAuction(t *testing.T) {
	state, _, p1, _ := newActiveGame(t)
	registry := reducers.NewRegistry()

	state, _, _ = registry.Dispatch(state, &commands.RollDiceCommand{
		BaseCommand: commands.BaseCommand{PlayerID: p1},
		Die1:        1,
		Die2:        2,
	})

	newState, evts, err := registry.Dispatch(state, &commands.DeclinePurchaseCommand{
		BaseCommand: commands.BaseCommand{PlayerID: p1},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !findEvtByType(evts, events.EventAuctionStarted) {
		t.Error("expected AuctionStarted event")
	}
	if newState.Auction == nil {
		t.Error("expected auction to be created")
	}
}

func TestEndTurn(t *testing.T) {
	state, _, p1, p2 := newActiveGame(t)
	registry := reducers.NewRegistry()

	// 1+3 lands on Income Tax: nothing to decide, so the turn can end.
	state, _, err := registry.Dispatch(state, &commands.RollDiceCommand{
		BaseCommand: commands.BaseCommand{PlayerID: p1}, Die1: 1, Die2: 3,
	})
	if err != nil {
		t.Fatalf("roll: %v", err)
	}

	newState, evts, err := registry.Dispatch(state, &commands.EndTurnCommand{
		BaseCommand: commands.BaseCommand{PlayerID: p1},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !findEvtByType(evts, events.EventTurnEnded) {
		t.Error("expected TurnEnded event")
	}
	if !findEvtByType(evts, events.EventTurnStarted) {
		t.Error("expected TurnStarted event")
	}
	if newState.Turn.ActivePlayerID != p2 {
		t.Errorf("expected player 2 to be active, got %s", newState.Turn.ActivePlayerID)
	}
	if newState.Turn.Phase != models.TurnPhasePreRoll {
		t.Errorf("expected PRE_ROLL phase, got %s", newState.Turn.Phase)
	}
}

func TestBuildHouse(t *testing.T) {
	state, _, p1, _ := newActiveGame(t)
	registry := reducers.NewRegistry()

	state.Ownerships[1] = models.Ownership{SpaceIndex: 1, OwnerID: p1}
	state.Ownerships[3] = models.Ownership{SpaceIndex: 3, OwnerID: p1}

	player := findPlayerState(&state, p1)
	addSpaceIndex(player, 1)
	addSpaceIndex(player, 3)
	state.Players[0] = *player

	newState, evts, err := registry.Dispatch(state, &commands.BuildHouseCommand{
		BaseCommand: commands.BaseCommand{PlayerID: p1},
		SpaceIndex:  1,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !findEvtByType(evts, events.EventHouseBuilt) {
		t.Error("expected HouseBuilt event")
	}

	o := newState.Ownerships[1]
	if o.Buildings != 1 {
		t.Errorf("expected 1 building, got %d", o.Buildings)
	}

	player = findPlayerState(&newState, p1)
	if player.Money != 1500-50 {
		t.Errorf("expected $1450, got %d", player.Money)
	}
	if newState.Bank.Houses != 31 {
		t.Errorf("expected 31 houses in bank, got %d", newState.Bank.Houses)
	}
}

func TestBuildHouseNoMonopoly(t *testing.T) {
	state, _, p1, _ := newActiveGame(t)
	registry := reducers.NewRegistry()

	state.Ownerships[1] = models.Ownership{SpaceIndex: 1, OwnerID: p1}

	player := findPlayerState(&state, p1)
	addSpaceIndex(player, 1)
	state.Players[0] = *player

	_, _, err := registry.Dispatch(state, &commands.BuildHouseCommand{
		BaseCommand: commands.BaseCommand{PlayerID: p1},
		SpaceIndex:  1,
	})
	if err == nil {
		t.Error("expected error without monopoly")
	}
}

func TestMortgageAndUnmortgage(t *testing.T) {
	state, _, p1, _ := newActiveGame(t)
	registry := reducers.NewRegistry()

	state.Ownerships[1] = models.Ownership{SpaceIndex: 1, OwnerID: p1}
	player := findPlayerState(&state, p1)
	addSpaceIndex(player, 1)
	state.Players[0] = *player

	newState, evts, err := registry.Dispatch(state, &commands.MortgagePropertyCommand{
		BaseCommand: commands.BaseCommand{PlayerID: p1},
		SpaceIndex:  1,
	})
	if err != nil {
		t.Fatalf("mortgage: %v", err)
	}
	if !findEvtByType(evts, events.EventPropertyMortgaged) {
		t.Error("expected PropertyMortgaged event")
	}
	o := newState.Ownerships[1]
	if !o.Mortgaged {
		t.Error("expected property to be mortgaged")
	}
	player = findPlayerState(&newState, p1)
	if player.Money != 1500+30 {
		t.Errorf("expected $1530, got %d", player.Money)
	}

	newState, evts, err = registry.Dispatch(newState, &commands.UnmortgagePropertyCommand{
		BaseCommand: commands.BaseCommand{PlayerID: p1},
		SpaceIndex:  1,
	})
	if err != nil {
		t.Fatalf("unmortgage: %v", err)
	}
	if !findEvtByType(evts, events.EventPropertyUnmortgaged) {
		t.Error("expected PropertyUnmortgaged event")
	}
	o = newState.Ownerships[1]
	if o.Mortgaged {
		t.Error("expected property to be unmortgaged")
	}
}

func TestTradeProposeAndAccept(t *testing.T) {
	state, _, p1, p2 := newActiveGame(t)
	registry := reducers.NewRegistry()

	state.Ownerships[1] = models.Ownership{SpaceIndex: 1, OwnerID: p1}
	state.Ownerships[6] = models.Ownership{SpaceIndex: 6, OwnerID: p2}
	player1 := findPlayerState(&state, p1)
	addSpaceIndex(player1, 1)
	player2 := findPlayerState(&state, p2)
	addSpaceIndex(player2, 6)
	state.Players[0] = *player1
	state.Players[1] = *player2

	newState, evts, err := registry.Dispatch(state, &commands.TradeProposeCommand{
		BaseCommand:     commands.BaseCommand{PlayerID: p1},
		RecipientID:     p2,
		OfferedMoney:    100,
		OfferedSpaces:   []int{1},
		RequestedMoney:  0,
		RequestedSpaces: []int{6},
	})
	if err != nil {
		t.Fatalf("trade propose: %v", err)
	}
	if !findEvtByType(evts, events.EventTradeProposed) {
		t.Error("expected TradeProposed event")
	}
	if len(newState.Trades) != 1 {
		t.Fatalf("expected 1 trade, got %d", len(newState.Trades))
	}

	newState, evts, err = registry.Dispatch(newState, &commands.TradeAcceptCommand{
		BaseCommand: commands.BaseCommand{PlayerID: p2},
		TradeID:     newState.Trades[0].TradeID,
	})
	if err != nil {
		t.Fatalf("trade accept: %v", err)
	}
	if !findEvtByType(evts, events.EventTradeAccepted) {
		t.Error("expected TradeAccepted event")
	}

	o1 := newState.Ownerships[1]
	if o1.OwnerID != p2 {
		t.Error("expected property 1 to belong to p2 after trade")
	}
	o6 := newState.Ownerships[6]
	if o6.OwnerID != p1 {
		t.Error("expected property 6 to belong to p1 after trade")
	}
	player1 = findPlayerState(&newState, p1)
	if player1.Money != 1500-100 {
		t.Errorf("expected p1 to have $1400, got %d", player1.Money)
	}
}

func TestPayJailFine(t *testing.T) {
	state, _, p1, _ := newActiveGame(t)
	registry := reducers.NewRegistry()

	player := findPlayerState(&state, p1)
	player.IsInJail = true
	player.JailState = models.JailState{TurnsInJail: 1}
	state.Players[0] = *player

	newState, evts, err := registry.Dispatch(state, &commands.PayJailFineCommand{
		BaseCommand: commands.BaseCommand{PlayerID: p1},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !findEvtByType(evts, events.EventReleasedFromJail) {
		t.Error("expected ReleasedFromJail event")
	}
	player = findPlayerState(&newState, p1)
	if player.IsInJail {
		t.Error("expected player to be out of jail")
	}
	if player.Money != 1500-50 {
		t.Errorf("expected $1450, got %d", player.Money)
	}
}

func TestDeclareBankruptcy(t *testing.T) {
	state, _, p1, p2 := newActiveGame(t)
	registry := reducers.NewRegistry()

	// p1 cannot cover Income Tax ($200): a debt to the bank is opened.
	state.Players[0].Money = 100
	state.Ownerships[39] = models.Ownership{SpaceIndex: 39, OwnerID: p1}
	addSpaceIndex(&state.Players[0], 39)

	state, evts, err := registry.Dispatch(state, &commands.RollDiceCommand{
		BaseCommand: commands.BaseCommand{PlayerID: p1}, Die1: 1, Die2: 3,
	})
	if err != nil {
		t.Fatalf("roll: %v", err)
	}
	if !findEvtByType(evts, events.EventDebtIncurred) {
		t.Fatal("expected DebtIncurred event")
	}
	if w := state.Turn.ActionWindow; w == nil || w.Type != models.ActionWaitingForBankruptcyResolution || w.PlayerID != p1 {
		t.Fatalf("expected bankruptcy resolution window for p1, got %+v", w)
	}

	newState, evts, err := registry.Dispatch(state, &commands.DeclareBankruptcyCommand{
		BaseCommand: commands.BaseCommand{PlayerID: p1},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !findEvtByType(evts, events.EventBankruptcyDeclared) {
		t.Error("expected BankruptcyDeclared event")
	}
	ended, ok := findEvt(evts, events.EventGameEnded).(*events.GameEndedEvent)
	if !ok || ended.WinnerID != p2 {
		t.Fatalf("expected GameEnded with p2 as winner, got %+v", ended)
	}
	if newState.Status != models.GameStatusFinished {
		t.Errorf("expected FINISHED status, got %s", newState.Status)
	}
	if _, owned := newState.Ownerships[39]; owned {
		t.Error("bank creditor: property should return to the bank")
	}
}

func TestDeclareBankruptcyWithoutDebtFails(t *testing.T) {
	state, _, p1, _ := newActiveGame(t)
	registry := reducers.NewRegistry()

	if _, _, err := registry.Dispatch(state, &commands.DeclareBankruptcyCommand{
		BaseCommand: commands.BaseCommand{PlayerID: p1},
	}); err == nil {
		t.Error("expected error declaring bankruptcy without a debt")
	}
}

func TestUnsupportedCommand(t *testing.T) {
	registry := reducers.NewRegistry()
	state := newTestState()

	_, _, err := registry.Dispatch(state, &commands.StartGameCommand{
		BaseCommand: commands.BaseCommand{PlayerID: uuid.New()},
	})
	if err == nil {
		t.Error("expected error for unsupported command path")
	}
}

func TestReplayDeterminism(t *testing.T) {
	state := newTestState()
	p1 := uuid.New()
	p2 := uuid.New()

	cmds := []commands.Command{
		&commands.CreateGameCommand{BaseCommand: commands.BaseCommand{PlayerID: p1}, Name: "Test"},
		&commands.JoinGameCommand{BaseCommand: commands.BaseCommand{PlayerID: p1}, Name: "Alice"},
		&commands.JoinGameCommand{BaseCommand: commands.BaseCommand{PlayerID: p2}, Name: "Bob"},
		&commands.StartGameCommand{BaseCommand: commands.BaseCommand{PlayerID: p1}},
	}

	r1, err := replay.Replay(state, cmds)
	if err != nil {
		t.Fatalf("first replay: %v", err)
	}

	r2, err := replay.Replay(state, cmds)
	if err != nil {
		t.Fatalf("second replay: %v", err)
	}

	if !replay.CompareStates(r1.FinalState, r2.FinalState) {
		t.Error("replays produced different states for same commands")
	}
}

func TestReplayWithDice(t *testing.T) {
	state := newTestState()
	p1 := uuid.New()
	p2 := uuid.New()

	cmds := []commands.Command{
		&commands.CreateGameCommand{BaseCommand: commands.BaseCommand{PlayerID: p1}, Name: "Test"},
		&commands.JoinGameCommand{BaseCommand: commands.BaseCommand{PlayerID: p1}, Name: "Alice"},
		&commands.JoinGameCommand{BaseCommand: commands.BaseCommand{PlayerID: p2}, Name: "Bob"},
		&commands.StartGameCommand{BaseCommand: commands.BaseCommand{PlayerID: p1}},
		&commands.RollDiceCommand{BaseCommand: commands.BaseCommand{PlayerID: p1}, Die1: 3, Die2: 4},
	}

	r1, err := replay.Replay(state, cmds)
	if err != nil {
		t.Fatalf("first replay: %v", err)
	}
	r2, err := replay.Replay(state, cmds)
	if err != nil {
		t.Fatalf("second replay: %v", err)
	}

	if !replay.CompareStates(r1.FinalState, r2.FinalState) {
		t.Error("non-deterministic replay with dice")
	}
}

func TestStateHash(t *testing.T) {
	state := newTestState()
	p1 := uuid.New()
	p2 := uuid.New()

	cmds1 := []commands.Command{
		&commands.CreateGameCommand{BaseCommand: commands.BaseCommand{PlayerID: p1}, Name: "Test"},
		&commands.JoinGameCommand{BaseCommand: commands.BaseCommand{PlayerID: p1}, Name: "Alice"},
		&commands.JoinGameCommand{BaseCommand: commands.BaseCommand{PlayerID: p2}, Name: "Bob"},
		&commands.StartGameCommand{BaseCommand: commands.BaseCommand{PlayerID: p1}},
	}
	cmds2 := make([]commands.Command, len(cmds1))
	copy(cmds2, cmds1)

	r1, _ := replay.Replay(state, cmds1)
	r2, _ := replay.Replay(state, cmds2)

	h1 := replay.StateHash(r1.FinalState)
	h2 := replay.StateHash(r2.FinalState)
	if h1 != h2 {
		t.Errorf("same state should produce same hash: %s vs %s", h1, h2)
	}
}

func TestTripleDoublesSendsToJail(t *testing.T) {
	state, _, p1, _ := newActiveGame(t)
	registry := reducers.NewRegistry()

	state, _, _ = registry.Dispatch(state, &commands.RollDiceCommand{
		BaseCommand: commands.BaseCommand{PlayerID: p1},
		Die1:        3, Die2: 3,
	})

	state.Turn.Phase = models.TurnPhasePreRoll
	state, _, _ = registry.Dispatch(state, &commands.RollDiceCommand{
		BaseCommand: commands.BaseCommand{PlayerID: p1},
		Die1:        4, Die2: 4,
	})

	state.Turn.Phase = models.TurnPhasePreRoll
	newState, evts, err := registry.Dispatch(state, &commands.RollDiceCommand{
		BaseCommand: commands.BaseCommand{PlayerID: p1},
		Die1:        2, Die2: 2,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !findEvtByType(evts, events.EventSentToJail) {
		t.Error("expected SentToJail event after triple doubles")
	}

	player := findPlayerState(&newState, p1)
	if !player.IsInJail {
		t.Error("expected player to be in jail")
	}
	if player.Position != 10 {
		t.Errorf("expected player at jail position 10, got %d", player.Position)
	}
}

func TestSellHouseAndHotel(t *testing.T) {
	state, _, p1, _ := newActiveGame(t)
	registry := reducers.NewRegistry()

	state.Ownerships[1] = models.Ownership{SpaceIndex: 1, OwnerID: p1, Buildings: 2}
	player := findPlayerState(&state, p1)
	addSpaceIndex(player, 1)
	state.Players[0] = *player

	newState, evts, err := registry.Dispatch(state, &commands.SellHouseCommand{
		BaseCommand: commands.BaseCommand{PlayerID: p1},
		SpaceIndex:  1,
	})
	if err != nil {
		t.Fatalf("sell house: %v", err)
	}
	if !findEvtByType(evts, events.EventHouseSold) {
		t.Error("expected HouseSold event")
	}
	if newState.Ownerships[1].Buildings != 1 {
		t.Errorf("expected 1 building, got %d", newState.Ownerships[1].Buildings)
	}
	player = findPlayerState(&newState, p1)
	if player.Money != 1500+25 {
		t.Errorf("expected $1525, got %d", player.Money)
	}

	state2 := newState
	state2.Ownerships[1] = models.Ownership{SpaceIndex: 1, OwnerID: p1, Buildings: models.HotelLevel}
	player2 := findPlayerState(&state2, p1)
	state2.Players[0] = *player2

	newState2, evts2, err := registry.Dispatch(state2, &commands.SellHotelCommand{
		BaseCommand: commands.BaseCommand{PlayerID: p1},
		SpaceIndex:  1,
	})
	if err != nil {
		t.Fatalf("sell hotel: %v", err)
	}
	if !findEvtByType(evts2, events.EventHotelSold) {
		t.Error("expected HotelSold event")
	}
	if newState2.Ownerships[1].Buildings != 4 {
		t.Errorf("expected 4 buildings after hotel sell, got %d", newState2.Ownerships[1].Buildings)
	}
}

func TestBuildHotel(t *testing.T) {
	state, _, p1, _ := newActiveGame(t)
	registry := reducers.NewRegistry()

	state.Ownerships[1] = models.Ownership{SpaceIndex: 1, OwnerID: p1, Buildings: 4}
	state.Ownerships[3] = models.Ownership{SpaceIndex: 3, OwnerID: p1, Buildings: 4}
	player := findPlayerState(&state, p1)
	addSpaceIndex(player, 1)
	addSpaceIndex(player, 3)
	state.Players[0] = *player

	newState, evts, err := registry.Dispatch(state, &commands.BuildHotelCommand{
		BaseCommand: commands.BaseCommand{PlayerID: p1},
		SpaceIndex:  1,
	})
	if err != nil {
		t.Fatalf("build hotel: %v", err)
	}
	if !findEvtByType(evts, events.EventHotelBuilt) {
		t.Error("expected HotelBuilt event")
	}
	if newState.Ownerships[1].Buildings != models.HotelLevel {
		t.Errorf("expected hotel (5), got %d", newState.Ownerships[1].Buildings)
	}
	if newState.Bank.Houses != 32+4 {
		t.Errorf("expected 36 houses returned to bank, got %d", newState.Bank.Houses)
	}
	if newState.Bank.Hotels != 11 {
		t.Errorf("expected 11 hotels in bank, got %d", newState.Bank.Hotels)
	}
}

func TestTradeReject(t *testing.T) {
	state, _, p1, p2 := newActiveGame(t)
	registry := reducers.NewRegistry()

	state.Ownerships[1] = models.Ownership{SpaceIndex: 1, OwnerID: p1}
	player1 := findPlayerState(&state, p1)
	addSpaceIndex(player1, 1)
	state.Players[0] = *player1

	newState, _, err := registry.Dispatch(state, &commands.TradeProposeCommand{
		BaseCommand:    commands.BaseCommand{PlayerID: p1},
		RecipientID:    p2,
		OfferedSpaces:  []int{1},
		RequestedMoney: 100,
	})
	if err != nil {
		t.Fatalf("trade propose: %v", err)
	}

	newState, evts, err := registry.Dispatch(newState, &commands.TradeRejectCommand{
		BaseCommand: commands.BaseCommand{PlayerID: p2},
		TradeID:     newState.Trades[0].TradeID,
	})
	if err != nil {
		t.Fatalf("trade reject: %v", err)
	}
	if !findEvtByType(evts, events.EventTradeRejected) {
		t.Error("expected TradeRejected event")
	}
	if newState.Trades[0].Status != models.TradeStatusRejected {
		t.Errorf("expected REJECTED status, got %s", newState.Trades[0].Status)
	}
}

func TestMortgageWithBuildingsFails(t *testing.T) {
	state, _, p1, _ := newActiveGame(t)
	registry := reducers.NewRegistry()

	state.Ownerships[1] = models.Ownership{SpaceIndex: 1, OwnerID: p1, Buildings: 2}
	player := findPlayerState(&state, p1)
	addSpaceIndex(player, 1)
	state.Players[0] = *player

	_, _, err := registry.Dispatch(state, &commands.MortgagePropertyCommand{
		BaseCommand: commands.BaseCommand{PlayerID: p1},
		SpaceIndex:  1,
	})
	if err == nil {
		t.Error("expected error mortgaging property with buildings")
	}
}

func TestBuildHouseEvenRule(t *testing.T) {
	state, _, p1, _ := newActiveGame(t)
	registry := reducers.NewRegistry()

	state.Ownerships[1] = models.Ownership{SpaceIndex: 1, OwnerID: p1, Buildings: 2}
	state.Ownerships[3] = models.Ownership{SpaceIndex: 3, OwnerID: p1, Buildings: 1}
	player := findPlayerState(&state, p1)
	addSpaceIndex(player, 1)
	addSpaceIndex(player, 3)
	state.Players[0] = *player

	_, _, err := registry.Dispatch(state, &commands.BuildHouseCommand{
		BaseCommand: commands.BaseCommand{PlayerID: p1},
		SpaceIndex:  1,
	})
	if err == nil {
		t.Error("expected error for even building rule violation")
	}
}

func TestUnsupportedCommandDispatch(t *testing.T) {
	registry := reducers.NewRegistry()
	state := newTestState()

	_, _, err := registry.Dispatch(state, &commands.DeclareBankruptcyCommand{
		BaseCommand: commands.BaseCommand{PlayerID: uuid.New()},
	})
	if err != nil && strings.Contains(err.Error(), "unsupported command") {
		t.Fatalf("DeclareBankruptcy should be registered: %v", err)
	}

	state2 := newTestState()
	state2.Status = models.GameStatusActive
	_, _, err = registry.Dispatch(state2, &commands.DeclareBankruptcyCommand{
		BaseCommand: commands.BaseCommand{PlayerID: uuid.New()},
	})
	if err == nil {
		t.Error("expected error for non-existent player")
	}
}

func addSpaceIndex(player *models.PlayerState, idx int) {
	for _, i := range player.OwnedSpaceIndexes {
		if i == idx {
			return
		}
	}
	player.OwnedSpaceIndexes = append(player.OwnedSpaceIndexes, idx)
}

func findPlayerState(state *models.GameState, id uuid.UUID) *models.PlayerState {
	for i := range state.Players {
		if state.Players[i].PlayerID == id {
			return &state.Players[i]
		}
	}
	return nil
}

func TestRollDiceInvalidDice(t *testing.T) {
	state, _, p1, _ := newActiveGame(t)
	registry := reducers.NewRegistry()

	for _, d := range [][2]int{{0, 3}, {7, 1}, {3, -1}} {
		_, _, err := registry.Dispatch(state, &commands.RollDiceCommand{
			BaseCommand: commands.BaseCommand{PlayerID: p1},
			Die1:        d[0],
			Die2:        d[1],
		})
		if err == nil {
			t.Errorf("expected error for dice %v", d)
		}
	}
}

func TestDispatchDoesNotMutateInput(t *testing.T) {
	state, _, p1, _ := newActiveGame(t)
	registry := reducers.NewRegistry()
	before := replay.StateHash(state)

	if _, _, err := registry.Dispatch(state, &commands.RollDiceCommand{
		BaseCommand: commands.BaseCommand{PlayerID: p1},
		Die1:        5,
		Die2:        6,
	}); err != nil {
		t.Fatalf("roll: %v", err)
	}

	if after := replay.StateHash(state); after != before {
		t.Error("Dispatch mutated the input state")
	}
}

func TestEventSequenceIsContinuous(t *testing.T) {
	state, gameID, p1, _ := newActiveGame(t)
	registry := reducers.NewRegistry()
	start := state.EventSequence

	newState, evts, err := registry.Dispatch(state, &commands.RollDiceCommand{
		BaseCommand: commands.BaseCommand{PlayerID: p1},
		Die1:        2,
		Die2:        3,
	})
	if err != nil {
		t.Fatalf("roll: %v", err)
	}
	for i, e := range evts {
		if e.GetSequenceNumber() != start+int64(i)+1 {
			t.Errorf("event %d: expected seq %d, got %d", i, start+int64(i)+1, e.GetSequenceNumber())
		}
		if e.GetGameID() != gameID {
			t.Errorf("event %d: expected game id %s, got %s", i, gameID, e.GetGameID())
		}
	}
	if newState.EventSequence != start+int64(len(evts)) {
		t.Errorf("expected state sequence %d, got %d", start+int64(len(evts)), newState.EventSequence)
	}
}

func TestBuyPropertyInsufficientFunds(t *testing.T) {
	state, _, p1, _ := newActiveGame(t)
	registry := reducers.NewRegistry()

	state, _, err := registry.Dispatch(state, &commands.RollDiceCommand{
		BaseCommand: commands.BaseCommand{PlayerID: p1},
		Die1:        1,
		Die2:        2,
	})
	if err != nil {
		t.Fatalf("roll: %v", err)
	}
	state.Players[0].Money = 10

	if _, _, err := registry.Dispatch(state, &commands.BuyPropertyCommand{
		BaseCommand: commands.BaseCommand{PlayerID: p1},
	}); err == nil {
		t.Error("expected error buying with insufficient funds")
	}
}

package runtime_test

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github/M2A96/Monopoly.git/internal/game/commands"
	"github/M2A96/Monopoly.git/internal/game/events"
	"github/M2A96/Monopoly.git/internal/game/models"
	"github/M2A96/Monopoly.git/internal/game/persistence/memory"
	"github/M2A96/Monopoly.git/internal/game/replay"
	"github/M2A96/Monopoly.git/internal/game/runtime"
	"github/M2A96/Monopoly.git/pkg/board"
	"github/M2A96/Monopoly.git/pkg/dice"
	"github.com/google/uuid"
)

// --- fixtures ---

type env struct {
	t     *testing.T
	log   *memory.CommandLog
	snaps *memory.SnapshotStore
	mgr   *runtime.Manager
	cfg   runtime.Config
	game  uuid.UUID
	ids   []uuid.UUID
}

func newEnv(t *testing.T, cfg runtime.Config) *env {
	t.Helper()
	if cfg.NewDice == nil {
		cfg.NewDice = func() dice.Provider { return dice.NewProvider(42) }
	}
	e := &env{
		t:     t,
		log:   memory.NewCommandLog(),
		snaps: memory.NewSnapshotStore(),
		cfg:   cfg,
		game:  uuid.New(),
	}
	e.mgr = runtime.NewManager(e.deps(), cfg)
	t.Cleanup(func() { _ = e.mgr.Shutdown(context.Background()) })
	return e
}

func (e *env) deps() runtime.Deps {
	return runtime.Deps{Log: e.log, Snapshots: e.snaps}
}

func (e *env) base(id uuid.UUID) commands.BaseCommand {
	return commands.BaseCommand{GameID: e.game, PlayerID: id}
}

func (e *env) submit(cmd commands.Command) (runtime.Result, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return e.mgr.Submit(ctx, e.game, cmd)
}

func (e *env) must(cmd commands.Command) runtime.Result {
	e.t.Helper()
	res, err := e.submit(cmd)
	if err != nil {
		e.t.Fatalf("%s: %v", cmd.GetType(), err)
	}
	return res
}

func (e *env) state() models.GameState {
	e.t.Helper()
	r, err := e.mgr.Activate(context.Background(), e.game)
	if err != nil {
		e.t.Fatal(err)
	}
	s, err := r.State(context.Background())
	if err != nil {
		e.t.Fatal(err)
	}
	return s
}

func (e *env) start(players int) {
	e.t.Helper()
	e.must(&commands.CreateGameCommand{BaseCommand: e.base(uuid.Nil), Name: "Runtime"})
	for i := 0; i < players; i++ {
		id := uuid.New()
		e.ids = append(e.ids, id)
		e.must(&commands.JoinGameCommand{BaseCommand: e.base(id), Name: fmt.Sprintf("P%d", i)})
	}
	e.must(&commands.StartGameCommand{BaseCommand: e.base(e.ids[0])})
}

// play drives the game with a simple policy for n accepted commands.
func (e *env) play(n int) {
	e.t.Helper()
	for i := 0; i < n; i++ {
		s := e.state()
		if s.Status != models.GameStatusActive {
			return
		}
		if _, err := e.submit(nextCommand(e, s)); err != nil {
			e.t.Fatalf("command %d: %v", i, err)
		}
	}
}

func nextCommand(e *env, s models.GameState) commands.Command {
	if len(s.Debts) > 0 {
		d := s.Debts[0]
		if p := player(s, d.DebtorID); p.Money >= d.Amount {
			return &commands.PayDebtCommand{BaseCommand: e.base(d.DebtorID)}
		}
		return &commands.DeclareBankruptcyCommand{BaseCommand: e.base(d.DebtorID)}
	}
	if s.Auction != nil {
		return &commands.AuctionPassCommand{BaseCommand: e.base(s.Auction.CurrentBidderID())}
	}
	active := s.Turn.ActivePlayerID
	if w := s.Turn.ActionWindow; w != nil && w.Type == models.ActionWaitingForPurchaseDecision {
		if player(s, active).Money >= board.SpaceByIndex[w.SpaceIndex].Price+100 {
			return &commands.BuyPropertyCommand{BaseCommand: e.base(active)}
		}
		return &commands.DeclinePurchaseCommand{BaseCommand: e.base(active)}
	}
	if s.Turn.Phase == models.TurnPhaseEnd {
		return &commands.EndTurnCommand{BaseCommand: e.base(active)}
	}
	return &commands.RollDiceCommand{BaseCommand: e.base(active)}
}

func player(s models.GameState, id uuid.UUID) *models.PlayerState {
	for i := range s.Players {
		if s.Players[i].PlayerID == id {
			return &s.Players[i]
		}
	}
	return nil
}

func hasEvent(evts []events.Event, typ events.EventType) bool {
	for _, ev := range evts {
		if ev.GetType() == typ {
			return true
		}
	}
	return false
}

// --- 4.1 / 4.2 processing loop ---

func TestRoomAppliesCommands(t *testing.T) {
	e := newEnv(t, runtime.Config{})
	e.start(2)
	res := e.must(&commands.RollDiceCommand{BaseCommand: e.base(e.ids[0])})
	if !hasEvent(res.Events, events.EventDiceRolled) {
		t.Fatal("expected DiceRolled event")
	}
	if res.CommandSeq != 5 || e.log.Len(e.game) != 5 {
		t.Fatalf("expected seq 5 and 5 logged commands, got %d / %d", res.CommandSeq, e.log.Len(e.game))
	}
}

func TestRejectedCommandIsNotLogged(t *testing.T) {
	e := newEnv(t, runtime.Config{})
	e.start(2)
	if _, err := e.submit(&commands.EndTurnCommand{BaseCommand: e.base(e.ids[1])}); err == nil {
		t.Fatal("expected out-of-turn EndTurn to be rejected")
	}
	if e.log.Len(e.game) != 4 {
		t.Fatalf("rejected command must not be logged, log has %d", e.log.Len(e.game))
	}
}

func TestWrongGameRejected(t *testing.T) {
	e := newEnv(t, runtime.Config{})
	e.start(2)
	r, _ := e.mgr.Get(e.game)
	_, err := r.Submit(context.Background(), &commands.EndTurnCommand{
		BaseCommand: commands.BaseCommand{GameID: uuid.New(), PlayerID: e.ids[0]},
	})
	if !errors.Is(err, runtime.ErrWrongGame) {
		t.Fatalf("expected ErrWrongGame, got %v", err)
	}
}

// --- 4.4 sequential processing ---

func TestConcurrentSubmittersAreSequenced(t *testing.T) {
	e := newEnv(t, runtime.Config{})
	e.start(4)

	// Trades between distinct pairs are valid in any order, so every one must be accepted.
	const perPlayer = 25
	var wg sync.WaitGroup
	for i, from := range e.ids {
		to := e.ids[(i+1)%len(e.ids)]
		wg.Add(1)
		go func(from, to uuid.UUID) {
			defer wg.Done()
			for k := 0; k < perPlayer; k++ {
				if _, err := e.submit(&commands.TradeProposeCommand{
					BaseCommand: e.base(from), RecipientID: to, OfferedMoney: 1,
				}); err != nil {
					t.Errorf("trade: %v", err)
				}
			}
		}(from, to)
	}
	wg.Wait()

	s := e.state()
	want := len(e.ids) * perPlayer
	if len(s.Trades) != want {
		t.Fatalf("expected %d trades, got %d", want, len(s.Trades))
	}
	entries, _ := e.log.Load(context.Background(), e.game, 0)
	for i, en := range entries {
		if en.Seq != int64(i+1) {
			t.Fatalf("log not contiguous at %d: seq %d", i, en.Seq)
		}
	}
	seen := map[uuid.UUID]bool{}
	for _, tr := range s.Trades {
		if seen[tr.TradeID] {
			t.Fatalf("duplicate trade id %s: concurrent mutation", tr.TradeID)
		}
		seen[tr.TradeID] = true
	}
}

// --- 4.10 dice injection ---

func TestRoomInjectsDiceAndLogsThem(t *testing.T) {
	e := newEnv(t, runtime.Config{})
	e.start(2)
	e.must(&commands.RollDiceCommand{BaseCommand: e.base(e.ids[0]), Die1: 6, Die2: 6})

	want1, want2 := dice.NewProvider(42).Roll()
	entries, _ := e.log.Load(context.Background(), e.game, 4)
	got := entries[0].Command.(*commands.RollDiceCommand)
	if got.Die1 != want1 || got.Die2 != want2 {
		t.Fatalf("expected injected dice (%d,%d), logged (%d,%d)", want1, want2, got.Die1, got.Die2)
	}
}

// --- 4.5 subscribers ---

func TestSubscriberReceivesOrderedUpdates(t *testing.T) {
	e := newEnv(t, runtime.Config{})
	e.start(2)
	r, _ := e.mgr.Get(e.game)
	ch, err := r.Subscribe(context.Background(), "spectator")
	if err != nil {
		t.Fatal(err)
	}
	first := <-ch
	if !first.Initial || first.CommandSeq != 4 || first.State.Status != models.GameStatusActive {
		t.Fatalf("unexpected initial update: %+v", first)
	}

	e.play(10)
	last := first.CommandSeq
	for i := 0; i < 10; i++ {
		upd := <-ch
		if upd.CommandSeq != last+1 {
			t.Fatalf("expected seq %d, got %d", last+1, upd.CommandSeq)
		}
		last = upd.CommandSeq
	}

	// Mutating a received state must not affect the room.
	first.State.Players[0].Money = -1
	if e.state().Players[0].Money < 0 {
		t.Fatal("subscriber mutation leaked into room state")
	}
}

func TestSlowSubscriberIsDroppedWithoutBlocking(t *testing.T) {
	e := newEnv(t, runtime.Config{SubscriberBuffer: 2})
	e.start(2)
	r, _ := e.mgr.Get(e.game)
	slow, _ := r.Subscribe(context.Background(), "slow")

	done := make(chan struct{})
	go func() { e.play(20); close(done) }()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("room blocked on a slow subscriber")
	}

	n := 0
	for range slow {
		n++
	}
	if n > 2 {
		t.Fatalf("slow subscriber should have been dropped after its buffer filled, got %d updates", n)
	}
}

// --- 4.6 snapshots / 4.7 recovery ---

func TestSnapshotEveryNCommands(t *testing.T) {
	e := newEnv(t, runtime.Config{SnapshotEveryCommands: 5})
	e.start(2)
	e.play(6)
	snap, err := e.snaps.Latest(context.Background(), e.game)
	if err != nil {
		t.Fatal(err)
	}
	if snap.CommandSeq != 10 {
		t.Fatalf("expected checkpoint at seq 10, got %d", snap.CommandSeq)
	}
}

func TestSnapshotInterval(t *testing.T) {
	e := newEnv(t, runtime.Config{SnapshotInterval: 20 * time.Millisecond})
	e.start(2)
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if snap, err := e.snaps.Latest(context.Background(), e.game); err == nil && snap.CommandSeq == 4 {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatal("expected a timed snapshot of the started game")
}

func TestCrashRecoveryFromCommandLogOnly(t *testing.T) {
	e := newEnv(t, runtime.Config{})
	e.start(3)
	e.play(60)
	want := replay.StateHash(e.state())

	// Simulate a crash: a fresh manager over the same storage, no snapshot was ever taken.
	mgr2 := runtime.NewManager(e.deps(), e.cfg)
	defer mgr2.Shutdown(context.Background())
	r, err := mgr2.Activate(context.Background(), e.game)
	if err != nil {
		t.Fatal(err)
	}
	s, _ := r.State(context.Background())
	if replay.StateHash(s) != want {
		t.Fatal("recovered state differs from live state")
	}
}

func TestRecoveryFromSnapshotPlusTrailingCommands(t *testing.T) {
	e := newEnv(t, runtime.Config{SnapshotEveryCommands: 25})
	e.start(3)
	e.play(60)
	want := replay.StateHash(e.state())
	snap, _ := e.snaps.Latest(context.Background(), e.game)
	if snap.CommandSeq == 0 || snap.CommandSeq == int64(e.log.Len(e.game)) {
		t.Fatalf("test needs trailing commands after the snapshot (snap %d, log %d)", snap.CommandSeq, e.log.Len(e.game))
	}

	mgr2 := runtime.NewManager(e.deps(), e.cfg)
	defer mgr2.Shutdown(context.Background())
	r, err := mgr2.Activate(context.Background(), e.game)
	if err != nil {
		t.Fatal(err)
	}
	s, _ := r.State(context.Background())
	if replay.StateHash(s) != want {
		t.Fatal("snapshot + trailing replay differs from live state")
	}
}

func TestSnapshotFromOtherRulesVersionIsIgnored(t *testing.T) {
	e := newEnv(t, runtime.Config{})
	e.start(2)
	e.play(10)
	want := replay.StateHash(e.state())

	bogus := models.GameState{GameID: e.game, Ownerships: map[int]models.Ownership{}}
	_ = e.snaps.Save(context.Background(), runtime.Snapshot{
		GameID: e.game, State: bogus, CommandSeq: int64(e.log.Len(e.game)), RulesVersion: "0",
	})

	mgr2 := runtime.NewManager(e.deps(), e.cfg)
	defer mgr2.Shutdown(context.Background())
	r, err := mgr2.Activate(context.Background(), e.game)
	if err != nil {
		t.Fatal(err)
	}
	s, _ := r.State(context.Background())
	if replay.StateHash(s) != want {
		t.Fatal("expected full replay to ignore an incompatible snapshot")
	}
}

func TestConcurrentActivationYieldsOneRoom(t *testing.T) {
	e := newEnv(t, runtime.Config{})
	e.start(2)
	_ = e.mgr.Shutdown(context.Background())

	mgr := runtime.NewManager(e.deps(), e.cfg)
	defer mgr.Shutdown(context.Background())
	var wg sync.WaitGroup
	rooms := make([]*runtime.Room, 20)
	for i := range rooms {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			rooms[i], _ = mgr.Activate(context.Background(), e.game)
		}(i)
	}
	wg.Wait()
	for _, r := range rooms {
		if r != rooms[0] || r == nil {
			t.Fatal("concurrent activation created more than one room")
		}
	}
}

// --- 4.8 idle cleanup ---

func TestIdleRoomStopsAndReactivates(t *testing.T) {
	e := newEnv(t, runtime.Config{IdleTimeout: 30 * time.Millisecond})
	e.start(2)
	r, _ := e.mgr.Get(e.game)

	select {
	case <-r.Done():
	case <-time.After(2 * time.Second):
		t.Fatal("idle room did not stop")
	}
	if e.mgr.Active() != 0 {
		t.Fatal("stopped room still registered")
	}
	snap, err := e.snaps.Latest(context.Background(), e.game)
	if err != nil || snap.CommandSeq != 4 {
		t.Fatalf("expected final snapshot at seq 4, got %+v / %v", snap.CommandSeq, err)
	}

	res := e.must(&commands.RollDiceCommand{BaseCommand: e.base(e.ids[0])})
	if res.CommandSeq != 5 {
		t.Fatalf("expected reactivated room to continue at seq 5, got %d", res.CommandSeq)
	}
}

// --- 4.11 turn timer ---

func TestTurnTimeoutResignsAbsentPlayer(t *testing.T) {
	e := newEnv(t, runtime.Config{TurnTimeout: 30 * time.Millisecond})
	e.start(3)
	r, _ := e.mgr.Get(e.game)
	ch, _ := r.Subscribe(context.Background(), "watcher")
	<-ch // initial

	select {
	case upd := <-ch:
		if !hasEvent(upd.Events, events.EventPlayerResigned) {
			t.Fatalf("expected PlayerResigned, got %v", upd.Events)
		}
		if p := player(upd.State, e.ids[0]); !p.IsBankrupt {
			t.Fatal("expected the idle active player to be eliminated")
		}
		if upd.State.Turn.ActivePlayerID != e.ids[1] {
			t.Fatal("expected the turn to pass to the next player")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("turn timeout never fired")
	}
}

func TestShutdownWritesFinalSnapshot(t *testing.T) {
	e := newEnv(t, runtime.Config{})
	e.start(2)
	e.play(3)
	if err := e.mgr.Shutdown(context.Background()); err != nil {
		t.Fatal(err)
	}
	snap, err := e.snaps.Latest(context.Background(), e.game)
	if err != nil || snap.CommandSeq != int64(e.log.Len(e.game)) {
		t.Fatalf("expected final snapshot at the last command, got %d / %v", snap.CommandSeq, err)
	}
	if _, err := e.submit(&commands.EndTurnCommand{BaseCommand: e.base(e.ids[0])}); !errors.Is(err, runtime.ErrManagerClosed) {
		t.Fatalf("expected ErrManagerClosed after shutdown, got %v", err)
	}
}

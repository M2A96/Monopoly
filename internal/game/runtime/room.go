// Package runtime runs active games in memory. Each Room owns one goroutine that is the
// single writer of its GameState: commands are processed strictly in arrival order.
package runtime

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github/M2A96/Monopoly.git/internal/game/commands"
	"github/M2A96/Monopoly.git/internal/game/events"
	"github/M2A96/Monopoly.git/internal/game/models"
	"github/M2A96/Monopoly.git/internal/game/reducers"
	"github/M2A96/Monopoly.git/pkg/dice"
)

var (
	ErrRoomClosed = errors.New("runtime: room closed")
	ErrWrongGame  = errors.New("runtime: command targets another game")
	ErrLogAppend  = errors.New("runtime: command could not be logged")
	ErrRecovery   = errors.New("runtime: recovery failed")
)

// Config tunes a room. Zero durations/counts disable the corresponding feature.
type Config struct {
	QueueSize             int           // buffered inbox size (default 64)
	SubscriberBuffer      int           // per-subscriber buffer (default 64)
	SnapshotEveryCommands int           // checkpoint after N commands
	SnapshotInterval      time.Duration // checkpoint every interval if state changed
	TurnTimeout           time.Duration // resign the awaited player after this long
	IdleTimeout           time.Duration // stop the room after this long without commands
	NewDice               func() dice.Provider
	Now                   func() time.Time
	OnError               func(gameID uuid.UUID, err error) // background errors (snapshots)
}

func (c Config) withDefaults() Config {
	if c.QueueSize <= 0 {
		c.QueueSize = 64
	}
	if c.SubscriberBuffer <= 0 {
		c.SubscriberBuffer = 64
	}
	if c.NewDice == nil {
		c.NewDice = dice.NewCryptoProvider
	}
	if c.Now == nil {
		c.Now = time.Now
	}
	if c.OnError == nil {
		c.OnError = func(uuid.UUID, error) {}
	}
	return c
}

// Deps are the collaborators a room needs.
type Deps struct {
	Log       CommandLog
	Snapshots SnapshotStore
	// NewReducer defaults to reducers.NewRegistry.
	NewReducer func() reducers.Reducer
}

// Update is pushed to subscribers after every accepted command.
// State is a private deep copy; Events must be treated as read-only.
type Update struct {
	GameID     uuid.UUID
	CommandSeq int64
	Events     []events.Event
	State      models.GameState
	Initial    bool // the first update after subscribing carries only the current state
}

// Result is returned to the submitter of an accepted command.
type Result struct {
	CommandSeq int64
	Events     []events.Event
}

type submitReq struct {
	cmd   commands.Command
	reply chan submitResp
}

type submitResp struct {
	res Result
	err error
}

type stateReq struct{ reply chan models.GameState }

type subReq struct {
	id    string
	reply chan (<-chan Update)
}

type unsubReq struct{ id string }

// Room is one active game.
type Room struct {
	gameID  uuid.UUID
	cfg     Config
	deps    Deps
	reducer reducers.Reducer
	dice    dice.Provider
	onStop  func(*Room)

	// Owned by the loop goroutine only.
	state       models.GameState
	seq         int64
	lastSnapSeq int64
	subs        map[string]chan Update

	inbox    chan any
	stop     chan struct{}
	stopOnce sync.Once
	done     chan struct{}
}

// StartRoom recovers the game from storage (Step 4.7) and starts its loop.
func StartRoom(ctx context.Context, gameID uuid.UUID, deps Deps, cfg Config, onStop func(*Room)) (*Room, error) {
	cfg = cfg.withDefaults()
	if deps.NewReducer == nil {
		deps.NewReducer = func() reducers.Reducer { return reducers.NewRegistry() }
	}
	reducer := deps.NewReducer()

	state, seq, snapSeq, err := recoverState(ctx, gameID, deps, reducer)
	if err != nil {
		return nil, err
	}

	r := &Room{
		gameID:      gameID,
		cfg:         cfg,
		deps:        deps,
		reducer:     reducer,
		dice:        cfg.NewDice(),
		onStop:      onStop,
		state:       state,
		seq:         seq,
		lastSnapSeq: snapSeq,
		subs:        make(map[string]chan Update),
		inbox:       make(chan any, cfg.QueueSize),
		stop:        make(chan struct{}),
		done:        make(chan struct{}),
	}
	go r.run()
	return r, nil
}

// recoverState loads the latest compatible snapshot and replays the commands after it.
func recoverState(ctx context.Context, gameID uuid.UUID, deps Deps, reducer reducers.Reducer) (models.GameState, int64, int64, error) {
	state := models.GameState{Ownerships: make(map[int]models.Ownership)}
	var from int64

	snap, err := deps.Snapshots.Latest(ctx, gameID)
	switch {
	case err == nil && snap.RulesVersion == reducers.RulesVersion:
		state, from = snap.State, snap.CommandSeq
	case err == nil:
		// Snapshot from another rule set: rebuild everything from the command log.
	case !errors.Is(err, ErrNoSnapshot):
		return state, 0, 0, fmt.Errorf("%w: load snapshot: %v", ErrRecovery, err)
	}
	snapSeq := from

	entries, err := deps.Log.Load(ctx, gameID, from)
	if err != nil {
		return state, 0, 0, fmt.Errorf("%w: load commands: %v", ErrRecovery, err)
	}
	for _, e := range entries {
		if e.Seq != from+1 {
			return state, 0, 0, fmt.Errorf("%w: command log gap at seq %d (have %d)", ErrRecovery, e.Seq, from)
		}
		state, _, err = reducer.Reduce(state, e.Command)
		if err != nil {
			return state, 0, 0, fmt.Errorf("%w: replay seq %d (%s): %v", ErrRecovery, e.Seq, e.Command.GetType(), err)
		}
		from = e.Seq
	}
	return state, from, snapSeq, nil
}

func (r *Room) GameID() uuid.UUID     { return r.gameID }
func (r *Room) Done() <-chan struct{} { return r.done }

// Submit queues a command and waits for its outcome. Commands from all callers are
// applied one at a time in queue order.
func (r *Room) Submit(ctx context.Context, cmd commands.Command) (Result, error) {
	if id := cmd.GetGameID(); id != uuid.Nil && id != r.gameID {
		return Result{}, ErrWrongGame
	}
	req := submitReq{cmd: cmd, reply: make(chan submitResp, 1)}
	if err := r.send(ctx, req); err != nil {
		return Result{}, err
	}
	select {
	case resp := <-req.reply:
		return resp.res, resp.err
	case <-r.done:
		select {
		case resp := <-req.reply:
			return resp.res, resp.err
		default:
			return Result{}, ErrRoomClosed
		}
	case <-ctx.Done():
		return Result{}, ctx.Err()
	}
}

// State returns a deep copy of the current state.
func (r *Room) State(ctx context.Context) (models.GameState, error) {
	req := stateReq{reply: make(chan models.GameState, 1)}
	if err := r.send(ctx, req); err != nil {
		return models.GameState{}, err
	}
	select {
	case s := <-req.reply:
		return s, nil
	case <-r.done:
		return models.GameState{}, ErrRoomClosed
	case <-ctx.Done():
		return models.GameState{}, ctx.Err()
	}
}

// Subscribe registers a subscriber. The channel first receives the current state
// (Initial), then one Update per accepted command. It is closed when the room stops,
// on Unsubscribe, or if the subscriber falls behind by more than SubscriberBuffer updates.
func (r *Room) Subscribe(ctx context.Context, id string) (<-chan Update, error) {
	req := subReq{id: id, reply: make(chan (<-chan Update), 1)}
	if err := r.send(ctx, req); err != nil {
		return nil, err
	}
	select {
	case ch := <-req.reply:
		return ch, nil
	case <-r.done:
		return nil, ErrRoomClosed
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (r *Room) Unsubscribe(id string) {
	select {
	case r.inbox <- unsubReq{id: id}:
	case <-r.done:
	}
}

// Close stops the room after persisting a final snapshot.
func (r *Room) Close(ctx context.Context) error {
	r.stopOnce.Do(func() { close(r.stop) })
	select {
	case <-r.done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (r *Room) send(ctx context.Context, req any) error {
	select {
	case <-r.done:
		return ErrRoomClosed
	default:
	}
	select {
	case r.inbox <- req:
		return nil
	case <-r.done:
		return ErrRoomClosed
	case <-ctx.Done():
		return ctx.Err()
	}
}

// --- loop (single writer) ---

func (r *Room) run() {
	defer close(r.done)

	var snapTick <-chan time.Time
	if r.cfg.SnapshotInterval > 0 {
		t := time.NewTicker(r.cfg.SnapshotInterval)
		defer t.Stop()
		snapTick = t.C
	}
	turnTimer := newStoppedTimer()
	idleTimer := newStoppedTimer()
	defer turnTimer.Stop()
	defer idleTimer.Stop()
	r.resetTurnTimer(turnTimer)
	resetTimer(idleTimer, r.cfg.IdleTimeout)

	for {
		select {
		case <-r.stop:
			r.shutdown()
			return
		case <-idleTimer.C:
			r.shutdown()
			return
		case <-snapTick:
			r.checkpoint()
		case <-turnTimer.C:
			r.handleTurnTimeout()
			r.resetTurnTimer(turnTimer)
		case msg := <-r.inbox:
			switch m := msg.(type) {
			case submitReq:
				res, err := r.apply(m.cmd)
				m.reply <- submitResp{res: res, err: err}
				if err == nil {
					r.resetTurnTimer(turnTimer)
				}
				resetTimer(idleTimer, r.cfg.IdleTimeout)
			case stateReq:
				m.reply <- r.state.Clone()
			case subReq:
				m.reply <- r.subscribe(m.id)
			case unsubReq:
				if ch, ok := r.subs[m.id]; ok {
					close(ch)
					delete(r.subs, m.id)
				}
			}
		}
	}
}

// apply runs one command: inject dice → validate+reduce → log → publish → checkpoint.
func (r *Room) apply(cmd commands.Command) (Result, error) {
	if roll, ok := cmd.(*commands.RollDiceCommand); ok {
		injected := *roll
		injected.Die1, injected.Die2 = r.dice.Roll()
		cmd = &injected
	}

	next, evts, err := r.reducer.Reduce(r.state, cmd)
	if err != nil {
		return Result{}, err
	}

	entry := LogEntry{Seq: r.seq + 1, Command: cmd, AcceptedAt: r.cfg.Now()}
	if err := r.deps.Log.Append(context.Background(), r.gameID, entry); err != nil {
		return Result{}, fmt.Errorf("%w: %v", ErrLogAppend, err)
	}

	r.state = next
	r.seq = entry.Seq
	r.publish(evts)

	if n := int64(r.cfg.SnapshotEveryCommands); n > 0 && r.seq-r.lastSnapSeq >= n {
		r.checkpoint()
	}
	return Result{CommandSeq: r.seq, Events: evts}, nil
}

func (r *Room) subscribe(id string) <-chan Update {
	if old, ok := r.subs[id]; ok {
		close(old)
	}
	ch := make(chan Update, r.cfg.SubscriberBuffer)
	ch <- Update{GameID: r.gameID, CommandSeq: r.seq, State: r.state.Clone(), Initial: true}
	r.subs[id] = ch
	return ch
}

// publish never blocks the loop: a subscriber whose buffer is full is dropped.
func (r *Room) publish(evts []events.Event) {
	for id, ch := range r.subs {
		upd := Update{GameID: r.gameID, CommandSeq: r.seq, Events: evts, State: r.state.Clone()}
		select {
		case ch <- upd:
		default:
			close(ch)
			delete(r.subs, id)
		}
	}
}

func (r *Room) checkpoint() {
	if r.seq == r.lastSnapSeq {
		return
	}
	snap := Snapshot{
		GameID:       r.gameID,
		State:        r.state.Clone(),
		CommandSeq:   r.seq,
		RulesVersion: reducers.RulesVersion,
		TakenAt:      r.cfg.Now(),
	}
	if err := r.deps.Snapshots.Save(context.Background(), snap); err != nil {
		r.cfg.OnError(r.gameID, fmt.Errorf("snapshot at seq %d: %w", r.seq, err))
		return
	}
	r.lastSnapSeq = r.seq
}

func (r *Room) shutdown() {
	r.checkpoint()
	for id, ch := range r.subs {
		close(ch)
		delete(r.subs, id)
	}
	if r.onStop != nil {
		r.onStop(r)
	}
}

// --- turn timer (Step 4.11) ---

// awaitedPlayer is the player the game is currently waiting on.
func awaitedPlayer(s models.GameState) uuid.UUID {
	switch {
	case len(s.Debts) > 0:
		return s.Debts[0].DebtorID
	case s.Auction != nil:
		return s.Auction.CurrentBidderID()
	default:
		return s.Turn.ActivePlayerID
	}
}

func (r *Room) resetTurnTimer(t *time.Timer) {
	if r.cfg.TurnTimeout <= 0 || r.state.Status != models.GameStatusActive {
		stopTimer(t)
		return
	}
	resetTimer(t, r.cfg.TurnTimeout)
}

func (r *Room) handleTurnTimeout() {
	if r.state.Status != models.GameStatusActive {
		return
	}
	who := awaitedPlayer(r.state)
	if who == uuid.Nil {
		return
	}
	_, err := r.apply(&commands.ResignCommand{
		BaseCommand: commands.BaseCommand{GameID: r.gameID, PlayerID: who},
		Reason:      commands.ResignTimeout,
	})
	if err != nil {
		r.cfg.OnError(r.gameID, fmt.Errorf("timeout resign of %s: %w", who, err))
	}
}

// --- timer helpers ---

func newStoppedTimer() *time.Timer {
	t := time.NewTimer(time.Hour)
	stopTimer(t)
	return t
}

func stopTimer(t *time.Timer) {
	if !t.Stop() {
		select {
		case <-t.C:
		default:
		}
	}
}

func resetTimer(t *time.Timer, d time.Duration) {
	stopTimer(t)
	if d > 0 {
		t.Reset(d)
	}
}

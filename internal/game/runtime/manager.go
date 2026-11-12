package runtime

import (
	"context"
	"errors"
	"sync"

	"github.com/google/uuid"
	"github/M2A96/Monopoly.git/internal/game/commands"
)

// Manager owns the set of active rooms on this node (Step 4.3). It guarantees at most
// one room per game: concurrent activations of the same game share a single recovery.
type Manager struct {
	deps Deps
	cfg  Config

	mu         sync.Mutex
	rooms      map[uuid.UUID]*Room
	activating map[uuid.UUID]*activation
	closed     bool
}

type activation struct {
	done chan struct{}
	room *Room
	err  error
}

var ErrManagerClosed = errors.New("runtime: manager shut down")

func NewManager(deps Deps, cfg Config) *Manager {
	return &Manager{
		deps:       deps,
		cfg:        cfg,
		rooms:      make(map[uuid.UUID]*Room),
		activating: make(map[uuid.UUID]*activation),
	}
}

// Get returns the active room for a game, if any.
func (m *Manager) Get(gameID uuid.UUID) (*Room, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	r, ok := m.rooms[gameID]
	return r, ok
}

// Active reports how many rooms are running.
func (m *Manager) Active() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.rooms)
}

// Activate returns the game's room, recovering it from storage if it is not running.
func (m *Manager) Activate(ctx context.Context, gameID uuid.UUID) (*Room, error) {
	m.mu.Lock()
	if m.closed {
		m.mu.Unlock()
		return nil, ErrManagerClosed
	}
	if r, ok := m.rooms[gameID]; ok {
		m.mu.Unlock()
		return r, nil
	}
	if a, ok := m.activating[gameID]; ok {
		m.mu.Unlock()
		select {
		case <-a.done:
			return a.room, a.err
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	a := &activation{done: make(chan struct{})}
	m.activating[gameID] = a
	m.mu.Unlock()

	a.room, a.err = StartRoom(ctx, gameID, m.deps, m.cfg, m.forget)

	m.mu.Lock()
	delete(m.activating, gameID)
	if a.err == nil {
		m.rooms[gameID] = a.room
	}
	m.mu.Unlock()
	close(a.done)
	return a.room, a.err
}

// Submit routes a command to its game's room, activating it if needed. If the room
// stops (idle cleanup) between lookup and delivery, it is reactivated once.
func (m *Manager) Submit(ctx context.Context, gameID uuid.UUID, cmd commands.Command) (Result, error) {
	for attempt := 0; ; attempt++ {
		r, err := m.Activate(ctx, gameID)
		if err != nil {
			return Result{}, err
		}
		res, err := r.Submit(ctx, cmd)
		if errors.Is(err, ErrRoomClosed) && attempt == 0 {
			<-r.Done()
			continue
		}
		return res, err
	}
}

// Shutdown stops every room, each persisting a final snapshot.
func (m *Manager) Shutdown(ctx context.Context) error {
	m.mu.Lock()
	m.closed = true
	rooms := make([]*Room, 0, len(m.rooms))
	for _, r := range m.rooms {
		rooms = append(rooms, r)
	}
	m.mu.Unlock()

	var errs []error
	for _, r := range rooms {
		if err := r.Close(ctx); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

// forget is called by a room's loop when it stops.
func (m *Manager) forget(r *Room) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if cur, ok := m.rooms[r.gameID]; ok && cur == r {
		delete(m.rooms, r.gameID)
	}
}

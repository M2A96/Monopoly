// Package memory provides in-process implementations of the runtime storage ports.
// They back tests and single-node development until Phase 6 adds durable stores.
package memory

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github/M2A96/Monopoly.git/internal/game/runtime"
)

type CommandLog struct {
	mu   sync.Mutex
	logs map[uuid.UUID][]runtime.LogEntry
}

func NewCommandLog() *CommandLog {
	return &CommandLog{logs: make(map[uuid.UUID][]runtime.LogEntry)}
}

func (l *CommandLog) Append(_ context.Context, gameID uuid.UUID, entry runtime.LogEntry) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	want := int64(len(l.logs[gameID])) + 1
	if entry.Seq != want {
		return fmt.Errorf("%w: got %d, want %d", runtime.ErrSeqConflict, entry.Seq, want)
	}
	l.logs[gameID] = append(l.logs[gameID], entry)
	return nil
}

func (l *CommandLog) Load(_ context.Context, gameID uuid.UUID, afterSeq int64) ([]runtime.LogEntry, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	all := l.logs[gameID]
	if afterSeq < 0 {
		afterSeq = 0
	}
	if afterSeq >= int64(len(all)) {
		return nil, nil
	}
	return append([]runtime.LogEntry(nil), all[afterSeq:]...), nil
}

// Len reports how many commands are logged for a game (test helper).
func (l *CommandLog) Len(gameID uuid.UUID) int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return len(l.logs[gameID])
}

type SnapshotStore struct {
	mu    sync.Mutex
	snaps map[uuid.UUID]runtime.Snapshot
	saves int
}

func NewSnapshotStore() *SnapshotStore {
	return &SnapshotStore{snaps: make(map[uuid.UUID]runtime.Snapshot)}
}

func (s *SnapshotStore) Save(_ context.Context, snap runtime.Snapshot) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if cur, ok := s.snaps[snap.GameID]; ok && cur.CommandSeq > snap.CommandSeq {
		return nil // never move a checkpoint backwards
	}
	snap.State = snap.State.Clone()
	s.snaps[snap.GameID] = snap
	s.saves++
	return nil
}

func (s *SnapshotStore) Latest(_ context.Context, gameID uuid.UUID) (runtime.Snapshot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	snap, ok := s.snaps[gameID]
	if !ok {
		return runtime.Snapshot{}, runtime.ErrNoSnapshot
	}
	snap.State = snap.State.Clone()
	return snap, nil
}

// Saves reports how many snapshots were written (test helper).
func (s *SnapshotStore) Saves() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.saves
}

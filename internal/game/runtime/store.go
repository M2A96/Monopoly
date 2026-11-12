package runtime

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github/M2A96/Monopoly.git/internal/game/commands"
	"github/M2A96/Monopoly.git/internal/game/models"
)

// LogEntry is one accepted command. Seq starts at 1 and is contiguous per game.
// The command carries any injected values (dice), so replay never needs the provider.
type LogEntry struct {
	Seq        int64
	Command    commands.Command
	AcceptedAt time.Time
}

// CommandLog is the append-only source of truth for a game (Decision 4.0).
type CommandLog interface {
	// Append must reject an entry whose Seq is not exactly last+1 with ErrSeqConflict.
	Append(ctx context.Context, gameID uuid.UUID, entry LogEntry) error
	// Load returns the entries with Seq > afterSeq, in order.
	Load(ctx context.Context, gameID uuid.UUID, afterSeq int64) ([]LogEntry, error)
}

// Snapshot is a checkpoint of a room's state after CommandSeq commands.
type Snapshot struct {
	GameID       uuid.UUID
	State        models.GameState
	CommandSeq   int64
	RulesVersion string
	TakenAt      time.Time
}

// SnapshotStore keeps the latest snapshot per game.
type SnapshotStore interface {
	Save(ctx context.Context, snap Snapshot) error
	// Latest returns ErrNoSnapshot when the game has none.
	Latest(ctx context.Context, gameID uuid.UUID) (Snapshot, error)
}

var (
	ErrSeqConflict = errors.New("command log: sequence conflict")
	ErrNoSnapshot  = errors.New("snapshot store: no snapshot")
)

package replay

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github/M2A96/Monopoly.git/internal/game/commands"
	"github/M2A96/Monopoly.git/internal/game/events"
	"github/M2A96/Monopoly.git/internal/game/models"
	"github/M2A96/Monopoly.git/internal/game/reducers"
)

type ReplayResult struct {
	FinalState models.GameState
	AllEvents  []events.Event
	Turns      int
}

func Replay(initialState models.GameState, cmds []commands.Command) (ReplayResult, error) {
	registry := reducers.NewRegistry()

	state := initialState
	if state.Ownerships == nil {
		state.Ownerships = make(map[int]models.Ownership)
	}

	var allEvents []events.Event
	maxTurns := 1000
	turnCount := 0

	for i, cmd := range cmds {
		newState, evts, err := registry.Dispatch(state, cmd)
		if err != nil {
			return ReplayResult{}, fmt.Errorf("command %d (%s) failed: %w", i, cmd.GetType(), err)
		}

		state = newState
		allEvents = append(allEvents, evts...)
		turnCount++

		if state.Status == models.GameStatusFinished {
			break
		}

		if turnCount > maxTurns {
			return ReplayResult{}, fmt.Errorf("replay exceeded %d turns", maxTurns)
		}
	}

	return ReplayResult{
		FinalState: state,
		AllEvents:  allEvents,
		Turns:      turnCount,
	}, nil
}

// ErrEventReplayNotImplemented is returned until events carry enough data to rebuild state.
// Until then, state is reconstructed by replaying the recorded commands (see Replay).
var ErrEventReplayNotImplemented = errors.New("replay from events not implemented: replay recorded commands instead")

func ReplayFromEvents(initialState models.GameState, recordedEvents []events.Event) (models.GameState, error) {
	_ = recordedEvents
	return initialState, ErrEventReplayNotImplemented
}

// StateHash returns a SHA-256 over the canonical JSON encoding of the full state.
// encoding/json sorts map keys, so equal states always hash equally.
func StateHash(state models.GameState) string {
	b, err := json.Marshal(state)
	if err != nil {
		// GameState contains only JSON-encodable types; this is a programming error.
		panic(fmt.Sprintf("replay: state not encodable: %v", err))
	}
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

// CompareStates reports whether two states are identical in every field.
func CompareStates(a, b models.GameState) bool {
	return StateHash(a) == StateHash(b)
}

func VerifyDeterminism(state models.GameState, cmds1, cmds2 []commands.Command) error {
	r1, err := Replay(state, cmds1)
	if err != nil {
		return fmt.Errorf("first replay failed: %w", err)
	}

	r2, err := Replay(state, cmds2)
	if err != nil {
		return fmt.Errorf("second replay failed: %w", err)
	}

	if !CompareStates(r1.FinalState, r2.FinalState) {
		return fmt.Errorf("non-deterministic: different final states for same commands")
	}

	return nil
}

type GameStateSnapshot struct {
	GameID   uuid.UUID
	State    models.GameState
	EventSeq int64
}

func Snapshot(state models.GameState, seq int64) GameStateSnapshot {
	return GameStateSnapshot{
		GameID:   state.GameID,
		State:    state,
		EventSeq: seq,
	}
}

func RestoreFromSnapshot(snap GameStateSnapshot) models.GameState {
	return snap.State
}

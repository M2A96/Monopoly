package replay_test

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github/M2A96/Monopoly.git/internal/game/commands"
	"github/M2A96/Monopoly.git/internal/game/models"
	"github/M2A96/Monopoly.git/internal/game/replay"
)

var (
	gameID = uuid.MustParse("00000000-0000-0000-0000-0000000000a1")
	p1     = uuid.MustParse("00000000-0000-0000-0000-0000000000b1")
	p2     = uuid.MustParse("00000000-0000-0000-0000-0000000000b2")
)

func setupCommands() []commands.Command {
	return []commands.Command{
		&commands.CreateGameCommand{BaseCommand: commands.BaseCommand{GameID: gameID, PlayerID: p1}, Name: "Replay"},
		&commands.JoinGameCommand{BaseCommand: commands.BaseCommand{GameID: gameID, PlayerID: p1}, Name: "Alice"},
		&commands.JoinGameCommand{BaseCommand: commands.BaseCommand{GameID: gameID, PlayerID: p2}, Name: "Bob"},
		&commands.StartGameCommand{BaseCommand: commands.BaseCommand{GameID: gameID, PlayerID: p1}},
	}
}

func emptyState() models.GameState {
	return models.GameState{Ownerships: make(map[int]models.Ownership)}
}

// playScript records a bot-driven game: randomness comes from a seeded dice provider
// outside the reducer and is injected into RollDice commands.
func playScript(t *testing.T, seed int64, turns int) []commands.Command {
	t.Helper()
	return playGame(t, seed, setupCommands(), turns*4).cmds
}

func TestReplayIsDeterministic(t *testing.T) {
	cmds := playScript(t, 42, 30)

	r1, err := replay.Replay(emptyState(), cmds)
	if err != nil {
		t.Fatalf("replay 1: %v", err)
	}
	r2, err := replay.Replay(emptyState(), cmds)
	if err != nil {
		t.Fatalf("replay 2: %v", err)
	}

	if replay.StateHash(r1.FinalState) != replay.StateHash(r2.FinalState) {
		t.Fatal("same commands produced different final states")
	}
	if len(r1.AllEvents) != len(r2.AllEvents) {
		t.Fatalf("event count differs: %d vs %d", len(r1.AllEvents), len(r2.AllEvents))
	}
	for i := range r1.AllEvents {
		a, b := r1.AllEvents[i], r2.AllEvents[i]
		if a.GetType() != b.GetType() || a.GetSequenceNumber() != b.GetSequenceNumber() {
			t.Fatalf("event %d differs: %s#%d vs %s#%d", i, a.GetType(), a.GetSequenceNumber(), b.GetType(), b.GetSequenceNumber())
		}
	}
}

func TestReplayMatchesLiveState(t *testing.T) {
	cmds := playScript(t, 7, 20)
	full, err := replay.Replay(emptyState(), cmds)
	if err != nil {
		t.Fatalf("full replay: %v", err)
	}

	// Replaying in two chunks (as recovery from a snapshot would) yields the same state.
	half := len(cmds) / 2
	first, err := replay.Replay(emptyState(), cmds[:half])
	if err != nil {
		t.Fatalf("first half: %v", err)
	}
	snap := replay.Snapshot(first.FinalState, first.FinalState.EventSequence)
	second, err := replay.Replay(replay.RestoreFromSnapshot(snap), cmds[half:])
	if err != nil {
		t.Fatalf("second half: %v", err)
	}

	if !replay.CompareStates(full.FinalState, second.FinalState) {
		t.Fatal("chunked replay diverged from full replay")
	}
}

func TestDifferentDiceDifferentState(t *testing.T) {
	a, err := replay.Replay(emptyState(), playScript(t, 1, 10))
	if err != nil {
		t.Fatal(err)
	}
	b, err := replay.Replay(emptyState(), playScript(t, 2, 10))
	if err != nil {
		t.Fatal(err)
	}
	if replay.CompareStates(a.FinalState, b.FinalState) {
		t.Fatal("different dice sequences produced identical states")
	}
}

func TestReplayStopsOnInvalidCommand(t *testing.T) {
	cmds := append(setupCommands(), &commands.RollDiceCommand{
		BaseCommand: commands.BaseCommand{GameID: gameID, PlayerID: p2}, // not p2's turn
		Die1:        1,
		Die2:        2,
	})
	if _, err := replay.Replay(emptyState(), cmds); err == nil {
		t.Fatal("expected replay to fail on out-of-turn roll")
	}
}

func TestReplayFromEventsFailsLoudly(t *testing.T) {
	_, err := replay.ReplayFromEvents(emptyState(), nil)
	if !errors.Is(err, replay.ErrEventReplayNotImplemented) {
		t.Fatalf("expected ErrEventReplayNotImplemented, got %v", err)
	}
}

func TestScriptCoversManyTurns(t *testing.T) {
	cmds := playScript(t, 42, 30)
	r, err := replay.Replay(emptyState(), cmds)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%d commands, %d events, turn %d", len(cmds), len(r.AllEvents), r.FinalState.Turn.Number)
	if len(cmds) < 40 {
		t.Fatalf("script too short to exercise replay: %d commands", len(cmds))
	}
}

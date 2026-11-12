package replay_test

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github/M2A96/Monopoly.git/internal/game/commands"
	"github/M2A96/Monopoly.git/internal/game/models"
	"github/M2A96/Monopoly.git/internal/game/replay"
)

func setupN(n int) []commands.Command {
	cmds := []commands.Command{&commands.CreateGameCommand{
		BaseCommand: commands.BaseCommand{GameID: gameID, PlayerID: p1}, Name: "Sim",
	}}
	for i := 0; i < n; i++ {
		id := uuid.NewSHA1(gameID, []byte(fmt.Sprintf("player-%d", i)))
		cmds = append(cmds, &commands.JoinGameCommand{BaseCommand: commands.BaseCommand{GameID: gameID, PlayerID: id}, Name: fmt.Sprintf("Bot%d", i)})
	}
	first := cmds[1].GetPlayerID()
	return append(cmds, &commands.StartGameCommand{BaseCommand: commands.BaseCommand{GameID: gameID, PlayerID: first}})
}

// TestFullGameSimulation plays many complete games with invariants checked after every
// command, and verifies that replaying the recorded commands reproduces the final state.
func TestFullGameSimulation(t *testing.T) {
	const maxCmds = 20000
	finished := 0
	for seed := int64(1); seed <= 30; seed++ {
		players := 2 + int(seed%4)
		t.Run(fmt.Sprintf("seed%d_%dp", seed, players), func(t *testing.T) {
			b := playGame(t, seed, setupN(players), maxCmds)
			if b.state.Status == models.GameStatusFinished {
				finished++
			}

			r, err := replay.Replay(emptyState(), b.cmds)
			if err != nil {
				t.Fatalf("replay: %v", err)
			}
			if !replay.CompareStates(r.FinalState, b.state) {
				t.Fatal("replay diverged from live game")
			}
		})
	}
	t.Logf("%d/30 games reached a winner within %d commands", finished, maxCmds)
	if finished == 0 {
		t.Error("no simulated game finished: bankruptcy/win path never exercised")
	}
}

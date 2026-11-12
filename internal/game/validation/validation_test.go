package validation_test

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github/M2A96/Monopoly.git/internal/game/commands"
	"github/M2A96/Monopoly.git/internal/game/models"
	"github/M2A96/Monopoly.git/internal/game/validation"
)

var (
	p1 = uuid.MustParse("00000000-0000-0000-0000-0000000000b1")
	p2 = uuid.MustParse("00000000-0000-0000-0000-0000000000b2")
)

func activeState() models.GameState {
	return models.GameState{
		GameID:     uuid.MustParse("00000000-0000-0000-0000-0000000000a1"),
		Status:     models.GameStatusActive,
		Players:    []models.PlayerState{{PlayerID: p1, Money: 1500}, {PlayerID: p2, Money: 1500}},
		Ownerships: map[int]models.Ownership{1: {SpaceIndex: 1, OwnerID: p1, Buildings: 2}},
		Turn:       models.Turn{Number: 1, ActivePlayerID: p1, Phase: models.TurnPhasePreRoll},
	}
}

func TestValidatorRejects(t *testing.T) {
	v := validation.NewValidator()
	base1 := commands.BaseCommand{PlayerID: p1}
	base2 := commands.BaseCommand{PlayerID: p2}

	waiting := activeState()
	waiting.Status = models.GameStatusWaiting

	endPhase := activeState()
	endPhase.Turn.Phase = models.TurnPhaseEnd

	cases := []struct {
		name  string
		state models.GameState
		cmd   commands.Command
	}{
		{"invalid turn", activeState(), &commands.RollDiceCommand{BaseCommand: base2, Die1: 1, Die2: 2}},
		{"invalid phase", endPhase, &commands.RollDiceCommand{BaseCommand: base1, Die1: 1, Die2: 2}},
		{"game not active", waiting, &commands.EndTurnCommand{BaseCommand: base1}},
		{"buy without pending decision", activeState(), &commands.BuyPropertyCommand{BaseCommand: base1}},
		{"hotel without 4 houses", activeState(), &commands.BuildHotelCommand{BaseCommand: base1, SpaceIndex: 1}},
		{"mortgage with buildings", activeState(), &commands.MortgagePropertyCommand{BaseCommand: base1, SpaceIndex: 1}},
		{"build on someone else's property", activeState(), &commands.BuildHouseCommand{BaseCommand: base2, SpaceIndex: 1}},
		{"jail fine when not in jail", activeState(), &commands.PayJailFineCommand{BaseCommand: base1}},
		{"trade with yourself", activeState(), &commands.TradeProposeCommand{BaseCommand: base1, RecipientID: p1}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := v.Validate(tc.state, tc.cmd)
			var ve *validation.ValidationError
			if !errors.As(err, &ve) {
				t.Fatalf("expected ValidationError, got %v", err)
			}
		})
	}
}

func TestValidatorAccepts(t *testing.T) {
	v := validation.NewValidator()
	if err := v.Validate(activeState(), &commands.RollDiceCommand{
		BaseCommand: commands.BaseCommand{PlayerID: p1}, Die1: 3, Die2: 4,
	}); err != nil {
		t.Fatalf("expected valid roll, got %v", err)
	}
}

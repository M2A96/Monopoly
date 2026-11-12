package reducers

import (
	"fmt"

	"github.com/google/uuid"
	"github/M2A96/Monopoly.git/internal/game/commands"
	"github/M2A96/Monopoly.git/internal/game/events"
	"github/M2A96/Monopoly.git/internal/game/models"
)

func reduceEndTurn(state models.GameState, cmd commands.Command) (models.GameState, []events.Event, error) {
	c, ok := cmd.(*commands.EndTurnCommand)
	if !ok {
		return state, nil, fmt.Errorf("expected EndTurnCommand, got %T", cmd)
	}

	if state.Status != models.GameStatusActive {
		return state, nil, fmt.Errorf("game not active")
	}

	if c.PlayerID != state.Turn.ActivePlayerID {
		return state, nil, fmt.Errorf("not your turn")
	}

	player := findPlayer(&state, c.PlayerID)
	if player == nil {
		return state, nil, fmt.Errorf("player not in game")
	}

	if player.IsBankrupt {
		return state, nil, fmt.Errorf("bankrupt player cannot end turn")
	}

	state = copyState(state)

	nextPlayerID := nextActivePlayer(&state, c.PlayerID)
	if nextPlayerID == uuid.Nil {
		return state, nil, fmt.Errorf("no active players remaining")
	}

	isWrap := false
	for _, p := range state.Players {
		if p.PlayerID == c.PlayerID {
			for _, q := range state.Players {
				if q.PlayerID == nextPlayerID {
					isWrap = true
					break
				}
			}
			break
		}
	}
	_ = isWrap

	turnNumber := state.Turn.Number
	nextIdx := findPlayerIndex(&state, nextPlayerID)
	currentIdx := findPlayerIndex(&state, c.PlayerID)
	if nextIdx < currentIdx {
		turnNumber++
	}

	evts := []events.Event{
		&events.TurnEndedEvent{
			BaseEvent:  events.BaseEvent{Type: events.EventTurnEnded},
			PlayerID:   c.PlayerID,
			TurnNumber: state.Turn.Number,
		},
	}

	state.Turn = models.Turn{
		Number:         turnNumber,
		ActivePlayerID: nextPlayerID,
		Phase:          models.TurnPhasePreRoll,
		DiceRoll:       nil,
		DoublesCount:   0,
		ActionWindow:   nil,
	}

	evts = append(evts, &events.TurnStartedEvent{
		BaseEvent:  events.BaseEvent{Type: events.EventTurnStarted},
		PlayerID:   nextPlayerID,
		TurnNumber: turnNumber,
	})

	return state, evts, nil
}

func reduceDeclareBankruptcy(state models.GameState, cmd commands.Command) (models.GameState, []events.Event, error) {
	c, ok := cmd.(*commands.DeclareBankruptcyCommand)
	if !ok {
		return state, nil, fmt.Errorf("expected DeclareBankruptcyCommand, got %T", cmd)
	}

	if state.Status != models.GameStatusActive {
		return state, nil, fmt.Errorf("game not active")
	}

	player := findPlayer(&state, c.PlayerID)
	if player == nil {
		return state, nil, fmt.Errorf("player not in game")
	}

	if player.IsBankrupt {
		return state, nil, fmt.Errorf("already bankrupt")
	}

	state = copyState(state)
	player = findPlayer(&state, c.PlayerID)
	player.IsBankrupt = true

	creditorID := uuid.Nil
	if state.Turn.ActionWindow != nil && state.Turn.ActionWindow.Type == models.ActionWaitingForBankruptcyResolution {
		creditorID = state.Turn.ActionWindow.PlayerID
	}

	creditor := findPlayer(&state, creditorID)

	for _, spaceIdx := range player.OwnedSpaceIndexes {
		o, ok := state.Ownerships[spaceIdx]
		if !ok {
			continue
		}
		if creditor != nil && creditorID != uuid.Nil {
			o.OwnerID = creditorID
			addSpaceIndex(creditor, spaceIdx)
		} else {
			o.OwnerID = uuid.Nil
			o.Mortgaged = true
			o.Buildings = 0
		}
		state.Ownerships[spaceIdx] = o
	}
	player.OwnedSpaceIndexes = nil

	if creditor != nil && creditorID != uuid.Nil {
		creditor.Money += player.Money
	}
	player.Money = 0

	state.Turn.ActionWindow = nil

	evts := []events.Event{
		&events.BankruptcyDeclaredEvent{
			BaseEvent:  events.BaseEvent{Type: events.EventBankruptcyDeclared},
			PlayerID:   c.PlayerID,
			CreditorID: creditorID,
		},
	}

	if activePlayers(&state) == 1 {
		var winnerID uuid.UUID
		for _, p := range state.Players {
			if !p.IsBankrupt {
				winnerID = p.PlayerID
				break
			}
		}
		state.Status = models.GameStatusFinished
		evts = append(evts, &events.GameEndedEvent{
			BaseEvent: events.BaseEvent{Type: events.EventGameEnded},
			WinnerID:  winnerID,
		})
	}

	return state, evts, nil
}

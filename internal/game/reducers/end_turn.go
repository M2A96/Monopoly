package reducers

import (
	"fmt"

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

	if len(state.Debts) > 0 || state.Auction != nil {
		return state, nil, fmt.Errorf("pending debts or auction must be resolved first")
	}

	if state.Turn.Phase != models.TurnPhaseEnd {
		return state, nil, fmt.Errorf("cannot end turn in phase %s", state.Turn.Phase)
	}

	state = copyState(state)
	return state, advanceTurn(&state, c.PlayerID), nil
}

func reducePayDebt(state models.GameState, cmd commands.Command) (models.GameState, []events.Event, error) {
	c, ok := cmd.(*commands.PayDebtCommand)
	if !ok {
		return state, nil, fmt.Errorf("expected PayDebtCommand, got %T", cmd)
	}

	if state.Status != models.GameStatusActive {
		return state, nil, fmt.Errorf("game not active")
	}

	idx := firstDebtOf(&state, c.PlayerID)
	if idx == -1 {
		return state, nil, fmt.Errorf("no pending debt")
	}

	debt := state.Debts[idx]
	player := findPlayer(&state, c.PlayerID)
	if player.Money < debt.Amount {
		return state, nil, fmt.Errorf("insufficient funds: have %d, owe %d (sell, mortgage or declare bankruptcy)", player.Money, debt.Amount)
	}

	state = copyState(state)
	player = findPlayer(&state, c.PlayerID)
	transfer(&state, player, debt.CreditorID, debt.Amount)
	state.Debts = append(state.Debts[:idx], state.Debts[idx+1:]...)
	if len(state.Debts) == 0 {
		state.Debts = nil
	}
	settlePhase(&state)

	return state, []events.Event{&events.DebtPaidEvent{
		BaseEvent:  events.BaseEvent{Type: events.EventDebtPaid},
		DebtorID:   debt.DebtorID,
		CreditorID: debt.CreditorID,
		Amount:     debt.Amount,
	}}, nil
}

// reduceDeclareBankruptcy eliminates a player who cannot pay a pending debt.
// Their assets go to the creditor of that debt (a player or the bank).
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

	idx := firstDebtOf(&state, c.PlayerID)
	if idx == -1 {
		return state, nil, fmt.Errorf("no pending debt: use Resign to leave the game")
	}
	creditorID := state.Debts[idx].CreditorID

	state = copyState(state)
	evts := []events.Event{&events.BankruptcyDeclaredEvent{
		BaseEvent:  events.BaseEvent{Type: events.EventBankruptcyDeclared},
		PlayerID:   c.PlayerID,
		CreditorID: creditorID,
	}}
	return state, append(evts, eliminate(&state, c.PlayerID, creditorID)...), nil
}

// reduceResign eliminates a player voluntarily or on timeout; assets return to the bank.
func reduceResign(state models.GameState, cmd commands.Command) (models.GameState, []events.Event, error) {
	c, ok := cmd.(*commands.ResignCommand)
	if !ok {
		return state, nil, fmt.Errorf("expected ResignCommand, got %T", cmd)
	}

	if state.Status != models.GameStatusActive {
		return state, nil, fmt.Errorf("game not active")
	}

	player := findPlayer(&state, c.PlayerID)
	if player == nil {
		return state, nil, fmt.Errorf("player not in game")
	}
	if player.IsBankrupt {
		return state, nil, fmt.Errorf("already eliminated")
	}

	reason := c.Reason
	if reason == "" {
		reason = commands.ResignVoluntary
	}

	state = copyState(state)
	evts := []events.Event{&events.PlayerResignedEvent{
		BaseEvent: events.BaseEvent{Type: events.EventPlayerResigned},
		PlayerID:  c.PlayerID,
		Reason:    string(reason),
	}}
	return state, append(evts, eliminate(&state, c.PlayerID, bank)...), nil
}

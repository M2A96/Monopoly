package reducers

import (
	"fmt"

	"github/M2A96/Monopoly.git/internal/game/commands"
	"github/M2A96/Monopoly.git/internal/game/events"
	"github/M2A96/Monopoly.git/internal/game/models"
	"github/M2A96/Monopoly.git/pkg/board"
)

func reduceMortgageProperty(state models.GameState, cmd commands.Command) (models.GameState, []events.Event, error) {
	c, ok := cmd.(*commands.MortgagePropertyCommand)
	if !ok {
		return state, nil, fmt.Errorf("expected MortgagePropertyCommand, got %T", cmd)
	}

	if state.Status != models.GameStatusActive {
		return state, nil, fmt.Errorf("game not active")
	}

	space, ok := board.SpaceByIndex[c.SpaceIndex]
	if !ok {
		return state, nil, fmt.Errorf("invalid space index")
	}

	o, owned := state.Ownerships[c.SpaceIndex]
	if !owned || o.OwnerID != c.PlayerID {
		return state, nil, fmt.Errorf("not your property")
	}

	if o.Mortgaged {
		return state, nil, fmt.Errorf("already mortgaged")
	}

	if o.Buildings > 0 {
		return state, nil, fmt.Errorf("must sell buildings before mortgaging")
	}

	state = copyState(state)
	player := findPlayer(&state, c.PlayerID)
	o = state.Ownerships[c.SpaceIndex]

	player.Money += space.MortgageValue
	o.Mortgaged = true
	state.Ownerships[c.SpaceIndex] = o

	return state, []events.Event{
		&events.PropertyMortgagedEvent{
			BaseEvent:  events.BaseEvent{Type: events.EventPropertyMortgaged},
			PlayerID:   c.PlayerID,
			SpaceIndex: c.SpaceIndex,
			Amount:     space.MortgageValue,
		},
	}, nil
}

func reduceUnmortgageProperty(state models.GameState, cmd commands.Command) (models.GameState, []events.Event, error) {
	c, ok := cmd.(*commands.UnmortgagePropertyCommand)
	if !ok {
		return state, nil, fmt.Errorf("expected UnmortgagePropertyCommand, got %T", cmd)
	}

	if state.Status != models.GameStatusActive {
		return state, nil, fmt.Errorf("game not active")
	}

	space, ok := board.SpaceByIndex[c.SpaceIndex]
	if !ok {
		return state, nil, fmt.Errorf("invalid space index")
	}

	o, owned := state.Ownerships[c.SpaceIndex]
	if !owned || o.OwnerID != c.PlayerID {
		return state, nil, fmt.Errorf("not your property")
	}

	if !o.Mortgaged {
		return state, nil, fmt.Errorf("not mortgaged")
	}

	cost := (space.MortgageValue * 110) / 100

	player := findPlayer(&state, c.PlayerID)
	if player == nil {
		return state, nil, fmt.Errorf("player not in game")
	}

	if player.Money < cost {
		return state, nil, fmt.Errorf("insufficient funds: have %d, need %d", player.Money, cost)
	}

	state = copyState(state)
	player = findPlayer(&state, c.PlayerID)
	o = state.Ownerships[c.SpaceIndex]

	player.Money -= cost
	o.Mortgaged = false
	state.Ownerships[c.SpaceIndex] = o

	return state, []events.Event{
		&events.PropertyUnmortgagedEvent{
			BaseEvent:  events.BaseEvent{Type: events.EventPropertyUnmortgaged},
			PlayerID:   c.PlayerID,
			SpaceIndex: c.SpaceIndex,
			Cost:       cost,
		},
	}, nil
}

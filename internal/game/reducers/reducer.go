package reducers

import (
	"fmt"

	"github/M2A96/Monopoly.git/internal/game/commands"
	"github/M2A96/Monopoly.git/internal/game/events"
	"github/M2A96/Monopoly.git/internal/game/models"
	"github/M2A96/Monopoly.git/internal/game/validation"
)

type ReducerFunc func(state models.GameState, cmd commands.Command) (models.GameState, []events.Event, error)

// Reducer is the Phase 2 engine contract: a pure state transition.
type Reducer interface {
	Reduce(state models.GameState, cmd commands.Command) (models.GameState, []events.Event, error)
}

var _ Reducer = (*Registry)(nil)

// Registry dispatches commands to their reducer. It holds no mutable game data:
// everything needed to reproduce a transition lives in GameState and the command.
type Registry struct {
	reducers  map[commands.CommandType]ReducerFunc
	validator *validation.Validator
}

func NewRegistry() *Registry {
	r := &Registry{
		reducers:  make(map[commands.CommandType]ReducerFunc),
		validator: validation.NewValidator(),
	}
	r.registerDefaults()
	return r
}

func (r *Registry) Register(cmdType commands.CommandType, fn ReducerFunc) {
	r.reducers[cmdType] = fn
}

func (r *Registry) Reduce(state models.GameState, cmd commands.Command) (models.GameState, []events.Event, error) {
	return r.Dispatch(state, cmd)
}

// Dispatch validates the command, runs its reducer and stamps the emitted events
// with the game ID and sequence numbers taken from state.
func (r *Registry) Dispatch(state models.GameState, cmd commands.Command) (models.GameState, []events.Event, error) {
	fn, ok := r.reducers[cmd.GetType()]
	if !ok {
		return state, nil, fmt.Errorf("unsupported command: %s", cmd.GetType())
	}

	if r.validator != nil {
		if err := r.validator.Validate(state, cmd); err != nil {
			return state, nil, err
		}
	}

	newState, evts, err := fn(state, cmd)
	if err != nil {
		return state, nil, err
	}

	seq := newState.EventSequence
	events.StampEvents(evts, newState.GameID, &seq)
	newState.EventSequence = seq

	return newState, evts, nil
}

func (r *Registry) registerDefaults() {
	r.Register(commands.CommandCreateGame, reduceCreateGame)
	r.Register(commands.CommandJoinGame, reduceJoinGame)
	r.Register(commands.CommandStartGame, reduceStartGame)
	r.Register(commands.CommandRollDice, reduceRollDice)
	r.Register(commands.CommandBuyProperty, reduceBuyProperty)
	r.Register(commands.CommandDeclinePurchase, reduceDeclinePurchase)
	r.Register(commands.CommandEndTurn, reduceEndTurn)
	r.Register(commands.CommandBuildHouse, reduceBuildHouse)
	r.Register(commands.CommandBuildHotel, reduceBuildHotel)
	r.Register(commands.CommandSellHouse, reduceSellHouse)
	r.Register(commands.CommandSellHotel, reduceSellHotel)
	r.Register(commands.CommandMortgageProperty, reduceMortgageProperty)
	r.Register(commands.CommandUnmortgageProperty, reduceUnmortgageProperty)
	r.Register(commands.CommandTradePropose, reduceTradePropose)
	r.Register(commands.CommandTradeAccept, reduceTradeAccept)
	r.Register(commands.CommandTradeReject, reduceTradeReject)
	r.Register(commands.CommandAuctionBid, reduceAuctionBid)
	r.Register(commands.CommandPayJailFine, reducePayJailFine)
	r.Register(commands.CommandUseGetOutOfJailCard, reduceUseGetOutOfJailCard)
	r.Register(commands.CommandDeclareBankruptcy, reduceDeclareBankruptcy)
}

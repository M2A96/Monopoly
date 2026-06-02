package models

import "github.com/google/uuid"

type TurnPhase string

type DiceRoll interface {
	Total() int
	IsDoubles() bool
}

const (
	TurnPhasePreRoll TurnPhase = "PRE_ROLL"
	TurnPhaseLand    TurnPhase = "LAND"
	TurnPhaseAction  TurnPhase = "ACTION"
	TurnPhaseEnd     TurnPhase = "END"
)

var _ DiceRoll = (*diceRoll)(nil)

// diceRoll holds one pair of dice values injected by the runtime.
// Reducers never generate randomness — dice values are always provided externally.
type diceRoll struct {
	Die1 int
	Die2 int
}

func (d *diceRoll) Total() int      { return d.Die1 + d.Die2 }
func (d *diceRoll) IsDoubles() bool { return d.Die1 == d.Die2 }

// NewDiceRoll constructs an immutable dice roll.
func NewDiceRoll(die1, die2 int) DiceRoll {
	return &diceRoll{Die1: die1, Die2: die2}
}

// Turn tracks the state of the current player's turn.
type Turn struct {
	Number         int
	ActivePlayerID uuid.UUID
	Phase          TurnPhase
	DiceRoll       DiceRoll      // nil until the player has rolled
	DoublesCount   int           // consecutive doubles this turn (3 sends player to jail)
	ActionWindow   *ActionWindow // the expected next action; nil if none pending
}

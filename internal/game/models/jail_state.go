package models

// JailState tracks how many turns a player has spent in jail.
// TurnsInJail increments each turn the player remains in jail (max 3).
// On the third turn the player must pay the fine and move.
type JailState struct {
	TurnsInJail int
}

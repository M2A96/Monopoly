package reducers

import (
	"fmt"

	"github.com/google/uuid"
	"github/M2A96/Monopoly.git/internal/game/commands"
	"github/M2A96/Monopoly.git/internal/game/events"
	"github/M2A96/Monopoly.git/internal/game/models"
)

func reduceCreateGame(state models.GameState, cmd commands.Command) (models.GameState, []events.Event, error) {
	c, ok := cmd.(*commands.CreateGameCommand)
	if !ok {
		return state, nil, fmt.Errorf("expected CreateGameCommand, got %T", cmd)
	}

	// GameID is injected by the caller (command or pre-seeded state) so the reducer stays pure.
	gameID := c.GetGameID()
	if gameID == uuid.Nil {
		gameID = state.GameID
	}
	if gameID == uuid.Nil {
		return state, nil, fmt.Errorf("game id required")
	}

	state = copyState(state)
	state.GameID = gameID
	state.Status = models.GameStatusWaiting
	state.Players = nil
	state.Bank = models.NewBank()
	if state.Ownerships == nil {
		state.Ownerships = make(map[int]models.Ownership)
	}
	state.CardDecks = models.CardDecks{
		Chance:         models.DefaultChanceDeck(),
		CommunityChest: models.DefaultCommunityChestDeck(),
	}
	state.Auction = nil
	state.Trades = nil
	state.Turn = models.Turn{}

	return state, []events.Event{
		&events.GameCreatedEvent{
			BaseEvent: events.BaseEvent{Type: events.EventGameCreated},
			Name:      c.Name,
		},
	}, nil
}

func reduceJoinGame(state models.GameState, cmd commands.Command) (models.GameState, []events.Event, error) {
	c, ok := cmd.(*commands.JoinGameCommand)
	if !ok {
		return state, nil, fmt.Errorf("expected JoinGameCommand, got %T", cmd)
	}

	if state.Status != models.GameStatusWaiting {
		return state, nil, fmt.Errorf("cannot join game in status %s", state.Status)
	}

	if len(state.Players) >= 8 {
		return state, nil, fmt.Errorf("game is full (max 8 players)")
	}

	for _, p := range state.Players {
		if p.PlayerID == c.PlayerID {
			return state, nil, fmt.Errorf("player already in game")
		}
	}

	state = copyState(state)
	order := len(state.Players)
	state.Players = append(state.Players, models.PlayerState{
		PlayerID:          c.PlayerID,
		Name:              c.Name,
		Money:             1500,
		Position:          0,
		IsInJail:          false,
		IsBankrupt:        false,
		OwnedSpaceIndexes: nil,
		GetOutOfJailCards: 0,
		PendingActions:    nil,
	})

	return state, []events.Event{
		&events.PlayerJoinedEvent{
			BaseEvent: events.BaseEvent{Type: events.EventPlayerJoined},
			PlayerID:  c.PlayerID,
			Name:      c.Name,
			Order:     order,
		},
	}, nil
}

func reduceStartGame(state models.GameState, cmd commands.Command) (models.GameState, []events.Event, error) {
	_ = cmd

	if state.Status != models.GameStatusWaiting {
		return state, nil, fmt.Errorf("game not in WAITING status")
	}

	if len(state.Players) < 2 {
		return state, nil, fmt.Errorf("need at least 2 players to start")
	}

	state = copyState(state)
	state.Status = models.GameStatusActive

	shuffleCards(state.CardDecks.Chance, state.GameID[:])
	shuffleCards(state.CardDecks.CommunityChest, state.GameID[8:])

	state.Turn = models.Turn{
		Number:         1,
		ActivePlayerID: state.Players[0].PlayerID,
		Phase:          models.TurnPhasePreRoll,
		DiceRoll:       nil,
		DoublesCount:   0,
		ActionWindow:   nil,
	}

	evts := []events.Event{
		&events.GameStartedEvent{
			BaseEvent: events.BaseEvent{Type: events.EventGameStarted},
		},
		&events.TurnStartedEvent{
			BaseEvent:  events.BaseEvent{Type: events.EventTurnStarted},
			PlayerID:   state.Players[0].PlayerID,
			TurnNumber: 1,
		},
	}

	return state, evts, nil
}

func shuffleCards(cards []models.Card, seed []byte) {
	n := len(cards)
	if n <= 1 {
		return
	}
	seedVal := int64(0)
	for _, b := range seed {
		seedVal = seedVal*31 + int64(b)
	}
	rng := newSimpleRNG(seedVal)
	for i := n - 1; i > 0; i-- {
		j := rng.Intn(i + 1)
		cards[i], cards[j] = cards[j], cards[i]
	}
}

type simpleRNG struct {
	state uint64
}

func newSimpleRNG(seed int64) *simpleRNG {
	if seed == 0 {
		seed = 1
	}
	return &simpleRNG{state: uint64(seed)}
}

func (r *simpleRNG) Intn(n int) int {
	r.state ^= r.state << 13
	r.state ^= r.state >> 7
	r.state ^= r.state << 17
	return int(r.state % uint64(n))
}

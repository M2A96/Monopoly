package reducers

import (
	"fmt"

	"github.com/google/uuid"
	"github/M2A96/Monopoly.git/internal/game/commands"
	"github/M2A96/Monopoly.git/internal/game/events"
	"github/M2A96/Monopoly.git/internal/game/models"
	"github/M2A96/Monopoly.git/pkg/board"
)

func reduceRollDice(state models.GameState, cmd commands.Command) (models.GameState, []events.Event, error) {
	c, ok := cmd.(*commands.RollDiceCommand)
	if !ok {
		return state, nil, fmt.Errorf("expected RollDiceCommand, got %T", cmd)
	}

	if state.Status != models.GameStatusActive {
		return state, nil, fmt.Errorf("game not active")
	}

	playerIdx := findPlayerIndex(&state, c.PlayerID)
	if playerIdx == -1 {
		return state, nil, fmt.Errorf("player not in game")
	}

	player := &state.Players[playerIdx]
	if player.IsBankrupt {
		return state, nil, fmt.Errorf("bankrupt player cannot act")
	}

	if c.PlayerID != state.Turn.ActivePlayerID {
		return state, nil, fmt.Errorf("not your turn")
	}

	if state.Turn.Phase != models.TurnPhasePreRoll {
		return state, nil, fmt.Errorf("cannot roll in phase %s", state.Turn.Phase)
	}

	if c.Die1 < 1 || c.Die1 > 6 || c.Die2 < 1 || c.Die2 > 6 {
		return state, nil, fmt.Errorf("invalid dice values")
	}

	state = copyState(state)
	// Re-acquire the pointer: the one taken above points into the pre-copy slice.
	player = &state.Players[playerIdx]
	diceRoll := models.NewDiceRoll(c.Die1, c.Die2)
	state.Turn.DiceRoll = diceRoll
	state.Turn.Phase = models.TurnPhaseLand

	evts := []events.Event{
		&events.DiceRolledEvent{
			BaseEvent: events.BaseEvent{Type: events.EventDiceRolled},
			PlayerID:  c.PlayerID,
			Die1:      c.Die1,
			Die2:      c.Die2,
		},
	}

	if player.IsInJail {
		return reduceJailRoll(&state, player, c, evts)
	}

	isDoubles := diceRoll.IsDoubles()
	if isDoubles {
		state.Turn.DoublesCount++
	}

	if state.Turn.DoublesCount >= 3 {
		return reduceTripleDoubles(&state, player, c.PlayerID, evts)
	}

	oldPos := player.Position
	total := diceRoll.Total()
	newPos := (oldPos + total) % 40
	player.Position = newPos

	if newPos < oldPos && oldPos != 30 {
		player.Money += 200
		evts = append(evts, &events.PassedGoEvent{
			BaseEvent: events.BaseEvent{Type: events.EventPassedGo},
			PlayerID:  c.PlayerID,
			Amount:    200,
		})
	}

	evts = append(evts, &events.PlayerMovedEvent{
		BaseEvent:    events.BaseEvent{Type: events.EventPlayerMoved},
		PlayerID:     c.PlayerID,
		FromPosition: oldPos,
		ToPosition:   newPos,
	})

	landEvts := handleLanding(&state, c.PlayerID, newPos, diceRoll.Total())
	evts = append(evts, landEvts...)

	if state.Turn.Phase == models.TurnPhaseLand {
		state.Turn.Phase = phaseAfterResolution(&state)
	}

	return state, evts, nil
}

func reduceJailRoll(state *models.GameState, player *models.PlayerState, c *commands.RollDiceCommand, evts []events.Event) (models.GameState, []events.Event, error) {
	diceRoll := models.NewDiceRoll(c.Die1, c.Die2)

	if diceRoll.IsDoubles() {
		player.IsInJail = false
		player.JailState = models.JailState{}
		// Leaving jail on doubles does not grant an extra roll.
		state.Turn.DoublesCount = 0

		evts = append(evts, &events.ReleasedFromJailEvent{
			BaseEvent: events.BaseEvent{Type: events.EventReleasedFromJail},
			PlayerID:  c.PlayerID,
			Method:    "DOUBLES",
		})

		oldPos := player.Position
		total := diceRoll.Total()
		newPos := (oldPos + total) % 40
		player.Position = newPos

		if newPos < oldPos {
			player.Money += 200
			evts = append(evts, &events.PassedGoEvent{
				BaseEvent: events.BaseEvent{Type: events.EventPassedGo},
				PlayerID:  c.PlayerID,
				Amount:    200,
			})
		}

		evts = append(evts, &events.PlayerMovedEvent{
			BaseEvent:    events.BaseEvent{Type: events.EventPlayerMoved},
			PlayerID:     c.PlayerID,
			FromPosition: oldPos,
			ToPosition:   newPos,
		})

		landEvts := handleLanding(state, c.PlayerID, newPos, diceRoll.Total())
		evts = append(evts, landEvts...)

		if state.Turn.Phase == models.TurnPhaseLand {
			state.Turn.Phase = models.TurnPhaseEnd
		}

		return *state, evts, nil
	}

	player.JailState.TurnsInJail++
	state.Turn.Phase = models.TurnPhaseEnd

	if player.JailState.TurnsInJail >= 3 {
		player.Money -= 50
		player.IsInJail = false
		player.JailState = models.JailState{}

		total := diceRoll.Total()
		oldPos := player.Position
		newPos := (oldPos + total) % 40
		player.Position = newPos

		evts = append(evts, &events.ReleasedFromJailEvent{
			BaseEvent: events.BaseEvent{Type: events.EventReleasedFromJail},
			PlayerID:  c.PlayerID,
			Method:    "FORCE_PAY",
		})

		if newPos < oldPos {
			player.Money += 200
			evts = append(evts, &events.PassedGoEvent{
				BaseEvent: events.BaseEvent{Type: events.EventPassedGo},
				PlayerID:  c.PlayerID,
				Amount:    200,
			})
		}

		evts = append(evts, &events.PlayerMovedEvent{
			BaseEvent:    events.BaseEvent{Type: events.EventPlayerMoved},
			PlayerID:     c.PlayerID,
			FromPosition: oldPos,
			ToPosition:   newPos,
		})

		landEvts := handleLanding(state, c.PlayerID, newPos, diceRoll.Total())
		evts = append(evts, landEvts...)
	}

	return *state, evts, nil
}

func reduceTripleDoubles(state *models.GameState, player *models.PlayerState, playerID uuid.UUID, evts []events.Event) (models.GameState, []events.Event, error) {
	player.IsInJail = true
	player.JailState = models.JailState{TurnsInJail: 0}
	player.Position = 10
	state.Turn.DoublesCount = 0
	state.Turn.Phase = models.TurnPhaseEnd
	state.Turn.ActionWindow = nil

	evts = append(evts, &events.SentToJailEvent{
		BaseEvent: events.BaseEvent{Type: events.EventSentToJail},
		PlayerID:  playerID,
	})

	return *state, evts, nil
}

func handleLanding(state *models.GameState, playerID uuid.UUID, spaceIndex int, diceTotal int) []events.Event {
	var evts []events.Event
	space, ok := board.SpaceByIndex[spaceIndex]
	if !ok {
		return evts
	}

	player := findPlayer(state, playerID)
	if player == nil {
		return evts
	}

	switch space.Type {
	case board.SpaceTypeGoToJail:
		player.IsInJail = true
		player.JailState = models.JailState{TurnsInJail: 0}
		player.Position = 10
		state.Turn.Phase = models.TurnPhaseEnd
		state.Turn.ActionWindow = nil
		evts = append(evts, &events.SentToJailEvent{
			BaseEvent: events.BaseEvent{Type: events.EventSentToJail},
			PlayerID:  playerID,
		})

	case board.SpaceTypeTax:
		amount := space.TaxAmount
		player.Money -= amount
		evts = append(evts, &events.TaxPaidEvent{
			BaseEvent:  events.BaseEvent{Type: events.EventTaxPaid},
			PlayerID:   playerID,
			SpaceIndex: spaceIndex,
			Amount:     amount,
		})

	case board.SpaceTypeProperty, board.SpaceTypeRailroad, board.SpaceTypeUtility:
		o, owned := state.Ownerships[spaceIndex]
		if !owned {
			if player.Money >= space.Price {
				state.Turn.Phase = models.TurnPhaseAction
				state.Turn.ActionWindow = &models.ActionWindow{
					Type:       models.ActionWaitingForPurchaseDecision,
					PlayerID:   playerID,
					SpaceIndex: spaceIndex,
				}
			}
		} else if o.OwnerID != playerID && !o.Mortgaged {
			rent := calculateRent(spaceIndex, diceTotal, state, o.OwnerID)
			player.Money -= rent
			owner := findPlayer(state, o.OwnerID)
			if owner != nil {
				owner.Money += rent
			}
			evts = append(evts, &events.RentChargedEvent{
				BaseEvent:  events.BaseEvent{Type: events.EventRentCharged},
				PayerID:    playerID,
				OwnerID:    o.OwnerID,
				SpaceIndex: spaceIndex,
				Amount:     rent,
			})
		}

	case board.SpaceTypeChance:
		if state.CardDecks.ChanceIndex < len(state.CardDecks.Chance) {
			card := state.CardDecks.Chance[state.CardDecks.ChanceIndex]
			state.CardDecks.ChanceIndex++
			evts = append(evts, &events.ChanceCardDrawnEvent{
				BaseEvent: events.BaseEvent{Type: events.EventChanceCardDrawn},
				PlayerID:  playerID,
				CardID:    card.ID,
			})
			cardEvts := applyCard(state, playerID, card)
			evts = append(evts, cardEvts...)
		}

	case board.SpaceTypeCommunity:
		if state.CardDecks.CommunityIndex < len(state.CardDecks.CommunityChest) {
			card := state.CardDecks.CommunityChest[state.CardDecks.CommunityIndex]
			state.CardDecks.CommunityIndex++
			evts = append(evts, &events.CommunityChestCardDrawnEvent{
				BaseEvent: events.BaseEvent{Type: events.EventCommunityChestCardDrawn},
				PlayerID:  playerID,
				CardID:    card.ID,
			})
			cardEvts := applyCard(state, playerID, card)
			evts = append(evts, cardEvts...)
		}
	}

	return evts
}

func applyCard(state *models.GameState, playerID uuid.UUID, card models.Card) []events.Event {
	var evts []events.Event
	player := findPlayer(state, playerID)
	if player == nil {
		return evts
	}

	switch card.Effect {
	case models.CardEffectMoveToGo:
		player.Position = 0
		player.Money += 200
		evts = append(evts, &events.PassedGoEvent{
			BaseEvent: events.BaseEvent{Type: events.EventPassedGo},
			PlayerID:  playerID,
			Amount:    200,
		})
		landEvts := handleLanding(state, playerID, 0, 0)
		evts = append(evts, landEvts...)

	case models.CardEffectMoveToSpace:
		oldPos := player.Position
		newPos := card.SpaceIndex
		if newPos < oldPos {
			player.Money += 200
			evts = append(evts, &events.PassedGoEvent{
				BaseEvent: events.BaseEvent{Type: events.EventPassedGo},
				PlayerID:  playerID,
				Amount:    200,
			})
		}
		player.Position = newPos
		evts = append(evts, &events.PlayerMovedEvent{
			BaseEvent:    events.BaseEvent{Type: events.EventPlayerMoved},
			PlayerID:     playerID,
			FromPosition: oldPos,
			ToPosition:   newPos,
		})
		landEvts := handleLanding(state, playerID, newPos, 0)
		evts = append(evts, landEvts...)

	case models.CardEffectGoToJail:
		player.IsInJail = true
		player.JailState = models.JailState{TurnsInJail: 0}
		player.Position = 10
		state.Turn.Phase = models.TurnPhaseEnd
		state.Turn.ActionWindow = nil
		evts = append(evts, &events.SentToJailEvent{
			BaseEvent: events.BaseEvent{Type: events.EventSentToJail},
			PlayerID:  playerID,
		})

	case models.CardEffectCollect:
		player.Money += card.Amount

	case models.CardEffectPay:
		player.Money -= card.Amount

	case models.CardEffectGetOutOfJail:
		player.GetOutOfJailCards++

	case models.CardEffectGoBack3Spaces:
		newPos := player.Position - 3
		if newPos < 0 {
			newPos += 40
		}
		oldPos := player.Position
		player.Position = newPos
		evts = append(evts, &events.PlayerMovedEvent{
			BaseEvent:    events.BaseEvent{Type: events.EventPlayerMoved},
			PlayerID:     playerID,
			FromPosition: oldPos,
			ToPosition:   newPos,
		})
		landEvts := handleLanding(state, playerID, newPos, 0)
		evts = append(evts, landEvts...)

	case models.CardEffectRepairsPerBuilding:
		houses, hotels := sumBuildings(state, playerID)
		cost := houses*card.HouseCost + hotels*card.HotelCost
		player.Money -= cost

	case models.CardEffectCollectFromPlayers:
		for i := range state.Players {
			if state.Players[i].PlayerID != playerID && !state.Players[i].IsBankrupt {
				state.Players[i].Money -= card.Amount
				player.Money += card.Amount
			}
		}

	case models.CardEffectPayToPlayers:
		for i := range state.Players {
			if state.Players[i].PlayerID != playerID && !state.Players[i].IsBankrupt {
				state.Players[i].Money += card.Amount
				player.Money -= card.Amount
			}
		}

	case models.CardEffectMoveToNearestRailroad:
		nearest := findNearestRailroad(player.Position)
		oldPos := player.Position
		if nearest < oldPos {
			player.Money += 200
			evts = append(evts, &events.PassedGoEvent{
				BaseEvent: events.BaseEvent{Type: events.EventPassedGo},
				PlayerID:  playerID,
				Amount:    200,
			})
		}
		player.Position = nearest
		evts = append(evts, &events.PlayerMovedEvent{
			BaseEvent:    events.BaseEvent{Type: events.EventPlayerMoved},
			PlayerID:     playerID,
			FromPosition: oldPos,
			ToPosition:   nearest,
		})
		landEvts := handleLanding(state, playerID, nearest, 0)
		evts = append(evts, landEvts...)

	case models.CardEffectMoveToNearestUtility:
		nearest := findNearestUtility(player.Position)
		oldPos := player.Position
		if nearest < oldPos {
			player.Money += 200
			evts = append(evts, &events.PassedGoEvent{
				BaseEvent: events.BaseEvent{Type: events.EventPassedGo},
				PlayerID:  playerID,
				Amount:    200,
			})
		}
		player.Position = nearest
		evts = append(evts, &events.PlayerMovedEvent{
			BaseEvent:    events.BaseEvent{Type: events.EventPlayerMoved},
			PlayerID:     playerID,
			FromPosition: oldPos,
			ToPosition:   nearest,
		})
		landEvts := handleLanding(state, playerID, nearest, 0)
		evts = append(evts, landEvts...)
	}

	return evts
}

func findNearestRailroad(fromPos int) int {
	railroads := []int{5, 15, 25, 35}
	for _, r := range railroads {
		if r > fromPos {
			return r
		}
	}
	return railroads[0]
}

func findNearestUtility(fromPos int) int {
	if fromPos < 12 {
		return 12
	}
	if fromPos < 28 {
		return 28
	}
	return 12
}

// phaseAfterResolution returns the phase once the current landing/action is resolved:
// a player who rolled doubles (and is still free) rolls again, otherwise the turn may end.
func phaseAfterResolution(state *models.GameState) models.TurnPhase {
	player := findPlayer(state, state.Turn.ActivePlayerID)
	if player != nil && !player.IsInJail && !player.IsBankrupt &&
		state.Turn.DiceRoll != nil && state.Turn.DiceRoll.IsDoubles() && state.Turn.DoublesCount > 0 {
		return models.TurnPhasePreRoll
	}
	return models.TurnPhaseEnd
}

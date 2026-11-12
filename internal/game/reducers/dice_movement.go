package reducers

import (
	"fmt"

	"github.com/google/uuid"
	"github/M2A96/Monopoly.git/internal/game/commands"
	"github/M2A96/Monopoly.git/internal/game/events"
	"github/M2A96/Monopoly.git/internal/game/models"
	"github/M2A96/Monopoly.git/pkg/board"
)

const (
	jailPosition = 10
	goSalary     = 200
	jailFine     = 50
	boardSize    = 40
)

// rentModifier adjusts rent for card-driven moves.
type rentModifier struct {
	railroadMultiplier int // "nearest railroad" cards: pay twice the rent
	utilityFactor      int // "nearest utility" card: pay 10x dice regardless of utilities owned
}

var noModifier = rentModifier{}

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

	if state.Players[playerIdx].IsBankrupt {
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
	player := &state.Players[playerIdx]
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
		evts = append(evts, rollInJail(&state, player, diceRoll)...)
	} else {
		if diceRoll.IsDoubles() {
			state.Turn.DoublesCount++
		}
		if state.Turn.DoublesCount >= 3 {
			evts = append(evts, sendToJail(&state, player)...)
		} else {
			evts = append(evts, moveBy(&state, player, diceRoll.Total())...)
		}
	}

	settlePhase(&state)
	return state, evts, nil
}

// rollInJail handles a roll by a jailed player: doubles release them (no extra roll);
// the third failed attempt forces the $50 fine and the player moves anyway.
func rollInJail(state *models.GameState, player *models.PlayerState, roll models.DiceRoll) []events.Event {
	if roll.IsDoubles() {
		release(player)
		state.Turn.DoublesCount = 0
		evts := []events.Event{releasedEvent(player.PlayerID, "DOUBLES")}
		return append(evts, moveBy(state, player, roll.Total())...)
	}

	player.JailState.TurnsInJail++
	if player.JailState.TurnsInJail < 3 {
		state.Turn.Phase = models.TurnPhaseEnd
		return nil
	}

	release(player)
	evts := []events.Event{releasedEvent(player.PlayerID, "FORCE_PAY")}
	evts = append(evts, charge(state, player.PlayerID, bank, jailFine, "JAIL_FINE")...)
	return append(evts, moveBy(state, player, roll.Total())...)
}

func release(player *models.PlayerState) {
	player.IsInJail = false
	player.JailState = models.JailState{}
}

func releasedEvent(playerID uuid.UUID, method string) events.Event {
	return &events.ReleasedFromJailEvent{
		BaseEvent: events.BaseEvent{Type: events.EventReleasedFromJail},
		PlayerID:  playerID,
		Method:    method,
	}
}

func sendToJail(state *models.GameState, player *models.PlayerState) []events.Event {
	player.IsInJail = true
	player.JailState = models.JailState{}
	player.Position = jailPosition
	state.Turn.DoublesCount = 0
	state.Turn.Phase = models.TurnPhaseEnd
	state.Turn.ActionWindow = nil
	return []events.Event{&events.SentToJailEvent{
		BaseEvent: events.BaseEvent{Type: events.EventSentToJail},
		PlayerID:  player.PlayerID,
	}}
}

// moveBy advances the player by steps, collecting GO salary when wrapping, then lands.
func moveBy(state *models.GameState, player *models.PlayerState, steps int) []events.Event {
	oldPos := player.Position
	newPos := (oldPos + steps) % boardSize
	return moveTo(state, player, newPos, newPos < oldPos, noModifier)
}

// moveTo places the player on newPos (forward movement), optionally paying GO salary.
func moveTo(state *models.GameState, player *models.PlayerState, newPos int, passedGo bool, mod rentModifier) []events.Event {
	var evts []events.Event
	oldPos := player.Position
	if passedGo {
		player.Money += goSalary
		evts = append(evts, &events.PassedGoEvent{
			BaseEvent: events.BaseEvent{Type: events.EventPassedGo},
			PlayerID:  player.PlayerID,
			Amount:    goSalary,
		})
	}
	player.Position = newPos
	evts = append(evts, &events.PlayerMovedEvent{
		BaseEvent:    events.BaseEvent{Type: events.EventPlayerMoved},
		PlayerID:     player.PlayerID,
		FromPosition: oldPos,
		ToPosition:   newPos,
	})
	return append(evts, land(state, player, newPos, mod)...)
}

func diceTotal(state *models.GameState) int {
	if state.Turn.DiceRoll == nil {
		return 0
	}
	return state.Turn.DiceRoll.Total()
}

func land(state *models.GameState, player *models.PlayerState, spaceIndex int, mod rentModifier) []events.Event {
	space, ok := board.SpaceByIndex[spaceIndex]
	if !ok {
		return nil
	}

	switch space.Type {
	case board.SpaceTypeGoToJail:
		return sendToJail(state, player)

	case board.SpaceTypeTax:
		evts := []events.Event{&events.TaxPaidEvent{
			BaseEvent:  events.BaseEvent{Type: events.EventTaxPaid},
			PlayerID:   player.PlayerID,
			SpaceIndex: spaceIndex,
			Amount:     space.TaxAmount,
		}}
		return append(evts, charge(state, player.PlayerID, bank, space.TaxAmount, "TAX")...)

	case board.SpaceTypeProperty, board.SpaceTypeRailroad, board.SpaceTypeUtility:
		o, owned := state.Ownerships[spaceIndex]
		if !owned {
			// Always offered; a player who cannot afford it declines and it goes to auction.
			state.Turn.Phase = models.TurnPhaseAction
			state.Turn.ActionWindow = &models.ActionWindow{
				Type:       models.ActionWaitingForPurchaseDecision,
				PlayerID:   player.PlayerID,
				SpaceIndex: spaceIndex,
			}
			return nil
		}
		if o.OwnerID == player.PlayerID || o.Mortgaged {
			return nil
		}
		rent := calculateRent(spaceIndex, diceTotal(state), state, o.OwnerID)
		switch {
		case space.Type == board.SpaceTypeRailroad && mod.railroadMultiplier > 1:
			rent *= mod.railroadMultiplier
		case space.Type == board.SpaceTypeUtility && mod.utilityFactor > 0:
			rent = diceTotal(state) * mod.utilityFactor
		}
		evts := []events.Event{&events.RentChargedEvent{
			BaseEvent:  events.BaseEvent{Type: events.EventRentCharged},
			PayerID:    player.PlayerID,
			OwnerID:    o.OwnerID,
			SpaceIndex: spaceIndex,
			Amount:     rent,
		}}
		return append(evts, charge(state, player.PlayerID, o.OwnerID, rent, "RENT")...)

	case board.SpaceTypeChance:
		card, ok := drawCard(state.CardDecks.Chance, &state.CardDecks.ChanceIndex, state.CardDecks.JailCardHolders)
		if !ok {
			return nil
		}
		evts := []events.Event{&events.ChanceCardDrawnEvent{
			BaseEvent: events.BaseEvent{Type: events.EventChanceCardDrawn},
			PlayerID:  player.PlayerID,
			CardID:    card.ID,
		}}
		return append(evts, applyCard(state, player, card)...)

	case board.SpaceTypeCommunity:
		card, ok := drawCard(state.CardDecks.CommunityChest, &state.CardDecks.CommunityIndex, state.CardDecks.JailCardHolders)
		if !ok {
			return nil
		}
		evts := []events.Event{&events.CommunityChestCardDrawnEvent{
			BaseEvent: events.BaseEvent{Type: events.EventCommunityChestCardDrawn},
			PlayerID:  player.PlayerID,
			CardID:    card.ID,
		}}
		return append(evts, applyCard(state, player, card)...)
	}

	return nil
}

func applyCard(state *models.GameState, player *models.PlayerState, card models.Card) []events.Event {
	pid := player.PlayerID
	reason := "CARD_" + card.ID

	switch card.Effect {
	case models.CardEffectMoveToGo:
		return moveTo(state, player, 0, true, noModifier)

	case models.CardEffectMoveToSpace:
		return moveTo(state, player, card.SpaceIndex, card.SpaceIndex < player.Position, noModifier)

	case models.CardEffectGoToJail:
		return sendToJail(state, player)

	case models.CardEffectCollect:
		player.Money += card.Amount

	case models.CardEffectPay:
		return charge(state, pid, bank, card.Amount, reason)

	case models.CardEffectGetOutOfJail:
		if state.CardDecks.JailCardHolders == nil {
			state.CardDecks.JailCardHolders = make(map[string]uuid.UUID)
		}
		state.CardDecks.JailCardHolders[card.ID] = pid
		player.GetOutOfJailCards++

	case models.CardEffectGoBack3Spaces:
		newPos := (player.Position - 3 + boardSize) % boardSize
		oldPos := player.Position
		player.Position = newPos
		evts := []events.Event{&events.PlayerMovedEvent{
			BaseEvent:    events.BaseEvent{Type: events.EventPlayerMoved},
			PlayerID:     pid,
			FromPosition: oldPos,
			ToPosition:   newPos,
		}}
		return append(evts, land(state, player, newPos, noModifier)...)

	case models.CardEffectRepairsPerBuilding:
		houses, hotels := sumBuildings(state, pid)
		return charge(state, pid, bank, houses*card.HouseCost+hotels*card.HotelCost, reason)

	case models.CardEffectCollectFromPlayers:
		var evts []events.Event
		for _, other := range otherActivePlayers(state, pid) {
			evts = append(evts, charge(state, other, pid, card.Amount, reason)...)
		}
		return evts

	case models.CardEffectPayToPlayers:
		var evts []events.Event
		for _, other := range otherActivePlayers(state, pid) {
			evts = append(evts, charge(state, pid, other, card.Amount, reason)...)
		}
		return evts

	case models.CardEffectMoveToNearestRailroad:
		nearest := findNearestRailroad(player.Position)
		mod := rentModifier{railroadMultiplier: 1}
		if card.DoubleTax {
			mod.railroadMultiplier = 2
		}
		return moveTo(state, player, nearest, nearest < player.Position, mod)

	case models.CardEffectMoveToNearestUtility:
		nearest := findNearestUtility(player.Position)
		return moveTo(state, player, nearest, nearest < player.Position, rentModifier{utilityFactor: 10})
	}

	return nil
}

func otherActivePlayers(state *models.GameState, except uuid.UUID) []uuid.UUID {
	var ids []uuid.UUID
	for _, p := range state.Players {
		if p.PlayerID != except && !p.IsBankrupt {
			ids = append(ids, p.PlayerID)
		}
	}
	return ids
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
	if state.Turn.DiceRoll == nil {
		// The active player has not rolled yet (e.g. the turn passed after a bankruptcy).
		return models.TurnPhasePreRoll
	}
	player := findPlayer(state, state.Turn.ActivePlayerID)
	if player != nil && !player.IsInJail && !player.IsBankrupt &&
		state.Turn.DiceRoll.IsDoubles() && state.Turn.DoublesCount > 0 {
		return models.TurnPhasePreRoll
	}
	return models.TurnPhaseEnd
}

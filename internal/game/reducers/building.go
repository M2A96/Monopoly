package reducers

import (
	"fmt"

	"github.com/google/uuid"
	"github/M2A96/Monopoly.git/internal/game/commands"
	"github/M2A96/Monopoly.git/internal/game/events"
	"github/M2A96/Monopoly.git/internal/game/models"
	"github/M2A96/Monopoly.git/pkg/board"
)

func reduceBuyProperty(state models.GameState, cmd commands.Command) (models.GameState, []events.Event, error) {
	c, ok := cmd.(*commands.BuyPropertyCommand)
	if !ok {
		return state, nil, fmt.Errorf("expected BuyPropertyCommand, got %T", cmd)
	}

	if state.Status != models.GameStatusActive {
		return state, nil, fmt.Errorf("game not active")
	}

	if state.Turn.Phase != models.TurnPhaseAction || state.Turn.ActionWindow == nil {
		return state, nil, fmt.Errorf("not in action phase with pending action")
	}

	if state.Turn.ActionWindow.Type != models.ActionWaitingForPurchaseDecision {
		return state, nil, fmt.Errorf("no purchase decision pending")
	}

	if state.Turn.ActionWindow.PlayerID != c.PlayerID {
		return state, nil, fmt.Errorf("not your purchase decision")
	}

	spaceIndex := state.Turn.ActionWindow.SpaceIndex
	space, ok := board.SpaceByIndex[spaceIndex]
	if !ok {
		return state, nil, fmt.Errorf("invalid space index")
	}

	player := findPlayer(&state, c.PlayerID)
	if player == nil {
		return state, nil, fmt.Errorf("player not in game")
	}

	if player.Money < space.Price {
		return state, nil, fmt.Errorf("insufficient funds: have %d, need %d", player.Money, space.Price)
	}

	if _, owned := state.Ownerships[spaceIndex]; owned {
		return state, nil, fmt.Errorf("property already owned")
	}

	state = copyState(state)
	player = findPlayer(&state, c.PlayerID)

	player.Money -= space.Price
	state.Ownerships[spaceIndex] = models.Ownership{
		SpaceIndex: spaceIndex,
		OwnerID:    c.PlayerID,
		Mortgaged:  false,
		Buildings:  0,
	}
	addSpaceIndex(player, spaceIndex)

	state.Turn.ActionWindow = nil
	state.Turn.Phase = phaseAfterResolution(&state)

	return state, []events.Event{
		&events.PropertyPurchasedEvent{
			BaseEvent:  events.BaseEvent{Type: events.EventPropertyPurchased},
			PlayerID:   c.PlayerID,
			SpaceIndex: spaceIndex,
			Price:      space.Price,
		},
	}, nil
}

func reduceDeclinePurchase(state models.GameState, cmd commands.Command) (models.GameState, []events.Event, error) {
	c, ok := cmd.(*commands.DeclinePurchaseCommand)
	if !ok {
		return state, nil, fmt.Errorf("expected DeclinePurchaseCommand, got %T", cmd)
	}

	if state.Status != models.GameStatusActive {
		return state, nil, fmt.Errorf("game not active")
	}

	if state.Turn.Phase != models.TurnPhaseAction || state.Turn.ActionWindow == nil {
		return state, nil, fmt.Errorf("not in action phase with pending action")
	}

	if state.Turn.ActionWindow.Type != models.ActionWaitingForPurchaseDecision {
		return state, nil, fmt.Errorf("no purchase decision pending")
	}

	if state.Turn.ActionWindow.PlayerID != c.PlayerID {
		return state, nil, fmt.Errorf("not your purchase decision")
	}

	spaceIndex := state.Turn.ActionWindow.SpaceIndex

	state = copyState(state)
	state.Auction = &models.Auction{
		SpaceIndex: spaceIndex,
		Bids:       make(map[uuid.UUID]int),
	}
	state.Auction.Bids[c.PlayerID] = 0
	state.Turn.ActionWindow = &models.ActionWindow{
		Type:       models.ActionWaitingForAuctionBid,
		PlayerID:   c.PlayerID,
		SpaceIndex: spaceIndex,
	}

	return state, []events.Event{
		&events.AuctionStartedEvent{
			BaseEvent:  events.BaseEvent{Type: events.EventAuctionStarted},
			SpaceIndex: spaceIndex,
		},
	}, nil
}

func reduceBuildHouse(state models.GameState, cmd commands.Command) (models.GameState, []events.Event, error) {
	c, ok := cmd.(*commands.BuildHouseCommand)
	if !ok {
		return state, nil, fmt.Errorf("expected BuildHouseCommand, got %T", cmd)
	}

	if state.Status != models.GameStatusActive {
		return state, nil, fmt.Errorf("game not active")
	}

	space, ok := board.SpaceByIndex[c.SpaceIndex]
	if !ok || space.Type != board.SpaceTypeProperty {
		return state, nil, fmt.Errorf("can only build on properties")
	}

	o, owned := state.Ownerships[c.SpaceIndex]
	if !owned || o.OwnerID != c.PlayerID {
		return state, nil, fmt.Errorf("not your property")
	}

	if o.Mortgaged {
		return state, nil, fmt.Errorf("property is mortgaged")
	}

	if !hasMonopoly(&state, c.PlayerID, c.SpaceIndex) {
		return state, nil, fmt.Errorf("need full color group to build")
	}

	if o.Buildings >= 4 {
		return state, nil, fmt.Errorf("max houses reached (use build hotel)")
	}

	group := colorGroupSpaces(space.ColorGroup)
	minBuildings := allSameBuildings(&state, group)
	if minBuildings > 0 && o.Buildings > minBuildings {
		return state, nil, fmt.Errorf("even building rule: must build equally across group")
	}

	if state.Bank.Houses <= 0 {
		return state, nil, fmt.Errorf("no houses available in bank")
	}

	player := findPlayer(&state, c.PlayerID)
	if player == nil {
		return state, nil, fmt.Errorf("player not in game")
	}

	if player.Money < space.HouseCost {
		return state, nil, fmt.Errorf("insufficient funds: have %d, need %d", player.Money, space.HouseCost)
	}

	state = copyState(state)
	player = findPlayer(&state, c.PlayerID)
	o = state.Ownerships[c.SpaceIndex]

	player.Money -= space.HouseCost
	o.Buildings++
	state.Ownerships[c.SpaceIndex] = o
	state.Bank.Houses--

	return state, []events.Event{
		&events.HouseBuiltEvent{
			BaseEvent:  events.BaseEvent{Type: events.EventHouseBuilt},
			PlayerID:   c.PlayerID,
			SpaceIndex: c.SpaceIndex,
			Cost:       space.HouseCost,
		},
	}, nil
}

func reduceBuildHotel(state models.GameState, cmd commands.Command) (models.GameState, []events.Event, error) {
	c, ok := cmd.(*commands.BuildHotelCommand)
	if !ok {
		return state, nil, fmt.Errorf("expected BuildHotelCommand, got %T", cmd)
	}

	if state.Status != models.GameStatusActive {
		return state, nil, fmt.Errorf("game not active")
	}

	space, ok := board.SpaceByIndex[c.SpaceIndex]
	if !ok || space.Type != board.SpaceTypeProperty {
		return state, nil, fmt.Errorf("can only build on properties")
	}

	o, owned := state.Ownerships[c.SpaceIndex]
	if !owned || o.OwnerID != c.PlayerID {
		return state, nil, fmt.Errorf("not your property")
	}

	if o.Mortgaged {
		return state, nil, fmt.Errorf("property is mortgaged")
	}

	if !hasMonopoly(&state, c.PlayerID, c.SpaceIndex) {
		return state, nil, fmt.Errorf("need full color group to build")
	}

	if o.Buildings != 4 {
		return state, nil, fmt.Errorf("need 4 houses before building hotel")
	}

	if state.Bank.Hotels <= 0 {
		return state, nil, fmt.Errorf("no hotels available in bank")
	}

	player := findPlayer(&state, c.PlayerID)
	if player == nil {
		return state, nil, fmt.Errorf("player not in game")
	}

	if player.Money < space.HouseCost {
		return state, nil, fmt.Errorf("insufficient funds: have %d, need %d", player.Money, space.HouseCost)
	}

	state = copyState(state)
	player = findPlayer(&state, c.PlayerID)
	o = state.Ownerships[c.SpaceIndex]

	player.Money -= space.HouseCost
	o.Buildings = models.HotelLevel
	state.Ownerships[c.SpaceIndex] = o
	state.Bank.Houses += 4
	state.Bank.Hotels--

	return state, []events.Event{
		&events.HotelBuiltEvent{
			BaseEvent:  events.BaseEvent{Type: events.EventHotelBuilt},
			PlayerID:   c.PlayerID,
			SpaceIndex: c.SpaceIndex,
			Cost:       space.HouseCost,
		},
	}, nil
}

func reduceSellHouse(state models.GameState, cmd commands.Command) (models.GameState, []events.Event, error) {
	c, ok := cmd.(*commands.SellHouseCommand)
	if !ok {
		return state, nil, fmt.Errorf("expected SellHouseCommand, got %T", cmd)
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

	if o.Buildings <= 0 || o.Buildings >= models.HotelLevel {
		return state, nil, fmt.Errorf("no houses to sell")
	}

	state = copyState(state)
	player := findPlayer(&state, c.PlayerID)
	o = state.Ownerships[c.SpaceIndex]

	proceeds := space.HouseCost / 2
	player.Money += proceeds
	o.Buildings--
	state.Ownerships[c.SpaceIndex] = o
	state.Bank.Houses++

	return state, []events.Event{
		&events.HouseSoldEvent{
			BaseEvent:  events.BaseEvent{Type: events.EventHouseSold},
			PlayerID:   c.PlayerID,
			SpaceIndex: c.SpaceIndex,
			Proceeds:   proceeds,
		},
	}, nil
}

func reduceSellHotel(state models.GameState, cmd commands.Command) (models.GameState, []events.Event, error) {
	c, ok := cmd.(*commands.SellHotelCommand)
	if !ok {
		return state, nil, fmt.Errorf("expected SellHotelCommand, got %T", cmd)
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

	if o.Buildings != models.HotelLevel {
		return state, nil, fmt.Errorf("no hotel to sell")
	}

	state = copyState(state)
	player := findPlayer(&state, c.PlayerID)
	o = state.Ownerships[c.SpaceIndex]

	proceeds := space.HouseCost / 2
	player.Money += proceeds
	o.Buildings = 4
	state.Ownerships[c.SpaceIndex] = o
	state.Bank.Houses -= 4
	state.Bank.Hotels++

	return state, []events.Event{
		&events.HotelSoldEvent{
			BaseEvent:  events.BaseEvent{Type: events.EventHotelSold},
			PlayerID:   c.PlayerID,
			SpaceIndex: c.SpaceIndex,
			Proceeds:   proceeds,
		},
	}, nil
}

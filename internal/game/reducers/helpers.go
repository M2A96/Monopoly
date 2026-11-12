package reducers

import (
	"github.com/google/uuid"
	"github/M2A96/Monopoly.git/internal/game/models"
	"github/M2A96/Monopoly.git/pkg/board"
)

func findPlayer(state *models.GameState, id uuid.UUID) *models.PlayerState {
	for i := range state.Players {
		if state.Players[i].PlayerID == id {
			return &state.Players[i]
		}
	}
	return nil
}

func findPlayerIndex(state *models.GameState, id uuid.UUID) int {
	for i := range state.Players {
		if state.Players[i].PlayerID == id {
			return i
		}
	}
	return -1
}

func isOwner(state *models.GameState, playerID uuid.UUID, spaceIndex int) bool {
	o, ok := state.Ownerships[spaceIndex]
	return ok && o.OwnerID == playerID && !o.Mortgaged
}

func hasMonopoly(state *models.GameState, playerID uuid.UUID, spaceIndex int) bool {
	space, ok := board.SpaceByIndex[spaceIndex]
	if !ok || space.ColorGroup == "" {
		return false
	}
	for _, s := range board.Spaces {
		if s.ColorGroup == space.ColorGroup && s.Type == board.SpaceTypeProperty {
			o, ok := state.Ownerships[s.Index]
			if !ok || o.OwnerID != playerID {
				return false
			}
		}
	}
	return true
}

func colorGroupSpaces(colorGroup board.ColorGroup) []int {
	var indices []int
	for _, s := range board.Spaces {
		if s.ColorGroup == colorGroup && s.Type == board.SpaceTypeProperty {
			indices = append(indices, s.Index)
		}
	}
	return indices
}

func countOwnedRailroads(state *models.GameState, playerID uuid.UUID) int {
	count := 0
	for _, s := range board.Spaces {
		if s.Type == board.SpaceTypeRailroad {
			o, ok := state.Ownerships[s.Index]
			if ok && o.OwnerID == playerID {
				count++
			}
		}
	}
	return count
}

func countOwnedUtilities(state *models.GameState, playerID uuid.UUID) int {
	count := 0
	for _, s := range board.Spaces {
		if s.Type == board.SpaceTypeUtility {
			o, ok := state.Ownerships[s.Index]
			if ok && o.OwnerID == playerID {
				count++
			}
		}
	}
	return count
}

func calculateRent(spaceIndex int, diceTotal int, state *models.GameState, ownerID uuid.UUID) int {
	space, ok := board.SpaceByIndex[spaceIndex]
	if !ok {
		return 0
	}

	switch space.Type {
	case board.SpaceTypeProperty:
		o, ok := state.Ownerships[spaceIndex]
		if !ok || o.Mortgaged {
			return 0
		}
		if o.Buildings > 0 {
			return space.Rent[o.Buildings]
		}
		if hasMonopoly(state, ownerID, spaceIndex) {
			return space.Rent[0] * 2
		}
		return space.Rent[0]

	case board.SpaceTypeRailroad:
		count := countOwnedRailroads(state, ownerID)
		if count == 0 {
			return 0
		}
		return space.Rent[count-1]

	case board.SpaceTypeUtility:
		count := countOwnedUtilities(state, ownerID)
		if count == 0 {
			return 0
		}
		if count == 1 {
			return diceTotal * 4
		}
		return diceTotal * 10
	}

	return 0
}

func copyState(state models.GameState) models.GameState {
	cp := state
	cp.Players = make([]models.PlayerState, len(state.Players))
	for i := range state.Players {
		cp.Players[i] = state.Players[i]
		cp.Players[i].OwnedSpaceIndexes = make([]int, len(state.Players[i].OwnedSpaceIndexes))
		copy(cp.Players[i].OwnedSpaceIndexes, state.Players[i].OwnedSpaceIndexes)
		cp.Players[i].PendingActions = make([]models.ActionWindow, len(state.Players[i].PendingActions))
		copy(cp.Players[i].PendingActions, state.Players[i].PendingActions)
	}
	cp.Ownerships = make(map[int]models.Ownership, len(state.Ownerships))
	for k, v := range state.Ownerships {
		cp.Ownerships[k] = v
	}
	cp.Trades = make([]models.Trade, len(state.Trades))
	copy(cp.Trades, state.Trades)
	if state.Debts != nil {
		cp.Debts = make([]models.Debt, len(state.Debts))
		copy(cp.Debts, state.Debts)
	}
	if state.Auction != nil {
		a := *state.Auction
		a.Bids = make(map[uuid.UUID]int, len(state.Auction.Bids))
		for k, v := range state.Auction.Bids {
			a.Bids[k] = v
		}
		a.Participants = append([]uuid.UUID(nil), state.Auction.Participants...)
		cp.Auction = &a
	}
	if state.Turn.ActionWindow != nil {
		w := *state.Turn.ActionWindow
		cp.Turn.ActionWindow = &w
	}
	if state.CardDecks.JailCardHolders != nil {
		cp.CardDecks.JailCardHolders = make(map[string]uuid.UUID, len(state.CardDecks.JailCardHolders))
		for k, v := range state.CardDecks.JailCardHolders {
			cp.CardDecks.JailCardHolders[k] = v
		}
	}
	cp.CardDecks.Chance = make([]models.Card, len(state.CardDecks.Chance))
	copy(cp.CardDecks.Chance, state.CardDecks.Chance)
	cp.CardDecks.CommunityChest = make([]models.Card, len(state.CardDecks.CommunityChest))
	copy(cp.CardDecks.CommunityChest, state.CardDecks.CommunityChest)
	return cp
}

func activePlayers(state *models.GameState) int {
	count := 0
	for _, p := range state.Players {
		if !p.IsBankrupt {
			count++
		}
	}
	return count
}

func nextActivePlayer(state *models.GameState, after uuid.UUID) uuid.UUID {
	started := false
	for _, p := range state.Players {
		if started && !p.IsBankrupt {
			return p.PlayerID
		}
		if p.PlayerID == after {
			started = true
		}
	}
	for _, p := range state.Players {
		if !p.IsBankrupt {
			return p.PlayerID
		}
	}
	return uuid.Nil
}

func addSpaceIndex(player *models.PlayerState, idx int) {
	for _, i := range player.OwnedSpaceIndexes {
		if i == idx {
			return
		}
	}
	player.OwnedSpaceIndexes = append(player.OwnedSpaceIndexes, idx)
}

func removeSpaceIndex(player *models.PlayerState, idx int) {
	for i, v := range player.OwnedSpaceIndexes {
		if v == idx {
			player.OwnedSpaceIndexes = append(player.OwnedSpaceIndexes[:i], player.OwnedSpaceIndexes[i+1:]...)
			return
		}
	}
}

func sumBuildings(state *models.GameState, playerID uuid.UUID) (houses, hotels int) {
	for _, o := range state.Ownerships {
		if o.OwnerID == playerID {
			if o.Buildings == models.HotelLevel {
				hotels++
			} else if o.Buildings > 0 {
				houses += o.Buildings
			}
		}
	}
	return
}

func allSameBuildings(state *models.GameState, group []int) int {
	if len(group) == 0 {
		return -1
	}
	min := -1
	for _, idx := range group {
		o, ok := state.Ownerships[idx]
		if !ok {
			return 0
		}
		if min == -1 || o.Buildings < min {
			min = o.Buildings
		}
	}
	return min
}

func maxBuildings(state *models.GameState, group []int) int {
	max := 0
	for _, idx := range group {
		if o, ok := state.Ownerships[idx]; ok && o.Buildings > max {
			max = o.Buildings
		}
	}
	return max
}

func groupHasMortgage(state *models.GameState, group []int) bool {
	for _, idx := range group {
		if o, ok := state.Ownerships[idx]; ok && o.Mortgaged {
			return true
		}
	}
	return false
}

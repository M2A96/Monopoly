package models

type CardType string
type CardEffectType string

const (
	CardTypeChance        CardType = "CHANCE"
	CardTypeCommunityChest CardType = "COMMUNITY_CHEST"
)

const (
	CardEffectCollect               CardEffectType = "COLLECT"
	CardEffectPay                   CardEffectType = "PAY"
	CardEffectMoveToGo              CardEffectType = "MOVE_TO_GO"
	CardEffectMoveToSpace           CardEffectType = "MOVE_TO_SPACE"
	CardEffectGoToJail              CardEffectType = "GO_TO_JAIL"
	CardEffectGetOutOfJail          CardEffectType = "GET_OUT_OF_JAIL"
	CardEffectMoveToNearestRailroad CardEffectType = "MOVE_TO_NEAREST_RAILROAD"
	CardEffectMoveToNearestUtility  CardEffectType = "MOVE_TO_NEAREST_UTILITY"
	CardEffectGoBack3Spaces         CardEffectType = "GO_BACK_3_SPACES"
	CardEffectRepairsPerBuilding    CardEffectType = "REPAIRS_PER_BUILDING"
	CardEffectCollectFromPlayers    CardEffectType = "COLLECT_FROM_PLAYERS"
	CardEffectPayToPlayers          CardEffectType = "PAY_TO_PLAYERS"
)

// Card is the static definition of a Chance or Community Chest card.
type Card struct {
	ID          string
	Type        CardType
	Description string
	Effect      CardEffectType
	Amount      int    // money collected/paid (COLLECT, PAY, MOVE_TO_GO, COLLECT_FROM_PLAYERS, PAY_TO_PLAYERS)
	HouseCost   int    // per-house charge (REPAIRS_PER_BUILDING)
	HotelCost   int    // per-hotel charge (REPAIRS_PER_BUILDING)
	SpaceIndex  int    // destination (MOVE_TO_SPACE)
	DoubleTax   bool   // pay double rent (MOVE_TO_NEAREST_RAILROAD)
}

// CardDecks holds the runtime state of both draw piles.
// Cards are shuffled once at game start; ChanceIndex/CommunityIndex advance on each draw.
type CardDecks struct {
	Chance         []Card
	CommunityChest []Card
	ChanceIndex    int
	CommunityIndex int
}

// DefaultChanceDeck returns the standard 16-card Chance deck in canonical order.
// The runtime must shuffle this before storing it in CardDecks.
func DefaultChanceDeck() []Card {
	return []Card{
		{ID: "CH01", Type: CardTypeChance, Description: "Advance to Go. Collect $200.", Effect: CardEffectMoveToGo, Amount: 200},
		{ID: "CH02", Type: CardTypeChance, Description: "Advance to Illinois Avenue.", Effect: CardEffectMoveToSpace, SpaceIndex: 24},
		{ID: "CH03", Type: CardTypeChance, Description: "Advance to St. Charles Place.", Effect: CardEffectMoveToSpace, SpaceIndex: 11},
		{ID: "CH04", Type: CardTypeChance, Description: "Advance token to nearest Railroad. If unowned buy it; if owned pay owner twice the rental.", Effect: CardEffectMoveToNearestRailroad, DoubleTax: true},
		{ID: "CH05", Type: CardTypeChance, Description: "Advance token to nearest Railroad. If unowned buy it; if owned pay owner twice the rental.", Effect: CardEffectMoveToNearestRailroad, DoubleTax: true},
		{ID: "CH06", Type: CardTypeChance, Description: "Advance token to nearest Utility. If unowned buy it; if owned throw dice and pay owner 10 times the amount shown.", Effect: CardEffectMoveToNearestUtility},
		{ID: "CH07", Type: CardTypeChance, Description: "Bank pays you dividend of $50.", Effect: CardEffectCollect, Amount: 50},
		{ID: "CH08", Type: CardTypeChance, Description: "Get out of Jail Free.", Effect: CardEffectGetOutOfJail},
		{ID: "CH09", Type: CardTypeChance, Description: "Go Back Three Spaces.", Effect: CardEffectGoBack3Spaces},
		{ID: "CH10", Type: CardTypeChance, Description: "Go to Jail. Go directly to Jail.", Effect: CardEffectGoToJail},
		{ID: "CH11", Type: CardTypeChance, Description: "Make general repairs on all your property. $25 per house, $100 per hotel.", Effect: CardEffectRepairsPerBuilding, HouseCost: 25, HotelCost: 100},
		{ID: "CH12", Type: CardTypeChance, Description: "Pay poor tax of $15.", Effect: CardEffectPay, Amount: 15},
		{ID: "CH13", Type: CardTypeChance, Description: "Take a trip to Reading Railroad.", Effect: CardEffectMoveToSpace, SpaceIndex: 5},
		{ID: "CH14", Type: CardTypeChance, Description: "Take a walk on the Boardwalk.", Effect: CardEffectMoveToSpace, SpaceIndex: 39},
		{ID: "CH15", Type: CardTypeChance, Description: "You have been elected Chairman of the Board. Pay each player $50.", Effect: CardEffectPayToPlayers, Amount: 50},
		{ID: "CH16", Type: CardTypeChance, Description: "Your building and loan matures. Collect $150.", Effect: CardEffectCollect, Amount: 150},
	}
}

// DefaultCommunityChestDeck returns the standard 16-card Community Chest deck in canonical order.
// The runtime must shuffle this before storing it in CardDecks.
func DefaultCommunityChestDeck() []Card {
	return []Card{
		{ID: "CC01", Type: CardTypeCommunityChest, Description: "Advance to Go. Collect $200.", Effect: CardEffectMoveToGo, Amount: 200},
		{ID: "CC02", Type: CardTypeCommunityChest, Description: "Bank error in your favor. Collect $200.", Effect: CardEffectCollect, Amount: 200},
		{ID: "CC03", Type: CardTypeCommunityChest, Description: "Doctor's fees. Pay $50.", Effect: CardEffectPay, Amount: 50},
		{ID: "CC04", Type: CardTypeCommunityChest, Description: "From sale of stock you get $50.", Effect: CardEffectCollect, Amount: 50},
		{ID: "CC05", Type: CardTypeCommunityChest, Description: "Get out of Jail Free.", Effect: CardEffectGetOutOfJail},
		{ID: "CC06", Type: CardTypeCommunityChest, Description: "Go to Jail. Go directly to Jail.", Effect: CardEffectGoToJail},
		{ID: "CC07", Type: CardTypeCommunityChest, Description: "Grand Opera Opening. Collect $50 from every player.", Effect: CardEffectCollectFromPlayers, Amount: 50},
		{ID: "CC08", Type: CardTypeCommunityChest, Description: "Holiday Fund matures. Receive $100.", Effect: CardEffectCollect, Amount: 100},
		{ID: "CC09", Type: CardTypeCommunityChest, Description: "Income tax refund. Collect $20.", Effect: CardEffectCollect, Amount: 20},
		{ID: "CC10", Type: CardTypeCommunityChest, Description: "It is your birthday. Collect $10 from every player.", Effect: CardEffectCollectFromPlayers, Amount: 10},
		{ID: "CC11", Type: CardTypeCommunityChest, Description: "Life insurance matures. Collect $100.", Effect: CardEffectCollect, Amount: 100},
		{ID: "CC12", Type: CardTypeCommunityChest, Description: "Pay hospital fees of $100.", Effect: CardEffectPay, Amount: 100},
		{ID: "CC13", Type: CardTypeCommunityChest, Description: "Pay school fees of $150.", Effect: CardEffectPay, Amount: 150},
		{ID: "CC14", Type: CardTypeCommunityChest, Description: "Receive $25 consultancy fee.", Effect: CardEffectCollect, Amount: 25},
		{ID: "CC15", Type: CardTypeCommunityChest, Description: "You are assessed for street repairs. $40 per house, $115 per hotel.", Effect: CardEffectRepairsPerBuilding, HouseCost: 40, HotelCost: 115},
		{ID: "CC16", Type: CardTypeCommunityChest, Description: "You have won second prize in a beauty contest. Collect $10.", Effect: CardEffectCollect, Amount: 10},
	}
}

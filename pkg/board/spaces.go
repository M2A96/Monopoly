package board

type SpaceType string
type ColorGroup string

const (
	SpaceTypeGO          SpaceType = "GO"
	SpaceTypeProperty    SpaceType = "PROPERTY"
	SpaceTypeRailroad    SpaceType = "RAILROAD"
	SpaceTypeUtility     SpaceType = "UTILITY"
	SpaceTypeTax         SpaceType = "TAX"
	SpaceTypeChance      SpaceType = "CHANCE"
	SpaceTypeCommunity   SpaceType = "COMMUNITY_CHEST"
	SpaceTypeJail        SpaceType = "JAIL"
	SpaceTypeFreeParking SpaceType = "FREE_PARKING"
	SpaceTypeGoToJail    SpaceType = "GO_TO_JAIL"
)

const (
	ColorBrown  ColorGroup = "BROWN"
	ColorLBlue  ColorGroup = "LIGHT_BLUE"
	ColorPink   ColorGroup = "PINK"
	ColorOrange ColorGroup = "ORANGE"
	ColorRed    ColorGroup = "RED"
	ColorYellow ColorGroup = "YELLOW"
	ColorGreen  ColorGroup = "GREEN"
	ColorDBlue  ColorGroup = "DARK_BLUE"
)

// BoardSpace is the static definition of a board space. Never mutate at runtime.
type BoardSpace struct {
	Index         int
	Type          SpaceType
	Name          string
	ColorGroup    ColorGroup
	Price         int
	// For PROPERTY: [no house, 1H, 2H, 3H, 4H, hotel]
	// For RAILROAD: [1 owned, 2 owned, 3 owned, 4 owned, 0, 0]
	// For UTILITY and special spaces: all zeros (rent computed at runtime)
	Rent          [6]int
	HouseCost     int
	MortgageValue int
	TaxAmount     int // for TAX spaces only
}

// Spaces is the static Monopoly board (positions 0–39). Never mutate at runtime.
var Spaces = []BoardSpace{
	// 0: GO
	{Index: 0, Type: SpaceTypeGO, Name: "Go"},
	// 1: Mediterranean Avenue (Brown)
	{Index: 1, Type: SpaceTypeProperty, Name: "Mediterranean Avenue", ColorGroup: ColorBrown,
		Price: 60, Rent: [6]int{2, 10, 30, 90, 160, 250}, HouseCost: 50, MortgageValue: 30},
	// 2: Community Chest
	{Index: 2, Type: SpaceTypeCommunity, Name: "Community Chest"},
	// 3: Baltic Avenue (Brown)
	{Index: 3, Type: SpaceTypeProperty, Name: "Baltic Avenue", ColorGroup: ColorBrown,
		Price: 60, Rent: [6]int{4, 20, 60, 180, 320, 450}, HouseCost: 50, MortgageValue: 30},
	// 4: Income Tax
	{Index: 4, Type: SpaceTypeTax, Name: "Income Tax", TaxAmount: 200},
	// 5: Reading Railroad
	{Index: 5, Type: SpaceTypeRailroad, Name: "Reading Railroad",
		Price: 200, Rent: [6]int{25, 50, 100, 200, 0, 0}, MortgageValue: 100},
	// 6: Oriental Avenue (Light Blue)
	{Index: 6, Type: SpaceTypeProperty, Name: "Oriental Avenue", ColorGroup: ColorLBlue,
		Price: 100, Rent: [6]int{6, 30, 90, 270, 400, 550}, HouseCost: 50, MortgageValue: 50},
	// 7: Chance
	{Index: 7, Type: SpaceTypeChance, Name: "Chance"},
	// 8: Vermont Avenue (Light Blue)
	{Index: 8, Type: SpaceTypeProperty, Name: "Vermont Avenue", ColorGroup: ColorLBlue,
		Price: 100, Rent: [6]int{6, 30, 90, 270, 400, 550}, HouseCost: 50, MortgageValue: 50},
	// 9: Connecticut Avenue (Light Blue)
	{Index: 9, Type: SpaceTypeProperty, Name: "Connecticut Avenue", ColorGroup: ColorLBlue,
		Price: 120, Rent: [6]int{8, 40, 100, 300, 450, 600}, HouseCost: 50, MortgageValue: 60},
	// 10: Jail / Just Visiting
	{Index: 10, Type: SpaceTypeJail, Name: "Jail"},
	// 11: St. Charles Place (Pink)
	{Index: 11, Type: SpaceTypeProperty, Name: "St. Charles Place", ColorGroup: ColorPink,
		Price: 140, Rent: [6]int{10, 50, 150, 450, 625, 750}, HouseCost: 100, MortgageValue: 70},
	// 12: Electric Company (Utility)
	{Index: 12, Type: SpaceTypeUtility, Name: "Electric Company",
		Price: 150, MortgageValue: 75},
	// 13: States Avenue (Pink)
	{Index: 13, Type: SpaceTypeProperty, Name: "States Avenue", ColorGroup: ColorPink,
		Price: 140, Rent: [6]int{10, 50, 150, 450, 625, 750}, HouseCost: 100, MortgageValue: 70},
	// 14: Virginia Avenue (Pink)
	{Index: 14, Type: SpaceTypeProperty, Name: "Virginia Avenue", ColorGroup: ColorPink,
		Price: 160, Rent: [6]int{12, 60, 180, 500, 700, 900}, HouseCost: 100, MortgageValue: 80},
	// 15: Pennsylvania Railroad
	{Index: 15, Type: SpaceTypeRailroad, Name: "Pennsylvania Railroad",
		Price: 200, Rent: [6]int{25, 50, 100, 200, 0, 0}, MortgageValue: 100},
	// 16: St. James Place (Orange)
	{Index: 16, Type: SpaceTypeProperty, Name: "St. James Place", ColorGroup: ColorOrange,
		Price: 180, Rent: [6]int{14, 70, 200, 550, 750, 950}, HouseCost: 100, MortgageValue: 90},
	// 17: Community Chest
	{Index: 17, Type: SpaceTypeCommunity, Name: "Community Chest"},
	// 18: Tennessee Avenue (Orange)
	{Index: 18, Type: SpaceTypeProperty, Name: "Tennessee Avenue", ColorGroup: ColorOrange,
		Price: 180, Rent: [6]int{14, 70, 200, 550, 750, 950}, HouseCost: 100, MortgageValue: 90},
	// 19: New York Avenue (Orange)
	{Index: 19, Type: SpaceTypeProperty, Name: "New York Avenue", ColorGroup: ColorOrange,
		Price: 200, Rent: [6]int{16, 80, 220, 600, 800, 1000}, HouseCost: 100, MortgageValue: 100},
	// 20: Free Parking
	{Index: 20, Type: SpaceTypeFreeParking, Name: "Free Parking"},
	// 21: Kentucky Avenue (Red)
	{Index: 21, Type: SpaceTypeProperty, Name: "Kentucky Avenue", ColorGroup: ColorRed,
		Price: 220, Rent: [6]int{18, 90, 250, 700, 875, 1050}, HouseCost: 150, MortgageValue: 110},
	// 22: Chance
	{Index: 22, Type: SpaceTypeChance, Name: "Chance"},
	// 23: Indiana Avenue (Red)
	{Index: 23, Type: SpaceTypeProperty, Name: "Indiana Avenue", ColorGroup: ColorRed,
		Price: 220, Rent: [6]int{18, 90, 250, 700, 875, 1050}, HouseCost: 150, MortgageValue: 110},
	// 24: Illinois Avenue (Red)
	{Index: 24, Type: SpaceTypeProperty, Name: "Illinois Avenue", ColorGroup: ColorRed,
		Price: 240, Rent: [6]int{20, 100, 300, 750, 925, 1100}, HouseCost: 150, MortgageValue: 120},
	// 25: B. & O. Railroad
	{Index: 25, Type: SpaceTypeRailroad, Name: "B. & O. Railroad",
		Price: 200, Rent: [6]int{25, 50, 100, 200, 0, 0}, MortgageValue: 100},
	// 26: Atlantic Avenue (Yellow)
	{Index: 26, Type: SpaceTypeProperty, Name: "Atlantic Avenue", ColorGroup: ColorYellow,
		Price: 260, Rent: [6]int{22, 110, 330, 800, 975, 1150}, HouseCost: 150, MortgageValue: 130},
	// 27: Ventnor Avenue (Yellow)
	{Index: 27, Type: SpaceTypeProperty, Name: "Ventnor Avenue", ColorGroup: ColorYellow,
		Price: 260, Rent: [6]int{22, 110, 330, 800, 975, 1150}, HouseCost: 150, MortgageValue: 130},
	// 28: Water Works (Utility)
	{Index: 28, Type: SpaceTypeUtility, Name: "Water Works",
		Price: 150, MortgageValue: 75},
	// 29: Marvin Gardens (Yellow)
	{Index: 29, Type: SpaceTypeProperty, Name: "Marvin Gardens", ColorGroup: ColorYellow,
		Price: 280, Rent: [6]int{24, 120, 360, 850, 1025, 1200}, HouseCost: 150, MortgageValue: 140},
	// 30: Go To Jail
	{Index: 30, Type: SpaceTypeGoToJail, Name: "Go To Jail"},
	// 31: Pacific Avenue (Green)
	{Index: 31, Type: SpaceTypeProperty, Name: "Pacific Avenue", ColorGroup: ColorGreen,
		Price: 300, Rent: [6]int{26, 130, 390, 900, 1100, 1275}, HouseCost: 200, MortgageValue: 150},
	// 32: North Carolina Avenue (Green)
	{Index: 32, Type: SpaceTypeProperty, Name: "North Carolina Avenue", ColorGroup: ColorGreen,
		Price: 300, Rent: [6]int{26, 130, 390, 900, 1100, 1275}, HouseCost: 200, MortgageValue: 150},
	// 33: Community Chest
	{Index: 33, Type: SpaceTypeCommunity, Name: "Community Chest"},
	// 34: Pennsylvania Avenue (Green)
	{Index: 34, Type: SpaceTypeProperty, Name: "Pennsylvania Avenue", ColorGroup: ColorGreen,
		Price: 320, Rent: [6]int{28, 150, 450, 1000, 1200, 1400}, HouseCost: 200, MortgageValue: 160},
	// 35: Short Line Railroad
	{Index: 35, Type: SpaceTypeRailroad, Name: "Short Line Railroad",
		Price: 200, Rent: [6]int{25, 50, 100, 200, 0, 0}, MortgageValue: 100},
	// 36: Chance
	{Index: 36, Type: SpaceTypeChance, Name: "Chance"},
	// 37: Park Place (Dark Blue)
	{Index: 37, Type: SpaceTypeProperty, Name: "Park Place", ColorGroup: ColorDBlue,
		Price: 350, Rent: [6]int{35, 175, 500, 1100, 1300, 1500}, HouseCost: 200, MortgageValue: 175},
	// 38: Luxury Tax
	{Index: 38, Type: SpaceTypeTax, Name: "Luxury Tax", TaxAmount: 100},
	// 39: Boardwalk (Dark Blue)
	{Index: 39, Type: SpaceTypeProperty, Name: "Boardwalk", ColorGroup: ColorDBlue,
		Price: 400, Rent: [6]int{50, 200, 600, 1400, 1700, 2000}, HouseCost: 200, MortgageValue: 200},
}

// SpaceByIndex provides O(1) lookup of board spaces by position. Never mutate at runtime.
var SpaceByIndex = func() map[int]BoardSpace {
	m := make(map[int]BoardSpace, len(Spaces))
	for _, s := range Spaces {
		m[s.Index] = s
	}
	return m
}()

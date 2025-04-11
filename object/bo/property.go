package bo

import (
	"database/sql"
	"encoding/json"
	"github/M2A96/Monopoly.git/object"
	"time"

	"github.com/google/uuid"
	"github.com/samber/lo"
)

type (
	Propertyer interface {
		GetID() int
		GetName() string
		GetColorGroup() string
		GetPrice() int
		GetHousePrice() int
		GetHotelPrice() int
		GetRent() int
		GetRentWithColorSet() int
		GetRentWith1House() int
		GetRentWith2Houses() int
		GetRentWith3Houses() int
		GetRentWith4Houses() int
		GetRentWithHotel() int
		GetMortgageValue() int
		GetOwnerID() *int
		GetHouses() int
		GetHasHotel() bool
		GetMortgaged() bool
	}

	// Property represents a property in the game
	property struct {
		id               int    `json:"id"`
		name             string `json:"name"`
		colorGroup       string `json:"color_group"`
		price            int    `json:"price"`
		housePrice       int    `json:"house_price"`
		hotelPrice       int    `json:"hotel_price"`
		rent             int    `json:"rent"`
		rentWithColorSet int    `json:"rent_with_color_set"`
		rentWith1House   int    `json:"rent_with_1_house"`
		rentWith2Houses  int    `json:"rent_with_2_houses"`
		rentWith3Houses  int    `json:"rent_with_3_houses"`
		rentWith4Houses  int    `json:"rent_with_4_houses"`
		rentWithHotel    int    `json:"rent_with_hotel"`
		mortgageValue    int    `json:"mortgage_value"`
		ownerID          *int   `json:"owner_id,omitempty"`
		houses           int    `json:"houses"`
		hasHotel         bool   `json:"has_hotel"`
		mortgaged        bool   `json:"mortgaged"`
	}
)

var (
	_ Propertyer       = (*property)(nil)
	_ object.GetMapper = (*property)(nil)
	_ json.Marshaler   = (*property)(nil)
)

// GetID implements Propertyer.
func (p *property) GetID() int {
	return p.id
}

// GetName implements Propertyer.
func (p *property) GetName() string {
	return p.name
}

// GetColorGroup implements Propertyer.
func (p *property) GetColorGroup() string {
	return p.colorGroup
}

// GetPrice implements Propertyer.
func (p *property) GetPrice() int {
	return p.price
}

// GetHousePrice implements Propertyer.
func (p *property) GetHousePrice() int {
	return p.housePrice
}

// GetHotelPrice implements Propertyer.
func (p *property) GetHotelPrice() int {
	return p.hotelPrice
}

// GetRent implements Propertyer.
func (p *property) GetRent() int {
	return p.rent
}

// GetRentWithColorSet implements Propertyer.
func (p *property) GetRentWithColorSet() int {
	return p.rentWithColorSet
}

// GetRentWith1House implements Propertyer.
func (p *property) GetRentWith1House() int {
	return p.rentWith1House
}

// GetRentWith2Houses implements Propertyer.
func (p *property) GetRentWith2Houses() int {
	return p.rentWith2Houses
}

// GetRentWith3Houses implements Propertyer.
func (p *property) GetRentWith3Houses() int {
	return p.rentWith3Houses
}

// GetRentWith4Houses implements Propertyer.
func (p *property) GetRentWith4Houses() int {
	return p.rentWith4Houses
}

// GetRentWithHotel implements Propertyer.
func (p *property) GetRentWithHotel() int {
	return p.rentWithHotel
}

// GetMortgageValue implements Propertyer.
func (p *property) GetMortgageValue() int {
	return p.mortgageValue
}

// GetOwnerID implements Propertyer.
func (p *property) GetOwnerID() *int {
	return p.ownerID
}

// GetHouses implements Propertyer.
func (p *property) GetHouses() int {
	return p.houses
}

// GetHasHotel implements Propertyer.
func (p *property) GetHasHotel() bool {
	return p.hasHotel
}

// GetMortgaged implements Propertyer.
func (p *property) GetMortgaged() bool {
	return p.mortgaged
}

// NewProperty creates a new property
func NewProperty(
	id uuid.UUID,
	propertyID int,
	name string,
	colorGroup string,
	price int,
	housePrice int,
	hotelPrice int,
	rent int,
	rentWithColorSet int,
	rentWith1House int,
	rentWith2Houses int,
	rentWith3Houses int,
	rentWith4Houses int,
	rentWithHotel int,
	mortgageValue int,
	ownerID *int,
	houses int,
	hasHotel bool,
	mortgaged bool,
	createdAt time.Time,
	updatedAt time.Time,
	deletedAt sql.NullTime,
) *property {
	return &property{
		id:               propertyID,
		name:             name,
		colorGroup:       colorGroup,
		price:            price,
		housePrice:       housePrice,
		hotelPrice:       hotelPrice,
		rent:             rent,
		rentWithColorSet: rentWithColorSet,
		rentWith1House:   rentWith1House,
		rentWith2Houses:  rentWith2Houses,
		rentWith3Houses:  rentWith3Houses,
		rentWith4Houses:  rentWith4Houses,
		rentWithHotel:    rentWithHotel,
		mortgageValue:    mortgageValue,
		ownerID:          ownerID,
		houses:           houses,
		hasHotel:         hasHotel,
		mortgaged:        mortgaged,
	}
}

// MarshalJSON implements json.Marshaler.
func (p *property) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.GetMap())
}

// GetMap implements object.GetMapper.
func (p *property) GetMap() map[string]any {
	return lo.Assign(
		map[string]any{
			"id":                  p.GetID(),
			"name":                p.GetName(),
			"color_group":         p.GetColorGroup(),
			"price":               p.GetPrice(),
			"house_price":         p.GetHousePrice(),
			"hotel_price":         p.GetHotelPrice(),
			"rent":                p.GetRent(),
			"rent_with_color_set": p.GetRentWithColorSet(),
			"rent_with_1_house":   p.GetRentWith1House(),
			"rent_with_2_houses":  p.GetRentWith2Houses(),
			"rent_with_3_houses":  p.GetRentWith3Houses(),
			"rent_with_4_houses":  p.GetRentWith4Houses(),
			"rent_with_hotel":     p.GetRentWithHotel(),
			"mortgage_value":      p.GetMortgageValue(),
			"owner_id":            p.GetOwnerID(),
			"houses":              p.GetHouses(),
			"has_hotel":           p.GetHasHotel(),
			"mortgaged":           p.GetMortgaged(),
		})
}

package dao

import (
	"encoding/json"
	"fmt"

	"github/M2A96/Monopoly.git/object"

	"github.com/google/uuid"

	"gorm.io/gorm"
)

type (
	// PropertyFilter is an interface.
	PropertyFilter interface {
		Filterer
		// GetIDs is a function.
		GetIDs() []uuid.UUID
		GetName() string
		GetColorGroup() string
		GetOwnerID() uuid.UUID
		GetHouses() int
		GetHasHotel() bool
		GetMortgaged() bool
	}

	propertyFilter struct {
		Filterer
		ids        []uuid.UUID
		name       string
		colorGroup string
		ownerID    uuid.UUID
		houses     int
		hasHotel   bool
		mortgaged  bool
	}
)

var (
	_ PropertyFilter   = (*propertyFilter)(nil)
	_ Filterer         = (*propertyFilter)(nil)
	_ json.Marshaler   = (*propertyFilter)(nil)
	_ object.GetMapper = (*propertyFilter)(nil)
)

func NewPropertyFilter(
	ids []uuid.UUID,
	name string,
	colorGroup string,
	ownerID uuid.UUID,
	houses int,
	hasHotel bool,
	mortgaged bool,
) PropertyFilter {
	return &propertyFilter{
		ids:        ids,
		name:       name,
		colorGroup: colorGroup,
		ownerID:    ownerID,
		houses:     houses,
		hasHotel:   hasHotel,
		mortgaged:  mortgaged,
	}
}

// GetMap implements object.GetMapper.
func (p *propertyFilter) GetMap() map[string]any {
	return map[string]any{
		"ids":         p.GetIDs(),
		"name":        p.GetName(),
		"color_group": p.GetColorGroup(),
		"owner_id":    p.GetOwnerID(),
		"houses":      p.GetHouses(),
		"has_hotel":   p.GetHasHotel(),
		"mortgaged":   p.GetMortgaged(),
	}
}

// MarshalJSON implements json.Marshaler.
func (p *propertyFilter) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.GetMap())
}

// Filter implements PropertyFilter.
// Subtle: this method shadows the method (Filterer).Filter of propertyFilter.Filterer.
func (p *propertyFilter) Filter(
	gormDB *gorm.DB,
) *gorm.DB {
	if len(p.GetIDs()) != 0 {
		gormDB = gormDB.
			Where(fmt.Sprintf(`%[1]s.id IN ?`, "property"), p.GetIDs())
	}

	if p.GetName() != "" {
		gormDB = gormDB.
			Where(fmt.Sprintf(`%[1]s.name = ?`, "property"), p.GetName())
	}

	if p.GetColorGroup() != "" {
		gormDB = gormDB.
			Where(fmt.Sprintf(`%[1]s.color_group = ?`, "property"), p.GetColorGroup())
	}

	if p.GetOwnerID() != uuid.Nil {
		gormDB = gormDB.
			Where(fmt.Sprintf(`%[1]s.owner_id = ?`, "property"), p.GetOwnerID())
	}

	if p.GetHouses() != 0 {
		gormDB = gormDB.
			Where(fmt.Sprintf(`%[1]s.houses = ?`, "property"), p.GetHouses())
	}

	if p.GetHasHotel() {
		gormDB = gormDB.
			Where(fmt.Sprintf(`%[1]s.has_hotel = ?`, "property"), p.GetHasHotel())
	}

	if p.GetMortgaged() {
		gormDB = gormDB.
			Where(fmt.Sprintf(`%[1]s.mortgaged = ?`, "property"), p.GetMortgaged())
	}

	return gormDB
}

// GetIDs implements PropertyFilter.
func (p *propertyFilter) GetIDs() []uuid.UUID {
	return p.ids
}

// GetName implements PropertyFilter.
func (p *propertyFilter) GetName() string {
	return p.name
}

// GetColorGroup implements PropertyFilter.
func (p *propertyFilter) GetColorGroup() string {
	return p.colorGroup
}

// GetOwnerID implements PropertyFilter.
func (p *propertyFilter) GetOwnerID() uuid.UUID {
	return p.ownerID
}

// GetHouses implements PropertyFilter.
func (p *propertyFilter) GetHouses() int {
	return p.houses
}

// GetHasHotel implements PropertyFilter.
func (p *propertyFilter) GetHasHotel() bool {
	return p.hasHotel
}

// GetMortgaged implements PropertyFilter.
func (p *propertyFilter) GetMortgaged() bool {
	return p.mortgaged
}

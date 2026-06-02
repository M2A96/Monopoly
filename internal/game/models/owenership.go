package models

import "github.com/google/uuid"

// Buildings represents a development level on a property.
// 0 = no buildings, 1–4 = houses, 5 = hotel.
const HotelLevel = 5

// Ownership records who owns a purchasable space and its current development state.
type Ownership struct {
	SpaceIndex int
	OwnerID    uuid.UUID
	Mortgaged  bool
	Buildings  int // 0–4 houses or 5 (HotelLevel) for a hotel
}

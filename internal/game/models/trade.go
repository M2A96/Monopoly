package models

import "github.com/google/uuid"

type TradeStatus string

const (
	TradeStatusPending  TradeStatus = "PENDING"
	TradeStatusAccepted TradeStatus = "ACCEPTED"
	TradeStatusRejected TradeStatus = "REJECTED"
	TradeStatusCanceled TradeStatus = "CANCELED"
)

// Trade represents a proposed exchange between two players.
type Trade struct {
	TradeID         uuid.UUID
	ProposerID      uuid.UUID
	RecipientID     uuid.UUID
	OfferedMoney    int
	OfferedSpaces   []int // space indices the proposer offers
	RequestedMoney  int
	RequestedSpaces []int // space indices the proposer requests
	Status          TradeStatus
}

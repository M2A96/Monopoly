package models

import "github.com/google/uuid"

// Debt is a payment the debtor could not cover when it was charged.
// CreditorID == uuid.Nil means the bank. The debtor must raise funds and pay,
// or declare bankruptcy, before the game continues.
type Debt struct {
	DebtorID   uuid.UUID
	CreditorID uuid.UUID
	Amount     int
	Reason     string
}

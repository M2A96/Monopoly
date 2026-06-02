package models

const (
	BankStartingHouses = 32
	BankStartingHotels = 12
)

// Bank tracks the finite supply of buildings.
// The bank's money supply is unlimited and not modeled here.
type Bank struct {
	Houses int // available house tokens (starts at 32)
	Hotels int // available hotel tokens (starts at 12)
}

func NewBank() Bank {
	return Bank{Houses: BankStartingHouses, Hotels: BankStartingHotels}
}

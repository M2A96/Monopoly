package dao

//go:generate mockgen -destination=../../test/v2/filter.go -package=test -mock_names=Filterer=MockFilter . Filterer

import "gorm.io/gorm"

// Filterer is an interface.
type Filterer interface {
	// Filter is a function.
	Filter(
		*gorm.DB,
	) *gorm.DB
}

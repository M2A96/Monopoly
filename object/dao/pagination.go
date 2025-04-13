package dao

//go:generate mockgen -destination=../../test/v2/pagination.go -package=test -mock_names=Paginationer=MockPagination . Paginationer

import (
	"encoding/json"

	"github/M2A96/Monopoly.git/object"

	"gorm.io/gorm"
)

type (
	// Paginationer is an interface.
	Paginationer interface {
		// GetCursorer is a function.
		GetCursorer() Cursorer
		// GetLimit is a function.
		GetLimit() uint32
		// Pagination is a function.
		Pagination(
			table string,
		) func(*gorm.DB) *gorm.DB
	}

	pagination struct {
		cursorer Cursorer
		limit    uint32
	}
)

var (
	_ Paginationer     = (*pagination)(nil)
	_ json.Marshaler   = (*pagination)(nil)
	_ object.GetMapper = (*pagination)(nil)
)

// NewPagination is a function.
func NewPagination(
	cursorer Cursorer,
	limit uint32,
) *pagination {
	return &pagination{
		cursorer: cursorer,
		limit:    limit,
	}
}

// GetCursorer is a function.
func (dao *pagination) GetCursorer() Cursorer {
	return dao.cursorer
}

// GetLimit is a function.
func (dao *pagination) GetLimit() uint32 {
	return dao.limit
}

// GetMap is a function.
func (dao *pagination) GetMap() map[string]any {
	return map[string]any{
		"cursorer": dao.GetCursorer(),
		"limit":    dao.GetLimit(),
	}
}

// MarshalJSON is a function.
// read more https://pkg.go.dev/encoding/json#Marshaler
func (dao *pagination) MarshalJSON() ([]byte, error) {
	return json.Marshal(dao.GetMap())
}

// Pagination is a function.
func (dao *pagination) Pagination(
	table string,
) func(*gorm.DB) *gorm.DB {
	return func(
		gormDB *gorm.DB,
	) *gorm.DB {
		if dao.GetCursorer() != nil {
			gormDB.Scopes(dao.GetCursorer().Query(table))
		}

		if dao.GetLimit() != 0 {
			gormDB.Limit(int(dao.GetLimit()) + 1)
		}

		return gormDB
	}
}

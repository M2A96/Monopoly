package dao

import (
	"database/sql"
	"github/M2A96/Monopoly.git/object"
	"time"
)

type (
	// CUDer is an interface.
	CUDer interface {
		object.GetMapper
		// GetCreatedAt is a function.
		GetCreatedAt() time.Time
		// GetUpdatedAt is a function.
		GetUpdatedAt() time.Time
		// GetDeletedAt is a function.
		GetDeletedAt() sql.NullTime
	}

	cud struct {
		createdAt time.Time
		updatedAt time.Time
		deletedAt sql.NullTime
	}
)

var (
	_ CUDer            = (*cud)(nil)
	_ object.GetMapper = (*cud)(nil)
)

// NewCUD is a function.
func NewCUD(
	createdAt time.Time,
	updatedAt time.Time,
	deletedAt sql.NullTime,
) *cud {
	return &cud{
		createdAt: createdAt,
		updatedAt: updatedAt,
		deletedAt: deletedAt,
	}
}

// NewCUDerFromMap is a function.
func NewCUDerFromMap(
	value map[string]any,
) (CUDer, error) {
	createdAt, ok := value["created_at"].(time.Time)
	if !ok {
		return nil, object.ErrTypeAssertion
	}

	updatedAt, ok := value["updated_at"].(time.Time)
	if !ok {
		return nil, object.ErrTypeAssertion
	}

	var deletedAt sql.NullTime

	if err := deletedAt.Scan(value["deleted_at"]); err != nil {
		return nil, err
	}

	return NewCUD(createdAt, updatedAt, deletedAt), nil
}

// CUDerComparer is a function.
func CUDerComparer(
	first CUDer,
	second CUDer,
) bool {
	return first.GetCreatedAt().Equal(second.GetCreatedAt()) &&
		first.GetUpdatedAt().Equal(second.GetUpdatedAt()) &&
		first.GetDeletedAt().Time.Equal(second.GetDeletedAt().Time) &&
		first.GetDeletedAt().Valid == second.GetDeletedAt().Valid
}

// GetCreatedAt is a function.
func (dao *cud) GetCreatedAt() time.Time {
	return dao.createdAt
}

// GetUpdatedAt is a function.
func (dao *cud) GetUpdatedAt() time.Time {
	return dao.updatedAt
}

// GetDeletedAt is a function.
func (dao *cud) GetDeletedAt() sql.NullTime {
	return dao.deletedAt
}

// GetMap is a function.
func (dao *cud) GetMap() map[string]any {
	return map[string]any{
		"created_at": dao.GetCreatedAt(),
		"updated_at": dao.GetUpdatedAt(),
		"deleted_at": dao.GetDeletedAt(),
	}
}

package bo

import (
	"github/M2A96/Monopoly.git/object"

	"github.com/google/uuid"
)

type (
	// BOer is an interface.
	BOer interface {
		// GetID is a function.
		GetID() uuid.UUID
	}

	bo struct {
		id uuid.UUID
	}
)

var (
	_ BOer             = (*bo)(nil)
	_ object.GetMapper = (*bo)(nil)
)

// NewBO is a function.
func NewBO(
	id uuid.UUID,
) *bo {
	return &bo{
		id: id,
	}
}

// BOerComparer is a function.
func BOerComparer(
	first BOer,
	second BOer,
) bool {
	return first.GetID() == second.GetID()
}

// GetID is a function.
func (bo *bo) GetID() uuid.UUID {
	return bo.id
}

// GetMap is a function.
func (bo *bo) GetMap() map[string]any {
	return map[string]any{
		"id": bo.GetID(),
	}
}

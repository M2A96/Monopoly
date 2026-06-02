package dao

import (
	"reflect"
	"sort"

	"github/M2A96/Monopoly.git/object"

	"github.com/google/uuid"
	"github.com/samber/lo"
)

type (
	// CUDIDer is an interface.
	CUDIDer interface {
		object.GetMapper
		// GetID is a function.
		GetID() map[string]uuid.UUID
	}

	cudID struct {
		id map[string]uuid.UUID
	}
)

var (
	_ CUDIDer          = (*cudID)(nil)
	_ object.GetMapper = (*cudID)(nil)
)

// NewCUDID is a function.
func NewCUDID(
	id map[string]uuid.UUID,
) *cudID {
	return &cudID{
		id: id,
	}
}

// NewCUDIDerFromMap is a function.
func NewCUDIDerFromMap(
	uuider object.UUIDer,
	value map[string]any,
) (CUDIDer, error) {
	result := map[string]uuid.UUID{}

	keys := lo.Keys(value)
	sort.Strings(keys)

	for _, key := range keys {
		valSTR, ok := value[key].(string)
		if !ok {
			return nil, object.ErrTypeAssertion
		}

		valUUID, err := uuider.Parse(valSTR)
		if err != nil {
			return nil, err
		}

		result[key] = valUUID
	}

	return NewCUDID(result), nil
}

// CUDIDerComparer is a function.
func CUDIDerComparer(
	first CUDIDer,
	second CUDIDer,
) bool {
	return reflect.DeepEqual(first, second)
}

// GetID is a function.
func (dao *cudID) GetID() map[string]uuid.UUID {
	return dao.id
}

// GetMap is a function.
func (dao *cudID) GetMap() map[string]any {
	result := map[string]any{}

	for key, value := range dao.GetID() {
		result[key] = value
	}

	return result
}

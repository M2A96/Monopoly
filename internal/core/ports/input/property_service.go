package input

import (
	"context"
	"github/M2A96/Monopoly.git/internal/core/domain/bo"
	"github/M2A96/Monopoly.git/object/dao"

	"github.com/google/uuid"
)

type (
	PropertyServicer interface {
		Get(
			context.Context,
			uuid.UUID,
		) (bo.Propertyer, error)

		List(
			context.Context,
			dao.Paginationer,
			dao.PropertyFilter,
		) ([]bo.Propertyer, dao.Cursorer, error)

		CreateProperty(
			context.Context,
			bo.Propertyer,
		) (uuid.UUID, error)
	}

	GetPropertyServicer interface {
		GetPropertyServicer() PropertyServicer
	}
)

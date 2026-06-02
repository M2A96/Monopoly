package output

import (
	"github/M2A96/Monopoly.git/config"
	"github/M2A96/Monopoly.git/log"
	"github/M2A96/Monopoly.git/object"
	"github/M2A96/Monopoly.git/object/dao"

	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
)

type (
	// PropertyRepositorier is an interface.
	PropertyRepositorier Repositorier[dao.Propertyer, dao.PropertyFilter]

	// GetPropertyRepositorier is an interface.
	GetPropertyRepositorier interface {
		// GetPropertyRepositorier is a function.
		GetPropertyRepositorier() PropertyRepositorier
	}

	propertyRepository struct {
		repository
	}

	propertyRepositoryOptioner = repositoryOptioner
)

// NewPropertyRepository is a function.
func NewPropertyRepository(
	configConfigger config.Configger,
	logRuntimeLogger log.RuntimeLogger,
	objectUUIDer object.UUIDer,
	traceTracer trace.Tracer,
	optioners ...propertyRepositoryOptioner,
) *propertyRepository {
	return &propertyRepository{
		repository: *NewRepository(
			configConfigger,
			logRuntimeLogger,
			objectUUIDer,
			traceTracer,
			optioners...,
		),
	}
}

// WithPropertyRepositoryTimer is a function.
func WithPropertyRepositoryTimer(
	objectTimer object.Timer,
) propertyRepositoryOptioner {
	return WithRepositoryTimer(objectTimer)
}

// WithPropertyRepositoryDB is a function.
func WithPropertyRepositoryDB(
	gormDB *gorm.DB,
) propertyRepositoryOptioner {
	return WithRepositoryDB(gormDB, object.URITableProperty)
}

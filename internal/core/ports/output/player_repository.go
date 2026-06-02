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
	// PlayerRepositorier is an interface.
	PlayerRepositorier Repositorier[dao.Player, dao.PlayerFilter]

	// GetPlayerRepositorier is an interface.
	GetPlayerRepositorier interface {
		// GetPlayerRepositorier is a function.
		GetPlayerRepositorier() PlayerRepositorier
	}

	playerRepository struct {
		repository
	}

	playerRepositoryOptioner = repositoryOptioner
)

// NewPlayerRepository is a function.
func NewPlayerRepository(
	configConfigger config.Configger,
	logRuntimeLogger log.RuntimeLogger,
	objectUUIDer object.UUIDer,
	traceTracer trace.Tracer,
	optioners ...playerRepositoryOptioner,
) *playerRepository {
	return &playerRepository{
		repository: *NewRepository(
			configConfigger,
			logRuntimeLogger,
			objectUUIDer,
			traceTracer,
			optioners...,
		),
	}
}

// WithPlayerRepositoryTimer is a function.
func WithPlayerRepositoryTimer(
	objectTimer object.Timer,
) playerRepositoryOptioner {
	return WithRepositoryTimer(objectTimer)
}

// WithPlayerRepositoryDB is a function.
func WithPlayerRepositoryDB(
	gormDB *gorm.DB,
) playerRepositoryOptioner {
	return WithRepositoryDB(gormDB, object.URITablePlayer)
}

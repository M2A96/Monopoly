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
	// GameStateRepositorier is an interface.
	GameStateRepositorier Repositorier[dao.GameStater, dao.GameStateFilter]

	// GetGameStateRepositorier is an interface.
	GetGameStateRepositorier interface {
		// GetGameStateRepositorier is a function.
		GetGameStateRepositorier() GameStateRepositorier
	}

	gameStateRepository struct {
		repository
	}

	gameStateRepositoryOptioner = repositoryOptioner
)

// NewGameStateRepository is a function.
func NewGameStateRepository(
	configConfigger config.Configger,
	logRuntimeLogger log.RuntimeLogger,
	objectUUIDer object.UUIDer,
	traceTracer trace.Tracer,
	optioners ...gameStateRepositoryOptioner,
) *gameStateRepository {
	return &gameStateRepository{
		repository: *NewRepository(
			configConfigger,
			logRuntimeLogger,
			objectUUIDer,
			traceTracer,
			optioners...,
		),
	}
}

// WithGameStateRepositoryTimer is a function.
func WithGameStateRepositoryTimer(
	objectTimer object.Timer,
) gameStateRepositoryOptioner {
	return WithRepositoryTimer(objectTimer)
}

// WithGameStateRepositoryDB is a function.
func WithGameStateRepositoryDB(
	gormDB *gorm.DB,
) gameStateRepositoryOptioner {
	return WithRepositoryDB(gormDB, "game_state")
}

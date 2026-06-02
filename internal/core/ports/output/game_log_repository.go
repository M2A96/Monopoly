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
	// GameLogRepositorier is an interface.
	GameLogRepositorier Repositorier[dao.GameLogger, dao.GameLogFilter]

	// GetGameLogRepositorier is an interface.
	GetGameLogRepositorier interface {
		// GetGameLogRepositorier is a function.
		GetGameLogRepositorier() GameLogRepositorier
	}

	gameLogRepository struct {
		repository
	}

	gameLogRepositoryOptioner = repositoryOptioner
)

// NewGameLogRepository is a function.
func NewGameLogRepository(
	configConfigger config.Configger,
	logRuntimeLogger log.RuntimeLogger,
	objectUUIDer object.UUIDer,
	traceTracer trace.Tracer,
	optioners ...gameLogRepositoryOptioner,
) *gameLogRepository {
	return &gameLogRepository{
		repository: *NewRepository(
			configConfigger,
			logRuntimeLogger,
			objectUUIDer,
			traceTracer,
			optioners...,
		),
	}
}

// WithGameLogRepositoryTimer is a function.
func WithGameLogRepositoryTimer(
	objectTimer object.Timer,
) gameLogRepositoryOptioner {
	return WithRepositoryTimer(objectTimer)
}

// WithGameLogRepositoryDB is a function.
func WithGameLogRepositoryDB(
	gormDB *gorm.DB,
) gameLogRepositoryOptioner {
	return WithRepositoryDB(gormDB, "game_log")
}

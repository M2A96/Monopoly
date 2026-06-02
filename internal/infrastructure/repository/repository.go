package repository

import (
	"github/M2A96/Monopoly.git/config"
	"github/M2A96/Monopoly.git/log"
	"github/M2A96/Monopoly.git/object"

	"go.opentelemetry.io/otel/trace"

	"gorm.io/gorm"
)

type (
	repository struct {
		configConfigger  config.Configger
		gormDB           *gorm.DB
		logRuntimeLogger log.RuntimeLogger
		objectTimer      object.Timer
		objectUUIDer     object.UUIDer
		traceTracer      trace.Tracer
	}

	repositoryOptioner interface {
		apply(*repository)
	}

	repositoryOptionerFunc func(*repository)
)

// GetTracer implements util.GetTracer.
func (repository *repository) GetTracer() trace.Tracer {
	return repository.traceTracer
}

// GetDB implements GetDBer.
func (repository *repository) GetDB() *gorm.DB {
	return repository.gormDB
}

// GetConfigger implements config.GetConfigger.
func (repository *repository) GetConfigger() config.Configger {
	return repository.configConfigger
}

// GetRuntimeLogger implements log.GetRuntimeLogger.
func (repository *repository) GetRuntimeLogger() log.RuntimeLogger {
	return repository.logRuntimeLogger
}

// GetTimer implements object.GetTimer.
func (repository *repository) GetTimer() object.Timer {
	return repository.objectTimer
}

// GetUUIDer implements object.GetUUIDer.
func (repository *repository) GetUUIDer() object.UUIDer {
	return repository.objectUUIDer
}

// apply implements repositoryOptioner.
func (f repositoryOptionerFunc) apply(repository *repository) {
	f(repository)
}

// NewRepository is a function.
func NewRepository(
	configConfigger config.Configger,
	logRuntimeLogger log.RuntimeLogger,
	objectUUIDer object.UUIDer,
	traceTracer trace.Tracer,
	optioners ...repositoryOptioner,
) *repository {
	repository := &repository{
		configConfigger:  configConfigger,
		gormDB:           nil,
		logRuntimeLogger: logRuntimeLogger,
		objectTimer:      nil,
		objectUUIDer:     objectUUIDer,
		traceTracer:      traceTracer,
	}

	for _, optioner := range optioners {
		optioner.apply(repository)
	}

	return repository
}

// WithRepositoryTimer is a function.
func WithRepositoryTimer(
	objectTimer object.Timer,
) repositoryOptionerFunc {
	return func(repository *repository) {
		repository.objectTimer = objectTimer
	}
}

// WithRepositoryDB is a function.
func WithRepositoryDB(
	gormDB *gorm.DB,
	tableName string,
) repositoryOptionerFunc {
	return func(repository *repository) {
		repository.gormDB = gormDB.Table(tableName)
	}
}

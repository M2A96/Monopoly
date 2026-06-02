package output

import (
	"context"
	"github/M2A96/Monopoly.git/config"
	"github/M2A96/Monopoly.git/log"
	"github/M2A96/Monopoly.git/object"
	"github/M2A96/Monopoly.git/object/dao"
	"github/M2A96/Monopoly.git/util"
	"time"

	"go.opentelemetry.io/otel/trace"

	"gorm.io/gorm"
)

type (
	// Repositorier is an interface.
	Repositorier[
		daoCUDer any,
		daoFilterer dao.Filterer,
	] interface {
		// Create is a function.
		Create(
			context.Context,
			daoCUDer,
		) (dao.CUDIDer, error)
		// Read is a function.
		Read(
			context.Context,
			dao.CUDIDer,
		) (daoCUDer, error)
		// ReadList is a function.
		ReadList(
			context.Context,
			dao.Paginationer,
			daoFilterer,
		) ([]daoCUDer, dao.Cursorer, error)
		// Update is a function.
		Update(
			context.Context,
			daoCUDer,
		) (time.Time, error)
		// Delete is a function.
		Delete(
			context.Context,
			dao.CUDIDer,
		) (time.Time, error)
	}

	// GetDBer is an interface.
	GetDBer interface {
		GetDB() *gorm.DB
	}

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

var (
	_ GetDBer              = (*repository)(nil)
	_ config.GetConfigger  = (*repository)(nil)
	_ log.GetRuntimeLogger = (*repository)(nil)
	_ object.GetTimer      = (*repository)(nil)
	_ object.GetUUIDer     = (*repository)(nil)
	_ util.GetTracer       = (*repository)(nil)
)

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

	return repository.WithOptioners(optioners...)
}

// WithRepositoryTimer is a function.
func WithRepositoryTimer(
	objectTimer object.Timer,
) repositoryOptioner {
	return repositoryOptionerFunc(func(
		config *repository,
	) {
		config.objectTimer = objectTimer
	})
}

// WithRepositoryDB is a function.
func WithRepositoryDB(
	gormDB *gorm.DB,
	table string,
) repositoryOptioner {
	return repositoryOptionerFunc(func(
		config *repository,
	) {
		config.gormDB = gormDB.
			Table(table).
			Session(&gorm.Session{
				DryRun:                   false,
				PrepareStmt:              true,
				NewDB:                    true,
				Initialized:              false,
				SkipHooks:                true,
				SkipDefaultTransaction:   true,
				DisableNestedTransaction: true,
				AllowGlobalUpdate:        false,
				FullSaveAssociations:     false,
				QueryFields:              true,
				Context:                  nil,
				Logger:                   nil,
				NowFunc:                  nil,
				CreateBatchSize:          0,
			})
	})
}

// GetDB is a function.
func (repository *repository) GetDB() *gorm.DB {
	return repository.gormDB
}

// GetConfigger is a function.
func (repository *repository) GetConfigger() config.Configger {
	return repository.configConfigger
}

// GetRuntimeLogger is a function.
func (repository *repository) GetRuntimeLogger() log.RuntimeLogger {
	return repository.logRuntimeLogger
}

// GetTimer is a function.
func (repository *repository) GetTimer() object.Timer {
	return repository.objectTimer
}

// GetUUIDer is a function.
func (repository *repository) GetUUIDer() object.UUIDer {
	return repository.objectUUIDer
}

// WithOptioners is a function.
func (repository *repository) WithOptioners(
	optioners ...repositoryOptioner,
) *repository {
	newRepository := repository.clone()
	for _, optioner := range optioners {
		optioner.apply(newRepository)
	}

	return newRepository
}

func (repository *repository) clone() *repository {
	newRepository := repository

	return newRepository
}

func (optionerFunc repositoryOptionerFunc) apply(
	repository *repository,
) {
	optionerFunc(repository)
}

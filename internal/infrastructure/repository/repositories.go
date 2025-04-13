package repository

import (
	"github/M2A96/Monopoly.git/config"
	"github/M2A96/Monopoly.git/internal/core/ports/output"
	"github/M2A96/Monopoly.git/log"
	"github/M2A96/Monopoly.git/object"

	"go.opentelemetry.io/otel/trace"
)

type (
	Repositories interface {
		output.GetGameRepositorier
	}

	GetRepositories interface {
		GetRepositories() Repositories
	}
	repositories struct {
		gameRepository output.GameRepositorier
	}

	optionRepositoriers interface {
		apply(*repositories)
	}

	optionRepositoriersFunc func(*repositories)
)

var (
	_ output.GetGameRepositorier = (*repositories)(nil)
)

// GetGameRepositorier implements GetGameRepositorier.
func (repository *repositories) GetGameRepositorier() output.GameRepositorier {
	return repository.gameRepository
}

func NewRepositories(
	optioners ...optionRepositoriers,
) *repositories {
	repositories := &repositories{
		gameRepository: nil,
	}

	return repositories.WithOptioners(optioners...)
}

// WithGameRepositorier is a function.
func WithGameRepositorier(
	configConfigger config.Configger,
	logRuntimeLogger log.RuntimeLogger,
	objectUUIDer object.UUIDer,
	traceTracer trace.Tracer,
	optioners ...gameRepositoryOptioner,
) optionRepositoriers {
	return optionRepositoriersFunc(func(
		repository *repositories,
	) {
		repository.gameRepository = NewGameRepository(
			configConfigger,
			logRuntimeLogger,
			objectUUIDer,
			traceTracer,
			optioners...,
		)
	})
}

// WithOptioners is a function.
func (repository *repositories) WithOptioners(
	optioners ...optionRepositoriers,
) *repositories {
	newRepository := repository.clone()
	for _, optioner := range optioners {
		optioner.apply(newRepository)
	}

	return newRepository
}

func (repository *repositories) clone() *repositories {
	newRepository := repository

	return newRepository
}

func (optionerFunc optionRepositoriersFunc) apply(
	repository *repositories,
) {
	optionerFunc(repository)
}

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
		output.GetPlayerRepositorier
		output.GetPropertyRepositorier
		output.GetGameLogRepositorier
		output.GetGameStateRepositorier
		output.GetTradeRequestRepositorier
	}

	GetRepositories interface {
		GetRepositories() Repositories
	}
	repositories struct {
		gameRepository         output.GameRepositorier
		playerRepository       output.PlayerRepositorier
		propertyRepository     output.PropertyRepositorier
		gameLogRepository      output.GameLogRepositorier
		gameStateRepository    output.GameStateRepositorier
		tradeRequestRepository output.TradeRequestRepositorier
	}

	optionRepositoriers interface {
		apply(*repositories)
	}

	optionRepositoriersFunc func(*repositories)
)

var (
	_ output.GetGameRepositorier         = (*repositories)(nil)
	_ output.GetPlayerRepositorier       = (*repositories)(nil)
	_ output.GetPropertyRepositorier     = (*repositories)(nil)
	_ output.GetGameLogRepositorier      = (*repositories)(nil)
	_ output.GetGameStateRepositorier    = (*repositories)(nil)
	_ output.GetTradeRequestRepositorier = (*repositories)(nil)
)

// GetGameRepositorier implements GetGameRepositorier.
func (repository *repositories) GetGameRepositorier() output.GameRepositorier {
	return repository.gameRepository
}

// GetPlayerRepositorier implements GetPlayerRepositorier.
func (repository *repositories) GetPlayerRepositorier() output.PlayerRepositorier {
	return repository.playerRepository
}

// GetPropertyRepositorier implements GetPropertyRepositorier.
func (repository *repositories) GetPropertyRepositorier() output.PropertyRepositorier {
	return repository.propertyRepository
}

// GetGameLogRepositorier implements GetGameLogRepositorier.
func (repository *repositories) GetGameLogRepositorier() output.GameLogRepositorier {
	return repository.gameLogRepository
}

// GetGameStateRepositorier implements GetGameStateRepositorier.
func (repository *repositories) GetGameStateRepositorier() output.GameStateRepositorier {
	return repository.gameStateRepository
}

// GetTradeRequestRepositorier implements GetTradeRequestRepositorier.
func (repository *repositories) GetTradeRequestRepositorier() output.TradeRequestRepositorier {
	return repository.tradeRequestRepository
}

func NewRepositories(
	optioners ...optionRepositoriers,
) *repositories {
	repositories := &repositories{
		gameRepository:         nil,
		playerRepository:       nil,
		propertyRepository:     nil,
		gameLogRepository:      nil,
		gameStateRepository:    nil,
		tradeRequestRepository: nil,
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

// WithPlayerRepositorier is a function.
func WithPlayerRepositorier(
	configConfigger config.Configger,
	logRuntimeLogger log.RuntimeLogger,
	objectUUIDer object.UUIDer,
	traceTracer trace.Tracer,
	optioners ...playerRepositoryOptioner,
) optionRepositoriers {
	return optionRepositoriersFunc(func(
		repository *repositories,
	) {
		repository.playerRepository = NewPlayerRepository(
			configConfigger,
			logRuntimeLogger,
			objectUUIDer,
			traceTracer,
			optioners...,
		)
	})
}

// WithPropertyRepositorier is a function.
func WithPropertyRepositorier(
	configConfigger config.Configger,
	logRuntimeLogger log.RuntimeLogger,
	objectUUIDer object.UUIDer,
	traceTracer trace.Tracer,
	optioners ...propertyRepositoryOptioner,
) optionRepositoriers {
	return optionRepositoriersFunc(func(
		repository *repositories,
	) {
		repository.propertyRepository = NewPropertyRepository(
			configConfigger,
			logRuntimeLogger,
			objectUUIDer,
			traceTracer,
			optioners...,
		)
	})
}

// WithGameLogRepositorier is a function.
func WithGameLogRepositorier(
	configConfigger config.Configger,
	logRuntimeLogger log.RuntimeLogger,
	objectUUIDer object.UUIDer,
	traceTracer trace.Tracer,
	optioners ...gameLogRepositoryOptioner,
) optionRepositoriers {
	return optionRepositoriersFunc(func(
		repository *repositories,
	) {
		repository.gameLogRepository = NewGameLogRepository(
			configConfigger,
			logRuntimeLogger,
			objectUUIDer,
			traceTracer,
			optioners...,
		)
	})
}

// WithGameStateRepositorier is a function.
func WithGameStateRepositorier(
	configConfigger config.Configger,
	logRuntimeLogger log.RuntimeLogger,
	objectUUIDer object.UUIDer,
	traceTracer trace.Tracer,
	optioners ...gameStateRepositoryOptioner,
) optionRepositoriers {
	return optionRepositoriersFunc(func(
		repository *repositories,
	) {
		repository.gameStateRepository = NewGameStateRepository(
			configConfigger,
			logRuntimeLogger,
			objectUUIDer,
			traceTracer,
			optioners...,
		)
	})
}

// WithTradeRequestRepositorier is a function.
func WithTradeRequestRepositorier(
	configConfigger config.Configger,
	logRuntimeLogger log.RuntimeLogger,
	objectUUIDer object.UUIDer,
	traceTracer trace.Tracer,
	optioners ...tradeRequestRepositoryOptioner,
) optionRepositoriers {
	return optionRepositoriersFunc(func(
		repository *repositories,
	) {
		repository.tradeRequestRepository = NewTradeRequestRepository(
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

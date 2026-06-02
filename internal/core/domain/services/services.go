package services

import (
	"github/M2A96/Monopoly.git/config"
	"github/M2A96/Monopoly.git/internal/core/ports/input"
	"github/M2A96/Monopoly.git/log"
	"github/M2A96/Monopoly.git/object"

	"go.opentelemetry.io/otel/trace"
)

type (
	Servicers interface {
		input.GetGameServicer
		input.GetPlayerServicer
		input.GetPropertyServicer
		input.GetGameLogServicer
		input.GetGameStateServicer
		input.GetTradeRequestServicer
	}

	// GetServicers is an interface.
	GetServicers interface {
		// GetServicers is a function.
		GetServicers() Servicers
	}

	// WithServicers is an interface.
	WithServicers interface {
		// WithServicers is a function.
		WithServicers(
			Servicers,
		)
	}

	optionServicer interface {
		apply(*service)
	}

	optionServicerFunc func(*service)

	service struct {
		gameService         input.GameServicer
		playerService       input.PlayerServicer
		propertyService     input.PropertyServicer
		gameLogService      input.GameLogServicer
		gameStateService    input.GameStateServicer
		tradeRequestService input.TradeRequestServicer
	}
)

var (
	_ input.GetGameServicer         = (*service)(nil)
	_ input.GetPlayerServicer       = (*service)(nil)
	_ input.GetPropertyServicer     = (*service)(nil)
	_ input.GetGameLogServicer      = (*service)(nil)
	_ input.GetGameStateServicer    = (*service)(nil)
	_ input.GetTradeRequestServicer = (*service)(nil)
	_ Servicers                     = (*service)(nil)
)

// GetGameServicer implements Servicers.
func (service *service) GetGameServicer() input.GameServicer {
	return service.gameService
}

// GetPlayerServicer implements Servicers.
func (service *service) GetPlayerServicer() input.PlayerServicer {
	return service.playerService
}

// GetPropertyServicer implements Servicers.
func (service *service) GetPropertyServicer() input.PropertyServicer {
	return service.propertyService
}

// GetGameLogServicer implements Servicers.
func (service *service) GetGameLogServicer() input.GameLogServicer {
	return service.gameLogService
}

// GetGameStateServicer implements Servicers.
func (service *service) GetGameStateServicer() input.GameStateServicer {
	return service.gameStateService
}

// GetTradeRequestServicer implements Servicers.
func (service *service) GetTradeRequestServicer() input.TradeRequestServicer {
	return service.tradeRequestService
}

// NewServices is a function.
func NewServices(
	optioners ...optionServicer,
) *service {
	service := &service{
		gameService:         nil,
		playerService:       nil,
		propertyService:     nil,
		gameLogService:      nil,
		gameStateService:    nil,
		tradeRequestService: nil,
	}

	return service.WithOptioners(optioners...)
}

// WithOptioners is a function.
func (service *service) WithOptioners(
	optioners ...optionServicer,
) *service {
	newService := service.clone()
	for _, optioner := range optioners {
		optioner.apply(newService)
	}

	return newService
}

func (service *service) clone() *service {
	newService := service

	return newService
}

func (optionerFunc optionServicerFunc) apply(
	service *service,
) {
	optionerFunc(service)
}

// WithGameService sets the game service.
func WithGameService(
	configConfigger config.Configger,
	logRuntimeLogger log.RuntimeLogger,
	objectUUIDer object.UUIDer,
	traceTracer trace.Tracer,
	optioner ...gameServiceOptioner,
) optionServicer {
	return optionServicerFunc(func(
		service *service,
	) {
		service.gameService = NewGameService(
			configConfigger,
			logRuntimeLogger,
			objectUUIDer,
			traceTracer,
			optioner...,
		)
	})
}

// WithPlayerService sets the player service.
func WithPlayerService(
	configConfigger config.Configger,
	logRuntimeLogger log.RuntimeLogger,
	objectUUIDer object.UUIDer,
	traceTracer trace.Tracer,
	optioner ...playerServiceOptioner,
) optionServicer {
	return optionServicerFunc(func(
		service *service,
	) {
		service.playerService = NewPlayerService(
			configConfigger,
			logRuntimeLogger,
			objectUUIDer,
			traceTracer,
			optioner...,
		)
	})
}

// WithPropertyService sets the property service.
func WithPropertyService(
	configConfigger config.Configger,
	logRuntimeLogger log.RuntimeLogger,
	objectUUIDer object.UUIDer,
	traceTracer trace.Tracer,
	optioner ...propertyServiceOptioner,
) optionServicer {
	return optionServicerFunc(func(
		service *service,
	) {
		service.propertyService = NewPropertyService(
			configConfigger,
			logRuntimeLogger,
			objectUUIDer,
			traceTracer,
			optioner...,
		)
	})
}

// WithGameLogService sets the game log service.
func WithGameLogService(
	configConfigger config.Configger,
	logRuntimeLogger log.RuntimeLogger,
	objectUUIDer object.UUIDer,
	traceTracer trace.Tracer,
	optioner ...gameLogServiceOptioner,
) optionServicer {
	return optionServicerFunc(func(
		service *service,
	) {
		service.gameLogService = NewGameLogService(
			configConfigger,
			logRuntimeLogger,
			objectUUIDer,
			traceTracer,
			optioner...,
		)
	})
}

// WithGameStateService sets the game state service.
func WithGameStateService(
	configConfigger config.Configger,
	logRuntimeLogger log.RuntimeLogger,
	objectUUIDer object.UUIDer,
	traceTracer trace.Tracer,
	optioner ...gameStateServiceOptioner,
) optionServicer {
	return optionServicerFunc(func(
		service *service,
	) {
		service.gameStateService = NewGameStateService(
			configConfigger,
			logRuntimeLogger,
			objectUUIDer,
			traceTracer,
			optioner...,
		)
	})
}

// WithTradeRequestService sets the trade request service.
func WithTradeRequestService(
	configConfigger config.Configger,
	logRuntimeLogger log.RuntimeLogger,
	objectUUIDer object.UUIDer,
	traceTracer trace.Tracer,
	optioner ...tradeRequestServiceOptioner,
) optionServicer {
	return optionServicerFunc(func(
		service *service,
	) {
		service.tradeRequestService = NewTradeRequestService(
			configConfigger,
			logRuntimeLogger,
			objectUUIDer,
			traceTracer,
			optioner...,
		)
	})
}

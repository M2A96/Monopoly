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
		gameService input.GameServicer
	}
)

var (
	_ input.GetGameServicer = (*service)(nil)
	_ Servicers             = (*service)(nil)
)

// GetGameServicer implements Servicers.
func (service *service) GetGameServicer() input.GameServicer {
	return service.gameService
}

// NewServices is a function.
func NewServices(
	optioners ...optionServicer,
) *service {
	service := &service{
		gameService: nil,
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

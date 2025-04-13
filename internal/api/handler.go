package api

import (
	"github/M2A96/Monopoly.git/config"
	"github/M2A96/Monopoly.git/log"
	"github/M2A96/Monopoly.git/object"
	"github/M2A96/Monopoly.git/util"

	"github.com/labstack/echo/v4"

	"go.opentelemetry.io/otel/trace"
)

type (
	Handler interface {
		RegisterRoutes(e *echo.Echo)
	}

	handler[ServiceServicer any] struct {
		configConfigger  config.Configger
		logRuntimeLogger log.RuntimeLogger
		objectUUIDer     object.UUIDer
		traceTracer      trace.Tracer
		serviceServicer  ServiceServicer
		utilValidationer util.Validationer
	}

	handlerOptioner[ServiceServicer any] interface {
		apply(*handler[ServiceServicer])
	}

	handlerOptionerFunc[ServiceServicer any] func(*handler[ServiceServicer])
)

var (
	_ config.GetConfigger  = (*handler[any])(nil)
	_ log.GetRuntimeLogger = (*handler[any])(nil)
	_ object.GetUUIDer     = (*handler[any])(nil)
	_ util.GetTracer       = (*handler[any])(nil)
)

// NewHandler is function.
func NewHandler[ServiceServicer any](
	configConfigger config.Configger,
	logRuntimeLogger log.RuntimeLogger,
	objectUUIDer object.UUIDer,
	traceTracer trace.Tracer,
	optioners ...handlerOptioner[ServiceServicer],
) *handler[ServiceServicer] {
	var serviceServicer ServiceServicer
	handler := &handler[ServiceServicer]{
		configConfigger:  configConfigger,
		logRuntimeLogger: logRuntimeLogger,
		objectUUIDer:     objectUUIDer,
		traceTracer:      traceTracer,
		serviceServicer:  serviceServicer,
		utilValidationer: nil,
	}

	return handler.WithOptioners(optioners...)
}

// WithHandlerServicer is a function.
func WithHandlerServicer[ServiceServicer any](
	serviceServicer ServiceServicer,
) handlerOptioner[ServiceServicer] {
	return handlerOptionerFunc[ServiceServicer](func(
		handler *handler[ServiceServicer],
	) {
		handler.serviceServicer = serviceServicer
	})
}

// WithHandlerValidationer is a function.
func WithHandlerValidationer[ServiceServicer any](
	utilValidationer util.Validationer,
) handlerOptioner[ServiceServicer] {
	return handlerOptionerFunc[ServiceServicer](func(
		handler *handler[ServiceServicer],
	) {
		handler.utilValidationer = utilValidationer
	})
}

// GetConfigger is a function.
func (handler *handler[ServiceServicer]) GetConfigger() config.Configger {
	return handler.configConfigger
}

// GetRuntimeLogger is a function.
func (handler *handler[ServiceServicer]) GetRuntimeLogger() log.RuntimeLogger {
	return handler.logRuntimeLogger
}

// GetUUIDer is a function.
func (handler *handler[ServiceServicer]) GetUUIDer() object.UUIDer {
	return handler.objectUUIDer
}

// GetTracer is a function.
func (handler *handler[ServiceServicer]) GetTracer() trace.Tracer {
	return handler.traceTracer
}

// GetServicer is a function.
func (handler *handler[ServiceServicer]) GetServicer() ServiceServicer {
	return handler.serviceServicer
}

// GetValidationer is a function.
func (handler *handler[ServiceServicer]) GetValidationer() util.Validationer {
	return handler.utilValidationer
}

// WithOptioners is a function.
func (handler *handler[ServiceServicer]) WithOptioners(
	optioners ...handlerOptioner[ServiceServicer],
) *handler[ServiceServicer] {
	newHandler := handler.clone()
	for _, optioner := range optioners {
		optioner.apply(newHandler)
	}

	return newHandler
}

func (handler *handler[ServiceServicer]) clone() *handler[ServiceServicer] {
	newHandler := handler

	return newHandler
}

func (optionerFunc handlerOptionerFunc[ServiceServicer]) apply(
	handler *handler[ServiceServicer],
) {
	optionerFunc(handler)
}

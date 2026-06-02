package services

import (
	"context"
	"database/sql"
	"github/M2A96/Monopoly.git/config"
	"github/M2A96/Monopoly.git/internal/core/domain/bo"
	"github/M2A96/Monopoly.git/internal/core/ports/input"
	"github/M2A96/Monopoly.git/internal/core/ports/output"
	"github/M2A96/Monopoly.git/log"
	"github/M2A96/Monopoly.git/object"
	"github/M2A96/Monopoly.git/object/dao"
	"github/M2A96/Monopoly.git/util"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type (
	propertyService struct {
		configConfigger                config.Configger
		logRuntimeLogger               log.RuntimeLogger
		objectUUIDer                   object.UUIDer
		repositoryPropertyRepositorier output.PropertyRepositorier
		traceTracer                    trace.Tracer
	}

	propertyServiceOptioner interface {
		apply(*propertyService)
	}

	propertyServiceOptionerFunc func(*propertyService)
)

var (
	_ input.PropertyServicer         = (*propertyService)(nil)
	_ config.GetConfigger            = (*propertyService)(nil)
	_ log.GetRuntimeLogger           = (*propertyService)(nil)
	_ object.GetUUIDer               = (*propertyService)(nil)
	_ output.GetPropertyRepositorier = (*propertyService)(nil)
	_ util.GetTracer                 = (*propertyService)(nil)
)

func NewPropertyService(
	configConfigger config.Configger,
	logRuntimeLogger log.RuntimeLogger,
	objectUUIDer object.UUIDer,
	traceTracer trace.Tracer,
	optioners ...propertyServiceOptioner,
) *propertyService {
	propertyService := &propertyService{
		configConfigger:                configConfigger,
		logRuntimeLogger:               logRuntimeLogger,
		objectUUIDer:                   objectUUIDer,
		repositoryPropertyRepositorier: nil,
		traceTracer:                    traceTracer,
	}

	return propertyService.WithOptioners(optioners...)
}

func (service *propertyService) WithOptioners(
	optioners ...propertyServiceOptioner,
) *propertyService {
	for _, optioner := range optioners {
		optioner.apply(service)
	}

	return service
}

func WithPropertyServicePropertyRepositorier(
	repositoryPropertyRepositorier output.PropertyRepositorier,
) propertyServiceOptioner {
	return propertyServiceOptionerFunc(func(
		config *propertyService,
	) {
		config.repositoryPropertyRepositorier = repositoryPropertyRepositorier
	})
}

func (p *propertyService) GetTracer() trace.Tracer {
	return p.traceTracer
}

func (p *propertyService) GetPropertyRepositorier() output.PropertyRepositorier {
	return p.repositoryPropertyRepositorier
}

func (p *propertyService) GetUUIDer() object.UUIDer {
	return p.objectUUIDer
}

func (p *propertyService) GetRuntimeLogger() log.RuntimeLogger {
	return p.logRuntimeLogger
}

func (p *propertyService) GetConfigger() config.Configger {
	return p.configConfigger
}

func (service *propertyService) Get(
	ctx context.Context,
	propertyID uuid.UUID,
) (bo.Propertyer, error) {
	var traceSpan trace.Span

	ctx, traceSpan = service.GetTracer().Start(
		ctx,
		"Get",
		trace.WithSpanKind(trace.SpanKindInternal),
	)
	defer traceSpan.End()

	utilRuntimeContext := util.NewRuntimeContext(ctx)
	utilSpanContext := util.NewSpanContext(traceSpan)
	fields := map[string]any{
		"name":        "Get",
		"rt_ctx":      utilRuntimeContext,
		"sp_ctx":      utilSpanContext,
		"config":      service.GetConfigger(),
		"property_id": propertyID,
	}

	service.GetRuntimeLogger().
		WithFields(fields).
		Info(object.URIEmpty)

	daoProperty, err := service.GetPropertyRepositorier().Read(
		ctx,
		dao.NewCUDID(map[string]uuid.UUID{"id": propertyID}),
	)
	if err != nil {
		service.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error("Failed to retrieve property")
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, "Failed to retrieve property")

		return nil, err
	}

	boProperty := bo.NewProperty(
		daoProperty.GetCUDIDer().GetID()["id"],
		daoProperty.GetName(),
		daoProperty.GetColorGroup(),
		daoProperty.GetPrice(),
		daoProperty.GetHousePrice(),
		daoProperty.GetHotelPrice(),
		daoProperty.GetRent(),
		daoProperty.GetRentWithColorSet(),
		daoProperty.GetRentWith1House(),
		daoProperty.GetRentWith2Houses(),
		daoProperty.GetRentWith3Houses(),
		daoProperty.GetRentWith4Houses(),
		daoProperty.GetRentWithHotel(),
		daoProperty.GetMortgageValue(),
		daoProperty.GetOwnerID(),
		daoProperty.GetHouses(),
		daoProperty.GetHasHotel(),
		daoProperty.GetMortgaged(),
	)

	service.GetRuntimeLogger().
		WithFields(fields).
		WithField(object.URIFieldBOProperty, boProperty).
		Debug(object.URIEmpty)

	return boProperty, nil
}

func (service *propertyService) List(
	ctx context.Context,
	pagination dao.Paginationer,
	filter dao.PropertyFilter,
) ([]bo.Propertyer, dao.Cursorer, error) {
	var traceSpan trace.Span

	ctx, traceSpan = service.GetTracer().Start(
		ctx,
		"List",
		trace.WithSpanKind(trace.SpanKindInternal),
	)
	defer traceSpan.End()

	utilRuntimeContext := util.NewRuntimeContext(ctx)
	utilSpanContext := util.NewSpanContext(traceSpan)
	fields := map[string]any{
		"name":       "List",
		"rt_ctx":     utilRuntimeContext,
		"sp_ctx":     utilSpanContext,
		"config":     service.GetConfigger(),
		"pagination": pagination,
		"filter":     filter,
	}

	service.GetRuntimeLogger().
		WithFields(fields).
		Info(object.URIEmpty)

	daoProperties, daoCursorer, err := service.GetPropertyRepositorier().
		ReadList(ctx, pagination, filter)
	if err != nil {
		service.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error("Failed to list properties")
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, "Failed to list properties")

		return nil, nil, err
	}

	service.GetRuntimeLogger().
		WithFields(fields).
		WithField(object.URIFieldDAOProperties, daoProperties).
		WithField(object.URIFieldDAOCursorer, daoCursorer).
		Debug(object.URIEmpty)

	boProperties := make([]bo.Propertyer, 0, len(daoProperties))

	for key, daoProperty := range daoProperties {
		service.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldKey, key).
			WithField(object.URIFieldDAOProperty, daoProperty).
			Debug(object.URIEmpty)

		boProperties = append(boProperties, bo.NewProperty(
			daoProperty.GetCUDIDer().GetID()["id"],
			daoProperty.GetName(),
			daoProperty.GetColorGroup(),
			daoProperty.GetPrice(),
			daoProperty.GetHousePrice(),
			daoProperty.GetHotelPrice(),
			daoProperty.GetRent(),
			daoProperty.GetRentWithColorSet(),
			daoProperty.GetRentWith1House(),
			daoProperty.GetRentWith2Houses(),
			daoProperty.GetRentWith3Houses(),
			daoProperty.GetRentWith4Houses(),
			daoProperty.GetRentWithHotel(),
			daoProperty.GetMortgageValue(),
			daoProperty.GetOwnerID(),
			daoProperty.GetHouses(),
			daoProperty.GetHasHotel(),
			daoProperty.GetMortgaged(),
		))
	}

	service.GetRuntimeLogger().
		WithFields(fields).
		WithField(object.URIFieldBOProperties, boProperties).
		Debug(object.URIEmpty)

	return boProperties, daoCursorer, nil
}

func (service *propertyService) CreateProperty(
	ctx context.Context,
	boProperty bo.Propertyer,
) (uuid.UUID, error) {
	var traceSpan trace.Span

	ctx, traceSpan = service.GetTracer().Start(
		ctx,
		"CreateProperty",
		trace.WithSpanKind(trace.SpanKindInternal),
	)
	defer traceSpan.End()

	utilRuntimeContext := util.NewRuntimeContext(ctx)
	utilSpanContext := util.NewSpanContext(traceSpan)
	fields := map[string]any{
		"name":        "CreateProperty",
		"rt_ctx":      utilRuntimeContext,
		"sp_ctx":      utilSpanContext,
		"config":      service.GetConfigger(),
		"bo_property": boProperty,
	}

	service.GetRuntimeLogger().
		WithFields(fields).
		Info(object.URIEmpty)

	daoProperty := dao.NewProperty(
		uuid.Nil,
		boProperty.GetName(),
		boProperty.GetColorGroup(),
		boProperty.GetPrice(),
		boProperty.GetHousePrice(),
		boProperty.GetHotelPrice(),
		boProperty.GetRent(),
		boProperty.GetRentWithColorSet(),
		boProperty.GetRentWith1House(),
		boProperty.GetRentWith2Houses(),
		boProperty.GetRentWith3Houses(),
		boProperty.GetRentWith4Houses(),
		boProperty.GetRentWithHotel(),
		boProperty.GetMortgageValue(),
		boProperty.GetOwnerID(),
		boProperty.GetHouses(),
		boProperty.GetHasHotel(),
		boProperty.GetMortgaged(),
		time.Time{},
		time.Time{},
		sql.NullTime{
			Time:  time.Time{},
			Valid: false,
		},
	)

	service.GetRuntimeLogger().
		WithFields(fields).
		WithField(object.URIFieldDAOProperty, daoProperty).
		Debug(object.URIEmpty)

	id, err := service.GetPropertyRepositorier().Create(ctx, daoProperty)
	if err != nil {
		service.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error(object.ErrPropertyServiceCreate.Error())
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, object.ErrPropertyServiceCreate.Error())

		return uuid.Nil, err
	}

	service.GetRuntimeLogger().
		WithFields(fields).
		WithField(object.URIFieldID, id).
		Debug(object.URIEmpty)

	return id.GetID()["id"], nil
}

func (service *propertyService) clone() *propertyService {
	newService := service

	return newService
}

func (optionerFunc propertyServiceOptionerFunc) apply(
	service *propertyService,
) {
	optionerFunc(service)
}

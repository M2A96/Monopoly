// internal/core/domain/services/property_service.go
package services

import (
	"context"
	"database/sql"
	"errors"
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

// PropertyService implements the input.PropertyServicer interface

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

// NewPropertyService creates a new property service instance
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

// WithOptioners applies the provided optioners to the property service
func (service *propertyService) WithOptioners(
	optioners ...propertyServiceOptioner,
) *propertyService {
	for _, optioner := range optioners {
		optioner.apply(service)
	}

	return service
}

// WithPropertyServicePropertyRepositorier is a function.
func WithPropertyServicePropertyRepositorier(
	repositoryPropertyRepositorier output.PropertyRepositorier,
) propertyServiceOptioner {
	return propertyServiceOptionerFunc(func(
		config *propertyService,
	) {
		config.repositoryPropertyRepositorier = repositoryPropertyRepositorier
	})
}

// GetTracer implements util.GetTracer.
func (p *propertyService) GetTracer() trace.Tracer {
	return p.traceTracer
}

// GetPropertyRepositorier implements output.GetPropertyRepositorier.
func (p *propertyService) GetPropertyRepositorier() output.PropertyRepositorier {
	return p.repositoryPropertyRepositorier
}

// GetUUIDer implements object.GetUUIDer.
func (p *propertyService) GetUUIDer() object.UUIDer {
	return p.objectUUIDer
}

// GetRuntimeLogger implements log.GetRuntimeLogger.
func (p *propertyService) GetRuntimeLogger() log.RuntimeLogger {
	return p.logRuntimeLogger
}

// GetConfigger implements config.GetConfigger.
func (p *propertyService) GetConfigger() config.Configger {
	return p.configConfigger
}

// Get implements input.PropertyServicer.
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

// List implements input.PropertyServicer.
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

	boProperties := make([]bo.Propertyer, len(daoProperties))

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

// CreateProperty implements input.PropertyServicer.
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

// UpdateProperty implements input.PropertyServicer.
func (service *propertyService) UpdateProperty(
	ctx context.Context,
	propertyID int,
	boProperty bo.Propertyer,
) error {
	var traceSpan trace.Span

	ctx, traceSpan = service.GetTracer().Start(
		ctx,
		"UpdateProperty",
		trace.WithSpanKind(trace.SpanKindInternal),
	)
	defer traceSpan.End()

	utilRuntimeContext := util.NewRuntimeContext(ctx)
	utilSpanContext := util.NewSpanContext(traceSpan)
	fields := map[string]any{
		"name":        "UpdateProperty",
		"rt_ctx":      utilRuntimeContext,
		"sp_ctx":      utilSpanContext,
		"config":      service.GetConfigger(),
		"property_id": propertyID,
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

	if _, err := service.GetPropertyRepositorier().
		Update(
			ctx,
			daoProperty,
		); err != nil {
		service.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error("Failed to update property")
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, "Failed to update property")

		return err
	}

	service.GetRuntimeLogger().
		WithFields(fields).
		Debug(object.URIEmpty)

	return nil
}

// AddHouse implements input.PropertyServicer.
func (service *propertyService) AddHouse(context.Context, int) error {
	panic("unimplemented")
}

// AddHotel implements input.PropertyServicer.
func (service *propertyService) AddHotel(
	ctx context.Context,
	propertyID uuid.UUID,
) error {
	var traceSpan trace.Span

	ctx, traceSpan = service.GetTracer().Start(
		ctx,
		"AddHotel",
		trace.WithSpanKind(trace.SpanKindInternal),
	)
	defer traceSpan.End()

	utilRuntimeContext := util.NewRuntimeContext(ctx)
	utilSpanContext := util.NewSpanContext(traceSpan)
	fields := map[string]any{
		"name":        "AddHotel",
		"rt_ctx":      utilRuntimeContext,
		"sp_ctx":      utilSpanContext,
		"config":      service.GetConfigger(),
		"property_id": propertyID,
	}

	service.GetRuntimeLogger().
		WithFields(fields).
		Info(object.URIEmpty)

	// Get the property first
	daoProperty, err := service.GetPropertyRepositorier().
		Read(
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

		return err
	}

	// Check if the property already has a hotel
	if daoProperty.GetHasHotel() {
		service.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error(object.ErrPropertyServiceAlreadyHasHotel.Error())
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, object.ErrPropertyServiceAlreadyHasHotel.Error())

		return object.ErrPropertyServiceAlreadyHasHotel
	}

	// Check if the property has 4 houses (required to add a hotel)
	if daoProperty.GetHouses() < 4 {
		err := errors.New("property must have 4 houses before adding a hotel")
		service.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error("Property must have 4 houses before adding a hotel")
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, "Property must have 4 houses before adding a hotel")

		return err
	}

	// Update the property to have a hotel and remove the houses
	daoProperty = dao.NewProperty(
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
		0,    // Reset houses to 0
		true, // Set hasHotel to true
		daoProperty.GetMortgaged(),
		daoProperty.GetCUDer().GetCreatedAt(),
		daoProperty.GetCUDer().GetUpdatedAt(),
		daoProperty.GetCUDer().GetDeletedAt(),
	)

	service.GetRuntimeLogger().
		WithFields(fields).
		WithField(object.URIFieldDAOProperty, daoProperty).
		Debug(object.URIEmpty)

	if _, err := service.GetPropertyRepositorier().
		Update(
			ctx,
			daoProperty,
		); err != nil {
		service.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error(object.ErrPropertyServiceUpdate.Error())
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, object.ErrPropertyServiceUpdate.Error())

		return err
	}

	service.GetRuntimeLogger().
		WithFields(fields).
		Debug(object.URIEmpty)

	return nil
}

// DeleteProperty implements input.PropertyServicer.
func (service *propertyService) DeleteProperty(
	ctx context.Context,
	propertyID uuid.UUID,
) error {
	var traceSpan trace.Span

	ctx, traceSpan = service.GetTracer().Start(
		ctx,
		"DeleteProperty",
		trace.WithSpanKind(trace.SpanKindInternal),
	)
	defer traceSpan.End()

	utilRuntimeContext := util.NewRuntimeContext(ctx)
	utilSpanContext := util.NewSpanContext(traceSpan)
	fields := map[string]any{
		"name":        "DeleteProperty",
		"rt_ctx":      utilRuntimeContext,
		"sp_ctx":      utilSpanContext,
		"config":      service.GetConfigger(),
		"property_id": propertyID,
	}

	service.GetRuntimeLogger().
		WithFields(fields).
		Info(object.URIEmpty)

	if _, err := service.GetPropertyRepositorier().
		Delete(
			ctx,
			dao.NewCUDID(map[string]uuid.UUID{"id": propertyID}),
		); err != nil {
		service.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error("Failed to delete property")
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, "Failed to delete property")

		return err
	}

	service.GetRuntimeLogger().
		WithFields(fields).
		Debug(object.URIEmpty)

	return nil
}

// BuyProperty implements input.PropertyServicer.
func (service *propertyService) BuyProperty(
	ctx context.Context,
	propertyID uuid.UUID,
	playerID uuid.UUID,
) error {
	var traceSpan trace.Span

	ctx, traceSpan = service.GetTracer().Start(
		ctx,
		"BuyProperty",
		trace.WithSpanKind(trace.SpanKindInternal),
	)
	defer traceSpan.End()

	utilRuntimeContext := util.NewRuntimeContext(ctx)
	utilSpanContext := util.NewSpanContext(traceSpan)
	fields := map[string]any{
		"name":        "BuyProperty",
		"rt_ctx":      utilRuntimeContext,
		"sp_ctx":      utilSpanContext,
		"config":      service.GetConfigger(),
		"property_id": propertyID,
		"player_id":   playerID,
	}

	service.GetRuntimeLogger().
		WithFields(fields).
		Info(object.URIEmpty)

	// Get the property first
	daoProperty, err := service.GetPropertyRepositorier().
		Read(
			ctx,
			dao.NewCUDID(map[string]uuid.UUID{"id": propertyID}),
		)
	if err != nil {
		service.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error(object.ErrPropertyServiceRead.Error())
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, object.ErrPropertyServiceRead.Error())

		return err
	}

	// Update the property with the new owner
	daoProperty = dao.NewProperty(
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
		playerID,
		daoProperty.GetHouses(),
		daoProperty.GetHasHotel(),
		daoProperty.GetMortgaged(),
		daoProperty.GetCUDer().GetCreatedAt(),
		daoProperty.GetCUDer().GetUpdatedAt(),
		daoProperty.GetCUDer().GetDeletedAt(),
	)

	service.GetRuntimeLogger().
		WithFields(fields).
		WithField(object.URIFieldDAOProperty, daoProperty).
		Debug(object.URIEmpty)

	if _, err := service.GetPropertyRepositorier().
		Update(
			ctx,
			daoProperty,
		); err != nil {
		service.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error(object.ErrPropertyServiceUpdateOwnerShip.Error())
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, object.ErrPropertyServiceUpdateOwnerShip.Error())

		return err
	}

	service.GetRuntimeLogger().
		WithFields(fields).
		Debug(object.URIEmpty)

	return nil
}

// MortgageProperty implements input.PropertyServicer.
func (service *propertyService) MortgageProperty(
	ctx context.Context,
	propertyID uuid.UUID,
	mortgaged bool,
) error {
	var traceSpan trace.Span

	ctx, traceSpan = service.GetTracer().Start(
		ctx,
		"MortgageProperty",
		trace.WithSpanKind(trace.SpanKindInternal),
	)
	defer traceSpan.End()

	utilRuntimeContext := util.NewRuntimeContext(ctx)
	utilSpanContext := util.NewSpanContext(traceSpan)
	fields := map[string]any{
		"name":        "MortgageProperty",
		"rt_ctx":      utilRuntimeContext,
		"sp_ctx":      utilSpanContext,
		"config":      service.GetConfigger(),
		"property_id": propertyID,
		"mortgaged":   mortgaged,
	}

	service.GetRuntimeLogger().
		WithFields(fields).
		Info(object.URIEmpty)

	// Get the property first
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

		return err
	}

	// Update the property with the new mortgage status
	daoProperty = dao.NewProperty(
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
		mortgaged,
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

	if _, err := service.GetPropertyRepositorier().
		Update(
			ctx,
			daoProperty,
		); err != nil {
		service.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error("Failed to update property mortgage status")
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, "Failed to update property mortgage status")

		return err
	}

	service.GetRuntimeLogger().
		WithFields(fields).
		Debug(object.URIEmpty)

	return nil
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

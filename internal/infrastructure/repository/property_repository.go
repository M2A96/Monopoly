package repository

import (
	"context"
	"database/sql"
	"fmt"
	"github/M2A96/Monopoly.git/config"
	"github/M2A96/Monopoly.git/internal/core/ports/output"
	"github/M2A96/Monopoly.git/log"
	"github/M2A96/Monopoly.git/object"
	"github/M2A96/Monopoly.git/object/dao"
	"github/M2A96/Monopoly.git/util"
	"time"

	"github.com/samber/lo"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
)

type (
	propertyRepository struct {
		repository
	}

	propertyRepositoryOptioner = repositoryOptioner
)

// Ensure propertyRepository implements output.PropertyRepositorier
var _ output.PropertyRepositorier = (*propertyRepository)(nil)

// NewPropertyRepository creates a new property repository instance
func NewPropertyRepository(
	configConfigger config.Configger,
	logRuntimeLogger log.RuntimeLogger,
	objectUUIDer object.UUIDer,
	traceTracer trace.Tracer,
	optioners ...propertyRepositoryOptioner,
) *propertyRepository {
	return &propertyRepository{
		repository: *NewRepository(
			configConfigger,
			logRuntimeLogger,
			objectUUIDer,
			traceTracer,
			optioners...,
		),
	}
}

// WithPropertyRepositoryTimer adds a timer to the property repository
func WithPropertyRepositoryTimer(
	objectTimer object.Timer,
) propertyRepositoryOptioner {
	return WithRepositoryTimer(objectTimer)
}

// WithPropertyRepositoryDB adds a database connection to the property repository
func WithPropertyRepositoryDB(
	gormDB *gorm.DB,
) propertyRepositoryOptioner {
	return WithRepositoryDB(gormDB, object.URITableProperty)
}

// Create implements the Create method of the PropertyRepositorier interface
func (repository *propertyRepository) Create(
	ctx context.Context,
	daoProperty dao.Propertyer,
) (dao.CUDIDer, error) {
	var traceSpan trace.Span

	ctx, traceSpan = repository.GetTracer().Start(
		ctx,
		"Create",
		trace.WithSpanKind(trace.SpanKindClient),
	)
	defer traceSpan.End()

	utilRuntimeContext := util.NewRuntimeContext(ctx)
	utilSpanContext := util.NewSpanContext(traceSpan)
	fields := map[string]any{
		"name":         "Create",
		"rt_ctx":       utilRuntimeContext,
		"sp_ctx":       utilSpanContext,
		"config":       repository.GetConfigger(),
		"dao_property": daoProperty,
	}

	repository.GetRuntimeLogger().
		WithFields(fields).
		Info(object.URIEmpty)

	id, err := repository.GetUUIDer().NewRandom()
	if err != nil {
		repository.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error(object.ErrUUIDerNewRandom.Error())
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, object.ErrUUIDerNewRandom.Error())

		return nil, err
	}

	repository.GetRuntimeLogger().
		WithFields(fields).
		WithField(object.URIFieldID, id).
		Debug(object.URIEmpty)

	nowUTC := repository.GetTimer().NowUTC()

	repository.GetRuntimeLogger().
		WithFields(fields).
		WithField(object.URIFieldNowUTC, nowUTC).
		Debug(object.URIEmpty)

	newProperty := dao.NewProperty(
		id,
		daoProperty.GetID(),
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
		nowUTC,
		nowUTC,
		sql.NullTime{
			Time:  time.Time{},
			Valid: false,
		},
	)

	repository.GetRuntimeLogger().
		WithFields(fields).
		WithField("dao_property", newProperty).
		Debug(object.URIEmpty)

	gormDB := repository.GetDB().
		WithContext(ctx).
		Create(newProperty.GetMap())
	if err = gormDB.Error; err != nil {
		repository.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error("Failed to create property")
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, "Failed to create property")

		return nil, err
	}

	return newProperty.GetCUDIDer(), nil
}

// Read implements the Read method of the PropertyRepositorier interface
func (repository *propertyRepository) Read(
	ctx context.Context,
	id dao.CUDIDer,
) (dao.Propertyer, error) {
	var traceSpan trace.Span

	ctx, traceSpan = repository.GetTracer().Start(
		ctx,
		"Read",
		trace.WithSpanKind(trace.SpanKindClient),
	)
	defer traceSpan.End()

	utilRuntimeContext := util.NewRuntimeContext(ctx)
	utilSpanContext := util.NewSpanContext(traceSpan)
	fields := map[string]any{
		"name":   "Read",
		"rt_ctx": utilRuntimeContext,
		"sp_ctx": utilSpanContext,
		"config": repository.GetConfigger(),
		"id":     id.GetID(),
	}

	repository.GetRuntimeLogger().
		WithFields(fields).
		Info(object.URIEmpty)

	result := map[string]any{}

	gormDB := repository.GetDB().
		WithContext(ctx).
		Where(lo.Assign(
			id.GetMap(),
			map[string]any{
				"deleted_at": nil,
			},
		)).
		Select(fmt.Sprintf(`%s.*`, object.URITableProperty)).
		Find(result)
	if err := gormDB.Error; err != nil {
		repository.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error("Failed to read property")
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, "Failed to read property")

		return nil, err
	}

	if gormDB.RowsAffected == 0 {
		err := fmt.Errorf("property not found")
		repository.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error(err.Error())
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, err.Error())

		return nil, err
	}

	repository.GetRuntimeLogger().
		WithFields(fields).
		WithField(object.URIFieldResult, result).
		Debug(object.URIEmpty)

	newProperty, err := dao.NewPropertyFromMap(repository.GetUUIDer(), result)
	if err != nil {
		repository.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error("Failed to create property from map")
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, "Failed to create property from map")

		return nil, err
	}
	repository.GetRuntimeLogger().
		WithFields(fields).
		WithField("dao_property", newProperty).
		Debug(object.URIEmpty)

	return newProperty, nil
}

// ReadList implements the ReadList method of the PropertyRepositorier interface
func (repository *propertyRepository) ReadList(
	ctx context.Context,
	pagination dao.Paginationer,
	filter dao.PropertyFilter,
) ([]dao.Propertyer, dao.Cursorer, error) {
	var traceSpan trace.Span

	ctx, traceSpan = repository.GetTracer().Start(
		ctx,
		"ReadList",
		trace.WithSpanKind(trace.SpanKindClient),
	)
	defer traceSpan.End()

	utilRuntimeContext := util.NewRuntimeContext(ctx)
	utilSpanContext := util.NewSpanContext(traceSpan)
	fields := map[string]any{
		"name":       "ReadList",
		"rt_ctx":     utilRuntimeContext,
		"sp_ctx":     utilSpanContext,
		"config":     repository.GetConfigger(),
		"pagination": pagination,
		"filter":     filter,
	}

	repository.GetRuntimeLogger().
		WithFields(fields).
		Info(object.URIEmpty)

	result := make([]map[string]any, 0, pagination.GetLimit()+1)

	gormDB := repository.GetDB().
		WithContext(ctx).
		Scopes(
			filter.Filter,
			pagination.Pagination(object.URITableProperty),
		).
		Where(map[string]any{
			"deleted_at": nil,
		}).
		Select(fmt.Sprintf(`%s.*`, object.URITableProperty)).
		Find(&result)
	if err := gormDB.Error; err != nil {
		repository.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error("Failed to read property list")
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, "Failed to read property list")

		return nil, nil, err
	}

	repository.GetRuntimeLogger().
		WithFields(fields).
		WithField(object.URIFieldResult, result).
		Debug(object.URIEmpty)

	daoPropertyers := make([]dao.Propertyer, 0, pagination.GetLimit())

	for key, value := range result {
		repository.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldKey, key).
			WithField(object.URIFieldValue, value).
			Debug(object.URIEmpty)

		if uint32(key) == pagination.GetLimit() {
			repository.GetRuntimeLogger().
				WithFields(fields).
				Debug(`uint32(key) == pagination.GetLimit()`)

			break
		}

		daoPropertyer, err := dao.NewPropertyFromMap(repository.GetUUIDer(), value)
		if err != nil {
			repository.GetRuntimeLogger().
				WithFields(fields).
				WithField(object.URIFieldError, err).
				Error("Failed to create property from map")
			traceSpan.RecordError(err)
			traceSpan.SetStatus(codes.Error, "Failed to create property from map")

			return nil, nil, err
		}

		repository.GetRuntimeLogger().
			WithFields(fields).
			WithField("dao_property", daoPropertyer).
			Debug(object.URIEmpty)

		daoPropertyers = append(daoPropertyers, daoPropertyer)
	}

	repository.GetRuntimeLogger().
		WithFields(fields).
		WithField("dao_properties", daoPropertyers).
		Debug(object.URIEmpty)

	var daoCursorer dao.Cursorer

	if pagination.GetLimit() < uint32(len(result)) {
		repository.GetRuntimeLogger().
			WithFields(fields).
			Debug(`pagination.GetLimit() < uint32(len(result))`)

		daoCursorer = dao.NewCursor(
			pagination.GetCursorer().GetOffset() + pagination.GetLimit(),
		)
	}

	repository.GetRuntimeLogger().
		WithFields(fields).
		WithField(object.URIFieldDAOCursorer, daoCursorer).
		Debug(object.URIEmpty)

	return daoPropertyers, daoCursorer, nil
}

// Update implements the Update method of the PropertyRepositorier interface
func (repository *propertyRepository) Update(
	ctx context.Context,
	property dao.Propertyer,
) (time.Time, error) {
	var traceSpan trace.Span

	ctx, traceSpan = repository.GetTracer().Start(
		ctx,
		"Update",
		trace.WithSpanKind(trace.SpanKindClient),
	)
	defer traceSpan.End()

	utilRuntimeContext := util.NewRuntimeContext(ctx)
	utilSpanContext := util.NewSpanContext(traceSpan)
	fields := map[string]any{
		"name":         "Update",
		"rt_ctx":       utilRuntimeContext,
		"sp_ctx":       utilSpanContext,
		"config":       repository.GetConfigger(),
		"property_id":  property.GetCUDIDer().GetID(),
		"dao_property": property,
	}

	repository.GetRuntimeLogger().
		WithFields(fields).
		Info(object.URIEmpty)

	nowUTC := repository.GetTimer().NowUTC()

	repository.GetRuntimeLogger().
		WithFields(fields).
		WithField(object.URIFieldNowUTC, nowUTC).
		Debug(object.URIEmpty)

	daoProperty := dao.NewProperty(
		property.GetCUDIDer().GetID()["id"],
		property.GetID(),
		property.GetName(),
		property.GetColorGroup(),
		property.GetPrice(),
		property.GetHousePrice(),
		property.GetHotelPrice(),
		property.GetRent(),
		property.GetRentWithColorSet(),
		property.GetRentWith1House(),
		property.GetRentWith2Houses(),
		property.GetRentWith3Houses(),
		property.GetRentWith4Houses(),
		property.GetRentWithHotel(),
		property.GetMortgageValue(),
		property.GetOwnerID(),
		property.GetHouses(),
		property.GetHasHotel(),
		property.GetMortgaged(),
		property.GetCUDer().GetCreatedAt(),
		nowUTC,
		sql.NullTime{
			Time:  time.Time{},
			Valid: false,
		},
	)

	repository.GetRuntimeLogger().
		WithFields(fields).
		WithField("dao_property", daoProperty).
		Debug(object.URIEmpty)

	// Update the property in the database
	gormDB := repository.GetDB().
		WithContext(ctx).
		Where(lo.Assign(
			daoProperty.GetCUDIDer().GetMap(),
			map[string]any{
				"deleted_at": nil,
			},
		)).
		Updates(daoProperty.GetMap())
	if err := gormDB.Error; err != nil {
		repository.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error("Failed to update property")
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, "Failed to update property")

		return time.Time{}, err
	}

	return nowUTC, nil
}

// Delete implements the Delete method of the PropertyRepositorier interface
func (repository *propertyRepository) Delete(
	ctx context.Context,
	id dao.CUDIDer,
) (time.Time, error) {
	var traceSpan trace.Span

	ctx, traceSpan = repository.GetTracer().Start(
		ctx,
		"Delete",
		trace.WithSpanKind(trace.SpanKindClient),
	)
	defer traceSpan.End()

	utilRuntimeContext := util.NewRuntimeContext(ctx)
	utilSpanContext := util.NewSpanContext(traceSpan)
	fields := map[string]any{
		"name":   "Delete",
		"rt_ctx": utilRuntimeContext,
		"sp_ctx": utilSpanContext,
		"config": repository.GetConfigger(),
		"id":     id.GetID(),
	}

	repository.GetRuntimeLogger().
		WithFields(fields).
		Info(object.URIEmpty)

	nowUTC := repository.GetTimer().NowUTC()

	repository.GetRuntimeLogger().
		WithFields(fields).
		WithField(object.URIFieldNowUTC, nowUTC).
		Debug(object.URIEmpty)

	// Delete the property from the database
	gormDB := repository.GetDB().
		WithContext(ctx).
		Where("id = ?", id.GetID()).
		Delete(nil)
	if err := gormDB.Error; err != nil {
		repository.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error("Failed to delete property")
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, "Failed to delete property")

		return time.Time{}, err
	}

	return nowUTC, nil
}

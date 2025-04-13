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
	tradeRequestRepository struct {
		repository
	}

	tradeRequestRepositoryOptioner = repositoryOptioner
)

// Ensure tradeRequestRepository implements output.TradeRequestRepositorier
var _ output.TradeRequestRepositorier = (*tradeRequestRepository)(nil)

// NewTradeRequestRepository creates a new trade request repository instance
func NewTradeRequestRepository(
	configConfigger config.Configger,
	logRuntimeLogger log.RuntimeLogger,
	objectUUIDer object.UUIDer,
	traceTracer trace.Tracer,
	optioners ...tradeRequestRepositoryOptioner,
) *tradeRequestRepository {
	return &tradeRequestRepository{
		repository: *NewRepository(
			configConfigger,
			logRuntimeLogger,
			objectUUIDer,
			traceTracer,
			optioners...,
		),
	}
}

// WithTradeRequestRepositoryTimer adds a timer to the trade request repository
func WithTradeRequestRepositoryTimer(
	objectTimer object.Timer,
) tradeRequestRepositoryOptioner {
	return WithRepositoryTimer(objectTimer)
}

// WithTradeRequestRepositoryDB adds a database connection to the trade request repository
func WithTradeRequestRepositoryDB(
	gormDB *gorm.DB,
) tradeRequestRepositoryOptioner {
	return WithRepositoryDB(gormDB, object.URITableTradeRequest)
}

// Create implements the Create method of the TradeRequestRepositorier interface
func (repository *tradeRequestRepository) Create(
	ctx context.Context,
	daoTradeRequest dao.TradeRequester,
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
		"name":              "Create",
		"rt_ctx":            utilRuntimeContext,
		"sp_ctx":            utilSpanContext,
		"config":            repository.GetConfigger(),
		"dao_trade_request": daoTradeRequest,
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

	newTradeRequest := dao.NewTradeRequest(
		id,
		daoTradeRequest.GetSenderID(),
		daoTradeRequest.GetSenderID(),
		daoTradeRequest.GetReceiverID(),
		daoTradeRequest.GetOfferingMoney(),
		daoTradeRequest.GetRequestingMoney(),
		daoTradeRequest.GetOfferingProperties(),
		daoTradeRequest.GetRequestingProperties(),
		daoTradeRequest.GetStatus(),
		nowUTC,
		nowUTC,
		sql.NullTime{
			Time:  time.Time{},
			Valid: false,
		},
	)

	repository.GetRuntimeLogger().
		WithFields(fields).
		WithField("dao_trade_request", newTradeRequest).
		Debug(object.URIEmpty)

	gormDB := repository.GetDB().
		WithContext(ctx).
		Create(newTradeRequest.GetMap())
	if err = gormDB.Error; err != nil {
		repository.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error("Failed to create trade request")
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, "Failed to create trade request")

		return nil, err
	}

	return newTradeRequest.GetCUDIDer(), nil
}

// Read implements the Read method of the TradeRequestRepositorier interface
func (repository *tradeRequestRepository) Read(
	ctx context.Context,
	id dao.CUDIDer,
) (dao.TradeRequester, error) {
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
		Select(fmt.Sprintf(`%s.*`, object.URITableTradeRequest)).
		Find(result)
	if err := gormDB.Error; err != nil {
		repository.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error("Failed to read trade request")
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, "Failed to read trade request")

		return nil, err
	}

	if len(result) == 0 {
		repository.GetRuntimeLogger().
			WithFields(fields).
			Error("Trade request not found")
		traceSpan.SetStatus(codes.Error, "Trade request not found")

		return nil, fmt.Errorf("trade request not found")
	}

	tradeRequest, err := dao.NewTradeRequestFromMap(
		repository.GetUUIDer(),
		result,
	)
	if err != nil {
		repository.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error("Failed to create trade request from map")
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, "Failed to create trade request from map")

		return nil, err
	}

	return tradeRequest, nil
}

// ReadList implements the ReadList method of the TradeRequestRepositorier interface
func (repository *tradeRequestRepository) ReadList(
	ctx context.Context,
	pagination dao.Paginationer,
	filter dao.TradeFilter,
) ([]dao.TradeRequester, dao.Cursorer, error) {
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
			pagination.Pagination(object.URITableTradeRequest),
		).
		Where(map[string]any{
			"deleted_at": nil,
		}).
		Select(fmt.Sprintf(`%s.*`, object.URITableTradeRequest)).
		Find(&result)
	if err := gormDB.Error; err != nil {
		repository.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error("Failed to read trade request list")
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, "Failed to read trade request list")

		return nil, nil, err
	}

	repository.GetRuntimeLogger().
		WithFields(fields).
		WithField(object.URIFieldResult, result).
		Debug(object.URIEmpty)

	daoTradeRequests := make([]dao.TradeRequester, 0, pagination.GetLimit())

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

		daoTradeRequest, err := dao.NewTradeRequestFromMap(repository.GetUUIDer(), value)
		if err != nil {
			repository.GetRuntimeLogger().
				WithFields(fields).
				WithField(object.URIFieldError, err).
				Error("Failed to create trade request from map")
			traceSpan.RecordError(err)
			traceSpan.SetStatus(codes.Error, "Failed to create trade request from map")

			return nil, nil, err
		}

		repository.GetRuntimeLogger().
			WithFields(fields).
			WithField("dao_trade_request", daoTradeRequest).
			Debug(object.URIEmpty)

		daoTradeRequests = append(daoTradeRequests, daoTradeRequest)
	}

	repository.GetRuntimeLogger().
		WithFields(fields).
		WithField("dao_trade_requests", daoTradeRequests).
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

	return daoTradeRequests, daoCursorer, nil
}

// Update implements the Update method of the TradeRequestRepositorier interface
func (repository *tradeRequestRepository) Update(
	ctx context.Context,
	tradeRequest dao.TradeRequester,
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
		"name":              "Update",
		"rt_ctx":            utilRuntimeContext,
		"sp_ctx":            utilSpanContext,
		"config":            repository.GetConfigger(),
		"dao_trade_request": tradeRequest,
	}

	repository.GetRuntimeLogger().
		WithFields(fields).
		Info(object.URIEmpty)

	nowUTC := repository.GetTimer().NowUTC()

	repository.GetRuntimeLogger().
		WithFields(fields).
		WithField(object.URIFieldNowUTC, nowUTC).
		Debug(object.URIEmpty)

	daoTradeRequest := dao.NewTradeRequest(
		tradeRequest.GetCUDIDer().GetID()["id"],
		tradeRequest.GetSenderID(),
		tradeRequest.GetSenderID(),
		tradeRequest.GetReceiverID(),
		tradeRequest.GetOfferingMoney(),
		tradeRequest.GetRequestingMoney(),
		tradeRequest.GetOfferingProperties(),
		tradeRequest.GetRequestingProperties(),
		tradeRequest.GetStatus(),
		nowUTC,
		nowUTC,
		sql.NullTime{
			Time:  time.Time{},
			Valid: false,
		},
	)

	gormDB := repository.GetDB().
		WithContext(ctx).
		Where(lo.Assign(
			daoTradeRequest.GetCUDIDer().GetMap(),
			map[string]any{
				"deleted_at": nil,
			},
		)).
		Updates(lo.Assign(
			daoTradeRequest.GetMap(),
			map[string]any{
				"updated_at": nowUTC,
			},
		))
	if err := gormDB.Error; err != nil {
		repository.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error("Failed to update trade request")
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, "Failed to update trade request")

		return time.Time{}, err
	}

	if gormDB.RowsAffected == 0 {
		repository.GetRuntimeLogger().
			WithFields(fields).
			Error("Trade request not found for update")
		traceSpan.SetStatus(codes.Error, "Trade request not found for update")

		return time.Time{}, fmt.Errorf("trade request not found for update")
	}

	return nowUTC, nil
}

// Delete implements the Delete method of the TradeRequestRepositorier interface
func (repository *tradeRequestRepository) Delete(
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

	gormDB := repository.GetDB().
		WithContext(ctx).
		Where(lo.Assign(
			id.GetMap(),
			map[string]any{
				"deleted_at": nil,
			},
		)).
		Updates(map[string]any{
			"deleted_at": nowUTC,
		})
	if err := gormDB.Error; err != nil {
		repository.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error("Failed to delete trade request")
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, "Failed to delete trade request")

		return time.Time{}, err
	}

	if gormDB.RowsAffected == 0 {
		repository.GetRuntimeLogger().
			WithFields(fields).
			Error("Trade request not found for delete")
		traceSpan.SetStatus(codes.Error, "Trade request not found for delete")

		return time.Time{}, fmt.Errorf("trade request not found for delete")
	}

	return nowUTC, nil
}

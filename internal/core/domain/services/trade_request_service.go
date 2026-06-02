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
	tradeRequestService struct {
		configConfigger                    config.Configger
		logRuntimeLogger                   log.RuntimeLogger
		objectUUIDer                       object.UUIDer
		repositoryTradeRequestRepositorier output.TradeRequestRepositorier
		objectTimer                        object.Timer
		traceTracer                        trace.Tracer
	}

	tradeRequestServiceOptioner interface {
		apply(*tradeRequestService)
	}

	tradeRequestServiceOptionerFunc func(*tradeRequestService)
)

var (
	_ input.TradeRequestServicer         = (*tradeRequestService)(nil)
	_ config.GetConfigger                = (*tradeRequestService)(nil)
	_ log.GetRuntimeLogger               = (*tradeRequestService)(nil)
	_ object.GetUUIDer                   = (*tradeRequestService)(nil)
	_ output.GetTradeRequestRepositorier = (*tradeRequestService)(nil)
	_ object.GetTimer                    = (*tradeRequestService)(nil)
	_ util.GetTracer                     = (*tradeRequestService)(nil)
)

func NewTradeRequestService(
	configConfigger config.Configger,
	logRuntimeLogger log.RuntimeLogger,
	objectUUIDer object.UUIDer,
	traceTracer trace.Tracer,
	optioners ...tradeRequestServiceOptioner,
) *tradeRequestService {
	tradeRequestService := &tradeRequestService{
		configConfigger:                    configConfigger,
		logRuntimeLogger:                   logRuntimeLogger,
		objectUUIDer:                       objectUUIDer,
		repositoryTradeRequestRepositorier: nil,
		objectTimer:                        nil,
		traceTracer:                        traceTracer,
	}

	return tradeRequestService.WithOptioners(optioners...)
}

func WithTradeRequestServiceTradeRequestRepositorier(
	repositoryTradeRequestRepositorier output.TradeRequestRepositorier,
) tradeRequestServiceOptioner {
	return tradeRequestServiceOptionerFunc(func(
		service *tradeRequestService,
	) {
		service.repositoryTradeRequestRepositorier = repositoryTradeRequestRepositorier
	})
}

func WithTradeRequestServiceTimer(
	objectTimer object.Timer,
) tradeRequestServiceOptioner {
	return tradeRequestServiceOptionerFunc(func(
		service *tradeRequestService,
	) {
		service.objectTimer = objectTimer
	})
}

func (service *tradeRequestService) WithOptioners(
	optioners ...tradeRequestServiceOptioner,
) *tradeRequestService {
	for _, optioner := range optioners {
		optioner.apply(service)
	}

	return service
}

func (service *tradeRequestService) GetTracer() trace.Tracer {
	return service.traceTracer
}

func (service *tradeRequestService) GetTradeRequestRepositorier() output.TradeRequestRepositorier {
	return service.repositoryTradeRequestRepositorier
}

func (service *tradeRequestService) GetUUIDer() object.UUIDer {
	return service.objectUUIDer
}

func (service *tradeRequestService) GetRuntimeLogger() log.RuntimeLogger {
	return service.logRuntimeLogger
}

func (service *tradeRequestService) GetConfigger() config.Configger {
	return service.configConfigger
}

func (service *tradeRequestService) GetTimer() object.Timer {
	return service.objectTimer
}

func (service *tradeRequestService) Get(
	ctx context.Context,
	id uuid.UUID,
) (bo.TradeRequester, error) {
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
		"name":   "Get",
		"rt_ctx": utilRuntimeContext,
		"sp_ctx": utilSpanContext,
		"config": service.GetConfigger(),
		"id":     id,
	}

	service.GetRuntimeLogger().
		WithFields(fields).
		Info(object.URIEmpty)

	cudIDer := dao.NewCUDID(map[string]uuid.UUID{"id": id})

	service.GetRuntimeLogger().
		WithFields(fields).
		WithField(object.URIFieldCUDIDer, cudIDer).
		Debug(object.URIEmpty)

	daoTradeRequester, err := service.GetTradeRequestRepositorier().Read(ctx, cudIDer)
	if err != nil {
		service.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error("Failed to read trade request")
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, "Failed to read trade request")

		return nil, err
	}

	service.GetRuntimeLogger().
		WithFields(fields).
		WithField(object.URIFieldDAOTradeRequester, daoTradeRequester).
		Debug(object.URIEmpty)

	boTradeRequester := bo.NewTradeRequest(
		daoTradeRequester.GetCUDIDer().GetID()["id"],
		daoTradeRequester.GetGameID(),
		daoTradeRequester.GetSenderID(),
		daoTradeRequester.GetReceiverID(),
		daoTradeRequester.GetOfferingMoney(),
		daoTradeRequester.GetRequestingMoney(),
		daoTradeRequester.GetOfferingProperties(),
		daoTradeRequester.GetRequestingProperties(),
		daoTradeRequester.GetStatus(),
	)

	service.GetRuntimeLogger().
		WithFields(fields).
		WithField(object.URIFieldBOTradeRequester, boTradeRequester).
		Debug(object.URIEmpty)

	return boTradeRequester, nil
}

func (service *tradeRequestService) List(
	ctx context.Context,
	daoPaginationer dao.Paginationer,
	daoTradeFilter dao.TradeFilter,
) ([]bo.TradeRequester, dao.Cursorer, error) {
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
		"name":             "List",
		"rt_ctx":           utilRuntimeContext,
		"sp_ctx":           utilSpanContext,
		"config":           service.GetConfigger(),
		"dao_paginationer": daoPaginationer,
		"dao_trade_filter": daoTradeFilter,
	}

	service.GetRuntimeLogger().
		WithFields(fields).
		Info(object.URIEmpty)

	daoTradeRequesters, daoCursorer, err := service.GetTradeRequestRepositorier().
		ReadList(ctx, daoPaginationer, daoTradeFilter)
	if err != nil {
		service.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error("Failed to read trade request list")
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, "Failed to read trade request list")

		return nil, nil, err
	}

	service.GetRuntimeLogger().
		WithFields(fields).
		WithField(object.URIFieldDAOTradeRequesters, daoTradeRequesters).
		WithField(object.URIFieldDAOCursorer, daoCursorer).
		Debug(object.URIEmpty)

	boTradeRequesters := make([]bo.TradeRequester, 0, len(daoTradeRequesters))

	for key, daoTradeRequester := range daoTradeRequesters {
		service.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldKey, key).
			WithField(object.URIFieldDAOTradeRequester, daoTradeRequester).
			Debug(object.URIEmpty)

		boTradeRequesters = append(boTradeRequesters, bo.NewTradeRequest(
			daoTradeRequester.GetCUDIDer().GetID()["id"],
			daoTradeRequester.GetGameID(),
			daoTradeRequester.GetSenderID(),
			daoTradeRequester.GetReceiverID(),
			daoTradeRequester.GetOfferingMoney(),
			daoTradeRequester.GetRequestingMoney(),
			daoTradeRequester.GetOfferingProperties(),
			daoTradeRequester.GetRequestingProperties(),
			daoTradeRequester.GetStatus(),
		))
	}

	service.GetRuntimeLogger().
		WithFields(fields).
		WithField(object.URIFieldBOTradeRequesters, boTradeRequesters).
		Debug(object.URIEmpty)

	return boTradeRequesters, daoCursorer, nil
}

func (service *tradeRequestService) CreateTradeRequest(
	ctx context.Context,
	boTradeRequester bo.TradeRequester,
) (uuid.UUID, error) {
	var traceSpan trace.Span

	ctx, traceSpan = service.GetTracer().Start(
		ctx,
		"CreateTradeRequest",
		trace.WithSpanKind(trace.SpanKindInternal),
	)
	defer traceSpan.End()

	utilRuntimeContext := util.NewRuntimeContext(ctx)
	utilSpanContext := util.NewSpanContext(traceSpan)
	fields := map[string]any{
		"name":               "CreateTradeRequest",
		"rt_ctx":             utilRuntimeContext,
		"sp_ctx":             utilSpanContext,
		"config":             service.GetConfigger(),
		"bo_trade_requester": boTradeRequester,
	}

	service.GetRuntimeLogger().
		WithFields(fields).
		Info(object.URIEmpty)

	nowUTC := service.GetTimer().NowUTC()

	daoTradeRequester := dao.NewTradeRequest(
		uuid.Nil,
		boTradeRequester.GetGameID(),
		boTradeRequester.GetSenderID(),
		boTradeRequester.GetReceiverID(),
		boTradeRequester.GetOfferingMoney(),
		boTradeRequester.GetRequestingMoney(),
		boTradeRequester.GetOfferingProperties(),
		boTradeRequester.GetRequestingProperties(),
		"pending",
		nowUTC,
		nowUTC,
		sql.NullTime{
			Time:  time.Time{},
			Valid: false,
		},
	)

	service.GetRuntimeLogger().
		WithFields(fields).
		WithField(object.URIFieldDAOTradeRequester, daoTradeRequester).
		Debug(object.URIEmpty)

	id, err := service.GetTradeRequestRepositorier().Create(ctx, daoTradeRequester)
	if err != nil {
		service.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error("Failed to create trade request")
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, "Failed to create trade request")

		return uuid.Nil, err
	}

	service.GetRuntimeLogger().
		WithFields(fields).
		WithField(object.URIFieldID, id).
		Debug(object.URIEmpty)

	return id.GetID()["id"], nil
}

func (service *tradeRequestService) GetTradeRequestsByGameID(
	ctx context.Context,
	gameID uuid.UUID,
	daoPaginationer dao.Paginationer,
) ([]bo.TradeRequester, dao.Cursorer, error) {
	var traceSpan trace.Span

	ctx, traceSpan = service.GetTracer().Start(
		ctx,
		"GetTradeRequestsByGameID",
		trace.WithSpanKind(trace.SpanKindInternal),
	)
	defer traceSpan.End()

	utilRuntimeContext := util.NewRuntimeContext(ctx)
	utilSpanContext := util.NewSpanContext(traceSpan)
	fields := map[string]any{
		"name":             "GetTradeRequestsByGameID",
		"rt_ctx":           utilRuntimeContext,
		"sp_ctx":           utilSpanContext,
		"config":           service.GetConfigger(),
		"game_id":          gameID,
		"dao_paginationer": daoPaginationer,
	}

	service.GetRuntimeLogger().
		WithFields(fields).
		Info(object.URIEmpty)

	daoTradeFilter := dao.NewTradeFilter(
		[]uuid.UUID{},
		uuid.Nil,
		uuid.Nil,
		"",
		gameID,
	)

	service.GetRuntimeLogger().
		WithFields(fields).
		WithField("dao_trade_filter", daoTradeFilter).
		Debug(object.URIEmpty)

	return service.List(ctx, daoPaginationer, daoTradeFilter)
}

func (service *tradeRequestService) GetTradeRequestsByPlayerID(
	ctx context.Context,
	playerID uuid.UUID,
	daoPaginationer dao.Paginationer,
) ([]bo.TradeRequester, dao.Cursorer, error) {
	var traceSpan trace.Span

	ctx, traceSpan = service.GetTracer().Start(
		ctx,
		"GetTradeRequestsByPlayerID",
		trace.WithSpanKind(trace.SpanKindInternal),
	)
	defer traceSpan.End()

	utilRuntimeContext := util.NewRuntimeContext(ctx)
	utilSpanContext := util.NewSpanContext(traceSpan)
	fields := map[string]any{
		"name":             "GetTradeRequestsByPlayerID",
		"rt_ctx":           utilRuntimeContext,
		"sp_ctx":           utilSpanContext,
		"config":           service.GetConfigger(),
		"player_id":        playerID,
		"dao_paginationer": daoPaginationer,
	}

	service.GetRuntimeLogger().
		WithFields(fields).
		Info(object.URIEmpty)

	daoTradeFilterSender := dao.NewTradeFilter(
		[]uuid.UUID{},
		playerID,
		uuid.Nil,
		"",
		uuid.Nil,
	)

	senderRequests, senderCursor, err := service.List(ctx, daoPaginationer, daoTradeFilterSender)
	if err != nil {
		service.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error("Failed to get trade requests where player is sender")
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, "Failed to get trade requests where player is sender")

		return nil, nil, err
	}

	daoTradeFilterReceiver := dao.NewTradeFilter(
		[]uuid.UUID{},
		uuid.Nil,
		playerID,
		"",
		uuid.Nil,
	)

	receiverRequests, receiverCursor, err := service.List(ctx, daoPaginationer, daoTradeFilterReceiver)
	if err != nil {
		service.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error("Failed to get trade requests where player is receiver")
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, "Failed to get trade requests where player is receiver")

		return nil, nil, err
	}

	combinedRequests := append(senderRequests, receiverRequests...)

	var combinedCursor dao.Cursorer
	if senderCursor.GetOffset() > receiverCursor.GetOffset() {
		combinedCursor = senderCursor
	} else {
		combinedCursor = receiverCursor
	}

	service.GetRuntimeLogger().
		WithFields(fields).
		WithField(object.URIFieldBOTradeRequesters, combinedRequests).
		WithField(object.URIFieldDAOCursorer, combinedCursor).
		Debug(object.URIEmpty)

	return combinedRequests, combinedCursor, nil
}

func (service *tradeRequestService) clone() *tradeRequestService {
	newService := service

	return newService
}

func (optionerFunc tradeRequestServiceOptionerFunc) apply(
	service *tradeRequestService,
) {
	optionerFunc(service)
}

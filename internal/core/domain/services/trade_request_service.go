// internal/core/domain/services/trade_request_service.go
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

// TradeRequestService implements the input.TradeRequestServicer interface
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

// NewTradeRequestService creates a new trade request service instance
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

// WithTradeRequestServiceTradeRequestRepositorier sets the trade request repository for the service
func WithTradeRequestServiceTradeRequestRepositorier(
	repositoryTradeRequestRepositorier output.TradeRequestRepositorier,
) tradeRequestServiceOptioner {
	return tradeRequestServiceOptionerFunc(func(
		service *tradeRequestService,
	) {
		service.repositoryTradeRequestRepositorier = repositoryTradeRequestRepositorier
	})
}

// WithTradeRequestServiceTimer sets the timer for the service
func WithTradeRequestServiceTimer(
	objectTimer object.Timer,
) tradeRequestServiceOptioner {
	return tradeRequestServiceOptionerFunc(func(
		service *tradeRequestService,
	) {
		service.objectTimer = objectTimer
	})
}

// WithOptioners applies the provided optioners to the service
func (service *tradeRequestService) WithOptioners(
	optioners ...tradeRequestServiceOptioner,
) *tradeRequestService {
	for _, optioner := range optioners {
		optioner.apply(service)
	}

	return service
}

// GetTracer implements util.GetTracer
func (service *tradeRequestService) GetTracer() trace.Tracer {
	return service.traceTracer
}

// GetTradeRequestRepositorier implements output.GetTradeRequestRepositorier
func (service *tradeRequestService) GetTradeRequestRepositorier() output.TradeRequestRepositorier {
	return service.repositoryTradeRequestRepositorier
}

// GetUUIDer implements object.GetUUIDer
func (service *tradeRequestService) GetUUIDer() object.UUIDer {
	return service.objectUUIDer
}

// GetRuntimeLogger implements log.GetRuntimeLogger
func (service *tradeRequestService) GetRuntimeLogger() log.RuntimeLogger {
	return service.logRuntimeLogger
}

// GetConfigger implements config.GetConfigger
func (service *tradeRequestService) GetConfigger() config.Configger {
	return service.configConfigger
}

// GetTimer implements object.GetTimer
func (service *tradeRequestService) GetTimer() object.Timer {
	return service.objectTimer
}

// Get implements input.TradeRequestServicer
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

// List implements input.TradeRequestServicer
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

// CreateTradeRequest implements input.TradeRequestServicer
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
		"pending", // Initial status is always pending
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

// UpdateTradeRequest implements input.TradeRequestServicer
func (service *tradeRequestService) UpdateTradeRequest(
	ctx context.Context,
	id uuid.UUID,
	boTradeRequester bo.TradeRequester,
) error {
	var traceSpan trace.Span

	ctx, traceSpan = service.GetTracer().Start(
		ctx,
		"UpdateTradeRequest",
		trace.WithSpanKind(trace.SpanKindInternal),
	)
	defer traceSpan.End()

	utilRuntimeContext := util.NewRuntimeContext(ctx)
	utilSpanContext := util.NewSpanContext(traceSpan)
	fields := map[string]any{
		"name":               "UpdateTradeRequest",
		"rt_ctx":             utilRuntimeContext,
		"sp_ctx":             utilSpanContext,
		"config":             service.GetConfigger(),
		"id":                 id,
		"bo_trade_requester": boTradeRequester,
	}

	service.GetRuntimeLogger().
		WithFields(fields).
		Info(object.URIEmpty)

	// First check if the trade request exists
	cudIDer := dao.NewCUDID(map[string]uuid.UUID{"id": id})
	existingTradeRequest, err := service.GetTradeRequestRepositorier().Read(ctx, cudIDer)
	if err != nil {
		service.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error("Failed to read trade request for update")
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, "Failed to read trade request for update")

		return err
	}

	nowUTC := service.GetTimer().NowUTC()

	// Create updated trade request with existing timestamps
	daoTradeRequester := dao.NewTradeRequest(
		id,
		boTradeRequester.GetGameID(),
		boTradeRequester.GetSenderID(),
		boTradeRequester.GetReceiverID(),
		boTradeRequester.GetOfferingMoney(),
		boTradeRequester.GetRequestingMoney(),
		boTradeRequester.GetOfferingProperties(),
		boTradeRequester.GetRequestingProperties(),
		"pending", // Initial status is always pending
		existingTradeRequest.GetCUDer().GetCreatedAt(),
		nowUTC,
		existingTradeRequest.GetCUDer().GetDeletedAt(),
	)

	service.GetRuntimeLogger().
		WithFields(fields).
		WithField(object.URIFieldDAOTradeRequester, daoTradeRequester).
		Debug(object.URIEmpty)

	_, err = service.GetTradeRequestRepositorier().
		Update(ctx, daoTradeRequester)
	if err != nil {
		service.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error("Failed to update trade request")
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, "Failed to update trade request")

		return err
	}

	return nil
}

// DeleteTradeRequest implements input.TradeRequestServicer
func (service *tradeRequestService) DeleteTradeRequest(
	ctx context.Context,
	id uuid.UUID,
) error {
	var traceSpan trace.Span

	ctx, traceSpan = service.GetTracer().Start(
		ctx,
		"DeleteTradeRequest",
		trace.WithSpanKind(trace.SpanKindInternal),
	)
	defer traceSpan.End()

	utilRuntimeContext := util.NewRuntimeContext(ctx)
	utilSpanContext := util.NewSpanContext(traceSpan)
	fields := map[string]any{
		"name":   "DeleteTradeRequest",
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

	_, err := service.GetTradeRequestRepositorier().
		Delete(
			ctx,
			cudIDer,
		)
	if err != nil {
		service.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error("Failed to delete trade request")
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, "Failed to delete trade request")

		return err
	}

	return nil
}

// AcceptTradeRequest implements input.TradeRequestServicer
func (service *tradeRequestService) AcceptTradeRequest(
	ctx context.Context,
	id uuid.UUID,
) error {
	var traceSpan trace.Span

	ctx, traceSpan = service.GetTracer().Start(
		ctx,
		"AcceptTradeRequest",
		trace.WithSpanKind(trace.SpanKindInternal),
	)
	defer traceSpan.End()

	utilRuntimeContext := util.NewRuntimeContext(ctx)
	utilSpanContext := util.NewSpanContext(traceSpan)
	fields := map[string]any{
		"name":   "AcceptTradeRequest",
		"rt_ctx": utilRuntimeContext,
		"sp_ctx": utilSpanContext,
		"config": service.GetConfigger(),
		"id":     id,
	}

	service.GetRuntimeLogger().
		WithFields(fields).
		Info(object.URIEmpty)

	// First get the existing trade request
	cudIDer := dao.NewCUDID(map[string]uuid.UUID{"id": id})
	boTradeRequester, err := service.GetTradeRequestRepositorier().Read(ctx, cudIDer)
	if err != nil {
		service.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error("Failed to read trade request for acceptance")
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, "Failed to read trade request for acceptance")

		return err
	}

	// Check if the trade request is in a state that can be accepted
	if boTradeRequester.GetStatus() != "pending" {
		service.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error("Trade request is not in pending status")
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, "Trade request is not in pending status")

		return err
	}

	nowUTC := service.GetTimer().NowUTC()

	// Update the trade request status to accepted
	updatedTradeRequest := dao.NewTradeRequest(id,
		boTradeRequester.GetGameID(),
		boTradeRequester.GetSenderID(),
		boTradeRequester.GetReceiverID(),
		boTradeRequester.GetOfferingMoney(),
		boTradeRequester.GetRequestingMoney(),
		boTradeRequester.GetOfferingProperties(),
		boTradeRequester.GetRequestingProperties(),
		"accepted", // Initial status is always pending
		boTradeRequester.GetCUDer().GetCreatedAt(),
		nowUTC,
		boTradeRequester.GetCUDer().GetDeletedAt(),
	)

	service.GetRuntimeLogger().
		WithFields(fields).
		WithField(object.URIFieldDAOTradeRequester, updatedTradeRequest).
		Debug(object.URIEmpty)

	_, err = service.GetTradeRequestRepositorier().
		Update(ctx, updatedTradeRequest)
	if err != nil {
		service.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error("Failed to update trade request status to accepted")
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, "Failed to update trade request status to accepted")

		return err
	}

	// TODO: Implement the actual property and money transfer logic here
	// This would involve updating player properties and balances

	return nil
}

// RejectTradeRequest implements input.TradeRequestServicer
func (service *tradeRequestService) RejectTradeRequest(
	ctx context.Context,
	id uuid.UUID,
) error {
	var traceSpan trace.Span

	ctx, traceSpan = service.GetTracer().Start(
		ctx,
		"RejectTradeRequest",
		trace.WithSpanKind(trace.SpanKindInternal),
	)
	defer traceSpan.End()

	utilRuntimeContext := util.NewRuntimeContext(ctx)
	utilSpanContext := util.NewSpanContext(traceSpan)
	fields := map[string]any{
		"name":   "RejectTradeRequest",
		"rt_ctx": utilRuntimeContext,
		"sp_ctx": utilSpanContext,
		"config": service.GetConfigger(),
		"id":     id,
	}

	service.GetRuntimeLogger().
		WithFields(fields).
		Info(object.URIEmpty)

	// First get the existing trade request
	cudIDer := dao.NewCUDID(map[string]uuid.UUID{"id": id})
	existingTradeRequest, err := service.GetTradeRequestRepositorier().Read(ctx, cudIDer)
	if err != nil {
		service.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error("Failed to read trade request for rejection")
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, "Failed to read trade request for rejection")

		return err
	}

	// Check if the trade request is in a state that can be rejected
	if existingTradeRequest.GetStatus() != "pending" {
		service.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error("Trade request is not in pending status")
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, "Trade request is not in pending status")

		return err
	}

	nowUTC := service.GetTimer().NowUTC()

	// Update the trade request status to rejected
	updatedTradeRequest := dao.NewTradeRequest(
		id,
		existingTradeRequest.GetGameID(),
		existingTradeRequest.GetSenderID(),
		existingTradeRequest.GetReceiverID(),
		existingTradeRequest.GetOfferingMoney(),
		existingTradeRequest.GetRequestingMoney(),
		existingTradeRequest.GetOfferingProperties(),
		existingTradeRequest.GetRequestingProperties(),
		"rejected",
		existingTradeRequest.GetCUDer().GetCreatedAt(),
		nowUTC,
		existingTradeRequest.GetCUDer().GetDeletedAt(),
	)

	service.GetRuntimeLogger().
		WithFields(fields).
		WithField(object.URIFieldDAOTradeRequester, updatedTradeRequest).
		Debug(object.URIEmpty)

	_, err = service.GetTradeRequestRepositorier().
		Update(ctx,
			updatedTradeRequest,
		)
	if err != nil {
		service.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error("Failed to update trade request status to rejected")
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, "Failed to update trade request status to rejected")

		return err
	}

	return nil
}

// GetTradeRequestsByGameID implements input.TradeRequestServicer
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

	// Create a filter for game ID
	daoTradeFilter := dao.NewTradeFilter(
		[]uuid.UUID{},
		uuid.Nil, // No specific sender
		uuid.Nil, // No specific receiver
		"",
		gameID,
	)

	service.GetRuntimeLogger().
		WithFields(fields).
		WithField("dao_trade_filter", daoTradeFilter).
		Debug(object.URIEmpty)

	// Use the List method with the game ID filter
	return service.List(ctx, daoPaginationer, daoTradeFilter)
}

// GetTradeRequestsByPlayerID implements input.TradeRequestServicer
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

	// Create a filter for player ID (as either sender or receiver)
	// We'll need to make two separate queries and combine the results

	// First, get trade requests where player is the sender
	daoTradeFilterSender := dao.NewTradeFilter(
		[]uuid.UUID{},
		playerID, // Player as sender
		uuid.Nil, // No specific game
		"",       // No specific receiver
		uuid.Nil, // No specific receiver
	)

	service.GetRuntimeLogger().
		WithFields(fields).
		WithField("dao_trade_filter_sender", daoTradeFilterSender).
		Debug(object.URIEmpty)

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

	// Then, get trade requests where player is the receiver
	daoTradeFilterReceiver := dao.NewTradeFilter(
		[]uuid.UUID{},
		uuid.Nil, // No specific game
		playerID, // No specific sender
		"",       // Player as receiver
		uuid.Nil, // No specific status
	)

	service.GetRuntimeLogger().
		WithFields(fields).
		WithField("dao_trade_filter_receiver", daoTradeFilterReceiver).
		Debug(object.URIEmpty)

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

	// Combine the results
	combinedRequests := append(senderRequests, receiverRequests...)

	// Use the cursor with the higher offset
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

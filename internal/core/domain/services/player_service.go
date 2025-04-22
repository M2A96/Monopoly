// internal/services/player_service.go
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

// PlayerService implements the input.PlayerServicer interface

type (
	playerService struct {
		configConfigger              config.Configger
		logRuntimeLogger             log.RuntimeLogger
		objectUUIDer                 object.UUIDer
		repositoryPlayerRepositorier output.PlayerRepositorier
		traceTracer                  trace.Tracer
	}

	playerServiceOptioner interface {
		apply(*playerService)
	}

	playerServiceOptionerFunc func(*playerService)
)

var (
	_ input.PlayerServicer         = (*playerService)(nil)
	_ config.GetConfigger          = (*playerService)(nil)
	_ log.GetRuntimeLogger         = (*playerService)(nil)
	_ object.GetUUIDer             = (*playerService)(nil)
	_ output.GetPlayerRepositorier = (*playerService)(nil)
	_ util.GetTracer               = (*playerService)(nil)
)

// NewPlayerService creates a new player service instance
func NewPlayerService(
	configConfigger config.Configger,
	logRuntimeLogger log.RuntimeLogger,
	objectUUIDer object.UUIDer,
	traceTracer trace.Tracer,
	optioners ...playerServiceOptioner,
) *playerService {
	playerService := &playerService{
		configConfigger:              configConfigger,
		logRuntimeLogger:             logRuntimeLogger,
		objectUUIDer:                 objectUUIDer,
		repositoryPlayerRepositorier: nil,
		traceTracer:                  traceTracer,
	}

	return playerService.WithOptioners(optioners...)
}

// WithPlayerServicePlayerRepositorier sets the player repository for the service
func WithPlayerServicePlayerRepositorier(
	repositoryPlayerRepositorier output.PlayerRepositorier,
) playerServiceOptioner {
	return playerServiceOptionerFunc(func(
		service *playerService,
	) {
		service.repositoryPlayerRepositorier = repositoryPlayerRepositorier
	})
}

// GetTracer implements util.GetTracer.
func (p *playerService) GetTracer() trace.Tracer {
	return p.traceTracer
}

// GetPlayerRepositorier implements output.GetPlayerRepositorier.
func (p *playerService) GetPlayerRepositorier() output.PlayerRepositorier {
	return p.repositoryPlayerRepositorier
}

// GetUUIDer implements object.GetUUIDer.
func (p *playerService) GetUUIDer() object.UUIDer {
	return p.objectUUIDer
}

// GetRuntimeLogger implements log.GetRuntimeLogger.
func (p *playerService) GetRuntimeLogger() log.RuntimeLogger {
	return p.logRuntimeLogger
}

// GetConfigger implements config.GetConfigger.
func (p *playerService) GetConfigger() config.Configger {
	return p.configConfigger
}

// CreatePlayer implements input.PlayerServicer.
func (service *playerService) CreatePlayer(
	ctx context.Context,
	boPlayer bo.Player,
) (uuid.UUID, error) {
	var traceSpan trace.Span

	ctx, traceSpan = service.GetTracer().Start(
		ctx,
		"CreatePlayer",
		trace.WithSpanKind(trace.SpanKindInternal),
	)
	defer traceSpan.End()

	utilRuntimeContext := util.NewRuntimeContext(ctx)
	utilSpanContext := util.NewSpanContext(traceSpan)
	fields := map[string]any{
		"name":      "CreatePlayer",
		"rt_ctx":    utilRuntimeContext,
		"sp_ctx":    utilSpanContext,
		"config":    service.GetConfigger(),
		"bo_player": boPlayer,
	}

	service.GetRuntimeLogger().
		WithFields(fields).
		Info(object.URIEmpty)

	daoPlayer := dao.NewPlayer(
		uuid.Nil,
		boPlayer.GetGameID(),
		boPlayer.GetName(),
		boPlayer.GetBalance(),
		boPlayer.GetBalance(),
		boPlayer.GetInJail(),
		boPlayer.GetJailTurns(),
		boPlayer.GetBankrupt(),
		time.Time{},
		time.Time{},
		sql.NullTime{
			Time:  time.Time{},
			Valid: false,
		},
	)

	service.GetRuntimeLogger().
		WithFields(fields).
		WithField(object.URIFieldDAOPlayer, daoPlayer).
		Debug(object.URIEmpty)

	id, err := service.GetPlayerRepositorier().Create(ctx, daoPlayer)
	if err != nil {
		service.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error("Failed to create player")
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, "Failed to create player")

		return uuid.Nil, err
	}

	service.GetRuntimeLogger().
		WithFields(fields).
		WithField(object.URIFieldID, id).
		Debug(object.URIEmpty)

	return id.GetID()["id"], nil
}

// Get implements input.PlayerServicer.
func (service *playerService) Get(
	ctx context.Context,
	playerID uuid.UUID,
) (bo.Player, error) {
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
		"name":      "Get",
		"rt_ctx":    utilRuntimeContext,
		"sp_ctx":    utilSpanContext,
		"config":    service.GetConfigger(),
		"player_id": playerID,
	}

	service.GetRuntimeLogger().
		WithFields(fields).
		Info(object.URIEmpty)

	daoPlayer, err := service.GetPlayerRepositorier().Read(
		ctx,
		dao.NewCUDID(map[string]uuid.UUID{"id": playerID}),
	)
	if err != nil {
		service.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error("Failed to retrieve player")
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, "Failed to retrieve player")

		return nil, err
	}
	service.GetRuntimeLogger().
		WithFields(fields).
		WithField(object.URIFieldDAOPlayer, daoPlayer).
		Debug(object.URIEmpty)

	player := bo.NewPlayer(
		daoPlayer.GetCUDIDer().GetID()["id"],
		daoPlayer.GetGameID(),
		daoPlayer.GetName(),
		daoPlayer.GetBalance(),
		daoPlayer.GetPosition(),
		daoPlayer.GetInJail(),
		daoPlayer.GetJailTurns(),
		daoPlayer.GetBankrupt(),
	)

	service.GetRuntimeLogger().
		WithFields(fields).
		WithField(object.URIFieldBOPlayer, player).
		Debug(object.URIEmpty)

	return player, nil
}

// List implements input.PlayerServicer.
func (service *playerService) List(
	ctx context.Context,
	pagination dao.Paginationer,
	filter dao.PlayerFilter,
) ([]bo.Player, dao.Cursorer, error) {
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

	daoPlayers, cursor, err := service.GetPlayerRepositorier().
		ReadList(
			ctx,
			pagination,
			filter,
		)
	if err != nil {
		service.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error("Failed to list players")
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, "Failed to list players")

		return nil, nil, err
	}

	boPlayers := make([]bo.Player, len(daoPlayers))
	for key, daoPlayer := range daoPlayers {
		service.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldKey, key).
			WithField(object.URIFieldDAOPlayer, boPlayers).
			Debug(object.URIEmpty)

		boPlayers = append(boPlayers, bo.NewPlayer(
			daoPlayer.GetCUDIDer().GetID()["id"],
			daoPlayer.GetGameID(),
			daoPlayer.GetName(),
			daoPlayer.GetBalance(),
			daoPlayer.GetPosition(),
			daoPlayer.GetInJail(),
			daoPlayer.GetJailTurns(),
			daoPlayer.GetBankrupt(),
		))
	}

	service.GetRuntimeLogger().
		WithFields(fields).
		WithField(object.URIFieldBOPlayers, boPlayers).
		Debug(object.URIEmpty)

	return boPlayers, cursor, nil
}

// UpdatePlayer implements input.PlayerServicer.
func (service *playerService) UpdatePlayer(
	ctx context.Context,
	playerID uuid.UUID,
	boPlayer bo.Player,
) error {
	var traceSpan trace.Span

	ctx, traceSpan = service.GetTracer().Start(
		ctx,
		"UpdatePlayer",
		trace.WithSpanKind(trace.SpanKindInternal),
	)
	defer traceSpan.End()

	utilRuntimeContext := util.NewRuntimeContext(ctx)
	utilSpanContext := util.NewSpanContext(traceSpan)
	fields := map[string]any{
		"name":      "UpdatePlayer",
		"rt_ctx":    utilRuntimeContext,
		"sp_ctx":    utilSpanContext,
		"config":    service.GetConfigger(),
		"player_id": playerID,
		"bo_player": boPlayer,
	}

	service.GetRuntimeLogger().
		WithFields(fields).
		Info(object.URIEmpty)

	daoPlayer := dao.NewPlayer(
		uuid.Nil,
		boPlayer.GetGameID(),
		boPlayer.GetName(),
		boPlayer.GetBalance(),
		boPlayer.GetBalance(),
		boPlayer.GetInJail(),
		boPlayer.GetJailTurns(),
		boPlayer.GetBankrupt(),
		time.Time{},
		time.Time{},
		sql.NullTime{
			Time:  time.Time{},
			Valid: false,
		},
	)

	service.GetRuntimeLogger().
		WithFields(fields).
		WithField(object.URIFieldDAOPlayer, daoPlayer).
		Debug(object.URIEmpty)

	if _, err := service.GetPlayerRepositorier().
		Update(
			ctx,
			daoPlayer,
		); err != nil {
		service.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error("Failed to update player")
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, "Failed to update player")

		return err
	}

	service.GetRuntimeLogger().
		WithFields(fields).
		Debug(object.URIEmpty)

	return nil
}

// DeletePlayer implements input.PlayerServicer.
func (service *playerService) DeletePlayer(
	ctx context.Context,
	playerID uuid.UUID,
) error {
	var traceSpan trace.Span

	ctx, traceSpan = service.GetTracer().Start(
		ctx,
		"DeletePlayer",
		trace.WithSpanKind(trace.SpanKindInternal),
	)
	defer traceSpan.End()

	utilRuntimeContext := util.NewRuntimeContext(ctx)
	utilSpanContext := util.NewSpanContext(traceSpan)
	fields := map[string]any{
		"name":      "DeletePlayer",
		"rt_ctx":    utilRuntimeContext,
		"sp_ctx":    utilSpanContext,
		"config":    service.GetConfigger(),
		"player_id": playerID,
	}

	service.GetRuntimeLogger().
		WithFields(fields).
		Info(object.URIEmpty)

	if _, err := service.GetPlayerRepositorier().
		Delete(
			ctx,
			dao.NewCUDID(map[string]uuid.UUID{"id": playerID}),
		); err != nil {
		service.GetRuntimeLogger().
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error("Failed to delete player")
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, "Failed to delete player")

		return err
	}

	service.GetRuntimeLogger().
		WithFields(fields).
		Debug(object.URIEmpty)

	return nil
}

// WithOptioners applies the given optioners to the player service
func (service *playerService) WithOptioners(
	optioners ...playerServiceOptioner,
) *playerService {
	newService := service.clone()
	for _, optioner := range optioners {
		optioner.apply(newService)
	}

	return newService
}

func (service *playerService) clone() *playerService {
	newService := service

	return newService
}

func (optionerFunc playerServiceOptionerFunc) apply(
	service *playerService,
) {
	optionerFunc(service)
}

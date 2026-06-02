package output

import (
	"github/M2A96/Monopoly.git/config"
	"github/M2A96/Monopoly.git/log"
	"github/M2A96/Monopoly.git/object"
	"github/M2A96/Monopoly.git/object/dao"

	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
)

type (
	// TradeRequestRepositorier is an interface.
	TradeRequestRepositorier Repositorier[dao.TradeRequester, dao.TradeFilter]

	// GetTradeRequestRepositorier is an interface.
	GetTradeRequestRepositorier interface {
		// GetTradeRequestRepositorier is a function.
		GetTradeRequestRepositorier() TradeRequestRepositorier
	}

	tradeRequestRepository struct {
		repository
	}

	tradeRequestRepositoryOptioner = repositoryOptioner
)

// NewTradeRequestRepository is a function.
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

// WithTradeRequestRepositoryTimer is a function.
func WithTradeRequestRepositoryTimer(
	objectTimer object.Timer,
) tradeRequestRepositoryOptioner {
	return WithRepositoryTimer(objectTimer)
}

// WithTradeRequestRepositoryDB is a function.
func WithTradeRequestRepositoryDB(
	gormDB *gorm.DB,
) tradeRequestRepositoryOptioner {
	return WithRepositoryDB(gormDB, "trade_request")
}

package output

import (
	"github/M2A96/Monopoly.git/object"
	"github/M2A96/Monopoly.git/object/dao"

	"gorm.io/gorm"
)

type (
	// GameRepositorier is an interface.
	GameRepositorier Repositorier[dao.Gamer, dao.GameFilter]

	// GetGameRepositorier is an interface.
	GetGameRepositorier interface {
		// GetGameRepositorier is a function.
		GetGameRepositorier() GameRepositorier
	}

	gameRepository struct {
		repository
	}

	gameRepositoryOptioner = repositoryOptioner
)

// WithGameRepositoryTimer is a function.
func WithGameRepositoryTimer(
	objectTimer object.Timer,
) gameRepositoryOptioner {
	return WithRepositoryTimer(objectTimer)
}

// WithGameRepositoryDB is a function.
func WithGameRepositoryDB(
	gormDB *gorm.DB,
) gameRepositoryOptioner {
	return WithRepositoryDB(gormDB, object.URITableGame)
}

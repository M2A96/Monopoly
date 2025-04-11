// internal/api/routes.go
package api

import (
	"yourproject/internal/database"

	"github.com/labstack/echo/v4"
)

func RegisterRoutes(e *echo.Echo, db *database.DB) {
	// Create game group
	gameGroup := e.Group("/games")

	// Game routes
	gameGroup.POST("", createGame(db))
	gameGroup.GET("/:id", getGame(db))
	gameGroup.PUT("/:id/start", startGame(db))
	gameGroup.GET("/:id/state", getGameState(db))

	// Player routes
	gameGroup.POST("/:id/players", addPlayer(db))
	gameGroup.PUT("/:id/players/:player_id/turn", executePlayerTurn(db))
	gameGroup.PUT("/:id/players/:player_id/buy", buyProperty(db))
	gameGroup.PUT("/:id/players/:player_id/build", buildHouse(db))

	// Property routes
	gameGroup.GET("/:id/properties", getProperties(db))
	gameGroup.PUT("/:id/properties/:property_id/mortgage", mortgageProperty(db))
	gameGroup.PUT("/:id/properties/:property_id/unmortgage", unmortgageProperty(db))

	// Game action routes
	gameGroup.POST("/:id/actions/trade", proposeTrade(db))
	gameGroup.GET("/:id/logs", getGameLogs(db))
}

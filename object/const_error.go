package object

import "errors"

var (
	ErrTypeAssertion            = errors.New("type assertion failed")
	ErrSDKResourceMerge         = errors.New("failed to merge resources")
	ErrBase64Decode2            = errors.New("unrecognized level")
	ErrSDKResourceNew           = errors.New("failed to create a new resource")
	ErrOTELExporterOTLPTraceNew = errors.New("failed to create a exporter otlp trace")
	ErrTracerProviderShutdown   = errors.New(
		"failed to shutdown traceTracer provider",
	)
	ErrUUIDerNewRandom                = errors.New("failed to new random uuid")
	ErrGameRepositoryCreate           = errors.New("failed to create game")
	ErrGameRepositoryRead             = errors.New("failed to read game")
	ErrGameRepositoryReadList         = errors.New("failed to read game list")
	ErrGameRepositoryUpdate           = errors.New("failed to update game")
	ErrGameRepositoryDelete           = errors.New("failed to delete game")
	ErrDAONewGameFromMap              = errors.New("failed to new game from map")
	ErrPlayerRepositoryCreate         = errors.New("failed to create player")
	ErrPlayerRepositoryRead           = errors.New("failed to read player")
	ErrPlayerRepositoryReadList       = errors.New("failed to read player list")
	ErrPlayerRepositoryUpdate         = errors.New("failed to update player")
	ErrPlayerRepositoryDelete         = errors.New("failed to delete player")
	ErrPlayerRepositoryReadNotFound   = errors.New("player not found")
	ErrDAONewPlayerFromMap            = errors.New("failed to new player from map")
	ErrPropertyRepositoryCreate       = errors.New("failed to create property")
	ErrPropertyRepositoryRead         = errors.New("failed to read property")
	ErrPropertyRepositoryReadList     = errors.New("failed to read property list")
	ErrPropertyRepositoryUpdate       = errors.New("failed to update property")
	ErrPropertyRepositoryDelete       = errors.New("failed to delete property")
	ErrDAONewPropertyFromMap          = errors.New("failed to new property from map")
	ErrGameStatusInvalid              = errors.New("invalid game status")
	ErrPlayerRepositoryUpdateNotFound = errors.New("player not found")
	ErrPlayerRepositoryNotFound       = errors.New("player not found")
	ErrDataVersion                    = errors.New("failed to retrieve data version")
	ErrSelectedValidateMapRule        = errors.New("failed to select validate map rule")
	ErrRegisterDefaultTranslations    = errors.New(
		"failed to register default translations",
	)
	ErrUUIDerParse            = errors.New("failed to parse uuid")
	ErrGameServiceGet         = errors.New("failed to get game")
	ErrGameServiceCreate      = errors.New("failed to create game")
	ErrGameServiceUpdate      = errors.New("failed to update game")
	ErrGameServiceDelete      = errors.New("failed to delete game")
	ErrGameServiceGetNotFound = errors.New("game not found")
	ErrGameServiceGetState    = errors.New(
		"failed to get game state",
	)
	ErrPropertyServiceCreate          = errors.New("failed to create property")
	ErrPropertyServiceGet             = errors.New("failed to get property")
	ErrPropertyServiceUpdate          = errors.New("failed to update property")
	ErrPropertyServiceUpdateOwnerShip = errors.New("failed to update property owner ship")
	ErrPropertyServiceDelete          = errors.New("failed to delete property")
	ErrPropertyServiceRead            = errors.New("failed to read property")
	ErrGameServiceStart               = errors.New("failed to start game")

	ErrJWKKeyParseKey  = errors.New("failed to parse key of jwk")
	ErrJWKPublicKeyOf  = errors.New("failed to public key of jwk")
	ErrJWKSetAddKey    = errors.New("failed to add key jwk set")
	ErrSQL             = errors.New("sql error")
	ErrGormOpen        = errors.New("failed to open gorm")
	ErrSQLDBFromGormDB = errors.New(
		"failed to receive sql db from gorm db",
	)
	ErrSQLDBClose                     = errors.New("failed to close sql db")
	ErrPropertyServiceAlreadyHasHotel = errors.New("property already has a hotel")
	ErrBase64Decode                   = errors.New("failed to decode base64")
	ErrGobDecoderDecode               = errors.New("failed to decode gob decoder")
	ErrPlayerHandlerDelete            = errors.New("player handler failed to delete player")
)

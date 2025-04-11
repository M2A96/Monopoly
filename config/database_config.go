package config

//go:generate mockgen -destination=../test/v2/database_config.go -package=test -mock_names=DatabaseConfigger=MockDatabaseConfig . DatabaseConfigger

import (
	"encoding/json"
	"github/M2A96/Monopoly.git/object"
)

type (
	// DatabaseConfigger is an interface.
	DatabaseConfigger interface {
		// GetDSN is a function.
		GetDSN() string
	}

	// GetDatabaseConfigger is an interface.
	GetDatabaseConfigger interface {
		// GetDatabaseConfigger is a function.
		GetDatabaseConfigger() DatabaseConfigger
	}

	databaseConfig struct {
		dsn string
	}

	databaseConfigOptioner interface {
		apply(*databaseConfig)
	}

	databaseConfigOptionerFunc func(*databaseConfig)
)

var (
	_ DatabaseConfigger = (*databaseConfig)(nil)
	_ json.Marshaler    = (*databaseConfig)(nil)
	_ object.GetMapper  = (*databaseConfig)(nil)
)

// NewDatabaseConfig is a function.
func NewDatabaseConfig(
	optioners ...databaseConfigOptioner,
) *databaseConfig {
	databaseConfig := &databaseConfig{
		dsn: object.URIEmpty,
	}

	return databaseConfig.WithOptioners(optioners...)
}

// WithDatabaseConfigDSN is a function.
func WithDatabaseConfigDSN(
	dsn string,
) databaseConfigOptioner {
	return databaseConfigOptionerFunc(func(
		config *databaseConfig,
	) {
		config.dsn = dsn
	})
}

// GetDSN is a function.
func (config *databaseConfig) GetDSN() string {
	return config.dsn
}

// GetMap is a function.
func (config *databaseConfig) GetMap() map[string]any {
	return map[string]any{
		"dsn": config.GetDSN(),
	}
}

// MarshalJSON is a function.
// read more https://pkg.go.dev/encoding/json#Marshaler
func (config *databaseConfig) MarshalJSON() ([]byte, error) {
	return json.Marshal(config.GetMap())
}

// WithOptioners is a function.
func (config *databaseConfig) WithOptioners(
	optioners ...databaseConfigOptioner,
) *databaseConfig {
	newConfig := config.clone()
	for _, optioner := range optioners {
		optioner.apply(newConfig)
	}

	return newConfig
}

func (config *databaseConfig) clone() *databaseConfig {
	newConfig := config

	return newConfig
}

func (optionerFunc databaseConfigOptionerFunc) apply(
	config *databaseConfig,
) {
	optionerFunc(config)
}

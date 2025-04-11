package config

//go:generate mockgen -destination=../test/v2/runtime_config.go -package=test -mock_names=RuntimeConfigger=MockRuntimeConfig . RuntimeConfigger

import (
	"encoding/json"
	"github/M2A96/Monopoly.git/object"
)

type (
	// RuntimeConfigger is an interface.
	RuntimeConfigger interface {
		// GetValidateMapRules is a function.
		GetValidateMapRules() map[string][]map[string]any
		// GetClientPaginationRequestSizeMax is a function.
		GetClientPaginationRequestSizeMax() int
	}

	// GetRuntimeConfigger is an interface.
	GetRuntimeConfigger interface {
		// GetRuntimeConfigger is a function.
		GetRuntimeConfigger() RuntimeConfigger
	}

	runtimeConfig struct {
		validateMapRules               map[string][]map[string]any
		clientPaginationRequestSizeMax int
	}

	runtimeConfigOptioner interface {
		apply(*runtimeConfig)
	}

	runtimeConfigOptionerFunc func(*runtimeConfig)
)

var (
	_ RuntimeConfigger = (*runtimeConfig)(nil)
	_ json.Marshaler   = (*runtimeConfig)(nil)
	_ object.GetMapper = (*runtimeConfig)(nil)
)

// NewRuntimeConfig is a function.
func NewRuntimeConfig(
	optioners ...runtimeConfigOptioner,
) *runtimeConfig {
	runtimeConfig := &runtimeConfig{
		validateMapRules:               map[string][]map[string]any{},
		clientPaginationRequestSizeMax: 0,
	}

	return runtimeConfig.WithOptioners(optioners...)
}

// WithRuntimeConfigValidateMapRules is a function.
func WithRuntimeConfigValidateMapRules(
	validateMapRules map[string][]map[string]any,
) runtimeConfigOptioner {
	return runtimeConfigOptionerFunc(func(
		config *runtimeConfig,
	) {
		config.validateMapRules = validateMapRules
	})
}

// WithRuntimeConfigClientPaginationRequestSizeMax is a function.
func WithRuntimeConfigClientPaginationRequestSizeMax(
	clientPaginationRequestSizeMax int,
) runtimeConfigOptioner {
	return runtimeConfigOptionerFunc(func(
		config *runtimeConfig,
	) {
		config.clientPaginationRequestSizeMax = clientPaginationRequestSizeMax
	})
}

// GetValidateMapRules is a function.
func (config *runtimeConfig) GetValidateMapRules() map[string][]map[string]any {
	return config.validateMapRules
}

// GetClientPaginationRequestSizeMax is a function.
func (config *runtimeConfig) GetClientPaginationRequestSizeMax() int {
	return config.clientPaginationRequestSizeMax
}

// GetMap is a function.
func (config *runtimeConfig) GetMap() map[string]any {
	return map[string]any{
		"validate_map_rules":                 config.GetValidateMapRules(),
		"client_pagination_request_size_max": config.GetClientPaginationRequestSizeMax(),
	}
}

// MarshalJSON is a function.
// read more https://pkg.go.dev/encoding/json#Marshaler
func (config *runtimeConfig) MarshalJSON() ([]byte, error) {
	return json.Marshal(config.GetMap())
}

// WithOptioners is a function.
func (config *runtimeConfig) WithOptioners(
	optioners ...runtimeConfigOptioner,
) *runtimeConfig {
	newConfig := config.clone()
	for _, optioner := range optioners {
		optioner.apply(newConfig)
	}

	return newConfig
}

func (config *runtimeConfig) clone() *runtimeConfig {
	newConfig := config

	return newConfig
}

func (optionerFunc runtimeConfigOptionerFunc) apply(
	config *runtimeConfig,
) {
	optionerFunc(config)
}

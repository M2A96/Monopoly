package config

//go:generate mockgen -destination=../test/v2/endpoint_config.go -package=test -mock_names=EndpointConfigger=MockEndpointConfig . EndpointConfigger

import (
	"encoding/json"
	"github/M2A96/Monopoly.git/object"
)

type (
	// EndpointConfigger describes a grpc endpoint.
	EndpointConfigger interface {
		// GetAddr is endpoint of the grpc service.
		GetAddr() string
		// GetNetwork is one of "tcp" or "unix" network type which is consistent to Addr.
		GetNetwork() string
	}

	endpointConfig struct {
		addr    string
		network string
	}

	endpointConfigOptioner interface {
		apply(*endpointConfig)
	}

	endpointConfigOptionerFunc func(*endpointConfig)
)

var (
	_ EndpointConfigger = (*endpointConfig)(nil)
	_ json.Marshaler    = (*endpointConfig)(nil)
	_ object.GetMapper  = (*endpointConfig)(nil)
)

// NewEndpointConfig is a function.
func NewEndpointConfig(
	optioners ...endpointConfigOptioner,
) *endpointConfig {
	endpointConfig := &endpointConfig{
		addr:    object.URIEmpty,
		network: object.URIEmpty,
	}

	return endpointConfig.WithOptioners(optioners...)
}

// WithEndpointConfigAddr is a function.
func WithEndpointConfigAddr(
	addr string,
) endpointConfigOptioner {
	return endpointConfigOptionerFunc(func(
		config *endpointConfig,
	) {
		config.addr = addr
	})
}

// WithEndpointConfigNetwork is a function.
func WithEndpointConfigNetwork(
	network string,
) endpointConfigOptioner {
	return endpointConfigOptionerFunc(func(
		config *endpointConfig,
	) {
		config.network = network
	})
}

// GetAddr is a function.
func (config *endpointConfig) GetAddr() string {
	return config.addr
}

// GetNetwork is a function.
func (config *endpointConfig) GetNetwork() string {
	return config.network
}

// GetMap is a function.
func (config *endpointConfig) GetMap() map[string]any {
	return map[string]any{
		"addr":    config.GetAddr(),
		"network": config.GetNetwork(),
	}
}

// MarshalJSON is a function.
// read more https://pkg.go.dev/encoding/json#Marshaler
func (config *endpointConfig) MarshalJSON() ([]byte, error) {
	return json.Marshal(config.GetMap())
}

// WithOptioners is a function.
func (config *endpointConfig) WithOptioners(
	optioners ...endpointConfigOptioner,
) *endpointConfig {
	newConfig := config.clone()
	for _, optioner := range optioners {
		optioner.apply(newConfig)
	}

	return newConfig
}

func (config *endpointConfig) clone() *endpointConfig {
	newConfig := config

	return newConfig
}

func (optionerFunc endpointConfigOptionerFunc) apply(
	config *endpointConfig,
) {
	optionerFunc(config)
}

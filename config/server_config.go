package config

//go:generate mockgen -destination=../test/v2/server_config.go -package=test -mock_names=ServerConfigger=MockServerConfig . ServerConfigger

import (
	"encoding/json"
	"github/M2A96/Monopoly.git/object"
)

type (
	// GetServerConfigger is an interface.
	GetServerConfigger interface {
		// GetServerConfigger is a function.
		GetServerConfigger() ServerConfigger
	}

	// ServerConfigger is configuration relevant to the grpc.
	ServerConfigger interface {
		// GetEndpointConfigger defines an endpoint of a grpc service.
		GetEndpointConfigger() EndpointConfigger
	}

	serverConfig struct {
		endpointConfigger EndpointConfigger
	}

	serverConfigOptioner interface {
		apply(*serverConfig)
	}

	serverConfigOptionerFunc func(*serverConfig)
)

var (
	_ ServerConfigger  = (*serverConfig)(nil)
	_ json.Marshaler   = (*serverConfig)(nil)
	_ object.GetMapper = (*serverConfig)(nil)
)

// NewServerConfig is a function.
func NewServerConfig(
	optioners ...serverConfigOptioner,
) *serverConfig {
	serverConfig := &serverConfig{
		endpointConfigger: &endpointConfig{
			addr:    object.URIEmpty,
			network: object.URIEmpty,
		},
	}

	return serverConfig.WithOptioners(optioners...)
}

// WithServerConfigEndpointConfigger is a function.
func WithServerConfigEndpointConfigger(
	endpointConfigger EndpointConfigger,
) serverConfigOptioner {
	return serverConfigOptionerFunc(func(
		config *serverConfig,
	) {
		config.endpointConfigger = endpointConfigger
	})
}

// GetEndpoint defines an endpoint of a grpc service.
func (config *serverConfig) GetEndpointConfigger() EndpointConfigger {
	return config.endpointConfigger
}

// GetMap is a function.
func (config *serverConfig) GetMap() map[string]any {
	return map[string]any{
		"endpoint_configger": config.GetEndpointConfigger(),
	}
}

// MarshalJSON is a function.
// read more https://pkg.go.dev/encoding/json#Marshaler
func (config *serverConfig) MarshalJSON() ([]byte, error) {
	return json.Marshal(config.GetMap())
}

// WithOptioners is a function.
func (config *serverConfig) WithOptioners(
	optioners ...serverConfigOptioner,
) *serverConfig {
	newConfig := config.clone()
	for _, optioner := range optioners {
		optioner.apply(newConfig)
	}

	return newConfig
}

func (config *serverConfig) clone() *serverConfig {
	newConfig := config

	return newConfig
}

func (optionerFunc serverConfigOptionerFunc) apply(
	config *serverConfig,
) {
	optionerFunc(config)
}

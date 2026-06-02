package config

//go:generate mockgen -destination=../test/v2/http_config.go -package=test -mock_names=HTTPConfigger=MockHTTPConfig . HTTPConfigger

import (
	"encoding/json"
	"github/M2A96/Monopoly.git/object"
)

type (
	// HTTPConfigger describes an HTTP endpoint configuration.
	HTTPConfigger interface {
		// GetAddr is the address to listen on for the HTTP server.
		GetAddr() string
	}

	// GetHTTPConfigger is an interface.
	GetHTTPConfigger interface {
		// GetHTTPConfigger is a function.
		GetHTTPConfigger() HTTPConfigger
	}

	httpConfig struct {
		addr string
	}

	httpConfigOptioner interface {
		apply(*httpConfig)
	}

	httpConfigOptionerFunc func(*httpConfig)
)

var (
	_ HTTPConfigger    = (*httpConfig)(nil)
	_ json.Marshaler   = (*httpConfig)(nil)
	_ object.GetMapper = (*httpConfig)(nil)
)

// NewHTTPConfig is a function.
func NewHTTPConfig(
	optioners ...httpConfigOptioner,
) *httpConfig {
	httpConfig := &httpConfig{
		addr: object.URIEmpty,
	}

	return httpConfig.WithOptioners(optioners...)
}

// WithHTTPConfigAddr is a function.
func WithHTTPConfigAddr(
	addr string,
) httpConfigOptioner {
	return httpConfigOptionerFunc(func(
		config *httpConfig,
	) {
		config.addr = addr
	})
}

// GetAddr is a function.
func (config *httpConfig) GetAddr() string {
	return config.addr
}

// GetMap is a function.
func (config *httpConfig) GetMap() map[string]any {
	return map[string]any{
		"addr": config.GetAddr(),
	}
}

// MarshalJSON is a function.
// read more https://pkg.go.dev/encoding/json#Marshaler
func (config *httpConfig) MarshalJSON() ([]byte, error) {
	return json.Marshal(config.GetMap())
}

// WithOptioners is a function.
func (config *httpConfig) WithOptioners(
	optioners ...httpConfigOptioner,
) *httpConfig {
	newConfig := config.clone()
	for _, optioner := range optioners {
		optioner.apply(newConfig)
	}

	return newConfig
}

func (config *httpConfig) clone() *httpConfig {
	newConfig := &httpConfig{
		addr: config.addr,
	}

	return newConfig
}

func (optionerFunc httpConfigOptionerFunc) apply(
	config *httpConfig,
) {
	optionerFunc(config)
}

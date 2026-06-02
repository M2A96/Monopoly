package config

//go:generate mockgen -destination=../test/v2/config.go -package=test -mock_names=Configger=MockConfig . Configger

import (
	"encoding/json"
	"github/M2A96/Monopoly.git/object"
)

type (
	// Configger interface is the core configuration.
	Configger interface {
		GetDatabaseConfigger
		// GetGatewayConfigger
		// GetGoogleConfigger
		GetHTTPConfigger
		GetLogConfigger
		GetOtelConfigger
		// GetPrometheusConfigger
		// GetPyroscopeConfigger
		// GetRedisConfigger
		GetRuntimeConfigger
		GetServerConfigger
	}

	// GetConfigger is an interface.
	GetConfigger interface {
		// GetConfigger is a function.
		GetConfigger() Configger
	}

	config struct {
		databaseConfigger DatabaseConfigger
		// gatewayConfigger    GatewayConfigger
		// googleConfigger     GoogleConfigger
		httpConfigger HTTPConfigger
		logConfigger  LogConfigger
		otelConfigger OtelConfigger
		// prometheusConfigger PrometheusConfigger
		// pyroscopeConfigger  PyroscopeConfigger
		// redisConfigger      RedisConfigger
		runtimeConfigger RuntimeConfigger
		serverConfigger  ServerConfigger
	}

	// ConfigOptioner is an interface.
	ConfigOptioner interface {
		apply(*config)
	}

	configOptionerFunc func(*config)
)

var (
	_ Configger            = (*config)(nil)
	_ GetDatabaseConfigger = (*config)(nil)
	// _ GetGatewayConfigger    = (*config)(nil)
	// _ GetGoogleConfigger     = (*config)(nil)
	_ GetHTTPConfigger = (*config)(nil)
	_ GetLogConfigger  = (*config)(nil)
	_ GetOtelConfigger = (*config)(nil)
	// _ GetPrometheusConfigger = (*config)(nil)
	// _ GetPyroscopeConfigger  = (*config)(nil)
	// _ GetRedisConfigger      = (*config)(nil)
	_ GetRuntimeConfigger = (*config)(nil)
	_ GetServerConfigger  = (*config)(nil)
	_ json.Marshaler      = (*config)(nil)
	_ object.GetMapper    = (*config)(nil)
)

// NewConfig constructs a Config struct which represents server settings,
// and populates it with default values.
func NewConfig(
	optioners ...ConfigOptioner,
) *config {
	config := &config{
		databaseConfigger: nil,
		// gatewayConfigger:    nil,
		// googleConfigger:     nil,
		httpConfigger: nil,
		logConfigger:  nil,
		// prometheusConfigger: nil,
		// redisConfigger:      nil,
		runtimeConfigger: nil,
		serverConfigger:  nil,
	}

	return config.WithOptioners(optioners...)
}

// WithConfigBinanceConfigger is a function.
// Func WithConfigBinanceConfigger(
// 	optioners ...binanceConfigOptioner,
// ) ConfigOptioner {
// 	return configOptionerFunc(func(
// 		config *config,
// 	) {
// 		config.binanceConfigger = NewBinanceConfig(optioners...)
// 	})
// }.

// WithConfigCasbinServerConfigger is a function.
// Func WithConfigCasbinServerConfigger(
// 	optioners ...casbinConfigOptioner,
// ) ConfigOptioner {
// 	return configOptionerFunc(func(
// 		config *config,
// 	) {
// 		config.casbinConfigger = NewCasbinConfig(optioners...)
// 	})
// }.

// WithConfigDatabaseConfigger is a function.
func WithConfigDatabaseConfigger(
	optioners ...databaseConfigOptioner,
) ConfigOptioner {
	return configOptionerFunc(func(
		config *config,
	) {
		config.databaseConfigger = NewDatabaseConfig(optioners...)
	})
}

// WithConfigGatewayConfigger is a function.
// func WithConfigGatewayConfigger(
// 	optioners ...gatewayConfigOptioner,
// ) ConfigOptioner {
// 	return configOptionerFunc(func(
// 		config *config,
// 	) {
// 		config.gatewayConfigger = NewGatewayConfig(optioners...)
// 	})
// }

// WithConfigGoogleConfigger is a function.
// func WithConfigGoogleConfigger(
// 	optioners ...googleConfigOptioner,
// ) ConfigOptioner {
// 	return configOptionerFunc(func(
// 		config *config,
// 	) {
// 		config.googleConfigger = NewGoogleConfig(optioners...)
// 	})
// }

// WithConfigHTTPConfigger is a function.
func WithConfigHTTPConfigger(
	optioners ...httpConfigOptioner,
) ConfigOptioner {
	return configOptionerFunc(func(
		config *config,
	) {
		config.httpConfigger = NewHTTPConfig(optioners...)
	})
}

// WithConfigLogConfigger is a function.
func WithConfigLogConfigger(
	optioners ...logConfigOptioner,
) ConfigOptioner {
	return configOptionerFunc(func(
		config *config,
	) {
		config.logConfigger = NewLogConfig(optioners...)
	})
}

// WithConfigOtelConfigger is a function.
func WithConfigOtelConfigger(
	optioners ...otelConfigOptioner,
) ConfigOptioner {
	return configOptionerFunc(func(
		config *config,
	) {
		config.otelConfigger = NewOtelConfig(optioners...)
	})
}

// WithConfigPODConfigger is a function.
// Func WithConfigPODConfigger(
// 	optioners ...podConfigOptioner,
// ) ConfigOptioner {
// 	return configOptionerFunc(func(
// 		config *config,
// 	) {
// 		config.podConfigger = NewPODConfig(optioners...)
// 	})
// }.

// WithConfigPrometheusConfigger is a function.
// func WithConfigPrometheusConfigger(
// 	optioners ...prometheusConfigOptioner,
// ) ConfigOptioner {
// 	return configOptionerFunc(func(
// 		config *config,
// 	) {
// 		config.prometheusConfigger = NewPrometheusConfig(optioners...)
// 	})
// }

// WithConfigPyroscopeConfigger is a function.
// func WithConfigPyroscopeConfigger(
// 	optioners ...pyroscopeConfigOptioner,
// ) ConfigOptioner {
// 	return configOptionerFunc(func(
// 		config *config,
// 	) {
// 		config.pyroscopeConfigger = NewPyroscopeConfig(optioners...)
// 	})
// }

// WithConfigRedisConfigger is a function.
// func WithConfigRedisConfigger(
// 	optioners ...redisConfigOptioner,
// ) ConfigOptioner {
// 	return configOptionerFunc(func(
// 		config *config,
// 	) {
// 		config.redisConfigger = NewRedisConfig(optioners...)
// 	})
// }

// WithConfigRedpandaConfigger is a function.
// Func WithConfigRedpandaConfigger(
// 	optioners ...redpandaConfigOptioner,
// ) ConfigOptioner {
// 	return configOptionerFunc(func(
// 		config *config,
// 	) {
// 		config.redpandaConfigger = NewRedpandaConfig(optioners...)
// 	})
// }.

// WithConfigRuntimeConfigger is a function.
func WithConfigRuntimeConfigger(
	optioners ...runtimeConfigOptioner,
) ConfigOptioner {
	return configOptionerFunc(func(
		config *config,
	) {
		config.runtimeConfigger = NewRuntimeConfig(optioners...)
	})
}

// WithConfigServerConfigger is a function.
func WithConfigServerConfigger(
	optioners ...serverConfigOptioner,
) ConfigOptioner {
	return configOptionerFunc(func(
		config *config,
	) {
		config.serverConfigger = NewServerConfig(optioners...)
	})
}

// WithConfigTapConfigger is a function.
// Func WithConfigTapConfigger(
// 	optioners ...tapConfigOptioner,
// ) ConfigOptioner {
// 	return configOptionerFunc(func(
// 		config *config,
// 	) {
// 		config.tapConfigger = NewTapConfig(optioners...)
// 	})
// }.

// GetBinanceConfigger is a function.
// Func (config *config) GetBinanceConfigger() BinanceConfigger {
// 	return config.binanceConfigger
// }.

// GetCasbinServerConfigger is a function.
// Func (config *config) GetCasbinServerConfigger() CasbinConfigger {
// 	return config.casbinConfigger
// }.

// GetDatabaseConfigger is a function.
func (config *config) GetDatabaseConfigger() DatabaseConfigger {
	return config.databaseConfigger
}

// GetGatewayConfigger is a function.
// func (config *config) GetGatewayConfigger() GatewayConfigger {
// 	return config.gatewayConfigger
// }

// GetGoogleConfigger is a function.
// func (config *config) GetGoogleConfigger() GoogleConfigger {
// 	return config.googleConfigger
// }

// GetHTTPConfigger is a function.
func (config *config) GetHTTPConfigger() HTTPConfigger {
	return config.httpConfigger
}

// GetLogConfigger is a function.
func (config *config) GetLogConfigger() LogConfigger {
	return config.logConfigger
}

// GetOtelConfigger is a function.
func (config *config) GetOtelConfigger() OtelConfigger {
	return config.otelConfigger
}

// // GetPODConfigger is a function.
// Func (config *config) GetPODConfigger() PODConfigger {
// 	return config.podConfigger
// }.

// GetPrometheusConfigger is a function.
// func (config *config) GetPrometheusConfigger() PrometheusConfigger {
// 	return config.prometheusConfigger
// }

// GetPyroscopeConfigger is a function.
// func (config *config) GetPyroscopeConfigger() PyroscopeConfigger {
// 	return config.pyroscopeConfigger
// }

// GetRedisConfigger is a function.
// func (config *config) GetRedisConfigger() RedisConfigger {
// 	return config.redisConfigger
// }

// GetRedpandaConfigger is a function.
// Func (config *config) GetRedpandaConfigger() RedpandaConfigger {
// 	return config.redpandaConfigger
// }.

// GetRuntimeConfigger is a function.
func (config *config) GetRuntimeConfigger() RuntimeConfigger {
	return config.runtimeConfigger
}

// GetServerConfigger is a function.
func (config *config) GetServerConfigger() ServerConfigger {
	return config.serverConfigger
}

// GetTapConfigger is a function.
// Func (config *config) GetTapConfigger() TapConfigger {
// 	return config.tapConfigger
// }.

// GetMap is a function.
func (config *config) GetMap() map[string]any {
	return map[string]any{
		"database_configger": config.GetDatabaseConfigger(),
		// "gateway_configger":    config.GetGatewayConfigger(),
		// "google_configger":     config.GetGoogleConfigger(),
		"http_configger": config.GetHTTPConfigger(),
		"log_configger":  config.GetLogConfigger(),
		"otel_configger": config.GetOtelConfigger(),
		// "prometheus_configger": config.GetPrometheusConfigger(),
		// "pyroscope_configger":  config.GetPyroscopeConfigger(),
		// "redis_configger":      config.GetRedisConfigger(),
		"runtime_configger": config.GetRuntimeConfigger(),
		"server_configger":  config.GetServerConfigger(),
	}
}

// MarshalJSON is a function.
// read more https://pkg.go.dev/encoding/json#Marshaler
func (config *config) MarshalJSON() ([]byte, error) {
	return json.Marshal(config.GetMap())
}

// WithOptioners is a function.
func (config *config) WithOptioners(
	optioners ...ConfigOptioner,
) *config {
	newConfig := config.clone()
	for _, optioner := range optioners {
		optioner.apply(newConfig)
	}

	return newConfig
}

func (config *config) clone() *config {
	newConfig := config

	return newConfig
}

func (optionerFunc configOptionerFunc) apply(
	config *config,
) {
	optionerFunc(config)
}

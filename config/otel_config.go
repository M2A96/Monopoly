package config

//go:generate mockgen -destination=../test/v2/otel_config.go -package=test -mock_names=OtelConfigger=MockOtelConfig . OtelConfigger

import (
	"encoding/json"
	"github/M2A96/Monopoly.git/object"
)

type (
	// OtelConfigger is an interface.
	OtelConfigger interface {
		// GetExporterOTLPTracesEndpoint is a function.
		GetExporterOTLPTracesEndpoint() string
		// GetInstrumentationName is a function.
		GetInstrumentationName() string
		// GetServiceInstanceID is a function.
		GetServiceInstanceID() string
		// GetServiceName is a function.
		GetServiceName() string
		// GetServiceNamespace is a function.
		GetServiceNamespace() string
		// GetServiceVersion is a function.
		GetServiceVersion() string
	}

	// GetOtelConfigger is an interface.
	GetOtelConfigger interface {
		// GetOtelConfigger is a function.
		GetOtelConfigger() OtelConfigger
	}

	otelConfig struct {
		exporterOTLPTracesEndpoint string
		instrumentationName        string
		serviceInstanceID          string
		serviceName                string
		serviceNamespace           string
		serviceVersion             string
	}

	otelConfigOptioner interface {
		apply(*otelConfig)
	}

	otelConfigOptionerFunc func(*otelConfig)
)

var (
	_ OtelConfigger    = (*otelConfig)(nil)
	_ json.Marshaler   = (*otelConfig)(nil)
	_ object.GetMapper = (*otelConfig)(nil)
)

// NewOtelConfig is a function.
func NewOtelConfig(
	optioners ...otelConfigOptioner,
) *otelConfig {
	otelConfig := &otelConfig{
		exporterOTLPTracesEndpoint: object.URIEmpty,
		instrumentationName:        object.URIEmpty,
		serviceInstanceID:          object.URIEmpty,
		serviceName:                object.URIEmpty,
		serviceNamespace:           object.URIEmpty,
		serviceVersion:             object.URIEmpty,
	}

	return otelConfig.WithOptioners(optioners...)
}

// WithOtelConfigExporterOTLPTracesEndpoint is a function.
func WithOtelConfigExporterOTLPTracesEndpoint(
	exporterOTLPTracesEndpoint string,
) otelConfigOptioner {
	return otelConfigOptionerFunc(func(
		config *otelConfig,
	) {
		config.exporterOTLPTracesEndpoint = exporterOTLPTracesEndpoint
	})
}

// WithOtelConfigInstrumentationName is a function.
func WithOtelConfigInstrumentationName(
	instrumentationName string,
) otelConfigOptioner {
	return otelConfigOptionerFunc(func(
		config *otelConfig,
	) {
		config.instrumentationName = instrumentationName
	})
}

// WithOtelConfigServiceInstanceID is a function.
func WithOtelConfigServiceInstanceID(
	serviceInstanceID string,
) otelConfigOptioner {
	return otelConfigOptionerFunc(func(
		config *otelConfig,
	) {
		config.serviceInstanceID = serviceInstanceID
	})
}

// WithOtelConfigServiceName is a function.
func WithOtelConfigServiceName(
	serviceName string,
) otelConfigOptioner {
	return otelConfigOptionerFunc(func(
		config *otelConfig,
	) {
		config.serviceName = serviceName
	})
}

// WithOtelConfigServiceNamespace is a function.
func WithOtelConfigServiceNamespace(
	serviceNamespace string,
) otelConfigOptioner {
	return otelConfigOptionerFunc(func(
		config *otelConfig,
	) {
		config.serviceNamespace = serviceNamespace
	})
}

// WithOtelConfigServiceVersion is a function.
func WithOtelConfigServiceVersion(
	serviceVersion string,
) otelConfigOptioner {
	return otelConfigOptionerFunc(func(
		config *otelConfig,
	) {
		config.serviceVersion = serviceVersion
	})
}

// GetExporterOTLPTracesEndpoint is a function.
func (config *otelConfig) GetExporterOTLPTracesEndpoint() string {
	return config.exporterOTLPTracesEndpoint
}

// GetInstrumentationName is a function.
func (config *otelConfig) GetInstrumentationName() string {
	return config.instrumentationName
}

// GetServiceInstanceID is a function.
func (config *otelConfig) GetServiceInstanceID() string {
	return config.serviceInstanceID
}

// GetServiceName is a function.
func (config *otelConfig) GetServiceName() string {
	return config.serviceName
}

// GetServiceNamespace is a function.
func (config *otelConfig) GetServiceNamespace() string {
	return config.serviceNamespace
}

// GetServiceVersion is a function.
func (config *otelConfig) GetServiceVersion() string {
	return config.serviceVersion
}

// GetMap is a function.
func (config *otelConfig) GetMap() map[string]any {
	return map[string]any{
		"exporter_otlp_traces_endpoint": config.GetExporterOTLPTracesEndpoint(),
		"instrumentation_name":          config.GetInstrumentationName(),
		"service_instance_id":           config.GetServiceInstanceID(),
		"service_name":                  config.GetServiceName(),
		"service_namespace":             config.GetServiceNamespace(),
		"service_version":               config.GetServiceVersion(),
	}
}

// MarshalJSON is a function.
// read more https://pkg.go.dev/encoding/json#Marshaler
func (config *otelConfig) MarshalJSON() ([]byte, error) {
	return json.Marshal(config.GetMap())
}

// WithOptioners is a function.
func (config *otelConfig) WithOptioners(
	optioners ...otelConfigOptioner,
) *otelConfig {
	newConfig := config.clone()
	for _, optioner := range optioners {
		optioner.apply(newConfig)
	}

	return newConfig
}

func (config *otelConfig) clone() *otelConfig {
	newConfig := config

	return newConfig
}

func (optionerFunc otelConfigOptionerFunc) apply(
	config *otelConfig,
) {
	optionerFunc(config)
}

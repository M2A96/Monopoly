package util

import (
	"context"

	"github/M2A96/Monopoly.git/config"
	"github/M2A96/Monopoly.git/log"
	"github/M2A96/Monopoly.git/object"

	pyroscopeOTELProfiling "github.com/grafana/otel-profiling-go"
	"go.opentelemetry.io/contrib/propagators/b3"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdkTrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	otelTrace "go.opentelemetry.io/otel/trace"
)

type (
	// Tracer is an interface.
	Tracer interface {
		// GetTracer is a function.
		GetTracer() otelTrace.Tracer
	}

	// GetTracer is an interface.
	GetTracer interface {
		// GetTracer is a function.
		GetTracer() otelTrace.Tracer
	}
)

// NewTracer is a function.
func NewTracer(
	ctx context.Context,
	configConfigger config.Configger,
	logRuntimeLogger log.RuntimeLogger,
) otelTrace.Tracer {
	runtimeContext := NewRuntimeContext(ctx)
	spanContext := NewSpanContext(nil)
	fields := map[string]any{
		"name":   "NewOpenTelemetry",
		"rt_ctx": runtimeContext,
		"sp_ctx": spanContext,
		"config": configConfigger,
	}

	logRuntimeLogger.
		WithFields(fields).
		Info(object.URIEmpty)

	sdkResourceResource := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceInstanceIDKey.String(
			configConfigger.GetOtelConfigger().GetServiceInstanceID(),
		),
		semconv.ServiceNameKey.String(configConfigger.GetOtelConfigger().GetServiceName()),
		semconv.ServiceNamespaceKey.String(
			configConfigger.GetOtelConfigger().GetServiceNamespace(),
		),
		semconv.ServiceVersionKey.String(
			configConfigger.GetOtelConfigger().GetServiceVersion(),
		),
	)

	logRuntimeLogger.
		WithFields(fields).
		WithField(object.URIFieldSDKResourceResource, sdkResourceResource).
		Debug(object.URIEmpty)

	newSDKResourceResource, err := resource.Merge(
		resource.Default(),
		sdkResourceResource,
	)
	if err != nil {
		logRuntimeLogger.
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error(object.ErrSDKResourceMerge.Error())
	}

	logRuntimeLogger.
		WithFields(fields).
		WithField(object.URIFieldSDKResourceResource, newSDKResourceResource).
		Debug(object.URIEmpty)

	sdkResourceResource3, err := resource.New(
		ctx,
		resource.WithContainer(),
		resource.WithFromEnv(),
		resource.WithOS(),
		resource.WithProcess(),
		/*
			otelSDKResource.WithAttributes
			otelSDKResource.WithContainerID
			otelSDKResource.WithDetectors
			otelSDKResource.WithHost
			otelSDKResource.WithHostID
			otelSDKResource.WithOSDescription
			otelSDKResource.WithOSType
			otelSDKResource.WithProcessCommandArgs
			otelSDKResource.WithProcessExecutableName
			otelSDKResource.WithProcessExecutablePath
			otelSDKResource.WithProcessOwner
			otelSDKResource.WithProcessPID
			otelSDKResource.WithProcessRuntimeDescription
			otelSDKResource.WithProcessRuntimeName
			otelSDKResource.WithProcessRuntimeVersion
			otelSDKResource.WithSchemaURL
			otelSDKResource.WithTelemetrySDK
		*/
	)
	if err != nil {
		logRuntimeLogger.
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error(object.ErrSDKResourceNew.Error())
	}

	logRuntimeLogger.
		WithFields(fields).
		WithField(object.URIFieldSDKResourceResource, sdkResourceResource3).
		Debug(object.URIEmpty)

	sdkResourceResource4, err := resource.Merge(newSDKResourceResource, sdkResourceResource3)
	if err != nil {
		logRuntimeLogger.
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error(object.ErrSDKResourceMerge.Error())
	}

	logRuntimeLogger.
		WithFields(fields).
		WithField(object.URIFieldSDKResourceResource, sdkResourceResource4).
		Debug(object.URIEmpty)

	exporterOLTPTrace, err := otlptrace.New(
		ctx,
		otlptracegrpc.NewClient(
			otlptracegrpc.WithEndpoint(
				configConfigger.GetOtelConfigger().GetExporterOTLPTracesEndpoint(),
			),
			otlptracegrpc.WithInsecure(),
			otlptracegrpc.WithTimeout(object.NUMGRPCClientTimeout),
			/*
				otelExportersOTLPTraceGRPC.WithCompressor
				otelExportersOTLPTraceGRPC.WithDialOption
				otelExportersOTLPTraceGRPC.WithGRPCConn
				otelExportersOTLPTraceGRPC.WithHeaders
				otelExportersOTLPTraceGRPC.WithReconnectionPeriod
				otelExportersOTLPTraceGRPC.WithRetry
				otelExportersOTLPTraceGRPC.WithServiceConfig
				otelExportersOTLPTraceGRPC.WithTLSCredentials
			*/
		),
	)
	if err != nil {
		logRuntimeLogger.
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error(object.ErrOTELExporterOTLPTraceNew.Error())
	}

	logRuntimeLogger.
		WithFields(fields).
		WithField(object.URIFieldExporterOTLPTrace, exporterOLTPTrace).
		Debug(object.URIEmpty)

	sdkTraceTracerProvider := sdkTrace.NewTracerProvider(
		sdkTrace.WithBatcher(
			exporterOLTPTrace,
			/*
				otelSDKTrace.WithBatchTimeout
				otelSDKTrace.WithBlocking
				otelSDKTrace.WithExportTimeout
				otelSDKTrace.WithMaxExportBatchSize
				otelSDKTrace.WithMaxQueueSize
			*/
		),
		sdkTrace.WithResource(sdkResourceResource4),
		/*
			otelSDKTrace.WithIDGenerator
			otelSDKTrace.WithRawSpanLimits
			otelSDKTrace.WithSampler
			otelSDKTrace.WithSpanLimits
			otelSDKTrace.WithSpanProcessor
			otelSDKTrace.WithSyncer
		*/
	)

	logRuntimeLogger.
		WithFields(fields).
		WithField(object.URIFieldTracerProvider, sdkTraceTracerProvider).
		Debug(object.URIEmpty)

	otel.SetTracerProvider(pyroscopeOTELProfiling.NewTracerProvider(
		sdkTraceTracerProvider,
		pyroscopeOTELProfiling.WithAppName(configConfigger.GetOtelConfigger().GetServiceName()),
		pyroscopeOTELProfiling.WithRootSpanOnly(true),
		pyroscopeOTELProfiling.WithAddSpanName(true),
		// PyroscopeOTELProfiling.WithPyroscopeURL(
		// 	configConfigger.GetPyroscopeConfigger().GetClientURL(),
		// ),.
		pyroscopeOTELProfiling.WithProfileBaselineLabels(map[string]string{}),
		pyroscopeOTELProfiling.WithProfileBaselineURL(true),
		pyroscopeOTELProfiling.WithProfileURL(true),
	))
	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			b3.New(
				b3.WithInjectEncoding(b3.B3SingleHeader),
			),
		),
	)

	traceTracer := otel.Tracer(
		configConfigger.GetOtelConfigger().GetInstrumentationName(),
		/*
			otelTrace.WithInstrumentationVersion
			otelTrace.WithInstrumentationAttributes
			otelTrace.WithSchemaURL
		*/
	)

	logRuntimeLogger.
		WithFields(fields).
		WithField(object.URIFieldTracer, traceTracer).
		Debug(object.URIEmpty)

	go func() {
		<-ctx.Done()

		logRuntimeLogger.
			WithFields(fields).
			Debug(`shutting down gracefully the tracing`)

		ctxWT, ctxWTCancelFunc := context.WithTimeout(ctx, object.NUMSystemGracefulShutdown)
		defer ctxWTCancelFunc()

		if err = sdkTraceTracerProvider.Shutdown(ctxWT); err != nil {
			logRuntimeLogger.
				WithFields(fields).
				WithField(object.URIFieldError, err).
				Error(object.ErrTracerProviderShutdown.Error())
		}
	}()

	return traceTracer
}

package main

import (
	"context"
	"encoding/json"
	"github/M2A96/Monopoly.git/config"
	"github/M2A96/Monopoly.git/internal/api"
	"github/M2A96/Monopoly.git/internal/core/domain/services"
	"github/M2A96/Monopoly.git/internal/infrastructure/repository"
	"github/M2A96/Monopoly.git/log"
	"github/M2A96/Monopoly.git/object"
	"github/M2A96/Monopoly.git/util"
	"net/http"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"go.opentelemetry.io/otel/codes"

	"go.opentelemetry.io/otel/trace"

	"go.uber.org/zap"

	"github.com/labstack/echo/v4"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/spf13/viper"
)

func main() {

	ctx := context.Background()
	http.DefaultClient.Timeout = object.NUMHTTPClientTimeout

	viper.AutomaticEnv()
	viper.SetDefault("DATABASE_DSN", "postgresql://root@127.0.0.1:26257/defaultdb?sslmode=disable")
	viper.SetDefault("LOG_FILE", "file.log")
	viper.SetDefault("LOG_FORMAT", "json")
	viper.SetDefault("LOG_LEVEL", "debug")
	viper.SetDefault("LOG_SQL_SLOW_THRESHOLD", "1s")
	viper.SetDefault("LOG_MAX_AGE", 0)
	viper.SetDefault("LOG_MAX_BACKUPS", 0)
	viper.SetDefault("LOG_MAX_SIZE", 1e2)
	viper.SetDefault("LOG_COMPRESS", false)
	viper.SetDefault("LOG_LOCAL_TIME", false)
	viper.SetDefault("LOG_ROTATION", false)
	viper.SetDefault("LOG_STDOUT", true)
	viper.SetDefault("OTEL_EXPORTER_OTLP_TRACES_ENDPOINT", "otelcol:4317")
	viper.SetDefault("OTEL_INSTRUMENTATION_NAME", object.URIEmpty)
	viper.SetDefault("OTEL_SERVICE_INSTANCE_ID", "00000000-0000-0000-0000-000000000000")
	viper.SetDefault("OTEL_SERVICE_NAME", "server")
	viper.SetDefault("OTEL_SERVICE_NAMESPACE", "server")
	viper.SetDefault("OTEL_SERVICE_VERSION", "0000.00.00-rc")
	viper.SetDefault("RUNTIME_VALIDATE_MAP_RULES", `{"rules":[{"version":"1"}]}`)
	viper.SetDefault("SERVER_ENDPOINT_ADDR", ":9090")
	viper.SetDefault("SERVER_ENDPOINT_NETWORK", "tcp")

	configConfig := config.NewConfig(
		config.WithConfigDatabaseConfigger(
			config.WithDatabaseConfigDSN(
				viper.GetString("DATABASE_DSN"),
			),
		),
		config.WithConfigLogConfigger(
			config.WithLogConfigFile(viper.GetString("LOG_FILE")),
			config.WithLogConfigFormat(viper.GetString("LOG_FORMAT")),
			config.WithLogConfigLevel(viper.GetString("LOG_LEVEL")),
			config.WithLogConfigSQLSlowThreshold(viper.GetDuration("LOG_SQL_SLOW_THRESHOLD")),
			config.WithLogConfigMaxAge(viper.GetInt("LOG_MAX_AGE")),
			config.WithLogConfigMaxBackups(viper.GetInt("LOG_MAX_BACKUPS")),
			config.WithLogConfigMaxSize(viper.GetInt("LOG_MAX_SIZE")),
			config.WithLogConfigCompress(viper.GetBool("LOG_COMPRESS")),
			config.WithLogConfigLocalTime(viper.GetBool("LOG_LOCAL_TIME")),
			config.WithLogConfigRotation(viper.GetBool("LOG_ROTATION")),
			config.WithLogConfigStdout(viper.GetBool("LOG_STDOUT")),
		),
		config.WithConfigOtelConfigger(
			config.WithOtelConfigExporterOTLPTracesEndpoint(
				viper.GetString("OTEL_EXPORTER_OTLP_TRACES_ENDPOINT"),
			),
			config.WithOtelConfigInstrumentationName(viper.GetString("OTEL_INSTRUMENTATION_NAME")),
			config.WithOtelConfigServiceInstanceID(viper.GetString("OTEL_SERVICE_INSTANCE_ID")),
			config.WithOtelConfigServiceName(viper.GetString("OTEL_SERVICE_NAME")),
			config.WithOtelConfigServiceNamespace(viper.GetString("OTEL_SERVICE_NAMESPACE")),
			config.WithOtelConfigServiceVersion(viper.GetString("OTEL_SERVICE_VERSION")),
		),
		config.WithConfigRuntimeConfigger(
			config.WithRuntimeConfigClientPaginationRequestSizeMax(
				viper.GetInt("RUNTIME_CLIENT_PAGINATION_REQUEST_SIZE_MAX"),
			),
			config.WithRuntimeConfigValidateMapRules(
				util.Cast(viper.GetStringMap("RUNTIME_VALIDATE_MAP_RULES")),
			),
		),
		config.WithConfigServerConfigger(
			config.WithServerConfigEndpointConfigger(
				config.NewEndpointConfig(
					config.WithEndpointConfigAddr(viper.GetString("SERVER_ENDPOINT_ADDR")),
					config.WithEndpointConfigNetwork(viper.GetString("SERVER_ENDPOINT_NETWORK")),
				),
			),
		),
	)

	zapLogger := log.NewZapLogger(configConfig)
	objectTime := object.NewTime()
	logGormLog := log.NewGormLog(
		configConfig,
		map[string]any{},
		objectTime,
		zapLogger.WithOptions(zap.AddCallerSkip(4)),
	)
	logRuntimeLog := log.NewRuntimeLog(
		configConfig,
		map[string]any{},
		zapLogger.WithOptions(zap.AddCallerSkip(1)),
	)
	objectUUID := object.NewUUID()
	traceTracer := util.NewTracer(ctx, configConfig, logRuntimeLog)

	var traceSpan trace.Span
	ctx, traceSpan = traceTracer.Start(
		ctx,
		"main",
		trace.WithSpanKind(trace.SpanKindServer))

	defer traceSpan.End()

	utilRuntimeContext := util.NewRuntimeContext(ctx)
	utilSpanContext := util.NewSpanContext(traceSpan)
	fields := map[string]any{
		"name":   "main",
		"rt_ctx": utilRuntimeContext,
		"sp_ctx": utilSpanContext,
		"config": configConfig,
	}

	logRuntimeLog.
		WithFields(fields).
		Info(object.URIEmpty)

	privateJWKSet := jwk.NewSet()

	for _, value := range object.JWKPubPrivKeys {
		jsonValue, _ := json.Marshal(value)

		privateJWKKey, err := jwk.ParseKey(jsonValue)
		if err != nil {
			logRuntimeLog.
				WithFields(fields).
				WithField(object.URIFieldError, err).
				Error(object.ErrJWKKeyParseKey.Error())
			traceSpan.RecordError(err)
			traceSpan.SetStatus(codes.Error, object.ErrJWKKeyParseKey.Error())

			return
		}

		if errJWKSetAddKey := privateJWKSet.AddKey(privateJWKKey); errJWKSetAddKey != nil {
			logRuntimeLog.
				WithFields(fields).
				WithField(object.URIFieldError, errJWKSetAddKey).
				Error(object.ErrJWKSetAddKey.Error())
			traceSpan.RecordError(errJWKSetAddKey)
			traceSpan.SetStatus(codes.Error, object.ErrJWKSetAddKey.Error())

			return
		}
	}

	gormDB, err := gorm.Open(
		postgres.Open(configConfig.GetDatabaseConfigger().GetDSN()),
		&gorm.Config{
			SkipDefaultTransaction: true,
			NamingStrategy:         nil,
			FullSaveAssociations:   false,
			Logger:                 logGormLog,
			NowFunc: func() time.Time {
				return object.NewTime().NowUTC()
			},
			DryRun:                                   false,
			PrepareStmt:                              true,
			DisableAutomaticPing:                     false,
			DisableForeignKeyConstraintWhenMigrating: true,
			IgnoreRelationshipsWhenMigrating:         false,
			DisableNestedTransaction:                 true,
			AllowGlobalUpdate:                        false,
			QueryFields:                              true,
			CreateBatchSize:                          0,
			TranslateError:                           false,
			ClauseBuilders:                           map[string]clause.ClauseBuilder{},
			ConnPool:                                 nil,
			Dialector:                                nil,
			Plugins:                                  map[string]gorm.Plugin{},
		},
	)
	if err != nil {
		logRuntimeLog.
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error(object.ErrGormOpen.Error())
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, object.ErrGormOpen.Error())

		return
	}

	sqlDB, err := gormDB.DB()
	if err != nil {
		logRuntimeLog.
			WithFields(fields).
			WithField(object.URIFieldError, err).
			Error(object.ErrSQLDBFromGormDB.Error())
		traceSpan.RecordError(err)
		traceSpan.SetStatus(codes.Error, object.ErrSQLDBFromGormDB.Error())
	}

	go func() {
		<-ctx.Done()

		logRuntimeLog.
			WithFields(fields).
			Debug(`closing the sql DB`)

		if err = sqlDB.Close(); err != nil {
			logRuntimeLog.
				WithFields(fields).
				WithField(object.URIFieldError, err).
				Error(object.ErrSQLDBClose.Error())
			traceSpan.RecordError(err)
			traceSpan.SetStatus(codes.Error, object.ErrSQLDBClose.Error())

			return
		}
	}()

	repositoryRepository := repository.NewRepositories(
		repository.WithGameRepositorier(
			configConfig,
			logRuntimeLog,
			objectUUID,
			traceTracer,
			repository.WithGameRepositoryDB(gormDB),
			repository.WithGameRepositoryTimer(objectTime),
		),
	)

	serviceService := services.NewServices(
		services.WithGameService(
			configConfig,
			logRuntimeLog,
			objectUUID,
			traceTracer,
			services.WithGameServiceGameRepositorier(
				repositoryRepository.GetGameRepositorier(),
			),
		),
	)

	gameHandler := api.NewGameHandler(
		configConfig,
		logRuntimeLog,
		objectUUID,
		traceTracer,
		api.WithGameHandlerGameServicer(
			serviceService.GetGameServicer(),
		),
	)

	e := echo.New()
	gameHandler.RegisterRoutes(e)
	e.Start(
		configConfig.GetServerConfigger().GetEndpointConfigger().GetAddr(),
	)

}

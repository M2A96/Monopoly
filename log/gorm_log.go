package log

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github/M2A96/Monopoly.git/config"
	"github/M2A96/Monopoly.git/object"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type gormLog struct {
	configConfigger config.Configger
	fields          map[string]any
	objectTimer     object.Timer
	zapLogger       *zap.Logger
	logger.Config
}

var (
	_ GetLogger           = (*gormLog)(nil)
	_ config.GetConfigger = (*gormLog)(nil)
	_ logger.Interface    = (*gormLog)(nil)
	_ object.GetTimer     = (*gormLog)(nil)
)

// NewGormLog is function.
func NewGormLog(
	configConfigger config.Configger,
	fields map[string]any,
	objectTimer object.Timer,
	zapLogger *zap.Logger,
) *gormLog {
	return &gormLog{
		configConfigger: configConfigger,
		objectTimer:     objectTimer,
		fields:          fields,
		zapLogger:       zapLogger,
		Config: logger.Config{
			SlowThreshold:             configConfigger.GetLogConfigger().GetSQLSlowThreshold(),
			Colorful:                  false,
			IgnoreRecordNotFoundError: false,
			ParameterizedQueries:      false,
			LogLevel:                  logger.Info,
		},
	}
}

// GetLogger is a function.
func (log *gormLog) GetLogger() *zap.Logger {
	return log.zapLogger
}

// GetConfigger is a function.
func (log *gormLog) GetConfigger() config.Configger {
	return log.configConfigger
}

// LogMode is a function.
func (log *gormLog) LogMode(
	level logger.LogLevel,
) logger.Interface {
	var zapCoreLevel zapcore.Level

	switch level {
	case logger.Error:
		zapCoreLevel = zapcore.ErrorLevel

	case logger.Warn:
		zapCoreLevel = zapcore.WarnLevel

	case logger.Info:
		zapCoreLevel = zapcore.DebugLevel

	default:
		zapCoreLevel = zapcore.DebugLevel
	}

	configConfig := config.NewConfig(
		config.WithConfigLogConfigger(
			config.WithLogConfigFile(log.GetConfigger().GetLogConfigger().GetFile()),
			config.WithLogConfigFormat(log.GetConfigger().GetLogConfigger().GetFormat()),
			config.WithLogConfigLevel(zapCoreLevel.String()),
			config.WithLogConfigSQLSlowThreshold(
				log.GetConfigger().GetLogConfigger().GetSQLSlowThreshold(),
			),
			config.WithLogConfigMaxAge(log.GetConfigger().GetLogConfigger().GetMaxAge()),
			config.WithLogConfigMaxBackups(
				log.GetConfigger().GetLogConfigger().GetMaxBackups(),
			),
			config.WithLogConfigMaxSize(log.GetConfigger().GetLogConfigger().GetMaxSize()),
			config.WithLogConfigCompress(log.GetConfigger().GetLogConfigger().GetCompress()),
			config.WithLogConfigLocalTime(log.GetConfigger().GetLogConfigger().GetLocalTime()),
			config.WithLogConfigRotation(log.GetConfigger().GetLogConfigger().GetRotation()),
			config.WithLogConfigStdout(log.GetConfigger().GetLogConfigger().GetStdout()),
		),
	)

	zapLogger := NewZapLogger(configConfig)

	return NewGormLog(
		configConfig,
		map[string]any{},
		log.GetTimer(),
		zapLogger,
	)
}

// Info is a function.
func (log *gormLog) Info(
	_ context.Context,
	format string,
	args ...any,
) {
	if log.GetLogger().Core().Enabled(zap.InfoLevel) {
		msg := fmt.Sprintf(format, args...)
		log.GetLogger().Info(msg, GetFileLine())
	}
}

// Warn is a function.
func (log *gormLog) Warn(
	_ context.Context,
	format string,
	args ...any,
) {
	if log.GetLogger().Core().Enabled(zap.WarnLevel) {
		msg := fmt.Sprintf(format, args...)
		log.GetLogger().Warn(msg, GetFileLine())
	}
}

// Error is a function.
func (log *gormLog) Error(
	_ context.Context,
	format string,
	args ...any,
) {
	if log.GetLogger().Core().Enabled(zap.ErrorLevel) {
		msg := fmt.Sprintf(format, args...)
		log.GetLogger().Error(msg, GetFileLine())
	}
}

// Trace is a function.
func (log *gormLog) Trace(
	ctx context.Context,
	begin time.Time,
	fc func() (sql string, rowsAffected int64),
	err error,
) {
	elapsed := log.GetTimer().Since(begin)
	sql, rows := fc()
	fields := map[string]any{
		"elapsed": float64(elapsed.Nanoseconds()) / 1e6,
		"rows":    rows,
		"sql":     sql,
	}

	if rows == -1 {
		fields[object.URIFieldRows] = -1
	}

	switch {
	case err != nil && (!errors.Is(err, gorm.ErrRecordNotFound) || !log.IgnoreRecordNotFoundError):
		fields[object.URIFieldError] = err
		log.
			WithFields(fields).
			Error(ctx, "%v", object.ErrSQL.Error())

	case log.SlowThreshold < elapsed && log.SlowThreshold != 0:
		fields["sql_slow_threshold"] = true
		log.
			WithFields(fields).
			Warn(ctx, "%v", object.URIEmpty)

	default:
		log.
			WithFields(fields).
			Info(ctx, "%v", object.URIEmpty)
	}
}

// WithFields implements RuntimeLogger.WithFieldsV1.
func (log *gormLog) WithFields(
	fields map[string]any,
) logger.Interface {
	newFields := make(map[string]any, len(fields))
	zapcoreFields := make([]zap.Field, 0, len(fields))

	for key, value := range fields {
		newFields[key] = value
		zapcoreFields = append(zapcoreFields, zap.Any(key, value))
	}

	return NewGormLog(
		log.GetConfigger(),
		newFields,
		log.GetTimer(),
		log.GetLogger().With(zapcoreFields...),
	)
}

// GetTimer is a function.
func (log *gormLog) GetTimer() object.Timer {
	return log.objectTimer
}

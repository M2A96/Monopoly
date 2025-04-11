package log

//go:generate mockgen -destination=../test/v2/runtime_log.go -package=test -mock_names=RuntimeLogger=MockRuntimeLog . RuntimeLogger

import (
	"fmt"
	"net/http"

	"github/M2A96/Monopoly.git/config"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type (
	// RuntimeLogger exposes a logging framework to use in modules.
	// It exposes level-specific logging functions and a set of common functions for compatibility.
	RuntimeLogger interface {
		// Log a message with optional arguments at DEBUG level. Arguments are handled in the manner of fmt.Printf.
		Debug(
			string,
			...any,
		)
		// Log a message with optional arguments at INFO level. Arguments are handled in the manner of fmt.Printf.
		Info(
			string,
			...any,
		)
		// Log a message with optional arguments at WARN level. Arguments are handled in the manner of fmt.Printf.
		Warn(
			string,
			...any,
		)
		// Log a message with optional arguments at ERROR level. Arguments are handled in the manner of fmt.Printf.
		Error(
			string,
			...any,
		)
		// Log a message with optional arguments at FATAL level. Arguments are handled in the manner of fmt.Printf.
		Fatal(
			string,
			...any,
		)
		// Return a runtimeLog with the specified field set so that they are included in subsequent logging calls.
		WithField(
			string,
			any,
		) RuntimeLogger
		// Return a runtimeLog with the specified fields set so that they are included in subsequent logging calls.
		WithFields(
			map[string]any,
		) RuntimeLogger
		// Returns the fields set in this RuntimeLogger.
		Fields() map[string]any
	}

	// GetRuntimeLogger is an interface.
	GetRuntimeLogger interface {
		// GetRuntimeLogger is a function.
		GetRuntimeLogger() RuntimeLogger
	}

	runtimeLog struct {
		configConfigger config.Configger
		fields          map[string]any
		zapLogger       *zap.Logger
	}
)

var (
	_ GetLogger            = (*runtimeLog)(nil)
	_ RuntimeLogger        = (*runtimeLog)(nil)
	_ config.GetConfigger  = (*runtimeLog)(nil)
	_ zapcore.LevelEnabler = (*runtimeLog)(nil)
)

// NewRuntimeLog is function.
func NewRuntimeLog(
	configConfigger config.Configger,
	fields map[string]any,
	zapLogger *zap.Logger,
) *runtimeLog {
	return &runtimeLog{
		configConfigger: configConfigger,
		fields:          fields,
		zapLogger:       zapLogger,
	}
}

// GetLogger is a function.
func (log *runtimeLog) GetLogger() *zap.Logger {
	return log.zapLogger
}

// GetConfigger is a function.
func (log *runtimeLog) GetConfigger() config.Configger {
	return log.configConfigger
}

// Debug implements RuntimeLogger.Debug.
func (log *runtimeLog) Debug(
	format string,
	args ...any,
) {
	if log.GetLogger().Core().Enabled(zap.DebugLevel) {
		msg := fmt.Sprintf(format, args...)
		log.GetLogger().Debug(msg)
	}
}

// Info implements RuntimeLogger.Info.
func (log *runtimeLog) Info(
	format string,
	args ...any,
) {
	if log.GetLogger().Core().Enabled(zap.InfoLevel) {
		msg := fmt.Sprintf(format, args...)
		log.GetLogger().Info(msg, GetFileLine())
	}
}

// Warn implements RuntimeLogger.Warn.
func (log *runtimeLog) Warn(
	format string,
	args ...any,
) {
	if log.GetLogger().Core().Enabled(zap.WarnLevel) {
		msg := fmt.Sprintf(format, args...)
		log.GetLogger().Warn(msg, GetFileLine())
	}
}

// Error implements RuntimeLogger.Error.
func (log *runtimeLog) Error(
	format string,
	args ...any,
) {
	if log.GetLogger().Core().Enabled(zap.ErrorLevel) {
		msg := fmt.Sprintf(format, args...)
		log.GetLogger().Error(msg, GetFileLine())
	}
}

// Fatal implements RuntimeLogger.Error.
func (log *runtimeLog) Fatal(
	format string,
	args ...any,
) {
	if log.GetLogger().Core().Enabled(zap.FatalLevel) {
		msg := fmt.Sprintf(format, args...)
		log.GetLogger().Fatal(msg, GetFileLine())
	}
}

// WithField implements RuntimeLogger.WithFieldV1.
func (log *runtimeLog) WithField(
	key string,
	value any,
) RuntimeLogger {
	return log.WithFields(map[string]any{
		key: value,
	})
}

// WithFields implements RuntimeLogger.WithFieldsV1.
func (log *runtimeLog) WithFields(
	fields map[string]any,
) RuntimeLogger {
	newFields := make(map[string]any, len(fields)+len(log.fields))
	zapcoreFields := make([]zap.Field, 0, len(fields)+len(log.fields))

	for key, value := range log.fields {
		newFields[key] = value
	}

	for key, value := range fields {
		if value, ok := value.(http.Request); ok && key == "req" {
			newFields["req_body"] = value.Body
			newFields["req_cookies"] = value.Cookies()
			newFields["req_header"] = value.Header
			zapcoreFields = append(zapcoreFields, zap.Any(key, value))

			continue
		}

		newFields[key] = value
		zapcoreFields = append(zapcoreFields, zap.Any(key, value))
	}

	return NewRuntimeLog(
		log.GetConfigger(),
		newFields,
		log.GetLogger().With(zapcoreFields...),
	)
}

// Fields implements RuntimeLogger.Fields.
func (log *runtimeLog) Fields() map[string]any {
	return log.fields
}

// Enabled is a function.
func (log *runtimeLog) Enabled(
	lvl zapcore.Level,
) bool {
	zapcoreLevel, err := zapcore.ParseLevel(log.GetConfigger().GetLogConfigger().GetLevel())
	if err != nil {
		zapcoreLevel = zapcore.InfoLevel
	}

	return zapcoreLevel <= lvl
}

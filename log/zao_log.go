package log

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github/M2A96/Monopoly.git/config"
	"github/M2A96/Monopoly.git/object"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type (
	// GetLogger is an enum.
	GetLogger interface {
		GetLogger() *zap.Logger
	}

	// LoggerFormat is an enum.
	LoggerFormat int8
)

const (
	// JSONFormat is a LoggingFormat.
	JSONFormat LoggerFormat = iota - 1
	// StackdriverFormat is a LoggingFormat.
	StackdriverFormat
)

// ParseLoggerFormat is a function.
func ParseLoggerFormat(
	text string,
) (LoggerFormat, error) {
	var loggerFormat LoggerFormat
	err := loggerFormat.UnmarshalText([]byte(text))

	return loggerFormat, err
}

// UnmarshalText is a function.
func (log *LoggerFormat) UnmarshalText(
	text []byte,
) error {
	if !log.IsUnmarshalText(text) && !log.IsUnmarshalText(bytes.ToLower(text)) {
		return fmt.Errorf("%w: %s", object.ErrBase64Decode2, text)
	}

	return nil
}

// IsUnmarshalText is a function.
func (log *LoggerFormat) IsUnmarshalText(
	text []byte,
) bool {
	switch string(text) {
	case object.URIEmpty, "JSON", "json":
		*log = JSONFormat

	case "STACKDRIVER", "stackdriver":
		*log = StackdriverFormat

	default:
		return false
	}

	return true
}

// NewZapLogger is a function.
func NewZapLogger(
	configConfigger config.Configger,
) *zap.Logger {
	zapcoreLevel, err := zapcore.ParseLevel(configConfigger.GetLogConfigger().GetLevel())
	if err != nil {
		zapcoreLevel = zapcore.InfoLevel
	}

	loggerFormat, err := ParseLoggerFormat(configConfigger.GetLogConfigger().GetFormat())
	if err != nil {
		loggerFormat = JSONFormat
	}

	consoleZapLogger := NewJSONLogger(os.Stdout, zapcoreLevel, loggerFormat)

	var jsonFileZapLogger *zap.Logger

	if configConfigger.GetLogConfigger().GetRotation() {
		jsonFileZapLogger = NewRotatingJSONFileLogger(
			consoleZapLogger,
			configConfigger,
			zapcoreLevel,
			loggerFormat,
		)
	} else {
		jsonFileZapLogger = NewJSONFileLogger(consoleZapLogger, configConfigger, zapcoreLevel, loggerFormat)
	}

	if jsonFileZapLogger != nil {
		multiZapLogger := NewMultiLogger(consoleZapLogger, jsonFileZapLogger)

		if configConfigger.GetLogConfigger().GetStdout() {
			RedirectStdLog(multiZapLogger)

			return multiZapLogger
		}

		RedirectStdLog(jsonFileZapLogger)

		return jsonFileZapLogger
	}

	RedirectStdLog(consoleZapLogger)

	return consoleZapLogger
}

// NewJSONFileLogger is a function.
func NewJSONFileLogger(
	consoleZapLogger *zap.Logger,
	configConfigger config.Configger,
	level zapcore.Level,
	format LoggerFormat,
) *zap.Logger {
	fileName := configConfigger.GetLogConfigger().GetFile()

	if fileName == object.URIEmpty {
		return nil
	}

	osFile, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o666)
	if err != nil {
		consoleZapLogger.Fatal("Could not create log file", zap.Error(err))

		return nil
	}

	return NewJSONLogger(osFile, level, format)
}

// NewRotatingJSONFileLogger is a function.
func NewRotatingJSONFileLogger(
	consoleZapLogger *zap.Logger,
	configConfigger config.Configger,
	level zapcore.Level,
	format LoggerFormat,
) *zap.Logger {
	fileName := configConfigger.GetLogConfigger().GetFile()
	if fileName == object.URIEmpty {
		consoleZapLogger.Fatal("Rotating log file is enabled but log file name is empty")

		return nil
	}

	logDir := filepath.Dir(fileName)
	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		if err = os.MkdirAll(logDir, 0o755); err != nil {
			consoleZapLogger.Fatal("Could not create log directory", zap.Error(err))

			return nil
		}
	}

	jsonEncoder := NewJSONEncoder(format)

	// Lumberjack.Logger is already safe for concurrent use, so we don't need to lock it.
	zapcoreWriteSyncer := zapcore.AddSync(&lumberjack.Logger{
		Filename:   fileName,
		MaxSize:    configConfigger.GetLogConfigger().GetMaxSize(),
		MaxAge:     configConfigger.GetLogConfigger().GetMaxAge(),
		MaxBackups: configConfigger.GetLogConfigger().GetMaxBackups(),
		LocalTime:  configConfigger.GetLogConfigger().GetLocalTime(),
		Compress:   configConfigger.GetLogConfigger().GetCompress(),
	})
	zapcoreCore := zapcore.NewCore(jsonEncoder, zapcoreWriteSyncer, level)
	optioners := []zap.Option{
		zap.AddCaller(),
	}

	return zap.New(zapcoreCore, optioners...)
}

// NewMultiLogger is a function.
func NewMultiLogger(
	loggers ...*zap.Logger,
) *zap.Logger {
	zapcoreCores := make([]zapcore.Core, 0, len(loggers))
	for _, logger := range loggers {
		zapcoreCores = append(zapcoreCores, logger.Core())
	}

	teeCore := zapcore.NewTee(zapcoreCores...)
	optioners := []zap.Option{
		zap.AddCaller(),
	}

	return zap.New(teeCore, optioners...)
}

// NewJSONLogger is a function.
func NewJSONLogger(
	output *os.File,
	level zapcore.Level,
	format LoggerFormat,
) *zap.Logger {
	jsonEncoder := NewJSONEncoder(format)

	zapcoreCore := zapcore.NewCore(jsonEncoder, zapcore.Lock(output), level)
	optioners := []zap.Option{
		zap.AddCaller(),
	}

	return zap.New(zapcoreCore, optioners...)
}

// NewJSONEncoder is a function.
func NewJSONEncoder(
	format LoggerFormat,
) zapcore.Encoder {
	if format == StackdriverFormat {
		return zapcore.NewJSONEncoder(zapcore.EncoderConfig{
			MessageKey:          "message",
			LevelKey:            "severity",
			TimeKey:             "time",
			NameKey:             "logger",
			CallerKey:           "caller",
			FunctionKey:         zapcore.OmitKey,
			StacktraceKey:       "stacktrace",
			SkipLineEnding:      false,
			LineEnding:          zapcore.DefaultLineEnding,
			EncodeLevel:         StackdriverLevelEncoder,
			EncodeTime:          zapcore.RFC3339NanoTimeEncoder,
			EncodeDuration:      zapcore.StringDurationEncoder,
			EncodeCaller:        zapcore.ShortCallerEncoder,
			EncodeName:          nil,
			NewReflectedEncoder: nil,
			ConsoleSeparator:    object.URIEmpty,
		})
	}

	return zapcore.NewJSONEncoder(zapcore.EncoderConfig{
		MessageKey:          "msg",
		LevelKey:            "level",
		TimeKey:             "ts",
		NameKey:             "logger",
		CallerKey:           "caller",
		FunctionKey:         zapcore.OmitKey,
		StacktraceKey:       "stacktrace",
		SkipLineEnding:      false,
		LineEnding:          zapcore.DefaultLineEnding,
		EncodeLevel:         zapcore.LowercaseLevelEncoder,
		EncodeTime:          zapcore.ISO8601TimeEncoder,
		EncodeDuration:      zapcore.StringDurationEncoder,
		EncodeCaller:        zapcore.ShortCallerEncoder,
		EncodeName:          nil,
		NewReflectedEncoder: nil,
		ConsoleSeparator:    object.URIEmpty,
	})
}

// StackdriverLevelEncoder is a function.
func StackdriverLevelEncoder(
	lvl zapcore.Level,
	enc zapcore.PrimitiveArrayEncoder,
) {
	switch lvl {
	case zapcore.DebugLevel:
		enc.AppendString("DEBUG")

	case zapcore.InfoLevel:
		enc.AppendString("INFO")

	case zapcore.WarnLevel:
		enc.AppendString("WARNING")

	case zapcore.ErrorLevel:
		enc.AppendString("ERROR")

	case zapcore.DPanicLevel:
		enc.AppendString("CRITICAL")

	case zapcore.PanicLevel:
		enc.AppendString("CRITICAL")

	case zapcore.FatalLevel:
		enc.AppendString("CRITICAL")

	case zapcore.InvalidLevel:
	default:
		enc.AppendString("DEFAULT")
	}
}

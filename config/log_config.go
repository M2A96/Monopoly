package config

//go:generate mockgen -destination=../test/v2/log_config.go -package=test -mock_names=LogConfigger=MockLogConfig . LogConfigger

import (
	"encoding/json"
	"github/M2A96/Monopoly.git/object"
	"time"
)

type (
	// LogConfigger is configuration relevant to logging levels and output.
	// Reference: https://godoc.org/gopkg.in/natefinch/lumberjack.v2
	LogConfigger interface {
		// GetFile log output to a file (as well as stdout if set). Make sure that the directory and the
		// file is writable.
		GetFile() string
		// GetFormat Set logging output format. Can either be 'JSON' or 'Stackdriver'. Default is 'JSON'.
		GetFormat() string
		// GetLevel log level to set. Valid values are 'debug', 'info', 'warn', 'error'. Default 'info'.
		GetLevel() string
		// GetSQLSlowThreshold is the threshold of execution of a sql.
		GetSQLSlowThreshold() time.Duration
		// GetMaxAge is the maximum number of days to retain old log files based on the timestamp encoded
		// in their filename.
		// The default is not to remove old log files based on age.
		GetMaxAge() int
		// GetMaxBackups is the maximum number of old log files to retain.
		// The default is to retain all old log files (though MaxAge may still cause them to get deleted.)
		GetMaxBackups() int
		// GetMaxSize is the maximum size in megabytes of the log file before it gets rotated. It defaults
		// to 100 megabytes.
		GetMaxSize() int
		// GetCompress determines if the rotated log files should be compressed using gzip.
		GetCompress() bool
		// GetLocalTime determines if the time used for formatting the timestamps in
		// backup files is the computer's local time. The default is to use UTC time.
		GetLocalTime() bool
		// GetRotation rotate log files. Default is false.
		GetRotation() bool
		// GetStdout log to standard console output (as well as to a file if set). Default true.
		GetStdout() bool
	}

	// GetLogConfigger is an interface.
	GetLogConfigger interface {
		// GetLogConfigger is a function.
		GetLogConfigger() LogConfigger
	}

	logConfig struct {
		file             string
		format           string
		level            string
		sqlSlowThreshold time.Duration
		maxAge           int
		maxBackups       int
		maxSize          int
		compress         bool
		localTime        bool
		rotation         bool
		stdout           bool
	}

	logConfigOptioner interface {
		apply(*logConfig)
	}

	logConfigOptionerFunc func(*logConfig)
)

var (
	_ LogConfigger     = (*logConfig)(nil)
	_ json.Marshaler   = (*logConfig)(nil)
	_ object.GetMapper = (*logConfig)(nil)
)

// NewLogConfig is a function.
func NewLogConfig(
	optioners ...logConfigOptioner,
) *logConfig {
	logConfig := &logConfig{
		file:             object.URIEmpty,
		format:           object.URIEmpty,
		level:            object.URIEmpty,
		sqlSlowThreshold: 0 * time.Nanosecond,
		maxAge:           0,
		maxBackups:       0,
		maxSize:          0,
		compress:         false,
		localTime:        false,
		rotation:         false,
		stdout:           false,
	}

	return logConfig.WithOptioners(optioners...)
}

// WithLogConfigFile is a function.
func WithLogConfigFile(
	file string,
) logConfigOptioner {
	return logConfigOptionerFunc(func(
		config *logConfig,
	) {
		config.file = file
	})
}

// WithLogConfigFormat is a function.
func WithLogConfigFormat(
	format string,
) logConfigOptioner {
	return logConfigOptionerFunc(func(
		config *logConfig,
	) {
		config.format = format
	})
}

// WithLogConfigLevel is a function.
func WithLogConfigLevel(
	level string,
) logConfigOptioner {
	return logConfigOptionerFunc(func(
		config *logConfig,
	) {
		config.level = level
	})
}

// WithLogConfigSQLSlowThreshold is a function.
func WithLogConfigSQLSlowThreshold(
	sqlSlowThreshold time.Duration,
) logConfigOptioner {
	return logConfigOptionerFunc(func(
		config *logConfig,
	) {
		config.sqlSlowThreshold = sqlSlowThreshold
	})
}

// WithLogConfigMaxAge is a function.
func WithLogConfigMaxAge(
	maxAge int,
) logConfigOptioner {
	return logConfigOptionerFunc(func(
		config *logConfig,
	) {
		config.maxAge = maxAge
	})
}

// WithLogConfigMaxBackups is a function.
func WithLogConfigMaxBackups(
	maxBackups int,
) logConfigOptioner {
	return logConfigOptionerFunc(func(
		config *logConfig,
	) {
		config.maxBackups = maxBackups
	})
}

// WithLogConfigMaxSize is a function.
func WithLogConfigMaxSize(
	maxSize int,
) logConfigOptioner {
	return logConfigOptionerFunc(func(
		config *logConfig,
	) {
		config.maxSize = maxSize
	})
}

// WithLogConfigCompress is a function.
func WithLogConfigCompress(
	compress bool,
) logConfigOptioner {
	return logConfigOptionerFunc(func(
		config *logConfig,
	) {
		config.compress = compress
	})
}

// WithLogConfigLocalTime is a function.
func WithLogConfigLocalTime(
	localTime bool,
) logConfigOptioner {
	return logConfigOptionerFunc(func(
		config *logConfig,
	) {
		config.localTime = localTime
	})
}

// WithLogConfigRotation is a function.
func WithLogConfigRotation(
	rotation bool,
) logConfigOptioner {
	return logConfigOptionerFunc(func(
		config *logConfig,
	) {
		config.rotation = rotation
	})
}

// WithLogConfigStdout is a function.
func WithLogConfigStdout(
	stdout bool,
) logConfigOptioner {
	return logConfigOptionerFunc(func(
		config *logConfig,
	) {
		config.stdout = stdout
	})
}

// GetFile is a function.
func (config *logConfig) GetFile() string {
	return config.file
}

// GetFormat is a function.
func (config *logConfig) GetFormat() string {
	return config.format
}

// GetLevel is a function.
func (config *logConfig) GetLevel() string {
	return config.level
}

// GetSQLSlowThreshold is a function.
func (config *logConfig) GetSQLSlowThreshold() time.Duration {
	return config.sqlSlowThreshold
}

// GetMaxAge is a function.
func (config *logConfig) GetMaxAge() int {
	return config.maxAge
}

// GetMaxBackups is a function.
func (config *logConfig) GetMaxBackups() int {
	return config.maxBackups
}

// GetMaxSize is a function.
func (config *logConfig) GetMaxSize() int {
	return config.maxSize
}

// GetCompress is a function.
func (config *logConfig) GetCompress() bool {
	return config.compress
}

// GetLocalTime is a function.
func (config *logConfig) GetLocalTime() bool {
	return config.localTime
}

// GetRotation is a function.
func (config *logConfig) GetRotation() bool {
	return config.rotation
}

// GetStdout is a function.
func (config *logConfig) GetStdout() bool {
	return config.stdout
}

// GetMap is a function.
func (config *logConfig) GetMap() map[string]any {
	return map[string]any{
		"file":               config.GetFile(),
		"format":             config.GetFormat(),
		"level":              config.GetLevel(),
		"sql_slow_threshold": config.GetSQLSlowThreshold(),
		"max_age":            config.GetMaxAge(),
		"max_backups":        config.GetMaxBackups(),
		"max_size":           config.GetMaxSize(),
		"compress":           config.GetCompress(),
		"local_time":         config.GetLocalTime(),
		"rotation":           config.GetRotation(),
		"stdout":             config.GetStdout(),
	}
}

// MarshalJSON is a function.
// read more https://pkg.go.dev/encoding/json#Marshaler
func (config *logConfig) MarshalJSON() ([]byte, error) {
	return json.Marshal(config.GetMap())
}

// WithOptioners is a function.
func (config *logConfig) WithOptioners(
	optioners ...logConfigOptioner,
) *logConfig {
	newConfig := config.clone()
	for _, optioner := range optioners {
		optioner.apply(newConfig)
	}

	return newConfig
}

func (config *logConfig) clone() *logConfig {
	newConfig := config

	return newConfig
}

func (optionerFunc logConfigOptionerFunc) apply(
	config *logConfig,
) {
	optionerFunc(config)
}

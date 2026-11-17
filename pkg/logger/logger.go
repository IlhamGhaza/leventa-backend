package logger

import (
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
)

var (
	log           zerolog.Logger
	logFileHandle *os.File
)

// Init initializes the global logger based on the environment. When logFile is
// not empty, JSON log lines are also appended to that file (rotation is left to
// the operating system or a log shipper).
func Init(env string, logFile string) {
	local := env == "local" || env == "development"

	var out io.Writer = os.Stdout
	if local {
		out = zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
	}
	if logFile != "" {
		if f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o640); err == nil {
			Close()
			logFileHandle = f
			out = zerolog.MultiLevelWriter(out, f)
		} else {
			// Keep running on stdout; report the problem once.
			stderr := zerolog.New(os.Stderr)
			stderr.Error().Err(err).Str("log_file", logFile).Msg("cannot open log file, logging to stdout only")
		}
	}

	ctx := zerolog.New(out).With().Timestamp()
	if local {
		ctx = ctx.Caller()
	}
	log = ctx.Logger()

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if local {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}
}

// Close releases the log file, if one was opened by Init.
func Close() {
	if logFileHandle != nil {
		_ = logFileHandle.Close()
		logFileHandle = nil
	}
}

// Get returns the global logger instance.
func Get() *zerolog.Logger {
	return &log
}

// Info logs an info-level message.
func Info() *zerolog.Event {
	return log.Info()
}

// Debug logs a debug-level message.
func Debug() *zerolog.Event {
	return log.Debug()
}

// Warn logs a warning-level message.
func Warn() *zerolog.Event {
	return log.Warn()
}

// Error logs an error-level message.
func Error() *zerolog.Event {
	return log.Error()
}

// Fatal logs a fatal-level message and exits.
func Fatal() *zerolog.Event {
	return log.Fatal()
}

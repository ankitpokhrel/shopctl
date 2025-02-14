package tlog

import (
	"fmt"
	"io"
	"log/slog"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/exp/zapslog"
	"go.uber.org/zap/zapcore"
)

// VerboseLevel is a log verbosity.
type VerboseLevel int32

const (
	// VL0 (VerboseLevel zero) is a default logging verbosity.
	VL0 VerboseLevel = iota
	// VL1 (VerboseLevel one) is the minimum logging verbosity.
	VL1
	// VL2 (VerboseLevel two) is an intermediate level of logging verbosity.
	VL2
	// VL3 (VerboseLevel three) provides the highest level of logging verbosity.
	VL3
)

var noopLogger = slog.New(zapslog.NewHandler(zapcore.NewNopCore()))

// Logger is an app logger.
type Logger struct {
	writer    *slog.Logger
	errWriter *slog.Logger
	verbosity VerboseLevel
}

// newConsole builds a sensible production Logger that writes InfoLevel and
// above logs to standard error as text.
//
// Logging is enabled at InfoLevel and above, and uses a console encoder.
// Logs are written to standard error.
// Stacktraces are included on logs of ErrorLevel and above.
// DPanicLevel logs will not panic, but will write a stacktrace.
func newConsole(options ...zap.Option) (*zap.Logger, error) {
	encoder := zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalColorLevelEncoder,
		EncodeTime:     zapcore.TimeEncoderOfLayout(time.RFC3339),
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	config := zap.Config{
		Level:             zap.NewAtomicLevelAt(zap.InfoLevel),
		Development:       false,
		DisableStacktrace: true,
		Sampling:          nil,
		Encoding:          "console",
		EncoderConfig:     encoder,
		OutputPaths:       []string{"stderr"},
		ErrorOutputPaths:  []string{"stderr"},
	}

	return config.Build(options...)
}

// New constructs a new logger.
func New(v VerboseLevel, quiet bool) *Logger {
	zapLogger := zap.Must(newConsole())
	defer func() { _ = zapLogger.Sync() }()

	w := slog.New(zapslog.NewHandler(zapLogger.Sugar().Desugar().Core()))

	var logger *slog.Logger
	if quiet {
		logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	} else {
		logger = w
	}
	return &Logger{
		writer:    logger,
		errWriter: w,
		verbosity: v,
	}
}

// V checks the verbosity level and returns the logger instance if verbosity is sufficient.
func (l *Logger) V(level VerboseLevel) *Logger {
	if l.verbosity >= level {
		return l
	}
	return &Logger{writer: noopLogger}
}

// Info logs informational messages.
func (l *Logger) Info(msg string, args ...any) {
	l.writer.Info(msg, args...)
}

// Infof logs formatted informational messages.
func (l *Logger) Infof(format string, args ...any) {
	l.writer.Info(fmt.Sprintf(format, args...))
}

// Warn logs warning messages.
func (l *Logger) Warn(msg string, args ...any) {
	l.writer.Warn(msg, args...)
}

// Warnf logs formatted warning messages.
func (l *Logger) Warnf(msg string, args ...any) {
	l.writer.Warn(fmt.Sprintf(msg, args...))
}

// Error logs error messages.
func (l *Logger) Error(msg string, args ...any) {
	l.errWriter.Error(msg, args...)
}

// Errorf logs formatted error messages.
func (l *Logger) Errorf(msg string, args ...any) {
	l.errWriter.Error(fmt.Sprintf(msg, args...))
}

// Debug logs debug messages for verbosity level >= VL3.
func (l *Logger) Debug(msg string, args ...any) {
	l.V(VL3).writer.Debug(msg, args...)
}

// Debugf logs formated debug messages for verbosity level >= VL3.
func (l *Logger) Debugf(msg string, args ...any) {
	l.V(VL3).writer.Debug(fmt.Sprintf(msg, args...))
}

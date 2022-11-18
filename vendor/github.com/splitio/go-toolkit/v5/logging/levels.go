package logging

import (
	"fmt"
	"math"
)

// Standard values
const (
	// Discard 0 value, so when can use it as "the lack of a logging level"
	_ = iota

	// LevelError log level
	LevelError

	// LevelWarning log level
	LevelWarning

	// LevelInfo log level
	LevelInfo

	// LevelDebug log level
	LevelDebug

	// LevelVerbose log level
	LevelVerbose
)

// Special values
const (
	// LevelNone implies that NOTHING will be logged, not even errors
	LevelNone = math.MinInt32

	// LevelAll implies that All logging levels will be recorded
	LevelAll = math.MaxInt32
)

// LevelFilteredLoggerWrapper forwards log message to delegate if level is set higher than incoming message
type LevelFilteredLoggerWrapper struct {
	level    int
	delegate LoggerInterface
}

// Error forwards error logging messages
func (l *LevelFilteredLoggerWrapper) Error(is ...interface{}) {
	if l.level >= LevelError {
		l.delegate.Error(is...)
	}
}

// Warning forwards warning logging messages
func (l *LevelFilteredLoggerWrapper) Warning(is ...interface{}) {
	if l.level >= LevelWarning {
		l.delegate.Warning(is...)
	}
}

// Info forwards info logging messages
func (l *LevelFilteredLoggerWrapper) Info(is ...interface{}) {
	if l.level >= LevelInfo {
		l.delegate.Info(is...)
	}
}

// Debug forwards debug logging messages
func (l *LevelFilteredLoggerWrapper) Debug(is ...interface{}) {
	if l.level >= LevelDebug {
		l.delegate.Debug(is...)
	}
}

// Verbose forwards verbose logging messages
func (l *LevelFilteredLoggerWrapper) Verbose(is ...interface{}) {
	if l.level >= LevelVerbose {
		l.delegate.Verbose(is...)
	}
}

var levels map[string]int = map[string]int{
	"ERROR":   LevelError,
	"WARNING": LevelWarning,
	"INFO":    LevelInfo,
	"DEBUG":   LevelDebug,
	"VERBOSE": LevelVerbose,
}

// Level gets current level
func Level(level string) int {
	l, ok := levels[level]
	if !ok {
		panic("Invalid log level " + level)
	}
	return l
}

// ExtendedLevelFilteredLoggerWrapper leveled logger improving message format performance
type ExtendedLevelFilteredLoggerWrapper struct {
	*LevelFilteredLoggerWrapper
}

// sprintf apply parameter function to parse the given message
func (l *ExtendedLevelFilteredLoggerWrapper) sprintf(format string, params func() []interface{}) string {
	return fmt.Sprintf(format, params()...)
}

// ErrorFn forwards error logging messages getting parameters from a function to improve performance
func (l *ExtendedLevelFilteredLoggerWrapper) ErrorFn(format string, params ParamsFn) {
	if l.level >= LevelError {
		l.delegate.Error(l.sprintf(format, params))
	}
}

// WarningFn forwards warning logging messages getting parameters from a function to improve performance
func (l *ExtendedLevelFilteredLoggerWrapper) WarningFn(format string, params ParamsFn) {
	if l.level >= LevelWarning {
		l.delegate.Warning(l.sprintf(format, params))
	}
}

// InfoFn forwards info logging messages getting parameters from a function to improve performance
func (l *ExtendedLevelFilteredLoggerWrapper) InfoFn(format string, params ParamsFn) {
	if l.level >= LevelInfo {
		l.delegate.Info(l.sprintf(format, params))
	}
}

// DebugFn forwards debug logging messages getting parameters from a function to improve performance
func (l *ExtendedLevelFilteredLoggerWrapper) DebugFn(format string, params ParamsFn) {
	if l.level >= LevelDebug {
		l.delegate.Debug(l.sprintf(format, params))
	}
}

// VerboseFn forwards verbose logging messages getting parameters from a function to improve performance
func (l *ExtendedLevelFilteredLoggerWrapper) VerboseFn(format string, params ParamsFn) {
	if l.level >= LevelVerbose {
		l.delegate.Verbose(l.sprintf(format, params))
	}
}

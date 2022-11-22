package logging

// LoggerInterface ...
// If a custom logger object is to be used, it should comply with the following
// interface. (Standard go-lang library log.Logger.Println method signature)
type LoggerInterface interface {
	Error(msg ...interface{})
	Warning(msg ...interface{})
	Info(msg ...interface{})
	Debug(msg ...interface{})
	Verbose(msg ...interface{})
}

// ParamsFn is a function that returns a slice of interface{}
type ParamsFn = func() []interface{}

// ExtendedLoggerInterface ...
// If a custom logger object is to be used, it should comply with the following
// interface. (Standard go-lang library log.Logger.Println method signature)
type ExtendedLoggerInterface interface {
	LoggerInterface
	ErrorFn(format string, params ParamsFn)
	WarningFn(format string, params ParamsFn)
	InfoFn(format string, params ParamsFn)
	DebugFn(format string, params ParamsFn)
	VerboseFn(format string, params ParamsFn)
}

package log

var (
	root = &logger{[]interface{}{}, new(swapHandler)}
)

////keystore -- watch.go -- loop()
//// New returns a new logger with the given context.
//// New is a convenient alias for Root().New
//func New(ctx ...interface{}) Logger {
//	return root.New(ctx...)
//}

// Debug is a convenient alias for Root().Debug
func Debug(msg string, ctx ...interface{}) {
	root.write(msg, LvlDebug, ctx)
}

// The following functions bypass the exported logger methods (logger.Debug,
// etc.) to keep the call depth the same for all paths to logger.write so
// runtime.Caller(2) always refers to the call site in client code.

// Trace is a convenient alias for Root().Trace
func Trace(msg string, ctx ...interface{}) {
	root.write(msg, LvlTrace, ctx)
}

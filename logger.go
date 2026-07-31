package gocfg

// Logger is the minimal logging contract used by gocfg.
//
// It is satisfied out of the box by *log.Logger from the standard library and
// by most third-party loggers (logrus, zerolog's compatibility wrappers, ...).
// Loggers exposing differently named methods can be adapted with LoggerFunc.
//
// gocfg never writes anywhere unless a Logger is explicitly configured through
// ConfigManager.WithLogger: by default the library is completely silent.
type Logger interface {
	Printf(format string, args ...interface{})
}

// LoggerFunc adapts a plain function to the Logger interface, which makes it
// possible to pass method values directly:
//
//	cfg.WithLogger(gocfg.LoggerFunc(sugar.Infof)) // zap
//	cfg.WithLogger(gocfg.LoggerFunc(t.Logf))      // testing
type LoggerFunc func(format string, args ...interface{})

// Printf implements Logger.
func (f LoggerFunc) Printf(format string, args ...interface{}) {
	f(format, args...)
}

// logf writes a diagnostic message if, and only if, a logger was configured.
func (c *ConfigManager) logf(format string, args ...interface{}) {
	if c.logger == nil {
		return
	}

	c.logger.Printf(format, args...)
}

package logger

import (
	"log/slog"
	"os"
)

// Config is the configuration for the log package. Within the configuration it
// is possible to set the log level and log format for our logger.
type Config struct {
	Format string     `json:"format" env:"FORMAT" enum:"console,json" default:"console" help:"Set the output format of the logs. Must be \"console\" or \"json\"."`
	Level  slog.Level `json:"level" env:"LEVEL" enum:"DEBUG,INFO,WARN,ERROR" default:"INFO" help:"Set the log level. Must be \"DEBUG\", \"INFO\", \"WARN\" or \"ERROR\"."`
}

func New(config Config) *slog.Logger {
	var handler slog.Handler

	if config.Format == "json" {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			AddSource: true,
			Level:     config.Level,
		})
	} else {
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			AddSource: true,
			Level:     config.Level,
		})
	}

	logger := slog.New(handler)
	slog.SetDefault(logger)

	return logger
}

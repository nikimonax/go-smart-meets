package logging

import (
	"log/slog"

	"go.uber.org/fx/fxevent"
)

func NewFxLogger(logger *slog.Logger) fxevent.Logger {
	fxLogger := &fxevent.SlogLogger{
		Logger: logger.With("package", "fx"),
	}
	fxLogger.UseLogLevel(slog.LevelDebug)
	return fxLogger
}

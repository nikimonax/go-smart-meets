package server

import (
	"log/slog"
	"os"
	"time"

	"go.uber.org/fx"

	"github.com/lmittmann/tint"

	"github.com/nikimonax/go-smart-meets/internal/server/config"
	"github.com/nikimonax/go-smart-meets/internal/server/module"
	"github.com/nikimonax/go-smart-meets/internal/shared/logging"
)

func New(cfg *config.Config) *fx.App {
	logger := slog.New(
		tint.NewTextHandler(os.Stdout, &tint.Options{
			Level:      slog.LevelDebug,
			TimeFormat: time.Kitchen,
		}),
	)
	slog.SetDefault(logger)

	return fx.New(
		fx.Supply(cfg),
		fx.Supply(logger),
		module.CoreModule(),
		module.DatabaseModule(),
		module.QueueModule(),
		module.QueueUIModule(),
		module.HTTPModule(),
		fx.WithLogger(logging.NewFxLogger),
	)
}

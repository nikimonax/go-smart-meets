package server

import (
	"go.uber.org/fx"

	"github.com/nikimonax/go-smart-meets/internal/server/background"
	"github.com/nikimonax/go-smart-meets/internal/server/config"
	"github.com/nikimonax/go-smart-meets/internal/server/module"
)

func New(cfg *config.Config) fx.App {
	// TODO: setup logging
	return *fx.New(
		fx.Supply(cfg),
		module.DatabaseModule(),
		module.QueueModule(),
		fx.Invoke(background.RegisterWorkers),
	)
}

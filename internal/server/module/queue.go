package module

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"go.uber.org/fx"

	"github.com/nikimonax/go-smart-meets/internal/server/background"
)

// depends on database module
func QueueModule() fx.Option {
	return fx.Module(
		"queue",
		fx.Provide(
			provideRiverWorkers,
			provideRiverClient,
		),
		fx.Invoke(
			background.RegisterWorkers,
			registerRiverClientLifecycle,
		),
	)
}

func provideRiverWorkers() *river.Workers {
	return river.NewWorkers()
}

func provideRiverClient(
	pool *pgxpool.Pool,
	workers *river.Workers,
	logger *slog.Logger,
) (*river.Client[pgx.Tx], error) {
	return river.NewClient(riverpgxv5.New(pool), &river.Config{
		Queues: map[string]river.QueueConfig{
			river.QueueDefault: {MaxWorkers: 20},
		},
		Logger: logger.With(
			"module", "queue",
			"package", "river",
		),
		Workers: workers,
	})
}

func registerRiverClientLifecycle(
	lc fx.Lifecycle,
	runCtx runtimeContext,
	client *river.Client[pgx.Tx],
) {
	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			return client.Start(runCtx)
		},
		OnStop: client.Stop,
	})
}

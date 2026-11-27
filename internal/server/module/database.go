package module

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"

	"github.com/nikimonax/go-smart-meets/internal/impl/database"
	"github.com/nikimonax/go-smart-meets/internal/server/config"
)

// depends on *config.Config
func DatabaseModule() fx.Option {
	return fx.Module(
		"database",
		fx.Provide(
			providePgxPool,
		),
		fx.Invoke(
			registerPgxPoolPing,
			registerPgxPoolClose,
			registerDatabaseMigrate,
		),
	)
}

func providePgxPool(
	cfg *config.Config,
) (*pgxpool.Pool, error) {
	return database.NewPool(cfg.Database.URL().String())
}

func registerPgxPoolPing(
	lc fx.Lifecycle,
	pool *pgxpool.Pool,
) {
	lc.Append(fx.StartHook(pool.Ping))
}

func registerPgxPoolClose(
	lc fx.Lifecycle,
	pool *pgxpool.Pool,
) {
	lc.Append(fx.StopHook(pool.Close))
}

func registerDatabaseMigrate(
	lc fx.Lifecycle,
	cfg *config.Config,
	pool *pgxpool.Pool,
) {
	if !cfg.Database.Migrate {
		return
	}

	lc.Append(fx.StartHook(func(ctx context.Context) error {
		return database.Migrate(ctx, pool)
	}))
}

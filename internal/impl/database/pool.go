package database

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPool(dsn string) (*pgxpool.Pool, error) {
	poolCfg, err := pgxpool.ParseConfig(dsn)

	if err != nil {
		return nil, err
	}

	// TODO: better pool options?

	return pgxpool.NewWithConfig(context.Background(), poolCfg)
}

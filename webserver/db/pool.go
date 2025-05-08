package db

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Pool struct {
	*pgxpool.Pool
}

func NewPool(ctx context.Context, connStr string) (pool *Pool, err error) {
	pgxPool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		return
	}

	if err = pgxPool.Ping(ctx); err != nil {
		return
	}

	pool = &Pool{
		pgxPool,
	}

	return
}

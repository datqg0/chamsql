package db

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

const (
	DatabaseTimeout = time.Second * 5
)

type Database struct {
	pool *pgxpool.Pool
}

func NewDatabase(uri string) (*Database, error) {
	config, err := pgxpool.ParseConfig(uri)
	if err != nil {
		return nil, err
	}

	config.MaxConns = 25
	config.MinConns = 5
	config.MaxConnLifetime = time.Hour
	config.MaxConnIdleTime = 30 * time.Minute

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(context.Background()); err != nil {
		pool.Close()
		return nil, err
	}

	log.Info().Msg("Database connection established")

	return &Database{pool: pool}, nil
}

func (d *Database) GetPool() *pgxpool.Pool {
	return d.pool
}

func (d *Database) Close() {
	if d.pool != nil {
		d.pool.Close()
		log.Info().Msg("Database connection closed")
	}
}

func (d *Database) WithContext(ctx context.Context) context.Context {
	ctx, _ = context.WithTimeout(ctx, DatabaseTimeout)
	return ctx
}

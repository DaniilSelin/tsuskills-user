package postgres

import (
	"context"
	"fmt"
	"time"

	"tsuskills-user/config"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Connect(ctx context.Context, cfg *config.PostgresConfig) (*pgxpool.Pool, error) {
	connString := cfg.Pool.ConnConfig.ConnString()
	connString = fmt.Sprintf("%s&search_path=%s,public", connString, cfg.Schema)

	poolCfg, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("unable to parse pool config: %w", err)
	}

	poolCfg.MaxConns = cfg.Pool.MaxConns
	poolCfg.MinConns = cfg.Pool.MinConns
	poolCfg.MaxConnLifetime = cfg.Pool.MaxConnLifetime
	poolCfg.MaxConnLifetimeJitter = cfg.Pool.MaxConnLifetimeJitter
	poolCfg.MaxConnIdleTime = cfg.Pool.MaxConnIdleTime
	poolCfg.HealthCheckPeriod = cfg.Pool.HealthCheckPeriod

	for attempt := 0; attempt <= cfg.ConnectRetries; attempt++ {
		pool, connErr := pgxpool.NewWithConfig(ctx, poolCfg)
		if connErr == nil {
			if pingErr := pool.Ping(ctx); pingErr == nil {
				return pool, nil
			}
			pool.Close()
		}
		if attempt == cfg.ConnectRetries {
			break
		}
		select {
		case <-time.After(cfg.ConnectRetryDelay):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	return nil, fmt.Errorf("failed to connect to postgres after %d attempts", cfg.ConnectRetries+1)
}

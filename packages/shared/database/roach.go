package database

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/yaninyzwitty/go-fx-v1/packages/shared/config"
	"github.com/yaninyzwitty/go-fx-v1/packages/shared/repository"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Params struct {
	fx.In

	Lifecycle fx.Lifecycle
	Cfg       *config.Config
	Log       *zap.Logger
}

// Module exports the database providers
// It provides both the connection pool and the queries repository
var Module = fx.Module("database",
	fx.Provide(NewPool),
	fx.Provide(NewQueries),
)

func NewPool(p Params) (*pgxpool.Pool, error) {
	config := p.Cfg

	if err := godotenv.Load(); err != nil {
		p.Log.Warn("No .env file found, proceeding with environment variables")
	}

	// Prefer environment variable, fall back to config file
	password := os.Getenv("DB_PASSWORD")
	if password == "" {
		password = config.DbConfig.Password
		if password == "" {
			return nil, fmt.Errorf("DB_PASSWORD env variable not set and password not found in config")
		}
	}

	// Properly encode username and password for URL userinfo
	// Use url.UserPassword which handles encoding correctly
	user := url.UserPassword(config.DbConfig.Username, password)

	// Build the connection URL properly
	connURL := &url.URL{
		Scheme:   "postgres",
		User:     user,
		Host:     fmt.Sprintf("%s:%d", config.DbConfig.Host, config.DbConfig.Port),
		Path:     "/" + config.DbConfig.Database,
		RawQuery: fmt.Sprintf("sslmode=%s", url.QueryEscape(config.DbConfig.SslMode)),
	}

	connString := connURL.String()

	pool, err := pgxpool.New(context.Background(), connString)
	if err != nil {
		p.Log.Error("Failed to connect to database", zap.Error(err))
		return nil, fmt.Errorf("failed to create pool: %w", err)
	}

	// Test the connection
	testCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := pool.Ping(testCtx); err != nil {
		pool.Close()
		p.Log.Error("Failed to ping database", zap.Error(err))
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Register lifecycle hooks to close pool on app shutdown
	p.Lifecycle.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			pool.Close()
			return nil
		},
	})
	return pool, nil

}

func NewQueries(pool *pgxpool.Pool) *repository.Queries {
	return repository.New(pool)
}

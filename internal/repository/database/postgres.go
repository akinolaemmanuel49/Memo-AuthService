package database

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/akinolaemmanuel49/Memo-AuthService/config"
	"github.com/akinolaemmanuel49/Memo-AuthService/internal/repository/dtos"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Database wraps the pgx pool and provides logging
type Database struct {
	pool   *pgxpool.Pool
	logger *slog.Logger
}

// NewPoolConfig creates a pgxpool.Config with the provided settings
func NewPoolConfig(dbConfig *config.DatabaseConfig) (*pgxpool.Config, error) {
	poolConfig, err := pgxpool.ParseConfig(dbConfig.URI)
	if err != nil {
		return nil, fmt.Errorf("error parsing database URI: %w", err)
	}

	poolConfig.MaxConns = dbConfig.MaxConns
	poolConfig.MinConns = dbConfig.MinConns
	poolConfig.MaxConnLifetime = dbConfig.MaxConnLifetime
	poolConfig.MaxConnIdleTime = dbConfig.MaxConnIdleTime
	poolConfig.HealthCheckPeriod = dbConfig.HealthCheckPeriod

	return poolConfig, nil
}

// New creates a new Database instance
func New(ctx context.Context, cfg *config.Config) (*Database, error) {
	poolConfig, err := NewPoolConfig(&cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("failed to create pool config: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	db := &Database{
		pool:   pool,
		logger: cfg.Logger.With(slog.String("component", "database")),
	}

	db.logger.Info("database connection established",
		"max_conns", cfg.Database.MaxConns,
		"min_conns", cfg.Database.MinConns,
	)

	return db, nil
}

// Close gracefully shuts down the database pool
func (db *Database) Close() {
	db.logger.Info("closing database connection pool")
	db.pool.Close()
}

// Pool returns the underlying connection pool
func (db *Database) Pool() *pgxpool.Pool {
	return db.pool
}

// HealthCheck performs a health check on the database
func (db *Database) HealthCheck(ctx context.Context) (interface{}, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := db.pool.Ping(timeoutCtx); err != nil {
		db.logger.Error("database health check failed", "error", err)
		return nil, fmt.Errorf("database health check failed: %w", err)
	}

	stats := db.pool.Stat()
	db.logger.Debug("database health check completed",
		"totalConnections", stats.TotalConns(),
		"acquiredConnections", stats.AcquiredConns(),
		"idleConnections", stats.IdleConns(),
	)

	metrics := dtos.DatabaseMetrics{
		TotalConnections:    stats.TotalConns(),
		AcquiredConnections: stats.AcquiredConns(),
		IdleConnections:     stats.IdleConns(),
	}

	return metrics, nil
}

// WithTransaction executes a function within a transaction
func (db *Database) WithTransaction(ctx context.Context, fn func(pgx.Tx) error) error {
	conn, err := db.pool.Acquire(ctx)
	if err != nil {
		db.logger.Error("failed to acquire connection", "error", err)
		return fmt.Errorf("failed to acquire connection: %w", err)
	}
	defer conn.Release()

	tx, err := conn.Begin(ctx)
	if err != nil {
		db.logger.Error("failed to begin transaction", "error", err)
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			db.logger.Error("transaction panic, rolling back", "panic", p)
			_ = tx.Rollback(ctx)
			panic(p)
		}
	}()

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			db.logger.Error("failed to rollback transaction",
				"rollback_error", rbErr,
				"original_error", err,
			)
		}
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		db.logger.Error("failed to commit transaction", "error", err)
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// Acquire gets a connection from the pool
func (db *Database) Acquire(ctx context.Context) (*pgxpool.Conn, error) {
	conn, err := db.pool.Acquire(ctx)
	if err != nil {
		db.logger.Error("failed to acquire connection", "error", err)
		return nil, fmt.Errorf("failed to acquire connection: %w", err)
	}
	return conn, nil
}

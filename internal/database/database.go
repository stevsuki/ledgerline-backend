// Package database: PostgreSQL connection and pool setup for GORM.
package database

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/stevensuki/ledgerline-backend/internal/config"
)

// slowQueryThreshold: anything slower is logged as a warning.
const slowQueryThreshold = 200 * time.Millisecond

// New opens the connection, configures the pool, then pings.
func New(ctx context.Context, cfg config.Database, log *slog.Logger, debug bool) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{
		Logger: newSlogLogger(log, slowQueryThreshold, debug),
		// TranslateError maps driver errors to GORM errors, so no repository reads SQLSTATE.
		TranslateError: true,
		NowFunc:        func() time.Time { return time.Now().UTC() },
	})
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get sql.DB: %w", err)
	}

	sqlDB.SetMaxOpenConns(int(cfg.MaxConns))
	sqlDB.SetMaxIdleConns(int(cfg.MinConns))
	sqlDB.SetConnMaxLifetime(cfg.MaxConnLifetime)
	sqlDB.SetConnMaxIdleTime(cfg.MaxConnIdleTime)

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(pingCtx); err != nil {
		_ = sqlDB.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}
	return db, nil
}

// Close shuts down the connection pool behind *gorm.DB.
func Close(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

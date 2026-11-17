package database

import (
	"context"
	"database/sql"
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"omni-relay/configs"
)

type Client struct {
	DB    *gorm.DB
	sqlDB *sql.DB
}

func Open(ctx context.Context, cfg configs.PostgresConfig) (*Client, error) {
	db, err := gorm.Open(postgres.Open(cfg.DSN), &gorm.Config{
		TranslateError: true,
		Logger:         logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get postgres connection pool: %w", err)
	}
	sqlDB.SetMaxOpenConns(int(cfg.MaxConns))
	sqlDB.SetMaxIdleConns(int(cfg.MaxConns))

	if err := sqlDB.PingContext(ctx); err != nil {
		_ = sqlDB.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}
	return &Client{DB: db, sqlDB: sqlDB}, nil
}

func (c *Client) Ping(ctx context.Context) error {
	return c.sqlDB.PingContext(ctx)
}

func (c *Client) Close() error {
	return c.sqlDB.Close()
}

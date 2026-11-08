package redis

import (
	"context"
	"crypto/tls"
	"fmt"

	redisclient "github.com/redis/go-redis/v9"

	"omni-relay/configs"
)

func Open(ctx context.Context, cfg configs.RedisConfig) (*redisclient.Client, error) {
	options := &redisclient.Options{
		Addr:     cfg.Address,
		Password: cfg.Password,
		DB:       cfg.DB,
	}
	if cfg.TLSEnabled {
		options.TLSConfig = &tls.Config{MinVersion: tls.VersionTLS12}
	}

	client := redisclient.NewClient(options)
	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("ping redis: %w", err)
	}
	return client, nil
}

package bootstrap

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	redisclient "github.com/redis/go-redis/v9"
	"github.com/twmb/franz-go/pkg/kgo"

	"omni-relay/configs"
	"omni-relay/internal/shared/infrastructure/database"
	"omni-relay/internal/shared/infrastructure/httpserver"
	kafkastore "omni-relay/internal/shared/infrastructure/kafka"
	redisstore "omni-relay/internal/shared/infrastructure/redis"
)

type App struct {
	config   configs.Config
	logger   *slog.Logger
	postgres *pgxpool.Pool
	redis    *redisclient.Client
	kafka    *kgo.Client
	server   *httpserver.Server
}

func New(ctx context.Context, cfg configs.Config, logger *slog.Logger) (*App, error) {
	postgres, err := database.Open(ctx, cfg.Postgres)
	if err != nil {
		return nil, err
	}

	redis, err := redisstore.Open(ctx, cfg.Redis)
	if err != nil {
		postgres.Close()
		return nil, err
	}

	kafka, err := kafkastore.Open(ctx, cfg.Kafka)
	if err != nil {
		_ = redis.Close()
		postgres.Close()
		return nil, err
	}

	dependencies := []httpserver.Dependency{
		{Name: "postgres", Check: postgres.Ping},
		{Name: "redis", Check: func(ctx context.Context) error {
			return redis.Ping(ctx).Err()
		}},
		{Name: "kafka", Check: kafka.Ping},
	}

	return &App{
		config:   cfg,
		logger:   logger,
		postgres: postgres,
		redis:    redis,
		kafka:    kafka,
		server:   httpserver.New(cfg.HTTP, logger, dependencies),
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	serverErr := make(chan error, 1)
	go func() {
		serverErr <- a.server.Run()
	}()

	a.logger.Info("server started",
		"environment", a.config.App.Environment,
		"address", a.config.HTTP.Address,
	)

	select {
	case err := <-serverErr:
		if err != nil {
			return fmt.Errorf("run http server: %w", err)
		}
		return nil
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), a.config.HTTP.ShutdownTimeout)
		defer cancel()
		if err := a.server.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("shutdown http server: %w", err)
		}
		select {
		case err := <-serverErr:
			return err
		case <-time.After(a.config.HTTP.ShutdownTimeout):
			return errorsFromTimeout(a.config.HTTP.ShutdownTimeout)
		}
	}
}

func (a *App) Close() {
	a.kafka.Close()
	if err := a.redis.Close(); err != nil {
		a.logger.Error("close redis", "error", err)
	}
	a.postgres.Close()
}

func errorsFromTimeout(timeout time.Duration) error {
	return fmt.Errorf("wait for http server shutdown after %s", timeout)
}

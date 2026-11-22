package bootstrap

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	redisclient "github.com/redis/go-redis/v9"
	"github.com/twmb/franz-go/pkg/kgo"

	"omni-relay/configs"
	accesskeyapp "omni-relay/internal/accesskey/application"
	accesskeypostgres "omni-relay/internal/accesskey/infrastructure/postgres"
	accesskeyredis "omni-relay/internal/accesskey/infrastructure/redis"
	accesskeysecurity "omni-relay/internal/accesskey/infrastructure/security"
	accesskeyhttp "omni-relay/internal/accesskey/interfaces/http"
	catalogapp "omni-relay/internal/catalog/application"
	catalogpostgres "omni-relay/internal/catalog/infrastructure/postgres"
	cataloghttp "omni-relay/internal/catalog/interfaces/http"
	identityapp "omni-relay/internal/identity/application"
	identitypostgres "omni-relay/internal/identity/infrastructure/postgres"
	identityredis "omni-relay/internal/identity/infrastructure/redis"
	"omni-relay/internal/identity/infrastructure/security"
	identityhttp "omni-relay/internal/identity/interfaces/http"
	relaypostgres "omni-relay/internal/relay/infrastructure/postgres"
	relayhttp "omni-relay/internal/relay/interfaces/http"
	"omni-relay/internal/shared/infrastructure/database"
	"omni-relay/internal/shared/infrastructure/httpserver"
	kafkastore "omni-relay/internal/shared/infrastructure/kafka"
	redisstore "omni-relay/internal/shared/infrastructure/redis"
	"omni-relay/internal/shared/infrastructure/secret"
)

type App struct {
	config   configs.Config
	logger   *slog.Logger
	postgres *database.Client
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
		_ = postgres.Close()
		return nil, err
	}

	kafka, err := kafkastore.Open(ctx, cfg.Kafka)
	if err != nil {
		_ = redis.Close()
		_ = postgres.Close()
		return nil, err
	}

	dependencies := []httpserver.Dependency{
		{Name: "postgres", Check: postgres.Ping},
		{Name: "redis", Check: func(ctx context.Context) error {
			return redis.Ping(ctx).Err()
		}},
		{Name: "kafka", Check: kafka.Ping},
	}
	userRepository := identitypostgres.NewUserRepository(postgres.DB)
	sessionStore := identityredis.NewSessionStore(redis)
	passwordService := security.NewPasswordService()
	identityService := identityapp.NewService(
		userRepository,
		passwordService,
		sessionStore,
		cfg.Identity.NewUserInitialQuota,
		cfg.Session.TTL,
	)
	identityHandler := identityhttp.NewHandler(
		identityService,
		logger,
		cfg.Session.TTL,
		cfg.Session.CookieSecure,
	)
	accessKeyRepository := accesskeypostgres.NewRepository(postgres.DB)
	credentialService := accesskeysecurity.NewCredentialService()
	balanceStore := accesskeyredis.NewBalanceStore(redis)
	accessKeyService := accesskeyapp.NewService(
		accessKeyRepository,
		userRepository,
		credentialService,
		balanceStore,
	)
	accessKeyHandler := accesskeyhttp.NewHandler(accessKeyService, identityService, logger)
	secretCipher, err := secret.NewCipher(cfg.Secret.EncryptionKey)
	if err != nil {
		kafka.Close()
		_ = redis.Close()
		_ = postgres.Close()
		return nil, err
	}
	catalogRepository := catalogpostgres.NewRepository(postgres.DB)
	catalogService := catalogapp.NewService(catalogRepository, secretCipher, cfg.Upstream.AllowHTTP)
	catalogHandler := cataloghttp.NewHandler(catalogService, identityService, logger)
	relayHandler := relayhttp.NewHandler(relaypostgres.NewRepository(postgres.DB), secretCipher, cfg.Upstream.Timeout, cfg.Upstream.MaxBodyBytes, logger)

	return &App{
		config:   cfg,
		logger:   logger,
		postgres: postgres,
		redis:    redis,
		kafka:    kafka,
		server: httpserver.New(
			cfg.HTTP,
			logger,
			dependencies,
			identityHandler.RegisterRoutes,
			accessKeyHandler.RegisterRoutes,
			catalogHandler.RegisterRoutes,
			func(mux *http.ServeMux) { relayHandler.RegisterRoutes(mux, accessKeyHandler) },
		),
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
	if err := a.postgres.Close(); err != nil {
		a.logger.Error("close postgres", "error", err)
	}
}

func errorsFromTimeout(timeout time.Duration) error {
	return fmt.Errorf("wait for http server shutdown after %s", timeout)
}

package configs

import (
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	App      AppConfig
	HTTP     HTTPConfig
	Postgres PostgresConfig
	Redis    RedisConfig
	Kafka    KafkaConfig
	Session  SessionConfig
	Secret   SecretConfig
	Upstream UpstreamConfig
}

type AppConfig struct {
	Environment string
}

type HTTPConfig struct {
	Address         string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
}

type PostgresConfig struct {
	DSN      string
	MaxConns int32
}

type RedisConfig struct {
	Address    string
	Password   string
	DB         int
	TLSEnabled bool
}

type KafkaConfig struct {
	Brokers       []string
	ClientID      string
	BillingTopic  string
	ConsumerGroup string
	BatchSize     int
	TLSEnabled    bool
	SASLUsername  string
	SASLPassword  string
}

type SessionConfig struct {
	TTL          time.Duration
	CookieSecure bool
}

type SecretConfig struct {
	EncryptionKey []byte
}

type UpstreamConfig struct {
	Timeout      time.Duration
	MaxBodyBytes int64
}

func Load() (Config, error) {
	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		return Config{}, fmt.Errorf("load .env: %w", err)
	}

	readTimeout, err := durationMS("HTTP_READ_TIMEOUT_MS", 10_000)
	if err != nil {
		return Config{}, err
	}
	writeTimeout, err := durationMS("HTTP_WRITE_TIMEOUT_MS", 30_000)
	if err != nil {
		return Config{}, err
	}
	shutdownTimeout, err := durationMS("HTTP_SHUTDOWN_TIMEOUT_MS", 10_000)
	if err != nil {
		return Config{}, err
	}
	maxConns, err := intValue("POSTGRES_MAX_CONNS", 20, 1, 10_000)
	if err != nil {
		return Config{}, err
	}
	redisDB, err := intValue("REDIS_DB", 0, 0, 15)
	if err != nil {
		return Config{}, err
	}
	redisTLS, err := boolValue("REDIS_TLS_ENABLED", false)
	if err != nil {
		return Config{}, err
	}
	kafkaBatchSize, err := intValue("KAFKA_BATCH_SIZE", 500, 1, 10_000)
	if err != nil {
		return Config{}, err
	}
	kafkaTLS, err := boolValue("KAFKA_TLS_ENABLED", false)
	if err != nil {
		return Config{}, err
	}
	sessionTTL, err := durationMS("SESSION_TTL_MS", 86_400_000)
	if err != nil {
		return Config{}, err
	}
	cookieSecure, err := boolValue("SESSION_COOKIE_SECURE", false)
	if err != nil {
		return Config{}, err
	}
	upstreamTimeout, err := durationMS("UPSTREAM_TIMEOUT_MS", 120_000)
	if err != nil {
		return Config{}, err
	}
	maxBodyBytes, err := int64Value("UPSTREAM_MAX_BODY_BYTES", 10_485_760, 1, 1<<30)
	if err != nil {
		return Config{}, err
	}
	encryptionKey, err := encryptionKey()
	if err != nil {
		return Config{}, err
	}

	cfg := Config{
		App: AppConfig{Environment: value("APP_ENV", "local")},
		HTTP: HTTPConfig{
			Address:         value("HTTP_ADDR", ":8080"),
			ReadTimeout:     readTimeout,
			WriteTimeout:    writeTimeout,
			ShutdownTimeout: shutdownTimeout,
		},
		Postgres: PostgresConfig{
			DSN:      strings.TrimSpace(os.Getenv("POSTGRES_DSN")),
			MaxConns: int32(maxConns),
		},
		Redis: RedisConfig{
			Address:    strings.TrimSpace(os.Getenv("REDIS_ADDR")),
			Password:   os.Getenv("REDIS_PASSWORD"),
			DB:         redisDB,
			TLSEnabled: redisTLS,
		},
		Kafka: KafkaConfig{
			Brokers:       splitCSV(os.Getenv("KAFKA_BROKERS")),
			ClientID:      value("KAFKA_CLIENT_ID", "omni-relay"),
			BillingTopic:  value("KAFKA_BILLING_TOPIC", "omni-relay.billing.settled.v1"),
			ConsumerGroup: value("KAFKA_CONSUMER_GROUP", "omni-relay-consumption-v1"),
			BatchSize:     kafkaBatchSize,
			TLSEnabled:    kafkaTLS,
			SASLUsername:  os.Getenv("KAFKA_SASL_USERNAME"),
			SASLPassword:  os.Getenv("KAFKA_SASL_PASSWORD"),
		},
		Session: SessionConfig{TTL: sessionTTL, CookieSecure: cookieSecure},
		Secret:  SecretConfig{EncryptionKey: encryptionKey},
		Upstream: UpstreamConfig{
			Timeout:      upstreamTimeout,
			MaxBodyBytes: maxBodyBytes,
		},
	}

	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func (c Config) Validate() error {
	var missing []string
	if c.Postgres.DSN == "" {
		missing = append(missing, "POSTGRES_DSN")
	}
	if c.Redis.Address == "" {
		missing = append(missing, "REDIS_ADDR")
	}
	if len(c.Kafka.Brokers) == 0 {
		missing = append(missing, "KAFKA_BROKERS")
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required environment variables: %s", strings.Join(missing, ", "))
	}
	if (c.Kafka.SASLUsername == "") != (c.Kafka.SASLPassword == "") {
		return errors.New("KAFKA_SASL_USERNAME and KAFKA_SASL_PASSWORD must be set together")
	}
	if len(c.Secret.EncryptionKey) != 32 {
		return errors.New("SECRET_ENCRYPTION_KEY must decode to exactly 32 bytes")
	}
	return nil
}

func value(name, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(name)); value != "" {
		return value
	}
	return fallback
}

func splitCSV(raw string) []string {
	var values []string
	for _, item := range strings.Split(raw, ",") {
		if item = strings.TrimSpace(item); item != "" {
			values = append(values, item)
		}
	}
	return values
}

func durationMS(name string, fallback int64) (time.Duration, error) {
	value, err := int64Value(name, fallback, 1, int64((24*time.Hour)/time.Millisecond)*30)
	if err != nil {
		return 0, err
	}
	return time.Duration(value) * time.Millisecond, nil
}

func intValue(name string, fallback, min, max int) (int, error) {
	raw := strings.TrimSpace(os.Getenv(name))
	if raw == "" {
		return fallback, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < min || value > max {
		return 0, fmt.Errorf("%s must be an integer between %d and %d", name, min, max)
	}
	return value, nil
}

func int64Value(name string, fallback, min, max int64) (int64, error) {
	raw := strings.TrimSpace(os.Getenv(name))
	if raw == "" {
		return fallback, nil
	}
	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || value < min || value > max {
		return 0, fmt.Errorf("%s must be an integer between %d and %d", name, min, max)
	}
	return value, nil
}

func boolValue(name string, fallback bool) (bool, error) {
	raw := strings.TrimSpace(os.Getenv(name))
	if raw == "" {
		return fallback, nil
	}
	value, err := strconv.ParseBool(raw)
	if err != nil {
		return false, fmt.Errorf("%s must be true or false", name)
	}
	return value, nil
}

func encryptionKey() ([]byte, error) {
	raw := strings.TrimSpace(os.Getenv("SECRET_ENCRYPTION_KEY"))
	if raw == "" {
		return nil, errors.New("missing required environment variable: SECRET_ENCRYPTION_KEY")
	}
	key, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		return nil, errors.New("SECRET_ENCRYPTION_KEY must be valid base64")
	}
	return key, nil
}

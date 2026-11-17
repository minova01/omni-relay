package redis

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strconv"
	"time"

	redisclient "github.com/redis/go-redis/v9"

	"omni-relay/internal/identity/domain"
)

const sessionKeyPrefix = "session:"

type SessionStore struct {
	client *redisclient.Client
}

func NewSessionStore(client *redisclient.Client) *SessionStore {
	return &SessionStore{client: client}
}

func (s *SessionStore) Create(ctx context.Context, userID int64, ttl time.Duration) (string, error) {
	token, err := sessionToken()
	if err != nil {
		return "", err
	}
	if err := s.client.Set(ctx, sessionKey(token), strconv.FormatInt(userID, 10), ttl).Err(); err != nil {
		return "", fmt.Errorf("create session: %w", err)
	}
	return token, nil
}

func (s *SessionStore) Delete(ctx context.Context, token string) error {
	if err := s.client.Del(ctx, sessionKey(token)).Err(); err != nil {
		return fmt.Errorf("delete session: %w", err)
	}
	return nil
}

func (s *SessionStore) GetUserID(ctx context.Context, token string) (int64, error) {
	value, err := s.client.Get(ctx, sessionKey(token)).Result()
	if err == redisclient.Nil {
		return 0, domain.ErrInvalidSession
	}
	if err != nil {
		return 0, fmt.Errorf("get session: %w", err)
	}
	userID, err := strconv.ParseInt(value, 10, 64)
	if err != nil || userID <= 0 {
		return 0, domain.ErrInvalidSession
	}
	return userID, nil
}

func sessionToken() (string, error) {
	value := make([]byte, 32)
	if _, err := rand.Read(value); err != nil {
		return "", fmt.Errorf("generate session token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(value), nil
}

func sessionKey(token string) string {
	hash := sha256.Sum256([]byte(token))
	return sessionKeyPrefix + hex.EncodeToString(hash[:])
}

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

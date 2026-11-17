package application

import (
	"context"
	"errors"
	"fmt"
	"time"

	"omni-relay/internal/accesskey/domain"
	identitydomain "omni-relay/internal/identity/domain"
)

const maxGenerateAttempts = 3

type Service struct {
	keys        domain.Repository
	users       identitydomain.UserRepository
	credentials CredentialService
	balances    BalanceStore
	now         func() time.Time
}

type CreateResult struct {
	APIKey    domain.APIKey
	Plaintext string
}

func NewService(
	keys domain.Repository,
	users identitydomain.UserRepository,
	credentials CredentialService,
	balances BalanceStore,
) *Service {
	return &Service{keys: keys, users: users, credentials: credentials, balances: balances, now: time.Now}
}

func (s *Service) Create(ctx context.Context, userID int64, name string) (CreateResult, error) {
	// 防止用户产出了一样的随机码，这里进行重试
	for range maxGenerateAttempts {
		credential, err := s.credentials.Generate()
		if err != nil {
			return CreateResult{}, err
		}
		key := domain.NewAPIKey(userID, name, credential.Prefix, credential.Hash, credential.Hint, s.now().UnixMilli())
		if err := s.keys.Create(ctx, &key); err != nil {
			if errors.Is(err, domain.ErrPrefixConflict) {
				continue
			}
			return CreateResult{}, err
		}
		return CreateResult{APIKey: key, Plaintext: credential.Plaintext}, nil
	}
	return CreateResult{}, fmt.Errorf("generate unique api key: %w", domain.ErrPrefixConflict)
}

func (s *Service) List(ctx context.Context, userID int64) ([]domain.APIKey, error) {
	return s.keys.ListByUser(ctx, userID)
}

func (s *Service) UpdateStatus(ctx context.Context, userID, keyID int64, status domain.Status) error {
	return s.keys.UpdateStatus(ctx, userID, keyID, status, s.now().UnixMilli())
}

func (s *Service) Delete(ctx context.Context, userID, keyID int64) error {
	return s.keys.SoftDelete(ctx, userID, keyID, s.now().UnixMilli())
}

func (s *Service) Authenticate(ctx context.Context, plaintext string) (Principal, error) {
	prefix, secret, err := s.credentials.Parse(plaintext)
	if err != nil {
		return Principal{}, domain.ErrInvalidAPIKey
	}
	key, err := s.keys.FindByPrefix(ctx, prefix)
	if errors.Is(err, domain.ErrAPIKeyNotFound) {
		return Principal{}, domain.ErrInvalidAPIKey
	}
	if err != nil {
		return Principal{}, err
	}
	if !s.credentials.Verify(key.KeyHash, secret) {
		return Principal{}, domain.ErrInvalidAPIKey
	}
	if key.Status != domain.StatusEnabled {
		return Principal{}, domain.ErrAPIKeyDisabled
	}
	user, err := s.users.FindByID(ctx, key.UserID)
	if errors.Is(err, identitydomain.ErrUserNotFound) {
		return Principal{}, domain.ErrInvalidAPIKey
	}
	if err != nil {
		return Principal{}, err
	}
	if user.Status != identitydomain.StatusEnabled {
		return Principal{}, identitydomain.ErrUserDisabled
	}
	balance, err := s.balances.GetOrInitialize(ctx, user.ID, user.Quota)
	if err != nil {
		return Principal{}, err
	}
	if balance <= 0 {
		return Principal{}, domain.ErrInsufficientQuota
	}
	return Principal{User: user, APIKey: key, Balance: balance}, nil
}

package application

import (
	"context"
	"errors"
	"time"

	"omni-relay/internal/identity/domain"
)

type Service struct {
	users        domain.UserRepository
	passwords    PasswordService
	sessions     SessionStore
	initialQuota int64
	sessionTTL   time.Duration
	now          func() time.Time
}

type LoginResult struct {
	User         domain.User
	SessionToken string
}

func NewService(
	users domain.UserRepository,
	passwords PasswordService,
	sessions SessionStore,
	initialQuota int64,
	sessionTTL time.Duration,
) *Service {
	return &Service{
		users:        users,
		passwords:    passwords,
		sessions:     sessions,
		initialQuota: initialQuota,
		sessionTTL:   sessionTTL,
		now:          time.Now,
	}
}

func (s *Service) Register(ctx context.Context, username, password string) (domain.User, error) {
	passwordHash, err := s.passwords.Hash(password)
	if err != nil {
		return domain.User{}, err
	}
	user := domain.NewUser(username, passwordHash, s.initialQuota, s.now().UnixMilli())
	if err := s.users.Create(ctx, &user); err != nil {
		return domain.User{}, err
	}
	return user, nil
}

func (s *Service) Login(ctx context.Context, username, password string) (LoginResult, error) {
	user, err := s.users.FindByUsername(ctx, username)
	if errors.Is(err, domain.ErrUserNotFound) {
		s.passwords.DummyVerify(password)
		return LoginResult{}, domain.ErrInvalidCredential
	}
	if err != nil {
		return LoginResult{}, err
	}
	if !s.passwords.Verify(user.PasswordHash, password) {
		return LoginResult{}, domain.ErrInvalidCredential
	}
	if user.Status != domain.StatusEnabled {
		return LoginResult{}, domain.ErrUserDisabled
	}

	token, err := s.sessions.Create(ctx, user.ID, s.sessionTTL)
	if err != nil {
		return LoginResult{}, err
	}
	return LoginResult{User: user, SessionToken: token}, nil
}

func (s *Service) Logout(ctx context.Context, token string) error {
	if token == "" {
		return nil
	}
	return s.sessions.Delete(ctx, token)
}

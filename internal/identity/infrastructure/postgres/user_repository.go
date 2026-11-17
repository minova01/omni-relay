package postgres

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"omni-relay/internal/identity/domain"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	err := r.db.WithContext(ctx).
		Select("username", "password_hash", "role", "status", "quota", "used_quota", "request_count", "created_at", "updated_at").
		Create(user).Error
	if err == nil {
		return nil
	}
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return domain.ErrUsernameTaken
	}
	return fmt.Errorf("create user: %w", err)
}

func (r *UserRepository) FindByUsername(ctx context.Context, username string) (domain.User, error) {
	var user domain.User
	err := r.db.WithContext(ctx).
		Select("id", "username", "password_hash", "role", "status", "quota", "used_quota", "request_count", "created_at", "updated_at").
		Where("username = ? AND deleted_at = 0", username).
		First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return domain.User{}, domain.ErrUserNotFound
	}
	if err != nil {
		return domain.User{}, fmt.Errorf("find user by username: %w", err)
	}
	return user, nil
}

func (r *UserRepository) FindByID(ctx context.Context, id int64) (domain.User, error) {
	var user domain.User
	err := r.db.WithContext(ctx).
		Select("id", "username", "password_hash", "role", "status", "quota", "used_quota", "request_count", "created_at", "updated_at").
		Where("id = ? AND deleted_at = 0", id).
		First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return domain.User{}, domain.ErrUserNotFound
	}
	if err != nil {
		return domain.User{}, fmt.Errorf("find user by id: %w", err)
	}
	return user, nil
}

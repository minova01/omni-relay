package postgres

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"omni-relay/internal/accesskey/domain"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, key *domain.APIKey) error {
	err := r.db.WithContext(ctx).
		Select("user_id", "name", "key_prefix", "key_hash", "key_hint", "status", "used_quota", "request_count", "last_used_at", "created_at", "updated_at").
		Create(key).Error
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return domain.ErrPrefixConflict
	}
	if err != nil {
		return fmt.Errorf("create api key: %w", err)
	}
	return nil
}

func (r *Repository) ListByUser(ctx context.Context, userID int64) ([]domain.APIKey, error) {
	var keys []domain.APIKey
	err := r.db.WithContext(ctx).
		Select("id", "user_id", "name", "key_prefix", "key_hint", "status", "used_quota", "request_count", "last_used_at", "created_at", "updated_at").
		Where("user_id = ? AND deleted_at = 0", userID).
		Order("created_at DESC, id DESC").
		Find(&keys).Error
	if err != nil {
		return nil, fmt.Errorf("list api keys: %w", err)
	}
	return keys, nil
}

func (r *Repository) FindByPrefix(ctx context.Context, prefix string) (domain.APIKey, error) {
	var key domain.APIKey
	err := r.db.WithContext(ctx).
		Select("id", "user_id", "name", "key_prefix", "key_hash", "key_hint", "status", "used_quota", "request_count", "last_used_at", "created_at", "updated_at").
		Where("key_prefix = ? AND deleted_at = 0", prefix).
		First(&key).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return domain.APIKey{}, domain.ErrAPIKeyNotFound
	}
	if err != nil {
		return domain.APIKey{}, fmt.Errorf("find api key by prefix: %w", err)
	}
	return key, nil
}

func (r *Repository) UpdateStatus(ctx context.Context, userID, keyID int64, status domain.Status, now int64) error {
	result := r.db.WithContext(ctx).
		Model(&domain.APIKey{}).
		Where("id = ? AND user_id = ? AND deleted_at = 0", keyID, userID).
		Updates(map[string]any{"status": status, "updated_at": now})
	if result.Error != nil {
		return fmt.Errorf("update api key status: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return domain.ErrAPIKeyNotFound
	}
	return nil
}

func (r *Repository) SoftDelete(ctx context.Context, userID, keyID, now int64) error {
	result := r.db.WithContext(ctx).
		Model(&domain.APIKey{}).
		Where("id = ? AND user_id = ? AND deleted_at = 0", keyID, userID).
		Updates(map[string]any{"deleted_at": now, "updated_at": now})
	if result.Error != nil {
		return fmt.Errorf("delete api key: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return domain.ErrAPIKeyNotFound
	}
	return nil
}

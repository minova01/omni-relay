package postgres

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"omni-relay/internal/catalog/domain"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateChannel(ctx context.Context, channel *domain.Channel) error {
	err := r.db.WithContext(ctx).
		Select("name", "channel_type", "base_url", "secret_ciphertext", "status", "used_quota", "request_count", "created_at", "updated_at").
		Create(channel).Error
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return domain.ErrChannelNameTaken
	}
	if err != nil {
		return fmt.Errorf("create channel: %w", err)
	}
	return nil
}

func (r *Repository) ListChannels(ctx context.Context) ([]domain.Channel, error) {
	var channels []domain.Channel
	err := r.db.WithContext(ctx).
		Select("id", "name", "channel_type", "base_url", "status", "used_quota", "request_count", "created_at", "updated_at").
		Where("deleted_at = 0").
		Order("created_at DESC, id DESC").
		Find(&channels).Error
	if err != nil {
		return nil, fmt.Errorf("list channels: %w", err)
	}
	return channels, nil
}

func (r *Repository) UpdateChannel(ctx context.Context, id int64, name, channelType, baseURL, encryptedSecret string, now int64) error {
	values := map[string]any{
		"name": name, "channel_type": channelType, "base_url": baseURL, "updated_at": now,
	}
	if encryptedSecret != "" {
		values["secret_ciphertext"] = encryptedSecret
	}
	result := r.db.WithContext(ctx).Model(&domain.Channel{}).
		Where("id = ? AND deleted_at = 0", id).Updates(values)
	if errors.Is(result.Error, gorm.ErrDuplicatedKey) {
		return domain.ErrChannelNameTaken
	}
	return mutationResult(result, domain.ErrChannelNotFound, "update channel")
}

func (r *Repository) UpdateChannelStatus(ctx context.Context, id int64, status domain.Status, now int64) error {
	result := r.db.WithContext(ctx).Model(&domain.Channel{}).
		Where("id = ? AND deleted_at = 0", id).
		Updates(map[string]any{"status": status, "updated_at": now})
	return mutationResult(result, domain.ErrChannelNotFound, "update channel status")
}

func (r *Repository) CreateModel(ctx context.Context, model *domain.Model) error {
	err := r.db.WithContext(ctx).
		Select("code", "display_name", "status", "input_price", "output_price", "created_at", "updated_at").
		Create(model).Error
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return domain.ErrModelCodeTaken
	}
	if err != nil {
		return fmt.Errorf("create model: %w", err)
	}
	return nil
}

func (r *Repository) ListModels(ctx context.Context) ([]domain.Model, error) {
	var models []domain.Model
	err := r.db.WithContext(ctx).
		Select("id", "code", "display_name", "status", "input_price", "output_price", "created_at", "updated_at").
		Where("deleted_at = 0").Order("created_at DESC, id DESC").Find(&models).Error
	if err != nil {
		return nil, fmt.Errorf("list models: %w", err)
	}
	return models, nil
}

func (r *Repository) UpdateModel(ctx context.Context, id int64, code, displayName string, inputPrice, outputPrice, now int64) error {
	result := r.db.WithContext(ctx).Model(&domain.Model{}).
		Where("id = ? AND deleted_at = 0", id).
		Updates(map[string]any{
			"code": code, "display_name": displayName, "input_price": inputPrice,
			"output_price": outputPrice, "updated_at": now,
		})
	if errors.Is(result.Error, gorm.ErrDuplicatedKey) {
		return domain.ErrModelCodeTaken
	}
	return mutationResult(result, domain.ErrModelNotFound, "update model")
}

func (r *Repository) UpdateModelStatus(ctx context.Context, id int64, status domain.Status, now int64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var model domain.Model
		if err := tx.Select("id").Where("id = ? AND deleted_at = 0", id).First(&model).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return domain.ErrModelNotFound
			}
			return fmt.Errorf("find model for status update: %w", err)
		}
		if status == domain.StatusEnabled {
			var count int64
			if err := tx.Model(&domain.ChannelModel{}).
				Where("model_id = ? AND status = ? AND deleted_at = 0", id, domain.StatusEnabled).
				Count(&count).Error; err != nil {
				return fmt.Errorf("count active model routes: %w", err)
			}
			if count == 0 {
				return domain.ErrModelHasNoActiveRoute
			}
		}
		if err := tx.Model(&domain.Model{}).Where("id = ?", id).
			Updates(map[string]any{"status": status, "updated_at": now}).Error; err != nil {
			return fmt.Errorf("update model status: %w", err)
		}
		return nil
	})
}

func (r *Repository) CreateChannelModel(ctx context.Context, mapping *domain.ChannelModel) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := ensureReference(tx, &domain.Channel{}, mapping.ChannelID, domain.ErrChannelNotFound); err != nil {
			return err
		}
		if err := ensureReference(tx, &domain.Model{}, mapping.ModelID, domain.ErrModelNotFound); err != nil {
			return err
		}
		err := tx.Select("channel_id", "model_id", "upstream_model", "priority", "supports_stream", "supports_tools", "status", "created_at", "updated_at").
			Create(mapping).Error
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return domain.ErrChannelModelTaken
		}
		if err != nil {
			return fmt.Errorf("create channel model: %w", err)
		}
		return nil
	})
}

func (r *Repository) ListChannelModels(ctx context.Context) ([]domain.ChannelModel, error) {
	var mappings []domain.ChannelModel
	err := r.db.WithContext(ctx).
		Select("id", "channel_id", "model_id", "upstream_model", "priority", "supports_stream", "supports_tools", "status", "created_at", "updated_at").
		Where("deleted_at = 0").Order("priority DESC, id ASC").Find(&mappings).Error
	if err != nil {
		return nil, fmt.Errorf("list channel models: %w", err)
	}
	return mappings, nil
}

func (r *Repository) UpdateChannelModel(ctx context.Context, id, channelID, modelID int64, upstreamModel string, priority int, supportsStream, supportsTools bool, now int64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := ensureReference(tx, &domain.Channel{}, channelID, domain.ErrChannelNotFound); err != nil {
			return err
		}
		if err := ensureReference(tx, &domain.Model{}, modelID, domain.ErrModelNotFound); err != nil {
			return err
		}
		result := tx.Model(&domain.ChannelModel{}).Where("id = ? AND deleted_at = 0", id).
			Updates(map[string]any{
				"channel_id": channelID, "model_id": modelID, "upstream_model": upstreamModel,
				"priority": priority, "supports_stream": supportsStream, "supports_tools": supportsTools,
				"updated_at": now,
			})
		if errors.Is(result.Error, gorm.ErrDuplicatedKey) {
			return domain.ErrChannelModelTaken
		}
		return mutationResult(result, domain.ErrChannelModelNotFound, "update channel model")
	})
}

func (r *Repository) UpdateChannelModelStatus(ctx context.Context, id int64, status domain.Status, now int64) error {
	result := r.db.WithContext(ctx).Model(&domain.ChannelModel{}).
		Where("id = ? AND deleted_at = 0", id).
		Updates(map[string]any{"status": status, "updated_at": now})
	return mutationResult(result, domain.ErrChannelModelNotFound, "update channel model status")
}

func (r *Repository) DeleteChannelModel(ctx context.Context, id, now int64) error {
	result := r.db.WithContext(ctx).Model(&domain.ChannelModel{}).
		Where("id = ? AND deleted_at = 0", id).
		Updates(map[string]any{"deleted_at": now, "updated_at": now})
	return mutationResult(result, domain.ErrChannelModelNotFound, "delete channel model")
}

func ensureReference(tx *gorm.DB, model any, id int64, notFound error) error {
	var count int64
	if err := tx.Model(model).Where("id = ? AND deleted_at = 0", id).Count(&count).Error; err != nil {
		return fmt.Errorf("check catalog reference: %w", err)
	}
	if count == 0 {
		return notFound
	}
	return nil
}

func mutationResult(result *gorm.DB, notFound error, operation string) error {
	if result.Error != nil {
		return fmt.Errorf("%s: %w", operation, result.Error)
	}
	if result.RowsAffected == 0 {
		return notFound
	}
	return nil
}

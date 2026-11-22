package postgres

import (
	"context"
	"errors"
	"omni-relay/internal/catalog/domain"

	"gorm.io/gorm"
)

var ErrModelUnavailable = errors.New("model unavailable")

type Route struct {
	Channel domain.Channel
	Mapping domain.ChannelModel
}
type Repository struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) *Repository { return &Repository{db: db} }

// transfer the model code to the corresponding channel and mapping
func (r *Repository) Select(ctx context.Context, code string) (Route, error) {
	var model domain.Model
	if err := r.db.WithContext(ctx).Where("code = ? AND status = 1 AND deleted_at = 0", code).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return Route{}, ErrModelUnavailable
		}
		return Route{}, err
	}
	var mapping domain.ChannelModel
	var channel domain.Channel
	err := r.db.WithContext(ctx).Model(&domain.ChannelModel{}).
		Select("channel_models.*").Joins("JOIN channels ON channels.id = channel_models.channel_id").
		Where("channel_models.model_id = ? AND channel_models.status = 1 AND channel_models.deleted_at = 0 AND channels.status = 1 AND channels.deleted_at = 0", model.ID).
		Order("channel_models.priority DESC, channel_models.id ASC").First(&mapping).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return Route{}, ErrModelUnavailable
	}
	if err != nil {
		return Route{}, err
	}
	if err = r.db.WithContext(ctx).Where("id = ? AND deleted_at = 0", mapping.ChannelID).First(&channel).Error; err != nil {
		return Route{}, err
	}
	return Route{Channel: channel, Mapping: mapping}, nil
}

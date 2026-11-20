package domain

import "context"

type Repository interface {
	CreateChannel(context.Context, *Channel) error
	ListChannels(context.Context) ([]Channel, error)
	UpdateChannel(context.Context, int64, string, string, string, string, int64) error
	UpdateChannelStatus(context.Context, int64, Status, int64) error

	CreateModel(context.Context, *Model) error
	ListModels(context.Context) ([]Model, error)
	UpdateModel(context.Context, int64, string, string, int64, int64, int64) error
	UpdateModelStatus(context.Context, int64, Status, int64) error

	CreateChannelModel(context.Context, *ChannelModel) error
	ListChannelModels(context.Context) ([]ChannelModel, error)
	UpdateChannelModel(context.Context, int64, int64, int64, string, int, bool, bool, int64) error
	UpdateChannelModelStatus(context.Context, int64, Status, int64) error
	DeleteChannelModel(context.Context, int64, int64) error
}

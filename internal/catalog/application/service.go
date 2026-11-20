package application

import (
	"context"
	"net/url"
	"time"

	"omni-relay/internal/catalog/domain"
)

type Service struct {
	repository domain.Repository
	cipher     SecretCipher
	allowHTTP  bool
	now        func() time.Time
}

type ChannelInput struct {
	Name        string
	ChannelType string
	BaseURL     string
	Secret      string
}

type ModelInput struct {
	Code        string
	DisplayName string
	InputPrice  int64
	OutputPrice int64
}

type ChannelModelInput struct {
	ChannelID      int64
	ModelID        int64
	UpstreamModel  string
	Priority       int
	SupportsStream bool
	SupportsTools  bool
}

func NewService(repository domain.Repository, cipher SecretCipher, allowHTTP bool) *Service {
	return &Service{repository: repository, cipher: cipher, allowHTTP: allowHTTP, now: time.Now}
}

func (s *Service) CreateChannel(ctx context.Context, input ChannelInput) (domain.Channel, error) {
	if !s.validBaseURL(input.BaseURL) {
		return domain.Channel{}, domain.ErrInvalidBaseURL
	}
	encryptedSecret, err := s.cipher.Encrypt(input.Secret)
	if err != nil {
		return domain.Channel{}, err
	}
	channel := domain.NewChannel(input.Name, input.ChannelType, input.BaseURL, encryptedSecret, s.now().UnixMilli())
	if err := s.repository.CreateChannel(ctx, &channel); err != nil {
		return domain.Channel{}, err
	}
	return channel, nil
}

func (s *Service) ListChannels(ctx context.Context) ([]domain.Channel, error) {
	return s.repository.ListChannels(ctx)
}

func (s *Service) UpdateChannel(ctx context.Context, id int64, input ChannelInput) error {
	if !s.validBaseURL(input.BaseURL) {
		return domain.ErrInvalidBaseURL
	}
	encryptedSecret := ""
	var err error
	if input.Secret != "" {
		encryptedSecret, err = s.cipher.Encrypt(input.Secret)
		if err != nil {
			return err
		}
	}
	return s.repository.UpdateChannel(ctx, id, input.Name, input.ChannelType, input.BaseURL, encryptedSecret, s.now().UnixMilli())
}

func (s *Service) UpdateChannelStatus(ctx context.Context, id int64, status domain.Status) error {
	return s.repository.UpdateChannelStatus(ctx, id, status, s.now().UnixMilli())
}

func (s *Service) CreateModel(ctx context.Context, input ModelInput) (domain.Model, error) {
	model := domain.NewModel(input.Code, input.DisplayName, input.InputPrice, input.OutputPrice, s.now().UnixMilli())
	if err := s.repository.CreateModel(ctx, &model); err != nil {
		return domain.Model{}, err
	}
	return model, nil
}

func (s *Service) ListModels(ctx context.Context) ([]domain.Model, error) {
	return s.repository.ListModels(ctx)
}

func (s *Service) UpdateModel(ctx context.Context, id int64, input ModelInput) error {
	return s.repository.UpdateModel(ctx, id, input.Code, input.DisplayName, input.InputPrice, input.OutputPrice, s.now().UnixMilli())
}

func (s *Service) UpdateModelStatus(ctx context.Context, id int64, status domain.Status) error {
	return s.repository.UpdateModelStatus(ctx, id, status, s.now().UnixMilli())
}

func (s *Service) CreateChannelModel(ctx context.Context, input ChannelModelInput) (domain.ChannelModel, error) {
	mapping := domain.NewChannelModel(input.ChannelID, input.ModelID, input.UpstreamModel, input.Priority, input.SupportsStream, input.SupportsTools, s.now().UnixMilli())
	if err := s.repository.CreateChannelModel(ctx, &mapping); err != nil {
		return domain.ChannelModel{}, err
	}
	return mapping, nil
}

func (s *Service) ListChannelModels(ctx context.Context) ([]domain.ChannelModel, error) {
	return s.repository.ListChannelModels(ctx)
}

func (s *Service) UpdateChannelModel(ctx context.Context, id int64, input ChannelModelInput) error {
	return s.repository.UpdateChannelModel(ctx, id, input.ChannelID, input.ModelID, input.UpstreamModel, input.Priority, input.SupportsStream, input.SupportsTools, s.now().UnixMilli())
}

func (s *Service) UpdateChannelModelStatus(ctx context.Context, id int64, status domain.Status) error {
	return s.repository.UpdateChannelModelStatus(ctx, id, status, s.now().UnixMilli())
}

func (s *Service) DeleteChannelModel(ctx context.Context, id int64) error {
	return s.repository.DeleteChannelModel(ctx, id, s.now().UnixMilli())
}

func (s *Service) validBaseURL(raw string) bool {
	parsed, err := url.ParseRequestURI(raw)
	if err != nil || parsed.Host == "" || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
		return false
	}
	return parsed.Scheme == "https" || (s.allowHTTP && parsed.Scheme == "http")
}

package domain

type Status int16

const (
	StatusEnabled  Status = 1
	StatusDisabled Status = 2
)

type Channel struct {
	ID               int64
	Name             string
	ChannelType      string
	BaseURL          string
	SecretCiphertext string
	Status           Status
	UsedQuota        int64
	RequestCount     int64
	CreatedAt        int64
	UpdatedAt        int64
}

type Model struct {
	ID          int64
	Code        string
	DisplayName string
	Status      Status
	InputPrice  int64
	OutputPrice int64
	CreatedAt   int64
	UpdatedAt   int64
}

type ChannelModel struct {
	ID             int64
	ChannelID      int64
	ModelID        int64
	UpstreamModel  string
	Priority       int
	SupportsStream bool
	SupportsTools  bool
	Status         Status
	CreatedAt      int64
	UpdatedAt      int64
}

func NewChannel(name, channelType, baseURL, encryptedSecret string, now int64) Channel {
	return Channel{
		Name:             name,
		ChannelType:      channelType,
		BaseURL:          baseURL,
		SecretCiphertext: encryptedSecret,
		Status:           StatusEnabled,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
}

func NewModel(code, displayName string, inputPrice, outputPrice, now int64) Model {
	return Model{
		Code:        code,
		DisplayName: displayName,
		Status:      StatusDisabled,
		InputPrice:  inputPrice,
		OutputPrice: outputPrice,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

func NewChannelModel(channelID, modelID int64, upstreamModel string, priority int, supportsStream, supportsTools bool, now int64) ChannelModel {
	return ChannelModel{
		ChannelID:      channelID,
		ModelID:        modelID,
		UpstreamModel:  upstreamModel,
		Priority:       priority,
		SupportsStream: supportsStream,
		SupportsTools:  supportsTools,
		Status:         StatusEnabled,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}

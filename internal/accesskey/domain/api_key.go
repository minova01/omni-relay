package domain

type Status int16

const (
	StatusEnabled  Status = 1
	StatusDisabled Status = 2
)

type APIKey struct {
	ID           int64
	UserID       int64
	Name         string
	KeyPrefix    string
	KeyHash      string
	KeyHint      string
	Status       Status
	UsedQuota    int64
	RequestCount int64
	LastUsedAt   int64
	CreatedAt    int64
	UpdatedAt    int64
}

func NewAPIKey(userID int64, name, prefix, hash, hint string, now int64) APIKey {
	return APIKey{
		UserID:    userID,
		Name:      name,
		KeyPrefix: prefix,
		KeyHash:   hash,
		KeyHint:   hint,
		Status:    StatusEnabled,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

package domain

type Role int16

const (
	RoleUser  Role = 1
	RoleAdmin Role = 10
)

type Status int16

const (
	StatusEnabled  Status = 1
	StatusDisabled Status = 2
)

type User struct {
	ID           int64
	Username     string
	PasswordHash string
	Role         Role
	Status       Status
	Quota        int64
	UsedQuota    int64
	RequestCount int64
	CreatedAt    int64
	UpdatedAt    int64
}

func NewUser(username, passwordHash string, initialQuota, now int64) User {
	return User{
		Username:     username,
		PasswordHash: passwordHash,
		Role:         RoleUser,
		Status:       StatusEnabled,
		Quota:        initialQuota,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

package model

import "omni-relay/internal/identity/domain"

type UserResponse struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	Role      int16  `json:"role"`
	Status    int16  `json:"status"`
	Quota     int64  `json:"quota"`
	CreatedAt int64  `json:"created_at"`
}

func NewUserResponse(user domain.User) UserResponse {
	return UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		Role:      int16(user.Role),
		Status:    int16(user.Status),
		Quota:     user.Quota,
		CreatedAt: user.CreatedAt,
	}
}

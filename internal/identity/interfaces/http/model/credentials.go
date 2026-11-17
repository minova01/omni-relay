package model

type CredentialsRequest struct {
	Username string `json:"username" validate:"required,username"`
	Password string `json:"password" validate:"required,password_bytes"`
}

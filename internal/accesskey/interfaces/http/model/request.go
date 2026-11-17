package model

type CreateAPIKeyRequest struct {
	Name string `json:"name" validate:"required,notblank,max=64"`
}

type UpdateAPIKeyStatusRequest struct {
	Status int16 `json:"status" validate:"required,oneof=1 2"`
}

package model

type CreateAPIKeyRequest struct {
	Name string `json:"name" validate:"required,notblank,max=64"`
}

type UpdateAPIKeyStatusRequest struct {
	ID     int64 `json:"id" validate:"required,gt=0"`
	Status int16 `json:"status" validate:"required,oneof=1 2"`
}

type DeleteAPIKeyRequest struct {
	ID int64 `json:"id" validate:"required,gt=0"`
}

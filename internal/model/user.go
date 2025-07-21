package model

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID          uuid.UUID `json:"id"`
	Login       string    `json:"login"`
	Password    string    `json:"-"`
	Token       string    `json:"-"`
	TokenExpiry time.Time `json:"-"`
	CreatedAt   time.Time `json:"created_at,omitempty"`
}

type UserResponse struct {
	ID        string    `json:"id"`
	Login     string    `json:"login"`
	CreatedAt time.Time `json:"created_at"`
}

type UserListResponse struct {
	Users []UserResponse `json:"users"`
	Total int            `json:"total"`
}

type UploadedFile struct {
	Filename string
	Data     []byte
	Size     int64
}

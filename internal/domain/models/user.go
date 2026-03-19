package models

import (
	"mapps_auth/internal/domain/utils"

	"github.com/google/uuid"
)

type User struct {
	ID       string
	Email    string
	PassHash string
	Username string
}

func NewUser(email, password, username string) (*User, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}
	hash, err := utils.HashPassword(password)
	if err != nil {
		return nil, err
	}
	return &User{
		ID:       id.String(),
		Email:    email,
		PassHash: hash,
		Username: username,
	}, nil
}

package domain

import "github.com/google/uuid"

type User struct {
	ID             uuid.UUID
	Name           string
	Email          string
	HashedPassword string
}

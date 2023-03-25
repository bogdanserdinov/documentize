package users

import (
	"context"
	"github.com/google/uuid"
	"time"
)

type DB interface {
	Create(ctx context.Context, user User) error
	Get(ctx context.Context, id uuid.UUID) (User, error)
	List(ctx context.Context) ([]User, error)
	UpdateStatus(ctx context.Context, id uuid.UUID) error
}

type User struct {
	ID        uuid.UUID
	Name      string
	Email     string
	Status    Status
	CreatedAt time.Time
}

type Status string

const (
	StatusUngenerated Status = "ungenerated"
	StatusGenerated   Status = "generated"
)

package model

import (
	"time"
)

type Dream struct {
	ID        int64
	Content   string
	CreatedAt time.Time
	UpdatedAt time.Time
}

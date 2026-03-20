package model

import "time"

type PrimingCache struct {
	Source    string
	Payload   []string
	FetchedAt time.Time
}

type PrimingLog struct {
	ID        int64
	CreatedAt time.Time
	Source    string
	Outcome   string
	Detail    string
	Content   string
}

type PrimingContent struct {
	ID        int64
	Source    string
	Title     string
	Content   string
	Category  string
	URL       string
	CreatedAt time.Time
	UpdatedAt time.Time
}

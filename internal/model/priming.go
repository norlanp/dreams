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

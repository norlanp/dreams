package priming

import (
	"net/http"
	"time"
)

func NewDefaultGenerator(store interface {
	LogStore
	analysisStore
	cacheStore
}) *Generator {
	client := &http.Client{Timeout: 8 * time.Second}
	return NewGenerator(
		store,
		NewPersonalizedSource(store),
		NewRedditSource(client, store),
		NewAISource(client, store),
		NewTemplateSource(),
	)
}

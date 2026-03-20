package priming

import (
	"net/http"
	"time"
)

type DefaultStore interface {
	LogStore
	analysisStore
	cacheStore
	contentStore
}

func NewDefaultGenerator(store DefaultStore) *Generator {
	client := &http.Client{Timeout: 8 * time.Second}
	return NewGenerator(
		store,
		NewPersonalizedSource(store),
		NewContentSource(store),
		NewAISource(client, store),
		NewTemplateSource(),
	)
}

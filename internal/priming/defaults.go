package priming

import (
	"net"
	"net/http"
	"time"
)

type DefaultStore interface {
	LogStore
	analysisStore
	cacheStore
	contentStore
}

func createPooledHTTPClient() *http.Client {
	return &http.Client{
		Timeout: 8 * time.Second,
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   5 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			TLSHandshakeTimeout:   5 * time.Second,
			ResponseHeaderTimeout: 5 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			MaxIdleConns:          10,
			MaxIdleConnsPerHost:   5,
			IdleConnTimeout:       30 * time.Second,
		},
	}
}

func NewDefaultGenerator(store DefaultStore) *Generator {
	client := createPooledHTTPClient()
	return NewGenerator(
		store,
		NewPersonalizedSource(store),
		NewContentSource(store),
		NewAISource(client, store),
		NewTemplateSource(),
	)
}

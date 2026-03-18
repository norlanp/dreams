package priming

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"dreams/internal/model"
)

type cacheStub struct {
	cache      *model.PrimingCache
	err        error
	saved      []string
	saveSource string
	saveCalls  int
	getCalls   int
}

type ttlCacheStub struct {
	cache      *model.PrimingCache
	saved      []string
	saveSource string
	saveCalls  int
	getCalls   int
	lastSource string
	lastNow    time.Time
	lastTTL    time.Duration
}

func (s *cacheStub) GetFreshPrimingCache(ctx context.Context, source string, now time.Time, ttl time.Duration) (*model.PrimingCache, error) {
	_ = ctx
	_ = source
	_ = now
	_ = ttl
	s.getCalls++
	if s.err != nil {
		return nil, s.err
	}
	return s.cache, nil
}

func (s *cacheStub) SavePrimingCache(ctx context.Context, source string, payload []string, fetchedAt time.Time) error {
	_ = ctx
	_ = fetchedAt
	s.saveCalls++
	s.saveSource = source
	s.saved = payload
	return nil
}

func (s *ttlCacheStub) GetFreshPrimingCache(ctx context.Context, source string, now time.Time, ttl time.Duration) (*model.PrimingCache, error) {
	_ = ctx
	s.getCalls++
	s.lastSource = source
	s.lastNow = now
	s.lastTTL = ttl

	if s.cache == nil {
		return nil, nil
	}
	if s.cache.FetchedAt.Before(now.Add(-ttl)) {
		return nil, nil
	}

	return s.cache, nil
}

func (s *ttlCacheStub) SavePrimingCache(ctx context.Context, source string, payload []string, fetchedAt time.Time) error {
	_ = ctx
	_ = fetchedAt
	s.saveCalls++
	s.saveSource = source
	s.saved = payload
	return nil
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestRedditSource_ShouldUseFreshCacheBeforeNetworkFetch(t *testing.T) {
	store := &cacheStub{cache: &model.PrimingCache{Source: string(SourceCommunity), Payload: []string{"cached post"}}}
	client := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return nil, errors.New("network should not be called")
	})}

	source := NewRedditSource(client, store)
	text, err := source.Next(context.Background())
	if err != nil {
		t.Fatalf("expected cache hit success, got %v", err)
	}
	if text != "cached post" {
		t.Fatalf("expected cached content, got %q", text)
	}
	if store.saveCalls != 0 {
		t.Fatalf("expected no cache write on hit, got %d", store.saveCalls)
	}
}

func TestRedditSource_ShouldFetchFilterAndCacheOnMiss(t *testing.T) {
	store := &cacheStub{}
	body := `{"data":{"children":[
		{"data":{"title":"Good","selftext":"` + strings.Repeat("a", 90) + `","stickied":false,"over_18":false}},
		{"data":{"title":"Short","selftext":"tiny","stickied":false,"over_18":false}}
	]}}`
	client := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(body)),
			Header:     make(http.Header),
		}, nil
	})}

	source := NewRedditSource(client, store)
	text, err := source.Next(context.Background())
	if err != nil {
		t.Fatalf("expected reddit fetch success, got %v", err)
	}
	if !strings.Contains(text, "Good") {
		t.Fatalf("expected rendered post title, got %q", text)
	}
	if store.saveCalls != 1 {
		t.Fatalf("expected cache save on miss, got %d", store.saveCalls)
	}
	if len(store.saved) != 1 {
		t.Fatalf("expected only text-rich post cached, got %d", len(store.saved))
	}
}

func TestRedditSource_ShouldTreatTTLBoundaryAsFreshInPrimingPath(t *testing.T) {
	now := time.Date(2026, 3, 1, 22, 0, 0, 0, time.UTC)
	store := &ttlCacheStub{cache: &model.PrimingCache{
		Source:    string(SourceCommunity),
		Payload:   []string{"boundary cached post"},
		FetchedAt: now.Add(-24 * time.Hour),
	}}

	networkCalls := 0
	client := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		_ = req
		networkCalls++
		return nil, errors.New("network should not be called when cache is fresh")
	})}

	source := NewRedditSource(client, store)
	source.nowFn = func() time.Time { return now }

	text, err := source.Next(context.Background())
	if err != nil {
		t.Fatalf("expected fresh cache boundary hit, got %v", err)
	}
	if text != "boundary cached post" {
		t.Fatalf("expected boundary cached payload, got %q", text)
	}
	if networkCalls != 0 {
		t.Fatalf("expected no network calls for boundary-fresh cache, got %d", networkCalls)
	}
	if store.lastTTL != 24*time.Hour {
		t.Fatalf("expected priming ttl 24h, got %s", store.lastTTL)
	}
	if store.lastSource != string(SourceCommunity) {
		t.Fatalf("expected community cache source, got %q", store.lastSource)
	}
	if store.saveCalls != 0 {
		t.Fatalf("expected no cache write on boundary-fresh hit, got %d", store.saveCalls)
	}
}

func TestRedditSource_ShouldFetchWhenCacheIsPastTTLInPrimingPath(t *testing.T) {
	now := time.Date(2026, 3, 1, 22, 0, 0, 0, time.UTC)
	store := &ttlCacheStub{cache: &model.PrimingCache{
		Source:    string(SourceCommunity),
		Payload:   []string{"stale post"},
		FetchedAt: now.Add(-24*time.Hour - time.Second),
	}}
	body := `{"data":{"children":[{"data":{"title":"Fresh","selftext":"` + strings.Repeat("b", 90) + `","stickied":false,"over_18":false}}]}}`

	networkCalls := 0
	client := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		networkCalls++
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(body)),
			Header:     make(http.Header),
		}, nil
	})}

	source := NewRedditSource(client, store)
	source.nowFn = func() time.Time { return now }

	text, err := source.Next(context.Background())
	if err != nil {
		t.Fatalf("expected fetch success for stale cache, got %v", err)
	}
	if !strings.Contains(text, "Fresh") {
		t.Fatalf("expected network-fetched content, got %q", text)
	}
	if networkCalls != 1 {
		t.Fatalf("expected one network fetch for stale cache, got %d", networkCalls)
	}
	if store.lastTTL != 24*time.Hour {
		t.Fatalf("expected priming ttl 24h, got %s", store.lastTTL)
	}
	if store.saveCalls != 1 {
		t.Fatalf("expected cache refresh write after stale fetch, got %d", store.saveCalls)
	}
}

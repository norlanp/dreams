package priming

import (
	"context"
	"errors"
	"fmt"
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

func starterThreadJSON(selftext string) string {
	return fmt.Sprintf(`[
		{"data":{"children":[{"data":{"title":"START HERE! - Beginner Guides, FAQs, and Resources","selftext":%q,"stickied":true,"over_18":false}}]}},
		{"data":{"children":[]}}
	]`, selftext)
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
	body := starterThreadJSON(strings.Join([]string{
		"For more on the basics, [jump into our Wiki](https://www.reddit.com/r/LucidDreaming/wiki/index) and [read the FAQ](https://www.reddit.com/r/LucidDreaming/wiki/faq).",
		"Increase your dream recall (by writing a dream journal), question your reality (with reality checks), and set the intention for lucidity: [Here is a quick beginner guide](https://www.reddit.com/r/LucidDreaming/comments/rsvp7/the_three_steps_for_learning_to_lucid_dream/).",
	}, "\n"))
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
	if !strings.Contains(text, "Guides and resources") {
		t.Fatalf("expected distilled guides content, got %q", text)
	}
	if !strings.Contains(text, "quick beginner guide") {
		t.Fatalf("expected beginner guide link in distilled content, got %q", text)
	}
	if store.saveCalls != 1 {
		t.Fatalf("expected cache save on miss, got %d", store.saveCalls)
	}
	if len(store.saved) != 1 {
		t.Fatalf("expected one distilled starter payload cached, got %d", len(store.saved))
	}
}

func TestRedditSource_ShouldRetryAfterForbiddenResponse(t *testing.T) {
	store := &cacheStub{}
	body := starterThreadJSON("[Quick beginner guide](https://www.reddit.com/r/LucidDreaming/comments/rsvp7/the_three_steps_for_learning_to_lucid_dream/)")

	callCount := 0
	requestedURLs := []string{}
	client := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		requestedURLs = append(requestedURLs, req.URL.String())
		callCount++
		if callCount == 1 {
			return &http.Response{
				StatusCode: http.StatusForbidden,
				Body:       io.NopCloser(strings.NewReader("forbidden")),
				Header:     make(http.Header),
			}, nil
		}

		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(body)),
			Header:     make(http.Header),
		}, nil
	})}

	source := NewRedditSource(client, store)
	text, err := source.Next(context.Background())
	if err != nil {
		t.Fatalf("expected reddit retry success, got %v", err)
	}
	if !strings.Contains(strings.ToLower(text), "quick beginner guide") {
		t.Fatalf("expected fallback endpoint content, got %q", text)
	}
	if callCount != 2 {
		t.Fatalf("expected retry after forbidden response, got %d calls", callCount)
	}
	if len(requestedURLs) != 2 {
		t.Fatalf("expected two reddit requests, got %d", len(requestedURLs))
	}
	if requestedURLs[0] != redditURL {
		t.Fatalf("expected primary reddit endpoint first, got %q", requestedURLs[0])
	}
	if requestedURLs[1] != redditFallbackURL {
		t.Fatalf("expected fallback reddit endpoint second, got %q", requestedURLs[1])
	}
	if store.saveCalls != 1 {
		t.Fatalf("expected cache save after retry success, got %d", store.saveCalls)
	}
}

func TestRedditSource_ShouldUseBundledGuidesWhenStarterThreadForbidden(t *testing.T) {
	store := &cacheStub{}
	callCount := 0
	requestedURLs := []string{}
	client := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		requestedURLs = append(requestedURLs, req.URL.String())
		callCount++
		return &http.Response{
			StatusCode: http.StatusForbidden,
			Body:       io.NopCloser(strings.NewReader("forbidden")),
			Header:     make(http.Header),
		}, nil
	})}

	source := NewRedditSource(client, store)
	text, err := source.Next(context.Background())
	if err != nil {
		t.Fatalf("expected bundled starter guides fallback, got %v", err)
	}
	if !strings.Contains(text, "Guides and resources") {
		t.Fatalf("expected bundled guides heading, got %q", text)
	}
	if !strings.Contains(text, "Three Steps Beginner Guide") {
		t.Fatalf("expected bundled beginner guide link, got %q", text)
	}
	if callCount != 2 {
		t.Fatalf("expected primary+fallback fetch attempts, got %d", callCount)
	}
	if len(requestedURLs) != 2 {
		t.Fatalf("expected two requests before bundled fallback, got %d", len(requestedURLs))
	}
	if requestedURLs[0] != redditURL {
		t.Fatalf("expected primary endpoint first, got %q", requestedURLs[0])
	}
	if requestedURLs[1] != redditFallbackURL {
		t.Fatalf("expected fallback endpoint second, got %q", requestedURLs[1])
	}
	if store.saveCalls != 1 {
		t.Fatalf("expected bundled guides to be cached, got %d save calls", store.saveCalls)
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
	body := starterThreadJSON("[Meditation for lucid dreaming](https://www.reddit.com/r/LucidDreaming/comments/36dvtb/meditation_for_lucid_dreaming_scientific_evidence/)")

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
	if !strings.Contains(text, "Meditation for lucid dreaming") {
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

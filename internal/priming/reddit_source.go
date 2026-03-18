package priming

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"dreams/internal/model"
)

const redditURL = "https://www.reddit.com/r/LucidDreaming/.json"

type cacheStore interface {
	GetFreshPrimingCache(ctx context.Context, source string, now time.Time, ttl time.Duration) (*model.PrimingCache, error)
	SavePrimingCache(ctx context.Context, source string, payload []string, fetchedAt time.Time) error
}

type RedditSource struct {
	httpClient *http.Client
	store      cacheStore
	nowFn      func() time.Time
	index      int
}

func NewRedditSource(client *http.Client, store cacheStore) *RedditSource {
	if client == nil {
		client = http.DefaultClient
	}
	return &RedditSource{httpClient: client, store: store, nowFn: func() time.Time { return time.Now().UTC() }}
}

func (s *RedditSource) Label() SourceLabel {
	return SourceCommunity
}

func (s *RedditSource) Next(ctx context.Context) (string, error) {
	items, err := s.cachedOrFetched(ctx)
	if err != nil {
		return "", err
	}
	if len(items) == 0 {
		return "", errSourceUnavailable
	}

	content := items[s.index%len(items)]
	s.index++
	return content, nil
}

func (s *RedditSource) cachedOrFetched(ctx context.Context) ([]string, error) {
	if s.store != nil {
		cache, err := s.store.GetFreshPrimingCache(ctx, string(SourceCommunity), s.nowFn(), 24*time.Hour)
		if err != nil {
			return nil, fmt.Errorf("failed to read community cache: %w", err)
		}
		if cache != nil && len(cache.Payload) > 0 {
			log.Printf("{\"event\":\"priming_cache\",\"source\":%q,\"result\":\"hit\"}", SourceCommunity)
			return cache.Payload, nil
		}
	}

	log.Printf("{\"event\":\"priming_cache\",\"source\":%q,\"result\":\"miss\"}", SourceCommunity)
	items, err := s.fetchReddit(ctx)
	if err != nil {
		return nil, err
	}

	if s.store != nil && len(items) > 0 {
		err := s.store.SavePrimingCache(ctx, string(SourceCommunity), items, s.nowFn())
		if err != nil {
			return nil, fmt.Errorf("failed to persist community cache: %w", err)
		}
	}

	return items, nil
}

func (s *RedditSource) fetchReddit(ctx context.Context) ([]string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, redditURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to build reddit request: %w", err)
	}
	req.Header.Set("User-Agent", "dreams-night-priming/1.0")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch reddit source: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("reddit returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read reddit response: %w", err)
	}

	return parseRedditPosts(body)
}

func parseRedditPosts(body []byte) ([]string, error) {
	type childData struct {
		Title    string `json:"title"`
		Selftext string `json:"selftext"`
		Stickied bool   `json:"stickied"`
		Over18   bool   `json:"over_18"`
	}
	type listing struct {
		Data struct {
			Children []struct {
				Data childData `json:"data"`
			} `json:"children"`
		} `json:"data"`
	}

	var parsed listing
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, fmt.Errorf("failed to parse reddit payload: %w", err)
	}

	items := make([]string, 0, len(parsed.Data.Children))
	for _, child := range parsed.Data.Children {
		text, ok := renderTextRichPost(child.Data)
		if ok {
			items = append(items, text)
		}
	}

	return items, nil
}

func renderTextRichPost(post struct {
	Title    string `json:"title"`
	Selftext string `json:"selftext"`
	Stickied bool   `json:"stickied"`
	Over18   bool   `json:"over_18"`
}) (string, bool) {
	if post.Stickied || post.Over18 {
		return "", false
	}

	selfText := strings.TrimSpace(post.Selftext)
	if len(selfText) < 80 {
		return "", false
	}

	title := strings.TrimSpace(post.Title)
	if title == "" {
		title = "Lucid Dreaming"
	}

	return fmt.Sprintf("%s\n\n%s", title, selfText), true
}

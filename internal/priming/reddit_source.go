package priming

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"dreams/internal/model"
)

const (
	redditURL         = "https://old.reddit.com/r/LucidDreaming/comments/73ih3x/start_here_beginner_guides_faqs_and_resources/.json"
	redditFallbackURL = "https://www.reddit.com/r/LucidDreaming/comments/73ih3x/start_here_beginner_guides_faqs_and_resources/.json"
)

var markdownLinkPattern = regexp.MustCompile(`\[([^\]]+)\]\((https?://[^)\s]+)\)`)

type redditStatusError struct {
	statusCode int
}

func (e redditStatusError) Error() string {
	return fmt.Sprintf("reddit returned status %d", e.statusCode)
}

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
	items, err := s.fetchRedditFromURL(ctx, redditURL)
	if err == nil {
		return items, nil
	}

	items, fallbackErr := s.fetchRedditFromURL(ctx, redditFallbackURL)
	if fallbackErr == nil {
		return items, nil
	}

	var primaryStatusErr redditStatusError
	var fallbackStatusErr redditStatusError
	if errors.As(err, &primaryStatusErr) && errors.As(fallbackErr, &fallbackStatusErr) {
		if primaryStatusErr.statusCode == http.StatusForbidden && fallbackStatusErr.statusCode == http.StatusForbidden {
			return []string{starterGuidesFallbackContent()}, nil
		}
	}

	return nil, fmt.Errorf("%w; fallback source failed: %v", err, fallbackErr)
}

func (s *RedditSource) fetchRedditFromURL(ctx context.Context, url string) ([]string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to build reddit request: %w", err)
	}
	req.Header.Set("User-Agent", "dreams-night-priming/1.0")
	req.Header.Set("Accept", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch reddit source: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, redditStatusError{statusCode: resp.StatusCode}
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read reddit response: %w", err)
	}

	return parseRedditPosts(body)
}

func parseRedditPosts(body []byte) ([]string, error) {
	type postData struct {
		Title    string `json:"title"`
		Selftext string `json:"selftext"`
	}
	type listing struct {
		Data struct {
			Children []struct {
				Data postData `json:"data"`
			} `json:"children"`
		} `json:"data"`
	}

	var parsed []listing
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, fmt.Errorf("failed to parse reddit payload: %w", err)
	}
	if len(parsed) == 0 || len(parsed[0].Data.Children) == 0 {
		return nil, errSourceUnavailable
	}

	post := parsed[0].Data.Children[0].Data
	text, ok := renderGuideResourcePost(post.Title, post.Selftext)
	if !ok {
		return nil, errSourceUnavailable
	}

	return []string{text}, nil
}

func renderGuideResourcePost(title, selftext string) (string, bool) {
	links := extractGuideResourceLinks(selftext)
	if len(links) == 0 {
		return "", false
	}

	headline := strings.TrimSpace(title)
	if headline == "" {
		headline = "Lucid Dreaming Starter Resources"
	}

	var b strings.Builder
	b.WriteString(headline)
	b.WriteString("\n\nGuides and resources:\n")

	limit := len(links)
	if limit > 5 {
		limit = 5
	}

	for i := 0; i < limit; i++ {
		b.WriteString("- ")
		b.WriteString(links[i].label)
		b.WriteString(": ")
		b.WriteString(links[i].url)
		b.WriteString("\n")
	}

	b.WriteString("\nBedtime focus: Dream journal recall, reality checks, and clear intention for lucidity.")
	return b.String(), true
}

type resourceLink struct {
	label string
	url   string
}

func extractGuideResourceLinks(selftext string) []resourceLink {
	lines := strings.Split(selftext, "\n")
	links := make([]resourceLink, 0, 8)
	seen := map[string]struct{}{}

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		lineLower := strings.ToLower(line)
		if strings.Contains(lineLower, "rules and guidelines") {
			continue
		}
		if !isGuideResourceLine(lineLower) {
			continue
		}

		matches := markdownLinkPattern.FindAllStringSubmatch(line, -1)
		for _, match := range matches {
			if len(match) != 3 {
				continue
			}

			label := sanitizeResourceLabel(match[1])
			url := strings.TrimSpace(match[2])
			if label == "" || url == "" {
				continue
			}
			if _, ok := seen[url]; ok {
				continue
			}

			seen[url] = struct{}{}
			links = append(links, resourceLink{label: label, url: url})
		}
	}

	return links
}

func isGuideResourceLine(line string) bool {
	keywords := []string{"guide", "faq", "wiki", "resource", "meditation", "myth", "beginner"}
	for _, keyword := range keywords {
		if strings.Contains(line, keyword) {
			return true
		}
	}

	return false
}

func sanitizeResourceLabel(label string) string {
	label = strings.ReplaceAll(label, "*", "")
	label = strings.TrimSpace(label)
	label = strings.Join(strings.Fields(label), " ")
	return label
}

func starterGuidesFallbackContent() string {
	return strings.Join([]string{
		"START HERE! - Beginner Guides, FAQs, and Resources",
		"",
		"Guides and resources:",
		"- Wiki: https://www.reddit.com/r/LucidDreaming/wiki/index",
		"- FAQ: https://www.reddit.com/r/LucidDreaming/wiki/faq",
		"- Beginner Q&A (Part 1): https://www.reddit.com/r/LucidDreaming/comments/3iplpa/beginners_qa/",
		"- Beginner FAQ Extended (Part 2): https://www.reddit.com/r/LucidDreaming/comments/4cpb6o/beginners_faq_extended/",
		"- Three Steps Beginner Guide: https://www.reddit.com/r/LucidDreaming/comments/rsvp7/the_three_steps_for_learning_to_lucid_dream/",
		"",
		"Bedtime focus: Dream journal recall, reality checks, and clear intention for lucidity.",
	}, "\n")
}

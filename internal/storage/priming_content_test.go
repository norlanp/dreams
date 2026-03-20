package storage

import (
	"context"
	"testing"
)

func TestSeedPrimingContent(t *testing.T) {
	repo := createTestRepository(t)
	t.Cleanup(func() {
		_ = repo.Close()
	})

	ctx := context.Background()

	if err := repo.SeedPrimingContent(ctx); err != nil {
		t.Fatalf("failed to seed priming content: %v", err)
	}

	content, err := repo.ListPrimingContent(ctx)
	if err != nil {
		t.Fatalf("failed to list priming content: %v", err)
	}

	if len(content) == 0 {
		t.Fatal("expected priming content to be seeded, got none")
	}

	expectedCount := 5
	if len(content) != expectedCount {
		t.Fatalf("expected %d priming content items, got %d", expectedCount, len(content))
	}

	foundTitles := make(map[string]bool)
	for _, item := range content {
		foundTitles[item.Title] = true
	}

	expectedTitles := []string{
		"Lucid Dreaming Wiki",
		"Frequently Asked Questions",
		"Beginner Q&A Part 1",
		"Beginner FAQ Extended Part 2",
		"Myths and Misconceptions",
	}

	for _, title := range expectedTitles {
		if !foundTitles[title] {
			t.Errorf("expected to find title %q", title)
		}
	}

	if err := repo.SeedPrimingContent(ctx); err != nil {
		t.Fatalf("failed to re-seed priming content: %v", err)
	}

	content2, err := repo.ListPrimingContent(ctx)
	if err != nil {
		t.Fatalf("failed to list priming content after re-seed: %v", err)
	}

	if len(content2) != len(content) {
		t.Fatalf("re-seeding should not add duplicates: expected %d, got %d", len(content), len(content2))
	}

	beginnerContent, err := repo.GetPrimingContentByCategory(ctx, "beginner")
	if err != nil {
		t.Fatalf("failed to get priming content by category: %v", err)
	}

	if len(beginnerContent) == 0 {
		t.Fatal("expected beginner category content")
	}
}

package storage

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"dreams/internal/model"
	"dreams/internal/storage/sqlc"
)

func TestRepository_ShouldHydrateClusterJSONFieldsFromStorage(t *testing.T) {
	repo := createTestRepository(t)
	t.Cleanup(func() {
		_ = repo.Close()
	})

	ctx := context.Background()
	analysis, err := repo.SaveAnalysis(ctx, time.Now().UTC(), 3, 1, `{"ok":true}`)
	if err != nil {
		t.Fatalf("expected analysis save to succeed: %v", err)
	}

	_, err = repo.SaveCluster(ctx, analysis.ID, 1, 3, `["flight","teeth"]`, `[11,22,33]`)
	if err != nil {
		t.Fatalf("expected cluster save to succeed: %v", err)
	}

	clusters, err := repo.GetAnalysisClusters(ctx, analysis.ID)
	if err != nil {
		t.Fatalf("expected clusters lookup to succeed: %v", err)
	}

	if len(clusters) != 1 {
		t.Fatalf("expected one cluster, got %d", len(clusters))
	}

	if strings.Join(clusters[0].TopTerms, ",") != "flight,teeth" {
		t.Fatalf("expected top terms to be hydrated from JSON, got %#v", clusters[0].TopTerms)
	}

	if len(clusters[0].DreamIDs) != 3 || clusters[0].DreamIDs[0] != 11 || clusters[0].DreamIDs[2] != 33 {
		t.Fatalf("expected dream ids to be hydrated from JSON, got %#v", clusters[0].DreamIDs)
	}
}

func TestRepositoryGetLatestAnalysis_ShouldFailFastOnInvalidStoredTimestamp(t *testing.T) {
	repo := createTestRepository(t)
	t.Cleanup(func() {
		_ = repo.Close()
	})

	ctx := context.Background()
	_, err := repo.queries.CreateAnalysis(ctx, sqlc.CreateAnalysisParams{
		AnalysisDate: "not-a-timestamp",
		DreamCount:   1,
		NClusters:    1,
		ResultsJson:  `{"ok":false}`,
		CreatedAt:    sql.NullTime{Time: time.Now().UTC(), Valid: true},
	})
	if err != nil {
		t.Fatalf("expected raw analysis insert to succeed: %v", err)
	}

	analysis, err := repo.GetLatestAnalysis(ctx)
	if err == nil {
		t.Fatalf("expected parse error, got analysis %#v", analysis)
	}

	if !strings.Contains(err.Error(), "failed to parse analysis timestamp") {
		t.Fatalf("expected parse context in error, got %v", err)
	}
}

func TestRepositorySaveAnalysisWithClusters_ShouldRollbackOnClusterInsertFailure(t *testing.T) {
	repo := createTestRepository(t)
	t.Cleanup(func() {
		_ = repo.Close()
	})

	ctx := context.Background()
	_, err := repo.db.ExecContext(ctx, `
		CREATE TRIGGER fail_cluster_insert
		BEFORE INSERT ON dream_clusters
		BEGIN
			SELECT RAISE(ABORT, 'cluster write blocked');
		END;
	`)
	if err != nil {
		t.Fatalf("expected trigger creation to succeed: %v", err)
	}

	_, err = repo.SaveAnalysisWithClusters(
		ctx,
		time.Now().UTC(),
		3,
		1,
		`{"ok":true}`,
		[]model.Cluster{{
			ClusterID:  1,
			DreamCount: 3,
			TopTerms:   []string{"flight", "teeth"},
			DreamIDs:   []int64{11, 22, 33},
		}},
	)
	if err == nil {
		t.Fatal("expected atomic save to fail")
	}

	if !strings.Contains(err.Error(), "failed to create cluster") {
		t.Fatalf("expected cluster save context in error, got %v", err)
	}

	history, err := repo.ListAnalysisHistory(ctx)
	if err != nil {
		t.Fatalf("expected history lookup to succeed: %v", err)
	}

	if len(history) != 0 {
		t.Fatalf("expected no persisted analysis after rollback, got %d", len(history))
	}
}

func createTestRepository(t *testing.T) *Repository {
	t.Helper()

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}

	err = os.Chdir(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("failed to change working directory to workspace root: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(cwd)
	})

	err = os.MkdirAll("./tmp", 0o755)
	if err != nil {
		t.Fatalf("failed to create tmp directory: %v", err)
	}

	tmpDir, err := os.MkdirTemp("./tmp", "repository-test-")
	if err != nil {
		t.Fatalf("failed to create test directory: %v", err)
	}
	t.Cleanup(func() {
		_ = os.RemoveAll(tmpDir)
	})

	repo, err := NewRepository(filepath.Join(tmpDir, "dreams.db"))
	if err != nil {
		t.Fatalf("failed to create repository: %v", err)
	}

	return repo
}

func TestRepositoryPrimingCache_ShouldRespectFreshnessTTL(t *testing.T) {
	repo := createTestRepository(t)
	t.Cleanup(func() {
		_ = repo.Close()
	})

	ctx := context.Background()
	fetchedAt := time.Date(2025, 3, 10, 22, 0, 0, 0, time.UTC)
	err := repo.SavePrimingCache(ctx, "Community", []string{"post-a", "post-b"}, fetchedAt)
	if err != nil {
		t.Fatalf("expected cache save to succeed: %v", err)
	}

	hit, err := repo.GetFreshPrimingCache(ctx, "Community", fetchedAt.Add(23*time.Hour), 24*time.Hour)
	if err != nil {
		t.Fatalf("expected fresh cache read to succeed: %v", err)
	}
	if hit == nil || len(hit.Payload) != 2 {
		t.Fatalf("expected fresh cache hit with payload, got %#v", hit)
	}

	miss, err := repo.GetFreshPrimingCache(ctx, "Community", fetchedAt.Add(25*time.Hour), 24*time.Hour)
	if err != nil {
		t.Fatalf("expected stale cache read to succeed: %v", err)
	}
	if miss != nil {
		t.Fatalf("expected stale cache miss, got %#v", miss)
	}
}

func TestRepositoryPrimingLog_ShouldPersistDisplayOutcomes(t *testing.T) {
	repo := createTestRepository(t)
	t.Cleanup(func() {
		_ = repo.Close()
	})

	ctx := context.Background()
	now := time.Date(2025, 3, 10, 23, 0, 0, 0, time.UTC)
	err := repo.SavePrimingLog(ctx, "Template", "success", "Recovered via fallback", "Prime tonight", now)
	if err != nil {
		t.Fatalf("expected priming log save to succeed: %v", err)
	}

	logs, err := repo.ListPrimingLogs(ctx)
	if err != nil {
		t.Fatalf("expected priming logs listing to succeed: %v", err)
	}
	if len(logs) != 1 {
		t.Fatalf("expected one priming log, got %d", len(logs))
	}
	if logs[0].Source != "Template" || logs[0].Outcome != "success" {
		t.Fatalf("expected persisted source/outcome, got %#v", logs[0])
	}
}
